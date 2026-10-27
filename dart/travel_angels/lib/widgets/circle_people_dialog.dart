import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/friend_invite.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/circle_person_row.dart';

class CirclePeopleDialog extends StatefulWidget {
  const CirclePeopleDialog({
    super.key,
    required this.service,
    required this.friends,
    required this.invites,
  });

  final FriendService service;
  final List<User> friends;
  final List<FriendInvite> invites;

  @override
  State<CirclePeopleDialog> createState() => _CirclePeopleDialogState();
}

class _CirclePeopleDialogState extends State<CirclePeopleDialog> {
  final _search = TextEditingController();
  final _sent = <String>{};
  final _sending = <String>{};
  final _inviteErrors = <String, String>{};
  AuthProvider? _auth;
  String? _identity;
  bool _initialized = false;
  bool _invalidated = false;
  bool _loading = false;
  String? _searchError;
  List<User> _results = [];
  Timer? _debounce;
  int _request = 0;

  bool get _current =>
      !_invalidated &&
      _identity?.isNotEmpty == true &&
      _auth?.user?.id == _identity;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider>();
    if (_auth != auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      auth.addListener(_authChanged);
    }
    if (!_initialized) {
      _initialized = true;
      _identity = auth.user?.id;
    }
    _authChanged();
  }

  void _authChanged() {
    if (!mounted || _current || _invalidated) return;
    _invalidated = true;
    _debounce?.cancel();
    ++_request;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      final route = ModalRoute.of(context);
      if (route != null && route.isActive) {
        Navigator.of(context).removeRoute(route);
      } else {
        setState(() {});
      }
    });
  }

  @override
  void dispose() {
    _auth?.removeListener(_authChanged);
    _debounce?.cancel();
    ++_request;
    _search.dispose();
    super.dispose();
  }

  void _changed(String value) {
    _debounce?.cancel();
    final request = ++_request;
    final query = value.trim().replaceFirst(RegExp(r'^@'), '');
    setState(() {
      _loading = query.isNotEmpty && _current;
      _searchError = null;
      _results = [];
    });
    if (query.isEmpty || !_current) return;
    _debounce = Timer(
      const Duration(milliseconds: 350),
      () => _find(query, request),
    );
  }

  Future<void> _find(String query, int request) async {
    if (!mounted || !_current || request != _request) return;
    setState(() {
      _loading = true;
      _searchError = null;
    });
    try {
      final people = await widget.service.searchUsers(query);
      if (!mounted || !_current || request != _request) return;
      final seen = <String>{};
      setState(() {
        _results = people.where((person) {
          final id = person.id;
          return id != null && id.isNotEmpty && id != _identity && seen.add(id);
        }).toList();
      });
    } catch (_) {
      if (!mounted || !_current || request != _request) return;
      setState(() => _searchError = 'Could not search people. Try again.');
    } finally {
      if (mounted && _current && request == _request) {
        setState(() => _loading = false);
      }
    }
  }

  bool _friend(String id) => widget.friends.any((friend) => friend.id == id);
  bool _incoming(String id) => widget.invites.any(
    (invite) => invite.inviteeId == _identity && invite.inviterId == id,
  );
  bool _outgoing(String id) =>
      _sent.contains(id) ||
      widget.invites.any(
        (invite) => invite.inviterId == _identity && invite.inviteeId == id,
      );

  Future<void> _invite(User person) async {
    final id = person.id;
    if (!_current ||
        id == null ||
        id.isEmpty ||
        id == _identity ||
        _friend(id) ||
        _incoming(id) ||
        _outgoing(id) ||
        _sending.contains(id)) {
      return;
    }
    setState(() {
      _sending.add(id);
      _inviteErrors.remove(id);
    });
    try {
      await widget.service.createFriendInvite(id);
      if (!mounted || !_current) return;
      setState(() => _sent.add(id));
    } catch (error) {
      if (!mounted || !_current) return;
      final data = error is DioException ? error.response?.data : null;
      final message = data is Map ? data['error'] : null;
      setState(() {
        _inviteErrors[id] = message is String && message.isNotEmpty
            ? message
            : 'Invite could not be sent. Try again.';
      });
    } finally {
      if (mounted && _current) setState(() => _sending.remove(id));
    }
  }

  Widget _action(User person) {
    final id = person.id!;
    if (_friend(id) || _outgoing(id)) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Text(
          _friend(id) ? 'In your circle' : 'Invite sent',
          key: ValueKey('circle-state-$id'),
          style: Theme.of(context).textTheme.bodySmall,
        ),
      );
    }
    if (_incoming(id)) {
      return TextButton(
        key: ValueKey('circle-review-$id'),
        onPressed: () => Navigator.of(context).pop('requests'),
        child: const Text('Review request'),
      );
    }
    return OutlinedButton.icon(
      key: ValueKey('circle-invite-$id'),
      onPressed: _sending.contains(id) ? null : () => _invite(person),
      icon: Icon(
        _sending.contains(id) ? Icons.hourglass_top : Icons.person_add_alt_1,
        size: 18,
      ),
      label: Text(
        _sending.contains(id)
            ? 'Sending…'
            : _inviteErrors.containsKey(id)
            ? 'Retry invite'
            : 'Invite',
      ),
      style: OutlinedButton.styleFrom(
        minimumSize: const Size(44, 44),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      ),
    );
  }

  Widget _body() {
    if (_search.text.trim().replaceFirst(RegExp(r'^@'), '').isEmpty) {
      return const Text('Search by username to invite someone to your circle.');
    }
    if (_loading) {
      return Semantics(
        label: 'Searching people',
        child: const LinearProgressIndicator(minHeight: 2),
      );
    }
    if (_searchError != null) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Semantics(
            liveRegion: true,
            child: Text(
              _searchError!,
              style: const TextStyle(color: AppTheme.error),
            ),
          ),
          TextButton(
            key: const ValueKey('circle-people-retry'),
            onPressed: () => _find(
              _search.text.trim().replaceFirst(RegExp(r'^@'), ''),
              ++_request,
            ),
            child: const Text('Retry search'),
          ),
        ],
      );
    }
    if (_results.isEmpty) {
      return const Text('No new people found. Try another username.');
    }
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        for (final person in _results) ...[
          CirclePersonRow(user: person, trailing: _action(person)),
          if (_inviteErrors[person.id] != null)
            Align(
              alignment: Alignment.centerLeft,
              child: Semantics(
                liveRegion: true,
                child: Text(
                  _inviteErrors[person.id]!,
                  style: const TextStyle(color: AppTheme.error),
                ),
              ),
            ),
          if (person != _results.last) const Divider(height: 1),
        ],
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    if (!_current) return const SizedBox.shrink();
    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 560),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 12, 8, 8),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      'Find people',
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                  ),
                  IconButton(
                    key: const ValueKey('circle-people-close'),
                    tooltip: 'Close people search',
                    onPressed: () => Navigator.of(context).pop(),
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: TextField(
                key: const ValueKey('circle-people-search'),
                controller: _search,
                onChanged: _changed,
                autocorrect: false,
                textInputAction: TextInputAction.search,
                onSubmitted: (query) {
                  _debounce?.cancel();
                  final normalized = query.trim().replaceFirst(
                    RegExp(r'^@'),
                    '',
                  );
                  if (normalized.isNotEmpty) _find(normalized, ++_request);
                },
                decoration: InputDecoration(
                  labelText: 'Username',
                  prefixIcon: const Icon(Icons.search, size: 20),
                  suffixIcon: _search.text.isEmpty
                      ? null
                      : IconButton(
                          tooltip: 'Clear search',
                          onPressed: () {
                            _search.clear();
                            _changed('');
                          },
                          icon: const Icon(Icons.close, size: 18),
                        ),
                ),
              ),
            ),
            Flexible(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(20),
                child: _body(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
