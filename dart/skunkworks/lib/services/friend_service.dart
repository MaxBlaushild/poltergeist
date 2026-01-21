import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/friend.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/services/api_client.dart';

class FriendService {
  final APIClient _apiClient;

  FriendService(this._apiClient);

  /// Gets the user's friends list
  /// 
  /// Returns list of friend users
  Future<List<User>> getFriends() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.friendsEndpoint,
      );

      return response
          .map((json) => User.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Gets pending friend invites
  /// 
  /// Returns list of friend invites
  Future<List<FriendInvite>> getFriendInvites() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.friendInvitesEndpoint,
      );

      return response
          .map((json) => FriendInvite.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Creates a friend invite
  /// 
  /// [inviteeID] - The ID of the user to invite
  /// 
  /// Returns the created friend invite
  Future<FriendInvite> createFriendInvite(String inviteeID) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.createFriendInviteEndpoint,
        data: {
          'inviteeID': inviteeID,
        },
      );

      return FriendInvite.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }

  /// Accepts a friend invite
  /// 
  /// [inviteID] - The ID of the invite to accept
  /// 
  /// Returns true if successful
  Future<bool> acceptFriendInvite(String inviteID) async {
    try {
      await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.acceptFriendInviteEndpoint,
        data: {
          'inviteID': inviteID,
        },
      );
      return true;
    } catch (e) {
      rethrow;
    }
  }

  /// Deletes a friend invite
  /// 
  /// [inviteID] - The ID of the invite to delete
  /// 
  /// Returns true if successful
  Future<bool> deleteFriendInvite(String inviteID) async {
    try {
      await _apiClient.delete(
        ApiConstants.deleteFriendInviteEndpoint(inviteID),
      );
      return true;
    } catch (e) {
      rethrow;
    }
  }

  /// Searches for users by username
  /// 
  /// [query] - The search query
  /// 
  /// Returns list of matching users
  Future<List<User>> searchUsers(String query) async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.searchUsersEndpoint(query),
      );

      return response
          .map((json) => User.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }
}

