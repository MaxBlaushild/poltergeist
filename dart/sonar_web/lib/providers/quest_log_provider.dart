import 'package:flutter/foundation.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../providers/tags_provider.dart';
import '../providers/zone_provider.dart';
import '../services/quest_log_service.dart';

/// POIs that appear in the quest log (root or unlocked via completed objectives).
List<PointOfInterest> getMapPointsOfInterest(List<Quest> quests) {
  final out = <PointOfInterest>[];
  for (final quest in quests) {
    void addFromNode(QuestNode node) {
      out.add(node.pointOfInterest);
      for (final o in node.objectives) {
        if (o.isCompleted && o.nextNode != null) {
          addFromNode(o.nextNode!);
        }
      }
    }
    addFromNode(quest.rootNode);
  }
  return out;
}

/// All POI IDs in a quest's tree (for tracking focus).
List<String> getAllPointsOfInterestIdsForQuest(Quest quest) {
  final ids = <String>[];
  void visit(QuestNode node) {
    ids.add(node.pointOfInterest.id);
    for (final o in node.objectives) {
      if (o.nextNode != null) visit(o.nextNode!);
    }
  }
  visit(quest.rootNode);
  return ids;
}

class QuestLogProvider with ChangeNotifier {
  final QuestLogService _service;
  final ZoneProvider _zone;
  final TagsProvider _tags;

  List<Quest> _quests = [];
  List<String> _trackedQuestIds = [];
  List<String> _trackedPointOfInterestIds = [];
  List<PointOfInterest> _pointsOfInterest = [];
  Map<String, List<QuestChallenge>> _pendingTasks = {};
  Map<String, List<QuestChallenge>> _completedTasks = {};
  bool _loading = false;
  String? _lastZoneId;
  List<String> _lastTagNames = [];

  QuestLogProvider(this._service, this._zone, this._tags) {
    _zone.addListener(_onZoneOrTagsChanged);
    _tags.addListener(_onZoneOrTagsChanged);
  }

  List<Quest> get quests => _quests;
  List<String> get trackedQuestIds => _trackedQuestIds;
  List<String> get trackedPointOfInterestIds => _trackedPointOfInterestIds;
  List<PointOfInterest> get pointsOfInterest => _pointsOfInterest;
  Map<String, List<QuestChallenge>> get pendingTasks => _pendingTasks;
  Map<String, List<QuestChallenge>> get completedTasks => _completedTasks;
  bool get loading => _loading;

  bool isRootNode(PointOfInterest poi) {
    return _quests.any(
      (q) => q.rootNode.pointOfInterest.id == poi.id,
    );
  }

  void _onZoneOrTagsChanged() {
    final zoneId = _zone.selectedZone?.id;
    final tagNames = _tagNamesFromSelection();
    if (zoneId == _lastZoneId &&
        _listEquals(tagNames, _lastTagNames)) {
      return;
    }
    _lastZoneId = zoneId;
    _lastTagNames = tagNames;
    if (zoneId != null && zoneId.isNotEmpty) {
      Future.microtask(() => refresh());
    } else {
      _quests = [];
      _trackedQuestIds = [];
      _trackedPointOfInterestIds = [];
      _pointsOfInterest = [];
      _pendingTasks = {};
      _completedTasks = {};
      notifyListeners();
    }
  }

  List<String> _tagNamesFromSelection() {
    final ids = _tags.selectedTagIds;
    return _tags.tags
        .where((t) => ids.contains(t.id))
        .map((t) => t.name)
        .toList();
  }

  bool _listEquals<T>(List<T> a, List<T> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }

  @override
  void dispose() {
    _zone.removeListener(_onZoneOrTagsChanged);
    _tags.removeListener(_onZoneOrTagsChanged);
    super.dispose();
  }

  Future<void> refresh() async {
    final zoneId = _zone.selectedZone?.id;
    if (zoneId == null || zoneId.isEmpty) return;
    _loading = true;
    notifyListeners();
    try {
      final tagNames = _tagNamesFromSelection();
      final log = await _service.getQuestLog(zoneId, tags: tagNames);
      _quests = log.quests;
      _pendingTasks = log.pendingTasks;
      _completedTasks = log.completedTasks;
      _trackedQuestIds = List.from(log.trackedQuestIds);
      _pointsOfInterest = getMapPointsOfInterest(log.quests);
      final tracked = log.quests
          .where((q) => log.trackedQuestIds.contains(q.id))
          .toList();
      _trackedPointOfInterestIds = tracked
          .expand((q) => getAllPointsOfInterestIdsForQuest(q))
          .toList();
      _lastZoneId = zoneId;
      _lastTagNames = tagNames;
    } catch (_) {
      _quests = [];
      _trackedQuestIds = [];
      _trackedPointOfInterestIds = [];
      _pointsOfInterest = [];
      _pendingTasks = {};
      _completedTasks = {};
    }
    _loading = false;
    notifyListeners();
  }

  Future<void> trackQuest(String questId) async {
    try {
      await _service.trackQuest(questId);
      await refresh();
    } catch (_) {}
  }

  Future<void> untrackQuest(String questId) async {
    try {
      await _service.untrackQuest(questId);
      await refresh();
    } catch (_) {}
  }

  Future<void> untrackAllQuests() async {
    try {
      await _service.untrackAllQuests();
      await refresh();
    } catch (_) {}
  }
}
