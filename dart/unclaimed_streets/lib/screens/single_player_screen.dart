import 'dart:async';
import 'dart:math' show Point;
import 'dart:math' as math;
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
import 'package:pointer_interceptor/pointer_interceptor.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/challenge.dart';
import '../models/healing_fountain.dart';
import '../models/monster.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../models/quest_node_challenge.dart';
import '../models/scenario.dart';
import '../models/treasure_chest.dart';
import '../models/tutorial.dart';
import '../models/zone.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/log_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/completed_task_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/quest_filter_provider.dart';
import '../providers/tags_provider.dart';
import '../providers/tutorial_replay_provider.dart';
import '../providers/zone_provider.dart';
import '../providers/map_focus_provider.dart';
import '../providers/party_provider.dart';
import '../services/media_service.dart';
import '../services/poi_service.dart';
import '../utils/poi_image_util.dart';
import '../utils/camera_capture.dart';
import '../constants/api_constants.dart';
import '../constants/gameplay_constants.dart';
import '../widgets/activity_feed_panel.dart';
import '../widgets/celebration_modal_manager.dart';
import '../widgets/character_panel.dart';
import '../widgets/healing_fountain_panel.dart';
import '../widgets/inventory_panel.dart';
import '../widgets/log_panel.dart';
import '../widgets/monster_battle_dialog.dart';
import '../widgets/monster_panel.dart';
import '../widgets/new_item_modal.dart';
import '../widgets/point_of_interest_panel.dart';
import '../widgets/quest_log_panel.dart';
import '../widgets/rpg_dialogue_modal.dart';
import '../widgets/scenario_panel.dart';
import '../widgets/tracked_quests_overlay.dart';
import '../widgets/shop_modal.dart';
import '../widgets/quest_filter_panel.dart';
import '../widgets/treasure_chest_panel.dart';
import '../widgets/used_item_modal.dart';
import '../widgets/zone_widget.dart';
import '../widgets/paper_texture.dart';
import '../widgets/party_member_map_strip.dart';
import 'layout_shell.dart';

const _chestImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';
const _scenarioMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png';
const _monsterMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png';
const _characterMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png';
const _challengeMysteryImageUrl = _scenarioMysteryImageUrl;
const _healingFountainFallbackImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';
const _legacyMysteryImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';
const _defeatedMonstersPrefsKeyPrefix = 'single_player_defeated_monsters';
const _discoveredCharactersPrefsKeyPrefix =
    'single_player_discovered_characters';
const _mapThumbnailVersion = 'v4';
const _standardMarkerThumbnailSize = 0.75;
const _poiImageLoadBatchSize = 24;
const _poiSymbolAddBatchSize = 32;
const _poiAssociationCoordinatePrecision = 4;
const _stamenWatercolorStyleBase =
    'https://tiles.stadiamaps.com/styles/stamen_watercolor.json';
const _stamenWatercolorApiKey = String.fromEnvironment(
  'STADIA_MAPS_API_KEY',
  defaultValue: '',
);
final String _stamenWatercolorStyle = _stamenWatercolorApiKey.isNotEmpty
    ? '$_stamenWatercolorStyleBase?api_key=$_stamenWatercolorApiKey'
    : _stamenWatercolorStyleBase;

class SinglePlayerScreen extends StatefulWidget {
  const SinglePlayerScreen({super.key});

  @override
  State<SinglePlayerScreen> createState() => _SinglePlayerScreenState();
}

class _SinglePlayerScreenState extends State<SinglePlayerScreen> {
  MapLibreMapController? _mapController;
  List<Zone> _zones = [];
  List<PointOfInterest> _pois = [];
  List<Character> _characters = [];
  List<TreasureChest> _treasureChests = [];
  List<HealingFountain> _healingFountains = [];
  List<Scenario> _scenarios = [];
  List<MonsterEncounter> _monsters = [];
  List<Challenge> _challenges = [];
  List<Line> _zoneLines = [];
  List<Fill> _zoneFills = [];
  List<Line> _questLines = [];
  List<Fill> _questFills = [];
  List<Symbol> _poiSymbols = [];
  final Map<String, Symbol> _poiSymbolById = {};
  int _poiMarkerGeneration = 0;
  List<Symbol> _questPoiHighlightSymbols = [];
  List<Symbol> _characterSymbols = [];
  final Map<String, List<Symbol>> _characterSymbolsById = {};
  List<Symbol> _chestSymbols = [];
  List<Circle> _chestCircles = [];
  final Map<String, Symbol> _chestSymbolById = {};
  final Map<String, Circle> _chestCircleById = {};
  final Map<String, bool> _chestCircleOpened = {};
  List<Symbol> _healingFountainSymbols = [];
  List<Circle> _healingFountainCircles = [];
  final Map<String, Symbol> _healingFountainSymbolById = {};
  final Map<String, Circle> _healingFountainCircleById = {};
  List<Symbol> _scenarioSymbols = [];
  List<Circle> _scenarioCircles = [];
  final Map<String, Symbol> _scenarioSymbolById = {};
  final Map<String, Circle> _scenarioCircleById = {};
  final Map<String, bool> _scenarioCircleMystery = {};
  final Map<String, bool> _scenarioQuestObjective = {};
  List<Symbol> _monsterSymbols = [];
  List<Circle> _monsterCircles = [];
  final Map<String, Symbol> _monsterSymbolById = {};
  final Map<String, Circle> _monsterCircleById = {};
  List<Symbol> _challengeSymbols = [];
  List<Circle> _challengeCircles = [];
  final Map<String, Symbol> _challengeSymbolById = {};
  final Map<String, Circle> _challengeCircleById = {};
  final Set<String> _resolvedScenarioIds = <String>{};
  final Set<String> _resolvedScenarioSignatures = <String>{};
  final Set<String> _openedTreasureChestIds = <String>{};
  final Set<String> _defeatedMonsterIds = <String>{};
  final Set<String> _discoveredCharacterIds = <String>{};
  String? _defeatedMonsterIdsUserId;
  String? _discoveredCharacterIdsUserId;
  final ZoneWidgetController _zoneWidgetController = ZoneWidgetController();
  Uint8List? _chestThumbnailBytes;
  bool _chestThumbnailAdded = false;
  Uint8List? _scenarioMysteryThumbnailBytes;
  bool _scenarioMysteryThumbnailAdded = false;
  Uint8List? _monsterMysteryThumbnailBytes;
  bool _monsterMysteryThumbnailAdded = false;
  Uint8List? _challengeMysteryThumbnailBytes;
  bool _challengeMysteryThumbnailAdded = false;
  bool _styleLoaded = false;
  bool _markersAdded = false;
  bool _addedMarkersWithEmptyDiscoveries = false;
  bool _mapLoadFailed = false;
  int _mapKey = 0;
  bool _hasAnimatedToUserLocation = false;
  QuestLogProvider? _questLogProvider;
  MapFocusProvider? _mapFocusProvider;
  CompletedTaskProvider? _completedTaskProvider;
  Map<String, dynamic>? _lastHandledCompletionModal;
  Timer? _questGlowTimer;
  bool _isQuestGlowPulsing = false;
  Timer? _questPoiPulseTimer;
  bool _isQuestPoiPulseActive = false;
  bool _questPoiPulseUp = false;
  Timer? _zoneAutoSelectTimer;
  String? _questLogRequestedZoneId;
  bool _questLogRefreshInFlight = false;
  bool _questAvailabilityRefreshInFlight = false;
  DateTime? _lastQuestLogRefreshAt;
  bool _questLogNeedsOverlayApply = false;
  bool _scenarioVisibilityRefreshPending = false;
  Future<void> _scenarioRefreshSequence = Future<void>.value();
  Future<void> _monsterRefreshSequence = Future<void>.value();
  Future<void> _challengeRefreshSequence = Future<void>.value();
  Set<String> _lastQuestPoiIds = <String>{};
  int _lastQuestPolygonHash = 0;
  String _lastMapFilterKey = '';
  QuestSubmissionOverlayPhase _questSubmissionPhase =
      QuestSubmissionOverlayPhase.hidden;
  String? _questSubmissionMessage;
  int? _questSubmissionScore;
  int? _questSubmissionDifficulty;
  int? _questSubmissionCombinedScore;
  List<String> _questSubmissionStatTags = const [];
  Map<String, int> _questSubmissionStatValues = const <String, int>{};
  int _questSubmissionRevealStep = 0;
  final List<Timer> _questSubmissionRevealTimers = [];
  final TrackedQuestsOverlayController _trackedQuestsController =
      TrackedQuestsOverlayController();
  String? _lastHandledMonsterBattleIntent;
  bool _handlingMonsterBattleIntent = false;
  TutorialStatus? _tutorialStatus;
  bool _tutorialStatusLoading = false;
  bool _tutorialStatusChecked = false;
  bool _tutorialDialogVisible = false;
  bool _tutorialActivationInFlight = false;
  bool _tutorialReplayPending = false;
  int _lastTutorialReplayRequestCount = 0;
  String? _tutorialFocusedScenarioId;
  String? _tutorialFocusedMonsterEncounterId;
  bool _tutorialNormalPinsRevealInProgress = false;
  bool _tutorialLoadoutPendingAfterCompletionModal = false;
  bool _tutorialRevealPendingAfterCompletionModal = false;
  bool _tutorialWelcomeOverlayVisible = false;
  double _tutorialWelcomeOverlayOpacity = 0.0;

