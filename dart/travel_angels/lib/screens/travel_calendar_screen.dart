import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_account_claim.dart';
import 'package:travel_angels/widgets/travel_calendar_delivery.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

/// An ongoing personal calendar, or its account-free shared browser view.
class TravelCalendarScreen extends StatefulWidget {
  const TravelCalendarScreen({
    super.key,
    this.calendarToken,
    this.stopToken,
    this.service,
    this.claimAccount,
  });

  final String? calendarToken;
  final String? stopToken;
  final TravelCalendarService? service;
  final Future<bool> Function(BuildContext, TravelCalendarService)?
  claimAccount;

  @override
  State<TravelCalendarScreen> createState() => _TravelCalendarScreenState();
}

class _TravelCalendarScreenState extends State<TravelCalendarScreen> {
  late final TravelCalendarService _service;
  final _search = TextEditingController();
  Map<String, dynamic> _calendar = {};
  List<TravelStop> _stops = [];
  List<Map<String, dynamic>> _cohostInvitations = [];
  List<Map<String, dynamic>> _editableCalendars = [];
  final Map<String, Map<String, dynamic>> _selections = {};
  bool _loading = true;
  bool _busy = false;
  bool _monthView = false;
  DateTime _month = DateTime.now();
  String _period = 'all';
  DateTimeRange? _dateFilter;
  String? _error;
  String? _selectionNotice;
  String? _selectionIdentity;
  String? _viewIdentity;
  AuthProvider? _auth;
  bool _initializing = true;
  bool _claimInProgress = false;
  bool _awaitingGuestVerification = false;
  int _identityGeneration = 0;
  int _loadGeneration = 0;
  bool get _shared => widget.calendarToken != null || widget.stopToken != null;
  bool get _canManageCalendar => !_shared && _calendar['canManage'] == true;
  bool get _canManageCalendarPermissions =>
      !_shared && _calendar['canManagePermissions'] == true;
  String get _selectionKey =>
      'travel_calendar_selections:${widget.calendarToken ?? widget.stopToken ?? 'personal'}:${_selectionIdentity ?? 'guest'}';
  Map<String, dynamic> get _accessQuery => {
    if (widget.calendarToken != null) 'calendarToken': widget.calendarToken,
    if (widget.stopToken != null) 'stopToken': widget.stopToken,
  };
  Map<String, Map<String, dynamic>> get _visibleSelections {
    final visibleIds = _stops.map((stop) => stop['id']).toSet();
    return Map.fromEntries(
      _selections.entries.where((entry) => visibleIds.contains(entry.key)),
    );
  }

  @override
  void initState() {
    super.initState();
    _service = widget.service ?? TravelCalendarService();
    _auth = context.read<AuthProvider?>();
    _selectionIdentity = _auth?.user?.id;
    _viewIdentity = _selectionIdentity;
    _auth?.addListener(_authChanged);
    _initialize();
  }

  @override
  void dispose() {
    _auth?.removeListener(_authChanged);
    _search.dispose();
    super.dispose();
  }

  Future<void> _initialize() async {
    await _restoreSelections();
    await _load();
    _initializing = false;
    _authChanged();
  }

  void _authChanged() {
    if (!mounted || _initializing) return;
    final identity = _auth?.user?.id;
    if (_claimInProgress) {
      if (identity != _viewIdentity) {
        ++_loadGeneration;
        setState(() {
          _viewIdentity = identity;
          _calendar = {};
          _stops = [];
          _cohostInvitations = [];
          _editableCalendars = [];
          _loading = true;
        });
      }
      return;
    }
    if (identity != _viewIdentity) _switchIdentity(identity);
  }

  Future<void> _switchIdentity(String? identity) async {
    final generation = ++_identityGeneration;
    ++_loadGeneration;
    // Never leave the previous account's manager-only data on screen.
    setState(() {
      _selectionIdentity = identity;
      _viewIdentity = identity;
      _stops = [];
      _calendar = {};
      _cohostInvitations = [];
      _editableCalendars = [];
      _selections.clear();
      _selectionNotice = null;
      _awaitingGuestVerification = false;
      _error = null;
      _loading = true;
    });
    await _restoreSelections();
    if (!mounted || generation != _identityGeneration) return;
    await _load();
  }

  Future<void> _restoreSelections() async {
    final key = _selectionKey;
    final generation = _identityGeneration;
    final prefs = await SharedPreferences.getInstance();
    if (!mounted || generation != _identityGeneration) return;
    final saved = prefs.getString(key);
    if (saved != null) {
      try {
        final selections = jsonDecode(saved);
        if (selections is Map) {
          for (final entry in selections.entries) {
            if (entry.value is Map) {
              _selections[entry.key as String] = Map<String, dynamic>.from(
                entry.value as Map,
              );
            }
          }
        }
      } on FormatException {
        await prefs.remove(key);
      }
    }
  }

  Future<void> _saveSelections() async {
    final key = _selectionKey;
    final saved = _selections.isEmpty ? null : jsonEncode(_selections);
    final prefs = await SharedPreferences.getInstance();
    if (saved == null) {
      await prefs.remove(key);
    } else {
      await prefs.setString(key, saved);
    }
  }

