import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/profile_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/user_level_service.dart';
import 'package:travel_angels/theme/app_theme.dart';

import 'library_test_support.dart';
import 'support/app_fonts.dart';

Finder _key(String value) => find.byKey(ValueKey(value));

class _ProfileAuth extends AuthProvider {
  _ProfileAuth() : super(AuthService(APIClient('https://example.test')));
  User current = User(
    id: 'me',
    name: 'Isabella Montague',
    username: 'isabella.travels',
    credits: 1250,
    bio: 'Slow mornings, good food, and a spreadsheet for every adventure.',
    locationAddress: 'Brooklyn, New York',
    createdAt: DateTime(2025, 6),
  );
  @override
  User get user => current;
  @override
  bool get isAuthenticated => true;
  @override
  bool get loading => false;
  @override
  Future<void> verifyToken() async {}
  void change(User next) {
    current = next;
    notifyListeners();
  }
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

class _RacingCounts extends LibraryFakeDocuments {
  final requests = <Completer<List<Map<String, dynamic>>>>[];
  @override
  Future<List<Map<String, dynamic>>> getLibraryDocuments(String scope) {
    final request = Completer<List<Map<String, dynamic>>>();
    requests.add(request);
    return request.future;
  }
}

Future<GlobalKey> _mount(
  WidgetTester tester,
  _ProfileAuth auth,
  LibraryFakeDocuments documents, {
  double width = 390,
  double scale = 1,
  VoidCallback? onOpenCalendar,
  VoidCallback? onOpenCircle,
}) async {
  final size = Size(width, width >= 600 ? 1000 : 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final level = _ProfileLevel();
  addTearDown(auth.dispose);
  addTearDown(level.dispose);
  final boundary = GlobalKey();
  await tester.pumpWidget(
    MultiProvider(
      providers: [
        ChangeNotifierProvider<AuthProvider>.value(value: auth),
        ChangeNotifierProvider<UserLevelProvider>.value(value: level),
      ],
      child: MaterialApp(
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        builder: (_, child) => MediaQuery(
          data: MediaQueryData(
            size: size,
            textScaler: TextScaler.linear(scale),
          ),
          child: child!,
        ),
        home: RepaintBoundary(
          key: boundary,
          child: ProfileScreen(
            documentService: documents,
            onOpenCalendar: onOpenCalendar,
            onOpenCircle: onOpenCircle,
          ),
        ),
      ),
    ),
  );
  await tester.pump();
  return boundary;
}

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _capture(
  WidgetTester tester,
  GlobalKey boundary,
  String name,
) async {
  if (!const bool.fromEnvironment('TRIP_SCREENSHOTS')) return;
  await tester.runAsync(() async {
    final image =
        await (boundary.currentContext!.findRenderObject()!
                as RenderRepaintBoundary)
            .toImage();
    final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-profile-redesign-$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

LibraryFakeDocuments _documents() => LibraryFakeDocuments([
  libraryGuide('mine', 'Summer in Italy', owner: 'me'),
  libraryGuide('lisbon', 'A slow week in Lisbon', saved: true),
  libraryGuide('budget', 'Portugal budget', format: 'spreadsheet', saved: true),
]);

void main() {
  setUpAll(loadAppFonts);
  for (final (width, scale) in [(1200.0, 1.0), (390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('profile fits ${width.toInt()}px with ${scale}x text', (
      tester,
    ) async {
      final auth = _ProfileAuth();
      final boundary = await _mount(
        tester,
        auth,
        _documents(),
        width: width,
        scale: scale,
      );
      await tester.pumpAndSettle();
      expect(find.text('Isabella Montague'), findsOneWidget);
      expect(find.text('@isabella.travels'), findsOneWidget);
      expect(find.text('Brooklyn, New York'), findsOneWidget);
      expect(find.text('4850 / 10000 XP'), findsOneWidget);
      expect(find.text('1250 credits'), findsOneWidget);
      expect(
        find.descendant(
          of: _key('profile-library-yours'),
          matching: find.text('1'),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: _key('profile-library-saved'),
          matching: find.text('2'),
        ),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
      await _capture(tester, boundary, '${width.toInt()}-${scale}x');

      await _tap(tester, _key('profile-edit'));
      expect(_key('profile-username'), findsOneWidget);
      await _tap(tester, _key('profile-edit-cancel'));
      await tester.ensureVisible(_key('profile-credits'));
      await tester.pumpAndSettle();
      expect(_key('profile-credits').hitTestable(), findsOneWidget);
      await _tap(tester, _key('profile-logout'));
      expect(
        find.text('You can sign back in with your phone number.'),
        findsOneWidget,
      );
      await _tap(tester, find.text('Cancel'));
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox());
    });
  }

  testWidgets(
    'profile shortcuts open their scope and update counts on return',
    (tester) async {
      final docs = _documents();
      var openedCalendar = false;
      var openedCircle = false;
      await _mount(
        tester,
        _ProfileAuth(),
        docs,
        onOpenCalendar: () => openedCalendar = true,
        onOpenCircle: () => openedCircle = true,
      );
      await tester.pumpAndSettle();
      await _tap(tester, _key('profile-calendar'));
      await _tap(tester, _key('profile-circle'));
      expect(openedCalendar && openedCircle, isTrue);
      await _tap(tester, _key('profile-library-saved'));
      expect(_key('library-card-lisbon'), findsOneWidget);
      expect(_key('library-card-mine'), findsNothing);
      await _tap(tester, _key('library-save-lisbon'));
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(
        find.descendant(
          of: _key('profile-library-saved'),
          matching: find.text('1'),
        ),
        findsOneWidget,
      );
      await _tap(tester, _key('profile-library-yours'));
      expect(_key('library-card-mine'), findsOneWidget);
      expect(_key('library-card-budget'), findsNothing);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'unavailable counts have retry instead of claiming an empty library',
    (tester) async {
      final docs = _documents()..listFailure = Exception('offline');
      await _mount(tester, _ProfileAuth(), docs);
      await tester.pumpAndSettle();
      expect(find.text('Guide counts are unavailable.'), findsOneWidget);
      expect(find.text('—'), findsNWidgets(2));
      docs.listFailure = null;
      await _tap(tester, _key('profile-library-retry'));
      expect(find.text('Guide counts are unavailable.'), findsNothing);
      expect(
        find.descendant(
          of: _key('profile-library-saved'),
          matching: find.text('2'),
        ),
        findsOneWidget,
      );
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'profile handles spaced names, broken photos, and late count responses',
    (tester) async {
      final docs = _RacingCounts();
      final auth = _ProfileAuth()
        ..current = User(
          id: 'me',
          name: '  Ada  Lovelace  ',
          profilePictureUrl: 'https://example.test/broken.png',
        );
      await _mount(tester, auth, docs);
      await tester.pump(const Duration(milliseconds: 100));
      expect(find.text('AL'), findsOneWidget);
      expect(tester.takeException(), isNull);
      auth.change(User(id: 'another', name: '  '));
      await tester.pump();
      expect(docs.requests, hasLength(4));
      docs.requests[2].complete([
        libraryGuide('new-1', 'New guide'),
        libraryGuide('new-2', 'Another guide'),
      ]);
      docs.requests[3].complete([libraryGuide('new-saved', 'New saved guide')]);
      await tester.pumpAndSettle();
      expect(
        find.descendant(
          of: _key('profile-library-yours'),
          matching: find.text('2'),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: _key('profile-library-saved'),
          matching: find.text('1'),
        ),
        findsOneWidget,
      );
      docs.requests[0].complete(
        List.generate(9, (i) => libraryGuide('old-$i', 'Old guide')),
      );
      docs.requests[1].complete(
        List.generate(
          8,
          (i) => libraryGuide('old-saved-$i', 'Old saved guide'),
        ),
      );
      await tester.pumpAndSettle();
      expect(
        find.descendant(
          of: _key('profile-library-yours'),
          matching: find.text('2'),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: _key('profile-library-saved'),
          matching: find.text('1'),
        ),
        findsOneWidget,
      );
      expect(
        find.byType(LinearProgressIndicator),
        findsOneWidget,
      ); // Only level progress remains.
      expect(find.text('Traveler'), findsOneWidget);
      expect(find.text('T'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox());
    },
  );
}
