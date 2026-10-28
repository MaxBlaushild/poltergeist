import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/login_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_account_claim.dart';

const _demoPhone = '+1 (202) 555-0101';
const _ordinaryPhone = '+1 (415) 555-2671';
const _nationalDemoPhone = '2025550101';
const _nationalOrdinaryPhone = '4155552671';
const _demoMessage =
    'Test account: enter your configured 6-digit test code. No text message will be sent.';
const _smsMessage =
    "We've just sent a 6-digit verification code. It may take a moment to arrive.";

class _AuthAPI extends APIClient {
  _AuthAPI() : super('https://example.test');

  final posts = <({String path, dynamic data})>[];
  Completer<dynamic>? pendingCodeRequest;

  @override
  Future<T> get<T>(String url, {Map<String, dynamic>? params}) async {
    throw StateError('Unexpected GET request: $url');
  }

  @override
  Future<T> post<T>(String url, {dynamic data}) async {
    posts.add((path: url, data: data));
    if (url == ApiConstants.verificationCodeEndpoint) {
      return (pendingCodeRequest == null
              ? <String, dynamic>{}
              : await pendingCodeRequest!.future)
          as T;
    }
    if (url == ApiConstants.loginEndpoint) {
      // No code is accepted by this fake, and no real account code is stored
      // in these tests. Authentication still has to reach the server boundary.
      final request = RequestOptions(path: url);
      throw DioException(
        requestOptions: request,
        type: DioExceptionType.badResponse,
        message: 'Invalid verification code',
        response: Response<Map<String, dynamic>>(
          requestOptions: request,
          statusCode: 401,
          data: {'error': 'Invalid verification code'},
        ),
      );
    }
    throw StateError('Unexpected POST request: $url');
  }
}

class _UnclaimedTravelService extends TravelCalendarService {
  int claimChecks = 0;

  @override
  Future<bool> guestNeedsClaim(String accountId) async {
    claimChecks++;
    throw StateError('An unauthenticated guest cannot claim an invitation');
  }
}

