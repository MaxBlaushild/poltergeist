import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

String travelError(Object error) {
  if (error is TravelCalendarException) return error.message;
  if (error is DioException) {
    final body = error.response?.data;
    if (body is Map && body['error'] is String) return body['error'] as String;
    return 'The request could not be completed. Please try again.';
  }
  return 'The request could not be completed. Please try again.';
}

List<Map<String, dynamic>> travelMaps(dynamic value) => value is List
    ? value.whereType<Map>().map((v) => Map<String, dynamic>.from(v)).toList()
    : [];

class TravelStopEditor extends StatefulWidget {
  const TravelStopEditor({
    super.key,
    required this.service,
    this.stop,
    this.calendars = const [],
  });
  final TravelCalendarService service;
  final TravelStop? stop;
  final List<Map<String, dynamic>> calendars;

  @override
  State<TravelStopEditor> createState() => _TravelStopEditorState();
}

class _TravelStopEditorState extends State<TravelStopEditor> {
  final _form = GlobalKey<FormState>();
  late final TextEditingController _title;
  late final TextEditingController _destination;
  late final TextEditingController _description;
  late final TextEditingController _timeZone;
  late String _precision;
  late String _status;
  late bool _joinOpen;
  DateTimeRange? _range;
  bool _busy = false;
  String? _error;
  String? _calendarId;

  @override
  void initState() {
    super.initState();
    final source = widget.stop?['draft'] is Map
        ? Map<String, dynamic>.from(widget.stop!['draft'] as Map)
        : widget.stop ?? {};
    _title = TextEditingController(text: source['title'] as String? ?? '');
    _destination = TextEditingController(
      text: source['destination'] as String? ?? '',
    );
    _description = TextEditingController(
      text: source['description'] as String? ?? '',
    );
    _timeZone = TextEditingController(
      text: source['timeZone'] as String? ?? 'UTC',
    );
    _calendarId = widget.calendars.firstOrNull?['id'] as String?;
    _precision = source['datePrecision'] as String? ?? 'tbd';
    _status = source['status'] as String? ?? 'tentative';
    _joinOpen = source['joinOpen'] as bool? ?? true;
    final start = travelDate(source['startDate']);
    final end = travelDate(source['endDate']);
    if (start != null && end != null && !end.isBefore(start)) {
      _range = DateTimeRange(start: start, end: end);
    }
  }

