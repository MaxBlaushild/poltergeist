import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:dio/dio.dart';

class AuthService {
  final APIClient _apiClient;
  static const String _tokenKey = 'token';
  int _sessionRevision = 0;

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

  /// Gets a verification code for the given phone number
  /// Returns true if user exists (login), false if new user (register)
  Future<bool> getVerificationCode(String phoneNumber) async {
    try {
      final formattedPhone = _formatPhoneNumber(phoneNumber);
      final response = await _apiClient.post(
        ApiConstants.verificationCodeEndpoint,
        data: {'phoneNumber': formattedPhone, 'appName': ApiConstants.appName},
      );

      return response != null;
    } catch (e) {
      rethrow;
    }
  }

  /// Attempts to login with phone number and verification code
  /// Returns the authenticated user and token
  Future<AuthResponse> login(String phoneNumber, String code) async {
    final session = ++_sessionRevision;
    try {
      final formattedPhone = _formatPhoneNumber(phoneNumber);
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.loginEndpoint,
        data: {'phoneNumber': formattedPhone, 'code': code},
      );

      final user = User.fromJson(response['user'] as Map<String, dynamic>);
      final token = response['token'] as String;

      // Store token
      final prefs = await SharedPreferences.getInstance();
      if (session != _sessionRevision) {
        throw StateError('Sign-in was cancelled');
      }
      await prefs.setString(_tokenKey, token);

      return AuthResponse(user: user, token: token);
    } catch (e) {
      rethrow;
    }
  }

  /// Validates username uniqueness
  /// Returns true if username is valid and available
  Future<bool> validateUsername(String username) async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.validateUsernameEndpoint(username),
      );
      return response['valid'] as bool? ?? false;
    } catch (e) {
      // If validation fails, assume invalid
      return false;
    }
  }

  /// Registers a new user with phone number and verification code
  /// Returns the authenticated user and token
  Future<AuthResponse> register(
    String phoneNumber,
    String code, {
    String? username,
    String? profilePictureUrl,
    DateTime? dateOfBirth,
    String? gender,
    double? latitude,
    double? longitude,
    String? locationAddress,
    String? bio,
  }) async {
    final session = ++_sessionRevision;
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

      if (dateOfBirth != null) {
        data['dateOfBirth'] =
            '${dateOfBirth.year}-${dateOfBirth.month.toString().padLeft(2, '0')}-${dateOfBirth.day.toString().padLeft(2, '0')}';
      }

      if (gender != null && gender.isNotEmpty) {
        data['gender'] = gender;
      }

      if (latitude != null) {
        data['latitude'] = latitude;
      }

      if (longitude != null) {
        data['longitude'] = longitude;
      }

      if (locationAddress != null && locationAddress.isNotEmpty) {
        data['locationAddress'] = locationAddress;
      }

      if (bio != null && bio.isNotEmpty) {
        data['bio'] = bio;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.registerEndpoint,
        data: data,
      );

      final user = User.fromJson(response['user'] as Map<String, dynamic>);
      final token = response['token'] as String;

      // Store token
      final prefs = await SharedPreferences.getInstance();
      if (session != _sessionRevision) {
        throw StateError('Sign-in was cancelled');
      }
      await prefs.setString(_tokenKey, token);

      return AuthResponse(user: user, token: token);
    } catch (e) {
      rethrow;
    }
  }

  /// Updates user profile (username, profile picture, and demographic/location fields)
  Future<User> updateProfile({
    String? username,
    String? profilePictureUrl,
    DateTime? dateOfBirth,
    String? gender,
    double? latitude,
    double? longitude,
    String? locationAddress,
    String? bio,
  }) async {
    try {
      final data = <String, dynamic>{};

      if (username != null && username.isNotEmpty) {
        data['username'] = username;
      }

      if (profilePictureUrl != null && profilePictureUrl.isNotEmpty) {
        data['profilePictureUrl'] = profilePictureUrl;
      }

      if (dateOfBirth != null) {
        data['dateOfBirth'] =
            '${dateOfBirth.year}-${dateOfBirth.month.toString().padLeft(2, '0')}-${dateOfBirth.day.toString().padLeft(2, '0')}';
      }

      if (gender != null && gender.isNotEmpty) {
        data['gender'] = gender;
      }

      if (latitude != null) {
        data['latitude'] = latitude;
      }

      if (longitude != null) {
        data['longitude'] = longitude;
      }

      if (locationAddress != null && locationAddress.isNotEmpty) {
        data['locationAddress'] = locationAddress;
      }

      if (bio != null) {
        data['bio'] = bio;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.profileEndpoint,
        data: data,
      );

      return User.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }

  /// Returns null only when no token exists or the server rejects the session.
  /// Temporary network/server failures leave the stored token available to retry.
  Future<User?> verifyToken() async {
    final session = _sessionRevision;
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString(_tokenKey);
    if (token == null) return null;
    try {
      Map<String, dynamic> data;
      try {
        data = await _apiClient.get<Map<String, dynamic>>(
          ApiConstants.whoamiEndpoint,
        );
      } on DioException catch (error) {
        // Keep compatibility with older deployments lacking whoami. A timeout,
        // rejected session or server outage is not a reason to change endpoints.
        if (error.response?.statusCode != 404 &&
            error.response?.statusCode != 405) {
          rethrow;
        }
        if (session != _sessionRevision ||
            prefs.getString(_tokenKey) != token) {
          throw StateError('Session changed during refresh');
        }
        data = await _apiClient.postForSession<Map<String, dynamic>>(
          ApiConstants.verifyTokenEndpoint,
          token: token,
          data: {'token': token},
        );
      }
      if (session != _sessionRevision) {
        throw StateError('Session changed during refresh');
      }
      final currentToken = prefs.getString(_tokenKey);
      if (currentToken == null) return null;
      if (currentToken != token) {
        throw StateError('Session changed during refresh');
      }
      final user = User.fromJson(data);
      if (user.id == null || user.id!.isEmpty) {
        throw const FormatException('Account response is missing its identity');
      }
      return user;
    } catch (error) {
      if (session != _sessionRevision) rethrow;
      final currentToken = prefs.getString(_tokenKey);
      if (currentToken == null) return null;
      if (currentToken != token) rethrow;
      if (error is DioException &&
          (error.response?.statusCode == 401 ||
              error.response?.statusCode == 403)) {
        await prefs.remove(_tokenKey);
        return null;
      }
      rethrow;
    }
  }

  /// A previous logout must not remove the token from a newer sign-in.
  Future<void> logout() async {
    final session = ++_sessionRevision;
    final prefs = await SharedPreferences.getInstance();
    if (session == _sessionRevision) await prefs.remove(_tokenKey);
  }

  /// Gets the stored token
  Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }
}

class AuthResponse {
  final User user;
  final String token;

  AuthResponse({required this.user, required this.token});
}
