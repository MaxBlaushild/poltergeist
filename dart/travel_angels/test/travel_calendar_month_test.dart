import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

import 'support/app_fonts.dart';

TravelStop _trip(
  String id,
  String title,
  String destination, {
  String start = '2028-02-10',
  String end = '2028-02-15',
  String precision = 'fixed',
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

Finder _key(String value) => find.byKey(ValueKey(value));

Finder _ribbons(String id) => find.byWidgetPredicate((widget) {
  final key = widget.key;
  return key is ValueKey<String> &&
      key.value.startsWith('travel-calendar-ribbon-') &&
      key.value.endsWith('-$id');
});

Future<GlobalKey> _mountMonth(
  WidgetTester tester, {
  required DateTime month,
  required List<TravelStop> stops,
  double width = 1200,
  double scale = 1,
  ValueChanged<TravelStop>? onStopTap,
  ValueChanged<DateTime>? onMonthChanged,
}) async {
  final size = Size(width, width > 600 ? 1000 : 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  var currentMonth = month;
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
            padding: EdgeInsets.symmetric(
              horizontal: width < 360 ? 4 : 16,
              vertical: 16,
            ),
            child: StatefulBuilder(
              builder: (context, setState) => TravelCalendarMonth(
                month: currentMonth,
                stops: stops,
                onMonthChanged: (next) {
                  setState(() => currentMonth = next);
                  onMonthChanged?.call(next);
                },
                onStopTap: onStopTap ?? (_) {},
              ),
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
      '/private/tmp/travel-angels-month-$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

Color? _ribbonColor(WidgetTester tester, String id) => tester
    .widget<Material>(
      find
          .descendant(of: _ribbons(id).first, matching: find.byType(Material))
          .first,
    )
    .color;

Future<void> _chooseMonth(
  WidgetTester tester, {
  required String year,
  required String month,
}) async {
  await tester.tap(_key('travel-calendar-month-picker'));
  await tester.pumpAndSettle();
  await tester.tap(_key('travel-calendar-year-selector'));
  await tester.pumpAndSettle();
  await tester.ensureVisible(find.text(year).last);
  await tester.pumpAndSettle();
  await tester.tap(find.text(year).last);
  await tester.pumpAndSettle();
  await tester.tap(_key('travel-calendar-month-selector'));
  await tester.pumpAndSettle();
  if (find.text(month).evaluate().isEmpty) {
    await tester.scrollUntilVisible(
      find.text(month),
      -120,
      scrollable: find.byType(Scrollable).last,
    );
  } else {
    await tester.ensureVisible(find.text(month).last);
  }
  await tester.pumpAndSettle();
  await tester.tap(find.text(month).last);
  await tester.pumpAndSettle();
  await tester.tap(_key('travel-calendar-month-confirm'));
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(loadAppFonts);

  testWidgets('ribbons continue across weeks and month boundaries', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    TravelStop? opened;
    await _mountMonth(
      tester,
      month: DateTime(2028, 2),
      stops: [
        _trip(
          'coast',
          'Spring along the coast',
          'Lisbon, Portugal',
          start: '2028-02-26',
          end: '2028-03-03',
        ),
      ],
      onStopTap: (stop) => opened = stop,
    );
    expect(_ribbons('coast'), findsNWidgets(2));
    expect(_key('travel-calendar-ribbon-2028-02-21-coast'), findsOneWidget);
    expect(_key('travel-calendar-ribbon-2028-02-28-coast'), findsOneWidget);
    final first = _key('travel-calendar-ribbon-2028-02-21-coast');
    expect(
      find.descendant(of: first, matching: find.textContaining('Spring along')),
      findsWidgets,
    );
    expect(
      find.descendant(of: first, matching: find.textContaining('Lisbon')),
      findsWidgets,
    );
    final description = tester.getSemantics(first).getSemanticsData().label;
    expect(description, contains('Spring along the coast'));
    expect(description, contains('Lisbon, Portugal'));
    expect(description, contains('continues into next week'));
    semantics.dispose();
    await tester.ensureVisible(first);
    await tester.tap(first);
    expect(opened?['id'], 'coast');

    await tester.ensureVisible(_key('travel-calendar-next'));
    await tester.tap(_key('travel-calendar-next'));
    await tester.pumpAndSettle();
    expect(find.text('March 2028'), findsOneWidget);
    expect(_ribbons('coast'), findsOneWidget);
    expect(_key('travel-calendar-ribbon-2028-02-28-coast'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('overlap overflow opens all trips for that day and navigates', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    TravelStop? opened;
    final trips = [
      _trip('a', 'Coastal mornings', 'Lisbon, Portugal'),
      _trip('b', 'City weekend', 'Porto, Portugal'),
      _trip('c', 'Winter sunshine', 'Faro, Portugal'),
      _trip('d', 'Island afternoons', 'Madeira, Portugal'),
      _trip('e', 'A few days together', 'Sintra, Portugal'),
    ];
    await _mountMonth(
      tester,
      month: DateTime(2028, 2),
      stops: trips,
      width: 390,
      onStopTap: (stop) => opened = stop,
    );
    final more = _key('travel-calendar-more-2028-02-10');
    expect(more, findsOneWidget);
    expect(
      tester.getSemantics(more).getSemanticsData().label,
      contains('more trips'),
    );
    semantics.dispose();
    await tester.ensureVisible(more);
    await tester.tap(more);
    await tester.pumpAndSettle();
    final sheet = find.byType(BottomSheet);
    expect(sheet, findsOneWidget);
    for (final trip in trips) {
      final tripCard = find.descendant(
        of: sheet,
        matching: find.byWidgetPredicate(
          (widget) =>
              widget is TravelStopCard && widget.stop['id'] == trip['id'],
        ),
      );
      await tester.scrollUntilVisible(
        tripCard,
        180,
        scrollable: find
            .descendant(of: sheet, matching: find.byType(Scrollable))
            .first,
      );
      await tester.pumpAndSettle();
      expect(tripCard, findsOneWidget);
    }
    final card = find.descendant(
      of: sheet,
      matching: find.byWidgetPredicate(
        (widget) => widget is TravelStopCard && widget.stop['id'] == 'e',
      ),
    );
    await tester.ensureVisible(card);
    await tester.tap(card);
    await tester.pumpAndSettle();
    expect(opened?['id'], 'e');
    expect(find.byType(BottomSheet), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('empty months navigate across years and return to today', (
    tester,
  ) async {
    DateTime? selected;
    await _mountMonth(
      tester,
      month: DateTime(2027, 12),
      stops: [],
      width: 390,
      onMonthChanged: (month) => selected = month,
    );
    await tester.tap(_key('travel-calendar-next'));
    await tester.pumpAndSettle();
    expect(selected, DateTime(2028, 1));
    expect(find.text('January 2028'), findsOneWidget);
    expect(_key('travel-calendar-day-2028-01-31'), findsOneWidget);
    await tester.tap(_key('travel-calendar-previous'));
    await tester.pumpAndSettle();
    expect(selected, DateTime(2027, 12));
    await tester.tap(_key('travel-calendar-today'));
    await tester.pumpAndSettle();
    expect(
      find.text(DateFormat.yMMMM().format(DateTime.now())),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets('month and year picker reaches leap days and empty years', (
    tester,
  ) async {
    DateTime? selected;
    await _mountMonth(
      tester,
      month: DateTime(2027, 12),
      stops: [],
      width: 320,
      scale: 2,
      onMonthChanged: (month) => selected = month,
    );
    await _chooseMonth(tester, year: '2028', month: 'February');
    expect(selected, DateTime(2028, 2));
    expect(find.text('February 2028'), findsOneWidget);
    expect(_key('travel-calendar-day-2028-02-29'), findsOneWidget);
    await _chooseMonth(tester, year: '2027', month: 'February');
    expect(selected, DateTime(2027, 2));
    expect(_key('travel-calendar-day-2027-02-28'), findsOneWidget);
    expect(_key('travel-calendar-day-2027-02-29'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'distinct trip colors stay stable across weeks months and ordering',
    (tester) async {
      final trips = [
        _trip(
          'a',
          'Along the coast',
          'Lisbon',
          start: '2028-02-26',
          end: '2028-03-03',
        ),
        _trip(
          'b',
          'A few days in town',
          'Porto',
          start: '2028-02-26',
          end: '2028-03-03',
        ),
      ];
      await _mountMonth(tester, month: DateTime(2028, 2), stops: trips);
      final firstColor = _ribbonColor(tester, 'a');
      final secondColor = _ribbonColor(tester, 'b');
      expect(firstColor, isNotNull);
      expect(secondColor, isNot(firstColor));
      for (final ribbon in _ribbons('a').evaluate()) {
        final background = find
            .descendant(
              of: find.byWidget(ribbon.widget),
              matching: find.byType(Material),
            )
            .first;
        expect(tester.widget<Material>(background).color, firstColor);
      }
      await tester.tap(_key('travel-calendar-next'));
      await tester.pumpAndSettle();
      expect(_ribbonColor(tester, 'a'), firstColor);
      expect(_ribbonColor(tester, 'b'), secondColor);
      await _mountMonth(
        tester,
        month: DateTime(2028, 2),
        stops: trips.reversed.toList(),
      );
      expect(_ribbonColor(tester, 'a'), firstColor);
      expect(_ribbonColor(tester, 'b'), secondColor);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'approximate and undecided trips stay outside exact day ribbons',
    (tester) async {
      TravelStop? opened;
      await _mountMonth(
        tester,
        month: DateTime(2028, 2),
        stops: [
          _trip('exact', 'A confirmed weekend', 'Porto'),
          _trip(
            'approx',
            'Sometime in February',
            'Lisbon',
            precision: 'approximate',
          ),
          _trip(
            'later',
            'Our next adventure',
            'Kyoto',
            precision: 'tbd',
            start: '',
            end: '',
          ),
        ],
        onStopTap: (stop) => opened = stop,
      );
      expect(_ribbons('exact'), findsWidgets);
      expect(_ribbons('approx'), findsNothing);
      expect(_ribbons('later'), findsNothing);
      expect(find.text('Approximate & unscheduled plans'), findsOneWidget);
      await tester.ensureVisible(_key('travel-calendar-unscheduled'));
      await tester.tap(find.text('Approximate & unscheduled plans'));
      await tester.pumpAndSettle();
      final undecided = find.byWidgetPredicate(
        (widget) => widget is TravelStopCard && widget.stop['id'] == 'later',
      );
      await tester.ensureVisible(undecided);
      await tester.tap(undecided);
      expect(opened?['id'], 'later');
      expect(tester.takeException(), isNull);
    },
  );

  for (final (width, scale) in [(1200.0, 1.0), (390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('month ribbons fit ${width.toInt()}px at ${scale}x text', (
      tester,
    ) async {
      final boundary = await _mountMonth(
        tester,
        month: DateTime(2028, 2),
        width: width,
        scale: scale,
        stops: [
          _trip(
            'long',
            'A slow journey through Portugal',
            'Lisbon & the Alentejo coast',
            start: '2028-02-05',
            end: '2028-02-20',
          ),
          _trip(
            'city',
            'A weekend with friends',
            'Porto, Portugal',
            start: '2028-02-10',
            end: '2028-02-14',
          ),
          _trip(
            'island',
            'Winter sunshine',
            'Madeira, Portugal',
            start: '2028-02-12',
            end: '2028-02-22',
          ),
          _trip(
            'single',
            'Lunch in the old town',
            'Sintra, Portugal',
            start: '2028-02-16',
            end: '2028-02-16',
          ),
        ],
      );
      expect(_ribbons('long'), findsWidgets);
      expect(_ribbons('city'), findsWidgets);
      for (final day in [10, 20, 29]) {
        final date = _key('travel-calendar-day-2028-02-$day');
        final label = find.descendant(of: date, matching: find.text('$day'));
        final paragraph = tester.renderObject<RenderParagraph>(label);
        expect(
          paragraph.didExceedMaxLines,
          isFalse,
          reason: 'Both digits of day $day must fit in the date cell.',
        );
        final glyphs = paragraph.getBoxesForSelection(
          const TextSelection(baseOffset: 0, extentOffset: 2),
        );
        expect(
          glyphs.map((box) => box.top).toSet(),
          hasLength(1),
          reason: 'Day $day must not wrap its second digit out of view.',
        );
      }
      expect(tester.takeException(), isNull);
      await _capture(
        tester,
        boundary,
        '${width.toInt()}${scale == 2 ? '-2x' : ''}',
      );
    });
  }
}
