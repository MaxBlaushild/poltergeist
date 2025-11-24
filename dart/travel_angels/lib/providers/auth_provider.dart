import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/auth_service.dart';

class AuthProvider extends ChangeNotifier {
  final AuthService _authService;
  User? _user;
  bool _loading = true;
  bool _isWaitingForVerificationCode = false;
  bool _isRegister = false;
  String? _error;

  AuthProvider(this._authService) {
    _initialize();
  }

  User? get user => _user;
  bool get loading => _loading;
  bool get isWaitingForVerificationCode => _isWaitingForVerificationCode;
  bool get isRegister => _isRegister;
  String? get error => _error;

  bool get isAuthenticated => _user != null;

  Future<void> _initialize() async {
    await verifyToken();
  }

  /// Verifies the stored token and loads user if valid
  Future<void> verifyToken() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      final user = await _authService.verifyToken();
      _user = user;
    } catch (e) {
      _error = 'Failed to verify token';
      _user = null;
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  /// Gets a verification code for the given phone number
  Future<void> getVerificationCode(String phoneNumber) async {
    _error = null;
    _isWaitingForVerificationCode = false;
    notifyListeners();

    try {
      final userExists = await _authService.getVerificationCode(phoneNumber);
      _isWaitingForVerificationCode = true;
      _isRegister = !userExists; // If user exists, it's login (not register)
      notifyListeners();
    } catch (e) {
      _error = _extractErrorMessage(e);
      _isWaitingForVerificationCode = false;
      notifyListeners();
      rethrow;
    }
  }

  /// Attempts to login or register with phone number and verification code
  Future<void> logister(String phoneNumber, String code) async {
    _error = null;
    notifyListeners();

    try {
      AuthResponse response;
      if (_isRegister) {
        // User doesn't exist, register them
        response = await _authService.register(phoneNumber, code);
      } else {
        // User exists, login them
        response = await _authService.login(phoneNumber, code);
      }

      _user = response.user;
      _isWaitingForVerificationCode = false;
      notifyListeners();
    } catch (e) {
      _error = _extractErrorMessage(e);
      notifyListeners();
      rethrow;
    }
  }

  /// Cancels the verification code flow and resets to phone input
  void cancelVerificationCode() {
    _isWaitingForVerificationCode = false;
    _isRegister = false;
    _error = null;
    notifyListeners();
  }

  /// Validates username uniqueness
  Future<bool> validateUsername(String username) async {
    try {
      return await _authService.validateUsername(username);
    } catch (e) {
      return false;
    }
  }

  /// Updates user profile (username and/or profile picture)
  Future<void> updateProfile({String? username, String? profilePictureUrl}) async {
    _error = null;
    notifyListeners();

    try {
      final updatedUser = await _authService.updateProfile(
        username: username,
        profilePictureUrl: profilePictureUrl,
      );
      _user = updatedUser;
      notifyListeners();
    } catch (e) {
      _error = _extractErrorMessage(e);
      notifyListeners();
      rethrow;
    }
  }

  /// Logs out the user
  Future<void> logout() async {
    await _authService.logout();
    _user = null;
    _isWaitingForVerificationCode = false;
    _isRegister = false;
    _error = null;
    notifyListeners();
  }

  /// Extracts error message from DioException response body
  String _extractErrorMessage(dynamic error) {
    if (error is DioException) {
      // Try to get error message from response body
      if (error.response?.data != null) {
        final data = error.response!.data;
        if (data is Map<String, dynamic> && data.containsKey('error')) {
          return data['error'] as String;
        }
        if (data is String) {
          return data;
        }
      }
      // Fall back to DioException message
      return error.message ?? 'An error occurred';
    }
    // Fall back for non-Dio errors
    return error.toString();
  }
}

