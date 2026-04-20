import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/base.dart';
import '../models/base_progression.dart';
import '../models/inventory_item.dart';
import '../models/tutorial.dart';
import '../models/zone.dart';
import '../providers/character_stats_provider.dart';
import '../providers/zone_provider.dart';
import '../services/base_service.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';
import '../widgets/paper_texture.dart';
import '../widgets/inventory_requirement_chip.dart';

const int _baseGridSize = 5;
const Color _roomBorderColor = Color(0xFF7B5A3B);
const Color _grassFallbackColor = Color(0xFF7AA65A);
const Color _buildSlotPlusColor = Color(0xE6100C08);
const String _hearthRecoveryStateKey = 'hearth_recovery';
const Duration _hearthRecoveryCooldownDuration = Duration(days: 1);

class BaseManagementScreen extends StatefulWidget {
  const BaseManagementScreen({
    super.key,
    required this.baseId,
    this.onTutorialProgressChanged,
  });

  final String baseId;
  final Future<void> Function()? onTutorialProgressChanged;

  @override
  State<BaseManagementScreen> createState() => _BaseManagementScreenState();
}

class _BaseManagementScreenState extends State<BaseManagementScreen> {
  final GlobalKey<_BaseManagementContentState> _contentKey =
      GlobalKey<_BaseManagementContentState>();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final contentState = _contentKey.currentState;
    return Scaffold(
      backgroundColor: theme.colorScheme.surface,
      appBar: AppBar(
        centerTitle: true,
        title: contentState?.buildHeaderTitle(theme) ?? const Text('Your Base'),
      ),
      body: PaperSheet(
        child: BaseManagementContent(
          key: _contentKey,
          baseId: widget.baseId,
          onTutorialProgressChanged: widget.onTutorialProgressChanged,
          onHeaderChanged: () {
            if (mounted) {
              setState(() {});
            }
          },
        ),
      ),
    );
  }
}

class BaseManagementSheet extends StatefulWidget {
  const BaseManagementSheet({
    super.key,
    required this.baseId,
    this.onClose,
    this.onTutorialProgressChanged,
  });

  final String baseId;
  final VoidCallback? onClose;
  final Future<void> Function()? onTutorialProgressChanged;

  @override
  State<BaseManagementSheet> createState() => _BaseManagementSheetState();
}

class _BaseManagementSheetState extends State<BaseManagementSheet> {
  final GlobalKey<_BaseManagementContentState> _contentKey =
      GlobalKey<_BaseManagementContentState>();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final contentState = _contentKey.currentState;
    final maxHeight = MediaQuery.sizeOf(context).height * 0.92;

