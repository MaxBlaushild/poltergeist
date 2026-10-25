import 'package:flutter/material.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

Future<bool> confirmTripImpact(
  BuildContext context,
  String title,
  Map<String, dynamic> preview, {
  String confirmLabel = 'Confirm changes',
}) async =>
    await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: SizedBox(
          width: 500,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  preview['message'] as String? ??
                      'Review the changes before confirming.',
                ),
                if (preview['audienceChange'] is Map) ...[
                  const SizedBox(height: 12),
                  Text(
                    'Inherited audience: ${tripMap(preview['audienceChange'])['fromVisibility'] ?? 'current'} → ${tripMap(preview['audienceChange'])['toVisibility'] ?? 'new owner’s calendar'}. Explicit trip permissions remain in place.',
                  ),
                  Text(
                    '${tripMap(preview['audienceChange'])['addedViewerCount'] ?? 0} people gain viewing access; ${tripMap(preview['audienceChange'])['removedViewerCount'] ?? 0} lose access.',
                  ),
                  for (final person in tripMaps(
                    tripMap(preview['audienceChange'])['addedViewers'],
                  ))
                    Text(
                      'Gains access: ${person['displayName'] ?? person['email'] ?? 'Viewer'}',
                    ),
                  for (final person in tripMaps(
                    tripMap(preview['audienceChange'])['removedViewers'],
                  ))
                    Text(
                      'Loses access: ${person['displayName'] ?? person['email'] ?? 'Viewer'}',
                    ),
                  if (tripMap(preview['audienceChange'])['linkAccessChanged'] ==
                      true)
                    const Text(
                      'Access through calendar sharing links will change.',
                    ),
                ],
                if ((preview['removedParticipationCount'] as num? ?? 0) > 0)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: Text(
                      '${preview['removedParticipationCount']} participation record(s) will end when viewing access is lost.',
                    ),
                  ),
                for (final person in tripMaps(preview['losingManagement']))
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      '${person['displayName'] ?? 'Collaborator'}: current management ends${person['pendingHostInvitation'] == true ? '; a full-host invitation awaits their acceptance' : ''}.',
                    ),
                  ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Back'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(confirmLabel),
          ),
        ],
      ),
    ) ??
    false;

class TravelTripHostsDialog extends StatefulWidget {
  const TravelTripHostsDialog({
    super.key,
    required this.stop,
    required this.service,
    required this.currentUserId,
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;
  final String? currentUserId;

  @override
  State<TravelTripHostsDialog> createState() => _TravelTripHostsDialogState();
}

class _TravelTripHostsDialogState extends State<TravelTripHostsDialog> {
  final _search = TextEditingController();
  late Map<String, dynamic> _stop;
  List<Map<String, dynamic>> _collaborators = [];
  List<Map<String, dynamic>> _results = [];
  bool _busy = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _stop = tripCopy(widget.stop);
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final results = await Future.wait([
        widget.service.get('/calendar/stops/${_stop['id']}/permissions'),
        widget.service.get('/calendar/stops/${_stop['id']}'),
      ]);
      if (mounted) {
        setState(() {
          _collaborators = tripMaps(results[0]['collaborators']);
          if (results[1]['stop'] is Map) _stop = tripMap(results[1]['stop']);
        });
      }
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _find() async {
    if (_search.text.trim().length < 2) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final result = await widget.service.get(
        '/users/search',
        query: {'query': _search.text.trim()},
      );
      if (mounted) {
        setState(() => _results = tripMaps(result['items'] ?? result['users']));
      }
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _invite(Map<String, dynamic> person) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.post(
        '/calendar/stops/${_stop['id']}/collaborators',
        data: {
          'revision': _stop['revision'],
          'userId': person['id'],
          'canManagePermissions': true,
        },
      );
      _results = [];
      _search.clear();
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

  Future<void> _remove(Map<String, dynamic> collaborator) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final path =
          '/calendar/stops/${_stop['id']}/collaborators/${collaborator['id']}';
      final preview = await widget.service.post(
        '$path/removal-preview',
        data: {'revision': _stop['revision']},
      );
      if (!mounted) return;
      setState(() => _busy = false);
      if (!await confirmTripImpact(context, 'Remove this host?', {
        ...preview,
        'message':
            'Their hosting rights end immediately. Separate viewer or attendee access is rechecked.',
      }, confirmLabel: 'Remove host')) {
        return;
      }
      if (!mounted) return;
      setState(() => _busy = true);
      await widget.service.delete(
        '$path?confirmRemoval=true',
        data: {'revision': _stop['revision'], 'confirmRemoval': true},
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
    title: const Text('Your hosting team'),
    content: SizedBox(
      width: 580,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Accepted hosts can edit and publish the trip, manage its audience, approve attendance, and invite or remove non-owner hosts. Only the owner can transfer or delete the trip.',
            ),
            if (_busy) const LinearProgressIndicator(),
            for (final host in tripMaps(_stop['hosts']))
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: TripAvatar(
                  name: host['displayName'] as String? ?? 'Host',
                  url: host['avatarUrl'] as String?,
                ),
                title: Text(host['displayName'] as String? ?? 'Host'),
                subtitle: Text(
                  '${host['isOwner'] == true ? 'Owner' : 'Host'} · ${host['isAttending'] == true ? 'Attending' : 'Organizing only'}',
                ),
              ),
            for (final collaborator in _collaborators)
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(
                  collaborator['displayName'] as String? ?? 'Invited host',
                ),
                subtitle: Text(
                  collaborator['accepted'] == true
                      ? 'Full host access'
                      : 'Invitation pending — no host access yet',
                ),
                trailing: collaborator['userId'] == _stop['ownerId']
                    ? null
                    : IconButton(
                        tooltip: 'Remove host invitation or access',
                        onPressed: _busy ? null : () => _remove(collaborator),
                        icon: const Icon(Icons.person_remove_outlined),
                      ),
              ),
            const Divider(height: 28),
            TextField(
              controller: _search,
              decoration: InputDecoration(
                labelText: 'Find an account by username',
                suffixIcon: IconButton(
                  tooltip: 'Search accounts',
                  onPressed: _busy ? null : _find,
                  icon: const Icon(Icons.search),
                ),
              ),
              onSubmitted: _busy ? null : (_) => _find(),
            ),
            for (final person in _results.where(
              (p) => p['id'] != _stop['ownerId'],
            ))
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(
                  person['name'] as String? ??
                      person['username'] as String? ??
                      'Traveler',
                ),
                subtitle: Text(person['username'] as String? ?? ''),
                trailing: TextButton(
                  onPressed: _busy ? null : () => _invite(person),
                  child: const Text('Invite as host'),
                ),
              ),
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
        onPressed: () => Navigator.pop(context, true),
        child: const Text('Done'),
      ),
    ],
  );
}

