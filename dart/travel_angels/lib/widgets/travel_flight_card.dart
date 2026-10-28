import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_layout.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

String flightText(dynamic value) => value is String ? value.trim() : '';
String flightName(Map<String, dynamic> leg) {
  final airline = flightText(leg['airline']);
  final number = flightText(leg['flightNumber']);
  if (RegExp(r'^[A-Z0-9]{2,3}$').hasMatch(airline.toUpperCase()) &&
      number.toUpperCase().startsWith(airline.toUpperCase())) {
    return number;
  }
  return [airline, number].where((value) => value.isNotEmpty).join(' ');
}

String flightDate(dynamic value) {
  final date = calendarCivilDate(value);
  return date == null ? 'Date TBD' : DateFormat.yMMMd().format(date);
}

String flightRoute(Map<String, dynamic> leg) =>
    '${flightText(leg['departureAirport']).isEmpty ? 'Departure TBD' : leg['departureAirport']} to '
    '${flightText(leg['arrivalAirport']).isEmpty ? 'Arrival TBD' : leg['arrivalAirport']}';

/// Prefer the host's booking link. A general web search is explicitly a search,
/// not a promise that a provider can sell this exact flight or current fare.
Uri flightSearchLink(Map<String, dynamic> entry) {
  final raw = flightText(entry['url']);
  final existing = Uri.tryParse(raw);
  if (existing != null &&
      const ['https', 'http'].contains(existing.scheme) &&
      existing.host.isNotEmpty &&
      existing.userInfo.isEmpty) {
    return existing;
  }
  final legs = tripMaps(entry['legs']);
  return Uri.https('www.google.com', '/search', {
    'q': [
      for (final leg in legs) ...[
        flightName(leg),
        flightText(leg['departureDate']),
        flightText(leg['departureAirport']),
        'to',
        flightText(leg['arrivalAirport']),
      ],
      'flight booking',
    ].where((value) => value.isNotEmpty).join(' '),
  });
}

String _selectionScope(
  Map<String, dynamic> selection,
  List<Map<String, dynamic>> legs,
) {
  final ids = (selection['legIds'] as List? ?? []).whereType<String>().toSet();
  final selected = legs.where((leg) => ids.contains(leg['id'])).toList();
  if (selected.length == legs.length || selected.isEmpty) return '';
  return selected.map(flightRoute).join(', ');
}

String _endpoint(String stopId, String flightId, String action) =>
    '/calendar/stops/${Uri.encodeComponent(stopId)}/flights/${Uri.encodeComponent(flightId)}/$action';

class TravelTripFlightCard extends StatefulWidget {
  const TravelTripFlightCard({
    super.key,
    required this.entry,
    this.actions,
    this.service,
    this.stopId,
    this.accessQuery = const {},
    this.draft = false,
    this.refreshToken = 0,
    this.onClaim,
    this.onJoinTrip,
  });
  final Map<String, dynamic> entry;
  final Widget? actions;
  final TravelCalendarService? service;
  final String? stopId;
  final Map<String, dynamic> accessQuery;
  final bool draft;
  final int refreshToken;
  final Future<void> Function()? onClaim;
  final Future<void> Function()? onJoinTrip;
  @override
  State<TravelTripFlightCard> createState() => _TravelTripFlightCardState();
}

class _TravelTripFlightCardState extends State<TravelTripFlightCard> {
  Map<String, dynamic>? _travelers;
  String? _error;
  bool _loading = false;
  bool _busy = false;
  bool _blocked = false;
  int _generation = 0;
  late String _fingerprint;
  AuthProvider? _auth;
  String? _identity;
  Route<void>? _dialogRoute;
  ValueNotifier<bool>? _dialogVisible;
  String get _currentFingerprint => jsonEncode([
    widget.entry,
    widget.stopId,
    widget.accessQuery,
    widget.draft,
  ]);
  bool get _connected =>
      widget.service != null &&
      flightText(widget.stopId).isNotEmpty &&
      flightText(widget.entry['id']).isNotEmpty &&
      !widget.draft;