    return PaperSheet(
      child: ConstrainedBox(
        constraints: BoxConstraints(maxHeight: maxHeight),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 8, 0),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: DefaultTextStyle(
                      style:
                          theme.textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.w800,
                          ) ??
                          const TextStyle(
                            fontSize: 22,
                            fontWeight: FontWeight.w800,
                          ),
                      child:
                          contentState?.buildHeaderTitle(theme) ??
                          const Text('Your Base'),
                    ),
                  ),
                  IconButton(
                    onPressed:
                        widget.onClose ?? () => Navigator.of(context).pop(),
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: BaseManagementContent(
                key: _contentKey,
                baseId: widget.baseId,
                onTutorialProgressChanged: widget.onTutorialProgressChanged,
                onHeaderChanged: () {
                  if (mounted) {
                    setState(() {});
                  }
                },
                padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class BaseManagementContent extends StatefulWidget {
  const BaseManagementContent({
    super.key,
    required this.baseId,
    this.onTutorialProgressChanged,
    this.onHeaderChanged,
    this.padding = const EdgeInsets.fromLTRB(16, 16, 16, 32),
  });

  final String baseId;
  final Future<void> Function()? onTutorialProgressChanged;
  final VoidCallback? onHeaderChanged;
  final EdgeInsets padding;

  @override
  State<BaseManagementContent> createState() => _BaseManagementContentState();
}

class _BaseManagementContentState extends State<BaseManagementContent> {
  BaseProgressionSnapshot? _snapshot;
  TutorialStatus? _tutorialStatus;
  List<BaseStructureDefinitionData> _catalog = const [];
  bool _loading = true;
  String? _error;
  String? _busyStructureKey;
  bool _editingBaseName = false;
  bool _savingBaseDetails = false;
  bool _selectingMoveRoom = false;
  String? _moveAnchorStructureKey;
  _GridCell? _buildSelectionCell;
  final TextEditingController _baseNameController = TextEditingController();
  final FocusNode _baseNameFocusNode = FocusNode();

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _baseNameController.dispose();
    _baseNameFocusNode.dispose();
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
      _notifyHeaderChanged();
      await _notifyTutorialProgressChanged();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = e.toString();
      });
      _notifyHeaderChanged();
    }
  }

  void _syncBaseEditorsToSnapshot() {
    final base = _snapshot?.base;
    if (base == null) {
      _baseNameController.text = '';
      return;
    }
    _baseNameController.text = base.name;
  }

  void _notifyHeaderChanged() {
    widget.onHeaderChanged?.call();
  }

  Future<void> _notifyTutorialProgressChanged() async {
    try {
      final status = await context.read<PoiService>().getTutorialStatus();
      if (mounted) {
        setState(() {
          _tutorialStatus = status;
        });
      }
    } catch (error) {
      debugPrint(
        'BaseManagementContent: tutorial status refresh failed: $error',
      );
    }

    final onTutorialProgressChanged = widget.onTutorialProgressChanged;
    if (onTutorialProgressChanged == null) return;
    try {
      await onTutorialProgressChanged();
    } catch (error) {
      debugPrint(
        'BaseManagementContent: tutorial progress refresh failed: $error',
      );
    }
  }

  UserBaseStructureData? get _moveAnchorStructure {
    final key = _moveAnchorStructureKey;
    if (key == null) return null;
    return _structureByKey[key];
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
      await _notifyTutorialProgressChanged();
      if (!mounted) return;
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
            description: _snapshot?.base?.description ?? '',
          );
      if (!mounted) return;
      setState(() {
        _snapshot = nextSnapshot;
        _editingBaseName = false;
        _savingBaseDetails = false;
      });
      _syncBaseEditorsToSnapshot();
      _notifyHeaderChanged();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Base details updated.')));
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _savingBaseDetails = false;
        _error = e.toString();
      });
      _notifyHeaderChanged();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(e.toString())));
    }
  }

  void _beginInlineEdit({
    required bool editingName,
    required TextEditingController controller,
    required FocusNode focusNode,
  }) {
    if (_savingBaseDetails) return;
    setState(() {
      _editingBaseName = editingName;
    });
    _notifyHeaderChanged();
    controller.selection = TextSelection.collapsed(
      offset: controller.text.length,
    );
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      focusNode.requestFocus();
    });
  }

  void _cancelBaseNameEdit() {
    _baseNameController.text = _snapshot?.base?.name ?? '';
    setState(() {
      _editingBaseName = false;
    });
    _notifyHeaderChanged();
  }

  String _ownerBaseTitle(BasePin base) {
    final username = base.owner.username.trim();
    if (username.isNotEmpty) {
      return "$username's Base";
    }
    final name = base.owner.name.trim();
    if (name.isNotEmpty) {
      return "$name's Base";
    }
    return 'Your Base';
  }

  String _displayBaseTitle(BasePin base) {
    final explicitName = base.name.trim();
    if (explicitName.isNotEmpty) {
      return explicitName;
    }
    return _ownerBaseTitle(base);
  }

  Widget buildHeaderTitle(ThemeData theme) {
    final snapshot = _snapshot;
    final base = snapshot?.base;
    if (snapshot == null || base == null) {
      return const Text('Your Base');
    }

    if (_editingBaseName) {
      return Wrap(
        crossAxisAlignment: WrapCrossAlignment.center,
        spacing: 4,
        runSpacing: 4,
        children: [
          SizedBox(
            width: 220,
            child: TextField(
              controller: _baseNameController,
              focusNode: _baseNameFocusNode,
              enabled: !_savingBaseDetails,
              textCapitalization: TextCapitalization.words,
              textInputAction: TextInputAction.done,
              onSubmitted: _savingBaseDetails
                  ? null
                  : (_) => _saveBaseDetails(),
              decoration: InputDecoration(
                isDense: true,
                hintText: _displayBaseTitle(base),
              ),
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          if (snapshot.canManage)
            _InlineEditActions(
              editing: true,
              saving: _savingBaseDetails,
              editTooltip: 'Edit base name',
              confirmTooltip: 'Save base details',
              cancelTooltip: 'Cancel name edit',
              onEdit: () {},
              onConfirm: _saveBaseDetails,
              onCancel: _cancelBaseNameEdit,
            ),
        ],
      );
    }

    return Text.rich(
      TextSpan(
        children: [
          TextSpan(text: _displayBaseTitle(base)),
          if (snapshot.canManage)
            WidgetSpan(
              alignment: PlaceholderAlignment.middle,
              child: Padding(
                padding: const EdgeInsets.only(left: 4),
                child: _InlineEditActions(
                  editing: false,
                  saving: _savingBaseDetails,
                  editTooltip: 'Edit base name',
                  confirmTooltip: 'Save base details',
                  cancelTooltip: 'Cancel name edit',
                  onEdit: () => _beginInlineEdit(
                    editingName: true,
                    controller: _baseNameController,
                    focusNode: _baseNameFocusNode,
                  ),
                  onConfirm: _saveBaseDetails,
                  onCancel: _cancelBaseNameEdit,
                ),
              ),
            ),
        ],
      ),
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: theme.textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
    );
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
      _selectingMoveRoom = false;
      _moveAnchorStructureKey = null;
      _buildSelectionCell = cell;
    });
  }

  void _beginMoveRoomSelection() {
    if (_snapshot?.canManage != true || _snapshot!.structures.isEmpty) return;
    setState(() {
      _buildSelectionCell = null;
      _selectingMoveRoom = true;
      _moveAnchorStructureKey = null;
      _error = null;
    });
  }

  void _cancelMoveRoomSelection() {
    setState(() {
      _selectingMoveRoom = false;
      _moveAnchorStructureKey = null;
    });
  }

  void _selectMoveAnchor(UserBaseStructureData structure) {
    setState(() {
      _selectingMoveRoom = false;
      _moveAnchorStructureKey = structure.structureKey;
      _buildSelectionCell = null;
      _error = null;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          'Moving ${_friendlyStructureName(structure.structureKey)}. Tap an empty tile to place it.',
        ),
      ),
    );
  }

  Future<void> _moveSelectedRoomTo(_GridCell targetCell) async {
    final anchor = _moveAnchorStructure;
    if (anchor == null) return;
    if (anchor.gridX == targetCell.gridX && anchor.gridY == targetCell.gridY) {
      return;
    }

    setState(() {
      _busyStructureKey = anchor.structureKey;
      _error = null;
    });
    try {
      final nextSnapshot = await context.read<BaseService>().moveRooms(
        anchorStructureKey: anchor.structureKey,
        structureKeys: [anchor.structureKey],
        targetGridX: targetCell.gridX,
        targetGridY: targetCell.gridY,
      );
      if (!mounted) return;
      setState(() {
        _snapshot = nextSnapshot;
        _busyStructureKey = null;
        _moveAnchorStructureKey = null;
        _selectingMoveRoom = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            '${_friendlyStructureName(anchor.structureKey)} moved.',
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

  Future<void> _showRoomDetails(UserBaseStructureData structure) async {
    final definition = _definitionForKey(structure.structureKey);
    if (definition == null) return;
    final hearthRecoveryState = _activeDailyStateForKey(
      _hearthRecoveryStateKey,
    );
    final hearthRecoveryInfo = structure.structureKey == 'hearth'
        ? _hearthRecoveryInfo(hearthRecoveryState)
        : null;
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      useSafeArea: true,
      builder: (context) {
        final maxHeight = MediaQuery.of(context).size.height * 0.82;
        return ConstrainedBox(
          constraints: BoxConstraints(maxHeight: maxHeight),
          child: _RoomDetailsSheet(
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
            hearthRecoveryInfo: hearthRecoveryInfo,
            onUseHearth: structure.structureKey == 'hearth'
                ? () async {
                    Navigator.of(context).pop();
                    await _useHearth();
                  }
                : null,
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
      },
    );
  }

  BaseDailyEffectData? _activeDailyStateForKey(String stateKey) {
    final snapshot = _snapshot;
    if (snapshot == null) return null;
    for (final effect in snapshot.activeDailyEffects) {
      if (effect.stateKey == stateKey) {
        return effect;
      }
    }
    return null;
  }

  _HearthRecoveryInfo _hearthRecoveryInfo(BaseDailyEffectData? effect) {
    final level = _structureByKey['hearth']?.level ?? 0;
    if (effect == null) {
      return _HearthRecoveryInfo(
        availableNow: true,
        hearthLevel: level,
        statusesApplied: 0,
        removedWounded: false,
      );
    }
    final state = effect.state;
    final nextAvailableAt = _parseDateTime(state['nextAvailableAt']);
    final lastUsedAt = _parseDateTime(state['usedAt']);
    final statusCount = _statusCount(state['statusesApplied']);
    final removedWounded = _hasStatusNamed(state['statusesRemoved'], 'Wounded');
    return _HearthRecoveryInfo(
      availableNow: nextAvailableAt == null
          ? false
          : nextAvailableAt.isBefore(DateTime.now()),
      nextAvailableAt: nextAvailableAt,
      lastUsedAt: lastUsedAt,
      hearthLevel: (state['hearthLevel'] as num?)?.toInt() ?? level,
      statusesApplied: statusCount,
      removedWounded: removedWounded,
    );
  }

  DateTime? _parseDateTime(Object? value) {
    if (value == null) return null;
    return DateTime.tryParse(value.toString())?.toLocal();
  }

  int _statusCount(Object? value) {
    if (value is List) {
      return value.length;
    }
    return 0;
  }

  bool _hasStatusNamed(Object? value, String name) {
    if (value is! List) return false;
    final target = name.trim().toLowerCase();
    for (final entry in value) {
      if (entry.toString().trim().toLowerCase() == target) {
        return true;
      }
    }
    return false;
  }

  String _formatHearthAvailability(DateTime? value) {
    if (value == null) {
      return 'Available tomorrow';
    }
    final local = value.toLocal();
    final hour = local.hour % 12 == 0 ? 12 : local.hour % 12;
    final minute = local.minute.toString().padLeft(2, '0');
    final suffix = local.hour >= 12 ? 'PM' : 'AM';
    return 'Available ${local.month}/${local.day} at $hour:$minute $suffix';
  }

  Future<void> _useHearth() async {
    if (_busyStructureKey != null) return;
    setState(() {
      _busyStructureKey = 'hearth';
      _error = null;
    });
    try {
      final response = await context.read<BaseService>().useHearth();
      if (!mounted) return;
      final nextSnapshot = BaseProgressionSnapshot.fromJson(response);
      final healthRestored = (response['healthRestored'] as num?)?.toInt() ?? 0;
      final manaRestored = (response['manaRestored'] as num?)?.toInt() ?? 0;
      final currentHealth = (response['currentHealth'] as num?)?.toInt();
      final currentMana = (response['currentMana'] as num?)?.toInt();
      final statusesApplied = _statusCount(response['statusesApplied']);
      final removedWounded = _hasStatusNamed(
        response['statusesRemoved'],
        'Wounded',
      );
      final statsProvider = context.read<CharacterStatsProvider>();
      if (currentHealth != null && currentMana != null) {
        await statsProvider.setHealthAndManaTo(
          health: currentHealth,
          mana: currentMana,
        );
      }
      await statsProvider.refresh(silent: true);
      if (!mounted) return;
      setState(() {
        _snapshot = nextSnapshot;
        _busyStructureKey = null;
      });
      _notifyHeaderChanged();
      await _notifyTutorialProgressChanged();
      if (!mounted) return;
      final statusFragments = <String>[];
      if (removedWounded) {
        statusFragments.add('cleared Wounded');
      }
      if (statusesApplied > 0) {
        statusFragments.add(
          'gained $statusesApplied blessing${statusesApplied == 1 ? '' : 's'}',
        );
      }
      final statusText = statusFragments.isEmpty
          ? ''
          : ', ${statusFragments.join(' and ')}';
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Recovered $healthRestored health and $manaRestored mana$statusText.',
          ),
        ),
      );
    } on DioException catch (error) {
      if (!mounted) return;
      final data = error.response?.data;
      final payload = data is Map
          ? Map<String, dynamic>.from(data)
          : const <String, dynamic>{};
      final nextAvailableAt = _parseDateTime(payload['nextAvailableAt']);
      final serverMessage =
          payload['error']?.toString().trim().isNotEmpty == true
          ? payload['error'].toString().trim()
          : null;
      setState(() {
        _busyStructureKey = null;
        if (payload.isNotEmpty && _snapshot != null) {
          final effects =
              List<BaseDailyEffectData>.from(
                _snapshot!.activeDailyEffects.where(
                  (effect) => effect.stateKey != _hearthRecoveryStateKey,
                ),
              )..add(
                BaseDailyEffectData(
                  stateKey: _hearthRecoveryStateKey,
                  state: <String, dynamic>{
                    'usedAt': payload['lastUsedAt'],
                    'nextAvailableAt': payload['nextAvailableAt'],
                    'hearthLevel': (_structureByKey['hearth']?.level ?? 0),
                    'statusesApplied': const <dynamic>[],
                    'statusesRemoved': const <dynamic>[],
                  },
                ),
              );
          _snapshot = BaseProgressionSnapshot(
            base: _snapshot!.base,
            resources: _snapshot!.resources,
            structures: _snapshot!.structures,
            activeDailyEffects: effects,
            grassTileUrls: _snapshot!.grassTileUrls,
            canManage: _snapshot!.canManage,
          );
        }
        _error = serverMessage ?? error.toString();
      });
      final message =
          error.response?.statusCode == 429 && nextAvailableAt != null
          ? 'Hearth recovery was already used today. ${_formatHearthAvailability(nextAvailableAt)}.'
          : serverMessage ?? error.toString();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
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

  Widget _buildBaseStats(ThemeData theme) {
    final snapshot = _snapshot;
    if (snapshot == null) return const SizedBox.shrink();

    final roomCount = snapshot.structures.length;
    final highestLevel = snapshot.structures.isEmpty
        ? 0
        : snapshot.structures
              .map((structure) => structure.level)
              .reduce((a, b) => a > b ? a : b);
    final expansionSites = _adjacentBuildCellKeys.length;
    final activeEffects = snapshot.activeDailyEffects.length;

    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Wrap(
        spacing: 10,
        runSpacing: 10,
        children: [
          _BaseStatCard(
            icon: Icons.home_work_outlined,
            label: 'Rooms',
            value: '$roomCount',
          ),
          _BaseStatCard(
            icon: Icons.north_east_outlined,
            label: 'Expansion Sites',
            value: '$expansionSites',
          ),
          _BaseStatCard(
            icon: Icons.stacked_line_chart,
            label: 'Highest Level',
            value: highestLevel > 0 ? 'Lv $highestLevel' : 'None',
          ),
          _BaseStatCard(
            icon: Icons.auto_awesome_outlined,
            label: 'Active Effects',
            value: '$activeEffects',
          ),
        ],
      ),
    );
  }

  Widget _buildMoveControls(ThemeData theme) {
    final snapshot = _snapshot;
    final anchor = _moveAnchorStructure;
    if (snapshot == null ||
        !snapshot.canManage ||
        snapshot.structures.isEmpty) {
      return const SizedBox.shrink();
    }

    String title;
    String description;
    if (_selectingMoveRoom) {
      title = 'Select Room To Move';
      description = 'Tap the room you want to reposition on the grid.';
    } else if (anchor != null) {
      title = 'Place ${_friendlyStructureName(anchor.structureKey)}';
      description =
          'Tap an empty tile to move it there, or choose a different room.';
    } else {
      title = 'Move Room';
      description = 'Reposition one room on your base grid.';
    }

    return Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: theme.colorScheme.surface.withValues(alpha: 0.78),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: theme.colorScheme.outlineVariant),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: theme.textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    description,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            if (_selectingMoveRoom || anchor != null)
              TextButton(
                onPressed: _busyStructureKey == null
                    ? _cancelMoveRoomSelection
                    : null,
                child: const Text('Cancel'),
              )
            else
              OutlinedButton(
                onPressed: _busyStructureKey == null
                    ? _beginMoveRoomSelection
                    : null,
                child: const Text('Move Room'),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildGrid(ThemeData theme) {
    final snapshot = _snapshot;
    if (snapshot == null) return const SizedBox.shrink();
    final moveAnchor = _moveAnchorStructure;
    final structuresByCell = <String, UserBaseStructureData>{};
    for (final structure in snapshot.structures) {
      structuresByCell['${structure.gridX}:${structure.gridY}'] = structure;
    }
    final adjacentBuildKeys = _adjacentBuildCellKeys;

    return LayoutBuilder(
      builder: (context, constraints) {
        const spacing = 0.0;
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
                            moveAnchor == null &&
                            !_selectingMoveRoom &&
                            structure == null &&
                            isAdjacentBuildCell,
                        isMoveAnchor:
                            moveAnchor?.structureKey == structure?.structureKey,
                        isMoveTarget:
                            moveAnchor != null &&
                            structure == null &&
                            !(_busyStructureKey != null),
                        canManage: snapshot.canManage,
                        onTap: () {
                          if (_busyStructureKey != null) {
                            return;
                          }
                          if (_selectingMoveRoom) {
                            if (structure != null) {
                              _selectMoveAnchor(structure);
                            }
                            return;
                          }
                          if (moveAnchor != null) {
                            if (structure == null) {
                              _moveSelectedRoomTo(cell);
                            } else {
                              _selectMoveAnchor(structure);
                            }
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

  Widget _buildTutorialObjectiveCard(ThemeData theme) {
    final tutorialStatus = _tutorialStatus;
    if (tutorialStatus == null || !tutorialStatus.isHearthStep) {
      return const SizedBox.shrink();
    }

    final hearthStructure = _structureByKey['hearth'];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 18),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: const Color(0xFFF9F1DC),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: const Color(0xFFD8B36B), width: 1.4),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Tutorial Objective',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF6A4A14),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                tutorialStatus.resolvedHearthObjectiveCopy,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: const Color(0xFF4C3824),
                  height: 1.35,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Open the hearth room and make a full recovery.',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: const Color(0xFF6A4A14),
                ),
              ),
              if (hearthStructure != null) ...[
                const SizedBox(height: 12),
                FilledButton.tonal(
                  onPressed: _busyStructureKey != null
                      ? null
                      : () async {
                          await _showRoomDetails(hearthStructure);
                        },
                  child: const Text('Open Hearth'),
                ),
              ],
            ],
          ),
        ),
      ],
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
            _buildBaseStats(theme),
            _buildMoveControls(theme),
            _buildGrid(theme),
            _buildTutorialObjectiveCard(theme),
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

class _InlineEditActions extends StatelessWidget {
  const _InlineEditActions({
    required this.editing,
    required this.saving,
    required this.editTooltip,
    required this.confirmTooltip,
    required this.cancelTooltip,
    required this.onEdit,
    required this.onConfirm,
    required this.onCancel,
  });

  final bool editing;
  final bool saving;
  final String editTooltip;
  final String confirmTooltip;
  final String cancelTooltip;
  final VoidCallback onEdit;
  final VoidCallback onConfirm;
  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    if (!editing) {
      return IconButton(
        tooltip: editTooltip,
        onPressed: saving ? null : onEdit,
        icon: const Icon(Icons.edit_outlined, size: 20),
        padding: EdgeInsets.zero,
        constraints: const BoxConstraints.tightFor(width: 28, height: 28),
        visualDensity: VisualDensity.compact,
      );
    }

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        IconButton(
          tooltip: cancelTooltip,
          onPressed: saving ? null : onCancel,
          icon: const Icon(Icons.close, size: 20),
          padding: EdgeInsets.zero,
          constraints: const BoxConstraints.tightFor(width: 28, height: 28),
          visualDensity: VisualDensity.compact,
        ),
        if (saving)
          const SizedBox(
            width: 28,
            height: 28,
            child: Center(
              child: SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            ),
          )
        else
          IconButton(
            tooltip: confirmTooltip,
            onPressed: onConfirm,
            icon: const Icon(Icons.check, size: 20),
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints.tightFor(width: 28, height: 28),
            visualDensity: VisualDensity.compact,
          ),
      ],
    );
  }
}

