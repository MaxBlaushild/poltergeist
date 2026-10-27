import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/login_screen.dart';
import 'package:travel_angels/screens/profile_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/user_level_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/home_widget.dart';
import 'package:travel_angels/widgets/main_scaffold.dart';

User _user(String id, {int credits = 10}) =>
    User(id: id, name: 'Traveler $id', credits: credits);
UserLevel _level(int value) => UserLevel(
  level: value,
  experiencePointsOnLevel: 10,
  totalExperiencePoints: 100,
  experienceToNextLevel: 200,
);
DioException _failure([int? status]) {
  final request = RequestOptions(path: ApiConstants.whoamiEndpoint);
  return DioException(
    requestOptions: request,
    type: status == null
        ? DioExceptionType.connectionTimeout
        : DioExceptionType.badResponse,
    response: status == null
        ? null
        : Response(requestOptions: request, statusCode: status),
  );
}

class _ControlledAuth extends AuthService {
  _ControlledAuth() : super(APIClient('https://example.test'));
  final verifications = <Completer<User?>>[];
  Completer<AuthResponse>? signingIn;
  Completer<User>? editing;
  @override
  Future<User?> verifyToken() {
    final result = Completer<User?>();
    verifications.add(result);
    return result.future;
  }

  @override
  Future<AuthResponse> login(String phoneNumber, String code) =>
      (signingIn = Completer<AuthResponse>()).future;
  @override
  Future<User> updateProfile({
    String? username,
    String? profilePictureUrl,
    DateTime? dateOfBirth,
    String? gender,
    double? latitude,
    double? longitude,
    String? locationAddress,
    String? bio,
  }) => (editing = Completer<User>()).future;
  @override
  Future<void> logout() async {}
}

class _SessionAPI extends APIClient {
  _SessionAPI() : super('https://example.test');
  Future<dynamic> Function(String)? onGet;
  Future<dynamic> Function(String, dynamic)? onPost;
  final posts = <String>[];
  @override
  Future<T> get<T>(String url, {Map<String, dynamic>? params}) async =>
      await onGet!(url) as T;
  @override
  Future<T> post<T>(String url, {dynamic data}) async {
    posts.add(url);
    return await onPost!(url, data) as T;
  }

  @override
  Future<T> postForSession<T>(
    String url, {
    required String token,
    dynamic data,
  }) => post<T>(url, data: data);
}

class _DelayedAdapter implements HttpClientAdapter {
  final arrived = Completer<RequestOptions>();
  final reply = Completer<ResponseBody>();
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? stream,
    Future<void>? cancelFuture,
  ) {
    arrived.complete(options);
    return reply.future;
  }

  @override
  void close({bool force = false}) {}
}

class _ControlledLevel extends UserLevelService {
  _ControlledLevel({this.immediate = false})
    : super(APIClient('https://example.test'));
  final bool immediate;
  final pending = <Completer<UserLevel>>[];
  int requests = 0;
  @override
  Future<UserLevel> getUserLevel() {
    requests++;
    if (immediate) return Future.value(_level(requests));
    final result = Completer<UserLevel>();
    pending.add(result);
    return result.future;
  }
}

