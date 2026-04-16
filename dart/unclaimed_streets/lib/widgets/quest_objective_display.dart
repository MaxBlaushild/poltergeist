import 'package:flutter/material.dart';

import '../models/point_of_interest.dart';
import '../models/quest_node.dart';
import '../models/quest_node_objective.dart';

const hiddenPoiPlaceholderImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';

List<String> questObjectiveLines(QuestNode? node) {
  if (node == null) return const [];

  final objectiveText = node.objectiveText.trim();
  if (objectiveText.isNotEmpty) {
    return [objectiveText];
  }

  final objectivePrompt = node.objective?.prompt.trim() ?? '';
  if (objectivePrompt.isNotEmpty) {
    return [objectivePrompt];
  }

  final objectiveType = node.objective?.type.trim() ?? '';

  if (_hasValue(node.fetchCharacterId) ||
      objectiveType == QuestNodeObjective.typeFetchQuest) {
    final characterName = node.fetchCharacter?.name.trim() ?? '';
    if (characterName.isNotEmpty) {
      return ['Deliver the required items to $characterName.'];
    }
    return const ['Deliver the required items to continue the quest.'];
  }
  if (_hasValue(node.scenarioId) ||
      objectiveType == QuestNodeObjective.typeScenario) {
    return const ['Complete the current scenario objective.'];
  }
  if (_hasValue(node.expositionId) ||
      objectiveType == QuestNodeObjective.typeExposition) {
    final title = node.exposition?.title.trim() ?? '';
    if (title.isNotEmpty) {
      return ['Complete the exposition: $title'];
    }
    return const ['Complete the current exposition objective.'];
  }
  if (_hasValue(node.monsterEncounterId) ||
      _hasValue(node.monsterId) ||
      objectiveType == QuestNodeObjective.typeMonsterEncounter ||
      objectiveType == QuestNodeObjective.typeMonster) {
    switch (_questEncounterType(node)) {
      case 'boss':
        return const ['Defeat the boss encounter.'];
      case 'raid':
        return const ['Complete the raid encounter.'];
      default:
        return const ['Defeat the current monster objective.'];
    }
  }
  if (_hasValue(node.challengeId) ||
      objectiveType == QuestNodeObjective.typeChallenge) {
    return const ['Complete the current challenge objective.'];
  }
  if (objectiveType == QuestNodeObjective.typeStoryFlag) {
    return const ['Continue the story until this objective completes.'];
  }

  final poiName = node.pointOfInterest?.name.trim() ?? '';
  if (poiName.isNotEmpty) {
    return ['Travel to $poiName.'];
  }
  return const ['Complete the current objective.'];
}

String questObjectiveSummary(QuestNode? node) {
  final lines = questObjectiveLines(node);
  if (lines.isEmpty) return '';
  return lines.join(' • ');
}

String? questObjectiveChallengeLabel(QuestNode? node) {
  switch (_questEncounterType(node)) {
    case 'boss':
      return 'Boss Encounter';
    case 'raid':
      return 'Raid Encounter';
    default:
      return null;
  }
}

bool isQuestPointOfInterestHidden(
  QuestNode? node,
  Set<String> discoveredPoiIds,
) {
  final poi = node?.pointOfInterest;
  if (poi == null) return false;
  return !discoveredPoiIds.contains(poi.id);
}

bool questNodeHasDirectFocusTarget(QuestNode? node) {
  if (node == null) return false;
  if (node.pointOfInterest != null) return true;
  if (_hasValue(node.fetchCharacterId) || _hasFetchCharacterLocation(node)) {
    return true;
  }
  return _hasValue(node.scenarioId) ||
      _hasValue(node.expositionId) ||
      _hasValue(node.monsterEncounterId) ||
      _hasValue(node.monsterId) ||
      _hasValue(node.challengeId) ||
      node.polygon.isNotEmpty;
}

class QuestObjectiveIcon extends StatelessWidget {
  const QuestObjectiveIcon({
    super.key,
    required this.node,
    required this.discoveredPoiIds,
    required this.size,
    required this.borderRadius,
    required this.iconColor,
    required this.backgroundColor,
  });

  final QuestNode? node;
  final Set<String> discoveredPoiIds;
  final double size;
  final double borderRadius;
  final Color iconColor;
  final Color backgroundColor;

  @override
  Widget build(BuildContext context) {
    final imageUrl = _objectiveImageUrl(node, discoveredPoiIds);
    if (imageUrl != null) {
      return ClipRRect(
        borderRadius: BorderRadius.circular(borderRadius),
        child: Image.network(
          imageUrl,
          width: size,
          height: size,
          fit: BoxFit.cover,
          errorBuilder: (_, _, _) => _ObjectiveIconFallback(
            node: node,
            size: size,
            borderRadius: borderRadius,
            iconColor: iconColor,
            backgroundColor: backgroundColor,
          ),
        ),
      );
    }

    return _ObjectiveIconFallback(
      node: node,
      size: size,
      borderRadius: borderRadius,
      iconColor: iconColor,
      backgroundColor: backgroundColor,
    );
  }
}

class QuestObjectiveChallengeBadge extends StatelessWidget {
  const QuestObjectiveChallengeBadge({super.key, required this.node});

  final QuestNode? node;

  @override
  Widget build(BuildContext context) {
    final label = questObjectiveChallengeLabel(node);
    if (label == null) {
      return const SizedBox.shrink();
    }
    final colors = _questChallengeBadgeColors(node);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: colors.background,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: colors.border),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(_questChallengeIcon(node), size: 14, color: colors.foreground),
          const SizedBox(width: 6),
          Text(
            label,
            style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: colors.foreground,
              fontWeight: FontWeight.w800,
              letterSpacing: 0.2,
            ),
          ),
        ],
      ),
    );
  }
}

