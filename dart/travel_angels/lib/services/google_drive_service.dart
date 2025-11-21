import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';

class GoogleDriveService {
  final APIClient _apiClient;

  GoogleDriveService(this._apiClient);

  /// Gets the connection status of Google Drive
  /// Returns a map with "connected" boolean
  Future<Map<String, dynamic>> getStatus() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.googleDriveStatusEndpoint,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }

  /// Gets the OAuth authorization URL for Google Drive
  /// Returns a map with "authUrl" and "state"
  Future<Map<String, dynamic>> getAuthUrl() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.googleDriveAuthEndpoint,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }

  /// Revokes Google Drive access
  /// Returns a map with "message"
  Future<Map<String, dynamic>> revoke() async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.googleDriveRevokeEndpoint,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }
}

