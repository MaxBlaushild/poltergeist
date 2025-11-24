import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:file_picker/file_picker.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/constants/api_constants.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _phoneController = TextEditingController();
  final _codeController = TextEditingController();
  final _usernameController = TextEditingController();
  final _phoneFocusNode = FocusNode();
  final _codeFocusNode = FocusNode();
  final _usernameFocusNode = FocusNode();

  bool _isPhoneValid = false;
  bool _shouldSetProfile = false;
  File? _profilePicture;
  bool _isValidatingUsername = false;
  String? _usernameValidationError;
  bool _isUsernameValid = false;

  @override
  void dispose() {
    _phoneController.dispose();
    _codeController.dispose();
    _usernameController.dispose();
    _phoneFocusNode.dispose();
    _codeFocusNode.dispose();
    _usernameFocusNode.dispose();
    super.dispose();
  }

  void _validatePhoneNumber(String value) {
    // Basic validation: non-empty and reasonable length (10-15 digits)
    final digitsOnly = value.replaceAll(RegExp(r'[^\d]'), '');
    setState(() {
      _isPhoneValid = digitsOnly.length >= 10 && digitsOnly.length <= 15;
    });
  }

  Future<void> _handleGetCode() async {
    if (!_isPhoneValid) return;

    final authProvider = context.read<AuthProvider>();
    final phoneNumber = _phoneController.text.trim();

    try {
      await authProvider.getVerificationCode(phoneNumber);
      _codeFocusNode.requestFocus();
    } catch (e) {
      // Error is handled by AuthProvider and displayed via error getter
    }
  }

  Future<void> _handleLogister() async {
    if (_codeController.text.length != 6) return;

    final authProvider = context.read<AuthProvider>();
    final phoneNumber = _phoneController.text.trim();
    final code = _codeController.text.trim();

    try {
      await authProvider.logister(phoneNumber, code);
      
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
    }
  }

  Future<void> _validateUsername(String username) async {
    if (username.isEmpty) {
      setState(() {
        _usernameValidationError = null;
        _isUsernameValid = false;
      });
      return;
    }

    // Basic format validation (alphanumeric only)
    if (!RegExp(r'^[a-zA-Z0-9]+$').hasMatch(username)) {
      setState(() {
        _usernameValidationError = 'Username must contain only letters and numbers';
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
      
      setState(() {
        _isValidatingUsername = false;
        _isUsernameValid = isValid;
        if (!isValid) {
          _usernameValidationError = 'Username already taken';
        }
      });
    } catch (e) {
      setState(() {
        _isValidatingUsername = false;
        _usernameValidationError = 'Failed to validate username';
        _isUsernameValid = false;
      });
    }
  }

  Future<void> _pickProfilePicture() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.image,
        allowMultiple: false,
      );

      if (result != null && result.files.single.path != null) {
        setState(() {
          _profilePicture = File(result.files.single.path!);
        });
      }
    } catch (e) {
      // Handle error
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to pick image: $e')),
        );
      }
    }
  }

  Future<void> _handleCompleteProfile() async {
    if (!_isUsernameValid || _usernameController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a valid username')),
      );
      return;
    }

    final authProvider = context.read<AuthProvider>();
    String? profilePictureUrl;

    // Upload profile picture if selected
    if (_profilePicture != null) {
      try {
        final apiClient = APIClient(ApiConstants.baseUrl);
        final mediaService = MediaService(apiClient);
        final user = authProvider.user;
        
        if (user != null && user.id != null) {
          profilePictureUrl = await mediaService.uploadProfilePicture(
            _profilePicture!,
            user.id!,
          );
          
          if (profilePictureUrl == null) {
            if (mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Failed to upload profile picture')),
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

    // Update profile with username and profile picture
    try {
      await authProvider.updateProfile(
        username: _usernameController.text.trim(),
        profilePictureUrl: profilePictureUrl,
      );
      
      // Navigation will be handled by main.dart based on auth state
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to update profile: $e')),
        );
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Travel Angels'),
      ),
      body: Consumer<AuthProvider>(
        builder: (context, authProvider, child) {
          // Profile setup step (only for registration)
          if (_shouldSetProfile) {
            return Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const Text(
                      'Set up your profile',
                      style: TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                      ),
                      textAlign: TextAlign.center,
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
                                ? const Icon(Icons.check, color: Colors.green)
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
                        style: TextStyle(
                          color: Colors.red.shade700,
                          fontSize: 12,
                        ),
                      ),
                    ],
                    const SizedBox(height: 24),
                    // Profile picture picker
                    const Text(
                      'Profile Picture (Optional)',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        if (_profilePicture != null) ...[
                          ClipRRect(
                            borderRadius: BorderRadius.circular(8),
                            child: Image.file(
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
                            onPressed: _pickProfilePicture,
                            icon: const Icon(Icons.image),
                            label: Text(_profilePicture == null ? 'Pick Image' : 'Change Image'),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 32),
                    // Complete profile button
                    ElevatedButton(
                      onPressed: _isUsernameValid ? _handleCompleteProfile : null,
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Complete Profile'),
                    ),
                  ],
                ),
              ),
            );
          }

          // Phone number and verification code step
          return Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const Text(
                    'Sign in or sign up',
                    style: TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 32),
                  // Phone number input
                  TextField(
                    controller: _phoneController,
                    focusNode: _phoneFocusNode,
                    keyboardType: TextInputType.phone,
                    decoration: const InputDecoration(
                      labelText: 'Phone Number',
                      hintText: 'Enter your phone number',
                      border: OutlineInputBorder(),
                    ),
                    onChanged: _validatePhoneNumber,
                  ),
                  if (authProvider.isWaitingForVerificationCode) ...[
                    const SizedBox(height: 16),
                    const Text(
                      "We've just sent a 6-digit verification code. It may take a moment to arrive.",
                      style: TextStyle(
                        fontSize: 14,
                        color: Colors.grey,
                      ),
                    ),
                    const SizedBox(height: 16),
                    // Verification code input
                    TextField(
                      controller: _codeController,
                      focusNode: _codeFocusNode,
                      keyboardType: TextInputType.number,
                      inputFormatters: [
                        FilteringTextInputFormatter.digitsOnly,
                        LengthLimitingTextInputFormatter(6),
                      ],
                      decoration: const InputDecoration(
                        labelText: 'Verification code',
                        hintText: 'Enter 6-digit code',
                        border: OutlineInputBorder(),
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
                        color: Colors.red.shade50,
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(color: Colors.red.shade200),
                      ),
                      child: Text(
                        authProvider.error!,
                        style: TextStyle(
                          color: Colors.red.shade700,
                          fontSize: 14,
                        ),
                      ),
                    ),
                  ],
                  const SizedBox(height: 24),
                  // Action buttons
                  if (!authProvider.isWaitingForVerificationCode && _isPhoneValid)
                    ElevatedButton(
                      onPressed: _handleGetCode,
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Get code'),
                    )
                  else if (authProvider.isWaitingForVerificationCode) ...[
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton(
                            onPressed: _handleCancel,
                            style: OutlinedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16),
                            ),
                            child: const Text('Change phone number'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          flex: 2,
                          child: ElevatedButton(
                            onPressed: _codeController.text.length == 6
                                ? _handleLogister
                                : null,
                            style: ElevatedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16),
                            ),
                            child: Text(
                              authProvider.isRegister ? 'Register' : 'Login',
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                  // Retry button when there's an error and not waiting for code
                  if (authProvider.error != null && 
                      !authProvider.isWaitingForVerificationCode)
                    Padding(
                      padding: const EdgeInsets.only(top: 12),
                      child: OutlinedButton(
                        onPressed: _handleRetry,
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 16),
                        ),
                        child: const Text('Try again'),
                      ),
                    ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}


