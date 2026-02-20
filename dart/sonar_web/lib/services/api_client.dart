import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/location.dart';

class ApiClient {
  final Dio _client;
  static const String _tokenKey = 'token';
  VoidCallback? _onAuthError;
  AppLocation? Function()? _getLocation;

  ApiClient(
    String baseUrl, {
    VoidCallback? onAuthError,
    AppLocation? Function()? getLocation,
  })  : _client = Dio(BaseOptions(baseUrl: baseUrl)),
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
    _client.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final prefs = await SharedPreferences.getInstance();
        final token = prefs.getString(_tokenKey);
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        final loc = _getLocation?.call();
        if (loc != null) {
          options.headers['X-User-Location'] = loc.headerValue;
        }
        return handler.next(options);
      },
      onResponse: (response, handler) => handler.next(response),
      onError: (error, handler) async {
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
    ));
  }

  Future<T> get<T>(
    String url, {
    Map<String, dynamic>? params,
    bool skipAuthError = false,
  }) async {
    final response = await _client.get<T>(
      url,
      queryParameters: params,
      options: Options(
        responseType: ResponseType.json,
        extra: {'skipAuthError': skipAuthError},
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
  Future<void> putRaw(String url, List<int> body, String contentType) async {
    await Dio().put<String>(
      url,
      data: body,
      options: Options(
        contentType: contentType,
        headers: {'Content-Type': contentType},
      ),
    );
  }
}