Future<AuthProvider> _signedIn(_ControlledAuth service) async {
  final auth = AuthProvider(service);
  service.verifications.single.complete(_user('a'));
  await Future<void>.delayed(Duration.zero);
  return auth;
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUp(() => SharedPreferences.setMockInitialValues({}));

  test(
    'refresh preserves established identity on transient errors but rejects invalid sessions',
    () async {
      final service = _ControlledAuth();
      final auth = await _signedIn(service);
      addTearDown(auth.dispose);
      final refresh = auth.verifyToken();
      expect(auth.loading, isFalse);
      expect(auth.user?.id, 'a');
      service.verifications.last.completeError(_failure());
      await refresh;
      expect(auth.isAuthenticated, isTrue);
      expect(auth.error, isNotNull);
      final invalid = auth.verifyToken();
      service.verifications.last.complete(null);
      await invalid;
      expect(auth.isAuthenticated, isFalse);
      expect(auth.loading, isFalse);
    },
  );

  test(
    'logout invalidates pending verification and profile edits immediately',
    () async {
      final service = _ControlledAuth();
      final auth = await _signedIn(service);
      addTearDown(auth.dispose);
      final refresh = auth.verifyToken();
      final edit = auth.updateProfile(bio: 'Old account');
      final logout = auth.logout();
      expect(auth.isAuthenticated, isFalse);
      service.verifications.last.complete(_user('a', credits: 99));
      service.editing!.complete(_user('a', credits: 100));
      await Future.wait([refresh, edit, logout]);
      expect(auth.user, isNull);
    },
  );

  test(
    'old account response cannot overwrite newer login or newer refresh',
    () async {
      final service = _ControlledAuth();
      final auth = await _signedIn(service);
      addTearDown(auth.dispose);
      final stale = auth.verifyToken();
      final oldResponse = service.verifications.last;
      final login = auth.logister('+15555555555', '123456');
      service.signingIn!.complete(AuthResponse(user: _user('b'), token: 'b'));
      await Future<void>.delayed(Duration.zero);
      oldResponse.complete(_user('a', credits: 1));
      await stale;
      expect(auth.user?.id, 'b');
      service.verifications.last.complete(_user('b', credits: 20));
      await login;
      final older = auth.verifyToken();
      final olderResponse = service.verifications.last;
      final newer = auth.verifyToken();
      service.verifications.last.complete(_user('b', credits: 30));
      await newer;
      olderResponse.complete(_user('b', credits: 2));
      await older;
      expect(auth.user?.credits, 30);
    },
  );

  test(
    'late sign-in completion after logout does not restore provider identity',
    () async {
      final service = _ControlledAuth();
      final auth = await _signedIn(service);
      addTearDown(auth.dispose);
      final login = auth.logister('+15555555555', '123456');
      await auth.logout();
      service.signingIn!.complete(AuthResponse(user: _user('b'), token: 'b'));
      await login;
      expect(auth.user, isNull);
      expect(service.verifications, hasLength(1));
    },
  );

  test(
    'AuthService keeps tokens on outages and only falls back for missing whoami',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'a'});
      final api = _SessionAPI();
      final service = AuthService(api);
      for (final error in [_failure(), _failure(503)]) {
        api.onGet = (_) async => throw error;
        await expectLater(service.verifyToken(), throwsA(same(error)));
        expect(await service.getToken(), 'a');
        expect(api.posts, isEmpty);
      }
      api.onGet = (_) async => throw _failure(404);
      api.onPost = (url, data) async {
        expect(url, ApiConstants.verifyTokenEndpoint);
        expect(data, {'token': 'a'});
        return {'id': 'a', 'name': 'Traveler'};
      };
      expect((await service.verifyToken())?.id, 'a');
      expect(api.posts, hasLength(1));
      api.onGet = (_) async => throw _failure(401);
      expect(await service.verifyToken(), isNull);
      expect(await service.getToken(), isNull);
      expect(api.posts, hasLength(1));
    },
  );

  test(
    'AuthService late login and verification cannot change newer stored credentials',
    () async {
      final api = _SessionAPI();
      final service = AuthService(api);
      final response = Completer<Map<String, dynamic>>();
      api.onPost = (_, _) => response.future;
      final login = service.login('+15555555555', '123456');
      await service.logout();
      response.complete({
        'user': {'id': 'a'},
        'token': 'old',
      });
      await expectLater(login, throwsStateError);
      expect(await service.getToken(), isNull);
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('token', 'a');
      final verification = Completer<Map<String, dynamic>>();
      api.onGet = (_) => verification.future;
      final refresh = service.verifyToken();
      await Future<void>.delayed(Duration.zero);
      await prefs.setString('token', 'b');
      verification.completeError(_failure(401));
      await expectLater(refresh, throwsA(isA<DioException>()));
      expect(await service.getToken(), 'b');
    },
  );

  test('profile service sends an explicit empty bio for clearing', () async {
    final api = _SessionAPI();
    api.onPost = (_, data) async {
      expect(data, {'bio': ''});
      return {'id': 'a', 'bio': ''};
    };
    expect((await AuthService(api).updateProfile(bio: '')).bio, isNull);
  });

  test(
    'delayed missing whoami cannot verify an old token after sign-in',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'a'});
      final api = _SessionAPI();
      final service = AuthService(api);
      final oldWhoami = Completer<Map<String, dynamic>>();
      api.onGet = (_) => oldWhoami.future;
      api.onPost = (url, data) async {
        expect(url, ApiConstants.loginEndpoint);
        return {
          'user': {'id': 'b'},
          'token': 'b',
        };
      };
      final refresh = service.verifyToken();
      await Future<void>.delayed(Duration.zero);
      await service.login('+15555555555', '123456');
      oldWhoami.completeError(_failure(404));
      await expectLater(refresh, throwsStateError);
      expect(api.posts, [ApiConstants.loginEndpoint]);
      expect(await service.getToken(), 'b');
    },
  );

  test(
    'legacy verification pins its rejected token without clearing a newer login',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'b'});
      final adapter = _DelayedAdapter();
      final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
        ..httpClientAdapter = adapter;
      var invalidations = 0;
      final api = APIClient(
        'https://example.test',
        client: dio,
        onAuthError: () => invalidations++,
      );
      final request = api.postForSession<Map<String, dynamic>>(
        ApiConstants.verifyTokenEndpoint,
        token: 'a',
        data: {'token': 'a'},
      );
      final options = await adapter.arrived.future;
      expect(options.headers['Authorization'], 'Bearer a');
      adapter.reply.complete(
        ResponseBody.fromString(
          '{}',
          401,
          headers: {
            Headers.contentTypeHeader: [Headers.jsonContentType],
          },
        ),
      );
      await expectLater(request, throwsA(isA<DioException>()));
      expect((await SharedPreferences.getInstance()).getString('token'), 'b');
      expect(invalidations, 0);
    },
  );

  for (final changed in [false, true]) {
    test(
      'API session invalidation respects request token (changed=$changed)',
      () async {
        SharedPreferences.setMockInitialValues({'token': 'a'});
        final adapter = _DelayedAdapter();
        final dio = Dio(BaseOptions(baseUrl: 'https://example.test'))
          ..httpClientAdapter = adapter;
        var invalidations = 0;
        final api = APIClient(
          'https://example.test',
          client: dio,
          onAuthError: () => invalidations++,
        );
        final request = api.get<Map<String, dynamic>>('/private');
        final options = await adapter.arrived.future;
        expect(options.headers['Authorization'], 'Bearer a');
        final prefs = await SharedPreferences.getInstance();
        if (changed) await prefs.setString('token', 'b');
        adapter.reply.complete(
          ResponseBody.fromString(
            jsonEncode({'error': 'expired'}),
            401,
            headers: {
              Headers.contentTypeHeader: [Headers.jsonContentType],
            },
          ),
        );
        await expectLater(request, throwsA(isA<DioException>()));
        expect(prefs.getString('token'), changed ? 'b' : null);
        expect(invalidations, changed ? 0 : 1);
      },
    );
  }

  test(
    'clearing user level rejects prior account result and permits next account fetch',
    () async {
      final service = _ControlledLevel();
      final provider = UserLevelProvider(service);
      addTearDown(provider.dispose);
      final old = provider.fetchUserLevel();
      provider.clear();
      final fresh = provider.fetchUserLevel();
      service.pending.last.complete(_level(2));
      await fresh;
      service.pending.first.complete(_level(99));
      await old;
      expect(provider.userLevel?.level, 2);
      expect(provider.loading, isFalse);
    },
  );

  testWidgets(
    'real HomeWidget keeps Profile selected through refresh and reloads levels after relogin',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      const links = MethodChannel('com.llfbandit.app_links/events');
      tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        links,
        (_) async => null,
      );
      addTearDown(
        () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
          links,
          null,
        ),
      );
      final service = _ControlledAuth();
      final auth = AuthProvider(service);
      final levels = _ControlledLevel(immediate: true);
      final level = UserLevelProvider(levels);
      addTearDown(auth.dispose);
      addTearDown(level.dispose);
      service.verifications.single.complete(_user('a'));
      await tester.pumpWidget(
        MultiProvider(
          providers: [
            ChangeNotifierProvider<AuthProvider>.value(value: auth),
            ChangeNotifierProvider<UserLevelProvider>.value(value: level),
          ],
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            home: const HomeWidget(),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Profile').last);
      await tester.pumpAndSettle();
      expect(find.byType(ProfileScreen), findsOneWidget);
      expect(
        tester.widget<NavigationBar>(find.byType(NavigationBar)).selectedIndex,
        3,
      );
      final shell = tester.state(find.byType(MainScaffold));
      final refresh = auth.verifyToken();
      await tester.pump();
      expect(find.byType(ProfileScreen), findsOneWidget);
      expect(tester.state(find.byType(MainScaffold)), same(shell));
      service.verifications.last.complete(_user('a', credits: 50));
      await refresh;
      await tester.pumpAndSettle();
      expect(tester.state(find.byType(MainScaffold)), same(shell));
      expect(
        tester.widget<NavigationBar>(find.byType(NavigationBar)).selectedIndex,
        3,
      );
      final offline = auth.verifyToken();
      service.verifications.last.completeError(_failure());
      await offline;
      await tester.pumpAndSettle();
      expect(find.byType(ProfileScreen), findsOneWidget);
      expect(tester.state(find.byType(MainScaffold)), same(shell));
      await auth.logout();
      await tester.pumpAndSettle();
      expect(find.byType(LoginScreen), findsOneWidget);
      expect(level.userLevel, isNull);
      final login = auth.logister('+15555555555', '123456');
      service.signingIn!.complete(AuthResponse(user: _user('b'), token: 'b'));
      await tester.pump();
      service.verifications.last.complete(_user('b'));
      await login;
      await tester.pumpAndSettle();
      expect(find.byType(MainScaffold), findsOneWidget);
      expect(tester.state(find.byType(MainScaffold)), isNot(same(shell)));
      expect(
        tester.widget<NavigationBar>(find.byType(NavigationBar)).selectedIndex,
        0,
      );
      expect(levels.requests, 2);
      expect(level.userLevel?.level, 2);
      expect(tester.takeException(), isNull);
    },
  );
}