class _BaseStatCard extends StatelessWidget {
  const _BaseStatCard({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      constraints: const BoxConstraints(minWidth: 132),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface.withValues(alpha: 0.72),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 18, color: theme.colorScheme.primary),
          const SizedBox(width: 10),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                value,
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w800,
                ),
              ),
              Text(
                label,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ],
          ),
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
    required this.isMoveAnchor,
    required this.isMoveTarget,
    required this.canManage,
    required this.onTap,
  });

  final double tileSize;
  final String grassTileUrl;
  final UserBaseStructureData? structure;
  final BaseStructureDefinitionData? definition;
  final BaseStructureLevelVisualData? visual;
  final bool showPlus;
  final bool isMoveAnchor;
  final bool isMoveTarget;
  final bool canManage;
  final VoidCallback onTap;

  String get _roomImageUrl {
    if (visual != null && visual!.topDownImageUrl.trim().isNotEmpty) {
      return visual!.topDownImageUrl.trim();
    }
    if (visual != null && visual!.topDownThumbnailUrl.trim().isNotEmpty) {
      return visual!.topDownThumbnailUrl.trim();
    }
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
    final hasRoom = structure != null && definition != null;
    final theme = Theme.of(context);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
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
            if (showPlus)
              Center(
                child: Icon(
                  Icons.add,
                  color: _buildSlotPlusColor,
                  size: tileSize * 0.28,
                ),
              ),
            if (isMoveTarget)
              Container(
                color: theme.colorScheme.primary.withValues(alpha: 0.12),
              ),
            if (hasRoom)
              Container(
                decoration: BoxDecoration(
                  border: Border.all(
                    color: isMoveAnchor
                        ? theme.colorScheme.primary
                        : _roomBorderColor,
                    width: isMoveAnchor ? 3 : 2,
                  ),
                ),
              ),
          ],
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

