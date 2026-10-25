import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

String tripLogisticsAudience(dynamic visibility) => switch (visibility) {
  'viewers' => 'Everyone who can view this trip',
  'hosts' => 'Hosts only',
  _ => 'Confirmed attendees and hosts',
};

class TravelTripLodgingCard extends StatelessWidget {
  const TravelTripLodgingCard({super.key, required this.entry, this.actions});
  final Map<String, dynamic> entry;
  final Widget? actions;
  @override
  Widget build(BuildContext context) {
    final details = tripMap(entry['privateDetails']);
    final cost = tripMap(entry['cost']);
    final checkIn = DateTime.tryParse(entry['checkInDate'] as String? ?? '');
    final checkOut = DateTime.tryParse(entry['checkOutDate'] as String? ?? '');
    final nights = checkIn != null && checkOut != null
        ? checkOut.difference(checkIn).inDays
        : null;
    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      padding: const EdgeInsets.all(22),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: tripSand),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: 8,
            runSpacing: 6,
            children: [
              TripPill(
                entry['kind'] == 'recommendation'
                    ? 'Recommended stay'
                    : 'Group stay',
                icon: Icons.bed_outlined,
              ),
              TripPill(
                '${entry['status'] ?? 'considering'}',
                color: tripMuted,
                background: tripPaper,
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            entry['name'] as String? ?? 'A place to stay',
            style: const TextStyle(
              color: tripInk,
              fontSize: 22,
              fontWeight: FontWeight.w600,
            ),
          ),
          if ((entry['generalLocation'] as String?)?.isNotEmpty == true)
            Padding(
              padding: const EdgeInsets.only(top: 6),
              child: Text(
                entry['generalLocation'] as String,
                style: const TextStyle(color: tripMuted),
              ),
            ),
          const SizedBox(height: 14),
          Text(
            entry['datesTbd'] == true || checkIn == null
                ? 'Stay dates to be decided'
                : '${DateFormat.MMMd().format(checkIn)} → ${checkOut == null ? 'TBD' : DateFormat.yMMMd().format(checkOut)}${nights != null && nights > 0 ? ' · $nights ${nights == 1 ? 'night' : 'nights'}' : ''}',
            style: const TextStyle(color: tripInk, fontWeight: FontWeight.w600),
          ),
          if ((entry['checkInTime'] as String?)?.isNotEmpty == true ||
              (entry['checkOutTime'] as String?)?.isNotEmpty == true)
            Text(
              'Check-in ${entry['checkInTime'] ?? 'TBD'} · Checkout ${entry['checkOutTime'] ?? 'TBD'} · ${entry['timeZone'] ?? 'UTC'}',
              style: const TextStyle(color: tripMuted),
            ),
          if ((entry['description'] as String?)?.isNotEmpty == true)
            Padding(
              padding: const EdgeInsets.only(top: 14),
              child: TripDescription(entry['description'] as String),
            ),
          if (cost.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Text(
                'Estimate: ${cost['currency']} ${cost['amount']} ${switch (cost['basis']) {
                  'night' => 'per night',
                  'person' => 'per person',
                  _ => 'total',
                }}',
                style: const TextStyle(color: tripMuted),
              ),
            ),
          if ((entry['roomsOrBeds'] as String?)?.isNotEmpty == true)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Text(entry['roomsOrBeds'] as String),
            ),
          if (details.isNotEmpty &&
              ((details['address'] as String?)?.isNotEmpty == true ||
                  (details['arrivalInstructions'] as String?)?.isNotEmpty ==
                      true)) ...[
            const Divider(height: 28),
            const Text(
              'Arrival details',
              style: TextStyle(color: tripInk, fontWeight: FontWeight.w600),
            ),
            if ((details['address'] as String?)?.isNotEmpty == true)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(details['address'] as String),
              ),
            if ((details['arrivalInstructions'] as String?)?.isNotEmpty == true)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(details['arrivalInstructions'] as String),
              ),
          ],
          if ((entry['url'] as String?)?.isNotEmpty == true)
            TextButton.icon(
              onPressed: () => openTripLink(context, entry['url'] as String),
              icon: const Icon(Icons.open_in_new, size: 16),
              label: const Text('View property'),
            ),
          if ((entry['hostNotes'] as String?)?.isNotEmpty == true)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Text(
                'Host notes: ${entry['hostNotes']}',
                style: const TextStyle(color: tripMuted),
              ),
            ),
          if (actions != null) ...[const Divider(height: 24), actions!],
        ],
      ),
    );
  }
}

