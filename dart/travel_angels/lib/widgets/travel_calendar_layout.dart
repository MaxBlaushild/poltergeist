/// Pure civil-date layout for the authorized stops supplied by the calendar.
/// UTC here represents a date, never a conversion of a destination instant.
DateTime? calendarCivilDate(dynamic value) {
  if (value is! String) return null;
  final match = RegExp(r'^(\d{4})-(\d{2})-(\d{2})$').firstMatch(value);
  if (match == null) return null;
  final year = int.parse(match[1]!);
  final month = int.parse(match[2]!);
  final day = int.parse(match[3]!);
  final date = DateTime.utc(year, month, day);
  return date.year == year && date.month == month && date.day == day
      ? date
      : null;
}

String calendarDateKey(DateTime date) =>
    '${date.year.toString().padLeft(4, '0')}-'
    '${date.month.toString().padLeft(2, '0')}-'
    '${date.day.toString().padLeft(2, '0')}';

/// Unlike String.hashCode, this remains stable across sessions and platforms.
/// The intermediate value stays below JavaScript's exact-integer limit.
int calendarTripColorIndex(String id) {
  var hash = 0;
  for (final code in id.codeUnits) {
    hash = (hash * 31 + code) % 2147483647;
  }
  return hash % 6;
}

class TravelCalendarDateRange {
  const TravelCalendarDateRange({
    required this.stop,
    required this.id,
    required this.start,
    required this.end,
  });

  final Map<String, dynamic> stop;
  final String id;
  final DateTime start;
  final DateTime end;
  int get colorIndex => calendarTripColorIndex(id);
}

class TravelCalendarRibbon {
  const TravelCalendarRibbon({
    required this.trip,
    required this.startColumn,
    required this.endColumn,
    required this.lane,
    required this.continuesBefore,
    required this.continuesAfter,
  });

  final TravelCalendarDateRange trip;
  final int startColumn;
  final int endColumn;
  final int lane;
  final bool continuesBefore;
  final bool continuesAfter;
  int get span => endColumn - startColumn + 1;
  bool includesColumn(int column) =>
      startColumn <= column && endColumn >= column;
}

class TravelCalendarWeekLayout {
  const TravelCalendarWeekLayout({
    required this.start,
    required this.ribbons,
    required this.laneLimit,
  });

  final DateTime start;
  final List<TravelCalendarRibbon> ribbons;
  final int laneLimit;

  Iterable<TravelCalendarRibbon> get visibleRibbons =>
      ribbons.where((ribbon) => ribbon.lane < laneLimit);

  int get visibleLaneCount => visibleRibbons.fold(
    0,
    (count, ribbon) => ribbon.lane + 1 > count ? ribbon.lane + 1 : count,
  );

  int hiddenCountOnDay(int column) => ribbons
      .where(
        (ribbon) => ribbon.lane >= laneLimit && ribbon.includesColumn(column),
      )
      .length;

  List<Map<String, dynamic>> stopsOnDay(int column) => ribbons
      .where((ribbon) => ribbon.includesColumn(column))
      .map((ribbon) => ribbon.trip.stop)
      .toList(growable: false);

  bool get hasOverflow => ribbons.any((ribbon) => ribbon.lane >= laneLimit);
}

class TravelCalendarMonthLayout {
  const TravelCalendarMonthLayout({
    required this.month,
    required this.weeks,
    required this.unscheduled,
    required this.hasTripsInMonth,
  });

  final DateTime month;
  final List<TravelCalendarWeekLayout> weeks;
  final List<Map<String, dynamic>> unscheduled;
  final bool hasTripsInMonth;
}

TravelCalendarMonthLayout layoutTravelCalendarMonth(
  DateTime month,
  Iterable<Map<String, dynamic>> stops, {
  int laneLimit = 3,
}) {
  if (laneLimit < 1) throw ArgumentError.value(laneLimit, 'laneLimit');
  final first = DateTime.utc(month.year, month.month);
  final last = DateTime.utc(month.year, month.month + 1, 0);
  final gridStart = first.subtract(Duration(days: first.weekday - 1));
  final weekCount = (last.difference(gridStart).inDays + 1 + 6) ~/ 7;
  final gridEnd = gridStart.add(Duration(days: weekCount * 7 - 1));
  final ranges = <TravelCalendarDateRange>[];
  final unscheduled = <Map<String, dynamic>>[];
  for (final stop in stops) {
    final start = calendarCivilDate(stop['startDate']);
    final end = calendarCivilDate(stop['endDate']);
    if (stop['datePrecision'] != 'fixed' ||
        start == null ||
        end == null ||
        end.isBefore(start)) {
      unscheduled.add(stop);
      continue;
    }
    if (start.isAfter(gridEnd) || end.isBefore(gridStart)) continue;
    final id = (stop['id'] as String?)?.trim();
    ranges.add(
      TravelCalendarDateRange(
        stop: stop,
        id: id != null && id.isNotEmpty
            ? id
            : '${calendarDateKey(start)}|${calendarDateKey(end)}|'
                  '${stop['title']}|${stop['destination']}',
        start: start,
        end: end,
      ),
    );
  }
  ranges.sort((a, b) {
    final startOrder = a.start.compareTo(b.start);
    if (startOrder != 0) return startOrder;
    final endOrder = b.end.compareTo(a.end);
    return endOrder != 0 ? endOrder : a.id.compareTo(b.id);
  });
  final seen = <String>{};
  final unique = ranges.where((trip) => seen.add(trip.id)).toList();
  final weeks = <TravelCalendarWeekLayout>[];
  for (var week = 0; week < weekCount; week++) {
    final weekStart = gridStart.add(Duration(days: week * 7));
    final weekEnd = weekStart.add(const Duration(days: 6));
    final laneEnds = <int>[];
    final ribbons = <TravelCalendarRibbon>[];
    for (final trip in unique) {
      if (trip.start.isAfter(weekEnd) || trip.end.isBefore(weekStart)) {
        continue;
      }
      final clippedStart = trip.start.isBefore(weekStart)
          ? weekStart
          : trip.start;
      final clippedEnd = trip.end.isAfter(weekEnd) ? weekEnd : trip.end;
      final startColumn = clippedStart.difference(weekStart).inDays;
      final endColumn = clippedEnd.difference(weekStart).inDays;
      var lane = laneEnds.indexWhere((end) => end < startColumn);
      if (lane == -1) {
        lane = laneEnds.length;
        laneEnds.add(endColumn);
      } else {
        laneEnds[lane] = endColumn;
      }
      ribbons.add(
        TravelCalendarRibbon(
          trip: trip,
          startColumn: startColumn,
          endColumn: endColumn,
          lane: lane,
          continuesBefore: trip.start.isBefore(weekStart),
          continuesAfter: trip.end.isAfter(weekEnd),
        ),
      );
    }
    weeks.add(
      TravelCalendarWeekLayout(
        start: weekStart,
        ribbons: List.unmodifiable(ribbons),
        laneLimit: laneLimit,
      ),
    );
  }
  return TravelCalendarMonthLayout(
    month: first,
    weeks: List.unmodifiable(weeks),
    unscheduled: List.unmodifiable(unscheduled),
    hasTripsInMonth: unique.any(
      (trip) => !trip.start.isAfter(last) && !trip.end.isBefore(first),
    ),
  );
}
