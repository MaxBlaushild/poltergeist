import 'point_of_interest.dart';

/// Lightweight challenge reference in objectives/tasks.
class QuestChallenge {
  final String id;
  final String question;
  final String pointOfInterestId;

  const QuestChallenge({
    required this.id,
    required this.question,
    required this.pointOfInterestId,
  });

  factory QuestChallenge.fromJson(Map<String, dynamic> json) {
    return QuestChallenge(
      id: json['id']?.toString() ?? '',
      question: json['question'] as String? ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString() ?? '',
    );
  }
}

class QuestObjective {
  final QuestChallenge challenge;
  final bool isCompleted;
  final QuestNode? nextNode;

  const QuestObjective({
    required this.challenge,
    required this.isCompleted,
    this.nextNode,
  });

  factory QuestObjective.fromJson(Map<String, dynamic> json) {
    QuestNode? next;
    final n = json['nextNode'];
    if (n is Map<String, dynamic>) {
      try {
        next = QuestNode.fromJson(n);
      } catch (_) {}
    }
    final ch = json['challenge'];
    return QuestObjective(
      challenge: ch is Map<String, dynamic>
          ? QuestChallenge.fromJson(ch)
          : const QuestChallenge(id: '', question: '', pointOfInterestId: ''),
      isCompleted: json['isCompleted'] as bool? ?? false,
      nextNode: next,
    );
  }
}

class QuestNode {
  final PointOfInterest pointOfInterest;
  final List<QuestObjective> objectives;

  const QuestNode({
    required this.pointOfInterest,
    this.objectives = const [],
  });

  factory QuestNode.fromJson(Map<String, dynamic> json) {
    final raw = json['objectives'];
    final list = <QuestObjective>[];
    if (raw is List) {
      for (final e in raw) {
        if (e is Map<String, dynamic>) {
          try {
            list.add(QuestObjective.fromJson(e));
          } catch (_) {}
        }
      }
    }
    final poi = json['pointOfInterest'];
    return QuestNode(
      pointOfInterest: poi is Map<String, dynamic>
          ? PointOfInterest.fromJson(poi)
          : throw Exception('missing pointOfInterest'),
      objectives: list,
    );
  }
}

class Quest {
  final String id;
  final String name;
  final String description;
  final String imageUrl;
  final bool isCompleted;
  final QuestNode rootNode;
  final int gold;
  final int? inventoryItemId;
  final String? questGiverCharacterId;
  final DateTime? turnedInAt;
  final bool readyToTurnIn;

  const Quest({
    required this.id,
    required this.name,
    required this.description,
    required this.imageUrl,
    required this.isCompleted,
    required this.rootNode,
    this.gold = 0,
    this.inventoryItemId,
    this.questGiverCharacterId,
    this.turnedInAt,
    this.readyToTurnIn = false,
  });

  factory Quest.fromJson(Map<String, dynamic> json) {
    final root = json['rootNode'];
    DateTime? turnedInAt;
    final t = json['turnedInAt'];
    if (t != null) {
      if (t is String) {
        turnedInAt = DateTime.tryParse(t);
      }
    }
    int? invItemId;
    final inv = json['inventoryItemId'];
    if (inv != null) {
      if (inv is int) {
        invItemId = inv;
      } else if (inv is num) {
        invItemId = inv.toInt();
      }
    }
    return Quest(
      id: json['id']?.toString() ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      imageUrl: json['imageUrl'] as String? ?? '',
      isCompleted: json['isCompleted'] as bool? ?? false,
      rootNode: root is Map<String, dynamic>
          ? QuestNode.fromJson(root)
          : throw Exception('missing rootNode'),
      gold: (json['gold'] as num?)?.toInt() ?? 0,
      inventoryItemId: invItemId,
      questGiverCharacterId: json['questGiverCharacterId']?.toString(),
      turnedInAt: turnedInAt,
      readyToTurnIn: json['readyToTurnIn'] as bool? ?? false,
    );
  }
}

class QuestLog {
  final List<Quest> quests;
  final Map<String, List<QuestChallenge>> pendingTasks;
  final Map<String, List<QuestChallenge>> completedTasks;
  final List<String> trackedQuestIds;

  const QuestLog({
    this.quests = const [],
    this.pendingTasks = const {},
    this.completedTasks = const {},
    this.trackedQuestIds = const [],
  });

  factory QuestLog.fromJson(Map<String, dynamic> json) {
    final qList = json['quests'];
    final quests = <Quest>[];
    if (qList is List) {
      for (final e in qList) {
        if (e is Map<String, dynamic>) {
          try {
            quests.add(Quest.fromJson(e));
          } catch (_) {}
        }
      }
    }
    final tracked = json['trackedQuestIds'];
    final trackedIds = <String>[];
    if (tracked is List) {
      for (final e in tracked) {
        final s = e?.toString();
        if (s != null && s.isNotEmpty) trackedIds.add(s);
      }
    }
    final pending = _parseTaskMap(json['pendingTasks']);
    final completed = _parseTaskMap(json['completedTasks']);
    return QuestLog(
      quests: quests,
      pendingTasks: pending,
      completedTasks: completed,
      trackedQuestIds: trackedIds,
    );
  }

  static Map<String, List<QuestChallenge>> _parseTaskMap(dynamic raw) {
    final out = <String, List<QuestChallenge>>{};
    if (raw == null || raw is! Map) return out;
    final map = Map<String, dynamic>.from(raw as Map);
    for (final entry in map.entries) {
      final key = entry.key.toString();
      final list = <QuestChallenge>[];
      final arr = entry.value;
      if (arr is List) {
        for (final e in arr) {
          if (e is Map<String, dynamic>) {
            try {
              final ch = e['challenge'];
              if (ch is Map<String, dynamic>) {
                list.add(QuestChallenge.fromJson(ch));
              }
            } catch (_) {}
          }
        }
      }
      out[key] = list;
    }
    return out;
  }
}

/// Collect all tag names from POIs in the quest tree.
List<String> getQuestTags(Quest quest) {
  final names = <String>{};
  void visit(QuestNode node) {
    for (final t in node.pointOfInterest.tags) {
      names.add(t.name);
    }
    for (final o in node.objectives) {
      final n = o.nextNode;
      if (n != null) visit(n);
    }
  }
  visit(quest.rootNode);
  return names.toList();
}
