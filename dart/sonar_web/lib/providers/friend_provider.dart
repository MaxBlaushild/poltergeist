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
  bool _loading = false;

  List<User> get friends => _friends;
  List<FriendInvite> get friendInvites => _friendInvites;
  List<User> get searchResults => _searchResults;
  bool get loading => _loading;

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
    notifyListeners();
  }

  Future<void> refresh() async {
    await Future.wait([fetchFriends(), fetchFriendInvites()]);
  }

  Future<void> searchForFriends(String query) async {
    if (query.trim().isEmpty) {
      _searchResults = [];
      notifyListeners();
      return;
    }
    try {
      _searchResults = await _friendService.searchUsers(query);
    } catch (_) {
      _searchResults = [];
    }
    notifyListeners();
  }

  Future<void> createFriendInvite(String inviteeId) async {
    await _friendService.createFriendInvite(inviteeId);
    await fetchFriendInvites();
    notifyListeners();
  }

  Future<void> acceptFriendInvite(String inviteId) async {
    await _friendService.acceptFriendInvite(inviteId);
    _friendInvites = _friendInvites.where((i) => i.id != inviteId).toList();
    await fetchFriends();
    notifyListeners();
  }

  Future<void> deleteFriendInvite(String inviteId) async {
    await _friendService.deleteFriendInvite(inviteId);
    _friendInvites = _friendInvites.where((i) => i.id != inviteId).toList();
    notifyListeners();
  }

  void clearSearch() {
    _searchResults = [];
    notifyListeners();
  }

  void clear() {
    _friends = [];
    _friendInvites = [];
    _searchResults = [];
    notifyListeners();
  }
}
