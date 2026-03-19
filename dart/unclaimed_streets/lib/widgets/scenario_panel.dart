import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/scenario.dart';
import '../providers/location_provider.dart';
import '../services/poi_service.dart';
import 'paper_texture.dart';

const _scenarioMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png';
const _legacyMysteryImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

class ScenarioPanel extends StatefulWidget {
  const ScenarioPanel({
    super.key,
    required this.scenario,
    required this.onClose,
    this.onPerformed,
  });

  final Scenario scenario;
  final VoidCallback onClose;
  final void Function(ScenarioPerformResult result)? onPerformed;

  @override
  State<ScenarioPanel> createState() => _ScenarioPanelState();
}

class _ScenarioPanelState extends State<ScenarioPanel>
    with SingleTickerProviderStateMixin {
  bool _loading = false;
  bool _rolling = false;
  bool _partySubmissionStatusLoading = true;
  bool _partySubmissionLocked = false;
  String? _partySubmissionStatus;
  String? _error;
  String _responseText = '';
  bool _attemptedLocally = false;
  ScenarioPerformResult? _result;
  late final AnimationController _diceController;
  late final Animation<double> _diceTilt;
  late final Animation<double> _dicePulse;
  Timer? _rollTicker;
  Timer? _partyStatusPollTicker;
  int _rollingValue = 1;
  final math.Random _rng = math.Random();

  @override
  void initState() {
    super.initState();
    _diceController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 520),
    );
    final curved = CurvedAnimation(
      parent: _diceController,
      curve: Curves.easeInOut,
    );
    _diceTilt = Tween<double>(begin: -0.09, end: 0.09).animate(curved);
    _dicePulse = Tween<double>(begin: 0.96, end: 1.06).animate(curved);
    unawaited(_refreshPartySubmissionStatus());
    _partyStatusPollTicker = Timer.periodic(const Duration(seconds: 3), (_) {
      unawaited(_refreshPartySubmissionStatus(silent: true));
    });
  }

  @override
  void dispose() {
    _rollTicker?.cancel();
    _partyStatusPollTicker?.cancel();
    _diceController.dispose();
    super.dispose();
  }

  bool get _attempted => widget.scenario.attemptedByUser || _attemptedLocally;

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

  String _errorMessage(Object error) {
    if (error is DioException && error.response?.data is Map) {
      final data = error.response!.data as Map<String, dynamic>;
      final msg = data['error'] ?? data['message'];
      if (msg != null && msg.toString().isNotEmpty) {
        return msg.toString();
      }
    }
    return 'Failed to perform scenario.';
  }

  Future<void> _refreshPartySubmissionStatus({bool silent = false}) async {
    if (!mounted) return;
    if (!silent) {
      setState(() => _partySubmissionStatusLoading = true);
    }
    try {
      final status = await context.read<PoiService>().getPartySubmissionStatus(
        contentType: 'scenario',
        contentId: widget.scenario.id,
      );
      if (!mounted) return;
      setState(() {
        _partySubmissionLocked = status.locked;
        _partySubmissionStatus = status.status;
        _partySubmissionStatusLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _partySubmissionStatusLoading = false;
      });
    }
  }

  void _startRollAnimation() {
    _rollTicker?.cancel();
    _diceController.repeat(reverse: true);
    setState(() {
      _rolling = true;
      _rollingValue = (_rng.nextInt(20) + 1);
    });
    _rollTicker = Timer.periodic(const Duration(milliseconds: 90), (_) {
      if (!mounted) return;
      setState(() {
        _rollingValue = (_rng.nextInt(20) + 1);
      });
    });
  }

  void _stopRollAnimation({int? finalRoll}) {
    _rollTicker?.cancel();
    _rollTicker = null;
    _diceController.stop();
    _diceController.value = 0;
    if (!mounted) return;
    setState(() {
      _rolling = false;
      if (finalRoll != null && finalRoll > 0) {
        _rollingValue = finalRoll;
      }
    });
  }

  Future<void> _perform({String? optionId}) async {
    if (_loading || _attempted) return;
    if (_partySubmissionLocked) {
      setState(() {
        _error = _partySubmissionStatus?.toLowerCase() == 'completed'
            ? 'A party member already resolved this scenario.'
            : 'A party member is already submitting this scenario.';
      });
      return;
    }

    if (widget.scenario.openEnded && _responseText.trim().isEmpty) {
      setState(() => _error = 'Write your response first.');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });
    _startRollAnimation();

    try {
      await Future<void>.delayed(const Duration(milliseconds: 850));
      if (!mounted) return;
      final result = await context.read<PoiService>().performScenario(
        widget.scenario.id,
        scenarioOptionId: optionId,
        responseText: widget.scenario.openEnded ? _responseText : null,
      );
      if (!mounted) return;
      _stopRollAnimation(finalRoll: result.roll);
      setState(() {
        _result = result;
        _attemptedLocally = true;
        _loading = false;
      });
      if (mounted) {
        Navigator.of(context).maybePop();
      }
      widget.onPerformed?.call(result);
    } catch (error) {
      if (!mounted) return;
      _stopRollAnimation();
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
    final distance = location == null
        ? null
        : _distanceMeters(
            location.latitude,
            location.longitude,
            widget.scenario.latitude,
            widget.scenario.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    final mysteryState = !withinRange;

    return DraggableScrollableSheet(
      initialChildSize: 0.88,
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
                  if (mysteryState)
                    Text(
                      'Mysterious Scenario',
                      style: theme.textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    )
                  else
                    const SizedBox.shrink(),
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
                          mysteryState
                              ? _scenarioMysteryImageUrl
                              : (widget.scenario.thumbnailUrl.isNotEmpty
                                    ? widget.scenario.thumbnailUrl
                                    : widget.scenario.imageUrl),
                          fit: BoxFit.cover,
                          errorBuilder: (_, _, _) => mysteryState
                              ? Image.network(
                                  _legacyMysteryImageUrl,
                                  fit: BoxFit.cover,
                                  errorBuilder: (_, _, _) => Container(
                                    color: theme
                                        .colorScheme
                                        .surfaceContainerHighest,
                                    child: const Icon(
                                      Icons.auto_awesome_outlined,
                                    ),
                                  ),
                                )
                              : Container(
                                  color:
                                      theme.colorScheme.surfaceContainerHighest,
                                  child: const Icon(
                                    Icons.auto_awesome_outlined,
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
                          _Chip(
                            icon: Icons.place_outlined,
                            label: '${distance.round()} m away',
                          ),
                        _Chip(
                          icon: Icons.shield_outlined,
                          label:
                              'Need ${kProximityUnlockRadiusMeters.round()} m',
                        ),
                      ],
                    ),
                    const SizedBox(height: 14),
                    if (mysteryState)
                      Text(
                        'This scenario remains a mystery until you are close enough to investigate.',
                        style: theme.textTheme.bodyMedium,
                      )
                    else ...[
                      Text(
                        widget.scenario.prompt,
                        style: theme.textTheme.bodyLarge,
                      ),
                      const SizedBox(height: 16),
                      AnimatedSwitcher(
                        duration: const Duration(milliseconds: 200),
                        switchInCurve: Curves.easeOut,
                        switchOutCurve: Curves.easeIn,
                        child: _rolling
                            ? Container(
                                key: const ValueKey('scenario-roll-animation'),
                                margin: const EdgeInsets.only(bottom: 16),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 10,
                                ),
                                decoration: BoxDecoration(
                                  color: theme
                                      .colorScheme
                                      .surfaceContainerHighest
                                      .withValues(alpha: 0.45),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: theme.colorScheme.outline.withValues(
                                      alpha: 0.24,
                                    ),
                                  ),
                                ),
                                child: Row(
                                  children: [
                                    AnimatedBuilder(
                                      animation: _diceController,
                                      builder: (context, child) {
                                        return Transform.rotate(
                                          angle: _diceTilt.value,
                                          child: Transform.scale(
                                            scale: _dicePulse.value,
                                            child: child,
                                          ),
                                        );
                                      },
                                      child: Icon(
                                        Icons.casino_rounded,
                                        color: theme.colorScheme.primary,
                                      ),
                                    ),
                                    const SizedBox(width: 10),
                                    Expanded(
                                      child: Text(
                                        'Rolling fate… d20: $_rollingValue',
                                        style: theme.textTheme.bodyMedium
                                            ?.copyWith(
                                              fontWeight: FontWeight.w600,
                                            ),
                                      ),
                                    ),
                                  ],
                                ),
                              )
                            : const SizedBox.shrink(
                                key: ValueKey('scenario-roll-animation-hidden'),
                              ),
                      ),
                      if (_attempted)
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.surfaceContainerHighest
                                .withValues(alpha: 0.45),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: theme.colorScheme.outline.withValues(
                                alpha: 0.2,
                              ),
                            ),
                          ),
                          child: Text(
                            _result?.reason.isNotEmpty == true
                                ? _result!.reason
                                : 'You have already resolved this scenario.',
                            style: theme.textTheme.bodyMedium,
                          ),
                        )
                      else if (_partySubmissionLocked)
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.surfaceContainerHighest
                                .withValues(alpha: 0.45),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: theme.colorScheme.outline.withValues(
                                alpha: 0.2,
                              ),
                            ),
                          ),
                          child: Text(
                            (_partySubmissionStatus ?? '').toLowerCase() ==
                                    'completed'
                                ? 'A party member already resolved this scenario.'
                                : 'A party member is submitting this scenario now.',
                            style: theme.textTheme.bodyMedium,
                          ),
                        )
                      else if (widget.scenario.openEnded) ...[
                        TextField(
                          minLines: 3,
                          maxLines: 6,
                          onChanged: (value) => _responseText = value,
                          decoration: const InputDecoration(
                            labelText: 'Your response',
                            hintText:
                                'Describe exactly how your character responds…',
                            border: OutlineInputBorder(),
                          ),
                        ),
                        const SizedBox(height: 12),
                        FilledButton(
                          onPressed: (_loading || _partySubmissionStatusLoading)
                              ? null
                              : () => _perform(),
                          child: Text(
                            _loading ? 'Resolving…' : 'Resolve Scenario',
                          ),
                        ),
                      ] else ...[
                        for (final option in widget.scenario.options) ...[
                          FilledButton.tonal(
                            onPressed:
                                (_loading || _partySubmissionStatusLoading)
                                ? null
                                : () => _perform(optionId: option.id),
                            child: Align(
                              alignment: Alignment.centerLeft,
                              child: Text(option.optionText),
                            ),
                          ),
                          const SizedBox(height: 8),
                        ],
                      ],
                    ],
                    if (_error != null) ...[
                      const SizedBox(height: 12),
                      Text(
                        _error!,
                        style: TextStyle(
                          color: theme.colorScheme.error,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                    if (_result != null) ...[
                      const SizedBox(height: 12),
                      TweenAnimationBuilder<double>(
                        tween: Tween<double>(begin: 0, end: 1),
                        duration: const Duration(milliseconds: 320),
                        curve: Curves.easeOutCubic,
                        builder: (context, value, child) {
                          return Opacity(
                            opacity: value,
                            child: Transform.translate(
                              offset: Offset(0, 10 * (1 - value)),
                              child: child,
                            ),
                          );
                        },
                        child: Container(
                          padding: const EdgeInsets.all(14),
                          decoration: BoxDecoration(
                            color: _result!.successful
                                ? const Color(0xFFDEEED8)
                                : const Color(0xFFF2DFDD),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color:
                                  (_result!.successful
                                          ? Colors.green.shade200
                                          : Colors.red.shade200)
                                      .withValues(alpha: 0.8),
                            ),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Icon(
                                    _result!.successful
                                        ? Icons.emoji_events_rounded
                                        : Icons.analytics_outlined,
                                    color: _result!.successful
                                        ? Colors.green.shade800
                                        : Colors.red.shade800,
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: Text(
                                      _result!.successful
                                          ? 'You cleared the check by ${(_result!.totalScore - _result!.threshold).abs()} points.'
                                          : 'You were ${(_result!.threshold - _result!.totalScore).abs()} points short.',
                                      style: theme.textTheme.bodyMedium
                                          ?.copyWith(
                                            fontWeight: FontWeight.w700,
                                          ),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 10),
                              Text(
                                'Score breakdown',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  fontWeight: FontWeight.w700,
                                  color: theme.colorScheme.onSurfaceVariant,
                                ),
                              ),
                              const SizedBox(height: 8),
                              Wrap(
                                spacing: 8,
                                runSpacing: 8,
                                children: [
                                  _ScoreBreakdownChip(
                                    icon: Icons.casino_rounded,
                                    label: 'Roll',
                                    value: _result!.roll,
                                  ),
                                  _ScoreBreakdownChip(
                                    icon: Icons.fitness_center_rounded,
                                    label: _formatStatLabel(_result!.statTag),
                                    value: _result!.statValue,
                                  ),
                                  _ScoreBreakdownChip(
                                    icon: Icons.workspace_premium_rounded,
                                    label: 'Proficiency',
                                    value: _result!.proficiencyBonus,
                                  ),
                                  if (_result!.responseScore > 0)
                                    _ScoreBreakdownChip(
                                      icon: Icons.psychology_alt_rounded,
                                      label: 'Response',
                                      value: _result!.responseScore,
                                    ),
                                  _ScoreBreakdownChip(
                                    icon: Icons.lightbulb_rounded,
                                    label: 'Creativity',
                                    value: _result!.creativityBonus,
                                  ),
                                ],
                              ),
                              const SizedBox(height: 10),
                              Text(
                                'Total ${_result!.totalScore} vs target ${_result!.threshold}',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                _result!.responseScore > 0
                                    ? 'Your score combines the roll, the chosen stat, training, the AI response score, and any creativity bonus.'
                                    : 'Your score combines the roll, the chosen stat, training, and any creativity bonus.',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: theme.colorScheme.onSurfaceVariant,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],
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

class _Chip extends StatelessWidget {
  const _Chip({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest.withValues(alpha: 0.6),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16),
          const SizedBox(width: 6),
          Text(label, style: theme.textTheme.bodySmall),
        ],
      ),
    );
  }
}

class _ScoreBreakdownChip extends StatelessWidget {
  const _ScoreBreakdownChip({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final int value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface.withValues(alpha: 0.55),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: theme.colorScheme.outline.withValues(alpha: 0.14),
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: theme.colorScheme.primary),
          const SizedBox(width: 6),
          Text(
            '$label ${value > 0 ? '+' : ''}$value',
            style: theme.textTheme.bodySmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

String _formatStatLabel(String statTag) {
  if (statTag.isEmpty) return 'Stat';
  return '${statTag[0].toUpperCase()}${statTag.substring(1)}';
}
