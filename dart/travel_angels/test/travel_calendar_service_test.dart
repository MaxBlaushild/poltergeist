import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';

class _Adapter implements HttpClientAdapter {
  _Adapter(this.reply);
  final ResponseBody Function(RequestOptions) reply;
  final List<RequestOptions> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? stream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return reply(options);
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody _json(Map<String, dynamic> body, {int status = 200}) =>
    ResponseBody.fromString(
      jsonEncode(body),
      status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() => SharedPreferences.setMockInitialValues({}));

  test('trip images use guest credentials and reject foreign origins', () async {
    SharedPreferences.setMockInitialValues({
      'token': 'account-token',
      TravelCalendarService.guestTokenKey: 'guest-token',
      TravelCalendarService.guestExpiryKey: DateTime.now()
          .add(const Duration(hours: 1))
          .toIso8601String(),
    });
    final adapter = _Adapter(
      (_) => ResponseBody.fromBytes(
        [137, 80, 78, 71],
        200,
        headers: {
          Headers.contentTypeHeader: ['image/png'],
        },
      ),
    );
    final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
      ..httpClientAdapter = adapter;
    final service = TravelCalendarService(client: dio);
    final bytes = await service.loadHero(
      '/travel-angels/calendar/stops/trip-123/hero/media-123',
      query: {'calendarToken': 'observed-link'},
    );
    expect(bytes, [137, 80, 78, 71]);
    expect(
      adapter.requests.single.headers['Authorization'],
      'Bearer account-token',
    );
    expect(
      adapter.requests.single.headers['X-Travel-Guest-Token'],
      'guest-token',
    );
    expect(
      adapter.requests.single.queryParameters['calendarToken'],
      'observed-link',
    );
    expect(adapter.requests.single.followRedirects, isFalse);
    for (final url in [
      'https://other.test/travel-angels/calendar/stops/trip-123/hero/media-123',
      '//other.test/travel-angels/calendar/stops/trip-123/hero/media-123',
      'file:/travel-angels/calendar/stops/trip-123/hero/media-123',
      'https://example.test/unrelated-image',
    ]) {
      await expectLater(
        service.loadHero(url),
        throwsA(isA<TravelCalendarException>()),
      );
    }
    expect(adapter.requests, hasLength(1));
  });

  test(
    'hero upload sends the revision and framing in one multipart request',
    () async {
      final adapter = _Adapter(
        (_) => _json({
          'stop': {'revision': 8},
        }),
      );
      final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
        ..httpClientAdapter = adapter;
      final result = await TravelCalendarService(client: dio).uploadHero(
        'trip-123',
        Uint8List.fromList([1, 2, 3]),
        'cover.png',
        altText: 'Cliffs above the Atlantic',
        focalX: .25,
        focalY: .7,
        revision: 7,
      );
      expect((result['stop'] as Map)['revision'], 8);
      final request = adapter.requests.single;
      expect(request.path, '/travel-angels/calendar/stops/trip-123/hero');
      final form = request.data as FormData;
      final fields = Map.fromEntries(form.fields);
      expect(fields['revision'], '7');
      expect(fields['altText'], 'Cliffs above the Atlantic');
      expect(fields['focalX'], '0.25');
      expect(form.files.single.value.filename, 'cover.png');
    },
  );

  test(
    'guest verification persists scoped access without creating app login',
    () async {
      final adapter = _Adapter(
        (request) => request.path.endsWith('/verify')
            ? _json({
                'guestToken': 'guest-secret',
                'recipientId': 'recipient-123',
                'expiresAt': DateTime.now()
                    .add(const Duration(hours: 1))
                    .toIso8601String(),
              })
            : _json({'stops': []}),
      );
      final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
        ..httpClientAdapter = adapter;
      final service = TravelCalendarService(client: dio);
      await service.post(
        '/calendar-guests/verify',
        data: {'challengeId': 'challenge', 'code': '123456'},
      );
      await service.get('/shared/calendars/share-token');
      final prefs = await SharedPreferences.getInstance();
      expect(prefs.getString('token'), isNull);
      expect(
        adapter.requests.last.headers['X-Travel-Guest-Token'],
        'guest-secret',
      );
      expect(adapter.requests.last.headers['Authorization'], isNull);
      expect(await service.hasGuestSession, isTrue);
    },
  );

  test('stop permission denial preserves the signed-in app account', () async {
    SharedPreferences.setMockInitialValues({'token': 'valid-account-token'});
    final adapter = _Adapter(
      (_) => _json({'error': 'This stop is restricted'}, status: 403),
    );
    final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
      ..httpClientAdapter = adapter;
    final service = TravelCalendarService(client: dio);
    await expectLater(
      service.get('/calendar/stops/forbidden'),
      throwsA(
        isA<TravelCalendarException>().having(
          (error) => error.statusCode,
          'status',
          403,
        ),
      ),
    );
    final prefs = await SharedPreferences.getInstance();
    expect(prefs.getString('token'), 'valid-account-token');
  });

  test(
    'expired guest access is not reused and never removes account auth',
    () async {
      SharedPreferences.setMockInitialValues({
        'token': 'account-token',
        TravelCalendarService.guestTokenKey: 'expired-guest',
        TravelCalendarService.guestExpiryKey: DateTime.now()
            .subtract(const Duration(minutes: 1))
            .toIso8601String(),
      });
      final adapter = _Adapter((_) => _json({'stops': []}));
      final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
        ..httpClientAdapter = adapter;
      await TravelCalendarService(client: dio).get('/shared/calendars/token');
      expect(adapter.requests.single.headers['X-Travel-Guest-Token'], isNull);
      expect(
        adapter.requests.single.headers['Authorization'],
        'Bearer account-token',
      );
    },
  );
}
