import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/constants/api_constants.dart';

class TravelCalendarException implements Exception {
  const TravelCalendarException(this.message, {this.statusCode, this.details});

  final String message;
  final int? statusCode;
  final Map<String, dynamic>? details;

  @override
  String toString() => message;
}

/// Calendar requests keep guest verification distinct from account login.
/// A forbidden stop is an access decision, never a reason to log out the app.
class TravelCalendarService {
  TravelCalendarService({Dio? client})
    : _client =
          client ??
          Dio(
            BaseOptions(
              baseUrl: ApiConstants.baseUrl,
              connectTimeout: const Duration(seconds: 20),
              receiveTimeout: const Duration(seconds: 45),
            ),
          );

  final Dio _client;
  static const guestTokenKey = 'travel_calendar_guest_token';
  static const guestExpiryKey = 'travel_calendar_guest_expires_at';
  static const guestRecipientKey = 'travel_calendar_guest_recipient';
  static const guestEmailKey = 'travel_calendar_guest_email';
  static const guestClaimedAccountKey = 'travel_calendar_guest_claimed_account';

  String sharedLink(String kind, String token) {
    const configured = String.fromEnvironment('TRAVEL_ANGELS_WEB_URL');
    final base = configured.isNotEmpty
        ? configured
        : kIsWeb
        ? Uri.base.origin
        : '';
    if (base.isEmpty) {
      throw const TravelCalendarException(
        'Set TRAVEL_ANGELS_WEB_URL to share browser links from the mobile app.',
      );
    }
    return '${base.replaceFirst(RegExp(r'/+$'), '')}/#/$kind/${Uri.encodeComponent(token)}';
  }

  Future<bool> get hasGuestSession async => await _guestToken() != null;

  Future<bool> guestNeedsClaim(String accountId) async {
    final prefs = await SharedPreferences.getInstance();
    final hasIdentity =
        prefs.containsKey(guestEmailKey) || prefs.containsKey(guestTokenKey);
    return hasIdentity && prefs.getString(guestClaimedAccountKey) != accountId;
  }

  Future<void> markGuestClaimed(String accountId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(guestClaimedAccountKey, accountId);
  }

  Future<void> clearGuestSession({bool keepIdentity = false}) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(guestTokenKey);
    await prefs.remove(guestExpiryKey);
    await prefs.remove(guestRecipientKey);
    if (!keepIdentity) {
      await prefs.remove(guestEmailKey);
      await prefs.remove(guestClaimedAccountKey);
    }
  }

  Future<String?> _guestToken() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString(guestTokenKey);
    final expiry = DateTime.tryParse(prefs.getString(guestExpiryKey) ?? '');
    if (token != null && (expiry == null || !expiry.isAfter(DateTime.now()))) {
      await clearGuestSession(keepIdentity: true);
      return null;
    }
    return token;
  }

  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) => _request('GET', path, query: query);

  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
  }) => _request('POST', path, data: data);

  Future<Map<String, dynamic>> put(String path, {Map<String, dynamic>? data}) =>
      _request('PUT', path, data: data);

  Future<Map<String, dynamic>> delete(
    String path, {
    Map<String, dynamic>? data,
  }) => _request('DELETE', path, data: data);

  Future<Map<String, dynamic>> _request(
    String method,
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
  }) async {
    final endpoint = path.startsWith('/travel-angels/')
        ? path
        : '/travel-angels${path.startsWith('/') ? path : '/$path'}';
    final prefs = await SharedPreferences.getInstance();
    final accountToken = prefs.getString('token');
    final guestToken = await _guestToken();
    try {
      final response = await _client.request<dynamic>(
        endpoint,
        data: data,
        queryParameters: query,
        options: Options(
          method: method,
          responseType: ResponseType.json,
          headers: {
            if (accountToken != null) 'Authorization': 'Bearer $accountToken',
            if (guestToken != null) 'X-Travel-Guest-Token': guestToken,
          },
        ),
      );
      final body = response.data;
      final result = body is Map
          ? Map<String, dynamic>.from(body)
          : <String, dynamic>{if (body is List) 'items': body};
      if (endpoint == '/travel-angels/calendar-guests/verification' &&
          data?['email'] is String) {
        await prefs.setString(guestEmailKey, (data!['email'] as String).trim());
      }
      if (endpoint == '/travel-angels/calendar-guests/verify' &&
          result['guestToken'] is String &&
          result['expiresAt'] is String) {
        await prefs.remove(guestClaimedAccountKey);
        await prefs.setString(guestTokenKey, result['guestToken'] as String);
        await prefs.setString(guestExpiryKey, result['expiresAt'] as String);
        if (result['recipientId'] is String) {
          await prefs.setString(
            guestRecipientKey,
            result['recipientId'] as String,
          );
        }
      }
      return result;
    } on DioException catch (error) {
      final body = error.response?.data;
      final details = body is Map ? Map<String, dynamic>.from(body) : null;
      throw TravelCalendarException(
        details?['error']?.toString() ??
            'Unable to reach Travel Angels. Please try again.',
        statusCode: error.response?.statusCode,
        details: details,
      );
    }
  }
}
