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

  test(
    'trip files carry scoped access without sending credentials elsewhere',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'account-token'});
      final adapter = _Adapter(
        (_) => ResponseBody.fromBytes(
          [37, 80, 68, 70],
          200,
          headers: {
            Headers.contentTypeHeader: ['application/pdf'],
          },
        ),
      );
      final service = TravelCalendarService(
        client: Dio(BaseOptions(baseUrl: 'https://example.test'))
          ..httpClientAdapter = adapter,
      );
      const path =
          '/travel-angels/calendar/stops/trip-123/documents/doc-123/file';
      expect(
        await service.loadTripDocumentFile(
          path,
          query: {'stopToken': 'trip-link'},
        ),
        [37, 80, 68, 70],
      );
      final request = adapter.requests.single;
      expect(request.headers['Authorization'], 'Bearer account-token');
      expect(request.queryParameters['stopToken'], 'trip-link');
      expect(request.followRedirects, isFalse);
      for (final url in [
        'https://other.test$path',
        '//other.test$path',
        'http://example.test$path',
        'https://user@example.test$path',
        'https://example.test:444$path',
        'file:$path',
        '$path?redirect=https://other.test',
        '$path#fragment',
        '/travel-angels/calendar/stops/trip-123/hero/doc-123',
        '/travel-angels/calendar/stops/trip-123/documents/doc-123/../../file',
      ]) {
        await expectLater(
          service.loadTripDocumentFile(url),
          throwsA(isA<TravelCalendarException>()),
        );
      }
      expect(adapter.requests, hasLength(1));
    },
  );

  test(
    'trip file access errors retain status and the account session',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'account-token'});
      for (final status in [302, 401, 403, 404]) {
        final adapter = _Adapter(
          (_) => ResponseBody.fromBytes(
            [],
            status,
            headers: {
              'location': ['https://other.test/file'],
            },
          ),
        );
        final service = TravelCalendarService(
          client: Dio(BaseOptions(baseUrl: 'https://example.test'))
            ..httpClientAdapter = adapter,
        );
        await expectLater(
          service.loadTripDocumentFile(
            '/travel-angels/calendar/stops/trip-123/documents/doc-123/file',
          ),
          throwsA(
            isA<TravelCalendarException>().having(
              (error) => error.statusCode,
              'status',
              status,
            ),
          ),
        );
        expect(adapter.requests, hasLength(1));
      }
      expect(
        (await SharedPreferences.getInstance()).getString('token'),
        'account-token',
      );
    },
  );

  test(
    'trip document upload keeps originals and rejects oversized files locally',
    () async {
      final adapter = _Adapter(
        (_) => _json({
          'document': {'id': 'doc-123'},
        }),
      );
      final service = TravelCalendarService(
        client: Dio(BaseOptions(baseUrl: 'https://example.test'))
          ..httpClientAdapter = adapter,
      );
      await service.uploadTripDocument(
        'trip-123',
        Uint8List.fromList([37, 80, 68, 70]),
        'itinerary.pdf',
        title: 'Our itinerary',
        query: {'stopToken': 'trip-link'},
      );
      final request = adapter.requests.single;
      expect(
        request.path,
        '/travel-angels/calendar/stops/trip-123/documents/upload',
      );
      expect(request.queryParameters['stopToken'], 'trip-link');
      final form = request.data as FormData;
      expect(Map.fromEntries(form.fields)['title'], 'Our itinerary');
      expect(form.files.single.key, 'file');
      expect(form.files.single.value.filename, 'itinerary.pdf');
      expect(form.files.single.value.length, 4);
      for (final bytes in [Uint8List(0), Uint8List(20 * 1024 * 1024 + 1)]) {
        expect(
          () => service.uploadTripDocument(
            'trip-123',
            bytes,
            'large.pdf',
            title: 'File',
          ),
          throwsA(isA<TravelCalendarException>()),
        );
      }
      expect(adapter.requests, hasLength(1));
    },
  );

  test('document mutations preserve trip link query access', () async {
    final adapter = _Adapter((_) => _json({'ok': true}));
    final service = TravelCalendarService(
      client: Dio(BaseOptions(baseUrl: 'https://example.test'))
        ..httpClientAdapter = adapter,
    );
    const path = '/calendar/stops/trip-123/documents/doc-123/comments';
    const query = {'stopToken': 'trip-link'};
    await service.post(path, data: {'body': 'Meet here'}, query: query);
    await service.put(
      '$path/comment-123',
      data: {'body': 'Meet there'},
      query: query,
    );
    await service.delete('$path/comment-123', query: query);
    expect(adapter.requests.map((request) => request.method), [
      'POST',
      'PUT',
      'DELETE',
    ]);
    for (final request in adapter.requests) {
      expect(request.queryParameters['stopToken'], 'trip-link');
    }
  });

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
