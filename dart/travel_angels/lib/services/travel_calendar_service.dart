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

  /// Upload directly to the protected trip endpoint. Media never inherits the
  /// public document library's sharing rules or exposes a permanent asset URL.
  Future<Map<String, dynamic>> uploadHero(
    String stopId,
    Uint8List bytes,
    String filename, {
    required String altText,
    required double focalX,
    required double focalY,
    required int revision,
    double cropX = 0,
    double cropY = 0,
    double cropWidth = 1,
    double cropHeight = 1,
  }) {
    if (bytes.length > 8 * 1024 * 1024) {
      throw const TravelCalendarException('Choose an image smaller than 8 MB.');
    }
    return _request(
      'POST',
      '/calendar/stops/${Uri.encodeComponent(stopId)}/hero',
      data: FormData.fromMap({
        'image': MultipartFile.fromBytes(bytes, filename: filename),
        'revision': revision,
        'altText': altText,
        'focalX': focalX,
        'focalY': focalY,
        'cropX': cropX,
        'cropY': cropY,
        'cropWidth': cropWidth,
        'cropHeight': cropHeight,
      }),
    );
  }

  Future<Map<String, dynamic>> uploadLodgingPhoto(
    String stopId,
    String entryId,
    Uint8List bytes,
    String filename, {
    required String altText,
    required int revision,
  }) {
    if (bytes.isEmpty || bytes.length > 8 * 1024 * 1024) {
      throw const TravelCalendarException('Choose an image up to 8 MB.');
    }
    return _request(
      'POST',
      '/calendar/stops/${Uri.encodeComponent(stopId)}/lodging/${Uri.encodeComponent(entryId)}/photos',
      data: FormData.fromMap({
        'image': MultipartFile.fromBytes(bytes, filename: filename),
        'revision': revision,
        'altText': altText,
      }),
    );
  }

  Future<Uint8List> loadHero(String url, {Map<String, dynamic>? query}) =>
      _loadProtectedImage(
        url,
        query,
        RegExp(
          r'^/travel-angels/calendar/stops/[A-Za-z0-9-]+/hero/[A-Za-z0-9-]+$',
        ),
      );

  Future<Uint8List> loadLodgingImage(
    String url, {
    Map<String, dynamic>? query,
  }) => _loadProtectedImage(
    url,
    query,
    RegExp(
      r'^/travel-angels/calendar/stops/[A-Za-z0-9-]+/(lodging/[A-Za-z0-9-]+/(photos/[A-Za-z0-9-]+|property/photos/[0-9]+)|lodging-properties/[A-Za-z0-9_-]+/photos/[0-9]+)$',
    ),
  );

  Future<Uint8List> _loadProtectedImage(
    String url,
    Map<String, dynamic>? query,
    RegExp allowedPath,
  ) async {
    final base = Uri.parse(_client.options.baseUrl);
    final candidate = Uri.tryParse(url);
    if (candidate == null ||
        candidate.hasFragment ||
        candidate.userInfo.isNotEmpty ||
        (candidate.hasScheme && !candidate.hasAuthority) ||
        (candidate.hasAuthority &&
            (candidate.scheme != base.scheme ||
                candidate.host != base.host ||
                candidate.port != base.port)) ||
        !allowedPath.hasMatch(candidate.path)) {
      throw const TravelCalendarException('This trip image is unavailable.');
    }
    try {
      final response = await _client.get<List<int>>(
        url,
        queryParameters: query,
        options: Options(
          responseType: ResponseType.bytes,
          followRedirects: false,
          headers: await _identityHeaders(),
        ),
      );
      return Uint8List.fromList(response.data ?? const []);
    } on DioException {
      throw const TravelCalendarException('This trip image is unavailable.');
    }
  }

  Future<Map<String, String>> _identityHeaders() async {
    final prefs = await SharedPreferences.getInstance();
    final accountToken = prefs.getString('token');
    final guestToken = await _guestToken();
    return {
      if (accountToken != null) 'Authorization': 'Bearer $accountToken',
      if (guestToken != null) 'X-Travel-Guest-Token': guestToken,
    };
  }

  Future<Map<String, dynamic>> _request(
    String method,
    String path, {
    Object? data,
    Map<String, dynamic>? query,
  }) async {
    final endpoint = path.startsWith('/travel-angels/')
        ? path
        : '/travel-angels${path.startsWith('/') ? path : '/$path'}';
    final prefs = await SharedPreferences.getInstance();
    try {
      final response = await _client.request<dynamic>(
        endpoint,
        data: data,
        queryParameters: query,
        options: Options(
          method: method,
          responseType: ResponseType.json,
          headers: await _identityHeaders(),
        ),
      );
      final body = response.data;
      final result = body is Map
          ? Map<String, dynamic>.from(body)
          : <String, dynamic>{if (body is List) 'items': body};
      if (endpoint == '/travel-angels/calendar-guests/verification' &&
          data is Map &&
          data['email'] is String) {
        await prefs.setString(guestEmailKey, (data['email'] as String).trim());
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
