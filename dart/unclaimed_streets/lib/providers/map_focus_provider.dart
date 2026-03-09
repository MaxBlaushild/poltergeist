import 'package:flutter/foundation.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';

class MapFocusProvider with ChangeNotifier {
  PointOfInterest? _pendingPoi;
  Quest? _pendingTurnInQuest;

  void focusPoi(PointOfInterest poi) {
    _pendingPoi = poi;
    _pendingTurnInQuest = null;
    notifyListeners();
  }

  void focusTurnInQuest(Quest quest) {
    _pendingTurnInQuest = quest;
    _pendingPoi = null;
    notifyListeners();
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
}
