import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

class TravelCalendarNotificationsDialog extends StatefulWidget {
  const TravelCalendarNotificationsDialog({super.key, required this.service});
  final TravelCalendarService service;

  @override
  State<TravelCalendarNotificationsDialog> createState() =>
      _TravelCalendarNotificationsDialogState();
}

class _TravelCalendarNotificationsDialogState
    extends State<TravelCalendarNotificationsDialog> {
  List<Map<String, dynamic>> _notifications = [];
  List<Map<String, dynamic>> _deliveries = [];
  bool _busy = true;
  String? _error;
  String? _confirmation;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final result = await widget.service.get('/calendar-notifications');
      if (mounted) {
        setState(() {
          _notifications = travelMaps(result['notifications']);
          _deliveries = travelMaps(result['deliveries']);
        });
      }
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _retry() async {
    setState(() {
      _busy = true;
      _error = null;
      _confirmation = null;
    });
    try {
      final result = await widget.service.post('/calendar-notifications/retry');
      if (mounted) {
        setState(
          () => _confirmation =
              '${result['pending'] ?? 0} failed email notice(s) queued to retry.',
        );
      }
      await _load();
    } catch (error) {
      if (mounted) {
        setState(() {
          _error = travelError(error);
          _busy = false;
        });
      }
    }
  }

  String _date(dynamic value) {
    final date = value is String ? DateTime.tryParse(value) : null;
    return date == null
        ? ''
        : DateFormat.yMMMd().add_jm().format(date.toLocal());
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Row(
      children: [
        const Expanded(child: Text('Travel notifications')),
        IconButton(
          tooltip: 'Refresh notifications',
          onPressed: _busy ? null : _load,
          icon: const Icon(Icons.refresh),
        ),
      ],
    ),
    content: SizedBox(
      width: 580,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Participation requests, approvals, and changes to your plans. Opening this panel does not change participation.',
            ),
            if (_busy) const LinearProgressIndicator(),
            if (_error != null)
              Text(
                _error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            if (_confirmation != null) Text(_confirmation!),
            if (!_busy && _notifications.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Text('No travel notifications yet.'),
              ),
            for (final notice in _notifications)
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const Icon(Icons.event_note_outlined),
                title: Text(
                  notice['message'] as String? ?? 'Travel plans updated.',
                ),
                subtitle: Text(_date(notice['createdAt'])),
                trailing: notice['stopId'] is String
                    ? const Icon(Icons.chevron_right)
                    : null,
                onTap: notice['stopId'] is String
                    ? () => Navigator.pop(context, notice['stopId'] as String)
                    : null,
              ),
            if (_deliveries.isNotEmpty) ...[
              const Divider(height: 24),
              Text(
                'Your email notices',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text(
                'Only your delivery statuses appear here. Access is checked again before a retry is delivered.',
              ),
              for (final delivery in _deliveries)
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    (delivery['event'] as String? ?? 'Travel update')
                        .replaceAll('_', ' '),
                  ),
                  subtitle: Text(
                    '${delivery['status'] ?? 'pending'} · ${_date(delivery['createdAt'])}${delivery['status'] == 'failed' && (delivery['lastError'] as String?)?.isNotEmpty == true ? '\n${delivery['lastError']}' : ''}',
                  ),
                ),
              if (_deliveries.any((delivery) => delivery['status'] == 'failed'))
                OutlinedButton.icon(
                  onPressed: _busy ? null : _retry,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Retry my failed email notices'),
                ),
            ],
          ],
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Done'),
      ),
    ],
  );
}

class TravelInvitationsDialog extends StatefulWidget {
  const TravelInvitationsDialog({
    super.key,
    required this.service,
    required this.calendarId,
    this.stop,
  });
  final TravelCalendarService service;
  final String calendarId;
  final TravelStop? stop;

  @override
  State<TravelInvitationsDialog> createState() =>
      _TravelInvitationsDialogState();
}

