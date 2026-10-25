import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/profile_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/user_level_service.dart';
import 'package:travel_angels/theme/app_theme.dart';

import 'support/app_fonts.dart';

class _ProfileAuth extends AuthProvider {
  _ProfileAuth() : super(AuthService(APIClient('https://example.test')));

  @override
  User get user => User(
    id: 'profile-traveler',
    name: 'Isabella Montague',
    username: 'isabella.travels',
    credits: 1250,
  );

  @override
  bool get isAuthenticated => true;

  @override
  bool get loading => false;

  @override
  Future<void> verifyToken() async {}
}

class _ProfileLevel extends UserLevelProvider {
  _ProfileLevel() : super(UserLevelService(APIClient('https://example.test')));

  @override
  UserLevel get userLevel => UserLevel(
    level: 12,
    experiencePointsOnLevel: 4850,
    totalExperiencePoints: 37480,
    experienceToNextLevel: 10000,
  );

  @override
  bool get loading => false;
}

void main() {
  setUpAll(loadAppFonts);

  for (final (width, scale) in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('profile fits ${width.toInt()}px with ${scale}x text', (
      tester,
    ) async {
      final size = Size(width, 844);
      await tester.binding.setSurfaceSize(size);
      addTearDown(() => tester.binding.setSurfaceSize(null));
      SharedPreferences.setMockInitialValues({});
      // Widget tests block outgoing HTTP. The default 400 response exercises
      // the document/Drive fallback without contacting an external service.
      const links = MethodChannel('com.llfbandit.app_links/events');
      tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        links,
        (_) async => null,
      );
      addTearDown(() {
        tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
          links,
          null,
        );
      });
      final auth = _ProfileAuth();
      final level = _ProfileLevel();
      final boundary = GlobalKey();
      addTearDown(auth.dispose);
      addTearDown(level.dispose);
      await tester.pumpWidget(
        MultiProvider(
          providers: [
            ChangeNotifierProvider<AuthProvider>.value(value: auth),
            ChangeNotifierProvider<UserLevelProvider>.value(value: level),
          ],
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            builder: (_, child) => MediaQuery(
              data: MediaQueryData(
                size: size,
                textScaler: TextScaler.linear(scale),
              ),
              child: child!,
            ),
            home: RepaintBoundary(key: boundary, child: const ProfileScreen()),
          ),
        ),
      );
      await tester.pumpAndSettle();
      // Let the blocked-HTTP Drive fallback snackbar finish before checking
      // that controls further down the scroll view can receive a tap.
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();
      expect(find.text('Isabella Montague'), findsOneWidget);
      expect(find.text('4850 / 10000 XP'), findsOneWidget);
      expect(find.text('1250 credits'), findsOneWidget);
      expect(tester.takeException(), isNull);
      if (width == 390 && const bool.fromEnvironment('TRIP_SCREENSHOTS')) {
        await tester.runAsync(() async {
          final image =
              await (boundary.currentContext!.findRenderObject()!
                      as RenderRepaintBoundary)
                  .toImage();
          final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
          await File(
            '/private/tmp/travel-angels-profile-390.png',
          ).writeAsBytes(bytes!.buffer.asUint8List());
          image.dispose();
        });
      }

      await tester.ensureVisible(find.text('Buy Credits'));
      await tester.pumpAndSettle();
      expect(find.text('Buy Credits').hitTestable(), findsOneWidget);
      await tester.ensureVisible(find.text('Log Out'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Log Out'));
      await tester.pumpAndSettle();
      expect(find.text('Are you sure you want to log out?'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      await tester.pumpWidget(const SizedBox.shrink());
    });
  }
}
