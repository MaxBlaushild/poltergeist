import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

typedef TravelStop = Map<String, dynamic>;

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
  'joined' => 'Joined',
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
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              CircleAvatar(
                backgroundColor: theme.colorScheme.primaryContainer,
                child: Icon(selected ? Icons.check : Icons.place_outlined),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      stop['title'] as String? ?? 'Travel stop',
                      style: theme.textTheme.titleMedium,
                    ),
                    const SizedBox(height: 3),
                    Text(stop['destination'] as String? ?? ''),
                    const SizedBox(height: 6),
                    Text(travelDateLabel(stop)),
                    const SizedBox(height: 8),
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
              const Icon(Icons.chevron_right),
            ],
          ),
        ),
      ),
    );
  }

  Widget _badge(BuildContext context, String label) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
    decoration: BoxDecoration(
      color: Theme.of(context).colorScheme.surfaceContainerHighest,
      borderRadius: BorderRadius.circular(6),
    ),
    child: Text(label, style: Theme.of(context).textTheme.labelSmall),
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

  @override
  Widget build(BuildContext context) {
    final first = DateTime(month.year, month.month);
    final start = first.subtract(Duration(days: first.weekday - 1));
    final days = DateTime(month.year, month.month + 1, 0).day;
    final cells = ((first.weekday - 1 + days) / 7).ceil() * 7;
    final fixed = stops.where((stop) => stop['datePrecision'] == 'fixed');
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            IconButton(
              tooltip: 'Previous month',
              onPressed: () =>
                  onMonthChanged(DateTime(month.year, month.month - 1)),
              icon: const Icon(Icons.chevron_left),
            ),
            Expanded(
              child: Text(
                DateFormat.yMMMM().format(month),
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.titleLarge,
              ),
            ),
            IconButton(
              tooltip: 'Next month',
              onPressed: () =>
                  onMonthChanged(DateTime(month.year, month.month + 1)),
              icon: const Icon(Icons.chevron_right),
            ),
            TextButton(
              onPressed: () => onMonthChanged(DateTime.now()),
              child: const Text('Today'),
            ),
          ],
        ),
        Row(
          children: [
            for (final day in ['M', 'T', 'W', 'T', 'F', 'S', 'S'])
              Expanded(child: Center(child: Text(day))),
          ],
        ),
        const SizedBox(height: 8),
        LayoutBuilder(
          builder: (context, constraints) {
            final compact = constraints.maxWidth < 600;
            return GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: cells,
              gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 7,
                mainAxisExtent: compact ? 76 : 120,
              ),
              itemBuilder: (context, index) {
                final date = DateTime(
                  start.year,
                  start.month,
                  start.day + index,
                );
                final matches = fixed.where((stop) {
                  final begins = travelDate(stop['startDate']);
                  final ends = travelDate(stop['endDate']) ?? begins;
                  return begins != null &&
                      ends != null &&
                      !date.isBefore(begins) &&
                      !date.isAfter(ends);
                }).toList();
                final theme = Theme.of(context);
                return Semantics(
                  label: DateFormat.yMMMMEEEEd().format(date),
                  child: Container(
                    margin: const EdgeInsets.all(2),
                    decoration: BoxDecoration(
                      color: date.month == month.month
                          ? theme.colorScheme.surfaceContainerLow
                          : theme.colorScheme.surface,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: InkWell(
                      onTap: matches.isEmpty
                          ? null
                          : () => showModalBottomSheet<void>(
                              context: context,
                              showDragHandle: true,
                              isScrollControlled: true,
                              builder: (sheetContext) => SafeArea(
                                child: ConstrainedBox(
                                  constraints: BoxConstraints(
                                    maxHeight:
                                        MediaQuery.sizeOf(sheetContext).height *
                                        .7,
                                  ),
                                  child: ListView(
                                    shrinkWrap: true,
                                    padding: const EdgeInsets.all(16),
                                    children: [
                                      Text(
                                        DateFormat.yMMMMEEEEd().format(date),
                                        style: theme.textTheme.titleLarge,
                                      ),
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
                            ),
                      child: Padding(
                        padding: const EdgeInsets.all(4),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('${date.day}'),
                            const SizedBox(height: 3),
                            if (matches.isNotEmpty)
                              Text(
                                compact
                                    ? '${matches.length} ${matches.length == 1 ? 'stop' : 'stops'}'
                                    : matches
                                          .take(2)
                                          .map((s) => s['title'])
                                          .join('\n'),
                                maxLines: compact ? 2 : 3,
                                overflow: TextOverflow.ellipsis,
                                style: theme.textTheme.labelSmall?.copyWith(
                                  color: theme.colorScheme.primary,
                                ),
                              ),
                          ],
                        ),
                      ),
                    ),
                  ),
                );
              },
            );
          },
        ),
        if (stops.any((s) => s['datePrecision'] != 'fixed')) ...[
          const SizedBox(height: 20),
          Text(
            'Approximate & unscheduled plans',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const Text('These plans do not reserve exact days on the calendar.'),
          for (final stop in stops.where((s) => s['datePrecision'] != 'fixed'))
            TravelStopCard(stop: stop, onTap: () => onStopTap(stop)),
        ],
      ],
    );
  }
}
