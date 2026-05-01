import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/location.dart';

class ApiClient {
  final Dio _client;
  static const String _tokenKey = 'token';
  static const Duration _connectTimeout = Duration(seconds: 15);
  static const Duration _receiveTimeout = Duration(seconds: 45);
  static const Duration _sendTimeout = Duration(minutes: 2);
  VoidCallback? _onAuthError;
  AppLocation? Function()? _getLocation;

  ApiClient(
    String baseUrl, {
    VoidCallback? onAuthError,
    AppLocation? Function()? getLocation,
  }) : _client = Dio(
         BaseOptions(
           baseUrl: baseUrl,
           connectTimeout: _connectTimeout,
           receiveTimeout: _receiveTimeout,
           sendTimeout: _sendTimeout,
         ),
       ),
       _onAuthError = onAuthError,
       _getLocation = getLocation {
    _setupInterceptors();
  }

  void setOnAuthError(VoidCallback? callback) {
    _onAuthError = callback;
  }

  void setGetLocation(AppLocation? Function()? getLocation) {
    _getLocation = getLocation;
  }

  void _setupInterceptors() {
    _client.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          options.extra['_requestStartedAt'] ??= DateTime.now();
          final prefs = await SharedPreferences.getInstance();
          final token = prefs.getString(_tokenKey);
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          final loc = _getLocation?.call();
          if (loc != null) {
            options.headers['X-User-Location'] = loc.headerValue;
          }
          final traceId = options.headers['X-Map-Trace-Id']?.toString().trim();
          final traceLabel = options.extra['traceLabel']?.toString().trim();
          if (kDebugMode &&
              ((traceId != null && traceId.isNotEmpty) ||
                  (traceLabel != null && traceLabel.isNotEmpty))) {
            debugPrint(
              'ApiClient trace start '
              '${options.method} ${options.uri} '
              'traceId=${traceId ?? '-'} traceLabel=${traceLabel ?? '-'}',
            );
          }
          return handler.next(options);
        },
        onResponse: (response, handler) {
          if (kDebugMode) {
            final startedAt =
                response.requestOptions.extra['_requestStartedAt'] as DateTime?;
            final traceId = response.requestOptions.headers['X-Map-Trace-Id']
                ?.toString()
                .trim();
            final traceLabel = response.requestOptions.extra['traceLabel']
                ?.toString()
                .trim();
            if ((traceId != null && traceId.isNotEmpty) ||
                (traceLabel != null && traceLabel.isNotEmpty)) {
              final elapsedMs = startedAt == null
                  ? -1
                  : DateTime.now().difference(startedAt).inMilliseconds;
              debugPrint(
                'ApiClient trace done '
                '${response.requestOptions.method} ${response.requestOptions.uri} '
                'status=${response.statusCode} '
                'elapsedMs=$elapsedMs '
                'traceId=${traceId ?? '-'} traceLabel=${traceLabel ?? '-'}',
              );
            }
          }
          return handler.next(response);
        },
        onError: (error, handler) async {
          if (kDebugMode) {
            final startedAt =
                error.requestOptions.extra['_requestStartedAt'] as DateTime?;
            final traceId = error.requestOptions.headers['X-Map-Trace-Id']
                ?.toString()
                .trim();
            final traceLabel = error.requestOptions.extra['traceLabel']
                ?.toString()
                .trim();
            if ((traceId != null && traceId.isNotEmpty) ||
                (traceLabel != null && traceLabel.isNotEmpty)) {
              final elapsedMs = startedAt == null
                  ? -1
                  : DateTime.now().difference(startedAt).inMilliseconds;
              debugPrint(
                'ApiClient trace error '
                '${error.requestOptions.method} ${error.requestOptions.uri} '
                'status=${error.response?.statusCode} '
                'elapsedMs=$elapsedMs '
                'traceId=${traceId ?? '-'} traceLabel=${traceLabel ?? '-'} '
                'error=${error.message}',
              );
            }
          }
          if (error.response?.statusCode == 401 ||
              error.response?.statusCode == 403) {
            final skipAuthError =
                error.requestOptions.extra['skipAuthError'] == true;
            if (kDebugMode) {
              final status = error.response?.statusCode;
              final method = error.requestOptions.method;
              final uri = error.requestOptions.uri;
              final hadAuthHeader =
                  error.requestOptions.headers['Authorization'] != null;
              debugPrint(
                'ApiClient auth error: $status $method $uri '
                '(skipAuthError=$skipAuthError, hadAuthHeader=$hadAuthHeader)',
              );
              debugPrint('ApiClient auth error body: ${error.response?.data}');
            }
            if (!skipAuthError) {
              final prefs = await SharedPreferences.getInstance();
              await prefs.remove(_tokenKey);
              _onAuthError?.call();
            }
          }
          return handler.next(error);
        },
      ),
    );
  }

  Future<T> get<T>(
    String url, {
    Map<String, dynamic>? params,
    bool skipAuthError = false,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
  }) async {
    final requestExtra = <String, dynamic>{
      'skipAuthError': skipAuthError,
      if (extra != null) ...extra,
    };
    final response = await _client.get<T>(
      url,
      queryParameters: params,
      options: Options(
        responseType: ResponseType.json,
        headers: headers,
        extra: requestExtra,
      ),
    );
    return response.data as T;
  }

  Future<T> post<T>(
    String url, {
    dynamic data,
    bool skipAuthError = false,
  }) async {
    final response = await _client.post<T>(
      url,
      data: data,
      options: Options(
        responseType: ResponseType.json,
        extra: {'skipAuthError': skipAuthError},
      ),
    );
    return response.data as T;
  }

  Future<T> put<T>(
    String url, {
    dynamic data,
    bool skipAuthError = false,
  }) async {
    final response = await _client.put<T>(
      url,
      data: data,
      options: Options(
        responseType: ResponseType.json,
        extra: {'skipAuthError': skipAuthError},
      ),
    );
    return response.data as T;
  }

  Future<T> patch<T>(
    String url, {
    dynamic data,
    bool skipAuthError = false,
  }) async {
    final response = await _client.patch<T>(
      url,
      data: data,
      options: Options(
        responseType: ResponseType.json,
        extra: {'skipAuthError': skipAuthError},
      ),
    );
    return response.data as T;
  }

  Future<T> delete<T>(
    String url, {
    dynamic data,
    bool skipAuthError = false,
  }) async {
    final response = await _client.delete<T>(
      url,
      data: data,
      options: Options(
        responseType: ResponseType.json,
        extra: {'skipAuthError': skipAuthError},
      ),
    );
    return response.data as T;
  }

  /// PUT raw bytes to a URL (e.g. presigned S3). No baseUrl.
  Future<Response<String>> putRaw(
    String url,
    List<int> body,
    String contentType, {
    void Function(int sent, int total)? onSendProgress,
    Duration? overallTimeout,
  }) async {
    final request =
        Dio(
          BaseOptions(
            connectTimeout: _connectTimeout,
            receiveTimeout: _receiveTimeout,
            sendTimeout: _sendTimeout,
          ),
        ).put<String>(
          url,
          data: body,
          onSendProgress: onSendProgress,
          options: Options(
            responseType: ResponseType.plain,
            contentType: contentType,
            headers: {'Content-Type': contentType},
          ),
        );
    if (overallTimeout == null) {
      return request;
    }
    return request.timeout(overallTimeout);
  }
}