class _TravelInvitationsDialogState extends State<TravelInvitationsDialog> {
  final _email = TextEditingController();
  final _name = TextEditingController();
  List<Map<String, dynamic>> _invitations = [];
  bool _busy = true;
  String? _error;
  Map<String, dynamic> get _scope => widget.stop == null
      ? {'calendarId': widget.calendarId}
      : {'stopId': widget.stop!['id']};

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _email.dispose();
    _name.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final data = await widget.service.get('/calendar-invites', query: _scope);
      if (mounted) {
        setState(() => _invitations = travelMaps(data['invitations']));
      }
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _send() async {
    final email = _email.text.trim();
    if (!RegExp(r'^[^\s@]+@[^\s@]+\.[^\s@]+$').hasMatch(email)) {
      setState(() => _error = 'Enter a valid email address.');
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.post(
        '/calendar-invites',
        data: {
          ..._scope,
          'email': email,
          'name': _name.text.trim(),
          'grantAccess': false,
        },
      );
      _email.clear();
      _name.clear();
      await _load();
    } catch (error) {
      if (mounted) {
        setState(() {
          _error = travelError(error);
          _busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(
      widget.stop == null
          ? 'Invite friends to your calendar'
          : 'Invite friends to this stop',
    ),
    content: SizedBox(
      width: 560,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Send a browser link by email. Add restricted viewers in Audience first. Invitations do not subscribe friends or join plans for them.',
            ),
            if (_busy) const LinearProgressIndicator(),
            const SizedBox(height: 16),
            TextField(
              controller: _email,
              keyboardType: TextInputType.emailAddress,
              autocorrect: false,
              decoration: const InputDecoration(labelText: 'Friend’s email'),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _name,
              decoration: const InputDecoration(labelText: 'Name (optional)'),
            ),
            const SizedBox(height: 12),
            FilledButton.icon(
              onPressed: _busy ? null : _send,
              icon: const Icon(Icons.send_outlined),
              label: const Text('Send invitation'),
            ),
            if (_invitations.isNotEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'Invitation history',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              for (final invitation in _invitations)
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    (invitation['name'] as String?)?.isNotEmpty == true
                        ? invitation['name'] as String
                        : invitation['email'] as String? ?? 'Invited friend',
                  ),
                  subtitle: Text(
                    '${invitation['email'] ?? ''}\n${invitation['status'] ?? 'Pending'} · ${invitation['claimed'] == true ? 'Account claimed' : 'Unclaimed'} · ${invitation['subscriptionStatus'] ?? 'Not following'}',
                  ),
                  isThreeLine: true,
                  trailing: invitation['status'] == 'failed'
                      ? TextButton(
                          onPressed: _busy
                              ? null
                              : () {
                                  _email.text =
                                      invitation['email'] as String? ?? '';
                                  _name.text =
                                      invitation['name'] as String? ?? '';
                                  _send();
                                },
                          child: const Text('Retry'),
                        )
                      : null,
                ),
            ],
            if (_error != null)
              Text(
                _error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
          ],
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Done'),
      ),
    ],
  );
}

class TravelBroadcastDialog extends StatefulWidget {
  const TravelBroadcastDialog({
    super.key,
    required this.service,
    required this.calendarId,
    required this.stops,
    this.initialStop,
  });
  final TravelCalendarService service;
  final String calendarId;
  final List<TravelStop> stops;
  final TravelStop? initialStop;

  @override
  State<TravelBroadcastDialog> createState() => _TravelBroadcastDialogState();
}

class _TravelBroadcastDialogState extends State<TravelBroadcastDialog> {
  final _message = TextEditingController();
  late bool _wholeCalendar;
  final Set<String> _stopIds = {};
  Map<String, dynamic>? _preview;
  List<Map<String, dynamic>> _broadcasts = [];
  bool _busy = false;
  String? _error;
  String? _success;

  @override
  void initState() {
    super.initState();
    _wholeCalendar = widget.initialStop == null;
    if (widget.initialStop != null) {
      _stopIds.add(widget.initialStop!['id'] as String);
    }
    _load();
  }

  @override
  void dispose() {
    _message.dispose();
    super.dispose();
  }

  Map<String, dynamic> get _payload => {
    if (_wholeCalendar) 'calendarId': widget.calendarId,
    if (!_wholeCalendar) 'stopIds': _stopIds.toList(),
    'message': _message.text.trim(),
  };

  Future<void> _load() async {
    try {
      final data = await widget.service.get(
        '/calendar-broadcasts',
        query: widget.initialStop == null
            ? {'calendarId': widget.calendarId}
            : {'stopId': widget.initialStop!['id']},
      );
      if (mounted) setState(() => _broadcasts = travelMaps(data['broadcasts']));
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    }
  }

