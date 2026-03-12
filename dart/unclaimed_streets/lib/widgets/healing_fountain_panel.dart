import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/healing_fountain.dart';
import '../providers/location_provider.dart';
import '../services/poi_service.dart';
import 'paper_texture.dart';

const _fallbackFountainImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';
const _discoveredFountainImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/healing-fountain-discovered.png';
const _healingFountainCooldownDuration = Duration(days: 7);

class HealingFountainPanel extends StatefulWidget {
  const HealingFountainPanel({
    super.key,
    required this.fountain,
    required this.onClose,
    this.onUsed,
    this.onUnlocked,
  });

  final HealingFountain fountain;
  final VoidCallback onClose;
  final void Function(Map<String, dynamic> result)? onUsed;
  final Future<void> Function(HealingFountain fountain)? onUnlocked;

  @override
  State<HealingFountainPanel> createState() => _HealingFountainPanelState();
}

class _HealingFountainPanelState extends State<HealingFountainPanel> {
  bool _loading = false;
  bool _justUnlocked = false;
  String? _error;
  Timer? _cooldownTicker;
  late HealingFountain _fountain;

  @override
  void initState() {
    super.initState();
    _fountain = widget.fountain;
    _syncCooldownTicker();
  }

  @override
  void dispose() {
    _cooldownTicker?.cancel();
    super.dispose();
  }

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

  double _cooldownProgress(Duration remaining) {
    final totalSeconds = _healingFountainCooldownDuration.inSeconds;
    if (totalSeconds <= 0) return 1;
    final clampedRemaining = remaining.inSeconds.clamp(0, totalSeconds);
    return 1 - (clampedRemaining / totalSeconds);
  }

  String _formatReadyAt(BuildContext context, DateTime dateTime) {
    final localizations = MaterialLocalizations.of(context);
    final use24HourFormat =
        MediaQuery.maybeOf(context)?.alwaysUse24HourFormat ?? false;
    final date = localizations.formatMediumDate(dateTime);
    final time = localizations.formatTimeOfDay(
      TimeOfDay.fromDateTime(dateTime),
      alwaysUse24HourFormat: use24HourFormat,
    );
    return '$date at $time';
  }

