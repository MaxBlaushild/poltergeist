import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../constants/api_constants.dart';
import '../models/user.dart';
import 'api_client.dart';

class AuthResponse {
  final User user;
  final String token;

  AuthResponse({required this.user, required this.token});
}

enum LogisterResult { done, needsProfileSetup }

class LogisterResponse {
  final LogisterResult result;
  final User? user;

  LogisterResponse({required this.result, this.user});
}

class AuthService {
  final ApiClient _apiClient;
  static const String _tokenKey = 'token';
  static const String _dryRunPhone = '+12025550102';
  static const String _dryRunCode = '123456';
  static const String _dryRunUserId = '00000000-0000-0000-0000-00000000d002';
  User? _dryRunUser;
  String? _pendingRegistrationPhone;
  String? _pendingRegistrationCode;

  AuthService(this._apiClient);

  String _formatPhoneNumber(String phoneNumber) {
    final cleaned = phoneNumber.replaceAll(RegExp(r'[^\d+]'), '');
    return cleaned.startsWith('+') ? cleaned : '+$cleaned';
  }

  bool _isDryRunPhone(String phoneNumber) => phoneNumber == _dryRunPhone;

  bool _isDryRunRegistration(String phoneNumber, String code) =>
      _isDryRunPhone(phoneNumber) && code == _dryRunCode;

  bool get isDryRunRegistrationActive => _dryRunUser != null;
  bool get hasPendingRegistration =>
      _pendingRegistrationPhone != null && _pendingRegistrationCode != null;

  /// Request verification code. Sets waiting state; actual API sends SMS.
  Future<void> getVerificationCode(String phoneNumber) async {
    final formatted = _formatPhoneNumber(phoneNumber);
    if (_isDryRunPhone(formatted)) {
      return;
    }

    await _apiClient.post<dynamic>(
      ApiConstants.verificationCodeEndpoint,
      data: {'phoneNumber': formatted, 'appName': ApiConstants.appName},
    );
  }

  /// Try login first. If the phone number is new, keep a pending registration
  /// locally and only create the user once profile setup is submitted.
  Future<LogisterResponse> logister(
    String phoneNumber,
    String code,
  ) async {
    final formatted = _formatPhoneNumber(phoneNumber);
    if (_isDryRunRegistration(formatted, code)) {
      final user = User(
        id: _dryRunUserId,
        phoneNumber: formatted,
        name: 'Dry Run User',
        username: '',
        profilePictureUrl: '',
      );
      _dryRunUser = user;
      return LogisterResponse(
        result: LogisterResult.needsProfileSetup,
        user: user,
      );
    }

    try {
      final data = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.loginEndpoint,
        data: {'phoneNumber': formatted, 'code': code},
      );
      final user = User.fromJson(data['user'] as Map<String, dynamic>);
      final token = data['token'] as String;
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, token);
      _pendingRegistrationPhone = null;
      _pendingRegistrationCode = null;
      return LogisterResponse(result: LogisterResult.done, user: user);
    } on DioException catch (error) {
      if (error.response?.statusCode != 404) {
        rethrow;
      }
      _pendingRegistrationPhone = formatted;
      _pendingRegistrationCode = code;
      return LogisterResponse(result: LogisterResult.needsProfileSetup);
    }
  }

  Future<User> completePendingRegistration({required String username}) async {
    if (_dryRunUser != null) {
      return _dryRunUser!;
    }
    final phone = _pendingRegistrationPhone;
    final code = _pendingRegistrationCode;
    if (phone == null || code == null) {
      throw StateError('No pending registration is available.');
    }

    final data = await _apiClient.post<Map<String, dynamic>>(
      ApiConstants.registerEndpoint,
      data: {
        'phoneNumber': phone,
        'code': code,
        'username': username,
      },
    );
    final user = User.fromJson(data['user'] as Map<String, dynamic>);
    final token = data['token'] as String;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
    _pendingRegistrationPhone = null;
    _pendingRegistrationCode = null;
    return user;
  }

  Future<User> whoami() async {
    if (_dryRunUser != null) {
      return _dryRunUser!;
    }
    final data = await _apiClient.get<Map<String, dynamic>>(
      ApiConstants.whoamiEndpoint,
    );
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
    if (_dryRunUser != null) return _dryRunUser;

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
    _dryRunUser = null;
    _pendingRegistrationPhone = null;
    _pendingRegistrationCode = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }

  void cancelPendingRegistration() {
    _dryRunUser = null;
    _pendingRegistrationPhone = null;
    _pendingRegistrationCode = null;
  }

  Future<String?> getStoredToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }
}
