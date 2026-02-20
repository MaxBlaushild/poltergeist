import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter/services.dart';

import '../constants/api_constants.dart';
import '../providers/auth_provider.dart';
import '../services/media_service.dart';

class LogisterModal extends StatelessWidget {
  const LogisterModal({
    super.key,
    required this.onSuccess,
    required this.onSkip,
    required this.mediaService,
  });

  final VoidCallback onSuccess;
  final VoidCallback onSkip;
  final MediaService mediaService;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Center(
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 980),
            child: DecoratedBox(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    colorScheme.surface.withOpacity(0.96),
                    colorScheme.surfaceVariant.withOpacity(0.94),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(28),
                border: Border.all(color: colorScheme.outlineVariant),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.08),
                    blurRadius: 24,
                    offset: const Offset(0, 16),
                  ),
                ],
              ),
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Consumer<AuthProvider>(
                  builder: (context, auth, _) {
                    return LayoutBuilder(
                      builder: (context, constraints) {
                        final isWide = constraints.maxWidth > 720;
                        final intro = _LogisterIntroPanel(
                          onSkip: onSkip,
                        );
                        final form = Container(
                          padding: const EdgeInsets.all(20),
                          decoration: BoxDecoration(
                            color: colorScheme.surface,
                            borderRadius: BorderRadius.circular(22),
                            border: Border.all(color: colorScheme.outlineVariant),
                          ),
                          child: _LogisterForm(
                            auth: auth,
                            mediaService: mediaService,
                            onSuccess: onSuccess,
                            onSkip: onSkip,
                          ),
                        );

                        if (isWide) {
                          return Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Expanded(child: intro),
                              const SizedBox(width: 24),
                              Expanded(child: form),
                            ],
                          );
                        }

                        return Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            intro,
                            const SizedBox(height: 24),
                            form,
                          ],
                        );
                      },
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
    required this.mediaService,
    required this.onSuccess,
    required this.onSkip,
  });

  final AuthProvider auth;
  final MediaService mediaService;
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
  PlatformFile? _pickedFile;

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
      if (!needsProfile && mounted) widget.onSuccess();
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Future<void> _pickFile() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.image,
      withData: true,
    );
    if (result != null && result.files.single.bytes != null) {
      setState(() => _pickedFile = result.files.single);
    }
  }

  Future<void> _submitProfileSetup() async {
    final username = _nameController.text.trim();
    final hasFile = _pickedFile != null && _pickedFile!.bytes != null;
    final hasUsername = username.length >= 2;
    if (!hasFile && !hasUsername) return;
    setState(() => _loading = true);
    try {
      String? profilePictureUrl;
      if (_pickedFile != null && _pickedFile!.bytes != null) {
        final user = widget.auth.user;
        if (user == null) {
          setState(() => _loading = false);
          return;
        }
        final ext = _pickedFile!.name.split('.').last.toLowerCase();
        if (ext.isEmpty) return;
        final key =
            '${user.id}-${DateTime.now().millisecondsSinceEpoch}.$ext';
        final url = await widget.mediaService.getPresignedUploadUrl(
          ApiConstants.crewProfileBucket,
          key,
        );
        if (url == null) {
          setState(() => _loading = false);
          return;
        }
        final contentType = ext == 'png'
            ? 'image/png'
            : ext == 'gif'
                ? 'image/gif'
                : 'image/jpeg';
        final ok = await widget.mediaService.uploadToPresigned(
          url,
          Uint8List.fromList(_pickedFile!.bytes!),
          contentType,
        );
        if (!ok) {
          setState(() => _loading = false);
          return;
        }
        profilePictureUrl = url.split('?').first;
      }
      await widget.auth.updateProfile(
        username: hasUsername ? username : null,
        profilePictureUrl: profilePictureUrl,
      );
      setState(() => _loading = false);
      if (mounted) widget.onSuccess();
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = widget.auth;
    final waiting = auth.isWaitingForVerificationCode;

    if (_showProfileSetup) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Forge your identity',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 16),
          Text(
            'Choose a call sign and crest so your crew can recognize you on the map.',
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
          const SizedBox(height: 12),
          ListTile(
            title: Text(_pickedFile?.name ?? 'Upload profile picture'),
            trailing: IconButton(
              icon: const Icon(Icons.upload_file),
              onPressed: _pickFile,
            ),
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: (_loading ||
                    ((_pickedFile == null || _pickedFile!.bytes == null) &&
                        _nameController.text.trim().length < 2))
                ? null
                : _submitProfileSetup,
            child: _loading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Set profile'),
          ),
        ],
      );
    }

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Signal the expedition gate',
          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.w600,
              ),
        ),
        const SizedBox(height: 8),
        Text(
          'Enter your phone number to receive the access pulse. Once verified, you can explore the world, gain strength, and complete quests with your crew.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.75),
              ),
        ),
        const SizedBox(height: 16),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: const [
            _StepChip(label: '1. Send signal'),
            _StepChip(label: '2. Confirm code'),
            _StepChip(label: '3. Launch map'),
          ],
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
              inputFormatters: [
                FilteringTextInputFormatter.digitsOnly,
              ],
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
            "We've sent a 6-digit access pulse. It may take a moment to arrive.",
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
                    : const Text('Send signal'),
              ),
            const SizedBox(width: 12),
            TextButton(
              onPressed: widget.onSkip,
              child: const Text('Return to landing'),
            ),
          ],
        ),
      ],
    );
  }
}