Finder _field(String label) => find.byWidgetPredicate(
  (widget) => widget is TextField && widget.decoration?.labelText == label,
);

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _pumpAuth(
  WidgetTester tester,
  AuthProvider auth, {
  _UnclaimedTravelService? claimService,
}) async {
  tester.view.physicalSize = const Size(390, 844);
  tester.view.devicePixelRatio = 1;
  tester.platformDispatcher.localesTestValue = const [Locale('en', 'US')];
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  addTearDown(tester.platformDispatcher.clearLocalesTestValue);
  addTearDown(auth.dispose);
  await tester.pumpWidget(
    ChangeNotifierProvider<AuthProvider>.value(
      value: auth,
      child: MaterialApp(
        theme: AppTheme.lightTheme,
        home: claimService == null
            ? const LoginScreen()
            : Scaffold(
                body: Builder(
                  builder: (context) => Center(
                    child: TextButton(
                      onPressed: () =>
                          claimTravelAccount(context, claimService),
                      child: const Text('Join shared trip'),
                    ),
                  ),
                ),
              ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  if (claimService != null) {
    await _tap(tester, find.text('Join shared trip'));
    expect(find.text('Claim your account'), findsOneWidget);
  }
}

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  for (final nationalInput in [false, true]) {
    testWidgets(
      'login demo skips SMS, rejects a wrong code, and resets to ordinary SMS '
      '(national input: $nationalInput)',
      (tester) async {
        final api = _AuthAPI();
        final auth = AuthProvider(AuthService(api));
        await _pumpAuth(tester, auth);
        expect(find.text('US +1'), findsOneWidget);

        await tester.enterText(
          _field('Phone Number'),
          nationalInput ? _nationalDemoPhone : _demoPhone,
        );
        await tester.pump();
        expect(find.text('Continue with test account'), findsOneWidget);
        expect(find.text('Get code'), findsNothing);
        await _tap(tester, find.text('Continue with test account'));

        expect(api.posts, isEmpty);
        expect(auth.verificationPhoneNumber, '+12025550101');
        expect(auth.isTestVerification, isTrue);
        expect(auth.isRegister, isFalse);
        expect(auth.isAuthenticated, isFalse);
        expect(find.text(_demoMessage), findsOneWidget);
        expect(find.text(_smsMessage), findsNothing);
        expect(
          tester.widget<TextField>(_field('Phone Number')).enabled,
          isFalse,
        );
        expect(
          tester.widget<TextField>(_field('Test code')).autofillHints,
          isNull,
        );

        await tester.enterText(_field('Test code'), '000000');
        await _tap(tester, find.text('Login'));
        expect(api.posts, hasLength(1));
        expect(api.posts.single.path, ApiConstants.loginEndpoint);
        expect(api.posts.single.data, {
          'phoneNumber': '+12025550101',
          'code': '000000',
        });
        expect(auth.isAuthenticated, isFalse);
        expect(
          (await SharedPreferences.getInstance()).getString('token'),
          isNull,
        );
        expect(find.text('Invalid verification code'), findsOneWidget);
        expect(find.text(_demoMessage), findsOneWidget);

        await _tap(tester, find.text('Change phone number'));
        expect(auth.isTestVerification, isFalse);
        expect(auth.verificationPhoneNumber, isNull);
        expect(find.text(_demoMessage), findsNothing);
        expect(find.text('Invalid verification code'), findsNothing);
        expect(
          tester.widget<TextField>(_field('Phone Number')).enabled,
          isTrue,
        );

        await tester.enterText(
          _field('Phone Number'),
          nationalInput ? _nationalOrdinaryPhone : _ordinaryPhone,
        );
        await tester.pump();
        expect(find.text('Continue with test account'), findsNothing);
        await _tap(tester, find.text('Get code'));
        expect(api.posts.last.path, ApiConstants.verificationCodeEndpoint);
        expect(api.posts.last.data, {
          'phoneNumber': '+14155552671',
          'appName': ApiConstants.appName,
        });
        expect(auth.verificationPhoneNumber, '+14155552671');
        expect(auth.isTestVerification, isFalse);
        expect(find.text(_smsMessage), findsOneWidget);
        expect(find.text(_demoMessage), findsNothing);
        expect(_field('Test code'), findsNothing);
        expect(
          tester.widget<TextField>(_field('Verification code')).autofillHints,
          contains(AutofillHints.oneTimeCode),
        );
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('login rejects short or malformed national phone input', (
    tester,
  ) async {
    final api = _AuthAPI();
    final auth = AuthProvider(AuthService(api));
    await _pumpAuth(tester, auth);

    for (final phone in ['202', 'abc', '202abc0101']) {
      await tester.enterText(_field('Phone Number'), phone);
      await tester.pump();
      expect(find.text('Get code'), findsNothing);
      expect(find.text('Continue with test account'), findsNothing);
      expect(auth.isWaitingForVerificationCode, isFalse);
    }
    expect(api.posts, isEmpty);
    expect(find.text('US +1'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('login locks the phone and guards a pending duplicate request', (
    tester,
  ) async {
    final api = _AuthAPI()..pendingCodeRequest = Completer<dynamic>();
    final auth = AuthProvider(AuthService(api));
    await _pumpAuth(tester, auth);
    await tester.enterText(_field('Phone Number'), _ordinaryPhone);
    await tester.pump();

    // Invoke the same callback twice before a rebuild, as rapid taps can do.
    final requestCode = tester
        .widget<ElevatedButton>(find.widgetWithText(ElevatedButton, 'Get code'))
        .onPressed!;
    requestCode();
    requestCode();
    await tester.pump();
    expect(api.posts, hasLength(1));
    expect(tester.widget<TextField>(_field('Phone Number')).enabled, isFalse);
    expect(
      tester
          .widget<ElevatedButton>(
            find.widgetWithText(ElevatedButton, 'Please wait…'),
          )
          .onPressed,
      isNull,
    );

    api.pendingCodeRequest!.complete(<String, dynamic>{});
    await tester.pumpAndSettle();
    expect(api.posts, hasLength(1));
    expect(find.text(_smsMessage), findsOneWidget);
    expect(tester.widget<TextField>(_field('Phone Number')).enabled, isFalse);
    await _tap(tester, find.text('Change phone number'));
    expect(tester.widget<TextField>(_field('Phone Number')).enabled, isTrue);
    expect(auth.verificationPhoneNumber, isNull);
    expect(tester.takeException(), isNull);
  });

  for (final nationalInput in [false, true]) {
    testWidgets(
      'shared trip claim demo skips SMS and cannot claim after rejected login '
      '(national input: $nationalInput)',
      (tester) async {
        final api = _AuthAPI();
        final auth = AuthProvider(AuthService(api));
        final service = _UnclaimedTravelService();
        await _pumpAuth(tester, auth, claimService: service);
        expect(find.text('US +1'), findsOneWidget);
        await tester.enterText(
          _field('Phone number'),
          nationalInput ? _nationalDemoPhone : _demoPhone,
        );
        await tester.pump();
        await _tap(tester, find.text('Continue with test account'));

        expect(api.posts, isEmpty);
        expect(auth.isTestVerification, isTrue);
        expect(auth.verificationPhoneNumber, '+12025550101');
        expect(find.text(_demoMessage), findsOneWidget);
        expect(_field('Test code'), findsOneWidget);
        expect(_field('Verification code'), findsNothing);
        expect(
          tester.widget<TextField>(_field('Phone number')).enabled,
          isFalse,
        );
        expect(find.text('Change number'), findsOneWidget);
        expect(find.text('Change number or resend code'), findsNothing);

        await tester.enterText(_field('Test code'), '000000');
        await _tap(tester, find.text('Verify account'));
        expect(api.posts.single.path, ApiConstants.loginEndpoint);
        expect(api.posts.single.data, {
          'phoneNumber': '+12025550101',
          'code': '000000',
        });
        expect(auth.isAuthenticated, isFalse);
        expect(service.claimChecks, 0);
        expect(find.text('Claim your account'), findsOneWidget);
        expect(
          find.textContaining('Invalid verification code'),
          findsOneWidget,
        );

        await _tap(tester, find.text('Change number'));
        expect(auth.isTestVerification, isFalse);
        expect(auth.verificationPhoneNumber, isNull);
        expect(find.text(_demoMessage), findsNothing);
        expect(_field('Test code'), findsNothing);
        await tester.enterText(
          _field('Phone number'),
          nationalInput ? _nationalOrdinaryPhone : _ordinaryPhone,
        );
        await tester.pump();
        expect(find.text('Continue with test account'), findsNothing);
        await _tap(tester, find.text('Send code'));
        expect(api.posts.last.path, ApiConstants.verificationCodeEndpoint);
        expect(api.posts.last.data, {
          'phoneNumber': '+14155552671',
          'appName': ApiConstants.appName,
        });
        expect(auth.verificationPhoneNumber, '+14155552671');
        expect(auth.isTestVerification, isFalse);
        expect(_field('Verification code'), findsOneWidget);
        expect(find.text(_demoMessage), findsNothing);
        expect(find.text('Change number or resend code'), findsOneWidget);
        expect(service.claimChecks, 0);

        await _tap(tester, find.text('Keep browsing'));
        expect(find.text('Claim your account'), findsNothing);
        expect(auth.isWaitingForVerificationCode, isFalse);
        expect(auth.verificationPhoneNumber, isNull);
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('shared trip claim rejects short or malformed national input', (
    tester,
  ) async {
    final api = _AuthAPI();
    final auth = AuthProvider(AuthService(api));
    final service = _UnclaimedTravelService();
    await _pumpAuth(tester, auth, claimService: service);
    for (final phone in ['202', 'abc', '202abc0101']) {
      await tester.enterText(_field('Phone number'), phone);
      await tester.pump();
      expect(find.text('Continue with test account'), findsNothing);
      await _tap(tester, find.text('Send code'));
      expect(
        find.text('Enter a valid phone number for the selected country.'),
        findsOneWidget,
      );
    }

    expect(api.posts, isEmpty);
    expect(auth.isAuthenticated, isFalse);
    expect(auth.isWaitingForVerificationCode, isFalse);
    expect(service.claimChecks, 0);
    expect(_field('Verification code'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('system back from claim clears the demo destination for login', (
    tester,
  ) async {
    final api = _AuthAPI();
    final auth = AuthProvider(AuthService(api));
    await _pumpAuth(tester, auth, claimService: _UnclaimedTravelService());
    await tester.enterText(_field('Phone number'), _demoPhone);
    await tester.pump();
    await _tap(tester, find.text('Continue with test account'));
    expect(auth.isTestVerification, isTrue);

    final navigator = Navigator.of(tester.element(find.byType(AlertDialog)));
    expect(await navigator.maybePop(), isTrue);
    await tester.pumpAndSettle();
    expect(find.text('Claim your account'), findsNothing);
    expect(auth.isWaitingForVerificationCode, isFalse);
    expect(auth.isTestVerification, isFalse);
    expect(auth.verificationPhoneNumber, isNull);

    await tester.pumpWidget(
      ChangeNotifierProvider<AuthProvider>.value(
        value: auth,
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const LoginScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.widget<TextField>(_field('Phone Number')).enabled, isTrue);
    expect(_field('Test code'), findsNothing);
    expect(find.text(_demoMessage), findsNothing);
    expect(find.text(_smsMessage), findsNothing);
    expect(api.posts, isEmpty);
    expect(tester.takeException(), isNull);
  });
}
