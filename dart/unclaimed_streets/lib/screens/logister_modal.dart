import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/notification_permission_service.dart';
import '../services/push_notification_service.dart';

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
            constraints: const BoxConstraints(maxWidth: 560),
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: colorScheme.surface.withOpacity(0.97),
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: colorScheme.outlineVariant),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.08),
                    blurRadius: 20,
                    offset: const Offset(0, 12),
                  ),
                ],
              ),
              child: Padding(
                padding: const EdgeInsets.all(24),
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
  bool _loading = false;
  bool _showProfileSetup = false;
  bool _showNotificationSetup = false;
  bool _notificationLoading = false;
  final NotificationPermissionService _notificationPermissionService =
      NotificationPermissionService();
  NotificationPermissionState _notificationPermissionState =
      NotificationPermissionState.notDetermined;

  @override
  void dispose() {
    _countryCodeController.dispose();
    _phoneController.dispose();
    _codeController.dispose();
    _nameController.dispose();
    super.dispose();
  }

  String _formattedPhoneNumber() {
    final code = _countryCodeController.text.replaceAll(RegExp(r'\D'), '');
    final local = _phoneController.text.replaceAll(RegExp(r'\D'), '');
    if (code.isEmpty && local.isEmpty) return '';
    return '+$code$local';
  }

  Future<void> _getCode() async {
    final phone = _formattedPhoneNumber();
    if (phone.isEmpty) return;
    setState(() => _loading = true);
    await widget.auth.getVerificationCode(phone);
    setState(() => _loading = false);
  }

  Future<void> _submit() async {
    final phone = _formattedPhoneNumber();
    final code = _codeController.text.trim();
    if (phone.isEmpty || code.isEmpty) return;
    setState(() => _loading = true);
    try {
      final needsProfile = await widget.auth.logister(phone, code);
      setState(() {
        _loading = false;
        _showProfileSetup = needsProfile;
      });
      if (!needsProfile && mounted) {
        await _registerPushTokenForCurrentUser(force: false);
        if (!mounted) return;
        widget.onSuccess();
      }
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Future<void> _submitProfileSetup() async {
    final username = _nameController.text.trim();
    final hasUsername = username.length >= 2;
    if (!hasUsername) return;
    setState(() => _loading = true);
    try {
      if (widget.auth.isDryRunRegistrationActive) {
        await widget.auth.logout();
        if (!mounted) return;
        setState(() => _loading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Dry-run registration complete. No user was saved.'),
          ),
        );
        widget.onSkip();
        return;
      }

      await widget.auth.updateProfile(username: username);
      if (!mounted) return;
      setState(() {
        _loading = false;
        _showProfileSetup = false;
        _showNotificationSetup = true;
      });
      await _loadNotificationPermissionState();
    } catch (_) {
      setState(() => _loading = false);
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

  @override
  Widget build(BuildContext context) {
    final auth = widget.auth;
    final waiting = auth.isWaitingForVerificationCode;
    final colorScheme = Theme.of(context).colorScheme;

    if (_showProfileSetup) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Set up your profile',
            style: Theme.of(
              context,
            ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 16),
          Text(
            'Choose a username so your crew can recognize you.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              color: Theme.of(context).colorScheme.onSurface.withOpacity(0.75),
            ),
          ),
          const SizedBox(height: 20),
          TextField(
            controller: _nameController,
            decoration: const InputDecoration(
              labelText: 'Username',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: (_loading || _nameController.text.trim().length < 2)
                ? null
                : _submitProfileSetup,
            child: _loading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Continue'),
          ),
        ],
      );
    }

    if (_showNotificationSetup) {
      final enabled =
          _notificationPermissionState == NotificationPermissionState.granted;
      final canRequest =
          _notificationPermissionState !=
          NotificationPermissionState.unsupported;
      return Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Stay in the loop',
            style: Theme.of(
              context,
            ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 16),
          Text(
            'Turn on notifications for party invites and combat updates.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              color: colorScheme.onSurface.withOpacity(0.75),
            ),
          ),
          const SizedBox(height: 20),
          Container(
            decoration: BoxDecoration(
              color: colorScheme.surfaceVariant.withOpacity(0.45),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: colorScheme.outlineVariant),
            ),
            child: SwitchListTile(
              value: enabled,
              onChanged: (_notificationLoading || !canRequest)
                  ? null
                  : _onNotificationToggle,
              title: const Text('Allow push notifications'),
              subtitle: Text(_notificationStatusText()),
              secondary: _notificationLoading
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.notifications_active_outlined),
            ),
          ),
          if (_notificationPermissionState ==
              NotificationPermissionState.denied)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Text(
                'Notifications are currently blocked. You can continue and enable them later in browser settings.',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: colorScheme.onSurface.withOpacity(0.72),
                ),
              ),
            ),
          const SizedBox(height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              FilledButton.icon(
                onPressed: _completeNotificationSetup,
                icon: const Icon(Icons.arrow_forward),
                label: const Text('Continue'),
              ),
              const SizedBox(width: 12),
              TextButton(
                onPressed: _completeNotificationSetup,
                child: const Text('Skip for now'),
              ),
            ],
          ),
        ],
      );
    }

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Sign in to Unclaimed Streets',
          style: Theme.of(
            context,
          ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        Text(
          'Enter your phone number to receive a code.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: Theme.of(context).colorScheme.onSurface.withOpacity(0.75),
          ),
        ),
        const SizedBox(height: 16),
        if (auth.error != null)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              auth.error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
        Builder(
          builder: (context) {
            final isNarrow = MediaQuery.of(context).size.width < 420;
            final countryField = TextField(
              controller: _countryCodeController,
              decoration: const InputDecoration(
                labelText: 'Country code',
                prefixText: '+',
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            );
            final phoneField = TextField(
              controller: _phoneController,
              decoration: const InputDecoration(
                labelText: 'Phone number',
                hintText: '234 567 8900',
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.phone,
              onSubmitted: (_) => _getCode(),
            );

            if (isNarrow) {
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
                SizedBox(width: 140, child: countryField),
                const SizedBox(width: 12),
                Expanded(child: phoneField),
              ],
            );
          },
        ),
        if (waiting) ...[
          const SizedBox(height: 12),
          const Text(
            "We've sent a 6-digit verification code. It may take a moment to arrive.",
            style: TextStyle(fontSize: 12),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _codeController,
            decoration: const InputDecoration(
              labelText: 'Verification code',
              border: OutlineInputBorder(),
            ),
            keyboardType: TextInputType.number,
            maxLength: 6,
          ),
        ],
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (waiting)
              FilledButton.icon(
                onPressed: _loading ? null : _submit,
                icon: const Icon(Icons.lock_open),
                label: _loading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Enter gate'),
              )
            else
              FilledButton.icon(
                onPressed: _loading ? null : _getCode,
                icon: const Icon(Icons.wifi_tethering),
                label: _loading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Send code'),
              ),
            const SizedBox(width: 12),
            TextButton(onPressed: widget.onSkip, child: const Text('Back')),
          ],
        ),
      ],
    );
  }
}
