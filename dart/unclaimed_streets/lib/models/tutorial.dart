import 'character.dart';

class TutorialStatus {
  final bool showWelcomeDialogue;
  final bool isCompleted;
  final String stage;
  final Character? character;
  final List<String> dialogue;
  final List<String> loadoutDialogue;
  final String? scenarioId;
  final String? monsterEncounterId;
  final List<int> requiredEquipItemIds;
  final List<int> completedEquipItemIds;
  final List<int> requiredUseItemIds;
  final List<int> completedUseItemIds;

  const TutorialStatus({
    required this.showWelcomeDialogue,
    required this.isCompleted,
    required this.stage,
    this.character,
    this.dialogue = const [],
    this.loadoutDialogue = const [],
    this.scenarioId,
    this.monsterEncounterId,
    this.requiredEquipItemIds = const [],
    this.completedEquipItemIds = const [],
    this.requiredUseItemIds = const [],
    this.completedUseItemIds = const [],
  });

  bool get hasActiveScenario =>
      stage == 'scenario' &&
      (scenarioId?.trim().isNotEmpty ?? false) &&
      !isCompleted;

  bool get isLoadoutStep => stage == 'loadout' && !isCompleted;

  bool get hasActiveMonsterEncounter =>
      stage == 'monster' &&
      (monsterEncounterId?.trim().isNotEmpty ?? false) &&
      !isCompleted;

  bool get hasOutstandingLoadoutRequirements =>
      _remaining(requiredEquipItemIds, completedEquipItemIds).isNotEmpty ||
      _remaining(requiredUseItemIds, completedUseItemIds).isNotEmpty;

  factory TutorialStatus.fromJson(Map<String, dynamic> json) {
    Character? character;
    final rawCharacter = json['character'];
    if (rawCharacter is Map<String, dynamic>) {
      character = Character.fromJson(rawCharacter);
    } else if (rawCharacter is Map) {
      character = Character.fromJson(Map<String, dynamic>.from(rawCharacter));
    }

    final dialogue = <String>[];
    final rawDialogue = json['dialogue'];
    if (rawDialogue is List) {
      for (final line in rawDialogue) {
        final value = line?.toString().trim() ?? '';
        if (value.isNotEmpty) {
          dialogue.add(value);
        }
      }
    }

    final loadoutDialogue = <String>[];
    final rawLoadoutDialogue = json['loadoutDialogue'];
    if (rawLoadoutDialogue is List) {
      for (final line in rawLoadoutDialogue) {
        final value = line?.toString().trim() ?? '';
        if (value.isNotEmpty) {
          loadoutDialogue.add(value);
        }
      }
    }

    return TutorialStatus(
      showWelcomeDialogue: json['showWelcomeDialogue'] as bool? ?? false,
      isCompleted: json['completedAt'] != null,
      stage: json['stage']?.toString().trim() ?? '',
      character: character,
      dialogue: dialogue,
      loadoutDialogue: loadoutDialogue,
      scenarioId: json['scenarioId']?.toString(),
      monsterEncounterId: json['monsterEncounterId']?.toString(),
      requiredEquipItemIds: _parseIntList(json['requiredEquipItemIds']),
      completedEquipItemIds: _parseIntList(json['completedEquipItemIds']),
      requiredUseItemIds: _parseIntList(json['requiredUseItemIds']),
      completedUseItemIds: _parseIntList(json['completedUseItemIds']),
    );
  }

  TutorialStatus copyWith({
    bool? showWelcomeDialogue,
    bool? isCompleted,
    String? stage,
    Character? character,
    List<String>? dialogue,
    List<String>? loadoutDialogue,
    String? scenarioId,
    String? monsterEncounterId,
    List<int>? requiredEquipItemIds,
    List<int>? completedEquipItemIds,
    List<int>? requiredUseItemIds,
    List<int>? completedUseItemIds,
  }) {
    return TutorialStatus(
      showWelcomeDialogue: showWelcomeDialogue ?? this.showWelcomeDialogue,
      isCompleted: isCompleted ?? this.isCompleted,
      stage: stage ?? this.stage,
      character: character ?? this.character,
      dialogue: dialogue ?? this.dialogue,
      loadoutDialogue: loadoutDialogue ?? this.loadoutDialogue,
      scenarioId: scenarioId ?? this.scenarioId,
      monsterEncounterId: monsterEncounterId ?? this.monsterEncounterId,
      requiredEquipItemIds: requiredEquipItemIds ?? this.requiredEquipItemIds,
      completedEquipItemIds:
          completedEquipItemIds ?? this.completedEquipItemIds,
      requiredUseItemIds: requiredUseItemIds ?? this.requiredUseItemIds,
      completedUseItemIds: completedUseItemIds ?? this.completedUseItemIds,
    );
  }

  List<int> get remainingEquipItemIds =>
      _remaining(requiredEquipItemIds, completedEquipItemIds);

  List<int> get remainingUseItemIds =>
      _remaining(requiredUseItemIds, completedUseItemIds);

  static List<int> _parseIntList(dynamic raw) {
    final values = <int>[];
    if (raw is List) {
      for (final entry in raw) {
        final value = entry is num ? entry.toInt() : int.tryParse('$entry');
        if (value != null && value > 0 && !values.contains(value)) {
          values.add(value);
        }
      }
    }
    return values;
  }

  static List<int> _remaining(List<int> required, List<int> completed) {
    final done = completed.toSet();
    return required.where((id) => !done.contains(id)).toList(growable: false);
  }
}
