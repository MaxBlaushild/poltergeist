import 'package:shared_preferences/shared_preferences.dart';

import '../constants/api_constants.dart';
import '../models/user.dart';
import 'api_client.dart';

class AuthResponse {
  final User user;
  final String token;

  AuthResponse({required this.user, required this.token});
}

enum LogisterResult {
  done,
  needsProfileSetup,
}

class AuthService {
  final ApiClient _apiClient;
  static const String _tokenKey = 'token';

  AuthService(this._apiClient);

  String _formatPhoneNumber(String phoneNumber) {
    final cleaned = phoneNumber.replaceAll(RegExp(r'[^\d+]'), '');
    return cleaned.startsWith('+') ? cleaned : '+$cleaned';
  }

  /// Request verification code. Sets waiting state; actual API sends SMS.
  Future<void> getVerificationCode(String phoneNumber) async {
    await _apiClient.post<dynamic>(
      ApiConstants.verificationCodeEndpoint,
      data: {
        'phoneNumber': _formatPhoneNumber(phoneNumber),
        'appName': ApiConstants.appName,
      },
    );
  }

  /// Try login, then register. Returns (result, user). needsProfileSetup when we registered.
  Future<(LogisterResult, User)> logister(String phoneNumber, String code) async {
    final formatted = _formatPhoneNumber(phoneNumber);
    try {
      final data = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.loginEndpoint,
        data: {'phoneNumber': formatted, 'code': code},
      );
      final user = User.fromJson(data['user'] as Map<String, dynamic>);
      final token = data['token'] as String;
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, token);
      return (LogisterResult.done, user);
    } catch (_) {
      // try register
    }

    final data = await _apiClient.post<Map<String, dynamic>>(
      ApiConstants.registerEndpoint,
      data: {'phoneNumber': formatted, 'code': code},
    );
    final user = User.fromJson(data['user'] as Map<String, dynamic>);
    final token = data['token'] as String;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
    return (LogisterResult.needsProfileSetup, user);
  }

  Future<User> whoami() async {
    final data = await _apiClient.get<Map<String, dynamic>>(ApiConstants.whoamiEndpoint);
    return User.fromJson(data);
  }

  Future<void> updateProfile({
    String? username,
    String? profilePictureUrl,
  }) async {
    final body = <String, dynamic>{};
    if (username != null && username.isNotEmpty) body['username'] = username;
    if (profilePictureUrl != null && profilePictureUrl.isNotEmpty) {
      body['profilePictureUrl'] = profilePictureUrl;
    }
    await _apiClient.post<dynamic>(ApiConstants.profileEndpoint, data: body);
  }

  Future<User?> verifyToken() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString(_tokenKey);
    if (token == null) return null;
    try {
      final userJson = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.verifyTokenEndpoint,
        data: {'token': token},
      );
      return User.fromJson(userJson);
    } catch (_) {
      await prefs.remove(_tokenKey);
      return null;
    }
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }

  Future<String?> getStoredToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }
}
