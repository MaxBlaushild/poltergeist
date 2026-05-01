import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/notification_permission_service.dart';
import '../services/push_notification_service.dart';

enum _LogisterStage { phone, profile, notifications }

class LogisterModal extends StatelessWidget {
  const LogisterModal({
    super.key,
    required this.onSuccess,
    required this.onSkip,
  });

  final VoidCallback onSuccess;
  final VoidCallback onSkip;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Center(
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 440),
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: colorScheme.surface.withValues(alpha: 0.98),
                borderRadius: BorderRadius.circular(28),
                border: Border.all(
                  color: colorScheme.outlineVariant.withValues(alpha: 0.7),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.06),
                    blurRadius: 36,
                    offset: const Offset(0, 18),
                  ),
                ],
              ),
              child: Padding(
                padding: const EdgeInsets.all(28),
                child: Consumer<AuthProvider>(
                  builder: (context, auth, _) {
                    return _LogisterForm(
                      auth: auth,
                      onSuccess: onSuccess,
                      onSkip: onSkip,
                    );
                  },
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _LogisterForm extends StatefulWidget {
  const _LogisterForm({
    required this.auth,
    required this.onSuccess,
    required this.onSkip,
  });

  final AuthProvider auth;
  final VoidCallback onSuccess;
  final VoidCallback onSkip;

  @override
  State<_LogisterForm> createState() => _LogisterFormState();
}

class _LogisterFormState extends State<_LogisterForm> {
  final _countryCodeController = TextEditingController(text: '1');
  final _phoneController = TextEditingController();
  final _codeController = TextEditingController();
  final _nameController = TextEditingController();
  bool _requestingCode = false;
  bool _submittingCode = false;
  bool _submittingProfile = false;
  bool _showProfileSetup = false;
  bool _showNotificationSetup = false;
  bool _notificationLoading = false;
  final NotificationPermissionService _notificationPermissionService =
      NotificationPermissionService();
  NotificationPermissionState _notificationPermissionState =
      NotificationPermissionState.notDetermined;

  @override
  void initState() {
    super.initState();
    _countryCodeController.addListener(_handleFormChanged);
    _phoneController.addListener(_handleFormChanged);
    _codeController.addListener(_handleFormChanged);
    _nameController.addListener(_handleFormChanged);
  }

  @override
  void dispose() {
    _countryCodeController.removeListener(_handleFormChanged);
    _phoneController.removeListener(_handleFormChanged);
    _codeController.removeListener(_handleFormChanged);
    _nameController.removeListener(_handleFormChanged);
    _countryCodeController.dispose();
    _phoneController.dispose();
    _codeController.dispose();
    _nameController.dispose();
    super.dispose();
  }

  void _handleFormChanged() {
    if (!mounted) return;
    if (widget.auth.error != null) {
      widget.auth.clearError();
    }
    setState(() {});
  }

  String _formattedPhoneNumber() {
    final code = _countryCodeController.text.replaceAll(RegExp(r'\D'), '');
    final local = _phoneController.text.replaceAll(RegExp(r'\D'), '');
    if (code.isEmpty && local.isEmpty) return '';
    return '+$code$local';
  }

  Future<void> _getCode({bool showConfirmation = false}) async {
    final phone = _formattedPhoneNumber();
    if (phone.isEmpty) return;
    setState(() => _requestingCode = true);
    final sent = await widget.auth.getVerificationCode(phone);
    if (!mounted) return;
    setState(() => _requestingCode = false);
    if (sent && showConfirmation) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Sent a fresh verification code.')),
      );
    }
  }

  Future<void> _submit() async {
    final phone = _formattedPhoneNumber();
    final code = _codeController.text.trim();
    if (phone.isEmpty || code.isEmpty) return;
    setState(() => _submittingCode = true);
    try {
      final needsProfile = await widget.auth.logister(phone, code);
      if (!mounted) return;
      setState(() {
        _submittingCode = false;
        _showProfileSetup = needsProfile;
      });
      if (!needsProfile && mounted) {
        await _registerPushTokenForCurrentUser(force: false);
        if (!mounted) return;
        widget.onSuccess();
      }
    } catch (_) {
      if (!mounted) return;
      setState(() => _submittingCode = false);
    }
  }

  Future<void> _submitProfileSetup() async {
    final username = _nameController.text.trim();
    final hasUsername = username.length >= 2;
    if (!hasUsername) return;
    setState(() => _submittingProfile = true);
    try {
      if (widget.auth.isDryRunRegistrationActive) {
        await widget.auth.logout();
        if (!mounted) return;
        setState(() => _submittingProfile = false);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Dry-run registration complete. No user was saved.'),
          ),
        );
        widget.onSkip();
        return;
      }

      await widget.auth.completeRegistration(username: username);
      if (!mounted) return;
      setState(() {
        _submittingProfile = false;
        _showProfileSetup = false;
        _showNotificationSetup = true;
      });
      await _loadNotificationPermissionState();
    } catch (_) {
      if (!mounted) return;
      setState(() => _submittingProfile = false);
    }
  }

  Future<void> _loadNotificationPermissionState() async {
    setState(() => _notificationLoading = true);
    try {
      final state = await _notificationPermissionService.getPermissionState();
      if (!mounted) return;
      setState(() {
        _notificationPermissionState = state;
        _notificationLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _notificationLoading = false);
    }
  }

  Future<void> _onNotificationToggle(bool value) async {
    if (!value) {
      if (_notificationPermissionState == NotificationPermissionState.granted &&
          mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Notifications are enabled. Use browser/system settings to disable.',
            ),
          ),
        );
      }
      return;
    }

    setState(() => _notificationLoading = true);
    try {
      final state = await _notificationPermissionService.requestPermission();
      if (!mounted) return;
      if (state == NotificationPermissionState.granted) {
        await _registerPushTokenForCurrentUser(force: true);
      }
      setState(() {
        _notificationPermissionState = state;
        _notificationLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _notificationLoading = false);
    }
  }

  String _notificationStatusText() {
    switch (_notificationPermissionState) {
      case NotificationPermissionState.granted:
        return 'Enabled';
      case NotificationPermissionState.denied:
        return 'Blocked';
      case NotificationPermissionState.unsupported:
        return 'Not supported on this device/browser';
      case NotificationPermissionState.notDetermined:
        return 'Not enabled yet';
    }
  }

  Future<void> _registerPushTokenForCurrentUser({required bool force}) async {
    final userId = widget.auth.user?.id;
    if (userId == null || userId.isEmpty) return;
    try {
      await context.read<PushNotificationService>().registerDeviceTokenForUser(
        userId,
        force: force,
      );
    } catch (_) {}
  }

  Future<void> _completeNotificationSetup() async {
    if (_notificationPermissionState == NotificationPermissionState.granted) {
      await _registerPushTokenForCurrentUser(force: false);
    }
    widget.auth.completeRegistrationFlow();
    if (!mounted) return;
    widget.onSuccess();
  }

  void _handleBack() {
    if (_showProfileSetup ||
        _showNotificationSetup ||
        widget.auth.isWaitingForVerificationCode) {
      widget.auth.cancelRegistrationFlow();
    }
    widget.onSkip();
  }

  _LogisterStage get _stage {
    if (_showProfileSetup) return _LogisterStage.profile;
    if (_showNotificationSetup) return _LogisterStage.notifications;
    return _LogisterStage.phone;
  }

  String get _phonePreview {
    final phone = _formattedPhoneNumber();
    return phone.isEmpty ? 'your phone' : phone;
  }

  InputDecoration _inputDecoration(
    BuildContext context,
    String label, {
    String? hintText,
    String? prefixText,
  }) {
    final colorScheme = Theme.of(context).colorScheme;
    return InputDecoration(
      labelText: label,
      hintText: hintText,
      prefixText: prefixText,
      filled: true,
      fillColor: colorScheme.surfaceContainerHighest.withValues(alpha: 0.26),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 18),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(16),
        borderSide: BorderSide(color: colorScheme.outlineVariant),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(16),
        borderSide: BorderSide(color: colorScheme.outlineVariant),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(16),
        borderSide: BorderSide(color: colorScheme.primary, width: 1.4),
      ),
    );
  }

  Widget _buildErrorBanner(BuildContext context, String message) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: colorScheme.errorContainer.withValues(alpha: 0.7),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            Icons.error_outline,
            size: 18,
            color: colorScheme.onErrorContainer,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: colorScheme.onErrorContainer,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPhoneStage(
    BuildContext context, {
    required bool waiting,
    required bool isCodeActionBusy,
    required bool canRequestCode,
    required bool canSubmitCode,
  }) {
    final colorScheme = Theme.of(context).colorScheme;
    final countryField = TextField(
      controller: _countryCodeController,
      decoration: _inputDecoration(
        context,
        'Country',
        hintText: '1',
        prefixText: '+',
      ),
      keyboardType: TextInputType.number,
      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
      textInputAction: TextInputAction.next,
      autofillHints: const [AutofillHints.telephoneNumberCountryCode],
    );
    final phoneField = TextField(
      controller: _phoneController,
      decoration: _inputDecoration(
        context,
        'Phone number',
        hintText: '234 567 8900',
      ),
      keyboardType: TextInputType.phone,
      textInputAction: waiting ? TextInputAction.next : TextInputAction.go,
      autofillHints: const [AutofillHints.telephoneNumber],
      onSubmitted: (_) {
        if (!waiting) {
          _getCode();
        }
      },
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          waiting
              ? 'Enter the 6-digit code sent to $_phonePreview.'
              : 'Use your phone number to enter StreetSekai.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: colorScheme.onSurface.withValues(alpha: 0.72),
            height: 1.45,
          ),
        ),
        const SizedBox(height: 20),
        LayoutBuilder(
          builder: (context, constraints) {
            if (constraints.maxWidth < 360) {
              return Column(
                children: [
                  countryField,
                  const SizedBox(height: 12),
                  phoneField,
                ],
              );
            }

            return Row(
              children: [
                SizedBox(width: 112, child: countryField),
                const SizedBox(width: 12),
                Expanded(child: phoneField),
              ],
            );
          },
        ),
        if (waiting) ...[
          const SizedBox(height: 12),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: colorScheme.surfaceContainerHighest.withValues(
                alpha: 0.24,
              ),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Text(
              'Code sent to $_phonePreview. It may take a moment to arrive.',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: colorScheme.onSurface.withValues(alpha: 0.72),
              ),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _codeController,
            decoration: _inputDecoration(
              context,
              'Verification code',
              hintText: '123456',
            ),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            maxLength: 6,
            textInputAction: TextInputAction.done,
            onSubmitted: (_) {
              if (canSubmitCode && !isCodeActionBusy) {
                _submit();
              }
            },
            autofillHints: const [AutofillHints.oneTimeCode],
          ),
        ],
        const SizedBox(height: 8),
        SizedBox(
          width: double.infinity,
          child: FilledButton(
            onPressed: waiting
                ? (isCodeActionBusy || !canSubmitCode ? null : _submit)
                : (isCodeActionBusy || !canRequestCode ? null : _getCode),
            style: FilledButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
            child: waiting
                ? (_submittingCode
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Continue'))
                : (_requestingCode
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Send code')),
          ),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            if (waiting)
              TextButton(
                onPressed: (isCodeActionBusy || !canRequestCode)
                    ? null
                    : () => _getCode(showConfirmation: true),
                child: Text(_requestingCode ? 'Sending...' : 'Resend code'),
              ),
            const Spacer(),
            TextButton(onPressed: _handleBack, child: const Text('Back')),
          ],
        ),
      ],
    );
  }

  Widget _buildProfileStage(BuildContext context, {required bool canSubmit}) {
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Choose the name your party will know you by.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: colorScheme.onSurface.withValues(alpha: 0.72),
            height: 1.45,
          ),
        ),
        const SizedBox(height: 20),
        TextField(
          controller: _nameController,
          decoration: _inputDecoration(
            context,
            'Username',
            hintText: 'Mapwalker',
          ),
          textInputAction: TextInputAction.done,
          onSubmitted: (_) {
            if (canSubmit && !_submittingProfile) {
              _submitProfileSetup();
            }
          },
          autofillHints: const [AutofillHints.username],
        ),
        const SizedBox(height: 20),
        SizedBox(
          width: double.infinity,
          child: FilledButton(
            onPressed: (_submittingProfile || !canSubmit)
                ? null
                : _submitProfileSetup,
            style: FilledButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
            child: _submittingProfile
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Continue'),
          ),
        ),
        const SizedBox(height: 12),
        Align(
          alignment: Alignment.centerRight,
          child: TextButton(onPressed: _handleBack, child: const Text('Back')),
        ),
      ],
    );
  }

  Widget _buildNotificationStage(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final enabled =
        _notificationPermissionState == NotificationPermissionState.granted;
    final canRequest =
        _notificationPermissionState != NotificationPermissionState.unsupported;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Turn on notifications for party invites, quest nudges, and combat updates. You can always change this later.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: colorScheme.onSurface.withValues(alpha: 0.72),
            height: 1.45,
          ),
        ),
        const SizedBox(height: 20),
        Container(
          decoration: BoxDecoration(
            color: colorScheme.surfaceContainerHighest.withValues(alpha: 0.22),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: colorScheme.outlineVariant),
          ),
          child: SwitchListTile(
            value: enabled,
            onChanged: (_notificationLoading || !canRequest)
                ? null
                : _onNotificationToggle,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 14,
              vertical: 4,
            ),
            title: const Text('Allow notifications'),
            subtitle: Text(_notificationStatusText()),
            secondary: _notificationLoading
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.notifications_none_rounded),
          ),
        ),
        if (_notificationPermissionState == NotificationPermissionState.denied)
          Padding(
            padding: const EdgeInsets.only(top: 12),
            child: Text(
              'Notifications are blocked right now. You can finish setup and enable them later in browser or system settings.',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: colorScheme.onSurface.withValues(alpha: 0.7),
                height: 1.4,
              ),
            ),
          ),
        const SizedBox(height: 20),
        SizedBox(
          width: double.infinity,
          child: FilledButton(
            onPressed: _completeNotificationSetup,
            style: FilledButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
            child: const Text('Finish setup'),
          ),
        ),
        const SizedBox(height: 12),
        Align(
          alignment: Alignment.centerRight,
          child: TextButton(
            onPressed: _completeNotificationSetup,
            child: const Text('Skip for now'),
          ),
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final auth = widget.auth;
    final waiting = auth.isWaitingForVerificationCode;
    final colorScheme = Theme.of(context).colorScheme;
    final isCodeActionBusy = _requestingCode || _submittingCode;
    final canRequestCode = _formattedPhoneNumber().isNotEmpty;
    final canSubmitCode =
        _formattedPhoneNumber().isNotEmpty &&
        _codeController.text.trim().length == 6;
    final canSubmitProfile = _nameController.text.trim().length >= 2;
    final stage = _stage;
    final stepLabel = switch (stage) {
      _LogisterStage.phone => 'Step 1 of 3',
      _LogisterStage.profile => 'Step 2 of 3',
      _LogisterStage.notifications => 'Step 3 of 3',
    };
    final title = switch (stage) {
      _LogisterStage.phone => waiting ? 'Check your code' : 'Sign in',
      _LogisterStage.profile => 'Set up your profile',
      _LogisterStage.notifications => 'Notifications',
    };

    final body = switch (stage) {
      _LogisterStage.phone => _buildPhoneStage(
        context,
        waiting: waiting,
        isCodeActionBusy: isCodeActionBusy,
        canRequestCode: canRequestCode,
        canSubmitCode: canSubmitCode,
      ),
      _LogisterStage.profile => _buildProfileStage(
        context,
        canSubmit: canSubmitProfile,
      ),
      _LogisterStage.notifications => _buildNotificationStage(context),
    };

    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 220),
      switchInCurve: Curves.easeOutCubic,
      switchOutCurve: Curves.easeInCubic,
      child: KeyedSubtree(
        key: ValueKey(stage),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'StreetSekai',
                        style: Theme.of(context).textTheme.labelLarge?.copyWith(
                          letterSpacing: 1.2,
                          fontWeight: FontWeight.w700,
                          color: colorScheme.primary,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        title,
                        style: Theme.of(context).textTheme.headlineSmall
                            ?.copyWith(fontWeight: FontWeight.w700),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 12),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    IconButton(
                      onPressed: _handleBack,
                      tooltip: 'Close',
                      visualDensity: VisualDensity.compact,
                      icon: const Icon(Icons.close),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: colorScheme.surfaceContainerHighest.withValues(
                          alpha: 0.4,
                        ),
                        borderRadius: BorderRadius.circular(999),
                      ),
                      child: Text(
                        stepLabel,
                        style: Theme.of(context).textTheme.labelMedium
                            ?.copyWith(
                              color: colorScheme.onSurface.withValues(
                                alpha: 0.7,
                              ),
                            ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
            if (auth.error != null) ...[
              const SizedBox(height: 18),
              _buildErrorBanner(context, auth.error!),
            ],
            const SizedBox(height: 20),
            body,
          ],
        ),
      ),
    );
  }
}
