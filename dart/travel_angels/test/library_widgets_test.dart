import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/screens/library_document_screen.dart';
import 'package:travel_angels/screens/library_screen.dart';
import 'package:travel_angels/theme/app_theme.dart';

import 'library_test_support.dart';
import 'support/app_fonts.dart';

Finder _key(String value) => find.byKey(ValueKey(value));

class _LibraryAuth extends AuthProvider {
  _LibraryAuth() : super(AuthService(APIClient('https://example.test')));
  User? current = User(id: 'me', name: 'You');
  @override
  User? get user => current;
  @override
  Future<void> verifyToken() async {}
  void change(User? next) {
    current = next;
    notifyListeners();
  }
}

class _RacingLibrary extends LibraryFakeDocuments {
  final network = Completer<List<Map<String, dynamic>>>();
  @override
  Future<List<Map<String, dynamic>>> getLibraryDocuments(String scope) async =>
      scope == 'network'
      ? network.future
      : [libraryGuide('mine', 'My private guide', owner: 'me')];
}

DioException _denied([int status = 403]) {
  final request = RequestOptions(path: '/travel-angels/documents/guide');
  return DioException(
    requestOptions: request,
    type: DioExceptionType.badResponse,
    response: Response(requestOptions: request, statusCode: status),
  );
}