  @override
  void dispose() {
    _title.dispose();
    _destination.dispose();
    _description.dispose();
    _timeZone.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!_form.currentState!.validate()) return;
    if (_precision != 'tbd' && _range == null) {
      setState(() => _error = 'Choose dates or select Dates to be decided.');
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final data = <String, dynamic>{
        if (widget.stop == null && _calendarId != null)
          'calendarId': _calendarId,
        'title': _title.text.trim(),
        'destination': _destination.text.trim(),
        'description': _description.text.trim(),
        'datePrecision': _precision,
        'timeZone': _timeZone.text.trim(),
        'status': _status,
        'joinOpen': _joinOpen,
        'startDate': _precision == 'tbd' ? '' : travelDateValue(_range!.start),
        'endDate': _precision == 'tbd' ? '' : travelDateValue(_range!.end),
      };
      if (widget.stop == null) {
        await widget.service.post('/calendar/stops', data: data);
      } else {
        await widget.service.put(
          '/calendar/stops/${widget.stop!['id']}',
          data: data,
        );
      }
      if (mounted) Navigator.pop(context, true);
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(widget.stop == null ? 'Add a travel stop' : 'Edit stop draft'),
    content: SizedBox(
      width: 520,
      child: SingleChildScrollView(
        child: Form(
          key: _form,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Save a draft, then review its audience and publish when ready.',
              ),
              const SizedBox(height: 16),
              if (widget.stop == null && widget.calendars.length > 1)
                DropdownButtonFormField<String>(
                  initialValue: _calendarId,
                  decoration: const InputDecoration(
                    labelText: 'Owning calendar',
                  ),
                  items:
                      {
                            for (final calendar in widget.calendars)
                              calendar['id']: calendar,
                          }.values
                          .map(
                            (calendar) => DropdownMenuItem<String>(
                              value: calendar['id'] as String,
                              child: Text(
                                calendar['title'] as String? ?? 'Calendar',
                              ),
                            ),
                          )
                          .toList(),
                  onChanged: (value) => setState(() => _calendarId = value),
                ),
              TextFormField(
                controller: _title,
                decoration: const InputDecoration(labelText: 'Title'),
                maxLength: 160,
                validator: (value) => value == null || value.trim().isEmpty
                    ? 'Give this stop a title.'
                    : null,
              ),
              TextFormField(
                controller: _destination,
                decoration: const InputDecoration(labelText: 'Destination'),
                maxLength: 200,
                validator: (value) => value == null || value.trim().isEmpty
                    ? 'Enter a destination.'
                    : null,
              ),
              DropdownButtonFormField<String>(
                initialValue: _precision,
                decoration: const InputDecoration(labelText: 'Date certainty'),
                items: const [
                  DropdownMenuItem(value: 'fixed', child: Text('Exact dates')),
                  DropdownMenuItem(
                    value: 'approximate',
                    child: Text('Approximate dates'),
                  ),
                  DropdownMenuItem(
                    value: 'tbd',
                    child: Text('Dates to be decided'),
                  ),
                ],
                onChanged: (value) => setState(() => _precision = value!),
              ),
              if (_precision != 'tbd')
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  child: OutlinedButton.icon(
                    icon: const Icon(Icons.date_range),
                    label: Text(
                      _range == null
                          ? 'Choose dates'
                          : '${travelDateValue(_range!.start)} – ${travelDateValue(_range!.end)}',
                    ),
                    onPressed: () async {
                      final range = await showDateRangePicker(
                        context: context,
                        firstDate: DateTime(2000),
                        lastDate: DateTime(2100),
                        initialDateRange: _range,
                        helpText: 'Destination local dates',
                      );
                      if (range != null && mounted) {
                        setState(() => _range = range);
                      }
                    },
                  ),
                ),
              DropdownButtonFormField<String>(
                initialValue: _status,
                decoration: const InputDecoration(labelText: 'Planning status'),
                items: const [
                  DropdownMenuItem(
                    value: 'tentative',
                    child: Text('Tentative'),
                  ),
                  DropdownMenuItem(
                    value: 'confirmed',
                    child: Text('Confirmed'),
                  ),
                  DropdownMenuItem(
                    value: 'cancelled',
                    child: Text('Cancelled'),
                  ),
                ],
                onChanged: (value) => setState(() => _status = value!),
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _timeZone,
                autocorrect: false,
                decoration: const InputDecoration(
                  labelText: 'Destination time zone',
                  helperText:
                      'Used to determine when this stop has ended. For example, Europe/Lisbon.',
                  helperMaxLines: 3,
                ),
                maxLength: 100,
                validator: (value) => value == null || value.trim().isEmpty
                    ? 'Enter a destination time zone, or UTC.'
                    : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _description,
                decoration: const InputDecoration(labelText: 'About this stop'),
                minLines: 3,
                maxLines: 6,
                maxLength: 5000,
              ),
              SwitchListTile.adaptive(
                contentPadding: EdgeInsets.zero,
                title: const Text('Open to friends joining'),
                subtitle: const Text('Requests need host approval.'),
                value: _joinOpen,
                onChanged: (value) => setState(() => _joinOpen = value),
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
    ),
    actions: [
      TextButton(
        onPressed: _busy ? null : () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(
        onPressed: _busy ? null : _save,
        child: Text(_busy ? 'Saving…' : 'Save draft'),
      ),
    ],
  );
}

class TravelPermissionsDialog extends StatefulWidget {
  const TravelPermissionsDialog({super.key, required this.service, this.stop});
  final TravelCalendarService service;
  final TravelStop? stop;

  @override
  State<TravelPermissionsDialog> createState() =>
      _TravelPermissionsDialogState();
}

class _TravelPermissionsDialogState extends State<TravelPermissionsDialog> {
  String _visibility = 'private';
  List<Map<String, dynamic>> _grants = [];
  final _email = TextEditingController();
  bool _busy = true;
  String? _error;
  String get _path => widget.stop == null
      ? '/calendar/permissions'
      : '/calendar/stops/${widget.stop!['id']}/permissions';

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _email.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final data = await widget.service.get(_path);
      if (!mounted) return;
      setState(() {
        _visibility = data['visibility'] as String? ?? 'private';
        _grants = travelMaps(data['grants']);
      });
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _addEmail() {
    final email = _email.text.trim().toLowerCase();
    if (!RegExp(r'^[^\s@]+@[^\s@]+\.[^\s@]+$').hasMatch(email)) {
      setState(() => _error = 'Enter a valid email address.');
      return;
    }
    setState(() {
      if (!_grants.any((grant) => grant['email'] == email)) {
        _grants.add({'email': email});
      }
      _email.clear();
      _error = null;
    });
  }

  Future<void> _save() async {
    if (_email.text.trim().isNotEmpty) {
      _addEmail();
      if (_error != null) return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final data = <String, dynamic>{
        'visibility': _visibility,
        'grants': _grants.map((grant) {
          if (grant['userId'] != null) return {'userId': grant['userId']};
          if (grant['recipientId'] != null) {
            return {'recipientId': grant['recipientId']};
          }
          return {'email': grant['email']};
        }).toList(),
      };
      final preview = await widget.service.post('$_path/preview', data: data);
      if (!mounted) return;
      final removed = preview['removedParticipationCount'] as int? ?? 0;
      setState(() => _busy = false);
      final confirm = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Confirm viewing permissions'),
          content: Text(
            '${travelVisibilityLabel(_visibility)}. '
            '${widget.stop == null ? 'This updates every stop that inherits this calendar, including future stops. Explicit stop audiences stay as set.' : 'This updates the published stop immediately. Draft content stays unpublished.'}'
            '\n\n${removed > 0 ? '$removed participation record(s) will be removed. Affected updates will pause. Restoring access does not restore participation.' : 'No active participation will end.'}'
            '${_visibility == 'link' ? '\n\nAnyone possessing the link can browse. Removing a named viewer will not prevent that person using this link.' : ''}',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Back'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Apply permissions'),
            ),
          ],
        ),
      );
      if (confirm != true || !mounted) return;
      setState(() => _busy = true);
      await widget.service.put(_path, data: {...data, 'confirmRemoval': true});
      if (mounted) Navigator.pop(context, true);
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(widget.stop == null ? 'Calendar audience' : 'Stop audience'),
    content: SizedBox(
      width: 480,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_busy) const LinearProgressIndicator(),
            const Text(
              'Viewing, managing, following updates, and joining are separate permissions.',
            ),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              key: ValueKey(_visibility),
              initialValue: _visibility,
              isExpanded: true,
              decoration: const InputDecoration(labelText: 'Who can view'),
              items: [
                for (final visibility in [
                  if (widget.stop != null) 'inherit',
                  'private',
                  'specific',
                  'link',
                ])
                  DropdownMenuItem(
                    value: visibility,
                    child: Text(travelVisibilityLabel(visibility)),
                  ),
              ],
              onChanged: _busy
                  ? null
                  : (value) => setState(() => _visibility = value!),
            ),
            const SizedBox(height: 12),
            Text(switch (_visibility) {
              'inherit' => 'Uses the owning calendar’s current audience.',
              'private' =>
                'Only the owner, calendar editors, and stop co-hosts.',
              'specific' =>
                'Only selected accounts or verified invited contacts, plus managers. This replaces the calendar audience.',
              _ =>
                'Unlisted. Anyone with a stop or calendar link can browse without signing in.',
            }),
            if (_visibility == 'specific') ...[
              const SizedBox(height: 12),
              for (final grant in _grants)
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    (grant['email'] ??
                            grant['name'] ??
                            grant['userId'] ??
                            'Viewer')
                        .toString(),
                  ),
                  trailing: IconButton(
                    tooltip: 'Remove viewer',
                    onPressed: _busy
                        ? null
                        : () => setState(() => _grants.remove(grant)),
                    icon: const Icon(Icons.close),
                  ),
                ),
              TextField(
                controller: _email,
                keyboardType: TextInputType.emailAddress,
                autocorrect: false,
                decoration: InputDecoration(
                  labelText: 'Viewer email',
                  helperText: 'An account is not required to view.',
                  suffixIcon: IconButton(
                    tooltip: 'Add viewer email',
                    onPressed: _busy ? null : _addEmail,
                    icon: const Icon(Icons.add),
                  ),
                ),
                onSubmitted: _busy ? null : (_) => _addEmail(),
              ),
              const SizedBox(height: 12),
              const Text(
                'Granting access does not send an invitation or subscribe anyone.',
              ),
            ],
            if (_error != null) ...[
              const SizedBox(height: 12),
              Text(
                _error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ],
          ],
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: _busy ? null : () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(
        onPressed: _busy ? null : _save,
        child: const Text('Review changes'),
      ),
    ],
  );
}

