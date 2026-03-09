import 'package:flutter/foundation.dart';

import '../models/friend_invite.dart';
import '../models/user.dart';
import '../services/friend_service.dart';

class FriendProvider with ChangeNotifier {
  final FriendService _friendService;

  FriendProvider(this._friendService);

  List<User> _friends = [];
  List<FriendInvite> _friendInvites = [];
  List<User> _searchResults = [];
  final bool _loading = false;
  bool _hasLoadedFriendInvites = false;
  String _searchQuery = '';

  List<User> get friends => _friends;
  List<FriendInvite> get friendInvites => _friendInvites;
  List<User> get searchResults => _searchResults;
  bool get loading => _loading;
  bool get hasLoadedFriendInvites => _hasLoadedFriendInvites;

  Future<void> fetchFriends() async {
    try {
      _friends = await _friendService.getFriends();
    } catch (_) {
      _friends = [];
    }
    notifyListeners();
  }

  Future<void> fetchFriendInvites() async {
    try {
      _friendInvites = await _friendService.getFriendInvites();
    } catch (_) {
      _friendInvites = [];
    }
    _hasLoadedFriendInvites = true;
    notifyListeners();
  }

  Future<void> refresh() async {
    await Future.wait([fetchFriends(), fetchFriendInvites()]);
  }

  Future<void> searchForFriends(String query) async {
    _searchQuery = query.trim();
    if (_searchQuery.isEmpty) {
      _searchResults = [];
      notifyListeners();
      return;
    }
    await _refreshSearchResults();
  }

  Future<void> _refreshSearchResults() async {
    try {
      _searchResults = await _friendService.searchUsers(_searchQuery);
    } catch (_) {
      _searchResults = [];
    }
    notifyListeners();
  }

  Future<void> createFriendInvite(String inviteeId) async {
    await _friendService.createFriendInvite(inviteeId);
    await fetchFriendInvites();
    await _refreshSearchResults();
  }

  Future<void> acceptFriendInvite(String inviteId) async {
    await _friendService.acceptFriendInvite(inviteId);
    _friendInvites = _friendInvites.where((i) => i.id != inviteId).toList();
    await fetchFriends();
    await fetchFriendInvites();
    await _refreshSearchResults();
  }

  Future<void> deleteFriendInvite(String inviteId) async {
    await _friendService.deleteFriendInvite(inviteId);
    _friendInvites = _friendInvites.where((i) => i.id != inviteId).toList();
    notifyListeners();
    await fetchFriendInvites();
    await _refreshSearchResults();
  }

  void clearSearch() {
    _searchQuery = '';
    _searchResults = [];
    notifyListeners();
  }

  void clear() {
    _friends = [];
    _friendInvites = [];
    _hasLoadedFriendInvites = false;
    _searchQuery = '';
    _searchResults = [];
    notifyListeners();
  }
}