class TravelTripFlightCard extends StatelessWidget {
  const TravelTripFlightCard({super.key, required this.entry, this.actions});
  final Map<String, dynamic> entry;
  final Widget? actions;
  String _time(Map<String, dynamic> leg, String prefix) {
    final date = DateTime.tryParse(leg['${prefix}Date'] as String? ?? '');
    return '${date == null ? 'Date TBD' : DateFormat.MMMd().format(date)}${(leg['${prefix}Time'] as String?)?.isNotEmpty == true ? ' · ${leg['${prefix}Time']}' : ' · Time TBD'}';
  }

  @override
  Widget build(BuildContext context) => Container(
    margin: const EdgeInsets.only(bottom: 14),
    padding: const EdgeInsets.all(22),
    decoration: BoxDecoration(
      color: Colors.white,
      borderRadius: BorderRadius.circular(18),
      border: Border.all(color: tripSand),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 8,
          runSpacing: 6,
          children: [
            TripPill(
              entry['kind'] == 'recommendation'
                  ? 'Recommended flight'
                  : 'Group flight',
              icon: Icons.flight_outlined,
            ),
            TripPill(
              '${entry['status'] ?? 'suggested'}',
              color: tripMuted,
              background: tripPaper,
            ),
          ],
        ),
        const SizedBox(height: 16),
        Text(
          entry['label'] as String? ?? 'Getting there',
          style: const TextStyle(
            color: tripInk,
            fontSize: 22,
            fontWeight: FontWeight.w600,
          ),
        ),
        for (final leg in tripMaps(entry['legs']))
          Padding(
            padding: const EdgeInsets.only(top: 20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '${leg['airline'] ?? ''} ${leg['flightNumber'] ?? ''}'
                          .trim()
                          .isEmpty
                      ? 'Flight to be decided'
                      : '${leg['airline'] ?? ''} ${leg['flightNumber'] ?? ''}',
                  style: const TextStyle(color: tripMuted, fontSize: 12),
                ),
                const SizedBox(height: 8),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            (leg['departureAirport'] as String?)?.isNotEmpty ==
                                    true
                                ? leg['departureAirport'] as String
                                : 'Departure TBD',
                            style: const TextStyle(
                              color: tripInk,
                              fontSize: 19,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          Text(_time(leg, 'departure')),
                          Text(
                            leg['departureTimeZone'] as String? ?? '',
                            style: const TextStyle(
                              color: tripMuted,
                              fontSize: 11,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Padding(
                      padding: EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 8,
                      ),
                      child: Icon(
                        Icons.arrow_forward,
                        color: tripOrange,
                        size: 20,
                      ),
                    ),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            (leg['arrivalAirport'] as String?)?.isNotEmpty ==
                                    true
                                ? leg['arrivalAirport'] as String
                                : 'Arrival TBD',
                            style: const TextStyle(
                              color: tripInk,
                              fontSize: 19,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          Text(_time(leg, 'arrival')),
                          Text(
                            leg['arrivalTimeZone'] as String? ?? '',
                            style: const TextStyle(
                              color: tripMuted,
                              fontSize: 11,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        if ((entry['notes'] as String?)?.isNotEmpty == true)
          Padding(
            padding: const EdgeInsets.only(top: 18),
            child: TripDescription(entry['notes'] as String),
          ),
        if ((entry['url'] as String?)?.isNotEmpty == true)
          TextButton.icon(
            onPressed: () => openTripLink(context, entry['url'] as String),
            icon: const Icon(Icons.open_in_new, size: 16),
            label: const Text('View flight details'),
          ),
        const Padding(
          padding: EdgeInsets.only(top: 10),
          child: Text(
            'Times are local to each airport. Travel plans do not confirm ticket purchases.',
            style: TextStyle(color: tripMuted, fontSize: 12),
          ),
        ),
        if ((entry['hostNotes'] as String?)?.isNotEmpty == true)
          Padding(
            padding: const EdgeInsets.only(top: 10),
            child: Text(
              'Host notes: ${entry['hostNotes']}',
              style: const TextStyle(color: tripMuted),
            ),
          ),
        if (actions != null) ...[const Divider(height: 24), actions!],
      ],
    ),
  );
}

class TravelTripLogisticsAudienceDialog extends StatefulWidget {
  const TravelTripLogisticsAudienceDialog({
    super.key,
    required this.stop,
    required this.entry,
    required this.kind,
    required this.service,
  });
  final Map<String, dynamic> stop, entry;
  final String kind;
  final TravelCalendarService service;
  @override
  State<TravelTripLogisticsAudienceDialog> createState() =>
      _TravelTripLogisticsAudienceDialogState();
}

class _TravelTripLogisticsAudienceDialogState
    extends State<TravelTripLogisticsAudienceDialog> {
  late String _visibility, _privateVisibility;
  bool _busy = false;
  String? _error;
  @override
  void initState() {
    super.initState();
    _visibility = widget.entry['visibility'] as String? ?? 'attendees';
    _privateVisibility =
        tripMap(widget.entry['privateDetails'])['visibility'] as String? ??
        'attendees';
  }

  Future<void> _save() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.put(
        '/calendar/stops/${widget.stop['id']}/logistics/permissions',
        data: {
          'revision': widget.stop['revision'],
          'kind': widget.kind,
          'id': widget.entry['id'],
          'visibility': _visibility,
          if (widget.kind == 'lodging') 'privateVisibility': _privateVisibility,
        },
      );
      if (mounted) Navigator.pop(context, true);
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => TripFormDialog(
    title: 'Who can see this ${widget.kind}?',
    busy: _busy,
    error: _error,
    onSave: _save,
    saveLabel: 'Apply audience now',
    children: [
      const Text(
        'Permissions change immediately in the current published plan. Content drafts stay unpublished. All viewers must still have access to the trip.',
      ),
      const SizedBox(height: 16),
      DropdownButtonFormField<String>(
        initialValue: _visibility,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Entry audience'),
        items: [
          for (final value in ['viewers', 'attendees', 'hosts'])
            DropdownMenuItem(
              value: value,
              child: Text(tripLogisticsAudience(value)),
            ),
        ],
        onChanged: (value) => setState(() => _visibility = value!),
      ),
      if (widget.kind == 'lodging') ...[
        const SizedBox(height: 16),
        DropdownButtonFormField<String>(
          initialValue: _privateVisibility,
          isExpanded: true,
          decoration: const InputDecoration(
            labelText: 'Exact address & arrival instructions',
          ),
          items: [
            for (final value in ['attendees', 'hosts'])
              DropdownMenuItem(
                value: value,
                child: Text(tripLogisticsAudience(value)),
              ),
          ],
          onChanged: (value) => setState(() => _privateVisibility = value!),
        ),
      ],
      const SizedBox(height: 16),
      const Text(
        'Named traveler assignments remain visible only to hosts and the person named. Host notes stay host-only.',
      ),
    ],
  );
}

class TravelTripLogisticsEditor extends StatefulWidget {
  const TravelTripLogisticsEditor({
    super.key,
    required this.stop,
    required this.kind,
    required this.service,
    this.entry,
  });
  final Map<String, dynamic> stop;
  final Map<String, dynamic>? entry;
  final String kind;
  final TravelCalendarService service;
  @override
  State<TravelTripLogisticsEditor> createState() =>
      _TravelTripLogisticsEditorState();
}

class _TravelTripLogisticsEditorState extends State<TravelTripLogisticsEditor> {
  final _form = GlobalKey<FormState>();
  late Map<String, dynamic> _entry;
  List<Map<String, dynamic>> _travelers = [];
  bool _busy = false;
  String? _error;
  bool get _lodging => widget.kind == 'lodging';
  @override
  void initState() {
    super.initState();
    _entry = tripCopy(
      widget.entry ??
          (_lodging
              ? {
                  'name': '',
                  'type': 'hotel',
                  'kind': 'group',
                  'visibility': 'attendees',
                  'status': 'considering',
                  'datesTbd': true,
                  'timeZone': widget.stop['timeZone'] ?? 'UTC',
                  'privateDetails': {'visibility': 'attendees'},
                  'cost': {'currency': 'USD', 'basis': 'night'},
                  'assignedUserIds': <String>[],
                }
              : {
                  'label': '',
                  'kind': 'group',
                  'visibility': 'attendees',
                  'status': 'suggested',
                  'legs': [
                    <String, dynamic>{
                      'departureTimeZone': 'UTC',
                      'arrivalTimeZone': 'UTC',
                    },
                  ],
                  'assignedUserIds': <String>[],
                }),
    );
    _loadTravelers();
  }

  Future<void> _loadTravelers() async {
    try {
      final result = await widget.service.get(
        '/calendar/stops/${widget.stop['id']}/participations',
      );
      if (mounted) {
        setState(
          () => _travelers = tripMaps(result['participations'])
              .where((p) => p['status'] == 'joined' || p['reserved'] == true)
              .toList(),
        );
      }
    } catch (_) {}
  }

  bool get _outside {
    final draft = tripDraft(widget.stop);
    final start = draft['startDate'] as String? ?? '';
    final end = draft['endDate'] as String? ?? '';
    if (start.isEmpty || end.isEmpty) return false;
    final dates = _lodging
        ? [_entry['checkInDate'], _entry['checkOutDate']]
        : tripMaps(_entry['legs'])
              .expand((leg) => [leg['departureDate'], leg['arrivalDate']])
              .toList();
    return dates
        .whereType<String>()
        .where((date) => date.isNotEmpty)
        .any((date) => date.compareTo(start) < 0 || date.compareTo(end) > 0);
  }

  Widget _text(
    Map<String, dynamic> record,
    String key,
    String label, {
    String? hint,
    int lines = 1,
    bool required = false,
    TextInputType? keyboard,
  }) => Padding(
    padding: const EdgeInsets.only(bottom: 14),
    child: TextFormField(
      key: ValueKey('${identityHashCode(record)}:$key'),
      initialValue: record[key]?.toString() ?? '',
      decoration: InputDecoration(labelText: label, hintText: hint),
      minLines: lines,
      maxLines: lines == 1 ? 1 : lines + 2,
      keyboardType: keyboard,
      validator: required
          ? (value) =>
                value == null || value.trim().isEmpty ? 'Enter $label.' : null
          : null,
      onChanged: (value) => setState(() => record[key] = value),
    ),
  );
  Widget _select(
    Map<String, dynamic> record,
    String key,
    String label,
    Map<String, String> choices,
  ) => Padding(
    padding: const EdgeInsets.only(bottom: 14),
    child: DropdownButtonFormField<String>(
      initialValue: choices.containsKey(record[key])
          ? record[key] as String
          : choices.keys.first,
      isExpanded: true,
      decoration: InputDecoration(labelText: label),
      items: choices.entries
          .map(
            (entry) =>
                DropdownMenuItem(value: entry.key, child: Text(entry.value)),
          )
          .toList(),
      onChanged: (value) => setState(() => record[key] = value),
    ),
  );
  Widget _heading(String title) => Padding(
    padding: const EdgeInsets.only(top: 12, bottom: 16),
    child: Text(
      title,
      style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 17),
    ),
  );
  Future<void> _save() async {
    if (!_form.currentState!.validate()) return;
    if (_outside && _entry['outsideTripDatesConfirmed'] != true) {
      setState(
        () => _error =
            'Review the dates outside this trip and acknowledge them below.',
      );
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final entry = tripCopy(_entry);
      if (_lodging &&
          tripMap(
            entry['cost'],
          )['amount'].toString().trim().replaceAll('null', '').isEmpty) {
        entry.remove('cost');
      }
      final key = _lodging ? 'lodging' : 'flights';
      final entries = tripMaps(tripDraft(widget.stop)[key]);
      final index = widget.entry == null
          ? -1
          : entries.indexWhere((value) => value['id'] == widget.entry!['id']);
      if (index < 0) {
        entries.add(entry);
      } else {
        entries[index] = entry;
      }
      await saveTripDraft(widget.service, widget.stop, {key: entries});
      if (mounted) Navigator.pop(context, true);
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final details = _entry['privateDetails'] is Map
        ? _entry['privateDetails'] as Map<String, dynamic>
        : <String, dynamic>{'visibility': 'attendees'};
    if (_lodging) _entry['privateDetails'] = details;
    final cost = _entry['cost'] is Map
        ? _entry['cost'] as Map<String, dynamic>
        : <String, dynamic>{'currency': 'USD', 'basis': 'night'};
    if (_lodging) _entry['cost'] = cost;
    final legs = _entry['legs'] is List
        ? (_entry['legs'] as List).cast<Map<String, dynamic>>()
        : <Map<String, dynamic>>[];
    final assigned = (_entry['assignedUserIds'] as List? ?? []).cast<String>();
    return TripFormDialog(
      title:
          '${widget.entry == null ? 'Add' : 'Edit'} ${_lodging ? 'lodging' : 'flight'}',
      busy: _busy,
      error: _error,
      onSave: _save,
      saveLabel: 'Save logistics draft',
      children: [
        Form(
          key: _form,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Save a draft, then publish when the plan is ready. These are travel plans, not reservations. Do not include confirmation numbers, tickets, passport or payment details.',
              ),
              const SizedBox(height: 20),
              _text(
                _entry,
                _lodging ? 'name' : 'label',
                _lodging ? 'Property name' : 'Itinerary name',
                required: true,
              ),
              _select(_entry, 'kind', 'Plan type', {
                'group': _lodging ? 'Group stay' : 'Group flight',
                'recommendation': _lodging
                    ? 'Recommended option'
                    : 'Recommended flight',
              }),
              _select(
                _entry,
                'status',
                'Planning status',
                _lodging
                    ? {
                        'considering': 'Considering',
                        'selected': 'Selected',
                        'booked': 'Booked',
                        'cancelled': 'Cancelled',
                      }
                    : {
                        'suggested': 'Suggested',
                        'planned': 'Planned',
                        'booked': 'Booked',
                        'cancelled': 'Cancelled',
                      },
              ),
              if (widget.entry == null)
                _select(_entry, 'visibility', 'Entry audience', {
                  for (final audience in ['attendees', 'viewers', 'hosts'])
                    audience: tripLogisticsAudience(audience),
                })
              else
                Padding(
                  padding: const EdgeInsets.only(bottom: 16),
                  child: Text(
                    'Audience: ${tripLogisticsAudience(_entry['visibility'])}. Use the entry’s audience control to change access immediately.',
                  ),
                ),
              if (_lodging) ...[
                _select(_entry, 'type', 'Accommodation type', {
                  'hotel': 'Hotel',
                  'rental': 'Vacation rental',
                  'hostel': 'Hostel',
                  'home': 'Private home',
                  'other': 'Other',
                }),
                _text(
                  _entry,
                  'generalLocation',
                  'General location',
                  hint: 'Neighborhood or city; exact address goes below',
                ),
                SwitchListTile.adaptive(
                  contentPadding: EdgeInsets.zero,
                  value: _entry['datesTbd'] != false,
                  title: const Text('Stay dates to be decided'),
                  onChanged: (value) =>
                      setState(() => _entry['datesTbd'] = value),
                ),
                if (_entry['datesTbd'] == false) ...[
                  _text(
                    _entry,
                    'checkInDate',
                    'Check-in date',
                    hint: 'YYYY-MM-DD',
                    required: true,
                  ),
                  _text(
                    _entry,
                    'checkOutDate',
                    'Checkout date',
                    hint: 'YYYY-MM-DD',
                    required: true,
                  ),
                  _text(
                    _entry,
                    'checkInTime',
                    'Check-in time (optional)',
                    hint: 'HH:mm',
                  ),
                  _text(
                    _entry,
                    'checkOutTime',
                    'Checkout time (optional)',
                    hint: 'HH:mm',
                  ),
                ],
                _text(
                  _entry,
                  'timeZone',
                  'Property time zone',
                  hint: 'Europe/Lisbon',
                  required: true,
                ),
                _text(_entry, 'description', 'About this stay', lines: 3),
                _text(
                  _entry,
                  'roomsOrBeds',
                  'Rooms or beds (optional)',
                  hint: 'For planning only; this does not reserve rooms',
                ),
                _heading('Optional indicative cost'),
                _text(
                  cost,
                  'amount',
                  'Amount',
                  keyboard: const TextInputType.numberWithOptions(
                    decimal: true,
                  ),
                ),
                _text(cost, 'currency', 'Currency', hint: 'USD'),
                _select(cost, 'basis', 'Cost basis', {
                  'night': 'Per night',
                  'person': 'Per person',
                  'total': 'Total',
                }),
                _heading('Restricted arrival details'),
                if (widget.entry == null)
                  _select(details, 'visibility', 'Address & arrival audience', {
                    'attendees': 'Confirmed attendees and hosts',
                    'hosts': 'Hosts only',
                  }),
                _text(details, 'address', 'Exact street address'),
                _text(
                  details,
                  'arrivalInstructions',
                  'Arrival instructions',
                  hint: 'No door codes, booking references or other secrets',
                  lines: 3,
                ),
              ] else ...[
                _heading('Flight legs'),
                const Text(
                  'Use local dates and times at each airport. Suggested routes can leave unknown details blank.',
                ),
                for (var index = 0; index < legs.length; index++)
                  Container(
                    margin: const EdgeInsets.only(top: 16, bottom: 8),
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(
                      border: Border.all(color: tripSand),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                'Leg ${index + 1}',
                                style: const TextStyle(
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ),
                            if (legs.length > 1)
                              IconButton(
                                tooltip: 'Remove flight leg',
                                onPressed: () =>
                                    setState(() => legs.removeAt(index)),
                                icon: const Icon(Icons.close),
                              ),
                          ],
                        ),
                        _text(legs[index], 'airline', 'Airline'),
                        _text(legs[index], 'flightNumber', 'Flight number'),
                        _text(
                          legs[index],
                          'departureAirport',
                          'Departure airport',
                          hint: 'JFK · New York',
                        ),
                        _text(
                          legs[index],
                          'departureDate',
                          'Departure date',
                          hint: 'YYYY-MM-DD',
                        ),
                        _text(
                          legs[index],
                          'departureTime',
                          'Departure time',
                          hint: 'HH:mm',
                        ),
                        _text(
                          legs[index],
                          'departureTimeZone',
                          'Departure time zone',
                          hint: 'America/New_York',
                        ),
                        _text(
                          legs[index],
                          'arrivalAirport',
                          'Arrival airport',
                          hint: 'LIS · Lisbon',
                        ),
                        _text(
                          legs[index],
                          'arrivalDate',
                          'Arrival date',
                          hint: 'YYYY-MM-DD',
                        ),
                        _text(
                          legs[index],
                          'arrivalTime',
                          'Arrival time',
                          hint: 'HH:mm',
                        ),
                        _text(
                          legs[index],
                          'arrivalTimeZone',
                          'Arrival time zone',
                          hint: 'Europe/Lisbon',
                        ),
                      ],
                    ),
                  ),
                OutlinedButton.icon(
                  onPressed: () => setState(
                    () => legs.add(<String, dynamic>{
                      'departureTimeZone': 'UTC',
                      'arrivalTimeZone': 'UTC',
                    }),
                  ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add connecting leg'),
                ),
                _text(_entry, 'notes', 'Itinerary notes', lines: 3),
              ],
              _text(
                _entry,
                'url',
                _lodging ? 'Listing or booking link' : 'Booking or search link',
                hint: 'https://…',
                keyboard: TextInputType.url,
              ),
              _heading('Named travelers · private'),
              const Text(
                'Associations are visible only to hosts and the named traveler. They do not confirm a room or ticket purchase.',
              ),
              if (_travelers.isEmpty)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 12),
                  child: Text(
                    'Confirm attendance before associating travelers.',
                  ),
                ),
              for (final traveler in _travelers)
                CheckboxListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(traveler['displayName'] as String? ?? 'Traveler'),
                  value: assigned.contains(traveler['userId']),
                  onChanged: (value) => setState(() {
                    final id = traveler['userId'] as String;
                    if (value == true) {
                      assigned.add(id);
                    } else {
                      assigned.remove(id);
                    }
                    _entry['assignedUserIds'] = assigned;
                  }),
                ),
              for (final id in assigned.where(
                (id) => !_travelers.any((person) => person['userId'] == id),
              ))
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Previously associated traveler'),
                  subtitle: const Text('Not currently confirmed'),
                  trailing: IconButton(
                    tooltip: 'Remove unavailable traveler',
                    onPressed: () => setState(() {
                      assigned.remove(id);
                      _entry['assignedUserIds'] = assigned;
                    }),
                    icon: const Icon(Icons.close),
                  ),
                ),
              _text(_entry, 'hostNotes', 'Host-only notes', lines: 3),
              if (_outside)
                CheckboxListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('I reviewed the dates outside this trip'),
                  subtitle: const Text(
                    'Keep them intentionally, such as arriving the night before.',
                  ),
                  value: _entry['outsideTripDatesConfirmed'] == true,
                  onChanged: (value) => setState(
                    () => _entry['outsideTripDatesConfirmed'] = value == true,
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}
