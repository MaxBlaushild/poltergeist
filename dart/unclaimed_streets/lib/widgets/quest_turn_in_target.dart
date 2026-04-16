import 'package:flutter/material.dart';

import '../models/character.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';

bool questIsAwaitingTurnIn(Quest quest) {
  return quest.readyToTurnIn ||
      (quest.turnedInAt == null &&
          quest.isAccepted &&
          quest.currentNode == null);
}

String questTurnInCharacterImageUrl(Character? character) {
  if (character == null) return '';

  final thumbnail = character.thumbnailUrl?.trim() ?? '';
  if (thumbnail.isNotEmpty) return thumbnail;

  final dialogue = character.dialogueImageUrl?.trim() ?? '';
  if (dialogue.isNotEmpty) return dialogue;

  final mapIcon = character.mapIconUrl?.trim() ?? '';
  if (mapIcon.isNotEmpty) return mapIcon;

  return '';
}

String questTurnInReceiverLabel(Character? character) {
  final name = character?.name.trim() ?? '';
  return name.isNotEmpty ? name : 'quest receiver';
}

String questTurnInLocationLabel({
  Character? character,
  PointOfInterest? pointOfInterest,
}) {
  final poiName = pointOfInterest?.name.trim() ?? '';
  if (poiName.isNotEmpty) return poiName;

  final characterName = character?.name.trim() ?? '';
  if (characterName.isNotEmpty) return '$characterName\'s location';

  return 'their location';
}

class QuestTurnInPortrait extends StatelessWidget {
  const QuestTurnInPortrait({
    super.key,
    required this.character,
    required this.size,
    required this.backgroundColor,
    required this.foregroundColor,
  });

  final Character? character;
  final double size;
  final Color backgroundColor;
  final Color foregroundColor;

  @override
  Widget build(BuildContext context) {
    final imageUrl = questTurnInCharacterImageUrl(character);

    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(size * 0.28),
      ),
      clipBehavior: Clip.antiAlias,
      child: imageUrl.isNotEmpty
          ? Image.network(
              imageUrl,
              fit: BoxFit.cover,
              errorBuilder: (_, _, _) =>
                  _QuestTurnInFallbackIcon(foregroundColor: foregroundColor),
            )
          : _QuestTurnInFallbackIcon(foregroundColor: foregroundColor),
    );
  }
}

class _QuestTurnInFallbackIcon extends StatelessWidget {
  const _QuestTurnInFallbackIcon({required this.foregroundColor});

  final Color foregroundColor;

  @override
  Widget build(BuildContext context) {
    return Center(child: Icon(Icons.person, color: foregroundColor, size: 22));
  }
}
