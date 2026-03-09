import 'character.dart';

class TutorialStatus {
  final bool showWelcomeDialogue;
  final Character? character;
  final List<String> dialogue;
  final String? scenarioId;

  const TutorialStatus({
    required this.showWelcomeDialogue,
    this.character,
    this.dialogue = const [],
    this.scenarioId,
  });

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

    return TutorialStatus(
      showWelcomeDialogue: json['showWelcomeDialogue'] as bool? ?? false,
      character: character,
      dialogue: dialogue,
      scenarioId: json['scenarioId']?.toString(),
    );
  }

  TutorialStatus copyWith({
    bool? showWelcomeDialogue,
    Character? character,
    List<String>? dialogue,
    String? scenarioId,
  }) {
    return TutorialStatus(
      showWelcomeDialogue: showWelcomeDialogue ?? this.showWelcomeDialogue,
      character: character ?? this.character,
      dialogue: dialogue ?? this.dialogue,
      scenarioId: scenarioId ?? this.scenarioId,
    );
  }
}