class _RoomDetailsSheet extends StatefulWidget {
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
    required this.hearthRecoveryInfo,
    required this.onUseHearth,
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
  final _HearthRecoveryInfo? hearthRecoveryInfo;
  final Future<void> Function()? onUseHearth;
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

  @override
  State<_RoomDetailsSheet> createState() => _RoomDetailsSheetState();
}

class _RoomDetailsSheetState extends State<_RoomDetailsSheet> {
  Timer? _cooldownTicker;
  BaseCraftingRecipesResponse? _craftingData;
  bool _loadingCrafting = false;
  String? _craftingError;
  String? _craftingRecipeId;
  List<Zone> _chaosZones = const [];
  List<ZoneGenre> _chaosGenres = const [];
  List<InventoryItem> _chaosInventoryItems = const [];
  List<OwnedInventoryItem> _chaosOwnedItems = const [];
  bool _loadingChaosEngine = false;
  String? _chaosEngineError;
  bool _usingChaosEngine = false;
  String? _selectedChaosZoneId;
  String? _selectedChaosGenreId;

  String? get _craftingStation {
    switch (widget.structure.structureKey) {
      case 'alchemy_lab':
        return 'alchemy';
      case 'workshop':
        return 'workshop';
      default:
        return null;
    }
  }

  bool get _isChaosEngineRoom =>
      widget.structure.structureKey == 'chaos_engine';

