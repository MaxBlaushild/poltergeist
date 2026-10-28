import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/models/travel_destination.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_calendar_layout.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

/// A list only groups records already authorized and filtered by the calendar.
class TravelTripGroups {
  const TravelTripGroups({
    required this.upcoming,
    required this.ongoing,
    required this.undated,
    required this.past,
  });

  final List<TravelStop> upcoming;
  final List<TravelStop> ongoing;
  final List<TravelStop> undated;
  final List<TravelStop> past;
}

class _DatedTrip {
  _DatedTrip(this.stop, this.inputIndex) {
    if (stop['datePrecision'] == 'tbd') return;
    final first = calendarCivilDate(stop['startDate']);
    final rawEnd = stop['endDate'];
    final last = rawEnd == null || rawEnd == ''
        ? first
        : calendarCivilDate(rawEnd);
    if (first == null || last == null || last.isBefore(first)) return;
    start = first;
    end = last;
  }

  final TravelStop stop;
  final int inputIndex;
  DateTime? start;
  DateTime? end;
}

/// Uses calendar dates rather than elapsed hours, including around DST and
/// midnight. Input order and maps are untouched; equal dates have stable ties.
TravelTripGroups groupTravelTrips(List<TravelStop> stops, {DateTime? today}) {
  final current = today ?? DateTime.now();
  final date = DateTime.utc(current.year, current.month, current.day);
  final upcoming = <_DatedTrip>[];
  final ongoing = <_DatedTrip>[];
  final undated = <_DatedTrip>[];
  final past = <_DatedTrip>[];
  for (var index = 0; index < stops.length; index++) {
    final trip = _DatedTrip(stops[index], index);
    if (trip.start == null) {
      undated.add(trip);
    } else if (!trip.start!.isBefore(date)) {
      upcoming.add(trip);
    } else if (!trip.end!.isBefore(date)) {
      ongoing.add(trip);
    } else {
      past.add(trip);
    }
  }

  int tie(_DatedTrip a, _DatedTrip b) {
    final title = '${a.stop['title'] ?? ''}'.toLowerCase().compareTo(
      '${b.stop['title'] ?? ''}'.toLowerCase(),
    );
    if (title != 0) return title;
    final id = '${a.stop['id'] ?? ''}'.compareTo('${b.stop['id'] ?? ''}');
    return id != 0 ? id : a.inputIndex.compareTo(b.inputIndex);
  }

  int chronological(_DatedTrip a, _DatedTrip b) {
    final start = a.start!.compareTo(b.start!);
    if (start != 0) return start;
    final end = a.end!.compareTo(b.end!);
    return end != 0 ? end : tie(a, b);
  }

  upcoming.sort(chronological);
  ongoing.sort(chronological);
  undated.sort(tie);
  past.sort((a, b) {
    final end = b.end!.compareTo(a.end!);
    if (end != 0) return end;
    final start = b.start!.compareTo(a.start!);
    return start != 0 ? start : tie(a, b);
  });
  List<TravelStop> records(List<_DatedTrip> items) =>
      List.unmodifiable(items.map((item) => item.stop));
  return TravelTripGroups(
    upcoming: records(upcoming),
    ongoing: records(ongoing),
    undated: records(undated),
    past: records(past),
  );
}

String _tripListDateLabel(TravelStop stop) {
  final trip = _DatedTrip(stop, 0);
  final start = trip.start;
  final end = trip.end;
  if (start == null || end == null) return 'Dates to be decided';
  final prefix = stop['datePrecision'] == 'approximate' ? 'Around ' : '';
  if (start == end) return '$prefix${DateFormat.yMMMd().format(start)}';
  if (start.year != end.year) {
    return '$prefix${DateFormat.yMMMd().format(start)} – '
        '${DateFormat.yMMMd().format(end)}';
  }
  final last = start.month == end.month
      ? '${end.day}'
      : DateFormat.MMMd().format(end);
  return '$prefix${DateFormat.MMMd().format(start)}–$last, ${end.year}';
}

/// An agenda of the entire supplied collection, independent of the month grid.
/// The surrounding page owns scrolling and the current filters.
class TravelTripList extends StatefulWidget {
  const TravelTripList({
    super.key,
    required this.stops,
    required this.onStopTap,
    this.selectedIds = const {},
    this.today,
    this.expandPast = false,
  });

  final List<TravelStop> stops;
  final ValueChanged<TravelStop> onStopTap;
  final Set<String> selectedIds;
  final DateTime? today;
  final bool expandPast;

  @override
  State<TravelTripList> createState() => _TravelTripListState();
}

class _TravelTripListState extends State<TravelTripList> {
  late bool _pastExpanded = widget.expandPast;

