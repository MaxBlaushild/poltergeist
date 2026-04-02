import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/character.dart';
import '../models/character_action.dart';
import '../models/location.dart';
import '../models/quest.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/completed_task_provider.dart';
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/quest_log_provider.dart';
import '../services/poi_service.dart';
import '../widgets/paper_texture.dart';
import 'rpg_dialogue_modal.dart';

const _unlockRadiusMeters = kProximityUnlockRadiusMeters;
const _placeholderImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png';

class CharacterPanel extends StatefulWidget {
  const CharacterPanel({
    super.key,
    required this.character,
    required this.hasDiscovered,
    required this.onClose,
    this.onUnlocked,
    this.onQuestAccepted,
    this.onStartDialogue,
    this.onStartShop,
  });

  final Character character;
  final bool hasDiscovered;
  final VoidCallback onClose;
  final Future<void> Function()? onUnlocked;
  final VoidCallback? onQuestAccepted;
  final void Function(BuildContext, Character, CharacterAction)?
  onStartDialogue;
  final void Function(BuildContext, Character, CharacterAction)? onStartShop;

  @override
  State<CharacterPanel> createState() => _CharacterPanelState();
}

class _CharacterPanelState extends State<CharacterPanel> {
  List<CharacterAction> _actions = [];
  bool _loadingActions = true;
  bool _unlocking = false;
  bool _justUnlocked = false;
  String? _unlockError;
  bool _acceptingQuest = false;
  bool _turningInQuest = false;

  @override
  void initState() {
    super.initState();
    _loadActions();
  }

  Future<void> _loadActions() async {
    setState(() => _loadingActions = true);
    try {
      final svc = context.read<PoiService>();
      _actions = await svc.getCharacterActions(widget.character.id);
    } catch (_) {
      _actions = [];
    }
    debugPrint(
      'CharacterPanel: loaded ${_actions.length} actions for ${widget.character.id}',
    );
    final hasTalkAction = _actions.any((action) => action.actionType == 'talk');
    if (!hasTalkAction) {
      final fallbackTalk = CharacterAction(
        id: 'local-talk-${widget.character.id}',
        createdAt: DateTime.now().toIso8601String(),
        updatedAt: DateTime.now().toIso8601String(),
        characterId: widget.character.id,
        actionType: 'talk',
        dialogue: const [
          DialogueMessage(speaker: 'character', text: '...', order: 0),
        ],
      );
      _actions = [fallbackTalk, ..._actions];
    }
    if (mounted) setState(() => _loadingActions = false);
  }

  CharacterAction? _firstActionOfType(String type) {
    for (final action in _actions) {
      if (action.actionType == type) return action;
    }
    return null;
  }

  CharacterAction? _firstActionOfTypes(List<String> types) {
    for (final action in _actions) {
      if (types.contains(action.actionType)) return action;
    }
    return null;
  }

  Quest? _questForAction(CharacterAction action) {
    final questId = action.questId;
    if (questId == null || questId.isEmpty) return null;
    final quests = context.read<QuestLogProvider>().quests;
    try {
      return quests.firstWhere((q) => q.id == questId);
    } catch (_) {
      return null;
    }
  }

  double _averageNodeDifficulty(Quest quest) {
    final nodeDifficulties = <double>[];
    for (final node in quest.nodes) {
      final difficulty = node.objective?.difficulty ?? 0;
      if (difficulty <= 0) continue;
      nodeDifficulties.add(difficulty.toDouble());
    }
    if (nodeDifficulties.isEmpty) return 0;
    final total = nodeDifficulties.reduce((a, b) => a + b);
    return total / nodeDifficulties.length;
  }

  int _roundToNearestFive(double value) {
    if (value.isNaN || value.isInfinite) return 0;
    return (value / 5).round() * 5;
  }

  Set<String> _questStatTags(Quest quest) {
    final tags = <String>{};
    for (final node in quest.nodes) {
      for (final tag in node.objective?.statTags ?? const <String>[]) {
        final normalized = tag.trim().toLowerCase();
        if (normalized.isNotEmpty) {
          tags.add(normalized);
        }
      }
    }
    return tags;
  }