Future<GlobalKey> _mount(
  WidgetTester tester,
  Widget child, {
  double width = 1200,
  double scale = 1,
}) async {
  final size = Size(width, width >= 600 ? 1000 : 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final boundary = GlobalKey();
  await tester.pumpWidget(
    RepaintBoundary(
      key: boundary,
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
        home: Scaffold(body: child),
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

Future<void> _select(WidgetTester tester, String key, String label) async {
  await _tap(tester, _key(key));
  if (find.text(label).evaluate().isEmpty) {
    await tester.scrollUntilVisible(
      find.text(label),
      100,
      scrollable: find.byType(Scrollable).last,
    );
  }
  await _tap(tester, find.text(label).last);
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
      '/private/tmp/travel-angels-library-$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

LibraryDocumentScreen _reader(
  LibraryFakeDocuments service, {
  String userId = 'me',
  Future<bool?> Function(BuildContext, dynamic)? onEdit,
}) => LibraryDocumentScreen(
  document: service.document('guide'),
  service: service,
  friendService: LibraryFakeFriends(),
  userId: userId,
  onEditDocument: onEdit,
);

void main() {
  setUpAll(loadAppFonts);

  testWidgets(
    'a linked original is distinguished from an imported saved copy',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Original Google document'),
      ]);
      service.documents['guide']!.addAll({
        'provider': 'google_docs',
        'content': null,
        'link': 'https://docs.google.com/document/d/example',
      });
      await _mount(tester, _reader(service));
      await tester.pumpAndSettle();
      expect(find.text('Linked original'), findsOneWidget);
      expect(find.text('Saved copy'), findsNothing);
      expect(
        find.textContaining('This guide links to the original file.'),
        findsOneWidget,
      );
      expect(_key('library-reader-source'), findsOneWidget);
      expect(_key('library-reader-content'), findsNothing);
      await tester.pumpWidget(const SizedBox());

      service.documents['guide']!['content'] =
          '# Imported notes\nA snapshot from the original.';
      await _mount(tester, _reader(service));
      await tester.pumpAndSettle();
      expect(find.text('Saved copy'), findsOneWidget);
      expect(find.text('Linked original'), findsNothing);
      expect(
        find.textContaining(
          'Changes to the original file do not sync automatically.',
        ),
        findsOneWidget,
      );
      expect(_key('library-reader-content'), findsOneWidget);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'a late network response cannot replace the selected Yours scope',
    (tester) async {
      final service = _RacingLibrary();
      await _mount(tester, LibraryScreen(service: service, userId: 'me'));
      await tester.tap(_key('library-tab-yours'));
      await tester.pumpAndSettle();
      expect(_key('library-card-mine'), findsOneWidget);
      service.network.complete([libraryGuide('old', 'Late network guide')]);
      await tester.pumpAndSettle();
      expect(_key('library-card-mine'), findsOneWidget);
      expect(_key('library-card-old'), findsNothing);
    },
  );

  testWidgets(
    'active authentication overrides a cached owner ID after logout and account switch',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Owner notes', owner: 'me'),
      ]);
      final auth = _LibraryAuth();
      await _mount(
        tester,
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: _reader(service),
        ),
      );
      await tester.pumpAndSettle();
      expect(_key('library-reader-share'), findsOneWidget);
      await _tap(tester, _key('library-reader-share'));
      auth.change(null);
      await tester.pumpAndSettle();
      expect(find.text('Who can see this?'), findsNothing);
      expect(_key('library-reader-content'), findsNothing);
      expect(_key('library-reader-share'), findsNothing);
      auth.change(User(id: 'another-person', name: 'Nora'));
      await tester.pumpAndSettle();
      expect(_key('library-reader-content'), findsOneWidget);
      expect(_key('library-reader-share'), findsNothing);
      await tester.pumpWidget(const SizedBox());
      auth.dispose();
    },
  );

  testWidgets(
    'owner sharing remains usable at 320px with twice-sized real fonts',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide(
          'guide',
          'A slow week in Lisbon',
          owner: 'me',
          visibility: 'private',
        ),
      ]);
      final boundary = await _mount(
        tester,
        _reader(service),
        width: 320,
        scale: 2,
      );
      await tester.pumpAndSettle();
      await _tap(tester, _key('library-reader-share'));
      await _select(tester, 'library-sharing-visibility', 'Selected friends');
      await _tap(tester, _key('library-sharing-friend-friend-1'));
      await _capture(tester, boundary, 'share-320-2x');
      await _tap(tester, _key('library-sharing-save'));
      expect(service.documents['guide']!['sharedWithUserIds'], ['friend-1']);
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'browse scopes, search destination or author, and filter format',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('lisbon', 'A slow week in Lisbon'),
        libraryGuide(
          'kyoto',
          'Japan budget',
          format: 'spreadsheet',
          destination: 'Kyoto, Japan',
          content: 'Place,Cost\nTemple,500',
          saved: true,
        ),
        libraryGuide(
          'mine',
          'My summer itinerary',
          owner: 'me',
          format: 'itinerary',
        ),
      ]);
      await _mount(tester, LibraryScreen(service: service, userId: 'me'));
      await tester.pumpAndSettle();
      expect(_key('library-card-lisbon'), findsOneWidget);
      expect(_key('library-card-kyoto'), findsOneWidget);
      expect(_key('library-card-mine'), findsNothing);
      await tester.enterText(_key('library-search'), 'KYOTO');
      await tester.pump();
      expect(_key('library-card-lisbon'), findsNothing);
      expect(_key('library-card-kyoto'), findsOneWidget);
      await tester.enterText(_key('library-search'), 'Maya');
      await tester.pump();
      expect(_key('library-card-lisbon'), findsOneWidget);
      await _select(tester, 'library-format-filter', 'Spreadsheets');
      expect(_key('library-card-lisbon'), findsNothing);
      expect(_key('library-card-kyoto'), findsOneWidget);
      await tester.enterText(_key('library-search'), '');
      await _select(tester, 'library-format-filter', 'All formats');
      await _tap(tester, _key('library-tab-saved'));
      expect(_key('library-card-kyoto'), findsOneWidget);
      expect(_key('library-card-lisbon'), findsNothing);
      await _tap(tester, _key('library-tab-yours'));
      expect(_key('library-card-mine'), findsOneWidget);
      expect(_key('library-card-kyoto'), findsNothing);
      expect(
        service.calls
            .where((call) => call['action'] == 'list')
            .map((call) => call['scope']),
        ['network', 'saved', 'yours'],
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'save is pending until confirmed, unsave leaves Saved, and denied save removes stale card',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Portugal favorites'),
      ]);
      service.saveGate = Completer<void>();
      await _mount(tester, LibraryScreen(service: service, userId: 'me'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(_key('library-save-guide'));
      await tester.pumpAndSettle();
      await tester.tap(_key('library-save-guide'));
      await tester.pump();
      expect(
        tester.widget<IconButton>(_key('library-save-guide')).onPressed,
        isNull,
      );
      expect(service.documents['guide']!['isSaved'], isFalse);
      service.saveGate!.complete();
      await tester.pumpAndSettle();
      expect(service.documents['guide']!['isSaved'], isTrue);
      service.saveGate = null;
      await _tap(tester, _key('library-tab-saved'));
      await _tap(tester, _key('library-save-guide'));
      expect(_key('library-card-guide'), findsNothing);
      expect(
        find.text('Save a guide to keep it close for your next trip.'),
        findsOneWidget,
      );
      await _tap(tester, _key('library-tab-network'));
      service.saveFailure = _denied();
      await _tap(tester, _key('library-save-guide'));
      expect(_key('library-card-guide'), findsNothing);
      expect(
        find.text('This guide is no longer available to you.'),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'loading and errors never masquerade as an empty library; retry and add reload',
    (tester) async {
      final service = LibraryFakeDocuments()..listGate = Completer<void>();
      var added = false;
      await _mount(
        tester,
        LibraryScreen(
          service: service,
          userId: 'me',
          onAddGuide: (_) async {
            added = true;
            service.documents['new'] = libraryGuide('new', 'New guide');
          },
        ),
      );
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(
        find.text('Guides shared by your friends will appear here.'),
        findsNothing,
      );
      service.listFailure = StateError('offline');
      service.listGate!.complete();
      await tester.pumpAndSettle();
      expect(
        find.text('Your library could not be loaded. Please try again.'),
        findsOneWidget,
      );
      service.listFailure = null;
      service.listGate = null;
      await _tap(tester, find.text('Retry'));
      expect(
        find.text('Guides shared by your friends will appear here.'),
        findsOneWidget,
      );
      await _tap(tester, _key('library-add-guide'));
      expect(added, isTrue);
      expect(_key('library-card-new'), findsOneWidget);
    },
  );

  testWidgets(
    'reader waits for fresh access, then hides cached content after revocation',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Private travel notes'),
      ])..readGate = Completer<void>();
      await _mount(tester, _reader(service));
      expect(find.text('Private travel notes'), findsNothing);
      expect(_key('library-reader-content'), findsNothing);
      expect(_key('library-reader-save'), findsNothing);
      service.readFailure = _denied(404);
      service.readGate!.complete();
      await tester.pumpAndSettle();
      expect(
        find.text('This guide is no longer available to you.'),
        findsOneWidget,
      );
      expect(_key('library-reader-content'), findsNothing);
      service.readFailure = null;
      service.readGate = null;
      await _tap(tester, find.text('Retry'));
      expect(find.text('Private travel notes'), findsOneWidget);
      expect(_key('library-reader-content'), findsOneWidget);
      service.readFailure = _denied();
      await tester.pump(const Duration(seconds: 31));
      await tester.pumpAndSettle();
      expect(_key('library-reader-content'), findsNothing);
      expect(_key('library-reader-trip'), findsNothing);
      expect(find.text('Private travel notes'), findsNothing);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'reader grants owner controls only to the matching signed-in account',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Shared guide'),
      ]);
      await _mount(tester, _reader(service));
      await tester.pumpAndSettle();
      for (final action in ['edit', 'share', 'delete']) {
        expect(_key('library-reader-$action'), findsNothing);
      }
      expect(_key('library-reader-save'), findsOneWidget);
      expect(_key('library-reader-trip'), findsOneWidget);
      expect(find.text('Shared with you'), findsOneWidget);
      await tester.pumpWidget(const SizedBox());

      service.documents['guide']!['userId'] = '';
      await _mount(tester, _reader(service, userId: ''));
      await tester.pumpAndSettle();
      expect(_key('library-reader-share'), findsNothing);
      await tester.pumpWidget(const SizedBox());

      service.documents['guide']!['userId'] = 'me';
      var edited = false;
      await _mount(
        tester,
        _reader(
          service,
          onEdit: (_, _) async {
            edited = true;
            service.documents['guide']!['title'] = 'Edited guide';
            return true;
          },
        ),
      );
      await tester.pumpAndSettle();
      for (final action in ['edit', 'share', 'delete']) {
        expect(_key('library-reader-$action'), findsOneWidget);
      }
      await _tap(tester, _key('library-reader-edit'));
      expect(edited, isTrue);
      expect(find.text('Edited guide'), findsOneWidget);
      await _tap(tester, _key('library-reader-delete'));
      await _tap(tester, find.text('Cancel'));
      expect(service.calls.any((call) => call['action'] == 'delete'), isFalse);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'sharing requires a selected friend and narrowing audience clears grants',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide(
          'guide',
          'My Portugal guide',
          owner: 'me',
          visibility: 'private',
        ),
      ]);
      await _mount(tester, _reader(service));
      await tester.pumpAndSettle();
      await _tap(tester, _key('library-reader-share'));
      await _select(tester, 'library-sharing-visibility', 'Selected friends');
      await _tap(tester, _key('library-sharing-save'));
      expect(
        find.text('Choose at least one friend, or select Only me.'),
        findsOneWidget,
      );
      expect(service.calls.any((call) => call['action'] == 'share'), isFalse);
      await _tap(tester, _key('library-sharing-friend-friend-2'));
      await _tap(tester, _key('library-sharing-save'));
      expect(service.documents['guide']!['visibility'], 'selected');
      expect(service.documents['guide']!['sharedWithUserIds'], ['friend-2']);
      await _tap(tester, _key('library-reader-share'));
      await _select(tester, 'library-sharing-visibility', 'Only me');
      await _tap(tester, _key('library-sharing-save'));
      expect(service.documents['guide']!['visibility'], 'private');
      expect(service.documents['guide']!['sharedWithUserIds'], isEmpty);
      expect(find.text('Only you'), findsOneWidget);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'personal trip references can be added and removed without changing sharing',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Shared Portugal guide', visibility: 'selected'),
      ]);
      service.documents['guide']!['sharedWithUserIds'] = ['me'];
      await _mount(tester, _reader(service));
      await tester.pumpAndSettle();
      await _tap(tester, _key('library-reader-trip'));
      expect(
        find.textContaining('does not change who can access'),
        findsOneWidget,
      );
      await _tap(tester, _key('library-trip-trip-1'));
      await _tap(tester, _key('library-trips-save'));
      expect(service.documents['guide']!['savedTripIds'], ['trip-1']);
      expect(find.textContaining('Our Portugal trip'), findsOneWidget);
      expect(service.documents['guide']!['visibility'], 'selected');
      expect(service.documents['guide']!['sharedWithUserIds'], ['me']);
      await _tap(tester, _key('library-reader-trip'));
      expect(
        tester.widget<CheckboxListTile>(_key('library-trip-trip-1')).value,
        isTrue,
      );
      await _tap(tester, _key('library-trip-trip-1'));
      await _tap(tester, _key('library-trips-save'));
      expect(service.documents['guide']!['savedTripIds'], isEmpty);
      expect(service.calls.any((call) => call['action'] == 'share'), isFalse);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets(
    'revoked sharing access closes its open dialog and clears the reader',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('guide', 'Owner notes', owner: 'me'),
      ]);
      await _mount(tester, _reader(service));
      await tester.pumpAndSettle();
      await _tap(tester, _key('library-reader-share'));
      expect(find.text('Who can see this?'), findsOneWidget);
      service.readFailure = _denied();
      await tester.pump(const Duration(seconds: 31));
      await tester.pumpAndSettle();
      expect(find.text('Who can see this?'), findsNothing);
      expect(_key('library-reader-content'), findsNothing);
      expect(_key('library-reader-share'), findsNothing);
      await tester.pumpWidget(const SizedBox());
    },
  );

  for (final viewport in [
    (1200.0, 1.0, '1200'),
    (390.0, 1.0, '390'),
    (320.0, 2.0, '320-2x'),
  ]) {
    testWidgets(
      'library and reader remain usable with real fonts at ${viewport.$3}',
      (tester) async {
        final service = LibraryFakeDocuments([
          libraryGuide('guide', 'A slow week in Lisbon'),
          libraryGuide(
            'budget',
            'Portugal budget',
            format: 'spreadsheet',
            content: 'Place,Cost\nAlfama,15 EUR\nSintra,20 EUR',
          ),
          libraryGuide('days', 'A few days on the coast', format: 'itinerary'),
        ]);
        final boundary = await _mount(
          tester,
          LibraryScreen(service: service, userId: 'me'),
          width: viewport.$1,
          scale: viewport.$2,
        );
        await tester.pumpAndSettle();
        await _capture(tester, boundary, viewport.$3);
        expect(_key('library-add-guide'), findsOneWidget);
        await _tap(tester, _key('library-save-guide'));
        expect(service.documents['guide']!['isSaved'], isTrue);
        expect(tester.takeException(), isNull);
        await tester.pumpWidget(const SizedBox());
        final readerBoundary = await _mount(
          tester,
          _reader(service),
          width: viewport.$1,
          scale: viewport.$2,
        );
        await tester.pumpAndSettle();
        await _capture(tester, readerBoundary, 'reader-${viewport.$3}');
        await _tap(tester, _key('library-reader-trip'));
        await _tap(tester, _key('library-trip-trip-2'));
        await _tap(tester, _key('library-trips-save'));
        expect(service.documents['guide']!['savedTripIds'], ['trip-2']);
        expect(tester.takeException(), isNull);
        await tester.pumpWidget(const SizedBox());
      },
    );
  }
}