class TravelGuestVerificationDialog extends StatefulWidget {
  const TravelGuestVerificationDialog({super.key, required this.service});
  final TravelCalendarService service;

  @override
  State<TravelGuestVerificationDialog> createState() =>
      _TravelGuestVerificationDialogState();
}

class _TravelGuestVerificationDialogState
    extends State<TravelGuestVerificationDialog> {
  final _email = TextEditingController();
  final _code = TextEditingController();
  String? _challenge;
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _email.dispose();
    _code.dispose();
    super.dispose();
  }

  Future<void> _submit({bool resend = false}) async {
    if (_email.text.trim().isEmpty ||
        (_challenge != null && !resend && _code.text.trim().isEmpty)) {
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      if (_challenge == null || resend) {
        final result = await widget.service.post(
          '/calendar-guests/verification',
          data: {'email': _email.text.trim()},
        );
        if (mounted) {
          setState(() => _challenge = result['challengeId'] as String);
        }
      } else {
        await widget.service.post(
          '/calendar-guests/verify',
          data: {'challengeId': _challenge, 'code': _code.text.trim()},
        );
        if (mounted) Navigator.pop(context, true);
      }
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Verify your email'),
    content: SizedBox(
      width: 420,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Verify your invited email to browse restricted plans or follow updates. This does not create or claim an account.',
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _email,
              enabled: !_busy && _challenge == null,
              keyboardType: TextInputType.emailAddress,
              autocorrect: false,
              decoration: const InputDecoration(labelText: 'Email address'),
            ),
            if (_challenge != null) ...[
              const SizedBox(height: 12),
              TextField(
                controller: _code,
                keyboardType: TextInputType.number,
                autofillHints: const [AutofillHints.oneTimeCode],
                decoration: const InputDecoration(
                  labelText: 'Verification code',
                ),
                onSubmitted: _busy ? null : (_) => _submit(),
              ),
              Wrap(
                children: [
                  TextButton(
                    onPressed: _busy ? null : () => _submit(resend: true),
                    child: const Text('Send a new code'),
                  ),
                  TextButton(
                    onPressed: _busy
                        ? null
                        : () => setState(() {
                            _challenge = null;
                            _code.clear();
                          }),
                    child: const Text('Use another email'),
                  ),
                ],
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
        onPressed: _busy ? null : () => Navigator.pop(context, false),
        child: const Text('Cancel'),
      ),
      FilledButton(
        onPressed: _busy ? null : _submit,
        child: Text(
          _busy
              ? 'Please wait…'
              : _challenge == null
              ? 'Email a code'
              : 'Verify email',
        ),
      ),
    ],
  );
}

