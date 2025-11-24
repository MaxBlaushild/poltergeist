import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/friend.dart';
import 'package:travel_angels/models/friend_invite.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/api_client.dart';

class FriendService {
  final APIClient _apiClient;

  FriendService(this._apiClient);

  /// Gets the user's friends list
  /// 
  /// Returns a list of User objects representing friends
  Future<List<User>> getFriends() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.friendsEndpoint,
      );
      return response
          .map((json) => User.fromJson(Map<String, dynamic>.from(json)))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Gets all friend invites (sent and received)
  /// 
  /// Returns a list of FriendInvite objects
  Future<List<FriendInvite>> getFriendInvites() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.friendInvitesEndpoint,
      );
      return response
          .map((json) => FriendInvite.fromJson(Map<String, dynamic>.from(json)))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Searches for users by username
  /// 
  /// [query] - The search query (username)
  /// 
  /// Returns a list of User objects matching the query
  Future<List<User>> searchUsers(String query) async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.searchUsersEndpoint(query),
      );
      return response
          .map((json) => User.fromJson(Map<String, dynamic>.from(json)))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Creates a friend invite
  /// 
  /// [inviteeId] - The ID of the user to invite
  /// 
  /// Returns the created FriendInvite
  Future<FriendInvite> createFriendInvite(String inviteeId) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.createFriendInviteEndpoint,
        data: {
          'inviteeID': inviteeId,
        },
      );
      return FriendInvite.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }

  /// Accepts a friend invite
  /// 
  /// [inviteId] - The ID of the invite to accept
  /// 
  /// Throws an exception if acceptance fails
  Future<void> acceptFriendInvite(String inviteId) async {
    try {
      await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.acceptFriendInviteEndpoint,
        data: {
          'inviteID': inviteId,
        },
      );
    } catch (e) {
      rethrow;
    }
  }

  /// Deletes/revokes a friend invite
  /// 
  /// [inviteId] - The ID of the invite to delete
  /// 
  /// Throws an exception if deletion fails
  Future<void> deleteFriendInvite(String inviteId) async {
    try {
      await _apiClient.delete<void>(
        ApiConstants.deleteFriendInviteEndpoint(inviteId),
      );
    } catch (e) {
      rethrow;
    }
  }
}

