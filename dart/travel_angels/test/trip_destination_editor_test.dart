import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/travel_destination.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';

import 'support/app_fonts.dart';

const lisbon = <String, dynamic>{
  'id': 'lisbon',
  'name': 'Lisbon',
  'kind': 'city',
  'countryCode': 'PT',
  'countryName': 'Portugal',
  'timeZone': 'Europe/Lisbon',
  'aliases': ['Lisboa'],
};
const tokyo = <String, dynamic>{
  'id': 'tokyo',
  'name': 'Tokyo',
  'kind': 'city',
  'countryName': 'Japan',
  'timeZone': 'Asia/Tokyo',
};

Finder key(String value) => find.byKey(ValueKey(value));

class DestinationService extends TravelCalendarService {
  final queries = <String>[];
  final saved = <Map<String, dynamic>>[];
  final pending = <String, Completer<Map<String, dynamic>>>{};
  List<Map<String, dynamic>> results = [lisbon, tokyo];
  Object? searchError;
  Object? saveError;
  bool lookupUnavailable = false;

  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    expect(path, '/destinations');
    final search = query?['q'] as String? ?? '';
    queries.add(search);
    if (pending[search] != null) return pending[search]!.future;
    if (searchError != null) throw searchError!;
    return {'destinations': results, 'lookupUnavailable': lookupUnavailable};
  }

  @override
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
  }) async {
    if (saveError != null) throw saveError!;
    saved.add({'path': path, ...?data});
    return {};
  }

  @override
  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
  }) => post(path, data: data);
}

class DestinationAuth extends AuthProvider {
  DestinationAuth() : super(AuthService(APIClient('https://example.test')));
  User? current = User(id: 'traveler');
  @override
  User? get user => current;
  @override
  Future<void> verifyToken() async {}
  void switchAccount() {
    current = User(id: 'someone-else');
    notifyListeners();
  }
}

Future<GlobalKey> openEditor(
  WidgetTester tester,
  DestinationService service, {
  Map<String, dynamic>? stop,
  DestinationAuth? auth,
  Size size = const Size(1200, 900),
  double scale = 1,
  double keyboard = 0,
  TargetPlatform? platform,
}) async {
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final boundary = GlobalKey();
  Widget app = RepaintBoundary(
    key: boundary,
    child: MaterialApp(
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme.copyWith(platform: platform),
      builder: (_, child) => MediaQuery(
        data: MediaQueryData(
          size: size,
          textScaler: TextScaler.linear(scale),
          viewInsets: EdgeInsets.only(bottom: keyboard),
        ),
        child: child!,
      ),
      home: Builder(
        builder: (context) => Scaffold(
          body: TextButton(
            onPressed: () => showDialog<void>(
              context: context,
              builder: (_) => TravelStopEditor(
                service: service,
                stop: stop,
                calendars: const [
                  {'id': 'my-calendar', 'title': 'My calendar'},
                ],
              ),
            ),
            child: const Text('Open editor'),
          ),
        ),
      ),
    ),
  );
  if (auth != null) {
    app = ChangeNotifierProvider<AuthProvider>.value(value: auth, child: app);
  }
  await tester.pumpWidget(app);
  await tester.tap(find.text('Open editor'));
  await tester.pumpAndSettle();
  return boundary;
}

Future<void> tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> search(WidgetTester tester, String text) async {
  await tester.ensureVisible(key('trip-location-search'));
  await tester.enterText(key('trip-location-search'), text);
  await tester.pump(const Duration(milliseconds: 350));
  await tester.pumpAndSettle();
}

Future<void> addLocation(WidgetTester tester, String id) async {
  await tap(tester, key('trip-location-search'));
  await tap(tester, key('trip-location-result-$id'));
}

