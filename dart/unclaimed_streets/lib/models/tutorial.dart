import 'character.dart';
import 'character_action.dart';

class TutorialStatus {
  final bool showWelcomeDialogue;
  final bool isCompleted;
  final String stage;
  final Character? character;
  final List<DialogueMessage> dialogue;
  final List<DialogueMessage> loadoutDialogue;
  final List<DialogueMessage> postMonsterDialogue;
  final List<DialogueMessage> baseKitDialogue;
  final List<DialogueMessage> postBaseDialogue;
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
    this.postMonsterDialogue = const [],
    this.baseKitDialogue = const [],
    this.postBaseDialogue = const [],
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

  bool get isBaseKitStep => stage == 'base_kit' && !isCompleted;

  bool get isPostMonsterDialogueStep =>
      stage == 'post_monster_dialogue' && !isCompleted;

  bool get isPostBaseDialogueStep =>
      stage == 'post_base_dialogue' && !isCompleted;

  bool get hasActiveMonsterEncounter =>
      stage == 'monster' &&
      (monsterEncounterId?.trim().isNotEmpty ?? false) &&
      !isCompleted;

  bool get shouldShowPostMonsterDialogue =>
      isPostMonsterDialogueStep &&
      postMonsterDialogue.isNotEmpty &&
      !isCompleted;

  bool get shouldShowPostBaseDialogue =>
      isPostBaseDialogueStep && postBaseDialogue.isNotEmpty && !isCompleted;

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

    final dialogue = <DialogueMessage>[];
    final rawDialogue = json['dialogue'];
    if (rawDialogue is List) {
      for (var index = 0; index < rawDialogue.length; index++) {
        final line = rawDialogue[index];
        if (line is Map<String, dynamic>) {
          dialogue.add(DialogueMessage.fromJson(line));
        } else if (line is Map) {
          dialogue.add(
            DialogueMessage.fromJson(Map<String, dynamic>.from(line)),
          );
        } else {
          final value = line?.toString().trim() ?? '';
          if (value.isNotEmpty) {
            dialogue.add(
              DialogueMessage(speaker: 'character', text: value, order: index),
            );
          }
        }
      }
    }

    final loadoutDialogue = <DialogueMessage>[];
    final rawLoadoutDialogue = json['loadoutDialogue'];
    if (rawLoadoutDialogue is List) {
      for (var index = 0; index < rawLoadoutDialogue.length; index++) {
        final line = rawLoadoutDialogue[index];
        if (line is Map<String, dynamic>) {
          loadoutDialogue.add(DialogueMessage.fromJson(line));
        } else if (line is Map) {
          loadoutDialogue.add(
            DialogueMessage.fromJson(Map<String, dynamic>.from(line)),
          );
        } else {
          final value = line?.toString().trim() ?? '';
          if (value.isNotEmpty) {
            loadoutDialogue.add(
              DialogueMessage(speaker: 'character', text: value, order: index),
            );
          }
        }
      }
    }

    final postMonsterDialogue = <DialogueMessage>[];
    final rawPostMonsterDialogue = json['postMonsterDialogue'];
    if (rawPostMonsterDialogue is List) {
      for (var index = 0; index < rawPostMonsterDialogue.length; index++) {
        final line = rawPostMonsterDialogue[index];
        if (line is Map<String, dynamic>) {
          postMonsterDialogue.add(DialogueMessage.fromJson(line));
        } else if (line is Map) {
          postMonsterDialogue.add(
            DialogueMessage.fromJson(Map<String, dynamic>.from(line)),
          );
        } else {
          final value = line?.toString().trim() ?? '';
          if (value.isNotEmpty) {
            postMonsterDialogue.add(
              DialogueMessage(speaker: 'character', text: value, order: index),
            );
          }
        }
      }
    }

    final baseKitDialogue = <DialogueMessage>[];
    final rawBaseKitDialogue = json['baseKitDialogue'];
    if (rawBaseKitDialogue is List) {
      for (var index = 0; index < rawBaseKitDialogue.length; index++) {
        final line = rawBaseKitDialogue[index];
        if (line is Map<String, dynamic>) {
          baseKitDialogue.add(DialogueMessage.fromJson(line));
        } else if (line is Map) {
          baseKitDialogue.add(
            DialogueMessage.fromJson(Map<String, dynamic>.from(line)),
          );
        } else {
          final value = line?.toString().trim() ?? '';
          if (value.isNotEmpty) {
            baseKitDialogue.add(
              DialogueMessage(speaker: 'character', text: value, order: index),
            );
          }
        }
      }
    }

    final postBaseDialogue = <DialogueMessage>[];
    final rawPostBaseDialogue = json['postBaseDialogue'];
    if (rawPostBaseDialogue is List) {
      for (var index = 0; index < rawPostBaseDialogue.length; index++) {
        final line = rawPostBaseDialogue[index];
        if (line is Map<String, dynamic>) {
          postBaseDialogue.add(DialogueMessage.fromJson(line));
        } else if (line is Map) {
          postBaseDialogue.add(
            DialogueMessage.fromJson(Map<String, dynamic>.from(line)),
          );
        } else {
          final value = line?.toString().trim() ?? '';
          if (value.isNotEmpty) {
            postBaseDialogue.add(
              DialogueMessage(speaker: 'character', text: value, order: index),
            );
          }
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
      postMonsterDialogue: postMonsterDialogue,
      baseKitDialogue: baseKitDialogue,
      postBaseDialogue: postBaseDialogue,
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
    List<DialogueMessage>? dialogue,
    List<DialogueMessage>? loadoutDialogue,
    List<DialogueMessage>? postMonsterDialogue,
    List<DialogueMessage>? baseKitDialogue,
    List<DialogueMessage>? postBaseDialogue,
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
      postMonsterDialogue: postMonsterDialogue ?? this.postMonsterDialogue,
      baseKitDialogue: baseKitDialogue ?? this.baseKitDialogue,
      postBaseDialogue: postBaseDialogue ?? this.postBaseDialogue,
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
