import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/monster.dart';
import '../providers/location_provider.dart';
import 'paper_texture.dart';

const _monsterMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png';
const _bossMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/boss-undiscovered.png';
const _raidMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/raid-undiscovered.png';
const _legacyMysteryImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

class MonsterPanel extends StatelessWidget {
  const MonsterPanel({
    super.key,
    required this.encounter,
    required this.onClose,
    this.onFight,
  });

  final MonsterEncounter encounter;
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

  String _mysteryImageUrlForEncounter(MonsterEncounter encounter) {
    if (encounter.isBossEncounter) return _bossMysteryImageUrl;
    if (encounter.isRaidEncounter) return _raidMysteryImageUrl;
    return _monsterMysteryImageUrl;
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
            encounter.latitude,
            encounter.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    final mysteryState = !withinRange;
    final canFight = onFight != null;
    final encounterTypeLabel = encounter.encounterTypeLabel;

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
                    mysteryState
                        ? 'Mysterious $encounterTypeLabel'
                        : encounter.name,
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
                              ? _mysteryImageUrlForEncounter(encounter)
                              : _encounterImageUrl,
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
                        if (!mysteryState)
                          _InfoChip(
                            icon: encounter.isRaidEncounter
                                ? Icons.groups_2_outlined
                                : encounter.isBossEncounter
                                ? Icons.workspace_premium_outlined
                                : Icons.pets_outlined,
                            label: encounterTypeLabel,
                          ),
                        _InfoChip(
                          icon: Icons.shield_outlined,
                          label:
                              'Need ${kProximityUnlockRadiusMeters.round()} m',
                        ),
                        if (!mysteryState)
                          _InfoChip(
                            icon: Icons.stars,
                            label:
                                '${encounter.monsters.length.clamp(1, 9)} monster${encounter.monsters.length == 1 ? '' : 's'}',
                          ),
                        if (!mysteryState && encounter.isRaidEncounter)
                          const _InfoChip(
                            icon: Icons.groups_outlined,
                            label: 'Balanced for 5 players',
                          ),
                        if (!mysteryState && encounter.isBossEncounter)
                          const _InfoChip(
                            icon: Icons.trending_up,
                            label: 'Scaled +5 levels',
                          ),
                      ],
                    ),
                    const SizedBox(height: 14),
                    if (mysteryState)
                      Text(
                        'This encounter remains a mystery until you are close enough to investigate.',
                        style: theme.textTheme.bodyMedium,
                      ),
                    if (!mysteryState &&
                        encounter.description.trim().isNotEmpty)
                      Text(
                        encounter.description,
                        style: theme.textTheme.bodyMedium,
                      ),
                    if (!mysteryState) ...[
                      if (encounter.description.trim().isNotEmpty)
                        const SizedBox(height: 12),
                      FilledButton.icon(
                        onPressed: canFight ? onFight : null,
                        icon: const Icon(Icons.sports_martial_arts),
                        label: const Text('Fight!'),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  String get _encounterImageUrl {
    final thumb = encounter.thumbnailUrl.trim();
    if (thumb.isNotEmpty) return thumb;
    final image = encounter.imageUrl.trim();
    if (image.isNotEmpty) return image;
    if (encounter.monsters.isNotEmpty) {
      final monsterThumb = encounter.monsters.first.thumbnailUrl.trim();
      if (monsterThumb.isNotEmpty) return monsterThumb;
      final monsterImage = encounter.monsters.first.imageUrl.trim();
      if (monsterImage.isNotEmpty) return monsterImage;
    }
    return '';
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