class _ObjectiveIconFallback extends StatelessWidget {
  const _ObjectiveIconFallback({
    required this.node,
    required this.size,
    required this.borderRadius,
    required this.iconColor,
    required this.backgroundColor,
  });

  final QuestNode? node;
  final double size;
  final double borderRadius;
  final Color iconColor;
  final Color backgroundColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(borderRadius),
      ),
      child: Icon(
        _objectiveIconData(node),
        size: size * 0.55,
        color: iconColor,
      ),
    );
  }
}

String? _objectiveImageUrl(QuestNode? node, Set<String> discoveredPoiIds) {
  final poi = node?.pointOfInterest;
  if (poi != null) {
    if (!discoveredPoiIds.contains(poi.id)) {
      return hiddenPoiPlaceholderImageUrl;
    }
    final poiImageUrl = _pointOfInterestImageUrl(poi);
    if (poiImageUrl != null) {
      return poiImageUrl;
    }
  }

  final objectiveThumbnail = node?.objective?.thumbnailUrl.trim() ?? '';
  if (objectiveThumbnail.isNotEmpty) {
    return objectiveThumbnail;
  }
  final objectiveImage = node?.objective?.imageUrl.trim() ?? '';
  if (objectiveImage.isNotEmpty) {
    return objectiveImage;
  }

  final fetchCharacterThumb = node?.fetchCharacter?.thumbnailUrl?.trim() ?? '';
  if (fetchCharacterThumb.isNotEmpty) {
    return fetchCharacterThumb;
  }
  final fetchCharacterDialogueImage =
      node?.fetchCharacter?.dialogueImageUrl?.trim() ?? '';
  if (fetchCharacterDialogueImage.isNotEmpty) {
    return fetchCharacterDialogueImage;
  }
  final fetchCharacterMapIcon = node?.fetchCharacter?.mapIconUrl?.trim() ?? '';
  if (fetchCharacterMapIcon.isNotEmpty) {
    return fetchCharacterMapIcon;
  }
  final expositionThumb = node?.exposition?.thumbnailUrl.trim() ?? '';
  if (expositionThumb.isNotEmpty) {
    return expositionThumb;
  }
  final expositionImage = node?.exposition?.imageUrl.trim() ?? '';
  if (expositionImage.isNotEmpty) {
    return expositionImage;
  }
  return null;
}

String? _pointOfInterestImageUrl(PointOfInterest poi) {
  final thumb = poi.thumbnailUrl?.trim() ?? '';
  if (thumb.isNotEmpty) return thumb;
  final image = poi.imageURL?.trim() ?? '';
  if (image.isNotEmpty) return image;
  return null;
}

bool _hasValue(String? value) => value != null && value.trim().isNotEmpty;

String _questEncounterType(QuestNode? node) {
  final encounterType =
      node?.objective?.encounterType.trim().toLowerCase() ?? '';
  switch (encounterType) {
    case 'boss':
      return 'boss';
    case 'raid':
      return 'raid';
    default:
      return 'monster';
  }
}

IconData _questChallengeIcon(QuestNode? node) {
  switch (_questEncounterType(node)) {
    case 'boss':
      return Icons.military_tech_outlined;
    case 'raid':
      return Icons.groups_outlined;
    default:
      return Icons.pets_outlined;
  }
}

({Color background, Color border, Color foreground}) _questChallengeBadgeColors(
  QuestNode? node,
) {
  switch (_questEncounterType(node)) {
    case 'boss':
      return (
        background: const Color(0xFFFFF0CC),
        border: const Color(0xFFE0A11B),
        foreground: const Color(0xFF6A4100),
      );
    case 'raid':
      return (
        background: const Color(0xFFFFE1DE),
        border: const Color(0xFFC53C2F),
        foreground: const Color(0xFF7A1712),
      );
    default:
      return (
        background: const Color(0xFFE2E8F0),
        border: const Color(0xFF94A3B8),
        foreground: const Color(0xFF334155),
      );
  }
}

bool _hasFetchCharacterLocation(QuestNode? node) {
  final character = node?.fetchCharacter;
  if (character == null) return false;
  if (_isValidCoordinate(
    character.pointOfInterestLat,
    character.pointOfInterestLng,
  )) {
    return true;
  }
  for (final location in character.locations) {
    if (_isValidCoordinate(location.latitude, location.longitude)) {
      return true;
    }
  }
  return false;
}

bool _isValidCoordinate(double? latitude, double? longitude) {
  if (latitude == null || longitude == null) return false;
  if (!latitude.isFinite || !longitude.isFinite) return false;
  return latitude.abs() <= 90 && longitude.abs() <= 180;
}

IconData _objectiveIconData(QuestNode? node) {
  if (_hasValue(node?.fetchCharacterId)) {
    return Icons.inventory_2_outlined;
  }
  if (_hasValue(node?.scenarioId)) {
    return Icons.auto_awesome_outlined;
  }
  if (_hasValue(node?.expositionId)) {
    return Icons.forum_outlined;
  }
  if (_hasValue(node?.monsterEncounterId) || _hasValue(node?.monsterId)) {
    return _questChallengeIcon(node);
  }
  if (_hasValue(node?.challengeId)) {
    return Icons.assignment_turned_in_outlined;
  }
  if (node?.pointOfInterest != null) {
    return Icons.place;
  }
  return Icons.explore_outlined;
}
