import 'package:flutter/foundation.dart';

import '../models/user.dart';
import '../services/auth_service.dart' show AuthService, LogisterResult;

class AuthProvider with ChangeNotifier {
  final AuthService _auth;
  User? _user;
  bool _loading = true;
  String? _error;
  bool _isWaitingForVerificationCode = false;
  bool _isRegistrationFlowActive = false;

  AuthProvider(this._auth) {
    _init();
  }

  User? get user => _user;
  bool get loading => _loading;
  String? get error => _error;
  bool get isWaitingForVerificationCode => _isWaitingForVerificationCode;
  bool get isAuthenticated => _user != null && !_isRegistrationFlowActive;
  bool get isDryRunRegistrationActive => _auth.isDryRunRegistrationActive;
  bool get isRegistrationFlowActive => _isRegistrationFlowActive;

  Future<void> _init() async {
    _user = await _auth.verifyToken();
    _loading = false;
    notifyListeners();
    if (_user == null) return;
    try {
      _user = await _auth.whoami();
      notifyListeners();
    } catch (_) {
      // Keep the verified user if refresh fails (e.g. transient network error).
    }
  }

  Future<void> getVerificationCode(String phoneNumber) async {
    _error = null;
    notifyListeners();
    try {
      await _auth.getVerificationCode(phoneNumber);
      _isWaitingForVerificationCode = true;
    } catch (e) {
      _error = e.toString();
    }
    notifyListeners();
  }

  /// Returns true if profile setup is needed (new user).
  Future<bool> logister(String phoneNumber, String code) async {
    _error = null;
    notifyListeners();
    try {
      final response = await _auth.logister(phoneNumber, code);
      _user = response.user;
      _isWaitingForVerificationCode = false;
      _isRegistrationFlowActive =
          response.result == LogisterResult.needsProfileSetup;
      notifyListeners();
      if (response.result == LogisterResult.done) {
        try {
          _user = await _auth.whoami();
          notifyListeners();
        } catch (_) {
          // If the refresh fails, keep the login response data.
        }
      }
      return response.result == LogisterResult.needsProfileSetup;
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  Future<void> completeRegistration({required String username}) async {
    _error = null;
    notifyListeners();
    try {
      _user = await _auth.completePendingRegistration(username: username);
      _isRegistrationFlowActive = true;
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  Future<void> updateProfile({
    String? username,
    String? profilePictureUrl,
  }) async {
    await _auth.updateProfile(
      username: username,
      profilePictureUrl: profilePictureUrl,
    );
    _user = await _auth.whoami();
    notifyListeners();
  }

  /// Refetch current user (e.g. after purchase/sell updates gold).
  Future<void> refresh() async {
    _user = await _auth.whoami();
    notifyListeners();
  }

  void completeRegistrationFlow() {
    if (!_isRegistrationFlowActive) return;
    _isRegistrationFlowActive = false;
    notifyListeners();
  }

  void setUser(User? u) {
    _user = u;
    notifyListeners();
  }

  Future<void> logout() async {
    await _auth.logout();
    _user = null;
    _error = null;
    _isWaitingForVerificationCode = false;
    _isRegistrationFlowActive = false;
    notifyListeners();
  }

  void cancelRegistrationFlow() {
    _auth.cancelPendingRegistration();
    _user = null;
    _error = null;
    _isWaitingForVerificationCode = false;
    _isRegistrationFlowActive = false;
    notifyListeners();
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
