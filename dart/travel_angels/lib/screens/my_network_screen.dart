import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/friend_invite.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/circle_people_dialog.dart';
import 'package:travel_angels/widgets/circle_person_row.dart';

enum _InviteAction { accept, decline, cancel }

class MyNetworkScreen extends StatefulWidget {
  const MyNetworkScreen({super.key, this.service});
  final FriendService? service;

  @override
  State<MyNetworkScreen> createState() => _MyNetworkScreenState();
}

class _MyNetworkScreenState extends State<MyNetworkScreen> {
  static const _pageSize = 10;
  late final _service =
      widget.service ?? FriendService(APIClient(ApiConstants.baseUrl));
  final _search = TextEditingController();
  final _searchFocus = FocusNode();
  final _busyInvites = <String>{};
  AuthProvider? _auth;
  String? _identity;
  List<User>? _friends;
  List<FriendInvite>? _invites;
  String? _friendsError;
  String? _invitesError;
  bool _loadingFriends = false;
  bool _loadingInvites = false;
  bool _searchExpanded = false;
  String _tab = 'friends';
  int _page = 0;
  int _accountVersion = 0;
  int _friendsVersion = 0;
  int _invitesVersion = 0;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider>();
    if (_auth != auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      auth.addListener(_authChanged);
      _authChanged();
    }
  }

  void _authChanged() {
    final identity = _auth?.user?.id;
    if (identity == _identity) return;
    _identity = identity;
    final account = ++_accountVersion;
    _friendsVersion++;
    _invitesVersion++;
    _search.clear();
    _searchFocus.unfocus();
    setState(() {
      _friends = null;
      _invites = null;
      _friendsError = null;
      _invitesError = null;
      _busyInvites.clear();
      _tab = 'friends';
      _page = 0;
      _searchExpanded = false;
      _loadingFriends = identity?.isNotEmpty == true;
      _loadingInvites = _loadingFriends;
    });
    if (identity?.isNotEmpty == true) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (_currentAccount(account)) _loadData();
      });
    }
  }

  bool _currentAccount(int version) =>
      mounted && version == _accountVersion && _identity?.isNotEmpty == true;

  @override
  void dispose() {
    _accountVersion++;
    _auth?.removeListener(_authChanged);
    _search.dispose();
    _searchFocus.dispose();
    super.dispose();
  }

  Future<void> _loadData() async =>
      Future.wait([_loadFriends(), _loadInvites()]);

  Future<void> _loadFriends() async {
    final account = _accountVersion;
    if (!_currentAccount(account)) return;
    final version = ++_friendsVersion;
    setState(() {
      _loadingFriends = true;
      _friendsError = null;
    });
    try {
      final friends = await _service.getFriends();
      if (!_currentAccount(account) || version != _friendsVersion) return;
      setState(() => _friends = friends);
    } catch (_) {
      if (!_currentAccount(account) || version != _friendsVersion) return;
      setState(() => _friendsError = 'Your friends could not be loaded.');
    } finally {
      if (_currentAccount(account) && version == _friendsVersion) {
        setState(() => _loadingFriends = false);
      }
    }
  }

  Future<void> _loadInvites() async {
    final account = _accountVersion;
    if (!_currentAccount(account)) return;
    final version = ++_invitesVersion;
    setState(() {
      _loadingInvites = true;
      _invitesError = null;
    });
    try {
      final invites = await _service.getFriendInvites();
      if (!_currentAccount(account) || version != _invitesVersion) return;
      setState(() => _invites = invites);
    } catch (_) {
      if (!_currentAccount(account) || version != _invitesVersion) return;
      setState(() => _invitesError = 'Your invitations could not be loaded.');
    } finally {
      if (_currentAccount(account) && version == _invitesVersion) {
        setState(() => _loadingInvites = false);
      }
    }
  }

  Future<void> _findPeople() async {
    final account = _accountVersion;
    final result = await showDialog<String>(
      context: context,
      builder: (_) => CirclePeopleDialog(
        service: _service,
        friends: _friends ?? [],
        invites: _invites ?? [],
      ),
    );
    if (!_currentAccount(account)) return;
    if (result == 'requests') {
      setState(() {
        _tab = 'requests';
        _search.clear();
      });
    }
    await _loadData();
  }

  Future<void> _act(FriendInvite invite, _InviteAction action) async {
    final account = _accountVersion;
    if (!_currentAccount(account) ||
        invite.id.isEmpty ||
        _busyInvites.contains(invite.id)) {
      return;
    }
    final ownsAction = action == _InviteAction.cancel
        ? invite.inviterId == _identity
        : invite.inviteeId == _identity;
    if (!ownsAction) return;
    setState(() => _busyInvites.add(invite.id));
    try {
      if (action == _InviteAction.accept) {
        await _service.acceptFriendInvite(invite.id);
      } else {
        await _service.deleteFriendInvite(invite.id);
      }
      if (!_currentAccount(account)) return;
      setState(
        () => _invites = _invites
            ?.where((entry) => entry.id != invite.id)
            .toList(),
      );
      if (action == _InviteAction.accept) {
        await _loadData();
      } else {
        await _loadInvites();
      }
      if (!mounted || !_currentAccount(account)) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(switch (action) {
            _InviteAction.accept => 'Request accepted',
            _InviteAction.decline => 'Request declined',
            _InviteAction.cancel => 'Invitation canceled',
          }),
        ),
      );
    } catch (_) {
      if (!_currentAccount(account)) return;
      // Acceptance can partially succeed on the server. Reconcile both lists
      // before offering another attempt instead of guessing the resulting state.
      if (action == _InviteAction.accept) await _loadData();
      if (!mounted || !_currentAccount(account)) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(switch (action) {
            _InviteAction.accept =>
              'The request could not be accepted. Please try again.',
            _InviteAction.decline =>
              'The request could not be declined. Please try again.',
            _InviteAction.cancel =>
              'The invitation could not be canceled. Please try again.',
          }),
        ),
      );
    } finally {
      if (_currentAccount(account)) {
        setState(() => _busyInvites.remove(invite.id));
      }
    }
  }

  bool _matches(User? user) {
    final query = _search.text.trim().toLowerCase();
    return query.isEmpty ||
        circlePersonName(user).toLowerCase().contains(query) ||
        (user?.username?.toLowerCase().contains(query) ?? false);
  }

  List<User> get _filteredFriends =>
      (_friends ?? []).where(_matches).toList()..sort(
        (a, b) => circlePersonName(
          a,
        ).toLowerCase().compareTo(circlePersonName(b).toLowerCase()),
      );

  List<FriendInvite> get _received => (_invites ?? [])
      .where((invite) => invite.inviteeId == _identity)
      .toList();
  List<FriendInvite> get _sent => (_invites ?? [])
      .where((invite) => invite.inviterId == _identity)
      .toList();

  void _toggleSearch() {
    setState(() => _searchExpanded = !_searchExpanded);
    if (_searchExpanded) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _searchExpanded) _searchFocus.requestFocus();
      });
    } else {
      _searchFocus.unfocus();
    }
  }

  Widget _tabs() => SingleChildScrollView(
    scrollDirection: Axis.horizontal,
    child: Row(
      children: [
        for (final tab in [
          ('friends', 'Friends'),
          ('requests', 'Requests'),
          ('sent', 'Sent'),
        ])
          Container(
            decoration: BoxDecoration(
              border: Border(
                bottom: BorderSide(
                  width: 2,
                  color: _tab == tab.$1 ? AppTheme.rust : Colors.transparent,
                ),
              ),
            ),
            child: Semantics(
              selected: _tab == tab.$1,
              button: true,
              child: TextButton(
                key: ValueKey('circle-tab-${tab.$1}'),
                onPressed: () => setState(() => _tab = tab.$1),
                style: TextButton.styleFrom(
                  minimumSize: const Size(44, 44),
                  padding: const EdgeInsets.symmetric(horizontal: 10),
                  foregroundColor: _tab == tab.$1
                      ? AppTheme.rust
                      : AppTheme.ink,
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(tab.$2),
                    if (tab.$1 == 'requests' && _received.isNotEmpty) ...[
                      const SizedBox(width: 6),
                      Badge(
                        backgroundColor: AppTheme.peach,
                        textColor: AppTheme.rust,
                        label: Text('${_received.length}'),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
      ],
    ),
  );

  Widget _searchField() => TextField(
    key: const ValueKey('circle-search'),
    controller: _search,
    focusNode: _searchFocus,
    textInputAction: TextInputAction.search,
    style: Theme.of(context).textTheme.bodyMedium,
    decoration: InputDecoration(
      isDense: true,
      hintText: 'Search by name or username',
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      prefixIcon: const Icon(Icons.search, size: 18),
      prefixIconConstraints: const BoxConstraints(minWidth: 40, minHeight: 44),
      suffixIcon: _search.text.isEmpty
          ? null
          : IconButton(
              tooltip: 'Clear search',
              icon: const Icon(Icons.clear, size: 18),
              onPressed: () => setState(() {
                _search.clear();
                _page = 0;
              }),
            ),
    ),
    onChanged: (_) => setState(() => _page = 0),
    onSubmitted: (_) => _searchFocus.unfocus(),
  );

  Widget _retry(String text, String key, VoidCallback retry) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: Wrap(
      spacing: 8,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        Text(text, style: Theme.of(context).textTheme.bodySmall),
        TextButton(
          key: ValueKey(key),
          onPressed: retry,
          child: const Text('Retry'),
        ),
      ],
    ),
  );

  Widget _empty(String text, {bool canFindPeople = false}) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 32),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          text,
          style: Theme.of(
            context,
          ).textTheme.bodyMedium?.copyWith(color: AppTheme.muted),
        ),
        if (canFindPeople) ...[
          const SizedBox(height: 8),
          TextButton.icon(
            onPressed: _findPeople,
            icon: const Icon(Icons.person_add_alt, size: 18),
            label: const Text('Find people'),
          ),
        ],
      ],
    ),
  );

  Widget _peopleLayout(List<Widget> people, double width) {
    final columns =
        width >= 760 && MediaQuery.textScalerOf(context).scale(14) <= 21
        ? 2
        : 1;
    return Wrap(
      spacing: 32,
      children: [
        for (final person in people)
          SizedBox(
            width: (width - (columns - 1) * 32) / columns,
            child: DecoratedBox(
              decoration: const BoxDecoration(
                border: Border(bottom: BorderSide(color: AppTheme.outline)),
              ),
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: person,
              ),
            ),
          ),
      ],
    );
  }

  Widget _friendsList(double width) {
    final friends = _filteredFriends;
    final pages = friends.isEmpty ? 1 : (friends.length / _pageSize).ceil();
    final page = _page.clamp(0, pages - 1);
    final start = page * _pageSize;
    final end = (start + _pageSize).clamp(0, friends.length);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_loadingFriends) const LinearProgressIndicator(minHeight: 2),
        if (_friendsError != null)
          _retry(_friendsError!, 'circle-friends-retry', _loadFriends),
        if (_friends != null)
          if (friends.isEmpty)
            _empty(
              _search.text.trim().isEmpty
                  ? 'Your circle starts with a friend.'
                  : 'No friends match your search.',
              canFindPeople: _search.text.trim().isEmpty,
            )
          else ...[
            _peopleLayout([
              for (final friend in friends.sublist(start, end))
                CirclePersonRow(
                  key: ValueKey('circle-friend-${friend.id}'),
                  user: friend,
                ),
            ], width),
            if (pages > 1)
              Padding(
                padding: const EdgeInsets.only(top: 12),
                child: Wrap(
                  alignment: WrapAlignment.spaceBetween,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  spacing: 8,
                  children: [
                    Text(
                      '${start + 1}–$end of ${friends.length} friends',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          key: const ValueKey('circle-page-previous'),
                          tooltip: 'Previous page',
                          onPressed: page == 0
                              ? null
                              : () => setState(() => _page = page - 1),
                          icon: const Icon(Icons.chevron_left, size: 20),
                        ),
                        Text(
                          '${page + 1} / $pages',
                          style: Theme.of(context).textTheme.bodySmall,
                        ),
                        IconButton(
                          key: const ValueKey('circle-page-next'),
                          tooltip: 'Next page',
                          onPressed: page + 1 >= pages
                              ? null
                              : () => setState(() => _page = page + 1),
                          icon: const Icon(Icons.chevron_right, size: 20),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
          ],
      ],
    );
  }

  Widget _invitesList(double width) {
    final incoming = _tab == 'requests';
    final invites = (incoming ? _received : _sent)
        .where((invite) => _matches(incoming ? invite.inviter : invite.invitee))
        .toList();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_loadingInvites) const LinearProgressIndicator(minHeight: 2),
        if (_invitesError != null)
          _retry(_invitesError!, 'circle-invites-retry', _loadInvites),
        if (_invites != null)
          if (invites.isEmpty)
            _empty(
              _search.text.trim().isNotEmpty
                  ? 'No invitations match your search.'
                  : incoming
                  ? 'No pending requests.'
                  : 'No invitations sent.',
            )
          else
            _peopleLayout([
              for (final invite in invites)
                CirclePersonRow(
                  key: ValueKey('circle-request-${invite.id}'),
                  user: incoming ? invite.inviter : invite.invitee,
                  detail: invite.createdAt == null
                      ? null
                      : '${incoming ? 'Received' : 'Sent'} ${DateFormat.MMMd().format(invite.createdAt!.toLocal())}',
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (_busyInvites.contains(invite.id))
                        const Padding(
                          padding: EdgeInsets.symmetric(horizontal: 12),
                          child: SizedBox.square(
                            dimension: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        ),
                      if (incoming) ...[
                        IconButton(
                          key: ValueKey('circle-accept-${invite.id}'),
                          tooltip:
                              'Accept request from ${circlePersonName(invite.inviter)}',
                          onPressed:
                              _busyInvites.contains(invite.id) ||
                                  _loadingInvites
                              ? null
                              : () => _act(invite, _InviteAction.accept),
                          color: AppTheme.rust,
                          icon: const Icon(Icons.check, size: 20),
                        ),
                        IconButton(
                          key: ValueKey('circle-decline-${invite.id}'),
                          tooltip:
                              'Decline request from ${circlePersonName(invite.inviter)}',
                          onPressed:
                              _busyInvites.contains(invite.id) ||
                                  _loadingInvites
                              ? null
                              : () => _act(invite, _InviteAction.decline),
                          icon: const Icon(Icons.close, size: 20),
                        ),
                      ] else
                        IconButton(
                          key: ValueKey('circle-cancel-${invite.id}'),
                          tooltip:
                              'Cancel invitation to ${circlePersonName(invite.invitee)}',
                          onPressed:
                              _busyInvites.contains(invite.id) ||
                                  _loadingInvites
                              ? null
                              : () => _act(invite, _InviteAction.cancel),
                          icon: const Icon(Icons.close, size: 20),
                        ),
                    ],
                  ),
                ),
            ], width),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    context.watch<AuthProvider>();
    if (_identity?.isNotEmpty != true) {
      return const Center(child: Text('Sign in to view your circle.'));
    }
    final compact =
        MediaQuery.sizeOf(context).width < 600 ||
        MediaQuery.textScalerOf(context).scale(14) > 21;
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 1040),
              child: Padding(
                padding: EdgeInsets.all(compact ? 16 : 24),
                child: LayoutBuilder(
                  builder: (context, constraints) => Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              'My circle',
                              style: Theme.of(context).textTheme.headlineMedium,
                            ),
                          ),
                          if (compact)
                            IconButton(
                              key: const ValueKey('circle-find-people'),
                              tooltip: 'Find people',
                              onPressed: _findPeople,
                              color: AppTheme.rust,
                              icon: const Icon(Icons.person_add_alt, size: 20),
                            )
                          else
                            TextButton.icon(
                              key: const ValueKey('circle-find-people'),
                              onPressed: _findPeople,
                              style: TextButton.styleFrom(
                                foregroundColor: AppTheme.rust,
                              ),
                              icon: const Icon(Icons.person_add_alt, size: 18),
                              label: const Text('Find people'),
                            ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Container(
                        decoration: const BoxDecoration(
                          border: Border(
                            bottom: BorderSide(color: AppTheme.outline),
                          ),
                        ),
                        child: Row(
                          children: [
                            Expanded(child: _tabs()),
                            const SizedBox(width: 4),
                            IconButton(
                              key: const ValueKey('circle-search-toggle'),
                              tooltip: _searchExpanded
                                  ? 'Hide search'
                                  : 'Search your circle',
                              color:
                                  _searchExpanded ||
                                      _search.text.trim().isNotEmpty
                                  ? AppTheme.rust
                                  : AppTheme.ink,
                              onPressed: _toggleSearch,
                              icon: Icon(
                                _searchExpanded ? Icons.close : Icons.search,
                                size: 20,
                              ),
                            ),
                          ],
                        ),
                      ),
                      if (_searchExpanded) ...[
                        const SizedBox(height: 12),
                        _searchField(),
                      ],
                      if (!_searchExpanded && _search.text.trim().isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 8),
                          child: Align(
                            alignment: Alignment.centerLeft,
                            child: InputChip(
                              label: Text(
                                'Search: ${_search.text.trim()}',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                              onPressed: _toggleSearch,
                              deleteButtonTooltipMessage: 'Clear search',
                              onDeleted: () => setState(() {
                                _search.clear();
                                _page = 0;
                              }),
                            ),
                          ),
                        ),
                      const SizedBox(height: 8),
                      if (_tab == 'friends')
                        _friendsList(constraints.maxWidth)
                      else
                        _invitesList(constraints.maxWidth),
                      const SizedBox(height: 24),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
