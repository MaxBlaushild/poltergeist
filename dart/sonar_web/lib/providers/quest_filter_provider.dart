import 'package:flutter/foundation.dart';

class QuestFilterProvider with ChangeNotifier {
  bool _showCurrentQuestPoints = false;
  bool _showQuestAvailablePoints = false;
  bool _enableTagFilter = false;
  bool _showTreasureChests = true;

  bool get showCurrentQuestPoints => _showCurrentQuestPoints;
  bool get showQuestAvailablePoints => _showQuestAvailablePoints;
  bool get enableTagFilter => _enableTagFilter;
  bool get showTreasureChests => _showTreasureChests;

  bool get hasAnyCategoryFilter =>
      _showCurrentQuestPoints || _showQuestAvailablePoints || !_showTreasureChests;

  void toggleCurrentQuestPoints() {
    _showCurrentQuestPoints = !_showCurrentQuestPoints;
    notifyListeners();
  }

  void toggleQuestAvailablePoints() {
    _showQuestAvailablePoints = !_showQuestAvailablePoints;
    notifyListeners();
  }

  void toggleTagFilter() {
    _enableTagFilter = !_enableTagFilter;
    notifyListeners();
  }

  void toggleTreasureChests() {
    _showTreasureChests = !_showTreasureChests;
    notifyListeners();
  }

  void clearAll() {
    _showCurrentQuestPoints = false;
    _showQuestAvailablePoints = false;
    _enableTagFilter = false;
    _showTreasureChests = true;
    notifyListeners();
  }
}
