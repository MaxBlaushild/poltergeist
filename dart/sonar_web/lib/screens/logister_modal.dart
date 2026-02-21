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
                      mediaService: mediaService,
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
            'Set up your profile',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 16),
          Text(
            'Add a name or a profile image so your crew can recognize you.',
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
          'Sign in to Unclaimed Streets',
          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.w600,
              ),
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
            TextButton(
              onPressed: widget.onSkip,
              child: const Text('Back'),
            ),
          ],
        ),
      ],
    );
  }
}
