import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/google_drive_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/permissions_panel.dart';

import 'support/app_fonts.dart';

class _ConnectionAuth extends AuthProvider {
  _ConnectionAuth() : super(AuthService(APIClient('https://example.test')));
  User? current = User(id: 'first-user', name: 'Maya');
  @override
  User? get user => current;
  @override
  Future<void> verifyToken() async {}
  void change(User? user) {
    current = user;
    notifyListeners();
  }
}

class _Drive extends GoogleDriveService {
  _Drive() : super(APIClient('https://example.test'));

  bool connected = false;
  int statusCalls = 0;
  int authCalls = 0;
  int revokeCalls = 0;
  Future<Map<String, dynamic>> Function()? status;
  Completer<Map<String, dynamic>>? authorization;
  Completer<Map<String, dynamic>>? revocation;

  @override
  Future<Map<String, dynamic>> getStatus() async {
    statusCalls++;
    return status == null ? {'connected': connected} : status!();
  }

  @override
  Future<Map<String, dynamic>> getAuthUrl() async {
    authCalls++;
    return authorization?.future ??
        {'authUrl': 'https://accounts.google.com/authorize'};
  }

  @override
  Future<Map<String, dynamic>> revoke() async {
    revokeCalls++;
    if (revocation != null) return revocation!.future;
    connected = false;
    return {};
  }
}

Finder _key(String key) => find.byKey(ValueKey(key));

