import 'package:flutter/foundation.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';

class MapFocusProvider with ChangeNotifier {
  PointOfInterest? _pendingPoi;
  PointOfInterest? _pendingMainStoryLeadPoi;
  Quest? _pendingTurnInQuest;
  QuestNode? _pendingNode;

  void focusPoi(PointOfInterest poi) {
    _pendingPoi = poi;
    _pendingMainStoryLeadPoi = null;
    _pendingTurnInQuest = null;
    _pendingNode = null;
    notifyListeners();
  }

  void focusMainStoryLead(PointOfInterest poi) {
    _pendingMainStoryLeadPoi = poi;
    _pendingPoi = null;
    _pendingTurnInQuest = null;
    _pendingNode = null;
    notifyListeners();
  }

  void focusTurnInQuest(Quest quest) {
    _pendingTurnInQuest = quest;
    _pendingMainStoryLeadPoi = null;
    _pendingPoi = null;
    _pendingNode = null;
    notifyListeners();
  }

  void focusNode(QuestNode node) {
    _pendingNode = node;
    _pendingTurnInQuest = null;
    _pendingMainStoryLeadPoi = null;
    _pendingPoi = null;
    notifyListeners();
  }

  PointOfInterest? consumeMainStoryLeadPoi() {
    final poi = _pendingMainStoryLeadPoi;
    _pendingMainStoryLeadPoi = null;
    return poi;
  }

  PointOfInterest? consumePoi() {
    final poi = _pendingPoi;
    _pendingPoi = null;
    return poi;
  }

  Quest? consumeTurnInQuest() {
    final quest = _pendingTurnInQuest;
    _pendingTurnInQuest = null;
    return quest;
  }

  QuestNode? consumeNode() {
    final node = _pendingNode;
    _pendingNode = null;
    return node;
  }
}