  int? get _chaosEngineRequiredInventoryItemId {
    final raw = widget.definition.effectConfig['requiredInventoryItemId'];
    if (raw is num) {
      final value = raw.toInt();
      return value > 0 ? value : null;
    }
    if (raw is String) {
      final value = int.tryParse(raw.trim());
      return value != null && value > 0 ? value : null;
    }
    return null;
  }

  InventoryItem? get _chaosEngineRequiredInventoryItem {
    final requiredInventoryItemId = _chaosEngineRequiredInventoryItemId;
    if (requiredInventoryItemId == null) {
      return null;
    }
    for (final item in _chaosInventoryItems) {
      if (item.id == requiredInventoryItemId) {
        return item;
      }
    }
    return null;
  }

  int get _chaosEngineOwnedQuantity {
    final requiredInventoryItemId = _chaosEngineRequiredInventoryItemId;
    if (requiredInventoryItemId == null) {
      return 0;
    }
    for (final item in _chaosOwnedItems) {
      if (item.inventoryItemId == requiredInventoryItemId) {
        return item.quantity;
      }
    }
    return 0;
  }

  Zone? get _selectedChaosZone {
    final zoneId = _selectedChaosZoneId;
    if (zoneId == null || zoneId.isEmpty) {
      return null;
    }
    for (final zone in _chaosZones) {
      if (zone.id == zoneId) {
        return zone;
      }
    }
    return null;
  }

  @override
  void initState() {
    super.initState();
    _syncCooldownTicker();
    _loadCraftingIfNeeded();
    _loadChaosEngineIfNeeded();
  }

  @override
  void didUpdateWidget(covariant _RoomDetailsSheet oldWidget) {
    super.didUpdateWidget(oldWidget);
    final current = widget.hearthRecoveryInfo;
    final previous = oldWidget.hearthRecoveryInfo;
    if (current?.availableNow != previous?.availableNow ||
        current?.nextAvailableAt != previous?.nextAvailableAt) {
      _syncCooldownTicker();
    }
    if (widget.structure.structureKey != oldWidget.structure.structureKey ||
        widget.structure.level != oldWidget.structure.level ||
        widget.canManage != oldWidget.canManage) {
      _loadCraftingIfNeeded(force: true);
      _loadChaosEngineIfNeeded(force: true);
    }
  }

  @override
  void dispose() {
    _cooldownTicker?.cancel();
    super.dispose();
  }

