import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/monster.dart';
import '../providers/location_provider.dart';
import 'paper_texture.dart';

class MonsterPanel extends StatelessWidget {
  const MonsterPanel({super.key, required this.monster, required this.onClose});

  final Monster monster;
  final VoidCallback onClose;

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
                    monster.name,
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
                          monster.thumbnailUrl.isNotEmpty
                              ? monster.thumbnailUrl
                              : monster.imageUrl,
                          fit: BoxFit.cover,
                          errorBuilder: (_, _, _) => Container(
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
                          icon: Icons.stars,
                          label: 'Level ${monster.level}',
                        ),
                        _InfoChip(
                          icon: Icons.gavel,
                          label:
                              'Damage ${monster.attackDamageMin}-${monster.attackDamageMax}',
                        ),
                        _InfoChip(
                          icon: Icons.swipe,
                          label: 'Swipes ${monster.attackSwipesPerAttack}',
                        ),
                        _InfoChip(
                          icon: Icons.favorite,
                          label: 'HP ${monster.health}/${monster.maxHealth}',
                        ),
                        _InfoChip(
                          icon: Icons.auto_fix_high,
                          label: 'Mana ${monster.mana}/${monster.maxMana}',
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    if (monster.weaponInventoryItemName.isNotEmpty)
                      Text(
                        'Weapon: ${monster.weaponInventoryItemName}',
                        style: theme.textTheme.bodyMedium,
                      ),
                    if (monster.template?.name.isNotEmpty == true)
                      Text(
                        'Template: ${monster.template!.name}',
                        style: theme.textTheme.bodyMedium,
                      ),
                    Text(
                      'STR ${monster.strength} · DEX ${monster.dexterity} · CON ${monster.constitution} · INT ${monster.intelligence} · WIS ${monster.wisdom} · CHA ${monster.charisma}',
                      style: theme.textTheme.bodySmall,
                    ),
                    if (monster.description.trim().isNotEmpty) ...[
                      const SizedBox(height: 10),
                      Text(
                        monster.description.trim(),
                        style: theme.textTheme.bodyMedium,
                      ),
                    ],
                    const SizedBox(height: 12),
                    _SectionCard(
                      title: 'Rewards',
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Experience: ${monster.rewardExperience}',
                            style: theme.textTheme.bodyMedium,
                          ),
                          Text(
                            'Gold: ${monster.rewardGold}',
                            style: theme.textTheme.bodyMedium,
                          ),
                          if (monster.itemRewards.isNotEmpty) ...[
                            const SizedBox(height: 8),
                            Text(
                              'Items',
                              style: theme.textTheme.labelLarge?.copyWith(
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: 4),
                            for (final reward in monster.itemRewards)
                              Text(
                                '• ${reward.inventoryItemName.isNotEmpty ? reward.inventoryItemName : 'Item #${reward.inventoryItemId}'} x${reward.quantity}',
                                style: theme.textTheme.bodySmall,
                              ),
                          ],
                        ],
                      ),
                    ),
                    const SizedBox(height: 10),
                    _SectionCard(
                      title: 'Spells',
                      child: monster.spells.isEmpty
                          ? Text(
                              'No spells',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            )
                          : Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                for (final spell in monster.spells)
                                  Text(
                                    '• ${spell.name}',
                                    style: theme.textTheme.bodySmall,
                                  ),
                              ],
                            ),
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

class _SectionCard extends StatelessWidget {
  const _SectionCard({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceVariant.withValues(alpha: 0.38),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: theme.colorScheme.outline.withValues(alpha: 0.2),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: theme.textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 6),
          child,
        ],
      ),
    );
  }
}
