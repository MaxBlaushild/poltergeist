import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// API Client for making HTTP requests to the Poltergeist API
/// Mirrors the functionality of the JavaScript API client
class APIClient {
  final Dio _client;
  static const String _tokenKey = 'token';
  VoidCallback? _onAuthError;

  /// Creates a new APIClient instance
  /// 
  /// [baseURL] - The base URL for the API
  /// [onAuthError] - Optional callback to be called when a 401/403 error occurs
  APIClient(String baseURL, {VoidCallback? onAuthError})
      : _client = Dio(BaseOptions(baseUrl: baseURL)),
        _onAuthError = onAuthError {
    _setupInterceptors();
  }

  /// Sets the callback to be called when a 401/403 error occurs
  void setOnAuthError(VoidCallback? callback) {
    _onAuthError = callback;
  }

  void _setupInterceptors() {
    // Request interceptor - adds authentication token
    _client.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        // Add authentication token from storage
        final prefs = await SharedPreferences.getInstance();
        final token = prefs.getString(_tokenKey);
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }

        return handler.next(options);
      },
      onResponse: (response, handler) {
        return handler.next(response);
      },
      onError: (error, handler) async {
        // Log error response body for debugging
        if (error.response != null) {
          print('API Error - Status: ${error.response?.statusCode}');
          print('API Error - URL: ${error.requestOptions.uri}');
          if (error.response?.data != null) {
            print('API Error - Response body: ${error.response?.data}');
          }
        }
        
        // Handle 401/403 errors by clearing invalid token
        if (error.response?.statusCode == 401 ||
            error.response?.statusCode == 403) {
          final prefs = await SharedPreferences.getInstance();
          await prefs.remove(_tokenKey);
          // Call the auth error callback if provided
          _onAuthError?.call();
        }
        return handler.next(error);
      },
    ));
  }

  /// Makes a GET request
  /// 
  /// [url] - The endpoint URL (relative to baseURL)
  /// [params] - Optional query parameters
  /// 
  /// Returns the parsed response data
  Future<T> get<T>(String url, {Map<String, dynamic>? params}) async {
    final response = await _client.get<T>(
      url,
      queryParameters: params,
      options: Options(responseType: ResponseType.json),
    );
    return response.data as T;
  }

  /// Makes a POST request
  /// 
  /// [url] - The endpoint URL (relative to baseURL)
  /// [data] - Optional request body data
  /// 
  /// Returns the parsed response data
  Future<T> post<T>(String url, {dynamic data}) async {
    final response = await _client.post<T>(
      url,
      data: data,
      options: Options(responseType: ResponseType.json),
    );
    return response.data as T;
  }

  /// Makes a PUT request
  /// 
  /// [url] - The endpoint URL (relative to baseURL)
  /// [data] - Optional request body data
  /// 
  /// Returns the parsed response data
  Future<T> put<T>(String url, {dynamic data}) async {
    final response = await _client.put<T>(
      url,
      data: data,
      options: Options(responseType: ResponseType.json),
    );
    return response.data as T;
  }

  /// Makes a PATCH request
  /// 
  /// [url] - The endpoint URL (relative to baseURL)
  /// [data] - Optional request body data
  /// 
  /// Returns the parsed response data
  Future<T> patch<T>(String url, {dynamic data}) async {
    final response = await _client.patch<T>(
      url,
      data: data,
      options: Options(responseType: ResponseType.json),
    );
    return response.data as T;
  }

  /// Makes a DELETE request
  /// 
  /// [url] - The endpoint URL (relative to baseURL)
  /// [data] - Optional request body data
  /// 
  /// Returns the parsed response data
  Future<T> delete<T>(String url, {dynamic data}) async {
    final response = await _client.delete<T>(
      url,
      data: data,
      options: Options(responseType: ResponseType.json),
    );
    return response.data as T;
  }
}
