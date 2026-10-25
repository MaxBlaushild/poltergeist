import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_calendar_layout.dart';

typedef TravelStop = Map<String, dynamic>;

class TravelCalendarEyebrow extends StatelessWidget {
  const TravelCalendarEyebrow(this.text, {super.key});
  final String text;

  @override
  Widget build(BuildContext context) => Text(
    text.toUpperCase(),
    style: Theme.of(context).textTheme.labelSmall?.copyWith(
      color: AppTheme.rust,
      letterSpacing: 1.4,
      fontWeight: FontWeight.normal,
      height: 1.5,
    ),
  );
}

class TravelCalendarDialogTitle extends StatelessWidget {
  const TravelCalendarDialogTitle(this.title, {super.key, required this.label});
  final String title;
  final String label;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    mainAxisSize: MainAxisSize.min,
    children: [
      TravelCalendarEyebrow(label),
      const SizedBox(height: 8),
      Text(title, style: Theme.of(context).textTheme.headlineSmall),
    ],
  );
}

class TravelCalendarCallout extends StatelessWidget {
  const TravelCalendarCallout(
    this.text, {
    super.key,
    this.icon = Icons.lock_outline,
  });
  final String text;
  final IconData icon;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(16),
    decoration: const BoxDecoration(
      color: AppTheme.peach,
      border: Border(left: BorderSide(color: AppTheme.rust, width: 2)),
    ),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 18, color: AppTheme.ink),
        const SizedBox(width: 12),
        Expanded(
          child: Text(text, style: Theme.of(context).textTheme.bodyMedium),
        ),
      ],
    ),
  );
}

DateTime? travelDate(dynamic value) {
  if (value is! String || value.isEmpty) return null;
  final date = DateTime.tryParse(value);
  return date == null ? null : DateTime(date.year, date.month, date.day);
}

String travelDateValue(DateTime value) =>
    DateFormat('yyyy-MM-dd').format(value);

String travelDateLabel(TravelStop stop) {
  final start = travelDate(stop['startDate']);
  final end = travelDate(stop['endDate']);
  if (stop['datePrecision'] == 'tbd' || start == null) {
    return 'Dates to be decided';
  }
  final prefix = stop['datePrecision'] == 'approximate' ? 'Around ' : '';
  final startLabel = DateFormat.yMMMd().format(start);
  return '$prefix$startLabel${end == null || end == start ? '' : ' – ${DateFormat.yMMMd().format(end)}'}';
}

String travelVisibilityLabel(String? value) => switch (value) {
  'private' => 'Private',
  'specific' => 'Specific people',
  'link' => 'Anyone with the link',
  _ => 'Inherit calendar',
};

String travelStatusLabel(dynamic value) => switch (value) {
  'needs_reconfirmation' => 'Needs reconfirmation',
  'interested' => 'Interested',
  'requested' => 'Requested',
  'joined' => 'Confirmed',
  'waitlisted' => 'Waitlisted',
  'declined' => 'Declined',
  'withdrawn' => 'Withdrawn',
  'removed' => 'Removed',
  'cancelled' => 'Cancelled',
  'confirmed' => 'Confirmed',
  _ => 'Tentative',
};

/// The API supplies only authorized stops. Filtering never creates busy markers
/// or counts for records the visitor cannot see.
List<TravelStop> filterTravelStops(
  List<TravelStop> stops, {
  String destination = '',
  String period = 'all',
  DateTime? today,
  DateTimeRange? range,
}) {
  final current = today ?? DateTime.now();
  final now = DateTime(current.year, current.month, current.day);
  final query = destination.trim().toLowerCase();
  return stops.where((stop) {
    if (query.isNotEmpty &&
        !'${stop['title'] ?? ''} ${stop['destination'] ?? ''}'
            .toLowerCase()
            .contains(query)) {
      return false;
    }
    final start = travelDate(stop['startDate']);
    final end = travelDate(stop['endDate']) ?? start;
    if (range != null &&
        (start == null ||
            end == null ||
            start.isAfter(range.end) ||
            end.isBefore(range.start))) {
      return false;
    }
    return switch (period) {
      'past' => end != null && end.isBefore(now),
      'current' =>
        start != null &&
            end != null &&
            !start.isAfter(now) &&
            !end.isBefore(now),
      'upcoming' => start != null && start.isAfter(now),
      'unscheduled' => start == null || stop['datePrecision'] != 'fixed',
      _ => true,
    };
  }).toList()..sort((a, b) {
    final first = travelDate(a['startDate']);
    final second = travelDate(b['startDate']);
    if (first == null) return second == null ? 0 : 1;
    if (second == null) return -1;
    return first.compareTo(second);
  });
}

