import 'package:flutter/material.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_layout.dart';
import 'package:travel_angels/widgets/travel_flight_card.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

class TravelFlightLookup extends StatefulWidget {
  const TravelFlightLookup({
    super.key,
    required this.stopId,
    required this.service,
    required this.onSelected,
    this.departureDate = '',
    this.busy = false,
  });
  final String stopId;
  final TravelCalendarService service;
  final ValueChanged<Map<String, dynamic>> onSelected;
  final String departureDate;
  final bool busy;
  @override
  State<TravelFlightLookup> createState() => _TravelFlightLookupState();
}

class _TravelFlightLookupState extends State<TravelFlightLookup> {
  final _number = TextEditingController();
  final _origin = TextEditingController();
  final _destination = TextEditingController();
  late final _date = TextEditingController(text: widget.departureDate);
  String _mode = 'number';
  List<Map<String, dynamic>> _results = [];
  String? _error;
  String? _message;
  bool _loading = false;
  int _generation = 0;
  @override
  void dispose() {
    _generation++;
    for (final controller in [_number, _origin, _destination, _date]) {
      controller.dispose();
    }
    super.dispose();
  }

  void _invalidate() {
    _generation++;
    setState(() {
      _results = [];
      _error = null;
      _message = null;
      _loading = false;
    });
  }