  Future<void> _load() async {
    final generation = ++_loadGeneration;
    if (mounted) {
      setState(() {
        _loading = true;
        _error = null;
      });
    }
    try {
      final result = await _service.get(
        widget.stopToken != null
            ? '/shared/stops/${Uri.encodeComponent(widget.stopToken!)}'
            : widget.calendarToken != null
            ? '/shared/calendars/${Uri.encodeComponent(widget.calendarToken!)}'
            : '/calendar',
      );
      final prefs = await SharedPreferences.getInstance();
      final accountId = _auth?.user?.id;
      final recoverableGuestSession =
          _shared &&
          (prefs.getString(TravelCalendarService.guestEmailKey)?.isNotEmpty ??
              false) &&
          !await _service.hasGuestSession &&
          (accountId == null || await _service.guestNeedsClaim(accountId));
      if (!mounted || generation != _loadGeneration) return;
      final singleStop = result['stop'];
      setState(() {
        _calendar = result['calendar'] is Map
            ? Map<String, dynamic>.from(result['calendar'] as Map)
            : {
                'id': singleStop is Map ? singleStop['calendarId'] : null,
                'title': 'Travel plans',
              };
        _stops = singleStop is Map
            ? [Map<String, dynamic>.from(singleStop)]
            : travelMaps(result['stops']);
        _cohostInvitations = travelMaps(result['cohostInvitations']);
        _editableCalendars = travelMaps(result['editableCalendars']);
        final visible = _stops.map((s) => s['id']).toSet();
        final unavailable = _selections.keys
            .where((id) => !visible.contains(id))
            .toList();
        _awaitingGuestVerification =
            unavailable.isNotEmpty && recoverableGuestSession;
        if (_awaitingGuestVerification) {
          // An expired guest capability cannot prove that a viewing grant was
          // revoked. Keep unsent choices, but render only current API results.
          _selectionNotice = null;
        } else if (unavailable.isNotEmpty) {
          for (final id in unavailable) {
            _selections.remove(id);
          }
          _selectionNotice =
              'Some selected plans are no longer available. Your other selections have been kept.';
        }
      });
      await _saveSelections();
    } catch (error) {
      if (mounted && generation == _loadGeneration) {
        setState(
          () => _error = _shared
              ? 'These plans are unavailable or require access. Verify the invited email to try again.'
              : travelError(error),
        );
      }
    } finally {
      if (mounted && generation == _loadGeneration) {
        setState(() => _loading = false);
      }
    }
  }

  void _message(String message) {
    if (mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
    }
  }