class TravelStopCard extends StatelessWidget {
  const TravelStopCard({
    super.key,
    required this.stop,
    required this.onTap,
    this.selected = false,
  });

  final TravelStop stop;
  final VoidCallback onTap;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final participation = stop['ownParticipation'];
    final start = travelDate(stop['startDate']);
    final end = travelDate(stop['endDate']);
    final fixed = stop['datePrecision'] == 'fixed' && start != null;
    final sameMonth =
        fixed &&
        end != null &&
        start.year == end.year &&
        start.month == end.month;
    return Material(
      color: selected ? AppTheme.peach : AppTheme.paper,
      shape: const Border(bottom: BorderSide(color: AppTheme.outline)),
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 8),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SizedBox(
                width: MediaQuery.sizeOf(context).width < 500 ? 72 : 86,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    TravelCalendarEyebrow(
                      fixed ? DateFormat.MMM().format(start) : 'Dates',
                    ),
                    const SizedBox(height: 4),
                    Text(
                      fixed
                          ? '${start.day}${sameMonth && start != end ? '–${end.day}' : ''}'
                          : stop['datePrecision'] == 'approximate' &&
                                start != null
                          ? '≈ ${start.day}'
                          : 'TBD',
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontSize: MediaQuery.sizeOf(context).width < 500
                            ? 25
                            : 27,
                        height: 1.1,
                      ),
                    ),
                    if (fixed)
                      Text(
                        '${start.year}',
                        style: theme.textTheme.labelSmall?.copyWith(
                          color: AppTheme.muted,
                        ),
                      ),
                    if (selected)
                      const Padding(
                        padding: EdgeInsets.only(top: 10),
                        child: Icon(
                          Icons.check_circle_outline,
                          size: 20,
                          color: AppTheme.rust,
                        ),
                      ),
                  ],
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    TravelCalendarEyebrow(stop['destination'] as String? ?? ''),
                    const SizedBox(height: 6),
                    Text(
                      stop['title'] as String? ?? 'Travel stop',
                      style: theme.textTheme.headlineSmall?.copyWith(
                        height: 1.15,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      travelDateLabel(stop),
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: AppTheme.muted,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 6,
                      runSpacing: 4,
                      children: [
                        _badge(context, travelStatusLabel(stop['status'])),
                        if (stop['canManage'] == true) ...[
                          _badge(
                            context,
                            stop['published'] == true ? 'Published' : 'Draft',
                          ),
                          _badge(
                            context,
                            travelVisibilityLabel(
                              stop['visibility'] as String?,
                            ),
                          ),
                        ],
                        if (participation is Map)
                          _badge(
                            context,
                            travelStatusLabel(participation['status']),
                          ),
                        if (selected)
                          _badge(context, 'Selected · not submitted'),
                      ],
                    ),
                  ],
                ),
              ),
              const Padding(
                padding: EdgeInsets.only(left: 8, top: 4),
                child: Icon(Icons.arrow_outward, size: 19, color: AppTheme.ink),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _badge(BuildContext context, String label) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
    decoration: BoxDecoration(
      color: label == 'Draft' ? AppTheme.peach : Colors.transparent,
      border: Border.all(
        color: label == 'Draft' ? AppTheme.rust : AppTheme.outline,
      ),
      borderRadius: BorderRadius.circular(2),
    ),
    child: Text(
      label,
      style: Theme.of(context).textTheme.labelSmall?.copyWith(
        color: label == 'Draft' ? AppTheme.rust : AppTheme.ink,
        height: 1.4,
      ),
    ),
  );
}