  Future<void> _search() async {
    if (_loading || widget.busy) return;
    final date = _date.text.trim();
    final number = _number.text.trim().toUpperCase().replaceAll(' ', '');
    final origin = _origin.text.trim().toUpperCase();
    final destination = _destination.text.trim().toUpperCase();
    if (calendarCivilDate(date) == null) {
      setState(() => _error = 'Enter a departure date as YYYY-MM-DD.');
      return;
    }
    if (_mode == 'number' &&
        !RegExp(r'^[A-Z0-9]{2,3}[0-9]{1,4}[A-Z]?$').hasMatch(number)) {
      setState(
        () => _error = 'Enter an airline and flight number, such as BA178.',
      );
      return;
    }
    if (_mode == 'route' &&
        (!RegExp(r'^[A-Z]{3}$').hasMatch(origin) ||
            !RegExp(r'^[A-Z]{3}$').hasMatch(destination) ||
            origin == destination)) {
      setState(
        () => _error =
            'Enter two different three-letter airport codes, such as JFK and LHR.',
      );
      return;
    }
    final generation = ++_generation;
    setState(() {
      _loading = true;
      _results = [];
      _error = null;
      _message = null;
    });
    try {
      final response = await widget.service.get(
        '/calendar/stops/${Uri.encodeComponent(widget.stopId)}/flight-search',
        query: {
          'departureDate': date,
          if (_mode == 'number') 'flightNumber': number,
          if (_mode == 'route') ...{
            'origin': origin,
            'destination': destination,
          },
        },
      );
      if (!mounted || generation != _generation) return;
      setState(() {
        _results = tripMaps(response['flights']);
        if (response['configured'] == false) {
          _message =
              'Flight lookup is not available yet. You can add the flight details manually below.';
        } else if (_results.isEmpty) {
          _message =
              'No flights found for this search. Try another date or add the details manually below.';
        }
      });
    } catch (error) {
      if (mounted && generation == _generation) {
        setState(
          () => _error =
              '${tripError(error)} You can still add the flight manually below.',
        );
      }
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loading = false);
      }
    }
  }

  Future<void> _pickDate() async {
    final value = calendarCivilDate(_date.text);
    final now = DateTime.now();
    final date = await showDatePicker(
      context: context,
      initialDate: value != null && value.year >= 2000 && value.year <= 2100
          ? value
          : DateTime(now.year, now.month, now.day),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100, 12, 31),
    );
    if (date != null && mounted) {
      _date.text = calendarDateKey(date);
      _invalidate();
    }
  }

  Widget _input(
    String label,
    TextEditingController controller,
    String key, {
    String? hint,
  }) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: TextField(
      key: ValueKey(key),
      controller: controller,
      enabled: !widget.busy,
      textCapitalization: TextCapitalization.characters,
      onChanged: (_) => _invalidate(),
      decoration: InputDecoration(labelText: label, hintText: hint),
      onSubmitted: (_) => _search(),
    ),
  );
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      Text('Find a flight', style: Theme.of(context).textTheme.titleLarge),
      const SizedBox(height: 8),
      Wrap(
        spacing: 8,
        runSpacing: 6,
        children: [
          for (final mode in ['number', 'route'])
            ChoiceChip(
              label: Text(mode == 'number' ? 'Flight number' : 'Route'),
              selected: _mode == mode,
              onSelected: widget.busy
                  ? null
                  : (_) {
                      _mode = mode;
                      _invalidate();
                    },
            ),
        ],
      ),
      const SizedBox(height: 12),
      if (_mode == 'number')
        _input(
          'Airline + flight number',
          _number,
          'flight-search-number',
          hint: 'BA178',
        )
      else ...[
        _input('From airport', _origin, 'flight-search-origin', hint: 'JFK'),
        _input(
          'To airport',
          _destination,
          'flight-search-destination',
          hint: 'LHR',
        ),
      ],
      TextField(
        key: const ValueKey('flight-search-date'),
        controller: _date,
        enabled: !widget.busy,
        onChanged: (_) => _invalidate(),
        keyboardType: TextInputType.datetime,
        decoration: InputDecoration(
          labelText: 'Departure date',
          hintText: 'YYYY-MM-DD',
          suffixIcon: IconButton(
            tooltip: 'Choose departure date',
            onPressed: widget.busy ? null : _pickDate,
            icon: const Icon(Icons.calendar_today_outlined, size: 18),
          ),
        ),
      ),
      Align(
        alignment: Alignment.centerLeft,
        child: TextButton.icon(
          key: const ValueKey('flight-search-submit'),
          onPressed: _loading || widget.busy ? null : _search,
          icon: _loading
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.search, size: 18),
          label: Text(_loading ? 'Finding flights…' : 'Search flights'),
        ),
      ),
      if (_error != null)
        Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: Text(
            _error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ),
      if (_message != null)
        Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: Text(_message!),
        ),
      for (var i = 0; i < _results.length; i++)
        Container(
          key: ValueKey('flight-search-result-$i'),
          padding: const EdgeInsets.all(12),
          margin: const EdgeInsets.only(bottom: 10),
          decoration: BoxDecoration(
            border: Border.all(color: tripOutline),
            borderRadius: BorderRadius.circular(3),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                flightName(_results[i]),
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 6),
              FlightRouteSummary(departure: _results[i], arrival: _results[i]),
              const SizedBox(height: 6),
              Text(
                '${flightDate(_results[i]['departureDate'])} · Times local to each airport',
                style: Theme.of(context).textTheme.bodySmall,
              ),
              if (flightText(_results[i]['operatingAirline']).isNotEmpty)
                Text(
                  'Operated by ${_results[i]['operatingAirline']}',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              Align(
                alignment: Alignment.centerLeft,
                child: TextButton(
                  key: ValueKey('flight-search-use-$i'),
                  onPressed: widget.busy
                      ? null
                      : () {
                          const fields = [
                            'airline',
                            'flightNumber',
                            'departureAirport',
                            'arrivalAirport',
                            'departureDate',
                            'arrivalDate',
                            'departureTime',
                            'arrivalTime',
                            'departureTimeZone',
                            'arrivalTimeZone',
                            'provider',
                            'providerId',
                            'operatingAirline',
                            'operatingFlightNumber',
                          ];
                          widget.onSelected({
                            for (final field in fields)
                              if (_results[i][field] != null)
                                field: _results[i][field],
                          });
                          _invalidate();
                          setState(
                            () => _message =
                                'Flight added below. Review its details before publishing.',
                          );
                        },
                  child: const Text('Use this flight'),
                ),
              ),
            ],
          ),
        ),
      const Divider(height: 28),
    ],
  );
}
