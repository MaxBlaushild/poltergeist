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
  int _sessionRevision = 0;
  int _verificationRevision = 0;
  int _codeRevision = 0;
  bool _loggingOut = false;
  bool _disposed = false;

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

  /// Refreshes account details without tearing down an authenticated screen.
  /// A null result is an absent/invalid session; transport failures are retryable.
  Future<void> verifyToken() async {
    if (_loggingOut || _disposed) return;
    final session = _sessionRevision;
    final request = ++_verificationRevision;
    bool isCurrent() =>
        !_disposed &&
        session == _sessionRevision &&
        request == _verificationRevision;
    _loading = _user == null;
    _error = null;
    notifyListeners();
    try {
      final user = await _authService.verifyToken();
      if (!isCurrent()) return;
      _user = user;
    } catch (e) {
      if (!isCurrent()) return;
      if (e is DioException &&
          (e.response?.statusCode == 401 || e.response?.statusCode == 403)) {
        _user = null;
      }
      _error = 'Unable to refresh your account. Please try again.';
    } finally {
      if (isCurrent()) {
        _loading = false;
        notifyListeners();
      }
    }
  }

  /// Gets a verification code for the given phone number
  Future<void> getVerificationCode(String phoneNumber) async {
    final session = _sessionRevision;
    final request = ++_codeRevision;
    bool isCurrent() =>
        !_disposed && session == _sessionRevision && request == _codeRevision;
    _error = null;
    _isWaitingForVerificationCode = false;
    notifyListeners();

    try {
      final userExists = await _authService.getVerificationCode(phoneNumber);
      if (!isCurrent()) return;
      _isWaitingForVerificationCode = true;
      _isRegister = !userExists; // If user exists, it's login (not register)
      notifyListeners();
    } catch (e) {
      if (!isCurrent()) return;
      _error = _extractErrorMessage(e);
      _isWaitingForVerificationCode = false;
      notifyListeners();
      rethrow;
    }
  }

  /// Attempts to login or register with phone number and verification code
  Future<void> logister(String phoneNumber, String code) async {
    final session = ++_sessionRevision;
    ++_verificationRevision;
    ++_codeRevision;
    _loggingOut = false;
    _loading = false;
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

      if (_disposed || session != _sessionRevision) return;
      _user = response.user;
      _isWaitingForVerificationCode = false;
      notifyListeners();

      // Refresh user details from server to ensure we have all profile fields
      await verifyToken();
    } catch (e) {
      if (_disposed || session != _sessionRevision) return;
      _error = _extractErrorMessage(e);
      notifyListeners();
      rethrow;
    }
  }

  /// Cancels the verification code flow and resets to phone input
  void cancelVerificationCode() {
    ++_codeRevision;
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
  Future<void> updateProfile({
    String? username,
    String? profilePictureUrl,
    DateTime? dateOfBirth,
    String? gender,
    double? latitude,
    double? longitude,
    String? locationAddress,
    String? bio,
  }) async {
    final session = _sessionRevision;
    final userId = _user?.id;
    _error = null;
    notifyListeners();

    try {
      final updatedUser = await _authService.updateProfile(
        username: username,
        profilePictureUrl: profilePictureUrl,
        dateOfBirth: dateOfBirth,
        gender: gender,
        latitude: latitude,
        longitude: longitude,
        locationAddress: locationAddress,
        bio: bio,
      );
      if (_disposed || session != _sessionRevision || _user?.id != userId) {
        return;
      }
      ++_verificationRevision;
      _loading = false;
      _user = updatedUser;
      notifyListeners();
    } catch (e) {
      if (_disposed || session != _sessionRevision) return;
      _error = _extractErrorMessage(e);
      notifyListeners();
      rethrow;
    }
  }

  /// Clears local identity immediately and invalidates pending account requests.
  Future<void> logout() async {
    final session = ++_sessionRevision;
    ++_verificationRevision;
    ++_codeRevision;
    _loggingOut = true;
    _user = null;
    _loading = false;
    _isWaitingForVerificationCode = false;
    _isRegister = false;
    _error = null;
    if (!_disposed) notifyListeners();
    try {
      await _authService.logout();
    } finally {
      if (session == _sessionRevision) _loggingOut = false;
    }
  }

  @override
  void dispose() {
    _disposed = true;
    ++_sessionRevision;
    ++_verificationRevision;
    ++_codeRevision;
    super.dispose();
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
