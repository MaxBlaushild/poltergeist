import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../constants/api_constants.dart';
import 'api_client.dart';

class MediaUploadResult {
  const MediaUploadResult({
    required this.success,
    required this.duration,
    required this.bytesSent,
    required this.totalBytes,
    this.statusCode,
    this.timedOut = false,
    this.errorDescription,
  });

  final bool success;
  final Duration duration;
  final int bytesSent;
  final int totalBytes;
  final int? statusCode;
  final bool timedOut;
  final String? errorDescription;
}

class MediaService {
  final ApiClient _apiClient;

  MediaService(this._apiClient);

  /// Returns presigned upload URL for the given bucket and key.
  Future<String?> getPresignedUploadUrl(
    String bucket,
    String key, {
    String debugLabel = 'upload',
  }) async {
    final startedAt = DateTime.now();
    debugPrint(
      '[media-upload][$debugLabel] requesting presigned URL '
      'bucket=$bucket key=$key',
    );
    try {
      final data = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.mediaUploadUrlEndpoint,
        data: {'bucket': bucket, 'key': key},
      );
      final url = data['url'] as String?;
      debugPrint(
        '[media-upload][$debugLabel] presigned URL ready '
        'elapsed=${_formatDuration(DateTime.now().difference(startedAt))} '
        'hasUrl=${url != null}',
      );
      return url;
    } on DioException catch (error, stackTrace) {
      debugPrint(
        '[media-upload][$debugLabel] failed to fetch presigned URL '
        'elapsed=${_formatDuration(DateTime.now().difference(startedAt))} '
        'status=${error.response?.statusCode} '
        'type=${error.type} '
        'message=${error.message} '
        'body=${_describeResponseBody(error.response?.data)}',
      );
      debugPrintStack(stackTrace: stackTrace);
      return null;
    } catch (error, stackTrace) {
      debugPrint(
        '[media-upload][$debugLabel] failed to fetch presigned URL '
        'elapsed=${_formatDuration(DateTime.now().difference(startedAt))} '
        'error=$error',
      );
      debugPrintStack(stackTrace: stackTrace);
      return null;
    }
  }

  /// Uploads bytes to a presigned S3 URL via PUT.
  Future<MediaUploadResult> uploadToPresigned(
    String url,
    Uint8List bytes,
    String contentType, {
    String debugLabel = 'upload',
    Duration timeout = const Duration(minutes: 3),
  }) async {
    final startedAt = DateTime.now();
    final uri = Uri.tryParse(url);
    var bytesSent = 0;
    var totalBytes = bytes.length;
    var nextProgressPercent = 25;
    var sawProgress = false;

    debugPrint(
      '[media-upload][$debugLabel] starting upload '
      'host=${uri?.host ?? 'unknown'} '
      'path=${uri?.path ?? '<invalid-url>'} '
      'bytes=$totalBytes '
      'contentType=$contentType '
      'timeout=${timeout.inSeconds}s',
    );

    final heartbeat = Timer.periodic(const Duration(seconds: 10), (_) {
      final elapsed = DateTime.now().difference(startedAt);
      final percent = totalBytes <= 0
          ? 'unknown'
          : ((bytesSent / totalBytes) * 100).clamp(0, 100).toStringAsFixed(1);
      debugPrint(
        '[media-upload][$debugLabel] still waiting '
        'elapsed=${_formatDuration(elapsed)} '
        'sent=$bytesSent '
        'total=$totalBytes '
        'percent=$percent',
      );
    });

    try {
      final response = await _apiClient.putRaw(
        url,
        bytes,
        contentType,
        overallTimeout: timeout,
        onSendProgress: (sent, total) {
          bytesSent = sent;
          if (total > 0) {
            totalBytes = total;
          }
          if (!sawProgress && sent > 0) {
            sawProgress = true;
            debugPrint(
              '[media-upload][$debugLabel] first bytes sent '
              'elapsed=${_formatDuration(DateTime.now().difference(startedAt))} '
              'sent=$sent '
              'total=$totalBytes',
            );
          }
          if (totalBytes <= 0) {
            return;
          }
          final percent = ((sent / totalBytes) * 100).floor();
          while (percent >= nextProgressPercent && nextProgressPercent <= 100) {
            debugPrint(
              '[media-upload][$debugLabel] progress=$nextProgressPercent% '
              'elapsed=${_formatDuration(DateTime.now().difference(startedAt))} '
              'sent=$sent '
              'total=$totalBytes',
            );
            nextProgressPercent += 25;
          }
        },
      );
      final duration = DateTime.now().difference(startedAt);
      debugPrint(
        '[media-upload][$debugLabel] upload complete '
        'elapsed=${_formatDuration(duration)} '
        'status=${response.statusCode} '
        'sent=$bytesSent '
        'total=$totalBytes',
      );
      return MediaUploadResult(
        success: true,
        duration: duration,
        bytesSent: bytesSent,
        totalBytes: totalBytes,
        statusCode: response.statusCode,
      );
    } on TimeoutException catch (error) {
      final duration = DateTime.now().difference(startedAt);
      debugPrint(
        '[media-upload][$debugLabel] upload timed out '
        'elapsed=${_formatDuration(duration)} '
        'sent=$bytesSent '
        'total=$totalBytes '
        'error=$error',
      );
      return MediaUploadResult(
        success: false,
        duration: duration,
        bytesSent: bytesSent,
        totalBytes: totalBytes,
        timedOut: true,
        errorDescription: error.toString(),
      );
    } on DioException catch (error, stackTrace) {
      final duration = DateTime.now().difference(startedAt);
      debugPrint(
        '[media-upload][$debugLabel] upload failed '
        'elapsed=${_formatDuration(duration)} '
        'sent=$bytesSent '
        'total=$totalBytes '
        'status=${error.response?.statusCode} '
        'type=${error.type} '
        'message=${error.message} '
        'body=${_describeResponseBody(error.response?.data)}',
      );
      debugPrintStack(stackTrace: stackTrace);
      return MediaUploadResult(
        success: false,
        duration: duration,
        bytesSent: bytesSent,
        totalBytes: totalBytes,
        statusCode: error.response?.statusCode,
        errorDescription: error.message,
      );
    } catch (error, stackTrace) {
      final duration = DateTime.now().difference(startedAt);
      debugPrint(
        '[media-upload][$debugLabel] upload failed '
        'elapsed=${_formatDuration(duration)} '
        'sent=$bytesSent '
        'total=$totalBytes '
        'error=$error',
      );
      debugPrintStack(stackTrace: stackTrace);
      return MediaUploadResult(
        success: false,
        duration: duration,
        bytesSent: bytesSent,
        totalBytes: totalBytes,
        errorDescription: error.toString(),
      );
    } finally {
      heartbeat.cancel();
    }
  }

  static String _formatDuration(Duration duration) {
    final minutes = duration.inMinutes;
    final seconds = duration.inSeconds.remainder(60);
    final tenths = (duration.inMilliseconds.remainder(1000) / 100).floor();
    if (minutes > 0) {
      return '${minutes}m ${seconds}s';
    }
    if (duration.inSeconds > 0) {
      return '${duration.inSeconds}.${tenths}s';
    }
    return '${duration.inMilliseconds}ms';
  }

  static String _describeResponseBody(dynamic data) {
    if (data == null) {
      return '<empty>';
    }
    final text = data.toString().replaceAll('\n', ' ').trim();
    if (text.isEmpty) {
      return '<empty>';
    }
    if (text.length <= 180) {
      return text;
    }
    return '${text.substring(0, 177)}...';
  }
}
