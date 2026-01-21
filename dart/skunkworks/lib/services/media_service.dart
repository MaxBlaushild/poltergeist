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
  /// [contentType] - Optional content type for the upload
  /// 
  /// Returns the presigned URL or null if failed
  Future<String?> getPresignedUploadURL(String bucket, String key, {String? contentType}) async {
    try {
      final data = <String, dynamic>{
        'bucket': bucket,
        'key': key,
      };
      if (contentType != null) {
        data['contentType'] = contentType;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.presignedUploadUrlEndpoint,
        data: data,
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
      
      print('Uploading to S3: ${presignedUrl.substring(0, presignedUrl.indexOf('?'))}...');
      print('Content-Type: $contentType');
      print('File size: ${fileBytes.length} bytes');
      
      final response = await dio.put(
        presignedUrl,
        data: fileBytes,
        options: Options(
          headers: {
            'Content-Type': contentType,
          },
          validateStatus: (status) => status! < 500, // Don't throw on 4xx errors
        ),
      );

      if (response.statusCode != 200) {
        print('Upload failed with status ${response.statusCode}: ${response.data}');
        return false;
      }

      return true;
    } catch (e) {
      print('Failed to upload media: $e');
      if (e is DioException) {
        print('DioException details: ${e.response?.statusCode} - ${e.response?.data}');
        print('Request URL: ${e.requestOptions.uri}');
      }
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
      final key = '$userId/posts/$timestamp.$extension';

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

      final presignedUrl = await getPresignedUploadURL(ApiConstants.postsBucket, key, contentType: contentType);
      if (presignedUrl == null) {
        print('Failed to get presigned URL');
        return null;
      }

      print('Got presigned URL: $presignedUrl');
      final uploadSuccess = await uploadMedia(presignedUrl, file);
      if (!uploadSuccess) {
        print('Failed to upload media');
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

