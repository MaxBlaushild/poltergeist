import 'package:flutter/material.dart';

import '../models/point_of_interest.dart';
import '../models/quest_node.dart';

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

  if (_hasValue(node.scenarioId)) {
    return const ['Complete the current scenario objective.'];
  }
  if (_hasValue(node.monsterEncounterId) || _hasValue(node.monsterId)) {
    return const ['Defeat the current monster objective.'];
  }
  if (_hasValue(node.challengeId)) {
    return const ['Complete the current challenge objective.'];
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

bool isQuestPointOfInterestHidden(
  QuestNode? node,
  Set<String> discoveredPoiIds,
) {
  final poi = node?.pointOfInterest;
  if (poi == null) return false;
  return !discoveredPoiIds.contains(poi.id);
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
  if (poi == null) return null;
  if (!discoveredPoiIds.contains(poi.id)) {
    return hiddenPoiPlaceholderImageUrl;
  }
  return _pointOfInterestImageUrl(poi);
}

String? _pointOfInterestImageUrl(PointOfInterest poi) {
  final thumb = poi.thumbnailUrl?.trim() ?? '';
  if (thumb.isNotEmpty) return thumb;
  final image = poi.imageURL?.trim() ?? '';
  if (image.isNotEmpty) return image;
  return null;
}

bool _hasValue(String? value) => value != null && value.trim().isNotEmpty;

IconData _objectiveIconData(QuestNode? node) {
  if (_hasValue(node?.scenarioId)) {
    return Icons.auto_awesome_outlined;
  }
  if (_hasValue(node?.monsterEncounterId) || _hasValue(node?.monsterId)) {
    return Icons.pets_outlined;
  }
  if (_hasValue(node?.challengeId)) {
    return Icons.assignment_turned_in_outlined;
  }
  if (node?.pointOfInterest != null) {
    return Icons.place;
  }
  return Icons.explore_outlined;
}
