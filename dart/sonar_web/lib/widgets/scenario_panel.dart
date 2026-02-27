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

class _ScenarioPanelState extends State<ScenarioPanel> {
  bool _loading = false;
  String? _error;
  String _responseText = '';
  bool _attemptedLocally = false;
  ScenarioPerformResult? _result;

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

  Future<void> _perform({String? optionId}) async {
    if (_loading || _attempted) return;

    if (widget.scenario.openEnded && _responseText.trim().isEmpty) {
      setState(() => _error = 'Write your response first.');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final result = await context.read<PoiService>().performScenario(
        widget.scenario.id,
        scenarioOptionId: optionId,
        responseText: widget.scenario.openEnded ? _responseText : null,
      );
      if (!mounted) return;
      setState(() {
        _result = result;
        _attemptedLocally = true;
        _loading = false;
      });
      widget.onPerformed?.call(result);
      if (optionId != null && optionId.isNotEmpty) {
        widget.onClose();
      }
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
                  Text(
                    mysteryState ? 'Mysterious Encounter' : 'Scenario',
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
                          mysteryState
                              ? _scenarioMysteryImageUrl
                              : (widget.scenario.thumbnailUrl.isNotEmpty
                                    ? widget.scenario.thumbnailUrl
                                    : widget.scenario.imageUrl),
                          fit: BoxFit.cover,
                          errorBuilder: (_, __, ___) => Container(
                            color: theme.colorScheme.surfaceVariant,
                            child: const Icon(Icons.auto_awesome_outlined),
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
                      if (_attempted)
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.surfaceVariant.withOpacity(
                              0.45,
                            ),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: theme.colorScheme.outline.withOpacity(0.2),
                            ),
                          ),
                          child: Text(
                            _result?.reason.isNotEmpty == true
                                ? _result!.reason
                                : 'You have already resolved this scenario.',
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
                          onPressed: _loading ? null : () => _perform(),
                          child: Text(
                            _loading ? 'Resolving…' : 'Resolve Scenario',
                          ),
                        ),
                      ] else ...[
                        for (final option in widget.scenario.options) ...[
                          FilledButton.tonal(
                            onPressed: _loading
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
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: _result!.successful
                              ? const Color(0xFFDEEED8)
                              : const Color(0xFFF2DFDD),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          'Roll ${_result!.roll} + stat ${_result!.statValue} + proficiency ${_result!.proficiencyBonus} + creativity ${_result!.creativityBonus} = ${_result!.totalScore} (need ${_result!.threshold})',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w700,
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
        color: theme.colorScheme.surfaceVariant.withOpacity(0.6),
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
