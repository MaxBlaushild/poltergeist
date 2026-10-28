import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/utils/auth_phone_number.dart';

class _API extends APIClient {
  _API() : super('https://example.test');

  final requests = <({String path, dynamic data})>[];
  Future<dynamic> Function(String, dynamic)? reply;

  @override
  Future<T> post<T>(String url, {dynamic data}) async {
    requests.add((path: url, data: data));
    return await reply?.call(url, data) as T;
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUp(() => SharedPreferences.setMockInitialValues({}));

  test('international input accepts formatting without guessing a country', () {
    const cases = {
      '+1 (415) 555-2671': '+14155552671',
      ' \t+44 20 7946 0958\n': '+442079460958',
      '+81.3.1234.5678': '+81312345678',
      '+14155552671': '+14155552671',
      '+12345678': '+12345678',
      '+123456789012345': '+123456789012345',
    };
    for (final entry in cases.entries) {
      expect(AuthPhoneNumber.normalize(entry.key), entry.value);
    }
    for (final input in [
      '',
      '4155552671',
      '00442079460958',
      '+04155552671',
      '+1234567',
      '+1234567890123456',
      '++14155552671',
      '+1415+5552671',
      '+1 415 555 2671 ext 2',
      '+1/415/555/2671',
      'phone +14155552671',
    ]) {
      expect(AuthPhoneNumber.normalize(input), isNull, reason: input);
    }
  });

  test('both allowlisted accounts are recognized after normalization', () {
    expect(AuthPhoneNumber.isTestAccount('+1 (440) 785-8475'), isTrue);
    expect(AuthPhoneNumber.isTestAccount('+1.202.555.0101'), isTrue);
    expect(AuthPhoneNumber.isTestAccount('+12025550102'), isFalse);
    expect(AuthPhoneNumber.isTestAccount('12025550101'), isFalse);
  });

  for (final phone in ['+1 (440) 785-8475', '+1 (202) 555-0101']) {
    test('test account $phone skips SMS without authenticating', () async {
      final api = _API();
      final auth = AuthService(api);
      expect(await auth.getVerificationCode(phone), isTrue);
      expect(api.requests, isEmpty);
      expect(await auth.getToken(), isNull);
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('token', 'existing-session');
      expect(await auth.getVerificationCode(phone), isTrue);
      expect(await auth.getToken(), 'existing-session');
      expect(api.requests, isEmpty);
    });
  }

  for (final exists in [true, false]) {
    test(
      'ordinary code request reaches API and preserves exists=$exists',
      () async {
        final api = _API()
          ..reply = (_, _) async => exists ? {'id': 'person'} : null;
        final auth = AuthService(api);
        expect(await auth.getVerificationCode('+44 (20) 7946-0958'), exists);
        expect(api.requests.single.path, ApiConstants.verificationCodeEndpoint);
        expect(api.requests.single.data, {
          'phoneNumber': '+442079460958',
          'appName': ApiConstants.appName,
        });
        expect(await auth.getToken(), isNull);
      },
    );
  }

  test(
    'invalid destinations reject every auth operation before API calls',
    () async {
      final api = _API();
      final auth = AuthService(api);
      final invalid = throwsA(
        isA<FormatException>().having(
          (error) => error.message,
          'message',
          contains('country code'),
        ),
      );
      await expectLater(auth.getVerificationCode('4155552671'), invalid);
      await expectLater(auth.login('4155552671', '918273'), invalid);
      await expectLater(auth.register('4155552671', '918273'), invalid);
      expect(api.requests, isEmpty);
      expect(await auth.getToken(), isNull);
    },
  );

  for (final register in [false, true]) {
    test(
      '${register ? 'register' : 'login'} sends canonical phone and caller code to backend',
      () async {
        final api = _API()
          ..reply = (_, _) async => {
            'user': {'id': 'backend-user', 'phoneNumber': '+12025550101'},
            'token': 'backend-issued-token',
          };
        final auth = AuthService(api);
        final result = register
            ? await auth.register('+1 (202) 555-0101', '918273')
            : await auth.login('+1 (202) 555-0101', '918273');
        expect(
          api.requests.single.path,
          register ? ApiConstants.registerEndpoint : ApiConstants.loginEndpoint,
        );
        expect(api.requests.single.data, {
          'phoneNumber': '+12025550101',
          'code': '918273',
        });
        expect(result.user.id, 'backend-user');
        expect(result.token, 'backend-issued-token');
        expect(await auth.getToken(), 'backend-issued-token');
      },
    );
  }

  test(
    'wrong code on a test account is rejected by backend without a token',
    () async {
      final request = RequestOptions(path: ApiConstants.loginEndpoint);
      final rejection = DioException(
        requestOptions: request,
        type: DioExceptionType.badResponse,
        response: Response(
          requestOptions: request,
          statusCode: 401,
          data: {'error': 'Invalid verification code'},
        ),
      );
      final api = _API()..reply = (_, _) async => throw rejection;
      final auth = AuthService(api);
      await expectLater(
        auth.login('+14407858475', '654321'),
        throwsA(same(rejection)),
      );
      expect(api.requests.single.path, ApiConstants.loginEndpoint);
      expect(api.requests.single.data['code'], '654321');
      expect(await auth.getToken(), isNull);
    },
  );

  test(
    'pending test-account login cannot restore a token after logout',
    () async {
      final response = Completer<Map<String, dynamic>>();
      final api = _API()..reply = (_, _) => response.future;
      final auth = AuthService(api);
      final login = auth.login('+12025550101', '918273');
      await auth.logout();
      response.complete({
        'user': {'id': 'backend-user'},
        'token': 'late-token',
      });
      await expectLater(login, throwsStateError);
      expect(await auth.getToken(), isNull);
    },
  );
}
