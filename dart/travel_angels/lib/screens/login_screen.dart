import 'dart:ui' show ImageDescriptor, ImmutableBuffer;
import 'package:flutter/material.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:file_picker/file_picker.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/services/location_service.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/widgets/location_picker.dart';
import 'package:travel_angels/utils/auth_phone_number.dart';
import 'package:travel_angels/widgets/international_phone_field.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({
    super.key,
    this.mediaService,
    this.locationService,
    this.showLocationMap,
  });

  final MediaService? mediaService;
  final LocationService? locationService;
  final bool? showLocationMap;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  String? _phoneNumber;
  final _codeController = TextEditingController();
  final _usernameController = TextEditingController();
  final _phoneFocusNode = FocusNode();
  final _codeFocusNode = FocusNode();
  final _usernameFocusNode = FocusNode();

  bool _isRequestingCode = false;
  bool _isSigningIn = false;
  bool _shouldSetProfile = false;
  Uint8List? _profilePicture;
  String? _profilePictureName;
  bool _pickingProfilePicture = false;
  int _usernameRevision = 0;
  bool _isValidatingUsername = false;
  String? _usernameValidationError;
  bool _isUsernameValid = false;

  // New registration fields
  DateTime? _dateOfBirth;
  String? _selectedGender;
  double? _selectedLatitude;
  double? _selectedLongitude;
  String? _selectedLocationAddress;
  final _bioController = TextEditingController();

  @override
  void dispose() {
    _codeController.dispose();
    _usernameController.dispose();
    _bioController.dispose();
    _phoneFocusNode.dispose();
    _codeFocusNode.dispose();
    _usernameFocusNode.dispose();
    super.dispose();
  }

  Future<void> _handleGetCode() async {
    if (_phoneNumber == null || _isRequestingCode) return;

    final authProvider = context.read<AuthProvider>();
    final phoneNumber = _phoneNumber!;
    setState(() => _isRequestingCode = true);

    try {
      await authProvider.getVerificationCode(phoneNumber);
      if (!mounted) return;
      if (authProvider.isWaitingForVerificationCode) {
        _codeFocusNode.requestFocus();
      }
    } catch (e) {
      // Error is handled by AuthProvider and displayed via error getter
    } finally {
      if (mounted) setState(() => _isRequestingCode = false);
    }
  }

  Future<void> _handleLogister() async {
    if (_codeController.text.length != 6 || _isSigningIn) return;

    final authProvider = context.read<AuthProvider>();
    final phoneNumber =
        authProvider.verificationPhoneNumber ?? _phoneNumber ?? '';
    final code = _codeController.text.trim();
    setState(() => _isSigningIn = true);

    try {
      await authProvider.logister(phoneNumber, code);
      if (!mounted) return;

      // If registering, show profile setup step
      if (authProvider.isRegister) {
        setState(() {
          _shouldSetProfile = true;
        });
      } else {
        // Navigation will be handled by main.dart based on auth state
      }
    } catch (e) {
      // Error is handled by AuthProvider and displayed via error getter
    } finally {
      if (mounted) setState(() => _isSigningIn = false);
    }
  }

  Future<void> _validateUsername(String username) async {
    final revision = ++_usernameRevision;
    if (username.isEmpty) {
      setState(() {
        _isValidatingUsername = false;
        _usernameValidationError = null;
        _isUsernameValid = false;
      });
      return;
    }

    // Basic format validation (alphanumeric only)
    if (!RegExp(r'^[a-zA-Z0-9]+$').hasMatch(username)) {
      setState(() {
        _isValidatingUsername = false;
        _usernameValidationError =
            'Username must contain only letters and numbers';
        _isUsernameValid = false;
      });
      return;
    }

    setState(() {
      _isValidatingUsername = true;
      _usernameValidationError = null;
    });

    try {
      final authProvider = context.read<AuthProvider>();
      final isValid = await authProvider.validateUsername(username);
      if (!mounted || revision != _usernameRevision) return;

      setState(() {
        _isValidatingUsername = false;
        _isUsernameValid = isValid;
        if (!isValid) {
          _usernameValidationError = 'Username already taken';
        }
      });
    } catch (e) {
      if (!mounted || revision != _usernameRevision) return;
      setState(() {
        _isValidatingUsername = false;
        _usernameValidationError = 'Failed to validate username';
        _isUsernameValid = false;
      });
    }
  }

  Future<void> _pickProfilePicture() async {
    if (_pickingProfilePicture) return;
    setState(() => _pickingProfilePicture = true);
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: const ['jpg', 'jpeg', 'png', 'webp'],
        allowMultiple: false,
        withData: true,
      );
      if (!mounted || result == null) return;
      final file = result.files.single;
      final bytes = file.bytes;
      if (bytes == null || bytes.isEmpty) {
        throw const FormatException(
          'This photo could not be read. Try another image.',
        );
      }
      if (bytes.length > 8 * 1024 * 1024) {
        throw const FormatException('Choose a photo smaller than 8 MB.');
      }
      final buffer = await ImmutableBuffer.fromUint8List(bytes);
      try {
        final descriptor = await ImageDescriptor.encoded(buffer);
        try {
          if (descriptor.width * descriptor.height > 20000000) {
            throw const FormatException(
              'Choose a photo smaller than 20 megapixels.',
            );
          }
        } finally {
          descriptor.dispose();
        }
      } finally {
        buffer.dispose();
      }
      if (!mounted) return;
      setState(() {
        _profilePicture = bytes;
        _profilePictureName = file.name;
      });
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              error is FormatException
                  ? error.message
                  : 'We could not open that photo. Choose a JPG, PNG, or WebP image.',
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _pickingProfilePicture = false);
    }
  }

  Future<void> _handleCompleteProfile() async {
    if (!_isUsernameValid || _usernameController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a valid username')),
      );
      return;
    }

    if (_dateOfBirth == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your date of birth')),
      );
      return;
    }

    if (_selectedGender == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your gender')),
      );
      return;
    }

    if (_selectedLatitude == null ||
        _selectedLongitude == null ||
        _selectedLocationAddress == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your location')),
      );
      return;
    }

    final authProvider = context.read<AuthProvider>();
    String? profilePictureUrl;

    // Upload profile picture if selected
    if (_profilePicture != null) {
      try {
        final apiClient = APIClient(ApiConstants.baseUrl);
        final mediaService = widget.mediaService ?? MediaService(apiClient);
        final user = authProvider.user;

        if (user != null && user.id != null) {
          profilePictureUrl = await mediaService.uploadProfilePictureBytes(
            _profilePicture!,
            _profilePictureName!,
            user.id!,
          );
          if (!mounted) return;

          if (profilePictureUrl == null) {
            if (mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('Failed to upload profile picture'),
                ),
              );
            }
            return;
          }
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to upload profile picture: $e')),
          );
        }
        return;
      }
    }

    // Update profile with all fields including new demographic and location data
    try {
      await authProvider.updateProfile(
        username: _usernameController.text.trim(),
        profilePictureUrl: profilePictureUrl,
        dateOfBirth: _dateOfBirth,
        gender: _selectedGender,
        latitude: _selectedLatitude,
        longitude: _selectedLongitude,
        locationAddress: _selectedLocationAddress,
        bio: _bioController.text.trim().isNotEmpty
            ? _bioController.text.trim()
            : null,
      );

      // Navigation will be handled by main.dart based on auth state
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Failed to update profile: $e')));
      }
    }
  }

  void _handleCancel() {
    final authProvider = context.read<AuthProvider>();
    authProvider.cancelVerificationCode();
    // Clear the code input field
    _codeController.clear();
    // Return focus to phone input
    _phoneFocusNode.requestFocus();
  }

  void _handleRetry() {
    final authProvider = context.read<AuthProvider>();
    authProvider.cancelVerificationCode();
    // Clear the code input field
    _codeController.clear();
    // Keep phone number for easy retry
    // Return focus to phone input
    _phoneFocusNode.requestFocus();
  }

  Widget _formSurface(List<Widget> children) => Center(
    child: SingleChildScrollView(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 32),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Container(
          padding: EdgeInsets.all(
            MediaQuery.sizeOf(context).width < 400 ? 20 : 32,
          ),
          decoration: BoxDecoration(
            color: AppTheme.peach,
            border: Border.all(color: AppTheme.ink),
            borderRadius: BorderRadius.circular(4),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'PLANS WITH FRIENDS',
                style: TextStyle(fontSize: 12, letterSpacing: 1.5),
              ),
              const SizedBox(height: 14),
              ...children,
            ],
          ),
        ),
      ),
    ),
  );

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const TravelAngelsWordmark()),
      body: Consumer<AuthProvider>(
        builder: (context, authProvider, child) {
          // Profile setup step (only for registration)
          if (_shouldSetProfile) {
            return _formSurface([
              const Text(
                'Set up your profile',
                style: TextStyle(
                  fontFamily: AppTheme.displayFont,
                  fontSize: 34,
                  height: 1.15,
                  fontWeight: FontWeight.w400,
                ),
                textAlign: TextAlign.start,
              ),
              const SizedBox(height: 32),
              // Username input
              TextField(
                controller: _usernameController,
                focusNode: _usernameFocusNode,
                decoration: InputDecoration(
                  labelText: 'Username',
                  hintText: 'Choose a username',
                  border: const OutlineInputBorder(),
                  suffixIcon: _isValidatingUsername
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: Padding(
                            padding: EdgeInsets.all(12.0),
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        )
                      : _isUsernameValid
                      ? const Icon(Icons.check, color: AppTheme.success)
                      : null,
                  errorText: _usernameValidationError,
                ),
                onChanged: (value) {
                  _validateUsername(value);
                },
                textInputAction: TextInputAction.done,
              ),
              if (_usernameValidationError != null) ...[
                const SizedBox(height: 8),
                Text(
                  _usernameValidationError!,
                  style: TextStyle(color: AppTheme.error, fontSize: 12),
                ),
              ],
              const SizedBox(height: 24),
              // Date of birth picker
              const Text(
                'Date of Birth *',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 8),
              InkWell(
                onTap: () async {
                  final DateTime? picked = await showDatePicker(
                    context: context,
                    initialDate:
                        _dateOfBirth ??
                        DateTime.now().subtract(const Duration(days: 365 * 18)),
                    firstDate: DateTime(1900),
                    lastDate: DateTime.now(),
                  );
                  if (mounted && picked != null) {
                    setState(() {
                      _dateOfBirth = picked;
                    });
                  }
                },
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    border: Border.all(color: AppTheme.muted),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        _dateOfBirth != null
                            ? '${_dateOfBirth!.year}-${_dateOfBirth!.month.toString().padLeft(2, '0')}-${_dateOfBirth!.day.toString().padLeft(2, '0')}'
                            : 'Select date of birth',
                        style: TextStyle(
                          color: _dateOfBirth != null
                              ? AppTheme.ink
                              : AppTheme.muted,
                        ),
                      ),
                      const Icon(Icons.calendar_today),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 24),
              // Gender dropdown
              const Text(
                'Gender *',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 8),
              DropdownButtonFormField<String>(
                initialValue: _selectedGender,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: 'Select gender',
                ),
                items: const [
                  DropdownMenuItem(value: 'Male', child: Text('Male')),
                  DropdownMenuItem(value: 'Female', child: Text('Female')),
                  DropdownMenuItem(
                    value: 'Non-binary',
                    child: Text('Non-binary'),
                  ),
                  DropdownMenuItem(
                    value: 'Prefer not to say',
                    child: Text('Prefer not to say'),
                  ),
                ],
                onChanged: (value) {
                  setState(() {
                    _selectedGender = value;
                  });
                },
              ),
              const SizedBox(height: 24),
              // Location picker
              const Text(
                'Location *',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 8),
              LocationPicker(
                locationService: widget.locationService,
                showMap: widget.showLocationMap,
                onLocationCleared: () {
                  setState(() {
                    _selectedLatitude = null;
                    _selectedLongitude = null;
                    _selectedLocationAddress = null;
                  });
                },
                onLocationSelected: (latitude, longitude, address) {
                  setState(() {
                    _selectedLatitude = latitude;
                    _selectedLongitude = longitude;
                    _selectedLocationAddress = address;
                  });
                },
              ),
              const SizedBox(height: 24),
              // Bio field
              TextField(
                controller: _bioController,
                decoration: const InputDecoration(
                  labelText: 'Bio (Optional)',
                  hintText: 'Tell us about yourself',
                  border: OutlineInputBorder(),
                ),
                maxLines: 3,
              ),
              const SizedBox(height: 24),
              // Profile picture picker
              const Text(
                'Profile Picture (Optional)',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  if (_profilePicture != null) ...[
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Image.memory(
                        _profilePicture!,
                        width: 100,
                        height: 100,
                        fit: BoxFit.cover,
                      ),
                    ),
                    const SizedBox(width: 16),
                  ],
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _pickingProfilePicture
                          ? null
                          : _pickProfilePicture,
                      icon: const Icon(Icons.image),
                      label: Text(
                        _profilePicture == null ? 'Pick Image' : 'Change Image',
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 32),
              // Complete profile button
              ElevatedButton(
                onPressed:
                    (_isUsernameValid &&
                        _dateOfBirth != null &&
                        _selectedGender != null &&
                        _selectedLatitude != null &&
                        _selectedLongitude != null)
                    ? _handleCompleteProfile
                    : null,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: const Text('Complete Profile'),
              ),
            ]);
          }

          // Phone number and verification code step
          return _formSurface([
            const Text(
              'Sign in or sign up',
              style: TextStyle(
                fontFamily: AppTheme.displayFont,
                fontSize: 34,
                height: 1.15,
                fontWeight: FontWeight.w400,
              ),
              textAlign: TextAlign.start,
            ),
            const SizedBox(height: 32),
            // Phone number input
            InternationalPhoneField(
              label: 'Phone Number',
              focusNode: _phoneFocusNode,
              enabled:
                  !_isRequestingCode &&
                  !_isSigningIn &&
                  !authProvider.isWaitingForVerificationCode,
              onChanged: (value) => setState(() => _phoneNumber = value),
            ),
            if (authProvider.isWaitingForVerificationCode) ...[
              const SizedBox(height: 16),
              Text(
                authProvider.isTestVerification
                    ? 'Test account: enter your configured 6-digit test code. No text message will be sent.'
                    : "We've just sent a 6-digit verification code. It may take a moment to arrive.",
                style: const TextStyle(fontSize: 14, color: AppTheme.muted),
              ),
              const SizedBox(height: 16),
              // Verification code input
              TextField(
                controller: _codeController,
                focusNode: _codeFocusNode,
                enabled: !_isSigningIn,
                keyboardType: TextInputType.number,
                autofillHints: authProvider.isTestVerification
                    ? null
                    : const [AutofillHints.oneTimeCode],
                inputFormatters: [
                  FilteringTextInputFormatter.digitsOnly,
                  LengthLimitingTextInputFormatter(6),
                ],
                decoration: InputDecoration(
                  labelText: authProvider.isTestVerification
                      ? 'Test code'
                      : 'Verification code',
                  hintText: 'Enter 6-digit code',
                  border: const OutlineInputBorder(),
                ),
                textInputAction: TextInputAction.done,
                onChanged: (_) => setState(() {}),
                onSubmitted: (_) => _handleLogister(),
              ),
            ],
            if (authProvider.error != null) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.errorContainer,
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: AppTheme.error),
                ),
                child: Text(
                  authProvider.error!,
                  style: TextStyle(color: AppTheme.error, fontSize: 14),
                ),
              ),
            ],
            const SizedBox(height: 24),
            // Action buttons
            if (!authProvider.isWaitingForVerificationCode &&
                _phoneNumber != null)
              ElevatedButton(
                onPressed: _isRequestingCode ? null : _handleGetCode,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: Text(
                  _isRequestingCode
                      ? 'Please wait…'
                      : AuthPhoneNumber.isTestAccount(_phoneNumber ?? '')
                      ? 'Continue with test account'
                      : 'Get code',
                ),
              )
            else if (authProvider.isWaitingForVerificationCode)
              _verificationActions(authProvider),
            // Retry button when there's an error and not waiting for code
            if (authProvider.error != null &&
                !authProvider.isWaitingForVerificationCode)
              Padding(
                padding: const EdgeInsets.only(top: 12),
                child: OutlinedButton(
                  onPressed: _isRequestingCode ? null : _handleRetry,
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                  ),
                  child: const Text('Try again'),
                ),
              ),
          ]);
        },
      ),
    );
  }

  Widget _verificationActions(AuthProvider authProvider) => LayoutBuilder(
    builder: (context, constraints) {
      final changePhone = OutlinedButton(
        onPressed: _isSigningIn ? null : _handleCancel,
        child: const Text('Change phone number', textAlign: TextAlign.center),
      );
      final submit = ElevatedButton(
        onPressed: !_isSigningIn && _codeController.text.length == 6
            ? _handleLogister
            : null,
        child: Text(
          _isSigningIn
              ? 'Signing in…'
              : authProvider.isRegister
              ? 'Register'
              : 'Login',
        ),
      );
      if (constraints.maxWidth < 420 ||
          MediaQuery.textScalerOf(context).scale(14) > 18) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [submit, const SizedBox(height: 12), changePhone],
        );
      }
      return Row(
        children: [
          Expanded(child: changePhone),
          const SizedBox(width: 12),
          Expanded(child: submit),
        ],
      );
    },
  );
}
