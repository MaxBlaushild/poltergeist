import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';
import 'package:travel_angels/widgets/travel_trip_list.dart';

import 'support/app_fonts.dart';

TravelStop _trip(
  String id,
  String title, {
  String? start = '2028-03-10',
  String? end = '2028-03-15',
  String precision = 'fixed',
  String destination = 'Kyoto, Japan',
}) => {
  'id': id,
  'title': title,
  'destination': destination,
  'startDate': start,
  'endDate': end,
  'datePrecision': precision,
  'status': 'confirmed',
  'published': true,
};

List<String> _ids(List<TravelStop> stops) =>
    stops.map((stop) => stop['id'] as String).toList();

List<String> _visibleIds(WidgetTester tester) => tester
    .widgetList<TravelTripListRow>(find.byType(TravelTripListRow))
    .map((row) => row.stop['id'] as String)
    .toList();

Future<GlobalKey> _mount(
  WidgetTester tester, {
  required List<TravelStop> stops,
  ValueChanged<TravelStop>? onStopTap,
  Set<String> selectedIds = const {},
  bool expandPast = false,
  double width = 390,
  double scale = 1,
}) async {
  final size = Size(width, 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final boundary = GlobalKey();
  await tester.pumpWidget(
    MaterialApp(
      theme: AppTheme.lightTheme,
      builder: (_, child) => MediaQuery(
        data: MediaQueryData(size: size, textScaler: TextScaler.linear(scale)),
        child: child!,
      ),
      home: RepaintBoundary(
        key: boundary,
        child: Scaffold(
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: TravelTripList(
              stops: stops,
              today: DateTime(2028, 3, 1),
              onStopTap: onStopTap ?? (_) {},
              selectedIds: selectedIds,
              expandPast: expandPast,
            ),
          ),
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return boundary;
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
      '/private/tmp/travel-angels-trip-list-$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(loadAppFonts);

  test(
    'upcoming dates lead, spanning multiple years and approximate dates',
    () {
      final stops = [
        _trip('next-year', 'Next year', start: '2029-01-01', end: '2029-01-02'),
        _trip('later', 'Later', start: '2028-12-10', end: '2028-12-15'),
        _trip('today', 'Today', start: '2028-03-01', end: '2028-03-04'),
        _trip(
          'around',
          'Approximate',
          start: '2028-03-07',
          end: '2028-03-14',
          precision: 'approximate',
        ),
        _trip('ongoing', 'Ongoing', start: '2028-02-26', end: '2028-03-02'),
        _trip('old', 'Old', start: '2026-12-10', end: '2026-12-15'),
        _trip('new-past', 'Recent', start: '2028-02-20', end: '2028-02-28'),
        _trip('tbd', 'Someday', precision: 'tbd'),
      ];
      final before = List<TravelStop>.from(stops);
      final groups = groupTravelTrips(
        stops,
        today: DateTime(2028, 3, 1, 23, 59),
      );
      expect(_ids(groups.upcoming), ['today', 'around', 'later', 'next-year']);
      expect(_ids(groups.ongoing), ['ongoing']);
      expect(_ids(groups.undated), ['tbd']);
      expect(_ids(groups.past), ['new-past', 'old']);
      expect(stops, orderedEquals(before));
      expect(identical(groups.upcoming.first, stops[2]), isTrue);
      expect(() => groups.upcoming.add(stops.first), throwsUnsupportedError);
    },
  );

  test('civil dates keep all of today inclusive through local midnight', () {
    final stops = [
      _trip('starting', 'Starts today', start: '2028-03-01', end: '2028-03-01'),
      _trip('ending', 'Ends today', start: '2028-02-28', end: '2028-03-01'),
      _trip('yesterday', 'Finished', start: '2028-02-27', end: '2028-02-29'),
    ];
    for (final today in [
      DateTime(2028, 3, 1),
      DateTime(2028, 3, 1, 23, 59, 59),
      DateTime.utc(2028, 3, 1, 5),
    ]) {
      final groups = groupTravelTrips(stops, today: today);
      expect(_ids(groups.upcoming), ['starting']);
      expect(_ids(groups.ongoing), ['ending']);
      expect(_ids(groups.past), ['yesterday']);
    }
    final tomorrow = groupTravelTrips(stops, today: DateTime(2028, 3, 2));
    expect(tomorrow.upcoming, isEmpty);
    expect(tomorrow.ongoing, isEmpty);
    expect(_ids(tomorrow.past), ['starting', 'ending', 'yesterday']);
  });

  test(
    'TBD ignores stale dates, invalid spans are undated, missing ends are one day',
    () {
      final groups = groupTravelTrips([
        _trip(
          'tbd',
          'A TBD',
          precision: 'tbd',
          start: '2020-01-01',
          end: '2020-02-01',
        ),
        _trip('missing', 'B Missing', start: null, end: null),
        _trip('reverse', 'C Reverse', start: '2028-04-10', end: '2028-04-01'),
        _trip('invalid', 'D Invalid', start: '2028-02-30', end: '2028-03-01'),
        _trip('single', 'Single day', start: '2028-03-02', end: null),
      ], today: DateTime(2028, 3, 1));
      expect(_ids(groups.undated), ['tbd', 'missing', 'reverse', 'invalid']);
      expect(_ids(groups.upcoming), ['single']);
      expect(groups.past, isEmpty);
    },
  );

  test('ties are deterministic across input ordering', () {
    final stops = [
      _trip('c', 'Bravo'),
      _trip('b', 'Alpha'),
      _trip('a', 'Alpha'),
      _trip('shorter', 'Zulu', end: '2028-03-12'),
    ];
    for (final items in [stops, stops.reversed.toList()]) {
      expect(
        _ids(groupTravelTrips(items, today: DateTime(2028, 3, 1)).upcoming),
        ['shorter', 'a', 'b', 'c'],
      );
    }
  });

  testWidgets(
    'upcoming, ongoing, and undated rows precede collapsed past trips',
    (tester) async {
      await _mount(
        tester,
        stops: [
          _trip('past', 'A past trip', start: '2028-01-01', end: '2028-01-05'),
          _trip('tbd', 'An open idea', precision: 'tbd'),
          _trip(
            'now',
            'Currently away',
            start: '2028-02-26',
            end: '2028-03-01',
          ),
          _trip('next', 'The next trip'),
        ],
      );
      expect(_visibleIds(tester), ['next', 'now', 'tbd']);
      expect(find.text('A past trip'), findsNothing);
      expect(find.text('Past trips · 1'), findsOneWidget);
      final toggle = find.byKey(const ValueKey('travel-trip-list-past-toggle'));
      await tester.ensureVisible(toggle);
      await tester.tap(toggle);
      await tester.pumpAndSettle();
      expect(_visibleIds(tester), ['next', 'now', 'tbd', 'past']);
      await tester.tap(toggle);
      await tester.pumpAndSettle();
      expect(find.text('A past trip'), findsNothing);
    },
  );

  testWidgets('explicit past filter expands newest past trips first', (
    tester,
  ) async {
    await _mount(
      tester,
      expandPast: true,
      stops: [
        _trip('old', 'Old trip', start: '2020-01-01', end: '2020-01-03'),
        _trip('recent', 'Recent trip', start: '2028-02-01', end: '2028-02-03'),
      ],
    );
    expect(_visibleIds(tester), ['recent', 'old']);
    expect(tester.takeException(), isNull);
  });

  testWidgets('row opens original stop and exposes selected state', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    final stop = _trip('next', 'Cherry blossoms');
    TravelStop? opened;
    await _mount(
      tester,
      stops: [stop],
      selectedIds: {'next'},
      onStopTap: (value) => opened = value,
    );
    final row = find.byType(TravelTripListRow);
    expect(find.text('Selected · not submitted'), findsOneWidget);
    expect(
      tester.getSemantics(row).flagsCollection.isSelected,
      ui.Tristate.isTrue,
    );
    await tester.tap(find.text('Cherry blossoms'));
    expect(identical(opened, stop), isTrue);
    semantics.dispose();
  });

  testWidgets('date labels retain years, approximate hints, and TBD meaning', (
    tester,
  ) async {
    await _mount(
      tester,
      stops: [
        _trip('around', 'Approximate', precision: 'approximate'),
        _trip('year', 'New Year', start: '2028-12-29', end: '2029-01-04'),
        _trip(
          'tbd',
          'Undecided',
          precision: 'tbd',
          start: '2020-01-01',
          end: '2020-01-02',
        ),
      ],
    );
    expect(find.text('Around Mar 10–15, 2028'), findsOneWidget);
    expect(find.text('Dec 29, 2028 – Jan 4, 2029'), findsOneWidget);
    expect(find.text('Dates to be decided'), findsOneWidget);
  });

  for (final (width, scale) in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('compact list fits ${width.toInt()}px at ${scale}x text', (
      tester,
    ) async {
      final boundary = await _mount(
        tester,
        width: width,
        scale: scale,
        stops: [
          _trip(
            'japan',
            'Spring in Japan',
            start: '2028-03-12',
            end: '2028-03-26',
            destination: 'Tokyo → Kyoto → Osaka',
          ),
          _trip(
            'italy',
            'A summer on the coast',
            start: '2028-06-02',
            end: '2028-06-14',
            destination: 'Amalfi Coast, Italy',
          ),
          _trip(
            'portugal',
            'A few weeks in Portugal',
            start: '2028-09-04',
            end: '2028-09-20',
            precision: 'approximate',
            destination: 'Lisbon → Porto',
          ),
          _trip(
            'now',
            'The long way home',
            start: '2028-02-25',
            end: '2028-03-03',
            destination: 'Melbourne, Australia',
          ),
          _trip(
            'idea',
            'A winter in the mountains',
            precision: 'tbd',
            destination: 'French Alps',
          ),
          _trip(
            'past',
            'Last summer',
            start: '2027-07-01',
            end: '2027-07-10',
            destination: 'Greece',
          ),
        ],
      );
      expect(tester.takeException(), isNull);
      await _capture(tester, boundary, '${width.toInt()}-${scale}x');
      await tester.ensureVisible(find.text('A winter in the mountains'));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.ensureVisible(
        find.byKey(const ValueKey('travel-trip-list-past-toggle')),
      );
      await tester.pumpAndSettle();
      await _capture(tester, boundary, '${width.toInt()}-${scale}x-past');
      await tester.tap(
        find.byKey(const ValueKey('travel-trip-list-past-toggle')),
      );
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Last summer'));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  }
}