  Future<void> _prepare() async {
    if (_message.text.trim().isEmpty || (!_wholeCalendar && _stopIds.isEmpty)) {
      setState(() => _error = 'Write a message and select at least one stop.');
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
      _success = null;
    });
    try {
      final preview = await widget.service.post(
        '/calendar-broadcasts/preview',
        data: _payload,
      );
      if (mounted) setState(() => _preview = preview);
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _send() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.post('/calendar-broadcasts', data: _payload);
      if (!mounted) return;
      setState(() {
        _preview = null;
        _message.clear();
        _success = 'Broadcast created. Delivery status is shown below.';
      });
      await _load();
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _retry(String id) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.post('/calendar-broadcasts/$id/retry');
      await _load();
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Broadcast a travel update'),
    content: SizedBox(
      width: 600,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Only people who follow these plans and can currently view them receive this update. A common message about several stops reaches only people authorized for all of them.',
            ),
            const SizedBox(height: 16),
            if (_preview == null) ...[
              if (widget.initialStop == null)
                SwitchListTile.adaptive(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Calendar followers'),
                  subtitle: const Text('Includes published stops you manage.'),
                  value: _wholeCalendar,
                  onChanged: _busy
                      ? null
                      : (value) => setState(() => _wholeCalendar = value),
                ),
              if (!_wholeCalendar)
                for (final stop in widget.stops.where(
                  (s) => s['published'] == true && s['canManage'] == true,
                ))
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text(stop['title'] as String? ?? 'Stop'),
                    subtitle: Text(stop['destination'] as String? ?? ''),
                    value: _stopIds.contains(stop['id']),
                    onChanged: _busy
                        ? null
                        : (value) => setState(() {
                            if (value == true) {
                              _stopIds.add(stop['id'] as String);
                            } else {
                              _stopIds.remove(stop['id']);
                            }
                          }),
                  ),
              TextField(
                controller: _message,
                minLines: 3,
                maxLines: 6,
                maxLength: 5000,
                decoration: const InputDecoration(
                  labelText: 'Message to friends',
                  hintText: 'What would you like them to know?',
                ),
              ),
            ] else ...[
              Text(
                'Preview · ${_preview!['recipientCount'] ?? 0} eligible recipients',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 12),
              SelectableText(_message.text.trim()),
              const SizedBox(height: 12),
              for (final recipient in travelMaps(_preview!['recipients']))
                Text(
                  (recipient['email'] ?? recipient['name'] ?? 'Follower')
                      .toString(),
                ),
              const SizedBox(height: 12),
              const Text(
                'Access is checked again before delivery, including retries.',
              ),
            ],
            if (_error != null)
              Text(
                _error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            if (_success != null) Text(_success!),
            if (_broadcasts.isNotEmpty) ...[
              const Divider(height: 32),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      'Delivery history',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  IconButton(
                    tooltip: 'Refresh delivery status',
                    onPressed: _busy ? null : _load,
                    icon: const Icon(Icons.refresh),
                  ),
                ],
              ),
              for (final broadcast in _broadcasts)
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          broadcast['message'] as String? ?? 'Travel update',
                        ),
                        Text('Status: ${broadcast['status'] ?? 'pending'}'),
                        for (final delivery in travelMaps(
                          broadcast['deliveries'],
                        ))
                          Text(
                            '${delivery['email'] ?? 'Recipient'} · ${delivery['status'] ?? 'pending'}',
                          ),
                        if (broadcast['status'] == 'failed' ||
                            travelMaps(
                              broadcast['deliveries'],
                            ).any((d) => d['status'] == 'failed'))
                          TextButton.icon(
                            onPressed: _busy
                                ? null
                                : () => _retry(broadcast['id'] as String),
                            icon: const Icon(Icons.refresh),
                            label: const Text('Retry failed deliveries'),
                          ),
                      ],
                    ),
                  ),
                ),
            ],
          ],
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: _busy ? null : () => Navigator.pop(context),
        child: const Text('Close'),
      ),
      if (_preview != null)
        TextButton(
          onPressed: _busy ? null : () => setState(() => _preview = null),
          child: const Text('Edit message'),
        ),
      FilledButton(
        onPressed: _busy
            ? null
            : _preview == null
            ? _prepare
            : (_preview!['recipientCount'] as num? ?? 0) > 0
            ? _send
            : null,
        child: Text(
          _busy
              ? 'Please wait…'
              : _preview == null
              ? 'Preview audience & message'
              : 'Send broadcast',
        ),
      ),
    ],
  );
}
