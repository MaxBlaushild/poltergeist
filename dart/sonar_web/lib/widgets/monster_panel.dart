import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/monster.dart';
import '../providers/location_provider.dart';
import 'paper_texture.dart';

const _monsterMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png';
const _legacyMysteryImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

class MonsterPanel extends StatelessWidget {
  const MonsterPanel({
    super.key,
    required this.monster,
    required this.onClose,
    this.onFight,
  });

  final Monster monster;
  final VoidCallback onClose;
  final VoidCallback? onFight;

  double _distanceMeters(double lat1, double lon1, double lat2, double lon2) {
    const earthRadiusMeters = 6371e3;
    final phi1 = lat1 * math.pi / 180;
    final phi2 = lat2 * math.pi / 180;
    final dPhi = (lat2 - lat1) * math.pi / 180;
    final dLambda = (lon2 - lon1) * math.pi / 180;
    final a =
        math.sin(dPhi / 2) * math.sin(dPhi / 2) +
        math.cos(phi1) *
            math.cos(phi2) *
            math.sin(dLambda / 2) *
            math.sin(dLambda / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return earthRadiusMeters * c;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final location = context.watch<LocationProvider>().location;
    final distance = location == null
        ? null
        : _distanceMeters(
            location.latitude,
            location.longitude,
            monster.latitude,
            monster.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    final mysteryState = !withinRange;
    final canFight = !mysteryState && onFight != null;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    mysteryState ? 'Mysterious Presence' : monster.name,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  IconButton(onPressed: onClose, icon: const Icon(Icons.close)),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    ClipRRect(
                      borderRadius: BorderRadius.circular(14),
                      child: AspectRatio(
                        aspectRatio: 1,
                        child: Image.network(
                          mysteryState
                              ? _monsterMysteryImageUrl
                              : (monster.thumbnailUrl.isNotEmpty
                                    ? monster.thumbnailUrl
                                    : monster.imageUrl),
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) =>
                              mysteryState
                              ? Image.network(
                                  _legacyMysteryImageUrl,
                                  fit: BoxFit.cover,
                                  errorBuilder: (context, error, stackTrace) =>
                                      Container(
                                        color: theme.colorScheme.surfaceVariant,
                                        child: const Icon(Icons.pets, size: 42),
                                      ),
                                )
                              : Container(
                                  color: theme.colorScheme.surfaceVariant,
                                  child: const Icon(Icons.pets, size: 42),
                                ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        if (distance != null)
                          _InfoChip(
                            icon: Icons.place_outlined,
                            label: '${distance.round()} m away',
                          ),
                        _InfoChip(
                          icon: Icons.shield_outlined,
                          label:
                              'Need ${kProximityUnlockRadiusMeters.round()} m',
                        ),
                        if (!mysteryState)
                          _InfoChip(
                            icon: Icons.stars,
                            label: 'Level ${monster.level}',
                          ),
                      ],
                    ),
                    const SizedBox(height: 14),
                    FilledButton.icon(
                      onPressed: canFight ? onFight : null,
                      icon: const Icon(Icons.sports_martial_arts),
                      label: Text(canFight ? 'Fight!' : 'Get closer to fight'),
                    ),
                    const SizedBox(height: 12),
                    if (!mysteryState && monster.description.trim().isNotEmpty)
                      Text(
                        monster.description,
                        style: theme.textTheme.bodyMedium,
                      ),
                    if (!mysteryState && monster.description.trim().isNotEmpty)
                      const SizedBox(height: 12),
                    if (mysteryState)
                      Text(
                        'The details of this monster are obscured until you are close enough to inspect it.',
                        style: theme.textTheme.bodyMedium,
                      ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _InfoChip extends StatelessWidget {
  const _InfoChip({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceVariant.withValues(alpha: 0.55),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14),
          const SizedBox(width: 6),
          Text(label, style: theme.textTheme.labelMedium),
        ],
      ),
    );
  }
}
