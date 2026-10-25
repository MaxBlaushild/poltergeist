import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

class TravelTripSettingsDialog extends StatefulWidget {
  const TravelTripSettingsDialog({
    super.key,
    required this.stop,
    required this.service,
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;

  @override
  State<TravelTripSettingsDialog> createState() =>
      _TravelTripSettingsDialogState();
}

class _TravelTripSettingsDialogState extends State<TravelTripSettingsDialog> {
  late final TextEditingController _maximum;
  late bool _waitlist;
  late bool _joinOpen;
  late bool _showCount;
  late String _roster;
  bool _busy = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    final settings = tripMap(widget.stop['tripSettings']);
    _maximum = TextEditingController(
      text: settings['maxAttendees']?.toString() ?? '',
    );
    _waitlist = settings['waitlistEnabled'] == true;
    _joinOpen = widget.stop['joinOpen'] == true;
    _showCount = settings['showAttendeeCount'] != false;
    _roster = settings['rosterVisibility'] as String? ?? 'attendees';
  }

  @override
  void dispose() {
    _maximum.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final text = _maximum.text.trim();
    final maximum = text.isEmpty ? null : int.tryParse(text);
    if (text.isNotEmpty && (maximum == null || maximum < 1)) {
      setState(
        () => _error =
            'Enter a positive whole number, or leave blank for unlimited.',
      );
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.put(
        '/calendar/stops/${widget.stop['id']}/settings',
        data: {
          'revision': widget.stop['revision'],
          'maxAttendees': maximum,
          'waitlistEnabled': _waitlist,
          'joinOpen': _joinOpen,
          'rosterVisibility': _roster,
          'showAttendeeCount': _showCount,
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
    title: 'Attendance & privacy',
    busy: _busy,
    error: _error,
    onSave: _save,
    saveLabel: 'Apply settings now',
    children: [
      const Text(
        'These settings take effect immediately. They do not publish your content draft.',
      ),
      const SizedBox(height: 20),
      SwitchListTile.adaptive(
        contentPadding: EdgeInsets.zero,
        value: _joinOpen,
        onChanged: (value) => setState(() => _joinOpen = value),
        title: const Text('Accept new attendance requests'),
        subtitle: const Text(
          'Takes effect now. Existing attendance and reservations remain in place.',
        ),
      ),
      TextField(
        controller: _maximum,
        keyboardType: TextInputType.number,
        inputFormatters: [FilteringTextInputFormatter.digitsOnly],
        decoration: const InputDecoration(
          labelText: 'Maximum attendees across the whole trip',
          hintText: 'Unlimited',
          helperText:
              'Counts distinct confirmed travelers, including attending hosts. Partial dates still use one place. This is not a daily or lodging limit.',
          helperMaxLines: 4,
        ),
      ),
      const SizedBox(height: 12),
      const Text(
        'Previously confirmed travelers who need to reconfirm retain reserved places. The maximum cannot be lower than occupied and reserved places.',
      ),
      SwitchListTile.adaptive(
        contentPadding: EdgeInsets.zero,
        value: _waitlist,
        onChanged: (value) => setState(() => _waitlist = value),
        title: const Text('Offer a waitlist when full'),
        subtitle: const Text(
          'Hosts choose whom to approve. Disabling this preserves existing waitlist entries.',
        ),
      ),
      const Divider(height: 28),
      DropdownButtonFormField<String>(
        initialValue: _roster,
        decoration: const InputDecoration(
          labelText: 'Who can see attendee names & avatars',
        ),
        isExpanded: true,
        items: const [
          DropdownMenuItem(
            value: 'attendees',
            child: Text('Hosts and confirmed attendees'),
          ),
          DropdownMenuItem(value: 'hosts', child: Text('Hosts only')),
        ],
        onChanged: (value) => setState(() => _roster = value!),
      ),
      const SizedBox(height: 12),
      const Text(
        'Guests and unconfirmed viewers never see attendee identities. A broader audience requires each affected attendee’s acknowledgement before their identity is shown.',
      ),
      SwitchListTile.adaptive(
        contentPadding: EdgeInsets.zero,
        value: _showCount,
        onChanged: (value) => setState(() => _showCount = value),
        title: const Text('Show attendance counts'),
        subtitle: const Text(
          'When hidden, viewers see availability without exact counts or remaining places.',
        ),
      ),
    ],
  );
}

class TravelTripConsentDialog extends StatefulWidget {
  const TravelTripConsentDialog({
    super.key,
    required this.stop,
    required this.service,
    this.hostAttendance = false,
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;
  final bool hostAttendance;

  @override
  State<TravelTripConsentDialog> createState() =>
      _TravelTripConsentDialogState();
}

class _TravelTripConsentDialogState extends State<TravelTripConsentDialog> {
  bool _acknowledged = false;
  bool _shareDates = false;
  bool _attending = true;
  late final TextEditingController _start;
  late final TextEditingController _end;
  bool _busy = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    final participation = tripMap(widget.stop['ownParticipation']);
    _shareDates = participation['shareDates'] == true;
    _start = TextEditingController(
      text:
          participation['startDate'] as String? ??
          widget.stop['startDate'] as String? ??
          '',
    );
    _end = TextEditingController(
      text:
          participation['endDate'] as String? ??
          widget.stop['endDate'] as String? ??
          '',
    );
    _attending =
        participation['status'] == 'joined' ||
        participation['reserved'] == true;
  }

  @override
  void dispose() {
    _start.dispose();
    _end.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if ((!widget.hostAttendance || _attending) && !_acknowledged) {
      setState(
        () => _error = 'Acknowledge the attendee audience before saving.',
      );
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.post(
        '/calendar/stops/${widget.stop['id']}/${widget.hostAttendance ? 'host-attendance' : 'roster-consent'}',
        data: {
          'revision': widget.stop['revision'],
          'rosterConsentVersion': widget.stop['rosterVersion'],
          'shareDates': _shareDates,
          if (widget.hostAttendance) 'attending': _attending,
          if (widget.hostAttendance &&
              _attending &&
              widget.stop['datePrecision'] == 'fixed') ...{
            'startDate': _start.text.trim(),
            'endDate': _end.text.trim(),
          },
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
  Widget build(BuildContext context) {
    final audience =
        tripMap(widget.stop['tripSettings'])['rosterVisibility'] == 'hosts'
        ? 'trip hosts'
        : 'trip hosts and confirmed attendees';
    return TripFormDialog(
      title: widget.hostAttendance
          ? 'Are you traveling too?'
          : 'Your attendee privacy',
      busy: _busy,
      error: _error,
      onSave: _save,
      children: [
        if (widget.hostAttendance) ...[
          SwitchListTile.adaptive(
            contentPadding: EdgeInsets.zero,
            value: _attending,
            onChanged: (value) => setState(() => _attending = value),
            title: const Text('I’m attending this trip'),
            subtitle: const Text(
              'Attending hosts use one place. You can organize without attending.',
            ),
          ),
          const SizedBox(height: 12),
          if (_attending && widget.stop['datePrecision'] == 'fixed') ...[
            TextField(
              controller: _start,
              decoration: const InputDecoration(
                labelText: 'My arrival date',
                hintText: 'YYYY-MM-DD',
              ),
              keyboardType: TextInputType.datetime,
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _end,
              decoration: const InputDecoration(
                labelText: 'My departure date',
                hintText: 'YYYY-MM-DD',
                helperText:
                    'Choose dates within the trip. Partial attendance still uses one place.',
                helperMaxLines: 3,
              ),
              keyboardType: TextInputType.datetime,
            ),
            const SizedBox(height: 16),
          ],
        ],
        Text(
          'Your name and avatar may be shown to $audience. Your contact information and private booking details are never shown on the roster.',
        ),
        CheckboxListTile(
          contentPadding: EdgeInsets.zero,
          title: Text('I understand who can see my name and avatar'),
          value: _acknowledged,
          onChanged: (value) => setState(() => _acknowledged = value == true),
        ),
        CheckboxListTile(
          contentPadding: EdgeInsets.zero,
          title: const Text(
            'Share my attendance dates with confirmed attendees',
          ),
          subtitle: const Text(
            'Otherwise, only you and hosts can see your dates.',
          ),
          value: _shareDates,
          onChanged: (value) => setState(() => _shareDates = value == true),
        ),
      ],
    );
  }
}

class TravelTripParticipantsDialog extends StatefulWidget {
  const TravelTripParticipantsDialog({
    super.key,
    required this.stop,
    required this.service,
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;

  @override
  State<TravelTripParticipantsDialog> createState() =>
      _TravelTripParticipantsDialogState();
}

class _TravelTripParticipantsDialogState
    extends State<TravelTripParticipantsDialog> {
  List<Map<String, dynamic>> _people = [];
  String _filter = 'all';
  bool _busy = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final result = await widget.service.get(
        '/calendar/stops/${widget.stop['id']}/participations',
      );
      if (mounted) {
        setState(() {
          _people = tripMaps(result['participations'])
            ..sort(
              (a, b) => (a['createdAt'] as String? ?? '').compareTo(
                b['createdAt'] as String? ?? '',
              ),
            );
        });
      }
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _review(Map<String, dynamic> person, String status) async {
    if (status == 'release_reservation' || status == 'removed') {
      final confirmed = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: Text(
            status == 'removed'
                ? 'Remove attendance?'
                : 'Release reserved place?',
          ),
          content: const Text(
            'This person’s place becomes available to others. They are notified according to their current access.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Confirm'),
            ),
          ],
        ),
      );
      if (confirmed != true) return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.put(
        '/calendar/stops/${widget.stop['id']}/participations/${person['id']}',
        data: {'status': status},
      );
      await _load();
    } catch (error) {
      if (mounted) {
        setState(() {
          _error = tripError(error);
          _busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Manage attendance'),
    content: SizedBox(
      width: 620,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Requests and waitlist entries are shown oldest first. You choose whom to approve; capacity is checked for every approval.',
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: _filter,
              decoration: const InputDecoration(labelText: 'Show'),
              items: [
                const DropdownMenuItem(value: 'all', child: Text('Everyone')),
                for (final status in [
                  'requested',
                  'waitlisted',
                  'joined',
                  'needs_reconfirmation',
                  'interested',
                  'removed',
                ])
                  DropdownMenuItem(
                    value: status,
                    child: Text(tripParticipationLabel(status)),
                  ),
              ],
              onChanged: (value) => setState(() => _filter = value!),
            ),
            if (_busy) const LinearProgressIndicator(),
            if (_error != null)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 12),
                child: Text(
                  _error!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (!_busy &&
                _people
                    .where((p) => _filter == 'all' || p['status'] == _filter)
                    .isEmpty)
              const Padding(
                padding: EdgeInsets.all(24),
                child: Text('No one in this group yet.'),
              ),
            for (final person in _people.where(
              (p) => _filter == 'all' || p['status'] == _filter,
            ))
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(14),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        person['displayName'] as String? ?? 'Traveler',
                        style: const TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: 16,
                        ),
                      ),
                      Text(
                        '${tripParticipationLabel(person['status'])}${person['reserved'] == true ? ' · Place reserved' : ''}',
                      ),
                      if ((person['startDate'] as String?)?.isNotEmpty == true)
                        Text('${person['startDate']} – ${person['endDate']}'),
                      Wrap(
                        spacing: 8,
                        children: [
                          if ([
                            'requested',
                            'waitlisted',
                          ].contains(person['status'])) ...[
                            FilledButton(
                              onPressed: _busy
                                  ? null
                                  : () => _review(person, 'joined'),
                              child: const Text('Approve'),
                            ),
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _review(person, 'declined'),
                              child: const Text('Decline'),
                            ),
                          ],
                          if (person['reserved'] == true)
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () =>
                                        _review(person, 'release_reservation'),
                              child: const Text('Release reservation'),
                            ),
                          if ([
                            'requested',
                            'waitlisted',
                            'joined',
                            'needs_reconfirmation',
                            'interested',
                          ].contains(person['status']))
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _review(person, 'removed'),
                              child: const Text('Remove'),
                            ),
                          if (person['status'] == 'removed')
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _review(person, 'withdrawn'),
                              child: const Text('Allow requests again'),
                            ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
          ],
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context, true),
        child: const Text('Done'),
      ),
    ],
  );
}