Future<void> capture(
  WidgetTester tester,
  GlobalKey boundary,
  String label,
) async {
  if (!const bool.fromEnvironment('TRIP_SCREENSHOTS')) return;
  await tester.pumpAndSettle();
  await tester.runAsync(() async {
    final image =
        await (boundary.currentContext!.findRenderObject()!
                as RenderRepaintBoundary)
            .toImage();
    final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-trip-editor-$label.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(loadAppFonts);

  test('legacy text stays one location; explicit empty array stays empty', () {
    final legacy = travelDestinations({
      'destination': 'Lisbon, Porto / somewhere else',
    });
    expect(legacy.single.name, 'Lisbon, Porto / somewhere else');
    expect(legacy.single.toSelectionJson(), {
      'name': 'Lisbon, Porto / somewhere else',
    });
    expect(
      travelDestinations({'destination': 'Lisbon', 'destinations': []}),
      isEmpty,
    );
    expect(TravelDestination.fromJson(lisbon).searchText, contains('lisboa'));
  });

  testWidgets('legacy draft saves intact with revision and logistics', (
    tester,
  ) async {
    final service = DestinationService();
    await openEditor(
      tester,
      service,
      stop: {
        'id': 'trip',
        'revision': 7,
        'draft': {
          'title': 'A long weekend',
          'destination': 'Lisbon / Porto, Portugal',
          'timeZone': 'Europe/Paris',
          'datePrecision': 'fixed',
          'startDate': '2099-04-03',
          'endDate': '2099-04-07',
          'hero': {'mediaId': 'private-image'},
          'lodging': [
            {'name': 'Harbor House'},
          ],
          'flights': [
            {'label': 'Outbound'},
          ],
        },
      },
    );
    expect(find.text('Lisbon / Porto, Portugal'), findsOneWidget);
    await tap(tester, key('trip-save-draft'));
    final data = service.saved.single;
    expect(data['destinations'], [
      {'name': 'Lisbon / Porto, Portugal'},
    ]);
    expect(data['destination'], 'Lisbon / Porto, Portugal');
    expect(data['timeZone'], 'Europe/Paris');
    expect(data['revision'], 7);
    expect(data['startDate'], '2099-04-03');
    expect(data['endDate'], '2099-04-07');
    expect(data['hero'], {'mediaId': 'private-image'});
    expect(data['lodging'], [
      {'name': 'Harbor House'},
    ]);
    expect(data['flights'], [
      {'label': 'Outbound'},
    ]);
  });

  testWidgets(
    'selects multiple locations, reorders, removes, and saves ordered ids',
    (tester) async {
      final service = DestinationService();
      await openEditor(tester, service, platform: TargetPlatform.android);
      await tester.enterText(key('trip-title'), 'Across the world');
      await addLocation(tester, 'lisbon');
      await addLocation(tester, 'tokyo');
      final tokyoRect = tester.getRect(key('trip-location-tokyo'));
      final lisbonTarget = tester.getCenter(key('trip-location-lisbon'));
      final drag = await tester.startGesture(
        tokyoRect.centerLeft + const Offset(24, 0),
      );
      await tester.pump(kLongPressTimeout + const Duration(milliseconds: 100));
      await drag.moveTo(lisbonTarget);
      await tester.pump();
      await drag.up();
      await tester.pumpAndSettle();
      await search(tester, 'Our mountain cabin');
      await tap(tester, key('trip-location-custom'));
      await tap(tester, find.byTooltip('Remove Our mountain cabin'));
      await tap(tester, key('trip-location-search'));
      expect(key('trip-location-result-lisbon'), findsNothing);
      expect(key('trip-location-result-tokyo'), findsNothing);
      await tap(tester, key('trip-save-draft'));
      final data = service.saved.single;
      expect(data['destinations'], [
        {'id': 'tokyo'},
        {'id': 'lisbon'},
      ]);
      expect(data['destination'], 'Tokyo → Lisbon');
      expect(data['calendarId'], 'my-calendar');
      expect(data['timeZone'], 'UTC');
      expect(service.queries, contains(''));
    },
  );

  testWidgets('infers a new trip time zone and preserves manual overrides', (
    tester,
  ) async {
    final service = DestinationService();
    await openEditor(tester, service);
    await tester.enterText(key('trip-title'), 'Portugal');
    await addLocation(tester, 'lisbon');
    await tap(tester, find.text('More options'));
    await tester.ensureVisible(key('trip-time-zone'));
    final field = tester.widget<TextFormField>(key('trip-time-zone'));
    expect(field.controller!.text, 'Europe/Lisbon');
    await tester.enterText(key('trip-time-zone'), 'America/New_York');
    await addLocation(tester, 'tokyo');
    await tap(tester, key('trip-save-draft'));
    expect(service.saved.single['timeZone'], 'America/New_York');
  });

  testWidgets('debounces and discards stale search responses', (tester) async {
    final service = DestinationService();
    final old = Completer<Map<String, dynamic>>();
    final latest = Completer<Map<String, dynamic>>();
    service.pending['old'] = old;
    service.pending['latest'] = latest;
    await openEditor(tester, service);
    await tester.enterText(key('trip-location-search'), 'o');
    await tester.pump(const Duration(milliseconds: 150));
    await tester.enterText(key('trip-location-search'), 'old');
    await tester.pump(const Duration(milliseconds: 350));
    await tester.enterText(key('trip-location-search'), 'latest');
    await tester.pump(const Duration(milliseconds: 350));
    latest.complete({
      'destinations': [tokyo],
    });
    await tester.pumpAndSettle();
    old.complete({
      'destinations': [lisbon],
    });
    await tester.pumpAndSettle();
    expect(service.queries, isNot(contains('o')));
    expect(key('trip-location-result-tokyo'), findsOneWidget);
    expect(key('trip-location-result-lisbon'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('search errors retry without losing selected locations', (
    tester,
  ) async {
    final service = DestinationService();
    await openEditor(tester, service);
    await addLocation(tester, 'lisbon');
    service.searchError = const TravelCalendarException(
      'Location search is unavailable.',
    );
    await search(tester, 'Tokyo');
    expect(find.text('Location search is unavailable.'), findsOneWidget);
    expect(key('trip-location-lisbon'), findsOneWidget);
    service.searchError = null;
    await tap(tester, find.text('Retry'));
    await tap(tester, key('trip-location-result-tokyo'));
    expect(key('trip-location-lisbon'), findsOneWidget);
    expect(key('trip-location-tokyo'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('pointer down keeps suggestions available until selection', (
    tester,
  ) async {
    final service = DestinationService();
    await openEditor(tester, service);
    await search(tester, 'Tokyo');
    expect(key('trip-location-custom'), findsNothing);
    final gesture = await tester.startGesture(
      tester.getCenter(key('trip-location-result-tokyo')),
      kind: ui.PointerDeviceKind.mouse,
    );
    await tester.pump();
    expect(key('trip-location-result-tokyo'), findsOneWidget);
    await gesture.up();
    await tester.pumpAndSettle();
    expect(key('trip-location-tokyo'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'provider lookup failures offer retry and explicit custom location',
    (tester) async {
      final service = DestinationService()
        ..results = []
        ..lookupUnavailable = true;
      await openEditor(tester, service);
      await search(tester, 'Our cabin');
      expect(
        find.text('Could not look up more places. Try again.'),
        findsOneWidget,
      );
      expect(find.text('Retry'), findsOneWidget);
      expect(key('trip-location-custom'), findsOneWidget);
      await tester.testTextInput.receiveAction(TextInputAction.search);
      await tester.pumpAndSettle();
      expect(key('trip-location-custom:our cabin'), findsNothing);
      await tap(tester, key('trip-location-search'));
      await tap(tester, key('trip-location-custom'));
      expect(key('trip-location-custom:our cabin'), findsOneWidget);
    },
  );

  testWidgets('save failure remains visible and preserves the completed form', (
    tester,
  ) async {
    final service = DestinationService()
      ..saveError = const TravelCalendarException(
        'Could not save this draft. Try again.',
      );
    await openEditor(tester, service);
    await tester.enterText(key('trip-title'), 'Try again');
    await addLocation(tester, 'lisbon');
    await tap(tester, key('trip-save-draft'));
    expect(
      find.text('Could not save this draft. Try again.').hitTestable(),
      findsOneWidget,
    );
    expect(key('trip-location-lisbon'), findsOneWidget);
    service.saveError = null;
    await tap(tester, key('trip-save-draft'));
    expect(service.saved.single['title'], 'Try again');
  });

  testWidgets(
    'choosing an exact catalog match replaces a legacy chip in place',
    (tester) async {
      final service = DestinationService();
      await openEditor(
        tester,
        service,
        stop: {
          'id': 'trip',
          'title': 'Japan',
          'destination': 'Tokyo',
          'timeZone': 'Europe/Paris',
        },
      );
      await addLocation(tester, 'tokyo');
      expect(key('trip-location-custom:tokyo'), findsNothing);
      expect(key('trip-location-tokyo'), findsOneWidget);
      await tap(tester, key('trip-save-draft'));
      expect(service.saved.single['destinations'], [
        {'id': 'tokyo'},
      ]);
      expect(service.saved.single['timeZone'], 'Europe/Paris');
    },
  );

  testWidgets('account changes close the draft and discard pending search', (
    tester,
  ) async {
    final service = DestinationService();
    final auth = DestinationAuth();
    addTearDown(auth.dispose);
    final response = Completer<Map<String, dynamic>>();
    service.pending['private'] = response;
    await openEditor(tester, service, auth: auth);
    await tester.enterText(key('trip-location-search'), 'private');
    await tester.pump(const Duration(milliseconds: 350));
    auth.switchAccount();
    await tester.pumpAndSettle();
    response.complete({
      'destinations': [lisbon],
    });
    await tester.pumpAndSettle();
    expect(find.byType(TravelStopEditor), findsNothing);
    expect(service.saved, isEmpty);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'removing the last location validates instead of reviving legacy text',
    (tester) async {
      final service = DestinationService();
      await openEditor(
        tester,
        service,
        stop: {'id': 'trip', 'title': 'A trip', 'destination': 'Lisbon'},
      );
      await tap(tester, find.byTooltip('Remove Lisbon'));
      await tap(tester, key('trip-save-draft'));
      expect(find.text('Add at least one location.'), findsOneWidget);
      expect(service.saved, isEmpty);
    },
  );

  testWidgets(
    'a selection of twelve keeps the add field visible and enforces the limit',
    (tester) async {
      final service = DestinationService();
      await openEditor(
        tester,
        service,
        stop: {
          'id': 'trip',
          'title': 'Many places',
          'destinations': List.generate(
            12,
            (index) => {'id': 'place-$index', 'name': 'Place $index'},
          ),
        },
      );
      await tester.ensureVisible(key('trip-location-search'));
      final field = tester.widget<TextField>(key('trip-location-search'));
      expect(field.enabled, false);
      expect(find.text('Remove a location to add another.'), findsOneWidget);
      await tap(tester, find.byTooltip('Remove Place 11'));
      expect(
        tester.widget<TextField>(key('trip-location-search')).enabled,
        true,
      );
      expect(tester.takeException(), isNull);
    },
  );

  for (final variant in [
    (name: 'desktop', size: const Size(1200, 900), scale: 1.0, keyboard: 0.0),
    (name: 'mobile', size: const Size(390, 844), scale: 1.0, keyboard: 0.0),
    (name: 'large-text', size: const Size(320, 740), scale: 2.0, keyboard: 0.0),
    (name: 'keyboard', size: const Size(390, 844), scale: 1.0, keyboard: 330.0),
  ]) {
    testWidgets('editor keeps save accessible on ${variant.name}', (
      tester,
    ) async {
      final service = DestinationService();
      final boundary = await openEditor(
        tester,
        service,
        size: variant.size,
        scale: variant.scale,
        keyboard: variant.keyboard,
      );
      await tester.enterText(key('trip-title'), 'Summer by the sea');
      await search(tester, 'Lis');
      await capture(tester, boundary, '${variant.name}-search');
      await tap(tester, key('trip-location-result-lisbon'));
      expect(
        tester.getRect(key('trip-save-draft')).bottom,
        lessThanOrEqualTo(variant.size.height - variant.keyboard),
      );
      await tester.ensureVisible(key('trip-description'));
      await tester.enterText(
        key('trip-description'),
        'A few unhurried days together.',
      );
      await capture(tester, boundary, variant.name);
      expect(tester.takeException(), isNull);
      await tap(tester, key('trip-save-draft'));
      expect(service.saved.single['title'], 'Summer by the sea');
      expect(service.saved.single['destinations'], [
        {'id': 'lisbon'},
      ]);
      expect(tester.takeException(), isNull);
    });
  }
}
