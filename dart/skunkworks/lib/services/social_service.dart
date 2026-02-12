import 'package:dio/dio.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/social_account.dart';
import 'package:skunkworks/services/api_client.dart';

class SocialService {
  final APIClient _apiClient;

  SocialService(this._apiClient);

  Future<List<SocialAccount>> getAccounts() async {
    final response = await _apiClient.get<List<dynamic>>(
      ApiConstants.socialAccountsEndpoint,
    );
    return response
        .map((item) => SocialAccount.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<String> getAuthUrl(String provider) async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.socialAuthEndpoint(provider),
      );
      final authUrl = response['authUrl'] as String?;
      if (authUrl == null || authUrl.isEmpty) {
        throw Exception('Failed to get auth URL');
      }
      return authUrl;
    } on DioException catch (e) {
      final data = e.response?.data;
      if (data is Map<String, dynamic>) {
        final message = data['error'];
        if (message is String && message.isNotEmpty) {
          throw Exception(message);
        }
      }
      rethrow;
    }
  }

  Future<void> revoke(String provider) async {
    await _apiClient.post(
      ApiConstants.socialRevokeEndpoint(provider),
    );
  }

  Future<Map<String, dynamic>> postToSocial({
    required String provider,
    required String postId,
  }) async {
    final response = await _apiClient.post<Map<String, dynamic>>(
      ApiConstants.socialPostEndpoint(provider),
      data: {
        'postId': postId,
      },
    );
    return response;
  }
}
