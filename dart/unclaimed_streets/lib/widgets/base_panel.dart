import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/base.dart';
import '../providers/location_provider.dart';
import '../screens/base_management_screen.dart';
import 'paper_texture.dart';

const _fallbackBaseImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/base-discovered.png';

class BasePanel extends StatelessWidget {
  const BasePanel({super.key, required this.base, required this.onClose});

  final BasePin base;
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

  String get _baseTitle {
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

  String get _baseImageUrl {
    final thumbnail = base.thumbnailUrl.trim();
    if (thumbnail.isNotEmpty) return thumbnail;
    final image = base.imageUrl.trim();
    if (image.isNotEmpty) return image;
    return _fallbackBaseImageUrl;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final maxHeight = MediaQuery.sizeOf(context).height * 0.82;
    final location = context.watch<LocationProvider>().location;
    final distance = location == null
        ? null
        : _distanceMeters(
            location.latitude,
            location.longitude,
            base.latitude,
            base.longitude,
          );
    final canEnterBase =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    final description = base.description.trim();

    return PaperSheet(
      child: ConstrainedBox(
        constraints: BoxConstraints(maxHeight: maxHeight),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _baseTitle,
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
                  IconButton(onPressed: onClose, icon: const Icon(Icons.close)),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
              child: Wrap(
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
            ),
            Expanded(
              child: canEnterBase
                  ? BaseManagementContent(
                      baseId: base.id,
                      padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
                    )
                  : SingleChildScrollView(
                      padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          ClipRRect(
                            borderRadius: BorderRadius.circular(14),
                            child: AspectRatio(
                              aspectRatio: 1,
                              child: Image.network(
                                _baseImageUrl,
                                fit: BoxFit.cover,
                                errorBuilder: (_, _, _) => Container(
                                  color:
                                      theme.colorScheme.surfaceContainerHighest,
                                  child: const Icon(
                                    Icons.home_work_outlined,
                                    size: 46,
                                  ),
                                ),
                              ),
                            ),
                          ),
                          if (description.isNotEmpty) ...[
                            const SizedBox(height: 14),
                            Text(
                              description,
                              style: theme.textTheme.bodyMedium,
                            ),
                          ],
                          const SizedBox(height: 16),
                          Text(
                            location == null
                                ? 'Enable location to see distance and enter this base.'
                                : 'Move within ${kProximityUnlockRadiusMeters.round()} m to enter this base.',
                            style: theme.textTheme.bodyMedium,
                          ),
                          if (distance != null)
                            Padding(
                              padding: const EdgeInsets.only(top: 6),
                              child: Text(
                                'You are ${distance.round()} m away.',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: theme.colorScheme.onSurface.withValues(
                                    alpha: 0.72,
                                  ),
                                ),
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
