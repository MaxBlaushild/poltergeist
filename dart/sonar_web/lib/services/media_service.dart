import 'dart:typed_data';

import '../constants/api_constants.dart';
import 'api_client.dart';

class MediaService {
  final ApiClient _apiClient;

  MediaService(this._apiClient);

  /// Returns presigned upload URL for the given bucket and key.
  Future<String?> getPresignedUploadUrl(String bucket, String key) async {
    try {
      final data = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.mediaUploadUrlEndpoint,
        data: {'bucket': bucket, 'key': key},
      );
      return data['url'] as String?;
    } catch (_) {
      return null;
    }
  }

  /// Uploads bytes to a presigned S3 URL via PUT.
  Future<bool> uploadToPresigned(
    String url,
    Uint8List bytes,
    String contentType,
  ) async {
    try {
      await _apiClient.putRaw(url, bytes, contentType);
      return true;
    } catch (_) {
      return false;
    }
  }
}
