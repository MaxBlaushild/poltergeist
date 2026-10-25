import 'dart:async';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';
import 'package:travel_angels/widgets/travel_trip_hero.dart';
import 'package:travel_angels/widgets/travel_trip_hosts.dart';
import 'package:travel_angels/widgets/travel_trip_logistics.dart';
import 'package:travel_angels/widgets/travel_trip_settings.dart';

/// The same canonical stop page serves public links and authenticated calendars.
/// Only the server-filtered DTO is rendered; drafts require management access.
class TravelTripScreen extends StatefulWidget {
  const TravelTripScreen({
    super.key,
    this.initialStop,
    required this.service,
    this.accessQuery = const {},
    required this.loadStop,
    required this.onAction,
    this.onVerify,
    this.onDeleted,
  });

  final Map<String, dynamic>? initialStop;
  final TravelCalendarService service;
  final Map<String, dynamic> accessQuery;
  final Future<Map<String, dynamic>> Function() loadStop;
  final Future<void> Function(String action, Map<String, dynamic> stop)
  onAction;
  final Future<void> Function()? onVerify;
  final VoidCallback? onDeleted;

  @override
  State<TravelTripScreen> createState() => _TravelTripScreenState();
}

class _TravelTripScreenState extends State<TravelTripScreen>
    with WidgetsBindingObserver {
  Map<String, dynamic>? _stop;
  String? _error;
  bool _loading = false;
  bool _busy = false;
  bool _showDraft = false;
  bool _sensitiveFlow = false;
  int _generation = 0;
  AuthProvider? _auth;
  String? _identity;
  Timer? _refreshTimer;
  final _sections = {
    for (final name in ['Overview', 'People', 'Lodging', 'Flights'])
      name: GlobalKey(),
  };

  bool get _manage => _stop?['canManage'] == true;
  bool get _owner => _identity != null && _identity == _stop?['ownerId'];
  bool get _host =>
      tripMaps(_stop?['hosts']).any((host) => host['userId'] == _identity);

  @override
  void initState() {
    super.initState();
    _stop = widget.initialStop;
    _showDraft = _stop?['published'] != true;
    WidgetsBinding.instance.addObserver(this);
    _refreshTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (_sensitiveFlow ||
          (!_busy && ModalRoute.of(context)?.isCurrent == true)) {
        _load(quiet: true);
      }
    });
    _load(quiet: _stop != null);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      _identity = auth?.user?.id;
      auth?.addListener(_authChanged);
    }
  }

  void _authChanged() {
    final next = _auth?.user?.id;
    if (next == _identity) return;
    _identity = next;
    _dismissSensitiveFlow();
    // Logout and account switches discard all protected text and image widgets.
    // Selection/claim continuation belongs to the calling calendar, not this DTO.
    setState(() {
      _stop = null;
      _showDraft = false;
    });
    _load();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) _load(quiet: true);
  }

  void _dismissSensitiveFlow() {
    if (!_sensitiveFlow || !mounted) return;
    _sensitiveFlow = false;
    final pageRoute = ModalRoute.of(context);
    if (pageRoute != null) {
      Navigator.of(
        context,
        rootNavigator: true,
      ).popUntil((route) => route == pageRoute);
    }
  }

  @override
  void dispose() {
    _generation++;
    _refreshTimer?.cancel();
    _auth?.removeListener(_authChanged);
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  Future<void> _load({bool quiet = false}) async {
    final generation = ++_generation;
    if (!quiet && mounted) {
      setState(() {
        _loading = true;
        _error = null;
      });
    }
    try {
      final stop = await widget.loadStop();
      if (!mounted || generation != _generation) return;
      if ((_manage && stop['canManage'] != true) ||
          (_stop?['canViewAttendees'] == true &&
              stop['canViewAttendees'] != true) ||
          (_stop?['rosterVersion'] != null &&
              _stop?['rosterVersion'] != stop['rosterVersion'])) {
        _dismissSensitiveFlow();
      }
      setState(() {
        _stop = stop;
        _error = null;
        _loading = false;
      });
    } catch (error) {
      if (!mounted || generation != _generation) return;
      final denied =
          error is TravelCalendarException &&
          [401, 403, 404, 410].contains(error.statusCode);
      if (denied) _dismissSensitiveFlow();
      setState(() {
        if (denied) _stop = null;
        _error = denied
            ? 'This trip is not available to your current account or link.'
            : tripError(error);
        _loading = false;
      });
    }
  }

  Future<void> _action(String action) async {
    if (_stop == null || _busy) return;
    final stop = _stop!;
    _sensitiveFlow = [
      'edit',
      'permissions',
      'share',
      'invite',
      'hosts',
      'preview',
      'participants',
      'links',
      'broadcast',
    ].contains(action);
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.onAction(action, stop);
      if (mounted) {
        if (action == 'edit') _showDraft = true;
        if (action == 'publish') _showDraft = false;
        await _load(quiet: true);
      }
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      _sensitiveFlow = false;
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _dialog(Widget dialog, {bool draft = false}) async {
    _sensitiveFlow = true;
    setState(() => _busy = true);
    final changed = await showDialog<bool>(
      context: context,
      builder: (_) => dialog,
    );
    _sensitiveFlow = false;
    if (!mounted) return;
    if (changed == true && draft) _showDraft = true;
    await _load(quiet: true);
    if (mounted) setState(() => _busy = false);
  }

  Future<void> _verify() async {
    if (widget.onVerify == null) return;
    setState(() => _busy = true);
    try {
      await widget.onVerify!();
      await _load();
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _hostAction(String action) async {
    final stop = _stop!;
    switch (action) {
      case 'hero':
        await _dialog(
          TravelTripHeroEditor(
            stop: stop,
            service: widget.service,
            accessQuery: widget.accessQuery,
          ),
          draft: true,
        );
      case 'settings':
        await _dialog(
          TravelTripSettingsDialog(stop: stop, service: widget.service),
        );
      case 'attendance':
        await _dialog(
          TravelTripConsentDialog(
            stop: stop,
            service: widget.service,
            hostAttendance: true,
          ),
        );
      case 'consent':
        await _dialog(
          TravelTripConsentDialog(stop: stop, service: widget.service),
        );
      case 'hosts':
        if (stop['hostModel'] == true) {
          await _dialog(
            TravelTripHostsDialog(
              stop: stop,
              service: widget.service,
              currentUserId: _identity ?? '',
            ),
          );
        } else {
          await _action('hosts');
        }
      case 'conversion':
        await _dialog(
          TravelTripConversionDialog(stop: stop, service: widget.service),
        );
      case 'transfer':
        await _dialog(
          TravelTripTransferDialog(stop: stop, service: widget.service),
        );
      case 'delete':
      case 'leave':
        final deleting = action == 'delete';
        final confirmed = await confirmTripImpact(
          context,
          deleting ? 'Delete this trip?' : 'Leave as a host?',
          {
            'message': deleting
                ? 'The trip and its shared links will no longer be available. This cannot be undone.'
                : 'You will lose host access. Your attendance may be removed if you no longer have permission to view the trip.',
          },
          confirmLabel: deleting ? 'Delete trip' : 'Leave host role',
        );
        if (!confirmed || !mounted) return;
        setState(() => _busy = true);
        try {
          if (deleting) {
            await widget.service.delete(
              '/calendar/stops/${stop['id']}',
              data: {'revision': stop['revision'], 'confirm': true},
            );
            if (mounted) {
              setState(() => _stop = null);
              widget.onDeleted?.call();
            }
          } else {
            await widget.service.post(
              '/calendar/stops/${stop['id']}/leave-host',
              data: {'revision': stop['revision'], 'confirmRemoval': true},
            );
            await _load();
          }
        } catch (error) {
          if (mounted) setState(() => _error = tripError(error));
        } finally {
          if (mounted) setState(() => _busy = false);
        }
      default:
        await _action(action);
    }
  }

  Future<void> _logistics(
    String kind,
    String action, [
    Map<String, dynamic>? entry,
  ]) async {
    final stop = _stop!;
    final field = kind == 'lodging' ? 'lodging' : 'flights';
    final entries = tripMaps(tripDraft(stop)[field]);
    final index = entries.indexWhere((item) => item['id'] == entry?['id']);
    if (action == 'add' || action == 'edit') {
      await _dialog(
        TravelTripLogisticsEditor(
          stop: stop,
          kind: kind,
          service: widget.service,
          entry: index < 0 ? entry : entries[index],
        ),
        draft: true,
      );
      return;
    }
    if (action == 'audience') {
      await _dialog(
        TravelTripLogisticsAudienceDialog(
          stop: stop,
          entry: entry!,
          kind: kind,
          service: widget.service,
        ),
      );
      return;
    }
    if (index < 0) return;
    if (action == 'remove') {
      final yes = await confirmTripImpact(
        context,
        'Remove ${kind == 'lodging' ? 'lodging' : 'flight'} from the draft?',
        {
          'message':
              'This changes your draft. Publish the trip when you are ready to share this removal.',
        },
        confirmLabel: 'Remove from draft',
      );
      if (!yes || !mounted) return;
      entries.removeAt(index);
    } else {
      final next = index + (action == 'up' ? -1 : 1);
      if (next < 0 || next >= entries.length) return;
      final moved = entries.removeAt(index);
      entries.insert(next, moved);
    }
    setState(() => _busy = true);
    try {
      await saveTripDraft(widget.service, stop, {field: entries});
      _showDraft = true;
      await _load(quiet: true);
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Widget _entryActions(
    String kind,
    Map<String, dynamic> entry,
    int index,
    int length,
  ) => Wrap(
    spacing: 4,
    children: [
      TextButton.icon(
        onPressed: _busy ? null : () => _logistics(kind, 'edit', entry),
        icon: const Icon(Icons.edit_outlined, size: 16),
        label: const Text('Edit draft'),
      ),
      if (_stop?['canManagePermissions'] == true && entry['id'] != null)
        TextButton.icon(
          onPressed: _busy ? null : () => _logistics(kind, 'audience', entry),
          icon: const Icon(Icons.lock_outline, size: 16),
          label: Text(tripLogisticsAudience(entry['visibility'])),
        ),
      if (index > 0)
        IconButton(
          tooltip: 'Move earlier',
          onPressed: _busy ? null : () => _logistics(kind, 'up', entry),
          icon: const Icon(Icons.arrow_upward, size: 18),
        ),
      if (index < length - 1)
        IconButton(
          tooltip: 'Move later',
          onPressed: _busy ? null : () => _logistics(kind, 'down', entry),
          icon: const Icon(Icons.arrow_downward, size: 18),
        ),
      IconButton(
        tooltip: 'Remove from draft',
        onPressed: _busy ? null : () => _logistics(kind, 'remove', entry),
        icon: const Icon(Icons.delete_outline, size: 18),
      ),
    ],
  );

  Widget _hostMenu() => PopupMenuButton<String>(
    tooltip: 'Manage trip',
    onSelected: _hostAction,
    itemBuilder: (_) => [
      const PopupMenuItem(value: 'edit', child: Text('Edit trip details')),
      const PopupMenuItem(value: 'hero', child: Text('Edit cover image')),
      if (_stop?['hostModel'] == true)
        const PopupMenuItem(
          value: 'settings',
          child: Text('Attendance & privacy'),
        ),
      const PopupMenuItem(
        value: 'participants',
        child: Text('Manage attendance'),
      ),
      if (_stop?['hostModel'] == true ||
          _stop?['canManageCollaborators'] == true)
        PopupMenuItem(
          value: 'hosts',
          child: Text(
            _stop?['hostModel'] == true
                ? 'Manage hosts'
                : 'Manage collaborators',
          ),
        ),
      if (_stop?['canManagePermissions'] == true) ...[
        const PopupMenuItem(
          value: 'permissions',
          child: Text('Who can view this trip'),
        ),
        const PopupMenuItem(value: 'share', child: Text('Share trip')),
        const PopupMenuItem(
          value: 'links',
          child: Text('Replace or disable share link'),
        ),
      ],
      const PopupMenuItem(value: 'invite', child: Text('Invite friends')),
      const PopupMenuItem(value: 'broadcast', child: Text('Send an update')),
      const PopupMenuItem(value: 'preview', child: Text('Preview as a viewer')),
      const PopupMenuDivider(),
      const PopupMenuItem(value: 'publish', child: Text('Publish draft')),
      if (_stop?['published'] == true)
        const PopupMenuItem(value: 'unpublish', child: Text('Unpublish trip')),
      if (_owner && _stop?['hostModel'] != true)
        const PopupMenuItem(
          value: 'conversion',
          child: Text('Upgrade to trip hosts'),
        ),
      if (_owner && _stop?['hostModel'] == true)
        const PopupMenuItem(
          value: 'transfer',
          child: Text('Transfer ownership'),
        ),
      if (!_owner && _host && _stop?['hostModel'] == true)
        const PopupMenuItem(value: 'leave', child: Text('Leave host role')),
      if (_owner)
        const PopupMenuItem(value: 'delete', child: Text('Delete trip')),
    ],
    child: const Padding(
      padding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.tune, size: 18),
          SizedBox(width: 8),
          Text('Manage trip'),
        ],
      ),
    ),
  );

  String? _joinAction() {
    final stop = _stop!;
    final own = tripMap(stop['ownParticipation']);
    if (stop['acceptingParticipation'] != true ||
        stop['published'] != true ||
        own['status'] == 'removed') {
      return null;
    }
    if (own['status'] == 'joined' || own['reserved'] == true) return 'request';
    if (own['status'] == 'requested') return 'request';
    if (own['status'] == 'waitlisted') return 'waitlist';
    if (stop['datePrecision'] != 'fixed') {
      return own['status'] == 'interested' ? null : 'interested';
    }
    final capacity = tripMap(stop['capacity']);
    if (capacity['full'] == true) {
      return capacity['waitlistEnabled'] == true ? 'waitlist' : null;
    }
    return 'request';
  }

  String _joinLabel(String action) {
    final own = tripMap(_stop?['ownParticipation']);
    if (own['status'] == 'requested') return 'Update request';
    if (own['status'] == 'waitlisted') return 'Update waitlist dates';
    if (action == 'waitlist') return 'Join the waitlist';
    if (action == 'interested') return 'I’m interested';
    if (own['status'] == 'needs_reconfirmation' || own['reserved'] == true) {
      return 'Review & reconfirm';
    }
    if (own['status'] == 'joined') return 'Change my dates';
    return 'Request to join';
  }

  Widget _joinButton() {
    final action = _joinAction();
    return FilledButton.icon(
      onPressed: action == null || _busy ? null : () => _action(action),
      icon: Icon(
        action == 'waitlist' ? Icons.hourglass_top_rounded : Icons.north_east,
        size: 18,
      ),
      label: Text(action == null ? _unavailableLabel() : _joinLabel(action)),
      style: FilledButton.styleFrom(
        backgroundColor: tripInk,
        foregroundColor: Colors.white,
        minimumSize: const Size.fromHeight(52),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
      ),
    );
  }

  String _unavailableLabel() {
    final stop = _stop!;
    final status = tripMap(stop['ownParticipation'])['status'];
    if (status == 'requested' || status == 'waitlisted') {
      return tripParticipationLabel(status);
    }
    if (status == 'removed') return 'Attendance unavailable';
    if (stop['status'] == 'cancelled') return 'Trip cancelled';
    if (stop['isPast'] == true) return 'This trip has ended';
    if (tripMap(stop['capacity'])['full'] == true) return 'This trip is full';
    return 'Joining is not open';
  }

  Widget _attendanceCard() {
    final stop = _stop!;
    final capacity = tripMap(stop['capacity']);
    final own = tripMap(stop['ownParticipation']);
    final showCount = capacity['showCount'] == true || _manage;
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: tripSand),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Text(
            'MAKE A LITTLE ROOM',
            style: TextStyle(
              color: tripMuted,
              fontSize: 10,
              letterSpacing: 2,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 10),
          const Text(
            'Come along.',
            style: TextStyle(
              color: tripInk,
              fontFamily: 'Georgia',
              fontSize: 32,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            tripDates(stop),
            style: const TextStyle(color: tripInk, height: 1.5),
          ),
          if (showCount && capacity['confirmedCount'] != null)
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: Text(
                '${capacity['confirmedCount']} confirmed${capacity['maximum'] != null ? ' · ${capacity['maximum']} places total' : ''}',
                style: const TextStyle(color: tripMuted),
              ),
            ),
          if (showCount && capacity['remainingPlaces'] != null)
            Text(
              '${capacity['remainingPlaces']} places available',
              style: const TextStyle(color: tripMuted),
            ),
          if (_manage && (capacity['reservedCount'] as num? ?? 0) > 0)
            Text(
              '${capacity['reservedCount']} reserved for reconfirmation',
              style: const TextStyle(color: tripMuted),
            ),
          if (own.isNotEmpty && own['status'] != 'withdrawn') ...[
            const SizedBox(height: 14),
            TripPill(
              tripParticipationLabel(own['status']),
              icon: Icons.person_outline,
            ),
            if (own['reserved'] == true)
              const Padding(
                padding: EdgeInsets.only(top: 8),
                child: Text(
                  'Your place is reserved while you reconfirm.',
                  style: TextStyle(color: tripMuted),
                ),
              ),
          ],
          const SizedBox(height: 20),
          if (_host && stop['hostModel'] == true)
            FilledButton.icon(
              onPressed: _busy ? null : () => _hostAction('attendance'),
              icon: const Icon(Icons.hiking),
              label: const Text('My host attendance'),
            )
          else
            _joinButton(),
          const SizedBox(height: 10),
          Text(
            _host
                ? 'Hosting and attending are separate. Choose whether you’re coming too.'
                : 'Choose all or part of the trip. Hosts confirm requests before you’re in.',
            style: const TextStyle(color: tripMuted, fontSize: 12, height: 1.5),
          ),
          if (!_host &&
              stop['acceptingParticipation'] == true &&
              stop['datePrecision'] == 'fixed' &&
              ![
                'joined',
                'requested',
                'interested',
                'waitlisted',
                'removed',
              ].contains(own['status']))
            TextButton(
              onPressed: _busy ? null : () => _action('interested'),
              child: const Text('Just interested for now'),
            ),
          OutlinedButton.icon(
            onPressed: _busy ? null : () => _action('follow'),
            icon: const Icon(Icons.notifications_none, size: 18),
            label: const Text('Follow trip updates'),
          ),
          if (own.isNotEmpty &&
              !['withdrawn', 'removed'].contains(own['status'])) ...[
            TextButton(
              onPressed: _busy ? null : () => _hostAction('consent'),
              child: const Text('My attendee privacy'),
            ),
            if (!_host)
              TextButton(
                onPressed: _busy ? null : () => _action('withdraw'),
                child: const Text('Withdraw my participation'),
              ),
          ],
        ],
      ),
    );
  }

  Widget _person(Map<String, dynamic> person, {bool host = false}) {
    final name = person['displayName'] as String? ?? 'Traveler';
    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        TripAvatar(name: name, url: person['avatarUrl'] as String?),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                name,
                style: const TextStyle(
                  fontWeight: FontWeight.w600,
                  color: tripInk,
                ),
              ),
              if (host)
                Text(
                  person['isOwner'] == true ? 'Trip owner' : 'Host',
                  style: const TextStyle(color: tripMuted, fontSize: 12),
                ),
              if (!host && (person['startDate'] as String?)?.isNotEmpty == true)
                Text(
                  '${person['startDate']} – ${person['endDate']}',
                  style: const TextStyle(color: tripMuted, fontSize: 12),
                ),
            ],
          ),
        ),
        if (host && person['isAttending'] == true) const TripPill('Coming too'),
      ],
    );
  }

  Widget _content(bool wide) {
    final stop = _stop!;
    final shown = _manage && _showDraft ? {...stop, ...tripDraft(stop)} : stop;
    final hosts = tripMaps(stop['hosts']);
    final attendees = tripMaps(stop['attendees']);
    final lodging = tripMaps(shown['lodging']);
    final flights = tripMaps(shown['flights']);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(wide ? 26 : 18),
          child: SizedBox(
            height: wide ? 300 : 230,
            child: TripProtectedHero(
              key: ValueKey('${stop['id']}:$_identity'),
              hero: tripMap(shown['hero']),
              service: widget.service,
              accessQuery: widget.accessQuery,
            ),
          ),
        ),
        const SizedBox(height: 28),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            TripPill(
              (shown['destination'] as String? ?? '').isEmpty
                  ? 'A new adventure'
                  : shown['destination'],
              icon: Icons.place_outlined,
            ),
            TripPill(
              shown['status'] == 'cancelled'
                  ? 'Cancelled'
                  : shown['status'] == 'confirmed'
                  ? 'Plans confirmed'
                  : 'Taking shape',
              background: const Color(0xFFE6EDDF),
            ),
            if (_manage)
              TripPill(
                _showDraft ? 'Host draft' : 'Published view',
                icon: _showDraft ? Icons.edit_note : Icons.public,
              ),
          ],
        ),
        const SizedBox(height: 18),
        Text(
          shown['title'] as String? ?? 'A trip together',
          style: TextStyle(
            fontFamily: 'Georgia',
            fontSize: wide ? 48 : 36,
            height: 1.07,
            letterSpacing: -1.2,
            color: tripInk,
          ),
        ),
        const SizedBox(height: 16),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(
              Icons.calendar_today_outlined,
              size: 18,
              color: tripMuted,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                tripDates(shown),
                style: const TextStyle(color: tripMuted, fontSize: 16),
              ),
            ),
          ],
        ),
        if (shown['datePrecision'] == 'fixed' &&
            (shown['timeZone'] as String?)?.isNotEmpty == true)
          Padding(
            padding: const EdgeInsets.only(left: 28, top: 5),
            child: Text(
              'Destination time · ${shown['timeZone']}',
              style: const TextStyle(color: tripMuted, fontSize: 12),
            ),
          ),
        if (hosts.isNotEmpty) ...[
          const SizedBox(height: 24),
          Wrap(
            spacing: 12,
            runSpacing: 8,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              for (final host in hosts.take(3))
                TripAvatar(
                  name: host['displayName'] as String? ?? 'Host',
                  url: host['avatarUrl'] as String?,
                  radius: 17,
                ),
              Text(
                'Hosted by ${hosts.map((h) => h['displayName'] ?? 'a friend').join(' & ')}',
                style: const TextStyle(color: tripMuted),
              ),
            ],
          ),
        ],
        const SizedBox(height: 24),
        if (_manage)
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: tripSand,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Wrap(
              spacing: 12,
              runSpacing: 8,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                Text(
                  _showDraft
                      ? 'You’re viewing your saved draft.'
                      : 'You’re viewing published content.',
                  style: const TextStyle(color: tripInk),
                ),
                if (stop['published'] == true)
                  TextButton(
                    onPressed: () => setState(() => _showDraft = !_showDraft),
                    child: Text(_showDraft ? 'View published' : 'View draft'),
                  ),
                if (_showDraft)
                  FilledButton(
                    onPressed: _busy ? null : () => _action('publish'),
                    child: const Text('Publish draft'),
                  ),
              ],
            ),
          ),
        if (stop['ownRosterConsentRequired'] == true)
          Padding(
            padding: const EdgeInsets.only(top: 16),
            child: TripEmptyState(
              icon: Icons.privacy_tip_outlined,
              title: 'Review your attendee privacy',
              message:
                  'The audience for the attendee list has changed. Your identity stays hidden until you acknowledge the new audience.',
              action: TextButton(
                onPressed: _busy ? null : () => _hostAction('consent'),
                child: const Text('Review privacy'),
              ),
            ),
          ),
        const Divider(color: tripSand),
        TripSection(
          key: _sections['Overview'],
          eyebrow: 'The plan',
          title: 'Good days ahead',
          child: (shown['description'] as String? ?? '').trim().isNotEmpty
              ? TripDescription(shown['description'])
              : const Text(
                  'The details are still taking shape. Follow along as the plan comes together.',
                  style: TextStyle(color: tripMuted, height: 1.6),
                ),
        ),
        if (!wide) _attendanceCard(),
        const Divider(color: tripSand),
        TripSection(
          key: _sections['People'],
          eyebrow: 'Better together',
          title: 'The people',
          action: _manage
              ? TextButton(
                  onPressed: _busy ? null : () => _action('participants'),
                  child: const Text('Manage attendance'),
                )
              : null,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (hosts.isNotEmpty) ...[
                const Text(
                  'YOUR HOSTS',
                  style: TextStyle(
                    fontSize: 10,
                    color: tripMuted,
                    letterSpacing: 1.8,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 16),
                for (final host in hosts)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: _person(host, host: true),
                  ),
                const SizedBox(height: 12),
              ],
              if (stop['canViewAttendees'] == true && attendees.isNotEmpty) ...[
                const Text(
                  'CONFIRMED TRAVELERS',
                  style: TextStyle(
                    fontSize: 10,
                    color: tripMuted,
                    letterSpacing: 1.8,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 16),
                for (final person in attendees)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: _person(person),
                  ),
              ] else
                TripEmptyState(
                  icon: Icons.people_outline,
                  title: stop['canViewAttendees'] == true
                      ? 'The group is coming together'
                      : 'A little privacy for the group',
                  message: stop['canViewAttendees'] == true
                      ? 'Confirmed travelers who have acknowledged the attendee audience appear here.'
                      : 'Attendee identities are shared only with the audience chosen by the hosts. You can still browse the plan and request to join.',
                ),
            ],
          ),
        ),
        const Divider(color: tripSand),
        TripSection(
          key: _sections['Lodging'],
          eyebrow: 'Settle in',
          title: 'Somewhere to stay',
          action: _manage
              ? TextButton.icon(
                  onPressed: _busy ? null : () => _logistics('lodging', 'add'),
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Add lodging'),
                )
              : null,
          child: lodging.isEmpty
              ? const TripEmptyState(
                  icon: Icons.bed_outlined,
                  title: 'Room for a plan',
                  message:
                      'Places to stay will appear here when the hosts share them with you.',
                )
              : Column(
                  children: [
                    for (var i = 0; i < lodging.length; i++)
                      Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: TravelTripLodgingCard(
                          entry: lodging[i],
                          actions: _manage
                              ? _entryActions(
                                  'lodging',
                                  lodging[i],
                                  i,
                                  lodging.length,
                                )
                              : null,
                        ),
                      ),
                  ],
                ),
        ),
        const Divider(color: tripSand),
        TripSection(
          key: _sections['Flights'],
          eyebrow: 'Getting there',
          title: 'The way there',
          action: _manage
              ? TextButton.icon(
                  onPressed: _busy ? null : () => _logistics('flight', 'add'),
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Add flights'),
                )
              : null,
          child: flights.isEmpty
              ? const TripEmptyState(
                  icon: Icons.flight_takeoff,
                  title: 'Meet you there',
                  message:
                      'Shared flight ideas and group plans will appear here when available to you.',
                )
              : Column(
                  children: [
                    for (var i = 0; i < flights.length; i++)
                      Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: TravelTripFlightCard(
                          entry: flights[i],
                          actions: _manage
                              ? _entryActions(
                                  'flight',
                                  flights[i],
                                  i,
                                  flights.length,
                                )
                              : null,
                        ),
                      ),
                  ],
                ),
        ),
        if (_owner && stop['hostModel'] != true)
          TripEmptyState(
            icon: Icons.groups_outlined,
            title: 'Give your trip its own hosts',
            message:
                'This trip still uses its existing calendar and collaborator permissions. Review who should become a full trip host before upgrading.',
            action: TextButton(
              onPressed: _busy ? null : () => _hostAction('conversion'),
              child: const Text('Review host upgrade'),
            ),
          ),
        const SizedBox(height: 36),
        const Center(
          child: Text(
            'TRAVEL ANGELS · GO TOGETHER',
            style: TextStyle(color: tripMuted, fontSize: 10, letterSpacing: 2),
          ),
        ),
        const SizedBox(height: 36),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final localTheme = Theme.of(context).copyWith(
      colorScheme: ColorScheme.fromSeed(seedColor: tripInk, surface: tripPaper),
      scaffoldBackgroundColor: tripPaper,
      dividerColor: tripSand,
    );
    return Theme(
      data: localTheme,
      child: Builder(
        builder: (context) => LayoutBuilder(
          builder: (context, constraints) {
            final wide = constraints.maxWidth >= 1000;
            return Scaffold(
              backgroundColor: tripPaper,
              appBar: AppBar(
                backgroundColor: tripPaper,
                foregroundColor: tripInk,
                surfaceTintColor: Colors.transparent,
                title: const Text(
                  'travel angels',
                  style: TextStyle(
                    fontFamily: 'Georgia',
                    fontSize: 23,
                    letterSpacing: -.5,
                  ),
                ),
                actions: [
                  IconButton(
                    tooltip: 'Refresh trip',
                    onPressed: _busy ? null : () => _load(),
                    icon: const Icon(Icons.refresh, size: 21),
                  ),
                  if (_manage && !_busy) _hostMenu(),
                  const SizedBox(width: 12),
                ],
                bottom: _stop == null
                    ? null
                    : PreferredSize(
                        preferredSize: const Size.fromHeight(48),
                        child: SingleChildScrollView(
                          scrollDirection: Axis.horizontal,
                          child: Row(
                            children: [
                              for (final name in _sections.keys)
                                TextButton(
                                  onPressed: () {
                                    final sectionContext =
                                        _sections[name]!.currentContext;
                                    if (sectionContext != null) {
                                      Scrollable.ensureVisible(
                                        sectionContext,
                                        duration:
                                            MediaQuery.disableAnimationsOf(
                                              context,
                                            )
                                            ? Duration.zero
                                            : const Duration(milliseconds: 250),
                                        alignment: .05,
                                      );
                                    }
                                  },
                                  child: Text(name),
                                ),
                            ],
                          ),
                        ),
                      ),
              ),
              body: _stop == null
                  ? Center(
                      child: Padding(
                        padding: const EdgeInsets.all(24),
                        child: _loading
                            ? const CircularProgressIndicator()
                            : ConstrainedBox(
                                constraints: const BoxConstraints(
                                  maxWidth: 480,
                                ),
                                child: TripEmptyState(
                                  icon: Icons.travel_explore,
                                  title: 'Let’s find your trip',
                                  message:
                                      _error ??
                                      'This trip is no longer available.',
                                  action: Wrap(
                                    spacing: 8,
                                    children: [
                                      TextButton(
                                        onPressed: () => _load(),
                                        child: const Text('Try again'),
                                      ),
                                      if (widget.onVerify != null)
                                        FilledButton(
                                          onPressed: _busy ? null : _verify,
                                          child: const Text(
                                            'Verify invited email',
                                          ),
                                        ),
                                    ],
                                  ),
                                ),
                              ),
                      ),
                    )
                  : Column(
                      children: [
                        if (_loading)
                          const LinearProgressIndicator(minHeight: 2),
                        if (_error != null)
                          MaterialBanner(
                            content: Text(_error!),
                            actions: [
                              TextButton(
                                onPressed: () => _load(),
                                child: const Text('Refresh'),
                              ),
                            ],
                          ),
                        Expanded(
                          child: Center(
                            child: ConstrainedBox(
                              constraints: const BoxConstraints(maxWidth: 1200),
                              child: wide
                                  ? Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Expanded(
                                          child: SingleChildScrollView(
                                            padding: const EdgeInsets.fromLTRB(
                                              28,
                                              16,
                                              32,
                                              28,
                                            ),
                                            child: _content(true),
                                          ),
                                        ),
                                        SizedBox(
                                          width: 330,
                                          child: SingleChildScrollView(
                                            padding: const EdgeInsets.fromLTRB(
                                              0,
                                              16,
                                              24,
                                              24,
                                            ),
                                            child: _attendanceCard(),
                                          ),
                                        ),
                                      ],
                                    )
                                  : SingleChildScrollView(
                                      padding: const EdgeInsets.fromLTRB(
                                        20,
                                        12,
                                        20,
                                        24,
                                      ),
                                      child: _content(false),
                                    ),
                            ),
                          ),
                        ),
                      ],
                    ),
              bottomNavigationBar: !wide && _stop != null && !_host
                  ? SafeArea(
                      child: Container(
                        padding: const EdgeInsets.fromLTRB(20, 10, 20, 10),
                        decoration: BoxDecoration(
                          color: tripPaper,
                          border: Border(top: BorderSide(color: tripSand)),
                        ),
                        child: _joinButton(),
                      ),
                    )
                  : null,
            );
          },
        ),
      ),
    );
  }
}
