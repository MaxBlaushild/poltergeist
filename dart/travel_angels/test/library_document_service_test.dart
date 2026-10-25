import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';

class _LibraryAPI extends APIClient {
  _LibraryAPI() : super('https://example.test');

  final calls = <Map<String, dynamic>>[];
  dynamic response;
  Object? failure;

  T _result<T>(String method, String path, dynamic data) {
    calls.add({'method': method, 'path': path, 'data': data});
    if (failure != null) throw failure!;
    return response as T;
  }

  @override
  Future<T> get<T>(String url, {Map<String, dynamic>? params}) async =>
      _result<T>('GET', url, params);

  @override
  Future<T> put<T>(String url, {dynamic data}) async =>
      _result<T>('PUT', url, data);

  @override
  Future<T> post<T>(String url, {dynamic data}) async =>
      _result<T>('POST', url, data);
}

class _LoopbackHttpOverrides extends HttpOverrides {}

Map<String, dynamic> _guide({Map<String, dynamic> fields = const {}}) => {
  'id': 'guide-1',
  'title': 'Lisbon notes',
  'provider': 'google_docs',
  'userId': 'traveler-1',
  'content': '# A favorite neighborhood\nWalk through Alfama at sunset.',
  ...fields,
};

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test(
    'document access denial preserves a valid session; expiry clears it',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'valid-session'});
      final server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      final requests = <String?>[];
      var status = HttpStatus.forbidden;
      final subscription = server.listen((request) async {
        requests.add(request.headers.value(HttpHeaders.authorizationHeader));
        request.response.statusCode = status;
        request.response.headers.contentType = ContentType.json;
        request.response.write(jsonEncode({'error': 'access denied'}));
        await request.response.close();
      });
      try {
        await HttpOverrides.runZoned(() async {
          var authErrors = 0;
          final api = APIClient(
            'http://127.0.0.1:${server.port}',
            onAuthError: () => authErrors++,
          );
          final service = DocumentService(api);
          final prefs = await SharedPreferences.getInstance();
          await expectLater(
            service.getDocument('revoked-guide'),
            throwsA(
              isA<DioException>().having(
                (error) => error.response?.statusCode,
                'status',
                HttpStatus.forbidden,
              ),
            ),
          );
          expect(prefs.getString('token'), 'valid-session');
          expect(authErrors, 0);

          status = HttpStatus.unauthorized;
          await expectLater(
            service.getDocument('revoked-guide'),
            throwsA(
              isA<DioException>().having(
                (error) => error.response?.statusCode,
                'status',
                HttpStatus.unauthorized,
              ),
            ),
          );
          expect(prefs.getString('token'), isNull);
          expect(authErrors, 1);
          expect(requests, ['Bearer valid-session', 'Bearer valid-session']);
        }, createHttpClient: _LoopbackHttpOverrides().createHttpClient);
      } finally {
        await subscription.cancel();
        await server.close(force: true);
      }
    },
  );

  test(
    'legacy documents decode without implying public access or saved state',
    () {
      final document = Document.fromJson(_guide());
      expect(document.visibility, 'private');
      expect(document.isSaved, isFalse);
      expect(document.sharedWithUserIds, isEmpty);
      expect(document.savedTripIds, isEmpty);
      expect(document.format, isEmpty);
      expect(document.destination, isEmpty);
      expect(document.personalNote, isEmpty);
      expect(document.provider, CloudDocumentProvider.googleDocs);
      expect(document.content, contains('Alfama'));

      final futureProvider = Document.fromJson(
        _guide(fields: {'provider': 'future_source', 'visibility': null}),
      );
      expect(futureProvider.provider, CloudDocumentProvider.unknown);
      expect(futureProvider.visibility, 'private');
    },
  );

  test(
    'reader metadata survives serialization without merging sharing and trip audiences',
    () {
      final document = Document.fromJson(
        _guide(
          fields: {
            'visibility': 'selected',
            'format': 'guide',
            'destination': 'Lisbon, Portugal',
            'personalNote': 'A lovely slow afternoon.',
            'isSaved': true,
            'sharedWithUserIds': ['friend-1'],
            'savedTripIds': ['trip-1', 'trip-2'],
            'user': {'id': 'traveler-1', 'name': 'Maya Chen'},
          },
        ),
      );
      final restored = Document.fromJson(
        jsonDecode(jsonEncode(document.toJson())) as Map<String, dynamic>,
      );
      expect(restored.visibility, 'selected');
      expect(restored.sharedWithUserIds, ['friend-1']);
      expect(restored.savedTripIds, ['trip-1', 'trip-2']);
      expect(restored.isSaved, isTrue);
      expect(restored.destination, 'Lisbon, Portugal');
      expect(restored.user?.name, 'Maya Chen');
      expect(restored.content, document.content);
      expect(restored.personalNote, document.personalNote);
    },
  );

  test(
    'library scopes preserve server access decisions and revoked reads propagate',
    () async {
      final api = _LibraryAPI();
      final service = DocumentService(api);
      api.response = [
        _guide(fields: {'visibility': 'selected', 'isSaved': true}),
      ];
      for (final scope in ['network', 'saved', 'yours']) {
        final documents = await service.getLibraryDocuments(scope);
        expect(api.calls.last['data'], {'scope': scope});
        expect(documents.single['visibility'], 'selected');
        expect(documents.single['isSaved'], isTrue);
      }
      final request = RequestOptions(path: '/travel-angels/documents/guide-1');
      final denied = DioException(
        requestOptions: request,
        response: Response(
          requestOptions: request,
          statusCode: 404,
          data: {'error': 'document not found'},
        ),
        type: DioExceptionType.badResponse,
      );
      api.failure = denied;
      final priorCalls = api.calls.length;
      await expectLater(service.getDocument('guide-1'), throwsA(same(denied)));
      expect(api.calls, hasLength(priorCalls + 1));
      expect(api.calls.last['path'], '/travel-angels/documents/guide-1');
    },
  );

  test(
    'unsaving, narrowing sharing and removing a trip retain explicit false and empty values',
    () async {
      final api = _LibraryAPI()..response = _guide();
      final service = DocumentService(api);
      await service.setDocumentSaved('guide-1', false);
      expect(api.calls.last, {
        'method': 'PUT',
        'path': '/travel-angels/documents/guide-1/saved',
        'data': {'saved': false},
      });
      await service.updateDocumentSharing('guide-1', visibility: 'private');
      expect(api.calls.last['data'], {'visibility': 'private', 'userIds': []});
      await service.setDocumentTrip('guide-1', 'trip-1', false);
      expect(api.calls.last, {
        'method': 'PUT',
        'path': '/travel-angels/documents/guide-1/trips/trip-1',
        'data': {'saved': false},
      });
    },
  );

  test(
    'editing can clear notes and locations without silently changing sharing',
    () async {
      final api = _LibraryAPI()..response = _guide();
      final service = DocumentService(api);
      await service.updateDocument(
        documentId: 'guide-1',
        title: 'An updated guide',
        content: '',
        personalNote: '',
        destination: '',
        locations: [],
      );
      final payload = api.calls.single['data'] as Map<String, dynamic>;
      expect(payload['content'], '');
      expect(payload['personalNote'], '');
      expect(payload['destination'], '');
      expect(payload['locations'], isEmpty);
      expect(payload, isNot(contains('visibility')));
      expect(payload, isNot(contains('userIds')));
      expect(payload, isNot(contains('savedTripIds')));
    },
  );
}
