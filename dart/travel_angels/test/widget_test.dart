import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'support/app_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/main.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/login_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/main_navbar.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

class _PreviewAuth extends AuthProvider {
  _PreviewAuth() : super(AuthService(APIClient('https://example.test')));
  @override
  Future<void> verifyToken() async {}
  @override
  bool get loading => false;
  @override
  bool get isWaitingForVerificationCode => true;
}

Future<void> _capture(WidgetTester tester, GlobalKey key, String name) async {
  if (!const bool.fromEnvironment('TRIP_SCREENSHOTS')) return;
  await tester.runAsync(() async {
    final image =
        await (key.currentContext!.findRenderObject()! as RenderRepaintBoundary)
            .toImage();
    final data = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-$name.png',
    ).writeAsBytes(data!.buffer.asUint8List());
    image.dispose();
  });
}

class _SharedCalendarService extends TravelCalendarService {
  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async => {
    'calendar': {
      'id': 'calendar-1',
      'title': 'Our travel calendar',
      'canManage': false,
    },
    'stops': [
      {
        'id': 'stop-1',
        'title': 'Meet us in Lisbon',
        'destination': 'Lisbon, Portugal',
        'startDate': travelDateValue(
          DateTime(DateTime.now().year, DateTime.now().month, 12),
        ),
        'endDate': travelDateValue(
          DateTime(DateTime.now().year, DateTime.now().month, 18),
        ),
        'datePrecision': 'fixed',
        'status': 'confirmed',
        'published': true,
        'joinOpen': true,
      },
    ],
  };
}

void main() {
  setUpAll(loadAppFonts);
  testWidgets('shared links open permitted plans without the app login gate', (
    tester,
  ) async {
    tester.view.devicePixelRatio = 1;
    tester.view.physicalSize = const Size(1200, 1000);
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.binding.setSurfaceSize(const Size(1200, 1000));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    SharedPreferences.setMockInitialValues({});
    final boundary = GlobalKey();
    await tester.pumpWidget(
      RepaintBoundary(
        key: boundary,
        child: MyApp(
          initialRoute: '/calendar/example_shared_token_123456789',
          calendarService: _SharedCalendarService(),
          listenForDeepLinks: false,
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.byType(TravelCalendarMonth), findsOneWidget);
    expect(
      tester
          .widget<TravelCalendarMonth>(find.byType(TravelCalendarMonth))
          .stops
          .single['title'],
      'Meet us in Lisbon',
    );
    expect(find.textContaining('Meet us in Lisbon'), findsWidgets);
    expect(find.text('Our travel calendar'), findsOneWidget);
    expect(find.text('Phone number'), findsNothing);
    expect(tester.takeException(), isNull);
    await _capture(tester, boundary, 'calendar-1200');
    tester.view.physicalSize = const Size(390, 844);
    await tester.binding.setSurfaceSize(const Size(390, 844));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await _capture(tester, boundary, 'calendar-390');
  });

  testWidgets('sign-in and navigation fit narrow screens with enlarged text', (
    tester,
  ) async {
    addTearDown(() => tester.binding.setSurfaceSize(null));
    final auth = _PreviewAuth();
    addTearDown(auth.dispose);
    for (final size in [
      const Size(1200, 900),
      const Size(390, 844),
      const Size(320, 740),
    ]) {
      await tester.binding.setSurfaceSize(size);
      final scale = size.width == 320 ? 2.0 : 1.0;
      final boundary = GlobalKey();
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            builder: (_, child) => MediaQuery(
              data: MediaQueryData(
                size: size,
                textScaler: TextScaler.linear(scale),
              ),
              child: child!,
            ),
            home: RepaintBoundary(key: boundary, child: const LoginScreen()),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Sign in or sign up'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await _capture(tester, boundary, 'login-${size.width.toInt()}');

      var destination = 0;
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          builder: (_, child) => MediaQuery(
            data: MediaQueryData(
              size: size,
              textScaler: TextScaler.linear(scale),
            ),
            child: child!,
          ),
          home: StatefulBuilder(
            builder: (context, setState) => Scaffold(
              appBar: size.width > 600
                  ? MainNavbarAppBar(
                      currentIndex: destination,
                      onDestinationChanged: (next) =>
                          setState(() => destination = next),
                    )
                  : null,
              bottomNavigationBar: size.width <= 600
                  ? MainNavbar(
                      currentIndex: destination,
                      onDestinationChanged: (next) =>
                          setState(() => destination = next),
                    )
                  : null,
              body: const SizedBox.expand(),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Library'));
      await tester.pumpAndSettle();
      expect(destination, NavDestination.library.index);
      expect(tester.takeException(), isNull);
    }
  });
}
