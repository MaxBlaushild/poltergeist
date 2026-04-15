import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/resource.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/location_provider.dart';
import '../providers/user_level_provider.dart';
import '../services/poi_service.dart';
import 'discovery_proximity_section.dart';
import 'paper_texture.dart';

class ResourcePanel extends StatefulWidget {
  const ResourcePanel({
    super.key,
    required this.resource,
    required this.onClose,
    this.onGathered,
  });

  final ResourceNode resource;
  final VoidCallback onClose;
  final void Function(Map<String, dynamic> rewardData)? onGathered;

  @override
  State<ResourcePanel> createState() => _ResourcePanelState();
}

class _ResourcePanelState extends State<ResourcePanel> {
  bool _loading = false;
  String? _error;

  double _distanceMeters(double lat1, double lon1, double lat2, double lon2) {
    const radius = 6371e3;
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
    return radius * c;
  }

  String _errorMessage(Object error) {
    if (error is DioException && error.response?.data is Map) {
      final data = error.response!.data as Map<dynamic, dynamic>;
      final message = data['error'] ?? data['message'];
      if (message != null && message.toString().trim().isNotEmpty) {
        return message.toString();
      }
    }
    return 'Unable to gather this resource right now.';
  }

  String _imageUrl() {
    final inventoryImage = widget.resource.inventoryItem?.imageUrl.trim() ?? '';
    if (inventoryImage.isNotEmpty) return inventoryImage;
    final icon = widget.resource.resourceType?.mapIconUrl.trim() ?? '';
    return icon;
  }

  Map<String, dynamic> _buildRewardModalData(Map<String, dynamic> response) {
    final rewardExperience =
        (response['rewardExperience'] as num?)?.toInt() ?? 0;
    final itemsAwarded =
        (response['itemsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((entry) => Map<String, dynamic>.from(entry))
            .toList() ??
        const <Map<String, dynamic>>[];
    final inventoryItem = widget.resource.inventoryItem;

    return {
      'resourceName': inventoryItem?.name ?? 'Resource',
      'rewardExperience': rewardExperience,
      'itemsAwarded': itemsAwarded.isNotEmpty
          ? itemsAwarded
          : [
              {
                'id': widget.resource.inventoryItemId,
                'name': inventoryItem?.name ?? 'Resource',
                'imageUrl': inventoryItem?.imageUrl ?? '',
                'quantity': widget.resource.quantity,
              },
            ],
    };
  }

  Future<void> _gather() async {
    if (_loading) return;
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final statsProvider = context.read<CharacterStatsProvider>();
      final previousLevel = statsProvider.level;
      final result = await context.read<PoiService>().gatherResource(
        widget.resource.id,
      );
      if (!mounted) return;
      final rewardData = _buildRewardModalData(result);
      await Future.wait([
        context.read<AuthProvider>().refresh(),
        statsProvider.refresh(silent: true),
        context.read<UserLevelProvider>().refresh(),
        context.read<ActivityFeedProvider>().refresh(),
      ]);
      if (!mounted) return;
      final modalData = {
        ...rewardData,
        'leveledUp': statsProvider.level > previousLevel,
        'previousLevel': previousLevel,
        'newLevel': statsProvider.level,
        'levelsGained': math.max(0, statsProvider.level - previousLevel),
      };
      setState(() => _loading = false);
      widget.onClose();
      widget.onGathered?.call(modalData);
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = _errorMessage(error);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final location = context.watch<LocationProvider>().location;
    final resourceTypeName = widget.resource.resourceType?.name.trim() ?? '';
    final distance = location == null
        ? null
        : _distanceMeters(
            location.latitude,
            location.longitude,
            widget.resource.latitude,
            widget.resource.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;

    if (!withinRange) {
      return AdaptivePaperSheet(
        maxHeightFactor: 0.45,
        header: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                widget.resource.inventoryItem?.name ?? 'Resource',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
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
          child: DiscoveryProximitySection(
            subjectLabel: 'resource',
            unlockRadiusMeters: kProximityUnlockRadiusMeters,
            distanceMeters: distance,
            hasProximityAccess: withinRange,
            liveWithinRange: withinRange,
            locationUnavailableText: 'Enable location to see distance.',
          ),
        ),
      );
    }

    return AdaptivePaperSheet(
      maxHeightFactor: 0.92,
      header: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              child: Text(
                widget.resource.inventoryItem?.name ?? 'Resource',
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
            ClipRRect(
              borderRadius: BorderRadius.circular(14),
              child: AspectRatio(
                aspectRatio: 1,
                child: _imageUrl().isEmpty
                    ? Container(
                        color: theme.colorScheme.surfaceContainerHighest,
                        alignment: Alignment.center,
                        child: const Icon(Icons.forest_outlined, size: 56),
                      )
                    : Image.network(
                        _imageUrl(),
                        fit: BoxFit.cover,
                        errorBuilder: (context, error, stackTrace) => Container(
                          color: theme.colorScheme.surfaceContainerHighest,
                          alignment: Alignment.center,
                          child: const Icon(Icons.forest_outlined, size: 56),
                        ),
                      ),
              ),
            ),
            const SizedBox(height: 16),
            if (resourceTypeName.isNotEmpty)
              Text(
                resourceTypeName,
                style: theme.textTheme.labelLarge?.copyWith(
                  color: theme.colorScheme.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
            const SizedBox(height: 8),
            Text(
              'Gather ${widget.resource.quantity}x ${widget.resource.inventoryItem?.name ?? 'resource'} from this node.',
              style: theme.textTheme.bodyLarge,
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                _ResourceMetaChip(
                  icon: Icons.inventory_2_outlined,
                  label: 'Quantity ${widget.resource.quantity}',
                ),
                _ResourceMetaChip(
                  icon: Icons.place_outlined,
                  label: '${distance.round()} m away',
                ),
              ],
            ),
            if (_error != null) ...[
              const SizedBox(height: 16),
              Text(
                _error!,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.error,
                ),
              ),
            ],
            const SizedBox(height: 20),
            FilledButton.icon(
              onPressed: _loading ? null : _gather,
              icon: const Icon(Icons.backpack_outlined),
              label: Text(_loading ? 'Gathering...' : 'Gather'),
            ),
          ],
        ),
      ),
    );
  }
}

class _ResourceMetaChip extends StatelessWidget {
  const _ResourceMetaChip({required this.icon, required this.label});

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
