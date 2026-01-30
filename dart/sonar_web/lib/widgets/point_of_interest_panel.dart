import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';

import '../models/point_of_interest.dart';
import '../providers/auth_provider.dart';
import '../providers/location_provider.dart';
import '../services/poi_service.dart';

const _placeholderImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

/// Unlock radius in meters. Must match backend (POST /sonar/pointOfInterest/unlock).
const _unlockRadiusMeters = 200.0;

/// Bottom-sheet content for a tapped point of interest.
/// When [hasDiscovered] is false: tags, how to unlock, distance, and Unlock button (disabled when too far).
/// When true: full panel with name, image, description, tags, etc.
class PointOfInterestPanel extends StatefulWidget {
  const PointOfInterestPanel({
    super.key,
    required this.pointOfInterest,
    required this.hasDiscovered,
    required this.onClose,
    this.onUnlocked,
  });

  final PointOfInterest pointOfInterest;
  final bool hasDiscovered;
  final VoidCallback onClose;
  /// Called after successful unlock (e.g. refresh discoveries and POI markers). Optional.
  final Future<void> Function()? onUnlocked;

  @override
  State<PointOfInterestPanel> createState() => _PointOfInterestPanelState();
}

class _PointOfInterestPanelState extends State<PointOfInterestPanel> {
  bool _loading = false;
  bool _justUnlocked = false;
  String? _error;

  static String _formatTagName(String name) {
    final parts = name.split('_');
    return parts
        .asMap()
        .entries
        .map((e) {
          final s = e.value;
          if (s.isEmpty) return '';
          return e.key == 0
              ? (s[0].toUpperCase() + s.substring(1).toLowerCase())
              : s.toLowerCase();
        })
        .join(' ');
  }

  double _haversineMeters(double lat1, double lon1, double lat2, double lon2) {
    const R = 6371e3;
    final phi1 = lat1 * math.pi / 180;
    final phi2 = lat2 * math.pi / 180;
    final dPhi = (lat2 - lat1) * math.pi / 180;
    final dLambda = (lon2 - lon1) * math.pi / 180;
    final a = math.sin(dPhi / 2) * math.sin(dPhi / 2) +
        math.cos(phi1) * math.cos(phi2) *
            math.sin(dLambda / 2) * math.sin(dLambda / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return R * c;
  }

  String _errorMessage(Object e) {
    if (e is DioException && e.response?.data is Map) {
      final d = e.response!.data as Map<String, dynamic>;
      final msg = d['error'] ?? d['message'];
      if (msg != null && msg.toString().isNotEmpty) return msg.toString();
    }
    return e.toString();
  }

  Future<void> _handleUnlock() async {
    if (_loading) return;
    final loc = context.read<LocationProvider>().location;
    if (loc == null) {
      setState(() => _error = 'Location not available. Enable location access.');
      return;
    }
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) {
      setState(() => _error = 'Please log in to unlock.');
      return;
    }
    final poi = widget.pointOfInterest;
    final plat = double.tryParse(poi.lat) ?? 0.0;
    final plng = double.tryParse(poi.lng) ?? 0.0;
    final dist = _haversineMeters(loc.latitude, loc.longitude, plat, plng);
    if (dist > _unlockRadiusMeters) {
      setState(() => _error = 'Too far away (${dist.round()} m). Get within ${_unlockRadiusMeters.round()} m to unlock.');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await context.read<PoiService>().unlockPointOfInterest(
            poi.id,
            loc.latitude,
            loc.longitude,
            userId: userId,
          );
      if (!mounted) return;
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _loading = false;
        _error = null;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Discovered!')),
      );
    } catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = _errorMessage(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final poi = widget.pointOfInterest;
    final showFull = widget.hasDiscovered || _justUnlocked;
    if (!showFull) {
      return _buildUndiscovered(context, poi);
    }
    return _buildFullPanel(context, poi);
  }