  @override
  void initState() {
    super.initState();
    _fingerprint = _currentFingerprint;
    _load();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_accountChanged);
      _auth = auth;
      _identity = auth?.user?.id;
      _auth?.addListener(_accountChanged);
    }
  }

  void _accountChanged() {
    if (_identity == _auth?.user?.id) return;
    _identity = _auth?.user?.id;
    _generation++;
    _closeDialog();
    setState(() {
      _travelers = null;
      _blocked = true;
      _loading = false;
      _error = null;
    });
  }

  @override
  void didUpdateWidget(covariant TravelTripFlightCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_fingerprint != _currentFingerprint ||
        oldWidget.service != widget.service) {
      _fingerprint = _currentFingerprint;
      _blocked = false;
      _closeDialog();
      _load();
    } else if (oldWidget.refreshToken != widget.refreshToken && !_busy) {
      _load(quiet: true);
    }
  }

  Future<void> _load({bool quiet = false}) async {
    if (quiet && _loading) return;
    final generation = ++_generation;
    setState(() {
      if (!quiet) _travelers = null;
      _error = null;
      _loading = !quiet && _connected && !_blocked;
    });
    if (!_connected || _blocked) return;
    try {
      final result = await widget.service!.get(
        _endpoint(widget.stopId!, widget.entry['id'] as String, 'travelers'),
        query: widget.accessQuery,
      );
      if (!mounted || generation != _generation) return;
      if (_travelers != null && _dialogRoute != null) {
        String version(Map<String, dynamic> value) => jsonEncode([
          value['flightVersion'],
          value['rosterVersion'],
          value['canSelect'],
          value['canViewTravelers'],
          value['canShareWithTravelers'],
          value['sharingAudience'],
          value['ownSelection'],
        ]);
        if (version(_travelers!) != version(result)) _closeDialog();
      }
      setState(() => _travelers = result);
    } catch (error) {
      if (!mounted || generation != _generation) return;
      final denied =
          error is TravelCalendarException &&
          const [401, 403, 404, 410].contains(error.statusCode);
      if (denied || !quiet) _closeDialog();
      setState(() {
        _error = tripError(error);
        if (denied) _blocked = true;
      });
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loading = false);
      }
    }
  }

  void _closeDialog() {
    final route = _dialogRoute;
    final visible = _dialogVisible;
    _dialogRoute = null;
    _dialogVisible = null;
    if (route == null) return;
    if (SchedulerBinding.instance.schedulerPhase !=
        SchedulerPhase.persistentCallbacks) {
      visible?.value = false;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (route.isActive) {
        visible?.value = false;
        route.navigator?.removeRoute(route);
      }
    });
  }

  Future<void> _dialog(Widget child) async {
    if (_dialogRoute != null || _blocked) return;
    final visible = ValueNotifier(true);
    final route = DialogRoute<void>(
      context: context,
      builder: (_) => ValueListenableBuilder<bool>(
        valueListenable: visible,
        builder: (_, show, _) => show ? child : const SizedBox.shrink(),
      ),
    );
    _dialogRoute = route;
    _dialogVisible = visible;
    await Navigator.of(context).push(route);
    if (_dialogRoute == route) {
      _dialogRoute = null;
      _dialogVisible = null;
    }
    visible.dispose();
  }

  Future<void> _select() async {
    if (_busy || _blocked || !_connected) return;
    await _load();
    if (!mounted || _blocked || _travelers == null) return;
    final state = _travelers!;
    final gate = state['requiresAccount'] == true
        ? widget.onClaim
        : state['requiresAttendance'] == true
        ? widget.onJoinTrip
        : null;
    if (state['canRemoveSelection'] != true &&
        (state['requiresAccount'] == true ||
            state['requiresAttendance'] == true)) {
      if (gate == null) return;
      setState(() => _busy = true);
      try {
        await gate();
        if (mounted && !_blocked) await _load();
      } catch (error) {
        if (mounted && !_blocked) setState(() => _error = tripError(error));
      } finally {
        if (mounted) setState(() => _busy = false);
      }
      return;
    }
    if (state['canSelect'] != true && state['canRemoveSelection'] != true) {
      return;
    }
    await _dialog(
      TravelFlightSelectionDialog(
        entry: state['flight'] is Map ? tripMap(state['flight']) : widget.entry,
        state: state,
        service: widget.service!,
        stopId: widget.stopId!,
        accessQuery: widget.accessQuery,
      ),
    );
    if (mounted && !_blocked) await _load();
  }

  @override
  void dispose() {
    _generation++;
    _auth?.removeListener(_accountChanged);
    _closeDialog();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_blocked) return const SizedBox.shrink();
    final entry = _travelers?['flight'] is Map
        ? tripMap(_travelers!['flight'])
        : widget.entry;
    final legs = tripMaps(entry['legs']);
    final first = legs.isEmpty ? <String, dynamic>{} : legs.first;
    final last = legs.isEmpty ? <String, dynamic>{} : legs.last;
    final own = tripMap(_travelers?['ownSelection']);
    final passengerNames = _travelers?['canViewTravelers'] == true
        ? tripMaps(_travelers?['travelers'])
        : <Map<String, dynamic>>[];
    final selectLabel = own['needsReconfirmation'] == true
        ? 'Review my flight'
        : own.isNotEmpty
        ? 'My flight plans'
        : "I’m on this flight";
    final canAct =
        _travelers?['canSelect'] == true ||
        _travelers?['canRemoveSelection'] == true ||
        (_travelers?['requiresAccount'] == true && widget.onClaim != null) ||
        (_travelers?['requiresAttendance'] == true &&
            widget.onJoinTrip != null);
    return Container(
      key: ValueKey('flight-card-${widget.entry['id']}'),
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: tripPaper,
        border: Border.all(color: tripOutline),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            flightText(entry['label']).isEmpty
                ? 'Getting there'
                : entry['label'] as String,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 4),
          Text(
            [
              entry['kind'] == 'recommendation'
                  ? 'Suggested flight'
                  : 'Group flight',
              switch (entry['status']) {
                'cancelled' => 'Cancelled',
                'booked' => 'Booked',
                'planned' => 'Planned',
                _ => 'Suggested',
              },
              flightDate(first['departureDate']),
              legs.isEmpty
                  ? 'Flight details TBD'
                  : legs.length == 1
                  ? 'Nonstop'
                  : '${legs.length - 1} ${legs.length == 2 ? 'connection' : 'connections'}',
            ].join(' · '),
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: 16),
          FlightRouteSummary(departure: first, arrival: last),
          const SizedBox(height: 10),
          Text(
            legs.map(flightName).where((name) => name.isNotEmpty).join(' · '),
            style: Theme.of(context).textTheme.bodySmall,
          ),
          if (own.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Text(
                own['needsReconfirmation'] == true
                    ? 'The itinerary changed. Review your flight plans.'
                    : own['status'] == 'booked'
                    ? [
                        'You’re booked',
                        _selectionScope(own, legs),
                        'self-reported',
                      ].where((value) => value.isNotEmpty).join(' · ')
                    : [
                        'You’re considering this flight',
                        _selectionScope(own, legs),
                      ].where((value) => value.isNotEmpty).join(' · '),
                style: Theme.of(
                  context,
                ).textTheme.bodySmall?.copyWith(color: tripRust),
              ),
            ),
          if (passengerNames.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Wrap(
                spacing: 12,
                runSpacing: 6,
                children: [
                  for (final person in passengerNames)
                    Text(
                      [
                        '${flightText(person['displayName'])}${person['isHost'] == true ? ' (host)' : ''}',
                        person['status'] == 'booked' ? 'booked' : 'considering',
                        _selectionScope(person, legs),
                      ].where((value) => value.isNotEmpty).join(' · '),
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                ],
              ),
            ),
          const SizedBox(height: 6),
          Wrap(
            spacing: 4,
            runSpacing: 2,
            children: [
              TextButton(
                key: const ValueKey('flight-details'),
                onPressed: () => _dialog(_FlightDetails(entry: entry)),
                child: const Text('Flight details'),
              ),
              TextButton.icon(
                key: const ValueKey('flight-find'),
                onPressed: () =>
                    openTripLink(context, flightSearchLink(entry).toString()),
                icon: const Icon(Icons.open_in_new, size: 15),
                label: const Text('Find this flight'),
              ),
              if (_connected && canAct)
                TextButton.icon(
                  key: const ValueKey('flight-select'),
                  onPressed: _busy || _loading ? null : _select,
                  icon: const Icon(Icons.person_add_alt_1_outlined, size: 16),
                  label: Text(selectLabel),
                ),
              if (_loading)
                const Padding(
                  padding: EdgeInsets.all(12),
                  child: SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ),
            ],
          ),
          if (_error != null)
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Text(
                    _error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
                TextButton(onPressed: _load, child: const Text('Retry')),
              ],
            ),
          if (widget.actions != null) ...[
            const Divider(height: 18),
            widget.actions!,
          ],
        ],
      ),
    );
  }
}

