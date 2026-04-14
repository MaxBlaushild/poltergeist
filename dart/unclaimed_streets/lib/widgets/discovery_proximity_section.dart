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
    final statusText = hasDistance
        ? liveWithinRange
              ? 'Within range! Tap Unlock to discover.'
              : hasProximityAccess
              ? 'You can still unlock this. You are $roundedDistance m away now.'
              : 'You are $roundedDistance m away.'
        : locationUnavailableText;
    final detailText = !hasDistance || liveWithinRange
        ? null
        : hasProximityAccess
        ? 'You were already within ${unlockRadiusMeters.round()} m, so access is still active.'
        : 'Move within ${unlockRadiusMeters.round()} m to unlock this $subjectLabel.';
    final statusColor = !hasDistance
        ? theme.colorScheme.onSurface.withValues(alpha: 0.6)
        : (liveWithinRange || hasProximityAccess)
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurface.withValues(alpha: 0.7);

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
        const SizedBox(height: 16),
        Text(
          'Visit this location to unlock this $subjectLabel. You must be within ${unlockRadiusMeters.round()} meters to discover it.',
          style: theme.textTheme.bodyLarge,
        ),
        const SizedBox(height: 12),
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
          const SizedBox(height: 6),
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
