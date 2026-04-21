import 'dart:async';
import 'dart:math' show Point;
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
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
import '../models/activity_feed.dart';
import '../models/base.dart';
import '../models/exposition.dart';
import '../models/healing_fountain.dart';
import '../models/inventory_item.dart';
import '../models/location.dart';
import '../models/monster.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../models/quest_node_objective.dart';
import '../models/resource.dart';
import '../models/scenario.dart';
import '../models/treasure_chest.dart';
import '../models/tutorial.dart';
import '../models/user.dart';
import '../models/zone.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/base_placement_provider.dart';
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/log_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/completed_task_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/quest_filter_provider.dart';
import '../providers/tags_provider.dart';
import '../providers/tutorial_replay_provider.dart';
import '../providers/user_level_provider.dart';
import '../providers/zone_provider.dart';
import '../providers/map_focus_provider.dart';
import '../providers/party_provider.dart';
import '../services/media_service.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';
import '../utils/poi_image_util.dart';
import '../utils/camera_capture.dart';
import '../constants/api_constants.dart';
import '../constants/gameplay_constants.dart';
import '../widgets/activity_feed_panel.dart';
import '../widgets/base_panel.dart';
import '../widgets/celebration_modal_manager.dart';
import '../widgets/character_panel.dart';
import '../widgets/healing_fountain_panel.dart';
import '../widgets/inventory_panel.dart';
import '../widgets/log_panel.dart';
import '../widgets/monster_battle_dialog.dart';
import '../widgets/monster_panel.dart';
import '../widgets/new_item_modal.dart';
import '../widgets/point_of_interest_panel.dart';
import '../widgets/resource_panel.dart';
import '../widgets/rpg_dialogue_modal.dart';
import '../widgets/scenario_panel.dart';
import '../widgets/tracked_quests_overlay.dart';
import '../widgets/shop_modal.dart';
import '../widgets/treasure_chest_panel.dart';
import '../widgets/tutorial_guide_chat_modal.dart';
import '../widgets/used_item_modal.dart';
import '../widgets/zone_widget.dart';
import '../widgets/paper_texture.dart';
import '../widgets/party_member_map_strip.dart';
import 'base_management_screen.dart';
import 'fetch_quest_turn_in_screen.dart';
import 'layout_shell.dart';

const _chestImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/treasure-chest-undiscovered.png';
const _scenarioMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png';
const _monsterMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png';
const _bossMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/boss-undiscovered.png';
const _raidMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/raid-undiscovered.png';
const _characterMysteryImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png';
const _challengeMysteryImageUrl = _scenarioMysteryImageUrl;
const _expositionMapIconImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/exposition-undiscovered.png';
const _healingFountainFallbackImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';
const _healingFountainDiscoveredImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/healing-fountain-discovered.png';
const _baseDiscoveredImageUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/base-discovered.png';
const _legacyMysteryImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';
const _defeatedMonstersPrefsKeyPrefix = 'single_player_defeated_monsters';
const _discoveredCharactersPrefsKeyPrefix =
    'single_player_discovered_characters';
const _tutorialGuideButtonAcknowledgedPrefsKeyPrefix =
    'single_player_tutorial_guide_button_acknowledged';
const _mapThumbnailVersion = 'v14';
const _standardMarkerThumbnailSize = 0.75;
const _baseMarkerSizeScale = 0.75;
const int _monsterBattleDefeatHealthFloorPercent = 30;
const int _monsterBattleDefeatManaFloorPercent = 25;
const int _monsterBattleDefeatStatusDurationMinutes = 15;
const _questPulseCoreColor = '#f7d46f';
const _questPulseMistColor = '#fff1c3';
const _questPulseRingColor = '#f1bb47';
const _mainStoryPulseCoreColor = '#b53a4b';
const _mainStoryPulseMistColor = '#f3cbd2';
const _mainStoryPulseRingColor = '#7a1823';
const _discoveryPulseCoreColor = '#f6d98c';
const _discoveryPulseMistColor = '#fff5d7';
const _discoveryPulseRingColor = '#f5c542';

int _monsterBattleDefeatResourceFloor(int maxResource, int floorPercent) {
  if (maxResource <= 0) return 0;
  final floor = ((maxResource * floorPercent) / 100).ceil();
  return math.max(1, math.min(maxResource, floor));
}

const _baseMarkerIconSize = 1.68 * _baseMarkerSizeScale;
const _basePlacementPreviewIconSize = 1.76 * _baseMarkerSizeScale;
const _baseFallbackCircleRadius = 42.0 * _baseMarkerSizeScale;
const _playerPresenceMarkerIconSize = 0.9;
const _playerPresenceAuraColor = '#5faab8';
const _playerPresencePulseColor = '#f4d989';
const _playerPresencePulseStrokeColor = '#f8ebc4';
const _playerPresenceConeFillColor = '#5faab8';
const _playerPresenceConeOutlineColor = '#f4e9d6';
const _playerPresenceConeMinLengthMeters = 28.0;
const _playerPresenceConeMaxLengthMeters = 46.0;
const _playerPresenceConeBaseHalfAngleDegrees = 26.0;
const _playerPresenceHeadingSpeedThreshold = 0.9;
const _playerPresenceBearingFallbackDistanceMeters = 6.0;
const _playerPresencePulseCycle = Duration(milliseconds: 1900);
const _playerPresencePulseFrameDelay = Duration(milliseconds: 90);
const _overlayRailButtonSize = 48.0;
const _overlayRailButtonSpacing = 12.0;
const _overlayRailButtonLeftInset = 16.0;
const _poiImageLoadBatchSize = 24;
const _poiSymbolAddBatchSize = 32;
const _zoneBaseContentFreshDuration = Duration(minutes: 2);
const _zoneBaseContentWarmupDebounce = Duration(milliseconds: 250);
const _zoneBaseContentWarmupThrottle = Duration(seconds: 2);
const _zoneBaseContentWarmCount = 4;
const _zoneBaseContentMaxCacheEntries = 6;
const _zoneBaseContentThumbnailWarmCount = 8;
const _defaultMapFocusZoom = 16.0;
const _trackedQuestOverlayFocusZoom = 14.0;
const _poiAssociationCoordinatePrecision = 4;
const _pinSelectionHitRadiusPx = 24.0;
const _playerUnderfootPinDistanceMeters = 18.0;
const _playerUnderfootPinAccuracyCapMeters = 24.0;
const _playerUnderfootTapHalfWidthPx = 36.0;
const _playerUnderfootTapTopReachPx = 72.0;
const _playerUnderfootTapBottomReachPx = 14.0;
const _transparentMapHaloColor = 'rgba(0,0,0,0)';
const _stamenWatercolorStyleBase =
    'https://tiles.stadiamaps.com/styles/stamen_watercolor.json';
const _stamenWatercolorApiKey = String.fromEnvironment(
  'STADIA_MAPS_API_KEY',
  defaultValue: '',
);
final String _stamenWatercolorStyle = _stamenWatercolorApiKey.isNotEmpty
    ? '$_stamenWatercolorStyleBase?api_key=$_stamenWatercolorApiKey'
    : _stamenWatercolorStyleBase;
final Set<Factory<OneSequenceGestureRecognizer>> _mapGestureRecognizers = {
  Factory<OneSequenceGestureRecognizer>(() => EagerGestureRecognizer()),
};

class SinglePlayerScreen extends StatefulWidget {
  const SinglePlayerScreen({super.key});

  @override
  State<SinglePlayerScreen> createState() => _SinglePlayerScreenState();
}

class _SinglePlayerScreenState extends State<SinglePlayerScreen>
    with TickerProviderStateMixin {
  MapLibreMapController? _mapController;
  List<Zone> _zones = [];
  List<PointOfInterest> _pois = [];
  List<Character> _characters = [];
  List<TreasureChest> _treasureChests = [];
  List<HealingFountain> _healingFountains = [];
  List<ResourceNode> _resources = [];
  List<BasePin> _bases = [];
  List<Scenario> _scenarios = [];
  List<Exposition> _expositions = [];
  List<MonsterEncounter> _monsters = [];
  List<Challenge> _challenges = [];
  List<Line> _zoneLines = [];
  List<Fill> _zoneFills = [];
  final Map<String, Fill> _zoneFillById = {};
  String? _renderedSelectedZoneId;
  List<Line> _questLines = [];
  List<Fill> _questFills = [];
  List<Symbol> _poiSymbols = [];
  final Map<String, Symbol> _poiSymbolById = {};
  final Set<String> _mapImageIds = <String>{};
  int _poiMarkerGeneration = 0;
  List<Symbol> _questPoiHighlightSymbols = [];
  List<Circle> _questPoiHighlightCircles = [];
  List<Symbol> _mainStoryPoiHighlightSymbols = [];
  List<Circle> _mainStoryPoiHighlightCircles = [];
  final Set<String> _activePulseKeys = <String>{};
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
  List<Symbol> _resourceSymbols = [];
  List<Circle> _resourceCircles = [];
  final Map<String, Symbol> _resourceSymbolById = {};
  final Map<String, Circle> _resourceCircleById = {};
  List<Symbol> _baseSymbols = [];
  List<Circle> _baseCircles = [];
  final Map<String, Symbol> _baseSymbolById = {};
  final Map<String, Circle> _baseCircleById = {};
  List<Symbol> _scenarioSymbols = [];
  List<Circle> _scenarioCircles = [];
  final Map<String, Symbol> _scenarioSymbolById = {};
  final Map<String, Circle> _scenarioCircleById = {};
  final Map<String, bool> _scenarioCircleMystery = {};
  final Map<String, bool> _scenarioQuestObjective = {};
  List<Symbol> _expositionSymbols = [];
  List<Circle> _expositionCircles = [];
  final Map<String, Symbol> _expositionSymbolById = {};
  final Map<String, Circle> _expositionCircleById = {};
  final Map<String, bool> _expositionQuestObjective = {};
  List<Symbol> _monsterSymbols = [];
  List<Circle> _monsterCircles = [];
  final Map<String, Symbol> _monsterSymbolById = {};
  final Map<String, Circle> _monsterCircleById = {};
  List<Symbol> _challengeSymbols = [];
  List<Circle> _challengeCircles = [];
  final Map<String, Symbol> _challengeSymbolById = {};
  final Map<String, Circle> _challengeCircleById = {};
  List<Line> _challengePolygonLines = [];
  List<Fill> _challengePolygonFills = [];
  final Map<String, Line> _challengePolygonLineById = {};
  final Map<String, Fill> _challengePolygonFillById = {};
  final Set<String> _resolvedScenarioIds = <String>{};
  final Set<String> _resolvedScenarioSignatures = <String>{};
  final Set<String> _openedTreasureChestIds = <String>{};
  final Set<String> _gatheredResourceIds = <String>{};
  final Set<String> _defeatedMonsterIds = <String>{};
  final Set<String> _discoveredCharacterIds = <String>{};
  String? _defeatedMonsterIdsUserId;
  String? _discoveredCharacterIdsUserId;
  final ZoneWidgetController _zoneWidgetController = ZoneWidgetController();
  Uint8List? _chestThumbnailBytes;
  bool _chestThumbnailAdded = false;
  Uint8List? _scenarioMysteryThumbnailBytes;
  bool _scenarioMysteryThumbnailAdded = false;
  Uint8List? _expositionMysteryThumbnailBytes;
  bool _expositionMysteryThumbnailAdded = false;
  final Map<String, Uint8List?> _monsterMysteryThumbnailBytesByType = {};
  final Set<String> _monsterMysteryThumbnailTypesAdded = {};
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
  ActivityFeedProvider? _activityFeedProvider;
  BasePlacementProvider? _basePlacementProvider;
  Map<String, dynamic>? _lastHandledCompletionModal;
  VoidCallback? _pendingCompletionModalDrainAction;
  final Set<String> _handledLevelUpActivityIds = <String>{};
  final Set<String> _zoneDiscoveryInFlightIds = <String>{};
  bool _levelUpActivityDrainScheduled = false;
  String _lastAuthenticatedUserId = '';
  Timer? _questGlowTimer;
  bool _isQuestGlowPulsing = false;
  Timer? _questPoiPulseTimer;
  String? _questLogRequestedZoneId;
  bool _questLogRefreshInFlight = false;
  bool _questAvailabilityLeadSyncPending = false;
  int _questAvailabilityLeadSyncGeneration = 0;
  bool _manualRefreshInFlight = false;
  int _skipQuestLogMapRefreshCount = 0;
  DateTime? _lastQuestLogRefreshAt;
  bool _questLogNeedsOverlayApply = false;
  _MapMarkerIsolation? _mapMarkerIsolation;
  int _mapMarkerIsolationToken = 0;
  bool _scenarioVisibilityRefreshPending = false;
  Future<void> _scenarioRefreshSequence = Future<void>.value();
  Future<void> _expositionRefreshSequence = Future<void>.value();
  Future<void> _monsterRefreshSequence = Future<void>.value();
  Future<void> _challengeRefreshSequence = Future<void>.value();
  Future<void> _zoneBoundaryRefreshSequence = Future<void>.value();
  Set<String> _lastQuestPoiIds = <String>{};
  Set<String> _lastTrackedQuestIds = <String>{};
  DateTime? _lastFeatureTapAt;
  Point<double>? _lastFeatureTapPoint;
  Set<String> _lastQuestTurnInCharacterIds = <String>{};
  String? _lastFeaturedMainStoryPulsePoiId;
  String? _lastFeaturedMainStoryPulseCharacterId;
  Map<String, String> _lastTrackedQuestObjectiveSignatures = <String, String>{};
  bool _hasTrackedQuestObjectiveSnapshot = false;
  String _lastTutorialTrackedObjectiveSignature = '';
  bool _hasTutorialTrackedObjectiveSnapshot = false;
  int _lastQuestPolygonHash = 0;
  String _lastMapFilterKey = '';
  bool _pinBatchRevealInProgress = false;
  int _zoneContentRequestVersion = 0;
  final Map<String, _ZoneBaseContentCacheEntry> _zoneBaseContentCache = {};
  final Map<String, Future<_ZoneBaseContent>> _zoneBaseContentRequests = {};
  final Map<String, _ZonePinContentCacheEntry> _zonePinContentCache = {};
  final Map<String, Future<_ZonePinContent>> _zonePinContentRequests = {};
  final Map<String, PointOfInterest> _knownPoiById = {};
  final Map<String, Character> _knownCharacterById = {};
  Timer? _zoneBaseContentWarmupTimer;
  String _lastZoneBaseContentWarmSignature = '';
  DateTime? _lastZoneBaseContentWarmAt;
  String? _renderedTreasureChestZoneId;
  String? _loadingZoneTransitionZoneId;
  QuestSubmissionOverlayPhase _questSubmissionPhase =
      QuestSubmissionOverlayPhase.hidden;
  String? _questSubmissionMessage;
  String? _questSubmissionStepLabel;
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
  bool _tutorialStatusReloadQueued = false;
  bool _tutorialStatusReloadQueuedPreserveCompletedReveal = false;
  bool _tutorialStatusChecked = false;
  bool _tutorialDialogVisible = false;
  bool _tutorialAdvanceInFlight = false;
  bool _tutorialActivationInFlight = false;
  bool _tutorialReplayResetInFlight = false;
  bool _tutorialReplayPending = false;
  int _lastTutorialReplayRequestCount = 0;
  bool _loadAllInFlight = false;
  int _loadAllGeneration = 0;
  String? _tutorialFocusedScenarioId;
  String? _tutorialFocusedMonsterEncounterId;
  bool _tutorialNormalPinsRevealInProgress = false;
  bool _tutorialLoadoutPendingAfterCompletionModal = false;
  bool _tutorialPostMonsterDialoguePendingAfterCompletionModal = false;
  bool _tutorialRevealPendingAfterCompletionModal = false;
  bool _tutorialWelcomeOverlayVisible = false;
  double _tutorialWelcomeOverlayOpacity = 0.0;
  late final AnimationController _tutorialGuideDockController;
  late final AnimationController _tutorialGuideButtonPulseController;
  bool _tutorialGuideDockVisible = false;
  Character? _tutorialGuideDockCharacter;
  String _tutorialGuideDockExcerpt = '';
  bool _tutorialGuideButtonAcknowledged = false;
  String _lastTutorialQuestSyncSignature = '';
  Set<String> _lastAcceptedTrackedTutorialQuestIds = <String>{};
  bool _hasAcceptedTrackedTutorialQuestSnapshot = false;
  String? _pendingBaseOwnedInventoryItemId;
  InventoryItem? _pendingBaseInventoryItem;
  LatLng? _pendingBaseSelection;
  bool _creatingBase = false;
  Symbol? _basePlacementPreviewSymbol;
  Uint8List? _basePlacementPreviewBytes;
  Symbol? _playerPresenceSymbol;
  Circle? _playerPresenceAuraCircle;
  Circle? _playerPresencePulseCircle;
  Fill? _playerPresenceConeFill;
  Line? _playerPresenceConeLine;
  Timer? _playerPresencePulseTimer;
  int _playerPresenceRefreshGeneration = 0;
  LatLng? _lastPlayerPresenceLatLng;
  double? _lastResolvedPlayerHeading;

  @override
  void initState() {
    super.initState();
    _tutorialGuideDockController =
        AnimationController(
          vsync: this,
          duration: const Duration(milliseconds: 820),
        )..addListener(() {
          if (!mounted || !_tutorialGuideDockVisible) return;
          setState(() {});
        });
    _tutorialGuideButtonPulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1700),
    );
    debugPrint('SinglePlayer: initState');
    _startMapLoadTimeout();
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
      _activityFeedProvider = context.read<ActivityFeedProvider>();
      _activityFeedProvider?.addListener(_onActivityFeedChanged);
      _basePlacementProvider = context.read<BasePlacementProvider>();
      _basePlacementProvider?.addListener(_onBasePlacementRequested);
      _lastAuthenticatedUserId =
          context.read<AuthProvider>().user?.id.trim() ?? '';
      unawaited(_restoreTutorialGuideButtonAcknowledgement());
      context.read<TutorialReplayProvider>().addListener(
        _onTutorialReplayRequested,
      );
      _onCompletedTaskModalChanged();
      _onActivityFeedChanged();
      _updateSelectedZoneFromLocation();
      _requestQuestLogIfReady();
      context.read<ActivityFeedProvider>().refresh();
      unawaited(context.read<PartyProvider>().fetchParty());
      unawaited(_loadTutorialStatus(force: true));
      _onTutorialReplayRequested();
      _onBasePlacementRequested();
      _queueLoadAllIfAuthenticated();
    });
  }

  @override
  void dispose() {
    _tutorialGuideDockController.dispose();
    _tutorialGuideButtonPulseController.dispose();
    _mapLoadTimeout?.cancel();
    _questGlowTimer?.cancel();
    _questPoiPulseTimer?.cancel();
    _zoneBaseContentWarmupTimer?.cancel();
    _playerPresencePulseTimer?.cancel();
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
      _activityFeedProvider?.removeListener(_onActivityFeedChanged);
    } catch (_) {}
    try {
      _basePlacementProvider?.removeListener(_onBasePlacementRequested);
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
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    if (_styleLoaded &&
        _mapController != null &&
        _markersAdded &&
        !_isTutorialMapFocusActive &&
        !_tutorialNormalPinsRevealInProgress &&
        _hasZoneBaseContentSnapshot(selectedZoneId) &&
        _hasZonePinContentSnapshot(selectedZoneId)) {
      _pinBatchRevealInProgress = true;
    }
    unawaited(_loadTreasureChestsForSelectedZone());
    _scheduleZoneBaseContentWarmup(immediate: true);
    unawaited(_addZoneBoundaries());
    if (_styleLoaded && _mapController != null && !_markersAdded) {
      unawaited(_addPoiMarkers());
    }
  }

  void _onLocationChanged() {
    if (!mounted) return;
    _updateSelectedZoneFromLocation();
    _requestQuestLogIfReady();
    _scheduleZoneBaseContentWarmup();
    _refreshScenarioVisibilityForLocationChange();
    _maybeShowTutorialDialogues();
    unawaited(_refreshPlayerPresence());
  }

  void _onActivityFeedChanged() {
    if (!mounted || _levelUpActivityDrainScheduled) return;
    _levelUpActivityDrainScheduled = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _levelUpActivityDrainScheduled = false;
      if (!mounted) return;
      unawaited(_drainPendingLevelUpActivities());
    });
  }

  Future<void> _drainPendingLevelUpActivities() async {
    final feed = _activityFeedProvider;
    final completedTaskProvider = _completedTaskProvider;
    if (feed == null || completedTaskProvider == null) return;

    final levelUpActivities =
        feed.activities
            .where(
              (activity) =>
                  activity.activityType == 'level_up' && !activity.seen,
            )
            .toList(growable: false)
          ..sort(_compareActivityChronology);
    final activityIdsToMarkSeen = <String>[];

    for (final activity in levelUpActivities) {
      final activityId = activity.id.trim();
      if (activityId.isEmpty || !_handledLevelUpActivityIds.add(activityId)) {
        continue;
      }

      final newLevel = _parseLevelUpActivityInt(activity.data['newLevel']);
      if (newLevel <= 0) continue;

      final previousLevel = _parseLevelUpActivityInt(
        activity.data['previousLevel'],
      );
      final explicitLevelsGained = _parseLevelUpActivityInt(
        activity.data['levelsGained'],
      );
      final levelsGained = explicitLevelsGained > 0
          ? explicitLevelsGained
          : (previousLevel > 0 && newLevel > previousLevel
                ? newLevel - previousLevel
                : 1);

      completedTaskProvider.queueLevelUpModal(
        newLevel: newLevel,
        previousLevel: previousLevel > 0 ? previousLevel : null,
        levelsGained: levelsGained,
      );

      if (!activity.seen) {
        activityIdsToMarkSeen.add(activityId);
      }
    }

    if (activityIdsToMarkSeen.isEmpty) return;
    try {
      await feed.markAsSeen(activityIdsToMarkSeen);
    } catch (_) {
      // Feed cleanup should never block level-up presentation.
    }
  }

  int _compareActivityChronology(ActivityFeed a, ActivityFeed b) {
    final aTime = DateTime.tryParse(a.createdAt);
    final bTime = DateTime.tryParse(b.createdAt);
    if (aTime != null && bTime != null) {
      return aTime.compareTo(bTime);
    }
    if (aTime != null) return -1;
    if (bTime != null) return 1;
    return a.createdAt.compareTo(b.createdAt);
  }

  int _parseLevelUpActivityInt(dynamic raw) {
    if (raw is num) return raw.toInt();
    return int.tryParse(raw?.toString().trim() ?? '') ?? 0;
  }

  void _refreshScenarioVisibilityForLocationChange() {
    if (_styleLoaded && _mapController != null && _markersAdded) {
      _scenarioVisibilityRefreshPending = false;
      if (_isTutorialMapFocusActive) {
        unawaited(
          (() async {
            await _refreshScenarioSymbols();
            await _refreshExpositionSymbols();
            await _refreshMonsterSymbols();
            await _refreshChallengeSymbols();
          })(),
        );
        return;
      }
      unawaited(
        (() async {
          await _refreshScenarioSymbols();
          await _refreshExpositionSymbols();
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
    final auth = context.read<AuthProvider>();
    final userId = auth.user?.id.trim() ?? '';
    if (userId != _lastAuthenticatedUserId) {
      _lastAuthenticatedUserId = userId;
      _handledLevelUpActivityIds.clear();
      _completedTaskProvider?.reset();
      _tutorialGuideButtonAcknowledged = false;
      _lastAcceptedTrackedTutorialQuestIds = <String>{};
      _hasAcceptedTrackedTutorialQuestSnapshot = false;
      unawaited(_restoreTutorialGuideButtonAcknowledgement());
    }
    if (auth.loading || !auth.isAuthenticated) {
      _tutorialGuideButtonPulseController.stop();
      unawaited(_clearPlayerPresenceOverlays());
      return;
    }
    _queueLoadAllIfAuthenticated();
    _requestQuestLogIfReady(force: true);
    unawaited(_restoreDefeatedMonsterIds(refreshMap: true));
    unawaited(_restoreDiscoveredCharacterIds(refreshMap: true));
    unawaited(_loadBases());
    unawaited(context.read<PartyProvider>().fetchParty());
    unawaited(_loadTutorialStatus(force: true));
    unawaited(_refreshPlayerPresence());
  }

  void _queueLoadAllIfAuthenticated() {
    if (!mounted || _loadAllInFlight) return;
    final auth = context.read<AuthProvider>();
    if (auth.loading || !auth.isAuthenticated) {
      return;
    }
    unawaited(_loadAll());
  }

  void _onTutorialReplayRequested() {
    if (!mounted) return;
    final provider = context.read<TutorialReplayProvider>();
    final requestCount = provider.requestCount;
    if (requestCount == _lastTutorialReplayRequestCount) return;
    _lastTutorialReplayRequestCount = requestCount;
    _tutorialReplayPending = true;
    unawaited(_resetTutorialForReplay());
  }

  Future<void> _resetTutorialForReplay() async {
    if (!mounted || _tutorialReplayResetInFlight) return;
    setState(() {
      _tutorialReplayResetInFlight = true;
      _tutorialFocusedScenarioId = null;
      _tutorialFocusedMonsterEncounterId = null;
      _tutorialNormalPinsRevealInProgress = false;
      _tutorialWelcomeOverlayVisible = false;
      _tutorialWelcomeOverlayOpacity = 0.0;
      _tutorialLoadoutPendingAfterCompletionModal = false;
      _tutorialPostMonsterDialoguePendingAfterCompletionModal = false;
      _tutorialRevealPendingAfterCompletionModal = false;
    });

    try {
      final status = await context.read<PoiService>().resetTutorial();
      if (!mounted) return;
      await _clearTutorialGuideButtonAcknowledgement();
      if (!mounted) return;
      await _applyTutorialStatusUpdate(status);
      await _rebuildMapPins();
    } catch (error) {
      if (!mounted) return;
      final message = PoiService.extractApiErrorMessage(
        error,
        'Failed to reset the tutorial.',
      );
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
      await _loadTutorialStatus(force: true, preserveCompletedReveal: true);
    } finally {
      if (mounted) {
        setState(() => _tutorialReplayResetInFlight = false);
      }
    }
    if (!mounted) return;
    _maybeShowTutorialDialogues();
  }

  bool get _isTutorialMapFocusActive =>
      !_tutorialNormalPinsRevealInProgress &&
      ((_tutorialFocusedScenarioId?.trim().isNotEmpty ?? false) ||
          (_tutorialFocusedMonsterEncounterId?.trim().isNotEmpty ?? false) ||
          (_tutorialStatus?.isLoadoutStep ?? false));

  bool get _shouldSuppressNormalMapPinsForTutorial {
    if (_tutorialReplayResetInFlight) return true;
    if (!_tutorialStatusChecked) return true;
    final status = _tutorialStatus;
    if (status == null) return false;
    return !status.isCompleted;
  }

  bool get _isPlacingBase => _pendingBaseOwnedInventoryItemId != null;

  BasePin? _baseById(String id) {
    for (final base in _bases) {
      if (base.id == id) return base;
    }
    return null;
  }

  BasePin? _currentUserBase() {
    for (final base in _bases) {
      if (_isCurrentUserBase(base)) return base;
    }
    return null;
  }

  bool _isCurrentUserBase(BasePin base) {
    final userId = context.read<AuthProvider>().user?.id ?? '';
    return userId.isNotEmpty && base.userId == userId;
  }

  bool _isTutorialFocusedScenarioId(String scenarioId) {
    final focusedId = _tutorialFocusedScenarioId?.trim() ?? '';
    return focusedId.isNotEmpty && focusedId == scenarioId.trim();
  }

  bool _isTutorialFocusedMonsterEncounterId(String encounterId) {
    final focusedId = _tutorialFocusedMonsterEncounterId?.trim() ?? '';
    return focusedId.isNotEmpty && focusedId == encounterId.trim();
  }

  double _mapMarkerStartingOpacity(double targetOpacity) {
    if (_pinBatchRevealInProgress) {
      return 0.0;
    }
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
    final drawerController = LayoutShellDrawerController.maybeReadOf(context);
    if (drawerController == null) return;
    if (status == null) {
      return;
    }
    final isInventoryTutorialStep =
        status.isLoadoutStep || status.isBaseKitStep;
    if (!isInventoryTutorialStep) {
      drawerController.stopInventoryTutorial();
      return;
    }
    final completionModalOpen = _completedTaskProvider?.currentModal != null;
    if (completionModalOpen ||
        _tutorialLoadoutPendingAfterCompletionModal ||
        _tutorialPostMonsterDialoguePendingAfterCompletionModal ||
        _tutorialRevealPendingAfterCompletionModal) {
      return;
    }
    drawerController.startInventoryTutorial(
      InventoryTutorialSession(
        dialogue: status.isBaseKitStep
            ? status.baseKitDialogue
            : status.loadoutDialogue,
        objectiveCopy: status.isBaseKitStep
            ? status.resolvedBaseKitObjectiveCopy
            : status.resolvedLoadoutObjectiveCopy,
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
    await _loadTutorialStatus(force: true, preserveCompletedReveal: true);
    await _rebuildMapPins();
    await _runTutorialWelcomeOverlaySequence();
  }

  Future<void> _beginTutorialLoadoutStep() async {
    if (!mounted) return;
    await _loadTutorialStatus(force: true);
    if (!mounted) return;
    final status = _tutorialStatus;
    if ((status?.isLoadoutStep ?? false) || (status?.isBaseKitStep ?? false)) {
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
    if ((status?.isPostScenarioDialogueStep ?? false) ||
        (status?.isPostWelcomeDialogueStep ?? false) ||
        (status?.isPostMonsterDialogueStep ?? false) ||
        (status?.isPostBaseDialogueStep ?? false)) {
      return;
    }
    if (status?.isCompleted ?? false) {
      await _beginTutorialNormalPinsReveal();
    }
  }

  Future<void> _beginTutorialPostMonsterDialogueStep() async {
    if (!mounted) return;
    await _loadTutorialStatus(force: true);
  }

  Future<void> _runTutorialWelcomeOverlaySequence() async {
    if (!mounted) return;
    const fadeInSteps = 4;
    for (var i = 1; i <= fadeInSteps; i++) {
      if (!mounted) return;
      setState(() => _tutorialWelcomeOverlayOpacity = i / fadeInSteps);
      await Future<void>.delayed(const Duration(milliseconds: 35));
    }
  }

  Future<void> _dismissTutorialWelcomeOverlay() async {
    if (!mounted || !_tutorialWelcomeOverlayVisible) return;
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
    await _updateNormalPinOpacities(
      c,
      1,
      questPoiIds: questPoiIds,
      discoveries: discoveries,
      mapContentPoiIds: mapContentPoiIds,
    );

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
      final pendingCompletionModalDrainAction =
          _pendingCompletionModalDrainAction;
      _pendingCompletionModalDrainAction = null;
      if (pendingCompletionModalDrainAction != null) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!mounted) return;
          pendingCompletionModalDrainAction();
        });
      }
      if (_tutorialLoadoutPendingAfterCompletionModal) {
        _tutorialLoadoutPendingAfterCompletionModal = false;
        unawaited(_beginTutorialLoadoutStep());
      }
      if (_tutorialPostMonsterDialoguePendingAfterCompletionModal) {
        _tutorialPostMonsterDialoguePendingAfterCompletionModal = false;
        unawaited(_beginTutorialPostMonsterDialogueStep());
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

  void _runAfterCompletionModals(VoidCallback action) {
    final provider = _completedTaskProvider;
    if (provider?.currentModal == null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        action();
      });
      return;
    }
    _pendingCompletionModalDrainAction = action;
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
        await _removeMonsterEncounterLocally(encounterId);
        if (!mounted) return;
        setState(() {
          _tutorialFocusedMonsterEncounterId = null;
        });
        _tutorialPostMonsterDialoguePendingAfterCompletionModal = true;
        return;
      default:
        return;
    }
  }

  Future<int> _refreshRewardDrivenPlayerState() async {
    final statsProvider = context.read<CharacterStatsProvider>();
    final authProvider = context.read<AuthProvider>();
    final activityFeedProvider = context.read<ActivityFeedProvider>();
    final userLevelProvider = context.read<UserLevelProvider>();
    await _runBestEffortRefresh(
      'character stats',
      statsProvider.refresh(silent: true),
      timeout: const Duration(seconds: 5),
    );
    unawaited(_runBestEffortRefresh('auth', authProvider.refresh()));
    unawaited(
      _runBestEffortRefresh('activity feed', activityFeedProvider.refresh()),
    );
    unawaited(_runBestEffortRefresh('user level', userLevelProvider.refresh()));
    return statsProvider.level;
  }

  Future<void> _runBestEffortRefresh(
    String label,
    Future<void> refresh, {
    Duration timeout = const Duration(seconds: 5),
  }) async {
    try {
      await refresh.timeout(timeout);
    } catch (error) {
      debugPrint('[SinglePlayerScreen] $label refresh skipped: $error');
    }
  }

  void _queueRewardLevelUpFromData(Map<String, dynamic> data) {
    final leveledUp = data['leveledUp'] == true;
    if (!leveledUp) return;
    final completedTaskProvider = context.read<CompletedTaskProvider>();
    completedTaskProvider.queueLevelUpModal(
      newLevel:
          (data['newLevel'] as num?)?.toInt() ??
          context.read<CharacterStatsProvider>().level,
      previousLevel: (data['previousLevel'] as num?)?.toInt(),
      levelsGained: (data['levelsGained'] as num?)?.toInt() ?? 1,
    );
  }

  Future<void> _removeMonsterEncounterLocally(String encounterId) async {
    final trimmedId = encounterId.trim();
    if (trimmedId.isEmpty) return;

    if (mounted) {
      setState(() {
        _monsters.removeWhere((item) => item.id == trimmedId);
      });
    } else {
      _monsters.removeWhere((item) => item.id == trimmedId);
    }

    final controller = _mapController;
    if (controller == null || !_styleLoaded) return;

    final symbol = _monsterSymbolById.remove(trimmedId);
    if (symbol != null) {
      _setQuestPoiHighlight(symbol, false);
      _monsterSymbols.remove(symbol);
      try {
        await controller.removeSymbols([symbol]);
      } catch (_) {}
    }

    final circle = _monsterCircleById.remove(trimmedId);
    if (circle != null) {
      _monsterCircles.remove(circle);
      try {
        await controller.removeCircle(circle);
      } catch (_) {}
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
    if (questLog.loading) {
      _setQuestAvailabilityLeadSyncPending(true);
      return;
    }
    _openTrackedQuestsForTutorialQuestAwards(questLog);
    _openTrackedQuestsForObjectiveUpdates(questLog);
    _applyQuestLogOverlaysIfChanged();
    if (_skipQuestLogMapRefreshCount > 0) {
      _skipQuestLogMapRefreshCount -= 1;
      _setQuestAvailabilityLeadSyncPending(false);
      return;
    }
    final syncGeneration = ++_questAvailabilityLeadSyncGeneration;
    _setQuestAvailabilityLeadSyncPending(true);
    // Quest availability is embedded in zone pin payloads, so quest-log changes
    // need a fresh pin fetch before showing any main-story availability lead.
    unawaited(_refreshQuestAvailabilityAfterQuestLogChange(syncGeneration));
  }

  void _onMapFocusRequest() {
    if (!mounted) return;
    final provider = context.read<MapFocusProvider>();
    final poi = provider.consumePoi();
    if (poi != null) {
      _focusQuestPoI(poi);
      return;
    }
    final node = provider.consumeNode();
    if (node != null) {
      _focusQuestNode(node);
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

  void _setQuestAvailabilityLeadSyncPending(bool pending) {
    if (_questAvailabilityLeadSyncPending == pending) {
      return;
    }
    void update() {
      _questAvailabilityLeadSyncPending = pending;
    }

    if (mounted) {
      setState(update);
    } else {
      update();
    }
  }

  Future<void> _refreshQuestAvailabilityAfterQuestLogChange(
    int syncGeneration,
  ) async {
    try {
      await _loadTreasureChestsForSelectedZone(forceRefreshZonePins: true);
    } catch (_) {
      // Best-effort refresh; keep existing pins if the forced reload fails.
    }
    if (!mounted || syncGeneration != _questAvailabilityLeadSyncGeneration) {
      return;
    }
    _setQuestAvailabilityLeadSyncPending(false);
    _applyQuestLogOverlaysIfChanged();
  }

  void _applyQuestLogOverlaysIfChanged() {
    if (!_styleLoaded || _mapController == null) {
      _questLogNeedsOverlayApply = true;
      return;
    }
    final questLog = context.read<QuestLogProvider>();
    final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
    final trackedQuestPoiIds = _trackedQuestPoiIdsForPulse(questLog);
    final turnInCharacterIds = _currentQuestTurnInCharacterIds(questLog);
    final featuredMainStoryPulseTarget = _featuredMainStoryPulseTarget(
      questLog,
    );
    final featuredMainStoryPulsePoiId = featuredMainStoryPulseTarget?.poiId;
    final featuredMainStoryPulseCharacterId =
        featuredMainStoryPulseTarget?.characterId;
    final trackedQuestIds = questLog.trackedQuestIds
        .map((id) => id.trim())
        .where((id) => id.isNotEmpty)
        .toSet();
    final trackedPolygons = _trackedQuestCurrentNodePolygons(questLog);
    final polygonHash = _hashQuestPolygons(trackedPolygons);
    debugPrint(
      'SinglePlayer: quest overlay check poiIds=${questPoiIds.length} trackedPoiIds=${trackedQuestPoiIds.length} polys=${trackedPolygons.length}',
    );
    final poiChanged = !_setEquals(_lastQuestPoiIds, questPoiIds);
    final trackedIdsChanged = !_setEquals(
      _lastTrackedQuestIds,
      trackedQuestIds,
    );
    final turnInCharactersChanged = !_setEquals(
      _lastQuestTurnInCharacterIds,
      turnInCharacterIds,
    );
    final featuredMainStoryPulsePoiChanged =
        _lastFeaturedMainStoryPulsePoiId != featuredMainStoryPulsePoiId;
    final featuredMainStoryPulseCharacterChanged =
        _lastFeaturedMainStoryPulseCharacterId !=
        featuredMainStoryPulseCharacterId;
    final polyChanged = polygonHash != _lastQuestPolygonHash;
    if (!poiChanged &&
        !turnInCharactersChanged &&
        !featuredMainStoryPulsePoiChanged &&
        !featuredMainStoryPulseCharacterChanged &&
        !polyChanged) {
      return;
    }
    final newlyAddedPoiIds = questPoiIds.difference(_lastQuestPoiIds);
    final removedPoiIds = _lastQuestPoiIds.difference(questPoiIds);
    final newlyAddedTurnInCharacterIds = turnInCharacterIds.difference(
      _lastQuestTurnInCharacterIds,
    );
    final removedTurnInCharacterIds = _lastQuestTurnInCharacterIds.difference(
      turnInCharacterIds,
    );
    final previousFeaturedMainStoryPulsePoiId =
        _lastFeaturedMainStoryPulsePoiId;
    final previousFeaturedMainStoryPulseCharacterId =
        _lastFeaturedMainStoryPulseCharacterId;
    _lastQuestPoiIds = questPoiIds;
    _lastTrackedQuestIds = trackedQuestIds;
    _lastQuestTurnInCharacterIds = turnInCharacterIds;
    _lastFeaturedMainStoryPulsePoiId = featuredMainStoryPulsePoiId;
    _lastFeaturedMainStoryPulseCharacterId = featuredMainStoryPulseCharacterId;
    _lastQuestPolygonHash = polygonHash;
    if (poiChanged && newlyAddedPoiIds.isNotEmpty) {
      for (final poiId in newlyAddedPoiIds.intersection(trackedQuestPoiIds)) {
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
    if (trackedIdsChanged) {
      for (final poiId in questPoiIds) {
        unawaited(_updatePoiSymbolForQuestState(poiId, isQuestCurrent: true));
      }
      for (final characterId in turnInCharacterIds) {
        final character = _characterById(characterId);
        if (character == null) continue;
        unawaited(_updateCharacterSymbolForState(character));
      }
      unawaited(_refreshScenarioSymbols());
      unawaited(_refreshExpositionSymbols());
      unawaited(_refreshMonsterSymbols());
      unawaited(_refreshChallengeSymbols());
    }
    if (turnInCharactersChanged) {
      final changedCharacterIds = <String>{
        ...newlyAddedTurnInCharacterIds,
        ...removedTurnInCharacterIds,
      };
      for (final characterId in changedCharacterIds) {
        final character = _characterById(characterId);
        if (character == null) continue;
        unawaited(_updateCharacterSymbolForState(character));
      }
    }
    if (featuredMainStoryPulsePoiChanged) {
      final changedPoiIds = <String>{
        if (previousFeaturedMainStoryPulsePoiId?.isNotEmpty == true)
          previousFeaturedMainStoryPulsePoiId!,
        if (featuredMainStoryPulsePoiId?.isNotEmpty == true)
          featuredMainStoryPulsePoiId!,
      };
      for (final poiId in changedPoiIds) {
        unawaited(
          _updatePoiSymbolForQuestState(
            poiId,
            isQuestCurrent: questPoiIds.contains(poiId),
          ),
        );
      }
    }
    if (featuredMainStoryPulseCharacterChanged) {
      final changedCharacterIds = <String>{
        if (previousFeaturedMainStoryPulseCharacterId?.isNotEmpty == true)
          previousFeaturedMainStoryPulseCharacterId!,
        if (featuredMainStoryPulseCharacterId?.isNotEmpty == true)
          featuredMainStoryPulseCharacterId!,
      };
      for (final characterId in changedCharacterIds) {
        final character = _characterById(characterId);
        if (character == null) continue;
        unawaited(_updateCharacterSymbolForState(character));
      }
    }
    if (polyChanged) {
      unawaited(_addQuestPolygons());
      for (final poly in trackedPolygons) {
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

  String _tutorialGuideButtonAcknowledgedPrefsKey(String userId) {
    return '${_tutorialGuideButtonAcknowledgedPrefsKeyPrefix}_$userId';
  }

  Future<void> _restoreTutorialGuideButtonAcknowledgement() async {
    if (!mounted) return;
    final userId = context.read<AuthProvider>().user?.id.trim() ?? '';
    if (userId.isEmpty) {
      setState(() => _tutorialGuideButtonAcknowledged = false);
      _syncTutorialGuideButtonPulse(_tutorialStatus);
      return;
    }

    final prefs = await SharedPreferences.getInstance();
    if (!mounted) return;
    final currentUserId = context.read<AuthProvider>().user?.id.trim() ?? '';
    if (currentUserId != userId) return;
    final acknowledged =
        prefs.getBool(_tutorialGuideButtonAcknowledgedPrefsKey(userId)) ??
        false;
    setState(() => _tutorialGuideButtonAcknowledged = acknowledged);
    _syncTutorialGuideButtonPulse(_tutorialStatus);
  }

  Future<void> _markTutorialGuideButtonAcknowledged() async {
    if (_tutorialGuideButtonAcknowledged) return;

    if (mounted) {
      setState(() => _tutorialGuideButtonAcknowledged = true);
    } else {
      _tutorialGuideButtonAcknowledged = true;
    }
    _syncTutorialGuideButtonPulse(_tutorialStatus);

    final userId = context.read<AuthProvider>().user?.id.trim() ?? '';
    if (userId.isEmpty) return;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_tutorialGuideButtonAcknowledgedPrefsKey(userId), true);
  }

  Future<void> _clearTutorialGuideButtonAcknowledgement() async {
    if (mounted) {
      setState(() => _tutorialGuideButtonAcknowledged = false);
    } else {
      _tutorialGuideButtonAcknowledged = false;
    }
    _syncTutorialGuideButtonPulse(_tutorialStatus);

    final userId = context.read<AuthProvider>().user?.id.trim() ?? '';
    if (userId.isEmpty) return;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tutorialGuideButtonAcknowledgedPrefsKey(userId));
  }

  void _syncTutorialGuideButtonPulse(TutorialStatus? status) {
    final shouldPulse =
        _isTutorialGuideButtonUnlocked(status) &&
        !_tutorialGuideDockVisible &&
        !_tutorialGuideButtonAcknowledged;
    if (shouldPulse) {
      if (!_tutorialGuideButtonPulseController.isAnimating) {
        _tutorialGuideButtonPulseController.repeat();
      }
      return;
    }
    if (_tutorialGuideButtonPulseController.isAnimating) {
      _tutorialGuideButtonPulseController.stop();
    }
    if (_tutorialGuideButtonPulseController.value != 0.0) {
      _tutorialGuideButtonPulseController.value = 0.0;
    }
  }

  String? _primaryTrackedTutorialQuestCarouselItemId(
    QuestLogProvider questLog,
  ) {
    final trackedQuestIds = questLog.trackedQuestIds
        .map((id) => id.trim())
        .where((id) => id.isNotEmpty)
        .toSet();
    if (trackedQuestIds.isEmpty) {
      return null;
    }

    for (final quest in questLog.quests) {
      final questId = quest.id.trim();
      if (questId.isEmpty ||
          !quest.isTutorial ||
          !quest.isAccepted ||
          !trackedQuestIds.contains(questId)) {
        continue;
      }
      return 'quest:$questId';
    }
    return null;
  }

  void _openTrackedQuestsForObjectiveUpdates(QuestLogProvider questLog) {
    final signatures = <String, String>{};
    for (final quest in questLog.quests) {
      if (!questLog.trackedQuestIds.contains(quest.id)) continue;
      if (!quest.isAccepted) continue;
      signatures[quest.id] = _trackedQuestObjectiveSignature(quest);
    }
    if (!_hasTrackedQuestObjectiveSnapshot) {
      _lastTrackedQuestObjectiveSignatures = signatures;
      _hasTrackedQuestObjectiveSnapshot = true;
      return;
    }
    var hasObjectiveUpdate = false;
    for (final entry in signatures.entries) {
      final previous = _lastTrackedQuestObjectiveSignatures[entry.key];
      if (previous != null && previous != entry.value) {
        hasObjectiveUpdate = true;
        break;
      }
    }
    _lastTrackedQuestObjectiveSignatures = signatures;
    if (!hasObjectiveUpdate) return;
    _trackedQuestsController.open();
  }

  void _openTrackedQuestsForTutorialQuestAwards(QuestLogProvider questLog) {
    final acceptedTrackedTutorialQuestIds = <String>{};
    for (final quest in questLog.quests) {
      final questId = quest.id.trim();
      if (questId.isEmpty ||
          !quest.isTutorial ||
          !quest.isAccepted ||
          !questLog.trackedQuestIds.contains(questId)) {
        continue;
      }
      acceptedTrackedTutorialQuestIds.add(questId);
    }

    if (!_hasAcceptedTrackedTutorialQuestSnapshot) {
      _lastAcceptedTrackedTutorialQuestIds = acceptedTrackedTutorialQuestIds;
      _hasAcceptedTrackedTutorialQuestSnapshot = true;
      return;
    }

    final newlyAcceptedTrackedTutorialQuestIds = acceptedTrackedTutorialQuestIds
        .difference(_lastAcceptedTrackedTutorialQuestIds);
    _lastAcceptedTrackedTutorialQuestIds = acceptedTrackedTutorialQuestIds;
    if (newlyAcceptedTrackedTutorialQuestIds.isEmpty) return;

    for (final quest in questLog.quests) {
      final questId = quest.id.trim();
      if (!newlyAcceptedTrackedTutorialQuestIds.contains(questId)) continue;
      _trackedQuestsController.open(itemId: 'quest:$questId');
      return;
    }
    _trackedQuestsController.open();
  }

  void _openTrackedQuestsForTutorialObjectiveUpdates(TutorialStatus? status) {
    final signature = _tutorialTrackedObjectiveSignature(status);
    if (!_hasTutorialTrackedObjectiveSnapshot) {
      _lastTutorialTrackedObjectiveSignature = signature;
      _hasTutorialTrackedObjectiveSnapshot = true;
      if (signature.isNotEmpty) {
        _trackedQuestsController.open();
      }
      return;
    }
    if (signature == _lastTutorialTrackedObjectiveSignature) {
      return;
    }
    _lastTutorialTrackedObjectiveSignature = signature;
    if (signature.isEmpty) {
      return;
    }
    _trackedQuestsController.open();
  }

  String _trackedQuestObjectiveSignature(Quest quest) {
    final node = quest.currentNode;
    final polygonHash = _hashQuestPolygons([
      node?.polygon ?? const <QuestNodePolygonPoint>[],
    ]);
    return [
      quest.id,
      quest.readyToTurnIn ? 'turnin' : 'active',
      node?.id ?? '',
      node?.objectiveText ?? '',
      node?.pointOfInterest?.id ?? '',
      node?.scenarioId ?? '',
      node?.expositionId ?? '',
      node?.monsterEncounterId ?? '',
      node?.monsterId ?? '',
      node?.challengeId ?? '',
      polygonHash.toString(),
    ].join('|');
  }

  String _tutorialTrackedObjectiveSignature(TutorialStatus? status) {
    if (status == null || status.isCompleted) {
      return '';
    }
    if (status.hasActiveScenario) {
      return [
        'scenario',
        status.scenarioId?.trim() ?? '',
        status.resolvedScenarioObjectiveCopy,
      ].join('|');
    }
    if (status.hasActiveMonsterEncounter) {
      return [
        'monster',
        status.monsterEncounterId?.trim() ?? '',
        status.resolvedMonsterObjectiveCopy,
      ].join('|');
    }
    return '';
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
    ids.addAll(_currentQuestTurnInPoiIds(questLog));
    return ids;
  }

  String? _poiThumbnailSourceUrl(PointOfInterest poi) {
    final thumb = poi.thumbnailUrl;
    if (thumb != null && thumb.isNotEmpty) return thumb;
    final image = poi.imageURL;
    if (image != null && image.isNotEmpty) return image;
    return null;
  }

  String _poiMarkerImageId(
    PointOfInterest poi, {
    required bool hasQuestMarker,
    required bool hasMainStoryAccent,
  }) {
    final categoryId = poi.markerCategory.wireValue;
    if (hasQuestMarker) {
      return hasMainStoryAccent
          ? 'poi_category_${categoryId}_main_story'
          : 'poi_category_${categoryId}_activity';
    }
    return 'poi_category_$categoryId';
  }

  Uint8List? _peekPoiMarkerImage(
    PointOfInterest poi, {
    required bool hasQuestMarker,
    required bool hasMainStoryAccent,
  }) {
    if (hasQuestMarker) {
      return hasMainStoryAccent
          ? peekPoiCategoryThumbnailWithMainStoryMarker(poi.markerCategory)
          : peekPoiCategoryThumbnailWithQuestMarker(poi.markerCategory);
    }
    return peekPoiCategoryThumbnail(poi.markerCategory);
  }

  Future<Uint8List?> _loadPoiMarkerImage(
    PointOfInterest poi, {
    required bool hasQuestMarker,
    required bool hasMainStoryAccent,
  }) {
    if (hasQuestMarker) {
      return hasMainStoryAccent
          ? loadPoiCategoryThumbnailWithMainStoryMarker(poi.markerCategory)
          : loadPoiCategoryThumbnailWithQuestMarker(poi.markerCategory);
    }
    return loadPoiCategoryThumbnail(poi.markerCategory);
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
    for (final exposition in _expositions) {
      addPoiId(exposition.pointOfInterestId);
      addByCoordinate(exposition.latitude, exposition.longitude);
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

  Set<String> _buildPoiIdsWithQuestMarkerContent() {
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

    for (final exposition in _expositions) {
      addPoiId(exposition.pointOfInterestId);
      addByCoordinate(exposition.latitude, exposition.longitude);
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

  bool _poiHasQuestMarkerContent(
    PointOfInterest poi, {
    required bool isQuestCurrent,
    Set<String>? questMarkerPoiIds,
  }) {
    if (poi.hasAvailableQuest || isQuestCurrent) return true;
    final contentIds =
        questMarkerPoiIds ?? _buildPoiIdsWithQuestMarkerContent();
    return contentIds.contains(poi.id);
  }

  Iterable<Quest> _trackedAcceptedQuests(QuestLogProvider questLog) sync* {
    final trackedIds = questLog.trackedQuestIds
        .map((id) => id.trim())
        .where((id) => id.isNotEmpty)
        .toSet();
    if (trackedIds.isEmpty) return;
    for (final quest in questLog.quests) {
      if (!quest.isAccepted || !trackedIds.contains(quest.id)) continue;
      yield quest;
    }
  }

  Set<String> _trackedQuestTurnInCharacterIds(QuestLogProvider questLog) {
    final ids = <String>{};
    for (final quest in _trackedAcceptedQuests(questLog)) {
      if (quest.readyToTurnIn && quest.questGiverCharacterId != null) {
        ids.add(quest.questGiverCharacterId!);
      }
      final currentNode = quest.currentNode;
      final fetchCharacterId = currentNode?.fetchCharacterId?.trim() ?? '';
      if (!quest.readyToTurnIn && fetchCharacterId.isNotEmpty) {
        ids.add(fetchCharacterId);
      }
    }
    return ids;
  }

  Set<String> _trackedQuestTurnInPoiIds(QuestLogProvider questLog) {
    final ids = <String>{};
    final turnInCharacterIds = _trackedQuestTurnInCharacterIds(questLog);
    for (final characterId in turnInCharacterIds) {
      final character = _characterById(characterId);
      if (character == null) continue;
      final poi = _poiForCharacter(character);
      final poiId = poi?.id.trim() ?? '';
      if (poiId.isNotEmpty) {
        ids.add(poiId);
      }
    }
    for (final quest in _trackedAcceptedQuests(questLog)) {
      if (!quest.readyToTurnIn) continue;
      final poi = _questReceiverPoiForQuest(quest);
      final poiId = poi?.id.trim() ?? '';
      if (poiId.isNotEmpty) {
        ids.add(poiId);
      }
    }
    return ids;
  }

  Set<String> _trackedQuestPoiIdsForPulse(QuestLogProvider questLog) {
    final ids = <String>{};
    for (final quest in _trackedAcceptedQuests(questLog)) {
      if (quest.readyToTurnIn) continue;
      final poiId = quest.currentNode?.pointOfInterest?.id.trim() ?? '';
      if (poiId.isNotEmpty) {
        ids.add(poiId);
      }
    }
    ids.addAll(_trackedQuestTurnInPoiIds(questLog));
    return ids;
  }

  bool _shouldPulseQuestPoi(
    PointOfInterest poi, {
    required bool isQuestCurrent,
    Set<String>? trackedQuestPoiIds,
  }) {
    if (poi.hasAvailableQuest || isQuestCurrent) return true;
    final trackedIds =
        trackedQuestPoiIds ??
        _trackedQuestPoiIdsForPulse(context.read<QuestLogProvider>());
    return trackedIds.contains(poi.id);
  }

  bool _shouldPulseQuestCharacter(
    Character character, {
    bool? isCurrentQuestTarget,
    bool? isTrackedQuestTarget,
  }) {
    if (character.hasAvailableQuest) return true;
    final currentQuestTarget =
        isCurrentQuestTarget ?? _isCurrentQuestTurnInCharacter(character.id);
    if (currentQuestTarget) return true;
    return isTrackedQuestTarget ?? _isTrackedQuestTurnInCharacter(character.id);
  }

  List<List<QuestNodePolygonPoint>> _trackedQuestCurrentNodePolygons(
    QuestLogProvider questLog,
  ) {
    return _trackedAcceptedQuests(questLog)
        .where((quest) => !quest.readyToTurnIn)
        .map(
          (quest) =>
              quest.currentNode?.polygon ?? const <QuestNodePolygonPoint>[],
        )
        .where((polygon) => polygon.isNotEmpty)
        .map((polygon) => List<QuestNodePolygonPoint>.from(polygon))
        .toList();
  }

  Set<String> _currentQuestTurnInCharacterIds(QuestLogProvider questLog) {
    final ids = <String>{};
    for (final quest in questLog.quests) {
      if (quest.readyToTurnIn && quest.questGiverCharacterId != null) {
        ids.add(quest.questGiverCharacterId!);
      }
      final currentNode = quest.currentNode;
      final fetchCharacterId = currentNode?.fetchCharacterId?.trim() ?? '';
      if (quest.isAccepted &&
          !quest.readyToTurnIn &&
          fetchCharacterId.isNotEmpty) {
        ids.add(fetchCharacterId);
      }
    }
    return ids;
  }

  Set<String> _currentMainStoryQuestPoiIds(QuestLogProvider questLog) {
    return questLog.quests
        .where((q) => q.isAccepted && q.isMainStory && !q.readyToTurnIn)
        .map((q) => q.currentNode?.pointOfInterest?.id ?? '')
        .where((id) => id.isNotEmpty)
        .toSet();
  }

  Set<String> _currentQuestTurnInPoiIds(QuestLogProvider questLog) {
    final ids = <String>{};
    final turnInCharacterIds = _currentQuestTurnInCharacterIds(questLog);
    for (final characterId in turnInCharacterIds) {
      final character = _characterById(characterId);
      if (character == null) continue;
      final poi = _poiForCharacter(character);
      final poiId = poi?.id.trim() ?? '';
      if (poiId.isNotEmpty) {
        ids.add(poiId);
      }
    }
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      if (!quest.readyToTurnIn) continue;
      final poi = _questReceiverPoiForQuest(quest);
      final poiId = poi?.id.trim() ?? '';
      if (poiId.isNotEmpty) {
        ids.add(poiId);
      }
    }
    return ids;
  }

  Set<String> _currentMainStoryTurnInCharacterIds(QuestLogProvider questLog) {
    return questLog.quests
        .where(
          (q) =>
              q.isMainStory &&
              q.readyToTurnIn &&
              q.questGiverCharacterId != null,
        )
        .map((q) => q.questGiverCharacterId!)
        .toSet();
  }

  Set<String> _currentMainStoryTurnInPoiIds(QuestLogProvider questLog) {
    final ids = <String>{};
    for (final quest in questLog.quests) {
      if (!quest.isMainStory || !quest.readyToTurnIn) continue;
      final poi = _questReceiverPoiForQuest(quest);
      final poiId = poi?.id.trim() ?? '';
      if (poiId.isNotEmpty) {
        ids.add(poiId);
      }
    }
    return ids;
  }

  bool _poiHasMainStoryAccent(PointOfInterest poi, QuestLogProvider questLog) {
    return poi.hasAvailableMainStoryQuest ||
        _currentMainStoryQuestPoiIds(questLog).contains(poi.id) ||
        _currentMainStoryTurnInPoiIds(questLog).contains(poi.id);
  }

  bool _characterHasMainStoryAccent(
    Character character,
    QuestLogProvider questLog,
  ) {
    return character.hasAvailableMainStoryQuest ||
        _currentMainStoryTurnInCharacterIds(questLog).contains(character.id);
  }

  PointOfInterest? _poiForCharacter(Character character) {
    final poiId = character.pointOfInterestId?.trim() ?? '';
    final knownPois = _knownPois().toList(growable: false);
    if (poiId.isNotEmpty) {
      for (final poi in knownPois) {
        if (poi.id == poiId) return poi;
      }
    }
    for (final poi in knownPois) {
      if (poi.characters.any((candidate) => candidate.id == character.id)) {
        return poi;
      }
    }
    final poiLat = character.pointOfInterestLat;
    final poiLng = character.pointOfInterestLng;
    if (poiLat != null && poiLng != null) {
      for (final poi in knownPois) {
        final lat = double.tryParse(poi.lat);
        final lng = double.tryParse(poi.lng);
        if (lat == null || lng == null) continue;
        if ((lat - poiLat).abs() < 0.0001 && (lng - poiLng).abs() < 0.0001) {
          return poi;
        }
      }
    }
    return null;
  }

  PointOfInterest? _syntheticPoiForCharacterLead(Character character) {
    final actualPoi = _poiForCharacter(character);
    if (actualPoi != null) return actualPoi;

    double? lat;
    double? lng;
    if (character.pointOfInterestLat != null &&
        character.pointOfInterestLng != null &&
        _isValidCharacterCoordinate(
          character.pointOfInterestLat!,
          character.pointOfInterestLng!,
        )) {
      lat = character.pointOfInterestLat;
      lng = character.pointOfInterestLng;
    } else {
      for (final location in character.locations) {
        if (_isValidCharacterCoordinate(
          location.latitude,
          location.longitude,
        )) {
          lat = location.latitude;
          lng = location.longitude;
          break;
        }
      }
    }
    if (lat == null || lng == null) return null;
    return PointOfInterest(
      id: 'main_story_character_${character.id}',
      name: 'their current location',
      lat: lat.toString(),
      lng: lng.toString(),
      characters: [character],
      hasAvailableMainStoryQuest: true,
    );
  }

  _MainStoryLead? _currentZoneMainStoryLead(
    QuestLogProvider questLog, {
    bool allowGlobalFallback = true,
    bool includeKnownPins = true,
    bool requireFreshZonePins = false,
  }) {
    if (requireFreshZonePins && _questAvailabilityLeadSyncPending) {
      return null;
    }
    final selectedZone = context.read<ZoneProvider>().selectedZone;
    final location = context.read<LocationProvider>().location;
    final userLat = location?.latitude;
    final userLng = location?.longitude;
    final poiCandidates = includeKnownPins ? _knownPois() : _pois;
    final characterCandidates = includeKnownPins
        ? _knownCharacters()
        : _characters;
    PointOfInterest? bestPoi;
    Character? bestCharacter;
    double? bestDistance;
    PointOfInterest? bestGlobalPoi;
    Character? bestGlobalCharacter;
    double? bestGlobalDistance;

    for (final poi in poiCandidates) {
      final poiLat = double.tryParse(poi.lat);
      final poiLng = double.tryParse(poi.lng);
      Character? featuredCharacter;
      for (final character in poi.characters) {
        if (character.hasAvailableMainStoryQuest) {
          featuredCharacter = character;
          break;
        }
      }
      final hasAvailableMainStoryLead =
          poi.hasAvailableMainStoryQuest || featuredCharacter != null;
      if (poiLat == null || poiLng == null || !hasAvailableMainStoryLead) {
        continue;
      }
      final distance = (userLat == null || userLng == null)
          ? 0.0
          : _haversineDistanceMeters(userLat, userLng, poiLat, poiLng);
      if (bestGlobalPoi == null ||
          bestGlobalDistance == null ||
          distance < bestGlobalDistance) {
        bestGlobalPoi = poi;
        bestGlobalCharacter = featuredCharacter;
        bestGlobalDistance = distance;
      }
      if (selectedZone == null ||
          !_isPointInZone(selectedZone, poiLat, poiLng)) {
        continue;
      }
      if (bestPoi == null || bestDistance == null || distance < bestDistance) {
        bestPoi = poi;
        bestCharacter = featuredCharacter;
        bestDistance = distance;
      }
    }

    for (final character in characterCandidates) {
      if (!character.hasAvailableMainStoryQuest) continue;
      final poi = _syntheticPoiForCharacterLead(character);
      if (poi == null) continue;
      final poiLat = double.tryParse(poi.lat);
      final poiLng = double.tryParse(poi.lng);
      if (poiLat == null || poiLng == null) continue;
      final distance = (userLat == null || userLng == null)
          ? 0.0
          : _haversineDistanceMeters(userLat, userLng, poiLat, poiLng);
      if (bestGlobalPoi == null ||
          bestGlobalDistance == null ||
          distance < bestGlobalDistance) {
        bestGlobalPoi = poi;
        bestGlobalCharacter = character;
        bestGlobalDistance = distance;
      }
      if (selectedZone == null ||
          !_isPointInZone(selectedZone, poiLat, poiLng)) {
        continue;
      }
      if (bestPoi == null || bestDistance == null || distance < bestDistance) {
        bestPoi = poi;
        bestCharacter = character;
        bestDistance = distance;
      }
    }

    if (allowGlobalFallback) {
      bestPoi ??= bestGlobalPoi;
      bestCharacter ??= bestGlobalCharacter;
      bestDistance ??= bestGlobalDistance;
    }
    if (bestPoi == null) return null;
    return _MainStoryLead(
      poi: bestPoi,
      character: bestCharacter,
      distanceMeters: bestDistance,
    );
  }

  double _haversineDistanceMeters(
    double lat1,
    double lng1,
    double lat2,
    double lng2,
  ) {
    const earthRadiusMeters = 6371000.0;
    final dLat = _degreesToRadians(lat2 - lat1);
    final dLng = _degreesToRadians(lng2 - lng1);
    final a =
        math.sin(dLat / 2) * math.sin(dLat / 2) +
        math.cos(_degreesToRadians(lat1)) *
            math.cos(_degreesToRadians(lat2)) *
            math.sin(dLng / 2) *
            math.sin(dLng / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return earthRadiusMeters * c;
  }

  double _degreesToRadians(double degrees) => degrees * math.pi / 180.0;

  Set<String> _currentQuestScenarioIds() {
    final questLog = context.read<QuestLogProvider>();
    return questLog.quests
        .where((q) => q.isAccepted)
        .map((q) => q.currentNode?.scenarioId?.trim() ?? '')
        .where((id) => id.isNotEmpty)
        .toSet();
  }

  Set<String> _currentQuestExpositionIds() {
    final questLog = context.read<QuestLogProvider>();
    return questLog.quests
        .where((q) => q.isAccepted)
        .map((q) => q.currentNode?.expositionId?.trim() ?? '')
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
    _playerPresencePulseTimer?.cancel();
    _playerPresencePulseTimer = null;
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
      _resourceSymbols = [];
      _resourceCircles = [];
      _resourceSymbolById.clear();
      _resourceCircleById.clear();
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
      _basePlacementPreviewSymbol = null;
      _basePlacementPreviewBytes = null;
      _challengePolygonLines = [];
      _challengePolygonFills = [];
      _challengePolygonLineById.clear();
      _challengePolygonFillById.clear();
      _scenarioMysteryThumbnailBytes = null;
      _scenarioMysteryThumbnailAdded = false;
      _monsterMysteryThumbnailBytesByType.clear();
      _monsterMysteryThumbnailTypesAdded.clear();
      _challengeMysteryThumbnailBytes = null;
      _challengeMysteryThumbnailAdded = false;
      _zoneLines = [];
      _zoneFills = [];
      _zoneFillById.clear();
      _mapImageIds.clear();
      _renderedSelectedZoneId = null;
      _renderedTreasureChestZoneId = null;
      _questLines = [];
      _characterSymbolsById.clear();
      _playerPresenceSymbol = null;
      _playerPresenceAuraCircle = null;
      _playerPresencePulseCircle = null;
      _playerPresenceConeFill = null;
      _playerPresenceConeLine = null;
      _playerPresenceRefreshGeneration = 0;
      _lastPlayerPresenceLatLng = null;
      _lastResolvedPlayerHeading = null;
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
        await _refreshPlayerPresence();
        await _addPoiMarkers();
        if (_scenarioVisibilityRefreshPending) {
          _scenarioVisibilityRefreshPending = false;
          await _refreshScenarioSymbols();
          await _refreshExpositionSymbols();
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
    if (_isPlacingBase) {
      unawaited(_hideMarkersForBasePlacement());
      unawaited(_refreshBasePlacementPreview());
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

  void _centerOnUserLocation() {
    final c = _mapController;
    final loc = context.read<LocationProvider>().location;
    if (c == null || loc == null) return;
    final lat = loc.latitude;
    final lng = loc.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180) {
      return;
    }
    final currentCamera = c.cameraPosition;
    try {
      c.animateCamera(
        CameraUpdate.newCameraPosition(
          CameraPosition(
            target: LatLng(lat, lng),
            zoom: currentCamera?.zoom ?? 15.5,
            tilt: currentCamera?.tilt ?? 0.0,
            bearing: currentCamera?.bearing ?? 0.0,
          ),
        ),
        duration: const Duration(milliseconds: 450),
      );
    } catch (_) {}
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

  bool _playerPresenceUsesPortrait(User? user) {
    if (user == null || !user.hasCustomizedPortrait) return false;
    final portraitUrl = user.profilePictureUrl.trim();
    if (portraitUrl.isEmpty) return false;
    final normalized = portraitUrl.toLowerCase();
    return !normalized.contains('character-undiscovered') &&
        !normalized.contains('loading-image');
  }

  double? _normalizedHeading(double? heading) {
    if (heading == null || !heading.isFinite) return null;
    final normalized = heading % 360;
    return normalized < 0 ? normalized + 360 : normalized;
  }

  double _radiansToDegrees(double radians) => radians * 180.0 / math.pi;

  double _bearingDegrees(double lat1, double lng1, double lat2, double lng2) {
    final phi1 = _degreesToRadians(lat1);
    final phi2 = _degreesToRadians(lat2);
    final lambda = _degreesToRadians(lng2 - lng1);
    final y = math.sin(lambda) * math.cos(phi2);
    final x =
        math.cos(phi1) * math.sin(phi2) -
        math.sin(phi1) * math.cos(phi2) * math.cos(lambda);
    return _normalizedHeading(_radiansToDegrees(math.atan2(y, x))) ?? 0.0;
  }

  LatLng _offsetLatLng(
    LatLng origin,
    double bearingDegrees,
    double distanceMeters,
  ) {
    const earthRadiusMeters = 6378137.0;
    final angularDistance = distanceMeters / earthRadiusMeters;
    final bearing = _degreesToRadians(bearingDegrees);
    final lat1 = _degreesToRadians(origin.latitude);
    final lng1 = _degreesToRadians(origin.longitude);

    final sinLat1 = math.sin(lat1);
    final cosLat1 = math.cos(lat1);
    final sinAngular = math.sin(angularDistance);
    final cosAngular = math.cos(angularDistance);

    final lat2 = math.asin(
      sinLat1 * cosAngular + cosLat1 * sinAngular * math.cos(bearing),
    );
    final lng2 =
        lng1 +
        math.atan2(
          math.sin(bearing) * sinAngular * cosLat1,
          cosAngular - sinLat1 * math.sin(lat2),
        );

    return LatLng(_radiansToDegrees(lat2), _radiansToDegrees(lng2));
  }

  double? _resolvePlayerHeading(AppLocation location) {
    final speed = location.speed ?? 0.0;
    final directHeading = _normalizedHeading(location.heading);
    if (directHeading != null &&
        speed >= _playerPresenceHeadingSpeedThreshold) {
      return directHeading;
    }

    final previous = _lastPlayerPresenceLatLng;
    if (previous != null) {
      final distance = _haversineDistanceMeters(
        previous.latitude,
        previous.longitude,
        location.latitude,
        location.longitude,
      );
      if (distance >= _playerPresenceBearingFallbackDistanceMeters) {
        return _bearingDegrees(
          previous.latitude,
          previous.longitude,
          location.latitude,
          location.longitude,
        );
      }
    }

    return _lastResolvedPlayerHeading;
  }

  List<LatLng> _buildPlayerFollowConeRing(
    LatLng origin,
    double headingDegrees, {
    required double speedMetersPerSecond,
  }) {
    final speedFactor = (speedMetersPerSecond / 5.0).clamp(0.0, 1.0).toDouble();
    final lengthMeters = _lerpDouble(
      _playerPresenceConeMinLengthMeters,
      _playerPresenceConeMaxLengthMeters,
      speedFactor,
    );
    final halfAngle = _lerpDouble(
      _playerPresenceConeBaseHalfAngleDegrees,
      18.0,
      speedFactor,
    );

    final ring = <LatLng>[origin];
    const arcSteps = 6;
    for (var step = 0; step <= arcSteps; step++) {
      final t = step / arcSteps;
      final angle = headingDegrees - halfAngle + (halfAngle * 2 * t);
      ring.add(_offsetLatLng(origin, angle, lengthMeters));
    }
    ring.add(origin);
    return ring;
  }

  String _playerPresenceImageIdFor(User? user) {
    final portraitUrl = user?.profilePictureUrl.trim() ?? '';
    final signature = _playerPresenceUsesPortrait(user)
        ? portraitUrl
        : 'fallback';
    return 'player_presence_marker_${Object.hash(user?.id ?? '', signature).abs()}';
  }

  CircleOptions _playerPresenceAuraOptions(
    LatLng geometry, {
    required double accuracyMeters,
  }) {
    final radius = (20.0 + (accuracyMeters / 12.0)).clamp(20.0, 30.0);
    return CircleOptions(
      geometry: geometry,
      circleRadius: radius.toDouble(),
      circleColor: _playerPresenceAuraColor,
      circleBlur: 0.36,
      circleOpacity: 0.18,
      circleStrokeWidth: 1.4,
      circleStrokeColor: _playerPresencePulseStrokeColor,
      circleStrokeOpacity: 0.16,
    );
  }

  CircleOptions _playerPresencePulseOptions(LatLng geometry) {
    final cycleMs = _playerPresencePulseCycle.inMilliseconds;
    final progress =
        (DateTime.now().millisecondsSinceEpoch % cycleMs) / cycleMs;
    final eased = Curves.easeOutCubic.transform(progress);
    final fade = 1.0 - Curves.easeInCubic.transform(progress);
    return CircleOptions(
      geometry: geometry,
      circleRadius: _lerpDouble(24.0, 48.0, eased),
      circleColor: _playerPresencePulseColor,
      circleBlur: 0.26,
      circleOpacity: 0.12 * fade,
      circleStrokeWidth: _lerpDouble(3.6, 0.9, eased),
      circleStrokeColor: _playerPresencePulseStrokeColor,
      circleStrokeOpacity: 0.52 * fade,
    );
  }

  Future<void> _clearPlayerPresenceOverlays({bool resetTracking = true}) async {
    _playerPresenceRefreshGeneration++;
    _playerPresencePulseTimer?.cancel();
    _playerPresencePulseTimer = null;

    final c = _mapController;
    final symbol = _playerPresenceSymbol;
    final aura = _playerPresenceAuraCircle;
    final pulse = _playerPresencePulseCircle;
    final coneFill = _playerPresenceConeFill;
    final coneLine = _playerPresenceConeLine;

    _playerPresenceSymbol = null;
    _playerPresenceAuraCircle = null;
    _playerPresencePulseCircle = null;
    _playerPresenceConeFill = null;
    _playerPresenceConeLine = null;
    if (resetTracking) {
      _lastPlayerPresenceLatLng = null;
      _lastResolvedPlayerHeading = null;
    }

    if (c == null) return;

    if (symbol != null) {
      try {
        await c.removeSymbols([symbol]);
      } catch (_) {}
    }
    for (final circle in [aura, pulse]) {
      if (circle == null) continue;
      try {
        await c.removeCircle(circle);
      } catch (_) {}
    }
    if (coneFill != null) {
      try {
        await c.removeFills([coneFill]);
      } catch (_) {}
    }
    if (coneLine != null) {
      try {
        await c.removeLines([coneLine]);
      } catch (_) {}
    }
  }

  void _ensurePlayerPresencePulseTimer() {
    if (_playerPresencePulseTimer != null) return;
    _playerPresencePulseTimer = Timer.periodic(
      _playerPresencePulseFrameDelay,
      (_) => unawaited(_updatePlayerPresencePulseFrame()),
    );
  }

  Future<void> _updatePlayerPresencePulseFrame() async {
    final c = _mapController;
    final pulse = _playerPresencePulseCircle;
    final geometry = _lastPlayerPresenceLatLng;
    if (c == null || pulse == null || geometry == null || !_styleLoaded) {
      return;
    }
    try {
      await c.updateCircle(pulse, _playerPresencePulseOptions(geometry));
    } catch (_) {}
  }

  Future<void> _refreshPlayerPresence() async {
    final generation = ++_playerPresenceRefreshGeneration;
    final c = _mapController;
    final location = context.read<LocationProvider>().location;
    final auth = context.read<AuthProvider>();

    if (c == null ||
        !_styleLoaded ||
        _mapLoadFailed ||
        location == null ||
        auth.loading ||
        !auth.isAuthenticated) {
      await _clearPlayerPresenceOverlays();
      return;
    }

    final lat = location.latitude;
    final lng = location.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180) {
      await _clearPlayerPresenceOverlays();
      return;
    }

    final authUser = auth.user;
    final usePortrait = _playerPresenceUsesPortrait(authUser);
    final portraitUrl = authUser?.profilePictureUrl.trim();
    final imageId = _playerPresenceImageIdFor(authUser);
    final imageBytes =
        peekPlayerPresenceMarker(portraitUrl, usePortrait: usePortrait) ??
        await loadPlayerPresenceMarker(portraitUrl, usePortrait: usePortrait);
    if (!mounted || generation != _playerPresenceRefreshGeneration) return;
    if (imageBytes == null) {
      await _clearPlayerPresenceOverlays(resetTracking: false);
      return;
    }

    await _ensureMapImage(c, imageId, imageBytes);
    if (!mounted || generation != _playerPresenceRefreshGeneration) return;

    final geometry = LatLng(lat, lng);
    final heading = _resolvePlayerHeading(location);

    final symbolOptions = SymbolOptions(
      geometry: geometry,
      iconImage: imageId,
      iconSize: _playerPresenceMarkerIconSize,
      iconOpacity: 1.0,
      iconAnchor: 'bottom',
      iconHaloColor: _transparentMapHaloColor,
      iconHaloWidth: 0.0,
      zIndex: 99,
    );
    if (_playerPresenceSymbol == null) {
      try {
        _playerPresenceSymbol = await c.addSymbol(symbolOptions, const {
          'type': 'playerPresence',
        });
      } catch (_) {}
    } else {
      try {
        await c.updateSymbol(_playerPresenceSymbol!, symbolOptions);
      } catch (_) {
        _playerPresenceSymbol = null;
      }
      if (_playerPresenceSymbol == null) {
        try {
          _playerPresenceSymbol = await c.addSymbol(symbolOptions, const {
            'type': 'playerPresence',
          });
        } catch (_) {}
      }
    }

    final auraOptions = _playerPresenceAuraOptions(
      geometry,
      accuracyMeters: location.accuracy,
    );
    if (_playerPresenceAuraCircle != null) {
      try {
        await c.updateCircle(_playerPresenceAuraCircle!, auraOptions);
      } catch (_) {
        _playerPresenceAuraCircle = null;
      }
    }
    if (_playerPresenceAuraCircle == null) {
      try {
        _playerPresenceAuraCircle = await c.addCircle(auraOptions, const {
          'type': 'playerPresenceAura',
        });
      } catch (_) {}
    }

    final pulseOptions = _playerPresencePulseOptions(geometry);
    if (_playerPresencePulseCircle != null) {
      try {
        await c.updateCircle(_playerPresencePulseCircle!, pulseOptions);
      } catch (_) {
        _playerPresencePulseCircle = null;
      }
    }
    if (_playerPresencePulseCircle == null) {
      try {
        _playerPresencePulseCircle = await c.addCircle(pulseOptions, const {
          'type': 'playerPresencePulse',
        });
      } catch (_) {}
    }

    if (heading == null) {
      if (_playerPresenceConeFill != null) {
        try {
          await c.removeFills([_playerPresenceConeFill!]);
        } catch (_) {}
        _playerPresenceConeFill = null;
      }
      if (_playerPresenceConeLine != null) {
        try {
          await c.removeLines([_playerPresenceConeLine!]);
        } catch (_) {}
        _playerPresenceConeLine = null;
      }
    } else {
      final speed = (location.speed ?? 0.0).clamp(0.0, 8.0).toDouble();
      final ring = _buildPlayerFollowConeRing(
        geometry,
        heading,
        speedMetersPerSecond: speed,
      );
      final fillOptions = FillOptions(
        geometry: [ring],
        fillColor: _playerPresenceConeFillColor,
        fillOpacity: 0.16,
        fillOutlineColor: _playerPresenceConeOutlineColor,
      );
      final lineOptions = LineOptions(
        geometry: ring,
        lineColor: _playerPresenceConeOutlineColor,
        lineWidth: 2.0,
        lineOpacity: 0.52,
        lineJoin: 'round',
        lineBlur: 0.1,
      );

      if (_playerPresenceConeFill != null) {
        try {
          await c.updateFill(_playerPresenceConeFill!, fillOptions);
        } catch (_) {
          _playerPresenceConeFill = null;
        }
      }
      if (_playerPresenceConeFill == null) {
        try {
          final fills = await c.addFills(
            [fillOptions],
            [
              {'type': 'playerPresenceCone'},
            ],
          );
          if (fills.isNotEmpty) {
            _playerPresenceConeFill = fills.first;
          }
        } catch (_) {}
      }

      if (_playerPresenceConeLine != null) {
        try {
          await c.updateLine(_playerPresenceConeLine!, lineOptions);
        } catch (_) {
          _playerPresenceConeLine = null;
        }
      }
      if (_playerPresenceConeLine == null) {
        try {
          final lines = await c.addLines(
            [lineOptions],
            [
              {'type': 'playerPresenceCone'},
            ],
          );
          if (lines.isNotEmpty) {
            _playerPresenceConeLine = lines.first;
          }
        } catch (_) {}
      }
    }

    _lastPlayerPresenceLatLng = geometry;
    if (heading != null) {
      _lastResolvedPlayerHeading = heading;
    }
    _ensurePlayerPresencePulseTimer();
  }

  Future<void> _refreshMapContent() async {
    if (_manualRefreshInFlight) {
      return;
    }
    final discoveriesProvider = context.read<DiscoveriesProvider>();
    final activityFeedProvider = context.read<ActivityFeedProvider>();
    final partyProvider = context.read<PartyProvider>();
    setState(() {
      _manualRefreshInFlight = true;
    });
    try {
      _updateSelectedZoneFromLocation();
      _requestQuestLogIfReady(force: true);
      await _loadAll();
      await _loadTreasureChestsForSelectedZone(
        forceRefreshZonePins: true,
        forceRefreshZoneBaseContent: true,
      );
      await _loadBases();
      unawaited(
        _runBestEffortRefresh('discoveries', discoveriesProvider.refresh()),
      );
      unawaited(
        _runBestEffortRefresh('activity feed', activityFeedProvider.refresh()),
      );
      unawaited(_runBestEffortRefresh('party', partyProvider.fetchParty()));
      unawaited(
        _loadTutorialStatus(force: true, preserveCompletedReveal: true),
      );
    } finally {
      if (mounted) {
        setState(() {
          _manualRefreshInFlight = false;
        });
      }
    }
  }

  void _flyToLocation(
    double lat,
    double lng, {
    double zoom = _defaultMapFocusZoom,
  }) {
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
          CameraPosition(target: LatLng(lat, lng), zoom: zoom),
        ),
        duration: const Duration(milliseconds: 500),
      );
    } catch (_) {}
  }

  void _focusQuestPoI(
    PointOfInterest poi, {
    double zoom = _defaultMapFocusZoom,
  }) async {
    final lat = double.tryParse(poi.lat) ?? 0.0;
    final lng = double.tryParse(poi.lng) ?? 0.0;
    final changedZone = _selectZoneForPoiIfDifferent(poi);
    _flyToLocation(lat, lng, zoom: zoom);
    unawaited(_pulsePoi(lat, lng));
    if (changedZone) return;
    final hasDiscovered = context.read<DiscoveriesProvider>().hasDiscovered(
      poi.id,
    );
    await _runWithMapMarkerIsolation(
      _mapMarkerIsolationForPoi(poi),
      () => _showPointOfInterestPanel(poi, hasDiscovered),
    );
  }

  void _previewTrackedQuestPoi(
    PointOfInterest poi, {
    double zoom = _trackedQuestOverlayFocusZoom,
  }) {
    final lat = double.tryParse(poi.lat) ?? 0.0;
    final lng = double.tryParse(poi.lng) ?? 0.0;
    _selectZoneForPoiIfDifferent(poi);
    _flyToLocation(lat, lng, zoom: zoom);
    unawaited(_pulsePoi(lat, lng));
  }

  void _focusTutorialScenarioFromTrackedOverlay({
    double zoom = _trackedQuestOverlayFocusZoom,
  }) {
    final scenarioId = _tutorialStatus?.scenarioId?.trim() ?? '';
    if (scenarioId.isEmpty) return;
    final scenario = _scenarioById(scenarioId);
    if (scenario == null) return;
    _selectZoneByIdIfDifferent(scenario.zoneId);
    _flyToLocation(scenario.latitude, scenario.longitude, zoom: zoom);
    unawaited(_pulsePoi(scenario.latitude, scenario.longitude));
  }

  void _focusTutorialMonsterFromTrackedOverlay({
    double zoom = _trackedQuestOverlayFocusZoom,
  }) {
    final encounterId = _tutorialStatus?.monsterEncounterId?.trim() ?? '';
    if (encounterId.isEmpty) return;
    final encounter = _monsterById(encounterId);
    if (encounter == null) return;
    _selectZoneByIdIfDifferent(encounter.zoneId);
    _flyToLocation(encounter.latitude, encounter.longitude, zoom: zoom);
    unawaited(_pulsePoi(encounter.latitude, encounter.longitude));
  }

  String _tutorialScenarioOverlayTitle(TutorialStatus? status) {
    final scenarioId = status?.scenarioId?.trim() ?? '';
    if (scenarioId.isEmpty) return 'Tutorial Scenario';
    final scenario = _scenarioById(scenarioId);
    if (scenario == null) return 'Tutorial Scenario';
    return _compactSelectionLabel(
      scenario.prompt,
      fallback: 'Tutorial Scenario',
    );
  }

  String _tutorialScenarioOverlayDetail(TutorialStatus? status) {
    if (!(status?.hasActiveScenario ?? false)) {
      return '';
    }
    return 'Scenario';
  }

  String _tutorialMonsterOverlayTitle(TutorialStatus? status) {
    final encounterId = status?.monsterEncounterId?.trim() ?? '';
    if (encounterId.isEmpty) return 'Tutorial Monster';
    final encounter = _monsterById(encounterId);
    final trimmedName = encounter?.name.trim() ?? '';
    if (trimmedName.isNotEmpty) {
      return trimmedName;
    }
    return 'Tutorial Monster';
  }

  String _tutorialMonsterOverlayDetail(TutorialStatus? status) {
    final encounterId = status?.monsterEncounterId?.trim() ?? '';
    if (encounterId.isEmpty) return '';
    final encounter = _monsterById(encounterId);
    if (encounter == null) return 'Monster Encounter';
    return encounter.encounterTypeLabel;
  }

  void _focusQuestTurnIn(Quest quest, {double zoom = _defaultMapFocusZoom}) {
    unawaited(_focusQuestTurnInFlow(quest, zoom: zoom));
  }

  void _returnToPlayerFromTrackedQuestOverlay() {
    final location = context.read<LocationProvider>().location;
    if (location == null) return;
    _updateSelectedZoneFromLocation(force: true);
    _flyToLocation(location.latitude, location.longitude);
  }

  _FeaturedMainStoryPulseTarget? _featuredMainStoryPulseTarget(
    QuestLogProvider questLog,
  ) {
    final lead = _currentZoneMainStoryLead(
      questLog,
      allowGlobalFallback: false,
      includeKnownPins: false,
      requireFreshZonePins: true,
    );
    if (lead == null) return null;

    final character = lead.character;
    if (character != null && _visibleCharacterPoints(character).isNotEmpty) {
      final characterId = character.id.trim();
      if (characterId.isNotEmpty) {
        return _FeaturedMainStoryPulseTarget.character(characterId);
      }
    }

    final poiId = lead.poi.id.trim();
    if (poiId.isEmpty) return null;
    return _FeaturedMainStoryPulseTarget.poi(poiId);
  }

  bool _selectZoneForQuestNodeIfDifferent(QuestNode node) {
    final poi = node.pointOfInterest;
    if (poi != null) {
      return _selectZoneForPoiIfDifferent(poi);
    }

    final fetchCharacter = _questNodeFetchCharacter(node);
    if (fetchCharacter != null) {
      final location = _questNodeFetchCharacterLocation(fetchCharacter);
      if (location != null) {
        return _selectZoneForCoordinatesIfDifferent(
          location.latitude,
          location.longitude,
        );
      }
    }

    final scenarioId = node.scenarioId?.trim() ?? '';
    if (scenarioId.isNotEmpty) {
      final scenario = _scenarioById(scenarioId);
      if (scenario != null) {
        return _selectZoneByIdIfDifferent(scenario.zoneId);
      }
    }

    final expositionId = node.expositionId?.trim() ?? '';
    if (expositionId.isNotEmpty) {
      final exposition = _expositionById(expositionId);
      if (exposition != null) {
        return _selectZoneByIdIfDifferent(exposition.zoneId);
      }
    }

    final encounterId = node.monsterEncounterId?.trim() ?? '';
    if (encounterId.isNotEmpty) {
      final encounter = _monsterById(encounterId);
      if (encounter != null) {
        return _selectZoneByIdIfDifferent(encounter.zoneId);
      }
    }

    final monsterId = node.monsterId?.trim() ?? '';
    if (monsterId.isNotEmpty) {
      final encounter = _monsterEncounterByMemberMonsterId(monsterId);
      if (encounter != null) {
        return _selectZoneByIdIfDifferent(encounter.zoneId);
      }
    }

    final challengeId = node.challengeId?.trim() ?? '';
    if (challengeId.isNotEmpty) {
      final challenge = _challengeById(challengeId);
      if (challenge != null) {
        return _selectZoneByIdIfDifferent(challenge.zoneId);
      }
    }

    return false;
  }

  Character? _questReceiverCharacterForQuest(Quest quest) {
    final embeddedQuestGiver = quest.questGiverCharacter;
    if (embeddedQuestGiver != null && embeddedQuestGiver.id.trim().isNotEmpty) {
      return embeddedQuestGiver;
    }
    final questGiverId = quest.questGiverCharacterId?.trim() ?? '';
    if (questGiverId.isEmpty) return null;
    return _characterById(questGiverId);
  }

  PointOfInterest? _questReceiverPoiForQuest(Quest quest) {
    final embeddedQuestGiverPoi = quest.questGiverPointOfInterest;
    if (embeddedQuestGiverPoi != null && embeddedQuestGiverPoi.id.isNotEmpty) {
      return embeddedQuestGiverPoi;
    }
    final character = _questReceiverCharacterForQuest(quest);
    if (character != null) {
      final actualPoi = _poiForCharacter(character);
      if (actualPoi != null) return actualPoi;
    }

    final questGiverId = quest.questGiverCharacterId?.trim() ?? '';
    if (questGiverId.isEmpty) return null;
    for (final candidate in _knownPois()) {
      if (candidate.characters.any((member) => member.id == questGiverId)) {
        return candidate;
      }
    }
    return null;
  }

  Future<void> _focusQuestTurnInFlow(
    Quest quest, {
    double zoom = _defaultMapFocusZoom,
  }) async {
    final questReceiver = _questReceiverCharacterForQuest(quest);
    final questReceiverPoi = _questReceiverPoiForQuest(quest);

    if (questReceiver != null) {
      final focusLocation = _questNodeFetchCharacterLocation(questReceiver);
      if (focusLocation != null) {
        _selectZoneForCoordinatesIfDifferent(
          focusLocation.latitude,
          focusLocation.longitude,
        );
      } else if (questReceiverPoi != null) {
        _selectZoneForPoiIfDifferent(questReceiverPoi);
      }
      if (focusLocation != null) {
        _flyToLocation(
          focusLocation.latitude,
          focusLocation.longitude,
          zoom: zoom,
        );
        unawaited(_pulsePoi(focusLocation.latitude, focusLocation.longitude));
      } else if (questReceiverPoi != null) {
        final lat = double.tryParse(questReceiverPoi.lat) ?? 0.0;
        final lng = double.tryParse(questReceiverPoi.lng) ?? 0.0;
        _flyToLocation(lat, lng, zoom: zoom);
        unawaited(_pulsePoi(lat, lng));
      }
      return;
    }

    if (questReceiverPoi == null) {
      return;
    }

    final lat = double.tryParse(questReceiverPoi.lat) ?? 0.0;
    final lng = double.tryParse(questReceiverPoi.lng) ?? 0.0;
    _selectZoneForPoiIfDifferent(questReceiverPoi);
    _flyToLocation(lat, lng, zoom: zoom);
    unawaited(_pulsePoi(lat, lng));
  }

  void _focusQuestNode(QuestNode node, {double zoom = _defaultMapFocusZoom}) {
    _selectZoneForQuestNodeIfDifferent(node);
    final poi = node.pointOfInterest;
    if (poi != null) {
      final lat = double.tryParse(poi.lat) ?? 0.0;
      final lng = double.tryParse(poi.lng) ?? 0.0;
      _flyToLocation(lat, lng, zoom: zoom);
      _pulsePoi(lat, lng);
      return;
    }
    final fetchCharacter = _questNodeFetchCharacter(node);
    if (fetchCharacter != null) {
      final focusLocation = _questNodeFetchCharacterLocation(fetchCharacter);
      if (focusLocation != null) {
        _flyToLocation(
          focusLocation.latitude,
          focusLocation.longitude,
          zoom: zoom,
        );
        _pulsePoi(focusLocation.latitude, focusLocation.longitude);
        return;
      }
    }
    final scenarioId = node.scenarioId?.trim() ?? '';
    if (scenarioId.isNotEmpty) {
      final scenario = _scenarioById(scenarioId);
      if (scenario != null) {
        _flyToLocation(scenario.latitude, scenario.longitude, zoom: zoom);
        _pulsePoi(scenario.latitude, scenario.longitude);
        return;
      }
    }
    final expositionId = node.expositionId?.trim() ?? '';
    if (expositionId.isNotEmpty) {
      final exposition = _expositionById(expositionId);
      if (exposition != null) {
        _flyToLocation(exposition.latitude, exposition.longitude, zoom: zoom);
        _pulsePoi(exposition.latitude, exposition.longitude);
        return;
      }
    }
    final encounterId = node.monsterEncounterId?.trim() ?? '';
    if (encounterId.isNotEmpty) {
      final encounter = _monsterById(encounterId);
      if (encounter != null) {
        _flyToLocation(encounter.latitude, encounter.longitude, zoom: zoom);
        _pulsePoi(encounter.latitude, encounter.longitude);
        return;
      }
    }
    final monsterId = node.monsterId?.trim() ?? '';
    if (monsterId.isNotEmpty) {
      final encounter = _monsterEncounterByMemberMonsterId(monsterId);
      if (encounter != null) {
        _flyToLocation(encounter.latitude, encounter.longitude, zoom: zoom);
        _pulsePoi(encounter.latitude, encounter.longitude);
        return;
      }
    }
    final challengeId = node.challengeId?.trim() ?? '';
    if (challengeId.isNotEmpty) {
      final challenge = _challengeById(challengeId);
      if (challenge != null) {
        final anchor = _challengeProximityAnchor(challenge);
        _flyToLocation(anchor.latitude, anchor.longitude, zoom: zoom);
        if (challenge.hasPolygon) {
          _pulsePolygon(challenge.polygonPoints);
        } else {
          _pulsePoi(anchor.latitude, anchor.longitude);
        }
        return;
      }
    }
    if (node.polygon.isNotEmpty) {
      final center = _polygonCenter(node.polygon);
      _flyToLocation(center.latitude, center.longitude, zoom: zoom);
      _pulsePolygon(node.polygon);
    }
  }

  Character? _questNodeFetchCharacter(QuestNode node) {
    final fetchCharacter = node.fetchCharacter;
    if (fetchCharacter != null) {
      return fetchCharacter;
    }
    final fetchCharacterId = node.fetchCharacterId?.trim() ?? '';
    if (fetchCharacterId.isEmpty) {
      return null;
    }
    return _characterById(fetchCharacterId);
  }

  LatLng? _questNodeFetchCharacterLocation(Character character) {
    if (_isValidMapCoordinate(
      character.pointOfInterestLat,
      character.pointOfInterestLng,
    )) {
      return LatLng(
        character.pointOfInterestLat!,
        character.pointOfInterestLng!,
      );
    }
    for (final location in character.locations) {
      if (_isValidMapCoordinate(location.latitude, location.longitude)) {
        return LatLng(location.latitude, location.longitude);
      }
    }
    return null;
  }

  bool _isValidMapCoordinate(double? latitude, double? longitude) {
    if (latitude == null || longitude == null) return false;
    if (!latitude.isFinite || !longitude.isFinite) return false;
    return latitude.abs() <= 90 && longitude.abs() <= 180;
  }

  String _mapMarkerIsolationKey(String type, String id) =>
      '${type.trim()}:${id.trim()}';

  _MapMarkerIsolation _mapMarkerIsolationForPoi(PointOfInterest poi) {
    return _MapMarkerIsolation(
      markerKeys: {_mapMarkerIsolationKey('poi', poi.id)},
    );
  }

  bool _isMapMarkerIsolationVisible(String type, String id) {
    final isolation = _mapMarkerIsolation;
    if (isolation == null) return true;
    return isolation.markerKeys.contains(_mapMarkerIsolationKey(type, id));
  }

  List<Challenge> _visibleStandalonePolygonChallenges() {
    if (_shouldSuppressNormalMapPinsForTutorial) {
      return const <Challenge>[];
    }
    return _challenges
        .where(
          (challenge) =>
              challenge.hasPolygon &&
              !_isChallengeRepresentedByPoi(challenge) &&
              !_usesDedicatedQuestChallengeUi(challenge),
        )
        .toList();
  }

  Future<void> _runWithMapMarkerIsolation(
    _MapMarkerIsolation isolation,
    Future<void> Function() action,
  ) async {
    final token = ++_mapMarkerIsolationToken;
    _mapMarkerIsolation = isolation;
    await _applyMapMarkerIsolation();
    try {
      await action();
    } finally {
      if (_mapMarkerIsolationToken == token) {
        _mapMarkerIsolation = null;
        await _applyMapMarkerIsolation();
      }
    }
  }

  Future<void> _applyMapMarkerIsolationIfNeeded() async {
    if (_mapMarkerIsolation == null) return;
    await _applyMapMarkerIsolation();
  }

  Future<void> _applyMapMarkerIsolation() async {
    final c = _mapController;
    if (c == null || !_styleLoaded || !_markersAdded) return;

    if (_mapMarkerIsolation == null) {
      await _updateNormalPinOpacities(
        c,
        1.0,
        questPoiIds: _currentQuestPoiIdsForFilter(
          context.read<QuestLogProvider>(),
        ),
        discoveries: context.read<DiscoveriesProvider>(),
        mapContentPoiIds: _buildPoiIdsWithMapContent(),
      );
      await _refreshChallengePolygonOverlays(
        c,
        _visibleStandalonePolygonChallenges(),
      );
      _ensureQuestPoiPulseTimer();
      return;
    }

    Future<void> updateSymbolVisibility(
      Symbol symbol, {
      required String type,
      required String id,
      required double visibleOpacity,
    }) async {
      final opacity = _isMapMarkerIsolationVisible(type, id)
          ? visibleOpacity
          : 0.0;
      try {
        await c.updateSymbol(symbol, SymbolOptions(iconOpacity: opacity));
      } catch (_) {}
    }

    Future<void> updateCircleVisibility(
      Circle circle, {
      required String type,
      required String id,
      required double visibleOpacity,
    }) async {
      final opacity = _isMapMarkerIsolationVisible(type, id)
          ? visibleOpacity
          : 0.0;
      try {
        await c.updateCircle(circle, CircleOptions(circleOpacity: opacity));
      } catch (_) {}
    }

    final questPoiIds = _currentQuestPoiIdsForFilter(
      context.read<QuestLogProvider>(),
    );
    final discoveries = context.read<DiscoveriesProvider>();
    final mapContentPoiIds = _buildPoiIdsWithMapContent();

    for (final entry in _poiSymbolById.entries) {
      final poi = _poiById(entry.key);
      if (poi == null) continue;
      final isQuestCurrent = questPoiIds.contains(poi.id);
      final undiscovered = !discoveries.hasDiscovered(poi.id);
      await updateSymbolVisibility(
        entry.value,
        type: 'poi',
        id: poi.id,
        visibleOpacity: _poiMarkerOpacity(
          poi,
          isQuestCurrent: isQuestCurrent,
          undiscovered: undiscovered,
          mapContentPoiIds: mapContentPoiIds,
        ),
      );
    }

    for (final entry in _characterSymbolsById.entries) {
      for (final symbol in entry.value) {
        await updateSymbolVisibility(
          symbol,
          type: 'character',
          id: entry.key,
          visibleOpacity: 1.0,
        );
      }
    }

    for (final entry in _chestSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'chest',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _chestCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'chest',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _healingFountainSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'healingFountain',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _healingFountainCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'healingFountain',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _resourceSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'resource',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _resourceCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'resource',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _baseSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'base',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _baseCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'base',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _scenarioSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'scenario',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _scenarioCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'scenario',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _expositionSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'exposition',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _expositionCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'exposition',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _monsterSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'monster',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _monsterCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'monster',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    for (final entry in _challengeSymbolById.entries) {
      await updateSymbolVisibility(
        entry.value,
        type: 'challenge',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }
    for (final entry in _challengeCircleById.entries) {
      await updateCircleVisibility(
        entry.value,
        type: 'challenge',
        id: entry.key,
        visibleOpacity: 1.0,
      );
    }

    await _refreshChallengePolygonOverlays(c, const <Challenge>[]);
    _ensureQuestPoiPulseTimer();
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

  String _pulseKeyForCoordinates(String namespace, double lat, double lng) {
    return '$namespace:${_coordinateKey(lat, lng) ?? '$lat,$lng'}';
  }

  double _lerpDouble(double start, double end, double t) {
    return start + ((end - start) * t);
  }

  Future<void> _animateFeatheredPulse(
    double lat,
    double lng, {
    required String pulseKey,
    required String coreColor,
    required String mistColor,
    required String ringColor,
    required double coreStartRadius,
    required double coreEndRadius,
    required double mistStartRadius,
    required double mistEndRadius,
    required double ringStartRadius,
    required double ringEndRadius,
    required double maxCoreOpacity,
    required double maxMistOpacity,
    required double maxRingOpacity,
    required double initialStrokeWidth,
    required int steps,
    required Duration frameDelay,
  }) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    if (!_activePulseKeys.add(pulseKey)) return;

    Circle? coreCircle;
    Circle? mistCircle;
    Circle? ringCircle;
    final geometry = LatLng(lat, lng);

    CircleOptions coreOptions(double progress) {
      final eased = Curves.easeOutCubic.transform(progress);
      final fade = 1.0 - Curves.easeInCubic.transform(progress);
      return CircleOptions(
        geometry: geometry,
        circleRadius: _lerpDouble(coreStartRadius, coreEndRadius, eased),
        circleColor: coreColor,
        circleOpacity: maxCoreOpacity * fade,
      );
    }

    CircleOptions mistOptions(double progress) {
      final eased = Curves.easeOutQuart.transform(progress);
      final bloom = (math.sin(progress * math.pi) * 1.08).clamp(0.0, 1.0);
      return CircleOptions(
        geometry: geometry,
        circleRadius: _lerpDouble(mistStartRadius, mistEndRadius, eased),
        circleColor: mistColor,
        circleOpacity: maxMistOpacity * bloom,
      );
    }

    CircleOptions ringOptions(double progress) {
      final eased = Curves.easeOutQuart.transform(progress);
      final fade = 1.0 - Curves.easeOutCubic.transform(progress);
      final strokeWidth = _lerpDouble(
        initialStrokeWidth,
        0.6,
        Curves.easeOut.transform(progress),
      );
      return CircleOptions(
        geometry: geometry,
        circleRadius: _lerpDouble(ringStartRadius, ringEndRadius, eased),
        circleColor: ringColor,
        circleOpacity: maxRingOpacity * fade,
        circleStrokeWidth: strokeWidth,
        circleStrokeColor: ringColor,
      );
    }

    try {
      coreCircle = await c.addCircle(coreOptions(0.0));
      mistCircle = await c.addCircle(mistOptions(0.0));
      ringCircle = await c.addCircle(ringOptions(0.0));
      final animatedCoreCircle = coreCircle;
      final animatedMistCircle = mistCircle;
      final animatedRingCircle = ringCircle;

      for (var step = 1; step <= steps; step++) {
        final progress = step / steps;
        await Future.wait([
          c.updateCircle(animatedCoreCircle, coreOptions(progress)),
          c.updateCircle(animatedMistCircle, mistOptions(progress)),
          c.updateCircle(animatedRingCircle, ringOptions(progress)),
        ]);
        if (step < steps) {
          await Future.delayed(frameDelay);
        }
      }
    } catch (_) {
      // Best-effort pulse only.
    } finally {
      for (final circle in [coreCircle, mistCircle, ringCircle]) {
        if (circle == null) continue;
        try {
          await c.removeCircle(circle);
        } catch (_) {}
      }
      _activePulseKeys.remove(pulseKey);
    }
  }

  Future<void> _pulsePoi(double lat, double lng) async {
    await _animateFeatheredPulse(
      lat,
      lng,
      pulseKey: _pulseKeyForCoordinates('poi', lat, lng),
      coreColor: _questPulseCoreColor,
      mistColor: _questPulseMistColor,
      ringColor: _questPulseRingColor,
      coreStartRadius: 8,
      coreEndRadius: 26,
      mistStartRadius: 18,
      mistEndRadius: 52,
      ringStartRadius: 14,
      ringEndRadius: 58,
      maxCoreOpacity: 0.18,
      maxMistOpacity: 0.16,
      maxRingOpacity: 0.22,
      initialStrokeWidth: 2.6,
      steps: 14,
      frameDelay: const Duration(milliseconds: 55),
    );
  }

  Future<void> _pulseMainStoryLeadFocus(double lat, double lng) async {
    await _animateFeatheredPulse(
      lat,
      lng,
      pulseKey: _pulseKeyForCoordinates('main_story_lead', lat, lng),
      coreColor: _mainStoryPulseCoreColor,
      mistColor: _mainStoryPulseMistColor,
      ringColor: _mainStoryPulseRingColor,
      coreStartRadius: 8,
      coreEndRadius: 26,
      mistStartRadius: 18,
      mistEndRadius: 52,
      ringStartRadius: 14,
      ringEndRadius: 58,
      maxCoreOpacity: 0.18,
      maxMistOpacity: 0.16,
      maxRingOpacity: 0.22,
      initialStrokeWidth: 2.6,
      steps: 14,
      frameDelay: const Duration(milliseconds: 55),
    );
  }

  Future<void> _pulseDiscoveredPoi(double lat, double lng) async {
    await _animateFeatheredPulse(
      lat,
      lng,
      pulseKey: _pulseKeyForCoordinates('discovery', lat, lng),
      coreColor: _discoveryPulseCoreColor,
      mistColor: _discoveryPulseMistColor,
      ringColor: _discoveryPulseRingColor,
      coreStartRadius: 6,
      coreEndRadius: 18,
      mistStartRadius: 12,
      mistEndRadius: 32,
      ringStartRadius: 10,
      ringEndRadius: 36,
      maxCoreOpacity: 0.14,
      maxMistOpacity: 0.12,
      maxRingOpacity: 0.18,
      initialStrokeWidth: 2.0,
      steps: 10,
      frameDelay: const Duration(milliseconds: 50),
    );
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
    if (_loadAllInFlight) return;
    final auth = context.read<AuthProvider>();
    if (auth.loading || !auth.isAuthenticated) {
      return;
    }
    _loadAllInFlight = true;
    final loadGeneration = ++_loadAllGeneration;
    debugPrint('SinglePlayer: _loadAll start');
    final svc = context.read<PoiService>();
    final discoveriesProvider = context.read<DiscoveriesProvider>();
    final zoneProvider = context.read<ZoneProvider>();
    try {
      final restoreDefeatedFuture = _restoreDefeatedMonsterIds();
      final restoreDiscoveredCharactersFuture =
          _restoreDiscoveredCharacterIds();
      final discoveriesFuture = discoveriesProvider.refresh();
      final zonesFuture = svc.getZones();
      final basesFuture = svc.getVisibleBases();
      final zones = await zonesFuture;
      if (!_isCurrentLoadGeneration(loadGeneration)) return;
      debugPrint('SinglePlayer: _loadAll shell ready: zones=${zones.length}');
      zoneProvider.setZones(zones);
      setState(() {
        _zones = zones;
        _pois = [];
        _characters = [];
        _markersAdded = false;
      });
      _updateSelectedZoneFromLocation();
      _requestQuestLogIfReady(force: true);
      _scheduleZoneBaseContentWarmup(immediate: true);
      if (_styleLoaded && _mapController != null) {
        unawaited(_addPoiMarkers());
        unawaited(_addZoneBoundaries());
      }
      unawaited(
        _hydrateWorldShellBackground(
          loadGeneration,
          restoreDefeatedFuture: restoreDefeatedFuture,
          restoreDiscoveredCharactersFuture: restoreDiscoveredCharactersFuture,
          discoveriesFuture: discoveriesFuture,
          basesFuture: basesFuture,
        ),
      );
      debugPrint('SinglePlayer: _loadAll shell staged');
    } catch (e, stackTrace) {
      debugPrint('SinglePlayer: _loadAll error: $e');
      debugPrint('SinglePlayer: _loadAll stack: $stackTrace');
      if (mounted) setState(() {});
    } finally {
      _loadAllInFlight = false;
    }
  }

  Future<void> _loadBases() async {
    try {
      final bases = await context.read<PoiService>().getVisibleBases();
      if (!mounted) return;
      setState(() {
        _bases = bases;
      });
      if (_styleLoaded && _mapController != null && _markersAdded) {
        await _refreshBaseSymbols();
        if (_isPlacingBase) {
          await _hideMarkersForBasePlacement();
          await _refreshBasePlacementPreview();
        }
      }
    } catch (e) {
      debugPrint('SinglePlayer: _loadBases error: $e');
    }
  }

  bool _isCurrentLoadGeneration(int generation) {
    return mounted && generation == _loadAllGeneration;
  }

  Future<void> _hydrateWorldShellBackground(
    int loadGeneration, {
    required Future<void> restoreDefeatedFuture,
    required Future<void> restoreDiscoveredCharactersFuture,
    required Future<void> discoveriesFuture,
    required Future<List<BasePin>> basesFuture,
  }) async {
    unawaited(_applyBootstrapBases(loadGeneration, basesFuture));
    unawaited(
      _applyBootstrapZoneContent(
        loadGeneration,
        restoreDefeatedFuture: restoreDefeatedFuture,
        restoreDiscoveredCharactersFuture: restoreDiscoveredCharactersFuture,
      ),
    );

    try {
      await discoveriesFuture;
    } catch (e) {
      debugPrint('SinglePlayer: discoveries bootstrap error: $e');
    }
  }

  Future<void> _applyBootstrapBases(
    int loadGeneration,
    Future<List<BasePin>> basesFuture,
  ) async {
    try {
      final bases = await basesFuture;
      if (!_isCurrentLoadGeneration(loadGeneration)) return;
      setState(() {
        _bases = bases;
      });
      if (_styleLoaded && _mapController != null && _markersAdded) {
        await _refreshBaseSymbols();
        if (_isPlacingBase) {
          await _hideMarkersForBasePlacement();
          await _refreshBasePlacementPreview();
        }
      }
    } catch (e) {
      debugPrint('SinglePlayer: bootstrap bases error: $e');
    }
  }

  Future<void> _applyBootstrapZoneContent(
    int loadGeneration, {
    required Future<void> restoreDefeatedFuture,
    required Future<void> restoreDiscoveredCharactersFuture,
  }) async {
    try {
      await Future.wait<void>([
        restoreDefeatedFuture,
        restoreDiscoveredCharactersFuture,
      ]);
      if (!_isCurrentLoadGeneration(loadGeneration)) return;
      await _loadTreasureChestsForSelectedZone();
    } catch (e) {
      debugPrint('SinglePlayer: bootstrap zone content error: $e');
    }
  }

  Future<bool> _applyTutorialStatusUpdate(
    TutorialStatus? status, {
    bool preserveCompletedReveal = false,
  }) async {
    if (!mounted || status == null) return false;
    final messenger = ScaffoldMessenger.maybeOf(context);

    setState(() {
      _tutorialStatus = status;
      _tutorialStatusChecked = true;
    });
    _syncQuestLogToTutorialStatus(status);
    await _ensureTutorialScenarioLoaded(status);
    await _ensureTutorialMonsterLoaded(status);
    await _syncTutorialMapModeFromStatus(status);
    if (!preserveCompletedReveal) {
      await _resetTutorialPresentationForInactiveStatus(status);
    }
    _syncTutorialInventorySession(status);
    _syncTutorialGuideButtonPulse(status);
    _openTrackedQuestsForTutorialObjectiveUpdates(status);
    if (_tutorialReplayPending &&
        (status.character == null || status.dialogue.isEmpty)) {
      _tutorialReplayPending = false;
      messenger?.showSnackBar(
        const SnackBar(content: Text('Tutorial is not configured yet.')),
      );
    }
    return true;
  }

  Future<void> _loadTutorialStatus({
    bool force = false,
    bool preserveCompletedReveal = false,
    bool triggerDialogues = true,
  }) async {
    if (!mounted) return;
    if (_tutorialStatusLoading) {
      if (force) {
        _tutorialStatusReloadQueued = true;
        _tutorialStatusReloadQueuedPreserveCompletedReveal =
            _tutorialStatusReloadQueuedPreserveCompletedReveal ||
            preserveCompletedReveal;
      }
      return;
    }
    if (!force && _tutorialStatusChecked) {
      if (triggerDialogues) {
        _maybeShowTutorialDialogues();
      }
      return;
    }

    _tutorialStatusLoading = true;
    var loadedStatus = false;
    final messenger = ScaffoldMessenger.maybeOf(context);
    try {
      final status = await context.read<PoiService>().getTutorialStatus();
      if (!mounted) return;
      loadedStatus = await _applyTutorialStatusUpdate(
        status,
        preserveCompletedReveal: preserveCompletedReveal,
      );
      if (!loadedStatus) {
        debugPrint(
          'SinglePlayer: tutorial status request returned no data; preserving existing tutorial state.',
        );
        if (_tutorialReplayPending && !_tutorialStatusChecked) {
          _tutorialReplayPending = false;
          messenger?.showSnackBar(
            const SnackBar(content: Text('Tutorial is not configured yet.')),
          );
        }
      }
    } finally {
      _tutorialStatusLoading = false;
    }

    if (loadedStatus && triggerDialogues) {
      _maybeShowTutorialDialogues();
    }
    if (_tutorialStatusReloadQueued) {
      final queuedPreserveCompletedReveal =
          _tutorialStatusReloadQueuedPreserveCompletedReveal;
      _tutorialStatusReloadQueued = false;
      _tutorialStatusReloadQueuedPreserveCompletedReveal = false;
      unawaited(
        _loadTutorialStatus(
          force: true,
          preserveCompletedReveal: queuedPreserveCompletedReveal,
          triggerDialogues: triggerDialogues,
        ),
      );
    }
  }

  Future<void> _refreshTutorialAfterBaseInteraction() async {
    await _loadTutorialStatus(force: true, preserveCompletedReveal: true);
    if (!mounted) return;

    try {
      await _loadTreasureChestsForSelectedZone(forceRefreshZonePins: true);
      if (!mounted) return;
      await _loadBases();
      if (!mounted) return;
      _requestQuestLogIfReady(force: true);
      await _rebuildMapPins();
    } catch (error) {
      debugPrint(
        'SinglePlayer: failed to refresh tutorial state after base interaction: $error',
      );
    }
  }

  Future<void> _resetTutorialPresentationForInactiveStatus(
    TutorialStatus? status,
  ) async {
    if (status == null) return;
    final tutorialStillActive =
        status.showWelcomeDialogue ||
        status.isPostWelcomeDialogueStep ||
        status.hasActiveScenario ||
        status.isPostScenarioDialogueStep ||
        status.isLoadoutStep ||
        status.isBaseKitStep ||
        status.hasActiveMonsterEncounter ||
        status.isPostMonsterDialogueStep ||
        status.isPostBaseDialogueStep;
    if (tutorialStillActive) return;

    final hasStalePresentation =
        (_tutorialFocusedScenarioId?.trim().isNotEmpty ?? false) ||
        (_tutorialFocusedMonsterEncounterId?.trim().isNotEmpty ?? false) ||
        _tutorialNormalPinsRevealInProgress ||
        _tutorialWelcomeOverlayVisible ||
        _tutorialWelcomeOverlayOpacity > 0.0 ||
        _tutorialGuideDockVisible;
    if (!hasStalePresentation) return;

    _tutorialGuideDockController.stop();
    setState(() {
      _tutorialFocusedScenarioId = null;
      _tutorialFocusedMonsterEncounterId = null;
      _tutorialNormalPinsRevealInProgress = false;
      _tutorialWelcomeOverlayVisible = false;
      _tutorialWelcomeOverlayOpacity = 0.0;
      _tutorialGuideDockVisible = false;
      _tutorialGuideDockCharacter = null;
      _tutorialGuideDockExcerpt = '';
      _tutorialLoadoutPendingAfterCompletionModal = false;
      _tutorialPostMonsterDialoguePendingAfterCompletionModal = false;
      _tutorialRevealPendingAfterCompletionModal = false;
    });
    await _rebuildMapPins();
  }

  void _syncQuestLogToTutorialStatus(TutorialStatus? status) {
    final nextSignature = _tutorialQuestSyncSignature(status);
    if (nextSignature == _lastTutorialQuestSyncSignature) {
      return;
    }
    _lastTutorialQuestSyncSignature = nextSignature;
    _requestQuestLogIfReady(force: true);
  }

  String _tutorialQuestSyncSignature(TutorialStatus? status) {
    if (status == null) {
      return 'inactive';
    }
    if (status.isCompleted) {
      return 'completed';
    }
    return [
      status.stage.trim(),
      status.scenarioId?.trim() ?? '',
      status.monsterEncounterId?.trim() ?? '',
      status.requiredEquipItemIds.join(','),
      status.requiredUseItemIds.join(','),
    ].join('|');
  }

  List<DialogueMessage> _tutorialGuideDialogueForStatus(
    TutorialStatus? status,
  ) {
    if (status == null) return const [];
    final primary = status.postWelcomeDialogue;
    if (primary.isNotEmpty) {
      return List<DialogueMessage>.from(primary)
        ..sort((a, b) => a.order.compareTo(b.order));
    }
    final fallback = status.dialogue;
    return List<DialogueMessage>.from(fallback)
      ..sort((a, b) => a.order.compareTo(b.order));
  }

  String _tutorialGuideDockExcerptForStatus(TutorialStatus? status) {
    final dialogue = _tutorialGuideDialogueForStatus(status);
    for (final line in dialogue) {
      final text = line.text.trim();
      if (text.isNotEmpty) return text;
    }
    return 'Tap the portrait on the left whenever you want to hear from me.';
  }

  bool _isTutorialGuideButtonUnlocked(TutorialStatus? status) {
    if (status == null || status.character == null) {
      return false;
    }
    return status.isPostWelcomeDialogueStep ||
        status.hasActiveScenario ||
        status.isPostScenarioDialogueStep ||
        status.isLoadoutStep ||
        status.isBaseKitStep ||
        status.hasActiveMonsterEncounter ||
        status.isPostMonsterDialogueStep ||
        status.isPostBaseDialogueStep ||
        status.isCompleted;
  }

  Future<void> _showTutorialGuideButtonInteraction() async {
    final status = _tutorialStatus;
    if (status != null &&
        status.isCompleted &&
        _tutorialGuideButtonAcknowledged) {
      await _showTutorialGuideSupportChat();
      return;
    }
    await _showTutorialGuideButtonDialogue();
  }

  String _tutorialGuideSupportGreetingForStatus(TutorialStatus? status) {
    final configuredGreeting = status?.guideSupportGreeting.trim() ?? '';
    if (configuredGreeting.isNotEmpty) {
      return configuredGreeting;
    }
    final characterName = status?.character?.name.trim() ?? 'your guide';
    return '$characterName is here to help. Ask about quests, combat, equipment, your base, or what to do next, and you will get an in-world answer.';
  }

  Future<void> _showTutorialGuideButtonDialogue() async {
    await _markTutorialGuideButtonAcknowledged();
    if (!mounted) return;
    final status = _tutorialStatus;
    final character = status?.character;
    final dialogue = _tutorialGuideDialogueForStatus(status);
    if (character == null || dialogue.isEmpty || _tutorialDialogVisible) {
      return;
    }

    final action = CharacterAction(
      id: 'tutorial-guide-button',
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
        useSafeArea: false,
        barrierDismissible: true,
        barrierColor: Colors.transparent,
        builder: (dialogContext) {
          return RpgDialogueModal(
            character: character,
            action: action,
            dialogueOverride: dialogue,
            finalStepLabel: 'Close',
            onClose: () => Navigator.of(dialogContext).pop(),
          );
        },
      );
    } finally {
      if (mounted) {
        setState(() => _tutorialDialogVisible = false);
      }
    }
  }

  Future<void> _showTutorialGuideSupportChat() async {
    final status = _tutorialStatus;
    final character = status?.character;
    if (character == null || _tutorialDialogVisible) {
      return;
    }

    setState(() => _tutorialDialogVisible = true);
    try {
      await showDialog<void>(
        context: context,
        useRootNavigator: true,
        useSafeArea: false,
        barrierDismissible: true,
        barrierColor: Colors.transparent,
        builder: (dialogContext) {
          return TutorialGuideChatModal(
            character: character,
            initialAssistantMessage: _tutorialGuideSupportGreetingForStatus(
              status,
            ),
            onClose: () => Navigator.of(dialogContext).pop(),
            onSendMessage: (message, history) async {
              try {
                final response = await context
                    .read<PoiService>()
                    .sendTutorialGuideChat(message: message, history: history);
                final answer = response.message.trim();
                if (answer.isNotEmpty) {
                  return answer;
                }
              } on DioException catch (error) {
                throw Exception(
                  PoiService.extractApiErrorMessage(
                    error,
                    'Guide support is unavailable right now.',
                  ),
                );
              }
              return '${character.name} pauses for a moment. "I do not have a clean answer for that right now."';
            },
          );
        },
      );
    } on DioException catch (error) {
      if (!mounted) return;
      final message = PoiService.extractApiErrorMessage(
        error,
        'Guide support is unavailable right now.',
      );
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
    } finally {
      if (mounted) {
        setState(() => _tutorialDialogVisible = false);
      }
    }
  }

  String _tutorialGuidePortraitUrl(Character? character) {
    if (character == null) return '';
    final dialogue = character.dialogueImageUrl?.trim() ?? '';
    if (dialogue.isNotEmpty) return dialogue;
    final thumbnail = character.thumbnailUrl?.trim() ?? '';
    if (thumbnail.isNotEmpty) return thumbnail;
    final mapIcon = character.mapIconUrl?.trim() ?? '';
    if (mapIcon.isNotEmpty) return mapIcon;
    return '';
  }

  Rect _tutorialGuideDockTargetRect(BuildContext context) {
    final topInset = MediaQuery.paddingOf(context).top;
    final targetTop =
        topInset +
        (3 * _overlayRailButtonSize) +
        (3 * _overlayRailButtonSpacing);
    return const Rect.fromLTWH(
      _overlayRailButtonLeftInset,
      0,
      _overlayRailButtonSize,
      _overlayRailButtonSize,
    ).shift(Offset(0, targetTop));
  }

  Future<void> _runTutorialGuideDockSequence(TutorialStatus status) async {
    final character = status.character;
    if (character == null || !mounted) return;

    setState(() {
      _tutorialGuideDockCharacter = character;
      _tutorialGuideDockExcerpt = _tutorialGuideDockExcerptForStatus(status);
      _tutorialGuideDockVisible = true;
    });
    _syncTutorialGuideButtonPulse(status);
    try {
      await _tutorialGuideDockController.forward(from: 0);
    } finally {
      if (mounted) {
        setState(() {
          _tutorialGuideDockVisible = false;
          _tutorialGuideDockCharacter = null;
          _tutorialGuideDockExcerpt = '';
        });
        _syncTutorialGuideButtonPulse(_tutorialStatus);
      } else {
        _tutorialGuideDockVisible = false;
        _tutorialGuideDockCharacter = null;
        _tutorialGuideDockExcerpt = '';
      }
    }
  }

  void _maybeShowTutorialDialogues() {
    if (!mounted ||
        _tutorialDialogVisible ||
        _tutorialAdvanceInFlight ||
        _tutorialActivationInFlight ||
        _tutorialReplayResetInFlight ||
        _tutorialGuideDockVisible) {
      return;
    }
    if (_completedTaskProvider?.currentModal != null) {
      return;
    }
    final status = _tutorialStatus;
    if (status == null || status.character == null) {
      return;
    }
    if (status.shouldShowPostWelcomeDialogue) {
      unawaited(_showTutorialProgressDialogue(status, stage: 'post_welcome'));
      return;
    }
    if (status.isPostWelcomeDialogueStep) {
      unawaited(
        _completeTutorialPostWelcomeStep(
          status,
          forceReplay: _tutorialReplayPending,
        ),
      );
      return;
    }
    if (status.shouldShowPostScenarioDialogue) {
      unawaited(_showTutorialProgressDialogue(status, stage: 'post_scenario'));
      return;
    }
    if (status.isPostScenarioDialogueStep) {
      unawaited(_advanceTutorialAfterDialogue('post_scenario_dialogue_closed'));
      return;
    }
    if (status.shouldShowPostMonsterDialogue) {
      unawaited(_showTutorialProgressDialogue(status, stage: 'post_monster'));
      return;
    }
    if (status.isPostMonsterDialogueStep) {
      unawaited(_advanceTutorialAfterDialogue('post_monster_dialogue_closed'));
      return;
    }
    if (status.shouldShowPostBaseDialogue) {
      unawaited(_showTutorialProgressDialogue(status, stage: 'post_base'));
      return;
    }
    if (status.isPostBaseDialogueStep) {
      unawaited(_completeTutorialAfterBaseDialogue());
      return;
    }
    if (status.dialogue.isEmpty) {
      if (status.showWelcomeDialogue || _tutorialReplayPending) {
        unawaited(
          _advanceTutorialAfterWelcome(
            status,
            forceReplay: _tutorialReplayPending,
          ),
        );
      }
      return;
    }
    if (!status.showWelcomeDialogue && !_tutorialReplayPending) {
      return;
    }
    final location = context.read<LocationProvider>().location;
    if (location == null && !_tutorialReplayPending) return;
    unawaited(
      _showTutorialWelcomeDialog(status, forceReplay: _tutorialReplayPending),
    );
  }

  Future<void> _showTutorialProgressDialogue(
    TutorialStatus status, {
    required String stage,
  }) async {
    final character = status.character;
    if (character == null || _tutorialDialogVisible) {
      return;
    }

    final rawDialogue = stage == 'post_base'
        ? status.postBaseDialogue
        : stage == 'post_welcome'
        ? status.postWelcomeDialogue
        : stage == 'post_scenario'
        ? status.postScenarioDialogue
        : status.postMonsterDialogue;
    if (rawDialogue.isEmpty) {
      await _completeTutorialDialogueStage(status, stage);
      return;
    }

    final dialogue = List<DialogueMessage>.from(rawDialogue)
      ..sort((a, b) => a.order.compareTo(b.order));
    final action = CharacterAction(
      id: 'tutorial-$stage',
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
        useSafeArea: false,
        barrierDismissible: false,
        barrierColor: Colors.transparent,
        builder: (dialogContext) {
          return PopScope(
            canPop: false,
            child: RpgDialogueModal(
              character: character,
              action: action,
              dialogueOverride: dialogue,
              showCloseButton: false,
              finalStepLabel: 'Continue',
              onClose: () async {
                Navigator.of(dialogContext).pop();
                await _completeTutorialDialogueStage(status, stage);
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

  Future<void> _completeTutorialDialogueStage(
    TutorialStatus status,
    String stage,
  ) async {
    switch (stage) {
      case 'post_welcome':
        await _completeTutorialPostWelcomeStep(
          status,
          forceReplay: _tutorialReplayPending,
        );
        return;
      case 'post_scenario':
        await _advanceTutorialAfterDialogue('post_scenario_dialogue_closed');
        return;
      case 'post_base':
        await _completeTutorialAfterBaseDialogue();
        return;
      default:
        await _advanceTutorialAfterDialogue('post_monster_dialogue_closed');
    }
  }

  Future<void> _advanceTutorialAfterWelcome(
    TutorialStatus status, {
    bool forceReplay = false,
  }) async {
    await _advanceTutorialStateAction('welcome_dialogue_closed');
    if (!mounted) return;
    final nextStatus = _tutorialStatus;
    if (nextStatus == null || nextStatus.stage.trim() == 'welcome') {
      return;
    }
    if (nextStatus.isPostWelcomeDialogueStep ||
        nextStatus.shouldShowPostWelcomeDialogue) {
      _maybeShowTutorialDialogues();
      return;
    }
    await _completeTutorialPostWelcomeStep(
      nextStatus,
      forceReplay: forceReplay,
    );
  }

  Future<void> _completeTutorialPostWelcomeStep(
    TutorialStatus status, {
    bool forceReplay = false,
  }) async {
    if (_tutorialGuideDockVisible || _tutorialActivationInFlight) {
      return;
    }
    await _runTutorialGuideDockSequence(status);
    if (!mounted) return;
    await _activateTutorialScenario(status, forceReplay: forceReplay);
  }

  Future<void> _advanceTutorialAfterDialogue(String action) async {
    await _advanceTutorialStateAction(action);
    if (!mounted) return;
    final currentStage = _tutorialStatus?.stage.trim() ?? '';
    final blockedStage = action == 'post_scenario_dialogue_closed'
        ? 'post_scenario_dialogue'
        : action == 'post_monster_dialogue_closed'
        ? 'post_monster_dialogue'
        : '';
    if (blockedStage.isNotEmpty && currentStage == blockedStage) {
      return;
    }
    _maybeShowTutorialDialogues();
  }

  Future<void> _completeTutorialAfterBaseDialogue() async {
    await _advanceTutorialStateAction(
      'post_base_dialogue_closed',
      preserveCompletedReveal: true,
    );
    if (!mounted) return;
    if ((_tutorialStatus?.stage.trim() ?? '') == 'post_base_dialogue') {
      return;
    }
    await _beginTutorialNormalPinsReveal();
  }

  Future<void> _advanceTutorialStateAction(
    String action, {
    bool preserveCompletedReveal = false,
  }) async {
    if (!mounted || _tutorialAdvanceInFlight) return;

    _tutorialAdvanceInFlight = true;
    try {
      final status = await context.read<PoiService>().advanceTutorial(action);
      if (!mounted) return;
      final updated = await _applyTutorialStatusUpdate(
        status,
        preserveCompletedReveal: preserveCompletedReveal,
      );
      if (!updated) {
        await _loadTutorialStatus(
          force: true,
          preserveCompletedReveal: preserveCompletedReveal,
          triggerDialogues: false,
        );
      }
    } finally {
      _tutorialAdvanceInFlight = false;
    }
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

    final dialogue = List<DialogueMessage>.from(status.dialogue)
      ..sort((a, b) => a.order.compareTo(b.order));
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
        useSafeArea: false,
        barrierDismissible: false,
        barrierColor: Colors.transparent,
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
                await _advanceTutorialAfterWelcome(
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
        await _loadTutorialStatus(force: true, preserveCompletedReveal: true);
        return;
      }

      setState(() {
        _tutorialReplayPending = false;
        _tutorialFocusedScenarioId = scenario.id;
        _tutorialFocusedMonsterEncounterId = null;
        _tutorialNormalPinsRevealInProgress = false;
        _tutorialLoadoutPendingAfterCompletionModal = false;
        _tutorialPostMonsterDialoguePendingAfterCompletionModal = false;
        _tutorialRevealPendingAfterCompletionModal = false;
        _scenarios = [
          scenario,
          ..._scenarios.where((item) => item.id != scenario.id),
        ];
      });
      final questLog = context.read<QuestLogProvider>();
      await _rebuildMapPins();
      if (!mounted) return;
      _flyToLocation(scenario.latitude, scenario.longitude);
      unawaited(_pulsePoi(scenario.latitude, scenario.longitude));
      await questLog.refresh();
      if (!mounted) return;
      final tutorialTrackedQuestItemId =
          _primaryTrackedTutorialQuestCarouselItemId(questLog);
      if (tutorialTrackedQuestItemId != null) {
        _trackedQuestsController.open(itemId: tutorialTrackedQuestItemId);
      }
      unawaited(_loadTutorialStatus(force: true));
    } catch (error) {
      await _loadTutorialStatus(force: true, preserveCompletedReveal: true);
      if (!mounted) return;

      final recoveredScenarioId =
          _tutorialStatus != null && _tutorialStatus!.hasActiveScenario
          ? _tutorialStatus!.scenarioId?.trim() ?? ''
          : '';
      if (recoveredScenarioId.isNotEmpty) {
        Scenario? recoveredScenario;
        for (final item in _scenarios) {
          if (item.id == recoveredScenarioId) {
            recoveredScenario = item;
            break;
          }
        }
        recoveredScenario ??= await context.read<PoiService>().getScenarioById(
          recoveredScenarioId,
        );
        if (recoveredScenario != null && mounted) {
          final recovered = recoveredScenario;
          setState(() {
            _tutorialReplayPending = false;
            _tutorialFocusedScenarioId = recovered.id;
            _tutorialFocusedMonsterEncounterId = null;
            _tutorialNormalPinsRevealInProgress = false;
            _tutorialLoadoutPendingAfterCompletionModal = false;
            _tutorialPostMonsterDialoguePendingAfterCompletionModal = false;
            _tutorialRevealPendingAfterCompletionModal = false;
            _scenarios = [
              recovered,
              ..._scenarios.where((item) => item.id != recovered.id),
            ];
          });
          await _rebuildMapPins();
          if (!mounted) return;
          _flyToLocation(
            recoveredScenario.latitude,
            recoveredScenario.longitude,
          );
          unawaited(
            _pulsePoi(recoveredScenario.latitude, recoveredScenario.longitude),
          );
          return;
        }
      }

      final message = PoiService.extractApiErrorMessage(
        error,
        'Failed to start the tutorial scenario.',
      );
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(message)));
      }
    } finally {
      if (mounted) {
        setState(() => _tutorialActivationInFlight = false);
      }
    }
  }

  Future<void> _loadTreasureChestsForSelectedZone({
    bool forceRefreshZonePins = false,
    bool forceRefreshZoneBaseContent = false,
  }) async {
    final requestVersion = ++_zoneContentRequestVersion;
    final zoneId =
        context.read<ZoneProvider>().selectedZone?.id ??
        (_zones.isNotEmpty ? _zones.first.id : null);
    if (zoneId == null) {
      if (mounted && requestVersion == _zoneContentRequestVersion) {
        setState(() {
          _pois = [];
          _characters = [];
          _treasureChests = [];
          _healingFountains = [];
          _resources = [];
          _scenarios = [];
          _expositions = [];
          _monsters = [];
          _challenges = [];
        });
      }
      _setLoadingZoneTransition(null);
      return;
    }
    final zoneChanged = _renderedTreasureChestZoneId != zoneId;
    final hasZoneSnapshot = _hasZoneBaseContentSnapshot(zoneId);
    final hasZonePinsSnapshot = _hasZonePinContentSnapshot(zoneId);
    _setLoadingZoneTransition(
      zoneChanged && !(hasZoneSnapshot && hasZonePinsSnapshot) ? zoneId : null,
    );
    if (zoneChanged &&
        hasZoneSnapshot &&
        hasZonePinsSnapshot &&
        _styleLoaded &&
        _mapController != null &&
        _markersAdded &&
        !_shouldSuppressNormalMapPinsForTutorial &&
        !_tutorialNormalPinsRevealInProgress) {
      _applyCachedZoneSnapshot(zoneId);
      await _refreshZoneScopedMapPins();
      if (!mounted || requestVersion != _zoneContentRequestVersion) return;
    }
    try {
      final svc = context.read<PoiService>();
      final pinContentFuture = _getZonePinContent(
        zoneId,
        svc: svc,
        forceRefresh: forceRefreshZonePins,
      );
      final baseContentFuture = _getZoneBaseContent(
        zoneId,
        svc: svc,
        forceRefresh: forceRefreshZoneBaseContent,
      );
      final pinContent = await pinContentFuture;
      final baseContent = await baseContentFuture;
      final chests = baseContent.treasureChests;
      final healingFountains = baseContent.healingFountains;
      final resources = baseContent.resources;
      final baseScenarios = baseContent.scenarios;
      final baseExpositions = baseContent.expositions;
      final baseMonsters = baseContent.monsters;
      final baseChallenges = baseContent.challenges;
      if (!mounted || requestVersion != _zoneContentRequestVersion) return;
      _rememberKnownZonePins(pinContent);
      final currentQuestScenarioIds = _currentQuestScenarioIds();
      final currentQuestExpositionIds = _currentQuestExpositionIds();
      final currentQuestMonsterIds = _currentQuestMonsterIds();
      final currentQuestChallengeIds = _currentQuestChallengeIds();

      final scenarioById = <String, Scenario>{
        for (final scenario in baseScenarios) scenario.id: scenario,
      };
      final expositionById = <String, Exposition>{
        for (final exposition in baseExpositions) exposition.id: exposition,
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
      for (final expositionId in currentQuestExpositionIds) {
        if (expositionById.containsKey(expositionId)) continue;
        final exposition = await svc.getExpositionById(expositionId);
        if (exposition == null) continue;
        if (exposition.zoneId != zoneId) continue;
        expositionById[exposition.id] = exposition;
      }
      final activeTutorialScenarioId =
          _tutorialStatus != null && _tutorialStatus!.hasActiveScenario
          ? _tutorialStatus!.scenarioId?.trim() ?? ''
          : '';
      if (activeTutorialScenarioId.isNotEmpty &&
          !scenarioById.containsKey(activeTutorialScenarioId)) {
        Scenario? tutorialScenario;
        for (final scenario in _scenarios) {
          if (scenario.id == activeTutorialScenarioId) {
            tutorialScenario = scenario;
            break;
          }
        }
        tutorialScenario ??= await svc.getScenarioById(
          activeTutorialScenarioId,
        );
        if (tutorialScenario != null && tutorialScenario.zoneId == zoneId) {
          scenarioById[tutorialScenario.id] = tutorialScenario;
        }
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
      final activeTutorialMonsterEncounterId =
          _tutorialStatus != null && _tutorialStatus!.hasActiveMonsterEncounter
          ? _tutorialStatus!.monsterEncounterId?.trim() ?? ''
          : '';
      if (activeTutorialMonsterEncounterId.isNotEmpty &&
          !monsterById.containsKey(activeTutorialMonsterEncounterId)) {
        MonsterEncounter? tutorialEncounter;
        for (final encounter in _monsters) {
          if (encounter.id == activeTutorialMonsterEncounterId) {
            tutorialEncounter = encounter;
            break;
          }
        }
        tutorialEncounter ??= await svc.getMonsterEncounterById(
          activeTutorialMonsterEncounterId,
        );
        if (tutorialEncounter != null && tutorialEncounter.zoneId == zoneId) {
          monsterById[tutorialEncounter.id] = tutorialEncounter;
        }
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
                scenario.id == activeTutorialScenarioId ||
                (!scenario.attemptedByUser &&
                    !_resolvedScenarioIds.contains(scenario.id) &&
                    !_resolvedScenarioSignatures.contains(
                      _scenarioSignature(scenario),
                    )),
          )
          .toList();
      final expositions = expositionById.values.toList();
      final monsters = monsterById.values
          .where(
            (monster) =>
                currentQuestMonsterIds.contains(monster.id) ||
                monster.id == activeTutorialMonsterEncounterId ||
                !_defeatedMonsterIds.contains(monster.id),
          )
          .toList();
      final challenges = challengeById.values.toList();
      if (!mounted || requestVersion != _zoneContentRequestVersion) return;
      _applyZoneContent(
        zoneId,
        pinContent: pinContent,
        treasureChests: chests
            .where((chest) => !_openedTreasureChestIds.contains(chest.id))
            .toList(growable: false),
        healingFountains: healingFountains,
        resources: resources
            .where(
              (resource) =>
                  !resource.gatheredByUser &&
                  !_gatheredResourceIds.contains(resource.id),
            )
            .toList(growable: false),
        scenarios: scenarios,
        expositions: expositions,
        monsters: monsters,
        challenges: challenges,
      );
      if (_styleLoaded && _mapController != null && _markersAdded) {
        await _refreshZoneScopedMapPins();
      }
      _clearLoadingZoneTransition(zoneId, requestVersion);
    } catch (e) {
      debugPrint('SinglePlayer: _loadTreasureChests/scenarios error: $e');
      if (mounted && requestVersion == _zoneContentRequestVersion) {
        setState(() {
          _pois = [];
          _characters = [];
          _treasureChests = [];
          _healingFountains = [];
          _resources = [];
          _scenarios = [];
          _expositions = [];
          _monsters = [];
          _challenges = [];
        });
        _renderedTreasureChestZoneId = null;
        if (_styleLoaded && _mapController != null && _markersAdded) {
          await _refreshZoneScopedMapPins();
        }
      }
      _clearLoadingZoneTransition(zoneId, requestVersion);
    }
  }

  void _applyZoneContent(
    String zoneId, {
    required _ZonePinContent pinContent,
    required List<TreasureChest> treasureChests,
    required List<HealingFountain> healingFountains,
    required List<ResourceNode> resources,
    required List<Scenario> scenarios,
    required List<Exposition> expositions,
    required List<MonsterEncounter> monsters,
    required List<Challenge> challenges,
  }) {
    void update() {
      _pois = pinContent.pointsOfInterest;
      _characters = pinContent.characters;
      _treasureChests = treasureChests;
      _healingFountains = healingFountains;
      _resources = resources;
      _scenarios = scenarios;
      _expositions = expositions;
      _monsters = monsters;
      _challenges = challenges;
      _renderedTreasureChestZoneId = zoneId;
    }

    if (mounted) {
      setState(update);
    } else {
      update();
    }
  }

  void _setLoadingZoneTransition(String? zoneId) {
    final normalizedZoneId = zoneId?.trim();
    if (_loadingZoneTransitionZoneId == normalizedZoneId) return;
    if (mounted) {
      setState(() => _loadingZoneTransitionZoneId = normalizedZoneId);
    } else {
      _loadingZoneTransitionZoneId = normalizedZoneId;
    }
  }

  void _clearLoadingZoneTransition(String zoneId, int requestVersion) {
    if (requestVersion != _zoneContentRequestVersion) return;
    if (_loadingZoneTransitionZoneId != zoneId) return;
    _setLoadingZoneTransition(null);
  }

  void _applyCachedZoneSnapshot(String zoneId) {
    final normalizedZoneId = zoneId.trim();
    if (normalizedZoneId.isEmpty) return;
    final cachedBase = _zoneBaseContentCache[normalizedZoneId];
    final cachedPins = _zonePinContentCache[normalizedZoneId];
    if (cachedBase == null || cachedPins == null) return;
    cachedBase.touch();
    cachedPins.touch();

    final currentQuestScenarioIds = _currentQuestScenarioIds();
    final currentQuestMonsterIds = _currentQuestMonsterIds();
    final activeTutorialScenarioId =
        _tutorialStatus != null && _tutorialStatus!.hasActiveScenario
        ? _tutorialStatus!.scenarioId?.trim() ?? ''
        : '';
    final activeTutorialMonsterEncounterId =
        _tutorialStatus != null && _tutorialStatus!.hasActiveMonsterEncounter
        ? _tutorialStatus!.monsterEncounterId?.trim() ?? ''
        : '';
    final content = cachedBase.content;
    _applyZoneContent(
      normalizedZoneId,
      pinContent: cachedPins.content,
      treasureChests: content.treasureChests
          .where((chest) => !_openedTreasureChestIds.contains(chest.id))
          .toList(growable: false),
      healingFountains: content.healingFountains,
      resources: content.resources
          .where(
            (resource) =>
                !resource.gatheredByUser &&
                !_gatheredResourceIds.contains(resource.id),
          )
          .toList(growable: false),
      scenarios: content.scenarios
          .where(
            (scenario) =>
                currentQuestScenarioIds.contains(scenario.id) ||
                scenario.id == activeTutorialScenarioId ||
                (!scenario.attemptedByUser &&
                    !_resolvedScenarioIds.contains(scenario.id) &&
                    !_resolvedScenarioSignatures.contains(
                      _scenarioSignature(scenario),
                    )),
          )
          .toList(growable: false),
      expositions: content.expositions,
      monsters: content.monsters
          .where(
            (monster) =>
                currentQuestMonsterIds.contains(monster.id) ||
                monster.id == activeTutorialMonsterEncounterId ||
                !_defeatedMonsterIds.contains(monster.id),
          )
          .toList(growable: false),
      challenges: content.challenges,
    );
  }

  Future<void> _refreshZoneScopedMapPins() async {
    if (!_styleLoaded || _mapController == null || !_markersAdded) return;
    if (_shouldSuppressNormalMapPinsForTutorial ||
        _tutorialNormalPinsRevealInProgress) {
      await _rebuildMapPins();
      return;
    }
    final controller = _mapController;
    if (controller == null) return;
    _pinBatchRevealInProgress = true;
    try {
      await _rebuildMapPins();
    } finally {
      _pinBatchRevealInProgress = false;
    }
    await _revealLoadedPins(controller);
  }

  Future<_ZoneBaseContent> _getZoneBaseContent(
    String zoneId, {
    PoiService? svc,
    bool forceRefresh = false,
  }) {
    final normalizedZoneId = zoneId.trim();
    final service = svc ?? context.read<PoiService>();
    if (forceRefresh) {
      return _fetchZoneBaseContent(
        normalizedZoneId,
        svc: service,
        forceRefresh: true,
      );
    }
    final cached = _zoneBaseContentCache[normalizedZoneId];
    if (cached != null) {
      cached.touch();
      if (!cached.isFresh) {
        // Keep movement snappy by serving the last snapshot immediately while
        // the selected zone refreshes in the background.
        unawaited(
          _warmZoneBaseContentInBackground(normalizedZoneId, svc: service),
        );
      }
      return Future<_ZoneBaseContent>.value(cached.content);
    }
    final inFlight = _zoneBaseContentRequests[normalizedZoneId];
    if (inFlight != null) return inFlight;
    return _fetchZoneBaseContent(normalizedZoneId, svc: service);
  }

  Future<_ZoneBaseContent> _fetchZoneBaseContent(
    String zoneId, {
    PoiService? svc,
    bool forceRefresh = false,
  }) {
    final normalizedZoneId = zoneId.trim();
    if (!forceRefresh) {
      final cached = _zoneBaseContentCache[normalizedZoneId];
      if (cached != null && cached.isFresh) {
        cached.touch();
        return Future<_ZoneBaseContent>.value(cached.content);
      }
    }

    final inFlight = _zoneBaseContentRequests[normalizedZoneId];
    if (inFlight != null) return inFlight;

    final service = svc ?? context.read<PoiService>();
    final request = () async {
      try {
        final results = await Future.wait<dynamic>([
          service.getTreasureChestsForZone(normalizedZoneId),
          service.getHealingFountainsForZone(normalizedZoneId),
          service.getResourcesForZone(normalizedZoneId),
          service.getScenariosForZone(normalizedZoneId),
          service.getExpositionsForZone(normalizedZoneId),
          service.getMonsterEncountersForZone(normalizedZoneId),
          service.getChallengesForZone(normalizedZoneId),
        ]);
        final content = _ZoneBaseContent(
          treasureChests: results[0] as List<TreasureChest>,
          healingFountains: results[1] as List<HealingFountain>,
          resources: results[2] as List<ResourceNode>,
          scenarios: results[3] as List<Scenario>,
          expositions: results[4] as List<Exposition>,
          monsters: results[5] as List<MonsterEncounter>,
          challenges: results[6] as List<Challenge>,
        );
        _storeZoneBaseContent(normalizedZoneId, content);
        return content;
      } finally {
        _zoneBaseContentRequests.remove(normalizedZoneId);
      }
    }();

    _zoneBaseContentRequests[normalizedZoneId] = request;
    return request;
  }

  Future<void> _warmZoneBaseContentInBackground(
    String zoneId, {
    PoiService? svc,
  }) async {
    try {
      await _fetchZoneBaseContent(zoneId, svc: svc, forceRefresh: true);
    } catch (_) {
      return;
    }
    if (!mounted) return;
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    if (selectedZoneId != zoneId) return;
    await _loadTreasureChestsForSelectedZone();
  }

  Future<_ZonePinContent> _getZonePinContent(
    String zoneId, {
    PoiService? svc,
    bool forceRefresh = false,
  }) {
    final normalizedZoneId = zoneId.trim();
    final service = svc ?? context.read<PoiService>();
    if (forceRefresh) {
      return _fetchZonePinContent(
        normalizedZoneId,
        svc: service,
        forceRefresh: true,
      );
    }
    final cached = _zonePinContentCache[normalizedZoneId];
    if (cached != null) {
      cached.touch();
      if (!cached.isFresh) {
        unawaited(
          _warmZonePinContentInBackground(normalizedZoneId, svc: service),
        );
      }
      return Future<_ZonePinContent>.value(cached.content);
    }
    final inFlight = _zonePinContentRequests[normalizedZoneId];
    if (inFlight != null) return inFlight;
    return _fetchZonePinContent(normalizedZoneId, svc: service);
  }

  Future<_ZonePinContent> _fetchZonePinContent(
    String zoneId, {
    PoiService? svc,
    bool forceRefresh = false,
  }) {
    final normalizedZoneId = zoneId.trim();
    if (!forceRefresh) {
      final cached = _zonePinContentCache[normalizedZoneId];
      if (cached != null && cached.isFresh) {
        cached.touch();
        return Future<_ZonePinContent>.value(cached.content);
      }
    }

    final inFlight = _zonePinContentRequests[normalizedZoneId];
    if (inFlight != null) return inFlight;

    final service = svc ?? context.read<PoiService>();
    final request = () async {
      try {
        final payload = await service.getZonePins(normalizedZoneId);
        final content = _ZonePinContent(
          pointsOfInterest: payload.pointsOfInterest,
          characters: payload.characters,
        );
        _storeZonePinContent(normalizedZoneId, content);
        return content;
      } finally {
        _zonePinContentRequests.remove(normalizedZoneId);
      }
    }();

    _zonePinContentRequests[normalizedZoneId] = request;
    return request;
  }

  Future<void> _warmZonePinContentInBackground(
    String zoneId, {
    PoiService? svc,
  }) async {
    try {
      await _fetchZonePinContent(zoneId, svc: svc, forceRefresh: true);
    } catch (_) {
      return;
    }
    if (!mounted) return;
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    if (selectedZoneId != zoneId) return;
    await _loadTreasureChestsForSelectedZone();
  }

  void _scheduleZoneBaseContentWarmup({bool immediate = false}) {
    if (!mounted || _zones.isEmpty) return;
    final location = context.read<LocationProvider>().location;
    final signature = _zoneBaseContentWarmSignature(location);
    final now = DateTime.now();
    final warmupAge = _lastZoneBaseContentWarmAt == null
        ? _zoneBaseContentWarmupThrottle
        : now.difference(_lastZoneBaseContentWarmAt!);
    final recentlyWarmed =
        _lastZoneBaseContentWarmAt != null &&
        warmupAge < _zoneBaseContentWarmupThrottle;
    if (!immediate &&
        signature == _lastZoneBaseContentWarmSignature &&
        recentlyWarmed) {
      return;
    }

    _zoneBaseContentWarmupTimer?.cancel();
    final delay = immediate
        ? Duration.zero
        : (recentlyWarmed
              ? _zoneBaseContentWarmupThrottle - warmupAge
              : _zoneBaseContentWarmupDebounce);
    _zoneBaseContentWarmupTimer = Timer(delay, () {
      if (!mounted) return;
      _lastZoneBaseContentWarmSignature = signature;
      _lastZoneBaseContentWarmAt = DateTime.now();
      unawaited(_prefetchZoneBaseContent(location: location));
    });
  }

  String _zoneBaseContentWarmSignature(AppLocation? location) {
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    return _prioritizedZoneBaseContentZoneIds(
      location: location,
      selectedZoneId: selectedZoneId,
    ).take(_zoneBaseContentWarmCount).join('|');
  }

  Future<void> _prefetchZoneBaseContent({AppLocation? location}) async {
    if (_zones.isEmpty) return;
    final svc = context.read<PoiService>();
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    final warmIds = _prioritizedZoneBaseContentZoneIds(
      location: location,
      selectedZoneId: selectedZoneId,
    ).take(_zoneBaseContentWarmCount).toList(growable: false);
    for (final zoneId in warmIds) {
      if (zoneId.isEmpty) continue;
      final pendingFetches = <Future<void>>[];
      final cached = _zoneBaseContentCache[zoneId];
      if (cached != null && cached.isFresh) {
        cached.touch();
      } else {
        pendingFetches.add(
          _fetchZoneBaseContent(
            zoneId,
            svc: svc,
            forceRefresh: cached != null,
          ).then((_) {}).catchError((_) {}),
        );
      }
      final pinCached = _zonePinContentCache[zoneId];
      if (pinCached != null && pinCached.isFresh) {
        pinCached.touch();
      } else {
        pendingFetches.add(
          _fetchZonePinContent(
            zoneId,
            svc: svc,
            forceRefresh: pinCached != null,
          ).then((_) {}).catchError((_) {}),
        );
      }
      if (pendingFetches.isNotEmpty) {
        await Future.wait(pendingFetches);
      }
    }
    _trimZoneBaseContentCache(pinnedZoneIds: warmIds.toSet());
    _trimZonePinContentCache(pinnedZoneIds: warmIds.toSet());
  }

  List<String> _prioritizedZoneBaseContentZoneIds({
    AppLocation? location,
    String? selectedZoneId,
  }) {
    final normalizedSelectedZoneId = selectedZoneId?.trim() ?? '';
    Zone? selectedZone;
    if (normalizedSelectedZoneId.isNotEmpty) {
      for (final zone in _zones) {
        if (zone.id == normalizedSelectedZoneId) {
          selectedZone = zone;
          break;
        }
      }
    }
    final anchorLatitude = location?.latitude ?? selectedZone?.latitude;
    final anchorLongitude = location?.longitude ?? selectedZone?.longitude;
    final candidates = _zones.where((zone) => zone.id.trim().isNotEmpty).map((
      zone,
    ) {
      final isSelected =
          normalizedSelectedZoneId.isNotEmpty &&
          zone.id == normalizedSelectedZoneId;
      final distance = anchorLatitude == null || anchorLongitude == null
          ? double.infinity
          : _distanceMeters(
              anchorLatitude,
              anchorLongitude,
              zone.latitude,
              zone.longitude,
            );
      return (zoneId: zone.id, isSelected: isSelected, distance: distance);
    }).toList();
    candidates.sort((a, b) {
      if (a.isSelected != b.isSelected) {
        return a.isSelected ? -1 : 1;
      }
      final distanceOrder = a.distance.compareTo(b.distance);
      if (distanceOrder != 0) return distanceOrder;
      return a.zoneId.compareTo(b.zoneId);
    });
    return candidates.map((candidate) => candidate.zoneId).toList();
  }

  void _storeZoneBaseContent(String zoneId, _ZoneBaseContent content) {
    _zoneBaseContentCache[zoneId] = _ZoneBaseContentCacheEntry(
      content: content,
    );
    _trimZoneBaseContentCache();
    unawaited(_warmZoneBaseContentThumbnails(content));
  }

  void _storeZonePinContent(String zoneId, _ZonePinContent content) {
    _zonePinContentCache[zoneId] = _ZonePinContentCacheEntry(content: content);
    _rememberKnownZonePins(content);
    _trimZonePinContentCache();
    unawaited(_warmZonePinContentThumbnails(content));
  }

  void _trimZoneBaseContentCache({Set<String> pinnedZoneIds = const {}}) {
    if (_zoneBaseContentCache.length <= _zoneBaseContentMaxCacheEntries) {
      return;
    }
    final selectedZoneId = mounted
        ? context.read<ZoneProvider>().selectedZone?.id
        : null;
    final protectedZoneIds = <String>{
      ...pinnedZoneIds,
      if (selectedZoneId != null && selectedZoneId.isNotEmpty) selectedZoneId,
    };
    final evictableEntries =
        _zoneBaseContentCache.entries
            .where((entry) => !protectedZoneIds.contains(entry.key))
            .toList()
          ..sort(
            (a, b) => a.value.lastAccessedAt.compareTo(b.value.lastAccessedAt),
          );
    while (_zoneBaseContentCache.length > _zoneBaseContentMaxCacheEntries &&
        evictableEntries.isNotEmpty) {
      final staleEntry = evictableEntries.removeAt(0);
      _zoneBaseContentCache.remove(staleEntry.key);
    }
  }

  void _trimZonePinContentCache({Set<String> pinnedZoneIds = const {}}) {
    if (_zonePinContentCache.length <= _zoneBaseContentMaxCacheEntries) {
      return;
    }
    final selectedZoneId = mounted
        ? context.read<ZoneProvider>().selectedZone?.id
        : null;
    final protectedZoneIds = <String>{
      ...pinnedZoneIds,
      if (selectedZoneId != null && selectedZoneId.isNotEmpty) selectedZoneId,
    };
    final evictableEntries =
        _zonePinContentCache.entries
            .where((entry) => !protectedZoneIds.contains(entry.key))
            .toList()
          ..sort(
            (a, b) => a.value.lastAccessedAt.compareTo(b.value.lastAccessedAt),
          );
    while (_zonePinContentCache.length > _zoneBaseContentMaxCacheEntries &&
        evictableEntries.isNotEmpty) {
      final staleEntry = evictableEntries.removeAt(0);
      _zonePinContentCache.remove(staleEntry.key);
    }
  }

  bool _hasZoneBaseContentSnapshot(String? zoneId) {
    final normalizedZoneId = zoneId?.trim() ?? '';
    if (normalizedZoneId.isEmpty) return false;
    final cached = _zoneBaseContentCache[normalizedZoneId];
    if (cached == null) return false;
    cached.touch();
    return true;
  }

  bool _hasZonePinContentSnapshot(String? zoneId) {
    final normalizedZoneId = zoneId?.trim() ?? '';
    if (normalizedZoneId.isEmpty) return false;
    final cached = _zonePinContentCache[normalizedZoneId];
    if (cached == null) return false;
    cached.touch();
    return true;
  }

  Future<void> _warmZoneBaseContentThumbnails(_ZoneBaseContent content) async {
    final urls = _zoneContentThumbnailUrls(
      content,
    ).take(_zoneBaseContentThumbnailWarmCount).toList(growable: false);
    if (urls.isEmpty) return;
    await Future.wait(
      urls.map((url) => loadPoiThumbnail(url).catchError((_) => null)),
    );
  }

  Iterable<String> _zoneContentThumbnailUrls(_ZoneBaseContent content) {
    final seen = <String>{};
    final urls = <String>[];

    void add(String rawUrl) {
      final normalized = rawUrl.trim();
      if (normalized.isEmpty || !seen.add(normalized)) return;
      urls.add(normalized);
    }

    for (final scenario in content.scenarios) {
      add(
        scenario.thumbnailUrl.isNotEmpty
            ? scenario.thumbnailUrl
            : scenario.imageUrl,
      );
    }
    for (final monster in content.monsters) {
      add(
        monster.thumbnailUrl.isNotEmpty
            ? monster.thumbnailUrl
            : monster.imageUrl,
      );
    }
    for (final challenge in content.challenges) {
      add(
        challenge.thumbnailUrl.isNotEmpty
            ? challenge.thumbnailUrl
            : challenge.imageUrl,
      );
    }
    for (final resource in content.resources) {
      add(resource.resourceType?.mapIconUrl ?? '');
    }
    return urls;
  }

  Future<void> _warmZonePinContentThumbnails(_ZonePinContent content) async {
    final futures = <Future<void>>[];
    final seenCategories = <PoiMarkerCategory>{};
    for (final poi in content.pointsOfInterest) {
      if (!seenCategories.add(poi.markerCategory)) continue;
      futures.add(
        loadPoiCategoryThumbnail(
          poi.markerCategory,
        ).then((_) {}).catchError((_) {}),
      );
      futures.add(
        loadPoiCategoryThumbnailWithQuestMarker(
          poi.markerCategory,
        ).then((_) {}).catchError((_) {}),
      );
      futures.add(
        loadPoiCategoryThumbnailWithMainStoryMarker(
          poi.markerCategory,
        ).then((_) {}).catchError((_) {}),
      );
    }

    final seenCharacterUrls = <String>{};
    for (final character in content.characters) {
      final thumbnail = character.thumbnailUrl?.trim() ?? '';
      if (thumbnail.isEmpty || !seenCharacterUrls.add(thumbnail)) continue;
      futures.add(loadPoiThumbnail(thumbnail).then((_) {}).catchError((_) {}));
    }

    if (futures.isEmpty) return;
    await Future.wait(futures);
  }

  void _rememberKnownZonePins(_ZonePinContent content) {
    for (final poi in content.pointsOfInterest) {
      if (poi.id.isEmpty) continue;
      _knownPoiById[poi.id] = poi;
    }
    for (final character in content.characters) {
      if (character.id.isEmpty) continue;
      _knownCharacterById[character.id] = character;
    }
  }

  Iterable<PointOfInterest> _knownPois() sync* {
    final seen = <String>{};
    for (final poi in _pois) {
      if (seen.add(poi.id)) yield poi;
    }
    for (final poi in _knownPoiById.values) {
      if (seen.add(poi.id)) yield poi;
    }
  }

  Iterable<Character> _knownCharacters() sync* {
    final seen = <String>{};
    for (final character in _characters) {
      if (seen.add(character.id)) yield character;
    }
    for (final character in _knownCharacterById.values) {
      if (seen.add(character.id)) yield character;
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

  String _monsterEncounterMarkerColor(MonsterEncounter encounter) {
    if (encounter.isBossEncounter) {
      return '#d97706';
    }
    if (encounter.isRaidEncounter) {
      return '#2563eb';
    }
    return '#b63f3f';
  }

  String _monsterMysteryImageUrlForEncounterType(String encounterType) {
    switch (encounterType.trim().toLowerCase()) {
      case 'boss':
        return _bossMysteryImageUrl;
      case 'raid':
        return _raidMysteryImageUrl;
      default:
        return _monsterMysteryImageUrl;
    }
  }

  String _monsterMysteryImageIdForEncounterType(String encounterType) {
    final normalized = encounterType.trim().toLowerCase();
    switch (normalized) {
      case 'boss':
        return 'boss';
      case 'raid':
        return 'raid';
      default:
        return 'monster';
    }
  }

  bool _isChallengeMystery(Challenge challenge) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return true;
    if (challenge.hasPolygon) {
      return !_isInsidePolygon(
        location.latitude,
        location.longitude,
        challenge.polygonPoints,
      );
    }
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
    if (challenge.hasPolygon) {
      return _polygonCenter(challenge.polygonPoints);
    }
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

  List<LatLng> _challengePolygonRing(Challenge challenge) {
    final ring = challenge.polygonPoints
        .map((point) => LatLng(point.latitude, point.longitude))
        .toList();
    if (ring.length > 1 &&
        (ring.first.latitude != ring.last.latitude ||
            ring.first.longitude != ring.last.longitude)) {
      ring.add(ring.first);
    }
    return ring;
  }

  Future<void> _clearChallengePolygonOverlays() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    if (_challengePolygonLines.isNotEmpty) {
      try {
        await c.removeLines(_challengePolygonLines);
      } catch (_) {}
      _challengePolygonLines.clear();
      _challengePolygonLineById.clear();
    }
    if (_challengePolygonFills.isNotEmpty) {
      try {
        await c.removeFills(_challengePolygonFills);
      } catch (_) {}
      _challengePolygonFills.clear();
      _challengePolygonFillById.clear();
    }
  }

  Future<void> _refreshChallengePolygonOverlays(
    MapLibreMapController c,
    List<Challenge> polygonChallenges,
  ) async {
    await _clearChallengePolygonOverlays();
    if (polygonChallenges.isEmpty) return;

    final fillOptions = <FillOptions>[];
    final fillData = <Map<String, String>>[];
    final lineOptions = <LineOptions>[];
    final lineData = <Map<String, String>>[];

    for (final challenge in polygonChallenges) {
      final ring = _challengePolygonRing(challenge);
      if (ring.length < 4) continue;
      final isCurrentQuestChallenge = _isCurrentQuestChallenge(challenge.id);
      final lineColor = isCurrentQuestChallenge ? '#f5c542' : '#2563eb';
      final fillColor = isCurrentQuestChallenge ? '#f5c542' : '#2563eb';
      final fillOpacity = isCurrentQuestChallenge ? 0.45 : 0.26;
      fillOptions.add(
        FillOptions(
          geometry: [ring],
          fillColor: fillColor,
          fillOpacity: fillOpacity,
        ),
      );
      fillData.add({'type': 'challenge-polygon', 'id': challenge.id});
      lineOptions.add(
        LineOptions(
          geometry: ring,
          lineColor: lineColor,
          lineWidth: isCurrentQuestChallenge ? 3.0 : 2.6,
          lineOpacity: 1.0,
        ),
      );
      lineData.add({'type': 'challenge-polygon', 'id': challenge.id});
    }

    if (fillOptions.isEmpty || lineOptions.isEmpty) return;
    try {
      final fills = await c.addFills(fillOptions, fillData);
      if (!mounted) return;
      _challengePolygonFills.addAll(fills);
      for (var i = 0; i < fills.length && i < fillData.length; i++) {
        final id = fillData[i]['id'];
        if (id == null || id.isEmpty) continue;
        _challengePolygonFillById[id] = fills[i];
      }
      final lines = await c.addLines(lineOptions, lineData);
      if (!mounted) return;
      _challengePolygonLines.addAll(lines);
      for (var i = 0; i < lines.length && i < lineData.length; i++) {
        final id = lineData[i]['id'];
        if (id == null || id.isEmpty) continue;
        _challengePolygonLineById[id] = lines[i];
      }
    } catch (_) {}
  }

  Scenario? _scenarioById(String id) {
    for (final scenario in _scenarios) {
      if (scenario.id == id) return scenario;
    }
    return null;
  }

  Exposition? _expositionById(String id) {
    for (final exposition in _expositions) {
      if (exposition.id == id) return exposition;
    }
    return null;
  }

  PointOfInterest? _poiById(String id) {
    for (final poi in _pois) {
      if (poi.id == id) return poi;
    }
    return _knownPoiById[id];
  }

  Character? _characterById(String id) {
    for (final character in _characters) {
      if (character.id == id) return character;
    }
    return _knownCharacterById[id];
  }

  TreasureChest? _treasureChestById(String id) {
    for (final chest in _treasureChests) {
      if (chest.id == id) return chest;
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

  bool _isScenarioRepresentedByPoi(Scenario scenario) {
    final poiId = scenario.pointOfInterestId?.trim() ?? '';
    if (poiId.isEmpty) return false;
    for (final poi in _pois) {
      if (poi.id == poiId) return true;
    }
    return false;
  }

  bool _isExpositionRepresentedByPoi(Exposition exposition) {
    final poiId = exposition.pointOfInterestId?.trim() ?? '';
    if (poiId.isEmpty) return false;
    for (final poi in _pois) {
      if (poi.id == poiId) return true;
    }
    return false;
  }

  bool _isMonsterRepresentedByPoi(MonsterEncounter encounter) {
    final poiId = encounter.pointOfInterestId?.trim() ?? '';
    if (poiId.isEmpty) return false;
    for (final poi in _pois) {
      if (poi.id == poiId) return true;
    }
    return false;
  }

  List<Scenario> _linkedScenariosForPoi(String poiId) {
    if (poiId.trim().isEmpty) return const [];
    return _scenarios.where((scenario) {
      final linkedPoiId = scenario.pointOfInterestId?.trim() ?? '';
      if (linkedPoiId.isEmpty || linkedPoiId != poiId) return false;
      return _activeQuestNodeForScenario(scenario.id) == null;
    }).toList();
  }

  List<Challenge> _linkedChallengesForPoi(String poiId) {
    if (poiId.trim().isEmpty) return const [];
    return _challenges.where((challenge) {
      final linkedPoiId = challenge.pointOfInterestId?.trim() ?? '';
      if (linkedPoiId.isEmpty || linkedPoiId != poiId) return false;
      return _activeQuestNodeForChallenge(challenge.id) == null;
    }).toList();
  }

  List<Exposition> _linkedExpositionsForPoi(String poiId) {
    if (poiId.trim().isEmpty) return const [];
    return _expositions.where((exposition) {
      final linkedPoiId = exposition.pointOfInterestId?.trim() ?? '';
      if (linkedPoiId.isEmpty || linkedPoiId != poiId) return false;
      return _activeQuestNodeForExposition(exposition.id) == null;
    }).toList();
  }

  List<MonsterEncounter> _linkedMonstersForPoi(String poiId) {
    if (poiId.trim().isEmpty) return const [];
    return _monsters.where((encounter) {
      final linkedPoiId = encounter.pointOfInterestId?.trim() ?? '';
      if (linkedPoiId.isEmpty || linkedPoiId != poiId) return false;
      return _activeQuestNodeForMonsterEncounter(encounter.id) == null;
    }).toList();
  }

  bool _usesDedicatedQuestChallengeUi(Challenge challenge) {
    if (_activeQuestNodeForChallenge(challenge.id) == null) {
      return false;
    }
    return challenge.hasPolygon || _isChallengeRepresentedByPoi(challenge);
  }

  HealingFountain? _healingFountainById(String id) {
    for (final fountain in _healingFountains) {
      if (fountain.id == id) return fountain;
    }
    return null;
  }

  void _syncHealingFountainState(HealingFountain updatedFountain) {
    final normalizedZoneId = updatedFountain.zoneId.trim();
    final cached = normalizedZoneId.isEmpty
        ? null
        : _zoneBaseContentCache[normalizedZoneId];
    if (cached != null) {
      final hasMatch = cached.content.healingFountains.any(
        (item) => item.id == updatedFountain.id,
      );
      if (hasMatch) {
        _zoneBaseContentCache[normalizedZoneId] = _ZoneBaseContentCacheEntry(
          content: _ZoneBaseContent(
            treasureChests: cached.content.treasureChests,
            healingFountains: cached.content.healingFountains
                .map(
                  (item) =>
                      item.id == updatedFountain.id ? updatedFountain : item,
                )
                .toList(growable: false),
            resources: cached.content.resources,
            scenarios: cached.content.scenarios,
            expositions: cached.content.expositions,
            monsters: cached.content.monsters,
            challenges: cached.content.challenges,
          ),
          fetchedAt: cached.fetchedAt,
          lastAccessedAt: cached.lastAccessedAt,
        );
      }
    }

    if (!mounted) return;
    setState(() {
      _healingFountains = _healingFountains
          .map((item) => item.id == updatedFountain.id ? updatedFountain : item)
          .toList(growable: false);
    });
  }

  ResourceNode? _resourceById(String id) {
    for (final resource in _resources) {
      if (resource.id == id) return resource;
    }
    return null;
  }

  String _resourceImageUrl(ResourceNode resource) {
    final icon = resource.resourceType?.mapIconUrl.trim() ?? '';
    if (icon.isNotEmpty) return icon;
    return _healingFountainFallbackImageUrl;
  }

  String _resourceTypeDisplayName(ResourceNode resource) {
    final resourceTypeName = resource.resourceType?.name.trim() ?? '';
    if (resourceTypeName.isNotEmpty) return resourceTypeName;
    final resourceTypeSlug = resource.resourceType?.slug.trim() ?? '';
    if (resourceTypeSlug.isNotEmpty) {
      return resourceTypeSlug
          .split(RegExp(r'[-_\s]+'))
          .where((segment) => segment.isNotEmpty)
          .map(
            (segment) =>
                segment[0].toUpperCase() + segment.substring(1).toLowerCase(),
          )
          .join(' ');
    }
    return 'Resource';
  }

  String _mysteriousResourceTitle(ResourceNode resource) {
    return 'Mysterious ${_resourceTypeDisplayName(resource)}';
  }

  bool _isResourceWithinRevealRange(ResourceNode resource) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return false;
    final distance = _distanceMeters(
      location.latitude,
      location.longitude,
      resource.latitude,
      resource.longitude,
    );
    return distance <= kProximityUnlockRadiusMeters;
  }

  String _resourceSelectionTitle(ResourceNode resource) {
    if (!_isResourceWithinRevealRange(resource)) {
      return _mysteriousResourceTitle(resource);
    }
    return _resourceTypeDisplayName(resource);
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
    final visibleScenarios = _shouldSuppressNormalMapPinsForTutorial
        ? _scenarios
              .where((scenario) => _isTutorialFocusedScenarioId(scenario.id))
              .toList()
        : _scenarios
              .where((scenario) => !_isScenarioRepresentedByPoi(scenario))
              .toList();

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
        _setQuestCircleHighlight(circle, false);
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
        _setQuestCircleHighlight(entry.value, false);
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
      // Keep scenario pins undiscovered on the map even after proximity unlocks
      // the scenario panel content.
      const mystery = true;
      final isCurrentQuestScenario = _isCurrentQuestScenario(scenario.id);
      final isTutorialScenario = _isTutorialFocusedScenarioId(scenario.id);
      final shouldShowQuestState = isCurrentQuestScenario || isTutorialScenario;
      final shouldAnimateQuest =
          _isTrackedQuestScenario(scenario.id) || isTutorialScenario;
      final existingSymbol = _scenarioSymbolById[scenario.id];
      final existingCircle = _scenarioCircleById[scenario.id];
      final needsRefresh =
          _scenarioCircleMystery[scenario.id] != mystery ||
          existingSymbol == null ||
          _scenarioQuestObjective[scenario.id] != shouldAnimateQuest;

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
            _setQuestCircleHighlight(existingCircle, false);
            try {
              await c.removeCircle(existingCircle);
            } catch (_) {}
            _scenarioCircles.remove(existingCircle);
            _scenarioCircleById.remove(scenario.id);
          }
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(scenario.latitude, scenario.longitude),
              iconImage: 'scenario_mystery_thumbnail_$_mapThumbnailVersion',
              iconSize: 0.74,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
              iconAnchor: 'center',
            ),
            {'type': 'scenario', 'id': scenario.id},
          );
          if (!mounted) return;
          _scenarioSymbols.add(symbol);
          _scenarioSymbolById[scenario.id] = symbol;
          _setQuestPoiHighlight(symbol, shouldAnimateQuest);
          _scenarioCircleMystery[scenario.id] = mystery;
          _scenarioQuestObjective[scenario.id] = shouldAnimateQuest;
        } else {
          _setQuestPoiHighlight(existingSymbol, shouldAnimateQuest);
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
          _scenarioQuestObjective[scenario.id] != shouldAnimateQuest) {
        if (existingCircle != null) {
          _setQuestCircleHighlight(existingCircle, false);
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
            circleColor: shouldShowQuestState ? '#e1b12c' : '#5a5560',
            circleStrokeWidth: 2,
            circleStrokeColor: '#ffffff',
          ),
          {'type': 'scenario', 'id': scenario.id},
        );
        if (!mounted) return;
        _scenarioCircles.add(circle);
        _scenarioCircleById[scenario.id] = circle;
        _scenarioCircleMystery[scenario.id] = mystery;
        _scenarioQuestObjective[scenario.id] = shouldAnimateQuest;
        _setQuestCircleHighlight(circle, shouldAnimateQuest);
      } else {
        _setQuestCircleHighlight(existingCircle, shouldAnimateQuest);
      }
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _loadExpositionMysteryThumbnail(MapLibreMapController c) async {
    if (_expositionMysteryThumbnailBytes == null) {
      try {
        _expositionMysteryThumbnailBytes = await loadPoiThumbnail(
          _expositionMapIconImageUrl,
        );
      } catch (_) {}
      _expositionMysteryThumbnailBytes ??= await loadPoiThumbnail(
        _scenarioMysteryImageUrl,
      );
      _expositionMysteryThumbnailBytes ??= await loadPoiThumbnail(
        _legacyMysteryImageUrl,
      );
    }
    if (_expositionMysteryThumbnailBytes != null &&
        !_expositionMysteryThumbnailAdded) {
      try {
        await c.addImage(
          'exposition_mystery_thumbnail_$_mapThumbnailVersion',
          _expositionMysteryThumbnailBytes!,
        );
        _expositionMysteryThumbnailAdded = true;
      } catch (_) {}
    }
  }

  Future<void> _refreshExpositionSymbols() {
    _expositionRefreshSequence = _expositionRefreshSequence.then((_) async {
      try {
        await _refreshExpositionSymbolsNow();
      } catch (e, st) {
        debugPrint('SinglePlayer: _refreshExpositionSymbols error: $e');
        debugPrint('SinglePlayer: _refreshExpositionSymbols stack: $st');
      }
    });
    return _expositionRefreshSequence;
  }

  Future<void> _refreshExpositionSymbolsNow() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    await _loadExpositionMysteryThumbnail(c);
    final visibleExpositions = _shouldSuppressNormalMapPinsForTutorial
        ? const <Exposition>[]
        : _expositions
              .where((exposition) => !_isExpositionRepresentedByPoi(exposition))
              .toList();

    final duplicateOrOrphanSymbols = <Symbol>[];
    for (final symbol in _expositionSymbols.toList()) {
      final id = _expositionIdFromData(symbol.data);
      if (id == null) {
        duplicateOrOrphanSymbols.add(symbol);
        continue;
      }
      final tracked = _expositionSymbolById[id];
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
        _expositionSymbols.remove(symbol);
      }
    }

    final duplicateOrOrphanCircles = <Circle>[];
    for (final circle in _expositionCircles.toList()) {
      final id = _expositionIdFromData(circle.data);
      if (id == null) {
        duplicateOrOrphanCircles.add(circle);
        continue;
      }
      final tracked = _expositionCircleById[id];
      if (tracked == null || !identical(tracked, circle)) {
        duplicateOrOrphanCircles.add(circle);
      }
    }
    if (duplicateOrOrphanCircles.isNotEmpty) {
      for (final circle in duplicateOrOrphanCircles) {
        _setQuestCircleHighlight(circle, false);
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _expositionCircles.remove(circle);
      }
    }

    final desiredIds = visibleExpositions
        .map((exposition) => exposition.id)
        .toSet();
    for (final entry in _expositionSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestPoiHighlight(entry.value, false);
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _expositionSymbols.remove(entry.value);
        _expositionSymbolById.remove(entry.key);
        _expositionQuestObjective.remove(entry.key);
      }
    }
    for (final entry in _expositionCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestCircleHighlight(entry.value, false);
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _expositionCircles.remove(entry.value);
        _expositionCircleById.remove(entry.key);
        _expositionQuestObjective.remove(entry.key);
      }
    }

    final canUseImages =
        _expositionMysteryThumbnailBytes != null &&
        _expositionMysteryThumbnailAdded;

    for (final exposition in visibleExpositions) {
      final isCurrentQuestExposition = _isCurrentQuestExposition(exposition.id);
      final shouldAnimateQuest = _isTrackedQuestExposition(exposition.id);
      final existingSymbol = _expositionSymbolById[exposition.id];
      final existingCircle = _expositionCircleById[exposition.id];
      final needsRefresh =
          existingSymbol == null ||
          _expositionQuestObjective[exposition.id] != shouldAnimateQuest;

      if (canUseImages) {
        if (needsRefresh) {
          if (existingSymbol != null) {
            _setQuestPoiHighlight(existingSymbol, false);
            try {
              await c.removeSymbols([existingSymbol]);
            } catch (_) {}
            _expositionSymbols.remove(existingSymbol);
            _expositionSymbolById.remove(exposition.id);
          }
          if (existingCircle != null) {
            _setQuestCircleHighlight(existingCircle, false);
            try {
              await c.removeCircle(existingCircle);
            } catch (_) {}
            _expositionCircles.remove(existingCircle);
            _expositionCircleById.remove(exposition.id);
          }
          const imageId = 'exposition_mystery_thumbnail_$_mapThumbnailVersion';
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(exposition.latitude, exposition.longitude),
              iconImage: imageId,
              iconSize: 0.74,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
              iconAnchor: 'center',
            ),
            {'type': 'exposition', 'id': exposition.id},
          );
          if (!mounted) return;
          _expositionSymbols.add(symbol);
          _expositionSymbolById[exposition.id] = symbol;
          _setQuestPoiHighlight(symbol, shouldAnimateQuest);
          _expositionQuestObjective[exposition.id] = shouldAnimateQuest;
        } else {
          _setQuestPoiHighlight(existingSymbol, shouldAnimateQuest);
        }
        continue;
      }

      if (existingSymbol != null) {
        _setQuestPoiHighlight(existingSymbol, false);
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _expositionSymbols.remove(existingSymbol);
        _expositionSymbolById.remove(exposition.id);
      }
      if (existingCircle == null ||
          _expositionQuestObjective[exposition.id] != shouldAnimateQuest) {
        if (existingCircle != null) {
          _setQuestCircleHighlight(existingCircle, false);
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _expositionCircles.remove(existingCircle);
          _expositionCircleById.remove(exposition.id);
        }
        final circle = await c.addCircle(
          CircleOptions(
            geometry: LatLng(exposition.latitude, exposition.longitude),
            circleRadius: 22,
            circleOpacity: _mapMarkerStartingOpacity(1.0),
            circleColor: isCurrentQuestExposition ? '#e1b12c' : '#d97706',
            circleStrokeWidth: 2,
            circleStrokeColor: '#ffffff',
          ),
          {'type': 'exposition', 'id': exposition.id},
        );
        if (!mounted) return;
        _expositionCircles.add(circle);
        _expositionCircleById[exposition.id] = circle;
        _expositionQuestObjective[exposition.id] = shouldAnimateQuest;
        _setQuestCircleHighlight(circle, shouldAnimateQuest);
      } else {
        _setQuestCircleHighlight(existingCircle, shouldAnimateQuest);
      }
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _loadMonsterMysteryThumbnail(MapLibreMapController c) async {
    for (final encounterType in const ['monster', 'boss', 'raid']) {
      final imageTypeId = _monsterMysteryImageIdForEncounterType(encounterType);
      if (!_monsterMysteryThumbnailBytesByType.containsKey(imageTypeId)) {
        Uint8List? bytes;
        try {
          bytes = await loadPoiThumbnail(
            _monsterMysteryImageUrlForEncounterType(encounterType),
          );
        } catch (_) {}
        bytes ??= await loadPoiThumbnail(_legacyMysteryImageUrl);
        _monsterMysteryThumbnailBytesByType[imageTypeId] = bytes;
      }
      final mysteryBytes = _monsterMysteryThumbnailBytesByType[imageTypeId];
      if (mysteryBytes != null &&
          !_monsterMysteryThumbnailTypesAdded.contains(imageTypeId)) {
        try {
          await c.addImage(
            'monster_mystery_thumbnail_${imageTypeId}_$_mapThumbnailVersion',
            mysteryBytes,
          );
          _monsterMysteryThumbnailTypesAdded.add(imageTypeId);
        } catch (_) {}
      }
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
    final visibleMonsters = _shouldSuppressNormalMapPinsForTutorial
        ? _monsters
              .where(
                (encounter) =>
                    _tutorialFocusedMonsterEncounterId?.trim().isNotEmpty ==
                        true
                    ? _isTutorialFocusedMonsterEncounterId(encounter.id)
                    : false,
              )
              .toList()
        : _monsters
              .where((encounter) => !_isMonsterRepresentedByPoi(encounter))
              .toList();

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
      for (final symbol in duplicateOrOrphanSymbols) {
        _setQuestPoiHighlight(symbol, false);
      }
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
        _setQuestCircleHighlight(circle, false);
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _monsterCircles.remove(circle);
      }
    }

    final desiredIds = visibleMonsters.map((monster) => monster.id).toSet();
    for (final entry in _monsterSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestPoiHighlight(entry.value, false);
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _monsterSymbols.remove(entry.value);
        _monsterSymbolById.remove(entry.key);
      }
    }
    for (final entry in _monsterCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestCircleHighlight(entry.value, false);
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
      final shouldShowQuestState = isCurrentQuestMonster || isTutorialMonster;
      final shouldAnimateQuest =
          _isTrackedQuestMonster(monster.id) || isTutorialMonster;
      final mystery = _isMonsterMystery(monster);
      String? symbolImageId;
      if (mystery) {
        final mysteryTypeId = _monsterMysteryImageIdForEncounterType(
          monster.encounterType,
        );
        if ((_monsterMysteryThumbnailBytesByType[mysteryTypeId]) != null &&
            _monsterMysteryThumbnailTypesAdded.contains(mysteryTypeId)) {
          symbolImageId =
              'monster_mystery_thumbnail_${mysteryTypeId}_$_mapThumbnailVersion';
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
          _setQuestCircleHighlight(existingCircle, false);
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
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
              iconAnchor: 'center',
              zIndex: shouldShowQuestState ? 4 : 2,
            ),
            {'type': 'monster', 'id': monster.id},
          );
          if (!mounted) return;
          _monsterSymbols.add(symbol);
          _monsterSymbolById[monster.id] = symbol;
          _setQuestPoiHighlight(symbol, shouldAnimateQuest);
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(monster.latitude, monster.longitude),
                iconImage: symbolImageId,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: _transparentMapHaloColor,
                iconHaloWidth: 0.0,
                zIndex: shouldShowQuestState ? 4 : 2,
              ),
            );
            _setQuestPoiHighlight(existingSymbol, shouldAnimateQuest);
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _monsterSymbolById[monster.id];
      if (existingSymbol != null) {
        _setQuestPoiHighlight(existingSymbol, false);
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _monsterSymbols.remove(existingSymbol);
        _monsterSymbolById.remove(monster.id);
      }

      final existingCircle = _monsterCircleById[monster.id];
      if (existingCircle != null) {
        _setQuestCircleHighlight(existingCircle, false);
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
          circleColor: shouldShowQuestState
              ? '#e1b12c'
              : (mystery ? '#5a5560' : _monsterEncounterMarkerColor(monster)),
          circleStrokeWidth: 2,
          circleStrokeColor: '#ffffff',
        ),
        {'type': 'monster', 'id': monster.id},
      );
      if (!mounted) return;
      _monsterCircles.add(circle);
      _monsterCircleById[monster.id] = circle;
      _setQuestCircleHighlight(circle, shouldAnimateQuest);
    }
    await _applyMapMarkerIsolationIfNeeded();
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

  bool _doesQuestNodeTargetMonster(QuestNode? node, String monsterId) {
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
    return (node?.monsterId ?? '').isNotEmpty && monsterId == node!.monsterId;
  }

  bool _isTrackedQuestScenario(String scenarioId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in _trackedAcceptedQuests(questLog)) {
      final node = quest.currentNode;
      if (node?.scenarioId == scenarioId) {
        return true;
      }
    }
    return false;
  }

  bool _isCurrentQuestExposition(String expositionId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node?.expositionId == expositionId) {
        return true;
      }
    }
    return false;
  }

  bool _isTrackedQuestExposition(String expositionId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in _trackedAcceptedQuests(questLog)) {
      final node = quest.currentNode;
      if (node?.expositionId == expositionId) {
        return true;
      }
    }
    return false;
  }

  bool _isCurrentQuestMonster(String monsterId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      if (_doesQuestNodeTargetMonster(quest.currentNode, monsterId)) {
        return true;
      }
    }
    return false;
  }

  bool _isTrackedQuestMonster(String monsterId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in _trackedAcceptedQuests(questLog)) {
      if (_doesQuestNodeTargetMonster(quest.currentNode, monsterId)) {
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

  bool _isTrackedQuestChallenge(String challengeId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in _trackedAcceptedQuests(questLog)) {
      final node = quest.currentNode;
      if (node?.challengeId == challengeId) {
        return true;
      }
    }
    return false;
  }

  bool _isCurrentQuestTurnInCharacter(String characterId) {
    final questLog = context.read<QuestLogProvider>();
    return _currentQuestTurnInCharacterIds(questLog).contains(characterId);
  }

  bool _isTrackedQuestTurnInCharacter(String characterId) {
    final questLog = context.read<QuestLogProvider>();
    return _trackedQuestTurnInCharacterIds(questLog).contains(characterId);
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
    final visibleChallenges = _shouldSuppressNormalMapPinsForTutorial
        ? const <Challenge>[]
        : _challenges;

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
      for (final symbol in duplicateOrOrphanSymbols) {
        _setQuestPoiHighlight(symbol, false);
      }
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
        _setQuestCircleHighlight(circle, false);
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _challengeCircles.remove(circle);
      }
    }

    final standaloneChallenges = visibleChallenges
        .where(
          (challenge) =>
              !_isChallengeRepresentedByPoi(challenge) &&
              !_usesDedicatedQuestChallengeUi(challenge),
        )
        .toList();
    final polygonChallenges = standaloneChallenges
        .where((challenge) => challenge.hasPolygon)
        .toList();
    await _refreshChallengePolygonOverlays(c, polygonChallenges);

    final pointChallenges = standaloneChallenges
        .where((challenge) => !challenge.hasPolygon)
        .toList();
    final desiredIds = pointChallenges.map((challenge) => challenge.id).toSet();
    for (final entry in _challengeSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestPoiHighlight(entry.value, false);
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _challengeSymbols.remove(entry.value);
        _challengeSymbolById.remove(entry.key);
      }
    }
    for (final entry in _challengeCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        _setQuestCircleHighlight(entry.value, false);
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _challengeCircles.remove(entry.value);
        _challengeCircleById.remove(entry.key);
      }
    }

    for (final challenge in pointChallenges) {
      final isCurrentQuestChallenge = _isCurrentQuestChallenge(challenge.id);
      final shouldAnimateQuest = _isTrackedQuestChallenge(challenge.id);
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
          _setQuestCircleHighlight(existingCircle, false);
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _challengeCircles.remove(existingCircle);
          _challengeCircleById.remove(challenge.id);
        }

        final existingSymbol = _challengeSymbolById[challenge.id];
        if (existingSymbol == null) {
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(challenge.latitude, challenge.longitude),
              iconImage: symbolImageId,
              iconSize: 0.74,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
              iconAnchor: 'center',
            ),
            {'type': 'challenge', 'id': challenge.id},
          );
          if (!mounted) return;
          _challengeSymbols.add(symbol);
          _challengeSymbolById[challenge.id] = symbol;
          _setQuestPoiHighlight(symbol, shouldAnimateQuest);
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(challenge.latitude, challenge.longitude),
                iconImage: symbolImageId,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: _transparentMapHaloColor,
                iconHaloWidth: 0.0,
              ),
            );
            _setQuestPoiHighlight(existingSymbol, shouldAnimateQuest);
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _challengeSymbolById[challenge.id];
      if (existingSymbol != null) {
        _setQuestPoiHighlight(existingSymbol, false);
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
        _setQuestCircleHighlight(circle, shouldAnimateQuest);
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
          _setQuestCircleHighlight(existingCircle, shouldAnimateQuest);
        } catch (_) {}
      }
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  String? _scenarioIdFromData(dynamic raw) {
    if (raw == null || raw is! Map) return null;
    final data = Map<String, dynamic>.from(raw);
    if (data['type']?.toString() != 'scenario') return null;
    final id = data['id']?.toString();
    if (id == null || id.isEmpty) return null;
    return id;
  }

  String? _expositionIdFromData(dynamic raw) {
    if (raw == null || raw is! Map) return null;
    final data = Map<String, dynamic>.from(raw);
    if (data['type']?.toString() != 'exposition') return null;
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

  String? _chestIdFromData(dynamic raw) {
    if (raw == null || raw is! Map) return null;
    final data = Map<String, dynamic>.from(raw);
    if (data['type']?.toString() != 'chest') return null;
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
    c.onFeatureTapped.add((point, _, _, _, annotation) {
      unawaited(_handleFeatureTap(point, annotation));
    });
  }

  Future<void> _handleFeatureTap(
    Point<double> point,
    Annotation? annotation,
  ) async {
    final data = _annotationData(annotation);
    if (data == null || data.isEmpty) return;
    final type = data['type']?.toString();
    final idStr = data['id']?.toString();
    if (type == null || idStr == null || idStr.isEmpty) return;
    if (!_isAnnotationTapEnabled(type, idStr)) return;

    _registerFeatureTap(point);

    if (_isSelectablePinType(type)) {
      final handled = await _maybeHandlePinTap(
        point,
        preferredType: type,
        preferredId: idStr,
      );
      if (handled || !mounted) return;
    }

    _openMapPinByTypeAndId(type, idStr);
  }

  Map<String, dynamic>? _annotationData(Annotation? annotation) {
    final raw = switch (annotation) {
      Symbol symbol => symbol.data,
      Circle circle => circle.data,
      Fill fill => fill.data,
      Line line => line.data,
      _ => null,
    };
    if (raw is! Map) return null;
    return Map<String, dynamic>.from(raw);
  }

  void _registerFeatureTap(Point<double> point) {
    _lastFeatureTapAt = DateTime.now();
    _lastFeatureTapPoint = point;
  }

  bool _shouldIgnoreMapClickForRecentFeatureTap(Point<double> point) {
    final lastAt = _lastFeatureTapAt;
    final lastPoint = _lastFeatureTapPoint;
    if (lastAt == null || lastPoint == null) return false;
    if (DateTime.now().difference(lastAt) > const Duration(milliseconds: 250)) {
      return false;
    }
    final dx = point.x - lastPoint.x;
    final dy = point.y - lastPoint.y;
    return math.sqrt(dx * dx + dy * dy) <= 8;
  }

  bool _isZoneScopedContentVisibleForSelectedZone(String zoneId) {
    final normalizedZoneId = zoneId.trim();
    if (normalizedZoneId.isEmpty) return false;
    final selectedZoneId =
        context.read<ZoneProvider>().selectedZone?.id.trim() ?? '';
    if (selectedZoneId.isNotEmpty) {
      return selectedZoneId == normalizedZoneId;
    }
    if (_zones.isEmpty) return false;
    return _zones.first.id == normalizedZoneId;
  }

  bool _poiMatchesActiveTagFilter(PointOfInterest poi) {
    final filters = context.read<QuestFilterProvider>();
    final tags = context.read<TagsProvider>();
    if (!filters.enableTagFilter || tags.selectedTagIds.isEmpty) {
      return true;
    }

    final selectedTagIds = tags.selectedTagIds;
    final selectedTagNames = tags.tags
        .where((tag) => selectedTagIds.contains(tag.id))
        .map((tag) => tag.name.toLowerCase())
        .toSet();
    return poi.tags.any(
      (tag) =>
          selectedTagIds.contains(tag.id) ||
          (tag.name.isNotEmpty &&
              selectedTagNames.contains(tag.name.toLowerCase())),
    );
  }

  bool _isPoiMarkerInteractable(PointOfInterest poi) {
    if (_shouldSuppressNormalMapPinsForTutorial) return false;
    if (!_isPoiInSelectedZone(poi)) return false;
    if (!_poiMatchesActiveTagFilter(poi)) return false;

    final questLog = context.read<QuestLogProvider>();
    final discoveries = context.read<DiscoveriesProvider>();
    final isQuestCurrent = _currentQuestPoiIdsForFilter(
      questLog,
    ).contains(poi.id);
    final undiscovered = !discoveries.hasDiscovered(poi.id);
    return _poiMarkerOpacity(
          poi,
          isQuestCurrent: isQuestCurrent,
          undiscovered: undiscovered,
          mapContentPoiIds: _buildPoiIdsWithMapContent(),
        ) >
        0.05;
  }

  bool _isAnnotationTapEnabled(String type, String id) {
    if (id.isEmpty) return false;
    if (type == 'zone') return true;
    if (_pinBatchRevealInProgress) return false;
    if (_tutorialNormalPinsRevealInProgress && !_isTutorialMapFocusActive) {
      return false;
    }

    final normalizedType = type == 'poiBorder' ? 'poi' : type;
    if (!_isMapMarkerIsolationVisible(normalizedType, id)) {
      return false;
    }

    switch (normalizedType) {
      case 'poi':
        final poi = _poiById(id);
        return poi != null && _isPoiMarkerInteractable(poi);
      case 'character':
        final character = _characterById(id);
        return character != null &&
            _visibleCharacterPoints(character).isNotEmpty;
      case 'chest':
        final chest = _treasureChestById(id);
        return chest != null &&
            chest.openedByUser != true &&
            _isZoneScopedContentVisibleForSelectedZone(chest.zoneId);
      case 'healingFountain':
        final fountain = _healingFountainById(id);
        return fountain != null &&
            _isZoneScopedContentVisibleForSelectedZone(fountain.zoneId);
      case 'resource':
        final resource = _resourceById(id);
        return resource != null &&
            !resource.gatheredByUser &&
            _isZoneScopedContentVisibleForSelectedZone(resource.zoneId);
      case 'base':
        return _baseById(id) != null;
      case 'scenario':
        final scenario = _scenarioById(id);
        return scenario != null &&
            !_isScenarioRepresentedByPoi(scenario) &&
            _isZoneScopedContentVisibleForSelectedZone(scenario.zoneId);
      case 'exposition':
        final exposition = _expositionById(id);
        return exposition != null &&
            !_isExpositionRepresentedByPoi(exposition) &&
            _isZoneScopedContentVisibleForSelectedZone(exposition.zoneId);
      case 'monster':
        final monster = _monsterById(id);
        return monster != null &&
            !_isMonsterRepresentedByPoi(monster) &&
            _isZoneScopedContentVisibleForSelectedZone(monster.zoneId);
      case 'challenge':
        final challenge = _challengeById(id);
        return challenge != null &&
            !_isChallengeRepresentedByPoi(challenge) &&
            !_usesDedicatedQuestChallengeUi(challenge) &&
            _isZoneScopedContentVisibleForSelectedZone(challenge.zoneId);
      default:
        return false;
    }
  }

  bool _isSelectablePinType(String type) {
    switch (type) {
      case 'poi':
      case 'poiBorder':
      case 'character':
      case 'chest':
      case 'healingFountain':
      case 'resource':
      case 'base':
      case 'scenario':
      case 'monster':
      case 'challenge':
        return true;
      default:
        return false;
    }
  }

  Future<bool> _maybeHandlePinTap(
    Point<double> point, {
    String? preferredType,
    String? preferredId,
  }) async {
    final candidates = await _pinSelectionCandidatesForPoint(
      point,
      preferredType: preferredType,
      preferredId: preferredId,
    );
    return _openPinSelectionCandidates(candidates);
  }

  List<_MapPinAnnotationSeed> _selectablePinAnnotationSeeds() {
    final annotations = <_MapPinAnnotationSeed>[];

    void addSymbolSeeds(Iterable<Symbol> symbols) {
      for (final symbol in symbols) {
        final data = _annotationData(symbol);
        if (data == null) continue;
        final type = data['type']?.toString();
        final id = data['id']?.toString();
        final geometry = symbol.options.geometry;
        final opacity = symbol.options.iconOpacity ?? 1.0;
        if (type == null ||
            id == null ||
            id.isEmpty ||
            !_isSelectablePinType(type) ||
            geometry == null ||
            opacity <= 0.05 ||
            !_isAnnotationTapEnabled(type, id)) {
          continue;
        }
        annotations.add(
          _MapPinAnnotationSeed(
            type: type,
            id: id,
            geometry: geometry,
            hitRadiusPx: _pinSelectionHitRadiusPx,
          ),
        );
      }
    }

    void addCircleSeeds(Iterable<Circle> circles) {
      for (final circle in circles) {
        final data = _annotationData(circle);
        if (data == null) continue;
        final type = data['type']?.toString();
        final id = data['id']?.toString();
        final geometry = circle.options.geometry;
        final opacity = circle.options.circleOpacity ?? 1.0;
        if (type == null ||
            id == null ||
            id.isEmpty ||
            !_isSelectablePinType(type) ||
            geometry == null ||
            opacity <= 0.05 ||
            !_isAnnotationTapEnabled(type, id)) {
          continue;
        }
        annotations.add(
          _MapPinAnnotationSeed(
            type: type,
            id: id,
            geometry: geometry,
            hitRadiusPx: math.max(
              _pinSelectionHitRadiusPx,
              (circle.options.circleRadius ?? 0) + 10,
            ),
          ),
        );
      }
    }

    addSymbolSeeds(_poiSymbols);
    addSymbolSeeds(_characterSymbols);
    addSymbolSeeds(_chestSymbols);
    addSymbolSeeds(_healingFountainSymbols);
    addSymbolSeeds(_resourceSymbols);
    addSymbolSeeds(_baseSymbols);
    addSymbolSeeds(_scenarioSymbols);
    addSymbolSeeds(_expositionSymbols);
    addSymbolSeeds(_monsterSymbols);
    addSymbolSeeds(_challengeSymbols);

    addCircleSeeds(_chestCircles);
    addCircleSeeds(_healingFountainCircles);
    addCircleSeeds(_resourceCircles);
    addCircleSeeds(_baseCircles);
    addCircleSeeds(_scenarioCircles);
    addCircleSeeds(_expositionCircles);
    addCircleSeeds(_monsterCircles);
    addCircleSeeds(_challengeCircles);

    return annotations;
  }

  Future<List<_MapPinSelectionCandidate>> _pinSelectionCandidatesForPoint(
    Point<double> point, {
    String? preferredType,
    String? preferredId,
  }) async {
    final controller = _mapController;
    if (controller == null || !_styleLoaded) return const [];

    final annotations = _selectablePinAnnotationSeeds();
    if (annotations.isEmpty) return const [];

    final screenPoints = await controller.toScreenLocationBatch(
      annotations.map((annotation) => annotation.geometry),
    );
    if (!mounted) return const [];

    final bestByKey = <String, _MapPinSelectionCandidate>{};
    for (var i = 0; i < annotations.length && i < screenPoints.length; i++) {
      final annotation = annotations[i];
      final screenPoint = screenPoints[i];
      final dx = screenPoint.x.toDouble() - point.x;
      final dy = screenPoint.y.toDouble() - point.y;
      final distance = math.sqrt(dx * dx + dy * dy);
      if (distance > annotation.hitRadiusPx) continue;

      final candidate = _buildPinSelectionCandidate(
        annotation.type,
        annotation.id,
        distance,
      );
      if (candidate == null) continue;

      final key = '${candidate.type}:${candidate.id}';
      final existing = bestByKey[key];
      if (existing == null || candidate.distance < existing.distance) {
        bestByKey[key] = candidate;
      }
    }

    return _sortPinSelectionCandidates(
      bestByKey.values.toList(),
      preferredType: preferredType,
      preferredId: preferredId,
    );
  }

  List<_MapPinSelectionCandidate> _pinSelectionCandidatesNearLocation(
    AppLocation location, {
    String? preferredType,
    String? preferredId,
  }) {
    if (!_styleLoaded) return const [];

    final lat = location.latitude;
    final lng = location.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180) {
      return const [];
    }

    final annotations = _selectablePinAnnotationSeeds();
    if (annotations.isEmpty) return const [];

    final maxDistanceMeters = _playerUnderfootSelectionRadiusMeters(location);
    final bestByKey = <String, _MapPinSelectionCandidate>{};
    for (final annotation in annotations) {
      final distance = _distanceMeters(
        lat,
        lng,
        annotation.geometry.latitude,
        annotation.geometry.longitude,
      );
      if (distance > maxDistanceMeters) continue;

      final candidate = _buildPinSelectionCandidate(
        annotation.type,
        annotation.id,
        distance,
      );
      if (candidate == null) continue;

      final key = '${candidate.type}:${candidate.id}';
      final existing = bestByKey[key];
      if (existing == null || candidate.distance < existing.distance) {
        bestByKey[key] = candidate;
      }
    }

    return _sortPinSelectionCandidates(
      bestByKey.values.toList(),
      preferredType: preferredType,
      preferredId: preferredId,
    );
  }

  List<_MapPinSelectionCandidate> _sortPinSelectionCandidates(
    List<_MapPinSelectionCandidate> candidates, {
    String? preferredType,
    String? preferredId,
  }) {
    candidates.sort((a, b) {
      final aPreferred = a.type == preferredType && a.id == preferredId;
      final bPreferred = b.type == preferredType && b.id == preferredId;
      if (aPreferred != bPreferred) return aPreferred ? -1 : 1;
      final distanceCompare = a.distance.compareTo(b.distance);
      if (distanceCompare != 0) return distanceCompare;
      return a.title.toLowerCase().compareTo(b.title.toLowerCase());
    });
    return candidates;
  }

  double _playerUnderfootSelectionRadiusMeters(AppLocation location) {
    final accuracy = location.accuracy;
    if (!accuracy.isFinite || accuracy <= 0) {
      return _playerUnderfootPinDistanceMeters;
    }
    return math.max(
      _playerUnderfootPinDistanceMeters,
      math.min(accuracy, _playerUnderfootPinAccuracyCapMeters),
    );
  }

  bool _openPinSelectionCandidates(List<_MapPinSelectionCandidate> candidates) {
    if (!mounted || candidates.isEmpty) return false;
    if (candidates.length == 1) {
      _openMapPinByTypeAndId(candidates.first.type, candidates.first.id);
      return true;
    }
    _showPinSelectionSheet(candidates);
    return true;
  }

  Future<bool> _maybeHandlePlayerUnderfootTap(Point<double> point) async {
    final controller = _mapController;
    final location = context.read<LocationProvider>().location;
    if (controller == null || !_styleLoaded || location == null) return false;

    final lat = location.latitude;
    final lng = location.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180) {
      return false;
    }

    final nearbyCandidates = _pinSelectionCandidatesNearLocation(location);
    if (nearbyCandidates.isEmpty) return false;

    try {
      final playerPoint = await controller.toScreenLocation(LatLng(lat, lng));
      if (!mounted) return false;
      final dx = (point.x - playerPoint.x.toDouble()).abs();
      final dy = point.y - playerPoint.y.toDouble();
      final tappedPlayerMarker =
          dx <= _playerUnderfootTapHalfWidthPx &&
          dy <= _playerUnderfootTapBottomReachPx &&
          dy >= -_playerUnderfootTapTopReachPx;
      if (!tappedPlayerMarker) return false;
      return _openPinSelectionCandidates(nearbyCandidates);
    } catch (_) {
      return false;
    }
  }

  _MapPinSelectionCandidate? _buildPinSelectionCandidate(
    String type,
    String id,
    double distance,
  ) {
    switch (type) {
      case 'poi':
      case 'poiBorder':
        final poi = _poiById(id);
        if (poi == null) return null;
        final hasDiscovered = context.read<DiscoveriesProvider>().hasDiscovered(
          poi.id,
        );
        return _MapPinSelectionCandidate(
          type: 'poi',
          id: poi.id,
          title: poi.name.trim().isNotEmpty
              ? poi.name.trim()
              : 'Point of Interest',
          imageUrl: _poiSelectionImageUrl(poi, hasDiscovered: hasDiscovered),
          distance: distance,
        );
      case 'character':
        final character = _characterById(id);
        if (character == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: character.name.trim().isNotEmpty
              ? character.name.trim()
              : 'Character',
          imageUrl: _characterSelectionImageUrl(character),
          distance: distance,
        );
      case 'chest':
        final chest = _treasureChestById(id);
        if (chest == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: 'Treasure Chest',
          imageUrl: _chestImageUrl,
          distance: distance,
        );
      case 'healingFountain':
        final fountain = _healingFountainById(id);
        if (fountain == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: fountain.name.trim().isNotEmpty
              ? fountain.name.trim()
              : 'Healing Fountain',
          imageUrl: _healingFountainImageUrl(fountain),
          distance: distance,
        );
      case 'resource':
        final resource = _resourceById(id);
        if (resource == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: _resourceSelectionTitle(resource),
          imageUrl: _resourceImageUrl(resource),
          distance: distance,
        );
      case 'base':
        final base = _baseById(id);
        if (base == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: base.name.trim().isNotEmpty ? base.name.trim() : 'Base',
          imageUrl: _baseSelectionImageUrl(base),
          distance: distance,
          useBaseDiamondMarker: true,
          isCurrentUserBase: _isCurrentUserBase(base),
        );
      case 'scenario':
        final scenario = _scenarioById(id);
        if (scenario == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: _compactSelectionLabel(scenario.prompt, fallback: 'Scenario'),
          imageUrl: _scenarioSelectionImageUrl(scenario),
          distance: distance,
        );
      case 'exposition':
        final exposition = _expositionById(id);
        if (exposition == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: _compactSelectionLabel(
            exposition.title,
            fallback: 'Exposition',
          ),
          imageUrl: _expositionSelectionImageUrl(exposition),
          distance: distance,
        );
      case 'monster':
        final monster = _monsterById(id);
        if (monster == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: monster.name.trim().isNotEmpty
              ? monster.name.trim()
              : 'Monster',
          imageUrl: _monsterSelectionImageUrl(monster),
          distance: distance,
        );
      case 'challenge':
        final challenge = _challengeById(id);
        if (challenge == null) return null;
        return _MapPinSelectionCandidate(
          type: type,
          id: id,
          title: _compactSelectionLabel(
            challenge.question,
            fallback: 'Challenge',
          ),
          imageUrl: _challengeSelectionImageUrl(challenge),
          distance: distance,
        );
      default:
        return null;
    }
  }

  String _compactSelectionLabel(String value, {required String fallback}) {
    final compact = value.replaceAll(RegExp(r'\s+'), ' ').trim();
    if (compact.isEmpty) return fallback;
    if (compact.length <= 72) return compact;
    return '${compact.substring(0, 69).trimRight()}...';
  }

  String _poiSelectionImageUrl(
    PointOfInterest poi, {
    required bool hasDiscovered,
  }) {
    if (!hasDiscovered) return _healingFountainFallbackImageUrl;
    return _poiThumbnailSourceUrl(poi) ?? _healingFountainFallbackImageUrl;
  }

  String _characterSelectionImageUrl(Character character) {
    final mapIcon = character.mapIconUrl?.trim() ?? '';
    if (mapIcon.isNotEmpty) return mapIcon;
    final thumbnail = character.thumbnailUrl?.trim() ?? '';
    if (thumbnail.isNotEmpty) return thumbnail;
    final dialogue = character.dialogueImageUrl?.trim() ?? '';
    if (dialogue.isNotEmpty) return dialogue;
    return _characterMysteryImageUrl;
  }

  String _baseSelectionImageUrl(BasePin base) {
    if (base.imageUrl.trim().isNotEmpty) return base.imageUrl.trim();
    if (base.thumbnailUrl.trim().isNotEmpty) return base.thumbnailUrl.trim();
    return _baseDiscoveredImageUrl;
  }

  String _scenarioSelectionImageUrl(Scenario scenario) {
    if (_isScenarioMystery(scenario)) {
      return _scenarioMysteryImageUrl;
    }
    if (scenario.thumbnailUrl.trim().isNotEmpty) {
      return scenario.thumbnailUrl.trim();
    }
    if (scenario.imageUrl.trim().isNotEmpty) return scenario.imageUrl.trim();
    return _scenarioMysteryImageUrl;
  }

  String _expositionSelectionImageUrl(Exposition exposition) {
    if (exposition.thumbnailUrl.trim().isNotEmpty) {
      return exposition.thumbnailUrl.trim();
    }
    if (exposition.imageUrl.trim().isNotEmpty) {
      return exposition.imageUrl.trim();
    }
    return _expositionMapIconImageUrl;
  }

  String _monsterSelectionImageUrl(MonsterEncounter monster) {
    if (_isMonsterMystery(monster)) {
      return _monsterMysteryImageUrlForEncounterType(monster.encounterType);
    }
    if (monster.thumbnailUrl.trim().isNotEmpty) {
      return monster.thumbnailUrl.trim();
    }
    if (monster.imageUrl.trim().isNotEmpty) return monster.imageUrl.trim();
    return _monsterMysteryImageUrlForEncounterType(monster.encounterType);
  }

  String _challengeSelectionImageUrl(Challenge challenge) {
    if (challenge.thumbnailUrl.trim().isNotEmpty) {
      return challenge.thumbnailUrl.trim();
    }
    if (challenge.imageUrl.trim().isNotEmpty) return challenge.imageUrl.trim();
    return _challengeMysteryImageUrl;
  }

  void _showPinSelectionSheet(List<_MapPinSelectionCandidate> candidates) {
    showDialog<void>(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) {
        return Dialog(
          backgroundColor: Theme.of(dialogContext).colorScheme.surface,
          insetPadding: const EdgeInsets.symmetric(
            horizontal: 24,
            vertical: 24,
          ),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(24, 22, 24, 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Pick one',
                  style: Theme.of(dialogContext).textTheme.titleLarge,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 20),
                Wrap(
                  alignment: WrapAlignment.center,
                  runAlignment: WrapAlignment.center,
                  spacing: 14,
                  runSpacing: 14,
                  children: [
                    for (final candidate in candidates)
                      Tooltip(
                        message: candidate.title,
                        child: Material(
                          color: Colors.transparent,
                          child: InkWell(
                            borderRadius: BorderRadius.circular(18),
                            onTap: () {
                              Navigator.of(dialogContext).pop();
                              WidgetsBinding.instance.addPostFrameCallback((_) {
                                if (!mounted) return;
                                _openMapPinByTypeAndId(
                                  candidate.type,
                                  candidate.id,
                                );
                              });
                            },
                            child: Ink(
                              width: 76,
                              height: 76,
                              decoration: BoxDecoration(
                                color: Theme.of(dialogContext)
                                    .colorScheme
                                    .surfaceContainerHighest
                                    .withValues(alpha: 0.55),
                                borderRadius: BorderRadius.circular(18),
                                border: Border.all(
                                  color: Theme.of(
                                    dialogContext,
                                  ).colorScheme.outlineVariant,
                                ),
                              ),
                              child: _buildPinSelectionPreview(
                                dialogContext,
                                candidate,
                              ),
                            ),
                          ),
                        ),
                      ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildPinSelectionPreview(
    BuildContext context,
    _MapPinSelectionCandidate candidate,
  ) {
    if (candidate.useBaseDiamondMarker) {
      return FutureBuilder<Uint8List?>(
        future: loadBaseDiamondMarker(
          isCurrentUserBase: candidate.isCurrentUserBase,
        ),
        builder: (context, snapshot) {
          final bytes = snapshot.data;
          if (bytes == null || bytes.isEmpty) {
            return Container(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              alignment: Alignment.center,
              padding: const EdgeInsets.all(10),
              child: const Icon(Icons.home_rounded, size: 28),
            );
          }
          return Padding(
            padding: const EdgeInsets.all(6),
            child: Image.memory(
              bytes,
              fit: BoxFit.contain,
              filterQuality: FilterQuality.none,
            ),
          );
        },
      );
    }

    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: Image.network(
        candidate.imageUrl,
        fit: BoxFit.cover,
        errorBuilder: (context, error, stackTrace) {
          return Container(
            color: Theme.of(context).colorScheme.surfaceContainerHighest,
            alignment: Alignment.center,
            child: Text(
              candidate.title.isNotEmpty
                  ? candidate.title.trimLeft().substring(0, 1).toUpperCase()
                  : '?',
              style: Theme.of(context).textTheme.titleLarge,
            ),
          );
        },
      ),
    );
  }

  void _openMapPinByTypeAndId(String type, String idStr) {
    if (!mounted || idStr.isEmpty) return;
    if (type == 'zone') {
      _selectZoneById(idStr);
      return;
    }
    if (type == 'character') {
      final character = _characterById(idStr);
      if (character != null) {
        _showCharacterPanel(character);
      }
      return;
    }
    if (type == 'chest') {
      final chest = _treasureChestById(idStr);
      if (chest != null) {
        _showTreasureChestPanel(chest);
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
    if (type == 'resource') {
      final resource = _resourceById(idStr);
      if (resource != null) {
        _showResourcePanel(resource);
      }
      return;
    }
    if (type == 'base') {
      final base = _baseById(idStr);
      if (base != null) {
        _showBasePanel(base);
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
    if (type == 'exposition') {
      final exposition = _expositionById(idStr);
      if (exposition != null) {
        final activeQuestEntry = _activeQuestNodeForExposition(exposition.id);
        unawaited(
          _showExpositionDialogue(
            exposition,
            quest: activeQuestEntry?.key,
            node: activeQuestEntry?.value,
          ),
        );
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
    if (type != 'poi' && type != 'poiBorder') return;
    final poi = _poiById(idStr);
    if (poi == null) return;
    _showPointOfInterestPanel(
      poi,
      context.read<DiscoveriesProvider>().hasDiscovered(idStr),
    );
  }

  void _showUndiscoveredZoneFeedback() {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Enter this zone to discover it.')),
    );
  }

  void _upsertZone(Zone zone) {
    context.read<ZoneProvider>().upsertZone(zone);
    if (!mounted) return;
    setState(() {
      final zones = List<Zone>.from(_zones);
      final existingIndex = zones.indexWhere((entry) => entry.id == zone.id);
      if (existingIndex >= 0) {
        zones[existingIndex] = zone;
      } else {
        zones.add(zone);
      }
      _zones = zones;
    });
  }

  Zone _zoneFromDiscoveryResponse(
    Map<String, dynamic> response,
    Zone fallback,
  ) {
    final rawZone = response['zone'];
    if (rawZone is Map<String, dynamic>) {
      return Zone.fromJson(rawZone);
    }
    if (rawZone is Map) {
      return Zone.fromJson(Map<String, dynamic>.from(rawZone));
    }
    return fallback.copyWith(discovered: true);
  }

  Map<String, dynamic>? _zoneDiscoveryRewardModalData(
    Zone zone,
    Map<String, dynamic> response,
  ) {
    final rewardExperience =
        (response['rewardExperience'] as num?)?.toInt() ?? 0;
    final rewardGold =
        (response['rewardGold'] as num?)?.toInt() ??
        (response['goldAwarded'] as num?)?.toInt() ??
        0;
    final itemsAwarded =
        (response['itemsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((entry) => Map<String, dynamic>.from(entry))
            .toList() ??
        const <Map<String, dynamic>>[];
    final spellsAwarded =
        (response['spellsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((entry) => Map<String, dynamic>.from(entry))
            .toList() ??
        const <Map<String, dynamic>>[];

    if (rewardExperience <= 0 &&
        rewardGold <= 0 &&
        itemsAwarded.isEmpty &&
        spellsAwarded.isEmpty) {
      return null;
    }

    return {
      'zoneName': zone.name,
      'rewardExperience': rewardExperience,
      'rewardGold': rewardGold,
      'goldAwarded': rewardGold,
      'itemsAwarded': itemsAwarded,
      'spellsAwarded': spellsAwarded,
    };
  }

  Future<void> _presentZoneDiscoveryRewards(
    Zone zone,
    Map<String, dynamic> response,
  ) async {
    if (response['alreadyDiscovered'] == true) return;
    final statsProvider = context.read<CharacterStatsProvider>();
    final previousLevel = statsProvider.level;
    await Future.wait([
      context.read<AuthProvider>().refresh(),
      statsProvider.refresh(silent: true),
      context.read<UserLevelProvider>().refresh(),
      context.read<ActivityFeedProvider>().refresh(),
    ]);
    if (!mounted) return;
    final modalData = _zoneDiscoveryRewardModalData(zone, response);
    final completedTaskProvider = context.read<CompletedTaskProvider>();
    if (modalData != null) {
      completedTaskProvider.showModal('zoneDiscovered', data: modalData);
    }
    completedTaskProvider.queueLevelUpFollowUpIfNeeded(
      previousLevel: previousLevel,
      currentLevel: statsProvider.level,
    );
  }

  Future<void> _discoverZoneForLocation(Zone zone, AppLocation location) async {
    final zoneId = zone.id.trim();
    if (zoneId.isEmpty || zone.discovered) return;
    if (!_zoneDiscoveryInFlightIds.add(zoneId)) return;

    try {
      final response = await context.read<PoiService>().discoverZone(
        zoneId,
        lat: location.latitude,
        lng: location.longitude,
      );
      if (!mounted) return;

      final updatedZone = _zoneFromDiscoveryResponse(response, zone);
      _upsertZone(updatedZone);
      await _presentZoneDiscoveryRewards(updatedZone, response);
      if (!mounted || !updatedZone.discovered) return;

      final currentLocation = context.read<LocationProvider>().location;
      if (currentLocation == null) return;
      final currentZone = context.read<ZoneProvider>().findZoneAtCoordinate(
        currentLocation.latitude,
        currentLocation.longitude,
      );
      if (currentZone?.id == updatedZone.id) {
        context.read<ZoneProvider>().setSelectedZone(updatedZone);
      }
    } on DioException catch (error) {
      debugPrint('SinglePlayer: zone discovery failed: $error');
      debugPrint(
        'SinglePlayer: zone discovery response=${error.response?.data}',
      );
    } catch (error, stackTrace) {
      debugPrint('SinglePlayer: zone discovery error: $error');
      debugPrint('SinglePlayer: zone discovery stack: $stackTrace');
    } finally {
      _zoneDiscoveryInFlightIds.remove(zoneId);
    }
  }

  void _selectZone(Zone? zone, {bool showUndiscoveredFeedback = false}) {
    if (zone != null && !zone.discovered) {
      if (showUndiscoveredFeedback) {
        _showUndiscoveredZoneFeedback();
      }
      return;
    }
    final zoneProvider = context.read<ZoneProvider>();
    zoneProvider.setSelectedZone(zone, manual: true);
    if (zone == null) {
      zoneProvider.unlockSelection();
      _zoneWidgetController.close();
    }
  }

  void _selectZoneById(String zoneId) {
    final zone = _zones.firstWhere(
      (z) => z.id == zoneId,
      orElse: () => const Zone(id: '', name: '', latitude: 0, longitude: 0),
    );
    if (zone.id.isEmpty) return;
    _selectZone(zone);
  }

  bool _selectZoneIfDifferent(Zone? zone) {
    final targetZoneId = zone?.id.trim() ?? '';
    if (targetZoneId.isEmpty) return false;
    if (zone?.discovered != true) return false;
    final currentZoneId = _effectiveSelectedZone()?.id.trim() ?? '';
    if (currentZoneId == targetZoneId) return false;
    _selectZone(zone);
    return true;
  }

  bool _selectZoneByIdIfDifferent(String? zoneId) {
    final normalizedZoneId = zoneId?.trim() ?? '';
    if (normalizedZoneId.isEmpty) return false;
    final zone = _zones.firstWhere(
      (z) => z.id == normalizedZoneId,
      orElse: () => const Zone(id: '', name: '', latitude: 0, longitude: 0),
    );
    if (zone.id.isEmpty) return false;
    return _selectZoneIfDifferent(zone);
  }

  bool _selectZoneForCoordinatesIfDifferent(double lat, double lng) {
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180) {
      return false;
    }
    final zone = context.read<ZoneProvider>().findZoneAtCoordinate(lat, lng);
    return _selectZoneIfDifferent(zone);
  }

  bool _selectZoneForPoiIfDifferent(PointOfInterest poi) {
    final lat = double.tryParse(poi.lat);
    final lng = double.tryParse(poi.lng);
    if (lat == null || lng == null) return false;
    return _selectZoneForCoordinatesIfDifferent(lat, lng);
  }

  void _handleMapClick(Point<double> point, LatLng coordinates) {
    if (_isPlacingBase) {
      setState(() {
        _pendingBaseSelection = coordinates;
      });
      unawaited(_refreshBasePlacementPreview());
      return;
    }
    unawaited(_handleMapClickAsync(point, coordinates));
  }

  Future<void> _handleMapClickAsync(
    Point<double> point,
    LatLng coordinates,
  ) async {
    if (_shouldIgnoreMapClickForRecentFeatureTap(point)) return;
    final handledPinTap = await _maybeHandlePinTap(point);
    if (handledPinTap || !mounted) return;
    final handledUnderfootTap = await _maybeHandlePlayerUnderfootTap(point);
    if (handledUnderfootTap || !mounted) return;
    final zone = context.read<ZoneProvider>().findZoneAtCoordinate(
      coordinates.latitude,
      coordinates.longitude,
    );
    _selectZone(zone, showUndiscoveredFeedback: zone != null);
  }

  void _onBasePlacementRequested() {
    final request = _basePlacementProvider?.pendingRequest;
    if (request == null || !mounted) return;
    _basePlacementProvider?.clearRequest();
    unawaited(
      _beginBasePlacement(request.ownedInventoryItem, request.inventoryItem),
    );
  }

  Future<void> _showPointOfInterestPanel(
    PointOfInterest poi,
    bool hasDiscovered,
  ) {
    Quest? questForPoi;
    QuestNode? nodeForPoi;
    final linkedScenarios = _linkedScenariosForPoi(poi.id);
    final linkedChallenges = _linkedChallengesForPoi(poi.id);
    final linkedExpositions = _linkedExpositionsForPoi(poi.id);
    final linkedMonsters = _linkedMonstersForPoi(poi.id);
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
    final objectiveQuest = questForPoi;
    final objectiveNode = nodeForPoi;

    return showModalBottomSheet(
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
        linkedScenarios: linkedScenarios,
        linkedChallenges: linkedChallenges,
        linkedExpositions: linkedExpositions,
        linkedMonsters: linkedMonsters,
        onClose: () => Navigator.of(context).pop(),
        onQuestObjectiveTap: objectiveQuest != null && objectiveNode != null
            ? () {
                Navigator.of(context).pop();
                WidgetsBinding.instance.addPostFrameCallback((_) {
                  if (!mounted) return;
                  _showQuestObjectiveSubmissionPanel(
                    objectiveQuest,
                    objectiveNode,
                  );
                });
              }
            : null,
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
        onScenarioTap: (scenario) {
          Navigator.of(context).pop();
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            _showScenarioPanel(scenario);
          });
        },
        onExpositionTap: (exposition) {
          Navigator.of(context).pop();
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            final activeQuestEntry = _activeQuestNodeForExposition(
              exposition.id,
            );
            if (activeQuestEntry != null) {
              unawaited(
                _showExpositionDialogue(
                  exposition,
                  quest: activeQuestEntry.key,
                  node: activeQuestEntry.value,
                ),
              );
            } else {
              unawaited(_showExpositionDialogue(exposition));
            }
          });
        },
        onMonsterTap: (encounter) {
          Navigator.of(context).pop();
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            _showMonsterPanel(encounter);
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
        onDiscoveryCelebrationsComplete: () {
          _runAfterCompletionModals(() {
            unawaited(_showPointOfInterestPanel(poi, true));
          });
        },
      ),
    );
  }

  Future<void> _refreshDiscoveredPoiMarkers() async {
    if (!mounted) return;
    if (!_styleLoaded || _mapController == null || !_markersAdded) return;
    final discoveries = context.read<DiscoveriesProvider>();
    if (discoveries.discoveredPoiIds.isEmpty) return;
    final questLog = context.read<QuestLogProvider>();
    final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
    for (final poiId in discoveries.discoveredPoiIds) {
      await _updatePoiSymbolForQuestState(
        poiId,
        isQuestCurrent: questPoiIds.contains(poiId),
      );
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
    final suppressNormalPins = _shouldSuppressNormalMapPinsForTutorial;
    final deferPinReveal =
        !suppressNormalPins && _tutorialNormalPinsRevealInProgress;
    _pinBatchRevealInProgress = deferPinReveal;
    debugPrint(
      'SinglePlayer: _addPoiMarkers start (pois=${_pois.length} chars=${_characters.length} chests=${_treasureChests.length} fountains=${_healingFountains.length} resources=${_resources.length} scenarios=${_scenarios.length} monsters=${_monsters.length} challenges=${_challenges.length})',
    );

    try {
      final questLog = context.read<QuestLogProvider>();
      final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
      final trackedQuestPoiIds = _trackedQuestPoiIdsForPulse(questLog);
      final featuredMainStoryPulseTarget = _featuredMainStoryPulseTarget(
        questLog,
      );
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
      _questPoiHighlightSymbols.clear();
      _questPoiHighlightCircles.clear();
      _mainStoryPoiHighlightSymbols.clear();
      _mainStoryPoiHighlightCircles.clear();
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
      if (_resourceSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_resourceSymbols);
        } catch (_) {}
        if (!mounted) return;
        _resourceSymbols.clear();
      }
      _resourceSymbolById.clear();
      if (_resourceCircles.isNotEmpty) {
        for (final circle in _resourceCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _resourceCircles.clear();
      }
      _resourceCircleById.clear();
      if (_baseSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_baseSymbols);
        } catch (_) {}
        if (!mounted) return;
        _baseSymbols.clear();
      }
      _baseSymbolById.clear();
      if (_baseCircles.isNotEmpty) {
        for (final circle in _baseCircles) {
          try {
            await c.removeCircle(circle);
          } catch (_) {}
        }
        if (!mounted) return;
        _baseCircles.clear();
      }
      _baseCircleById.clear();
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
      final mainStoryPlaceholderFuture = loadPoiThumbnailWithMainStoryMarker(
        null,
      ).catchError((_) => null);
      final characterPlaceholderFuture = loadPoiThumbnail(
        _characterMysteryImageUrl,
      ).catchError((_) => null);
      final characterAvailablePlaceholderFuture =
          loadPoiThumbnailWithQuestMarker(
            _characterMysteryImageUrl,
          ).catchError((_) => null);
      final characterMainStoryPlaceholderFuture =
          loadPoiThumbnailWithMainStoryMarker(
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
      final mainStoryPlaceholderBytes = await mainStoryPlaceholderFuture;
      if (mainStoryPlaceholderBytes != null) {
        try {
          await c.addImage(
            'poi_placeholder_main_story_$_mapThumbnailVersion',
            mainStoryPlaceholderBytes,
          );
        } catch (_) {}
      }
      final characterPlaceholderBytes = await characterPlaceholderFuture;
      final characterAvailablePlaceholderBytes =
          await characterAvailablePlaceholderFuture;
      final characterMainStoryPlaceholderBytes =
          await characterMainStoryPlaceholderFuture;

      final chestBytes = await chestFuture;
      if (chestBytes != null) {
        _chestThumbnailBytes = chestBytes;
        _chestThumbnailAdded = true;
        try {
          await c.addImage('chest_thumbnail_$_mapThumbnailVersion', chestBytes);
        } catch (_) {}
      }

      if (suppressNormalPins) {
        _pinBatchRevealInProgress = false;
        await _refreshTreasureChestSymbols();
        await _refreshHealingFountainSymbols();
        await _refreshResourceSymbols();
        await _refreshBaseSymbols();
        await _refreshScenarioSymbols();
        await _refreshExpositionSymbols();
        await _refreshMonsterSymbols();
        await _refreshChallengeSymbols();
        _ensureQuestPoiPulseTimer();
        debugPrint('SinglePlayer: _addPoiMarkers done (tutorial-suppressed)');
        return;
      }

      final discoveries = context.read<DiscoveriesProvider>();
      final hadEmptyDiscoveries = discoveries.discoveries.isEmpty;
      final poiImageUpdates = <_PoiImageUpdate>[];
      final poiSymbolRequests = <_PoiSymbolRequest>[];
      final mapContentPoiIds = _buildPoiIdsWithMapContent();
      final questMarkerPoiIds = _buildPoiIdsWithQuestMarkerContent();
      for (final poi in _pois) {
        if (!_isPoiInSelectedZone(poi)) continue;
        final lat = double.tryParse(poi.lat) ?? 0.0;
        final lng = double.tryParse(poi.lng) ?? 0.0;
        final useRealImage = discoveries.hasDiscovered(poi.id);
        final undiscovered = !useRealImage;
        final isQuestCurrent = questPoiIds.contains(poi.id);
        final hasMainStoryAccent = _poiHasMainStoryAccent(poi, questLog);
        final shouldPulseLikeQuest = _shouldPulseQuestPoi(
          poi,
          isQuestCurrent: isQuestCurrent,
          trackedQuestPoiIds: trackedQuestPoiIds,
        );
        final shouldPulseLikeMainStory =
            featuredMainStoryPulseTarget?.poiId == poi.id;
        final hasMapContent = _poiHasMapContent(
          poi,
          isQuestCurrent: isQuestCurrent,
          mapContentPoiIds: mapContentPoiIds,
        );
        final hasQuestMarkerContent = _poiHasQuestMarkerContent(
          poi,
          isQuestCurrent: isQuestCurrent,
          questMarkerPoiIds: questMarkerPoiIds,
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
        String? symbolImageId;
        if (hasQuestMarkerContent && hasMainStoryAccent) {
          placeholderId = mainStoryPlaceholderBytes != null
              ? 'poi_placeholder_main_story'
              : null;
        } else if (hasQuestMarkerContent) {
          placeholderId = availablePlaceholderBytes != null
              ? 'poi_placeholder_available'
              : null;
        } else {
          placeholderId = placeholderBytes != null ? 'poi_placeholder' : null;
        }

        var addedSymbol = false;
        if (useRealImage) {
          final cachedImageBytes = _peekPoiMarkerImage(
            poi,
            hasQuestMarker: hasQuestMarkerContent,
            hasMainStoryAccent: hasMainStoryAccent,
          );
          if (cachedImageBytes != null) {
            symbolImageId = _poiMarkerImageId(
              poi,
              hasQuestMarker: hasQuestMarkerContent,
              hasMainStoryAccent: hasMainStoryAccent,
            );
            await _ensureMapImage(
              c,
              '${symbolImageId}_$_mapThumbnailVersion',
              cachedImageBytes,
            );
          }
        }
        if (symbolImageId != null || placeholderId != null) {
          final versionedId = symbolImageId != null
              ? '${symbolImageId}_$_mapThumbnailVersion'
              : '${placeholderId}_$_mapThumbnailVersion';
          poiSymbolRequests.add(
            _PoiSymbolRequest(
              poiId: poi.id,
              isQuestCurrent: isQuestCurrent,
              shouldPulseLikeQuest: shouldPulseLikeQuest,
              shouldPulseLikeMainStory: shouldPulseLikeMainStory,
              options: SymbolOptions(
                geometry: LatLng(lat, lng),
                iconImage: versionedId,
                iconSize: isQuestCurrent ? 0.82 : 0.75,
                iconHaloColor: _transparentMapHaloColor,
                iconHaloWidth: 0.0,
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
                zIndex: isQuestCurrent ? 4 : 2,
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

        if (useRealImage && symbolImageId == null) {
          poiImageUpdates.add(
            _PoiImageUpdate(
              poi: poi,
              isQuestCurrent: isQuestCurrent,
              hasMapContent: hasMapContent,
              hasQuestMarker: hasQuestMarkerContent,
              hasMainStoryAccent: hasMainStoryAccent,
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

      for (final ch in _characters) {
        final hasDiscovered = _hasDiscoveredCharacter(ch);
        final visiblePoints = _visibleCharacterPoints(ch);
        if (visiblePoints.isEmpty) continue;
        final thumbnailUrl = hasDiscovered ? ch.thumbnailUrl : null;
        final hasQuestAvailable = ch.hasAvailableQuest;
        final hasMainStoryAccent = _characterHasMainStoryAccent(ch, questLog);
        final shouldUseTurnInHalo = _isCurrentQuestTurnInCharacter(ch.id);
        final shouldPulseLikeQuest = _shouldPulseQuestCharacter(
          ch,
          isCurrentQuestTarget: shouldUseTurnInHalo,
          isTrackedQuestTarget: _isTrackedQuestTurnInCharacter(ch.id),
        );
        final shouldPulseLikeMainStory =
            featuredMainStoryPulseTarget?.characterId == ch.id;
        final useAccentMarker =
            hasQuestAvailable || shouldUseTurnInHalo || hasMainStoryAccent;
        Uint8List? markerBytes;
        String? markerId;
        if (thumbnailUrl != null && thumbnailUrl.isNotEmpty) {
          try {
            markerBytes = useAccentMarker
                ? (hasMainStoryAccent
                      ? await loadPoiThumbnailWithMainStoryMarker(thumbnailUrl)
                      : await loadPoiThumbnailWithQuestMarker(thumbnailUrl))
                : await loadPoiThumbnail(thumbnailUrl);
            if (markerBytes != null) {
              markerId = useAccentMarker
                  ? (hasMainStoryAccent
                        ? 'character_${ch.id}_main_story'
                        : 'character_${ch.id}_quest')
                  : 'character_${ch.id}';
            }
          } catch (_) {}
        }

        if (markerBytes == null) {
          markerBytes = useAccentMarker
              ? (hasMainStoryAccent
                    ? (characterMainStoryPlaceholderBytes ??
                          characterAvailablePlaceholderBytes ??
                          mainStoryPlaceholderBytes ??
                          availablePlaceholderBytes)
                    : (characterAvailablePlaceholderBytes ??
                          availablePlaceholderBytes))
              : (characterPlaceholderBytes ?? placeholderBytes);
          markerId = useAccentMarker
              ? (hasMainStoryAccent
                    ? 'character_placeholder_main_story'
                    : 'character_placeholder_available')
              : 'character_placeholder';
        }

        if (markerBytes != null && markerId != null) {
          final versionedId = '${markerId}_$_mapThumbnailVersion';
          try {
            await c.addImage(versionedId, markerBytes);
          } catch (_) {}
          for (final point in visiblePoints) {
            final sym = await c.addSymbol(
              SymbolOptions(
                geometry: point,
                iconImage: versionedId,
                iconSize: _standardMarkerThumbnailSize,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: _transparentMapHaloColor,
                iconHaloWidth: 0.0,
                iconAnchor: 'center',
                zIndex: shouldUseTurnInHalo ? 4 : 2,
              ),
              {'type': 'character', 'id': ch.id, 'name': ch.name},
            );
            if (!mounted) return;
            _characterSymbols.add(sym);
            (_characterSymbolsById[ch.id] ??= []).add(sym);
            _setPoiPulseHighlightState(
              sym,
              shouldPulseLikeQuest: shouldPulseLikeQuest,
              shouldPulseLikeMainStory: shouldPulseLikeMainStory,
            );
          }
          continue;
        }

        for (final point in visiblePoints) {
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
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
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
      await _refreshResourceSymbols();
      await _refreshBaseSymbols();
      await _refreshScenarioSymbols();
      await _refreshExpositionSymbols();
      await _refreshMonsterSymbols();
      await _refreshChallengeSymbols();
      if (_scenarioVisibilityRefreshPending) {
        _scenarioVisibilityRefreshPending = false;
        await _refreshScenarioSymbols();
        await _refreshExpositionSymbols();
        await _refreshMonsterSymbols();
        await _refreshChallengeSymbols();
      }
      if (deferPinReveal) {
        _pinBatchRevealInProgress = false;
        await _revealLoadedPins(c);
      }
      _ensureQuestPoiPulseTimer();
      if (mounted && hadEmptyDiscoveries) {
        setState(() => _addedMarkersWithEmptyDiscoveries = true);
      }
      await _applyMapMarkerIsolationIfNeeded();
      debugPrint('SinglePlayer: _addPoiMarkers done');
    } catch (e, st) {
      _pinBatchRevealInProgress = false;
      debugPrint('SinglePlayer: _addPoiMarkers error: $e');
      debugPrint('SinglePlayer: _addPoiMarkers stack: $st');
      if (mounted) setState(() => _markersAdded = false);
    } finally {
      _pinBatchRevealInProgress = false;
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
            return (request: request, symbol: sym);
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
        _setPoiPulseHighlightState(
          result.symbol,
          shouldPulseLikeQuest: result.request.shouldPulseLikeQuest,
          shouldPulseLikeMainStory: result.request.shouldPulseLikeMainStory,
        );
      }
    }
    await _applyMapMarkerIsolationIfNeeded();
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
      final trackedQuestPoiIds = _trackedQuestPoiIdsForPulse(questLog);
      final mapContentPoiIds = _buildPoiIdsWithMapContent();
      final questMarkerPoiIds = _buildPoiIdsWithQuestMarkerContent();
      await Future.wait(
        results.map(
          (result) => _applyPoiImageUpdate(
            c,
            markerGeneration,
            result,
            currentQuestPoiIds,
            trackedQuestPoiIds,
            mapContentPoiIds,
            questMarkerPoiIds,
          ),
        ),
      );
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _applyPoiImageUpdate(
    MapLibreMapController c,
    int markerGeneration,
    _PoiImageUpdateResult result,
    Set<String> currentQuestPoiIds,
    Set<String> trackedQuestPoiIds,
    Set<String> mapContentPoiIds,
    Set<String> questMarkerPoiIds,
  ) async {
    final bytes = result.bytes;
    final imageId = result.imageId;
    if (bytes == null || imageId == null) return;
    if (!mounted || markerGeneration != _poiMarkerGeneration) return;

    final isQuestCurrentNow = currentQuestPoiIds.contains(result.update.poi.id);
    final shouldPulseLikeQuest = _shouldPulseQuestPoi(
      result.update.poi,
      isQuestCurrent: isQuestCurrentNow,
      trackedQuestPoiIds: trackedQuestPoiIds,
    );
    final shouldPulseLikeMainStory =
        _featuredMainStoryPulseTarget(
          context.read<QuestLogProvider>(),
        )?.poiId ==
        result.update.poi.id;
    final hasMapContentNow = _poiHasMapContent(
      result.update.poi,
      isQuestCurrent: isQuestCurrentNow,
      mapContentPoiIds: mapContentPoiIds,
    );
    final hasQuestMarkerNow = _poiHasQuestMarkerContent(
      result.update.poi,
      isQuestCurrent: isQuestCurrentNow,
      questMarkerPoiIds: questMarkerPoiIds,
    );
    if (isQuestCurrentNow != result.update.isQuestCurrent ||
        hasMapContentNow != result.update.hasMapContent ||
        hasQuestMarkerNow != result.update.hasQuestMarker) {
      return;
    }

    final versionedId = '${imageId}_$_mapThumbnailVersion';
    await _ensureMapImage(c, versionedId, bytes);
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
            iconHaloColor: _transparentMapHaloColor,
            iconHaloWidth: 0.0,
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
            zIndex: isQuestCurrentNow ? 4 : 2,
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
        _setPoiPulseHighlightState(
          newSym,
          shouldPulseLikeQuest: shouldPulseLikeQuest,
          shouldPulseLikeMainStory: shouldPulseLikeMainStory,
        );
        await _applyMapMarkerIsolationIfNeeded();
      } catch (_) {}
      return;
    }

    try {
      await c.updateSymbol(
        sym,
        SymbolOptions(
          iconImage: versionedId,
          iconSize: isQuestCurrentNow ? 0.82 : 0.75,
          iconHaloColor: _transparentMapHaloColor,
          iconHaloWidth: 0.0,
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
          zIndex: isQuestCurrentNow ? 4 : 2,
        ),
      );
      _setPoiPulseHighlightState(
        sym,
        shouldPulseLikeQuest: shouldPulseLikeQuest,
        shouldPulseLikeMainStory: shouldPulseLikeMainStory,
      );
    } catch (_) {}
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<_PoiImageUpdateResult> _loadPoiImageUpdate(
    _PoiImageUpdate update,
  ) async {
    Uint8List? imageBytes;
    String? imageId;
    if (update.hasQuestMarker) {
      imageBytes = await _loadPoiMarkerImage(
        update.poi,
        hasQuestMarker: true,
        hasMainStoryAccent: update.hasMainStoryAccent,
      );
      if (imageBytes != null) {
        imageId = _poiMarkerImageId(
          update.poi,
          hasQuestMarker: true,
          hasMainStoryAccent: update.hasMainStoryAccent,
        );
      }
    } else {
      imageBytes = await _loadPoiMarkerImage(
        update.poi,
        hasQuestMarker: false,
        hasMainStoryAccent: false,
      );
      if (imageBytes != null) {
        imageId = _poiMarkerImageId(
          update.poi,
          hasQuestMarker: false,
          hasMainStoryAccent: false,
        );
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
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    final visibleTreasureChests =
        (_shouldSuppressNormalMapPinsForTutorial
                ? const <TreasureChest>[]
                : _treasureChests)
            .where(
              (t) =>
                  t.openedByUser != true &&
                  (selectedZoneId == null ||
                      selectedZoneId.isEmpty ||
                      t.zoneId == selectedZoneId),
            )
            .toList();
    final desiredIds = visibleTreasureChests
        .where((t) => t.openedByUser != true)
        .map((t) => t.id)
        .toSet();

    final duplicateOrOrphanSymbols = <Symbol>[];
    for (final symbol in _chestSymbols.toList()) {
      final id = _chestIdFromData(symbol.data);
      if (id == null) {
        duplicateOrOrphanSymbols.add(symbol);
        continue;
      }
      final tracked = _chestSymbolById[id];
      if (tracked == null || !identical(tracked, symbol)) {
        duplicateOrOrphanSymbols.add(symbol);
      }
    }
    if (duplicateOrOrphanSymbols.isNotEmpty) {
      try {
        await c.removeSymbols(duplicateOrOrphanSymbols);
      } catch (_) {}
      for (final symbol in duplicateOrOrphanSymbols) {
        _chestSymbols.remove(symbol);
      }
    }

    final duplicateOrOrphanCircles = <Circle>[];
    for (final circle in _chestCircles.toList()) {
      final id = _chestIdFromData(circle.data);
      if (id == null) {
        duplicateOrOrphanCircles.add(circle);
        continue;
      }
      final tracked = _chestCircleById[id];
      if (tracked == null || !identical(tracked, circle)) {
        duplicateOrOrphanCircles.add(circle);
      }
    }
    if (duplicateOrOrphanCircles.isNotEmpty) {
      for (final circle in duplicateOrOrphanCircles) {
        try {
          await c.removeCircle(circle);
        } catch (_) {}
        _chestCircles.remove(circle);
      }
    }

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

    for (final tc in visibleTreasureChests) {
      if (useImage) {
        final existing = _chestSymbolById[tc.id];
        if (existing == null) {
          final sym = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              iconImage: 'chest_thumbnail_$_mapThumbnailVersion',
              iconSize: 0.75,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
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
    await _applyMapMarkerIsolationIfNeeded();
  }

  String _healingFountainImageUrl(HealingFountain fountain) {
    final thumbnail = fountain.thumbnailUrl.trim();
    if (!fountain.discovered) {
      return _healingFountainFallbackImageUrl;
    }
    if (thumbnail.isEmpty || thumbnail == _healingFountainFallbackImageUrl) {
      return _healingFountainDiscoveredImageUrl;
    }
    return thumbnail;
  }

  Future<void> _refreshHealingFountainSymbols() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;

    final visibleHealingFountains = _shouldSuppressNormalMapPinsForTutorial
        ? const <HealingFountain>[]
        : _healingFountains;
    final desiredIds = visibleHealingFountains
        .map((fountain) => fountain.id)
        .toSet();

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

    for (final fountain in visibleHealingFountains) {
      if (fountain.id.isEmpty) continue;
      if (!fountain.latitude.isFinite || !fountain.longitude.isFinite) continue;
      if (fountain.latitude == 0.0 && fountain.longitude == 0.0) continue;

      final discovered = fountain.discovered;
      final circleColor = discovered
          ? (fountain.availableNow ? '#2ecc71' : '#7f8c8d')
          : '#3388ff';
      final imageSource = _healingFountainImageUrl(fountain);
      Uint8List? imageBytes;
      try {
        imageBytes = await loadPoiThumbnail(imageSource);
      } catch (_) {}

      if (imageBytes != null) {
        final imageKey = imageSource.hashCode.abs();
        final imageId =
            'healing_fountain_${fountain.id}_${imageKey}_$_mapThumbnailVersion';
        await _ensureMapImage(c, imageId, imageBytes);

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
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
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
                iconHaloColor: _transparentMapHaloColor,
                iconHaloWidth: 0.0,
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
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _refreshResourceSymbols() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;

    final visibleResources = _shouldSuppressNormalMapPinsForTutorial
        ? const <ResourceNode>[]
        : _resources;
    final desiredIds = visibleResources.map((resource) => resource.id).toSet();

    for (final entry in _resourceSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _resourceSymbols.remove(entry.value);
        _resourceSymbolById.remove(entry.key);
      }
    }
    for (final entry in _resourceCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _resourceCircles.remove(entry.value);
        _resourceCircleById.remove(entry.key);
      }
    }

    for (final resource in visibleResources) {
      if (resource.id.isEmpty) continue;
      if (!resource.latitude.isFinite || !resource.longitude.isFinite) continue;
      if (resource.latitude == 0.0 && resource.longitude == 0.0) continue;

      final imageSource = _resourceImageUrl(resource);
      Uint8List? imageBytes;
      if (imageSource.isNotEmpty) {
        try {
          imageBytes = await loadPoiThumbnail(imageSource);
        } catch (_) {}
      }

      if (imageBytes != null) {
        final imageKey = imageSource.hashCode.abs();
        final imageId =
            'resource_${resource.id}_${imageKey}_$_mapThumbnailVersion';
        await _ensureMapImage(c, imageId, imageBytes);

        final existingCircle = _resourceCircleById[resource.id];
        if (existingCircle != null) {
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _resourceCircles.remove(existingCircle);
          _resourceCircleById.remove(resource.id);
        }

        final existingSymbol = _resourceSymbolById[resource.id];
        if (existingSymbol == null) {
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(resource.latitude, resource.longitude),
              iconImage: imageId,
              iconSize: 0.72,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
              iconAnchor: 'center',
            ),
            {'type': 'resource', 'id': resource.id},
          );
          if (!mounted) return;
          _resourceSymbols.add(symbol);
          _resourceSymbolById[resource.id] = symbol;
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(resource.latitude, resource.longitude),
                iconImage: imageId,
                iconSize: 0.72,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                iconHaloColor: _transparentMapHaloColor,
                iconHaloWidth: 0.0,
              ),
            );
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _resourceSymbolById[resource.id];
      if (existingSymbol != null) {
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _resourceSymbols.remove(existingSymbol);
        _resourceSymbolById.remove(resource.id);
      }

      final existingCircle = _resourceCircleById[resource.id];
      if (existingCircle == null) {
        final circle = await c.addCircle(
          CircleOptions(
            geometry: LatLng(resource.latitude, resource.longitude),
            circleRadius: 20,
            circleOpacity: _mapMarkerStartingOpacity(1.0),
            circleColor: '#4f7d4f',
            circleStrokeWidth: 2,
            circleStrokeColor: '#f4e9d6',
          ),
          {'type': 'resource', 'id': resource.id},
        );
        if (!mounted) return;
        _resourceCircles.add(circle);
        _resourceCircleById[resource.id] = circle;
      } else {
        try {
          await c.updateCircle(
            existingCircle,
            CircleOptions(
              geometry: LatLng(resource.latitude, resource.longitude),
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleRadius: 20,
              circleColor: '#4f7d4f',
              circleStrokeWidth: 2,
              circleStrokeColor: '#f4e9d6',
            ),
          );
        } catch (_) {}
      }
    }

    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _refreshBaseSymbols() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final currentUserId = context.read<AuthProvider>().user?.id ?? '';

    final visibleBases = _shouldSuppressNormalMapPinsForTutorial
        ? const <BasePin>[]
        : _bases;
    final desiredIds = visibleBases.map((base) => base.id).toSet();

    for (final entry in _baseSymbolById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeSymbols([entry.value]);
        } catch (_) {}
        _baseSymbols.remove(entry.value);
        _baseSymbolById.remove(entry.key);
      }
    }
    for (final entry in _baseCircleById.entries.toList()) {
      if (!desiredIds.contains(entry.key)) {
        try {
          await c.removeCircle(entry.value);
        } catch (_) {}
        _baseCircles.remove(entry.value);
        _baseCircleById.remove(entry.key);
      }
    }

    for (final base in visibleBases) {
      if (base.id.isEmpty) continue;
      if (!base.latitude.isFinite || !base.longitude.isFinite) continue;
      if (base.latitude == 0.0 && base.longitude == 0.0) continue;
      final isCurrentUserBase =
          currentUserId.isNotEmpty && base.userId == currentUserId;
      final imageId = isCurrentUserBase
          ? 'base_diamond_marker_self_v6'
          : 'base_diamond_marker_v6';

      Uint8List? imageBytes;
      try {
        imageBytes = await loadBaseDiamondMarker(
          isCurrentUserBase: isCurrentUserBase,
        );
      } catch (_) {}

      if (imageBytes != null) {
        await _ensureMapImage(c, imageId, imageBytes);

        final existingCircle = _baseCircleById[base.id];
        if (existingCircle != null) {
          try {
            await c.removeCircle(existingCircle);
          } catch (_) {}
          _baseCircles.remove(existingCircle);
          _baseCircleById.remove(base.id);
        }

        final existingSymbol = _baseSymbolById[base.id];
        if (existingSymbol == null) {
          final symbol = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(base.latitude, base.longitude),
              iconImage: imageId,
              iconSize: _baseMarkerIconSize,
              iconOpacity: _mapMarkerStartingOpacity(1.0),
              iconHaloColor: _transparentMapHaloColor,
              iconHaloWidth: 0.0,
              iconAnchor: 'center',
              zIndex: isCurrentUserBase ? 4 : 3,
            ),
            {'type': 'base', 'id': base.id},
          );
          if (!mounted) return;
          _baseSymbols.add(symbol);
          _baseSymbolById[base.id] = symbol;
        } else {
          try {
            await c.updateSymbol(
              existingSymbol,
              SymbolOptions(
                geometry: LatLng(base.latitude, base.longitude),
                iconImage: imageId,
                iconSize: _baseMarkerIconSize,
                iconOpacity: _mapMarkerStartingOpacity(1.0),
                zIndex: isCurrentUserBase ? 4 : 3,
              ),
            );
          } catch (_) {}
        }
        continue;
      }

      final existingSymbol = _baseSymbolById[base.id];
      if (existingSymbol != null) {
        try {
          await c.removeSymbols([existingSymbol]);
        } catch (_) {}
        _baseSymbols.remove(existingSymbol);
        _baseSymbolById.remove(base.id);
      }

      final existingCircle = _baseCircleById[base.id];
      if (existingCircle == null) {
        final circle = await c.addCircle(
          CircleOptions(
            geometry: LatLng(base.latitude, base.longitude),
            circleRadius: _baseFallbackCircleRadius,
            circleOpacity: _mapMarkerStartingOpacity(1.0),
            circleColor: isCurrentUserBase ? '#5fabb8' : '#8c6239',
            circleStrokeWidth: 2,
            circleStrokeColor: isCurrentUserBase ? '#f7d672' : '#f4e9d6',
          ),
          {'type': 'base', 'id': base.id},
        );
        if (!mounted) return;
        _baseCircles.add(circle);
        _baseCircleById[base.id] = circle;
      } else {
        try {
          await c.updateCircle(
            existingCircle,
            CircleOptions(
              geometry: LatLng(base.latitude, base.longitude),
              circleOpacity: _mapMarkerStartingOpacity(1.0),
              circleRadius: _baseFallbackCircleRadius,
              circleColor: isCurrentUserBase ? '#5fabb8' : '#8c6239',
              circleStrokeWidth: 2,
              circleStrokeColor: isCurrentUserBase ? '#f7d672' : '#f4e9d6',
            ),
          );
        } catch (_) {}
      }
    }
    await _applyMapMarkerIsolationIfNeeded();
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

  Zone? _effectiveSelectedZone() {
    return context.read<ZoneProvider>().selectedZone;
  }

  bool _isPoiInSelectedZone(PointOfInterest poi) {
    final selectedZone = _effectiveSelectedZone();
    if (selectedZone == null) return false;
    final lat = double.tryParse(poi.lat);
    final lng = double.tryParse(poi.lng);
    if (lat == null || lng == null) return false;
    return _isPointInZone(selectedZone, lat, lng);
  }

  double _poiMarkerOpacity(
    PointOfInterest poi, {
    required bool isQuestCurrent,
    required bool undiscovered,
    Set<String>? mapContentPoiIds,
  }) {
    if (!_isPoiInSelectedZone(poi)) return 0.0;
    if (isQuestCurrent || !undiscovered) return 1.0;
    final hasMapContent = _poiHasMapContent(
      poi,
      isQuestCurrent: isQuestCurrent,
      mapContentPoiIds: mapContentPoiIds,
    );
    if (!hasMapContent) return 0.0;
    return 1.0;
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
    final hex = color.toARGB32().toRadixString(16).padLeft(8, '0').substring(2);
    return '#$hex';
  }

  String _shroudedToneForZone(Zone zone, {int salt = 0}) {
    final seed =
        '${zone.id}|${zone.latitude.toStringAsFixed(4)}|${zone.longitude.toStringAsFixed(4)}|shrouded|$salt';
    int hash = 0;
    for (final code in seed.codeUnits) {
      hash = 0x1fffffff & (hash + code);
      hash = 0x1fffffff & (hash + ((0x0007ffff & hash) << 10));
      hash ^= (hash >> 6);
    }
    hash = 0x1fffffff & (hash + ((0x03ffffff & hash) << 3));
    hash ^= (hash >> 11);
    hash = 0x1fffffff & (hash + ((0x00003fff & hash) << 15));
    final hue = 208 + (hash % 14); // 208–221
    final saturation = 18 + ((hash >> 8) % 10); // 18–27
    final lightness = 18 + ((hash >> 16) % 7); // 18–24
    final color = HSLColor.fromAHSL(
      1,
      hue.toDouble(),
      saturation / 100,
      lightness / 100,
    ).toColor();
    final hex = color.toARGB32().toRadixString(16).padLeft(8, '0').substring(2);
    return '#$hex';
  }

  Color? _parseHexColor(String? raw) {
    final trimmed = raw?.trim().toLowerCase() ?? '';
    if (!RegExp(r'^#[0-9a-f]{6}$').hasMatch(trimmed)) {
      return null;
    }
    final value = int.tryParse(trimmed.substring(1), radix: 16);
    if (value == null) {
      return null;
    }
    return Color(0xFF000000 | value);
  }

  String _hexFromColor(Color color) {
    final hex = color.toARGB32().toRadixString(16).padLeft(8, '0').substring(2);
    return '#$hex';
  }

  Color? _zoneKindBaseColor(Zone zone) => _parseHexColor(zone.kindOverlayColor);

  String _zoneKindShroudedFillColor(Color baseColor) {
    final hsl = HSLColor.fromColor(baseColor);
    final shrouded = hsl
        .withSaturation((hsl.saturation * 0.34).clamp(0.08, 0.24).toDouble())
        .withLightness((hsl.lightness * 0.4).clamp(0.13, 0.23).toDouble())
        .toColor();
    return _hexFromColor(shrouded);
  }

  String _zoneKindInnerAccentColor(Color baseColor) {
    final hsl = HSLColor.fromColor(baseColor);
    final accent = hsl
        .withSaturation((hsl.saturation * 1.08).clamp(0.2, 0.78).toDouble())
        .withLightness((hsl.lightness * 0.58).clamp(0.18, 0.34).toDouble())
        .toColor();
    return _hexFromColor(accent);
  }

  bool _isUndiscoveredZone(Zone zone) => !zone.discovered;

  String _zoneFillColor(Zone zone, {int salt = 0}) {
    final kindColor = _zoneKindBaseColor(zone);
    if (kindColor != null) {
      return _isUndiscoveredZone(zone)
          ? _zoneKindShroudedFillColor(kindColor)
          : _hexFromColor(kindColor);
    }
    if (_isUndiscoveredZone(zone)) {
      return _shroudedToneForZone(zone, salt: salt);
    }
    return _earthToneForZone(zone, salt: salt);
  }

  double _zoneFillOpacity(Zone zone) {
    return _isUndiscoveredZone(zone) ? 0.68 : 0.4;
  }

  String _zoneOuterLineColor(Zone zone) {
    return _isUndiscoveredZone(zone) ? '#0d1218' : '#000000';
  }

  double _zoneOuterLineOpacity(Zone zone) {
    return _isUndiscoveredZone(zone) ? 0.42 : 0.18;
  }

  double _zoneOuterLineWidth(Zone zone) {
    return _isUndiscoveredZone(zone) ? 8.0 : 7.0;
  }

  double _zoneOuterLineBlur(Zone zone) {
    return _isUndiscoveredZone(zone) ? 2.1 : 1.6;
  }

  String _zoneInnerLineColor(Zone zone) {
    if (_isUndiscoveredZone(zone)) {
      return '#d8c28b';
    }
    final kindColor = _zoneKindBaseColor(zone);
    if (kindColor != null) {
      return _zoneKindInnerAccentColor(kindColor);
    }
    return '#000000';
  }

  double _zoneInnerLineOpacity(Zone zone) {
    return _isUndiscoveredZone(zone) ? 0.9 : 0.95;
  }

  double _zoneInnerLineWidth(Zone zone) {
    return _isUndiscoveredZone(zone) ? 2.2 : 2.8;
  }

  Future<void> _addZoneBoundaries() {
    _zoneBoundaryRefreshSequence = _zoneBoundaryRefreshSequence.then((_) async {
      try {
        await _addZoneBoundariesNow();
      } catch (e, st) {
        debugPrint('SinglePlayer: _addZoneBoundaries error: $e');
        debugPrint('SinglePlayer: _addZoneBoundaries stack: $st');
      }
    });
    return _zoneBoundaryRefreshSequence;
  }

  Future<void> _addZoneBoundariesNow() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) {
      debugPrint(
        'SinglePlayer: _addZoneBoundaries skip (controller=${c != null} styleLoaded=$_styleLoaded)',
      );
      return;
    }
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    if (_zoneLines.isNotEmpty) {
      await _syncZoneFillSelection(selectedZoneId);
      return;
    }

    if (_zoneFills.isNotEmpty) {
      try {
        await c.removeFills(_zoneFills);
      } catch (_) {}
      if (!mounted) return;
      _zoneFills.clear();
      _zoneFillById.clear();
    }
    final options = <LineOptions>[];
    final lineData = <Map>[];
    final fillOptions = <FillOptions>[];
    final fillData = <Map>[];
    for (var i = 0; i < _zones.length; i++) {
      final z = _zones[i];
      final ring = _zoneRing(z);
      if (ring.length < 2) continue;
      if (selectedZoneId == null || selectedZoneId != z.id) {
        fillOptions.add(
          FillOptions(
            geometry: [ring],
            fillColor: _zoneFillColor(z, salt: i),
            fillOpacity: _zoneFillOpacity(z),
          ),
        );
        fillData.add({'type': 'zone', 'id': z.id});
      }
      options.add(
        LineOptions(
          geometry: ring,
          lineColor: _zoneOuterLineColor(z),
          lineWidth: _zoneOuterLineWidth(z),
          lineOpacity: _zoneOuterLineOpacity(z),
          lineBlur: _zoneOuterLineBlur(z),
          lineJoin: 'round',
        ),
      );
      lineData.add({'type': 'zone', 'id': z.id});
      options.add(
        LineOptions(
          geometry: ring,
          lineColor: _zoneInnerLineColor(z),
          lineWidth: _zoneInnerLineWidth(z),
          lineOpacity: _zoneInnerLineOpacity(z),
          lineJoin: 'round',
        ),
      );
      lineData.add({'type': 'zone', 'id': z.id});
    }
    debugPrint(
      'SinglePlayer: _addZoneBoundaries zones=${_zones.length} rings=${options.length}',
    );
    if (options.isEmpty && fillOptions.isEmpty) return;
    if (fillOptions.isNotEmpty) {
      final fills = await c.addFills(fillOptions, fillData);
      if (!mounted) return;
      _zoneFills.addAll(fills);
      for (var i = 0; i < fills.length && i < fillData.length; i++) {
        final zoneId = fillData[i]['id']?.toString();
        if (zoneId == null || zoneId.isEmpty) continue;
        _zoneFillById[zoneId] = fills[i];
      }
    }
    final lines = await c.addLines(options, lineData);
    if (!mounted) return;
    _zoneLines.addAll(lines);
    _renderedSelectedZoneId = selectedZoneId;
    debugPrint('SinglePlayer: _addZoneBoundaries added ${lines.length} lines');
  }

  Future<void> _syncZoneFillSelection(String? selectedZoneId) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    if (_renderedSelectedZoneId == selectedZoneId) return;

    final previousZoneId = _renderedSelectedZoneId;
    if (previousZoneId != null && previousZoneId.isNotEmpty) {
      await _ensureZoneFill(previousZoneId);
    }
    if (selectedZoneId != null && selectedZoneId.isNotEmpty) {
      await _removeZoneFill(selectedZoneId);
    }
    _renderedSelectedZoneId = selectedZoneId;
  }

  Future<void> _ensureZoneFill(String zoneId) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    if (_zoneFillById.containsKey(zoneId)) return;
    final zoneIndex = _zones.indexWhere((zone) => zone.id == zoneId);
    if (zoneIndex < 0) return;
    final zone = _zones[zoneIndex];
    final ring = _zoneRing(zone);
    if (ring.length < 2) return;
    try {
      final fills = await c.addFills(
        [
          FillOptions(
            geometry: [ring],
            fillColor: _zoneFillColor(zone, salt: zoneIndex),
            fillOpacity: _zoneFillOpacity(zone),
          ),
        ],
        [
          {'type': 'zone', 'id': zone.id},
        ],
      );
      if (!mounted || fills.isEmpty) return;
      final fill = fills.first;
      _zoneFills.add(fill);
      _zoneFillById[zone.id] = fill;
    } catch (_) {}
  }

  Future<void> _removeZoneFill(String zoneId) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final fill = _zoneFillById.remove(zoneId);
    if (fill == null) return;
    _zoneFills.remove(fill);
    try {
      await c.removeFills([fill]);
    } catch (_) {}
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
    final polygons = _trackedQuestCurrentNodePolygons(questLog);
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
    _mainStoryPoiHighlightSymbols.remove(sym);
    _questPoiHighlightSymbols.remove(sym);
    if (enabled) {
      _questPoiHighlightSymbols.add(sym);
    }
    _ensureQuestPoiPulseTimer();
  }

  void _setQuestCircleHighlight(Circle circle, bool enabled) {
    _mainStoryPoiHighlightCircles.remove(circle);
    _questPoiHighlightCircles.remove(circle);
    if (enabled) {
      _questPoiHighlightCircles.add(circle);
    }
    _ensureQuestPoiPulseTimer();
  }

  void _setMainStoryPoiHighlight(Symbol sym, bool enabled) {
    _questPoiHighlightSymbols.remove(sym);
    _mainStoryPoiHighlightSymbols.remove(sym);
    if (enabled) {
      _mainStoryPoiHighlightSymbols.add(sym);
    }
    _ensureQuestPoiPulseTimer();
  }

  void _setPoiPulseHighlightState(
    Symbol sym, {
    required bool shouldPulseLikeQuest,
    required bool shouldPulseLikeMainStory,
  }) {
    if (shouldPulseLikeMainStory) {
      _setMainStoryPoiHighlight(sym, true);
      return;
    }
    _setQuestPoiHighlight(sym, shouldPulseLikeQuest);
  }

  Future<void> _revealLoadedPins(MapLibreMapController c) async {
    if (!_styleLoaded || !_markersAdded) return;
    await _updateNormalPinOpacities(
      c,
      1.0,
      questPoiIds: _currentQuestPoiIdsForFilter(
        context.read<QuestLogProvider>(),
      ),
      discoveries: context.read<DiscoveriesProvider>(),
      mapContentPoiIds: _buildPoiIdsWithMapContent(),
    );
  }

  Future<void> _removePoiSymbol(String poiId) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final existing = _poiSymbolById.remove(poiId);
    if (existing == null) return;
    _poiSymbols.remove(existing);
    _setQuestPoiHighlight(existing, false);
    try {
      await c.removeSymbols([existing]);
    } catch (_) {}
  }

  void _ensureQuestPoiPulseTimer() {
    _questPoiHighlightSymbols.removeWhere((symbol) {
      final geometry = symbol.options.geometry;
      final opacity = symbol.options.iconOpacity ?? 1.0;
      return geometry == null || opacity <= 0.05;
    });
    _questPoiHighlightCircles.removeWhere((circle) {
      final geometry = circle.options.geometry;
      final opacity = circle.options.circleOpacity ?? 1.0;
      return geometry == null || opacity <= 0.05;
    });
    _mainStoryPoiHighlightSymbols.removeWhere((symbol) {
      final geometry = symbol.options.geometry;
      final opacity = symbol.options.iconOpacity ?? 1.0;
      return geometry == null || opacity <= 0.05;
    });
    _mainStoryPoiHighlightCircles.removeWhere((circle) {
      final geometry = circle.options.geometry;
      final opacity = circle.options.circleOpacity ?? 1.0;
      return geometry == null || opacity <= 0.05;
    });
    if (_questPoiHighlightSymbols.isEmpty &&
        _questPoiHighlightCircles.isEmpty &&
        _mainStoryPoiHighlightSymbols.isEmpty &&
        _mainStoryPoiHighlightCircles.isEmpty) {
      _questPoiPulseTimer?.cancel();
      _questPoiPulseTimer = null;
      return;
    }
    if (_questPoiPulseTimer != null) return;

    Future<void> pulseTrackedAnnotations() async {
      final symbolSnapshot = List<Symbol>.from(_questPoiHighlightSymbols);
      for (final symbol in symbolSnapshot) {
        final geometry = symbol.options.geometry;
        final opacity = symbol.options.iconOpacity ?? 1.0;
        if (geometry == null || opacity <= 0.05) continue;
        unawaited(_pulsePoi(geometry.latitude, geometry.longitude));
      }
      final circleSnapshot = List<Circle>.from(_questPoiHighlightCircles);
      for (final circle in circleSnapshot) {
        final geometry = circle.options.geometry;
        final opacity = circle.options.circleOpacity ?? 1.0;
        if (geometry == null || opacity <= 0.05) continue;
        unawaited(_pulsePoi(geometry.latitude, geometry.longitude));
      }
      final mainStorySymbolSnapshot = List<Symbol>.from(
        _mainStoryPoiHighlightSymbols,
      );
      for (final symbol in mainStorySymbolSnapshot) {
        final geometry = symbol.options.geometry;
        final opacity = symbol.options.iconOpacity ?? 1.0;
        if (geometry == null || opacity <= 0.05) continue;
        unawaited(
          _pulseMainStoryLeadFocus(geometry.latitude, geometry.longitude),
        );
      }
      final mainStoryCircleSnapshot = List<Circle>.from(
        _mainStoryPoiHighlightCircles,
      );
      for (final circle in mainStoryCircleSnapshot) {
        final geometry = circle.options.geometry;
        final opacity = circle.options.circleOpacity ?? 1.0;
        if (geometry == null || opacity <= 0.05) continue;
        unawaited(
          _pulseMainStoryLeadFocus(geometry.latitude, geometry.longitude),
        );
      }
    }

    unawaited(pulseTrackedAnnotations());
    _questPoiPulseTimer = Timer.periodic(const Duration(milliseconds: 1800), (
      _,
    ) {
      unawaited(pulseTrackedAnnotations());
    });
  }

  Future<void> _updatePoiSymbolForQuestState(
    String poiId, {
    required bool isQuestCurrent,
  }) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    if (_shouldSuppressNormalMapPinsForTutorial) {
      await _removePoiSymbol(poiId);
      await _applyMapMarkerIsolationIfNeeded();
      return;
    }
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
    final questLog = context.read<QuestLogProvider>();
    final useRealImage = discoveries.hasDiscovered(poi.id);
    final undiscovered = !useRealImage;
    final shouldPulseLikeQuest = _shouldPulseQuestPoi(
      poi,
      isQuestCurrent: isQuestCurrent,
      trackedQuestPoiIds: _trackedQuestPoiIdsForPulse(questLog),
    );
    final shouldPulseLikeMainStory =
        _featuredMainStoryPulseTarget(questLog)?.poiId == poi.id;
    final mapContentPoiIds = _buildPoiIdsWithMapContent();
    final questMarkerPoiIds = _buildPoiIdsWithQuestMarkerContent();
    final hasMapContent = _poiHasMapContent(
      poi,
      isQuestCurrent: isQuestCurrent,
      mapContentPoiIds: mapContentPoiIds,
    );
    final hasQuestMarkerContent = _poiHasQuestMarkerContent(
      poi,
      isQuestCurrent: isQuestCurrent,
      questMarkerPoiIds: questMarkerPoiIds,
    );
    final hasMainStoryAccent = _poiHasMainStoryAccent(poi, questLog);
    final hasCharacter = poi.characters.isNotEmpty;
    final baseEligible =
        _isPoiInSelectedZone(poi) &&
        (!undiscovered || hasCharacter || hasMapContent);
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
      await _applyMapMarkerIsolationIfNeeded();
      return;
    }

    final sym = _poiSymbolById[poiId];

    late final String imageId;
    Uint8List? imageBytes;
    if (hasQuestMarkerContent) {
      imageBytes = useRealImage
          ? await _loadPoiMarkerImage(
              poi,
              hasQuestMarker: true,
              hasMainStoryAccent: hasMainStoryAccent,
            )
          : null;
      imageId = imageBytes != null
          ? _poiMarkerImageId(
              poi,
              hasQuestMarker: true,
              hasMainStoryAccent: hasMainStoryAccent,
            )
          : (hasMainStoryAccent
                ? 'poi_placeholder_main_story'
                : 'poi_placeholder_available');
    } else if (useRealImage) {
      imageBytes = await _loadPoiMarkerImage(
        poi,
        hasQuestMarker: false,
        hasMainStoryAccent: false,
      );
      imageId = imageBytes != null
          ? _poiMarkerImageId(
              poi,
              hasQuestMarker: false,
              hasMainStoryAccent: false,
            )
          : 'poi_placeholder';
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
        } else if (imageId == 'poi_placeholder_main_story') {
          imageBytes = await loadPoiThumbnailWithMainStoryMarker(null);
        }
      } catch (_) {}
    }
    if (imageBytes != null) {
      await _ensureMapImage(c, versionedId, imageBytes);
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
            iconHaloColor: _transparentMapHaloColor,
            iconHaloWidth: 0.0,
            iconOpacity: _poiMarkerOpacity(
              poi,
              isQuestCurrent: isQuestCurrent,
              undiscovered: undiscovered,
              mapContentPoiIds: mapContentPoiIds,
            ),
            iconAnchor: 'center',
            zIndex: isQuestCurrent ? 4 : 2,
          ),
          {'type': 'poi', 'id': poi.id, 'name': poi.name},
        );
        if (!mounted) return;
        _poiSymbols.add(newSym);
        _poiSymbolById[poi.id] = newSym;
        _setPoiPulseHighlightState(
          newSym,
          shouldPulseLikeQuest: shouldPulseLikeQuest,
          shouldPulseLikeMainStory: shouldPulseLikeMainStory,
        );
        await _applyMapMarkerIsolationIfNeeded();
      } catch (_) {}
      return;
    }

    try {
      await c.updateSymbol(
        sym,
        SymbolOptions(
          iconImage: versionedId,
          iconSize: isQuestCurrent ? 0.82 : 0.75,
          iconHaloColor: _transparentMapHaloColor,
          iconHaloWidth: 0.0,
          iconOpacity: _poiMarkerOpacity(
            poi,
            isQuestCurrent: isQuestCurrent,
            undiscovered: undiscovered,
            mapContentPoiIds: mapContentPoiIds,
          ),
          iconAnchor: 'center',
          zIndex: isQuestCurrent ? 4 : 2,
        ),
      );
      _setPoiPulseHighlightState(
        sym,
        shouldPulseLikeQuest: shouldPulseLikeQuest,
        shouldPulseLikeMainStory: shouldPulseLikeMainStory,
      );
    } catch (_) {}
    await _applyMapMarkerIsolationIfNeeded();
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

  Future<void> _ensureMapImage(
    MapLibreMapController controller,
    String imageId,
    Uint8List imageBytes,
  ) async {
    if (_mapImageIds.contains(imageId)) return;
    try {
      await controller.addImage(imageId, imageBytes);
    } catch (_) {
      // MapLibre throws when an image already exists; treat that as satisfied.
    }
    _mapImageIds.add(imageId);
  }

  bool _hasDiscoveredCharacter(Character character) {
    if (_isCurrentQuestTurnInCharacter(character.id)) {
      return true;
    }
    if (!_isCharacterDiscoveryManaged(character)) return true;
    final poiId = _characterDiscoveryPoiId(character);
    if (poiId.isNotEmpty) {
      return context.read<DiscoveriesProvider>().hasDiscovered(poiId);
    }
    return _discoveredCharacterIds.contains(character.id);
  }

  List<LatLng> _visibleCharacterPoints(Character ch) {
    if (_shouldSuppressNormalMapPinsForTutorial) {
      return const <LatLng>[];
    }
    final points = ch.locations
        .map((loc) => LatLng(loc.latitude, loc.longitude))
        .where((p) => p.latitude != 0 || p.longitude != 0)
        .toList();
    if (points.isEmpty) return const <LatLng>[];

    final selectedZone = _effectiveSelectedZone();
    if (selectedZone == null) return const <LatLng>[];

    return points
        .where(
          (point) =>
              _isPointInZone(selectedZone, point.latitude, point.longitude),
        )
        .toList();
  }

  Future<void> _removeCharacterSymbols(String characterId) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final existingSymbols = _characterSymbolsById.remove(characterId);
    if (existingSymbols == null || existingSymbols.isEmpty) return;
    for (final symbol in existingSymbols) {
      _setQuestPoiHighlight(symbol, false);
    }
    _characterSymbols.removeWhere((symbol) => existingSymbols.contains(symbol));
    try {
      await c.removeSymbols(existingSymbols);
    } catch (_) {}
  }

  Future<void> _updateCharacterSymbolsForState(
    List<Character> characters,
  ) async {
    for (final ch in characters) {
      await _updateCharacterSymbolForState(ch);
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _updateCharacterSymbolForState(Character ch) async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final visiblePoints = _visibleCharacterPoints(ch);
    await _removeCharacterSymbols(ch.id);
    if (visiblePoints.isEmpty) return;

    final hasDiscovered = _hasDiscoveredCharacter(ch);
    final thumbnailUrl = hasDiscovered ? ch.thumbnailUrl : null;
    final hasQuestAvailable = ch.hasAvailableQuest;
    final questLog = context.read<QuestLogProvider>();
    final hasMainStoryAccent = _characterHasMainStoryAccent(ch, questLog);
    final shouldUseTurnInHalo = _isCurrentQuestTurnInCharacter(ch.id);
    final shouldPulseLikeQuest = _shouldPulseQuestCharacter(
      ch,
      isCurrentQuestTarget: shouldUseTurnInHalo,
      isTrackedQuestTarget: _isTrackedQuestTurnInCharacter(ch.id),
    );
    final shouldPulseLikeMainStory =
        _featuredMainStoryPulseTarget(questLog)?.characterId == ch.id;
    final useAccentMarker =
        hasQuestAvailable || shouldUseTurnInHalo || hasMainStoryAccent;
    Uint8List? imageBytes;
    String? imageId;
    if (thumbnailUrl != null && thumbnailUrl.isNotEmpty) {
      try {
        imageBytes = useAccentMarker
            ? (hasMainStoryAccent
                  ? await loadPoiThumbnailWithMainStoryMarker(thumbnailUrl)
                  : await loadPoiThumbnailWithQuestMarker(thumbnailUrl))
            : await loadPoiThumbnail(thumbnailUrl);
        if (imageBytes != null) {
          imageId = useAccentMarker
              ? (hasMainStoryAccent
                    ? 'character_${ch.id}_main_story'
                    : 'character_${ch.id}_quest')
              : 'character_${ch.id}';
        }
      } catch (_) {}
    }
    if (imageBytes == null) {
      try {
        imageBytes = useAccentMarker
            ? (hasMainStoryAccent
                  ? await loadPoiThumbnailWithMainStoryMarker(
                      _characterMysteryImageUrl,
                    )
                  : await loadPoiThumbnailWithQuestMarker(
                      _characterMysteryImageUrl,
                    ))
            : await loadPoiThumbnail(_characterMysteryImageUrl);
        if (imageBytes != null) {
          imageId = useAccentMarker
              ? (hasMainStoryAccent
                    ? 'character_placeholder_main_story'
                    : 'character_placeholder_available')
              : 'character_placeholder';
        }
      } catch (_) {}
    }
    if (imageBytes == null) {
      try {
        imageBytes = useAccentMarker
            ? (hasMainStoryAccent
                  ? await loadPoiThumbnailWithMainStoryMarker(null)
                  : await loadPoiThumbnailWithQuestMarker(null))
            : await loadPoiThumbnail(null);
        if (imageBytes != null) {
          imageId = useAccentMarker
              ? (hasMainStoryAccent
                    ? 'character_placeholder_main_story'
                    : 'character_placeholder_available')
              : 'character_placeholder';
        }
      } catch (_) {}
    }
    if (imageBytes == null || imageId == null) return;
    final versionedId = '${imageId}_$_mapThumbnailVersion';
    try {
      await c.addImage(versionedId, imageBytes);
    } catch (_) {}
    for (final point in visiblePoints) {
      try {
        final sym = await c.addSymbol(
          SymbolOptions(
            geometry: point,
            iconImage: versionedId,
            iconSize: _standardMarkerThumbnailSize,
            iconOpacity: _mapMarkerStartingOpacity(1.0),
            iconHaloColor: _transparentMapHaloColor,
            iconHaloWidth: 0.0,
            iconAnchor: 'center',
            zIndex: shouldUseTurnInHalo ? 4 : 2,
          ),
          {'type': 'character', 'id': ch.id, 'name': ch.name},
        );
        if (!mounted) return;
        _characterSymbols.add(sym);
        (_characterSymbolsById[ch.id] ??= []).add(sym);
        _setPoiPulseHighlightState(
          sym,
          shouldPulseLikeQuest: shouldPulseLikeQuest,
          shouldPulseLikeMainStory: shouldPulseLikeMainStory,
        );
      } catch (_) {}
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _refreshCharacterDiscoveryMarkers({
    bool undiscoveredOnly = false,
  }) async {
    if (!_styleLoaded || _mapController == null || !_markersAdded) return;
    for (final ch in _characters) {
      if (!_isCharacterDiscoveryManaged(ch)) continue;
      if (undiscoveredOnly && _hasDiscoveredCharacter(ch)) continue;
      await _updateCharacterSymbolForState(ch);
    }
    await _applyMapMarkerIsolationIfNeeded();
  }

  Future<void> _refreshCharacterMarkersForPoi(String poiId) async {
    if (poiId.isEmpty) return;
    final linkedCharacters = _characters
        .where((character) => _characterDiscoveryPoiId(character) == poiId)
        .toList();
    if (linkedCharacters.isEmpty) return;
    await _updateCharacterSymbolsForState(linkedCharacters);
    await _applyMapMarkerIsolationIfNeeded();
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
    await updateSymbolOpacity(_resourceSymbolById.values);
    await updateCircleOpacity(_resourceCircleById);
    await updateSymbolOpacity(_baseSymbolById.values);
    await updateCircleOpacity(_baseCircleById);
    await updateSymbolOpacity(_scenarioSymbolById.values);
    await updateCircleOpacity(_scenarioCircleById);
    await updateSymbolOpacity(_expositionSymbolById.values);
    await updateCircleOpacity(_expositionCircleById);
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
    String? standaloneChallengeZoneId,
    String? challengeImageHeroTag,
  }) async {
    final parentContext = context;
    final textController = TextEditingController();
    CapturedImage? capturedImage;
    PlatformFile? capturedVideo;
    bool uploadingSubmission = false;
    final objective = node.objective;
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
              final selectedObjectiveHeroTag = objective == null
                  ? null
                  : (challengeImageHeroTag ??
                        _challengeImageHeroTag(objective.id));
              final statValues = context.watch<CharacterStatsProvider>().stats;
              final statTags = (objective?.statTags ?? const [])
                  .map((tag) => tag.trim().toLowerCase())
                  .where((tag) => tag.isNotEmpty)
                  .toList();
              final difficultyValue = objective?.difficulty ?? 0;
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
                  if ((objective?.prompt.trim() ?? '').isNotEmpty)
                    Text(
                      objective!.prompt,
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  if (objective != null &&
                      (objective.imageUrl.isNotEmpty ||
                          objective.thumbnailUrl.isNotEmpty)) ...[
                    const SizedBox(height: 10),
                    Hero(
                      tag: selectedObjectiveHeroTag!,
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(12),
                        child: Image.network(
                          objective.thumbnailUrl.isNotEmpty
                              ? objective.thumbnailUrl
                              : objective.imageUrl,
                          fit: BoxFit.cover,
                          height: 220,
                          width: double.infinity,
                          errorBuilder: (_, __, ___) => const SizedBox.shrink(),
                        ),
                      ),
                    ),
                  ],
                  if (objective != null) ...[
                    const SizedBox(height: 6),
                    Text(
                      'Difficulty: ${objective.difficulty}',
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
                            final submissionLogLabel =
                                standaloneChallengeId == null
                                ? 'quest-node:${node.id}'
                                : 'challenge:$standaloneChallengeId';
                            void logSubmission(String message) {
                              debugPrint(
                                '[challenge-submission][$submissionLogLabel] '
                                '$message',
                              );
                            }

                            final startedAt = DateTime.now();
                            setModalState(() => uploadingSubmission = true);
                            Navigator.of(context).pop();
                            void updateLoadingStep(String stepLabel) {
                              _setQuestSubmissionOverlay(
                                QuestSubmissionOverlayPhase.loading,
                                stepLabel: stepLabel,
                              );
                            }

                            updateLoadingStep(
                              isPhotoSubmission
                                  ? 'Preparing photo upload...'
                                  : isVideoSubmission
                                  ? 'Preparing video upload...'
                                  : 'Sending answer to the Dungeonmaster...',
                            );
                            logSubmission(
                              'starting submission '
                              'type=$submissionType '
                              'hasText=${trimmedText.isNotEmpty} '
                              'hasPhoto=${capturedImage != null} '
                              'hasVideo=${capturedVideo != null}',
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
                              logSubmission(
                                'requesting photo upload URL '
                                'bytes=${capturedImage!.bytes.length} '
                                'mime=${capturedImage!.mimeType ?? 'image/jpeg'} '
                                'key=$key',
                              );
                              updateLoadingStep('Preparing photo upload...');
                              final url = await mediaService
                                  .getPresignedUploadUrl(
                                    ApiConstants.crewPointsOfInterestBucket,
                                    key,
                                    debugLabel: '$submissionLogLabel:photo',
                                  );
                              if (url == null) {
                                logSubmission('failed to prepare photo upload');
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
                              updateLoadingStep('Uploading photo...');
                              final uploadResult = await mediaService
                                  .uploadToPresigned(
                                    url,
                                    Uint8List.fromList(capturedImage!.bytes),
                                    capturedImage!.mimeType ?? 'image/jpeg',
                                    debugLabel: '$submissionLogLabel:photo',
                                  );
                              if (!uploadResult.success) {
                                logSubmission(
                                  'photo upload failed '
                                  'timedOut=${uploadResult.timedOut} '
                                  'status=${uploadResult.statusCode} '
                                  'duration=${uploadResult.duration} '
                                  'error=${uploadResult.errorDescription}',
                                );
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
                                  message: uploadResult.timedOut
                                      ? 'Photo upload timed out. Please try again.'
                                      : 'Failed to upload photo.',
                                );
                                return;
                              }
                              imageSubmissionUrl = url.split('?').first;
                              logSubmission(
                                'photo upload complete objectUrl=$imageSubmissionUrl',
                              );
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
                              logSubmission(
                                'requesting video upload URL '
                                'bytes=${capturedVideo!.bytes?.length ?? 0} '
                                'mime=${_mimeTypeFromFile(capturedVideo!) ?? 'video/mp4'} '
                                'key=$key',
                              );
                              updateLoadingStep('Preparing video upload...');
                              final url = await mediaService
                                  .getPresignedUploadUrl(
                                    ApiConstants.crewPointsOfInterestBucket,
                                    key,
                                    debugLabel: '$submissionLogLabel:video',
                                  );
                              if (url == null) {
                                logSubmission('failed to prepare video upload');
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
                              updateLoadingStep('Uploading video...');
                              final uploadResult = await mediaService
                                  .uploadToPresigned(
                                    url,
                                    Uint8List.fromList(bytes),
                                    _mimeTypeFromFile(capturedVideo!) ??
                                        'video/mp4',
                                    debugLabel: '$submissionLogLabel:video',
                                  );
                              if (!uploadResult.success) {
                                logSubmission(
                                  'video upload failed '
                                  'timedOut=${uploadResult.timedOut} '
                                  'status=${uploadResult.statusCode} '
                                  'duration=${uploadResult.duration} '
                                  'error=${uploadResult.errorDescription}',
                                );
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
                                  message: uploadResult.timedOut
                                      ? 'Video upload timed out. Please try again.'
                                      : 'Failed to upload video.',
                                );
                                return;
                              }
                              videoSubmissionUrl = url.split('?').first;
                              logSubmission(
                                'video upload complete objectUrl=$videoSubmissionUrl',
                              );
                            }
                            late final Map<String, dynamic> resp;
                            final previousLevel = context
                                .read<CharacterStatsProvider>()
                                .level;
                            updateLoadingStep(
                              'Waiting for the Dungeonmaster to judge your submission...',
                            );
                            logSubmission(
                              'sending submission '
                              'hasText=${isTextSubmission ? trimmedText.isNotEmpty : false} '
                              'hasImageUrl=${imageSubmissionUrl != null} '
                              'hasVideoUrl=${videoSubmissionUrl != null} '
                              'standaloneChallengeId=$standaloneChallengeId',
                            );
                            try {
                              resp = standaloneChallengeId == null
                                  ? await questLogProvider.submitQuestNode(
                                      node.id,
                                      textSubmission: isTextSubmission
                                          ? trimmedText
                                          : null,
                                      imageSubmissionUrl: imageSubmissionUrl,
                                      videoSubmissionUrl: videoSubmissionUrl,
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
                              logSubmission(
                                'submission failed '
                                'duration=${DateTime.now().difference(startedAt)} '
                                'error=$error',
                              );
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
                            logSubmission(
                              'submission complete '
                              'success=$success '
                              'reason="$reason" '
                              'score=$score '
                              'duration=${DateTime.now().difference(startedAt)}',
                            );
                            if (mounted) {
                              _dismissQuestSubmissionOverlay();
                            }
                            if (success && standaloneChallengeId != null) {
                              unawaited(
                                _refreshStandaloneChallengeZoneContent(
                                  standaloneChallengeZoneId,
                                ),
                              );
                            }
                            final currentLevel =
                                await _refreshRewardDrivenPlayerState();
                            if (!mounted || !parentContext.mounted) return;
                            if (mounted && parentContext.mounted) {
                              final completedTaskProvider = parentContext
                                  .read<CompletedTaskProvider>();
                              completedTaskProvider.showModal(
                                'challengeOutcome',
                                data: {
                                  ...Map<String, dynamic>.from(resp),
                                  if (baseMessage.isNotEmpty &&
                                      (resp['reason']?.toString().trim() ?? '')
                                          .isEmpty)
                                    'reason': baseMessage,
                                  if (score != null) 'score': score,
                                  if (difficulty != null)
                                    'difficulty': difficulty,
                                  if (combined != null)
                                    'combinedScore': combined,
                                  if (statTags != null) 'statTags': statTags,
                                  if (statValues != null)
                                    'statValues': statValues,
                                },
                              );
                              completedTaskProvider
                                  .queueLevelUpFollowUpIfNeeded(
                                    previousLevel: previousLevel,
                                    currentLevel: currentLevel,
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

  void _updateSelectedZoneFromLocation({bool force = false}) {
    final location = context.read<LocationProvider>().location;
    if (location == null || _zones.isEmpty) return;

    final zoneProvider = context.read<ZoneProvider>();
    final zone = zoneProvider.findZoneAtCoordinate(
      location.latitude,
      location.longitude,
    );
    if (force) {
      zoneProvider.unlockSelection();
    }
    if (zone == null) {
      zoneProvider.setSelectedZone(null);
      return;
    }
    if (!zone.discovered) {
      zoneProvider.setSelectedZone(null);
      unawaited(_discoverZoneForLocation(zone, location));
      return;
    }
    zoneProvider.setSelectedZone(zone);
  }

  void _setQuestSubmissionOverlay(
    QuestSubmissionOverlayPhase phase, {
    String? message,
    String? stepLabel,
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
      _questSubmissionStepLabel = phase == QuestSubmissionOverlayPhase.loading
          ? stepLabel
          : null;
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
      _questSubmissionStepLabel = null;
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
    const overlayButtonSize = _overlayRailButtonSize;
    const overlayButtonSpacing = _overlayRailButtonSpacing;
    const overlayButtonCount = 3;
    const topOverlayLaneGap = 20.0;
    const partyStripTileWidth = 58.0;
    const partyStripToggleGap = 2.0;
    const partyStripToggleWidth = 26.0;
    const partyStripHorizontalPadding = 12.0;
    const trackedQuestOverlayCollapsedHeight = 48.0;
    const bottomOverlayPadding = 24.0;
    const zoneLoadingSpinnerSize = 28.0;
    const polygonActionGap = 12.0;
    const stackedPolygonActionSpacing = 56.0;
    final screenWidth = MediaQuery.sizeOf(context).width;
    final tutorialStatus = _tutorialStatus;
    final showTutorialGuideButton =
        _isTutorialGuideButtonUnlocked(tutorialStatus) &&
        !_tutorialGuideDockVisible;
    final pulseTutorialGuideButton =
        showTutorialGuideButton && !_tutorialGuideButtonAcknowledged;
    final overlayButtonTotalCount =
        overlayButtonCount + (showTutorialGuideButton ? 1 : 0);
    final overlayButtonStackHeight =
        overlayButtonSize * overlayButtonTotalCount +
        overlayButtonSpacing * math.max(0, overlayButtonTotalCount - 1);
    final tutorialScenarioTrackedObjective =
        tutorialStatus?.hasActiveScenario ?? false
        ? tutorialStatus!.resolvedScenarioObjectiveCopy
        : '';
    final tutorialMonsterTrackedObjective =
        tutorialStatus?.hasActiveMonsterEncounter ?? false
        ? tutorialStatus!.resolvedMonsterObjectiveCopy
        : '';
    final authUser = context.watch<AuthProvider>().user;
    final party = context.watch<PartyProvider>().party;
    final hasPartyMapStrip = authUser != null;
    final partyMemberCount = authUser == null
        ? 0
        : <String>{
            authUser.id,
            if (party != null)
              ...party.members
                  .map((member) => member.id)
                  .where((id) => id.isNotEmpty),
          }.length;
    final hasExpandablePartyStrip = partyMemberCount > 1;
    final partyStripClosedWidth = hasPartyMapStrip
        ? partyStripTileWidth +
              (hasExpandablePartyStrip
                  ? partyStripToggleGap + partyStripToggleWidth
                  : 0.0) +
              partyStripHorizontalPadding
        : 0.0;
    final zoneWidgetLeftInset = 16.0 + overlayButtonSize + topOverlayLaneGap;
    final zoneWidgetRightInset =
        16.0 + partyStripClosedWidth + topOverlayLaneGap;
    final zoneWidgetAvailableWidth = math.max(
      0.0,
      screenWidth - zoneWidgetLeftInset - zoneWidgetRightInset,
    );
    final hasTrackedQuestOverlay =
        tutorialScenarioTrackedObjective.isNotEmpty ||
        tutorialMonsterTrackedObjective.isNotEmpty ||
        questLog.quests.any(
          (quest) => questLog.trackedQuestIds.contains(quest.id),
        );
    final polygonActionBottom =
        MediaQuery.paddingOf(context).bottom +
        bottomOverlayPadding +
        (hasTrackedQuestOverlay ? trackedQuestOverlayCollapsedHeight : 0) +
        polygonActionGap;
    final loadingZoneSpinnerDockHeight = hasTrackedQuestOverlay
        ? trackedQuestOverlayCollapsedHeight
        : zoneLoadingSpinnerSize;
    final loadingZoneSpinnerBottom =
        MediaQuery.paddingOf(context).bottom +
        bottomOverlayPadding +
        ((loadingZoneSpinnerDockHeight - zoneLoadingSpinnerSize) / 2);
    Quest? polygonQuest;
    QuestNode? polygonNode;
    Challenge? polygonChallenge;
    MapEntry<Quest, QuestNode>? polygonChallengeQuestEntry;
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

      for (final challenge in _challenges) {
        if (!challenge.hasPolygon) continue;
        if (_usesDedicatedQuestChallengeUi(challenge)) continue;
        if (!_isInsidePolygon(
          loc.latitude,
          loc.longitude,
          challenge.polygonPoints,
        )) {
          continue;
        }
        final activeQuestEntry = _activeQuestNodeForChallenge(challenge.id);
        if (activeQuestEntry != null) {
          polygonChallenge = challenge;
          polygonChallengeQuestEntry = activeQuestEntry;
          break;
        }
        polygonChallenge ??= challenge;
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
        !_markersAdded &&
        !_isPlacingBase) {
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
          MapLibreMap(
            key: ValueKey(_mapKey),
            initialCameraPosition: initialPosition,
            styleString: _stamenWatercolorStyle,
            gestureRecognizers: _mapGestureRecognizers,
            scrollGesturesEnabled: true,
            zoomGesturesEnabled: true,
            rotateGesturesEnabled: true,
            tiltGesturesEnabled: true,
            onMapCreated: (c) {
              debugPrint('SinglePlayer: map created');
              _mapController = c;
              _setupTapHandlers(c);
            },
            onMapClick: _handleMapClick,
            onStyleLoadedCallback: _onMapStyleLoaded,
            myLocationEnabled: false,
            compassEnabled: false,
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
            if (!_isPlacingBase)
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
                                                  color: const Color(
                                                    0xFFB87333,
                                                  ),
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
                                  _OverlayButton(
                                    icon: Icons.refresh,
                                    onTap: _refreshMapContent,
                                    isBusy: _manualRefreshInFlight,
                                  ),
                                  const SizedBox(height: 12),
                                  _OverlayButton(
                                    icon: Icons.my_location,
                                    onTap: _centerOnUserLocation,
                                  ),
                                  if (showTutorialGuideButton) ...[
                                    const SizedBox(height: 12),
                                    AnimatedBuilder(
                                      animation:
                                          _tutorialGuideButtonPulseController,
                                      builder: (context, _) {
                                        return _OverlayPortraitButton(
                                          imageUrl: _tutorialGuidePortraitUrl(
                                            tutorialStatus?.character,
                                          ),
                                          onTap:
                                              _showTutorialGuideButtonInteraction,
                                          pulseProgress:
                                              pulseTutorialGuideButton
                                              ? _tutorialGuideButtonPulseController
                                                    .value
                                              : null,
                                        );
                                      },
                                    ),
                                  ],
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
            if (!_isPlacingBase)
              Positioned(
                left: 16,
                bottom: loadingZoneSpinnerBottom,
                child: IgnorePointer(
                  ignoring: true,
                  child: AnimatedSwitcher(
                    duration: const Duration(milliseconds: 180),
                    switchInCurve: Curves.easeOutCubic,
                    switchOutCurve: Curves.easeInCubic,
                    child: _loadingZoneTransitionZoneId == null
                        ? const SizedBox.shrink(
                            key: ValueKey('zone_transition_hidden'),
                          )
                        : _buildZoneTransitionSpinner(
                            key: ValueKey(_loadingZoneTransitionZoneId),
                          ),
                  ),
                ),
              ),
            if (!_isPlacingBase)
              Positioned(
                top: 0,
                left: zoneWidgetLeftInset,
                right: zoneWidgetRightInset,
                child: PointerInterceptor(
                  child: SafeArea(
                    bottom: false,
                    child: Align(
                      alignment: Alignment.topCenter,
                      child: ConstrainedBox(
                        constraints: BoxConstraints(
                          maxWidth: zoneWidgetAvailableWidth,
                        ),
                        child: ZoneWidget(
                          controller: _zoneWidgetController,
                          expandedHeight: 260,
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            if (!_isPlacingBase && hasPartyMapStrip)
              Positioned(
                top: 0,
                right: 16,
                child: PointerInterceptor(
                  child: SafeArea(
                    bottom: false,
                    child: const Align(
                      alignment: Alignment.topRight,
                      child: PartyMemberMapStrip(),
                    ),
                  ),
                ),
              ),
            if (!_isPlacingBase && polygonQuest != null && polygonNode != null)
              Positioned(
                left: 16,
                right: 16,
                bottom: polygonActionBottom,
                child: FilledButton(
                  onPressed: () => _showQuestObjectiveSubmissionPanel(
                    polygonQuest!,
                    polygonNode!,
                  ),
                  child: Text('Quest: ${polygonQuest.name}'),
                ),
              ),
            if (!_isPlacingBase && polygonChallenge != null)
              Positioned(
                left: 16,
                right: 16,
                bottom: polygonQuest != null && polygonNode != null
                    ? polygonActionBottom + stackedPolygonActionSpacing
                    : polygonActionBottom,
                child: FilledButton(
                  onPressed: () {
                    final challenge = polygonChallenge!;
                    final questEntry = polygonChallengeQuestEntry;
                    if (questEntry != null) {
                      _showStandaloneQuestChallengeSubmissionModal(
                        questEntry.key,
                        questEntry.value,
                        challenge,
                      );
                      return;
                    }
                    _showStandaloneChallengeSubmissionModal(challenge);
                  },
                  child: Text(
                    polygonChallengeQuestEntry != null
                        ? 'Challenge: ${polygonChallengeQuestEntry.key.name}'
                        : 'Challenge Area',
                  ),
                ),
              ),
            if (_isPlacingBase)
              Positioned(
                left: 16,
                right: 16,
                bottom: polygonActionBottom,
                child: PointerInterceptor(
                  child: Card(
                    elevation: 10,
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _pendingBaseInventoryItem == null
                                ? 'Choose a base location'
                                : 'Use ${_pendingBaseInventoryItem!.name}',
                            style: Theme.of(context).textTheme.titleMedium
                                ?.copyWith(fontWeight: FontWeight.w700),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _pendingBaseSelection == null
                                ? 'Tap anywhere on the map to place your base pin.'
                                : 'Does this location look correct? Tap elsewhere to move it.',
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                          const SizedBox(height: 14),
                          Row(
                            children: [
                              Expanded(
                                child: OutlinedButton(
                                  onPressed: _creatingBase
                                      ? null
                                      : _cancelBasePlacement,
                                  child: const Text('Cancel'),
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: FilledButton(
                                  onPressed:
                                      _pendingBaseSelection == null ||
                                          _creatingBase
                                      ? null
                                      : _confirmBasePlacement,
                                  child: Text(
                                    _creatingBase ? 'Creating...' : 'Continue',
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
            if (!_isPlacingBase && hasTrackedQuestOverlay)
              Positioned(
                left: 0,
                right: 0,
                bottom: 0,
                child: PointerInterceptor(
                  child: SafeArea(
                    top: false,
                    child: Padding(
                      padding: const EdgeInsets.only(
                        bottom: bottomOverlayPadding,
                      ),
                      child: Align(
                        alignment: Alignment.bottomCenter,
                        child: TrackedQuestsOverlay(
                          controller: _trackedQuestsController,
                          expandUpwards: true,
                          collapsedHeight: trackedQuestOverlayCollapsedHeight,
                          onFocusPoI: (poi) => _focusQuestPoI(
                            poi,
                            zoom: _trackedQuestOverlayFocusZoom,
                          ),
                          onFocusNode: (node) => _focusQuestNode(
                            node,
                            zoom: _trackedQuestOverlayFocusZoom,
                          ),
                          onFocusTurnInQuest: (quest) => _focusQuestTurnIn(
                            quest,
                            zoom: _trackedQuestOverlayFocusZoom,
                          ),
                          onPreviewPoI: (poi) => _previewTrackedQuestPoi(
                            poi,
                            zoom: _trackedQuestOverlayFocusZoom,
                          ),
                          onPreviewNode: (node) => _focusQuestNode(
                            node,
                            zoom: _trackedQuestOverlayFocusZoom,
                          ),
                          onPreviewTurnInQuest: (quest) => _focusQuestTurnIn(
                            quest,
                            zoom: _trackedQuestOverlayFocusZoom,
                          ),
                          resolveQuestReceiverCharacter:
                              _questReceiverCharacterForQuest,
                          resolveQuestReceiverPoi: _questReceiverPoiForQuest,
                          tutorialScenarioTitle: _tutorialScenarioOverlayTitle(
                            tutorialStatus,
                          ),
                          tutorialScenarioDetail:
                              _tutorialScenarioOverlayDetail(tutorialStatus),
                          tutorialScenarioObjectiveCopy:
                              tutorialScenarioTrackedObjective,
                          onFocusTutorialScenario:
                              tutorialScenarioTrackedObjective.isEmpty
                              ? null
                              : () => _focusTutorialScenarioFromTrackedOverlay(
                                  zoom: _trackedQuestOverlayFocusZoom,
                                ),
                          onPreviewTutorialScenario:
                              tutorialScenarioTrackedObjective.isEmpty
                              ? null
                              : () => _focusTutorialScenarioFromTrackedOverlay(
                                  zoom: _trackedQuestOverlayFocusZoom,
                                ),
                          tutorialMonsterTitle: _tutorialMonsterOverlayTitle(
                            tutorialStatus,
                          ),
                          tutorialMonsterDetail: _tutorialMonsterOverlayDetail(
                            tutorialStatus,
                          ),
                          tutorialMonsterObjectiveCopy:
                              tutorialMonsterTrackedObjective,
                          onFocusTutorialMonster:
                              tutorialMonsterTrackedObjective.isEmpty
                              ? null
                              : () => _focusTutorialMonsterFromTrackedOverlay(
                                  zoom: _trackedQuestOverlayFocusZoom,
                                ),
                          onPreviewTutorialMonster:
                              tutorialMonsterTrackedObjective.isEmpty
                              ? null
                              : () => _focusTutorialMonsterFromTrackedOverlay(
                                  zoom: _trackedQuestOverlayFocusZoom,
                                ),
                          onCloseOverlay:
                              _returnToPlayerFromTrackedQuestOverlay,
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
                                          isLoading
                                              ? 'Submitting...'
                                              : 'Calculating...',
                                          style: Theme.of(
                                            context,
                                          ).textTheme.bodySmall,
                                        ),
                                      ),
                                      if ((_questSubmissionStepLabel ?? '')
                                          .isNotEmpty) ...[
                                        const SizedBox(height: 6),
                                        Center(
                                          child: Text(
                                            _questSubmissionStepLabel!,
                                            textAlign: TextAlign.center,
                                            style: Theme.of(context)
                                                .textTheme
                                                .bodySmall
                                                ?.copyWith(
                                                  color: Theme.of(context)
                                                      .colorScheme
                                                      .onSurface
                                                      .withValues(alpha: 0.72),
                                                ),
                                          ),
                                        ),
                                      ],
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
          if (_tutorialGuideDockVisible)
            Positioned.fill(child: _buildTutorialGuideDockOverlay(context)),
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
                        const SizedBox(height: 18),
                        SizedBox(
                          width: double.infinity,
                          child: FilledButton(
                            onPressed: _dismissTutorialWelcomeOverlay,
                            style: FilledButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 14),
                              backgroundColor: const Color(0xFF8C5A14),
                              foregroundColor: const Color(0xFFF8EFD8),
                            ),
                            child: const Text('Continue'),
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

  Widget _buildTutorialGuideDockOverlay(BuildContext context) {
    final character = _tutorialGuideDockCharacter;
    if (character == null) {
      return const SizedBox.shrink();
    }

    double lerp(num a, num b, double t) => a + ((b - a) * t);

    final theme = Theme.of(context);
    final progress = _tutorialGuideDockController.value.clamp(0.0, 1.0);
    final collapseProgress = (progress / 0.42).clamp(0.0, 1.0);
    final travelProgress = ((progress - 0.42) / 0.58).clamp(0.0, 1.0);
    final collapseCurve = Curves.easeInOutCubic.transform(collapseProgress);
    final travelCurve = Curves.easeInOutCubic.transform(travelProgress);
    final screenSize = MediaQuery.sizeOf(context);
    final safePadding = MediaQuery.paddingOf(context);
    final startWidth = math.min(screenSize.width - 44, 336.0).toDouble();
    const startHeight = 228.0;
    const portraitStartSize = 104.0;
    final targetRect = _tutorialGuideDockTargetRect(context);
    final collapsedLeft = (screenSize.width - portraitStartSize) / 2;
    final collapsedTop = math.max(
      safePadding.top + 32,
      (screenSize.height - portraitStartSize) / 2 - 8,
    );
    final startLeft = (screenSize.width - startWidth) / 2;
    final startTop = math.max(
      safePadding.top + 20,
      (screenSize.height - startHeight) / 2 - 24,
    );

    final width = travelProgress > 0
        ? lerp(portraitStartSize, targetRect.width, travelCurve)
        : lerp(startWidth, portraitStartSize, collapseCurve);
    final height = travelProgress > 0
        ? lerp(portraitStartSize, targetRect.height, travelCurve)
        : lerp(startHeight, portraitStartSize, collapseCurve);
    final left = travelProgress > 0
        ? lerp(collapsedLeft, targetRect.left, travelCurve)
        : lerp(startLeft, collapsedLeft, collapseCurve);
    final top = travelProgress > 0
        ? lerp(collapsedTop, targetRect.top, travelCurve)
        : lerp(startTop, collapsedTop, collapseCurve);
    final shellOpacity = travelProgress > 0 ? 1.0 - travelCurve : 1.0;
    final textOpacity = 1.0 - collapseCurve;
    final portraitScale = travelProgress > 0
        ? lerp(1.0, 0.84, travelCurve)
        : lerp(1.0, 1.06, 1.0 - collapseCurve);

    return IgnorePointer(
      ignoring: true,
      child: Stack(
        children: [
          Positioned(
            left: left,
            top: top,
            width: width,
            height: height,
            child: PaperTexture(
              borderRadius: BorderRadius.circular(28),
              opacity: 0.14 * math.max(0.0, math.min(1.0, shellOpacity)),
              child: Container(
                padding: EdgeInsets.all(travelProgress > 0 ? 0 : 16),
                decoration: BoxDecoration(
                  color: const Color(
                    0xFFF7EBD1,
                  ).withValues(alpha: 0.98 * shellOpacity),
                  borderRadius: BorderRadius.circular(
                    lerp(28, 14, travelCurve).toDouble(),
                  ),
                  border: Border.all(
                    color: const Color(
                      0xFFD2B26C,
                    ).withValues(alpha: 0.9 * shellOpacity),
                    width: 1.3,
                  ),
                  boxShadow: shellOpacity <= 0.05
                      ? const []
                      : [
                          BoxShadow(
                            color: Colors.black.withValues(
                              alpha: 0.26 * shellOpacity,
                            ),
                            blurRadius: 24,
                            offset: const Offset(0, 12),
                          ),
                        ],
                ),
                child: Stack(
                  children: [
                    if (travelProgress == 0)
                      Opacity(
                        opacity: math.max(0.0, math.min(1.0, textOpacity)),
                        child: Align(
                          alignment: Alignment.bottomCenter,
                          child: Padding(
                            padding: const EdgeInsets.only(
                              top: 120,
                              left: 10,
                              right: 10,
                              bottom: 10,
                            ),
                            child: Container(
                              padding: const EdgeInsets.fromLTRB(
                                14,
                                12,
                                14,
                                12,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.white.withValues(alpha: 0.82),
                                borderRadius: BorderRadius.circular(18),
                                border: Border.all(
                                  color: const Color(
                                    0xFFD9C89C,
                                  ).withValues(alpha: 0.92),
                                ),
                              ),
                              child: Text(
                                _tutorialGuideDockExcerpt,
                                maxLines: 3,
                                overflow: TextOverflow.fade,
                                textAlign: TextAlign.center,
                                style: theme.textTheme.bodyMedium?.copyWith(
                                  color: const Color(0xFF4F3B1D),
                                  height: 1.35,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                    Center(
                      child: Transform.scale(
                        scale: portraitScale.toDouble(),
                        child: _buildTutorialGuidePortraitFrame(
                          imageUrl: _tutorialGuidePortraitUrl(character),
                          size: travelProgress > 0 ? width : portraitStartSize,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTutorialGuidePortraitFrame({
    required String imageUrl,
    required double size,
  }) {
    final clampedSize = math.max(size, 40.0).toDouble();
    return Container(
      width: clampedSize,
      height: clampedSize,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E2BC),
        borderRadius: BorderRadius.circular(
          math.min(clampedSize * 0.22, 18).toDouble(),
        ),
        border: Border.all(
          color: const Color(0xFFD2B26C).withValues(alpha: 0.94),
          width: 1.3,
        ),
        boxShadow: const [
          BoxShadow(
            color: Colors.black26,
            blurRadius: 18,
            offset: Offset(0, 8),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(
          math.min(clampedSize * 0.19, 16).toDouble(),
        ),
        child: imageUrl.isNotEmpty
            ? Image.network(
                imageUrl,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) => Icon(
                  Icons.person,
                  size: clampedSize * 0.5,
                  color: const Color(0xFF8C5A14),
                ),
              )
            : Icon(
                Icons.person,
                size: clampedSize * 0.5,
                color: const Color(0xFF8C5A14),
              ),
      ),
    );
  }

  Widget _buildZoneTransitionSpinner({Key? key}) {
    return PaperTexture(
      key: key,
      borderRadius: BorderRadius.circular(999),
      opacity: 0.08,
      child: Container(
        width: 28,
        height: 28,
        decoration: BoxDecoration(
          color: const Color(0xFFF6E8C9).withValues(alpha: 0.94),
          shape: BoxShape.circle,
          border: Border.all(
            color: const Color(0xFFD0AE6B).withValues(alpha: 0.9),
          ),
          boxShadow: const [
            BoxShadow(
              color: Colors.black26,
              blurRadius: 16,
              offset: Offset(0, 8),
            ),
          ],
        ),
        child: const Padding(
          padding: EdgeInsets.all(7),
          child: CircularProgressIndicator(
            strokeWidth: 2.2,
            valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF8A5A14)),
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

  double? _baseDistanceFromCurrentLocation(BasePin base) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return null;
    return _distanceMeters(
      location.latitude,
      location.longitude,
      base.latitude,
      base.longitude,
    );
  }

  void _openBaseManagement(BasePin base) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (sheetContext) => BaseManagementSheet(
        baseId: base.id,
        onClose: () => Navigator.of(sheetContext).pop(),
        onTutorialProgressChanged: _refreshTutorialAfterBaseInteraction,
      ),
    ).whenComplete(() {
      if (!mounted) return;
      unawaited(_loadBases());
    });
  }

  void _showBasePanel(BasePin base) {
    final distance = _baseDistanceFromCurrentLocation(base);
    final canEnterBase =
        distance != null && distance <= kProximityUnlockRadiusMeters;
    if (canEnterBase) {
      _openBaseManagement(base);
      return;
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (sheetContext) => BasePanel(
        base: base,
        onClose: () => Navigator.of(sheetContext).pop(),
        onOpenBaseManagement: () {
          Navigator.of(sheetContext).pop();
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            _openBaseManagement(base);
          });
        },
      ),
    );
  }

  Future<void> _hideMarkersForBasePlacement() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;

    Future<void> removeSymbols(List<Symbol> symbols) async {
      if (symbols.isEmpty) return;
      try {
        await c.removeSymbols(symbols);
      } catch (_) {}
      symbols.clear();
    }

    Future<void> removeCircleMap(
      Map<String, Circle> circles,
      List<Circle> backingList,
    ) async {
      if (circles.isEmpty) return;
      for (final circle in circles.values.toList()) {
        try {
          await c.removeCircle(circle);
        } catch (_) {}
      }
      circles.clear();
      backingList.clear();
    }

    await removeSymbols(_poiSymbols);
    _poiSymbolById.clear();
    _questPoiHighlightSymbols.clear();
    _questPoiHighlightCircles.clear();
    _questPoiPulseTimer?.cancel();
    _questPoiPulseTimer = null;
    await removeSymbols(_characterSymbols);
    _characterSymbolsById.clear();
    await removeSymbols(_chestSymbols);
    _chestSymbolById.clear();
    await removeCircleMap(_chestCircleById, _chestCircles);
    await removeSymbols(_healingFountainSymbols);
    _healingFountainSymbolById.clear();
    await removeCircleMap(_healingFountainCircleById, _healingFountainCircles);
    await removeSymbols(_resourceSymbols);
    _resourceSymbolById.clear();
    await removeCircleMap(_resourceCircleById, _resourceCircles);
    await removeSymbols(_baseSymbols);
    _baseSymbolById.clear();
    await removeCircleMap(_baseCircleById, _baseCircles);
    await removeSymbols(_scenarioSymbols);
    _scenarioSymbolById.clear();
    await removeCircleMap(_scenarioCircleById, _scenarioCircles);
    await removeSymbols(_expositionSymbols);
    _expositionSymbolById.clear();
    await removeCircleMap(_expositionCircleById, _expositionCircles);
    await removeSymbols(_monsterSymbols);
    _monsterSymbolById.clear();
    await removeCircleMap(_monsterCircleById, _monsterCircles);
    await removeSymbols(_challengeSymbols);
    _challengeSymbolById.clear();
    await removeCircleMap(_challengeCircleById, _challengeCircles);
    _markersAdded = false;
  }

  Future<void> _restoreMarkersAfterBasePlacement() async {
    if (!_styleLoaded || _mapController == null) return;
    await _addPoiMarkers();
    await _refreshDiscoveredPoiMarkers();
    _applyQuestLogOverlaysIfChanged();
  }

  Future<void> _clearBasePlacementPreview() async {
    final c = _mapController;
    final symbol = _basePlacementPreviewSymbol;
    _basePlacementPreviewSymbol = null;
    if (c == null || symbol == null) return;
    try {
      await c.removeSymbols([symbol]);
    } catch (_) {}
  }

  Future<void> _refreshBasePlacementPreview() async {
    if (!_isPlacingBase || _pendingBaseSelection == null) {
      await _clearBasePlacementPreview();
      return;
    }

    final c = _mapController;
    if (c == null || !_styleLoaded) return;

    _basePlacementPreviewBytes ??= await loadBaseDiamondMarker(
      isCurrentUserBase: true,
    );
    final previewBytes = _basePlacementPreviewBytes;
    if (previewBytes == null) return;

    const imageId = 'base_placement_preview_diamond_self_v6';
    await _ensureMapImage(c, imageId, previewBytes);

    final selection = _pendingBaseSelection!;
    final existing = _basePlacementPreviewSymbol;
    if (existing == null) {
      final symbol = await c.addSymbol(
        SymbolOptions(
          geometry: selection,
          iconImage: imageId,
          iconSize: _basePlacementPreviewIconSize,
          iconOpacity: 1,
          iconHaloColor: _transparentMapHaloColor,
          iconHaloWidth: 0.0,
          iconAnchor: 'center',
          zIndex: 7,
        ),
        const {'type': 'basePlacementPreview'},
      );
      if (!mounted) return;
      _basePlacementPreviewSymbol = symbol;
      return;
    }

    try {
      await c.updateSymbol(
        existing,
        SymbolOptions(
          geometry: selection,
          iconImage: imageId,
          iconSize: _basePlacementPreviewIconSize,
          iconOpacity: 1,
        ),
      );
    } catch (_) {}
  }

  Future<void> _beginBasePlacement(
    OwnedInventoryItem owned,
    InventoryItem item,
  ) async {
    if (!mounted) return;
    setState(() {
      _pendingBaseOwnedInventoryItemId = owned.id;
      _pendingBaseInventoryItem = item;
      _pendingBaseSelection = null;
      _creatingBase = false;
    });
    _zoneWidgetController.close();
    final loc = context.read<LocationProvider>().location;
    final mapController = _mapController;
    if (loc != null && mapController != null) {
      final lat = loc.latitude;
      final lng = loc.longitude;
      if (lat.isFinite && lng.isFinite && lat.abs() <= 90 && lng.abs() <= 180) {
        try {
          await mapController.animateCamera(
            CameraUpdate.newCameraPosition(
              CameraPosition(target: LatLng(lat, lng), zoom: 15.5),
            ),
            duration: const Duration(milliseconds: 450),
          );
        } catch (_) {}
      }
    }
    await _hideMarkersForBasePlacement();
    await _refreshBasePlacementPreview();
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Tap the map to choose where to establish your base.'),
      ),
    );
  }

  void _cancelBasePlacement() {
    if (!mounted) return;
    final wasPlacingBase = _isPlacingBase;
    setState(() {
      _pendingBaseOwnedInventoryItemId = null;
      _pendingBaseInventoryItem = null;
      _pendingBaseSelection = null;
      _creatingBase = false;
    });
    if (wasPlacingBase) {
      unawaited(_clearBasePlacementPreview());
      unawaited(_restoreMarkersAfterBasePlacement());
    }
  }

  Future<void> _confirmBasePlacement() async {
    final ownedInventoryItemId = _pendingBaseOwnedInventoryItemId;
    final selection = _pendingBaseSelection;
    if (ownedInventoryItemId == null || selection == null || _creatingBase) {
      return;
    }

    final existingBase = _currentUserBase();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(
          existingBase == null ? 'Establish base here?' : 'Move base here?',
        ),
        content: Text(
          existingBase == null
              ? 'This will consume the item and place your base at the selected location.'
              : 'This will consume the item and move your current base pin to the selected location.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Submit'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    setState(() {
      _creatingBase = true;
    });

    try {
      await context.read<InventoryService>().useItem(
        ownedInventoryItemId,
        baseLatitude: selection.latitude,
        baseLongitude: selection.longitude,
      );
      if (!mounted) return;
      await _loadBases();
      if (!mounted) return;
      _cancelBasePlacement();
      await _loadTutorialStatus(force: true, preserveCompletedReveal: true);
      if (!mounted) return;
      final mapController = _mapController;
      if (mapController != null) {
        await mapController.animateCamera(CameraUpdate.newLatLng(selection));
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            existingBase == null
                ? 'Base established.'
                : 'Base moved to the new location.',
          ),
        ),
      );
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _creatingBase = false;
      });
      var message = e.toString();
      if (e is DioException) {
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final error = data['error'];
          if (error is String && error.trim().isNotEmpty) {
            message = error;
          }
        }
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(message)));
    }
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
      useSafeArea: false,
      barrierDismissible: true,
      barrierColor: Colors.transparent,
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
        onStatusChanged: (updatedFountain) {
          if (!mounted) return;
          _syncHealingFountainState(updatedFountain);
          unawaited(_refreshHealingFountainSymbols());
        },
        onUnlocked: (unlockedFountain) async {
          final zoneProvider = context.read<ZoneProvider>();
          if (!mounted) return;
          final currentFountain = _healingFountainById(fountain.id) ?? fountain;
          _syncHealingFountainState(
            currentFountain.copyWith(
              discovered: true,
              thumbnailUrl: unlockedFountain.thumbnailUrl,
            ),
          );
          await _refreshHealingFountainSymbols();
          _invalidateZoneBaseContent(fountain.zoneId);
          final selectedZoneId = zoneProvider.selectedZone?.id;
          if (selectedZoneId == fountain.zoneId) {
            await _loadTreasureChestsForSelectedZone();
          }
        },
        onUsed: (result) {
          if (!mounted) return;
          final lastUsedAt = DateTime.tryParse(
            result['lastUsedAt']?.toString() ?? '',
          )?.toLocal();
          final nextAvailableAt = DateTime.tryParse(
            result['nextAvailableAt']?.toString() ?? '',
          )?.toLocal();
          final currentFountain = _healingFountainById(fountain.id) ?? fountain;

          _syncHealingFountainState(
            currentFountain.copyWith(
              availableNow: false,
              lastUsedAt: lastUsedAt,
              nextAvailableAt: nextAvailableAt,
              cooldownSecondsRemaining:
                  (result['cooldownSecondsRemaining'] as num?)?.toInt() ?? 0,
            ),
          );

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

  void _showResourcePanel(ResourceNode resource) {
    final parentContext = context;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => ResourcePanel(
        resource: resource,
        onClose: () => Navigator.of(context).pop(),
        onGathered: (rewardData) {
          if (!mounted) return;
          setState(() {
            _gatheredResourceIds.add(resource.id);
            _resources = _resources
                .where((item) => item.id != resource.id)
                .toList(growable: false);
          });
          _invalidateZoneBaseContent(resource.zoneId);
          unawaited(_refreshResourceSymbols());
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            final completedTaskProvider = parentContext
                .read<CompletedTaskProvider>();
            completedTaskProvider.showModal(
              'resourceGathered',
              data: rewardData,
            );
            _queueRewardLevelUpFromData(rewardData);
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
            final completedTaskProvider = parentContext
                .read<CompletedTaskProvider>();
            completedTaskProvider.showModal(
              'treasureChestOpened',
              data: rewardData,
            );
            _queueRewardLevelUpFromData(rewardData);
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

  MapEntry<Quest, QuestNode>? _activeQuestNodeForScenario(String scenarioId) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node == null) continue;
      if (node.scenarioId == scenarioId) {
        return MapEntry(quest, node);
      }
    }
    return null;
  }

  MapEntry<Quest, QuestNode>? _activeQuestNodeForExposition(
    String expositionId,
  ) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node == null) continue;
      if (node.expositionId == expositionId) {
        return MapEntry(quest, node);
      }
    }
    return null;
  }

  MapEntry<Quest, QuestNode>? _activeQuestNodeForMonsterEncounter(
    String encounterId,
  ) {
    final questLog = context.read<QuestLogProvider>();
    for (final quest in questLog.quests) {
      if (!quest.isAccepted) continue;
      final node = quest.currentNode;
      if (node == null) continue;
      if (node.monsterEncounterId == encounterId) {
        return MapEntry(quest, node);
      }
      if (node.monsterId?.trim().isNotEmpty == true) {
        final encounter = _monsterEncounterByMemberMonsterId(node.monsterId!);
        if (encounter != null && encounter.id == encounterId) {
          return MapEntry(quest, node);
        }
      }
    }
    return null;
  }

  void _showQuestObjectiveSubmissionPanel(Quest quest, QuestNode node) {
    final fetchCharacterId = node.fetchCharacterId?.trim() ?? '';
    if (fetchCharacterId.isNotEmpty) {
      unawaited(_openFetchQuestTurnInScreen(quest));
      return;
    }
    final scenarioId = node.scenarioId?.trim() ?? '';
    if (scenarioId.isNotEmpty) {
      final scenario = _scenarioById(scenarioId);
      if (scenario != null) {
        _showScenarioPanel(scenario);
        return;
      }
    }
    final expositionId = node.expositionId?.trim() ?? '';
    if (expositionId.isNotEmpty) {
      final exposition = node.exposition ?? _expositionById(expositionId);
      if (exposition != null) {
        unawaited(
          _showExpositionDialogue(exposition, quest: quest, node: node),
        );
        return;
      }
    }
    final challengeId = node.challengeId?.trim() ?? '';
    if (challengeId.isNotEmpty) {
      final challenge = _challengeById(challengeId);
      if (challenge != null) {
        _showChallengePanel(challenge);
        return;
      }
    }
    final encounterId = node.monsterEncounterId?.trim() ?? '';
    if (encounterId.isNotEmpty) {
      final encounter = _monsterById(encounterId);
      if (encounter != null) {
        _showMonsterPanel(encounter);
        return;
      }
    }
    _showQuestNodeSubmissionModal(quest.name, node);
  }

  Future<void> _openFetchQuestTurnInScreen(Quest quest) async {
    final delivered = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => FetchQuestTurnInScreen(questId: quest.id),
      ),
    );
    if (delivered != true || !mounted) return;
    await context.read<QuestLogProvider>().refresh();
  }

  Character? _defaultExpositionSpeaker(Exposition exposition) {
    for (final message in exposition.dialogue) {
      final characterId = message.characterId?.trim() ?? '';
      if (characterId.isEmpty) continue;
      final character = _characterById(characterId);
      if (character != null) {
        return character;
      }
    }
    return null;
  }

  Map<String, Character> _speakerCharacterMapForExposition(
    Exposition exposition,
  ) {
    final speakers = <String, Character>{};
    for (final message in exposition.dialogue) {
      final characterId = message.characterId?.trim() ?? '';
      if (characterId.isEmpty || speakers.containsKey(characterId)) continue;
      final character = _characterById(characterId);
      if (character != null) {
        speakers[characterId] = character;
      }
    }
    return speakers;
  }

  double? _expositionDistanceMeters(Exposition exposition) {
    final location = context.read<LocationProvider>().location;
    if (location == null) return null;
    return _distanceMeters(
      location.latitude,
      location.longitude,
      exposition.latitude,
      exposition.longitude,
    );
  }

  String _expositionDisplayTitle(Exposition exposition) {
    final title = exposition.title.trim();
    if (title.isNotEmpty) return title;
    return 'Nearby Dialogue';
  }

  Future<void> _showExpositionTooFarDialog(
    Exposition exposition,
    double distanceMeters,
  ) {
    final title = _expositionDisplayTitle(exposition);
    return showModalBottomSheet<void>(
      context: context,
      isScrollControlled: false,
      useSafeArea: true,
      backgroundColor: Colors.transparent,
      builder: (sheetContext) {
        final theme = Theme.of(sheetContext);
        final colorScheme = theme.colorScheme;
        return AdaptivePaperSheet(
          maxHeightFactor: 0.4,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
          header: Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    title,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                IconButton(
                  onPressed: () => Navigator.of(sheetContext).pop(),
                  icon: const Icon(Icons.close),
                  visualDensity: VisualDensity.compact,
                  splashRadius: 20,
                ),
              ],
            ),
          ),
          body: Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  'You are too far away to hear this conversation.',
                  textAlign: TextAlign.center,
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: colorScheme.onSurface.withValues(alpha: 0.8),
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(height: 16),
                Wrap(
                  alignment: WrapAlignment.center,
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _MiniInfoChip(
                      icon: Icons.place_outlined,
                      label: '${distanceMeters.round()} m away',
                    ),
                    _MiniInfoChip(
                      icon: Icons.hearing_outlined,
                      label: 'Need ${kProximityUnlockRadiusMeters.round()} m',
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Future<void> _removeExpositionLocally(String expositionId) async {
    final trimmedId = expositionId.trim();
    if (trimmedId.isEmpty) return;

    if (mounted) {
      setState(() {
        _expositions.removeWhere((item) => item.id == trimmedId);
      });
    } else {
      _expositions.removeWhere((item) => item.id == trimmedId);
    }

    await _refreshExpositionSymbols();
  }

  Future<void> _showExpositionDialogue(
    Exposition exposition, {
    Quest? quest,
    QuestNode? node,
  }) async {
    final distanceMeters = _expositionDistanceMeters(exposition);
    if (distanceMeters != null &&
        distanceMeters > kProximityUnlockRadiusMeters) {
      await _showExpositionTooFarDialog(exposition, distanceMeters);
      return;
    }

    final defaultSpeaker = _defaultExpositionSpeaker(exposition);
    if (defaultSpeaker == null) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Dialogue participants are still loading.'),
        ),
      );
      return;
    }

    final action = CharacterAction(
      id: 'exposition-${exposition.id}',
      createdAt: '',
      updatedAt: '',
      characterId: defaultSpeaker.id,
      actionType: 'exposition',
      dialogue: List<DialogueMessage>.from(exposition.dialogue)
        ..sort((a, b) => a.order.compareTo(b.order)),
    );
    final speakerCharacters = _speakerCharacterMapForExposition(exposition);
    final parentContext = context;

    await showDialog<void>(
      context: context,
      useRootNavigator: true,
      useSafeArea: false,
      barrierDismissible: false,
      barrierColor: Colors.transparent,
      builder: (dialogContext) {
        return PopScope(
          canPop: false,
          child: RpgDialogueModal(
            character: defaultSpeaker,
            action: action,
            dialogueOverride: action.dialogue,
            showCloseButton: false,
            finalStepLabel: 'Complete',
            speakerCharacters: speakerCharacters,
            onClose: () async {
              Navigator.of(dialogContext).pop();
              final previousLevel = parentContext
                  .read<CharacterStatsProvider>()
                  .level;
              try {
                final result = await parentContext
                    .read<PoiService>()
                    .performExposition(exposition.id);
                await _removeExpositionLocally(exposition.id);
                _invalidateZoneBaseContent(exposition.zoneId);
                await parentContext.read<QuestLogProvider>().refresh();
                final currentLevel = await _refreshRewardDrivenPlayerState();
                if (!mounted || !parentContext.mounted) return;
                parentContext.read<CompletedTaskProvider>().showModal(
                  'expositionOutcome',
                  data: {
                    'expositionId': result.expositionId,
                    'title': result.title.isNotEmpty
                        ? result.title
                        : exposition.title,
                    'questName': quest?.name,
                    'questNodeId': node?.id,
                    'successful': result.successful,
                    'rewardExperience': result.rewardExperience,
                    'rewardGold': result.rewardGold,
                    'baseResourcesAwarded': result.baseResourcesAwarded,
                    'itemsAwarded': result.itemsAwarded,
                    'spellsAwarded': result.spellsAwarded
                        .map((spell) => spell.toJson())
                        .toList(),
                    'awardedRewards': result.awardedRewards,
                  },
                );
                parentContext
                    .read<CompletedTaskProvider>()
                    .queueLevelUpFollowUpIfNeeded(
                      previousLevel: previousLevel,
                      currentLevel: currentLevel,
                    );
              } catch (error) {
                if (!mounted || !parentContext.mounted) return;
                ScaffoldMessenger.of(parentContext).showSnackBar(
                  SnackBar(
                    content: Text(
                      PoiService.extractApiErrorMessage(
                        error,
                        'Failed to complete exposition.',
                      ),
                    ),
                  ),
                );
              }
            },
          ),
        );
      },
    );
  }

  void _showChallengePanel(Challenge challenge) {
    final activeQuestEntry = _activeQuestNodeForChallenge(challenge.id);
    final questName = activeQuestEntry?.key.name;
    final anchor = _challengeProximityAnchor(challenge);
    final location = context.read<LocationProvider>().location;
    final distance = location == null || challenge.hasPolygon
        ? null
        : _distanceMeters(
            location.latitude,
            location.longitude,
            anchor.latitude,
            anchor.longitude,
          );
    final withinRange = location != null
        ? (challenge.hasPolygon
              ? _isInsidePolygon(
                  location.latitude,
                  location.longitude,
                  challenge.polygonPoints,
                )
              : distance != null && distance <= kProximityUnlockRadiusMeters)
        : false;
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
                    if (!mysteryState) ...[
                      Hero(
                        tag: _challengeImageHeroTag(challenge.id),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(14),
                          child: AspectRatio(
                            aspectRatio: 1,
                            child: Image.network(
                              challenge.thumbnailUrl.isNotEmpty
                                  ? challenge.thumbnailUrl
                                  : challenge.imageUrl,
                              fit: BoxFit.cover,
                              errorBuilder: (_, _, _) => Container(
                                color: theme.colorScheme.surfaceVariant,
                                child: const Icon(Icons.auto_awesome_outlined),
                              ),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                    ],
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
      polygon: challenge.hasPolygon ? challenge.polygonPoints : node.polygon,
      objective: QuestNodeObjective(
        id: challenge.id,
        type: QuestNodeObjective.typeChallenge,
        prompt: challenge.question,
        description: challenge.description,
        imageUrl: challenge.imageUrl,
        thumbnailUrl: challenge.thumbnailUrl,
        reward: challenge.reward,
        inventoryItemId: challenge.inventoryItemId,
        submissionType: submissionType,
        difficulty: challenge.difficulty,
        statTags: challenge.statTags,
        proficiency: challenge.proficiency,
      ),
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
      polygon: challenge.polygonPoints,
      objective: QuestNodeObjective(
        id: challenge.id,
        type: QuestNodeObjective.typeChallenge,
        prompt: challenge.question,
        description: challenge.description,
        imageUrl: challenge.imageUrl,
        thumbnailUrl: challenge.thumbnailUrl,
        reward: challenge.reward,
        inventoryItemId: challenge.inventoryItemId,
        submissionType: submissionType,
        difficulty: challenge.difficulty,
        statTags: challenge.statTags,
        proficiency: challenge.proficiency,
      ),
    );
    return _showQuestNodeSubmissionModal(
      'Challenge',
      syntheticNode,
      standaloneChallengeId: challenge.id,
      standaloneChallengeZoneId: challenge.zoneId,
      challengeImageHeroTag: challengeImageHeroTag,
    );
  }

  void _invalidateZoneBaseContent(String zoneId) {
    final normalized = zoneId.trim();
    if (normalized.isEmpty) return;
    _zoneBaseContentCache.remove(normalized);
    _zoneBaseContentRequests.remove(normalized);
  }

  Future<void> _refreshStandaloneChallengeZoneContent(String? zoneId) async {
    final normalized = zoneId?.trim() ?? '';
    if (normalized.isEmpty) return;
    _invalidateZoneBaseContent(normalized);
    final selectedZoneId = context.read<ZoneProvider>().selectedZone?.id;
    if (selectedZoneId != normalized) return;
    await _loadTreasureChestsForSelectedZone();
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

  Future<MonsterEncounter> _hydrateMonsterEncounterForBattle(
    MonsterEncounter encounter,
    String battleMonsterId,
  ) async {
    final refreshedMonster = await context.read<PoiService>().getMonsterById(
      battleMonsterId,
    );
    if (refreshedMonster == null) {
      return encounter;
    }

    final sourceMembers = encounter.members.isNotEmpty
        ? encounter.members
        : encounter.monsters
              .asMap()
              .entries
              .map(
                (entry) => MonsterEncounterMember(
                  slot: entry.key + 1,
                  monster: entry.value,
                ),
              )
              .toList(growable: false);

    final hydratedMembers = sourceMembers
        .map((member) {
          if (member.monster.id != battleMonsterId) {
            return member;
          }
          return MonsterEncounterMember(
            slot: member.slot,
            monster: refreshedMonster,
          );
        })
        .toList(growable: false);

    final resolvedMembers = hydratedMembers.isNotEmpty
        ? hydratedMembers
        : <MonsterEncounterMember>[
            MonsterEncounterMember(slot: 1, monster: refreshedMonster),
          ];
    final resolvedMonsters = resolvedMembers
        .map((member) => member.monster)
        .toList(growable: false);

    return MonsterEncounter(
      id: encounter.id,
      name: encounter.name,
      description: encounter.description,
      imageUrl: encounter.imageUrl,
      thumbnailUrl: encounter.thumbnailUrl,
      encounterType: encounter.encounterType,
      rewardMode: encounter.rewardMode,
      randomRewardSize: encounter.randomRewardSize,
      rewardExperience: encounter.rewardExperience,
      rewardGold: encounter.rewardGold,
      itemRewards: encounter.itemRewards,
      zoneId: encounter.zoneId,
      pointOfInterestId: encounter.pointOfInterestId,
      latitude: encounter.latitude,
      longitude: encounter.longitude,
      monsterCount: resolvedMonsters.length,
      members: resolvedMembers,
      monsters: resolvedMonsters,
    );
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
        battleDetail = await poiService.startMonsterBattle(
          battleMonsterId,
          monsterEncounterId: monster.id.trim().isNotEmpty ? monster.id : null,
        );
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
    final battleEncounter = await _hydrateMonsterEncounterForBattle(
      monster,
      battleMonsterId,
    );
    try {
      result = await showDialog<MonsterBattleResult>(
        context: parentContext,
        useRootNavigator: true,
        useSafeArea: false,
        barrierDismissible: false,
        builder: (_) => MonsterBattleDialog(
          encounter: battleEncounter,
          isPartyBattle: effectivePartyBattle,
          battleMonsterId: battleMonsterId,
          battleId: resolvedBattleId.isNotEmpty ? resolvedBattleId : null,
        ),
      );
    } finally {
      if (!effectivePartyBattle) {
        try {
          await poiService.endMonsterBattle(
            battleMonsterId,
            outcome: result?.outcome.name,
          );
        } catch (error) {
          debugPrint('Failed to end server monster battle: $error');
        }
      }
    }

    if (!mounted || result == null) return;
    debugPrint(
      '[monster-rewards][client][result] '
      'encounter=${monster.id} outcome=${result.outcome} '
      'rewardExperience=${result.rewardExperience} rewardGold=${result.rewardGold} '
      'itemsAwarded=${result.itemsAwarded.length}',
    );
    final statsProvider = context.read<CharacterStatsProvider>();
    final previousLevel = statsProvider.level;
    final questLogProvider = context.read<QuestLogProvider>();

    if (result.outcome == MonsterBattleOutcome.escaped) {
      await statsProvider.setHealthAndManaTo(
        health: result.playerHealthRemaining,
        mana: result.playerManaRemaining,
      );
      return;
    }

    final defeatHealthSetTo = _monsterBattleDefeatResourceFloor(
      statsProvider.maxHealth,
      _monsterBattleDefeatHealthFloorPercent,
    );
    final defeatManaSetTo = _monsterBattleDefeatResourceFloor(
      statsProvider.maxMana,
      _monsterBattleDefeatManaFloorPercent,
    );
    final healthSetTo = result.outcome == MonsterBattleOutcome.defeat
        ? math.max(result.playerHealthRemaining, defeatHealthSetTo)
        : result.playerHealthRemaining;
    final manaSetTo = result.outcome == MonsterBattleOutcome.defeat
        ? math.max(result.playerManaRemaining, defeatManaSetTo)
        : result.playerManaRemaining;

    await statsProvider.setHealthAndManaTo(
      health: healthSetTo,
      mana: manaSetTo,
    );

    if (result.outcome == MonsterBattleOutcome.victory) {
      await _refreshRewardDrivenPlayerState();
    }

    if (result.outcome == MonsterBattleOutcome.victory) {
      final wasTutorialMonster = _isTutorialFocusedMonsterEncounterId(
        monster.id,
      );
      await _removeMonsterEncounterLocally(monster.id);
      if (!mounted) return;
      setState(() {
        _defeatedMonsterIds.add(monster.id);
        if (wasTutorialMonster) {
          _tutorialFocusedMonsterEncounterId = null;
          _tutorialPostMonsterDialoguePendingAfterCompletionModal = true;
        }
      });
      await _persistDefeatedMonsterIds();
      _skipQuestLogMapRefreshCount += 1;
      await questLogProvider.refresh();
      if (_skipQuestLogMapRefreshCount > 0) {
        _skipQuestLogMapRefreshCount = 0;
      }
      if (!mounted || !parentContext.mounted) return;

      final completedTaskProvider = parentContext.read<CompletedTaskProvider>();
      completedTaskProvider.showModal(
        'monsterBattleVictory',
        data: {
          'monsterEncounterId': monster.id,
          'monsterName': monster.name,
          'rewardExperience': result.rewardExperience,
          'rewardGold': result.rewardGold,
          'baseResourcesAwarded': result.baseResourcesAwarded,
          'itemsAwarded': result.itemsAwarded,
        },
      );
      completedTaskProvider.queueLevelUpFollowUpIfNeeded(
        previousLevel: previousLevel,
        currentLevel: statsProvider.level,
      );
      return;
    }

    if (!mounted || !parentContext.mounted) return;
    parentContext.read<CompletedTaskProvider>().showModal(
      'monsterBattleDefeat',
      data: {
        'monsterName': monster.name,
        'healthSetTo': healthSetTo,
        'manaSetTo': manaSetTo,
        'statusName': 'Wounded',
        'statusDurationMinutes': _monsterBattleDefeatStatusDurationMinutes,
      },
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
          final previousLevel = context.read<CharacterStatsProvider>().level;

          await _removeScenarioLocally(
            scenario.id,
            performedScenarioId: result.scenarioId,
            fallbackScenario: scenario,
          );
          unawaited(_loadTreasureChestsForSelectedZone());
          final currentLevel = await _refreshRewardDrivenPlayerState();
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
            final completedTaskProvider = parentContext
                .read<CompletedTaskProvider>();
            completedTaskProvider.showModal(
              'scenarioOutcome',
              data: {
                'scenarioId': result.scenarioId,
                'scenarioPrompt': scenario.prompt,
                'successful': result.successful,
                'outcomeText': outcomeText,
                'questHandoffs': result.questHandoffs
                    .map((handoff) => handoff.toJson())
                    .toList(),
                'reason': result.reason,
                'roll': result.roll,
                'statTag': result.statTag,
                'statValue': result.statValue,
                'proficiencies': result.proficiencies,
                'proficiencyBonus': result.proficiencyBonus,
                'responseScore': result.responseScore,
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
                'baseResourcesAwarded': result.baseResourcesAwarded,
                'itemsAwarded': result.itemsAwarded,
                'itemChoiceRewards': result.itemChoiceRewards,
                'spellsAwarded': result.spellsAwarded
                    .map((spell) => spell.toJson())
                    .toList(),
              },
            );
            completedTaskProvider.queueLevelUpFollowUpIfNeeded(
              previousLevel: previousLevel,
              currentLevel: currentLevel,
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
      builder: (context) => AdaptivePaperSheet(
        maxHeightFactor: 0.95,
        header: Container(
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
        body: const Padding(
          padding: EdgeInsets.all(16),
          child: ActivityFeedPanel(),
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
      builder: (context) => AdaptivePaperSheet(
        maxHeightFactor: 0.95,
        header: Container(
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
        body: const Padding(padding: EdgeInsets.all(16), child: LogPanel()),
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
        color: theme.colorScheme.surfaceContainerHighest.withValues(
          alpha: 0.55,
        ),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: theme.colorScheme.outline.withValues(alpha: 0.2),
        ),
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

class _OverlayButton extends StatelessWidget {
  const _OverlayButton({
    required this.icon,
    required this.onTap,
    this.isBusy = false,
  });

  final IconData icon;
  final VoidCallback onTap;
  final bool isBusy;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final surfaceColor = theme.colorScheme.surface.withValues(alpha: 0.95);
    final borderColor = theme.colorScheme.outlineVariant;
    final foregroundColor = theme.colorScheme.onSurface;
    return Material(
      color: surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: borderColor),
      ),
      child: InkWell(
        onTap: isBusy ? null : onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: SizedBox(
            width: 24,
            height: 24,
            child: isBusy
                ? CircularProgressIndicator(
                    strokeWidth: 2.4,
                    color: foregroundColor,
                  )
                : Icon(icon, size: 24, color: foregroundColor),
          ),
        ),
      ),
    );
  }
}

class _OverlayPortraitButton extends StatelessWidget {
  const _OverlayPortraitButton({
    required this.imageUrl,
    required this.onTap,
    this.pulseProgress,
  });

  final String imageUrl;
  final VoidCallback onTap;
  final double? pulseProgress;

  Widget _buildPulseRing(double progress) {
    final curved = Curves.easeOutCubic.transform(progress.clamp(0.0, 1.0));
    final scale = 1.0 + (0.55 * curved);
    final opacity = (1.0 - curved).clamp(0.0, 1.0);
    return Transform.scale(
      scale: scale,
      child: Container(
        width: 56,
        height: 56,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: const Color(0xFFF5C542).withValues(alpha: 0.12 * opacity),
          border: Border.all(
            color: const Color(0xFFD2B26C).withValues(alpha: 0.82 * opacity),
            width: 1.5,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFFF4D989).withValues(alpha: 0.24 * opacity),
              blurRadius: 14,
              spreadRadius: 1.5,
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final surfaceColor = theme.colorScheme.surface.withValues(alpha: 0.95);
    final borderColor = theme.colorScheme.outlineVariant;
    final primaryPulseProgress = pulseProgress?.clamp(0.0, 1.0);
    final secondaryPulseProgress = primaryPulseProgress == null
        ? null
        : ((primaryPulseProgress + 0.5) % 1.0);
    return Stack(
      clipBehavior: Clip.none,
      alignment: Alignment.center,
      children: [
        if (secondaryPulseProgress != null)
          IgnorePointer(child: _buildPulseRing(secondaryPulseProgress)),
        if (primaryPulseProgress != null)
          IgnorePointer(child: _buildPulseRing(primaryPulseProgress)),
        Material(
          color: surfaceColor,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: BorderSide(color: borderColor),
          ),
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(12),
            child: Padding(
              padding: const EdgeInsets.all(4),
              child: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: const Color(0xFFF3E2BC),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: const Color(0xFFD2B26C).withValues(alpha: 0.92),
                  ),
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(9),
                  child: imageUrl.isNotEmpty
                      ? Image.network(
                          imageUrl,
                          fit: BoxFit.cover,
                          errorBuilder: (_, _, _) => const Icon(Icons.person),
                        )
                      : const Icon(Icons.person),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _MainStoryLead {
  const _MainStoryLead({
    required this.poi,
    this.character,
    this.distanceMeters,
  });

  final PointOfInterest poi;
  final Character? character;
  final double? distanceMeters;
}

class _FeaturedMainStoryPulseTarget {
  const _FeaturedMainStoryPulseTarget.poi(this.poiId) : characterId = null;

  const _FeaturedMainStoryPulseTarget.character(this.characterId)
    : poiId = null;

  final String? poiId;
  final String? characterId;
}

class _MapMarkerIsolation {
  const _MapMarkerIsolation({this.markerKeys = const <String>{}});

  final Set<String> markerKeys;
}

class _MapPinAnnotationSeed {
  const _MapPinAnnotationSeed({
    required this.type,
    required this.id,
    required this.geometry,
    required this.hitRadiusPx,
  });

  final String type;
  final String id;
  final LatLng geometry;
  final double hitRadiusPx;
}

class _MapPinSelectionCandidate {
  const _MapPinSelectionCandidate({
    required this.type,
    required this.id,
    required this.title,
    required this.imageUrl,
    required this.distance,
    this.useBaseDiamondMarker = false,
    this.isCurrentUserBase = false,
  });

  final String type;
  final String id;
  final String title;
  final String imageUrl;
  final double distance;
  final bool useBaseDiamondMarker;
  final bool isCurrentUserBase;
}

class _PoiSymbolRequest {
  const _PoiSymbolRequest({
    required this.poiId,
    required this.isQuestCurrent,
    required this.shouldPulseLikeQuest,
    required this.shouldPulseLikeMainStory,
    required this.options,
    required this.data,
  });

  final String poiId;
  final bool isQuestCurrent;
  final bool shouldPulseLikeQuest;
  final bool shouldPulseLikeMainStory;
  final SymbolOptions options;
  final Map<String, dynamic> data;
}

class _PoiImageUpdate {
  const _PoiImageUpdate({
    required this.poi,
    required this.isQuestCurrent,
    required this.hasMapContent,
    required this.hasQuestMarker,
    required this.hasMainStoryAccent,
    required this.undiscovered,
  });

  final PointOfInterest poi;
  final bool isQuestCurrent;
  final bool hasMapContent;
  final bool hasQuestMarker;
  final bool hasMainStoryAccent;
  final bool undiscovered;
}

class _PoiImageUpdateResult {
  const _PoiImageUpdateResult(this.update, this.imageId, this.bytes);

  final _PoiImageUpdate update;
  final String? imageId;
  final Uint8List? bytes;
}

class _ZonePinContent {
  const _ZonePinContent({
    required this.pointsOfInterest,
    required this.characters,
  });

  final List<PointOfInterest> pointsOfInterest;
  final List<Character> characters;
}

class _ZonePinContentCacheEntry {
  _ZonePinContentCacheEntry({
    required this.content,
    DateTime? fetchedAt,
    DateTime? lastAccessedAt,
  }) : fetchedAt = fetchedAt ?? DateTime.now(),
       lastAccessedAt = lastAccessedAt ?? DateTime.now();

  final _ZonePinContent content;
  final DateTime fetchedAt;
  DateTime lastAccessedAt;

  bool get isFresh =>
      DateTime.now().difference(fetchedAt) <= _zoneBaseContentFreshDuration;

  void touch() {
    lastAccessedAt = DateTime.now();
  }
}

class _ZoneBaseContent {
  const _ZoneBaseContent({
    required this.treasureChests,
    required this.healingFountains,
    required this.resources,
    required this.scenarios,
    required this.expositions,
    required this.monsters,
    required this.challenges,
  });

  final List<TreasureChest> treasureChests;
  final List<HealingFountain> healingFountains;
  final List<ResourceNode> resources;
  final List<Scenario> scenarios;
  final List<Exposition> expositions;
  final List<MonsterEncounter> monsters;
  final List<Challenge> challenges;
}

class _ZoneBaseContentCacheEntry {
  _ZoneBaseContentCacheEntry({
    required this.content,
    DateTime? fetchedAt,
    DateTime? lastAccessedAt,
  }) : fetchedAt = fetchedAt ?? DateTime.now(),
       lastAccessedAt = lastAccessedAt ?? DateTime.now();

  final _ZoneBaseContent content;
  final DateTime fetchedAt;
  DateTime lastAccessedAt;

  bool get isFresh =>
      DateTime.now().difference(fetchedAt) <= _zoneBaseContentFreshDuration;

  void touch() {
    lastAccessedAt = DateTime.now();
  }
}