  double _averageStatValue(Map<String, int> stats, Set<String> tags) {
    if (stats.isEmpty) return 0;
    if (tags.isEmpty) {
      final values = stats.values;
      final total = values.fold<int>(0, (sum, value) => sum + value);
      return total / values.length;
    }
    var total = 0;
    var count = 0;
    for (final tag in tags) {
      if (!stats.containsKey(tag)) continue;
      total += stats[tag] ?? 0;
      count += 1;
    }
    if (count == 0) return 0;
    return total / count;
  }

  Color _difficultyColor(double statAverage, int difficulty) {
    if (statAverage > difficulty) {
      return const Color(0xFFC9C2B2);
    }
    if (statAverage > difficulty - 25) {
      return const Color(0xFF6F8F5E);
    }
    if (statAverage > difficulty - 50) {
      return const Color(0xFFC89A3A);
    }
    return const Color(0xFFA35B4B);
  }

  Widget _buildQuestDifficultySummary(BuildContext context, Quest quest) {
    final statsProvider = context.watch<CharacterStatsProvider>();
    final avgDifficulty = _averageNodeDifficulty(quest);
    final roundedDifficulty = _roundToNearestFive(avgDifficulty);
    final tags = _questStatTags(quest);
    final statAverage = _averageStatValue(statsProvider.stats, tags);
    final difficultyColor = _difficultyColor(statAverage, roundedDifficulty);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Theme.of(context).colorScheme.outlineVariant),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              'Difficulty',
              style: Theme.of(
                context,
              ).textTheme.labelLarge?.copyWith(fontWeight: FontWeight.w600),
            ),
          ),
          Text(
            '$roundedDifficulty',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
              color: difficultyColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuestDifficultySummaryFromMetadata(
    BuildContext context,
    double avgDifficulty,
    List<String> tags,
  ) {
    final statsProvider = context.watch<CharacterStatsProvider>();
    final roundedDifficulty = _roundToNearestFive(avgDifficulty);
    final tagSet = tags
        .map((tag) => tag.trim().toLowerCase())
        .where((tag) => tag.isNotEmpty)
        .toSet();
    final statAverage = _averageStatValue(statsProvider.stats, tagSet);
    final difficultyColor = _difficultyColor(statAverage, roundedDifficulty);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Theme.of(context).colorScheme.outlineVariant),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              'Difficulty',
              style: Theme.of(
                context,
              ).textTheme.labelLarge?.copyWith(fontWeight: FontWeight.w600),
            ),
          ),
          Text(
            '$roundedDifficulty',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
              color: difficultyColor,
            ),
          ),
        ],
      ),
    );
  }

  List<DialogueMessage> _buildQuestAcceptanceDialogue(
    Quest? quest,
    CharacterAction action,
  ) {
    final questLines = (quest?.acceptanceDialogue ?? const [])
        .map((line) => line.trim())
        .where((line) => line.isNotEmpty)
        .toList();
    if (questLines.isNotEmpty) {
      return [
        for (var i = 0; i < questLines.length; i++)
          DialogueMessage(speaker: 'character', text: questLines[i], order: i),
      ];
    }

    if (action.dialogue.isNotEmpty) {
      return action.dialogue;
    }

    final actionLines = action.questAcceptanceDialogue
        .map((line) => line.trim())
        .where((line) => line.isNotEmpty)
        .toList();
    if (actionLines.isNotEmpty) {
      return [
        for (var i = 0; i < actionLines.length; i++)
          DialogueMessage(speaker: 'character', text: actionLines[i], order: i),
      ];
    }

    final fallback =
        quest?.description.trim() ?? action.questDescription?.trim() ?? '';
    if (fallback.isNotEmpty) {
      return [DialogueMessage(speaker: 'character', text: fallback, order: 0)];
    }

    return const [DialogueMessage(speaker: 'character', text: '...', order: 0)];
  }

  Future<void> _showQuestAcceptanceDialog(CharacterAction action) async {
    if (_acceptingQuest) return;
    final quest = _questForAction(action);
    final questId = action.questId;
    final metaDifficulty = action.questAverageDifficulty;
    final metaTags = action.questStatTags;
    final accepted = await showDialog<bool>(
      context: context,
      useRootNavigator: true,
      barrierDismissible: true,
      builder: (dialogContext) {
        final footer = metaDifficulty != null
            ? _buildQuestDifficultySummaryFromMetadata(
                dialogContext,
                metaDifficulty,
                metaTags,
              )
            : (quest == null
                  ? (questId == null
                        ? null
                        : _QuestDifficultyFooter(questId: questId))
                  : _buildQuestDifficultySummary(dialogContext, quest));
        return RpgDialogueModal(
          character: widget.character,
          action: action,
          dialogueOverride: _buildQuestAcceptanceDialogue(quest, action),
          footerContent: footer,
          primaryActionLabel: 'Accept quest',
          secondaryActionLabel: 'Decline',
          onPrimaryAction: () => Navigator.of(dialogContext).pop(true),
          onSecondaryAction: () => Navigator.of(dialogContext).pop(false),
          onClose: () => Navigator.of(dialogContext).pop(false),
        );
      },
    );
    if (accepted == true) {
      await _handleQuest(action);
    }
  }

  Future<void> _handleQuest(CharacterAction action) async {
    final questId = action.questId;
    if (questId == null) return;
    final location = context.read<LocationProvider>().location;
    final distance = _questDistanceFrom(location);
    final proximityBlockedReason = _questAcceptDisabledReason(
      location,
      distance,
    );
    if (proximityBlockedReason != null) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(proximityBlockedReason)));
      }
      return;
    }
    setState(() => _acceptingQuest = true);
    try {
      await context.read<PoiService>().acceptQuest(
        characterId: widget.character.id,
        questId: questId,
      );
      await context.read<QuestLogProvider>().refresh();
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('Quest accepted')));
        widget.onClose();
        widget.onQuestAccepted?.call();
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('Failed to accept quest')));
      }
    } finally {
      if (mounted) setState(() => _acceptingQuest = false);
    }
  }

  Future<void> _handleTurnIn(Quest quest, CharacterAction action) async {
    final questId = action.questId ?? quest.id;
    if (questId.isEmpty) return;
    setState(() => _turningInQuest = true);
    try {
      final resp = await context.read<QuestLogProvider>().turnInQuest(questId);
      if (mounted) {
        try {
          await context.read<AuthProvider>().refresh();
        } catch (_) {}
        context.read<CompletedTaskProvider>().showModal(
          'questCompleted',
          data: {'questName': quest.name, ...resp},
        );
        widget.onClose();
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to turn in quest')),
        );
      }
    } finally {
      if (mounted) setState(() => _turningInQuest = false);
    }
  }

  Quest? _questReadyToTurnIn(CharacterAction action) {
    final questId = action.questId;
    if (questId == null || questId.isEmpty) return null;
    final quests = context.read<QuestLogProvider>().quests;
    try {
      return quests.firstWhere((q) => q.id == questId && q.readyToTurnIn);
    } catch (_) {
      return null;
    }
  }

  bool get _hasCharacterLocation {
    return _questGiverCoordinates.isNotEmpty;
  }

  List<AppLocation> get _questGiverCoordinates {
    final coordinates = <AppLocation>[];
    bool isValid(double lat, double lng) {
      if (!lat.isFinite || !lng.isFinite) return false;
      if (lat.abs() > 90 || lng.abs() > 180) return false;
      return lat != 0 || lng != 0;
    }

    final poiLat = widget.character.pointOfInterestLat;
    final poiLng = widget.character.pointOfInterestLng;
    if (poiLat != null && poiLng != null && isValid(poiLat, poiLng)) {
      return [AppLocation(latitude: poiLat, longitude: poiLng, accuracy: 0)];
    }

    for (final loc in widget.character.locations) {
      if (!isValid(loc.latitude, loc.longitude)) continue;
      coordinates.add(
        AppLocation(
          latitude: loc.latitude,
          longitude: loc.longitude,
          accuracy: 0,
        ),
      );
    }

    return coordinates;
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

  double? _questDistanceFrom(AppLocation? location) {
    if (!_hasCharacterLocation || location == null) return null;
    var nearestMeters = double.infinity;
    for (final point in _questGiverCoordinates) {
      final distance = _distanceMeters(
        location.latitude,
        location.longitude,
        point.latitude,
        point.longitude,
      );
      if (distance < nearestMeters) nearestMeters = distance;
    }
    return nearestMeters.isFinite ? nearestMeters : null;
  }

  String? _questAcceptDisabledReason(
    AppLocation? location,
    double? distanceMeters,
  ) {
    if (!_hasCharacterLocation) return null;
    if (location == null) {
      return 'Enable location to accept this quest.';
    }
    if (distanceMeters == null) {
      return 'Character location unavailable.';
    }
    if (distanceMeters > _unlockRadiusMeters) {
      return '${distanceMeters.round()} m away. Need ${_unlockRadiusMeters.round()} m.';
    }
    return null;
  }

  bool get _isDiscoveryManaged {
    return _hasCharacterLocation;
  }

  bool _isAlreadyDiscovered() {
    if (widget.hasDiscovered || _justUnlocked) return true;
    final poiId = widget.character.pointOfInterestId?.trim() ?? '';
    if (poiId.isEmpty) return false;
    try {
      return context.read<DiscoveriesProvider>().hasDiscovered(poiId);
    } catch (_) {
      return false;
    }
  }

  String _errorMessage(Object e) {
    if (e is DioException && e.response?.data is Map) {
      final d = e.response!.data as Map<String, dynamic>;
      final msg = d['error'] ?? d['message'];
      if (msg != null && msg.toString().isNotEmpty) return msg.toString();
    }
    return e.toString();
  }

  bool _isDiscoveryDuplicateError(Object e) {
    final msg = _errorMessage(e).toLowerCase();
    final mentionsDiscovery =
        msg.contains('discover') || msg.contains('point_of_interest');
    final mentionsDuplicate =
        msg.contains('duplicate') ||
        msg.contains('already') ||
        msg.contains('unique') ||
        msg.contains('constraint');
    return mentionsDiscovery && mentionsDuplicate;
  }

  Future<void> _handleUnlock() async {
    if (_unlocking) return;
    if (_isAlreadyDiscovered()) {
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _unlockError = null;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Already discovered.')));
      return;
    }
    final poiId = widget.character.pointOfInterestId?.trim() ?? '';
    final loc = context.read<LocationProvider>().location;
    if (loc == null) {
      setState(
        () => _unlockError = 'Location not available. Enable location access.',
      );
      return;
    }
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) {
      setState(() => _unlockError = 'Please log in to discover.');
      return;
    }
    final distance = _questDistanceFrom(loc);
    if (distance == null) {
      setState(() => _unlockError = 'Character location unavailable.');
      return;
    }
    if (distance > _unlockRadiusMeters) {
      setState(() {
        _unlockError =
            'Too far away (${distance.round()} m). Get within ${_unlockRadiusMeters.round()} m to unlock.';
      });
      return;
    }
    setState(() {
      _unlocking = true;
      _unlockError = null;
    });
    if (poiId.isEmpty) {
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _unlocking = false;
        _unlockError = null;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Discovered!')));
      return;
    }
    try {
      await context.read<PoiService>().unlockPointOfInterest(
        poiId,
        loc.latitude,
        loc.longitude,
        userId: userId,
      );
      if (!mounted) return;
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _unlocking = false;
        _unlockError = null;
      });
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Discovered!')));
    } catch (e) {
      if (_isDiscoveryDuplicateError(e)) {
        if (!mounted) return;
        await widget.onUnlocked?.call();
        if (!mounted) return;
        setState(() {
          _justUnlocked = true;
          _unlocking = false;
          _unlockError = null;
        });
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('Already discovered.')));
        return;
      }
      if (!mounted) return;
      setState(() {
        _unlocking = false;
        _unlockError = _errorMessage(e);
      });
    }
  }

  Future<void> _showCharacterImageDialog(String imageUrl) async {
    await showDialog<void>(
      context: context,
      barrierColor: Colors.black54,
      builder: (dialogContext) {
        final theme = Theme.of(dialogContext);
        return Dialog(
          backgroundColor: Colors.transparent,
          insetPadding: const EdgeInsets.all(24),
          child: Stack(
            alignment: Alignment.topRight,
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.2),
                      blurRadius: 18,
                      offset: const Offset(0, 10),
                    ),
                  ],
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: InteractiveViewer(
                    minScale: 1,
                    maxScale: 4,
                    child: Image.network(
                      imageUrl,
                      width: 360,
                      height: 360,
                      fit: BoxFit.cover,
                      errorBuilder: (_, _, _) => Container(
                        width: 360,
                        height: 360,
                        color: theme.colorScheme.surfaceContainerHighest,
                        child: const Icon(Icons.person, size: 96),
                      ),
                    ),
                  ),
                ),
              ),
              IconButton(
                onPressed: () => Navigator.of(dialogContext).pop(),
                icon: const Icon(Icons.close),
                style: IconButton.styleFrom(
                  backgroundColor: theme.colorScheme.surfaceContainerHighest,
                  shape: const CircleBorder(),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final showFull =
        widget.hasDiscovered || _justUnlocked || !_isDiscoveryManaged;
    if (!showFull) {
      return _buildUndiscovered(context);
    }
    final talkAction = _firstActionOfType('talk');
    final shopAction = _firstActionOfType('shop');
    final questActions = _actions
        .where(
          (action) =>
              ['giveQuest', 'quest', 'quests'].contains(action.actionType),
        )
        .where((action) => action.questId != null && action.questId!.isNotEmpty)
        .toList();
    final rawImageUrl =
        widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;
    final imageUrl = (rawImageUrl != null && rawImageUrl.isNotEmpty)
        ? rawImageUrl
        : null;
    final userLocation = context.watch<LocationProvider>().location;
    final questDistance = _questDistanceFrom(userLocation);
    final questAcceptDisabledReason = _questAcceptDisabledReason(
      userLocation,
      questDistance,
    );

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          widget.character.name,
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                      ),
                      IconButton(
                        onPressed: widget.onClose,
                        icon: const Icon(Icons.close),
                      ),
                    ],
                  ),
                  if (imageUrl != null) ...[
                    const SizedBox(height: 12),
                    Material(
                      color: Colors.transparent,
                      child: InkWell(
                        borderRadius: BorderRadius.circular(16),
                        onTap: () => _showCharacterImageDialog(imageUrl),
                        child: Ink(
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: Theme.of(
                                context,
                              ).colorScheme.outlineVariant,
                            ),
                          ),
                          child: ClipRRect(
                            borderRadius: BorderRadius.circular(16),
                            child: SizedBox(
                              height: 172,
                              child: Image.network(
                                imageUrl,
                                fit: BoxFit.cover,
                                errorBuilder: (_, _, _) => Container(
                                  color: Theme.of(
                                    context,
                                  ).colorScheme.surfaceContainerHighest,
                                  child: const Icon(Icons.person, size: 64),
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                  if (widget.character.description != null &&
                      widget.character.description!.isNotEmpty) ...[
                    const SizedBox(height: 10),
                    Text(
                      widget.character.description!,
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ],
              ),
            ),
            Expanded(
              child: _loadingActions
                  ? const Center(child: CircularProgressIndicator())
                  : _actions.isEmpty
                  ? const Center(child: Text('No actions available'))
                  : ListView(
                      controller: scrollController,
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                      children: [
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Colors.black87,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.white70, width: 2),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              ...questActions.map((action) {
                                final quest = _questForAction(action);
                                final questReadyToTurnIn = _questReadyToTurnIn(
                                  action,
                                );
                                if (questReadyToTurnIn != null) {
                                  final turnInPrefix =
                                      questReadyToTurnIn.isMainStory
                                      ? 'Main Story: '
                                      : '';
                                  return _DialogueChoiceButton(
                                    label: _turningInQuest
                                        ? 'Turning in…'
                                        : '${turnInPrefix}Turn in: ${questReadyToTurnIn.name}',
                                    icon: Icons.assignment_turned_in,
                                    onTap: _turningInQuest
                                        ? null
                                        : () => _handleTurnIn(
                                            questReadyToTurnIn,
                                            action,
                                          ),
                                  );
                                }
                                if (quest?.isAccepted == true) {
                                  return const SizedBox.shrink();
                                }
                                final questAcceptBlocked =
                                    !_acceptingQuest &&
                                    questAcceptDisabledReason != null;
                                return _DialogueChoiceButton(
                                  label: _acceptingQuest
                                      ? 'Accepting quest…'
                                      : '${action.isMainStoryQuest ? 'Main Story: ' : ''}Accept: ${quest?.name ?? action.questName ?? 'Quest'}',
                                  icon: Icons.assignment_turned_in,
                                  subtitle: questAcceptBlocked
                                      ? questAcceptDisabledReason
                                      : null,
                                  onTap:
                                      _acceptingQuest ||
                                          questAcceptDisabledReason != null
                                      ? null
                                      : () =>
                                            _showQuestAcceptanceDialog(action),
                                );
                              }),
                              if (shopAction != null)
                                _DialogueChoiceButton(
                                  label: 'Shop',
                                  icon: Icons.storefront,
                                  onTap: widget.onStartShop == null
                                      ? null
                                      : () {
                                          widget.onStartShop!(
                                            context,
                                            widget.character,
                                            shopAction,
                                          );
                                          widget.onClose();
                                        },
                                ),
                              if (talkAction != null)
                                _DialogueChoiceButton(
                                  label: 'Talk',
                                  icon: Icons.chat_bubble_outline,
                                  onTap: widget.onStartDialogue == null
                                      ? null
                                      : () {
                                          widget.onStartDialogue!(
                                            context,
                                            widget.character,
                                            talkAction,
                                          );
                                        },
                                ),
                            ],
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

  Widget _buildUndiscovered(BuildContext context) {
    final location = context.watch<LocationProvider>().location;
    final distance = _questDistanceFrom(location);
    final withinRange = distance != null && distance <= _unlockRadiusMeters;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
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
                  ClipRRect(
                    borderRadius: BorderRadius.circular(14),
                    child: AspectRatio(
                      aspectRatio: 1,
                      child: Image.network(
                        _placeholderImageUrl,
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) => Container(
                          color: Theme.of(context).colorScheme.surfaceVariant,
                          child: const Icon(Icons.person_search, size: 72),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Visit this location to unlock this character. You must be within ${_unlockRadiusMeters.round()} meters to discover it.',
                    style: Theme.of(context).textTheme.bodyLarge,
                  ),
                  const SizedBox(height: 12),
                  if (distance != null)
                    Text(
                      withinRange
                          ? 'Within range! Tap Unlock to discover.'
                          : 'You are ${distance.round()} m away.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: withinRange
                            ? Theme.of(context).colorScheme.primary
                            : Theme.of(
                                context,
                              ).colorScheme.onSurface.withValues(alpha: 0.7),
                        fontWeight: withinRange ? FontWeight.w600 : null,
                      ),
                    )
                  else
                    Text(
                      _hasCharacterLocation
                          ? 'Enable location to see distance.'
                          : 'Character location unavailable.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Theme.of(
                          context,
                        ).colorScheme.onSurface.withValues(alpha: 0.6),
                      ),
                    ),
                  if (_unlockError != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      _unlockError!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ],
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: (_unlocking || !withinRange)
                        ? null
                        : _handleUnlock,
                    child: Text(
                      _unlocking
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

class _QuestDifficultyFooter extends StatefulWidget {
  const _QuestDifficultyFooter({required this.questId});

  final String questId;

  @override
  State<_QuestDifficultyFooter> createState() => _QuestDifficultyFooterState();
}

class _QuestDifficultyFooterState extends State<_QuestDifficultyFooter> {
  late Future<Quest?> _questFuture;

  @override
  void initState() {
    super.initState();
    _questFuture = _resolveQuest();
  }

  Future<Quest?> _resolveQuest() async {
    final quests = context.read<QuestLogProvider>().quests;
    try {
      return quests.firstWhere((q) => q.id == widget.questId);
    } catch (_) {}
    return context.read<PoiService>().getQuestById(widget.questId);
  }

  double _averageNodeDifficulty(Quest quest) {
    final nodeDifficulties = <double>[];
    for (final node in quest.nodes) {
      final difficulty = node.objective?.difficulty ?? 0;
      if (difficulty <= 0) continue;
      nodeDifficulties.add(difficulty.toDouble());
    }
    if (nodeDifficulties.isEmpty) return 0;
    final total = nodeDifficulties.reduce((a, b) => a + b);
    return total / nodeDifficulties.length;
  }

  int _roundToNearestFive(double value) {
    if (value.isNaN || value.isInfinite) return 0;
    return (value / 5).round() * 5;
  }

  Set<String> _questStatTags(Quest quest) {
    final tags = <String>{};
    for (final node in quest.nodes) {
      for (final tag in node.objective?.statTags ?? const <String>[]) {
        final normalized = tag.trim().toLowerCase();
        if (normalized.isNotEmpty) {
          tags.add(normalized);
        }
      }
    }
    return tags;
  }

  double _averageStatValue(Map<String, int> stats, Set<String> tags) {
    if (stats.isEmpty) return 0;
    if (tags.isEmpty) {
      final values = stats.values;
      final total = values.fold<int>(0, (sum, value) => sum + value);
      return total / values.length;
    }
    var total = 0;
    var count = 0;
    for (final tag in tags) {
      if (!stats.containsKey(tag)) continue;
      total += stats[tag] ?? 0;
      count += 1;
    }
    if (count == 0) return 0;
    return total / count;
  }

  Color _difficultyColor(double statAverage, int difficulty) {
    if (statAverage > difficulty) {
      return const Color(0xFFCBD5E1);
    }
    if (statAverage > difficulty - 25) {
      return const Color(0xFF22C55E);
    }
    if (statAverage > difficulty - 50) {
      return const Color(0xFFF59E0B);
    }
    return const Color(0xFFEF4444);
  }

  Widget _buildQuestDifficultySummary(
    BuildContext context,
    Quest quest,
    Map<String, int> stats,
  ) {
    final avgDifficulty = _averageNodeDifficulty(quest);
    final roundedDifficulty = _roundToNearestFive(avgDifficulty);
    final tags = _questStatTags(quest);
    final statAverage = _averageStatValue(stats, tags);
    final difficultyColor = _difficultyColor(statAverage, roundedDifficulty);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Theme.of(context).colorScheme.outlineVariant),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              'Difficulty',
              style: Theme.of(
                context,
              ).textTheme.labelLarge?.copyWith(fontWeight: FontWeight.w600),
            ),
          ),
          Text(
            '$roundedDifficulty',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
              color: difficultyColor,
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final stats = context.watch<CharacterStatsProvider>().stats;
    return FutureBuilder<Quest?>(
      future: _questFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: Theme.of(context).colorScheme.outlineVariant,
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    'Difficulty',
                    style: Theme.of(context).textTheme.labelLarge?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ],
            ),
          );
        }
        final quest = snapshot.data;
        if (quest == null) {
          return Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: Theme.of(context).colorScheme.outlineVariant,
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    'Difficulty',
                    style: Theme.of(context).textTheme.labelLarge?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                Text(
                  '—',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                    color: Theme.of(context).colorScheme.outline,
                  ),
                ),
              ],
            ),
          );
        }
        return _buildQuestDifficultySummary(context, quest, stats);
      },
    );
  }
}

class _DialogueChoiceButton extends StatelessWidget {
  const _DialogueChoiceButton({
    required this.label,
    required this.icon,
    this.subtitle,
    this.onTap,
  });

  final String label;
  final IconData icon;
  final String? subtitle;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 12),
          decoration: BoxDecoration(
            color: onTap == null ? Colors.white12 : Colors.white10,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.white30),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.only(top: 1),
                child: Icon(
                  icon,
                  color: onTap == null ? Colors.white38 : Colors.white,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      label,
                      style: TextStyle(
                        color: onTap == null ? Colors.white38 : Colors.white,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    if (subtitle != null) ...[
                      const SizedBox(height: 2),
                      Text(
                        subtitle!,
                        style: const TextStyle(
                          color: Colors.white54,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
