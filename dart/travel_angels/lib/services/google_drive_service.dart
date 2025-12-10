import 'package:dio/dio.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document_location.dart';
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

  /// Lists Google Drive files
  /// Returns a FileListResponse with files and nextPageToken
  Future<Map<String, dynamic>> listFiles({
    int? pageSize,
    String? pageToken,
    String? query,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (pageSize != null) queryParams['pageSize'] = pageSize.toString();
      if (pageToken != null) queryParams['pageToken'] = pageToken;
      if (query != null) queryParams['q'] = query;

      print('[GoogleDriveService] Listing files with params: $queryParams');
      print('[GoogleDriveService] Endpoint: ${ApiConstants.googleDriveFilesEndpoint}');
      
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.googleDriveFilesEndpoint,
        params: queryParams.isEmpty ? null : queryParams,
      );
      
      print('[GoogleDriveService] Successfully received response: ${response.keys}');
      return response;
    } catch (e) {
      print('[GoogleDriveService] Error listing files: $e');
      print('[GoogleDriveService] Error type: ${e.runtimeType}');
      if (e is DioException) {
        print('[GoogleDriveService] DioException details:');
        print('  - Response status: ${e.response?.statusCode}');
        print('  - Response status message: ${e.response?.statusMessage}');
        print('  - Response data: ${e.response?.data}');
        print('  - Request path: ${e.requestOptions.path}');
        print('  - Request method: ${e.requestOptions.method}');
        print('  - Request headers: ${e.requestOptions.headers}');
        print('  - Error message: ${e.message}');
        print('  - Error type: ${e.type}');
      }
      rethrow;
    }
  }

  /// Imports a Google Drive document
  /// fileId: The Google Drive file ID
  /// importType: Either "import" or "reference"
  /// locations: Optional list of document locations
  /// Returns the created document
  Future<Map<String, dynamic>> importDocument(
    String fileId,
    String importType, {
    List<DocumentLocation>? locations,
  }) async {
    try {
      final data = <String, dynamic>{
        'fileId': fileId,
        'importType': importType,
      };

      if (locations != null && locations.isNotEmpty) {
        data['locations'] = locations.map((loc) => {
          'placeId': loc.placeId,
          'name': loc.name,
          'formattedAddress': loc.formattedAddress,
          'latitude': loc.latitude,
          'longitude': loc.longitude,
          'type': loc.locationType.name,
        }).toList();
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.googleDriveImportDocumentEndpoint,
        data: data,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }
}