class TravelCalendarMonth extends StatelessWidget {
  const TravelCalendarMonth({
    super.key,
    required this.month,
    required this.stops,
    required this.onMonthChanged,
    required this.onStopTap,
  });

  final DateTime month;
  final List<TravelStop> stops;
  final ValueChanged<DateTime> onMonthChanged;
  final ValueChanged<TravelStop> onStopTap;

  static const _palettes = <(Color, Color)>[
    (Color(0xFFDCE8F2), Color(0xFF1C4E69)),
    (Color(0xFFF6D8C5), Color(0xFFA53B18)),
    (Color(0xFFDDE9D5), Color(0xFF3E684B)),
    (Color(0xFFE8DDF0), Color(0xFF735584)),
    (Color(0xFFF3E7B9), Color(0xFF826323)),
    (Color(0xFFF1DCE3), Color(0xFF8B4D63)),
  ];

  Future<void> _chooseMonth(BuildContext context) async {
    final choice = await showDialog<DateTime>(
      context: context,
      builder: (_) => _TravelCalendarMonthChooser(month: month),
    );
    if (context.mounted && choice != null) {
      onMonthChanged(choice);
    }
  }

  Widget _navigation(BuildContext context, double width) {
    final theme = Theme.of(context);
    final scaler = MediaQuery.textScalerOf(context);
    final headingStyle = theme.textTheme.headlineSmall?.copyWith(
      color: AppTheme.ink,
      fontSize: width < 600 ? 24 : null,
      height: 1.15,
    );
    double textWidth(String text, TextStyle? style) {
      final painter = TextPainter(
        text: TextSpan(text: text, style: style),
        textDirection: Directionality.of(context),
        textScaler: scaler,
        maxLines: 1,
      )..layout();
      final measured = painter.width;
      painter.dispose();
      return measured;
    }

    final headingLabel = DateFormat.yMMMM().format(month);
    final controlsWidth =
        88 +
        6 +
        math.max(44, textWidth('Today', theme.textTheme.labelLarge) + 16);
    final stacked =
        width < 350 ||
        scaler.scale(14) > 21 ||
        textWidth(headingLabel, headingStyle) + 36 + controlsWidth > width;
    final heading = TextButton(
      key: const ValueKey('travel-calendar-month-picker'),
      onPressed: () => _chooseMonth(context),
      style: TextButton.styleFrom(
        alignment: Alignment.centerLeft,
        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 8),
        minimumSize: const Size(44, 44),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Flexible(child: Text(headingLabel, style: headingStyle)),
          const SizedBox(width: 6),
          const Icon(Icons.expand_more, size: 22),
        ],
      ),
    );
    final controls = Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        IconButton(
          key: const ValueKey('travel-calendar-previous'),
          tooltip: 'Previous month',
          constraints: const BoxConstraints.tightFor(width: 44, height: 44),
          onPressed: month.year <= 2000 && month.month == 1
              ? null
              : () => onMonthChanged(DateTime(month.year, month.month - 1)),
          icon: const Icon(Icons.chevron_left),
        ),
        IconButton(
          key: const ValueKey('travel-calendar-next'),
          tooltip: 'Next month',
          constraints: const BoxConstraints.tightFor(width: 44, height: 44),
          onPressed: month.year >= 2100 && month.month == 12
              ? null
              : () => onMonthChanged(DateTime(month.year, month.month + 1)),
          icon: const Icon(Icons.chevron_right),
        ),
        const SizedBox(width: 6),
        TextButton(
          key: const ValueKey('travel-calendar-today'),
          onPressed: () {
            final today = DateTime.now();
            onMonthChanged(DateTime(today.year, today.month));
          },
          style: TextButton.styleFrom(
            minimumSize: const Size(44, 44),
            padding: const EdgeInsets.symmetric(horizontal: 8),
          ),
          child: const Text('Today'),
        ),
      ],
    );
    return stacked
        ? Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              heading,
              Align(alignment: Alignment.centerRight, child: controls),
            ],
          )
        : Row(
            children: [
              Expanded(child: heading),
              controls,
            ],
          );
  }

  @override
  Widget build(BuildContext context) => LayoutBuilder(
    builder: (context, constraints) {
      final compact = constraints.maxWidth < 600;
      final scaler = MediaQuery.textScalerOf(context);
      final layout = layoutTravelCalendarMonth(
        month,
        stops,
        laneLimit: compact ? 2 : 3,
      );
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _navigation(context, constraints.maxWidth),
          const SizedBox(height: 12),
          DecoratedBox(
            decoration: BoxDecoration(
              border: Border.all(color: AppTheme.outline),
              borderRadius: BorderRadius.circular(3),
            ),
            child: Padding(
              padding: const EdgeInsets.all(1),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(2),
                child: Column(
                  children: [
                    _weekdays(
                      context,
                      constraints.maxWidth < 450 || scaler.scale(12) > 18,
                    ),
                    for (final week in layout.weeks)
                      _week(context, week, constraints.maxWidth - 2, compact),
                  ],
                ),
              ),
            ),
          ),
          if (!layout.hasTripsInMonth)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 16),
              child: Text(
                'No trips this month.',
                textAlign: TextAlign.center,
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(color: AppTheme.muted),
              ),
            ),
          if (layout.unscheduled.isNotEmpty) ...[
            const SizedBox(height: 20),
            ExpansionTile(
              key: const ValueKey('travel-calendar-unscheduled'),
              tilePadding: const EdgeInsets.symmetric(
                horizontal: 4,
                vertical: 4,
              ),
              childrenPadding: const EdgeInsets.only(top: 8),
              title: Text(
                'Approximate & unscheduled plans',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              subtitle: const Padding(
                padding: EdgeInsets.only(top: 6),
                child: Text(
                  'These plans do not reserve exact days on the calendar.',
                ),
              ),
              children: [
                for (final stop in layout.unscheduled)
                  TravelStopCard(stop: stop, onTap: () => onStopTap(stop)),
              ],
            ),
          ],
        ],
      );
    },
  );

  Widget _weekdays(BuildContext context, bool short) {
    const names = [
      'Monday',
      'Tuesday',
      'Wednesday',
      'Thursday',
      'Friday',
      'Saturday',
      'Sunday',
    ];
    return Row(
      children: [
        for (final name in names)
          Expanded(
            child: Semantics(
              label: name,
              excludeSemantics: true,
              child: Container(
                alignment: Alignment.center,
                padding: const EdgeInsets.symmetric(vertical: 9),
                decoration: const BoxDecoration(
                  color: AppTheme.peach,
                  border: Border(bottom: BorderSide(color: AppTheme.outline)),
                ),
                child: Text(
                  short ? name[0] : name.substring(0, 3).toUpperCase(),
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    color: AppTheme.ink,
                    fontSize: 11,
                    height: 1.3,
                    letterSpacing: short ? 0 : 1,
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }

  Widget _week(
    BuildContext context,
    TravelCalendarWeekLayout week,
    double width,
    bool compact,
  ) {
    final scaler = MediaQuery.textScalerOf(context);
    final cellWidth = width / 7;
    final dayHeight = math.max(30.0, scaler.scale(12) * 1.3 + 12);
    final ribbonHeight = math.max(
      44.0,
      scaler.scale(12) * 1.25 + scaler.scale(11) * 1.25 + 10,
    );
    final overflowHeight = math.max(44.0, scaler.scale(12) * 1.3 + 12);
    final ribbonBottom = dayHeight + week.visibleLaneCount * (ribbonHeight + 4);
    final height = math.max(
      compact ? 80.0 : 104.0,
      ribbonBottom + (week.hasOverflow ? overflowHeight : 0) + 5,
    );
    final today = DateTime.now();
    return SizedBox(
      height: height,
      child: Stack(
        children: [
          Positioned.fill(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                for (var column = 0; column < 7; column++)
                  Expanded(
                    child: _day(
                      context,
                      week.start.add(Duration(days: column)),
                      today,
                      dayHeight,
                    ),
                  ),
              ],
            ),
          ),
          for (final ribbon in week.visibleRibbons)
            Positioned(
              left:
                  ribbon.startColumn * cellWidth +
                  (cellWidth < 48 || ribbon.continuesBefore ? 0 : 2),
              width:
                  ribbon.span * cellWidth -
                  (cellWidth < 48 || ribbon.continuesBefore ? 0 : 2) -
                  (cellWidth < 48 || ribbon.continuesAfter ? 0 : 2),
              top: dayHeight + ribbon.lane * (ribbonHeight + 4),
              height: ribbonHeight,
              child: _ribbon(context, week, ribbon),
            ),
          if (week.hasOverflow)
            for (var column = 0; column < 7; column++)
              if (week.hiddenCountOnDay(column) > 0)
                Positioned(
                  left: column * cellWidth,
                  width: cellWidth,
                  top: ribbonBottom,
                  height: overflowHeight,
                  child: _overflow(context, week, column, cellWidth < 72),
                ),
        ],
      ),
    );
  }

  Widget _day(
    BuildContext context,
    DateTime date,
    DateTime today,
    double dayHeight,
  ) {
    final isToday =
        date.year == today.year &&
        date.month == today.month &&
        date.day == today.day;
    return Semantics(
      key: ValueKey('travel-calendar-day-${calendarDateKey(date)}'),
      label: DateFormat.yMMMMEEEEd().format(date),
      excludeSemantics: true,
      child: Container(
        alignment: Alignment.topRight,
        decoration: BoxDecoration(
          color: date.month == month.month
              ? AppTheme.paper
              : AppTheme.cream.withValues(alpha: .35),
          border: const Border(
            bottom: BorderSide(color: AppTheme.outline, width: .5),
            right: BorderSide(color: AppTheme.outline, width: .5),
          ),
        ),
        child: SizedBox(
          height: dayHeight,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 2, vertical: 3),
            child: Container(
              alignment: Alignment.center,
              padding: const EdgeInsets.symmetric(horizontal: 2),
              decoration: isToday
                  ? const BoxDecoration(
                      color: AppTheme.ink,
                      shape: BoxShape.circle,
                    )
                  : null,
              child: Text(
                '${date.day}',
                maxLines: 1,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  fontSize: 12,
                  height: 1.3,
                  color: isToday
                      ? AppTheme.paper
                      : date.month == month.month
                      ? AppTheme.ink
                      : AppTheme.muted,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _ribbon(
    BuildContext context,
    TravelCalendarWeekLayout week,
    TravelCalendarRibbon ribbon,
  ) {
    final trip = ribbon.trip;
    final title = (trip.stop['title'] as String?)?.trim();
    final destination = (trip.stop['destination'] as String?)?.trim() ?? '';
    final name = title?.isNotEmpty == true ? title! : 'Untitled trip';
    final description =
        '$name${destination.isEmpty ? '' : ', $destination'}, '
        '${DateFormat.yMMMMd().format(trip.start)} through ${DateFormat.yMMMMd().format(trip.end)}'
        '${ribbon.continuesBefore ? ', continues from previous week' : ''}'
        '${ribbon.continuesAfter ? ', continues into next week' : ''}';
    final palette = _palettes[trip.colorIndex];
    final radius = BorderRadius.horizontal(
      left: Radius.circular(ribbon.continuesBefore ? 0 : 3),
      right: Radius.circular(ribbon.continuesAfter ? 0 : 3),
    );
    return Semantics(
      key: ValueKey(
        'travel-calendar-ribbon-${calendarDateKey(week.start)}-${trip.id}',
      ),
      label: description,
      button: true,
      excludeSemantics: true,
      onTap: () => onStopTap(trip.stop),
      child: Tooltip(
        message: description,
        excludeFromSemantics: true,
        child: Material(
          color: palette.$1,
          borderRadius: radius,
          clipBehavior: Clip.antiAlias,
          child: InkWell(
            onTap: () => onStopTap(trip.stop),
            child: DecoratedBox(
              decoration: BoxDecoration(
                border: Border(
                  left: BorderSide(
                    color: palette.$2,
                    width: ribbon.continuesBefore ? 1 : 3,
                  ),
                ),
              ),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                        fontSize: 12,
                        height: 1.25,
                        color: AppTheme.ink,
                      ),
                    ),
                    if (destination.isNotEmpty)
                      Text(
                        destination,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          fontSize: 11,
                          height: 1.25,
                          color: AppTheme.ink,
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _overflow(
    BuildContext context,
    TravelCalendarWeekLayout week,
    int column,
    bool short,
  ) {
    final date = week.start.add(Duration(days: column));
    final count = week.hiddenCountOnDay(column);
    final label =
        '+$count more trips on ${DateFormat.yMMMMEEEEd().format(date)}';
    void open() => _showDay(context, date, week.stopsOnDay(column));
    return Semantics(
      key: ValueKey('travel-calendar-more-${calendarDateKey(date)}'),
      label: label,
      button: true,
      excludeSemantics: true,
      onTap: open,
      child: Tooltip(
        message: label,
        excludeFromSemantics: true,
        child: TextButton(
          style: TextButton.styleFrom(
            minimumSize: const Size(44, 44),
            padding: const EdgeInsets.symmetric(horizontal: 2),
            foregroundColor: AppTheme.ink,
          ),
          onPressed: open,
          child: Text(
            short ? '+$count' : '+$count more',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(fontSize: 12, height: 1.3),
          ),
        ),
      ),
    );
  }

  void _showDay(BuildContext context, DateTime date, List<TravelStop> matches) {
    showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (sheetContext) => SafeArea(
        child: ConstrainedBox(
          constraints: BoxConstraints(
            maxHeight: MediaQuery.sizeOf(sheetContext).height * .75,
          ),
          child: ListView(
            shrinkWrap: true,
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
            children: [
              Text(
                DateFormat.yMMMMEEEEd().format(date),
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 16),
              for (final stop in matches)
                TravelStopCard(
                  stop: stop,
                  onTap: () {
                    Navigator.pop(sheetContext);
                    onStopTap(stop);
                  },
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TravelCalendarMonthChooser extends StatefulWidget {
  const _TravelCalendarMonthChooser({required this.month});
  final DateTime month;

  @override
  State<_TravelCalendarMonthChooser> createState() =>
      _TravelCalendarMonthChooserState();
}

class _TravelCalendarMonthChooserState
    extends State<_TravelCalendarMonthChooser> {
  late int _year = widget.month.year.clamp(2000, 2100);
  late int _month = widget.month.month;

  @override
  Widget build(BuildContext context) => AlertDialog(
    insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
    title: const Text('Choose month and year'),
    scrollable: true,
    content: SizedBox(
      width: 340,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          DropdownButtonFormField<int>(
            key: const ValueKey('travel-calendar-year-selector'),
            initialValue: _year,
            isExpanded: true,
            menuMaxHeight: 280,
            decoration: const InputDecoration(labelText: 'Year'),
            items: [
              for (var year = 2000; year <= 2100; year++)
                DropdownMenuItem(value: year, child: Text('$year')),
            ],
            onChanged: (value) {
              if (value != null) setState(() => _year = value);
            },
          ),
          const SizedBox(height: 20),
          DropdownButtonFormField<int>(
            key: const ValueKey('travel-calendar-month-selector'),
            initialValue: _month,
            isExpanded: true,
            menuMaxHeight: 280,
            decoration: const InputDecoration(labelText: 'Month'),
            items: [
              for (var month = 1; month <= 12; month++)
                DropdownMenuItem(
                  value: month,
                  child: Text(DateFormat.MMMM().format(DateTime(2000, month))),
                ),
            ],
            onChanged: (value) {
              if (value != null) setState(() => _month = value);
            },
          ),
        ],
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(
        key: const ValueKey('travel-calendar-month-confirm'),
        onPressed: () => Navigator.pop(context, DateTime(_year, _month)),
        child: const Text('Show month'),
      ),
    ],
  );
}
