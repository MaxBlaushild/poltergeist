import 'dart:io';
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';

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
      } else if (extension == 'mp4') {
        contentType = 'video/mp4';
      } else if (extension == 'mov') {
        contentType = 'video/quicktime';
      } else if (extension == 'm4v') {
        contentType = 'video/x-m4v';
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

  /// Uploads a profile picture and returns the final S3 URL
  /// 
  /// [file] - The image file to upload
  /// [userId] - The user ID for generating the key
  /// 
  /// Returns the final S3 URL or null if upload failed
  Future<String?> uploadProfilePicture(File file, String userId) async {
    try {
      final timestamp = DateTime.now().millisecondsSinceEpoch.toString();
      final fileName = file.path.split('/').last;
      final extension = fileName.split('.').last.toLowerCase();
      final key = '$userId-$timestamp.$extension';

      final presignedUrl = await getPresignedUploadURL('crew-profile-icons', key);
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
      print('Failed to upload profile picture: $e');
      return null;
    }
  }

  /// Uploads an edited video and returns the final S3 URL.
  ///
  /// [file] - The video file to upload (e.g. exported from video editor)
  /// [userId] - The user ID for generating the key
  ///
  /// Returns the final S3 URL or null if upload failed.
  Future<String?> uploadVideo(File file, String userId) async {
    try {
      final timestamp = DateTime.now().millisecondsSinceEpoch.toString();
      final fileName = file.path.split('/').last;
      final extension = fileName.split('.').last.toLowerCase();
      if (extension != 'mp4' && extension != 'mov' && extension != 'm4v') {
        print('Unsupported video extension: $extension');
        return null;
      }
      final key = 'videos/$userId-$timestamp.$extension';

      final presignedUrl = await getPresignedUploadURL('crew-profile-icons', key);
      if (presignedUrl == null) {
        return null;
      }

      final uploadSuccess = await uploadMedia(presignedUrl, file);
      if (!uploadSuccess) {
        return null;
      }

      final uri = Uri.parse(presignedUrl);
      final finalUrl = '${uri.scheme}://${uri.host}${uri.path}';
      return finalUrl;
    } catch (e) {
      print('Failed to upload video: $e');
      return null;
    }
  }

  /// Downloads a video from a URL to a temporary file
  ///
  /// [url] - The URL of the video to download
  ///
  /// Returns the downloaded File or null if download failed
  Future<File?> downloadVideo(String url) async {
    try {
      final dio = Dio();
      final response = await dio.get<List<int>>(
        url,
        options: Options(responseType: ResponseType.bytes),
      );

      if (response.statusCode != 200 || response.data == null) {
        return null;
      }

      final tempDir = await getTemporaryDirectory();
      final extension = _getVideoExtension(url);
      final file = File('${tempDir.path}/video_${DateTime.now().millisecondsSinceEpoch}.$extension');
      await file.writeAsBytes(response.data!);
      return file;
    } catch (e) {
      print('Failed to download video: $e');
      return null;
    }
  }

  String _getVideoExtension(String url) {
    final uri = Uri.parse(url);
    final path = uri.path.toLowerCase();
    if (path.endsWith('.mp4')) return 'mp4';
    if (path.endsWith('.mov')) return 'mov';
    if (path.endsWith('.m4v')) return 'm4v';
    return 'mp4'; // Default
  }
}