class TravelCollaboratorsDialog extends StatefulWidget {
  const TravelCollaboratorsDialog({
    super.key,
    required this.service,
    this.stop,
  });
  final TravelCalendarService service;
  final TravelStop? stop;

  @override
  State<TravelCollaboratorsDialog> createState() =>
      _TravelCollaboratorsDialogState();
}

class _TravelCollaboratorsDialogState extends State<TravelCollaboratorsDialog> {
  final _query = TextEditingController();
  List<User> _users = [];
  List<Map<String, dynamic>> _collaborators = [];
  bool _permissions = false;
  bool _busy = true;
  String? _error;
  String get _base => widget.stop == null
      ? '/calendar'
      : '/calendar/stops/${widget.stop!['id']}';

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _query.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final result = await widget.service.get('$_base/permissions');
      if (mounted) {
        setState(() => _collaborators = travelMaps(result['collaborators']));
      }
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _search() async {
    if (_query.text.trim().length < 2) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final result = await FriendService(
        APIClient(ApiConstants.baseUrl),
      ).searchUsers(_query.text.trim());
      if (mounted) setState(() => _users = result);
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _add(User user) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.post(
        '$_base/collaborators',
        data: {'userId': user.id, 'canManagePermissions': _permissions},
      );
      if (mounted) setState(() => _users = []);
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

  Future<void> _remove(Map<String, dynamic> collaborator) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final preview = await widget.service.post(
        '$_base/collaborators/${collaborator['id']}/removal-preview',
      );
      if (!mounted) return;
      final count = preview['removedParticipationCount'] as int? ?? 0;
      if (count > 0) {
        setState(() => _busy = false);
        final confirm = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('Remove collaborator access?'),
            content: Text(
              '$count participation record(s) will end because management was this person’s last viewing access. Other active viewing permissions, if any, will remain.',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('Keep collaborator'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('Remove collaborator'),
              ),
            ],
          ),
        );
        if (confirm != true || !mounted) return;
        setState(() => _busy = true);
      }
      await widget.service.delete(
        '$_base/collaborators/${collaborator['id']}?confirmRemoval=true',
      );
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
    title: Text(widget.stop == null ? 'Calendar editors' : 'Stop co-hosts'),
    content: SizedBox(
      width: 500,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              widget.stop == null
                  ? 'Editors can see and manage every stop owned by this calendar, including private stops. References to another calendar keep their own permissions.'
                  : 'Co-hosts can manage only this stop. They accept the invitation in their calendar before it appears there. Adding it to their calendar does not expand its audience.',
            ),
            if (_busy) const LinearProgressIndicator(),
            for (final person in _collaborators)
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(
                  (person['displayName'] ??
                          person['name'] ??
                          person['username'] ??
                          person['userId'] ??
                          'Collaborator')
                      .toString(),
                ),
                subtitle: Text(
                  '${person['accepted'] == true ? 'Active' : 'Awaiting acceptance'}${person['canManagePermissions'] == true ? ' · Can manage audience' : ''}',
                ),
                trailing: IconButton(
                  tooltip: 'Remove collaborator',
                  onPressed: _busy ? null : () => _remove(person),
                  icon: const Icon(Icons.person_remove_outlined),
                ),
              ),
            const SizedBox(height: 16),
            TextField(
              controller: _query,
              decoration: InputDecoration(
                labelText: 'Find an account by username',
                suffixIcon: IconButton(
                  tooltip: 'Search accounts',
                  onPressed: _busy ? null : _search,
                  icon: const Icon(Icons.search),
                ),
              ),
              onSubmitted: _busy ? null : (_) => _search(),
            ),
            CheckboxListTile(
              contentPadding: EdgeInsets.zero,
              title: Text(
                widget.stop == null
                    ? 'Allow managing stop audiences'
                    : 'Allow managing this stop’s audience',
              ),
              subtitle: const Text(
                'Viewing permissions can reveal or hide plans.',
              ),
              value: _permissions,
              onChanged: _busy
                  ? null
                  : (value) => setState(() => _permissions = value!),
            ),
            for (final user in _users)
              ListTile(
                title: Text(user.name ?? user.username ?? 'Traveler'),
                subtitle: Text(user.username ?? ''),
                trailing: TextButton(
                  onPressed: _busy ? null : () => _add(user),
                  child: Text(
                    widget.stop == null ? 'Add editor' : 'Invite co-host',
                  ),
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

class TravelStopPreviewDialog extends StatefulWidget {
  const TravelStopPreviewDialog({
    super.key,
    required this.service,
    required this.stop,
  });
  final TravelCalendarService service;
  final TravelStop stop;

  @override
  State<TravelStopPreviewDialog> createState() =>
      _TravelStopPreviewDialogState();
}

class _TravelStopPreviewDialogState extends State<TravelStopPreviewDialog> {
  final _email = TextEditingController();
  String _audience = 'anonymous';
  Map<String, dynamic>? _preview;
  bool _busy = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _email.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _busy = true;
      _error = null;
      _preview = null;
    });
    try {
      final result = await widget.service.post(
        '/calendar/stops/${widget.stop['id']}/preview',
        data: {
          'audience': _audience,
          if (_audience == 'recipient') 'email': _email.text.trim(),
        },
      );
      if (mounted) setState(() => _preview = result);
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final stop = _preview?['stop'] is Map
        ? Map<String, dynamic>.from(_preview!['stop'] as Map)
        : null;
    return AlertDialog(
      title: const Text('Preview before publishing'),
      content: SizedBox(
        width: 500,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'See whether a visitor could view this draft after publication using the current audience and active share links.',
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                initialValue: _audience,
                decoration: const InputDecoration(labelText: 'Preview as'),
                items: const [
                  DropdownMenuItem(
                    value: 'anonymous',
                    child: Text('Anonymous visitor'),
                  ),
                  DropdownMenuItem(
                    value: 'recipient',
                    child: Text('Invited email recipient'),
                  ),
                ],
                onChanged: _busy
                    ? null
                    : (value) {
                        setState(() {
                          _audience = value!;
                          _preview = null;
                        });
                        if (value == 'anonymous') _load();
                      },
              ),
              if (_audience == 'recipient') ...[
                const SizedBox(height: 12),
                TextField(
                  controller: _email,
                  keyboardType: TextInputType.emailAddress,
                  autocorrect: false,
                  decoration: const InputDecoration(
                    labelText: 'Recipient email',
                  ),
                  onChanged: (_) => setState(() => _preview = null),
                ),
                const SizedBox(height: 12),
                OutlinedButton(
                  onPressed: _busy || _email.text.trim().isEmpty ? null : _load,
                  child: const Text('Preview this recipient'),
                ),
              ],
              if (_busy) const LinearProgressIndicator(),
              if (_error != null)
                Text(
                  _error!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              if (_preview != null) ...[
                const Divider(height: 32),
                if (_preview!['canView'] != true || stop == null)
                  const Text(
                    'This visitor cannot view the stop. The shared page will not reveal its details.',
                  )
                else ...[
                  Text(
                    stop['title'] as String? ?? '',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  Text(stop['destination'] as String? ?? ''),
                  Text(travelDateLabel(stop)),
                  const SizedBox(height: 12),
                  Text(stop['description'] as String? ?? ''),
                ],
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
}

Future<void> showTravelShareLink(
  BuildContext context,
  String link,
) => showDialog<void>(
  context: context,
  builder: (context) => AlertDialog(
    title: const Text('Share this plan'),
    content: SizedBox(
      width: 460,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'The link opens in a browser. Each stop keeps its viewing permissions.',
          ),
          const SizedBox(height: 16),
          SelectableText(link),
        ],
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Done'),
      ),
      FilledButton.icon(
        onPressed: () async {
          await Clipboard.setData(ClipboardData(text: link));
          if (context.mounted) {
            Navigator.pop(context);
            ScaffoldMessenger.of(
              context,
            ).showSnackBar(const SnackBar(content: Text('Link copied')));
          }
        },
        icon: const Icon(Icons.copy),
        label: const Text('Copy link'),
      ),
    ],
  ),
);