  @override
  void initState() {
    super.initState();
    debugPrint('SinglePlayer: initState');
    _startMapLoadTimeout();
    _loadAll();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<ZoneProvider>().addListener(_onZoneChanged);
      context.read<LocationProvider>().addListener(_onLocationChanged);
      context.read<AuthProvider>().addListener(_onAuthChanged);
      context.read<QuestFilterProvider>().addListener(_onFilterChanged);
      context.read<TagsProvider>().addListener(_onTagFilterChanged);
      unawaited(context.read<LocationProvider>().ensureLoaded());
      _questLogProvider = context.read<QuestLogProvider>();
      _questLogProvider?.addListener(_onQuestLogChanged);
      _mapFocusProvider = context.read<MapFocusProvider>();
      _mapFocusProvider?.addListener(_onMapFocusRequest);
      _completedTaskProvider = context.read<CompletedTaskProvider>();
      _completedTaskProvider?.addListener(_onCompletedTaskModalChanged);
      context.read<TutorialReplayProvider>().addListener(
        _onTutorialReplayRequested,
      );
      _onCompletedTaskModalChanged();
      _updateSelectedZoneFromLocation();
      _requestQuestLogIfReady();
      context.read<ActivityFeedProvider>().refresh();
      unawaited(context.read<PartyProvider>().fetchParty());
      unawaited(_loadTutorialStatus(force: true));
      _onTutorialReplayRequested();
    });
  }

  @override
  void dispose() {
    _mapLoadTimeout?.cancel();
    _questGlowTimer?.cancel();
    _questPoiPulseTimer?.cancel();
    _zoneAutoSelectTimer?.cancel();
    _clearQuestSubmissionRevealTimers();
    _trackedQuestsController.dispose();
    try {
      context.read<ZoneProvider>().removeListener(_onZoneChanged);
    } catch (_) {}
    try {
      context.read<LocationProvider>().removeListener(_onLocationChanged);
    } catch (_) {}
    try {
      context.read<AuthProvider>().removeListener(_onAuthChanged);
    } catch (_) {}
    try {
      context.read<QuestFilterProvider>().removeListener(_onFilterChanged);
    } catch (_) {}
    try {
      context.read<TagsProvider>().removeListener(_onTagFilterChanged);
    } catch (_) {}
    try {
      _questLogProvider?.removeListener(_onQuestLogChanged);
    } catch (_) {}
    try {
      _mapFocusProvider?.removeListener(_onMapFocusRequest);
    } catch (_) {}
    try {
      _completedTaskProvider?.removeListener(_onCompletedTaskModalChanged);
    } catch (_) {}
    try {
      context.read<TutorialReplayProvider>().removeListener(
        _onTutorialReplayRequested,
      );
    } catch (_) {}
    super.dispose();
  }

  void _onZoneChanged() {
    if (!mounted) return;
    unawaited(_loadTreasureChestsForSelectedZone());
    unawaited(_addZoneBoundaries());
    if (_styleLoaded && _mapController != null) {
      if (_markersAdded) {
        unawaited(_refreshUndiscoveredPoiOpacitiesForZone());
      } else {
        unawaited(_addPoiMarkers());
      }
    }
    _requestQuestLogIfReady();
  }

  void _onLocationChanged() {
    if (!mounted) return;
    _updateSelectedZoneFromLocation();
    _requestQuestLogIfReady();
    _refreshScenarioVisibilityForLocationChange();
    _maybeShowTutorialWelcome();
  }

  void _refreshScenarioVisibilityForLocationChange() {
    if (_styleLoaded && _mapController != null && _markersAdded) {
      _scenarioVisibilityRefreshPending = false;
      if (_isTutorialMapFocusActive) {
        unawaited(
          (() async {
            await _refreshScenarioSymbols();
            await _refreshMonsterSymbols();
            await _refreshChallengeSymbols();
          })(),
        );
        return;
      }
      unawaited(
        (() async {
          await _refreshScenarioSymbols();
          await _refreshMonsterSymbols();
          await _refreshChallengeSymbols();
        })(),
      );
      return;
    }
    _scenarioVisibilityRefreshPending = true;
  }

  void _onAuthChanged() {
    if (!mounted) return;
    _requestQuestLogIfReady(force: true);
    unawaited(_restoreDefeatedMonsterIds(refreshMap: true));
    unawaited(_restoreDiscoveredCharacterIds(refreshMap: true));
    unawaited(context.read<PartyProvider>().fetchParty());
    unawaited(_loadTutorialStatus(force: true));
  }

  void _onTutorialReplayRequested() {
    if (!mounted) return;
    final provider = context.read<TutorialReplayProvider>();
    final requestCount = provider.requestCount;
    if (requestCount == _lastTutorialReplayRequestCount) return;
    _lastTutorialReplayRequestCount = requestCount;
    _tutorialReplayPending = true;
    unawaited(_loadTutorialStatus(force: true));
  }

  bool get _isTutorialMapFocusActive =>
      !_tutorialNormalPinsRevealInProgress &&
      ((_tutorialFocusedScenarioId?.trim().isNotEmpty ?? false) ||
          (_tutorialFocusedMonsterEncounterId?.trim().isNotEmpty ?? false) ||
          (_tutorialStatus?.isLoadoutStep ?? false));

  bool _isTutorialFocusedScenarioId(String scenarioId) {
    final focusedId = _tutorialFocusedScenarioId?.trim() ?? '';
    return focusedId.isNotEmpty && focusedId == scenarioId.trim();
  }

  bool _isTutorialFocusedMonsterEncounterId(String encounterId) {
    final focusedId = _tutorialFocusedMonsterEncounterId?.trim() ?? '';
    return focusedId.isNotEmpty && focusedId == encounterId.trim();
  }

  double _mapMarkerStartingOpacity(double targetOpacity) {
    if (_tutorialNormalPinsRevealInProgress && !_isTutorialMapFocusActive) {
      return 0.0;
    }
    return targetOpacity;
  }

  Future<void> _rebuildMapPins() async {
    if (!_styleLoaded || _mapController == null) return;
    if (mounted) {
      setState(() => _markersAdded = false);
    } else {
      _markersAdded = false;
    }
    await _addPoiMarkers();
  }

  Future<void> _ensureTutorialScenarioLoaded(TutorialStatus? status) async {
    if (status == null || !status.hasActiveScenario) return;
    final scenarioId = status.scenarioId?.trim() ?? '';
    if (scenarioId.isEmpty) return;
    final alreadyLoaded = _scenarios.any(
      (scenario) => scenario.id == scenarioId,
    );
    if (alreadyLoaded) return;
    final scenario = await context.read<PoiService>().getScenarioById(
      scenarioId,
    );
    if (!mounted || scenario == null) return;
    setState(() {
      _scenarios = [
        scenario,
        ..._scenarios.where((item) => item.id != scenario.id),
      ];
    });
  }

  Future<void> _ensureTutorialMonsterLoaded(TutorialStatus? status) async {
    if (status == null || !status.hasActiveMonsterEncounter) return;
    final encounterId = status.monsterEncounterId?.trim() ?? '';
    if (encounterId.isEmpty) return;
    final alreadyLoaded = _monsters.any(
      (encounter) => encounter.id == encounterId,
    );
    if (alreadyLoaded) return;
    final encounter = await context.read<PoiService>().getMonsterEncounterById(
      encounterId,
    );
    if (!mounted || encounter == null) return;
    setState(() {
      _monsters = [
        encounter,
        ..._monsters.where((item) => item.id != encounter.id),
      ];
    });
  }

  Future<void> _syncTutorialMapModeFromStatus(TutorialStatus? status) async {
    final nextFocusedScenarioId = status != null && status.hasActiveScenario
        ? status.scenarioId?.trim() ?? ''
        : '';
    final nextFocusedMonsterEncounterId =
        status != null && status.hasActiveMonsterEncounter
        ? status.monsterEncounterId?.trim() ?? ''
        : '';
    final previousFocusedScenarioId = _tutorialFocusedScenarioId?.trim() ?? '';
    final previousFocusedMonsterEncounterId =
        _tutorialFocusedMonsterEncounterId?.trim() ?? '';
    if (nextFocusedScenarioId == previousFocusedScenarioId &&
        nextFocusedMonsterEncounterId == previousFocusedMonsterEncounterId) {
      return;
    }
    if (mounted) {
      setState(() {
        _tutorialFocusedScenarioId = nextFocusedScenarioId.isEmpty
            ? null
            : nextFocusedScenarioId;
        _tutorialFocusedMonsterEncounterId =
            nextFocusedMonsterEncounterId.isEmpty
            ? null
            : nextFocusedMonsterEncounterId;
      });
    } else {
      _tutorialFocusedScenarioId = nextFocusedScenarioId.isEmpty
          ? null
          : nextFocusedScenarioId;
      _tutorialFocusedMonsterEncounterId = nextFocusedMonsterEncounterId.isEmpty
          ? null
          : nextFocusedMonsterEncounterId;
    }
    await _rebuildMapPins();
    if (nextFocusedMonsterEncounterId.isNotEmpty &&
        nextFocusedMonsterEncounterId != previousFocusedMonsterEncounterId) {
      final encounter = _monsters.firstWhere(
        (entry) => entry.id == nextFocusedMonsterEncounterId,
        orElse: () => const MonsterEncounter(
          id: '',
          name: '',
          zoneId: '',
          latitude: 0,
          longitude: 0,
        ),
      );
      if (encounter.id.isNotEmpty) {
        _flyToLocation(encounter.latitude, encounter.longitude);
        unawaited(_pulsePoi(encounter.latitude, encounter.longitude));
      }
    }
  }

  void _syncTutorialInventorySession(TutorialStatus? status) {
    final drawerController = LayoutShellDrawerController.maybeOf(context);
    if (drawerController == null) return;
    if (status == null || !status.isLoadoutStep) {
      drawerController.stopInventoryTutorial();
      return;
    }
    drawerController.startInventoryTutorial(
      InventoryTutorialSession(
        dialogue: status.loadoutDialogue,
        requiredEquipItemIds: status.requiredEquipItemIds,
        completedEquipItemIds: status.completedEquipItemIds,
        requiredUseItemIds: status.requiredUseItemIds,
        completedUseItemIds: status.completedUseItemIds,
        onProgressChanged: () => _loadTutorialStatus(force: true),
      ),
    );
  }

  Future<void> _beginTutorialNormalPinsReveal() async {
    if (!mounted) return;
    setState(() {
      _tutorialFocusedScenarioId = null;
      _tutorialFocusedMonsterEncounterId = null;
      _tutorialNormalPinsRevealInProgress = true;
      _tutorialWelcomeOverlayVisible = true;
      _tutorialWelcomeOverlayOpacity = 0.0;
    });
    await _loadTutorialStatus(force: true);
    await _rebuildMapPins();
    await _runTutorialWelcomeOverlaySequence();
  }

  Future<void> _beginTutorialLoadoutStep() async {
    if (!mounted) return;
    await _loadTutorialStatus(force: true);
    if (!mounted) return;
    final status = _tutorialStatus;
    if (status?.isLoadoutStep ?? false) {
      setState(() {
        _tutorialFocusedScenarioId = null;
        _tutorialFocusedMonsterEncounterId = null;
      });
      await _rebuildMapPins();
      return;
    }
    if (status?.hasActiveMonsterEncounter ?? false) {
      await _rebuildMapPins();
      return;
    }
    if (status?.isCompleted ?? false) {
      await _beginTutorialNormalPinsReveal();
    }
  }

  Future<void> _runTutorialWelcomeOverlaySequence() async {
    if (!mounted) return;
    const fadeInSteps = 5;
    for (var i = 1; i <= fadeInSteps; i++) {
      if (!mounted) return;
      setState(() => _tutorialWelcomeOverlayOpacity = i / fadeInSteps);
      await Future<void>.delayed(const Duration(milliseconds: 50));
    }
    await Future<void>.delayed(const Duration(milliseconds: 1800));
    if (!mounted) return;

    final c = _mapController;
    if (c == null || !_styleLoaded) {
      setState(() {
        _tutorialWelcomeOverlayVisible = false;
        _tutorialWelcomeOverlayOpacity = 0.0;
        _tutorialNormalPinsRevealInProgress = false;
      });
      return;
    }

    final questPoiIds = _currentQuestPoiIdsForFilter(
      context.read<QuestLogProvider>(),
    );
    final discoveries = context.read<DiscoveriesProvider>();
    final mapContentPoiIds = _buildPoiIdsWithMapContent();
    const fadeOutSteps = 16;
    for (var i = 0; i <= fadeOutSteps; i++) {
      if (!mounted || !_tutorialNormalPinsRevealInProgress) return;
      final progress = i / fadeOutSteps;
      setState(() => _tutorialWelcomeOverlayOpacity = 1.0 - progress);
      await _updateNormalPinOpacities(
        c,
        progress,
        questPoiIds: questPoiIds,
        discoveries: discoveries,
        mapContentPoiIds: mapContentPoiIds,
      );
      if (i != fadeOutSteps) {
        await Future<void>.delayed(const Duration(milliseconds: 90));
      }
    }

    if (!mounted) return;
    setState(() {
      _tutorialWelcomeOverlayVisible = false;
      _tutorialWelcomeOverlayOpacity = 0.0;
      _tutorialNormalPinsRevealInProgress = false;
    });
    _ensureQuestPoiPulseTimer();
  }

  void _onCompletedTaskModalChanged() {
    final provider = _completedTaskProvider;
    if (!mounted || provider == null) return;
    final modal = provider.currentModal;
    if (modal == null) {
      _lastHandledCompletionModal = null;
      if (_tutorialLoadoutPendingAfterCompletionModal) {
        _tutorialLoadoutPendingAfterCompletionModal = false;
        unawaited(_beginTutorialLoadoutStep());
      }
      if (_tutorialRevealPendingAfterCompletionModal) {
        _tutorialRevealPendingAfterCompletionModal = false;
        unawaited(_beginTutorialNormalPinsReveal());
      }
      return;
    }
    if (identical(modal, _lastHandledCompletionModal)) {
      return;
    }
    _lastHandledCompletionModal = modal;
    unawaited(_reconcileMapFromCompletionModal(modal));
  }

  Future<void> _reconcileMapFromCompletionModal(
    Map<String, dynamic> modal,
  ) async {
    final type = modal['type']?.toString().trim() ?? '';
    final rawData = modal['data'];
    final data = rawData is Map
        ? Map<String, dynamic>.from(rawData)
        : const <String, dynamic>{};

    switch (type) {
      case 'scenarioOutcome':
        final scenarioId = data['scenarioId']?.toString().trim() ?? '';
        if (scenarioId.isEmpty) return;
        final wasTutorialScenario = _isTutorialFocusedScenarioId(scenarioId);
        await _removeScenarioLocally(
          scenarioId,
          performedScenarioId: scenarioId,
        );
        if (wasTutorialScenario) {
          _tutorialLoadoutPendingAfterCompletionModal = true;
        }
        return;
      case 'monsterBattleVictory':
        final encounterId = data['monsterEncounterId']?.toString().trim() ?? '';
        if (encounterId.isEmpty) return;
        final wasTutorialMonster = _isTutorialFocusedMonsterEncounterId(
          encounterId,
        );
        if (!wasTutorialMonster) return;
        setState(() {
          _tutorialFocusedMonsterEncounterId = null;
          _monsters.removeWhere((item) => item.id == encounterId);
        });
        _tutorialRevealPendingAfterCompletionModal = true;
        return;
      default:
        return;
    }
  }

  String _defeatedMonstersPrefsKey(String userId) {
    return '$_defeatedMonstersPrefsKeyPrefix:$userId';
  }

  String _discoveredCharactersPrefsKey(String userId) {
    return '$_discoveredCharactersPrefsKeyPrefix:$userId';
  }

  Future<void> _restoreDefeatedMonsterIds({bool refreshMap = false}) async {
    final auth = context.read<AuthProvider>();
    if (auth.loading) return;
    final userId = auth.user?.id;
    if (_defeatedMonsterIdsUserId == userId) return;

    final prefs = await SharedPreferences.getInstance();
    final storedIds = userId == null || userId.isEmpty
        ? const <String>[]
        : (prefs.getStringList(_defeatedMonstersPrefsKey(userId)) ??
              const <String>[]);
    if (!mounted) return;

    setState(() {
      _defeatedMonsterIdsUserId = userId;
      _defeatedMonsterIds
        ..clear()
        ..addAll(storedIds.where((id) => id.trim().isNotEmpty));
      if (_monsters.isNotEmpty) {
        _monsters = _monsters
            .where((monster) => !_defeatedMonsterIds.contains(monster.id))
            .toList();
      }
    });

    if (refreshMap && _styleLoaded && _mapController != null && _markersAdded) {
      await _refreshMonsterSymbols();
    }
  }

  Future<void> _persistDefeatedMonsterIds() async {
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) return;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(
      _defeatedMonstersPrefsKey(userId),
      _defeatedMonsterIds.toList(growable: false),
    );
  }

  Future<void> _restoreDiscoveredCharacterIds({bool refreshMap = false}) async {
    final auth = context.read<AuthProvider>();
    if (auth.loading) return;
    final userId = auth.user?.id;
    if (_discoveredCharacterIdsUserId == userId) return;

    final prefs = await SharedPreferences.getInstance();
    final storedIds = userId == null || userId.isEmpty
        ? const <String>[]
        : (prefs.getStringList(_discoveredCharactersPrefsKey(userId)) ??
              const <String>[]);
    if (!mounted) return;

    setState(() {
      _discoveredCharacterIdsUserId = userId;
      _discoveredCharacterIds
        ..clear()
        ..addAll(storedIds.where((id) => id.trim().isNotEmpty));
    });

    if (refreshMap && _styleLoaded && _mapController != null && _markersAdded) {
      await _refreshCharacterDiscoveryMarkers();
    }
  }

  Future<void> _persistDiscoveredCharacterIds() async {
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) return;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(
      _discoveredCharactersPrefsKey(userId),
      _discoveredCharacterIds.toList(growable: false),
    );
  }

  Future<void> _markCharacterDiscovered(String characterId) async {
    final normalized = characterId.trim();
    if (normalized.isEmpty) return;
    if (!_discoveredCharacterIds.contains(normalized)) {
      setState(() => _discoveredCharacterIds.add(normalized));
    }
    await _persistDiscoveredCharacterIds();
  }

  void _onFilterChanged() {
    if (!mounted) return;
    _maybeApplyMapFilters();
  }

  void _onTagFilterChanged() {
    if (!mounted) return;
    _maybeApplyMapFilters();
  }

  void _maybeApplyMapFilters() {
    final key = _buildMapFilterKey();
    if (key == _lastMapFilterKey) return;
    _lastMapFilterKey = key;
    if (_styleLoaded && _mapController != null) {
      setState(() => _markersAdded = false);
      unawaited(_addPoiMarkers());
    }
  }

  void _requestQuestLogIfReady({bool force = false}) {
    final auth = context.read<AuthProvider>();
    if (auth.loading) {
      return;
    }
    if (!auth.isAuthenticated) return;
    _updateSelectedZoneFromLocation();
    final zoneId = context.read<ZoneProvider>().selectedZone?.id;
    if (zoneId == null || zoneId.isEmpty) {
      final loc = context.read<LocationProvider>().location;
      if (loc == null) return;
      if (!force &&
          _lastQuestLogRefreshAt != null &&
          DateTime.now().difference(_lastQuestLogRefreshAt!) <
              const Duration(seconds: 20)) {
        return;
      }
      _refreshQuestLog();
      return;
    }
    if (!force && _questLogRequestedZoneId == zoneId) return;
    _questLogRequestedZoneId = zoneId;
    _refreshQuestLog();
  }

  void _onQuestLogChanged() {
    if (!mounted) return;
    final questLog = context.read<QuestLogProvider>();
    if (questLog.loading) return;
    _applyQuestLogOverlaysIfChanged();
    unawaited(_refreshScenarioSymbols());
    unawaited(_refreshMonsterSymbols());
    unawaited(_refreshChallengeSymbols());
    unawaited(_loadTreasureChestsForSelectedZone());
    unawaited(_refreshQuestAvailabilityMarkers());
  }

  void _onMapFocusRequest() {
    if (!mounted) return;
    final provider = context.read<MapFocusProvider>();
    final poi = provider.consumePoi();
    if (poi != null) {
      _focusQuestPoI(poi);
      return;
    }
    final quest = provider.consumeTurnInQuest();
    if (quest != null) {
      _focusQuestTurnIn(quest);
    }
  }

  void _refreshQuestLog() {
    if (_questLogRefreshInFlight) return;
    _questLogRefreshInFlight = true;
    _lastQuestLogRefreshAt = DateTime.now();
    unawaited(() async {
      try {
        await context.read<QuestLogProvider>().refresh();
      } finally {
        _questLogRefreshInFlight = false;
      }
    }());
  }

  void _applyQuestLogOverlaysIfChanged() {
    if (!_styleLoaded || _mapController == null) {
      _questLogNeedsOverlayApply = true;
      return;
    }
    final questLog = context.read<QuestLogProvider>();
    final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
    final polygonHash = _hashQuestPolygons(questLog.currentNodePolygons);
    debugPrint(
      'SinglePlayer: quest overlay check poiIds=${questPoiIds.length} polys=${questLog.currentNodePolygons.length}',
    );
    final poiChanged = !_setEquals(_lastQuestPoiIds, questPoiIds);
    final polyChanged = polygonHash != _lastQuestPolygonHash;
    if (!poiChanged && !polyChanged) return;
    final newlyAddedPoiIds = questPoiIds.difference(_lastQuestPoiIds);
    final removedPoiIds = _lastQuestPoiIds.difference(questPoiIds);
    _lastQuestPoiIds = questPoiIds;
    _lastQuestPolygonHash = polygonHash;
    if (poiChanged && newlyAddedPoiIds.isNotEmpty) {
      for (final poiId in newlyAddedPoiIds) {
        unawaited(_updatePoiSymbolForQuestState(poiId, isQuestCurrent: true));
        final poi = _pois.firstWhere(
          (p) => p.id == poiId,
          orElse: () => PointOfInterest(id: '', name: '', lat: '0', lng: '0'),
        );
        if (poi.id.isEmpty) continue;
        final lat = double.tryParse(poi.lat) ?? 0.0;
        final lng = double.tryParse(poi.lng) ?? 0.0;
        unawaited(_pulsePoi(lat, lng));
      }
    }
    if (poiChanged && removedPoiIds.isNotEmpty) {
      for (final poiId in removedPoiIds) {
        unawaited(_updatePoiSymbolForQuestState(poiId, isQuestCurrent: false));
      }
    }
    if (polyChanged) {
      unawaited(_addQuestPolygons());
      for (final poly in questLog.currentNodePolygons) {
        unawaited(_pulsePolygon(poly));
      }
    }
  }

  bool _setEquals<T>(Set<T> a, Set<T> b) {
    if (a.length != b.length) return false;
    for (final v in a) {
      if (!b.contains(v)) return false;
    }
    return true;
  }

  int _hashQuestPolygons(List<List<QuestNodePolygonPoint>> polygons) {
    var hash = polygons.length;
    for (final poly in polygons) {
      hash = hash * 31 + poly.length;
      for (final p in poly) {
        hash = hash * 31 + (p.latitude * 100000).round();
        hash = hash * 31 + (p.longitude * 100000).round();
      }
    }
    return hash;
  }

  String _buildMapFilterKey() {
    final filters = context.read<QuestFilterProvider>();
    final tags = context.read<TagsProvider>();
    final selectedIds = tags.selectedTagIds.toList()..sort();
    return [filters.enableTagFilter, selectedIds.join(',')].join('|');
  }

  Set<String> _currentQuestPoiIdsForFilter(QuestLogProvider questLog) {
    final ids = questLog.currentNodePoiIds.toSet();
    if (_pois.isEmpty) return ids;
    final turnInCharacterIds = _currentQuestTurnInCharacterIds(questLog);
    if (turnInCharacterIds.isEmpty) return ids;
    for (final poi in _pois) {
      if (poi.characters.any((ch) => turnInCharacterIds.contains(ch.id))) {
        ids.add(poi.id);
      }
    }
    return ids;
  }

  String? _poiThumbnailSourceUrl(PointOfInterest poi) {
    final thumb = poi.thumbnailUrl;
    if (thumb != null && thumb.isNotEmpty) return thumb;
    final image = poi.imageURL;
    if (image != null && image.isNotEmpty) return image;
    return null;
  }

  String _normalizePoiId(String? rawId) {
    return rawId?.trim() ?? '';
  }

  String? _coordinateKey(double latitude, double longitude) {
    if (!latitude.isFinite || !longitude.isFinite) return null;
    if (latitude.abs() > 90 || longitude.abs() > 180) return null;
    return '${latitude.toStringAsFixed(_poiAssociationCoordinatePrecision)},${longitude.toStringAsFixed(_poiAssociationCoordinatePrecision)}';
  }

  String? _poiCoordinateKey(PointOfInterest poi) {
    final latitude = double.tryParse(poi.lat);
    final longitude = double.tryParse(poi.lng);
    if (latitude == null || longitude == null) return null;
    return _coordinateKey(latitude, longitude);
  }

  Map<String, Set<String>> _buildPoiCoordinateIndex() {
    final index = <String, Set<String>>{};
    for (final poi in _pois) {
      final key = _poiCoordinateKey(poi);
      if (key == null) continue;
      (index[key] ??= <String>{}).add(poi.id);
    }
    return index;
  }

  Set<String> _buildPoiIdsWithMapContent() {
    final ids = <String>{};
    final byCoordinate = _buildPoiCoordinateIndex();

    void addPoiId(String? poiId) {
      final normalizedId = _normalizePoiId(poiId);
      if (normalizedId.isEmpty) return;
      ids.add(normalizedId);
    }

    void addByCoordinate(double latitude, double longitude) {
      final key = _coordinateKey(latitude, longitude);
      if (key == null) return;
      final matchedPoiIds = byCoordinate[key];
      if (matchedPoiIds == null || matchedPoiIds.isEmpty) return;
      ids.addAll(matchedPoiIds);
    }

    for (final challenge in _challenges) {
      addPoiId(challenge.pointOfInterestId);
      addByCoordinate(challenge.latitude, challenge.longitude);
    }
    for (final scenario in _scenarios) {
      addPoiId(scenario.pointOfInterestId);
      addByCoordinate(scenario.latitude, scenario.longitude);
    }
    for (final monster in _monsters) {
      addPoiId(monster.pointOfInterestId);
      addByCoordinate(monster.latitude, monster.longitude);
    }

    return ids;
  }

  bool _poiHasMapContent(
    PointOfInterest poi, {
    required bool isQuestCurrent,
    Set<String>? mapContentPoiIds,
  }) {
    if (poi.hasAvailableQuest || isQuestCurrent) return true;
    final contentIds = mapContentPoiIds ?? _buildPoiIdsWithMapContent();
    return contentIds.contains(poi.id);
  }

  Set<String> _currentQuestTurnInCharacterIds(QuestLogProvider questLog) {
    return questLog.quests
        .where((q) => q.readyToTurnIn && q.questGiverCharacterId != null)
        .map((q) => q.questGiverCharacterId!)
        .toSet();
  }

  Set<String> _currentQuestScenarioIds() {
    final questLog = context.read<QuestLogProvider>();
    return questLog.quests
        .where((q) => q.isAccepted)
        .map((q) => q.currentNode?.scenarioId?.trim() ?? '')
        .where((id) => id.isNotEmpty)
        .toSet();
  }

  Set<String> _currentQuestMonsterIds() {
    final questLog = context.read<QuestLogProvider>();
    return questLog.quests
        .where((q) => q.isAccepted)
        .map((q) {
          final node = q.currentNode;
          final encounterID = node?.monsterEncounterId?.trim() ?? '';
          if (encounterID.isNotEmpty) return encounterID;
          return node?.monsterId?.trim() ?? '';
        })
        .where((id) => id.isNotEmpty)
        .toSet();
  }

  Set<String> _currentQuestChallengeIds() {
    final questLog = context.read<QuestLogProvider>();
    return questLog.quests
        .where((q) => q.isAccepted)
        .map((q) => q.currentNode?.challengeId?.trim() ?? '')
        .where((id) => id.isNotEmpty)
        .toSet();
  }

  Timer? _mapLoadTimeout;

  void _startMapLoadTimeout() {
    _mapLoadTimeout?.cancel();
    _mapLoadTimeout = Timer(const Duration(seconds: 15), () {
      if (mounted && !_styleLoaded && !_mapLoadFailed) {
        debugPrint('SinglePlayer: map style load timeout (15s)');
        setState(() => _mapLoadFailed = true);
      }
    });
  }

  void _retryMap() {
    _mapLoadTimeout?.cancel();
    setState(() {
      _mapLoadFailed = false;
      _styleLoaded = false;
      _markersAdded = false;
      _addedMarkersWithEmptyDiscoveries = false;
      _hasAnimatedToUserLocation = false;
      _mapController = null;
      _mapKey++;
      _poiSymbols = [];
      _chestSymbols = [];
      _chestCircles = [];
      _chestSymbolById.clear();
      _chestCircleById.clear();
      _chestCircleOpened.clear();
      _healingFountainSymbols = [];
      _healingFountainCircles = [];
      _healingFountainSymbolById.clear();
      _healingFountainCircleById.clear();
      _chestThumbnailBytes = null;
      _chestThumbnailAdded = false;
      _scenarioSymbols = [];
      _scenarioCircles = [];
      _scenarioSymbolById.clear();
      _scenarioCircleById.clear();
      _scenarioCircleMystery.clear();
      _scenarioQuestObjective.clear();
      _monsterSymbols = [];
      _monsterCircles = [];
      _monsterSymbolById.clear();
      _monsterCircleById.clear();
      _challengeSymbols = [];
      _challengeCircles = [];
      _challengeSymbolById.clear();
      _challengeCircleById.clear();
      _scenarioMysteryThumbnailBytes = null;
      _scenarioMysteryThumbnailAdded = false;
      _monsterMysteryThumbnailBytes = null;
      _monsterMysteryThumbnailAdded = false;
      _challengeMysteryThumbnailBytes = null;
      _challengeMysteryThumbnailAdded = false;
      _questLines = [];
      _characterSymbolsById.clear();
    });
    _startMapLoadTimeout();
  }

  void _onMapStyleLoaded() {
    debugPrint('SinglePlayer: map style loaded');
    _mapLoadTimeout?.cancel();
    _mapLoadTimeout = null;
    if (!mounted) return;
    setState(() => _styleLoaded = true);
    unawaited(
      (() async {
        await _setSymbolOverlap();
        await _addPoiMarkers();
        if (_scenarioVisibilityRefreshPending) {
          _scenarioVisibilityRefreshPending = false;
          await _refreshScenarioSymbols();
          await _refreshMonsterSymbols();
          await _refreshChallengeSymbols();
        }
      })(),
    );
    if (_zones.isNotEmpty) unawaited(_addZoneBoundaries());
    unawaited(_addQuestPolygons());
    if (_questLogNeedsOverlayApply) {
      _questLogNeedsOverlayApply = false;
      _applyQuestLogOverlaysIfChanged();
    }
    _animateToUserLocationIfReady();
  }

  Future<void> _setSymbolOverlap() async {
    final c = _mapController;
    if (c == null) return;
    try {
      await c.setSymbolIconAllowOverlap(true);
      await c.setSymbolIconIgnorePlacement(true);
    } catch (e) {
      debugPrint('SinglePlayer: setSymbolOverlap error: $e');
    }
  }

  void _animateToUserLocationIfReady() {
    if (!mounted ||
        !_styleLoaded ||
        _mapLoadFailed ||
        _hasAnimatedToUserLocation)
      return;
    final c = _mapController;
    final loc = context.read<LocationProvider>().location;
    if (c == null || loc == null) return;
    final lat = loc.latitude;
    final lng = loc.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180)
      return;
    _hasAnimatedToUserLocation = true;
    setState(() {});
    Future.delayed(const Duration(milliseconds: 400), () {
      if (!mounted) return;
      final controller = _mapController;
      if (controller == null) return;
      try {
        controller.animateCamera(
          CameraUpdate.newCameraPosition(
            CameraPosition(target: LatLng(lat, lng), zoom: 15),
          ),
          duration: const Duration(milliseconds: 600),
        );
      } catch (_) {}
    });
  }

  void _centerOnUserLocation() {
    if (!_styleLoaded || _mapLoadFailed) return;
    final c = _mapController;
    final loc = context.read<LocationProvider>().location;
    if (c == null || loc == null) return;
    final lat = loc.latitude;
    final lng = loc.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180)
      return;
    try {
      c.animateCamera(
        CameraUpdate.newCameraPosition(
          CameraPosition(target: LatLng(lat, lng), zoom: 15),
        ),
        duration: const Duration(milliseconds: 500),
      );
    } catch (_) {}
  }

  void _flyToLocation(double lat, double lng) {
    final c = _mapController;
    if (c == null ||
        !lat.isFinite ||
        !lng.isFinite ||
        lat.abs() > 90 ||
        lng.abs() > 180)
      return;
    try {
      c.animateCamera(
        CameraUpdate.newCameraPosition(
          CameraPosition(target: LatLng(lat, lng), zoom: 16),
        ),
        duration: const Duration(milliseconds: 500),
      );
    } catch (_) {}
  }

  void _focusQuestPoI(PointOfInterest poi) {
    final lat = double.tryParse(poi.lat) ?? 0.0;
    final lng = double.tryParse(poi.lng) ?? 0.0;
    _flyToLocation(lat, lng);
    final hasDiscovered = context.read<DiscoveriesProvider>().hasDiscovered(
      poi.id,
    );
    _showPointOfInterestPanel(poi, hasDiscovered);
  }

  void _focusQuestTurnIn(Quest quest) {
    final questGiverId = quest.questGiverCharacterId;
    if (questGiverId == null || questGiverId.isEmpty) {
      return;
    }
    PointOfInterest? poi;
    for (final candidate in _pois) {
      final hasMatch = candidate.characters.any((c) => c.id == questGiverId);
      if (hasMatch) {
        poi = candidate;
        break;
      }
    }
    if (poi == null) {
      return;
    }
    final lat = double.tryParse(poi.lat) ?? 0.0;
    final lng = double.tryParse(poi.lng) ?? 0.0;
    _flyToLocation(lat, lng);
    _pulsePoi(lat, lng);
  }

  void _focusQuestNode(QuestNode node) {
    final poi = node.pointOfInterest;
    if (poi != null) {
      final lat = double.tryParse(poi.lat) ?? 0.0;
      final lng = double.tryParse(poi.lng) ?? 0.0;
      _flyToLocation(lat, lng);
      _pulsePoi(lat, lng);
      return;
    }
    final scenarioId = node.scenarioId?.trim() ?? '';
    if (scenarioId.isNotEmpty) {
      final scenario = _scenarioById(scenarioId);
      if (scenario != null) {
        _flyToLocation(scenario.latitude, scenario.longitude);
        _pulsePoi(scenario.latitude, scenario.longitude);
        return;
      }
    }
    final encounterId = node.monsterEncounterId?.trim() ?? '';
    if (encounterId.isNotEmpty) {
      final encounter = _monsterById(encounterId);
      if (encounter != null) {
        _flyToLocation(encounter.latitude, encounter.longitude);
        _pulsePoi(encounter.latitude, encounter.longitude);
        return;
      }
    }
    final monsterId = node.monsterId?.trim() ?? '';
    if (monsterId.isNotEmpty) {
      final encounter = _monsterEncounterByMemberMonsterId(monsterId);
      if (encounter != null) {
        _flyToLocation(encounter.latitude, encounter.longitude);
        _pulsePoi(encounter.latitude, encounter.longitude);
        return;
      }
    }
    final challengeId = node.challengeId?.trim() ?? '';
    if (challengeId.isNotEmpty) {
      final challenge = _challengeById(challengeId);
      if (challenge != null) {
        final anchor = _challengeProximityAnchor(challenge);
        _flyToLocation(anchor.latitude, anchor.longitude);
        _pulsePoi(anchor.latitude, anchor.longitude);
        return;
      }
    }
    if (node.polygon.isNotEmpty) {
      final center = _polygonCenter(node.polygon);
      _flyToLocation(center.latitude, center.longitude);
      _pulsePolygon(node.polygon);
    }
  }

  LatLng _polygonCenter(List<QuestNodePolygonPoint> polygon) {
    double latSum = 0;
    double lngSum = 0;
    var count = 0;
    for (final p in polygon) {
      latSum += p.latitude;
      lngSum += p.longitude;
      count++;
    }
    if (count == 0) return const LatLng(0, 0);
    return LatLng(latSum / count, lngSum / count);
  }

  Future<void> _pulsePoi(double lat, double lng) async {
    final c = _mapController;
    if (c == null) return;
    try {
      final circle = await c.addCircle(
        CircleOptions(
          geometry: LatLng(lat, lng),
          circleRadius: 36,
          circleColor: '#f5c542',
          circleOpacity: 0.35,
          circleStrokeWidth: 3,
          circleStrokeColor: '#f5c542',
        ),
      );
      await Future.delayed(const Duration(milliseconds: 500));
      try {
        await c.removeCircle(circle);
      } catch (_) {}
    } catch (_) {}
  }

  Future<void> _pulseDiscoveredPoi(double lat, double lng) async {
    final c = _mapController;
    if (c == null) return;
    try {
      final circle = await c.addCircle(
        CircleOptions(
          geometry: LatLng(lat, lng),
          circleRadius: 24,
          circleColor: '#f5c542',
          circleOpacity: 0.25,
          circleStrokeWidth: 2,
          circleStrokeColor: '#f5c542',
        ),
      );
      await Future.delayed(const Duration(milliseconds: 350));
      try {
        await c.removeCircle(circle);
      } catch (_) {}
    } catch (_) {}
  }

  Future<void> _pulsePolygon(List<QuestNodePolygonPoint> polygon) async {
    final c = _mapController;
    if (c == null || polygon.length < 3) return;
    final ring = polygon.map((p) => LatLng(p.latitude, p.longitude)).toList();
    if (ring.length > 1 &&
        (ring.first.latitude != ring.last.latitude ||
            ring.first.longitude != ring.last.longitude)) {
      ring.add(ring.first);
    }
    try {
      final lines = await c.addLines([
        LineOptions(
          geometry: ring,
          lineColor: '#f5c542',
          lineWidth: 8.0,
          lineOpacity: 0.8,
        ),
      ]);
      await Future.delayed(const Duration(milliseconds: 500));
      try {
        await c.removeLines(lines);
      } catch (_) {}
    } catch (_) {}
  }

  Future<void> _pulseQuestGlow(
    List<List<QuestNodePolygonPoint>> polygons,
  ) async {
    if (_isQuestGlowPulsing) return;
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    _isQuestGlowPulsing = true;
    final options = <LineOptions>[];
    for (final poly in polygons) {
      if (poly.length < 3) continue;
      final ring = poly.map((p) => LatLng(p.latitude, p.longitude)).toList();
      if (ring.length > 1 &&
          (ring.first.latitude != ring.last.latitude ||
              ring.first.longitude != ring.last.longitude)) {
        ring.add(ring.first);
      }
      options.add(
        LineOptions(
          geometry: ring,
          lineColor: '#f5c542',
          lineWidth: 8.0,
          lineOpacity: 0.35,
        ),
      );
    }
    if (options.isEmpty) {
      _isQuestGlowPulsing = false;
      return;
    }
    try {
      final lines = await c.addLines(options);
      await Future.delayed(const Duration(milliseconds: 500));
      try {
        await c.removeLines(lines);
      } catch (_) {}
    } catch (_) {}
    _isQuestGlowPulsing = false;
  }

  Future<void> _loadAll() async {
    debugPrint('SinglePlayer: _loadAll start');
    final svc = context.read<PoiService>();
    final discoveriesProvider = context.read<DiscoveriesProvider>();
    final zoneProvider = context.read<ZoneProvider>();
    try {
      await _restoreDefeatedMonsterIds();
      await _restoreDiscoveredCharacterIds();
      await discoveriesProvider.refresh();
      final zones = await svc.getZones();
      final pois = await svc.getPointsOfInterest();
      final characters = await svc.getCharacters();
      if (!mounted) return;
      debugPrint(
        'SinglePlayer: _loadAll data: zones=${zones.length} pois=${pois.length} chars=${characters.length}',
      );
      zoneProvider.setZones(zones);
      setState(() {
        _zones = zones;
        _pois = pois;
        _characters = characters;
        _markersAdded = false;
      });
      _updateSelectedZoneFromLocation();
      _requestQuestLogIfReady(force: true);
      await _loadTreasureChestsForSelectedZone();
      await _addPoiMarkers();
      await _addZoneBoundaries();
      debugPrint('SinglePlayer: _loadAll done');
    } catch (e, stackTrace) {
      debugPrint('SinglePlayer: _loadAll error: $e');
      debugPrint('SinglePlayer: _loadAll stack: $stackTrace');
      if (mounted) setState(() {});
    }
  }

  Future<void> _loadTutorialStatus({bool force = false}) async {
    if (!mounted || _tutorialStatusLoading) return;
    if (!force && _tutorialStatusChecked) {
      _maybeShowTutorialWelcome();
      return;
    }

    _tutorialStatusLoading = true;
    try {
      final status = await context.read<PoiService>().getTutorialStatus();
      if (!mounted) return;
      setState(() {
        _tutorialStatus = status;
        _tutorialStatusChecked = true;
      });
      await _ensureTutorialScenarioLoaded(status);
      await _ensureTutorialMonsterLoaded(status);
      await _syncTutorialMapModeFromStatus(status);
      _syncTutorialInventorySession(status);
      if (_tutorialReplayPending &&
          (status == null ||
              status.character == null ||
              status.dialogue.isEmpty)) {
        _tutorialReplayPending = false;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tutorial is not configured yet.')),
        );
      }
    } finally {
      _tutorialStatusLoading = false;
    }

    _maybeShowTutorialWelcome();
  }

  void _maybeShowTutorialWelcome() {
    if (!mounted || _tutorialDialogVisible || _tutorialActivationInFlight) {
      return;
    }
    final status = _tutorialStatus;
    if (status == null || status.character == null || status.dialogue.isEmpty) {
      return;
    }
    if (!status.showWelcomeDialogue && !_tutorialReplayPending) {
      return;
    }
    final location = context.read<LocationProvider>().location;
    if (location == null) return;
    unawaited(
      _showTutorialWelcomeDialog(status, forceReplay: _tutorialReplayPending),
    );
  }

  Future<void> _showTutorialWelcomeDialog(
    TutorialStatus status, {
    bool forceReplay = false,
  }) async {
    final character = status.character;
    if (character == null ||
        status.dialogue.isEmpty ||
        _tutorialDialogVisible) {
      return;
    }

    final dialogue = <DialogueMessage>[
      for (int index = 0; index < status.dialogue.length; index++)
        DialogueMessage(
          speaker: 'character',
          text: status.dialogue[index],
          order: index,
        ),
    ];
    final action = CharacterAction(
      id: 'tutorial-welcome',
      createdAt: '',
      updatedAt: '',
      characterId: character.id,
      actionType: 'tutorial',
      dialogue: dialogue,
    );

    setState(() => _tutorialDialogVisible = true);
    try {
      await showDialog<void>(
        context: context,
        useRootNavigator: true,
        barrierDismissible: false,
        builder: (dialogContext) {
          return PopScope(
            canPop: false,
            child: RpgDialogueModal(
              character: character,
              action: action,
              dialogueOverride: dialogue,
              showCloseButton: false,
              finalStepLabel: 'Begin',
              onClose: () async {
                Navigator.of(dialogContext).pop();
                await _activateTutorialScenario(
                  status,
                  forceReplay: forceReplay,
                );
              },
            ),
          );
        },
      );
    } finally {
      if (mounted) {
        setState(() => _tutorialDialogVisible = false);
      }
    }
  }

  Future<void> _activateTutorialScenario(
    TutorialStatus status, {
    bool forceReplay = false,
  }) async {
    if (!mounted || _tutorialActivationInFlight) return;

    setState(() {
      _tutorialActivationInFlight = true;
      _tutorialStatus = status.copyWith(showWelcomeDialogue: false);
    });

    try {
      final scenario = await context.read<PoiService>().activateTutorial(
        force: forceReplay,
      );
      if (!mounted) return;
      if (scenario == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Failed to start the tutorial scenario.'),
          ),
        );
        await _loadTutorialStatus(force: true);
        return;
      }

      setState(() {
        _tutorialReplayPending = false;
        _tutorialFocusedScenarioId = scenario.id;
        _tutorialFocusedMonsterEncounterId = null;
        _tutorialNormalPinsRevealInProgress = false;
        _tutorialLoadoutPendingAfterCompletionModal = false;
        _tutorialRevealPendingAfterCompletionModal = false;
        _scenarios = [
          scenario,
          ..._scenarios.where((item) => item.id != scenario.id),
        ];
      });
      await _rebuildMapPins();
      if (!mounted) return;
      _flyToLocation(scenario.latitude, scenario.longitude);
      unawaited(_pulsePoi(scenario.latitude, scenario.longitude));
      unawaited(_loadTutorialStatus(force: true));
    } finally {
      if (mounted) {
        setState(() => _tutorialActivationInFlight = false);
      }
    }
  }

  Future<void> _loadTreasureChestsForSelectedZone() async {
    final zoneId =
        context.read<ZoneProvider>().selectedZone?.id ??
        (_zones.isNotEmpty ? _zones.first.id : null);
    if (zoneId == null) {
      if (mounted) {
        setState(() {
          _treasureChests = [];
          _healingFountains = [];
          _scenarios = [];
          _monsters = [];
          _challenges = [];
        });
      }
      return;
    }
    try {
      final svc = context.read<PoiService>();
      final chestsFuture = svc.getTreasureChestsForZone(zoneId);
      final healingFountainsFuture = svc.getHealingFountainsForZone(zoneId);
      final scenariosFuture = svc.getScenariosForZone(zoneId);
      final monstersFuture = svc.getMonsterEncountersForZone(zoneId);
      final challengesFuture = svc.getChallengesForZone(zoneId);
      final chests = await chestsFuture;
      final healingFountains = await healingFountainsFuture;
      final baseScenarios = await scenariosFuture;
      final baseMonsters = await monstersFuture;
      final baseChallenges = await challengesFuture;
      if (!mounted) return;
      final currentQuestScenarioIds = _currentQuestScenarioIds();
      final currentQuestMonsterIds = _currentQuestMonsterIds();
      final currentQuestChallengeIds = _currentQuestChallengeIds();

      final scenarioById = <String, Scenario>{
        for (final scenario in baseScenarios) scenario.id: scenario,
      };
      final monsterById = <String, MonsterEncounter>{
        for (final monster in baseMonsters) monster.id: monster,
      };
      final monsterByMemberID = <String, MonsterEncounter>{};
      for (final encounter in baseMonsters) {
        for (final member in encounter.members) {
          if (member.monster.id.isNotEmpty) {
            monsterByMemberID[member.monster.id] = encounter;
          }
        }
        for (final monster in encounter.monsters) {
          if (monster.id.isNotEmpty) {
            monsterByMemberID[monster.id] = encounter;
          }
        }
      }
      final challengeById = <String, Challenge>{
        for (final challenge in baseChallenges) challenge.id: challenge,
      };

      for (final scenarioId in currentQuestScenarioIds) {
        if (scenarioById.containsKey(scenarioId)) continue;
        final scenario = await svc.getScenarioById(scenarioId);
        if (scenario == null) continue;
        if (scenario.zoneId != zoneId) continue;
        scenarioById[scenario.id] = scenario;
      }
      for (final monsterId in currentQuestMonsterIds) {
        if (monsterById.containsKey(monsterId)) continue;
        if (monsterByMemberID.containsKey(monsterId)) {
          final encounter = monsterByMemberID[monsterId];
          if (encounter != null) {
            monsterById[encounter.id] = encounter;
          }
          continue;
        }
        final encounter = await svc.getMonsterEncounterById(monsterId);
        if (encounter != null && encounter.zoneId == zoneId) {
          monsterById[encounter.id] = encounter;
          continue;
        }

        // Backward compatibility: if a quest node still carries a legacy
        // monsterId, synthesize a single-member encounter.
        final legacyMonster = await svc.getMonsterById(monsterId);
        if (legacyMonster == null) continue;
        if (legacyMonster.zoneId != zoneId) continue;
        monsterById[legacyMonster.id] = MonsterEncounter(
          id: legacyMonster.id,
          name: '${legacyMonster.name} Encounter',
          description: legacyMonster.description,
          imageUrl: legacyMonster.imageUrl,
          thumbnailUrl: legacyMonster.thumbnailUrl,
          zoneId: legacyMonster.zoneId,
          latitude: legacyMonster.latitude,
          longitude: legacyMonster.longitude,
          monsterCount: 1,
          members: [MonsterEncounterMember(slot: 1, monster: legacyMonster)],
          monsters: [legacyMonster],
        );
      }
      for (final challengeId in currentQuestChallengeIds) {
        if (challengeById.containsKey(challengeId)) continue;
        final challenge = await svc.getChallengeById(challengeId);
        if (challenge == null) continue;
        if (challenge.zoneId != zoneId) continue;
        challengeById[challenge.id] = challenge;
      }

      final scenarios = scenarioById.values
          .where(
            (scenario) =>
                (currentQuestScenarioIds.contains(scenario.id)) ||
                (!scenario.attemptedByUser &&
                    !_resolvedScenarioIds.contains(scenario.id) &&
                    !_resolvedScenarioSignatures.contains(
                      _scenarioSignature(scenario),
                    )),
          )
          .toList();
      final monsters = monsterById.values
          .where(
            (monster) =>
                currentQuestMonsterIds.contains(monster.id) ||
                !_defeatedMonsterIds.contains(monster.id),
          )
          .toList();
      final challenges = challengeById.values.toList();
      if (!mounted) return;
      setState(() {
        _treasureChests = chests
            .where((chest) => !_openedTreasureChestIds.contains(chest.id))
            .toList();
        _healingFountains = healingFountains;
        _scenarios = scenarios;
        _monsters = monsters;
        _challenges = challenges;
      });
      if (_styleLoaded && _mapController != null && _markersAdded) {
        if (_isTutorialMapFocusActive || _tutorialNormalPinsRevealInProgress) {
          await _rebuildMapPins();
          return;
        }
        await _refreshTreasureChestSymbols();
        await _refreshHealingFountainSymbols();
        await _refreshScenarioSymbols();
        await _refreshMonsterSymbols();
        await _refreshChallengeSymbols();
        await _refreshUndiscoveredPoiOpacitiesForZone();
      }
    } catch (e) {
      debugPrint('SinglePlayer: _loadTreasureChests/scenarios error: $e');
      if (mounted) {
        setState(() {
          _treasureChests = [];
          _healingFountains = [];
          _scenarios = [];
          _monsters = [];
          _challenges = [];
        });
        if (_styleLoaded && _mapController != null && _markersAdded) {
          if (_isTutorialMapFocusActive ||
              _tutorialNormalPinsRevealInProgress) {
            await _rebuildMapPins();
            return;
          }
          await _refreshTreasureChestSymbols();
          await _refreshHealingFountainSymbols();
          await _refreshScenarioSymbols();
          await _refreshMonsterSymbols();
          await _refreshChallengeSymbols();
          await _refreshUndiscoveredPoiOpacitiesForZone();
        }
      }
    }
  }

  bool _isScenarioMystery(Scenario scenario) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return true;
    final distance = _distanceMeters(
      location.latitude,
      location.longitude,
      scenario.latitude,
      scenario.longitude,
    );
    return distance > kProximityUnlockRadiusMeters;
  }

  bool _isMonsterMystery(MonsterEncounter monster) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return true;
    final distance = _distanceMeters(
      location.latitude,
      location.longitude,
      monster.latitude,
      monster.longitude,
    );
    return distance > kProximityUnlockRadiusMeters;
  }

  bool _isChallengeMystery(Challenge challenge) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return true;
    final anchor = _challengeProximityAnchor(challenge);
    final distance = _distanceMeters(
      location.latitude,
      location.longitude,
      anchor.latitude,
      anchor.longitude,
    );
    return distance > kProximityUnlockRadiusMeters;
  }

  LatLng _challengeProximityAnchor(Challenge challenge) {
    final poiId = challenge.pointOfInterestId?.trim() ?? '';
    if (poiId.isNotEmpty) {
      for (final poi in _pois) {
        if (poi.id != poiId) continue;
        final lat = double.tryParse(poi.lat);
        final lng = double.tryParse(poi.lng);
        if (lat != null &&
            lng != null &&
            lat.isFinite &&
            lng.isFinite &&
            lat.abs() <= 90 &&
            lng.abs() <= 180) {
          return LatLng(lat, lng);
        }
        break;
      }
    }
    return LatLng(challenge.latitude, challenge.longitude);
  }

  Scenario? _scenarioById(String id) {
    for (final scenario in _scenarios) {
      if (scenario.id == id) return scenario;
    }
    return null;
  }

  PointOfInterest? _poiById(String id) {
    for (final poi in _pois) {
      if (poi.id == id) return poi;
    }
    return null;
  }

  MonsterEncounter? _monsterById(String id) {
    for (final monster in _monsters) {
      if (monster.id == id) return monster;
    }
    return null;
  }

  MonsterEncounter? _monsterEncounterByMemberMonsterId(String monsterId) {
    for (final encounter in _monsters) {
      for (final member in encounter.members) {
        if (member.monster.id == monsterId) {
          return encounter;
        }
      }
      for (final monster in encounter.monsters) {
        if (monster.id == monsterId) {
          return encounter;
        }
      }
    }
    return null;
  }

  Challenge? _challengeById(String id) {
    for (final challenge in _challenges) {
      if (challenge.id == id) return challenge;
    }
    return null;
  }

  String _challengeImageHeroTag(String challengeId) =>
      'challenge-image-$challengeId';

  bool _isChallengeRepresentedByPoi(Challenge challenge) {
    final poiId = challenge.pointOfInterestId?.trim() ?? '';
    if (poiId.isEmpty) return false;
    for (final poi in _pois) {
      if (poi.id == poiId) return true;
    }
    return false;
  }

  List<Challenge> _linkedChallengesForPoi(String poiId) {
    if (poiId.trim().isEmpty) return const [];
    return _challenges.where((challenge) {
      final linkedPoiId = challenge.pointOfInterestId?.trim() ?? '';
      return linkedPoiId.isNotEmpty && linkedPoiId == poiId;
    }).toList();
  }

  HealingFountain? _healingFountainById(String id) {
    for (final fountain in _healingFountains) {
      if (fountain.id == id) return fountain;
    }
    return null;
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

  Future<void> _loadScenarioMysteryThumbnail(MapLibreMapController c) async {
    if (_scenarioMysteryThumbnailBytes == null) {
      try {
        _scenarioMysteryThumbnailBytes = await loadPoiThumbnail(
          _scenarioMysteryImageUrl,
        );
      } catch (_) {}
      _scenarioMysteryThumbnailBytes ??= await loadPoiThumbnail(
        _legacyMysteryImageUrl,
      );
    }
    if (_scenarioMysteryThumbnailBytes != null &&
        !_scenarioMysteryThumbnailAdded) {
      try {
        await c.addImage(
          'scenario_mystery_thumbnail_$_mapThumbnailVersion',
          _scenarioMysteryThumbnailBytes!,
        );
        _scenarioMysteryThumbnailAdded = true;
      } catch (_) {}
    }
  }

  Future<String?> _ensureScenarioVisibleThumbnail(
    MapLibreMapController c,
    Scenario scenario,
  ) async {
    final source = scenario.thumbnailUrl.isNotEmpty
        ? scenario.thumbnailUrl
        : scenario.imageUrl;
    if (source.isEmpty) return null;

    final imageBytes = await loadPoiThumbnail(source);
    if (imageBytes == null) return null;

    final imageId = 'scenario_${scenario.id}_$_mapThumbnailVersion';
    try {
      await c.addImage(imageId, imageBytes);
    } catch (_) {}
    return imageId;
  }

  Future<void> _refreshScenarioSymbols() {
    _scenarioRefreshSequence = _scenarioRefreshSequence.then((_) async {
      try {
        await _refreshScenarioSymbolsNow();
      } catch (e, st) {
        debugPrint('SinglePlayer: _refreshScenarioSymbols error: $e');
        debugPrint('SinglePlayer: _refreshScenarioSymbols stack: $st');
      }
    });
    return _scenarioRefreshSequence;
  }

  Future<void> _refreshScenarioSymbolsNow() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    await _loadScenarioMysteryThumbnail(c);
    final visibleScenarios = _isTutorialMapFocusActive
        ? _scenarios
              .where((scenario) => _isTutorialFocusedScenarioId(scenario.id))
              .toList()
        : _scenarios;

    // Remove untracked/duplicate scenario symbols that can appear due to
    // overlapping async refresh calls.
    final duplicateOrOrphanSymbols = <Symbol>[];
    for (final symbol in _scenarioSymbols.toList()) {
      final id = _scenarioIdFromData(symbol.data);
      if (id == null) {
        duplicateOrOrphanSymbols.add(symbol);
        continue;
      }
      final tracked = _scenarioSymbolById[id];
      if (tracked == null || !identical(tracked, symbol)) {
        duplicateOrOrphanSymbols.add(symbol);
      }
    }
    if (duplicateOrOrphanSymbols.isNotEmpty) {
      for (final symbol in duplicateOrOrphanSymbols) {
        _setQuestPoiHighlight(symbol, false);
      }
      try {
        await c.removeSymbols(duplicateOrOrphanSymbols);
      } catch (_) {}
      for (final symbol in duplicateOrOrphanSymbols) {
        _scenarioSymbols.remove(symbol);
      }
    }

    final duplicateOrOrphanCircles = <Circle>[];
    for (final circle in _scenarioCircles.toList()) {
      final id = _scenarioIdFromData(circle.data);
      if (id == null) {
        duplicateOrOrphanCircles.add(circle);
        continue;
      }
      final tracked = _scenarioCircleById[id];
      if (tracked == null || !identical(tracked, circle)) {
        duplicateOrOrphanCircles.add(circle);
      }
    }
    if (duplicateOrOrphanCircles.isNotEmpty) {
      for (final circle in duplicateOrOrphanCircles) {
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _scenarioCircles.remove(circle);
      }
    }

    final desiredIds = visibleScenarios.map((scenario) => scenario.id).toSet();
    for (final entry in _scenarioSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestPoiHighlight(entry.value, false);
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _scenarioSymbols.remove(entry.value);
        _scenarioSymbolById.remove(entry.key);
        _scenarioQuestObjective.remove(entry.key);
      }
    }
    for (final entry in _scenarioCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _scenarioCircles.remove(entry.value);
        _scenarioCircleById.remove(entry.key);
        _scenarioCircleMystery.remove(entry.key);
        _scenarioQuestObjective.remove(entry.key);
      }
    }

    final canUseImages =
        _scenarioMysteryThumbnailBytes != null &&
        _scenarioMysteryThumbnailAdded;

    for (final scenario in visibleScenarios) {
      final mystery = _isScenarioMystery(scenario);
      final isCurrentQuestScenario = _isCurrentQuestScenario(scenario.id);
      final isTutorialScenario = _isTutorialFocusedScenarioId(scenario.id);
      final shouldPulseLikeQuest = isCurrentQuestScenario || isTutorialScenario;
      final existingSymbol = _scenarioSymbolById[scenario.id];
      final existingCircle = _scenarioCircleById[scenario.id];
      final needsRefresh =
          _scenarioCircleMystery[scenario.id] != mystery ||
          existingSymbol == null ||
          _scenarioQuestObjective[scenario.id] != shouldPulseLikeQuest;

      if (canUseImages) {
        if (needsRefresh) {
          if (existingSymbol != null) {
            _setQuestPoiHighlight(existingSymbol, false);
            try {
              await c.removeSymbols([existingSymbol]);
            } catch (_) {}
            _scenarioSymbols.remove(existingSymbol);
            _scenarioSymbolById.remove(scenario.id);
          }
          if (existingCircle != null) {
            try {
              await c.removeCircle(existingCircle);
            } catch (_) {}
            _scenarioCircles.remove(existingCircle);
            _scenarioCircleById.remove(scenario.id);
          }
          var imageId = 'scenario_mystery_thumbnail_$_mapThumbnailVersion';
          if (!mystery) {
            final visibleImageId = await _ensureScenarioVisibleThumbnail(
              c,
              scenario,
            );
            if (visibleImageId != null) {
              imageId = visibleImageId;
            }
          }
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(scenario.latitude, scenario.longitude),
              iconImage: imageId,
              iconSize: 0.74,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: shouldPulseLikeQuest ? '#e1b12c' : '#000000',
              iconHaloWidth: shouldPulseLikeQuest ? 1.15 : 0.75,
              iconAnchor: 'center',
            ),
            {'type': 'scenario', 'id': scenario.id},
          );
          if (!mounted) return;
          _scenarioSymbols.add(symbol);
          _scenarioSymbolById[scenario.id] = symbol;
          _setQuestPoiHighlight(symbol, shouldPulseLikeQuest);
          _scenarioCircleMystery[scenario.id] = mystery;
          _scenarioQuestObjective[scenario.id] = shouldPulseLikeQuest;
        }
        continue;
      }

      if (existingSymbol != null) {
        _setQuestPoiHighlight(existingSymbol, false);
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _scenarioSymbols.remove(existingSymbol);
        _scenarioSymbolById.remove(scenario.id);
      }
      if (existingCircle == null ||
          _scenarioCircleMystery[scenario.id] != mystery ||
          _scenarioQuestObjective[scenario.id] != isCurrentQuestScenario) {
        if (existingCircle != null) {
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _scenarioCircles.remove(existingCircle);
          _scenarioCircleById.remove(scenario.id);
        }
        final circle = await c.addCircle(
          CircleOptions(
            geometry: LatLng(scenario.latitude, scenario.longitude),
            circleRadius: 23,
            circleOpacity: _mapMarkerStartingOpacity(1.0),
            circleColor: shouldPulseLikeQuest
                ? '#e1b12c'
                : (mystery ? '#5a5560' : '#4f8cff'),
            circleStrokeWidth: 2,
            circleStrokeColor: '#ffffff',
          ),
          {'type': 'scenario', 'id': scenario.id},
        );
        if (!mounted) return;
        _scenarioCircles.add(circle);
        _scenarioCircleById[scenario.id] = circle;
        _scenarioCircleMystery[scenario.id] = mystery;
        _scenarioQuestObjective[scenario.id] = shouldPulseLikeQuest;
      }
    }
  }

  Future<void> _loadMonsterMysteryThumbnail(MapLibreMapController c) async {
    if (_monsterMysteryThumbnailBytes == null) {
      try {
        _monsterMysteryThumbnailBytes = await loadPoiThumbnail(
          _monsterMysteryImageUrl,
        );
      } catch (_) {}
      _monsterMysteryThumbnailBytes ??= await loadPoiThumbnail(
        _legacyMysteryImageUrl,
      );
    }
    if (_monsterMysteryThumbnailBytes != null &&
        !_monsterMysteryThumbnailAdded) {
      try {
        await c.addImage(
          'monster_mystery_thumbnail_$_mapThumbnailVersion',
          _monsterMysteryThumbnailBytes!,
        );
        _monsterMysteryThumbnailAdded = true;
      } catch (_) {}
    }
  }

  Future<void> _refreshMonsterSymbols() {
    _monsterRefreshSequence = _monsterRefreshSequence.then((_) async {
      try {
        await _refreshMonsterSymbolsNow();
      } catch (e, st) {
        debugPrint('SinglePlayer: _refreshMonsterSymbols error: $e');
        debugPrint('SinglePlayer: _refreshMonsterSymbols stack: $st');
      }
    });
    return _monsterRefreshSequence;
  }

  Future<void> _refreshMonsterSymbolsNow() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    await _loadMonsterMysteryThumbnail(c);
    final visibleMonsters = _isTutorialMapFocusActive
        ? _monsters
              .where(
                (encounter) =>
                    _tutorialFocusedMonsterEncounterId?.trim().isNotEmpty ==
                        true
                    ? _isTutorialFocusedMonsterEncounterId(encounter.id)
                    : false,
              )
              .toList()
        : _monsters;

    // Remove untracked/duplicate monster symbols that can appear due to
    // overlapping async refresh calls.
    final duplicateOrOrphanSymbols = <Symbol>[];
    for (final symbol in _monsterSymbols.toList()) {
      final id = _monsterIdFromData(symbol.data);
      if (id == null) {
        duplicateOrOrphanSymbols.add(symbol);
        continue;
      }
      final tracked = _monsterSymbolById[id];
      if (tracked == null || !identical(tracked, symbol)) {
        duplicateOrOrphanSymbols.add(symbol);
      }
    }
    if (duplicateOrOrphanSymbols.isNotEmpty) {
      try {
        await c.removeSymbols(duplicateOrOrphanSymbols);
      } catch (_) {}
      for (final symbol in duplicateOrOrphanSymbols) {
        _monsterSymbols.remove(symbol);
      }
    }

    final duplicateOrOrphanCircles = <Circle>[];
    for (final circle in _monsterCircles.toList()) {
      final id = _monsterIdFromData(circle.data);
      if (id == null) {
        duplicateOrOrphanCircles.add(circle);
        continue;
      }
      final tracked = _monsterCircleById[id];
      if (tracked == null || !identical(tracked, circle)) {
        duplicateOrOrphanCircles.add(circle);
      }
    }
    if (duplicateOrOrphanCircles.isNotEmpty) {
      for (final circle in duplicateOrOrphanCircles) {
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _monsterCircles.remove(circle);
      }
    }

    final desiredIds = visibleMonsters.map((monster) => monster.id).toSet();
    for (final entry in _monsterSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _monsterSymbols.remove(entry.value);
        _monsterSymbolById.remove(entry.key);
      }
    }
    for (final entry in _monsterCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _monsterCircles.remove(entry.value);
        _monsterCircleById.remove(entry.key);
      }
    }

    for (final monster in visibleMonsters) {
      final isCurrentQuestMonster = _isCurrentQuestMonster(monster.id);
      final isTutorialMonster = _isTutorialFocusedMonsterEncounterId(
        monster.id,
      );
      final mystery = _isMonsterMystery(monster);
      String? symbolImageId;
      if (mystery) {
        if (_monsterMysteryThumbnailBytes != null &&
            _monsterMysteryThumbnailAdded) {
          symbolImageId = 'monster_mystery_thumbnail_$_mapThumbnailVersion';
        }
      } else {
        final sourceUrl = monster.thumbnailUrl.isNotEmpty
            ? monster.thumbnailUrl
            : monster.imageUrl;
        if (sourceUrl.isNotEmpty) {
          try {
            final imageBytes = await loadPoiThumbnail(sourceUrl);
            if (imageBytes != null) {
              symbolImageId = 'monster_${monster.id}_$_mapThumbnailVersion';
              try {
                await c.addImage(symbolImageId, imageBytes);
              } catch (_) {}
            }
          } catch (_) {}
        }
      }

      if (symbolImageId != null) {
        final existingCircle = _monsterCircleById[monster.id];
        if (existingCircle != null) {
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _monsterCircles.remove(existingCircle);
          _monsterCircleById.remove(monster.id);
        }

        final existingSymbol = _monsterSymbolById[monster.id];
        if (existingSymbol == null) {
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(monster.latitude, monster.longitude),
              iconImage: symbolImageId,
              iconSize: 0.78,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: (isCurrentQuestMonster || isTutorialMonster)
                  ? '#e1b12c'
                  : '#000000',
              iconHaloWidth: (isCurrentQuestMonster || isTutorialMonster)
                  ? 1.15
                  : 0.75,
              iconAnchor: 'center',
            ),
            {'type': 'monster', 'id': monster.id},
          );
          if (!mounted) return;
          _monsterSymbols.add(symbol);
          _monsterSymbolById[monster.id] = symbol;
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(monster.latitude, monster.longitude),
                iconImage: symbolImageId,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: (isCurrentQuestMonster || isTutorialMonster)
                    ? '#e1b12c'
                    : '#000000',
                iconHaloWidth: (isCurrentQuestMonster || isTutorialMonster)
                    ? 1.15
                    : 0.75,
              ),
            );
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _monsterSymbolById[monster.id];
      if (existingSymbol != null) {
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _monsterSymbols.remove(existingSymbol);
        _monsterSymbolById.remove(monster.id);
      }

      final existingCircle = _monsterCircleById[monster.id];
      if (existingCircle != null) {
        try {
          await c.removeCircle(existingCircle);
        } catch (_) {}
        _monsterCircles.remove(existingCircle);
        _monsterCircleById.remove(monster.id);
      }
      final circle = await c.addCircle(
        CircleOptions(
          geometry: LatLng(monster.latitude, monster.longitude),
          circleRadius: 24,
          circleOpacity: _mapMarkerStartingOpacity(1.0),
          circleColor: (isCurrentQuestMonster || isTutorialMonster)
              ? '#e1b12c'
              : (mystery ? '#5a5560' : '#b63f3f'),
          circleStrokeWidth: 2,
          circleStrokeColor: '#ffffff',
        ),
        {'type': 'monster', 'id': monster.id},
      );
      if (!mounted) return;
      _monsterCircles.add(circle);
      _monsterCircleById[monster.id] = circle;
    }
  }

  bool _isCurrentQuestScenario(String scenarioId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node?.scenarioId == scenarioId) {
        return true;
      }
    }
    return false;
  }

  bool _isCurrentQuestMonster(String monsterId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node?.monsterEncounterId == monsterId) {
        return true;
      }
      if (node?.monsterId == monsterId) {
        return true;
      }
      if ((node?.monsterId ?? '').isNotEmpty) {
        final encounter = _monsterEncounterByMemberMonsterId(node!.monsterId!);
        if (encounter != null && encounter.id == monsterId) {
          return true;
        }
      }
      if ((node?.monsterEncounterId ?? '').isNotEmpty) {
        final encounterId = node!.monsterEncounterId!;
        final encounter = _monsterById(encounterId);
        if (encounter != null) {
          final hasMemberMatch =
              encounter.monsters.any((m) => m.id == monsterId) ||
              encounter.members.any((m) => m.monster.id == monsterId);
          if (hasMemberMatch) {
            return true;
          }
        }
      }
      if ((node?.monsterId ?? '').isNotEmpty && monsterId == node!.monsterId) {
        return true;
      }
    }
    return false;
  }

  bool _isCurrentQuestChallenge(String challengeId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node?.challengeId == challengeId) {
        return true;
      }
    }
    return false;
  }

  Future<void> _loadChallengeMysteryThumbnail(MapLibreMapController c) async {
    if (_challengeMysteryThumbnailBytes == null) {
      try {
        _challengeMysteryThumbnailBytes = await loadPoiThumbnail(
          _challengeMysteryImageUrl,
        );
      } catch (_) {}
      _challengeMysteryThumbnailBytes ??= await loadPoiThumbnail(
        _legacyMysteryImageUrl,
      );
    }
    if (_challengeMysteryThumbnailBytes != null &&
        !_challengeMysteryThumbnailAdded) {
      try {
        await c.addImage(
          'challenge_mystery_thumbnail_$_mapThumbnailVersion',
          _challengeMysteryThumbnailBytes!,
        );
        _challengeMysteryThumbnailAdded = true;
      } catch (_) {}
    }
  }

  Future<String?> _ensureChallengeVisibleThumbnail(
    MapLibreMapController c,
    Challenge challenge,
  ) async {
    final source = challenge.thumbnailUrl.isNotEmpty
        ? challenge.thumbnailUrl
        : challenge.imageUrl;
    if (source.isEmpty) return null;

    final imageBytes = await loadPoiThumbnail(source);
    if (imageBytes == null) return null;

    final imageId = 'challenge_${challenge.id}_$_mapThumbnailVersion';
    try {
      await c.addImage(imageId, imageBytes);
    } catch (_) {}
    return imageId;
  }

  Future<void> _refreshChallengeSymbols() {
    _challengeRefreshSequence = _challengeRefreshSequence.then((_) async {
      try {
        await _refreshChallengeSymbolsNow();
      } catch (e, st) {
        debugPrint('SinglePlayer: _refreshChallengeSymbols error: $e');
        debugPrint('SinglePlayer: _refreshChallengeSymbols stack: $st');
      }
    });
    return _challengeRefreshSequence;
  }

  Future<void> _refreshChallengeSymbolsNow() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    await _loadChallengeMysteryThumbnail(c);

    final duplicateOrOrphanSymbols = <Symbol>[];
    for (final symbol in _challengeSymbols.toList()) {
      final id = _challengeIdFromData(symbol.data);
      if (id == null) {
        duplicateOrOrphanSymbols.add(symbol);
        continue;
      }
      final tracked = _challengeSymbolById[id];
      if (tracked == null || !identical(tracked, symbol)) {
        duplicateOrOrphanSymbols.add(symbol);
      }
    }
    if (duplicateOrOrphanSymbols.isNotEmpty) {
      try {
        await c.removeSymbols(duplicateOrOrphanSymbols);
      } catch (_) {}
      for (final symbol in duplicateOrOrphanSymbols) {
        _challengeSymbols.remove(symbol);
      }
    }

    final duplicateOrOrphanCircles = <Circle>[];
    for (final circle in _challengeCircles.toList()) {
      final id = _challengeIdFromData(circle.data);
      if (id == null) {
        duplicateOrOrphanCircles.add(circle);
        continue;
      }
      final tracked = _challengeCircleById[id];
      if (tracked == null || !identical(tracked, circle)) {
        duplicateOrOrphanCircles.add(circle);
      }
    }
    if (duplicateOrOrphanCircles.isNotEmpty) {
      for (final circle in duplicateOrOrphanCircles) {
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _challengeCircles.remove(circle);
      }
    }

    final standaloneChallenges = _challenges
        .where((challenge) => !_isChallengeRepresentedByPoi(challenge))
        .toList();
    final desiredIds = standaloneChallenges
        .map((challenge) => challenge.id)
        .toSet();
    for (final entry in _challengeSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _challengeSymbols.remove(entry.value);
        _challengeSymbolById.remove(entry.key);
      }
    }
    for (final entry in _challengeCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _challengeCircles.remove(entry.value);
        _challengeCircleById.remove(entry.key);
      }
    }

    for (final challenge in standaloneChallenges) {
      final isCurrentQuestChallenge = _isCurrentQuestChallenge(challenge.id);
      final mystery = _isChallengeMystery(challenge);
      String? symbolImageId;

      if (mystery) {
        if (_challengeMysteryThumbnailBytes != null &&
            _challengeMysteryThumbnailAdded) {
          symbolImageId = 'challenge_mystery_thumbnail_$_mapThumbnailVersion';
        }
      } else {
        try {
          symbolImageId = await _ensureChallengeVisibleThumbnail(c, challenge);
        } catch (_) {}
      }

      if (symbolImageId != null) {
        final existingCircle = _challengeCircleById[challenge.id];
        if (existingCircle != null) {
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _challengeCircles.remove(existingCircle);
          _challengeCircleById.remove(challenge.id);
        }

        final existingSymbol = _challengeSymbolById[challenge.id];
        final iconHaloColor = isCurrentQuestChallenge ? '#e1b12c' : '#000000';
        final iconHaloWidth = isCurrentQuestChallenge ? 1.15 : 0.75;
        if (existingSymbol == null) {
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(challenge.latitude, challenge.longitude),
              iconImage: symbolImageId,
              iconSize: 0.74,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: iconHaloColor,
              iconHaloWidth: iconHaloWidth,
              iconAnchor: 'center',
            ),
            {'type': 'challenge', 'id': challenge.id},
          );
          if (!mounted) return;
          _challengeSymbols.add(symbol);
          _challengeSymbolById[challenge.id] = symbol;
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(challenge.latitude, challenge.longitude),
                iconImage: symbolImageId,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: iconHaloColor,
                iconHaloWidth: iconHaloWidth,
              ),
            );
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _challengeSymbolById[challenge.id];
      if (existingSymbol != null) {
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _challengeSymbols.remove(existingSymbol);
        _challengeSymbolById.remove(challenge.id);
      }

      final circleColor = isCurrentQuestChallenge
          ? '#e1b12c'
          : (mystery ? '#5a5560' : '#2a9d8f');
      final existingCircle = _challengeCircleById[challenge.id];
      if (existingCircle == null) {
        final circle = await c.addCircle(
          CircleOptions(
            geometry: LatLng(challenge.latitude, challenge.longitude),
            circleRadius: 20,
            circleOpacity: _mapMarkerStartingOpacity(1.0),
            circleColor: circleColor,
            circleStrokeWidth: 2,
            circleStrokeColor: '#ffffff',
          ),
          {'type': 'challenge', 'id': challenge.id},
        );
        if (!mounted) return;
        _challengeCircles.add(circle);
        _challengeCircleById[challenge.id] = circle;
      } else {
        try {
          await c.updateCircle(
            existingCircle,
            CircleOptions(
              geometry: LatLng(challenge.latitude, challenge.longitude),
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleColor: circleColor,
            ),
          );
        } catch (_) {}
      }
    }
  }

  String? _scenarioIdFromData(dynamic raw) {
    if (raw == null || raw is! Map) return null;
    final data = Map<String, dynamic>.from(raw);
    if (data['type']?.toString() != 'scenario') return null;
    final id = data['id']?.toString();
    if (id == null || id.isEmpty) return null;
    return id;
  }

  String? _monsterIdFromData(dynamic raw) {
    if (raw == null || raw is! Map) return null;
    final data = Map<String, dynamic>.from(raw);
    if (data['type']?.toString() != 'monster') return null;
    final id = data['id']?.toString();
    if (id == null || id.isEmpty) return null;
    return id;
  }

  String? _challengeIdFromData(dynamic raw) {
    if (raw == null || raw is! Map) return null;
    final data = Map<String, dynamic>.from(raw);
    if (data['type']?.toString() != 'challenge') return null;
    final id = data['id']?.toString();
    if (id == null || id.isEmpty) return null;
    return id;
  }

  String? _extensionFromMime(String? mimeType, String? filename) {
    final name = filename ?? '';
    final dot = name.lastIndexOf('.');
    if (dot != -1 && dot < name.length - 1) {
      return name.substring(dot + 1).toLowerCase();
    }
    switch (mimeType) {
      case 'image/png':
        return 'png';
      case 'image/gif':
        return 'gif';
      case 'image/webp':
        return 'webp';
      case 'image/jpeg':
      case 'image/jpg':
        return 'jpg';
      default:
        return null;
    }
  }

  String? _mimeTypeFromFile(PlatformFile file) {
    final ext = (file.extension ?? _extensionFromMime(null, file.name))
        ?.toLowerCase();
    switch (ext) {
      case 'png':
        return 'image/png';
      case 'gif':
        return 'image/gif';
      case 'webp':
        return 'image/webp';
      case 'jpg':
      case 'jpeg':
        return 'image/jpeg';
      case 'mp4':
        return 'video/mp4';
      case 'mov':
        return 'video/quicktime';
      case 'm4v':
        return 'video/x-m4v';
      case 'webm':
        return 'video/webm';
      default:
        return null;
    }
  }

  String _formatStatLabel(String raw) {
    final trimmed = raw.trim().toLowerCase();
    if (trimmed.isEmpty) return raw;
    return trimmed[0].toUpperCase() + trimmed.substring(1);
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

  double _averageStatValue(Map<String, int> stats, List<String> tags) {
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

  void _clearQuestSubmissionRevealTimers() {
    for (final timer in _questSubmissionRevealTimers) {
      timer.cancel();
    }
    _questSubmissionRevealTimers.clear();
  }

  void _startQuestSubmissionRevealSequence() {
    _clearQuestSubmissionRevealTimers();
    const initialDelay = Duration(milliseconds: 250);
    const stepDelay = Duration(milliseconds: 320);
    var delay = initialDelay;
    for (var step = 1; step <= 5; step++) {
      _questSubmissionRevealTimers.add(
        Timer(delay, () {
          if (!mounted) return;
          setState(() => _questSubmissionRevealStep = step);
        }),
      );
      delay += stepDelay;
    }
  }

  Widget _buildRevealSection(int step, Widget child) {
    final visible = _questSubmissionRevealStep >= step;
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 240),
      switchInCurve: Curves.easeOutBack,
      switchOutCurve: Curves.easeIn,
      transitionBuilder: (child, animation) => FadeTransition(
        opacity: animation,
        child: ScaleTransition(
          scale: Tween<double>(begin: 0.96, end: 1).animate(animation),
          child: child,
        ),
      ),
      child: visible ? child : const SizedBox.shrink(),
    );
  }

  void _setupTapHandlers(MapLibreMapController c) {
    c.onCircleTapped.add((circle) {
      final raw = circle.data;
      if (raw == null) return;
      final data = Map<String, dynamic>.from(raw);
      final type = data['type'] as String?;
      final idStr = data['id']?.toString();
      if (type == null || idStr == null || idStr.isEmpty) return;
      if (type == 'character') {
        final ch = _characters.where((x) => x.id == idStr).toList();
        if (ch.isNotEmpty) {
          _showCharacterPanel(ch.first);
        }
        return;
      }
      if (type == 'chest') {
        final tc = _treasureChests.where((t) => t.id == idStr).toList();
        if (tc.isNotEmpty) {
          _showTreasureChestPanel(tc.first);
        }
        return;
      }
      if (type == 'healingFountain') {
        final fountain = _healingFountainById(idStr);
        if (fountain != null) {
          _showHealingFountainPanel(fountain);
        }
        return;
      }
      if (type == 'scenario') {
        final scenario = _scenarioById(idStr);
        if (scenario != null) {
          _showScenarioPanel(scenario);
        }
        return;
      }
      if (type == 'monster') {
        final monster = _monsterById(idStr);
        if (monster != null) {
          _showMonsterPanel(monster);
        }
        return;
      }
      if (type == 'challenge') {
        final challenge = _challengeById(idStr);
        if (challenge != null) {
          _showChallengePanel(challenge);
        }
        return;
      }
      if (type == 'poi') {
        final pois = _pois.where((p) => p.id == idStr).toList();
        if (pois.isNotEmpty && mounted) {
          _showPointOfInterestPanel(
            pois.first,
            context.read<DiscoveriesProvider>().hasDiscovered(idStr),
          );
        }
      }
    });
    c.onSymbolTapped.add((symbol) {
      try {
        debugPrint('SinglePlayer: onSymbolTapped');
        final raw = symbol.data;
        if (raw == null) {
          debugPrint('SinglePlayer: symbol tap data is null');
          return;
        }
        final data = Map<String, dynamic>.from(raw);
        if (data.isEmpty) {
          debugPrint('SinglePlayer: symbol tap data is null or not Map');
          return;
        }
        final type = data['type'] as String?;
        final idStr = data['id']?.toString();
        if (type == null || idStr == null || idStr.isEmpty) return;
        if (type == 'chest') {
          final tc = _treasureChests.where((t) => t.id == idStr).toList();
          if (tc.isNotEmpty && mounted) {
            _showTreasureChestPanel(tc.first);
          }
          return;
        }
        if (type == 'healingFountain') {
          final fountain = _healingFountainById(idStr);
          if (fountain != null && mounted) {
            _showHealingFountainPanel(fountain);
          }
          return;
        }
        if (type == 'scenario') {
          final scenario = _scenarioById(idStr);
          if (scenario != null && mounted) {
            _showScenarioPanel(scenario);
          }
          return;
        }
        if (type == 'monster') {
          final monster = _monsterById(idStr);
          if (monster != null && mounted) {
            _showMonsterPanel(monster);
          }
          return;
        }
        if (type == 'challenge') {
          final challenge = _challengeById(idStr);
          if (challenge != null && mounted) {
            _showChallengePanel(challenge);
          }
          return;
        }
        if (type == 'character') {
          final ch = _characters.where((x) => x.id == idStr).toList();
          if (ch.isNotEmpty) {
            _showCharacterPanel(ch.first);
          }
          return;
        }
        if (type == 'poiBorder') {
          final pois = _pois.where((p) => p.id == idStr).toList();
          if (pois.isNotEmpty && mounted) {
            _showPointOfInterestPanel(
              pois.first,
              context.read<DiscoveriesProvider>().hasDiscovered(idStr),
            );
          }
          return;
        }
        if (type != 'poi') {
          debugPrint('SinglePlayer: symbol tap skip type=$type id=$idStr');
          return;
        }
        final pois = _pois.where((p) => p.id == idStr).toList();
        if (pois.isEmpty) {
          debugPrint('SinglePlayer: symbol tap POI not found id=$idStr');
          return;
        }
        debugPrint(
          'SinglePlayer: symbol tap POI found id=$idStr mounted=$mounted',
        );
        if (!mounted) {
          debugPrint('SinglePlayer: symbol tap unmounted');
          return;
        }
        debugPrint('SinglePlayer: showing POI panel ${pois.first.name}');
        _showPointOfInterestPanel(
          pois.first,
          context.read<DiscoveriesProvider>().hasDiscovered(idStr),
        );
      } catch (e, st) {
        debugPrint('SinglePlayer: symbol tap error: $e');
        debugPrint('SinglePlayer: symbol tap stack: $st');
      }
    });
    c.onFillTapped.add((fill) {
      final raw = fill.data;
      if (raw == null) return;
      final data = Map<String, dynamic>.from(raw);
      final type = data['type'] as String?;
      final idStr = data['id']?.toString();
      if (type == 'zone' && idStr != null && idStr.isNotEmpty) {
        _selectZoneById(idStr);
      }
    });
    c.onLineTapped.add((line) {
      final raw = line.data;
      if (raw == null) return;
      final data = Map<String, dynamic>.from(raw);
      final type = data['type'] as String?;
      final idStr = data['id']?.toString();
      if (type == 'zone' && idStr != null && idStr.isNotEmpty) {
        _selectZoneById(idStr);
      }
    });
  }

  void _selectZone(Zone? zone) {
    final zoneProvider = context.read<ZoneProvider>();
    zoneProvider.setSelectedZone(zone, manual: true);
    if (zone != null) {
      _scheduleZoneAutoResume();
    } else {
      _zoneAutoSelectTimer?.cancel();
      zoneProvider.unlockSelection();
      _zoneWidgetController.close();
    }
  }

  void _scheduleZoneAutoResume() {
    _zoneAutoSelectTimer?.cancel();
    _zoneAutoSelectTimer = Timer(const Duration(seconds: 30), () {
      if (!mounted) return;
      final zoneProvider = context.read<ZoneProvider>();
      zoneProvider.unlockSelection();
      _updateSelectedZoneFromLocation();
    });
  }

  void _selectZoneById(String zoneId) {
    final zone = _zones.firstWhere(
      (z) => z.id == zoneId,
      orElse: () => const Zone(id: '', name: '', latitude: 0, longitude: 0),
    );
    if (zone.id.isEmpty) return;
    _selectZone(zone);
  }

  void _handleMapClick(Point<double> point, LatLng coordinates) {
    final zone = context.read<ZoneProvider>().findZoneAtCoordinate(
      coordinates.latitude,
      coordinates.longitude,
    );
    _selectZone(zone);
  }

  void _showPointOfInterestPanel(PointOfInterest poi, bool hasDiscovered) {
    Quest? questForPoi;
    QuestNode? nodeForPoi;
    final linkedChallenges = _linkedChallengesForPoi(poi.id);
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      final node = quest.currentNode;
      if (!quest.isAccepted || node?.pointOfInterest == null) continue;
      if (node!.pointOfInterest!.id == poi.id) {
        questForPoi = quest;
        nodeForPoi = node;
        break;
      }
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => PointOfInterestPanel(
        pointOfInterest: poi,
        hasDiscovered: hasDiscovered,
        quest: questForPoi,
        questNode: nodeForPoi,
        linkedChallenges: linkedChallenges,
        onClose: () => Navigator.of(context).pop(),
        onChallengeTap: (challenge) {
          Navigator.of(context).pop();
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            final activeQuestEntry = _activeQuestNodeForChallenge(challenge.id);
            if (activeQuestEntry != null) {
              _showStandaloneQuestChallengeSubmissionModal(
                activeQuestEntry.key,
                activeQuestEntry.value,
                challenge,
              );
            } else {
              _showStandaloneChallengeSubmissionModal(challenge);
            }
          });
        },
        onQuestSubmissionState: _setQuestSubmissionOverlay,
        onCharacterTap: (character) {
          Navigator.of(context).pop();
          _showCharacterPanel(character);
        },
        onUnlocked: () async {
          await context.read<DiscoveriesProvider>().refresh();
          if (!mounted) return;
          await _refreshCharacterMarkersForPoi(poi.id);
          if (!mounted) return;
          final questLog = context.read<QuestLogProvider>();
          final isQuestCurrent = _currentQuestPoiIdsForFilter(
            questLog,
          ).contains(poi.id);
          unawaited(
            _updatePoiSymbolForQuestState(
              poi.id,
              isQuestCurrent: isQuestCurrent,
            ),
          );
          final lat = double.tryParse(poi.lat) ?? 0.0;
          final lng = double.tryParse(poi.lng) ?? 0.0;
          if (lat != 0.0 || lng != 0.0) {
            unawaited(_pulseDiscoveredPoi(lat, lng));
          }
        },
      ),
    );
  }

  Future<void> _refreshDiscoveredPoiMarkers() async {
    if (!mounted) return;
    if (!_styleLoaded || _mapController == null || !_markersAdded) return;
    final discoveries = context.read<DiscoveriesProvider>();
    if (discoveries.discoveries.isEmpty) return;
    final questLog = context.read<QuestLogProvider>();
    final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
    final discoveredPoiIds = <String>{
      for (final d in discoveries.discoveries) d.pointOfInterestId,
    };
    for (final poiId in discoveredPoiIds) {
      await _updatePoiSymbolForQuestState(
        poiId,
        isQuestCurrent: questPoiIds.contains(poiId),
      );
    }
  }

  Future<void> _refreshQuestAvailabilityMarkers() async {
    if (_questAvailabilityRefreshInFlight) return;
    if (!_styleLoaded || _mapController == null || !_markersAdded) return;
    _questAvailabilityRefreshInFlight = true;
    try {
      final svc = context.read<PoiService>();
      final pois = await svc.getPointsOfInterest();
      final characters = await svc.getCharacters();
      if (!mounted) return;
      final oldPoiById = {for (final p in _pois) p.id: p};
      final oldCharById = {for (final c in _characters) c.id: c};
      setState(() {
        _pois = pois;
        _characters = characters;
      });
      final questLog = context.read<QuestLogProvider>();
      final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
      for (final poi in pois) {
        final oldPoi = oldPoiById[poi.id];
        if (oldPoi == null) continue;
        if (oldPoi.hasAvailableQuest == poi.hasAvailableQuest) continue;
        unawaited(
          _updatePoiSymbolForQuestState(
            poi.id,
            isQuestCurrent: questPoiIds.contains(poi.id),
          ),
        );
      }
      final changedCharacters = <Character>[];
      for (final ch in characters) {
        final oldCh = oldCharById[ch.id];
        if (oldCh == null) continue;
        if (oldCh.hasAvailableQuest == ch.hasAvailableQuest) continue;
        changedCharacters.add(ch);
      }
      if (changedCharacters.isNotEmpty) {
        await _updateCharacterSymbolsForState(changedCharacters);
      }
    } catch (_) {
      // Best-effort refresh; ignore failures.
    } finally {
      _questAvailabilityRefreshInFlight = false;
    }
  }

  Future<void> _addPoiMarkers() async {
    if (!_styleLoaded || _markersAdded) {
      debugPrint(
        'SinglePlayer: _addPoiMarkers skip (styleLoaded=$_styleLoaded markersAdded=$_markersAdded)',
      );
      return;
    }
    final c = _mapController;
    if (c == null) {
      debugPrint('SinglePlayer: _addPoiMarkers skip (no controller)');
      return;
    }
    _markersAdded = true;
    final markerGeneration = ++_poiMarkerGeneration;
    final tutorialMapFocused = _isTutorialMapFocusActive;
    debugPrint(
      'SinglePlayer: _addPoiMarkers start (pois=${_pois.length} chars=${_characters.length} chests=${_treasureChests.length} fountains=${_healingFountains.length} scenarios=${_scenarios.length} monsters=${_monsters.length} challenges=${_challenges.length})',
    );

    try {
      final questLog = context.read<QuestLogProvider>();
      final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
      final filters = context.read<QuestFilterProvider>();
      final tags = context.read<TagsProvider>();
      final tagFilterActive =
          filters.enableTagFilter && tags.selectedTagIds.isNotEmpty;
      final selectedTagIds = tags.selectedTagIds;
      final selectedTagNames = tagFilterActive
          ? tags.tags
                .where((t) => selectedTagIds.contains(t.id))
                .map((t) => t.name.toLowerCase())
                .toSet()
          : <String>{};
      debugPrint(
        'SinglePlayer: _addPoiMarkers questPoiIds=${questPoiIds.length}',
      );
      _lastMapFilterKey = _buildMapFilterKey();
      if (_poiSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_poiSymbols);
        } catch (_) {}
        if (!mounted) return;
        _poiSymbols.clear();
      }
      _poiSymbolById.clear();
      if (_questPoiHighlightSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_questPoiHighlightSymbols);
        } catch (_) {}
        if (!mounted) return;
        _questPoiHighlightSymbols.clear();
      }
      _questPoiPulseTimer?.cancel();
      _questPoiPulseTimer = null;
      if (_characterSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_characterSymbols);
        } catch (_) {}
        if (!mounted) return;
        _characterSymbols.clear();
      }
      _characterSymbolsById.clear();
      if (_chestSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_chestSymbols);
        } catch (_) {}
        if (!mounted) return;
        _chestSymbols.clear();
      }
      _chestSymbolById.clear();
      if (_chestCircles.isNotEmpty) {
        for (final circle in _chestCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _chestCircles.clear();
      }
      _chestCircleById.clear();
      _chestCircleOpened.clear();
      if (_healingFountainSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_healingFountainSymbols);
        } catch (_) {}
        if (!mounted) return;
        _healingFountainSymbols.clear();
      }
      _healingFountainSymbolById.clear();
      if (_healingFountainCircles.isNotEmpty) {
        for (final circle in _healingFountainCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _healingFountainCircles.clear();
      }
      _healingFountainCircleById.clear();
      if (_scenarioSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_scenarioSymbols);
        } catch (_) {}
        if (!mounted) return;
        _scenarioSymbols.clear();
      }
      _scenarioSymbolById.clear();
      if (_scenarioCircles.isNotEmpty) {
        for (final circle in _scenarioCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _scenarioCircles.clear();
      }
      _scenarioCircleById.clear();
      _scenarioCircleMystery.clear();
      if (_monsterSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_monsterSymbols);
        } catch (_) {}
        if (!mounted) return;
        _monsterSymbols.clear();
      }
      _monsterSymbolById.clear();
      if (_monsterCircles.isNotEmpty) {
        for (final circle in _monsterCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _monsterCircles.clear();
      }
      _monsterCircleById.clear();
      if (_challengeSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_challengeSymbols);
        } catch (_) {}
        if (!mounted) return;
        _challengeSymbols.clear();
      }
      _challengeSymbolById.clear();
      if (_challengeCircles.isNotEmpty) {
        for (final circle in _challengeCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _challengeCircles.clear();
      }
      _challengeCircleById.clear();
      try {
        await c.clearCircles();
      } catch (_) {}

      final placeholderFuture = loadPoiThumbnail(null).catchError((_) => null);
      final availablePlaceholderFuture = loadPoiThumbnailWithQuestMarker(
        null,
      ).catchError((_) => null);
      final characterPlaceholderFuture = loadPoiThumbnail(
        _characterMysteryImageUrl,
      ).catchError((_) => null);
      final characterAvailablePlaceholderFuture =
          loadPoiThumbnailWithQuestMarker(
            _characterMysteryImageUrl,
          ).catchError((_) => null);
      final chestFuture = loadPoiThumbnail(
        _chestImageUrl,
      ).catchError((_) => null);

      final placeholderBytes = await placeholderFuture;
      if (placeholderBytes != null) {
        try {
          await c.addImage(
            'poi_placeholder_$_mapThumbnailVersion',
            placeholderBytes,
          );
        } catch (_) {}
      }
      final availablePlaceholderBytes = await availablePlaceholderFuture;
      if (availablePlaceholderBytes != null) {
        try {
          await c.addImage(
            'poi_placeholder_available_$_mapThumbnailVersion',
            availablePlaceholderBytes,
          );
        } catch (_) {}
      }
      final characterPlaceholderBytes = await characterPlaceholderFuture;
      final characterAvailablePlaceholderBytes =
          await characterAvailablePlaceholderFuture;

      final chestBytes = await chestFuture;
      if (chestBytes != null) {
        _chestThumbnailBytes = chestBytes;
        _chestThumbnailAdded = true;
        try {
          await c.addImage('chest_thumbnail_$_mapThumbnailVersion', chestBytes);
        } catch (_) {}
      }

      if (tutorialMapFocused) {
        await _refreshScenarioSymbols();
        await _refreshMonsterSymbols();
        await _refreshChallengeSymbols();
        _ensureQuestPoiPulseTimer();
        debugPrint('SinglePlayer: _addPoiMarkers done (tutorial focus)');
        return;
      }

      for (final ch in _characters) {
        final points = ch.locations
            .map((loc) => LatLng(loc.latitude, loc.longitude))
            .where((p) => p.latitude != 0 || p.longitude != 0)
            .toList();

        if (points.isEmpty) continue;

        final hasDiscovered = _hasDiscoveredCharacter(ch);
        final thumbnailUrl = hasDiscovered ? ch.thumbnailUrl : null;
        final hasQuestAvailable = ch.hasAvailableQuest;
        Uint8List? markerBytes;
        String? markerId;
        if (thumbnailUrl != null && thumbnailUrl.isNotEmpty) {
          try {
            markerBytes = hasQuestAvailable
                ? await loadPoiThumbnailWithQuestMarker(thumbnailUrl)
                : await loadPoiThumbnail(thumbnailUrl);
            if (markerBytes != null) {
              markerId = hasQuestAvailable
                  ? 'character_${ch.id}_quest'
                  : 'character_${ch.id}';
            }
          } catch (_) {}
        }

        if (markerBytes == null) {
          markerBytes = hasQuestAvailable
              ? (characterAvailablePlaceholderBytes ??
                    availablePlaceholderBytes)
              : (characterPlaceholderBytes ?? placeholderBytes);
          markerId = hasQuestAvailable
              ? 'character_placeholder_available'
              : 'character_placeholder';
        }

        if (markerBytes != null && markerId != null) {
          final versionedId = '${markerId}_$_mapThumbnailVersion';
          try {
            await c.addImage(versionedId, markerBytes);
          } catch (_) {}
          for (final point in points) {
            final sym = await c.addSymbol(
              SymbolOptions(
                geometry: point,
                iconImage: versionedId,
                iconSize: _standardMarkerThumbnailSize,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: '#000000',
                iconHaloWidth: 0.75,
                iconAnchor: 'center',
              ),
              {'type': 'character', 'id': ch.id, 'name': ch.name},
            );
            if (!mounted) return;
            _characterSymbols.add(sym);
            (_characterSymbolsById[ch.id] ??= []).add(sym);
          }
          continue;
        }

        for (final point in points) {
          await c.addCircle(
            CircleOptions(
              geometry: point,
              circleRadius: 30,
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleColor: '#ff8833',
              circleStrokeWidth: 2,
              circleStrokeColor: '#ffffff',
            ),
            {'type': 'character', 'id': ch.id, 'name': ch.name},
          );
        }
      }
      for (final tc in _treasureChests) {
        if (tc.openedByUser == true) continue;
        if (chestBytes != null) {
          final sym = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              iconImage: 'chest_thumbnail_$_mapThumbnailVersion',
              iconSize: 0.75,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: '#000000',
              iconHaloWidth: 0.75,
              iconAnchor: 'center',
            ),
            {'type': 'chest', 'id': tc.id},
          );
          if (!mounted) return;
          _chestSymbols.add(sym);
          _chestSymbolById[tc.id] = sym;
        } else {
          final circle = await c.addCircle(
            CircleOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              circleRadius: 24,
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleColor: tc.openedByUser == true ? '#888888' : '#ffcc00',
              circleStrokeWidth: 2,
              circleStrokeColor: '#ffffff',
            ),
            {'type': 'chest', 'id': tc.id},
          );
          if (!mounted) return;
          _chestCircles.add(circle);
          _chestCircleById[tc.id] = circle;
          _chestCircleOpened[tc.id] = tc.openedByUser == true;
        }
      }
      await _refreshHealingFountainSymbols();
      await _refreshScenarioSymbols();
      await _refreshMonsterSymbols();
      await _refreshChallengeSymbols();

      final discoveries = context.read<DiscoveriesProvider>();
      final hadEmptyDiscoveries = discoveries.discoveries.isEmpty;
      final poiImageUpdates = <_PoiImageUpdate>[];
      final poiSymbolRequests = <_PoiSymbolRequest>[];
      final mapContentPoiIds = _buildPoiIdsWithMapContent();
      for (final poi in _pois) {
        final lat = double.tryParse(poi.lat) ?? 0.0;
        final lng = double.tryParse(poi.lng) ?? 0.0;
        final useRealImage = discoveries.hasDiscovered(poi.id);
        final undiscovered = !useRealImage;
        final imageUrl = useRealImage ? _poiThumbnailSourceUrl(poi) : null;
        final isQuestCurrent = questPoiIds.contains(poi.id);
        final hasMapContent = _poiHasMapContent(
          poi,
          isQuestCurrent: isQuestCurrent,
          mapContentPoiIds: mapContentPoiIds,
        );
        final hasCharacter = poi.characters.isNotEmpty;
        final baseEligible = !undiscovered || hasCharacter || hasMapContent;
        if (!baseEligible) {
          continue;
        }
        if (tagFilterActive) {
          final matchesTag = poi.tags.any(
            (tag) =>
                selectedTagIds.contains(tag.id) ||
                (tag.name.isNotEmpty &&
                    selectedTagNames.contains(tag.name.toLowerCase())),
          );
          if (!matchesTag) continue;
        }

        String? placeholderId;
        if (hasMapContent) {
          placeholderId = availablePlaceholderBytes != null
              ? 'poi_placeholder_available'
              : null;
        } else {
          placeholderId = placeholderBytes != null ? 'poi_placeholder' : null;
        }

        var addedSymbol = false;
        if (placeholderId != null) {
          final versionedId = '${placeholderId}_$_mapThumbnailVersion';
          poiSymbolRequests.add(
            _PoiSymbolRequest(
              poiId: poi.id,
              isQuestCurrent: isQuestCurrent,
              options: SymbolOptions(
                geometry: LatLng(lat, lng),
                iconImage: versionedId,
                iconSize: isQuestCurrent ? 0.82 : 0.75,
                iconHaloColor: '#000000',
                iconHaloWidth: isQuestCurrent ? 0.0 : 0.75,
                iconHaloBlur: 0.0,
                iconOpacity: _mapMarkerStartingOpacity(
                  _poiMarkerOpacity(
                    poi,
                    isQuestCurrent: isQuestCurrent,
                    undiscovered: undiscovered,
                    mapContentPoiIds: mapContentPoiIds,
                  ),
                ),
                iconAnchor: 'center',
                zIndex: 2,
              ),
              data: {'type': 'poi', 'id': poi.id, 'name': poi.name},
            ),
          );
          addedSymbol = true;
        }

        if (!addedSymbol && !useRealImage) {
          c.addCircle(
            CircleOptions(
              geometry: LatLng(lat, lng),
              circleRadius: 24,
              circleColor: '#3388ff',
              circleOpacity: _mapMarkerStartingOpacity(
                _poiMarkerOpacity(
                  poi,
                  isQuestCurrent: isQuestCurrent,
                  undiscovered: undiscovered,
                  mapContentPoiIds: mapContentPoiIds,
                ),
              ),
              circleStrokeWidth: 2,
              circleStrokeColor: isQuestCurrent ? '#f5c542' : '#ffffff',
            ),
            {'type': 'poi', 'id': poi.id, 'name': poi.name},
          );
        }

        if (useRealImage && imageUrl != null) {
          poiImageUpdates.add(
            _PoiImageUpdate(
              poi: poi,
              imageUrl: imageUrl,
              isQuestCurrent: isQuestCurrent,
              hasMapContent: hasMapContent,
              undiscovered: undiscovered,
            ),
          );
        }
      }
      if (poiSymbolRequests.isNotEmpty) {
        await _addPoiSymbolsBatched(c, markerGeneration, poiSymbolRequests);
      }
      if (poiImageUpdates.isNotEmpty) {
        if (_tutorialNormalPinsRevealInProgress) {
          await _loadPoiImagesAndUpdate(markerGeneration, poiImageUpdates);
        } else {
          unawaited(_loadPoiImagesAndUpdate(markerGeneration, poiImageUpdates));
        }
      }
      if (_scenarioVisibilityRefreshPending) {
        _scenarioVisibilityRefreshPending = false;
        await _refreshScenarioSymbols();
        await _refreshMonsterSymbols();
        await _refreshChallengeSymbols();
      }
      _ensureQuestPoiPulseTimer();
      if (mounted && hadEmptyDiscoveries) {
        setState(() => _addedMarkersWithEmptyDiscoveries = true);
      }
      debugPrint('SinglePlayer: _addPoiMarkers done');
    } catch (e, st) {
      debugPrint('SinglePlayer: _addPoiMarkers error: $e');
      debugPrint('SinglePlayer: _addPoiMarkers stack: $st');
      if (mounted) setState(() => _markersAdded = false);
    }
  }

  Future<void> _addPoiSymbolsBatched(
    MapLibreMapController c,
    int markerGeneration,
    List<_PoiSymbolRequest> requests,
  ) async {
    if (requests.isEmpty) return;
    for (var i = 0; i < requests.length; i += _poiSymbolAddBatchSize) {
      if (!mounted || markerGeneration != _poiMarkerGeneration) return;
      final end = i + _poiSymbolAddBatchSize;
      final batch = requests.sublist(
        i,
        end > requests.length ? requests.length : end,
      );
      final results = await Future.wait(
        batch.map((request) async {
          try {
            final sym = await c.addSymbol(request.options, request.data);
            return _PoiSymbolResult(request, sym);
          } catch (_) {
            return null;
          }
        }),
      );
      if (!mounted || markerGeneration != _poiMarkerGeneration) return;
      for (final result in results) {
        if (result == null) continue;
        _poiSymbols.add(result.symbol);
        _poiSymbolById[result.request.poiId] = result.symbol;
        if (result.request.isQuestCurrent) {
          _setQuestPoiHighlight(result.symbol, true);
        }
      }
    }
  }

  Future<void> _loadPoiImagesAndUpdate(
    int markerGeneration,
    List<_PoiImageUpdate> updates,
  ) async {
    if (updates.isEmpty) return;
    for (var i = 0; i < updates.length; i += _poiImageLoadBatchSize) {
      if (!mounted || markerGeneration != _poiMarkerGeneration) return;
      final end = i + _poiImageLoadBatchSize;
      final batch = updates.sublist(
        i,
        end > updates.length ? updates.length : end,
      );
      final results = await Future.wait(
        batch.map((update) => _loadPoiImageUpdate(update)),
      );
      if (!mounted || markerGeneration != _poiMarkerGeneration) return;
      final c = _mapController;
      if (c == null || !_styleLoaded) return;
      final questLog = context.read<QuestLogProvider>();
      final currentQuestPoiIds = _currentQuestPoiIdsForFilter(questLog);
      final mapContentPoiIds = _buildPoiIdsWithMapContent();
      await Future.wait(
        results.map(
          (result) => _applyPoiImageUpdate(
            c,
            markerGeneration,
            result,
            currentQuestPoiIds,
            mapContentPoiIds,
          ),
        ),
      );
    }
  }

  Future<void> _applyPoiImageUpdate(
    MapLibreMapController c,
    int markerGeneration,
    _PoiImageUpdateResult result,
    Set<String> currentQuestPoiIds,
    Set<String> mapContentPoiIds,
  ) async {
    final bytes = result.bytes;
    final imageId = result.imageId;
    if (bytes == null || imageId == null) return;
    if (!mounted || markerGeneration != _poiMarkerGeneration) return;

    final isQuestCurrentNow = currentQuestPoiIds.contains(result.update.poi.id);
    final hasMapContentNow = _poiHasMapContent(
      result.update.poi,
      isQuestCurrent: isQuestCurrentNow,
      mapContentPoiIds: mapContentPoiIds,
    );
    if (isQuestCurrentNow != result.update.isQuestCurrent ||
        hasMapContentNow != result.update.hasMapContent) {
      return;
    }

    final versionedId = '${imageId}_$_mapThumbnailVersion';
    try {
      await c.addImage(versionedId, bytes);
    } catch (_) {}
    if (!mounted || markerGeneration != _poiMarkerGeneration) return;

    final sym = _poiSymbolById[result.update.poi.id];
    if (sym == null) {
      final lat = double.tryParse(result.update.poi.lat) ?? 0.0;
      final lng = double.tryParse(result.update.poi.lng) ?? 0.0;
      try {
        final newSym = await c.addSymbol(
          SymbolOptions(
            geometry: LatLng(lat, lng),
            iconImage: versionedId,
            iconSize: isQuestCurrentNow ? 0.82 : 0.75,
            iconHaloColor: '#000000',
            iconHaloWidth: isQuestCurrentNow ? 0.0 : 0.75,
            iconHaloBlur: 0.0,
            iconOpacity: _mapMarkerStartingOpacity(
              _poiMarkerOpacity(
                result.update.poi,
                isQuestCurrent: isQuestCurrentNow,
                undiscovered: result.update.undiscovered,
                mapContentPoiIds: mapContentPoiIds,
              ),
            ),
            iconAnchor: 'center',
            zIndex: 2,
          ),
          {
            'type': 'poi',
            'id': result.update.poi.id,
            'name': result.update.poi.name,
          },
        );
        if (!mounted || markerGeneration != _poiMarkerGeneration) return;
        _poiSymbols.add(newSym);
        _poiSymbolById[result.update.poi.id] = newSym;
        _setQuestPoiHighlight(newSym, isQuestCurrentNow);
      } catch (_) {}
      return;
    }

    try {
      await c.updateSymbol(
        sym,
        SymbolOptions(
          iconImage: versionedId,
          iconSize: isQuestCurrentNow ? 0.82 : 0.75,
          iconHaloColor: '#000000',
          iconHaloWidth: isQuestCurrentNow ? 0.0 : 0.75,
          iconHaloBlur: 0.0,
          iconOpacity: _mapMarkerStartingOpacity(
            _poiMarkerOpacity(
              result.update.poi,
              isQuestCurrent: isQuestCurrentNow,
              undiscovered: result.update.undiscovered,
              mapContentPoiIds: mapContentPoiIds,
            ),
          ),
          iconAnchor: 'center',
          zIndex: 2,
        ),
      );
      _setQuestPoiHighlight(sym, isQuestCurrentNow);
    } catch (_) {}
  }

  Future<_PoiImageUpdateResult> _loadPoiImageUpdate(
    _PoiImageUpdate update,
  ) async {
    Uint8List? imageBytes;
    String? imageId;
    if (update.hasMapContent) {
      imageBytes = await loadPoiThumbnailWithQuestMarker(update.imageUrl);
      if (imageBytes != null) {
        imageId = 'poi_${update.poi.id}_activity';
      }
    } else {
      imageBytes = await loadPoiThumbnail(update.imageUrl);
      if (imageBytes != null) {
        imageId = 'poi_${update.poi.id}';
      }
    }
    return _PoiImageUpdateResult(update, imageId, imageBytes);
  }

  Future<void> _refreshTreasureChestSymbols() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;

    if (_chestThumbnailBytes == null) {
      try {
        _chestThumbnailBytes = await loadPoiThumbnail(_chestImageUrl);
      } catch (_) {}
    }
    if (_chestThumbnailBytes != null && !_chestThumbnailAdded) {
      try {
        await c.addImage(
          'chest_thumbnail_$_mapThumbnailVersion',
          _chestThumbnailBytes!,
        );
        _chestThumbnailAdded = true;
      } catch (_) {}
    }

    final useImage = _chestThumbnailBytes != null;
    final desiredIds = _treasureChests
        .where((t) => t.openedByUser != true)
        .map((t) => t.id)
        .toSet();

    for (final entry in _chestSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key) || !useImage) {
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _chestSymbols.remove(entry.value);
        _chestSymbolById.remove(entry.key);
      }
    }
    for (final entry in _chestCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key) || useImage) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _chestCircles.remove(entry.value);
        _chestCircleById.remove(entry.key);
        _chestCircleOpened.remove(entry.key);
      }
    }

    for (final tc in _treasureChests) {
      if (tc.openedByUser == true) continue;
      if (useImage) {
        final existing = _chestSymbolById[tc.id];
        if (existing == null) {
          final sym = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              iconImage: 'chest_thumbnail_$_mapThumbnailVersion',
              iconSize: 0.75,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: '#000000',
              iconHaloWidth: 0.75,
              iconAnchor: 'center',
            ),
            {'type': 'chest', 'id': tc.id},
          );
          if (!mounted) return;
          _chestSymbols.add(sym);
          _chestSymbolById[tc.id] = sym;
        } else {
          try {
            await c.updateSymbol(
              existing,
              SymbolOptions(
                geometry: LatLng(tc.latitude, tc.longitude),
                iconOpacity: _mapMarkerStartingOpacity(1.0),
              ),
            );
          } catch (_) {}
        }
      } else {
        final existing = _chestCircleById[tc.id];
        final opened = tc.openedByUser == true;
        final needsUpdate =
            existing == null || _chestCircleOpened[tc.id] != opened;
        if (needsUpdate) {
          if (existing != null) {
            try {
              await c.removeCircle(existing);
            } catch (_) {}
            _chestCircles.remove(existing);
            _chestCircleById.remove(tc.id);
          }
          final circle = await c.addCircle(
            CircleOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              circleRadius: 24,
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleColor: opened ? '#888888' : '#ffcc00',
              circleStrokeWidth: 2,
              circleStrokeColor: '#ffffff',
            ),
            {'type': 'chest', 'id': tc.id},
          );
          if (!mounted) return;
          _chestCircles.add(circle);
          _chestCircleById[tc.id] = circle;
          _chestCircleOpened[tc.id] = opened;
        }
      }
    }
  }

  Future<void> _refreshHealingFountainSymbols() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;

    final desiredIds = _healingFountains.map((fountain) => fountain.id).toSet();

    for (final entry in _healingFountainSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _healingFountainSymbols.remove(entry.value);
        _healingFountainSymbolById.remove(entry.key);
      }
    }
    for (final entry in _healingFountainCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _healingFountainCircles.remove(entry.value);
        _healingFountainCircleById.remove(entry.key);
      }
    }

    for (final fountain in _healingFountains) {
      if (fountain.id.isEmpty) continue;
      if (!fountain.latitude.isFinite || !fountain.longitude.isFinite) continue;
      if (fountain.latitude == 0.0 && fountain.longitude == 0.0) continue;

      final haloColor = fountain.availableNow ? '#2ecc71' : '#888888';
      final discovered = fountain.discovered;
      final effectiveHaloColor = discovered ? haloColor : '#000000';
      final circleColor = discovered
          ? (fountain.availableNow ? '#2ecc71' : '#7f8c8d')
          : '#3388ff';
      final imageSource = discovered && fountain.thumbnailUrl.trim().isNotEmpty
          ? fountain.thumbnailUrl.trim()
          : _healingFountainFallbackImageUrl;
      Uint8List? imageBytes;
      try {
        imageBytes = await loadPoiThumbnail(imageSource);
      } catch (_) {}

      if (imageBytes != null) {
        final imageId = 'healing_fountain_${fountain.id}_$_mapThumbnailVersion';
        try {
          await c.addImage(imageId, imageBytes);
        } catch (_) {}

        final existingCircle = _healingFountainCircleById[fountain.id];
        if (existingCircle != null) {
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _healingFountainCircles.remove(existingCircle);
          _healingFountainCircleById.remove(fountain.id);
        }

        final existingSymbol = _healingFountainSymbolById[fountain.id];
        if (existingSymbol == null) {
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(fountain.latitude, fountain.longitude),
              iconImage: imageId,
              iconSize: 0.8,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: effectiveHaloColor,
              iconHaloWidth: 1.0,
              iconAnchor: 'center',
            ),
            {'type': 'healingFountain', 'id': fountain.id},
          );
          if (!mounted) return;
          _healingFountainSymbols.add(symbol);
          _healingFountainSymbolById[fountain.id] = symbol;
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(fountain.latitude, fountain.longitude),
                iconImage: imageId,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: effectiveHaloColor,
                iconHaloWidth: 1.0,
              ),
            );
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _healingFountainSymbolById[fountain.id];
      if (existingSymbol != null) {
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _healingFountainSymbols.remove(existingSymbol);
        _healingFountainSymbolById.remove(fountain.id);
      }

      final existingCircle = _healingFountainCircleById[fountain.id];
      if (existingCircle == null) {
        final circle = await c.addCircle(
          CircleOptions(
            geometry: LatLng(fountain.latitude, fountain.longitude),
            circleRadius: 20,
            circleOpacity: _mapMarkerStartingOpacity(1.0),
            circleColor: circleColor,
            circleStrokeWidth: 2,
            circleStrokeColor: '#ffffff',
          ),
          {'type': 'healingFountain', 'id': fountain.id},
        );
        if (!mounted) return;
        _healingFountainCircles.add(circle);
        _healingFountainCircleById[fountain.id] = circle;
      } else {
        try {
          await c.updateCircle(
            existingCircle,
            CircleOptions(
              geometry: LatLng(fountain.latitude, fountain.longitude),
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleColor: circleColor,
            ),
          );
        } catch (_) {}
      }
    }
  }

  List<LatLng> _zoneRing(Zone z) {
    final ring = z.ring;
    if (ring != null && ring.isNotEmpty) {
      final list = ring.map((c) => LatLng(c.latitude, c.longitude)).toList();
      if (list.length > 1 &&
          (list.first.latitude != list.last.latitude ||
              list.first.longitude != list.last.longitude)) {
        list.add(list.first);
      }
      return list;
    }
    // boundary is a WKT string, not usable directly - rely on points/boundaryCoords
    return [];
  }

  bool _isPointInZone(Zone zone, double lat, double lng) {
    final ring = zone.ring;
    if (ring == null || ring.length < 3) return false;

    var inside = false;
    var j = ring.length - 1;
    for (var i = 0; i < ring.length; i++) {
      final xi = ring[i].longitude;
      final yi = ring[i].latitude;
      final xj = ring[j].longitude;
      final yj = ring[j].latitude;
      final intersect =
          ((yi > lat) != (yj > lat)) &&
          (lng < (xj - xi) * (lat - yi) / (yj - yi + 0.0) + xi);
      if (intersect) inside = !inside;
      j = i;
    }
    return inside;
  }

  double _poiMarkerOpacity(
    PointOfInterest poi, {
    required bool isQuestCurrent,
    required bool undiscovered,
    Set<String>? mapContentPoiIds,
  }) {
    if (isQuestCurrent || !undiscovered) return 1.0;
    final hasMapContent = _poiHasMapContent(
      poi,
      isQuestCurrent: isQuestCurrent,
      mapContentPoiIds: mapContentPoiIds,
    );
    if (!hasMapContent) return 0.0;
    if (!mounted) return 0.5;

    final selectedZone = context.read<ZoneProvider>().selectedZone;
    if (selectedZone == null) return 0.5;

    final lat = double.tryParse(poi.lat);
    final lng = double.tryParse(poi.lng);
    if (lat == null || lng == null) return 0.5;

    return _isPointInZone(selectedZone, lat, lng) ? 1.0 : 0.5;
  }

  String _earthToneForZone(Zone zone, {int salt = 0}) {
    final seed =
        '${zone.id}|${zone.name}|${zone.latitude.toStringAsFixed(4)}|${zone.longitude.toStringAsFixed(4)}|$salt';
    int hash = 0;
    for (final code in seed.codeUnits) {
      hash = 0x1fffffff & (hash + code);
      hash = 0x1fffffff & (hash + ((0x0007ffff & hash) << 10));
      hash ^= (hash >> 6);
    }
    hash = 0x1fffffff & (hash + ((0x03ffffff & hash) << 3));
    hash ^= (hash >> 11);
    hash = 0x1fffffff & (hash + ((0x00003fff & hash) << 15));
    final hue = 20 + (hash % 50); // 20–69
    final saturation = 30 + ((hash >> 8) % 18); // 30–47
    final lightness = 40 + ((hash >> 16) % 18); // 40–57
    final color = HSLColor.fromAHSL(
      1,
      hue.toDouble(),
      saturation / 100,
      lightness / 100,
    ).toColor();
    final hex = color.value.toRadixString(16).padLeft(8, '0').substring(2);
    return '#$hex';
  }

  Future<void> _addZoneBoundaries() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) {
      debugPrint(
        'SinglePlayer: _addZoneBoundaries skip (controller=${c != null} styleLoaded=$_styleLoaded)',
      );
      return;
    }
    if (_zoneFills.isNotEmpty) {
      try {
        await c.removeFills(_zoneFills);
      } catch (_) {}
      if (!mounted) return;
      _zoneFills.clear();
    }
    if (_zoneLines.isNotEmpty) {
      try {
        await c.removeLines(_zoneLines);
      } catch (_) {}
      if (!mounted) return;
      _zoneLines.clear();
    }
    final options = <LineOptions>[];
    final lineData = <Map>[];
    final fillOptions = <FillOptions>[];
    final fillData = <Map>[];
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    for (var i = 0; i < _zones.length; i++) {
      final z = _zones[i];
      final ring = _zoneRing(z);
      if (ring.length < 2) continue;
      if (selectedZoneId == null || selectedZoneId != z.id) {
        fillOptions.add(
          FillOptions(
            geometry: [ring],
            fillColor: _earthToneForZone(z, salt: i),
            fillOpacity: 0.4,
          ),
        );
        fillData.add({'type': 'zone', 'id': z.id});
      }
      options.add(
        LineOptions(
          geometry: ring,
          lineColor: '#000000',
          lineWidth: 7.0,
          lineOpacity: 0.18,
          lineBlur: 1.6,
          lineJoin: 'round',
        ),
      );
      lineData.add({'type': 'zone', 'id': z.id});
      options.add(
        LineOptions(
          geometry: ring,
          lineColor: '#000000',
          lineWidth: 2.8,
          lineOpacity: 0.95,
          lineJoin: 'round',
        ),
      );
      lineData.add({'type': 'zone', 'id': z.id});
    }
    debugPrint(
      'SinglePlayer: _addZoneBoundaries zones=${_zones.length} rings=${options.length}',
    );
    if (options.isEmpty && fillOptions.isEmpty) return;
    try {
      if (fillOptions.isNotEmpty) {
        final fills = await c.addFills(fillOptions, fillData);
        if (!mounted) return;
        _zoneFills.addAll(fills);
      }
      final lines = await c.addLines(options, lineData);
      if (!mounted) return;
      _zoneLines.addAll(lines);
      debugPrint(
        'SinglePlayer: _addZoneBoundaries added ${lines.length} lines',
      );
    } catch (e, st) {
      debugPrint('SinglePlayer: _addZoneBoundaries error: $e');
      debugPrint('SinglePlayer: _addZoneBoundaries stack: $st');
    }
  }

  Future<void> _addQuestPolygons() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) {
      return;
    }
    _questGlowTimer?.cancel();
    _questGlowTimer = null;
    if (_questLines.isNotEmpty) {
      try {
        await c.removeLines(_questLines);
      } catch (_) {}
      if (!mounted) return;
      _questLines.clear();
    }
    if (_questFills.isNotEmpty) {
      try {
        await c.removeFills(_questFills);
      } catch (_) {}
      if (!mounted) return;
      _questFills.clear();
    }

    final questLog = context.read<QuestLogProvider>();
    final polygons = questLog.currentNodePolygons;
    if (polygons.isEmpty) return;

    final options = <LineOptions>[];
    final fillOptions = <FillOptions>[];
    for (final poly in polygons) {
      if (poly.length < 3) continue;
      final ring = poly.map((p) => LatLng(p.latitude, p.longitude)).toList();
      if (ring.length > 1 &&
          (ring.first.latitude != ring.last.latitude ||
              ring.first.longitude != ring.last.longitude)) {
        ring.add(ring.first);
      }
      fillOptions.add(
        FillOptions(geometry: [ring], fillColor: '#f5c542', fillOpacity: 0.5),
      );
      options.add(
        LineOptions(
          geometry: ring,
          lineColor: '#f5c542',
          lineWidth: 3.0,
          lineOpacity: 1.0,
        ),
      );
    }

    if (options.isEmpty || fillOptions.isEmpty) return;
    try {
      final fills = await c.addFills(fillOptions);
      if (!mounted) return;
      _questFills.addAll(fills);
      final lines = await c.addLines(options);
      if (!mounted) return;
      _questLines.addAll(lines);
      _questGlowTimer = Timer.periodic(const Duration(milliseconds: 1400), (_) {
        unawaited(_pulseQuestGlow(polygons));
      });
    } catch (e, st) {
      debugPrint('SinglePlayer: _addQuestPolygons error: $e');
      debugPrint('SinglePlayer: _addQuestPolygons stack: $st');
    }
  }

  void _setQuestPoiHighlight(Symbol sym, bool enabled) {
    if (enabled) {
      if (!_questPoiHighlightSymbols.contains(sym)) {
        _questPoiHighlightSymbols.add(sym);
      }
    } else {
      _questPoiHighlightSymbols.remove(sym);
    }
    _ensureQuestPoiPulseTimer();
  }

  Future<void> _refreshUndiscoveredPoiOpacitiesForZone() async {
    final c = _mapController;
    if (c == null || !_styleLoaded || !_markersAdded) return;

    final questPoiIds = _currentQuestPoiIdsForFilter(
      context.read<QuestLogProvider>(),
    );
    final discoveries = context.read<DiscoveriesProvider>();
    final mapContentPoiIds = _buildPoiIdsWithMapContent();

    for (final poi in _pois) {
      final symbol = _poiSymbolById[poi.id];
      if (symbol == null) continue;

      final isQuestCurrent = questPoiIds.contains(poi.id);
      final undiscovered = !discoveries.hasDiscovered(poi.id);
      if (!undiscovered || isQuestCurrent) continue;

      final opacity = _poiMarkerOpacity(
        poi,
        isQuestCurrent: isQuestCurrent,
        undiscovered: undiscovered,
        mapContentPoiIds: mapContentPoiIds,
      );
      try {
        await c.updateSymbol(symbol, SymbolOptions(iconOpacity: opacity));
      } catch (_) {}
    }
  }

  void _ensureQuestPoiPulseTimer() {
    if (_tutorialNormalPinsRevealInProgress) {
      _questPoiPulseTimer?.cancel();
      _questPoiPulseTimer = null;
      return;
    }
    if (_questPoiHighlightSymbols.isEmpty) {
      _questPoiPulseTimer?.cancel();
      _questPoiPulseTimer = null;
      return;
    }
    _questPoiPulseTimer ??= Timer.periodic(const Duration(milliseconds: 1200), (
      _,
    ) {
      unawaited(_pulseQuestPoiBorders());
    });
  }

  Future<void> _updatePoiSymbolForQuestState(
    String poiId, {
    required bool isQuestCurrent,
  }) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    PointOfInterest? poi;
    for (final candidate in _pois) {
      if (candidate.id == poiId) {
        poi = candidate;
        break;
      }
    }
    if (poi == null) return;

    final filters = context.read<QuestFilterProvider>();
    final tags = context.read<TagsProvider>();
    final tagFilterActive =
        filters.enableTagFilter && tags.selectedTagIds.isNotEmpty;
    final selectedTagIds = tags.selectedTagIds;
    final selectedTagNames = tagFilterActive
        ? tags.tags
              .where((t) => selectedTagIds.contains(t.id))
              .map((t) => t.name.toLowerCase())
              .toSet()
        : <String>{};
    final discoveries = context.read<DiscoveriesProvider>();
    final useRealImage = discoveries.hasDiscovered(poi.id);
    final undiscovered = !useRealImage;
    final imageUrl = useRealImage ? _poiThumbnailSourceUrl(poi) : null;
    final mapContentPoiIds = _buildPoiIdsWithMapContent();
    final hasMapContent = _poiHasMapContent(
      poi,
      isQuestCurrent: isQuestCurrent,
      mapContentPoiIds: mapContentPoiIds,
    );
    final hasCharacter = poi.characters.isNotEmpty;
    final baseEligible = !undiscovered || hasCharacter || hasMapContent;
    final tagMatch =
        !tagFilterActive ||
        poi.tags.any(
          (tag) =>
              selectedTagIds.contains(tag.id) ||
              (tag.name.isNotEmpty &&
                  selectedTagNames.contains(tag.name.toLowerCase())),
        );
    if (!(baseEligible && tagMatch)) {
      final existing = _poiSymbolById.remove(poiId);
      if (existing != null) {
        _poiSymbols.remove(existing);
        _setQuestPoiHighlight(existing, false);
        try {
          await c.removeSymbols([existing]);
        } catch (_) {}
      }
      return;
    }

    final sym = _poiSymbolById[poiId];

    late final String imageId;
    Uint8List? imageBytes;
    if (hasMapContent) {
      imageBytes = await loadPoiThumbnailWithQuestMarker(imageUrl);
      imageId = imageBytes != null
          ? 'poi_${poi.id}_activity'
          : 'poi_placeholder_available';
    } else if (useRealImage) {
      imageBytes = await loadPoiThumbnail(imageUrl);
      imageId = imageBytes != null ? 'poi_${poi.id}' : 'poi_placeholder';
    } else {
      imageId = 'poi_placeholder';
    }
    final versionedId = '${imageId}_$_mapThumbnailVersion';
    if (imageBytes == null) {
      try {
        if (imageId == 'poi_placeholder') {
          imageBytes = await loadPoiThumbnail(null);
        } else if (imageId == 'poi_placeholder_available') {
          imageBytes = await loadPoiThumbnailWithQuestMarker(null);
        }
      } catch (_) {}
    }
    if (imageBytes != null) {
      try {
        await c.addImage(versionedId, imageBytes);
      } catch (_) {}
    }

    if (sym == null) {
      final lat = double.tryParse(poi.lat) ?? 0.0;
      final lng = double.tryParse(poi.lng) ?? 0.0;
      try {
        final newSym = await c.addSymbol(
          SymbolOptions(
            geometry: LatLng(lat, lng),
            iconImage: versionedId,
            iconSize: isQuestCurrent ? 0.82 : 0.75,
            iconHaloColor: '#000000',
            iconHaloWidth: isQuestCurrent ? 0.0 : 0.75,
            iconOpacity: _poiMarkerOpacity(
              poi,
              isQuestCurrent: isQuestCurrent,
              undiscovered: undiscovered,
              mapContentPoiIds: mapContentPoiIds,
            ),
            iconAnchor: 'center',
            zIndex: 2,
          ),
          {'type': 'poi', 'id': poi.id, 'name': poi.name},
        );
        if (!mounted) return;
        _poiSymbols.add(newSym);
        _poiSymbolById[poi.id] = newSym;
        _setQuestPoiHighlight(newSym, isQuestCurrent);
      } catch (_) {}
      return;
    }

    try {
      await c.updateSymbol(
        sym,
        SymbolOptions(
          iconImage: versionedId,
          iconSize: isQuestCurrent ? 0.82 : 0.75,
          iconHaloColor: '#000000',
          iconHaloWidth: isQuestCurrent ? 0.0 : 0.75,
          iconOpacity: _poiMarkerOpacity(
            poi,
            isQuestCurrent: isQuestCurrent,
            undiscovered: undiscovered,
            mapContentPoiIds: mapContentPoiIds,
          ),
          iconAnchor: 'center',
          zIndex: 2,
        ),
      );
      _setQuestPoiHighlight(sym, isQuestCurrent);
    } catch (_) {}
  }

  String _characterDiscoveryPoiId(Character character) {
    return character.pointOfInterestId?.trim() ?? '';
  }

  bool _isValidCharacterCoordinate(double lat, double lng) {
    if (!lat.isFinite || !lng.isFinite) return false;
    if (lat.abs() > 90 || lng.abs() > 180) return false;
    return lat != 0 || lng != 0;
  }

  bool _isCharacterDiscoveryManaged(Character character) {
    if (_characterDiscoveryPoiId(character).isNotEmpty) return true;
    if (character.pointOfInterestLat != null &&
        character.pointOfInterestLng != null &&
        _isValidCharacterCoordinate(
          character.pointOfInterestLat!,
          character.pointOfInterestLng!,
        )) {
      return true;
    }
    if (character.locations.any(
      (loc) => _isValidCharacterCoordinate(loc.latitude, loc.longitude),
    )) {
      return true;
    }
    return false;
  }

  bool _hasDiscoveredCharacter(Character character) {
    if (!_isCharacterDiscoveryManaged(character)) return true;
    final poiId = _characterDiscoveryPoiId(character);
    if (poiId.isNotEmpty) {
      return context.read<DiscoveriesProvider>().hasDiscovered(poiId);
    }
    return _discoveredCharacterIds.contains(character.id);
  }

  Future<void> _updateCharacterSymbolsForState(
    List<Character> characters,
  ) async {
    for (final ch in characters) {
      await _updateCharacterSymbolForState(ch);
    }
  }

  Future<void> _updateCharacterSymbolForState(Character ch) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final symbols = _characterSymbolsById[ch.id];
    if (symbols == null || symbols.isEmpty) return;
    final hasDiscovered = _hasDiscoveredCharacter(ch);
    final thumbnailUrl = hasDiscovered ? ch.thumbnailUrl : null;
    final hasQuestAvailable = ch.hasAvailableQuest;
    Uint8List? imageBytes;
    String? imageId;
    if (thumbnailUrl != null && thumbnailUrl.isNotEmpty) {
      try {
        imageBytes = hasQuestAvailable
            ? await loadPoiThumbnailWithQuestMarker(thumbnailUrl)
            : await loadPoiThumbnail(thumbnailUrl);
        if (imageBytes != null) {
          imageId = hasQuestAvailable
              ? 'character_${ch.id}_quest'
              : 'character_${ch.id}';
        }
      } catch (_) {}
    }
    if (imageBytes == null) {
      try {
        imageBytes = hasQuestAvailable
            ? await loadPoiThumbnailWithQuestMarker(_characterMysteryImageUrl)
            : await loadPoiThumbnail(_characterMysteryImageUrl);
        if (imageBytes != null) {
          imageId = hasQuestAvailable
              ? 'character_placeholder_available'
              : 'character_placeholder';
        }
      } catch (_) {}
    }
    if (imageBytes == null) {
      try {
        imageBytes = hasQuestAvailable
            ? await loadPoiThumbnailWithQuestMarker(null)
            : await loadPoiThumbnail(null);
        if (imageBytes != null) {
          imageId = hasQuestAvailable
              ? 'character_placeholder_available'
              : 'character_placeholder';
        }
      } catch (_) {}
    }
    if (imageBytes == null || imageId == null) return;
    final versionedId = '${imageId}_$_mapThumbnailVersion';
    try {
      await c.addImage(versionedId, imageBytes);
    } catch (_) {}
    for (final sym in symbols) {
      try {
        await c.updateSymbol(
          sym,
          SymbolOptions(
            iconImage: versionedId,
            iconSize: _standardMarkerThumbnailSize,
            iconHaloColor: '#000000',
            iconHaloWidth: 0.75,
            iconAnchor: 'center',
          ),
        );
      } catch (_) {}
    }
  }

  Future<void> _refreshCharacterDiscoveryMarkers() async {
    if (!_styleLoaded || _mapController == null || !_markersAdded) return;
    for (final ch in _characters) {
      if (!_isCharacterDiscoveryManaged(ch)) continue;
      await _updateCharacterSymbolForState(ch);
    }
  }

  Future<void> _refreshCharacterMarkersForPoi(String poiId) async {
    if (poiId.isEmpty) return;
    final linkedCharacters = _characters
        .where((character) => _characterDiscoveryPoiId(character) == poiId)
        .toList();
    if (linkedCharacters.isEmpty) return;
    await _updateCharacterSymbolsForState(linkedCharacters);
  }

  Future<void> _pulseQuestPoiBorders() async {
    if (_tutorialNormalPinsRevealInProgress) return;
    if (_isQuestPoiPulseActive) return;
    if (_questPoiHighlightSymbols.isEmpty) return;
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    _isQuestPoiPulseActive = true;
    _questPoiPulseUp = !_questPoiPulseUp;
    final targetSize = _questPoiPulseUp ? 0.86 : 0.78;
    final targetOpacity = _questPoiPulseUp ? 1.0 : 0.85;
    try {
      for (final border in _questPoiHighlightSymbols) {
        await c.updateSymbol(
          border,
          SymbolOptions(iconSize: targetSize, iconOpacity: targetOpacity),
        );
      }
    } catch (_) {}
    _isQuestPoiPulseActive = false;
  }

  Future<void> _updateNormalPinOpacities(
    MapLibreMapController c,
    double progress, {
    required Set<String> questPoiIds,
    required DiscoveriesProvider discoveries,
    required Set<String> mapContentPoiIds,
  }) async {
    for (final entry in _poiSymbolById.entries) {
      final poiId = entry.key;
      final poi = _poiById(poiId);
      if (poi == null) continue;
      final isQuestCurrent = questPoiIds.contains(poi.id);
      final undiscovered = !discoveries.hasDiscovered(poi.id);
      final targetOpacity = _poiMarkerOpacity(
        poi,
        isQuestCurrent: isQuestCurrent,
        undiscovered: undiscovered,
        mapContentPoiIds: mapContentPoiIds,
      );
      try {
        await c.updateSymbol(
          entry.value,
          SymbolOptions(iconOpacity: targetOpacity * progress),
        );
      } catch (_) {}
    }

    Future<void> updateSymbolOpacity(Iterable<Symbol> symbols) async {
      for (final symbol in symbols) {
        try {
          await c.updateSymbol(symbol, SymbolOptions(iconOpacity: progress));
        } catch (_) {}
      }
    }

    Future<void> updateCircleOpacity(Map<String, Circle> circles) async {
      for (final circle in circles.values) {
        try {
          await c.updateCircle(circle, CircleOptions(circleOpacity: progress));
        } catch (_) {}
      }
    }

    await updateSymbolOpacity(_characterSymbols);
    await updateSymbolOpacity(_chestSymbolById.values);
    await updateCircleOpacity(_chestCircleById);
    await updateSymbolOpacity(_healingFountainSymbolById.values);
    await updateCircleOpacity(_healingFountainCircleById);
    await updateSymbolOpacity(_scenarioSymbolById.values);
    await updateCircleOpacity(_scenarioCircleById);
    await updateSymbolOpacity(_monsterSymbolById.values);
    await updateCircleOpacity(_monsterCircleById);
    await updateSymbolOpacity(_challengeSymbolById.values);
    await updateCircleOpacity(_challengeCircleById);
  }

  bool _isInsidePolygon(
    double lat,
    double lng,
    List<QuestNodePolygonPoint> polygon,
  ) {
    if (polygon.length < 3) return false;
    var inside = false;
    for (var i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
      final xi = polygon[i].longitude;
      final yi = polygon[i].latitude;
      final xj = polygon[j].longitude;
      final yj = polygon[j].latitude;
      final intersect =
          ((yi > lat) != (yj > lat)) &&
          (lng < (xj - xi) * (lat - yi) / (yj - yi + 0.0) + xi);
      if (intersect) inside = !inside;
    }
    return inside;
  }

  Future<void> _showQuestNodeSubmissionModal(
    String title,
    QuestNode node, {
    String? standaloneChallengeId,
    String? challengeImageHeroTag,
  }) async {
    final parentContext = context;
    final textController = TextEditingController();
    CapturedImage? capturedImage;
    PlatformFile? capturedVideo;
    bool uploadingSubmission = false;
    String? selectedChallengeId = node.challenges.isNotEmpty
        ? node.challenges.first.id
        : null;
    final questLogProvider = context.read<QuestLogProvider>();
    final poiService = context.read<PoiService>();
    final mediaService = context.read<MediaService>();
    final userId = context.read<AuthProvider>().user?.id ?? 'anonymous';

    await showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) {
        return Padding(
          padding: EdgeInsets.only(
            left: 16,
            right: 16,
            bottom: MediaQuery.viewInsetsOf(context).bottom + 24,
            top: 16,
          ),
          child: StatefulBuilder(
            builder: (context, setModalState) {
              final canUseCamera =
                  kIsWeb ||
                  defaultTargetPlatform == TargetPlatform.iOS ||
                  defaultTargetPlatform == TargetPlatform.android;
              final submissionType = node.submissionType;
              final isTextSubmission =
                  submissionType == QuestNode.submissionTypeText;
              final isPhotoSubmission =
                  submissionType == QuestNode.submissionTypePhoto;
              final isVideoSubmission =
                  submissionType == QuestNode.submissionTypeVideo;
              final selectedChallenge = node.challenges.isEmpty
                  ? null
                  : (selectedChallengeId == null
                        ? node.challenges.first
                        : node.challenges.firstWhere(
                            (c) => c.id == selectedChallengeId,
                            orElse: () => node.challenges.first,
                          ));
              final selectedChallengeHeroTag = selectedChallenge == null
                  ? null
                  : (challengeImageHeroTag ??
                        _challengeImageHeroTag(selectedChallenge.id));
              final statValues = context.watch<CharacterStatsProvider>().stats;
              final statTags = (selectedChallenge?.statTags ?? const [])
                  .map((tag) => tag.trim().toLowerCase())
                  .where((tag) => tag.isNotEmpty)
                  .toList();
              final difficultyValue = selectedChallenge?.difficulty ?? 0;
              final statAverage = _averageStatValue(statValues, statTags);
              final difficultyColor = _difficultyColor(
                statAverage,
                difficultyValue,
              );
              return Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    title,
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (node.challenges.length > 1)
                    DropdownButtonFormField<String>(
                      value: selectedChallengeId,
                      items: node.challenges
                          .map(
                            (c) => DropdownMenuItem(
                              value: c.id,
                              child: Text(c.question),
                            ),
                          )
                          .toList(),
                      onChanged: (value) {
                        setModalState(() => selectedChallengeId = value);
                      },
                      decoration: const InputDecoration(
                        labelText: 'Challenge',
                        border: OutlineInputBorder(),
                      ),
                    )
                  else if (node.challenges.isNotEmpty)
                    Text(
                      node.challenges.first.question,
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  if (selectedChallenge != null &&
                      (selectedChallenge.imageUrl.isNotEmpty ||
                          selectedChallenge.thumbnailUrl.isNotEmpty)) ...[
                    const SizedBox(height: 10),
                    Hero(
                      tag: selectedChallengeHeroTag!,
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(12),
                        child: Image.network(
                          selectedChallenge.thumbnailUrl.isNotEmpty
                              ? selectedChallenge.thumbnailUrl
                              : selectedChallenge.imageUrl,
                          fit: BoxFit.cover,
                          height: 220,
                          width: double.infinity,
                          errorBuilder: (_, __, ___) => const SizedBox.shrink(),
                        ),
                      ),
                    ),
                  ],
                  if (selectedChallenge != null) ...[
                    const SizedBox(height: 6),
                    Text(
                      'Difficulty: ${selectedChallenge.difficulty}',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: difficultyColor,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                  if (statTags.isNotEmpty) ...[
                    const SizedBox(height: 12),
                    Text(
                      'Stat modifiers',
                      style: Theme.of(context).textTheme.labelLarge?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: statTags.map((tag) {
                        final label = _formatStatLabel(tag);
                        final value =
                            statValues[tag] ??
                            CharacterStatsProvider.baseStatValue;
                        return Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: Theme.of(
                              context,
                            ).colorScheme.surfaceVariant.withOpacity(0.6),
                            borderRadius: BorderRadius.circular(999),
                            border: Border.all(
                              color: Theme.of(
                                context,
                              ).colorScheme.outlineVariant,
                            ),
                          ),
                          child: Text(
                            '+$value $label',
                            style: Theme.of(context).textTheme.labelSmall,
                          ),
                        );
                      }).toList(),
                    ),
                  ],
                  const SizedBox(height: 12),
                  if (isTextSubmission) ...[
                    TextField(
                      controller: textController,
                      decoration: const InputDecoration(
                        labelText: 'Answer',
                        border: OutlineInputBorder(),
                      ),
                      maxLines: 3,
                    ),
                    const SizedBox(height: 12),
                  ],
                  if (isPhotoSubmission) ...[
                    if (canUseCamera)
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: uploadingSubmission
                                  ? null
                                  : () async {
                                      final result =
                                          await captureImageFromCamera();
                                      if (!mounted) return;
                                      if (result == null ||
                                          result.bytes.isEmpty) {
                                        ScaffoldMessenger.of(
                                          context,
                                        ).showSnackBar(
                                          const SnackBar(
                                            content: Text('No photo captured.'),
                                          ),
                                        );
                                        return;
                                      }
                                      setModalState(
                                        () => capturedImage = result,
                                      );
                                    },
                              icon: const Icon(Icons.photo_camera),
                              label: const Text('Take photo'),
                            ),
                          ),
                          if (capturedImage != null) ...[
                            const SizedBox(width: 12),
                            TextButton(
                              onPressed: () =>
                                  setModalState(() => capturedImage = null),
                              child: const Text('Clear'),
                            ),
                          ],
                        ],
                      ),
                    if (capturedImage != null) ...[
                      const SizedBox(height: 12),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.memory(
                          capturedImage!.bytes,
                          height: 160,
                          fit: BoxFit.cover,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Captured photo will be uploaded on submit.',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                    const SizedBox(height: 12),
                  ],
                  if (isVideoSubmission) ...[
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: uploadingSubmission
                                ? null
                                : () async {
                                    final result = await FilePicker.platform
                                        .pickFiles(
                                          type: FileType.video,
                                          withData: true,
                                        );
                                    if (!mounted) return;
                                    if (result == null ||
                                        result.files.isEmpty) {
                                      ScaffoldMessenger.of(
                                        context,
                                      ).showSnackBar(
                                        const SnackBar(
                                          content: Text('No video selected.'),
                                        ),
                                      );
                                      return;
                                    }
                                    setModalState(
                                      () => capturedVideo = result.files.first,
                                    );
                                  },
                            icon: const Icon(Icons.videocam),
                            label: Text(
                              capturedVideo == null
                                  ? 'Select video'
                                  : 'Replace video',
                            ),
                          ),
                        ),
                        if (capturedVideo != null) ...[
                          const SizedBox(width: 12),
                          TextButton(
                            onPressed: () =>
                                setModalState(() => capturedVideo = null),
                            child: const Text('Clear'),
                          ),
                        ],
                      ],
                    ),
                    if (capturedVideo != null) ...[
                      const SizedBox(height: 8),
                      Text(
                        'Selected video: ${capturedVideo!.name}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Video will be uploaded on submit.',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                    const SizedBox(height: 12),
                  ],
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: uploadingSubmission
                        ? null
                        : () async {
                            final trimmedText = textController.text.trim();
                            if (isTextSubmission && trimmedText.isEmpty) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Please enter an answer.'),
                                ),
                              );
                              return;
                            }
                            if (isPhotoSubmission && capturedImage == null) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Please capture a photo.'),
                                ),
                              );
                              return;
                            }
                            if (isVideoSubmission && capturedVideo == null) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Please select a video.'),
                                ),
                              );
                              return;
                            }
                            if (standaloneChallengeId != null) {
                              try {
                                final status = await poiService
                                    .getPartySubmissionStatus(
                                      contentType: 'challenge',
                                      contentId: standaloneChallengeId,
                                    );
                                if (status.locked) {
                                  if (!context.mounted) return;
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(
                                      content: Text(
                                        status.isCompleted
                                            ? 'A party member already resolved this challenge.'
                                            : 'A party member is already submitting this challenge.',
                                      ),
                                    ),
                                  );
                                  return;
                                }
                              } catch (_) {}
                            }
                            final startedAt = DateTime.now();
                            setModalState(() => uploadingSubmission = true);
                            Navigator.of(context).pop();
                            _setQuestSubmissionOverlay(
                              QuestSubmissionOverlayPhase.loading,
                            );
                            String? imageSubmissionUrl;
                            String? videoSubmissionUrl;
                            if (isPhotoSubmission && capturedImage != null) {
                              final ext =
                                  _extensionFromMime(
                                    capturedImage!.mimeType,
                                    capturedImage!.name,
                                  ) ??
                                  'jpg';
                              final key =
                                  'quest-submissions/$userId/${DateTime.now().millisecondsSinceEpoch}.$ext';
                              final url = await mediaService
                                  .getPresignedUploadUrl(
                                    ApiConstants.crewPointsOfInterestBucket,
                                    key,
                                  );
                              if (url == null) {
                                final elapsed = DateTime.now().difference(
                                  startedAt,
                                );
                                if (elapsed <
                                    const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                _setQuestSubmissionOverlay(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to prepare image upload.',
                                );
                                return;
                              }
                              final ok = await mediaService.uploadToPresigned(
                                url,
                                Uint8List.fromList(capturedImage!.bytes),
                                capturedImage!.mimeType ?? 'image/jpeg',
                              );
                              if (!ok) {
                                final elapsed = DateTime.now().difference(
                                  startedAt,
                                );
                                if (elapsed <
                                    const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                _setQuestSubmissionOverlay(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to upload photo.',
                                );
                                return;
                              }
                              imageSubmissionUrl = url.split('?').first;
                            }
                            if (isVideoSubmission && capturedVideo != null) {
                              final ext =
                                  _extensionFromMime(
                                    _mimeTypeFromFile(capturedVideo!),
                                    capturedVideo!.name,
                                  ) ??
                                  'mp4';
                              final key =
                                  'quest-submissions/$userId/${DateTime.now().millisecondsSinceEpoch}.$ext';
                              final url = await mediaService
                                  .getPresignedUploadUrl(
                                    ApiConstants.crewPointsOfInterestBucket,
                                    key,
                                  );
                              if (url == null) {
                                final elapsed = DateTime.now().difference(
                                  startedAt,
                                );
                                if (elapsed <
                                    const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                _setQuestSubmissionOverlay(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to prepare video upload.',
                                );
                                return;
                              }
                              final bytes = capturedVideo!.bytes;
                              if (bytes == null || bytes.isEmpty) {
                                final elapsed = DateTime.now().difference(
                                  startedAt,
                                );
                                if (elapsed <
                                    const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                _setQuestSubmissionOverlay(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to read video data.',
                                );
                                return;
                              }
                              final ok = await mediaService.uploadToPresigned(
                                url,
                                Uint8List.fromList(bytes),
                                _mimeTypeFromFile(capturedVideo!) ??
                                    'video/mp4',
                              );
                              if (!ok) {
                                final elapsed = DateTime.now().difference(
                                  startedAt,
                                );
                                if (elapsed <
                                    const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                _setQuestSubmissionOverlay(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to upload video.',
                                );
                                return;
                              }
                              videoSubmissionUrl = url.split('?').first;
                            }
                            late final Map<String, dynamic> resp;
                            try {
                              resp = standaloneChallengeId == null
                                  ? await questLogProvider
                                        .submitQuestNodeChallenge(
                                          node.id,
                                          questNodeChallengeId:
                                              selectedChallengeId,
                                          textSubmission: isTextSubmission
                                              ? trimmedText
                                              : null,
                                          imageSubmissionUrl:
                                              imageSubmissionUrl,
                                          videoSubmissionUrl:
                                              videoSubmissionUrl,
                                        )
                                  : await poiService.submitChallenge(
                                      standaloneChallengeId,
                                      textSubmission: isTextSubmission
                                          ? trimmedText
                                          : null,
                                      imageSubmissionUrl: imageSubmissionUrl,
                                      videoSubmissionUrl: videoSubmissionUrl,
                                    );
                            } catch (error) {
                              final elapsed = DateTime.now().difference(
                                startedAt,
                              );
                              if (elapsed < const Duration(milliseconds: 700)) {
                                await Future<void>.delayed(
                                  const Duration(milliseconds: 700),
                                );
                              }
                              var message = 'Submission failed.';
                              if (error is DioException &&
                                  error.response?.data is Map) {
                                final data = Map<String, dynamic>.from(
                                  error.response!.data as Map,
                                );
                                final apiMessage =
                                    data['error']?.toString().trim() ?? '';
                                if (apiMessage.isNotEmpty) {
                                  message = apiMessage;
                                }
                              }
                              _setQuestSubmissionOverlay(
                                QuestSubmissionOverlayPhase.failure,
                                message: message,
                              );
                              return;
                            }
                            final elapsed = DateTime.now().difference(
                              startedAt,
                            );
                            if (elapsed < const Duration(milliseconds: 700)) {
                              await Future<void>.delayed(
                                const Duration(milliseconds: 700),
                              );
                            }
                            final success = resp['successful'] == true;
                            final reason = resp['reason']?.toString() ?? '';
                            final score = (resp['score'] as num?)?.toInt();
                            final difficulty = (resp['difficulty'] as num?)
                                ?.toInt();
                            final combined = (resp['combinedScore'] as num?)
                                ?.toInt();
                            final statTags = (resp['statTags'] as List?)
                                ?.map((tag) => tag.toString())
                                .toList();
                            final statValues = (resp['statValues'] as Map?)
                                ?.map(
                                  (key, value) => MapEntry(
                                    key.toString(),
                                    (value as num?)?.toInt() ?? 0,
                                  ),
                                );
                            final baseMessage = success
                                ? (reason.isNotEmpty
                                      ? reason
                                      : 'Challenge completed!')
                                : (reason.isNotEmpty
                                      ? reason
                                      : 'Submission failed');
                            _setQuestSubmissionOverlay(
                              success
                                  ? QuestSubmissionOverlayPhase.success
                                  : QuestSubmissionOverlayPhase.failure,
                              message: baseMessage,
                              score: score,
                              difficulty: difficulty,
                              combinedScore: combined,
                              statTags: statTags,
                              statValues: statValues,
                            );
                            if (standaloneChallengeId != null &&
                                mounted &&
                                parentContext.mounted) {
                              parentContext
                                  .read<CompletedTaskProvider>()
                                  .showModal(
                                    'challengeOutcome',
                                    data: Map<String, dynamic>.from(resp),
                                  );
                            }
                          },
                    child: const Text('Submit'),
                  ),
                ],
              );
            },
          ),
        );
      },
    );
  }

  void _updateSelectedZoneFromLocation() {
    final location = context.read<LocationProvider>().location;
    if (location == null || _zones.isEmpty) return;

    final zoneProvider = context.read<ZoneProvider>();
    final zone = zoneProvider.findZoneAtCoordinate(
      location.latitude,
      location.longitude,
    );
    zoneProvider.setSelectedZone(zone);
  }

  void _setQuestSubmissionOverlay(
    QuestSubmissionOverlayPhase phase, {
    String? message,
    int? score,
    int? difficulty,
    int? combinedScore,
    List<String>? statTags,
    Map<String, int>? statValues,
  }) {
    final hasDetails =
        score != null ||
        difficulty != null ||
        combinedScore != null ||
        statTags != null ||
        statValues != null;
    _clearQuestSubmissionRevealTimers();
    setState(() {
      _questSubmissionPhase = phase;
      _questSubmissionMessage = message;
      if (phase == QuestSubmissionOverlayPhase.loading || !hasDetails) {
        _questSubmissionScore = null;
        _questSubmissionDifficulty = null;
        _questSubmissionCombinedScore = null;
        _questSubmissionStatTags = const [];
        _questSubmissionStatValues = const <String, int>{};
        _questSubmissionRevealStep = 0;
      } else {
        _questSubmissionScore = score;
        _questSubmissionDifficulty = difficulty;
        _questSubmissionCombinedScore = combinedScore;
        _questSubmissionStatTags = statTags ?? const [];
        _questSubmissionStatValues = statValues ?? const <String, int>{};
        _questSubmissionRevealStep = 0;
      }
    });
    if (phase != QuestSubmissionOverlayPhase.loading && hasDetails) {
      _startQuestSubmissionRevealSequence();
    } else if (phase != QuestSubmissionOverlayPhase.loading) {
      setState(() => _questSubmissionRevealStep = 5);
    }
    if (phase != QuestSubmissionOverlayPhase.loading && hasDetails) {
      unawaited(context.read<CharacterStatsProvider>().refresh());
    }
  }

  void _dismissQuestSubmissionOverlay() {
    if (!mounted) return;
    _clearQuestSubmissionRevealTimers();
    setState(() {
      _questSubmissionPhase = QuestSubmissionOverlayPhase.hidden;
      _questSubmissionMessage = null;
      _questSubmissionScore = null;
      _questSubmissionDifficulty = null;
      _questSubmissionCombinedScore = null;
      _questSubmissionStatTags = const [];
      _questSubmissionStatValues = const <String, int>{};
      _questSubmissionRevealStep = 0;
    });
  }

  @override
  Widget build(BuildContext context) {
    final routeUri = GoRouterState.of(context).uri;
    _consumeMonsterBattleRouteIntent(routeUri);
    final loc = context.watch<LocationProvider>().location;
    final discoveries = context.watch<DiscoveriesProvider>();
    final questLog = context.watch<QuestLogProvider>();
    final lat = loc?.latitude ?? 0.0;
    final lng = loc?.longitude ?? 0.0;
    final initialPosition = CameraPosition(target: LatLng(lat, lng), zoom: 15);
    const overlayButtonSize = 48.0;
    const overlayButtonSpacing = 12.0;
    const overlayButtonCount = 3;
    final overlayButtonStackHeight =
        overlayButtonSize * overlayButtonCount +
        overlayButtonSpacing * (overlayButtonCount - 1);
    final authUser = context.watch<AuthProvider>().user;
    context.watch<PartyProvider>().party;
    final hasPartyMapStrip = authUser != null;
    final hasTrackedQuestOverlay = questLog.quests.any(
      (quest) => questLog.trackedQuestIds.contains(quest.id),
    );

    Quest? polygonQuest;
    QuestNode? polygonNode;
    if (loc != null) {
      for (final quest in questLog.quests) {
        final node = quest.currentNode;
        if (!quest.isAccepted || node == null || node.polygon.isEmpty) {
          continue;
        }
        if (_isInsidePolygon(loc.latitude, loc.longitude, node.polygon)) {
          polygonQuest = quest;
          polygonNode = node;
          break;
        }
      }
    }

    if (_styleLoaded &&
        !_mapLoadFailed &&
        _mapController != null &&
        loc != null &&
        !_hasAnimatedToUserLocation) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        _animateToUserLocationIfReady();
      });
    }

    // Retry adding markers/zones when we have style + data but haven't added yet (e.g. style loaded after _loadAll)
    if (_styleLoaded &&
        !_mapLoadFailed &&
        _mapController != null &&
        _zones.isNotEmpty &&
        !_markersAdded) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        unawaited(_addPoiMarkers());
        unawaited(_addZoneBoundaries());
        unawaited(_addQuestPolygons());
      });
    }

    // Re-add POI markers when discoveries load after we had added with empty (auth ready late)
    if (!discoveries.loading &&
        discoveries.discoveries.isNotEmpty &&
        _addedMarkersWithEmptyDiscoveries &&
        _styleLoaded &&
        !_mapLoadFailed &&
        _mapController != null &&
        _pois.isNotEmpty &&
        _markersAdded) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        setState(() => _addedMarkersWithEmptyDiscoveries = false);
        unawaited(_refreshDiscoveredPoiMarkers());
        unawaited(_refreshCharacterDiscoveryMarkers());
      });
    }

    return Scaffold(
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      body: Stack(
        children: [
          Listener(
            behavior: HitTestBehavior.translucent,
            onPointerDown: (event) {
              if (kDebugMode) {
                debugPrint(
                  'SinglePlayer: map pointer down at ${event.position}',
                );
              }
            },
            child: MapLibreMap(
              key: ValueKey(_mapKey),
              initialCameraPosition: initialPosition,
              styleString: _stamenWatercolorStyle,
              minMaxZoomPreference: const MinMaxZoomPreference(null, 16),
              onMapCreated: (c) {
                debugPrint('SinglePlayer: map created');
                _mapController = c;
                _setupTapHandlers(c);
              },
              onMapClick: _handleMapClick,
              onStyleLoadedCallback: _onMapStyleLoaded,
              myLocationEnabled: true,
              compassEnabled: false,
            ),
          ),
          if (!_styleLoaded && !_mapLoadFailed)
            Positioned.fill(
              child: Container(
                color: Theme.of(context).scaffoldBackgroundColor,
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const CircularProgressIndicator(),
                      const SizedBox(height: 16),
                      Text(
                        'Loading map...',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    ],
                  ),
                ),
              ),
            ),
          if (_mapLoadFailed)
            Positioned.fill(
              child: Container(
                color: Theme.of(context).scaffoldBackgroundColor,
                padding: const EdgeInsets.all(24),
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.error_outline,
                        size: 48,
                        color: Theme.of(context).colorScheme.error,
                      ),
                      const SizedBox(height: 16),
                      Text(
                        'Map failed to load',
                        style: Theme.of(context).textTheme.titleLarge,
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Check your connection and try again.',
                        style: Theme.of(context).textTheme.bodyMedium,
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 24),
                      FilledButton.icon(
                        onPressed: _retryMap,
                        icon: const Icon(Icons.refresh),
                        label: const Text('Retry'),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          if (_styleLoaded && !_mapLoadFailed) ...[
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: PointerInterceptor(
                child: Listener(
                  behavior: HitTestBehavior.translucent,
                  onPointerDown: (event) {
                    if (kDebugMode) {
                      debugPrint(
                        'SinglePlayer: top controls pointer down at ${event.position}',
                      );
                    }
                  },
                  child: SafeArea(
                    bottom: false,
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
                      child: SizedBox(
                        width: double.infinity,
                        child: SizedBox(
                          height: overlayButtonStackHeight,
                          child: Align(
                            alignment: Alignment.topLeft,
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Consumer<ActivityFeedProvider>(
                                  builder: (context, feed, _) {
                                    final hasUnseen =
                                        feed.unseenActivities.isNotEmpty;
                                    return Stack(
                                      clipBehavior: Clip.none,
                                      children: [
                                        _OverlayButton(
                                          icon: Icons.campaign,
                                          onTap: () {
                                            if (kDebugMode) {
                                              debugPrint(
                                                'SinglePlayer: notifications tapped',
                                              );
                                            }
                                            _showActivityFeed(context);
                                          },
                                        ),
                                        if (hasUnseen)
                                          Positioned(
                                            top: -2,
                                            right: -2,
                                            child: Container(
                                              width: 10,
                                              height: 10,
                                              decoration: BoxDecoration(
                                                color: const Color(0xFFB87333),
                                                shape: BoxShape.circle,
                                                border: Border.all(
                                                  color: Theme.of(
                                                    context,
                                                  ).colorScheme.surface,
                                                  width: 1,
                                                ),
                                                boxShadow: const [
                                                  BoxShadow(
                                                    color: Colors.black26,
                                                    blurRadius: 4,
                                                    offset: Offset(0, 2),
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ),
                                      ],
                                    );
                                  },
                                ),
                                const SizedBox(height: 12),
                                Consumer2<QuestFilterProvider, TagsProvider>(
                                  builder: (context, filters, tags, _) {
                                    final hasActiveFilters =
                                        filters.enableTagFilter &&
                                        tags.selectedTagIds.isNotEmpty;
                                    return Stack(
                                      clipBehavior: Clip.none,
                                      children: [
                                        _OverlayButton(
                                          icon: Icons.tune,
                                          onTap: () {
                                            if (kDebugMode) {
                                              debugPrint(
                                                'SinglePlayer: filters tapped',
                                              );
                                            }
                                            _showTagFilter(context);
                                          },
                                        ),
                                        if (hasActiveFilters)
                                          Positioned(
                                            top: -2,
                                            right: -2,
                                            child: Container(
                                              width: 10,
                                              height: 10,
                                              decoration: BoxDecoration(
                                                color: const Color(0xFFF5C542),
                                                shape: BoxShape.circle,
                                                border: Border.all(
                                                  color: Colors.white,
                                                  width: 1,
                                                ),
                                                boxShadow: const [
                                                  BoxShadow(
                                                    color: Colors.black26,
                                                    blurRadius: 4,
                                                    offset: Offset(0, 2),
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ),
                                      ],
                                    );
                                  },
                                ),
                                const SizedBox(height: 12),
                                _OverlayButton(
                                  icon: Icons.my_location,
                                  onTap: _centerOnUserLocation,
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            if (hasPartyMapStrip || hasTrackedQuestOverlay)
              Positioned(
                top: 0,
                left: 16,
                right: 16,
                child: PointerInterceptor(
                  child: SafeArea(
                    bottom: false,
                    child: Align(
                      alignment: Alignment.topRight,
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          if (hasPartyMapStrip) const PartyMemberMapStrip(),
                          if (hasPartyMapStrip && hasTrackedQuestOverlay)
                            const SizedBox(height: 12),
                          if (hasTrackedQuestOverlay)
                            TrackedQuestsOverlay(
                              controller: _trackedQuestsController,
                              onFocusPoI: _focusQuestPoI,
                              onFocusNode: _focusQuestNode,
                              onOpenQuestDetails: _openQuestLogForQuest,
                            ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            if (polygonQuest != null && polygonNode != null)
              Positioned(
                left: 16,
                right: 16,
                bottom: 92,
                child: FilledButton(
                  onPressed: () => _showQuestNodeSubmissionModal(
                    polygonQuest!.name,
                    polygonNode!,
                  ),
                  child: Text('Quest: ${polygonQuest.name}'),
                ),
              ),
            Positioned(
              left: 0,
              right: 0,
              bottom: 0,
              child: PointerInterceptor(
                child: SafeArea(
                  top: false,
                  child: Padding(
                    padding: const EdgeInsets.only(bottom: 24),
                    child: Align(
                      alignment: Alignment.bottomCenter,
                      child: ZoneWidget(
                        controller: _zoneWidgetController,
                        expandUpwards: true,
                        expandedHeight: 260,
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ],
          if (_questSubmissionPhase != QuestSubmissionOverlayPhase.hidden)
            Positioned.fill(
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap:
                    _questSubmissionPhase ==
                            QuestSubmissionOverlayPhase.loading ||
                        _questSubmissionRevealStep < 5
                    ? null
                    : _dismissQuestSubmissionOverlay,
                child: AnimatedOpacity(
                  duration: const Duration(milliseconds: 200),
                  opacity: 1,
                  child: Container(
                    color: Colors.black.withOpacity(0.45),
                    child: Center(
                      child: LayoutBuilder(
                        builder: (context, constraints) {
                          final isLoading =
                              _questSubmissionPhase ==
                              QuestSubmissionOverlayPhase.loading;
                          final isFailure =
                              _questSubmissionPhase ==
                              QuestSubmissionOverlayPhase.failure;
                          final accentColor = isFailure
                              ? Theme.of(context).colorScheme.error
                              : const Color(0xFFF5C542);
                          final statsProvider = context
                              .watch<CharacterStatsProvider>();
                          final statValues =
                              _questSubmissionStatValues.isNotEmpty
                              ? _questSubmissionStatValues
                              : statsProvider.stats;
                          final statTags = _questSubmissionStatTags;
                          final hasDetails =
                              _questSubmissionScore != null ||
                              _questSubmissionDifficulty != null ||
                              _questSubmissionCombinedScore != null ||
                              statTags.isNotEmpty ||
                              _questSubmissionStatValues.isNotEmpty;
                          final scoreValue = _questSubmissionScore ?? 0;
                          final combinedValue =
                              _questSubmissionCombinedScore ??
                              (scoreValue +
                                  statTags.fold<int>(
                                    0,
                                    (sum, tag) =>
                                        sum +
                                        (statValues[tag] ??
                                            CharacterStatsProvider
                                                .baseStatValue),
                                  ));
                          final difficultyValue =
                              _questSubmissionDifficulty ?? 0;
                          final availableWidth = constraints.maxWidth * 0.9;
                          final maxWidth = availableWidth > 420
                              ? 420.0
                              : availableWidth;
                          final minWidth = maxWidth < 280 ? maxWidth : 280.0;
                          final borderRadius = BorderRadius.circular(20);

                          return ConstrainedBox(
                            constraints: BoxConstraints(
                              minWidth: minWidth,
                              maxWidth: maxWidth,
                            ),
                            child: PaperTexture(
                              borderRadius: borderRadius,
                              opacity: 0.1,
                              child: Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 20,
                                  vertical: 22,
                                ),
                                decoration: BoxDecoration(
                                  color: Theme.of(
                                    context,
                                  ).colorScheme.surface.withOpacity(0.98),
                                  borderRadius: borderRadius,
                                  border: Border.all(
                                    color: accentColor.withOpacity(0.9),
                                    width: 1.5,
                                  ),
                                  boxShadow: const [
                                    BoxShadow(
                                      color: Colors.black26,
                                      blurRadius: 18,
                                      offset: Offset(0, 8),
                                    ),
                                  ],
                                ),
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  crossAxisAlignment:
                                      CrossAxisAlignment.stretch,
                                  children: [
                                    Center(
                                      child: Text(
                                        "Dungeonmaster's score",
                                        textAlign: TextAlign.center,
                                        style: Theme.of(context)
                                            .textTheme
                                            .titleMedium
                                            ?.copyWith(
                                              fontWeight: FontWeight.w700,
                                            ),
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    if (isLoading ||
                                        _questSubmissionRevealStep == 0) ...[
                                      Center(
                                        child: Text(
                                          'Calculating...',
                                          style: Theme.of(
                                            context,
                                          ).textTheme.bodySmall,
                                        ),
                                      ),
                                      const SizedBox(height: 10),
                                      LinearProgressIndicator(
                                        minHeight: 6,
                                        color: accentColor,
                                        backgroundColor: accentColor
                                            .withOpacity(0.15),
                                      ),
                                    ],
                                    if (hasDetails) ...[
                                      _buildRevealSection(
                                        1,
                                        Center(
                                          child: Column(
                                            children: [
                                              Text(
                                                '$scoreValue',
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .displaySmall
                                                    ?.copyWith(
                                                      fontWeight:
                                                          FontWeight.w700,
                                                      color: accentColor,
                                                    ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                      _buildRevealSection(
                                        2,
                                        SizedBox(
                                          width: double.infinity,
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              const SizedBox(height: 16),
                                              Text(
                                                'Stat modifiers',
                                                style: Theme.of(
                                                  context,
                                                ).textTheme.labelLarge,
                                              ),
                                              const SizedBox(height: 6),
                                              if (statTags.isEmpty)
                                                Text(
                                                  'None',
                                                  style: Theme.of(
                                                    context,
                                                  ).textTheme.bodySmall,
                                                )
                                              else
                                                Wrap(
                                                  spacing: 8,
                                                  runSpacing: 8,
                                                  children: statTags.map((tag) {
                                                    final label =
                                                        _formatStatLabel(tag);
                                                    final value =
                                                        statValues[tag] ??
                                                        CharacterStatsProvider
                                                            .baseStatValue;
                                                    return Container(
                                                      padding:
                                                          const EdgeInsets.symmetric(
                                                            horizontal: 10,
                                                            vertical: 6,
                                                          ),
                                                      decoration: BoxDecoration(
                                                        color: Theme.of(context)
                                                            .colorScheme
                                                            .surfaceVariant
                                                            .withOpacity(0.6),
                                                        borderRadius:
                                                            BorderRadius.circular(
                                                              999,
                                                            ),
                                                        border: Border.all(
                                                          color: accentColor
                                                              .withOpacity(0.2),
                                                        ),
                                                      ),
                                                      child: Text(
                                                        '+$value $label',
                                                        style: Theme.of(
                                                          context,
                                                        ).textTheme.labelSmall,
                                                      ),
                                                    );
                                                  }).toList(),
                                                ),
                                            ],
                                          ),
                                        ),
                                      ),
                                      _buildRevealSection(
                                        3,
                                        SizedBox(
                                          width: double.infinity,
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              const SizedBox(height: 16),
                                              Text(
                                                'Difficulty',
                                                style: Theme.of(
                                                  context,
                                                ).textTheme.labelLarge,
                                              ),
                                              const SizedBox(height: 4),
                                              Text(
                                                '$difficultyValue',
                                                style: Theme.of(
                                                  context,
                                                ).textTheme.bodySmall,
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                      _buildRevealSection(
                                        4,
                                        SizedBox(
                                          width: double.infinity,
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              const SizedBox(height: 16),
                                              Text(
                                                'Scoring',
                                                style: Theme.of(
                                                  context,
                                                ).textTheme.labelLarge,
                                              ),
                                              const SizedBox(height: 6),
                                              if (statTags.isNotEmpty)
                                                Text(
                                                  'Modifiers: ${statTags.map((tag) {
                                                    final label = _formatStatLabel(tag);
                                                    final value = statValues[tag] ?? CharacterStatsProvider.baseStatValue;
                                                    return '+$value $label';
                                                  }).join(' · ')}',
                                                  style: Theme.of(
                                                    context,
                                                  ).textTheme.bodySmall,
                                                ),
                                              Text(
                                                () {
                                                  final parts = <String>[
                                                    'Score $scoreValue',
                                                  ];
                                                  for (final tag in statTags) {
                                                    final label =
                                                        _formatStatLabel(tag);
                                                    final value =
                                                        statValues[tag] ??
                                                        CharacterStatsProvider
                                                            .baseStatValue;
                                                    parts.add('+$value $label');
                                                  }
                                                  return '${parts.join(' ')} = $combinedValue';
                                                }(),
                                                style: Theme.of(
                                                  context,
                                                ).textTheme.bodySmall,
                                              ),
                                              Text(
                                                'Difficulty = $difficultyValue',
                                                style: Theme.of(
                                                  context,
                                                ).textTheme.bodySmall,
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                      _buildRevealSection(
                                        5,
                                        Column(
                                          children: [
                                            const SizedBox(height: 18),
                                            Row(
                                              mainAxisAlignment:
                                                  MainAxisAlignment.center,
                                              children: [
                                                Icon(
                                                  _questSubmissionPhase ==
                                                          QuestSubmissionOverlayPhase
                                                              .success
                                                      ? Icons.emoji_events
                                                      : Icons
                                                            .sentiment_very_dissatisfied,
                                                  size: 22,
                                                  color: accentColor,
                                                ),
                                                const SizedBox(width: 8),
                                                Text(
                                                  _questSubmissionPhase ==
                                                          QuestSubmissionOverlayPhase
                                                              .success
                                                      ? 'Victory!'
                                                      : 'Defeat',
                                                  style: Theme.of(context)
                                                      .textTheme
                                                      .titleMedium
                                                      ?.copyWith(
                                                        fontWeight:
                                                            FontWeight.w700,
                                                        color: accentColor,
                                                      ),
                                                ),
                                              ],
                                            ),
                                            const SizedBox(height: 8),
                                            Text(
                                              _questSubmissionMessage ??
                                                  (_questSubmissionPhase ==
                                                          QuestSubmissionOverlayPhase
                                                              .success
                                                      ? 'Challenge completed!'
                                                      : 'Submission failed'),
                                              textAlign: TextAlign.center,
                                              style: Theme.of(
                                                context,
                                              ).textTheme.bodySmall,
                                            ),
                                            const SizedBox(height: 12),
                                            Text(
                                              'Tap anywhere to dismiss.',
                                              style: Theme.of(
                                                context,
                                              ).textTheme.bodySmall,
                                            ),
                                          ],
                                        ),
                                      ),
                                    ] else if (!isLoading) ...[
                                      const SizedBox(height: 16),
                                      Row(
                                        mainAxisAlignment:
                                            MainAxisAlignment.center,
                                        children: [
                                          Icon(
                                            _questSubmissionPhase ==
                                                    QuestSubmissionOverlayPhase
                                                        .success
                                                ? Icons.emoji_events
                                                : Icons.error,
                                            size: 22,
                                            color: accentColor,
                                          ),
                                          const SizedBox(width: 8),
                                          Flexible(
                                            child: Text(
                                              _questSubmissionMessage ??
                                                  (_questSubmissionPhase ==
                                                          QuestSubmissionOverlayPhase
                                                              .success
                                                      ? 'Challenge completed!'
                                                      : 'Submission failed'),
                                              textAlign: TextAlign.center,
                                              style: Theme.of(context)
                                                  .textTheme
                                                  .titleMedium
                                                  ?.copyWith(
                                                    fontWeight: FontWeight.w700,
                                                  ),
                                            ),
                                          ),
                                        ],
                                      ),
                                      const SizedBox(height: 12),
                                      Text(
                                        'Tap anywhere to dismiss.',
                                        textAlign: TextAlign.center,
                                        style: Theme.of(
                                          context,
                                        ).textTheme.bodySmall,
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                  ),
                ),
              ),
            ),
          if (_tutorialWelcomeOverlayVisible)
            Positioned.fill(child: _buildTutorialWelcomeOverlay(context)),
          const CelebrationModalManager(),
          const NewItemModal(),
          const UsedItemModal(),
          // Shop and dialogue are opened from the character panel.
        ],
      ),
    );
  }

  Widget _buildTutorialWelcomeOverlay(BuildContext context) {
    final theme = Theme.of(context);
    return IgnorePointer(
      ignoring: false,
      child: Opacity(
        opacity: _tutorialWelcomeOverlayOpacity.clamp(0.0, 1.0),
        child: Container(
          color: const Color(0xFF19140D).withValues(alpha: 0.58),
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 360),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                child: PaperTexture(
                  borderRadius: BorderRadius.circular(28),
                  opacity: 0.12,
                  child: Container(
                    padding: const EdgeInsets.fromLTRB(28, 30, 28, 26),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF8EFD8).withValues(alpha: 0.96),
                      borderRadius: BorderRadius.circular(28),
                      border: Border.all(
                        color: const Color(0xFFD2B26C).withValues(alpha: 0.9),
                        width: 1.5,
                      ),
                      boxShadow: const [
                        BoxShadow(
                          color: Colors.black26,
                          blurRadius: 24,
                          offset: Offset(0, 12),
                        ),
                      ],
                    ),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 62,
                          height: 62,
                          decoration: BoxDecoration(
                            color: const Color(
                              0xFFF5C542,
                            ).withValues(alpha: 0.2),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: const Color(0xFFD4A72C),
                              width: 1.4,
                            ),
                          ),
                          child: const Icon(
                            Icons.explore_rounded,
                            color: Color(0xFF8C5A14),
                            size: 30,
                          ),
                        ),
                        const SizedBox(height: 18),
                        Text(
                          'Welcome to Unclaimed Streets',
                          textAlign: TextAlign.center,
                          style: GoogleFonts.cinzelDecorative(
                            textStyle: theme.textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF3D2B13),
                              height: 1.15,
                            ),
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          'The streets are open to you now. Make your mark.',
                          textAlign: TextAlign.center,
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: const Color(0xFF5F4A28),
                            height: 1.4,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _showCharacterPanel(Character ch) async {
    var openTrackedQuests = false;
    final hasDiscovered = _hasDiscoveredCharacter(ch);
    await showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => CharacterPanel(
        character: ch,
        hasDiscovered: hasDiscovered,
        onClose: () => Navigator.of(context).pop(),
        onUnlocked: () async {
          final poiId = _characterDiscoveryPoiId(ch);
          if (poiId.isNotEmpty) {
            await context.read<DiscoveriesProvider>().refresh();
          } else {
            await _markCharacterDiscovered(ch.id);
          }
          if (!mounted) return;
          await _updateCharacterSymbolForState(ch);
          if (!mounted) return;
          if (poiId.isNotEmpty) {
            final questLog = context.read<QuestLogProvider>();
            final isQuestCurrent = _currentQuestPoiIdsForFilter(
              questLog,
            ).contains(poiId);
            unawaited(
              _updatePoiSymbolForQuestState(
                poiId,
                isQuestCurrent: isQuestCurrent,
              ),
            );
          }
        },
        onQuestAccepted: () => openTrackedQuests = true,
        onStartDialogue: (dialogContext, character, action) {
          debugPrint(
            'SinglePlayer: onStartDialogue character=${character.id} action=${action.id}',
          );
          _showDialogueModal(dialogContext, character, action);
        },
        onStartShop: (dialogContext, character, action) {
          _showShopModal(dialogContext, character, action);
        },
      ),
    );
    if (!mounted || !openTrackedQuests) return;
    _trackedQuestsController.open();
  }

  Future<void> _showShopModal(
    BuildContext dialogContext,
    Character character,
    CharacterAction action,
  ) async {
    if (!dialogContext.mounted) {
      debugPrint(
        'SinglePlayer: showShopModal skipped (dialogContext unmounted)',
      );
      return;
    }
    debugPrint(
      'SinglePlayer: showShopModal open character=${character.id} action=${action.id}',
    );
    await showModalBottomSheet<void>(
      context: dialogContext,
      isScrollControlled: true,
      useRootNavigator: true,
      useSafeArea: false,
      backgroundColor: Theme.of(dialogContext).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      isDismissible: true,
      builder: (context) {
        debugPrint('SinglePlayer: showShopModal builder');
        return SafeArea(
          top: false,
          child: ShopModal(
            character: character,
            action: action,
            onClose: () => Navigator.of(context).pop(),
          ),
        );
      },
    );
    if (dialogContext.mounted && Navigator.of(dialogContext).canPop()) {
      Navigator.of(dialogContext).pop();
    }
  }

  Future<void> _showDialogueModal(
    BuildContext dialogContext,
    Character character,
    CharacterAction action,
  ) async {
    if (!dialogContext.mounted) {
      debugPrint(
        'SinglePlayer: showDialogueModal skipped (dialogContext unmounted)',
      );
      return;
    }
    debugPrint(
      'SinglePlayer: showDialogueModal open character=${character.id} action=${action.id}',
    );
    await showDialog<void>(
      context: dialogContext,
      useRootNavigator: true,
      barrierDismissible: true,
      builder: (context) {
        debugPrint('SinglePlayer: showDialogueModal builder');
        return RpgDialogueModal(
          character: character,
          action: action,
          onClose: () => Navigator.of(context).pop(),
        );
      },
    );
    if (dialogContext.mounted && Navigator.of(dialogContext).canPop()) {
      Navigator.of(dialogContext).pop();
    }
  }

  void _showHealingFountainPanel(HealingFountain fountain) {
    final parentContext = context;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => HealingFountainPanel(
        fountain: fountain,
        onClose: () => Navigator.of(context).pop(),
        onUnlocked: () async {
          if (!mounted) return;
          setState(() {
            _healingFountains = _healingFountains
                .map(
                  (item) => item.id == fountain.id
                      ? item.copyWith(discovered: true)
                      : item,
                )
                .toList(growable: false);
          });
          await _refreshHealingFountainSymbols();
        },
        onUsed: (result) {
          if (!mounted) return;
          final lastUsedAt = DateTime.tryParse(
            result['lastUsedAt']?.toString() ?? '',
          )?.toLocal();
          final nextAvailableAt = DateTime.tryParse(
            result['nextAvailableAt']?.toString() ?? '',
          )?.toLocal();

          setState(() {
            _healingFountains = _healingFountains
                .map(
                  (item) => item.id == fountain.id
                      ? item.copyWith(
                          availableNow: false,
                          lastUsedAt: lastUsedAt,
                          nextAvailableAt: nextAvailableAt,
                          cooldownSecondsRemaining:
                              (result['cooldownSecondsRemaining'] as num?)
                                  ?.toInt() ??
                              0,
                        )
                      : item,
                )
                .toList(growable: false);
          });

          unawaited(_refreshHealingFountainSymbols());
          unawaited(
            context.read<CharacterStatsProvider>().refresh(silent: true),
          );
          unawaited(_loadTreasureChestsForSelectedZone());

          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            parentContext.read<CompletedTaskProvider>().showModal(
              'healingFountainUsed',
              data: {
                'healthRestored':
                    (result['healthRestored'] as num?)?.toInt() ?? 0,
                'manaRestored': (result['manaRestored'] as num?)?.toInt() ?? 0,
                'nextAvailableAt': result['nextAvailableAt']?.toString() ?? '',
              },
            );
          });
        },
      ),
    );
  }

  void _showTreasureChestPanel(TreasureChest tc) {
    final parentContext = context;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => TreasureChestPanel(
        treasureChest: tc,
        onClose: () {
          Navigator.of(context).pop();
          _loadTreasureChestsForSelectedZone();
        },
        onOpened: (rewardData) {
          if (!mounted) return;
          setState(() {
            _openedTreasureChestIds.add(tc.id);
            _treasureChests.removeWhere((chest) => chest.id == tc.id);
          });
          unawaited(_refreshTreasureChestSymbols());
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            parentContext.read<CompletedTaskProvider>().showModal(
              'treasureChestOpened',
              data: rewardData,
            );
          });
        },
      ),
    );
  }

  MapEntry<Quest, QuestNode>? _activeQuestNodeForChallenge(String challengeId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node == null) continue;
      if (node.challengeId == challengeId) {
        return MapEntry(quest, node);
      }
    }
    return null;
  }

  void _showChallengePanel(Challenge challenge) {
    final activeQuestEntry = _activeQuestNodeForChallenge(challenge.id);
    final questName = activeQuestEntry?.key.name;
    final anchor = _challengeProximityAnchor(challenge);
    final location = context.read<LocationProvider>().location;
    final distance = location == null
        ? null
        : _distanceMeters(
            location.latitude,
            location.longitude,
            anchor.latitude,
            anchor.longitude,
          );
    final withinRange =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    final mysteryState = !withinRange;
    final canSubmit = !mysteryState;
    var partySubmissionStatusLoading = !mysteryState;
    var partySubmissionLocked = false;
    String? partySubmissionStatus;
    var statusPollingStarted = false;
    var sheetClosed = false;
    Timer? statusPollTimer;

    Future<void> refreshPartySubmissionStatus(
      StateSetter setSheetState, {
      bool silent = false,
    }) async {
      if (mysteryState || sheetClosed) return;
      if (!silent) {
        setSheetState(() => partySubmissionStatusLoading = true);
      }
      try {
        final status = await context
            .read<PoiService>()
            .getPartySubmissionStatus(
              contentType: 'challenge',
              contentId: challenge.id,
            );
        if (sheetClosed) return;
        setSheetState(() {
          partySubmissionLocked = status.locked;
          partySubmissionStatus = status.status;
          partySubmissionStatusLoading = false;
        });
      } catch (_) {
        if (sheetClosed) return;
        setSheetState(() {
          partySubmissionStatusLoading = false;
        });
      }
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (sheetContext) {
        final theme = Theme.of(sheetContext);
        return StatefulBuilder(
          builder: (modalContext, setModalState) {
            if (!statusPollingStarted && !mysteryState) {
              statusPollingStarted = true;
              unawaited(
                refreshPartySubmissionStatus(setModalState, silent: false),
              );
              statusPollTimer = Timer.periodic(const Duration(seconds: 3), (_) {
                unawaited(
                  refreshPartySubmissionStatus(setModalState, silent: true),
                );
              });
            }

            final lockedByParty = partySubmissionLocked;
            final submitEnabled =
                canSubmit && !partySubmissionStatusLoading && !lockedByParty;

            return Padding(
              padding: const EdgeInsets.fromLTRB(16, 20, 16, 24),
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          mysteryState ? 'Mysterious Challenge' : 'Challenge',
                          style: theme.textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        IconButton(
                          onPressed: () => Navigator.of(sheetContext).pop(),
                          icon: const Icon(Icons.close),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Hero(
                      tag: _challengeImageHeroTag(challenge.id),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(14),
                        child: AspectRatio(
                          aspectRatio: 1,
                          child: Image.network(
                            mysteryState
                                ? _challengeMysteryImageUrl
                                : (challenge.thumbnailUrl.isNotEmpty
                                      ? challenge.thumbnailUrl
                                      : challenge.imageUrl),
                            fit: BoxFit.cover,
                            errorBuilder: (_, _, _) => mysteryState
                                ? Image.network(
                                    _legacyMysteryImageUrl,
                                    fit: BoxFit.cover,
                                    errorBuilder: (_, _, _) => Container(
                                      color: theme.colorScheme.surfaceVariant,
                                      child: const Icon(
                                        Icons.auto_awesome_outlined,
                                      ),
                                    ),
                                  )
                                : Container(
                                    color: theme.colorScheme.surfaceVariant,
                                    child: const Icon(
                                      Icons.auto_awesome_outlined,
                                    ),
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
                          _MiniInfoChip(
                            icon: Icons.place_outlined,
                            label: '${distance.round()} m away',
                          ),
                        _MiniInfoChip(
                          icon: Icons.shield_outlined,
                          label:
                              'Need ${kProximityUnlockRadiusMeters.round()} m',
                        ),
                        if (!mysteryState)
                          _MiniInfoChip(
                            icon: Icons.edit_note_outlined,
                            label: challenge.submissionType.toUpperCase(),
                          ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    if (mysteryState)
                      Text(
                        'This challenge remains a mystery until you are close enough to investigate.',
                        style: theme.textTheme.bodyMedium,
                      )
                    else ...[
                      Text(
                        challenge.question,
                        style: theme.textTheme.bodyLarge,
                      ),
                      if (challenge.description.trim().isNotEmpty) ...[
                        const SizedBox(height: 10),
                        Text(
                          challenge.description.trim(),
                          style: theme.textTheme.bodyMedium,
                        ),
                      ],
                      const SizedBox(height: 10),
                      Text(
                        'Difficulty: ${challenge.difficulty}',
                        style: theme.textTheme.bodySmall,
                      ),
                      if (lockedByParty) ...[
                        const SizedBox(height: 10),
                        Text(
                          (partySubmissionStatus ?? '').toLowerCase() ==
                                  'completed'
                              ? 'A party member already resolved this challenge.'
                              : 'A party member is submitting this challenge now.',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ],
                    if (!mysteryState) ...[
                      const SizedBox(height: 16),
                      FilledButton(
                        onPressed: submitEnabled
                            ? () async {
                                if (activeQuestEntry != null) {
                                  await _showStandaloneQuestChallengeSubmissionModal(
                                    activeQuestEntry.key,
                                    activeQuestEntry.value,
                                    challenge,
                                    challengeImageHeroTag:
                                        _challengeImageHeroTag(challenge.id),
                                  );
                                } else {
                                  await _showStandaloneChallengeSubmissionModal(
                                    challenge,
                                    challengeImageHeroTag:
                                        _challengeImageHeroTag(challenge.id),
                                  );
                                }
                                if (!sheetContext.mounted) return;
                                Navigator.of(sheetContext).pop();
                              }
                            : null,
                        child: Text(
                          activeQuestEntry == null
                              ? 'Submit Challenge'
                              : 'Submit for quest: $questName',
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            );
          },
        );
      },
    ).whenComplete(() {
      sheetClosed = true;
      statusPollTimer?.cancel();
    });
  }

  Future<void> _showStandaloneQuestChallengeSubmissionModal(
    Quest quest,
    QuestNode node,
    Challenge challenge, {
    String? challengeImageHeroTag,
  }) {
    final submissionType = challenge.submissionType.trim().isNotEmpty
        ? challenge.submissionType
        : node.submissionType;
    final syntheticNode = QuestNode(
      id: node.id,
      orderIndex: node.orderIndex,
      submissionType: submissionType,
      pointOfInterest: node.pointOfInterest,
      scenarioId: node.scenarioId,
      monsterId: node.monsterId,
      monsterEncounterId: node.monsterEncounterId,
      challengeId: node.challengeId,
      polygon: node.polygon,
      challenges: [
        QuestNodeChallenge(
          id: challenge.id,
          tier: 0,
          question: challenge.question,
          imageUrl: challenge.imageUrl,
          thumbnailUrl: challenge.thumbnailUrl,
          reward: challenge.reward,
          inventoryItemId: challenge.inventoryItemId,
          difficulty: challenge.difficulty,
          statTags: challenge.statTags,
          proficiency: challenge.proficiency,
        ),
      ],
    );
    return _showQuestNodeSubmissionModal(
      quest.name,
      syntheticNode,
      challengeImageHeroTag: challengeImageHeroTag,
    );
  }

  Future<void> _showStandaloneChallengeSubmissionModal(
    Challenge challenge, {
    String? challengeImageHeroTag,
  }) {
    final submissionType = challenge.submissionType.trim().isNotEmpty
        ? challenge.submissionType
        : QuestNode.submissionTypePhoto;
    final syntheticNode = QuestNode(
      id: challenge.id,
      orderIndex: 0,
      submissionType: submissionType,
      challengeId: challenge.id,
      challenges: [
        QuestNodeChallenge(
          id: challenge.id,
          tier: 0,
          question: challenge.question,
          imageUrl: challenge.imageUrl,
          thumbnailUrl: challenge.thumbnailUrl,
          reward: challenge.reward,
          inventoryItemId: challenge.inventoryItemId,
          difficulty: challenge.difficulty,
          statTags: challenge.statTags,
          proficiency: challenge.proficiency,
        ),
      ],
    );
    return _showQuestNodeSubmissionModal(
      'Challenge',
      syntheticNode,
      standaloneChallengeId: challenge.id,
      challengeImageHeroTag: challengeImageHeroTag,
    );
  }

  void _showMonsterPanel(MonsterEncounter monster) {
    final parentContext = context;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (sheetContext) => MonsterPanel(
        encounter: monster,
        onClose: () => Navigator.of(sheetContext).pop(),
        onFight: () {
          Navigator.of(sheetContext).pop();
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            unawaited(_startMonsterBattle(monster, parentContext));
          });
        },
      ),
    );
  }

  String _primaryMonsterIdForEncounter(MonsterEncounter encounter) {
    for (final monster in encounter.monsters) {
      final id = monster.id.trim();
      if (id.isNotEmpty) return id;
    }
    for (final member in encounter.members) {
      final id = member.monster.id.trim();
      if (id.isNotEmpty) return id;
    }
    return '';
  }

  void _consumeMonsterBattleRouteIntent(Uri routeUri) {
    final joinMonsterId =
        routeUri.queryParameters['joinMonsterId']?.trim() ?? '';
    final battleId = routeUri.queryParameters['battleId']?.trim() ?? '';
    final isPartyBattle = routeUri.queryParameters['partyBattle'] == '1';
    if (!isPartyBattle) return;
    if (joinMonsterId.isEmpty && battleId.isEmpty) return;
    if (_handlingMonsterBattleIntent) return;
    final requestKey = routeUri.toString();
    if (_lastHandledMonsterBattleIntent == requestKey) return;
    _lastHandledMonsterBattleIntent = requestKey;
    _handlingMonsterBattleIntent = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        _handlingMonsterBattleIntent = false;
        return;
      }
      unawaited(_launchMonsterBattleFromRouteIntent(joinMonsterId, battleId));
    });
  }

  Future<void> _launchMonsterBattleFromRouteIntent(
    String joinMonsterId,
    String battleId,
  ) async {
    try {
      _clearMonsterBattleRouteIntentFromUri();
      final poiService = context.read<PoiService>();
      var resolvedMonsterId = joinMonsterId.trim();
      final trimmedBattleId = battleId.trim();
      if (resolvedMonsterId.isEmpty && trimmedBattleId.isNotEmpty) {
        final detail = await poiService.getMonsterBattleStatusById(
          trimmedBattleId,
        );
        resolvedMonsterId = _stringFromBattleDetail(detail, 'monsterId').trim();
      }
      if (resolvedMonsterId.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Could not join battle: battle target not found.'),
            ),
          );
        }
        return;
      }
      final encounter = await _resolveMonsterEncounterForInvite(
        resolvedMonsterId,
      );
      if (encounter == null) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Could not join battle: encounter not found.'),
            ),
          );
        }
        return;
      }
      if (!mounted) return;
      await _startMonsterBattle(
        encounter,
        context,
        isPartyBattle: true,
        skipStartRequest: true,
        battleId: trimmedBattleId.isEmpty ? null : trimmedBattleId,
      );
    } finally {
      _handlingMonsterBattleIntent = false;
    }
  }

  String _stringFromBattleDetail(Map<String, dynamic> detail, String key) {
    final battleRaw = detail['battle'];
    if (battleRaw is Map<String, dynamic>) {
      return battleRaw[key]?.toString() ?? '';
    }
    if (battleRaw is Map) {
      return Map<String, dynamic>.from(battleRaw)[key]?.toString() ?? '';
    }
    return '';
  }

  void _clearMonsterBattleRouteIntentFromUri() {
    final uri = GoRouterState.of(context).uri;
    final hasIntentParams =
        uri.queryParameters.containsKey('joinMonsterId') ||
        uri.queryParameters.containsKey('partyBattle') ||
        uri.queryParameters.containsKey('inviteId') ||
        uri.queryParameters.containsKey('battleId');
    if (!hasIntentParams) return;
    final query = Map<String, String>.from(uri.queryParameters);
    query.remove('joinMonsterId');
    query.remove('partyBattle');
    query.remove('inviteId');
    query.remove('battleId');
    final cleaned = Uri(
      path: uri.path,
      queryParameters: query.isEmpty ? null : query,
    );
    if (cleaned.toString() == uri.toString()) return;
    context.replace(cleaned.toString());
  }

  Future<MonsterEncounter?> _resolveMonsterEncounterForInvite(
    String battleMonsterId,
  ) async {
    final localEncounter =
        _monsterEncounterByMemberMonsterId(battleMonsterId) ??
        _monsterById(battleMonsterId);
    if (localEncounter != null) return localEncounter;

    final poiService = context.read<PoiService>();
    final encounter = await poiService.getMonsterEncounterById(battleMonsterId);
    if (encounter != null) return encounter;

    final monster = await poiService.getMonsterById(battleMonsterId);
    if (monster == null) return null;
    return MonsterEncounter(
      id: monster.id,
      name: '${monster.name} Encounter',
      description: monster.description,
      imageUrl: monster.imageUrl,
      thumbnailUrl: monster.thumbnailUrl,
      zoneId: monster.zoneId,
      latitude: monster.latitude,
      longitude: monster.longitude,
      monsterCount: 1,
      members: [MonsterEncounterMember(slot: 1, monster: monster)],
      monsters: [monster],
    );
  }

  bool _battleWaitingOnInvites(Map<String, dynamic> battleDetail) {
    final pendingRaw = battleDetail['pendingResponses'];
    var pendingResponses = 0;
    if (pendingRaw is num) {
      pendingResponses = pendingRaw.toInt();
    } else {
      pendingResponses = int.tryParse(pendingRaw?.toString() ?? '') ?? 0;
    }

    final battleRaw = battleDetail['battle'];
    final battle = battleRaw is Map<String, dynamic>
        ? battleRaw
        : (battleRaw is Map ? Map<String, dynamic>.from(battleRaw) : null);
    final state = (battle?['state']?.toString() ?? '').trim();
    return pendingResponses > 0 || state == 'pending_party_responses';
  }

  bool _isPartyBattleDetail(Map<String, dynamic> battleDetail) {
    final pendingRaw = battleDetail['pendingResponses'];
    final pendingResponses = pendingRaw is num
        ? pendingRaw.toInt()
        : int.tryParse(pendingRaw?.toString() ?? '') ?? 0;
    if (pendingResponses > 0) return true;

    final invitesRaw = battleDetail['invites'];
    if (invitesRaw is List && invitesRaw.isNotEmpty) return true;

    final participantsRaw = battleDetail['participants'];
    if (participantsRaw is List && participantsRaw.length > 1) return true;

    return false;
  }

  Future<bool> _waitForPartyBattleReady(
    BuildContext parentContext,
    String battleMonsterId,
    Map<String, dynamic> initialBattleDetail,
    String? battleId,
  ) async {
    if (!_battleWaitingOnInvites(initialBattleDetail)) {
      return true;
    }
    final poiService = context.read<PoiService>();
    final deadline = DateTime.now().add(const Duration(seconds: 75));
    var latestDetail = initialBattleDetail;
    var waitingModalVisible = false;
    unawaited(
      showDialog<void>(
        context: parentContext,
        useRootNavigator: true,
        barrierDismissible: false,
        builder: (dialogContext) {
          waitingModalVisible = true;
          return PopScope(
            canPop: false,
            child: const AlertDialog(
              title: Text('Waiting For Party'),
              content: SizedBox(
                width: 300,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    CircularProgressIndicator(),
                    SizedBox(height: 14),
                    Text(
                      'Combat will start after all party invites are accepted, declined, or expire.',
                      textAlign: TextAlign.center,
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );

    try {
      while (mounted && parentContext.mounted) {
        if (!_battleWaitingOnInvites(latestDetail)) {
          return true;
        }
        if (DateTime.now().isAfter(deadline)) {
          return false;
        }
        await Future<void>.delayed(const Duration(seconds: 2));
        latestDetail = (battleId != null && battleId.trim().isNotEmpty)
            ? await poiService.getMonsterBattleStatusById(battleId.trim())
            : await poiService.getMonsterBattleStatus(battleMonsterId);
      }
      return false;
    } on DioException catch (error) {
      if (error.response?.statusCode == 404) {
        return false;
      }
      rethrow;
    } finally {
      if (waitingModalVisible && mounted && parentContext.mounted) {
        final navigator = Navigator.of(parentContext, rootNavigator: true);
        if (navigator.canPop()) {
          navigator.pop();
        }
      }
    }
  }

  Future<void> _startMonsterBattle(
    MonsterEncounter monster,
    BuildContext parentContext, {
    bool isPartyBattle = false,
    bool skipStartRequest = false,
    String? battleId,
  }) async {
    final battleMonsterId = _primaryMonsterIdForEncounter(monster);
    if (battleMonsterId.isEmpty) {
      if (mounted && parentContext.mounted) {
        ScaffoldMessenger.of(parentContext).showSnackBar(
          const SnackBar(
            content: Text('Could not start battle: monster id missing.'),
          ),
        );
      }
      return;
    }

    final poiService = context.read<PoiService>();
    Map<String, dynamic> battleDetail = const {};
    var resolvedBattleId = battleId?.trim() ?? '';
    var effectivePartyBattle = isPartyBattle;
    try {
      if (skipStartRequest) {
        battleDetail = resolvedBattleId.isNotEmpty
            ? await poiService.getMonsterBattleStatusById(resolvedBattleId)
            : await poiService.getMonsterBattleStatus(battleMonsterId);
      } else {
        battleDetail = await poiService.startMonsterBattle(battleMonsterId);
      }
      final battleRaw = battleDetail['battle'];
      final battle = battleRaw is Map<String, dynamic>
          ? battleRaw
          : (battleRaw is Map ? Map<String, dynamic>.from(battleRaw) : null);
      final fetchedBattleId = (battle?['id']?.toString() ?? '').trim();
      if (fetchedBattleId.isNotEmpty) {
        resolvedBattleId = fetchedBattleId;
      }
      if (_isPartyBattleDetail(battleDetail)) {
        effectivePartyBattle = true;
      }
    } catch (error) {
      var message = 'Could not start battle.';
      if (error is DioException && error.response?.data is Map) {
        final data = Map<String, dynamic>.from(error.response!.data as Map);
        final apiMessage = data['error']?.toString().trim() ?? '';
        if (apiMessage.isNotEmpty) {
          message = apiMessage;
        }
      }
      if (mounted && parentContext.mounted) {
        ScaffoldMessenger.of(
          parentContext,
        ).showSnackBar(SnackBar(content: Text(message)));
      }
      return;
    }

    bool readyForCombat;
    try {
      readyForCombat = await _waitForPartyBattleReady(
        parentContext,
        battleMonsterId,
        battleDetail,
        resolvedBattleId.isNotEmpty ? resolvedBattleId : null,
      );
    } catch (_) {
      if (mounted && parentContext.mounted) {
        ScaffoldMessenger.of(parentContext).showSnackBar(
          const SnackBar(
            content: Text('Could not verify party battle readiness.'),
          ),
        );
      }
      return;
    }
    if (!readyForCombat) {
      if (mounted && parentContext.mounted) {
        ScaffoldMessenger.of(parentContext).showSnackBar(
          const SnackBar(
            content: Text(
              'Party battle did not become ready. Try joining again.',
            ),
          ),
        );
      }
      return;
    }

    MonsterBattleResult? result;
    try {
      result = await showDialog<MonsterBattleResult>(
        context: parentContext,
        useRootNavigator: true,
        useSafeArea: false,
        barrierDismissible: false,
        builder: (_) => MonsterBattleDialog(
          encounter: monster,
          isPartyBattle: effectivePartyBattle,
          battleMonsterId: battleMonsterId,
          battleId: resolvedBattleId.isNotEmpty ? resolvedBattleId : null,
        ),
      );
    } finally {
      if (!effectivePartyBattle) {
        try {
          await poiService.endMonsterBattle(battleMonsterId);
        } catch (error) {
          debugPrint('Failed to end server monster battle: $error');
        }
      }
    }

    if (!mounted || result == null) return;
    final statsProvider = context.read<CharacterStatsProvider>();

    if (result.outcome == MonsterBattleOutcome.escaped) {
      await statsProvider.setHealthAndManaTo(
        health: result.playerHealthRemaining,
        mana: result.playerManaRemaining,
      );
      return;
    }

    if (result.outcome == MonsterBattleOutcome.victory) {
      await statsProvider.setHealthAndManaTo(
        health: result.playerHealthRemaining,
        mana: result.playerManaRemaining,
      );
    }

    if (result.outcome == MonsterBattleOutcome.victory) {
      setState(() {
        _defeatedMonsterIds.add(monster.id);
        _monsters.removeWhere((item) => item.id == monster.id);
      });
      await _persistDefeatedMonsterIds();
      await _refreshMonsterSymbols();
      if (!mounted || !parentContext.mounted) return;

      final itemTotals = <int, Map<String, dynamic>>{};
      for (final enemy in monster.monsters) {
        for (final reward in enemy.itemRewards) {
          final quantity = reward.quantity > 0 ? reward.quantity : 1;
          final entry = itemTotals.putIfAbsent(reward.inventoryItemId, () {
            return <String, dynamic>{
              'id': reward.inventoryItemId,
              'name': reward.inventoryItemName.isNotEmpty
                  ? reward.inventoryItemName
                  : 'Item #${reward.inventoryItemId}',
              'imageUrl': reward.inventoryItemImageUrl,
              'quantity': 0,
            };
          });
          entry['quantity'] = (entry['quantity'] as int) + quantity;
          if ((entry['imageUrl'] as String).isEmpty &&
              reward.inventoryItemImageUrl.isNotEmpty) {
            entry['imageUrl'] = reward.inventoryItemImageUrl;
          }
        }
      }
      final itemsAwarded = itemTotals.values.toList();
      parentContext.read<CompletedTaskProvider>().showModal(
        'monsterBattleVictory',
        data: {
          'monsterEncounterId': monster.id,
          'monsterName': monster.name,
          'rewardExperience': monster.totalRewardExperience,
          'rewardGold': monster.totalRewardGold,
          'itemsAwarded': itemsAwarded,
        },
      );
      return;
    }

    await statsProvider.setHealthAndManaTo(
      health: 1,
      mana: result.playerManaRemaining,
    );
    if (!mounted || !parentContext.mounted) return;
    parentContext.read<CompletedTaskProvider>().showModal(
      'monsterBattleDefeat',
      data: {'monsterName': monster.name, 'healthSetTo': 1},
    );
  }

  Future<void> _removeScenarioLocally(
    String scenarioId, {
    String? performedScenarioId,
    Scenario? fallbackScenario,
  }) async {
    if (!mounted) return;

    final scenarioIds = <String>{};
    final trimmedTappedId = scenarioId.trim();
    if (trimmedTappedId.isNotEmpty) {
      scenarioIds.add(trimmedTappedId);
    }
    final trimmedPerformedId = performedScenarioId?.trim() ?? '';
    if (trimmedPerformedId.isNotEmpty) {
      scenarioIds.add(trimmedPerformedId);
    }
    if (scenarioIds.isEmpty) return;

    final scenariosToRemove = _scenarios
        .where(
          (item) =>
              scenarioIds.contains(item.id) ||
              _scenarioMatchesFallback(item, fallbackScenario),
        )
        .toList();

    _resolvedScenarioIds.addAll(scenarioIds);
    for (final item in scenariosToRemove) {
      _resolvedScenarioSignatures.add(_scenarioSignature(item));
    }
    if (scenariosToRemove.isEmpty && fallbackScenario != null) {
      _resolvedScenarioSignatures.add(_scenarioSignature(fallbackScenario));
    }
    setState(() {
      _scenarios.removeWhere(
        (item) =>
            scenarioIds.contains(item.id) ||
            _scenarioMatchesFallback(item, fallbackScenario),
      );
    });

    await _refreshScenarioSymbols();
  }

  String _scenarioSignature(Scenario scenario) {
    final prompt = scenario.prompt.trim().toLowerCase();
    final lat = scenario.latitude.toStringAsFixed(5);
    final lng = scenario.longitude.toStringAsFixed(5);
    return '${scenario.zoneId}|$lat|$lng|$prompt';
  }

  bool _scenarioMatchesFallback(Scenario scenario, Scenario? fallback) {
    if (fallback == null) return false;
    const epsilon = 0.000001;
    final sameLat = (scenario.latitude - fallback.latitude).abs() <= epsilon;
    final sameLng = (scenario.longitude - fallback.longitude).abs() <= epsilon;
    if (!sameLat || !sameLng) return false;
    if (scenario.zoneId != fallback.zoneId) return false;

    final prompt = scenario.prompt.trim();
    final fallbackPrompt = fallback.prompt.trim();
    return prompt.isNotEmpty &&
        fallbackPrompt.isNotEmpty &&
        prompt == fallbackPrompt;
  }

  void _showScenarioPanel(Scenario scenario) {
    final parentContext = context;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => ScenarioPanel(
        scenario: scenario,
        onClose: () => Navigator.of(context).pop(),
        onPerformed: (result) async {
          if (!mounted) return;

          await _removeScenarioLocally(
            scenario.id,
            performedScenarioId: result.scenarioId,
            fallbackScenario: scenario,
          );
          unawaited(_loadTreasureChestsForSelectedZone());
          if (!mounted || !parentContext.mounted) return;

          ScenarioOption? selectedOption;
          if (result.scenarioOptionId != null &&
              result.scenarioOptionId!.isNotEmpty) {
            for (final option in scenario.options) {
              if (option.id == result.scenarioOptionId) {
                selectedOption = option;
                break;
              }
            }
          }

          final outcomeText = result.outcomeText.trim().isNotEmpty
              ? result.outcomeText.trim()
              : result.successful
              ? (selectedOption?.successText.trim().isNotEmpty == true
                    ? selectedOption!.successText.trim()
                    : 'Your approach succeeds.')
              : (selectedOption?.failureText.trim().isNotEmpty == true
                    ? selectedOption!.failureText.trim()
                    : 'Your approach falls short.');

          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted || !parentContext.mounted) return;
            parentContext.read<CompletedTaskProvider>().showModal(
              'scenarioOutcome',
              data: {
                'scenarioId': result.scenarioId,
                'scenarioPrompt': scenario.prompt,
                'successful': result.successful,
                'outcomeText': outcomeText,
                'reason': result.reason,
                'roll': result.roll,
                'statTag': result.statTag,
                'statValue': result.statValue,
                'proficiencies': result.proficiencies,
                'proficiencyBonus': result.proficiencyBonus,
                'creativityBonus': result.creativityBonus,
                'totalScore': result.totalScore,
                'threshold': result.threshold,
                'failureHealthDrained': result.failureHealthDrained,
                'failureManaDrained': result.failureManaDrained,
                'failureStatusesApplied': result.failureStatusesApplied
                    .map((status) => status.toJson())
                    .toList(),
                'successHealthRestored': result.successHealthRestored,
                'successManaRestored': result.successManaRestored,
                'successStatusesApplied': result.successStatusesApplied
                    .map((status) => status.toJson())
                    .toList(),
                'rewardExperience': result.rewardExperience,
                'rewardGold': result.rewardGold,
                'itemsAwarded': result.itemsAwarded,
                'itemChoiceRewards': result.itemChoiceRewards,
                'spellsAwarded': result.spellsAwarded
                    .map((spell) => spell.toJson())
                    .toList(),
              },
            );
          });
        },
      ),
    );
  }

  void _showActivityFeed(BuildContext context) {
    context.read<ActivityFeedProvider>().refresh();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => PaperSheet(
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Activity Feed',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  child: const Padding(
                    padding: EdgeInsets.all(16),
                    child: ActivityFeedPanel(),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showInventory(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => PaperSheet(
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Inventory',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  child: InventoryPanel(
                    onClose: () => Navigator.of(context).pop(),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _openQuestLogForQuest(Quest quest) {
    _showQuestLog(context, initialSelectedQuest: quest);
  }

  void _showQuestLog(BuildContext context, {Quest? initialSelectedQuest}) {
    _refreshQuestLog();
    context.read<TagsProvider>().refresh();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => PaperSheet(
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Quest Log',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: QuestLogPanel(
                  onClose: () => Navigator.of(context).pop(),
                  onFocusPoI: _focusQuestPoI,
                  onFocusTurnInQuest: _focusQuestTurnIn,
                  initialSelectedQuest: initialSelectedQuest,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showLog(BuildContext context) {
    context.read<LogProvider>().refresh();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => PaperSheet(
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Log', style: Theme.of(context).textTheme.titleLarge),
                    IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  child: const Padding(
                    padding: EdgeInsets.all(16),
                    child: LogPanel(),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showTagFilter(BuildContext context) {
    context.read<TagsProvider>().refresh();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: false,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => SafeArea(
        top: false,
        child: DraggableScrollableSheet(
          initialChildSize: 0.9,
          minChildSize: 0.6,
          maxChildSize: 0.95,
          builder: (_, scrollController) => PaperSheet(
            child: Column(
              children: [
                Container(
                  padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 36,
                        height: 4,
                        margin: const EdgeInsets.only(bottom: 8),
                        decoration: BoxDecoration(
                          color: Theme.of(
                            context,
                          ).colorScheme.onSurface.withValues(alpha: 0.3),
                          borderRadius: BorderRadius.circular(999),
                        ),
                      ),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            'Filters',
                            style: Theme.of(context).textTheme.titleLarge,
                          ),
                          IconButton(
                            onPressed: () => Navigator.of(context).pop(),
                            icon: const Icon(Icons.close),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: SingleChildScrollView(
                    controller: scrollController,
                    child: const Padding(
                      padding: EdgeInsets.all(16),
                      child: QuestFilterPanel(),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _MiniInfoChip extends StatelessWidget {
  const _MiniInfoChip({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceVariant.withValues(alpha: 0.55),
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

class _MapButton extends StatelessWidget {
  const _MapButton({required this.label, required this.onTap});

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final surfaceColor = theme.colorScheme.surface.withValues(alpha: 0.95);
    final borderColor = theme.colorScheme.outlineVariant;
    final textStyle = GoogleFonts.cinzel(
      fontWeight: FontWeight.w600,
      color: theme.colorScheme.onSurface,
    );
    return Material(
      color: surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: borderColor),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          child: Text(label, style: textStyle),
        ),
      ),
    );
  }
}

class _OverlayButton extends StatelessWidget {
  const _OverlayButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final surfaceColor = theme.colorScheme.surface.withValues(alpha: 0.95);
    final borderColor = theme.colorScheme.outlineVariant;
    return Material(
      color: surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: borderColor),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Icon(icon, size: 24, color: theme.colorScheme.onSurface),
        ),
      ),
    );
  }
}

class _PoiSymbolRequest {
  const _PoiSymbolRequest({
    required this.poiId,
    required this.isQuestCurrent,
    required this.options,
    required this.data,
  });

  final String poiId;
  final bool isQuestCurrent;
  final SymbolOptions options;
  final Map<String, dynamic> data;
}

class _PoiSymbolResult {
  const _PoiSymbolResult(this.request, this.symbol);

  final _PoiSymbolRequest request;
  final Symbol symbol;
}

class _PoiImageUpdate {
  const _PoiImageUpdate({
    required this.poi,
    required this.imageUrl,
    required this.isQuestCurrent,
    required this.hasMapContent,
    required this.undiscovered,
  });

  final PointOfInterest poi;
  final String? imageUrl;
  final bool isQuestCurrent;
  final bool hasMapContent;
  final bool undiscovered;
}

class _PoiImageUpdateResult {
  const _PoiImageUpdateResult(this.update, this.imageId, this.bytes);

  final _PoiImageUpdate update;
  final String? imageId;
  final Uint8List? bytes;
}
