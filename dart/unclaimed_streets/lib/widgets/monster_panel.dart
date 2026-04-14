import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/monster.dart';
import '../providers/location_provider.dart';
import '../utils/sticky_proximity_access.dart';
import 'paper_texture.dart';

class MonsterPanel extends StatefulWidget {
  const MonsterPanel({
    super.key,
    required this.encounter,
    required this.onClose,
    this.onFight,
  });

  final MonsterEncounter encounter;
  final VoidCallback onClose;
  final VoidCallback? onFight;

  @override
  State<MonsterPanel> createState() => _MonsterPanelState();
}

class _MonsterPanelState extends State<MonsterPanel> {
  final StickyProximityAccess _proximityAccess = StickyProximityAccess();

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
            widget.encounter.latitude,
            widget.encounter.longitude,
          );
    final liveWithinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    final hasProximityAccess = _proximityAccess.resolve(
      currentLocation: location,
      withinRange: liveWithinRange,
    );
    final mysteryState = !hasProximityAccess;
    final canFight = widget.onFight != null;
    final encounterTypeLabel = widget.encounter.encounterTypeLabel;
    return AdaptivePaperSheet(
      maxHeightFactor: mysteryState ? 0.62 : 0.95,
      header: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              child: Text(
                mysteryState
                    ? 'Mysterious $encounterTypeLabel'
                    : widget.encounter.name,
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
            IconButton(
              onPressed: widget.onClose,
              icon: const Icon(Icons.close),
            ),
          ],
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (!mysteryState) ...[
              ClipRRect(
                borderRadius: BorderRadius.circular(14),
                child: AspectRatio(
                  aspectRatio: 1,
                  child: Image.network(
                    _encounterImageUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) => Container(
                      color: theme.colorScheme.surfaceContainerHighest,
                      child: const Icon(Icons.pets, size: 42),
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 12),
            ],
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
                    icon: widget.encounter.isRaidEncounter
                        ? Icons.groups_2_outlined
                        : widget.encounter.isBossEncounter
                        ? Icons.workspace_premium_outlined
                        : Icons.pets_outlined,
                    label: encounterTypeLabel,
                  ),
                _InfoChip(
                  icon: Icons.shield_outlined,
                  label: 'Need ${kProximityUnlockRadiusMeters.round()} m',
                ),
                if (!mysteryState)
                  _InfoChip(
                    icon: Icons.stars,
                    label:
                        '${widget.encounter.monsters.length.clamp(1, 9)} monster${widget.encounter.monsters.length == 1 ? '' : 's'}',
                  ),
                if (!mysteryState && widget.encounter.isRaidEncounter)
                  const _InfoChip(
                    icon: Icons.groups_outlined,
                    label: 'Balanced for 5 players',
                  ),
                if (!mysteryState && widget.encounter.isBossEncounter)
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
            if (!mysteryState && widget.encounter.description.trim().isNotEmpty)
              Text(
                widget.encounter.description,
                style: theme.textTheme.bodyMedium,
              ),
            if (!mysteryState) ...[
              if (widget.encounter.description.trim().isNotEmpty)
                const SizedBox(height: 12),
              FilledButton.icon(
                onPressed: canFight ? widget.onFight : null,
                icon: const Icon(Icons.sports_martial_arts),
                label: const Text('Fight!'),
              ),
            ],
          ],
        ),
      ),
    );
  }

  String get _encounterImageUrl {
    final thumb = widget.encounter.thumbnailUrl.trim();
    if (thumb.isNotEmpty) return thumb;
    final image = widget.encounter.imageUrl.trim();
    if (image.isNotEmpty) return image;
    if (widget.encounter.monsters.isNotEmpty) {
      final monsterThumb = widget.encounter.monsters.first.thumbnailUrl.trim();
      if (monsterThumb.isNotEmpty) return monsterThumb;
      final monsterImage = widget.encounter.monsters.first.imageUrl.trim();
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
        color: theme.colorScheme.surfaceContainerHighest.withValues(
          alpha: 0.55,
        ),
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
