import 'package:flutter/foundation.dart';

/// Minimal provider for CelebrationModalManager.
/// Modal queue is populated when tasks complete (not wired in this polish pass).
class CompletedTaskProvider with ChangeNotifier {
  Map<String, dynamic>? _currentModal;
  final List<Map<String, dynamic>> _queuedModals = [];

  Map<String, dynamic>? get currentModal => _currentModal;

  void showModal(String type, {Map<String, dynamic>? data}) {
    final next = {'type': type, 'data': data ?? {}};
    if (_currentModal == null) {
      _currentModal = next;
    } else {
      _queuedModals.add(next);
    }
    notifyListeners();
  }

  void queueLevelUpFollowUpIfNeeded({
    required int previousLevel,
    required int currentLevel,
  }) {
    if (currentLevel <= previousLevel) return;
    queueLevelUpModal(
      previousLevel: previousLevel,
      newLevel: currentLevel,
      levelsGained: currentLevel - previousLevel,
    );
  }

  void queueLevelUpModal({
    required int newLevel,
    int? previousLevel,
    int levelsGained = 1,
  }) {
    showModal(
      'levelUp',
      data: {
        if (previousLevel != null) 'previousLevel': previousLevel,
        'newLevel': newLevel,
        'levelsGained': levelsGained,
      },
    );
  }

  void clearModal() {
    if (_queuedModals.isNotEmpty) {
      _currentModal = _queuedModals.removeAt(0);
    } else {
      _currentModal = null;
    }
    notifyListeners();
  }
}
