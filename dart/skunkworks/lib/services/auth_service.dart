import 'package:shared_preferences/shared_preferences.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:dio/dio.dart';

class AuthService {
  final APIClient _apiClient;
  static const String _tokenKey = 'token';
  static const String _demoPhone = '+14407858475';

  AuthService(this._apiClient);

  /// Formats phone number to ensure it starts with +
  String _formatPhoneNumber(String phoneNumber) {
    // Remove all non-digit characters except +
    String cleaned = phoneNumber.replaceAll(RegExp(r'[^\d+]'), '');
    
    // If it doesn't start with +, add it
    if (!cleaned.startsWith('+')) {
      cleaned = '+$cleaned';
    }
    
    return cleaned;
  }

  bool _isDemoPhone(String formattedPhone) => formattedPhone == _demoPhone;

  /// Gets a verification code for the given phone number
  /// Returns true if user exists (login), false if new user (register)
  Future<bool> getVerificationCode(String phoneNumber) async {
    try {
      final formattedPhone = _formatPhoneNumber(phoneNumber);

      // Demo account: skip API call (no SMS), always treat as existing user (login).
      if (_isDemoPhone(formattedPhone)) {
        return true;
      }

      final response = await _apiClient.post(
        ApiConstants.verificationCodeEndpoint,
        data: {
          'phoneNumber': formattedPhone,
          'appName': ApiConstants.appName,
        },
      );

      // If response contains a user object (Map with id field), user exists (login)
      // If response is null or doesn't have id, user doesn't exist (register)
      if (response is Map<String, dynamic> && response.containsKey('id')) {
        return true;
      }
      return false;
    } catch (e) {
      // If we get an error, the user might not exist
      // Re-throw to let the caller handle it
      rethrow;
    }
  }

  /// Attempts to login with phone number and verification code
  /// Returns the authenticated user and token
  Future<AuthResponse> login(String phoneNumber, String code) async {
    try {
      final formattedPhone = _formatPhoneNumber(phoneNumber);
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.loginEndpoint,
        data: {
          'phoneNumber': formattedPhone,
          'code': code,
        },
      );

      final user = User.fromJson(response['user'] as Map<String, dynamic>);
      final token = response['token'] as String;

      // Store token
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, token);

      return AuthResponse(user: user, token: token);
    } catch (e) {
      rethrow;
    }
  }

  /// Registers a new user with phone number and verification code
  /// Returns the authenticated user and token
  Future<AuthResponse> register(
    String phoneNumber,
    String code, {
    String? username,
    String? profilePictureUrl,
  }) async {
    try {
      final formattedPhone = _formatPhoneNumber(phoneNumber);
      final data = <String, dynamic>{
        'phoneNumber': formattedPhone,
        'code': code,
      };

      if (username != null && username.isNotEmpty) {
        data['username'] = username;
      }

      if (profilePictureUrl != null && profilePictureUrl.isNotEmpty) {
        data['profilePictureUrl'] = profilePictureUrl;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.registerEndpoint,
        data: data,
      );

      final user = User.fromJson(response['user'] as Map<String, dynamic>);
      final token = response['token'] as String;

      // Store token
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, token);

      return AuthResponse(user: user, token: token);
    } catch (e) {
      rethrow;
    }
  }

  /// Verifies the stored token and returns the user if valid
  Future<User?> verifyToken() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final token = prefs.getString(_tokenKey);
      if (token == null) {
        return null;
      }

      // Use verifyToken endpoint
      final dio = Dio(BaseOptions(baseUrl: ApiConstants.baseUrl));
      final response = await dio.post<Map<String, dynamic>>(
        ApiConstants.verifyTokenEndpoint,
        data: {
          'token': token,
        },
      );

      return User.fromJson(response.data!);
    } catch (e) {
      // Token is invalid, clear it
      await logout();
      return null;
    }
  }

  /// Logs out the user by clearing the stored token
  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }

  /// Gets the stored token
  Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }

  /// Updates the user profile
  /// 
  /// [username] - Optional username to update
  /// [profilePictureUrl] - Optional profile picture URL to update
  /// [category] - Optional category to update
  /// [ageRange] - Optional age range to update
  /// [bio] - Optional bio to update
  /// 
  /// Returns the updated user
  Future<User> updateProfile({
    String? username,
    String? profilePictureUrl,
    String? category,
    String? ageRange,
    String? bio,
  }) async {
    try {
      final data = <String, dynamic>{};
      if (username != null) data['username'] = username;
      if (profilePictureUrl != null) data['profilePictureUrl'] = profilePictureUrl;
      if (category != null) data['category'] = category;
      if (ageRange != null) data['ageRange'] = ageRange;
      if (bio != null) data['bio'] = bio;

      final response = await _apiClient.put<Map<String, dynamic>>(
        ApiConstants.updateProfileEndpoint,
        data: data,
      );

      return User.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }
}

class AuthResponse {
  final User user;
  final String token;

  AuthResponse({required this.user, required this.token});
}