class TravelTripConversionDialog extends StatefulWidget {
  const TravelTripConversionDialog({
    super.key,
    required this.stop,
    required this.service,
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;
  @override
  State<TravelTripConversionDialog> createState() =>
      _TravelTripConversionDialogState();
}

class _TravelTripConversionDialogState
    extends State<TravelTripConversionDialog> {
  List<Map<String, dynamic>> _candidates = [];
  final Set<String> _selected = {};
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
        '/calendar/stops/${widget.stop['id']}/host-conversion',
      );
      if (mounted) setState(() => _candidates = tripMaps(result['candidates']));
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _convert() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    final data = {
      'revision': widget.stop['revision'],
      'hostUserIds': _selected.toList(),
    };
    try {
      final preview = await widget.service.post(
        '/calendar/stops/${widget.stop['id']}/host-conversion/preview',
        data: data,
      );
      if (!mounted) return;
      setState(() => _busy = false);
      if (!await confirmTripImpact(
        context,
        'Review the new hosting team',
        preview,
        confirmLabel: 'Confirm & send host invitations',
      )) {
        return;
      }
      if (!mounted) return;
      setState(() => _busy = true);
      await widget.service.post(
        '/calendar/stops/${widget.stop['id']}/host-conversion',
        data: {...data, 'confirm': true},
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
    title: 'Choose your trip hosts',
    busy: _busy,
    error: _error,
    onSave: _convert,
    saveLabel: 'Preview role changes',
    children: [
      const Text(
        'This older trip still uses calendar editor and collaborator permissions. Choose who should receive an invitation for full trip hosting rights. No one is upgraded automatically; they must accept.',
      ),
      const SizedBox(height: 16),
      for (final person in _candidates)
        CheckboxListTile(
          contentPadding: EdgeInsets.zero,
          title: Text(person['displayName'] as String? ?? 'Collaborator'),
          subtitle: Text(
            '${person['source'] ?? 'Existing collaborator'} · ${person['canManagePermissions'] == true ? 'Can manage audience' : 'Limited permissions'}',
          ),
          value: _selected.contains(person['userId']),
          onChanged: _busy
              ? null
              : (value) => setState(() {
                  if (value == true) {
                    _selected.add(person['userId'] as String);
                  } else {
                    _selected.remove(person['userId']);
                  }
                }),
        ),
      const SizedBox(height: 12),
      const Text(
        'After conversion, calendar editing alone no longer grants trip management. The next step shows whose current rights end and any affected attendance.',
      ),
    ],
  );
}

class TravelTripTransferDialog extends StatefulWidget {
  const TravelTripTransferDialog({
    super.key,
    required this.stop,
    required this.service,
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;
  @override
  State<TravelTripTransferDialog> createState() =>
      _TravelTripTransferDialogState();
}

class _TravelTripTransferDialogState extends State<TravelTripTransferDialog> {
  String? _userId;
  bool _busy = false;
  String? _error;
  Future<void> _transfer() async {
    if (_userId == null) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    final data = {'revision': widget.stop['revision'], 'userId': _userId};
    try {
      final preview = await widget.service.post(
        '/calendar/stops/${widget.stop['id']}/transfer-ownership/preview',
        data: data,
      );
      if (!mounted) return;
      setState(() => _busy = false);
      if (!await confirmTripImpact(
        context,
        'Transfer trip ownership?',
        preview,
        confirmLabel: 'Transfer ownership',
      )) {
        return;
      }
      if (!mounted) return;
      setState(() => _busy = true);
      await widget.service.post(
        '/calendar/stops/${widget.stop['id']}/transfer-ownership',
        data: {
          ...data,
          'confirm': true,
          'audienceFingerprint': preview['audienceFingerprint'],
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
    title: 'Transfer ownership',
    busy: _busy,
    error: _error,
    onSave: _userId == null ? null : _transfer,
    saveLabel: 'Review audience impact',
    children: [
      const Text(
        'Choose an accepted host. Explicit trip permissions stay the same; inherited viewing access may change with the new owner’s calendar and must be reviewed.',
      ),
      const SizedBox(height: 16),
      DropdownButtonFormField<String>(
        initialValue: _userId,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'New owner'),
        items: tripMaps(widget.stop['hosts'])
            .where((host) => host['isOwner'] != true)
            .map(
              (host) => DropdownMenuItem(
                value: host['userId'] as String,
                child: Text(host['displayName'] as String? ?? 'Host'),
              ),
            )
            .toList(),
        onChanged: (value) => setState(() => _userId = value),
      ),
    ],
  );
}