  @override
  void didUpdateWidget(covariant TravelTripList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.expandPast != widget.expandPast) {
      _pastExpanded = widget.expandPast;
    }
  }

  Widget _heading(String label) => Padding(
    padding: const EdgeInsets.fromLTRB(0, 20, 0, 8),
    child: TravelCalendarEyebrow(label),
  );

  Widget _row(TravelStop stop, {bool ongoing = false}) => TravelTripListRow(
    key: ValueKey('travel-trip-list-row-${stop['id']}'),
    stop: stop,
    selected: widget.selectedIds.contains(stop['id']),
    ongoing: ongoing,
    onTap: () => widget.onStopTap(stop),
  );

  @override
  Widget build(BuildContext context) {
    final groups = groupTravelTrips(widget.stops, today: widget.today);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (groups.upcoming.isNotEmpty) ...[
          _heading('Upcoming trips'),
          for (final stop in groups.upcoming) _row(stop),
        ],
        if (groups.ongoing.isNotEmpty) ...[
          _heading('Happening now'),
          for (final stop in groups.ongoing) _row(stop, ongoing: true),
        ],
        if (groups.undated.isNotEmpty) ...[
          _heading('Dates to come'),
          for (final stop in groups.undated) _row(stop),
        ],
        if (groups.past.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.only(top: 16),
            child: Semantics(
              expanded: _pastExpanded,
              child: TextButton(
                key: const ValueKey('travel-trip-list-past-toggle'),
                onPressed: () => setState(() => _pastExpanded = !_pastExpanded),
                style: TextButton.styleFrom(
                  alignment: Alignment.centerLeft,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Past trips · ${groups.past.length}',
                        style: Theme.of(
                          context,
                        ).textTheme.labelLarge?.copyWith(color: AppTheme.muted),
                      ),
                    ),
                    Icon(
                      _pastExpanded ? Icons.expand_less : Icons.expand_more,
                      size: 20,
                      color: AppTheme.muted,
                    ),
                  ],
                ),
              ),
            ),
          ),
          if (_pastExpanded)
            for (final stop in groups.past) _row(stop),
        ],
      ],
    );
  }
}

class TravelTripListRow extends StatelessWidget {
  const TravelTripListRow({
    super.key,
    required this.stop,
    required this.onTap,
    this.selected = false,
    this.ongoing = false,
  });

  final TravelStop stop;
  final VoidCallback onTap;
  final bool selected;
  final bool ongoing;

  // Matches the six ribbon accents in TravelCalendarMonth.
  static const _accents = [
    Color(0xFF1C4E69),
    Color(0xFFA53B18),
    Color(0xFF3E684B),
    Color(0xFF735584),
    Color(0xFF826323),
    Color(0xFF8B4D63),
  ];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final title = (stop['title'] as String?)?.trim();
    final destination = travelDestinationLabel(stop).replaceAll(' → ', ' · ');
    final participation = stop['ownParticipation'];
    final status = <String>[
      if (ongoing) 'Happening now',
      if (stop['canManage'] == true && stop['published'] != true)
        'Draft'
      else if (stop['status'] == 'cancelled')
        'Cancelled'
      else if (participation is Map)
        travelStatusLabel(participation['status'])
      else if (stop['status'] != 'confirmed')
        travelStatusLabel(stop['status']),
      if (selected) 'Selected · not submitted',
    ];
    final color = _accents[calendarTripColorIndex('${stop['id'] ?? title}')];
    return Semantics(
      selected: selected,
      child: Material(
        color: selected ? AppTheme.peach : AppTheme.paper,
        shape: const Border(bottom: BorderSide(color: AppTheme.outline)),
        child: InkWell(
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 15),
            child: IntrinsicHeight(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  SizedBox(width: 3, child: ColoredBox(color: color)),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(
                              child: Text(
                                title == null || title.isEmpty
                                    ? 'Travel stop'
                                    : title,
                                style: theme.textTheme.titleLarge?.copyWith(
                                  fontSize: 22,
                                  height: 1.15,
                                ),
                              ),
                            ),
                            Padding(
                              padding: const EdgeInsets.only(
                                left: 8,
                                top: 3,
                                right: 2,
                              ),
                              child: Icon(
                                selected
                                    ? Icons.check_circle_outline
                                    : Icons.chevron_right,
                                size: 19,
                                color: selected
                                    ? AppTheme.rust
                                    : AppTheme.muted,
                              ),
                            ),
                          ],
                        ),
                        if (destination.isNotEmpty) ...[
                          const SizedBox(height: 5),
                          Text(
                            destination,
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: AppTheme.muted,
                              height: 1.3,
                            ),
                          ),
                        ],
                        const SizedBox(height: 8),
                        Text(
                          _tripListDateLabel(stop),
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: AppTheme.ink,
                            height: 1.3,
                          ),
                        ),
                        if (status.isNotEmpty) ...[
                          const SizedBox(height: 5),
                          Text(
                            status.join(' · '),
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: ongoing ? AppTheme.success : AppTheme.rust,
                              height: 1.4,
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
