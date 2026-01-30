import 'package:flutter/foundation.dart';

import '../models/party.dart';
import '../models/party_invite.dart';
import '../models/user.dart';
import '../services/party_service.dart';

class PartyProvider with ChangeNotifier {
  final PartyService _partyService;

  PartyProvider(this._partyService);

  Party? _party;
  List<PartyInvite> _partyInvites = [];
  bool _loading = true;
  Object? _error;

  Party? get party => _party;
  List<PartyInvite> get partyInvites => _partyInvites;
  bool get loading => _loading;
  Object? get error => _error;

  Future<void> fetchParty() async {
    _loading = true;
    _error = null;
    notifyListeners();
    try {
      _party = await _partyService.getParty();
    } catch (e) {
      _party = null;
      _error = e;
    }
    _loading = false;
    notifyListeners();
  }

  Future<void> fetchPartyInvites() async {
    try {
      _partyInvites = await _partyService.getPartyInvites();
    } catch (_) {
      _partyInvites = [];
    }
    notifyListeners();
  }

  Future<void> refresh() async {
    await Future.wait([fetchParty(), fetchPartyInvites()]);
  }

  Future<void> leaveParty() async {
    await _partyService.leaveParty();
    _party = null;
    await fetchPartyInvites();
    notifyListeners();
  }

  Future<void> setLeader(User leader) async {
    await _partyService.setLeader(leader);
    await fetchParty();
    notifyListeners();
  }

  Future<void> inviteToParty(User invitee) async {
    await _partyService.inviteToParty(invitee);
    await fetchPartyInvites();
    notifyListeners();
  }

  Future<void> acceptPartyInvite(String inviteId) async {
    await _partyService.acceptPartyInvite(inviteId);
    _partyInvites = _partyInvites.where((i) => i.id != inviteId).toList();
    await fetchParty();
    notifyListeners();
  }

  Future<void> rejectPartyInvite(String inviteId) async {
    await _partyService.rejectPartyInvite(inviteId);
    _partyInvites = _partyInvites.where((i) => i.id != inviteId).toList();
    notifyListeners();
  }

  void clear() {
    _party = null;
    _partyInvites = [];
    _error = null;
    notifyListeners();
  }
}
