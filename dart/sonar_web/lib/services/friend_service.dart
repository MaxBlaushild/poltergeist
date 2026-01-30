import '../constants/api_constants.dart';
import '../models/friend_invite.dart';
import '../models/user.dart';
import 'api_client.dart';

class FriendService {
  final ApiClient _api;

  FriendService(this._api);

  Future<List<User>> searchUsers(String query) async {
    if (query.trim().isEmpty) return [];
    try {
      final list = await _api.get<List<dynamic>>(
        ApiConstants.usersSearchEndpoint,
        params: {'query': query},
      );
      return list
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .where((u) =>
              u.id.isNotEmpty && u.id != '00000000-0000-0000-0000-000000000000')
          .toList();
    } on Exception catch (_) {
      return [];
    }
  }

  Future<List<User>> getFriends() async {
    try {
      final list = await _api.get<List<dynamic>>(ApiConstants.friendsEndpoint);
      return list
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .toList();
    } on Exception catch (_) {
      return [];
    }
  }

  Future<List<FriendInvite>> getFriendInvites() async {
    try {
      final list =
          await _api.get<List<dynamic>>(ApiConstants.friendInvitesEndpoint);
      return list
          .map((e) => FriendInvite.fromJson(e as Map<String, dynamic>))
          .toList();
    } on Exception catch (_) {
      return [];
    }
  }

  Future<void> createFriendInvite(String inviteeId) async {
    await _api.post<dynamic>(
      ApiConstants.friendInvitesCreateEndpoint,
      data: {'inviteeID': inviteeId},
    );
  }

  /// Backend expects inviteId (lowercase d).
  Future<void> acceptFriendInvite(String inviteId) async {
    await _api.post<dynamic>(
      ApiConstants.friendInvitesAcceptEndpoint,
      data: {'inviteId': inviteId},
    );
  }

  Future<void> deleteFriendInvite(String inviteId) async {
    await _api.delete<dynamic>('${ApiConstants.friendInvitesEndpoint}/$inviteId');
  }
}
