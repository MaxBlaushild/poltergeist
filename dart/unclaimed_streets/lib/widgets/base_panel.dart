import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/base.dart';
import '../providers/location_provider.dart';
import 'paper_texture.dart';

class BasePanel extends StatelessWidget {
  const BasePanel({
    super.key,
    required this.base,
    required this.onClose,
    this.onOpenBaseManagement,
  });

  final BasePin base;
  final VoidCallback onClose;
  final VoidCallback? onOpenBaseManagement;

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

  String _baseTitle(BasePin base) {
    final explicitName = base.name.trim();
    if (explicitName.isNotEmpty) {
      return explicitName;
    }
    final owner = base.owner;
    final preferredName = owner.username.trim().isNotEmpty
        ? owner.username.trim()
        : owner.name.trim().isNotEmpty
        ? owner.name.trim()
        : owner.displayName.replaceFirst('@', '').trim();
    if (preferredName.isEmpty) {
      return 'Base';
    }
    return "$preferredName's Base";
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
            base.latitude,
            base.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    return AdaptivePaperSheet(
      maxHeightFactor: 0.42,
      header: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 8, 0),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Padding(
                        padding: const EdgeInsets.only(top: 2),
                        child: Icon(
                          Icons.home_work_outlined,
                          color: theme.colorScheme.primary,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _baseTitle(base),
                              style: theme.textTheme.titleLarge?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              base.owner.displayName,
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.onSurface.withValues(
                                  alpha: 0.7,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(onPressed: onClose, icon: const Icon(Icons.close)),
              ],
            ),
          ],
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
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
                  label: 'Need ${kProximityUnlockRadiusMeters.round()} m',
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              location == null
                  ? 'Enable location to see distance and enter this base.'
                  : withinRange
                  ? 'You are close enough to enter this base.'
                  : 'You are too far away to enter this base right now.',
              style: theme.textTheme.bodyMedium,
            ),
            if (location != null && !withinRange) ...[
              const SizedBox(height: 6),
              Text(
                'Move within ${kProximityUnlockRadiusMeters.round()} m of the base pin to open base management.',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ],
            if (withinRange && onOpenBaseManagement != null) ...[
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: onOpenBaseManagement,
                  child: const Text('Open Base Management'),
                ),
              ),
            ],
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
