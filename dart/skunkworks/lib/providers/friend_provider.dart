import 'package:flutter/material.dart';
import 'package:skunkworks/models/friend.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/services/friend_service.dart';

class FriendProvider extends ChangeNotifier {
  final FriendService _friendService;
  List<User> _friends = [];
  List<FriendInvite> _friendInvites = [];
  List<User> _searchResults = [];
  bool _loading = false;
  String? _error;

  FriendProvider(this._friendService);

  List<User> get friends => _friends;
  List<FriendInvite> get friendInvites => _friendInvites;
  List<User> get searchResults => _searchResults;
  bool get loading => _loading;
  String? get error => _error;

  /// Loads the user's friends list
  Future<void> loadFriends() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      _friends = await _friendService.getFriends();
    } catch (e) {
      _error = e.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  /// Loads pending friend invites
  Future<void> loadInvites() async {
    _error = null;
    notifyListeners();

    try {
      _friendInvites = await _friendService.getFriendInvites();
    } catch (e) {
      _error = e.toString();
    } finally {
      notifyListeners();
    }
  }

  /// Searches for users by username
  /// 
  /// [query] - The search query
  Future<void> searchUsers(String query) async {
    if (query.trim().isEmpty) {
      _searchResults = [];
      notifyListeners();
      return;
    }

    _error = null;
    notifyListeners();

    try {
      _searchResults = await _friendService.searchUsers(query);
    } catch (e) {
      _error = e.toString();
      _searchResults = [];
    } finally {
      notifyListeners();
    }
  }

  /// Sends a friend invite
  /// 
  /// [inviteeID] - The ID of the user to invite
  Future<void> sendInvite(String inviteeID) async {
    _error = null;
    notifyListeners();

    try {
      await _friendService.createFriendInvite(inviteeID);
      await loadInvites();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Accepts a friend invite
  /// 
  /// [inviteID] - The ID of the invite to accept
  Future<void> acceptInvite(String inviteID) async {
    _error = null;
    notifyListeners();

    try {
      await _friendService.acceptFriendInvite(inviteID);
      await loadInvites();
      await loadFriends();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Deletes a friend invite
  /// 
  /// [inviteID] - The ID of the invite to delete
  Future<void> deleteInvite(String inviteID) async {
    _error = null;
    notifyListeners();

    try {
      await _friendService.deleteFriendInvite(inviteID);
      await loadInvites();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }
}

