import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:dio/dio.dart';

class AuthService {
  final APIClient _apiClient;
  static const String _tokenKey = 'token';

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
        data: {
          'phoneNumber': formattedPhone,
          'appName': ApiConstants.appName,
        },
      );

      return response != null;
    } catch (e) {
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
        data['dateOfBirth'] = '${dateOfBirth.year}-${dateOfBirth.month.toString().padLeft(2, '0')}-${dateOfBirth.day.toString().padLeft(2, '0')}';
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
        data['dateOfBirth'] = '${dateOfBirth.year}-${dateOfBirth.month.toString().padLeft(2, '0')}-${dateOfBirth.day.toString().padLeft(2, '0')}';
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
        ApiConstants.profileEndpoint,
        data: data,
      );

      return User.fromJson(response);
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

      // Use whoami endpoint which returns user in same format as login/register
      try {
        final user = await _apiClient.get<Map<String, dynamic>>(
          ApiConstants.whoamiEndpoint,
        );
        return User.fromJson(user);
      } catch (e) {
        // If whoami fails, fall back to verifyToken endpoint
        final dio = Dio(BaseOptions(baseUrl: ApiConstants.baseUrl));
        final response = await dio.post<Map<String, dynamic>>(
          ApiConstants.verifyTokenEndpoint,
          data: {
            'token': token,
          },
        );

        return User.fromJson(response.data!);
      }
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
}

class AuthResponse {
  final User user;
  final String token;

  AuthResponse({required this.user, required this.token});
}

