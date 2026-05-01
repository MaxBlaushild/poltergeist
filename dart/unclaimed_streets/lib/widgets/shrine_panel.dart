import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../models/shrine.dart';
import '../services/poi_service.dart';
import 'paper_texture.dart';

const _shrineFallbackImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/shrine-discovered.png';

class ShrinePanel extends StatefulWidget {
  const ShrinePanel({
    super.key,
    required this.shrine,
    required this.onClose,
    this.onUsed,
    this.onStatusChanged,
  });

  final Shrine shrine;
  final VoidCallback onClose;
  final void Function(Map<String, dynamic> result)? onUsed;
  final void Function(Shrine shrine)? onStatusChanged;

  @override
  State<ShrinePanel> createState() => _ShrinePanelState();
}

class _ShrinePanelState extends State<ShrinePanel>
    with SingleTickerProviderStateMixin {
  bool _loading = false;
  bool _statusRefreshing = false;
  String? _error;
  Timer? _cooldownTicker;
  late Shrine _shrine;
  late final AnimationController _shakeController;

  @override
  void initState() {
    super.initState();
    _shrine = widget.shrine;
    _shakeController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 420),
    );
    _statusRefreshing = _shrine.id.trim().isNotEmpty;
    _syncCooldownTicker();
    if (_statusRefreshing) {
      unawaited(_refreshShrineStatus());
    }
  }

  @override
  void dispose() {
    _cooldownTicker?.cancel();
    _shakeController.dispose();
    super.dispose();
  }

  double _cooldownProgress(Duration remaining) {
    final totalSeconds = _shrine.cooldownSeconds > 0
        ? _shrine.cooldownSeconds
        : const Duration(days: 7).inSeconds;
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
    final shouldTick = !_shrine.availableNow && _shrine.nextAvailableAt != null;
    if (!shouldTick) {
      _cooldownTicker?.cancel();
      _cooldownTicker = null;
      return;
    }
    _cooldownTicker ??= Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      final nextAvailableAt = _shrine.nextAvailableAt;
      if (nextAvailableAt == null || nextAvailableAt.isBefore(DateTime.now())) {
        _cooldownTicker?.cancel();
        _cooldownTicker = null;
      }
      setState(() {});
    });
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

  Future<void> _refreshShrineStatus() async {
    final latestShrine = await context.read<PoiService>().getShrineById(
      _shrine.id,
    );
    if (!mounted) return;

    if (latestShrine == null) {
      setState(() => _statusRefreshing = false);
      return;
    }

    setState(() {
      _shrine = latestShrine;
      _statusRefreshing = false;
    });
    _syncCooldownTicker();
    widget.onStatusChanged?.call(latestShrine);
  }

  Future<void> _playShrineHaptics() async {
    await HapticFeedback.heavyImpact();
    await Future<void>.delayed(const Duration(milliseconds: 80));
    await HapticFeedback.mediumImpact();
    await Future<void>.delayed(const Duration(milliseconds: 70));
    await HapticFeedback.heavyImpact();
  }

  Future<void> _playShrineShake() async {
    _shakeController
      ..stop()
      ..reset();
    await _shakeController.forward();
  }

  Future<void> _useShrine() async {
    if (_loading) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await context.read<PoiService>().useShrine(_shrine.id);
      if (!mounted) return;
      await _playShrineHaptics();
      await _playShrineShake();
      if (!mounted) return;
      setState(() => _loading = false);
      widget.onClose();
      widget.onUsed?.call(result);
    } catch (error) {
      if (!mounted) return;
      String message = _errorMessage(
        error,
        fallback: 'Unable to invoke this shrine right now.',
      );
      if (error is DioException && error.response?.data is Map) {
        final data = Map<String, dynamic>.from(
          error.response!.data as Map<dynamic, dynamic>,
        );
        final nextAvailableAt = DateTime.tryParse(
          data['nextAvailableAt']?.toString() ?? '',
        )?.toLocal();
        setState(() {
          _shrine = _shrine.copyWith(
            availableNow: data['availableNow'] as bool? ?? false,
            lastUsedAt: DateTime.tryParse(
              data['lastUsedAt']?.toString() ?? '',
            )?.toLocal(),
            nextAvailableAt: nextAvailableAt,
            cooldownSecondsRemaining:
                (data['cooldownSecondsRemaining'] as num?)?.toInt() ??
                _shrine.cooldownSecondsRemaining,
          );
        });
        if (message.toLowerCase().contains('cooldown') &&
            nextAvailableAt != null) {
          message =
              'This shrine is still recharging. It will be ready ${_formatReadyAt(context, nextAvailableAt)}.';
        }
        widget.onStatusChanged?.call(_shrine);
        _syncCooldownTicker();
      }
      setState(() {
        _loading = false;
        _error = message;
      });
    }
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
            'Shrine Recharging',
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

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final nextAvailableAt = _shrine.nextAvailableAt;
    final cooldownRemaining = Duration(
      seconds: math.max(0, _shrine.cooldownSecondsRemaining),
    );
    final thumbnailUrl = _shrine.mapMarkerUrl.trim().isNotEmpty
        ? _shrine.mapMarkerUrl.trim()
        : _shrineFallbackImageUrl;

    return PaperTexture(
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          child: AnimatedBuilder(
            animation: _shakeController,
            builder: (context, child) {
              final progress = _shakeController.value;
              final shakeOffset =
                  math.sin(progress * math.pi * 8) * (1 - progress) * 18;
              return Transform.translate(
                offset: Offset(shakeOffset, 0),
                child: child,
              );
            },
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Center(
                    child: Container(
                      width: 44,
                      height: 4,
                      decoration: BoxDecoration(
                        color: theme.colorScheme.outlineVariant,
                        borderRadius: BorderRadius.circular(999),
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(18),
                        child: Image.network(
                          thumbnailUrl,
                          width: 80,
                          height: 80,
                          fit: BoxFit.cover,
                          errorBuilder: (_, __, ___) => Container(
                            width: 80,
                            height: 80,
                            color: theme.colorScheme.surfaceContainerHighest,
                            alignment: Alignment.center,
                            child: const Icon(Icons.auto_awesome_rounded),
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _shrine.name.trim().isNotEmpty
                                  ? _shrine.name.trim()
                                  : 'Shrine',
                              style: theme.textTheme.headlineSmall?.copyWith(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 6),
                            Text(
                              _shrine.blessingName.trim().isNotEmpty
                                  ? _shrine.blessingName.trim()
                                  : 'Mystic Blessing',
                              style: theme.textTheme.titleMedium?.copyWith(
                                color: theme.colorScheme.primary,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  if (_shrine.description.trim().isNotEmpty)
                    Text(
                      _shrine.description.trim(),
                      style: theme.textTheme.bodyLarge?.copyWith(height: 1.5),
                    ),
                  if (_shrine.effectDescription.trim().isNotEmpty) ...[
                    const SizedBox(height: 18),
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(18),
                        color: const Color(0xFFF4EDFF),
                        border: Border.all(color: const Color(0xFFE2D0FF)),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Blessing Effect',
                            style: theme.textTheme.labelLarge?.copyWith(
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF6D28D9),
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _shrine.effectDescription.trim(),
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: const Color(0xFF4C1D95),
                              height: 1.45,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                  const SizedBox(height: 18),
                  if (!_shrine.availableNow && nextAvailableAt != null)
                    _buildCooldownCard(
                      context,
                      cooldownRemaining,
                      nextAvailableAt,
                    )
                  else
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(18),
                        color: const Color(0xFFEEF9F2),
                        border: Border.all(color: const Color(0xFFCBE9D4)),
                      ),
                      child: Text(
                        'Invoke this shrine to receive a day-long blessing scaled to your level.',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: const Color(0xFF166534),
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  if (_error != null) ...[
                    const SizedBox(height: 14),
                    Text(
                      _error!,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.error,
                      ),
                    ),
                  ],
                  const SizedBox(height: 20),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton(
                          onPressed: _loading ? null : widget.onClose,
                          child: const Text('Close'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: FilledButton.icon(
                          onPressed:
                              _loading ||
                                  _statusRefreshing ||
                                  !_shrine.availableNow
                              ? null
                              : _useShrine,
                          icon: _loading
                              ? const SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Icon(Icons.auto_awesome_rounded),
                          label: Text(
                            _shrine.availableNow
                                ? 'Invoke Shrine'
                                : 'Recharging',
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
