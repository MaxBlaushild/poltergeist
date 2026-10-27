import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/screens/travel_trip_screen.dart';
import 'package:travel_angels/widgets/travel_account_claim.dart';
import 'package:travel_angels/widgets/travel_calendar_delivery.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';
import 'package:travel_angels/widgets/travel_trip_settings.dart';

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
  bool _monthView = true;
  bool _showFilters = false;
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
        calendars: [
          _calendar,
          ..._editableCalendars,
        ].where((calendar) => calendar['canCreateTrips'] != false).toList(),
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
      calendarId: (stop?['calendarId'] ?? _calendar['id']) as String,
      stop: stop,
    ),
  );

  Future<void> _broadcast([TravelStop? stop]) => showDialog<void>(
    context: context,
    builder: (context) => TravelBroadcastDialog(
      service: _service,
      calendarId: (stop?['calendarId'] ?? _calendar['id']) as String,
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
        data: {
          'enabled': true,
          'rotate': false,
          if (stop?['revision'] != null) 'revision': stop!['revision'],
        },
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
      if (stop?['revision'] != null) 'revision': stop!['revision'],
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

  Future<void> _selectStop(
    TravelStop stop, {
    String? preset,
    bool submitImmediately = false,
  }) async {
    final id = stop['id'] as String;
    final prior =
        _selections[id] ??
        (stop['ownParticipation'] is Map
            ? Map<String, dynamic>.from(stop['ownParticipation'] as Map)
            : <String, dynamic>{});
    String action =
        preset ??
        ([
              'requested',
              'joined',
              'needs_reconfirmation',
            ].contains(prior['status'])
            ? 'requested'
            : 'interested');
    final settings = stop['tripSettings'] is Map
        ? stop['tripSettings'] as Map
        : const {};
    final capacity = stop['capacity'] is Map
        ? stop['capacity'] as Map
        : const {};
    final canRequest =
        capacity['full'] != true ||
        prior['reserved'] == true ||
        [
          'joined',
          'requested',
          'needs_reconfirmation',
        ].contains(prior['status']);
    final canWaitlist =
        settings['waitlistEnabled'] == true || prior['status'] == 'waitlisted';
    if (action == 'requested' && !canRequest) {
      action = canWaitlist ? 'waitlisted' : 'interested';
    }
    if (action == 'waitlisted' && !canWaitlist) action = 'interested';
    final rosterVersion = stop['rosterVersion'] as int?;
    bool rosterConsent =
        rosterVersion != null && prior['rosterConsentVersion'] == rosterVersion;
    bool shareDates = prior['shareDates'] == true;
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
                      if (fixed && range != null && canRequest)
                        const DropdownMenuItem(
                          value: 'requested',
                          child: Text('Request to join'),
                        ),
                      if (fixed && range != null && canWaitlist)
                        const DropdownMenuItem(
                          value: 'waitlisted',
                          child: Text('Join waitlist'),
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
                  if (rosterVersion != null) ...[
                    CheckboxListTile(
                      contentPadding: EdgeInsets.zero,
                      value: rosterConsent,
                      onChanged: (value) =>
                          setDialogState(() => rosterConsent = value ?? false),
                      title: Text(
                        settings['rosterVisibility'] == 'hosts'
                            ? 'Trip hosts can see my name and avatar.'
                            : 'Trip hosts and confirmed attendees can see my name and avatar when I attend.',
                      ),
                      subtitle: const Text(
                        'Your contact details stay private. Required to request attendance or join the waitlist.',
                      ),
                    ),
                    CheckboxListTile(
                      contentPadding: EdgeInsets.zero,
                      value: shareDates,
                      onChanged: (value) =>
                          setDialogState(() => shareDates = value ?? false),
                      title: const Text(
                        'Also share my selected dates with confirmed attendees',
                      ),
                      subtitle: const Text(
                        'Optional. Hosts can always see the dates you request.',
                      ),
                    ),
                  ],
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
              onPressed:
                  rosterVersion != null &&
                      action != 'interested' &&
                      !rosterConsent
                  ? null
                  : () => Navigator.pop(context, {
                      'stopId': id,
                      'status': action,
                      'startDate': range == null
                          ? ''
                          : travelDateValue(range!.start),
                      'endDate': range == null
                          ? ''
                          : travelDateValue(range!.end),
                      if (rosterConsent) 'rosterConsentVersion': rosterVersion,
                      'shareDates': shareDates,
                    }),
              child: Text(submitImmediately ? 'Continue' : 'Add to selections'),
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
      if (submitImmediately) {
        await _reviewSelections(onlyStopId: id);
      } else {
        _message('Selection saved. Review and submit when ready.');
      }
    }
  }

  Future<void> _reviewSelections({String? onlyStopId}) async {
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
    final reviewSelections = Map<String, Map<String, dynamic>>.fromEntries(
      _visibleSelections.entries.where(
        (entry) => onlyStopId == null || entry.key == onlyStopId,
      ),
    );
    if (reviewSelections.isEmpty) return;
    for (final entry in reviewSelections.entries) {
      final current = _stops
          .where((stop) => stop['id'] == entry.key)
          .firstOrNull;
      if (current?['rosterVersion'] != null &&
          entry.value['status'] != 'interested' &&
          entry.value['rosterConsentVersion'] != current!['rosterVersion']) {
        _message(
          'The attendee privacy setting changed. Review it before submitting.',
        );
        await _selectStop(current, submitImmediately: true);
        return;
      }
    }
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
          data: {if (stop['revision'] != null) 'revision': stop['revision']},
        );
      });
    }
  }

  Future<void> _participants(TravelStop stop) async {
    await showDialog<bool>(
      context: context,
      builder: (context) =>
          TravelTripParticipantsDialog(service: _service, stop: stop),
    );
    if (mounted) await _load();
  }

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

  Future<void> _acceptHost(Map<String, dynamic> invitation) async {
    bool attending = false;
    bool consent = false;
    final fullHost = invitation['isHost'] == true;
    final version = invitation['rosterVersion'] as int?;
    final choice = await showDialog<bool>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, update) => AlertDialog(
          title: const Text('Accept hosting invitation'),
          content: SizedBox(
            width: 440,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  fullHost
                      ? 'As a host you can edit this trip, manage its audience, invite hosts, and approve attendees.'
                      : 'Accept the collaboration permissions granted by the trip owner.',
                ),
                if (fullHost) ...[
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    value: attending,
                    onChanged: version == null
                        ? null
                        : (value) => update(() => attending = value ?? false),
                    title: const Text('I am also attending'),
                    subtitle: const Text(
                      'Attending hosts count toward the trip maximum. You can change this on the trip page.',
                    ),
                  ),
                  if (attending)
                    CheckboxListTile(
                      contentPadding: EdgeInsets.zero,
                      value: consent,
                      onChanged: (value) =>
                          update(() => consent = value ?? false),
                      title: Text(
                        invitation['rosterVisibility'] == 'hosts'
                            ? 'Hosts can see my attendance.'
                            : 'Hosts and confirmed attendees can see my name and avatar.',
                      ),
                    ),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: attending && !consent
                  ? null
                  : () => Navigator.pop(context, true),
              child: const Text('Accept'),
            ),
          ],
        ),
      ),
    );
    if (choice != true || !mounted) return;
    await _mutate(() async {
      await _service.post(
        '/calendar/collaborators/${invitation['id']}/accept',
        data: {
          'attending': attending,
          if (consent && version != null) 'rosterConsentVersion': version,
          'shareDates': false,
        },
      );
    });
  }

  Future<TravelStop> _loadTrip([String? stopId]) async {
    final accountId = _auth?.user?.id;
    final generation = _identityGeneration;
    final result = await _service.get(
      widget.stopToken != null
          ? '/shared/stops/${Uri.encodeComponent(widget.stopToken!)}'
          : '/calendar/stops/${Uri.encodeComponent(stopId!)}',
      query: _accessQuery,
    );
    if (!mounted ||
        generation != _identityGeneration ||
        accountId != _auth?.user?.id ||
        result['stop'] is! Map) {
      throw const TravelCalendarException(
        'This trip is unavailable or requires access.',
      );
    }
    final trip = Map<String, dynamic>.from(result['stop'] as Map);
    if (mounted) {
      final index = _stops.indexWhere((s) => s['id'] == trip['id']);
      setState(() {
        if (index < 0) {
          _stops.add(trip);
        } else {
          _stops[index] = trip;
        }
      });
    }
    return trip;
  }

  Future<void> _showStop(TravelStop stop) async {
    await Navigator.of(context).push<void>(
      MaterialPageRoute(
        builder: (_) => TravelTripScreen(
          initialStop: stop,
          service: _service,
          accessQuery: _accessQuery,
          loadStop: () => _loadTrip(stop['id'] as String),
          onAction: _tripAction,
          onVerify: _verify,
          onDeleted: () {
            Navigator.of(context).pop();
          },
        ),
      ),
    );
    if (mounted) await _load();
  }

  Future<void> _tripAction(String action, TravelStop stop) async {
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
      case 'hosts':
        await _collaborators(stop);
      case 'invite':
        await _invitations(stop);
      case 'broadcast':
        await _broadcast(stop);
      case 'select':
        await _selectStop(stop);
      case 'request':
      case 'interested':
      case 'waitlist':
        await _selectStop(
          stop,
          preset: action == 'request'
              ? 'requested'
              : action == 'waitlist'
              ? 'waitlisted'
              : 'interested',
          submitImmediately: true,
        );
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

  void _calendarAction(String action) {
    switch (action) {
      case 'follow':
        _follow();
      case 'share':
        _share();
      case 'invite':
        _invitations();
      case 'broadcast':
        _broadcast();
      case 'verify':
        _verify();
      case 'rename':
        _rename();
      case 'permissions':
        _permissions();
      case 'editors':
        _collaborators();
      case 'links':
        _manageLinks();
    }
  }

  Widget _calendarHeading() {
    final theme = Theme.of(context);
    final buttonStyle = ButtonStyle(
      minimumSize: const WidgetStatePropertyAll(Size(0, 36)),
      padding: const WidgetStatePropertyAll(
        EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      ),
      textStyle: WidgetStatePropertyAll(theme.textTheme.labelMedium),
    );
    final controls = Wrap(
      spacing: 6,
      runSpacing: 4,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        if (_canManageCalendar && _calendar['canCreateTrips'] != false)
          FilledButton.icon(
            style: buttonStyle,
            onPressed: _busy ? null : _edit,
            icon: const Icon(Icons.add, size: 18),
            label: const Text('Create trip'),
          ),
        if (_shared)
          OutlinedButton.icon(
            style: buttonStyle,
            onPressed: _busy ? null : _follow,
            icon: const Icon(Icons.notifications_none, size: 18),
            label: const Text('Follow updates'),
          ),
        PopupMenuButton<bool>(
          tooltip: 'Calendar view',
          initialValue: _monthView,
          onSelected: (value) => setState(() => _monthView = value),
          itemBuilder: (context) => [
            CheckedPopupMenuItem(
              value: true,
              checked: _monthView,
              child: const Text('Month'),
            ),
            CheckedPopupMenuItem(
              value: false,
              checked: !_monthView,
              child: const Text('Agenda'),
            ),
          ],
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 12),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _monthView ? 'Month' : 'Agenda',
                  style: theme.textTheme.labelMedium,
                ),
                const SizedBox(width: 4),
                const Icon(Icons.expand_more, size: 18),
              ],
            ),
          ),
        ),
        IconButton(
          tooltip: 'Filter trips',
          isSelected: _showFilters,
          onPressed: () => setState(() => _showFilters = !_showFilters),
          icon: const Icon(Icons.tune, size: 20),
          selectedIcon: const Icon(Icons.tune, size: 20),
          style: IconButton.styleFrom(
            backgroundColor: _showFilters ? AppTheme.peach : null,
          ),
        ),
      ],
    );
    if (!_shared) return controls;
    return LayoutBuilder(
      builder: (context, constraints) {
        final title = Text(
          _calendar['title'] as String? ?? 'Travel plans',
          style: theme.textTheme.titleLarge,
        );
        if (constraints.maxWidth < 700 ||
            MediaQuery.textScalerOf(context).scale(14) > 20) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [title, const SizedBox(height: 4), controls],
          );
        }
        return Row(
          children: [
            Expanded(child: title),
            const SizedBox(width: 16),
            controls,
          ],
        );
      },
    );
  }

  Widget _calendarFilters() => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      if (_showFilters) ...[
        const SizedBox(height: 10),
        TextField(
          controller: _search,
          decoration: const InputDecoration(
            labelText: 'Filter by destination or title',
            prefixIcon: Icon(Icons.search, size: 20),
            isDense: true,
          ),
          onChanged: (_) => setState(() {}),
        ),
        const SizedBox(height: 4),
        Wrap(
          spacing: 12,
          runSpacing: 4,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            SizedBox(
              width: MediaQuery.sizeOf(context).width < 600
                  ? MediaQuery.sizeOf(context).width -
                        (MediaQuery.sizeOf(context).width < 360 ? 8 : 24)
                  : 260,
              child: DropdownButton<String>(
                isExpanded: true,
                value: _period,
                items: const [
                  DropdownMenuItem(value: 'all', child: Text('All dates')),
                  DropdownMenuItem(value: 'upcoming', child: Text('Upcoming')),
                  DropdownMenuItem(
                    value: 'current',
                    child: Text('Happening now'),
                  ),
                  DropdownMenuItem(value: 'past', child: Text('Past')),
                  DropdownMenuItem(
                    value: 'unscheduled',
                    child: Text('Approximate / TBD'),
                  ),
                ],
                onChanged: (value) => setState(() => _period = value!),
              ),
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
              icon: const Icon(Icons.date_range, size: 18),
              label: const Text('Date range'),
            ),
          ],
        ),
      ],
      if (_search.text.isNotEmpty || _period != 'all' || _dateFilter != null)
        Padding(
          padding: const EdgeInsets.only(top: 6),
          child: Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              if (_search.text.isNotEmpty)
                InputChip(
                  label: Text(_search.text, maxLines: 2),
                  deleteButtonTooltipMessage: 'Clear destination filter',
                  onDeleted: () => setState(_search.clear),
                ),
              if (_period != 'all')
                InputChip(
                  label: Text(switch (_period) {
                    'upcoming' => 'Upcoming',
                    'current' => 'Happening now',
                    'past' => 'Past',
                    _ => 'Approximate / TBD',
                  }),
                  deleteButtonTooltipMessage: 'Clear period filter',
                  onDeleted: () => setState(() => _period = 'all'),
                ),
              if (_dateFilter != null)
                InputChip(
                  label: Text(
                    '${travelDateValue(_dateFilter!.start)} – ${travelDateValue(_dateFilter!.end)}',
                    maxLines: 2,
                  ),
                  deleteButtonTooltipMessage: 'Clear date filter',
                  onDeleted: () => setState(() => _dateFilter = null),
                ),
            ],
          ),
        ),
    ],
  );

  @override
  Widget build(BuildContext context) {
    if (widget.stopToken != null) {
      return TravelTripScreen(
        service: _service,
        accessQuery: _accessQuery,
        loadStop: _loadTrip,
        onAction: _tripAction,
        onVerify: _verify,
      );
    }
    final theme = Theme.of(context);
    final visibleSelections = _visibleSelections;
    final filtered = filterTravelStops(
      _stops,
      destination: _search.text,
      period: _period,
      range: _dateFilter,
    );
    return Scaffold(
      backgroundColor: AppTheme.paper,
      appBar: AppBar(
        title: _shared
            ? const TravelAngelsWordmark()
            : Text(_calendar['title'] as String? ?? 'Travel calendar'),
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
          if (_shared || _canManageCalendar || _canManageCalendarPermissions)
            PopupMenuButton<String>(
              tooltip: 'Calendar actions',
              enabled: !_busy,
              onSelected: _calendarAction,
              itemBuilder: (context) => [
                if (_shared) ...[
                  const PopupMenuItem(
                    value: 'follow',
                    child: Text('Follow / manage updates'),
                  ),
                  const PopupMenuItem(
                    value: 'verify',
                    child: Text('Verify invited email'),
                  ),
                ],
                if (_canManageCalendarPermissions)
                  const PopupMenuItem(
                    value: 'share',
                    child: Text('Share calendar'),
                  ),
                if (_canManageCalendar) ...[
                  const PopupMenuItem(
                    value: 'invite',
                    child: Text('Invite friends'),
                  ),
                  const PopupMenuItem(
                    value: 'broadcast',
                    child: Text('Broadcast update'),
                  ),
                ],
                if (_canManageCalendarPermissions) ...[
                  const PopupMenuDivider(),
                  const PopupMenuItem(
                    value: 'rename',
                    child: Text('Rename calendar'),
                  ),
                  const PopupMenuItem(
                    value: 'permissions',
                    child: Text('Default audience'),
                  ),
                  const PopupMenuItem(
                    value: 'editors',
                    child: Text('Calendar editors'),
                  ),
                  const PopupMenuItem(
                    value: 'links',
                    child: Text('Manage share link'),
                  ),
                ],
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
                  physics: const AlwaysScrollableScrollPhysics(),
                  padding: EdgeInsets.symmetric(
                    horizontal: MediaQuery.sizeOf(context).width > 1320
                        ? (MediaQuery.sizeOf(context).width - 1280) / 2
                        : MediaQuery.sizeOf(context).width < 360
                        ? 4
                        : MediaQuery.sizeOf(context).width < 600
                        ? 12
                        : 24,
                    vertical: 16,
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
                      _calendarHeading(),
                      _calendarFilters(),
                      const SizedBox(height: 12),
                      for (final invitation in _cohostInvitations)
                        Card(
                          color: AppTheme.peach,
                          child: ListTile(
                            leading: const Icon(Icons.group_add_outlined),
                            title: Text(
                              (invitation['tripTitle'] ?? invitation['title'])
                                      as String? ??
                                  'Hosting invitation',
                            ),
                            subtitle: Text(
                              invitation['isHost'] == true
                                  ? '${invitation['ownerDisplayName'] ?? 'The trip owner'} invited you to host. Review the role and choose whether you’re attending.'
                                  : 'Accept to show this shared trip on your calendar.',
                            ),
                            trailing: TextButton(
                              onPressed: _busy
                                  ? null
                                  : () => _acceptHost(invitation),
                              child: const Text('Accept'),
                            ),
                          ),
                        ),
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
                          color: AppTheme.peach,
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                Text(
                                  '${visibleSelections.length} ${visibleSelections.length == 1 ? 'stop' : 'stops'} selected · nothing submitted',
                                  style: theme.textTheme.titleLarge,
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
                      if (_monthView)
                        TravelCalendarMonth(
                          month: _month,
                          stops: filtered,
                          onMonthChanged: (month) =>
                              setState(() => _month = month),
                          onStopTap: _showStop,
                        )
                      else if (filtered.isEmpty)
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 36),
                          child: Column(
                            children: [
                              Icon(
                                Icons.explore_outlined,
                                size: 48,
                                color: AppTheme.rust,
                              ),
                              const SizedBox(height: 16),
                              Text(
                                _stops.isEmpty
                                    ? 'No travel stops to show yet'
                                    : 'No plans match these filters',
                                style: theme.textTheme.headlineSmall,
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
                            style: theme.textTheme.headlineSmall,
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
