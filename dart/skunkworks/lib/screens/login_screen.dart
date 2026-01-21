import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/providers/auth_provider.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _phoneController = TextEditingController();
  final _codeController = TextEditingController();
  final _phoneFocusNode = FocusNode();
  final _codeFocusNode = FocusNode();

  bool _isPhoneValid = false;

  @override
  void dispose() {
    _phoneController.dispose();
    _codeController.dispose();
    _phoneFocusNode.dispose();
    _codeFocusNode.dispose();
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
      // Navigation will be handled by main.dart based on auth state
    } catch (e) {
      // Error is handled by AuthProvider and displayed via error getter
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 40.0),
            child: Consumer<AuthProvider>(
              builder: (context, authProvider, child) {
                return Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // App branding area (Instagram-like)
                    const SizedBox(height: 60),
                    Text(
                      'Verifiable SN',
                      style: TextStyle(
                        fontSize: 48,
                        fontWeight: FontWeight.w400,
                        letterSpacing: -1.0,
                        color: Colors.black87,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 40),

                    // Phone number or verification code input
                    AnimatedSwitcher(
                      duration: const Duration(milliseconds: 300),
                      child: !authProvider.isWaitingForVerificationCode
                          ? _buildPhoneInput(authProvider)
                          : _buildCodeInput(authProvider),
                    ),

                    // Error message
                    if (authProvider.error != null) ...[
                      const SizedBox(height: 20),
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
                          textAlign: TextAlign.center,
                        ),
                      ),
                    ],

                    const SizedBox(height: 20),

                    // Action buttons
                    if (!authProvider.isWaitingForVerificationCode) ...[
                      _buildGetCodeButton(authProvider),
                    ] else ...[
                      _buildCodeButtons(authProvider),
                    ],
                  ],
                );
              },
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildPhoneInput(AuthProvider authProvider) {
    return Column(
      key: const ValueKey('phone'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _phoneController,
          focusNode: _phoneFocusNode,
          keyboardType: TextInputType.phone,
          decoration: InputDecoration(
            labelText: 'Phone Number',
            hintText: 'Phone number',
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 16,
            ),
            filled: true,
            fillColor: Colors.grey.shade50,
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
              borderSide: BorderSide(color: Colors.grey.shade600),
            ),
          ),
          style: const TextStyle(fontSize: 16),
          onChanged: _validatePhoneNumber,
          textInputAction: TextInputAction.done,
          onSubmitted: (_) => _handleGetCode(),
        ),
      ],
    );
  }

  Widget _buildCodeInput(AuthProvider authProvider) {
    return Column(
      key: const ValueKey('code'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          "We've sent a 6-digit verification code to your phone. It may take a moment to arrive.",
          style: TextStyle(
            fontSize: 14,
            color: Colors.grey.shade600,
          ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 24),
        TextField(
          controller: _codeController,
          focusNode: _codeFocusNode,
          keyboardType: TextInputType.number,
          inputFormatters: [
            FilteringTextInputFormatter.digitsOnly,
            LengthLimitingTextInputFormatter(6),
          ],
          decoration: InputDecoration(
            labelText: 'Verification Code',
            hintText: 'Enter 6-digit code',
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 16,
            ),
            filled: true,
            fillColor: Colors.grey.shade50,
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
              borderSide: BorderSide(color: Colors.grey.shade600),
            ),
          ),
          style: const TextStyle(
            fontSize: 16,
            letterSpacing: 4,
          ),
          textAlign: TextAlign.center,
          textInputAction: TextInputAction.done,
          onChanged: (_) => setState(() {}),
          onSubmitted: (_) => _handleLogister(),
        ),
      ],
    );
  }

  Widget _buildGetCodeButton(AuthProvider authProvider) {
    return SizedBox(
      height: 44,
      child: ElevatedButton(
        onPressed: _isPhoneValid ? _handleGetCode : null,
        style: ElevatedButton.styleFrom(
          backgroundColor: _isPhoneValid
              ? Colors.blue.shade600
              : Colors.blue.shade300,
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(4),
          ),
          elevation: 0,
        ),
        child: const Text(
          'Next',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }

  Widget _buildCodeButtons(AuthProvider authProvider) {
    return Row(
      children: [
        Expanded(
          child: TextButton(
            onPressed: _handleCancel,
            style: TextButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 12),
            ),
            child: Text(
              'Change number',
              style: TextStyle(
                color: Colors.grey.shade700,
                fontSize: 14,
              ),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          flex: 2,
          child: SizedBox(
            height: 44,
            child: ElevatedButton(
              onPressed: _codeController.text.length == 6
                  ? _handleLogister
                  : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: _codeController.text.length == 6
                    ? Colors.blue.shade600
                    : Colors.blue.shade300,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(4),
                ),
                elevation: 0,
              ),
              child: Text(
                authProvider.isRegister ? 'Register' : 'Log In',
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}