class FlightRouteSummary extends StatelessWidget {
  const FlightRouteSummary({
    super.key,
    required this.departure,
    required this.arrival,
  });
  final Map<String, dynamic> departure;
  final Map<String, dynamic> arrival;
  @override
  Widget build(BuildContext context) {
    final start = calendarCivilDate(departure['departureDate']);
    final end = calendarCivilDate(arrival['arrivalDate']);
    final days = start == null || end == null
        ? 0
        : end.difference(start).inDays;
    Widget side(Map<String, dynamic> leg, String prefix) => Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          flightText(leg['${prefix}Airport']).isEmpty
              ? 'TBD'
              : leg['${prefix}Airport'] as String,
          style: Theme.of(
            context,
          ).textTheme.headlineSmall?.copyWith(fontSize: 26),
        ),
        const SizedBox(height: 3),
        Text(
          flightText(leg['${prefix}Time']).isEmpty
              ? 'Time TBD'
              : leg['${prefix}Time'] as String,
          softWrap: false,
          style: Theme.of(context).textTheme.titleMedium,
        ),
        if (prefix == 'arrival' && days != 0)
          Text(
            '${days > 0 ? '+' : ''}$days ${days.abs() == 1 ? 'day' : 'days'}',
            style: Theme.of(
              context,
            ).textTheme.bodySmall?.copyWith(color: tripRust),
          ),
      ],
    );
    return LayoutBuilder(
      builder: (context, constraints) {
        final scale = MediaQuery.textScalerOf(context).scale(26) / 26;
        if (constraints.maxWidth / scale < 220) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              side(departure, 'departure'),
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 10),
                child: Align(
                  alignment: Alignment.centerLeft,
                  child: Icon(Icons.arrow_downward, size: 18, color: tripRust),
                ),
              ),
              side(arrival, 'arrival'),
            ],
          );
        }
        return Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(child: side(departure, 'departure')),
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 10, vertical: 5),
              child: Icon(Icons.arrow_forward, size: 18, color: tripRust),
            ),
            Expanded(child: side(arrival, 'arrival')),
          ],
        );
      },
    );
  }
}

