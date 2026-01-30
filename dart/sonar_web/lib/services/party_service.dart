import '../constants/api_constants.dart';
import '../models/party.dart';
import '../models/party_invite.dart';
import '../models/user.dart';
import 'api_client.dart';

class PartyService {
  final ApiClient _api;

  PartyService(this._api);

  /// Returns party when user is in one. 404 when not in a party.
  Future<Party?> getParty() async {
    try {
      final data = await _api.get<Map<String, dynamic>>(ApiConstants.partyEndpoint);
      return Party.fromJson(data);
    } on Exception catch (_) {
      return null;
    }
  }

  Future<List<PartyInvite>> getPartyInvites() async {
    try {
      final list = await _api.get<List<dynamic>>(ApiConstants.partyInvitesEndpoint);
      return list
          .map((e) => PartyInvite.fromJson(e as Map<String, dynamic>))
          .toList();
    } on Exception catch (_) {
      return [];
    }
  }

  Future<void> leaveParty() async {
    await _api.post<dynamic>(ApiConstants.partyLeaveEndpoint);
  }

  /// Backend uses leaderID; path is /sonar/party/setLeader.
  Future<void> setLeader(User leader) async {
    await _api.post<dynamic>(
      ApiConstants.partySetLeaderEndpoint,
      data: {'leaderID': leader.id},
    );
  }

  Future<void> inviteToParty(User invitee) async {
    await _api.post<dynamic>(
      ApiConstants.partyInvitesEndpoint,
      data: {'inviteeID': invitee.id},
    );
  }

  Future<void> acceptPartyInvite(String inviteId) async {
    await _api.post<dynamic>(
      ApiConstants.partyInvitesAcceptEndpoint,
      data: {'inviteID': inviteId},
    );
  }

  Future<void> rejectPartyInvite(String inviteId) async {
    await _api.post<dynamic>(
      ApiConstants.partyInvitesRejectEndpoint,
      data: {'inviteID': inviteId},
    );
  }
}
