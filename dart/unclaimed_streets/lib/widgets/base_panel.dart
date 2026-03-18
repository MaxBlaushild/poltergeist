import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/base.dart';
import '../models/base_progression.dart';
import '../providers/auth_provider.dart';
import '../services/base_service.dart';
import 'paper_texture.dart';

class BasePanel extends StatefulWidget {
  const BasePanel({super.key, required this.base, required this.onClose});

  final BasePin base;
  final VoidCallback onClose;

  @override
  State<BasePanel> createState() => _BasePanelState();
}

class _BasePanelState extends State<BasePanel> {
  BaseProgressionSnapshot? _snapshot;
  List<BaseStructureDefinitionData> _catalog = const [];
  bool _loading = true;
  String? _error;
  String? _busyStructureKey;

  bool get _isOwner {
    final userId = context.read<AuthProvider>().user?.id ?? '';
    return userId.isNotEmpty && userId == widget.base.userId;
  }

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    if (!_isOwner) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = null;
        });
      }
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final service = context.read<BaseService>();
      final results = await Future.wait<dynamic>([
        service.getMyBase(),
        service.getCatalog(),
      ]);
      if (!mounted) return;
      setState(() {
        _snapshot = results[0] as BaseProgressionSnapshot;
        _catalog = (results[1] as List<BaseStructureDefinitionData>)
          ..sort((a, b) => a.sortOrder.compareTo(b.sortOrder));
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = e.toString();
      });
    }
  }

  Future<void> _mutateStructure(
    BaseStructureDefinitionData definition,
    bool isUpgrade,
  ) async {
    setState(() {
      _busyStructureKey = definition.key;
      _error = null;
    });
    try {
      final service = context.read<BaseService>();
      final nextSnapshot = isUpgrade
          ? await service.upgradeStructure(definition.key)
          : await service.buildStructure(definition.key);
      if (!mounted) return;
      setState(() {
        _snapshot = nextSnapshot;
        _busyStructureKey = null;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            isUpgrade
                ? '${definition.name} upgraded.'
                : '${definition.name} built.',
          ),
        ),
      );
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _busyStructureKey = null;
        _error = e.toString();
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(e.toString())));
    }
  }

  Map<String, int> get _resourceAmounts {
    final snapshot = _snapshot;
    if (snapshot == null) return const <String, int>{};
    final values = <String, int>{};
    for (final resource in snapshot.resources) {
      values[resource.resourceKey] = resource.amount;
    }
    return values;
  }

  Map<String, int> get _structureLevels {
    final snapshot = _snapshot;
    if (snapshot == null) return const <String, int>{};
    final values = <String, int>{};
    for (final structure in snapshot.structures) {
      values[structure.structureKey] = structure.level;
    }
    return values;
  }

  List<BaseStructureCostData> _costsForLevel(
    BaseStructureDefinitionData definition,
    int level,
  ) {
    return definition.levelCosts.where((cost) => cost.level == level).toList();
  }

  bool _hasMetPrerequisites(BaseStructureDefinitionData definition) {
    final required = definition.prereqConfig['requiredStructures'];
    if (required is! List) return true;
    final levels = _structureLevels;
    for (final entry in required) {
      if (entry is! Map) continue;
      final mapped = Map<String, dynamic>.from(entry);
      final key = mapped['key']?.toString() ?? '';
      final level = (mapped['level'] as num?)?.toInt() ?? 1;
      if (key.isEmpty) continue;
      if ((levels[key] ?? 0) < level) {
        return false;
      }
    }
    return true;
  }

  String _prerequisiteText(BaseStructureDefinitionData definition) {
    final required = definition.prereqConfig['requiredStructures'];
    if (required is! List || required.isEmpty) return '';
    final pieces = <String>[];
    for (final entry in required) {
      if (entry is! Map) continue;
      final mapped = Map<String, dynamic>.from(entry);
      final key = mapped['key']?.toString() ?? '';
      final level = (mapped['level'] as num?)?.toInt() ?? 1;
      if (key.isEmpty) continue;
      pieces.add('${_friendlyStructureName(key)} Lv $level');
    }
    return pieces.join(', ');
  }

  bool _canAfford(List<BaseStructureCostData> costs) {
    final amounts = _resourceAmounts;
    for (final cost in costs) {
      if ((amounts[cost.resourceKey] ?? 0) < cost.amount) {
        return false;
      }
    }
    return true;
  }

  String _friendlyResourceName(String key) {
    switch (key) {
      case 'arcane_dust':
        return 'Arcane Dust';
      case 'monster_parts':
        return 'Monster Parts';
      case 'relic_shards':
        return 'Relic Shards';
      default:
        final text = key.replaceAll('_', ' ');
        if (text.isEmpty) return key;
        return text
            .split(' ')
            .map((part) {
              if (part.isEmpty) return part;
              return '${part[0].toUpperCase()}${part.substring(1)}';
            })
            .join(' ');
    }
  }

  String _friendlyStructureName(String key) {
    for (final definition in _catalog) {
      if (definition.key == key && definition.name.trim().isNotEmpty) {
        return definition.name.trim();
      }
    }
    return _friendlyResourceName(key);
  }

  Widget _buildOwnerContent(BuildContext context) {
    final theme = Theme.of(context);
    if (_loading) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 28),
        child: Center(child: CircularProgressIndicator()),
      );
    }
    if (_error != null && _snapshot == null) {
      return Padding(
        padding: const EdgeInsets.only(top: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'We could not load your base details.',
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            Text(_error!, style: theme.textTheme.bodyMedium),
            const SizedBox(height: 12),
            OutlinedButton(onPressed: _load, child: const Text('Try again')),
          ],
        ),
      );
    }

    final snapshot = _snapshot;
    if (snapshot == null) {
      return const SizedBox.shrink();
    }

    final structures = _structureLevels;
    final resources = snapshot.resources.where((entry) => entry.amount > 0);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 16),
        Text(
          'Base Materials',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 10),
        if (resources.isEmpty)
          Text(
            'You have not collected any base materials yet.',
            style: theme.textTheme.bodyMedium,
          )
        else
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: resources
                .map(
                  (resource) => Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.surfaceContainerHighest
                          .withValues(alpha: 0.42),
                      borderRadius: BorderRadius.circular(999),
                      border: Border.all(
                        color: theme.colorScheme.outlineVariant,
                      ),
                    ),
                    child: Text(
                      '${_friendlyResourceName(resource.resourceKey)}: ${resource.amount}',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                )
                .toList(),
          ),
        const SizedBox(height: 20),
        Text(
          'Structures',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 10),
        ..._catalog.map((definition) {
          final currentLevel = structures[definition.key] ?? 0;
          final isBuilt = currentLevel > 0;
          final nextLevel = isBuilt ? currentLevel + 1 : 1;
          final isMaxed = currentLevel >= definition.maxLevel;
          final costs = _costsForLevel(definition, nextLevel);
          final hasPrereqs = _hasMetPrerequisites(definition);
          final canAfford = _canAfford(costs);
          final isBusy = _busyStructureKey == definition.key;
          final buttonEnabled =
              !isBusy &&
              !isMaxed &&
              hasPrereqs &&
              canAfford &&
              !_loading &&
              (isBuilt || definition.key != 'hearth' || currentLevel == 0);
          final actionLabel = isMaxed
              ? 'Max level'
              : isBuilt
              ? 'Upgrade to Lv $nextLevel'
              : 'Build';

          return Container(
            margin: const EdgeInsets.only(bottom: 12),
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest.withValues(
                alpha: 0.28,
              ),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: theme.colorScheme.outlineVariant),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            definition.name,
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            isBuilt
                                ? 'Level $currentLevel of ${definition.maxLevel}'
                                : 'Not built yet',
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surface,
                        borderRadius: BorderRadius.circular(999),
                        border: Border.all(
                          color: theme.colorScheme.outlineVariant,
                        ),
                      ),
                      child: Text(
                        definition.category.toUpperCase(),
                        style: theme.textTheme.labelSmall?.copyWith(
                          fontWeight: FontWeight.w700,
                          letterSpacing: 0.6,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                Text(
                  definition.description,
                  style: theme.textTheme.bodyMedium?.copyWith(height: 1.35),
                ),
                if (!hasPrereqs) ...[
                  const SizedBox(height: 10),
                  Text(
                    'Requires ${_prerequisiteText(definition)}',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.error,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
                if (costs.isNotEmpty) ...[
                  const SizedBox(height: 10),
                  Text(
                    'Cost',
                    style: theme.textTheme.labelLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: costs
                        .map(
                          (cost) => Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 10,
                              vertical: 6,
                            ),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surface,
                              borderRadius: BorderRadius.circular(999),
                              border: Border.all(
                                color: theme.colorScheme.outlineVariant,
                              ),
                            ),
                            child: Text(
                              '${_friendlyResourceName(cost.resourceKey)} ${cost.amount}',
                              style: theme.textTheme.bodySmall?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        )
                        .toList(),
                  ),
                ],
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    onPressed: buttonEnabled
                        ? () => _mutateStructure(definition, isBuilt)
                        : null,
                    child: isBusy
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(actionLabel),
                  ),
                ),
              ],
            ),
          );
        }),
        if (_error != null) ...[
          const SizedBox(height: 6),
          Text(
            _error!,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.error,
            ),
          ),
        ],
        if (snapshot.activeDailyEffects.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(
            'Active Base Effects',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          ...snapshot.activeDailyEffects.map(
            (effect) => Container(
              margin: const EdgeInsets.only(bottom: 8),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: theme.colorScheme.surface,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: theme.colorScheme.outlineVariant),
              ),
              child: Text(
                _friendlyResourceName(effect.stateKey),
                style: theme.textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ),
        ],
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final maxHeight = MediaQuery.sizeOf(context).height * 0.82;
    return PaperSheet(
      child: ConstrainedBox(
        constraints: BoxConstraints(maxHeight: maxHeight),
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Base',
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
              ClipRRect(
                borderRadius: BorderRadius.circular(14),
                child: AspectRatio(
                  aspectRatio: 1,
                  child: Image.network(
                    widget.base.thumbnailUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (_, _, _) => Container(
                      color: theme.colorScheme.surfaceContainerHighest,
                      child: const Icon(Icons.home_work_outlined, size: 48),
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest.withValues(
                    alpha: 0.42,
                  ),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      _isOwner
                          ? 'Your Base'
                          : '${widget.base.owner.displayName}\'s Base',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    if (widget.base.owner.secondaryName.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        widget.base.owner.secondaryName,
                        style: theme.textTheme.bodyMedium,
                      ),
                    ],
                    const SizedBox(height: 10),
                    Text(
                      _isOwner
                          ? 'Gather materials from the world, then shape your base into a place of recovery, study, and power.'
                          : 'A marked home base shared on the map with trusted allies.',
                      style: theme.textTheme.bodyMedium?.copyWith(height: 1.35),
                    ),
                  ],
                ),
              ),
              if (_isOwner) _buildOwnerContent(context),
            ],
          ),
        ),
      ),
    );
  }
}
