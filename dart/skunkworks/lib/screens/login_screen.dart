import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:intl_phone_field/intl_phone_field.dart';
import 'package:intl_phone_field/phone_number.dart';

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
  String _countryCode = '+1'; // Default to US
  String _phoneNumber = '';
  PhoneNumber? _phoneNumberObj;

  @override
  void dispose() {
    _phoneController.dispose();
    _codeController.dispose();
    _phoneFocusNode.dispose();
    _codeFocusNode.dispose();
    super.dispose();
  }

  void _validatePhoneNumber(PhoneNumber? phoneNumber) {
    if (phoneNumber == null || phoneNumber.number.isEmpty) {
      setState(() {
        _isPhoneValid = false;
        _phoneNumber = '';
        _phoneNumberObj = null;
      });
      return;
    }
    
    // Store the phone number object and extracted values
    _phoneNumberObj = phoneNumber;
    _phoneNumber = phoneNumber.number;
    _countryCode = phoneNumber.countryCode;
    
    // Basic validation: reasonable length (at least 4 digits, max 15)
    final digitsOnly = phoneNumber.number.replaceAll(RegExp(r'[^\d]'), '');
    setState(() {
      _isPhoneValid = digitsOnly.length >= 4 && digitsOnly.length <= 15;
    });
  }

  String _getFullPhoneNumber() {
    // Use the complete phone number from the PhoneNumber object if available
    if (_phoneNumberObj != null) {
      // Ensure the number starts with '+'
      final completeNumber = _phoneNumberObj!.completeNumber;
      if (completeNumber.startsWith('+')) {
        return completeNumber;
      } else {
        // If it doesn't start with '+', add it
        return '+$completeNumber';
      }
    }
    // Fallback: combine country code and phone number with '+' prefix
    final digitsOnly = _phoneNumber.replaceAll(RegExp(r'[^\d]'), '');
    // Ensure country code has '+' prefix
    final countryCode = _countryCode.startsWith('+') ? _countryCode : '+$_countryCode';
    return '$countryCode$digitsOnly';
  }

  Future<void> _handleGetCode() async {
    if (!_isPhoneValid) return;

    final authProvider = context.read<AuthProvider>();
    // Get full phone number with country code and '+' prefix
    final phoneNumber = _getFullPhoneNumber();

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
    // Get full phone number with country code and '+' prefix
    final phoneNumber = _getFullPhoneNumber();
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
      backgroundColor: AppColors.warmWhite,
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
                      'Vera',
                      style: TextStyle(
                        fontSize: 48,
                        fontWeight: FontWeight.w400,
                        letterSpacing: -1.0,
                        color: AppColors.graphiteInk,
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
                          color: AppColors.coralPop.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(4),
                          border: Border.all(color: AppColors.coralPop.withOpacity(0.3)),
                        ),
                        child: Text(
                          authProvider.error!,
                          style: TextStyle(
                            color: AppColors.coralPop,
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
        IntlPhoneField(
          controller: _phoneController,
          focusNode: _phoneFocusNode,
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
          initialCountryCode: 'US',
          onCountryChanged: (country) {
            setState(() {
              _countryCode = '+${country.dialCode}';
            });
          },
          onChanged: (phone) {
            _validatePhoneNumber(phone);
          },
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
            color: AppColors.graphiteInk.withOpacity(0.6),
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
              ? AppColors.softRealBlue
              : AppColors.softRealBlue.withOpacity(0.5),
          foregroundColor: AppColors.warmWhite,
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
                  color: AppColors.graphiteInk.withOpacity(0.7),
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
                    ? AppColors.softRealBlue
                    : AppColors.softRealBlue.withOpacity(0.5),
                foregroundColor: AppColors.warmWhite,
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