Future<void> _mount(
  WidgetTester tester,
  _Drive service, {
  Stream<Uri>? stream,
  Future<bool> Function(Uri)? launch,
  double width = 800,
  double textScale = 1,
  AuthProvider? auth,
}) async {
  final size = Size(width, 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final app = MaterialApp(
    theme: AppTheme.lightTheme,
    builder: (_, child) => MediaQuery(
      data: MediaQueryData(
        size: size,
        textScaler: TextScaler.linear(textScale),
      ),
      child: child!,
    ),
    home: Scaffold(
      appBar: AppBar(title: const Text('Connected accounts')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: PermissionsPanel(
          service: service,
          deepLinkStream: stream,
          listenForDeepLinks: stream != null,
          launchAuthorization: launch ?? (_) async => true,
        ),
      ),
    ),
  );
  await tester.pumpWidget(
    auth == null
        ? app
        : ChangeNotifierProvider<AuthProvider>.value(value: auth, child: app),
  );
  await tester.pump();
}

Future<void> _tapToggle(WidgetTester tester) async {
  await tester.tap(_key('drive-connection-toggle'));
  await tester.pump();
}

Future<void> _resume(WidgetTester tester) async {
  tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
  tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
  await tester.pump();
}

void main() {
  setUpAll(loadAppFonts);

  testWidgets(
    'account changes discard old connection status and logout stops fetching',
    (tester) async {
      final auth = _ConnectionAuth();
      final old = Completer<Map<String, dynamic>>();
      final service = _Drive()..status = () => old.future;
      await _mount(tester, service, auth: auth);
      service.status = () async => {'connected': false};
      auth.change(User(id: 'second-user', name: 'Nora'));
      await tester.pumpAndSettle();
      expect(service.statusCalls, 2);
      expect(find.text('Not connected'), findsOneWidget);
      old.complete({'connected': true});
      await tester.pumpAndSettle();
      expect(find.text('Not connected'), findsOneWidget);
      auth.change(null);
      await tester.pumpAndSettle();
      expect(
        find.text('Sign in to manage connected accounts.'),
        findsOneWidget,
      );
      expect(
        tester.widget<Switch>(_key('drive-connection-toggle')).onChanged,
        isNull,
      );
      await _resume(tester);
      await tester.pumpAndSettle();
      expect(service.statusCalls, 2);
      await tester.pumpWidget(const SizedBox());
      auth.dispose();
    },
  );

  testWidgets(
    'logout during authorization prevents opening a previous account connection',
    (tester) async {
      final auth = _ConnectionAuth();
      final service = _Drive()
        ..authorization = Completer<Map<String, dynamic>>();
      var launched = false;
      await _mount(
        tester,
        service,
        auth: auth,
        launch: (_) async {
          launched = true;
          return true;
        },
      );
      await tester.pumpAndSettle();
      await _tapToggle(tester);
      auth.change(null);
      service.authorization!.complete({
        'authUrl': 'https://accounts.google.com/authorize',
      });
      await tester.pumpAndSettle();
      expect(launched, isFalse);
      expect(
        find.text('Sign in to manage connected accounts.'),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox());
      auth.dispose();
    },
  );

  for (final fails in [false, true]) {
    testWidgets(
      'leaving during a status ${fails ? 'failure' : 'response'} is safe',
      (tester) async {
        final completion = Completer<Map<String, dynamic>>();
        final service = _Drive()..status = () => completion.future;
        await _mount(tester, service);
        expect(find.text('Checking connection…'), findsOneWidget);
        expect(
          tester.widget<Switch>(_key('drive-connection-toggle')).onChanged,
          isNull,
        );
        await tester.pumpWidget(const SizedBox());
        if (fails) {
          completion.completeError(StateError('offline'));
        } else {
          completion.complete({'connected': true});
        }
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
      },
    );

    testWidgets(
      'leaving during authorization ${fails ? 'failure' : 'response'} does not launch or update the page',
      (tester) async {
        final service = _Drive()
          ..authorization = Completer<Map<String, dynamic>>();
        var launched = false;
        await _mount(
          tester,
          service,
          launch: (_) async {
            launched = true;
            return true;
          },
        );
        await tester.pumpAndSettle();
        await _tapToggle(tester);
        expect(service.authCalls, 1);
        await tester.pumpWidget(const SizedBox());
        if (fails) {
          service.authorization!.completeError(StateError('offline'));
        } else {
          service.authorization!.complete({
            'authUrl': 'https://accounts.google.com/authorize',
          });
        }
        await tester.pumpAndSettle();
        expect(launched, isFalse);
        expect(tester.takeException(), isNull);
      },
    );

    testWidgets(
      'leaving during disconnect ${fails ? 'failure' : 'response'} is safe',
      (tester) async {
        final service = _Drive()
          ..connected = true
          ..revocation = Completer<Map<String, dynamic>>();
        await _mount(tester, service);
        await tester.pumpAndSettle();
        await _tapToggle(tester);
        expect(service.revokeCalls, 1);
        await tester.pumpWidget(const SizedBox());
        if (fails) {
          service.revocation!.completeError(StateError('offline'));
        } else {
          service.revocation!.complete({});
        }
        await tester.pumpAndSettle();
        expect(service.statusCalls, 1);
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets(
    'status failures stay inline, expose Retry and preserve an unknown connection state',
    (tester) async {
      final service = _Drive()
        ..status = () async => throw StateError('internal secret details');
      await _mount(tester, service, width: 320, textScale: 2);
      await tester.pumpAndSettle();
      expect(find.text('Connection status unavailable'), findsOneWidget);
      expect(_key('drive-connection-error'), findsOneWidget);
      expect(find.byType(SnackBar), findsNothing);
      expect(find.textContaining('internal secret details'), findsNothing);
      expect(
        tester.widget<Switch>(_key('drive-connection-toggle')).onChanged,
        isNull,
      );
      service.status = null;
      await tester.ensureVisible(_key('drive-connection-retry'));
      await tester.pumpAndSettle();
      await tester.tap(_key('drive-connection-retry'));
      await tester.pumpAndSettle();
      expect(find.text('Not connected'), findsOneWidget);
      expect(_key('drive-connection-error'), findsNothing);
      expect(_key('drive-connection-retry'), findsNothing);
      expect(find.text('Refresh connection'), findsNothing);
      expect(
        tester.widget<Switch>(_key('drive-connection-toggle')).onChanged,
        isNotNull,
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'OAuth callback and browser return check the server instead of claiming success',
    (tester) async {
      final service = _Drive();
      final links = StreamController<Uri>();
      final launched = <Uri>[];
      await _mount(
        tester,
        service,
        stream: links.stream,
        launch: (uri) async {
          launched.add(uri);
          return true;
        },
      );
      await tester.pumpAndSettle();
      await _tapToggle(tester);
      await tester.pumpAndSettle();
      expect(launched.single.host, 'accounts.google.com');
      expect(
        find.textContaining('Complete authorization in your browser'),
        findsOneWidget,
      );
      expect(
        tester.widget<Switch>(_key('drive-connection-toggle')).value,
        isFalse,
      );
      links.add(
        Uri.parse(
          'travelangels://oauth-callback?success=true&service=google-drive',
        ),
      );
      await tester.pumpAndSettle();
      expect(service.statusCalls, 2);
      expect(find.text('Not connected'), findsOneWidget);
      expect(find.byType(SnackBar), findsNothing);
      service.connected = true;
      await _resume(tester);
      await tester.pumpAndSettle();
      expect(find.text('Connected'), findsOneWidget);
      expect(
        find.textContaining('Complete authorization in your browser'),
        findsNothing,
      );
      await _tapToggle(tester);
      await tester.pumpAndSettle();
      expect(service.revokeCalls, 1);
      expect(find.text('Not connected'), findsOneWidget);
      await tester.pumpWidget(const SizedBox());
      expect(links.hasListener, isFalse);
      unawaited(links.close());
      await tester.pump();
    },
  );

  testWidgets('a later resume check wins over an older status response', (
    tester,
  ) async {
    final old = Completer<Map<String, dynamic>>();
    final service = _Drive()..status = () => old.future;
    await _mount(tester, service);
    service.status = () async => {'connected': false};
    await _resume(tester);
    await tester.pumpAndSettle();
    expect(find.text('Not connected'), findsOneWidget);
    old.complete({'connected': true});
    await tester.pumpAndSettle();
    expect(find.text('Not connected'), findsOneWidget);
    expect(
      tester.widget<Switch>(_key('drive-connection-toggle')).value,
      isFalse,
    );
  });

  testWidgets('failed launch and failed disconnect retain confirmed status', (
    tester,
  ) async {
    final service = _Drive();
    await _mount(tester, service, launch: (_) async => false);
    await tester.pumpAndSettle();
    await _tapToggle(tester);
    await tester.pumpAndSettle();
    expect(find.text('Not connected'), findsOneWidget);
    expect(
      find.text(
        'Google Drive authorization could not be opened. Please try again.',
      ),
      findsOneWidget,
    );
    expect(
      find.textContaining('Complete authorization in your browser'),
      findsNothing,
    );
    service.connected = true;
    await _resume(tester);
    await tester.pumpAndSettle();
    service.revocation = Completer<Map<String, dynamic>>();
    await _tapToggle(tester);
    service.revocation!.completeError(StateError('offline'));
    await tester.pumpAndSettle();
    expect(find.text('Connected'), findsOneWidget);
    expect(
      find.text('Google Drive could not be disconnected. Please try again.'),
      findsOneWidget,
    );
    expect(
      tester.widget<Switch>(_key('drive-connection-toggle')).value,
      isTrue,
    );
    expect(tester.takeException(), isNull);
  });
}
