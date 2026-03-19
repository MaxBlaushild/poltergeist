import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/base_progression.dart';
import '../services/base_service.dart';
import '../widgets/paper_texture.dart';

const int _baseGridSize = 5;
const Color _roomBorderColor = Color(0xFF7B5A3B);
const Color _grassFallbackColor = Color(0xFF7AA65A);

class BaseManagementScreen extends StatelessWidget {
  const BaseManagementScreen({super.key, required this.baseId});

  final String baseId;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      backgroundColor: theme.colorScheme.surface,
      appBar: AppBar(title: const Text('Your Base')),
      body: PaperSheet(child: BaseManagementContent(baseId: baseId)),
    );
  }
}

class BaseManagementContent extends StatefulWidget {
  const BaseManagementContent({
    super.key,
    required this.baseId,
    this.padding = const EdgeInsets.fromLTRB(16, 16, 16, 32),
  });

  final String baseId;
  final EdgeInsets padding;

  @override
  State<BaseManagementContent> createState() => _BaseManagementContentState();
}

class _BaseManagementContentState extends State<BaseManagementContent> {
  BaseProgressionSnapshot? _snapshot;
  List<BaseStructureDefinitionData> _catalog = const [];
  bool _loading = true;
  String? _error;
  String? _busyStructureKey;
  bool _editingBaseDetails = false;
  bool _savingBaseDetails = false;
  _GridCell? _buildSelectionCell;
  String? _moveAnchorStructureKey;
  Set<String> _moveStructureKeys = const <String>{};
  _GridCell? _moveTargetCell;
  final TextEditingController _baseNameController = TextEditingController();
  final TextEditingController _baseDescriptionController =
      TextEditingController();

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _baseNameController.dispose();
    _baseDescriptionController.dispose();
    super.dispose();
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
      _syncBaseEditorsToSnapshot();
      _syncMoveStateToSnapshot();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = e.toString();
      });
    }
  }

  void _syncBaseEditorsToSnapshot() {
    final base = _snapshot?.base;
    if (base == null) {
      _baseNameController.text = '';
      _baseDescriptionController.text = '';
      return;
    }
    _baseNameController.text = base.name;
    _baseDescriptionController.text = base.description;
  }

  void _syncMoveStateToSnapshot() {
    final anchorKey = _moveAnchorStructureKey;
    if (anchorKey == null) return;
    final structure = _structureByKey[anchorKey];
    if (structure == null) {
      _cancelMoveMode();
      return;
    }
    if (!mounted) return;
    setState(() {
      _moveStructureKeys = _moveStructureKeys
          .where((key) => _structureByKey.containsKey(key))
          .toSet();
      if (!_moveStructureKeys.contains(anchorKey)) {
        _moveStructureKeys = <String>{anchorKey, ..._moveStructureKeys};
      }
      _moveTargetCell = _GridCell(structure.gridX, structure.gridY);
    });
  }

  Future<void> _mutateStructure(
    BaseStructureDefinitionData definition,
    bool isUpgrade, {
    int? gridX,
    int? gridY,
  }) async {
    setState(() {
      _busyStructureKey = definition.key;
      _error = null;
    });
    try {
      final service = context.read<BaseService>();
      final nextSnapshot = isUpgrade
          ? await service.upgradeStructure(definition.key)
          : await service.buildStructure(
              definition.key,
              gridX: gridX!,
              gridY: gridY!,
            );
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
      _syncMoveStateToSnapshot();
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

  Future<void> _moveSelectedRooms() async {
    final anchorKey = _moveAnchorStructureKey;
    final target = _moveTargetCell;
    if (anchorKey == null || target == null || !_isCurrentMoveTargetValid) {
      return;
    }
    setState(() {
      _busyStructureKey = anchorKey;
      _error = null;
    });
    try {
      final nextSnapshot = await context.read<BaseService>().moveRooms(
        anchorStructureKey: anchorKey,
        structureKeys: _moveStructureKeys.toList(growable: false),
        targetGridX: target.gridX,
        targetGridY: target.gridY,
      );
      if (!mounted) return;
      setState(() {
        _snapshot = nextSnapshot;
        _busyStructureKey = null;
      });
      _cancelMoveMode();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Base rooms moved.')));
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

  Future<void> _saveBaseDetails() async {
    if (_snapshot?.canManage != true) return;
    setState(() {
      _savingBaseDetails = true;
      _error = null;
    });
    try {
      final nextSnapshot = await context
          .read<BaseService>()
          .updateMyBaseDetails(
            name: _baseNameController.text,
            description: _baseDescriptionController.text,
          );
      if (!mounted) return;
      setState(() {
        _snapshot = nextSnapshot;
        _editingBaseDetails = false;
        _savingBaseDetails = false;
      });
      _syncBaseEditorsToSnapshot();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Base details updated.')));
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _savingBaseDetails = false;
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

  Map<String, UserBaseStructureData> get _structureByKey {
    final snapshot = _snapshot;
    if (snapshot == null) return const <String, UserBaseStructureData>{};
    final values = <String, UserBaseStructureData>{};
    for (final structure in snapshot.structures) {
      values[structure.structureKey] = structure;
    }
    return values;
  }

  Map<String, int> get _structureLevels {
    final values = <String, int>{};
    for (final structure in _structureByKey.values) {
      values[structure.structureKey] = structure.level;
    }
    return values;
  }

  Map<String, _GridCell> get _occupiedCells {
    final values = <String, _GridCell>{};
    for (final structure in _structureByKey.values) {
      values['${structure.gridX}:${structure.gridY}'] = _GridCell(
        structure.gridX,
        structure.gridY,
      );
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

  BaseStructureDefinitionData? _definitionForKey(String key) {
    for (final definition in _catalog) {
      if (definition.key == key) return definition;
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

  IconData _materialIcon(String key) {
    switch (key) {
      case 'timber':
        return Icons.park;
      case 'stone':
        return Icons.landscape;
      case 'iron':
        return Icons.hardware;
      case 'herbs':
        return Icons.local_florist;
      case 'monster_parts':
        return Icons.pets;
      case 'arcane_dust':
        return Icons.auto_awesome;
      case 'relic_shards':
        return Icons.diamond;
      default:
        return Icons.inventory_2;
    }
  }

  Color _materialAccentColor(String key) {
    switch (key) {
      case 'timber':
        return const Color(0xFF8D6E63);
      case 'stone':
        return const Color(0xFF78909C);
      case 'iron':
        return const Color(0xFF546E7A);
      case 'herbs':
        return const Color(0xFF43A047);
      case 'monster_parts':
        return const Color(0xFFC62828);
      case 'arcane_dust':
        return const Color(0xFF6A5ACD);
      case 'relic_shards':
        return const Color(0xFF00897B);
      default:
        return const Color(0xFF616161);
    }
  }

  bool _isWithinGrid(int gridX, int gridY) {
    return gridX >= 0 &&
        gridX < _baseGridSize &&
        gridY >= 0 &&
        gridY < _baseGridSize;
  }

  bool _isAdjacent(_GridCell a, _GridCell b) {
    return (a.gridX - b.gridX).abs() + (a.gridY - b.gridY).abs() == 1;
  }

  Set<String> get _adjacentBuildCellKeys {
    final keys = <String>{};
    for (final structure in _structureByKey.values) {
      for (final delta in const <_GridCell>[
        _GridCell(-1, 0),
        _GridCell(1, 0),
        _GridCell(0, -1),
        _GridCell(0, 1),
      ]) {
        final nextX = structure.gridX + delta.gridX;
        final nextY = structure.gridY + delta.gridY;
        if (!_isWithinGrid(nextX, nextY)) continue;
        if (_occupiedCells.containsKey('$nextX:$nextY')) continue;
        keys.add('$nextX:$nextY');
      }
    }
    return keys;
  }

  List<BaseStructureDefinitionData> _buildOptionsForCell(_GridCell cell) {
    return _catalog.where((definition) {
      if ((_structureLevels[definition.key] ?? 0) > 0) return false;
      return _hasMetPrerequisites(definition);
    }).toList()..sort((a, b) => a.sortOrder.compareTo(b.sortOrder));
  }

  void _startMoveMode(String structureKey) {
    final structure = _structureByKey[structureKey];
    if (structure == null) return;
    setState(() {
      _moveAnchorStructureKey = structureKey;
      _moveStructureKeys = <String>{structureKey};
      _moveTargetCell = _GridCell(structure.gridX, structure.gridY);
    });
  }

  void _cancelMoveMode() {
    if (!mounted) return;
    setState(() {
      _moveAnchorStructureKey = null;
      _moveStructureKeys = const <String>{};
      _moveTargetCell = null;
    });
  }

  bool _canMoveSelectionTo(_GridCell targetCell) {
    final anchorKey = _moveAnchorStructureKey;
    if (anchorKey == null) return false;
    final anchor = _structureByKey[anchorKey];
    if (anchor == null) return false;

    final deltaX = targetCell.gridX - anchor.gridX;
    final deltaY = targetCell.gridY - anchor.gridY;
    final occupiedByUnselected = <String>{};
    for (final structure in _structureByKey.values) {
      if (!_moveStructureKeys.contains(structure.structureKey)) {
        occupiedByUnselected.add('${structure.gridX}:${structure.gridY}');
      }
    }

    final projectedPositions = <String, _GridCell>{};
    for (final structure in _structureByKey.values) {
      final moving = _moveStructureKeys.contains(structure.structureKey);
      final position = moving
          ? _GridCell(structure.gridX + deltaX, structure.gridY + deltaY)
          : _GridCell(structure.gridX, structure.gridY);
      if (!_isWithinGrid(position.gridX, position.gridY)) {
        return false;
      }
      final key = '${position.gridX}:${position.gridY}';
      if (projectedPositions.containsKey(key)) return false;
      if (moving && occupiedByUnselected.contains(key)) return false;
      projectedPositions[key] = position;
    }
    return _isProjectedLayoutConnected(projectedPositions.values.toList());
  }

  bool _isProjectedLayoutConnected(List<_GridCell> cells) {
    if (cells.length <= 1) return true;
    final remaining = cells.toSet();
    final queue = <_GridCell>[cells.first];
    final visited = <_GridCell>{cells.first};
    while (queue.isNotEmpty) {
      final current = queue.removeAt(0);
      for (final candidate in remaining) {
        if (visited.contains(candidate)) continue;
        if (_isAdjacent(current, candidate)) {
          visited.add(candidate);
          queue.add(candidate);
        }
      }
    }
    return visited.length == cells.length;
  }

  bool get _isCurrentMoveTargetValid {
    final target = _moveTargetCell;
    if (target == null) return false;
    return _canMoveSelectionTo(target);
  }

  void _toggleLinkedRoom(String structureKey) {
    final anchorKey = _moveAnchorStructureKey;
    if (anchorKey == null || structureKey == anchorKey) return;
    final next = Set<String>.from(_moveStructureKeys);
    if (!next.add(structureKey)) {
      next.remove(structureKey);
    }
    setState(() {
      _moveStructureKeys = next;
    });
  }

  void _showBuildOptions(_GridCell cell) {
    if (!_snapshot!.canManage) return;
    final options = _buildOptionsForCell(cell);
    if (options.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('No rooms are available to build there yet.'),
        ),
      );
      return;
    }
    setState(() {
      _buildSelectionCell = cell;
    });
  }

  Future<void> _showRoomDetails(UserBaseStructureData structure) async {
    final definition = _definitionForKey(structure.structureKey);
    if (definition == null) return;
    final shouldStartMove = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      builder: (context) => _RoomDetailsSheet(
        definition: definition,
        structure: structure,
        canManage: _snapshot?.canManage == true,
        isBusy: _busyStructureKey == definition.key,
        isMaxed: structure.level >= definition.maxLevel,
        resourceAmounts: _resourceAmounts,
        materialIcon: _materialIcon,
        materialAccentColor: _materialAccentColor,
        costs: _costsForLevel(definition, structure.level + 1),
        canAffordUpgrade: _canAfford(
          _costsForLevel(definition, structure.level + 1),
        ),
        visual: _visualForLevel(definition, structure.level),
        onUpgrade:
            structure.level < definition.maxLevel &&
                _snapshot?.canManage == true &&
                _canAfford(_costsForLevel(definition, structure.level + 1))
            ? () async {
                Navigator.of(context).pop(false);
                await _mutateStructure(definition, true);
              }
            : null,
      ),
    );
    if (shouldStartMove == true) {
      _startMoveMode(structure.structureKey);
    }
  }

  Widget _buildMaterialStrip(ThemeData theme) {
    final snapshot = _snapshot;
    if (snapshot == null) return const SizedBox.shrink();
    final resources = snapshot.resources;
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

  Widget _buildCostBadge(ThemeData theme, BaseStructureCostData cost) {
    final available = _resourceAmounts[cost.resourceKey] ?? 0;
    final hasEnough = available >= cost.amount;
    final accentColor = hasEnough
        ? _materialAccentColor(cost.resourceKey)
        : theme.colorScheme.error;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: accentColor, width: 1.5),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(_materialIcon(cost.resourceKey), size: 18, color: accentColor),
          const SizedBox(width: 6),
          Text(
            '$available / ${cost.amount}',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.bold,
              color: accentColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBuildSelectionView(
    ThemeData theme,
    List<BaseStructureDefinitionData> options,
  ) {
    final selectedCell = _buildSelectionCell;
    if (selectedCell == null) {
      return const SizedBox.shrink();
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            IconButton(
              onPressed: () {
                setState(() {
                  _buildSelectionCell = null;
                });
              },
              icon: const Icon(Icons.arrow_back),
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints.tightFor(width: 36, height: 36),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Text(
          'Choose a room to add on this tile.',
          style: theme.textTheme.bodyMedium,
        ),
        const SizedBox(height: 16),
        ...options.map((definition) {
          final costs = _costsForLevel(definition, 1);
          final isAffordable = _canAfford(costs);
          final hasPrereqs = _hasMetPrerequisites(definition);
          final isBusy = _busyStructureKey == definition.key;
          final previewVisual = _visualForLevel(definition, 1);
          final previewImageUrl =
              (previewVisual?.imageUrl.trim().isNotEmpty ?? false)
              ? previewVisual!.imageUrl.trim()
              : (previewVisual?.thumbnailUrl.trim().isNotEmpty ?? false)
              ? previewVisual!.thumbnailUrl.trim()
              : '';
          return Container(
            margin: const EdgeInsets.only(bottom: 12),
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: theme.colorScheme.outlineVariant),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (previewImageUrl.isNotEmpty) ...[
                  ClipRRect(
                    borderRadius: BorderRadius.circular(14),
                    child: AspectRatio(
                      aspectRatio: 1.45,
                      child: Image.network(
                        previewImageUrl,
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) =>
                            _RoomFallbackLabel(title: definition.name),
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                ],
                Text(
                  definition.name,
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 6),
                Text(definition.description),
                if (!hasPrereqs) ...[
                  const SizedBox(height: 8),
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
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: costs
                        .map((cost) => _buildCostBadge(theme, cost))
                        .toList(),
                  ),
                ],
                const SizedBox(height: 10),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    onPressed: hasPrereqs && isAffordable && !isBusy
                        ? () async {
                            await _mutateStructure(
                              definition,
                              false,
                              gridX: selectedCell.gridX,
                              gridY: selectedCell.gridY,
                            );
                            if (!mounted) return;
                            setState(() {
                              _buildSelectionCell = null;
                            });
                          }
                        : null,
                    child: isBusy
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(
                            hasPrereqs && isAffordable
                                ? 'Build ${definition.name}'
                                : !hasPrereqs
                                ? 'Locked'
                                : 'Need More Materials',
                          ),
                  ),
                ),
              ],
            ),
          );
        }),
      ],
    );
  }

  Widget _buildGrid(ThemeData theme) {
    final snapshot = _snapshot;
    if (snapshot == null) return const SizedBox.shrink();
    final structuresByCell = <String, UserBaseStructureData>{};
    for (final structure in snapshot.structures) {
      structuresByCell['${structure.gridX}:${structure.gridY}'] = structure;
    }
    final adjacentBuildKeys = _adjacentBuildCellKeys;

    return LayoutBuilder(
      builder: (context, constraints) {
        const spacing = 6.0;
        final tileSize = (constraints.maxWidth - (spacing * 4)) / 5;
        return Column(
          children: List.generate(_baseGridSize, (row) {
            return Padding(
              padding: EdgeInsets.only(
                bottom: row == _baseGridSize - 1 ? 0 : spacing,
              ),
              child: Row(
                children: List.generate(_baseGridSize, (column) {
                  final cell = _GridCell(column, row);
                  final structure = structuresByCell['$column:$row'];
                  final isAdjacentBuildCell = adjacentBuildKeys.contains(
                    '$column:$row',
                  );
                  final isMoveTarget = _moveTargetCell == cell;
                  final moveOverlayColor = _moveAnchorStructureKey == null
                      ? null
                      : (_canMoveSelectionTo(cell)
                            ? const Color(0x663FAE5A)
                            : const Color(0x66C94B4B));

                  return Padding(
                    padding: EdgeInsets.only(
                      right: column == _baseGridSize - 1 ? 0 : spacing,
                    ),
                    child: SizedBox(
                      width: tileSize,
                      height: tileSize,
                      child: _BaseGridTile(
                        tileSize: tileSize,
                        grassTileUrl:
                            snapshot.grassTileUrls['$column:$row'] ?? '',
                        structure: structure,
                        definition: structure == null
                            ? null
                            : _definitionForKey(structure.structureKey),
                        visual: structure == null
                            ? null
                            : _visualForLevel(
                                _definitionForKey(structure.structureKey)!,
                                structure.level,
                              ),
                        showPlus:
                            snapshot.canManage &&
                            structure == null &&
                            _moveAnchorStructureKey == null &&
                            isAdjacentBuildCell,
                        moveOverlayColor: moveOverlayColor,
                        isSelectedMoveTarget: isMoveTarget,
                        isLockedRoom:
                            structure != null &&
                            _moveStructureKeys.contains(structure.structureKey),
                        canManage: snapshot.canManage,
                        onTap: () {
                          if (_moveAnchorStructureKey != null) {
                            setState(() {
                              _moveTargetCell = cell;
                            });
                            return;
                          }
                          if (structure != null) {
                            _showRoomDetails(structure);
                            return;
                          }
                          if (isAdjacentBuildCell) {
                            _showBuildOptions(cell);
                          }
                        },
                      ),
                    ),
                  );
                }),
              ),
            );
          }),
        );
      },
    );
  }

  Widget _buildMoveControls(ThemeData theme) {
    final anchorKey = _moveAnchorStructureKey;
    if (anchorKey == null) return const SizedBox.shrink();
    return Container(
      margin: const EdgeInsets.only(top: 14),
      padding: const EdgeInsets.all(14),
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
          Text(
            'Move Rooms',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Tap other room chips to lock them into this move as a block, then tap a tile on the grid to choose the destination.',
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.35),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _structureByKey.values.map((structure) {
              final isAnchor = structure.structureKey == anchorKey;
              final selected = _moveStructureKeys.contains(
                structure.structureKey,
              );
              return FilterChip(
                label: Text(_friendlyStructureName(structure.structureKey)),
                selected: selected,
                onSelected: isAnchor
                    ? null
                    : (_) => _toggleLinkedRoom(structure.structureKey),
              );
            }).toList(),
          ),
          const SizedBox(height: 12),
          if (_moveTargetCell != null)
            Text(
              _isCurrentMoveTargetValid
                  ? 'That destination works. Confirm when you are ready.'
                  : 'That destination would break the layout or overlap another room.',
              style: theme.textTheme.bodySmall?.copyWith(
                color: _isCurrentMoveTargetValid
                    ? theme.colorScheme.primary
                    : theme.colorScheme.error,
                fontWeight: FontWeight.w700,
              ),
            ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: OutlinedButton(
                  onPressed: _busyStructureKey == null ? _cancelMoveMode : null,
                  child: const Text('Cancel'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton(
                  onPressed:
                      _busyStructureKey == null && _isCurrentMoveTargetValid
                      ? _moveSelectedRooms
                      : null,
                  child: _busyStructureKey == anchorKey
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Move Here'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildActiveEffects(ThemeData theme) {
    final effects =
        _snapshot?.activeDailyEffects ?? const <BaseDailyEffectData>[];
    if (effects.isEmpty) return const SizedBox.shrink();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 18),
        Text(
          'Active Base Effects',
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 10),
        ...effects.map(
          (effect) => Container(
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest.withValues(
                alpha: 0.3,
              ),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: theme.colorScheme.outlineVariant),
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
    );
  }

  Widget _buildBaseFlavorSection(ThemeData theme) {
    final snapshot = _snapshot;
    final base = snapshot?.base;
    if (snapshot == null || base == null) {
      return const SizedBox.shrink();
    }

    final baseName = base.name.trim();
    final description = base.description.trim();
    final hasFlavor = baseName.isNotEmpty || description.isNotEmpty;
    if (!snapshot.canManage && !hasFlavor) {
      return const SizedBox.shrink();
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 18),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest.withValues(
          alpha: 0.22,
        ),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: _editingBaseDetails && snapshot.canManage
          ? Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Edit Base Flavor',
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _baseNameController,
                  enabled: !_savingBaseDetails,
                  textCapitalization: TextCapitalization.words,
                  decoration: const InputDecoration(
                    labelText: 'Base name',
                    hintText: 'Give your base a name',
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _baseDescriptionController,
                  enabled: !_savingBaseDetails,
                  textCapitalization: TextCapitalization.sentences,
                  minLines: 3,
                  maxLines: 5,
                  decoration: const InputDecoration(
                    labelText: 'Base description',
                    hintText: 'Describe what makes this base yours',
                  ),
                ),
                const SizedBox(height: 14),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                        onPressed: _savingBaseDetails
                            ? null
                            : () {
                                _syncBaseEditorsToSnapshot();
                                setState(() {
                                  _editingBaseDetails = false;
                                });
                              },
                        child: const Text('Cancel'),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: FilledButton(
                        onPressed: _savingBaseDetails ? null : _saveBaseDetails,
                        child: _savingBaseDetails
                            ? const SizedBox(
                                width: 18,
                                height: 18,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Text('Save'),
                      ),
                    ),
                  ],
                ),
              ],
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (baseName.isNotEmpty)
                  Text(
                    baseName,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                if (baseName.isNotEmpty && description.isNotEmpty)
                  const SizedBox(height: 8),
                Text(
                  description.isNotEmpty
                      ? description
                      : snapshot.canManage
                      ? 'Give your base a name and description so it feels like home.'
                      : 'This base has not been described yet.',
                  style: theme.textTheme.bodyMedium?.copyWith(height: 1.4),
                ),
                if (snapshot.canManage) ...[
                  const SizedBox(height: 14),
                  Align(
                    alignment: Alignment.centerLeft,
                    child: OutlinedButton.icon(
                      onPressed: () {
                        _syncBaseEditorsToSnapshot();
                        setState(() {
                          _editingBaseDetails = true;
                        });
                      },
                      icon: const Icon(Icons.edit_outlined),
                      label: const Text('Edit Details'),
                    ),
                  ),
                ],
              ],
            ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final snapshot = _snapshot;
    final buildSelectionCell = _buildSelectionCell;
    final buildOptions = buildSelectionCell == null
        ? const <BaseStructureDefinitionData>[]
        : _buildOptionsForCell(buildSelectionCell);

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: widget.padding,
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
          else if (buildSelectionCell != null)
            _buildBuildSelectionView(theme, buildOptions)
          else ...[
            _buildBaseFlavorSection(theme),
            _buildGrid(theme),
            _buildMoveControls(theme),
            if (snapshot!.canManage) ...[
              const SizedBox(height: 18),
              Text(
                'Materials',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 10),
              _buildMaterialStrip(theme),
            ],
            _buildActiveEffects(theme),
            if (_error != null) ...[
              const SizedBox(height: 12),
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
    );
  }
}

class _BaseGridTile extends StatelessWidget {
  const _BaseGridTile({
    required this.tileSize,
    required this.grassTileUrl,
    required this.structure,
    required this.definition,
    required this.visual,
    required this.showPlus,
    required this.moveOverlayColor,
    required this.isSelectedMoveTarget,
    required this.isLockedRoom,
    required this.canManage,
    required this.onTap,
  });

  final double tileSize;
  final String grassTileUrl;
  final UserBaseStructureData? structure;
  final BaseStructureDefinitionData? definition;
  final BaseStructureLevelVisualData? visual;
  final bool showPlus;
  final Color? moveOverlayColor;
  final bool isSelectedMoveTarget;
  final bool isLockedRoom;
  final bool canManage;
  final VoidCallback onTap;

  String get _roomImageUrl {
    if (visual != null && visual!.imageUrl.trim().isNotEmpty) {
      return visual!.imageUrl.trim();
    }
    if (visual != null && visual!.thumbnailUrl.trim().isNotEmpty) {
      return visual!.thumbnailUrl.trim();
    }
    return '';
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasRoom = structure != null && definition != null;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: ClipRRect(
          borderRadius: BorderRadius.circular(12),
          child: Stack(
            fit: StackFit.expand,
            children: [
              _GrassTile(url: grassTileUrl),
              if (hasRoom)
                _roomImageUrl.isNotEmpty
                    ? Image.network(
                        _roomImageUrl,
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) =>
                            _RoomFallbackLabel(title: definition!.name),
                      )
                    : _RoomFallbackLabel(title: definition!.name),
              if (moveOverlayColor != null)
                ColoredBox(color: moveOverlayColor!),
              if (showPlus)
                Center(
                  child: Icon(
                    Icons.add,
                    color: _roomBorderColor.withValues(alpha: 0.38),
                    size: tileSize * 0.28,
                  ),
                ),
              if (hasRoom)
                Container(
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: isLockedRoom
                          ? theme.colorScheme.primary
                          : _roomBorderColor,
                      width: isLockedRoom ? 3 : 2,
                    ),
                  ),
                ),
              if (isSelectedMoveTarget)
                Container(
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: Colors.white, width: 3),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _GrassTile extends StatelessWidget {
  const _GrassTile({required this.url});

  final String url;

  @override
  Widget build(BuildContext context) {
    if (url.trim().isEmpty) {
      return const ColoredBox(color: _grassFallbackColor);
    }
    return Image.network(
      url,
      fit: BoxFit.cover,
      errorBuilder: (_, _, _) => const ColoredBox(color: _grassFallbackColor),
    );
  }
}

class _RoomFallbackLabel extends StatelessWidget {
  const _RoomFallbackLabel({required this.title});

  final String title;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ColoredBox(
      color: theme.colorScheme.surface.withValues(alpha: 0.92),
      child: Center(
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Text(
            title,
            textAlign: TextAlign.center,
            style: theme.textTheme.labelMedium?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ),
    );
  }
}

class _RoomDetailsSheet extends StatelessWidget {
  const _RoomDetailsSheet({
    required this.definition,
    required this.structure,
    required this.canManage,
    required this.isBusy,
    required this.isMaxed,
    required this.resourceAmounts,
    required this.materialIcon,
    required this.materialAccentColor,
    required this.costs,
    required this.canAffordUpgrade,
    required this.visual,
    required this.onUpgrade,
  });

  final BaseStructureDefinitionData definition;
  final UserBaseStructureData structure;
  final bool canManage;
  final bool isBusy;
  final bool isMaxed;
  final Map<String, int> resourceAmounts;
  final IconData Function(String key) materialIcon;
  final Color Function(String key) materialAccentColor;
  final List<BaseStructureCostData> costs;
  final bool canAffordUpgrade;
  final BaseStructureLevelVisualData? visual;
  final Future<void> Function()? onUpgrade;

  String get _visualUrl {
    if (visual != null && visual!.imageUrl.trim().isNotEmpty) {
      return visual!.imageUrl.trim();
    }
    if (visual != null && visual!.thumbnailUrl.trim().isNotEmpty) {
      return visual!.thumbnailUrl.trim();
    }
    return '';
  }

  Widget _buildCostBadge(BuildContext context, BaseStructureCostData cost) {
    final theme = Theme.of(context);
    final available = resourceAmounts[cost.resourceKey] ?? 0;
    final hasEnough = available >= cost.amount;
    final accentColor = hasEnough
        ? materialAccentColor(cost.resourceKey)
        : theme.colorScheme.error;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: accentColor, width: 1.5),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(materialIcon(cost.resourceKey), size: 18, color: accentColor),
          const SizedBox(width: 6),
          Text(
            '$available / ${cost.amount}',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.bold,
              color: accentColor,
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              definition.name,
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'Level ${structure.level} of ${definition.maxLevel}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 16),
            if (_visualUrl.isNotEmpty)
              ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: AspectRatio(
                  aspectRatio: 1.4,
                  child: Image.network(
                    _visualUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (_, _, _) =>
                        _RoomFallbackLabel(title: definition.name),
                  ),
                ),
              ),
            if (_visualUrl.isEmpty) ...[
              ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: SizedBox(
                  height: 180,
                  width: double.infinity,
                  child: _RoomFallbackLabel(title: definition.name),
                ),
              ),
            ],
            const SizedBox(height: 16),
            Text(
              definition.description,
              style: theme.textTheme.bodyMedium?.copyWith(height: 1.4),
            ),
            if (costs.isNotEmpty) ...[
              const SizedBox(height: 16),
              Text(
                'Upgrade Cost',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: costs
                    .map((cost) => _buildCostBadge(context, cost))
                    .toList(),
              ),
            ],
            const SizedBox(height: 18),
            if (canManage && !isMaxed)
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: isBusy || !canAffordUpgrade || onUpgrade == null
                      ? null
                      : () => onUpgrade!.call(),
                  child: isBusy
                      ? const SizedBox(
                          height: 18,
                          width: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(
                          canAffordUpgrade
                              ? 'Upgrade Room'
                              : 'Need More Materials',
                        ),
                ),
              ),
            if (canManage) ...[
              const SizedBox(height: 10),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Move Room'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

@immutable
class _GridCell {
  const _GridCell(this.gridX, this.gridY);

  final int gridX;
  final int gridY;

  @override
  bool operator ==(Object other) {
    return other is _GridCell && other.gridX == gridX && other.gridY == gridY;
  }

  @override
  int get hashCode => Object.hash(gridX, gridY);
}