  void _syncCooldownTicker() {
    final hearthInfo = widget.hearthRecoveryInfo;
    final shouldTick =
        hearthInfo != null &&
        !hearthInfo.availableNow &&
        hearthInfo.nextAvailableAt != null;
    if (!shouldTick) {
      _cooldownTicker?.cancel();
      _cooldownTicker = null;
      return;
    }
    _cooldownTicker ??= Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      final nextAvailableAt = widget.hearthRecoveryInfo?.nextAvailableAt;
      if (nextAvailableAt == null || !nextAvailableAt.isAfter(DateTime.now())) {
        _cooldownTicker?.cancel();
        _cooldownTicker = null;
      }
      setState(() {});
    });
  }

  Future<void> _loadCraftingIfNeeded({bool force = false}) async {
    final station = _craftingStation;
    if (station == null || !widget.canManage) {
      if (!mounted) return;
      setState(() {
        _craftingData = null;
        _loadingCrafting = false;
        _craftingError = null;
      });
      return;
    }
    if (!force && (_loadingCrafting || _craftingData != null)) {
      return;
    }
    setState(() {
      _loadingCrafting = true;
      _craftingError = null;
    });
    try {
      final data = await context.read<BaseService>().getCraftingRecipes(
        station,
      );
      if (!mounted) return;
      setState(() {
        _craftingData = data;
        _loadingCrafting = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingCrafting = false;
        _craftingError = e.toString();
      });
    }
  }

  Future<void> _loadChaosEngineIfNeeded({bool force = false}) async {
    if (!_isChaosEngineRoom || !widget.canManage) {
      if (!mounted) return;
      setState(() {
        _chaosZones = const [];
        _chaosGenres = const [];
        _chaosInventoryItems = const [];
        _chaosOwnedItems = const [];
        _loadingChaosEngine = false;
        _chaosEngineError = null;
      });
      return;
    }
    if (!force &&
        (_loadingChaosEngine ||
            _chaosZones.isNotEmpty ||
            _chaosGenres.isNotEmpty ||
            _chaosInventoryItems.isNotEmpty)) {
      return;
    }

    setState(() {
      _loadingChaosEngine = true;
      _chaosEngineError = null;
    });
    try {
      final results = await Future.wait([
        context.read<PoiService>().getZones(),
        context.read<BaseService>().getZoneGenres(),
        context.read<InventoryService>().getInventoryItems(preferCache: true),
        context.read<InventoryService>().getOwnedInventoryItems(
          preferCache: true,
        ),
      ]);

      final zones = List<Zone>.from(results[0] as List<Zone>)
        ..sort(
          (left, right) =>
              left.name.toLowerCase().compareTo(right.name.toLowerCase()),
        );
      final genres = List<ZoneGenre>.from(results[1] as List<ZoneGenre>)
        ..sort((left, right) {
          final sortCompare = left.sortOrder.compareTo(right.sortOrder);
          if (sortCompare != 0) {
            return sortCompare;
          }
          return left.name.toLowerCase().compareTo(right.name.toLowerCase());
        });
      final inventoryItems = List<InventoryItem>.from(
        results[2] as List<InventoryItem>,
      );
      final ownedItems = List<OwnedInventoryItem>.from(
        results[3] as List<OwnedInventoryItem>,
      );

      final nextZoneId = zones.any((zone) => zone.id == _selectedChaosZoneId)
          ? _selectedChaosZoneId
          : (zones.isNotEmpty ? zones.first.id : null);
      final nextGenreId =
          genres.any((genre) => genre.id == _selectedChaosGenreId)
          ? _selectedChaosGenreId
          : (genres.isNotEmpty ? genres.first.id : null);

      if (!mounted) return;
      setState(() {
        _chaosZones = zones;
        _chaosGenres = genres;
        _chaosInventoryItems = inventoryItems;
        _chaosOwnedItems = ownedItems;
        _selectedChaosZoneId = nextZoneId;
        _selectedChaosGenreId = nextGenreId;
        _loadingChaosEngine = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingChaosEngine = false;
        _chaosEngineError = e.toString();
      });
    }
  }

  Future<void> _craftRecipe(BaseCraftingRecipeData recipe) async {
    final station = _craftingStation;
    if (station == null || _craftingRecipeId != null) return;
    setState(() {
      _craftingRecipeId = recipe.id;
      _craftingError = null;
    });
    try {
      final response = await context.read<BaseService>().craftRecipe(
        station,
        recipe.id,
      );
      if (!mounted) return;
      await _loadCraftingIfNeeded(force: true);
      if (!mounted) return;
      final craftedName = response['craftedItem'] is Map
          ? ((response['craftedItem'] as Map)['name']?.toString() ?? '')
          : '';
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            craftedName.isNotEmpty ? 'Crafted $craftedName.' : 'Crafted item.',
          ),
        ),
      );
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _craftingError = e.toString();
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(e.toString())));
    } finally {
      if (mounted) {
        setState(() {
          _craftingRecipeId = null;
        });
      }
    }
  }

  Future<void> _useChaosEngine() async {
    final zoneId = _selectedChaosZoneId;
    final genreId = _selectedChaosGenreId;
    if (_usingChaosEngine || zoneId == null || genreId == null) {
      return;
    }
    final zoneProvider = context.read<ZoneProvider>();
    final inventoryService = context.read<InventoryService>();

    setState(() {
      _usingChaosEngine = true;
      _chaosEngineError = null;
    });
    try {
      final response = await context.read<BaseService>().useChaosEngine(
        zoneId: zoneId,
        genreId: genreId,
      );
      final rawZone = response['zone'];
      Zone? updatedZone;
      if (rawZone is Map<String, dynamic>) {
        updatedZone = Zone.fromJson(rawZone);
      } else if (rawZone is Map) {
        updatedZone = Zone.fromJson(Map<String, dynamic>.from(rawZone));
      }

      late final List<Zone> nextZones;
      if (updatedZone == null) {
        nextZones = _chaosZones;
      } else {
        final resolvedUpdatedZone = updatedZone;
        nextZones = _chaosZones
            .map(
              (zone) => zone.id == resolvedUpdatedZone.id
                  ? resolvedUpdatedZone
                  : zone,
            )
            .toList(growable: false);
      }
      zoneProvider.setZones(nextZones);

      final refreshedOwnedItems = await inventoryService
          .refreshOwnedInventoryItems();

      if (!mounted) return;
      setState(() {
        _chaosZones = nextZones;
        if (refreshedOwnedItems != null) {
          _chaosOwnedItems = refreshedOwnedItems;
        }
        _usingChaosEngine = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            response['message']?.toString() ?? 'Chaos Engine activated.',
          ),
        ),
      );
    } on DioException catch (error) {
      final responseData = error.response?.data;
      final message = responseData is Map
          ? responseData['error']?.toString()
          : error.message;
      if (!mounted) return;
      setState(() {
        _usingChaosEngine = false;
        _chaosEngineError = message ?? 'Failed to activate the Chaos Engine.';
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(_chaosEngineError!)));
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _usingChaosEngine = false;
        _chaosEngineError = e.toString();
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(_chaosEngineError!)));
    }
  }

  double _cooldownProgress(_HearthRecoveryInfo hearthInfo, Duration remaining) {
    final totalDuration =
        hearthInfo.lastUsedAt != null && hearthInfo.nextAvailableAt != null
        ? hearthInfo.nextAvailableAt!.difference(hearthInfo.lastUsedAt!)
        : _hearthRecoveryCooldownDuration;
    final totalSeconds = totalDuration.inSeconds > 0
        ? totalDuration.inSeconds
        : _hearthRecoveryCooldownDuration.inSeconds;
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

  Widget _buildCooldownCard(
    BuildContext context,
    _HearthRecoveryInfo hearthInfo,
    Duration remaining,
    DateTime nextAvailableAt,
  ) {
    final theme = Theme.of(context);
    final progress = _cooldownProgress(hearthInfo, remaining);
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
            'Hearth Recharging',
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

  Widget _buildCostBadge(BuildContext context, BaseStructureCostData cost) {
    final theme = Theme.of(context);
    final available = widget.resourceAmounts[cost.resourceKey] ?? 0;
    final hasEnough = available >= cost.amount;
    final accentColor = hasEnough
        ? widget.materialAccentColor(cost.resourceKey)
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
          Icon(
            widget.materialIcon(cost.resourceKey),
            size: 18,
            color: accentColor,
          ),
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

  Widget _buildIngredientChip(
    BuildContext context,
    BaseCraftingIngredientData ingredient,
  ) {
    return InventoryRequirementChip(
      item: ingredient.item,
      quantity: ingredient.quantity,
      ownedQuantity: ingredient.ownedQuantity,
    );
  }

  Widget _buildCraftingCard(
    BuildContext context,
    BaseCraftingRecipeData recipe,
  ) {
    final theme = Theme.of(context);
    final imageUrl = recipe.resultItem.imageUrl.trim();
    final busy = _craftingRecipeId == recipe.id;
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        color: theme.colorScheme.surfaceContainerHighest,
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: SizedBox(
                  width: 60,
                  height: 60,
                  child: imageUrl.isNotEmpty
                      ? Image.network(
                          imageUrl,
                          fit: BoxFit.cover,
                          errorBuilder: (_, _, _) =>
                              _RoomFallbackLabel(title: recipe.resultItem.name),
                        )
                      : _RoomFallbackLabel(title: recipe.resultItem.name),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      recipe.resultItem.name,
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Tier ${recipe.tier} ${recipe.isPublic ? 'Public' : 'Private'} Recipe',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: recipe.ingredients
                .map((ingredient) => _buildIngredientChip(context, ingredient))
                .toList(growable: false),
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: FilledButton.tonal(
              onPressed: busy || !recipe.canCraft
                  ? null
                  : () => _craftRecipe(recipe),
              child: busy
                  ? const SizedBox(
                      height: 18,
                      width: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(recipe.canCraft ? 'Craft' : 'Need Ingredients'),
            ),
          ),
        ],
      ),
    );
  }

  List<ZoneGenreScore> _sortedZoneGenreScores(Zone? zone) {
    if (zone == null || zone.genreScores.isEmpty) {
      return const <ZoneGenreScore>[];
    }
    final sorted = List<ZoneGenreScore>.from(zone.genreScores);
    sorted.sort((left, right) {
      final scoreCompare = right.score.compareTo(left.score);
      if (scoreCompare != 0) {
        return scoreCompare;
      }
      final sortCompare = left.genre.sortOrder.compareTo(right.genre.sortOrder);
      if (sortCompare != 0) {
        return sortCompare;
      }
      return left.genre.name.toLowerCase().compareTo(
        right.genre.name.toLowerCase(),
      );
    });
    return sorted;
  }

  Widget _buildChaosEnginePanel(BuildContext context) {
    final theme = Theme.of(context);
    if (!widget.canManage) {
      return const SizedBox.shrink();
    }
    if (_loadingChaosEngine) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_chaosEngineError != null) {
      return Text(
        _chaosEngineError!,
        style: theme.textTheme.bodySmall?.copyWith(
          color: theme.colorScheme.error,
        ),
      );
    }

    final requiredItem = _chaosEngineRequiredInventoryItem;
    final selectedZone = _selectedChaosZone;
    final sortedScores = _sortedZoneGenreScores(selectedZone);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Chaos Engine',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          if (requiredItem == null)
            Text(
              'This room has not been configured with a fuel item yet.',
              style: theme.textTheme.bodyMedium,
            )
          else ...[
            Text(
              'Fuel item: ${requiredItem.name}',
              style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'Owned: $_chaosEngineOwnedQuantity',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              key: ValueKey('chaos-zone-${_selectedChaosZoneId ?? ''}'),
              initialValue: _selectedChaosZoneId,
              decoration: const InputDecoration(
                labelText: 'Target Zone',
                border: OutlineInputBorder(),
              ),
              items: _chaosZones
                  .map(
                    (zone) => DropdownMenuItem<String>(
                      value: zone.id,
                      child: Text(zone.name),
                    ),
                  )
                  .toList(growable: false),
              onChanged: (value) {
                setState(() {
                  _selectedChaosZoneId = value;
                });
              },
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              key: ValueKey('chaos-genre-${_selectedChaosGenreId ?? ''}'),
              initialValue: _selectedChaosGenreId,
              decoration: const InputDecoration(
                labelText: 'Genre to Increase',
                border: OutlineInputBorder(),
              ),
              items: _chaosGenres
                  .map(
                    (genre) => DropdownMenuItem<String>(
                      value: genre.id,
                      child: Text(genre.name),
                    ),
                  )
                  .toList(growable: false),
              onChanged: (value) {
                setState(() {
                  _selectedChaosGenreId = value;
                });
              },
            ),
            if (selectedZone != null) ...[
              const SizedBox(height: 12),
              Text(
                'Current zone genre scores',
                style: theme.textTheme.bodySmall?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 8),
              if (sortedScores.isEmpty)
                Text(
                  'No genre alignment yet.',
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                )
              else
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: sortedScores
                      .map(
                        (score) => Chip(
                          label: Text('${score.genre.name} ${score.score}'),
                          visualDensity: VisualDensity.compact,
                        ),
                      )
                      .toList(growable: false),
                ),
            ],
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: FilledButton.tonal(
                onPressed:
                    _usingChaosEngine ||
                        _selectedChaosZoneId == null ||
                        _selectedChaosGenreId == null ||
                        _chaosEngineOwnedQuantity <= 0
                    ? null
                    : _useChaosEngine,
                child: _usingChaosEngine
                    ? const SizedBox(
                        height: 18,
                        width: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(
                        _chaosEngineOwnedQuantity > 0
                            ? 'Spend 1 Item to Shift the Zone'
                            : 'Need Fuel Item',
                      ),
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
    final hearthInfo = widget.hearthRecoveryInfo;
    final nextAvailableAt = hearthInfo?.nextAvailableAt;
    final remaining = nextAvailableAt?.difference(DateTime.now());
    final cooldownActive =
        hearthInfo != null &&
        !hearthInfo.availableNow &&
        remaining != null &&
        !remaining.isNegative;
    return SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.definition.name,
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'Level ${widget.structure.level} of ${widget.definition.maxLevel}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 16),
            if (widget._visualUrl.isNotEmpty)
              ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: AspectRatio(
                  aspectRatio: 1.4,
                  child: Image.network(
                    widget._visualUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (_, _, _) =>
                        _RoomFallbackLabel(title: widget.definition.name),
                  ),
                ),
              ),
            if (widget._visualUrl.isEmpty) ...[
              ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: SizedBox(
                  height: 180,
                  width: double.infinity,
                  child: _RoomFallbackLabel(title: widget.definition.name),
                ),
              ),
            ],
            const SizedBox(height: 16),
            Text(
              widget.definition.description,
              style: theme.textTheme.bodyMedium?.copyWith(height: 1.4),
            ),
            if (hearthInfo != null) ...[
              const SizedBox(height: 16),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (hearthInfo.availableNow)
                      Text(
                        hearthInfo.hearthLevel >= 3
                            ? 'Fully recover once per day and gain the rank 2 and rank 3 hearth blessings.'
                            : hearthInfo.hearthLevel >= 2
                            ? 'Fully recover once per day and gain the rank 2 hearth blessings.'
                            : 'Fully recover once per day at this hearth.',
                        style: theme.textTheme.bodyMedium,
                      ),
                    if (cooldownActive) ...[
                      if (hearthInfo.availableNow) const SizedBox(height: 14),
                      _buildCooldownCard(
                        context,
                        hearthInfo,
                        remaining,
                        nextAvailableAt!,
                      ),
                    ],
                    if (hearthInfo.statusesApplied > 0) ...[
                      const SizedBox(height: 6),
                      Text(
                        '${hearthInfo.statusesApplied} blessing${hearthInfo.statusesApplied == 1 ? '' : 's'} applied on the last use.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    if (hearthInfo.removedWounded) ...[
                      const SizedBox(height: 6),
                      Text(
                        'Wounded was cleared on the last use.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    const SizedBox(height: 12),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.tonal(
                        onPressed:
                            widget.isBusy ||
                                !hearthInfo.availableNow ||
                                widget.onUseHearth == null
                            ? null
                            : () => widget.onUseHearth!.call(),
                        child: widget.isBusy
                            ? const SizedBox(
                                height: 18,
                                width: 18,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : Text(
                                hearthInfo.availableNow
                                    ? 'Make Full Recovery'
                                    : hearthInfo.nextAvailableAt != null
                                    ? hearthInfo.formattedNextAvailableLabel
                                    : 'Available Tomorrow',
                              ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
            if (_isChaosEngineRoom && widget.canManage) ...[
              const SizedBox(height: 16),
              _buildChaosEnginePanel(context),
            ],
            if (_craftingStation != null && widget.canManage) ...[
              const SizedBox(height: 16),
              Text(
                'Crafting',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 8),
              if (_loadingCrafting)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 12),
                  child: Center(child: CircularProgressIndicator()),
                )
              else if (_craftingError != null)
                Text(
                  _craftingError!,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.error,
                  ),
                )
              else if ((_craftingData?.recipes.isEmpty ?? true))
                Text(
                  'No recipes are available here yet.',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                )
              else
                ..._craftingData!.recipes.map(
                  (recipe) => _buildCraftingCard(context, recipe),
                ),
            ],
            if (widget.costs.isNotEmpty) ...[
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
                children: widget.costs
                    .map((cost) => _buildCostBadge(context, cost))
                    .toList(),
              ),
              if (widget.canManage && !widget.isMaxed) ...[
                const SizedBox(height: 14),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    onPressed:
                        widget.isBusy ||
                            !widget.canAffordUpgrade ||
                            widget.onUpgrade == null
                        ? null
                        : () => widget.onUpgrade!.call(),
                    child: widget.isBusy
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(
                            widget.canAffordUpgrade
                                ? 'Upgrade Room'
                                : 'Need More Materials',
                          ),
                  ),
                ),
              ],
            ],
            const SizedBox(height: 18),
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

class _HearthRecoveryInfo {
  const _HearthRecoveryInfo({
    required this.availableNow,
    required this.hearthLevel,
    required this.statusesApplied,
    required this.removedWounded,
    this.nextAvailableAt,
    this.lastUsedAt,
  });

  final bool availableNow;
  final DateTime? nextAvailableAt;
  final DateTime? lastUsedAt;
  final int hearthLevel;
  final int statusesApplied;
  final bool removedWounded;

  String get formattedNextAvailableAt {
    final value = nextAvailableAt;
    if (value == null) {
      return 'tomorrow';
    }
    return _formatCompactDateTime(value);
  }

  String get formattedLastUsedAt {
    final value = lastUsedAt;
    if (value == null) {
      return 'today';
    }
    return _formatCompactDateTime(value);
  }

  String get formattedNextAvailableLabel {
    final value = nextAvailableAt;
    if (value == null) {
      return 'Available Tomorrow';
    }
    return 'Ready ${_formatCompactDateTime(value)}';
  }
}

String _formatCompactDateTime(DateTime value) {
  final local = value.toLocal();
  final hour = local.hour % 12 == 0 ? 12 : local.hour % 12;
  final minute = local.minute.toString().padLeft(2, '0');
  final suffix = local.hour >= 12 ? 'PM' : 'AM';
  return '${local.month}/${local.day} $hour:$minute $suffix';
}
