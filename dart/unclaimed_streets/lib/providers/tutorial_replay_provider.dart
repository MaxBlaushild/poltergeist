import 'package:flutter/foundation.dart';

class TutorialReplayProvider with ChangeNotifier {
  int _requestCount = 0;

  int get requestCount => _requestCount;

  void requestReplay() {
    _requestCount += 1;
    notifyListeners();
  }
}