  Future<void> _mutate(
    Future<void> Function() action, {
    String? success,
  }) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      await action();
      await _load();
      if (success != null) _message(success);
    } catch (error) {
      _message(travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _verify() async {
    final verified = await showDialog<bool>(
      context: context,
      builder: (context) => TravelGuestVerificationDialog(service: _service),
    );
    if (verified == true) await _load();
  }

  Future<void> _edit([TravelStop? stop]) async {
    final saved = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => TravelStopEditor(
        service: _service,
        stop: stop,
        calendars: [_calendar, ..._editableCalendars],
      ),
    );
    if (saved == true) await _load();
  }

  Future<void> _permissions([TravelStop? stop]) async {
    final saved = await showDialog<bool>(
      context: context,
      builder: (context) =>
          TravelPermissionsDialog(service: _service, stop: stop),
    );
    if (saved == true) await _load();
  }

  Future<void> _collaborators([TravelStop? stop]) async {
    await showDialog<bool>(
      context: context,
      builder: (context) =>
          TravelCollaboratorsDialog(service: _service, stop: stop),
    );
    await _load();
  }

  Future<void> _invitations([TravelStop? stop]) => showDialog<void>(
    context: context,
    builder: (context) => TravelInvitationsDialog(
      service: _service,
      calendarId: _calendar['id'] as String,
      stop: stop,
    ),
  );

  Future<void> _broadcast([TravelStop? stop]) => showDialog<void>(
    context: context,
    builder: (context) => TravelBroadcastDialog(
      service: _service,
      calendarId: _calendar['id'] as String,
      stops: _stops,
      initialStop: stop,
    ),
  );

  Future<void> _share([TravelStop? stop]) async {
    try {
      final scope = stop ?? _calendar;
      final enabled = scope['sharingEnabled'] == true;
      final token = scope['shareToken'] as String?;
      if (enabled && token != null && token.isNotEmpty) {
        final link = scope['shareUrl'] as String?;
        await showTravelShareLink(
          context,
          link != null && Uri.tryParse(link)?.hasAuthority == true
              ? link
              : _service.sharedLink(stop == null ? 'calendar' : 'stops', token),
        );
        return;
      }
      final enable = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Enable a share link'),
          content: const Text(
            'Enable a browser link for this scope. Each stop keeps its audience. Specific people must still verify their email, and private stops remain hidden.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Enable link'),
            ),
          ],
        ),
      );
      if (enable != true) return;
      final result = await _service.post(
        stop == null
            ? '/calendar/share-link'
            : '/calendar/stops/${stop['id']}/share-link',
        data: {'enabled': true, 'rotate': false},
      );
      await _load();
      if (!mounted) return;
      final data = result['calendar'] is Map
          ? Map<String, dynamic>.from(result['calendar'] as Map)
          : result['stop'] is Map
          ? Map<String, dynamic>.from(result['stop'] as Map)
          : result;
      final newToken = data['shareToken'] as String?;
      final latest = stop == null
          ? _calendar
          : _stops.where((s) => s['id'] == stop['id']).firstOrNull;
      final actualToken = newToken ?? latest?['shareToken'] as String?;
      final link =
          data['shareUrl'] as String? ?? latest?['shareUrl'] as String?;
      if (actualToken == null && link == null) {
        throw StateError('No share link returned');
      }
      await showTravelShareLink(
        context,
        link != null && Uri.tryParse(link)?.hasAuthority == true
            ? link
            : _service.sharedLink(
                stop == null ? 'calendar' : 'stops',
                actualToken!,
              ),
      );
    } catch (error) {
      _message(travelError(error));
    }
  }

  Future<void> _manageLinks([TravelStop? stop]) async {
    final choice = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Manage share link'),
        content: const Text(
          'Replacing or disabling a link invalidates the old link. Named viewers and other active sharing paths may still provide access. Delivered messages cannot be recalled.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, 'disable'),
            child: const Text('Disable link'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, 'rotate'),
            child: const Text('Replace link'),
          ),
        ],
      ),
    );
    if (choice == null) return;
    final path = stop == null
        ? '/calendar/share-link'
        : '/calendar/stops/${stop['id']}/share-link';
    final data = <String, dynamic>{
      'enabled': choice == 'rotate',
      'rotate': choice == 'rotate',
    };
    setState(() => _busy = true);
    try {
      try {
        await _service.post(path, data: data);
      } on TravelCalendarException catch (error) {
        if (error.statusCode != 409 ||
            error.details?['removedParticipationCount'] == null) {
          rethrow;
        }
        if (!mounted) return;
        setState(() => _busy = false);
        final confirm = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('Confirm ending participation'),
            content: Text(
              '${error.details!['removedParticipationCount']} participation record(s) will end because these people lose their last viewing access. Updates will pause. Restoring access will not restore participation.',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('Keep current link'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('Apply link change'),
              ),
            ],
          ),
        );
        if (confirm != true || !mounted) return;
        setState(() => _busy = true);
        await _service.post(path, data: {...data, 'confirmRemoval': true});
      }
      await _load();
      _message(
        choice == 'rotate' ? 'Share link replaced.' : 'Share link disabled.',
      );
    } catch (error) {
      _message(travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _rename() async {
    var proposedTitle = _calendar['title'] as String? ?? 'My travel calendar';
    final title = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Calendar name'),
        content: TextFormField(
          initialValue: proposedTitle,
          onChanged: (value) => proposedTitle = value,
          maxLength: 160,
          decoration: const InputDecoration(labelText: 'Name'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () {
              if (proposedTitle.trim().isNotEmpty) {
                Navigator.pop(context, proposedTitle.trim());
              }
            },
            child: const Text('Save'),
          ),
        ],
      ),
    );
    if (title != null) {
      await _mutate(() async {
        await _service.put('/calendar', data: {'title': title});
      });
    }
  }

  bool _eligible(TravelStop stop) {
    final participation = stop['ownParticipation'];
    return stop['published'] == true &&
        stop['joinOpen'] == true &&
        stop['status'] != 'cancelled' &&
        stop['acceptingParticipation'] != false &&
        !(participation is Map && participation['status'] == 'removed');
  }

  Future<void> _selectStop(TravelStop stop) async {
    final id = stop['id'] as String;
    final prior =
        _selections[id] ??
        (stop['ownParticipation'] is Map
            ? Map<String, dynamic>.from(stop['ownParticipation'] as Map)
            : <String, dynamic>{});
    String action =
        [
          'requested',
          'joined',
          'needs_reconfirmation',
        ].contains(prior['status'])
        ? 'requested'
        : 'interested';
    final fixed = stop['datePrecision'] == 'fixed';
    if (!fixed) action = 'interested';
    final first = travelDate(stop['startDate']);
    final last = travelDate(stop['endDate']);
    DateTimeRange? range;
    if (fixed && first != null && last != null) {
      final selectedStart = travelDate(prior['startDate']);
      final selectedEnd = travelDate(prior['endDate']);
      final safeStart =
          selectedStart != null &&
              !selectedStart.isBefore(first) &&
              !selectedStart.isAfter(last)
          ? selectedStart
          : first;
      final safeEnd =
          selectedEnd != null &&
              !selectedEnd.isAfter(last) &&
              !selectedEnd.isBefore(first)
          ? selectedEnd
          : last;
      range = DateTimeRange(
        start: safeEnd.isBefore(safeStart) ? first : safeStart,
        end: safeEnd,
      );
    }
    final choice = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text(stop['title'] as String? ?? 'Choose your dates'),
          content: SizedBox(
            width: 460,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(travelDateLabel(stop)),
                  if (stop['datePrecision'] == 'fixed' &&
                      (stop['timeZone'] as String?)?.isNotEmpty == true)
                    Text(
                      'Destination time zone: ${stop['timeZone']}',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  const SizedBox(height: 16),
                  DropdownButtonFormField<String>(
                    initialValue: action,
                    decoration: const InputDecoration(
                      labelText: 'Your interest',
                    ),
                    items: [
                      const DropdownMenuItem(
                        value: 'interested',
                        child: Text('Interested'),
                      ),
                      if (fixed && range != null)
                        const DropdownMenuItem(
                          value: 'requested',
                          child: Text('Request to join'),
                        ),
                    ],
                    onChanged: (value) => setDialogState(() => action = value!),
                  ),
                  if (range != null) ...[
                    const SizedBox(height: 16),
                    OutlinedButton.icon(
                      icon: const Icon(Icons.date_range),
                      label: Text(
                        '${travelDateValue(range!.start)} – ${travelDateValue(range!.end)}',
                      ),
                      onPressed: () async {
                        final picked = await showDateRangePicker(
                          context: context,
                          firstDate: first!,
                          lastDate: last!,
                          initialDateRange: range,
                          helpText: 'Choose all or part of this stop',
                        );
                        if (picked != null && context.mounted) {
                          setDialogState(() => range = picked);
                        }
                      },
                    ),
                    const Text(
                      'Dates are local to the destination. You can join for only part of a stop.',
                    ),
                  ] else ...[
                    const SizedBox(height: 12),
                    const Text(
                      'You can express interest while dates are approximate or undecided. Joining needs exact dates.',
                    ),
                  ],
                  const SizedBox(height: 16),
                  const Text(
                    'Nothing is submitted yet. You will review your selections and claim or sign into your account before confirming. Joining requires host approval.',
                  ),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, {
                'stopId': id,
                'status': action,
                'startDate': range == null ? '' : travelDateValue(range!.start),
                'endDate': range == null ? '' : travelDateValue(range!.end),
              }),
              child: const Text('Add to selections'),
            ),
          ],
        ),
      ),
    );
    if (choice != null && mounted) {
      setState(() {
        _selections[id] = choice;
        _selectionNotice = null;
      });
      await _saveSelections();
      _message('Selection saved. Review and submit when ready.');
    }
  }

  Future<void> _reviewSelections() async {
    if (_selections.isEmpty) return;
    final authenticated =
        context.read<AuthProvider?>()?.isAuthenticated ?? false;
    final currentAccountId = _auth?.user?.id;
    final needsClaim =
        !authenticated ||
        _selectionIdentity != currentAccountId ||
        await _service.guestNeedsClaim(currentAccountId ?? '');
    if (!mounted) return;
    if (needsClaim) {
      final oldKey = _selectionKey;
      _claimInProgress = true;
      try {
        final claimed = await (widget.claimAccount ?? claimTravelAccount)(
          context,
          _service,
        );
        if (!mounted) return;
        _viewIdentity = _auth?.user?.id;
        if (!claimed) {
          // A recoverable claim failure keeps the original selections, without
          // attaching them to a different account before a successful claim.
          await _load();
          return;
        }
        _selectionIdentity = _viewIdentity;
        final prefs = await SharedPreferences.getInstance();
        await prefs.remove(oldKey);
        await _saveSelections();
        await _load();
      } finally {
        _claimInProgress = false;
        _authChanged();
      }
      if (!mounted || _selections.isEmpty || _error != null) return;
    }
    if (!mounted) return;
    final reviewSelections = _visibleSelections;
    if (reviewSelections.isEmpty) return;
    final reviewIdentityGeneration = _identityGeneration;
    final reviewAccountId = _auth?.user?.id;
    final titles = {for (final stop in _stops) stop['id']: stop['title']};
    final submit = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Review your selections'),
        content: SizedBox(
          width: 520,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Text(
                  'Submit interest or requests for the stops below. A host must approve each request before you are joined.',
                ),
                const SizedBox(height: 16),
                for (final selection in reviewSelections.values)
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text(
                      titles[selection['stopId']] as String? ?? 'Travel stop',
                    ),
                    subtitle: Text(
                      '${travelStatusLabel(selection['status'])}\n${selection['startDate'] == '' ? 'Dates to be decided' : '${selection['startDate']} – ${selection['endDate']}'}',
                    ),
                    isThreeLine: true,
                  ),
                const Text(
                  'Current access, availability, and dates are checked again when you submit.',
                ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Keep reviewing'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Submit selections'),
          ),
        ],
      ),
    );
    if (submit != true || !mounted) return;
    if (reviewIdentityGeneration != _identityGeneration ||
        reviewAccountId != _auth?.user?.id) {
      _message(
        'Your account changed. Review your selections again before submitting.',
      );
      return;
    }
    setState(() => _busy = true);
    int submitted = 0;
    final failures = <String>[];
    for (final entry in reviewSelections.entries) {
      try {
        await _service.post(
          '/calendar/stops/${entry.key}/participation',
          data: {...entry.value, ..._accessQuery},
        );
        _selections.remove(entry.key);
        submitted++;
      } catch (error) {
        failures.add(travelError(error));
      }
    }
    await _saveSelections();
    await _load();
    if (!mounted) return;
    setState(() {
      _busy = false;
      _selectionNotice = failures.isEmpty
          ? null
          : 'Some selections could not be submitted. Review them and try again. ${failures.first}';
    });
    _message(
      '$submitted ${submitted == 1 ? 'selection' : 'selections'} submitted${failures.isEmpty ? '.' : '; ${failures.length} need review.'}',
    );
  }

  Future<void> _follow([TravelStop? stop]) async {
    if (!await _service.hasGuestSession) {
      if (!mounted) return;
      final verified = await showDialog<bool>(
        context: context,
        builder: (context) => TravelGuestVerificationDialog(service: _service),
      );
      if (verified != true || !mounted) return;
    }
    bool subscribed = false;
    try {
      final result = await _service.get('/calendar-subscriptions');
      subscribed = travelMaps(result['subscriptions']).any(
        (s) =>
            (stop == null
                ? s['calendarId'] == _calendar['id']
                : s['stopId'] == stop['id']) &&
            s['status'] == 'active',
      );
    } catch (error) {
      _message(travelError(error));
      return;
    }
    if (!mounted) return;
    final choice = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(subscribed ? 'Manage updates' : 'Follow travel updates'),
        content: Text(
          stop == null
              ? 'Receive updates for the stops you can view on this calendar, including newly published plans. Following does not join any stop.'
              : 'Receive updates only for this stop while you can view it. Following does not join the stop.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, !subscribed),
            child: Text(subscribed ? 'Unsubscribe' : 'Follow updates'),
          ),
        ],
      ),
    );
    if (choice == null) return;
    try {
      await _service.post(
        '/calendar-subscriptions',
        data: {
          if (stop == null) 'calendarId': _calendar['id'],
          if (stop != null) 'stopId': stop['id'],
          'subscribed': choice,
          if (widget.calendarToken != null || widget.stopToken != null)
            'shareToken': widget.stopToken ?? widget.calendarToken,
        },
      );
      _message(
        choice
            ? 'You are following updates. You have not joined any plans.'
            : 'Unsubscribed from updates.',
      );
    } catch (error) {
      _message(travelError(error));
    }
  }

  Future<void> _publish(TravelStop stop, bool published) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(published ? 'Publish this stop?' : 'Unpublish this stop?'),
        content: Text(
          published
              ? 'Published details become visible to the current audience. Material date or destination changes require participants to reconfirm. A general broadcast is sent separately.'
              : 'This removes the stop from guest views. Affected participants receive a minimal notice and will need to reconfirm after republication.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(published ? 'Publish' : 'Unpublish'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await _mutate(() async {
        await _service.post(
          '/calendar/stops/${stop['id']}/${published ? 'publish' : 'unpublish'}',
        );
      });
    }
  }

  Future<void> _participants(TravelStop stop) => showDialog<void>(
    context: context,
    builder: (context) => _TravelParticipantsDialog(
      service: _service,
      stop: stop,
      onChanged: _load,
    ),
  );

  Future<void> _notifications() async {
    final stopId = await showDialog<String>(
      context: context,
      builder: (context) =>
          TravelCalendarNotificationsDialog(service: _service),
    );
    if (stopId == null || !mounted) return;
    try {
      final result = await _service.get('/calendar/stops/$stopId');
      if (!mounted) return;
      if (result['stop'] is Map) {
        await _showStop(Map<String, dynamic>.from(result['stop'] as Map));
      }
    } catch (error) {
      _message(travelError(error));
    }
  }

  Future<void> _showStop(TravelStop stop) async {
    final action = await showModalBottomSheet<String>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (context) => SafeArea(
        child: ConstrainedBox(
          constraints: BoxConstraints(
            maxHeight: MediaQuery.sizeOf(context).height * .85,
          ),
          child: SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(24, 4, 24, 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  stop['title'] as String? ?? 'Travel stop',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  stop['destination'] as String? ?? '',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                Text(travelDateLabel(stop)),
                if (stop['datePrecision'] == 'fixed' &&
                    (stop['timeZone'] as String?)?.isNotEmpty == true)
                  Text(
                    'Destination time zone: ${stop['timeZone']}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                const SizedBox(height: 12),
                Text(
                  '${travelStatusLabel(stop['status'])}${stop['published'] == true ? '' : ' · Draft'}',
                ),
                if ((stop['description'] as String?)?.isNotEmpty == true) ...[
                  const SizedBox(height: 16),
                  Text(stop['description'] as String),
                ],
                if (stop['canManage'] == true) ...[
                  const Divider(height: 32),
                  Text(
                    'Manage this stop',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  Text(
                    'Audience: ${travelVisibilityLabel(stop['visibility'] as String?)}',
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      OutlinedButton.icon(
                        onPressed: () => Navigator.pop(context, 'edit'),
                        icon: const Icon(Icons.edit_outlined),
                        label: const Text('Edit draft'),
                      ),
                      OutlinedButton.icon(
                        onPressed: () => Navigator.pop(context, 'preview'),
                        icon: const Icon(Icons.visibility_outlined),
                        label: const Text('Preview draft'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.pop(context, 'publish'),
                        child: Text(
                          stop['published'] == true
                              ? 'Publish draft changes'
                              : 'Publish stop',
                        ),
                      ),
                      if (stop['published'] == true)
                        TextButton(
                          onPressed: () => Navigator.pop(context, 'unpublish'),
                          child: const Text('Unpublish'),
                        ),
                      OutlinedButton(
                        onPressed: () => Navigator.pop(context, 'participants'),
                        child: const Text('Participants'),
                      ),
                      if (stop['published'] == true) ...[
                        OutlinedButton(
                          onPressed: () => Navigator.pop(context, 'invite'),
                          child: const Text('Invite friends'),
                        ),
                        OutlinedButton(
                          onPressed: () => Navigator.pop(context, 'broadcast'),
                          child: const Text('Broadcast update'),
                        ),
                      ],
                      if (stop['canManagePermissions'] == true) ...[
                        OutlinedButton(
                          onPressed: () =>
                              Navigator.pop(context, 'permissions'),
                          child: const Text('Audience'),
                        ),
                        OutlinedButton(
                          onPressed: () => Navigator.pop(context, 'share'),
                          child: const Text('Share link'),
                        ),
                        TextButton(
                          onPressed: () => Navigator.pop(context, 'links'),
                          child: const Text('Replace / disable link'),
                        ),
                      ],
                      if (stop['canManageCollaborators'] == true ||
                          _canManageCalendarPermissions &&
                              stop['calendarId'] == _calendar['id'])
                        OutlinedButton(
                          onPressed: () =>
                              Navigator.pop(context, 'collaborators'),
                          child: const Text('Co-hosts'),
                        ),
                    ],
                  ),
                ],
                if (stop['published'] == true) ...[
                  const Divider(height: 32),
                  if (stop['ownParticipation'] is Map) ...[
                    Text(
                      'Your participation: ${travelStatusLabel((stop['ownParticipation'] as Map)['status'])}',
                    ),
                    if ((stop['ownParticipation'] as Map)['status'] ==
                        'needs_reconfirmation')
                      const Text(
                        'The plan changed. Review the current dates and resubmit for approval.',
                      ),
                    if (![
                      'withdrawn',
                      'removed',
                      'cancelled',
                    ].contains((stop['ownParticipation'] as Map)['status']))
                      TextButton(
                        onPressed: () => Navigator.pop(context, 'withdraw'),
                        child: const Text('Withdraw participation'),
                      ),
                  ],
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      if (_eligible(stop))
                        FilledButton.icon(
                          onPressed: () => Navigator.pop(context, 'select'),
                          icon: const Icon(Icons.add_task),
                          label: Text(
                            _selections.containsKey(stop['id'])
                                ? 'Edit selection'
                                : 'Choose dates / interest',
                          ),
                        ),
                      OutlinedButton.icon(
                        onPressed: () => Navigator.pop(context, 'follow'),
                        icon: const Icon(Icons.notifications_none),
                        label: const Text('Follow / manage updates'),
                      ),
                    ],
                  ),
                  if (!_eligible(stop))
                    const Padding(
                      padding: EdgeInsets.only(top: 12),
                      child: Text(
                        'This stop is not accepting new participation.',
                      ),
                    ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
    if (!mounted) return;
    switch (action) {
      case 'edit':
        await _edit(stop);
      case 'permissions':
        await _permissions(stop);
      case 'share':
        await _share(stop);
      case 'links':
        await _manageLinks(stop);
      case 'publish':
        await _publish(stop, true);
      case 'unpublish':
        await _publish(stop, false);
      case 'participants':
        await _participants(stop);
      case 'collaborators':
        await _collaborators(stop);
      case 'invite':
        await _invitations(stop);
      case 'broadcast':
        await _broadcast(stop);
      case 'select':
        await _selectStop(stop);
      case 'follow':
        await _follow(stop);
      case 'preview':
        await _preview(stop);
      case 'withdraw':
        await _mutate(() async {
          await _service.post(
            '/calendar/stops/${stop['id']}/participation',
            data: {'status': 'withdrawn', ..._accessQuery},
          );
        }, success: 'Participation withdrawn.');
    }
  }

  Future<void> _preview(TravelStop stop) => showDialog<void>(
    context: context,
    builder: (context) =>
        TravelStopPreviewDialog(service: _service, stop: stop),
  );

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final visibleSelections = _visibleSelections;
    final filtered = filterTravelStops(
      _stops,
      destination: _search.text,
      period: _period,
      range: _dateFilter,
    );
    return Scaffold(
      appBar: AppBar(
        title: Text(
          _shared ? 'Travel Angels · shared plans' : 'Travel calendar',
        ),
        actions: [
          if (!_shared)
            IconButton(
              tooltip: 'Travel notifications',
              onPressed: _busy ? null : _notifications,
              icon: const Icon(Icons.notifications_none),
            ),
          IconButton(
            tooltip: 'Refresh calendar',
            onPressed: _loading || _busy ? null : _load,
            icon: const Icon(Icons.refresh),
          ),
          if (_canManageCalendarPermissions)
            PopupMenuButton<String>(
              tooltip: 'Calendar settings',
              onSelected: (value) {
                switch (value) {
                  case 'rename':
                    _rename();
                  case 'permissions':
                    _permissions();
                  case 'editors':
                    _collaborators();
                  case 'links':
                    _manageLinks();
                }
              },
              itemBuilder: (context) => const [
                PopupMenuItem(value: 'rename', child: Text('Rename calendar')),
                PopupMenuItem(
                  value: 'permissions',
                  child: Text('Default audience'),
                ),
                PopupMenuItem(
                  value: 'editors',
                  child: Text('Calendar editors'),
                ),
                PopupMenuItem(value: 'links', child: Text('Manage share link')),
              ],
            ),
        ],
      ),
      body: SafeArea(
        child: _loading && _stops.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : RefreshIndicator(
                onRefresh: _load,
                child: ListView(
                  padding: EdgeInsets.symmetric(
                    horizontal: MediaQuery.sizeOf(context).width > 1100
                        ? (MediaQuery.sizeOf(context).width - 1040) / 2
                        : 16,
                    vertical: 20,
                  ),
                  children: [
                    if (_loading || _busy) const LinearProgressIndicator(),
                    if (_error != null) ...[
                      Card(
                        child: Padding(
                          padding: const EdgeInsets.all(24),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Icon(Icons.lock_outline, size: 36),
                              const SizedBox(height: 12),
                              Text(_error!),
                              const SizedBox(height: 12),
                              Wrap(
                                spacing: 8,
                                children: [
                                  FilledButton(
                                    onPressed: _load,
                                    child: const Text('Try again'),
                                  ),
                                  if (_shared)
                                    OutlinedButton(
                                      onPressed: _verify,
                                      child: const Text('Verify invited email'),
                                    ),
                                ],
                              ),
                            ],
                          ),
                        ),
                      ),
                    ] else ...[
                      Text(
                        _calendar['title'] as String? ?? 'Travel plans',
                        style: theme.textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        _shared
                            ? 'Browse the plans shared with you. Follow updates without an account, then claim an account when you want to join.'
                            : 'All your travel plans, in one ongoing calendar. Choose who can see each stop and invite friends along.',
                      ),
                      const SizedBox(height: 20),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          if (_canManageCalendar)
                            FilledButton.icon(
                              onPressed: _busy ? null : _edit,
                              icon: const Icon(Icons.add),
                              label: const Text('Add stop'),
                            ),
                          if (_canManageCalendarPermissions)
                            OutlinedButton.icon(
                              onPressed: _busy ? null : _share,
                              icon: const Icon(Icons.link),
                              label: const Text('Share calendar'),
                            ),
                          if (_canManageCalendar) ...[
                            OutlinedButton.icon(
                              onPressed: _busy ? null : _invitations,
                              icon: const Icon(Icons.person_add_outlined),
                              label: const Text('Invite friends'),
                            ),
                            OutlinedButton.icon(
                              onPressed: _busy ? null : _broadcast,
                              icon: const Icon(Icons.campaign_outlined),
                              label: const Text('Broadcast update'),
                            ),
                          ],
                          if (_shared) ...[
                            if (widget.stopToken == null)
                              OutlinedButton.icon(
                                onPressed: _busy ? null : _follow,
                                icon: const Icon(Icons.notifications_none),
                                label: const Text('Follow / manage updates'),
                              ),
                            TextButton(
                              onPressed: _verify,
                              child: const Text('Verify invited email'),
                            ),
                          ],
                        ],
                      ),
                      for (final invitation in _cohostInvitations)
                        Card(
                          child: ListTile(
                            leading: const Icon(Icons.group_add_outlined),
                            title: Text(
                              invitation['title'] as String? ??
                                  'Co-host invitation',
                            ),
                            subtitle: const Text(
                              'Accept to show this shared stop on your calendar.',
                            ),
                            trailing: TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _mutate(() async {
                                      await _service.post(
                                        '/calendar/collaborators/${invitation['id']}/accept',
                                      );
                                    }),
                              child: const Text('Accept'),
                            ),
                          ),
                        ),
                      const SizedBox(height: 24),
                      TextField(
                        controller: _search,
                        decoration: const InputDecoration(
                          labelText: 'Filter by destination or title',
                          prefixIcon: Icon(Icons.search),
                          border: OutlineInputBorder(),
                        ),
                        onChanged: (_) => setState(() {}),
                      ),
                      const SizedBox(height: 12),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        children: [
                          SegmentedButton<bool>(
                            segments: const [
                              ButtonSegment(
                                value: false,
                                icon: Icon(Icons.view_agenda_outlined),
                                label: Text('Agenda'),
                              ),
                              ButtonSegment(
                                value: true,
                                icon: Icon(Icons.calendar_month_outlined),
                                label: Text('Month'),
                              ),
                            ],
                            selected: {_monthView},
                            onSelectionChanged: (selection) =>
                                setState(() => _monthView = selection.first),
                          ),
                          DropdownButton<String>(
                            value: _period,
                            items: const [
                              DropdownMenuItem(
                                value: 'all',
                                child: Text('All dates'),
                              ),
                              DropdownMenuItem(
                                value: 'upcoming',
                                child: Text('Upcoming'),
                              ),
                              DropdownMenuItem(
                                value: 'current',
                                child: Text('Happening now'),
                              ),
                              DropdownMenuItem(
                                value: 'past',
                                child: Text('Past'),
                              ),
                              DropdownMenuItem(
                                value: 'unscheduled',
                                child: Text('Approximate / TBD'),
                              ),
                            ],
                            onChanged: (value) =>
                                setState(() => _period = value!),
                          ),
                          TextButton.icon(
                            onPressed: () async {
                              final range = await showDateRangePicker(
                                context: context,
                                firstDate: DateTime(2000),
                                lastDate: DateTime(2100),
                                initialDateRange: _dateFilter,
                                helpText: 'Filter travel dates',
                              );
                              if (range != null && mounted) {
                                setState(() => _dateFilter = range);
                              }
                            },
                            icon: const Icon(Icons.date_range),
                            label: Text(
                              _dateFilter == null
                                  ? 'Date range'
                                  : '${travelDateValue(_dateFilter!.start)} – ${travelDateValue(_dateFilter!.end)}',
                            ),
                          ),
                          if (_dateFilter != null)
                            IconButton(
                              tooltip: 'Clear date filter',
                              onPressed: () =>
                                  setState(() => _dateFilter = null),
                              icon: const Icon(Icons.close),
                            ),
                        ],
                      ),
                      const SizedBox(height: 20),
                      if (_selectionNotice != null)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 16),
                          child: Text(
                            _selectionNotice!,
                            style: TextStyle(color: theme.colorScheme.error),
                          ),
                        ),
                      if (_awaitingGuestVerification)
                        Card(
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                const Text(
                                  'Verify your invited email to restore saved selections for restricted plans.',
                                ),
                                const SizedBox(height: 8),
                                OutlinedButton(
                                  onPressed: _busy || _loading ? null : _verify,
                                  child: const Text('Restore saved selections'),
                                ),
                              ],
                            ),
                          ),
                        ),
                      if (visibleSelections.isNotEmpty)
                        Card(
                          color: theme.colorScheme.primaryContainer,
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                Text(
                                  '${visibleSelections.length} ${visibleSelections.length == 1 ? 'stop' : 'stops'} selected · nothing submitted',
                                  style: theme.textTheme.titleMedium,
                                ),
                                const SizedBox(height: 8),
                                for (final entry in visibleSelections.entries)
                                  Row(
                                    children: [
                                      Expanded(
                                        child: Text(
                                          '${_stops.where((s) => s['id'] == entry.key).firstOrNull?['title'] ?? 'Travel stop'} · ${travelStatusLabel(entry.value['status'])}',
                                        ),
                                      ),
                                      IconButton(
                                        tooltip: 'Remove selection',
                                        onPressed: _busy
                                            ? null
                                            : () {
                                                setState(
                                                  () => _selections.remove(
                                                    entry.key,
                                                  ),
                                                );
                                                _saveSelections();
                                              },
                                        icon: const Icon(Icons.close),
                                      ),
                                    ],
                                  ),
                                FilledButton(
                                  onPressed: _busy ? null : _reviewSelections,
                                  child: Text(
                                    context
                                                .watch<AuthProvider?>()
                                                ?.isAuthenticated ==
                                            true
                                        ? 'Review selections'
                                        : 'Claim account & review selections',
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      if (filtered.isEmpty)
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 36),
                          child: Column(
                            children: [
                              Icon(
                                Icons.explore_outlined,
                                size: 48,
                                color: theme.colorScheme.primary,
                              ),
                              const SizedBox(height: 16),
                              Text(
                                _stops.isEmpty
                                    ? 'No travel stops to show yet'
                                    : 'No plans match these filters',
                                style: theme.textTheme.titleLarge,
                              ),
                              const SizedBox(height: 8),
                              Text(
                                _canManageCalendar
                                    ? 'Add a destination to start planning. New stops begin as private drafts.'
                                    : 'Only published plans you can view appear here.',
                                textAlign: TextAlign.center,
                              ),
                            ],
                          ),
                        )
                      else if (_monthView)
                        TravelCalendarMonth(
                          month: _month,
                          stops: filtered,
                          onMonthChanged: (month) =>
                              setState(() => _month = month),
                          onStopTap: _showStop,
                        )
                      else ...[
                        for (final stop in filtered.where(
                          (s) => s['datePrecision'] == 'fixed',
                        ))
                          TravelStopCard(
                            stop: stop,
                            selected: _selections.containsKey(stop['id']),
                            onTap: () => _showStop(stop),
                          ),
                        if (filtered.any(
                          (s) => s['datePrecision'] != 'fixed',
                        )) ...[
                          const SizedBox(height: 20),
                          Text(
                            'Approximate & unscheduled plans',
                            style: theme.textTheme.titleMedium,
                          ),
                          const SizedBox(height: 8),
                          for (final stop in filtered.where(
                            (s) => s['datePrecision'] != 'fixed',
                          ))
                            TravelStopCard(
                              stop: stop,
                              selected: _selections.containsKey(stop['id']),
                              onTap: () => _showStop(stop),
                            ),
                        ],
                      ],
                    ],
                    const SizedBox(height: 24),
                  ],
                ),
              ),
      ),
    );
  }
}

