import 'package:flutter/material.dart';

class DiscoveryProximitySection extends StatelessWidget {
  const DiscoveryProximitySection({
    super.key,
    required this.subjectLabel,
    required this.unlockRadiusMeters,
    required this.distanceMeters,
    required this.hasProximityAccess,
    required this.liveWithinRange,
    this.locationUnavailableText = 'Enable location to see distance.',
  });

  final String subjectLabel;
  final double unlockRadiusMeters;
  final double? distanceMeters;
  final bool hasProximityAccess;
  final bool liveWithinRange;
  final String locationUnavailableText;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasDistance = distanceMeters != null;
    final roundedDistance = distanceMeters?.round();
    final statusText = !hasDistance
        ? locationUnavailableText
        : liveWithinRange
        ? 'Within range! Tap Unlock to discover this $subjectLabel.'
        : hasProximityAccess
        ? 'You can still unlock this $subjectLabel.'
        : null;
    final detailText = !hasDistance || liveWithinRange
        ? null
        : hasProximityAccess
        ? 'Access stays active after your first in-range visit.'
        : null;
    final statusColor = !hasDistance
        ? theme.colorScheme.onSurface.withValues(alpha: 0.6)
        : (liveWithinRange || hasProximityAccess)
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurface.withValues(alpha: 0.7);
    final hasSupportingCopy = statusText != null || detailText != null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            if (hasDistance)
              _DiscoveryProximityChip(
                icon: Icons.place_outlined,
                label: '$roundedDistance m away',
              ),
            _DiscoveryProximityChip(
              icon: Icons.shield_outlined,
              label: 'Need ${unlockRadiusMeters.round()} m',
            ),
          ],
        ),
        if (hasSupportingCopy) ...[const SizedBox(height: 12)],
        if (statusText != null)
          Text(
            statusText,
            style: theme.textTheme.bodyMedium?.copyWith(
              color: statusColor,
              fontWeight: liveWithinRange || hasProximityAccess
                  ? FontWeight.w600
                  : null,
            ),
          ),
        if (detailText != null) ...[
          if (statusText != null) const SizedBox(height: 6),
          Text(
            detailText,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ],
    );
  }
}

class _DiscoveryProximityChip extends StatelessWidget {
  const _DiscoveryProximityChip({required this.icon, required this.label});

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
        border: Border.all(
          color: theme.colorScheme.outline.withValues(alpha: 0.2),
        ),
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
