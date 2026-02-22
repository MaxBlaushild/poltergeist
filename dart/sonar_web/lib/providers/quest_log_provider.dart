import 'package:flutter/foundation.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../providers/quest_filter_provider.dart';
import '../providers/tags_provider.dart';
import '../providers/zone_provider.dart';
import '../services/quest_log_service.dart';

/// POIs that appear in the quest log (current node only).
List<PointOfInterest> getMapPointsOfInterest(List<Quest> quests) {
  final out = <PointOfInterest>[];
  for (final quest in quests) {
    final node = quest.currentNode;
    if (node?.pointOfInterest != null) {
      out.add(node!.pointOfInterest!);
    }
  }
  return out;
}

/// All POI IDs in a quest's current node (for tracking focus).
List<String> getAllPointsOfInterestIdsForQuest(Quest quest) {
  final node = quest.currentNode;
  if (node?.pointOfInterest == null) return [];
  return [node!.pointOfInterest!.id];
}

class QuestLogProvider with ChangeNotifier {
  final QuestLogService _service;
  final ZoneProvider _zone;
  final TagsProvider _tags;
  final QuestFilterProvider _filters;

  List<Quest> _quests = [];
  List<Quest> _completedQuests = [];
  List<String> _trackedQuestIds = [];
  List<String> _trackedPointOfInterestIds = [];
  List<PointOfInterest> _pointsOfInterest = [];
  List<String> _currentNodePoiIds = [];
  List<List<QuestNodePolygonPoint>> _currentNodePolygons = [];
  bool _loading = false;
  String? _lastZoneId;
  List<String> _lastTagNames = [];

  QuestLogProvider(this._service, this._zone, this._tags, this._filters) {
    _zone.addListener(_onZoneOrTagsChanged);
    _filters.addListener(_onZoneOrTagsChanged);
  }

  List<Quest> get quests => _quests;
  List<Quest> get completedQuests => _completedQuests;
  List<String> get trackedQuestIds => _trackedQuestIds;
  List<String> get trackedPointOfInterestIds => _trackedPointOfInterestIds;
  List<PointOfInterest> get pointsOfInterest => _pointsOfInterest;
  List<String> get currentNodePoiIds => _currentNodePoiIds;
  List<List<QuestNodePolygonPoint>> get currentNodePolygons => _currentNodePolygons;
  bool get loading => _loading;

  bool isRootNode(PointOfInterest poi) {
    return _quests.any(
      (q) => q.currentNode?.pointOfInterest?.id == poi.id,
    );
  }

  void _onZoneOrTagsChanged() {
    final zoneId = _zone.selectedZone?.id;
    final tagNames = _tagNamesFromSelection();
    final effectiveZoneId = zoneId ?? '';
    if (effectiveZoneId == (_lastZoneId ?? '') &&
        _listEquals(tagNames, _lastTagNames)) {
      return;
    }
    _lastZoneId = effectiveZoneId;
    _lastTagNames = tagNames;
    Future.microtask(() => refresh());
  }

  List<String> _tagNamesFromSelection() {
    return [];
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
    _filters.removeListener(_onZoneOrTagsChanged);
    super.dispose();
  }

  Future<void> refresh() async {
    final zoneId = _zone.selectedZone?.id;
    _loading = true;
    notifyListeners();
    try {
      final tagNames = _tagNamesFromSelection();
      final log = await _service.getQuestLog(zoneId: zoneId, tags: tagNames);
      _quests = log.quests;
      _completedQuests = log.completedQuests;
      _trackedQuestIds = List.from(log.trackedQuestIds);
      _pointsOfInterest = getMapPointsOfInterest(log.quests);
      final tracked = log.quests
          .where((q) => log.trackedQuestIds.contains(q.id))
          .toList();
      _trackedPointOfInterestIds = tracked
          .expand((q) => getAllPointsOfInterestIdsForQuest(q))
          .toList();
      _currentNodePoiIds = log.quests
          .where((q) => q.isAccepted)
          .expand((q) => getAllPointsOfInterestIdsForQuest(q))
          .toList();
      _currentNodePolygons = log.quests
          .where((q) => q.isAccepted)
          .map((q) => q.currentNode?.polygon ?? const <QuestNodePolygonPoint>[])
          .where((poly) => poly.isNotEmpty)
          .map((poly) => List<QuestNodePolygonPoint>.from(poly))
          .toList();
      _lastZoneId = zoneId ?? '';
      _lastTagNames = tagNames;
    } catch (_) {
      _quests = [];
      _completedQuests = [];
      _trackedQuestIds = [];
      _trackedPointOfInterestIds = [];
      _pointsOfInterest = [];
      _currentNodePoiIds = [];
      _currentNodePolygons = [];
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

  /// Turn in a completed quest. Returns the response (goldAwarded, itemAwarded).
  Future<Map<String, dynamic>> turnInQuest(String questId) async {
    final resp = await _service.turnInQuest(questId);
    await refresh();
    return resp;
  }

  Future<Map<String, dynamic>> submitQuestNodeChallenge(
    String questNodeId, {
    String? questNodeChallengeId,
    String? textSubmission,
    String? imageSubmissionUrl,
    String? videoSubmissionUrl,
  }) async {
    final resp = await _service.submitQuestNodeChallenge(
      questNodeId,
      questNodeChallengeId: questNodeChallengeId,
      textSubmission: textSubmission,
      imageSubmissionUrl: imageSubmissionUrl,
      videoSubmissionUrl: videoSubmissionUrl,
    );
    await refresh();
    return resp;
  }
}