  void _syncCooldownTicker() {
    final shouldTick =
        !_fountain.availableNow && _fountain.nextAvailableAt != null;
    if (!shouldTick) {
      _cooldownTicker?.cancel();
      _cooldownTicker = null;
      return;
    }
    _cooldownTicker ??= Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      final nextAvailableAt = _fountain.nextAvailableAt;
      if (nextAvailableAt == null || nextAvailableAt.isBefore(DateTime.now())) {
        _cooldownTicker?.cancel();
        _cooldownTicker = null;
      }
      setState(() {});
    });
  }

  Widget _buildCooldownCard(
    BuildContext context,
    Duration remaining,
    DateTime nextAvailableAt,
  ) {
    final theme = Theme.of(context);
    final progress = _cooldownProgress(remaining);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(18),
        color: theme.colorScheme.surfaceContainerHighest,
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Fountain Recharging',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 12),
          ClipRRect(
            borderRadius: BorderRadius.circular(999),
            child: LinearProgressIndicator(
              value: progress,
              minHeight: 10,
              backgroundColor: theme.colorScheme.surface,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            'Ready again ${_formatReadyAt(context, nextAvailableAt)}',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
  }

  DateTime? _parseDateTime(dynamic raw) {
    if (raw == null) return null;
    final text = raw.toString().trim();
    if (text.isEmpty) return null;
    return DateTime.tryParse(text)?.toLocal();
  }

  String _errorMessage(Object error, {String fallback = 'Request failed.'}) {
    if (error is DioException && error.response?.data is Map) {
      final data = Map<String, dynamic>.from(
        error.response!.data as Map<dynamic, dynamic>,
      );
      final rawMessage = data['error'] ?? data['message'];
      if (rawMessage != null) {
        final text = rawMessage.toString().trim();
        if (text.isNotEmpty) return text;
      }
    }
    return fallback;
  }

  bool _isAlreadyDiscoveredError(Object error) {
    final message = _errorMessage(error).toLowerCase();
    return message.contains('healing fountain') &&
        message.contains('discover') &&
        message.contains('already');
  }

  bool get _isDiscovered => _fountain.discovered || _justUnlocked;

  String _resolvedThumbnailUrl(HealingFountain fountain) {
    final raw = fountain.thumbnailUrl.trim();
    if (!fountain.discovered) {
      return _fallbackFountainImageUrl;
    }
    if (raw.isEmpty || raw == _fallbackFountainImageUrl) {
      return _discoveredFountainImageUrl;
    }
    return raw;
  }

  Future<void> _completeUnlock([Map<String, dynamic>? result]) async {
    final rawThumbnailUrl = result?['thumbnailUrl']?.toString().trim() ?? '';
    final thumbnailUrl = rawThumbnailUrl.isNotEmpty
        ? rawThumbnailUrl
        : _discoveredFountainImageUrl;
    final unlockedFountain = _fountain.copyWith(
      discovered: true,
      thumbnailUrl: thumbnailUrl,
    );
    await widget.onUnlocked?.call(unlockedFountain);
    if (!mounted) return;
    setState(() {
      _loading = false;
      _justUnlocked = true;
      _fountain = unlockedFountain;
    });
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(const SnackBar(content: Text('Discovered!')));
  }

  Future<void> _unlockFountain() async {
    if (_loading) return;
    final location = context.read<LocationProvider>().location;
    if (location == null) {
      setState(
        () => _error = 'Location not available. Enable location access.',
      );
      return;
    }
    final distance = _distanceMeters(
      location.latitude,
      location.longitude,
      _fountain.latitude,
      _fountain.longitude,
    );
    if (distance > kProximityUnlockRadiusMeters) {
      setState(
        () => _error =
            'Too far away (${distance.round()} m). Get within ${kProximityUnlockRadiusMeters.round()} m to unlock.',
      );
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await context.read<PoiService>().unlockHealingFountain(
        _fountain.id,
      );
      await _completeUnlock(result);
    } catch (error) {
      if (_isAlreadyDiscoveredError(error)) {
        await _completeUnlock();
        return;
      }
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = _errorMessage(
          error,
          fallback: 'Unable to discover this healing fountain right now.',
        );
      });
    }
  }

  Future<void> _useFountain() async {
    if (_loading) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await context.read<PoiService>().useHealingFountain(
        _fountain.id,
      );
      if (!mounted) return;
      setState(() => _loading = false);
      widget.onClose();
      widget.onUsed?.call(result);
    } catch (error) {
      if (!mounted) return;
      String message = _errorMessage(
        error,
        fallback: 'Unable to use healing fountain right now.',
      );
      if (error is DioException && error.response?.data is Map) {
        final data = Map<String, dynamic>.from(
          error.response!.data as Map<dynamic, dynamic>,
        );
        final nextAvailableAt = _parseDateTime(data['nextAvailableAt']);
        if (nextAvailableAt != null) {
          _fountain = _fountain.copyWith(
            availableNow: false,
            nextAvailableAt: nextAvailableAt,
            lastUsedAt:
                _parseDateTime(data['lastUsedAt']) ?? _fountain.lastUsedAt,
            cooldownSecondsRemaining:
                (data['cooldownSecondsRemaining'] as num?)?.toInt() ??
                _fountain.cooldownSecondsRemaining,
          );
          _syncCooldownTicker();
        }
      }
      setState(() {
        _loading = false;
        _error = message;
      });
    }
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
            _fountain.latitude,
            _fountain.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    if (!_isDiscovered) {
      return _buildUndiscovered(context, distance, withinRange);
    }
    final now = DateTime.now();
    final nextAvailableAt = _fountain.nextAvailableAt;
    final remaining = nextAvailableAt?.difference(now);
    final cooldownActive =
        !_fountain.availableNow && remaining != null && !remaining.isNegative;
    final buttonDisabled = _loading || !withinRange || cooldownActive;
    final buttonLabel = !withinRange
        ? 'Too far away'
        : (_loading ? 'Restoring...' : 'Restore Health & Mana');

    final thumbnail = _resolvedThumbnailUrl(_fountain);

    return DraggableScrollableSheet(
      initialChildSize: 0.86,
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
                    _fountain.name.isNotEmpty
                        ? _fountain.name
                        : 'Healing Fountain',
                    style: theme.textTheme.titleLarge?.copyWith(
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
                          thumbnail,
                          fit: BoxFit.cover,
                          errorBuilder: (_, _, _) => Container(
                            color: theme.colorScheme.surfaceContainerHighest,
                            child: const Icon(
                              Icons.water_drop_outlined,
                              size: 46,
                            ),
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
                        _InfoChip(
                          icon: Icons.refresh_outlined,
                          label: cooldownActive ? 'Weekly cooldown' : 'Ready',
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Text(
                      _fountain.description.trim().isNotEmpty
                          ? _fountain.description.trim()
                          : 'Touch the fountain to fully restore health and mana. Each healing fountain can be used once every 7 days.',
                      style: theme.textTheme.bodyMedium,
                    ),
                    if (cooldownActive && nextAvailableAt != null) ...[
                      const SizedBox(height: 14),
                      _buildCooldownCard(context, remaining, nextAvailableAt),
                    ],
                    if (_error != null) ...[
                      const SizedBox(height: 12),
                      Text(
                        _error!,
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.error,
                        ),
                      ),
                    ],
                    const SizedBox(height: 18),
                    FilledButton.icon(
                      onPressed: buttonDisabled ? null : _useFountain,
                      icon: const Icon(Icons.auto_fix_high),
                      label: Text(buttonLabel),
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

  Widget _buildUndiscovered(
    BuildContext context,
    double? distance,
    bool withinRange,
  ) {
    final theme = Theme.of(context);
    return DraggableScrollableSheet(
      initialChildSize: 0.86,
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
                  Row(
                    children: [
                      Icon(
                        Icons.lock_outline,
                        size: 28,
                        color: theme.colorScheme.primary,
                      ),
                      const SizedBox(width: 10),
                      Text(
                        'Undiscovered',
                        style: theme.textTheme.titleLarge?.copyWith(
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
                  Text(
                    'Visit this location to unlock this healing fountain. You must be within ${kProximityUnlockRadiusMeters.round()} meters to discover it.',
                    style: theme.textTheme.bodyLarge,
                  ),
                  const SizedBox(height: 16),
                  if (distance != null)
                    Text(
                      withinRange
                          ? 'Within range! Tap Unlock to discover.'
                          : 'You are ${distance.round()} m away.',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: withinRange
                            ? theme.colorScheme.primary
                            : theme.colorScheme.onSurface.withValues(
                                alpha: 0.7,
                              ),
                        fontWeight: withinRange ? FontWeight.w600 : null,
                      ),
                    )
                  else
                    Text(
                      'Enable location to see distance.',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurface.withValues(
                          alpha: 0.6,
                        ),
                      ),
                    ),
                  if (_error != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      _error!,
                      style: TextStyle(color: theme.colorScheme.error),
                    ),
                  ],
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: (_loading || !withinRange)
                        ? null
                        : _unlockFountain,
                    child: Text(
                      _loading
                          ? 'Unlocking...'
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
