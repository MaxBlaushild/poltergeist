import 'package:flutter/foundation.dart';

/// Minimal provider for CelebrationModalManager.
/// Modal queue is populated when tasks complete (not wired in this polish pass).
class CompletedTaskProvider with ChangeNotifier {
  Map<String, dynamic>? _currentModal;

  Map<String, dynamic>? get currentModal => _currentModal;

  void showModal(String type, {Map<String, dynamic>? data}) {
    _currentModal = {'type': type, 'data': data ?? {}};
    notifyListeners();
  }

  void clearModal() {
    _currentModal = null;
    notifyListeners();
  }
}
