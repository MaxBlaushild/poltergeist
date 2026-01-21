import 'dart:io';
import 'package:dio/dio.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/services/api_client.dart';

class MediaService {
  final APIClient _apiClient;

  MediaService(this._apiClient);

  /// Gets a presigned upload URL from the backend
  /// 
  /// [bucket] - The S3 bucket name
  /// [key] - The S3 object key
  /// 
  /// Returns the presigned URL or null if failed
  Future<String?> getPresignedUploadURL(String bucket, String key) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.presignedUploadUrlEndpoint,
        data: {
          'bucket': bucket,
          'key': key,
        },
      );

      return response['url'] as String?;
    } catch (e) {
      print('Failed to get presigned upload URL: $e');
      return null;
    }
  }

  /// Uploads a file to S3 using a presigned URL
  /// 
  /// [presignedUrl] - The presigned URL from getPresignedUploadURL
  /// [file] - The file to upload
  /// 
  /// Returns true if upload was successful
  Future<bool> uploadMedia(String presignedUrl, File file) async {
    try {
      final fileName = file.path.split('/').last;
      final extension = fileName.split('.').last.toLowerCase();
      
      // Determine content type based on extension
      String contentType = 'application/octet-stream';
      if (extension == 'jpg' || extension == 'jpeg') {
        contentType = 'image/jpeg';
      } else if (extension == 'png') {
        contentType = 'image/png';
      } else if (extension == 'webp') {
        contentType = 'image/webp';
      } else if (extension == 'gif') {
        contentType = 'image/gif';
      }

      final dio = Dio();
      final fileBytes = await file.readAsBytes();
      
      final response = await dio.put(
        presignedUrl,
        data: fileBytes,
        options: Options(
          headers: {
            'Content-Type': contentType,
          },
        ),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Failed to upload media: $e');
      return false;
    }
  }

  /// Uploads a post image and returns the final S3 URL
  /// 
  /// [file] - The image file to upload
  /// [userId] - The user ID for generating the key
  /// 
  /// Returns the final S3 URL or null if upload failed
  Future<String?> uploadPostImage(File file, String userId) async {
    try {
      final timestamp = DateTime.now().millisecondsSinceEpoch.toString();
      final fileName = file.path.split('/').last;
      final extension = fileName.split('.').last.toLowerCase();
      final key = '$userId/$timestamp.$extension';

      final presignedUrl = await getPresignedUploadURL(ApiConstants.postsBucket, key);
      if (presignedUrl == null) {
        return null;
      }

      final uploadSuccess = await uploadMedia(presignedUrl, file);
      if (!uploadSuccess) {
        return null;
      }

      // Extract the final S3 URL from the presigned URL (remove query parameters)
      final uri = Uri.parse(presignedUrl);
      final finalUrl = '${uri.scheme}://${uri.host}${uri.path}';
      
      return finalUrl;
    } catch (e) {
      print('Failed to upload post image: $e');
      return null;
    }
  }
}

