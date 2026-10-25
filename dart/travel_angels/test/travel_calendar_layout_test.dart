import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/widgets/travel_calendar_layout.dart';

Map<String, dynamic> _stop(String id, String start, String end) => {
  'id': id,
  'title': 'Trip $id',
  'destination': 'Lisbon, Portugal',
  'datePrecision': 'fixed',
  'startDate': start,
  'endDate': end,
};

List<String> _occupiedDates(TravelCalendarMonthLayout layout, String id) => [
  for (final week in layout.weeks)
    for (final ribbon in week.ribbons.where((r) => r.trip.id == id))
      for (
        var column = ribbon.startColumn;
        column <= ribbon.endColumn;
        column++
      )
        calendarDateKey(week.start.add(Duration(days: column))),
];

void main() {
  test(
    'civil dates reject rollover and instants instead of shifting dates',
    () {
      expect(calendarCivilDate('2028-02-29'), DateTime.utc(2028, 2, 29));
      for (final invalid in [
        '2027-02-29',
        '2028-02-30',
        '2028-13-01',
        '2028-00-10',
        '2028-02-00',
        '2028-2-01',
        '2028-02-29T23:00:00-05:00',
        '',
        null,
      ]) {
        expect(calendarCivilDate(invalid), isNull, reason: '$invalid');
      }
    },
  );

  test('months use four, five, or six complete Monday-first weeks', () {
    for (final (month, count) in [
      (DateTime(2021, 2), 4),
      (DateTime(2028, 2), 5),
      (DateTime(2026, 8), 6),
    ]) {
      final layout = layoutTravelCalendarMonth(month, []);
      expect(layout.weeks.length, count);
      expect(layout.weeks.first.start.weekday, DateTime.monday);
      final days = [
        for (final week in layout.weeks)
          for (var column = 0; column < 7; column++)
            week.start.add(Duration(days: column)),
      ];
      expect(days.toSet().length, count * 7);
      expect(
        days.first.isAfter(DateTime.utc(month.year, month.month)),
        isFalse,
      );
      expect(
        days.last.isBefore(DateTime.utc(month.year, month.month + 1, 0)),
        isFalse,
      );
      expect(layout.hasTripsInMonth, isFalse);
    }
  });

  test('inclusive ranges cross leap day, week, and month without gaps', () {
    final layout = layoutTravelCalendarMonth(DateTime(2028, 2), [
      _stop('coast', '2028-02-26', '2028-03-03'),
    ]);
    expect(_occupiedDates(layout, 'coast'), [
      '2028-02-26',
      '2028-02-27',
      '2028-02-28',
      '2028-02-29',
      '2028-03-01',
      '2028-03-02',
      '2028-03-03',
    ]);
    final ribbons = layout.weeks.expand((week) => week.ribbons).toList();
    expect(ribbons.map((r) => r.span), [2, 5]);
    expect(ribbons.first.continuesBefore, isFalse);
    expect(ribbons.first.continuesAfter, isTrue);
    expect(ribbons.last.continuesBefore, isTrue);
    expect(ribbons.last.continuesAfter, isFalse);
  });

  test('civil ranges preserve all days across DST and year boundaries', () {
    for (final (month, start, end, expected) in [
      (
        DateTime(2027, 3),
        '2027-03-13',
        '2027-03-15',
        ['2027-03-13', '2027-03-14', '2027-03-15'],
      ),
      (
        DateTime(2027, 11),
        '2027-11-06',
        '2027-11-08',
        ['2027-11-06', '2027-11-07', '2027-11-08'],
      ),
      (
        DateTime(2028, 1),
        '2027-12-30',
        '2028-01-04',
        [
          '2027-12-30',
          '2027-12-31',
          '2028-01-01',
          '2028-01-02',
          '2028-01-03',
          '2028-01-04',
        ],
      ),
    ]) {
      final layout = layoutTravelCalendarMonth(month, [
        _stop('trip', start, end),
      ]);
      expect(_occupiedDates(layout, 'trip'), expected);
    }
  });

  test(
    'overlap lanes do not collide and day counts include only supplied trips',
    () {
      final stops = [
        _stop('long', '2027-07-05', '2027-07-09'),
        _stop('one-day', '2027-07-06', '2027-07-06'),
        _stop('later', '2027-07-08', '2027-07-10'),
        _stop('weekend', '2027-07-06', '2027-07-08'),
      ];
      final layout = layoutTravelCalendarMonth(
        DateTime(2027, 7),
        stops,
        laneLimit: 2,
      );
      final week = layout.weeks.singleWhere(
        (w) => w.start == DateTime.utc(2027, 7, 5),
      );
      for (var column = 0; column < 7; column++) {
        final occupants = week.ribbons
            .where((r) => r.includesColumn(column))
            .toList();
        expect(occupants.map((r) => r.lane).toSet().length, occupants.length);
        final visible = week.visibleRibbons
            .where((r) => r.includesColumn(column))
            .length;
        expect(
          visible + week.hiddenCountOnDay(column),
          week.stopsOnDay(column).length,
        );
      }
      expect(week.hiddenCountOnDay(1), 1);
      expect(week.stopsOnDay(1).map((s) => s['id']).toSet(), {
        'long',
        'one-day',
        'weekend',
      });
      expect(_occupiedDates(layout, 'one-day'), ['2027-07-06']);
      final restricted = layoutTravelCalendarMonth(DateTime(2027, 7), [
        stops.first,
      ], laneLimit: 2);
      expect(restricted.weeks.every((w) => !w.hasOverflow), isTrue);
      expect(
        restricted.weeks.expand((w) => w.ribbons).map((r) => r.trip.id).toSet(),
        {'long'},
      );
    },
  );

  test(
    'input order and duplicates do not change ribbon lanes or trip colors',
    () {
      final stops = [
        for (final id in ['a', 'b', 'c', 'd', 'e', 'f'])
          _stop(id, '2027-07-05', '2027-07-09'),
      ];
      final first = layoutTravelCalendarMonth(DateTime(2027, 7), stops);
      final second = layoutTravelCalendarMonth(DateTime(2027, 7), [
        ...stops.reversed,
        stops.first,
      ]);
      List<String> signature(TravelCalendarMonthLayout layout) => [
        for (final week in layout.weeks)
          for (final r in week.ribbons)
            '${calendarDateKey(week.start)}:${r.trip.id}:${r.lane}:${r.trip.colorIndex}',
      ];
      expect(signature(second), signature(first));
      expect(
        stops
            .map((s) => calendarTripColorIndex(s['id'] as String))
            .toSet()
            .length,
        6,
      );
    },
  );

  test('uncertain or invalid dates never occupy exact calendar days', () {
    final uncertain = [
      {
        ..._stop('approximate', '2027-07-05', '2027-07-09'),
        'datePrecision': 'approximate',
      },
      {..._stop('tbd', '', ''), 'datePrecision': 'tbd'},
      _stop('reversed', '2027-07-09', '2027-07-05'),
      _stop('invalid', '2027-02-29', '2027-03-01'),
    ];
    final layout = layoutTravelCalendarMonth(DateTime(2027, 7), uncertain);
    expect(layout.weeks.expand((w) => w.ribbons), isEmpty);
    expect(layout.unscheduled.map((s) => s['id']), [
      'approximate',
      'tbd',
      'reversed',
      'invalid',
    ]);
    expect(layout.hasTripsInMonth, isFalse);
  });
}