class FlightLegTimeline extends StatelessWidget {
  const FlightLegTimeline({super.key, required this.legs});
  final List<Map<String, dynamic>> legs;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      for (var i = 0; i < legs.length; i++) ...[
        if (i > 0)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 16),
            child: Text(
              'Connection in ${flightText(legs[i]['departureAirport'])}',
              style: Theme.of(
                context,
              ).textTheme.labelMedium?.copyWith(color: tripRust),
            ),
          ),
        Text(
          flightName(legs[i]).isEmpty
              ? 'Flight to be decided'
              : flightName(legs[i]),
          style: Theme.of(context).textTheme.titleMedium,
        ),
        if (flightText(legs[i]['operatingAirline']).isNotEmpty)
          Text(
            'Operated by ${legs[i]['operatingAirline']} ${flightText(legs[i]['operatingFlightNumber'])}',
            style: Theme.of(context).textTheme.bodySmall,
          ),
        const SizedBox(height: 8),
        FlightRouteSummary(departure: legs[i], arrival: legs[i]),
        const SizedBox(height: 8),
        for (final prefix in ['departure', 'arrival'])
          Text(
            '${prefix == 'departure' ? 'Departs' : 'Arrives'} ${flightDate(legs[i]['${prefix}Date'])} · ${flightText(legs[i]['${prefix}TimeZone'])}',
            style: Theme.of(context).textTheme.bodySmall,
          ),
      ],
    ],
  );
}

