import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:dio/dio.dart';

import '../models/point_of_interest.dart';
import '../models/fetch_quest_turn_in.dart';
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
    if (quest.readyToTurnIn) continue;
    final node = quest.currentNode;
    if (node?.pointOfInterest != null) {
      out.add(node!.pointOfInterest!);
    }
  }
  return out;
}

/// All POI IDs in a quest's current node (for tracking focus).
List<String> getAllPointsOfInterestIdsForQuest(Quest quest) {
  if (quest.readyToTurnIn) return [];
  final node = quest.currentNode;
  if (node?.pointOfInterest == null) return [];
  return [node!.pointOfInterest!.id];
}

class QuestLogProvider with ChangeNotifier {
  static const Duration _mutationRefreshTimeout = Duration(seconds: 8);

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
    _tags.addListener(_onZoneOrTagsChanged);
    _filters.addListener(_onZoneOrTagsChanged);
  }

  List<Quest> get quests => _quests;
  List<Quest> get completedQuests => _completedQuests;
  List<String> get trackedQuestIds => _trackedQuestIds;
  List<String> get trackedPointOfInterestIds => _trackedPointOfInterestIds;
  List<PointOfInterest> get pointsOfInterest => _pointsOfInterest;
  List<String> get currentNodePoiIds => _currentNodePoiIds;
  List<List<QuestNodePolygonPoint>> get currentNodePolygons =>
      _currentNodePolygons;
  bool get loading => _loading;

  bool isRootNode(PointOfInterest poi) {
    return _quests.any((q) => q.currentNode?.pointOfInterest?.id == poi.id);
  }

  List<String> _normalizeTrackedQuestIds(List<String> trackedQuestIds) {
    final normalized = <String>[];
    final seen = <String>{};
    for (final questId in trackedQuestIds) {
      final trimmed = questId.trim();
      if (trimmed.isEmpty || !seen.add(trimmed)) continue;
      normalized.add(trimmed);
    }
    return normalized;
  }

  List<Quest> _promoteTutorialTrackedQuests(
    List<Quest> quests,
    List<String> trackedQuestIds,
  ) {
    final trackedIds = trackedQuestIds.toSet();
    if (trackedIds.isEmpty) {
      return List<Quest>.from(quests);
    }

    final tutorialTracked = <Quest>[];
    final remaining = <Quest>[];
    for (final quest in quests) {
      final questId = quest.id.trim();
      if (quest.isTutorial && trackedIds.contains(questId)) {
        tutorialTracked.add(quest);
      } else {
        remaining.add(quest);
      }
    }

    if (tutorialTracked.isEmpty) {
      return List<Quest>.from(quests);
    }
    return [...tutorialTracked, ...remaining];
  }

  List<String> _promoteTutorialTrackedQuestIds(
    List<String> trackedQuestIds,
    List<Quest> orderedQuests,
  ) {
    if (trackedQuestIds.isEmpty) {
      return const [];
    }

    final trackedIdSet = trackedQuestIds.toSet();
    final tutorialTrackedIds = <String>[];
    final seen = <String>{};

    for (final quest in orderedQuests) {
      final questId = quest.id.trim();
      if (!quest.isTutorial ||
          !trackedIdSet.contains(questId) ||
          !seen.add(questId)) {
        continue;
      }
      tutorialTrackedIds.add(questId);
    }

    if (tutorialTrackedIds.isEmpty) {
      return List<String>.from(trackedQuestIds);
    }

    final orderedIds = <String>[...tutorialTrackedIds];
    for (final questId in trackedQuestIds) {
      if (seen.add(questId)) {
        orderedIds.add(questId);
      }
    }
    return orderedIds;
  }

  List<Quest> _trackedQuestsForState(
    List<Quest> quests,
    List<String> trackedQuestIds,
  ) {
    final trackedIds = trackedQuestIds.toSet();
    if (trackedIds.isEmpty) {
      return const [];
    }
    return quests
        .where((quest) => trackedIds.contains(quest.id.trim()))
        .toList();
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
    if (!_filters.enableTagFilter) return [];
    final selectedIds = _tags.selectedTagIds;
    if (selectedIds.isEmpty) return [];
    final names = <String>[];
    for (final tag in _tags.tags) {
      if (selectedIds.contains(tag.id) && tag.name.isNotEmpty) {
        names.add(tag.name);
      }
    }
    names.sort();
    return names;
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
      final normalizedTrackedQuestIds = _normalizeTrackedQuestIds(
        log.trackedQuestIds,
      );
      final orderedQuests = _promoteTutorialTrackedQuests(
        log.quests,
        normalizedTrackedQuestIds,
      );
      final orderedTrackedQuestIds = _promoteTutorialTrackedQuestIds(
        normalizedTrackedQuestIds,
        orderedQuests,
      );
      _quests = orderedQuests;
      _completedQuests = log.completedQuests;
      _trackedQuestIds = orderedTrackedQuestIds;
      _pointsOfInterest = getMapPointsOfInterest(orderedQuests);
      final tracked = _trackedQuestsForState(
        orderedQuests,
        orderedTrackedQuestIds,
      );
      _trackedPointOfInterestIds = tracked
          .expand((q) => getAllPointsOfInterestIdsForQuest(q))
          .toList();
      _currentNodePoiIds = orderedQuests
          .where((q) => q.isAccepted)
          .expand((q) => getAllPointsOfInterestIdsForQuest(q))
          .toList();
      _currentNodePolygons = orderedQuests
          .where((q) => q.isAccepted && !q.readyToTurnIn)
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
    unawaited(_refreshAfterMutation('turnInQuest'));
    return resp;
  }

  Future<Map<String, dynamic>> submitQuestNode(
    String questNodeId, {
    String? textSubmission,
    String? imageSubmissionUrl,
    String? videoSubmissionUrl,
  }) async {
    final resp = await _service.submitQuestNode(
      questNodeId,
      textSubmission: textSubmission,
      imageSubmissionUrl: imageSubmissionUrl,
      videoSubmissionUrl: videoSubmissionUrl,
    );
    unawaited(_refreshAfterMutation('submitQuestNode'));
    return resp;
  }

  Future<FetchQuestTurnInDetails> getFetchQuestTurnIn(String questId) {
    return _service.getFetchQuestTurnIn(questId);
  }

  Future<Map<String, dynamic>> submitFetchQuestTurnIn(String questId) async {
    final resp = await _service.submitFetchQuestTurnIn(questId);
    unawaited(_refreshAfterMutation('submitFetchQuestTurnIn'));
    return resp;
  }

  Future<void> _refreshAfterMutation(String source) async {
    try {
      await refresh().timeout(_mutationRefreshTimeout);
    } catch (error) {
      debugPrint('[QuestLogProvider] $source refresh skipped: $error');
    }
  }

  Future<String?> shareQuest(String questId, String targetUserId) async {
    try {
      await _service.shareQuest(questId, targetUserId);
      return null;
    } catch (e) {
      if (e is DioException) {
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final message = data['error'];
          if (message is String && message.trim().isNotEmpty) {
            return message.trim();
          }
        }
      }
      return 'Failed to share quest.';
    }
  }

  Future<String?> forgetQuest(String questId) async {
    try {
      await _service.forgetQuest(questId);
      await _refreshAfterMutation('forgetQuest');
      return null;
    } catch (e) {
      if (e is DioException) {
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final message = data['error'];
          if (message is String && message.trim().isNotEmpty) {
            return message.trim();
          }
        }
      }
      return 'Failed to forget quest.';
    }
  }
}
