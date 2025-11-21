import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';

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


