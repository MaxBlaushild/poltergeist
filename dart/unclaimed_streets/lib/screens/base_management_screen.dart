import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/base_progression.dart';
import '../services/base_service.dart';
import '../widgets/paper_texture.dart';

class BaseManagementScreen extends StatefulWidget {
  const BaseManagementScreen({super.key, required this.baseId});

  final String baseId;

  @override
  State<BaseManagementScreen> createState() => _BaseManagementScreenState();
}

class _BaseManagementScreenState extends State<BaseManagementScreen> {
  BaseProgressionSnapshot? _snapshot;
  List<BaseStructureDefinitionData> _catalog = const [];
  bool _loading = true;
  String? _error;
  String? _busyStructureKey;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final service = context.read<BaseService>();
      final results = await Future.wait<dynamic>([
        service.getBaseById(widget.baseId),
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

  BaseStructureLevelVisualData? _visualForLevel(
    BaseStructureDefinitionData definition,
    int level,
  ) {
    for (final visual in definition.levelVisuals) {
      if (visual.level == level) {
        return visual;
      }
    }
    return null;
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

  Widget _buildMaterialStrip(ThemeData theme) {
    final snapshot = _snapshot;
    if (snapshot == null) return const SizedBox.shrink();
    final resources = snapshot.resources.where((entry) => entry.amount > 0);
    if (resources.isEmpty) {
      return Text(
        'You have not collected any base materials yet.',
        style: theme.textTheme.bodyMedium,
      );
    }
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: resources
          .map(
            (resource) => Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: theme.colorScheme.surfaceContainerHighest.withValues(
                  alpha: 0.5,
                ),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(color: theme.colorScheme.outlineVariant),
              ),
              child: Text(
                '${_friendlyResourceName(resource.resourceKey)}: ${resource.amount}',
                style: theme.textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          )
          .toList(),
    );
  }

  Widget _buildStructureCard(
    ThemeData theme,
    BaseStructureDefinitionData definition,
  ) {
    final currentLevel = _structureLevels[definition.key] ?? 0;
    final isBuilt = currentLevel > 0;
    final nextLevel = isBuilt ? currentLevel + 1 : 1;
    final isMaxed = currentLevel >= definition.maxLevel;
    final costs = _costsForLevel(definition, nextLevel);
    final hasPrereqs = _hasMetPrerequisites(definition);
    final canAfford = _canAfford(costs);
    final isBusy = _busyStructureKey == definition.key;
    final canManage = _snapshot?.canManage == true;
    final displayLevel = isBuilt ? currentLevel : nextLevel;
    final displayVisual = _visualForLevel(definition, displayLevel);
    final visualUrl = (displayVisual?.imageUrl.trim().isNotEmpty ?? false)
        ? displayVisual!.imageUrl.trim()
        : (displayVisual?.thumbnailUrl.trim().isNotEmpty ?? false)
        ? displayVisual!.thumbnailUrl.trim()
        : '';
    final actionLabel = isMaxed
        ? 'Max level'
        : isBuilt
        ? 'Upgrade to Lv $nextLevel'
        : 'Build room';
    final buttonEnabled =
        canManage &&
        !isBusy &&
        !isMaxed &&
        hasPrereqs &&
        canAfford &&
        !_loading;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest.withValues(
          alpha: 0.28,
        ),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (visualUrl.isNotEmpty) ...[
            ClipRRect(
              borderRadius: BorderRadius.circular(14),
              child: AspectRatio(
                aspectRatio: 1.45,
                child: Image.network(
                  visualUrl,
                  fit: BoxFit.cover,
                  errorBuilder: (_, _, _) => _buildRoomImagePlaceholder(
                    theme,
                    definition,
                    displayLevel,
                  ),
                  loadingBuilder: (context, child, progress) {
                    if (progress == null) return child;
                    return _buildRoomImagePlaceholder(
                      theme,
                      definition,
                      displayLevel,
                    );
                  },
                ),
              ),
            ),
            const SizedBox(height: 14),
          ] else ...[
            _buildRoomImagePlaceholder(theme, definition, displayLevel),
            const SizedBox(height: 14),
          ],
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
                          ? 'Built, level $currentLevel of ${definition.maxLevel}'
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
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Text(
                  definition.category.toUpperCase(),
                  style: theme.textTheme.labelSmall?.copyWith(
                    fontWeight: FontWeight.w700,
                    letterSpacing: 0.7,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            definition.description,
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.4),
          ),
          if (!hasPrereqs) ...[
            const SizedBox(height: 10),
            Text(
              'Requires ${_prerequisiteText(definition)}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.error,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
          if (costs.isNotEmpty) ...[
            const SizedBox(height: 12),
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
          const SizedBox(height: 14),
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
          if (!canManage) ...[
            const SizedBox(height: 10),
            Text(
              'Only the owner of this base can build or upgrade rooms.',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildRoomImagePlaceholder(
    ThemeData theme,
    BaseStructureDefinitionData definition,
    int displayLevel,
  ) {
    return Container(
      height: 164,
      width: double.infinity,
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      alignment: Alignment.center,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            definition.name,
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 6),
          Text(
            'Level $displayLevel preview coming soon',
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final snapshot = _snapshot;
    final canManage = snapshot?.canManage == true;
    final baseTitle = canManage
        ? 'Your Base'
        : snapshot?.base?.owner.displayName ?? 'Base';
    final builtDefinitions = _catalog
        .where((definition) => (_structureLevels[definition.key] ?? 0) > 0)
        .toList();
    final buildableDefinitions = _catalog.where((definition) {
      final level = _structureLevels[definition.key] ?? 0;
      return level < definition.maxLevel;
    }).toList();

    return Scaffold(
      backgroundColor: theme.colorScheme.surface,
      appBar: AppBar(title: Text(baseTitle)),
      body: PaperSheet(
        child: RefreshIndicator(
          onRefresh: _load,
          child: ListView(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
            children: [
              if (_loading && snapshot == null)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 60),
                  child: Center(child: CircularProgressIndicator()),
                )
              else if (_error != null && snapshot == null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'We could not open this base.',
                        style: theme.textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(_error!, style: theme.textTheme.bodyMedium),
                      const SizedBox(height: 12),
                      OutlinedButton(
                        onPressed: _load,
                        child: const Text('Try again'),
                      ),
                    ],
                  ),
                )
              else if (snapshot?.base == null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Text(
                    'You do not have a base yet. Use a Home Base Kit on the map to establish one.',
                    style: theme.textTheme.bodyLarge,
                  ),
                )
              else ...[
                Container(
                  padding: const EdgeInsets.all(18),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.surfaceContainerHighest.withValues(
                      alpha: 0.36,
                    ),
                    borderRadius: BorderRadius.circular(18),
                    border: Border.all(color: theme.colorScheme.outlineVariant),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Base Overview',
                        style: theme.textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        (snapshot?.base?.description.trim().isNotEmpty ?? false)
                            ? snapshot!.base!.description.trim()
                            : canManage
                            ? 'Build new rooms, strengthen the ones you already have, and turn your base into a real source of power.'
                            : 'Take a look through this base and see which rooms have already been built.',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          height: 1.4,
                        ),
                      ),
                    ],
                  ),
                ),
                if (canManage) ...[
                  const SizedBox(height: 18),
                  Text(
                    'Materials',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 10),
                  _buildMaterialStrip(theme),
                  const SizedBox(height: 22),
                ] else ...[
                  const SizedBox(height: 18),
                  Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.surfaceContainerHighest
                          .withValues(alpha: 0.28),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: theme.colorScheme.outlineVariant,
                      ),
                    ),
                    child: Text(
                      'You can look around, but only the owner can spend materials or expand this base.',
                      style: theme.textTheme.bodyMedium?.copyWith(height: 1.35),
                    ),
                  ),
                  const SizedBox(height: 22),
                ],
                Text(
                  'Built Rooms',
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 10),
                if (builtDefinitions.isEmpty)
                  Text(
                    canManage
                        ? 'Only your Hearth stands right now. Build outward from there.'
                        : 'This base is just getting started.',
                    style: theme.textTheme.bodyMedium,
                  )
                else
                  ...builtDefinitions.map(
                    (definition) => _buildStructureCard(theme, definition),
                  ),
                const SizedBox(height: 10),
                Text(
                  canManage
                      ? 'Rooms To Build Or Improve'
                      : 'Other Possible Rooms',
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 10),
                ...buildableDefinitions.map(
                  (definition) => _buildStructureCard(theme, definition),
                ),
                if (canManage && snapshot!.activeDailyEffects.isNotEmpty) ...[
                  const SizedBox(height: 12),
                  Text(
                    'Active Base Effects',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 10),
                  ...snapshot.activeDailyEffects.map(
                    (effect) => Container(
                      margin: const EdgeInsets.only(bottom: 8),
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceContainerHighest
                            .withValues(alpha: 0.3),
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(
                          color: theme.colorScheme.outlineVariant,
                        ),
                      ),
                      child: Text(
                        _friendlyResourceName(effect.stateKey),
                        style: theme.textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                ],
                if (_error != null) ...[
                  const SizedBox(height: 8),
                  Text(
                    _error!,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.error,
                    ),
                  ),
                ],
              ],
            ],
          ),
        ),
      ),
    );
  }
}