class _LogisterIntroPanel extends StatelessWidget {
  const _LogisterIntroPanel({required this.onSkip});

  final VoidCallback onSkip;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Welcome to the Sonar Dart expedition',
          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.w600,
              ),
        ),
        const SizedBox(height: 12),
        Text(
          'You are about to step into a world where real places hide mythic layers. Discover new locations, gain strength with every mission, and complete quests to unlock the next realm.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: colorScheme.onSurface.withOpacity(0.75),
                height: 1.4,
              ),
        ),
        const SizedBox(height: 20),
        Wrap(
          spacing: 12,
          runSpacing: 12,
          children: const [
            _LegendItem(
              icon: Icons.public,
              title: 'Explore the world',
              description: 'Track real-world landmarks and hidden routes.',
            ),
            _LegendItem(
              icon: Icons.auto_awesome,
              title: 'Discover mythic sites',
              description: 'Find ley lines and spectral waypoints.',
            ),
            _LegendItem(
              icon: Icons.fitness_center,
              title: 'Gain strength',
              description: 'Earn upgrades and new abilities.',
            ),
            _LegendItem(
              icon: Icons.task_alt,
              title: 'Complete quests',
              description: 'Finish chains to open new realms.',
            ),
          ],
        ),
        const SizedBox(height: 20),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: colorScheme.primary.withOpacity(0.1),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: colorScheme.primary.withOpacity(0.3)),
          ),
          child: Row(
            children: [
              Icon(Icons.wifi_tethering, color: colorScheme.primary),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  'Signal status: strong. Your crew is ready for launch.',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurface.withOpacity(0.8),
                      ),
                ),
              ),
              TextButton(
                onPressed: onSkip,
                child: const Text('Not now'),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _LegendItem extends StatelessWidget {
  const _LegendItem({
    required this.icon,
    required this.title,
    required this.description,
  });

  final IconData icon;
  final String title;
  final String description;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      width: 220,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: colorScheme.surface.withOpacity(0.92),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: colorScheme.primary),
          const SizedBox(height: 8),
          Text(
            title,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 6),
          Text(
            description,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: colorScheme.onSurface.withOpacity(0.7),
                ),
          ),
        ],
      ),
    );
  }
}

class _StepChip extends StatelessWidget {
  const _StepChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: colorScheme.primary.withOpacity(0.12),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: colorScheme.primary.withOpacity(0.35)),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: colorScheme.primary,
              fontWeight: FontWeight.w600,
            ),
      ),
    );
  }
}