class _TravelParticipantsDialog extends StatefulWidget {
  const _TravelParticipantsDialog({
    required this.service,
    required this.stop,
    required this.onChanged,
  });
  final TravelCalendarService service;
  final TravelStop stop;
  final Future<void> Function() onChanged;

  @override
  State<_TravelParticipantsDialog> createState() =>
      _TravelParticipantsDialogState();
}

class _TravelParticipantsDialogState extends State<_TravelParticipantsDialog> {
  List<Map<String, dynamic>> _participants = [];
  bool _busy = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final data = await widget.service.get(
        '/calendar/stops/${widget.stop['id']}/participations',
      );
      if (mounted) {
        setState(() => _participants = travelMaps(data['participations']));
      }
    } catch (error) {
      if (mounted) setState(() => _error = travelError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _update(Map<String, dynamic> participant, String status) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.service.put(
        '/calendar/stops/${widget.stop['id']}/participations/${participant['id']}',
        data: {'status': status},
      );
      await _load();
      await widget.onChanged();
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
    title: const Text('Participants'),
    content: SizedBox(
      width: 600,
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Only approved Joined participants are confirmed. Date changes need renewed approval.',
            ),
            if (_busy) const LinearProgressIndicator(),
            if (!_busy && _participants.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Text('No participation yet.'),
              ),
            for (final participant in _participants)
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        (participant['displayName'] ??
                                participant['name'] ??
                                participant['username'] ??
                                participant['userId'] ??
                                'Traveler')
                            .toString(),
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      Text(travelStatusLabel(participant['status'])),
                      Text(
                        '${participant['startDate'] ?? ''} – ${participant['endDate'] ?? ''}',
                      ),
                      Wrap(
                        spacing: 8,
                        children: [
                          if (participant['status'] == 'requested') ...[
                            FilledButton(
                              onPressed: _busy
                                  ? null
                                  : () => _update(participant, 'joined'),
                              child: const Text('Approve'),
                            ),
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _update(participant, 'declined'),
                              child: const Text('Decline'),
                            ),
                          ],
                          if ([
                            'joined',
                            'requested',
                            'interested',
                            'needs_reconfirmation',
                          ].contains(participant['status']))
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _update(participant, 'removed'),
                              child: const Text('Remove'),
                            ),
                          if (participant['status'] == 'removed')
                            TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _update(participant, 'withdrawn'),
                              child: const Text('Allow requesting again'),
                            ),
                        ],
                      ),
                    ],
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
        onPressed: () => Navigator.pop(context),
        child: const Text('Done'),
      ),
    ],
  );
}