class _FlightDetails extends StatelessWidget {
  const _FlightDetails({required this.entry});
  final Map<String, dynamic> entry;
  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(
      flightText(entry['label']).isEmpty
          ? 'Flight details'
          : entry['label'] as String,
    ),
    content: SizedBox(
      width: 520,
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            FlightLegTimeline(legs: tripMaps(entry['legs'])),
            const SizedBox(height: 20),
            const Text(
              'Times are local to each airport. Check the airline for current schedules and availability.',
            ),
            if (flightText(entry['notes']).isNotEmpty) ...[
              const SizedBox(height: 16),
              TripDescription(entry['notes'] as String),
            ],
            if (flightText(entry['hostNotes']).isNotEmpty) ...[
              const SizedBox(height: 16),
              Text('Host notes: ${entry['hostNotes']}'),
            ],
            TextButton.icon(
              onPressed: () async {
                final lines = tripMaps(entry['legs']).map(
                  (leg) =>
                      '${flightName(leg)} · ${flightRoute(leg)}\n${flightDate(leg['departureDate'])} ${flightText(leg['departureTime'])} ${flightText(leg['departureTimeZone'])}',
                );
                await Clipboard.setData(
                  ClipboardData(text: lines.join('\n\n')),
                );
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Flight details copied.')),
                  );
                }
              },
              icon: const Icon(Icons.copy_outlined, size: 16),
              label: const Text('Copy flight details'),
            ),
          ],
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Close'),
      ),
    ],
  );
}

class TravelFlightSelectionDialog extends StatefulWidget {
  const TravelFlightSelectionDialog({
    super.key,
    required this.entry,
    required this.state,
    required this.service,
    required this.stopId,
    this.accessQuery = const {},
  });
  final Map<String, dynamic> entry;
  final Map<String, dynamic> state;
  final TravelCalendarService service;
  final String stopId;
  final Map<String, dynamic> accessQuery;
  @override
  State<TravelFlightSelectionDialog> createState() =>
      _TravelFlightSelectionDialogState();
}