  Widget _buildUndiscovered(BuildContext context, PointOfInterest poi) {
    final loc = context.watch<LocationProvider>().location;
    final plat = double.tryParse(poi.lat) ?? 0.0;
    final plng = double.tryParse(poi.lng) ?? 0.0;
    final distance = loc != null
        ? _haversineMeters(loc.latitude, loc.longitude, plat, plng)
        : null;
    final withinRange = distance != null && distance <= _unlockRadiusMeters;
    final tags = poi.tags;

    return DraggableScrollableSheet(
      initialChildSize: 0.7,
      minChildSize: 0.3,
      maxChildSize: 0.95,
      builder: (_, scrollController) => Container(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
        ),
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.lock_outline,
                        size: 28,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      const SizedBox(width: 10),
                      Text(
                        'Undiscovered',
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                    ],
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                children: [
                  // Tags
                  if (tags.isNotEmpty) ...[
                    Text(
                      'Tags',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.8),
                          ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 6,
                      children: tags
                          .map(
                            (t) => Chip(
                              label: Text(
                                _formatTagName(t.name),
                                style: const TextStyle(fontSize: 12),
                              ),
                              materialTapTargetSize:
                                  MaterialTapTargetSize.shrinkWrap,
                            ),
                          )
                          .toList(),
                    ),
                    const SizedBox(height: 20),
                  ],
                  // How to unlock
                  Text(
                    'Visit this location to unlock this point of interest. You must be within ${_unlockRadiusMeters.round()} meters to discover it.',
                    style: Theme.of(context).textTheme.bodyLarge,
                  ),
                  const SizedBox(height: 16),
                  // Distance
                  if (distance != null)
                    Text(
                      withinRange
                          ? 'Within range! Tap Unlock to discover.'
                          : 'You are ${distance.round()} m away.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: withinRange
                                ? Theme.of(context).colorScheme.primary
                                : Theme.of(context)
                                    .colorScheme
                                    .onSurface
                                    .withValues(alpha: 0.7),
                            fontWeight: withinRange ? FontWeight.w600 : null,
                          ),
                    )
                  else
                    Text(
                      'Enable location to see distance.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.6),
                          ),
                    ),
                  if (_error != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      _error!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                        fontSize: 14,
                      ),
                    ),
                  ],
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: (_loading || !withinRange) ? null : _handleUnlock,
                    child: Text(
                      _loading
                          ? 'Unlocking…'
                          : !withinRange
                              ? 'Too far to unlock'
                              : 'Unlock',
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFullPanel(BuildContext context, PointOfInterest poi) {
    final imageUrl = (poi.imageURL != null && poi.imageURL!.isNotEmpty)
        ? poi.imageURL!
        : _placeholderImageUrl;
    final tags = poi.tags;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.3,
      maxChildSize: 0.95,
      builder: (_, scrollController) => Container(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
        ),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          poi.name,
                          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                        if (poi.originalName != null &&
                            poi.originalName!.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          Text(
                            poi.originalName!,
                            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: Theme.of(context)
                                      .colorScheme
                                      .onSurface
                                      .withValues(alpha: 0.7),
                                ),
                          ),
                        ],
                        if (poi.googleMapsPlaceId != null &&
                            poi.googleMapsPlaceId!.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          GestureDetector(
                            onTap: () async {
                              final uri = Uri.parse(
                                'https://www.google.com/maps/place/?q=place_id:${poi.googleMapsPlaceId}',
                              );
                              try {
                                await launchUrl(uri, mode: LaunchMode.externalApplication);
                              } catch (_) {}
                            },
                            child: Text(
                              'View on Google Maps',
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.primary,
                                decoration: TextDecoration.underline,
                                fontSize: 14,
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(12),
                    child: Image.network(
                      imageUrl,
                      height: 200,
                      width: double.infinity,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        height: 200,
                        color: Colors.grey.shade300,
                        child: const Icon(Icons.image_not_supported, size: 48),
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (tags.isNotEmpty) ...[
                    Text(
                      'Tags',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.8),
                          ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 6,
                      children: tags
                          .map(
                            (t) => Chip(
                              label: Text(
                                _formatTagName(t.name),
                                style: const TextStyle(fontSize: 12),
                              ),
                              materialTapTargetSize:
                                  MaterialTapTargetSize.shrinkWrap,
                            ),
                          )
                          .toList(),
                    ),
                    const SizedBox(height: 16),
                  ],
                  if (poi.description != null && poi.description!.isNotEmpty) ...[
                    Text(
                      poi.description!,
                      style: Theme.of(context).textTheme.bodyLarge,
                    ),
                    const SizedBox(height: 12),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
