import 'package:flutter/foundation.dart';

class QuestFilterProvider with ChangeNotifier {
  bool _enableTagFilter = false;

  bool get enableTagFilter => _enableTagFilter;

  void toggleTagFilter() {
    _enableTagFilter = !_enableTagFilter;
    notifyListeners();
  }

  void clearAll() {
    _enableTagFilter = false;
    notifyListeners();
  }
}