class _TravelFlightSelectionDialogState
    extends State<TravelFlightSelectionDialog> {
  late Map<String, dynamic> _state;
  late String _status;
  late Set<String> _legIds;
  late bool _share;
  bool _reviewed = false;
  bool _busy = false;
  bool _conflict = false;
  String? _error;
  List<Map<String, dynamic>> get _legs => tripMaps(widget.entry['legs']);
  Map<String, dynamic> get _own => tripMap(_state['ownSelection']);
  bool get _canEdit => _state['canSelect'] == true;
  bool get _canShare => _canEdit && _state['canShareWithTravelers'] == true;
  bool get _shareWithHosts => _state['sharingAudience'] == 'hosts';
  @override
  void initState() {
    super.initState();
    _state = widget.state;
    _status = _own['status'] == 'booked' ? 'booked' : 'considering';
    final validIds = _legs
        .map((leg) => flightText(leg['id']))
        .where((id) => id.isNotEmpty)
        .toSet();
    final selected = (_own['legIds'] as List? ?? [])
        .whereType<String>()
        .toSet();
    _legIds = _own.isEmpty ? validIds : selected.intersection(validIds);
    _share =
        _canShare &&
        _own['rosterConsentRequired'] != true &&
        _own['shareWithTravelers'] == true;
  }

  Future<void> _save({bool remove = false}) async {
    if (_busy || _conflict || (!remove && !_canEdit)) return;
    if (!remove && _legIds.isEmpty) {
      setState(() => _error = 'Choose at least one flight leg.');
      return;
    }
    if (!remove && _own['needsReconfirmation'] == true && !_reviewed) {
      setState(
        () => _error = 'Review the changed itinerary and confirm below.',
      );
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final path = _endpoint(
        widget.stopId,
        widget.entry['id'] as String,
        'selection',
      );
      final queryPath = Uri(
        path: path,
        queryParameters: widget.accessQuery.isEmpty
            ? null
            : widget.accessQuery.map(
                (key, value) => MapEntry(key, value.toString()),
              ),
      ).toString();
      await widget.service.post(
        queryPath,
        data: {
          'status': remove ? 'removed' : _status,
          'legIds': remove ? <String>[] : _legIds.toList(),
          'flightVersion': _state['flightVersion'],
          'shareWithTravelers': !remove && _share && _canShare,
          'rosterVersion': _state['rosterVersion'],
        },
      );
      if (mounted) Navigator.pop(context);
    } catch (error) {
      if (!mounted) return;
      if (error is TravelCalendarException &&
          const [401, 403, 404, 410].contains(error.statusCode)) {
        Navigator.pop(context);
        return;
      }
      setState(() => _error = tripError(error));
      if (error is TravelCalendarException && error.statusCode == 409) {
        // A new itinerary needs a new, deliberate review from the updated card.
        setState(() {
          _conflict = true;
          _error =
              'The flight or sharing settings changed. Close this dialog and review the current details before saving.';
        });
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => TripFormDialog(
    title: 'Your flight plans',
    busy: _busy,
    error: _error,
    saveLabel: _own['needsReconfirmation'] == true
        ? 'Confirm my plans'
        : 'Save my plans',
    onSave: _conflict || !_canEdit ? null : _save,
    children: [
      Text(
        flightText(widget.entry['label']),
        style: Theme.of(context).textTheme.titleMedium,
      ),
      const SizedBox(height: 12),
      DropdownButtonFormField<String>(
        key: const ValueKey('flight-selection-status'),
        initialValue: _status,
        decoration: const InputDecoration(labelText: 'My plans'),
        style: Theme.of(context).textTheme.bodyMedium,
        isExpanded: true,
        items: const [
          DropdownMenuItem(value: 'considering', child: Text('Considering')),
          DropdownMenuItem(value: 'booked', child: Text('Booked')),
        ],
        onChanged: _busy || !_canEdit
            ? null
            : (value) => setState(() => _status = value!),
      ),
      const SizedBox(height: 8),
      const Text(
        'Booking status is self-reported. Travel Angels does not issue or verify tickets.',
      ),
      const SizedBox(height: 16),
      if (_legs.length > 1)
        CheckboxListTile(
          contentPadding: EdgeInsets.zero,
          title: const Text('All flight legs'),
          value: _legs.every((leg) => _legIds.contains(leg['id'])),
          onChanged: _busy || !_canEdit
              ? null
              : (selected) => setState(
                  () => _legIds = selected == true
                      ? _legs
                            .map((leg) => flightText(leg['id']))
                            .where((id) => id.isNotEmpty)
                            .toSet()
                      : {},
                ),
        ),
      for (final leg in _legs)
        CheckboxListTile(
          key: ValueKey('flight-selection-leg-${leg['id']}'),
          contentPadding: EdgeInsets.zero,
          value: _legIds.contains(leg['id']),
          title: Text('${flightName(leg)} · ${flightRoute(leg)}'),
          subtitle: Text(
            '${flightDate(leg['departureDate'])} · ${flightText(leg['departureTime'])}',
          ),
          onChanged: _busy || !_canEdit || flightText(leg['id']).isEmpty
              ? null
              : (selected) => setState(() {
                  if (selected == true) {
                    _legIds.add(leg['id'] as String);
                  } else {
                    _legIds.remove(leg['id']);
                  }
                }),
        ),
      if (_canEdit)
        CheckboxListTile(
          key: const ValueKey('flight-selection-share'),
          contentPadding: EdgeInsets.zero,
          title: Text(
            _shareWithHosts
                ? 'Share with trip hosts'
                : 'Share with fellow travelers',
          ),
          subtitle: Text(
            !_canShare
                ? 'Review the trip’s attendee-list privacy settings before sharing.'
                : _shareWithHosts
                ? 'Let the trip hosts see your name and flight plans.'
                : 'Let attendees who can view this flight see your name and plans.',
          ),
          value: _share,
          onChanged: _busy || !_canShare
              ? null
              : (value) => setState(() => _share = value == true),
        )
      else
        const Text('Your plans are private unless you choose to share them.'),
      if (_own['needsReconfirmation'] == true) ...[
        const SizedBox(height: 12),
        FlightLegTimeline(legs: _legs),
        CheckboxListTile(
          key: const ValueKey('flight-selection-reconfirm'),
          contentPadding: EdgeInsets.zero,
          title: const Text('I reviewed the updated itinerary'),
          value: _reviewed,
          onChanged: _busy || !_canEdit
              ? null
              : (value) => setState(() => _reviewed = value == true),
        ),
      ],
      if (_own.isNotEmpty)
        TextButton(
          key: const ValueKey('flight-selection-remove'),
          onPressed: _busy || _conflict ? null : () => _save(remove: true),
          child: const Text('Remove my flight plans'),
        ),
    ],
  );
}
