import 'dart:async';
import 'dart:math' show Point;
import 'dart:typed_data';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
import 'package:pointer_interceptor/pointer_interceptor.dart';
import 'package:provider/provider.dart';
import '../config/router.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../models/treasure_chest.dart';
import '../models/zone.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/log_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/quest_filter_provider.dart';
import '../providers/tags_provider.dart';
import '../providers/zone_provider.dart';
import '../providers/map_focus_provider.dart';
import '../services/media_service.dart';
import '../services/poi_service.dart';
import '../utils/poi_image_util.dart';
import '../utils/camera_capture.dart';
import '../constants/api_constants.dart';
import '../widgets/activity_feed_panel.dart';
import '../widgets/celebration_modal_manager.dart';
import '../widgets/character_panel.dart';
import '../widgets/inventory_panel.dart';
import '../widgets/log_panel.dart';
import '../widgets/new_item_modal.dart';
import '../widgets/point_of_interest_panel.dart';
import '../widgets/quest_log_panel.dart';
import '../widgets/rpg_dialogue_modal.dart';
import '../widgets/tracked_quests_overlay.dart';
import '../widgets/shop_modal.dart';
import '../widgets/quest_filter_panel.dart';
import '../widgets/treasure_chest_panel.dart';
import '../widgets/used_item_modal.dart';
import '../widgets/zone_widget.dart';
import '../widgets/paper_texture.dart';

const _chestImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';
const _mapThumbnailVersion = 'v4';
const _stamenWatercolorStyleBase =
    'https://tiles.stadiamaps.com/styles/stamen_watercolor.json';
const _stamenWatercolorApiKey =
    String.fromEnvironment('STADIA_MAPS_API_KEY', defaultValue: '');
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
  List<Line> _zoneLines = [];
  List<Fill> _zoneFills = [];
  List<Line> _questLines = [];
  List<Fill> _questFills = [];
  List<Symbol> _poiSymbols = [];
  final Map<String, Symbol> _poiSymbolById = {};
  List<Symbol> _questPoiHighlightSymbols = [];
  List<Symbol> _characterSymbols = [];
  List<Symbol> _chestSymbols = [];
  List<Circle> _chestCircles = [];
  final Map<String, Symbol> _chestSymbolById = {};
  final Map<String, Circle> _chestCircleById = {};
  final Map<String, bool> _chestCircleOpened = {};
  final ZoneWidgetController _zoneWidgetController = ZoneWidgetController();
  Uint8List? _chestThumbnailBytes;
  bool _chestThumbnailAdded = false;
  bool _styleLoaded = false;
  bool _markersAdded = false;
  bool _addedMarkersWithEmptyDiscoveries = false;
  bool _mapLoadFailed = false;
  int _mapKey = 0;
  bool _hasAnimatedToUserLocation = false;
  QuestLogProvider? _questLogProvider;
  MapFocusProvider? _mapFocusProvider;
  Timer? _questGlowTimer;
  bool _isQuestGlowPulsing = false;
  Timer? _questPoiPulseTimer;
  bool _isQuestPoiPulseActive = false;
  bool _questPoiPulseUp = false;
  Timer? _zoneAutoSelectTimer;
  String? _questLogRequestedZoneId;
  bool _questLogRefreshInFlight = false;
  DateTime? _lastQuestLogRefreshAt;
  bool _questLogNeedsOverlayApply = false;
  Set<String> _lastQuestPoiIds = <String>{};
  int _lastQuestPolygonHash = 0;
  String _lastMapFilterKey = '';
  QuestSubmissionOverlayPhase _questSubmissionPhase = QuestSubmissionOverlayPhase.hidden;
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
      _updateSelectedZoneFromLocation();
      _requestQuestLogIfReady();
      context.read<ActivityFeedProvider>().refresh();
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
    super.dispose();
  }

  void _onZoneChanged() {
    if (!mounted) return;
    unawaited(_loadTreasureChestsForSelectedZone());
    unawaited(_addZoneBoundaries());
    _requestQuestLogIfReady();
  }

  void _onLocationChanged() {
    if (!mounted) return;
    _updateSelectedZoneFromLocation();
    _requestQuestLogIfReady();
  }

  void _onAuthChanged() {
    if (!mounted) return;
    _requestQuestLogIfReady(force: true);
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
          orElse: () => PointOfInterest(
            id: '',
            name: '',
            lat: '0',
            lng: '0',
          ),
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
    return [
      filters.showCurrentQuestPoints,
      filters.showQuestAvailablePoints,
      filters.showTreasureChests,
    ].join('|');
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

  Set<String> _currentQuestTurnInCharacterIds(QuestLogProvider questLog) {
    return questLog.quests
        .where((q) => q.readyToTurnIn && q.questGiverCharacterId != null)
        .map((q) => q.questGiverCharacterId!)
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
      _chestThumbnailBytes = null;
      _chestThumbnailAdded = false;
      _questLines = [];
    });
    _startMapLoadTimeout();
  }

  void _onMapStyleLoaded() {
    debugPrint('SinglePlayer: map style loaded');
    _mapLoadTimeout?.cancel();
    _mapLoadTimeout = null;
    if (!mounted) return;
    setState(() => _styleLoaded = true);
    unawaited((() async {
      await _setSymbolOverlap();
      await _addPoiMarkers();
    })());
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
    if (!mounted || !_styleLoaded || _mapLoadFailed || _hasAnimatedToUserLocation) return;
    final c = _mapController;
    final loc = context.read<LocationProvider>().location;
    if (c == null || loc == null) return;
    final lat = loc.latitude;
    final lng = loc.longitude;
    if (!lat.isFinite || !lng.isFinite || lat.abs() > 90 || lng.abs() > 180) return;
    _hasAnimatedToUserLocation = true;
    setState(() {});
    Future.delayed(const Duration(milliseconds: 400), () {
      if (!mounted) return;
      final controller = _mapController;
      if (controller == null) return;
      try {
        controller.animateCamera(
          CameraUpdate.newCameraPosition(CameraPosition(
            target: LatLng(lat, lng),
            zoom: 15,
          )),
          duration: const Duration(milliseconds: 600),
        );
      } catch (_) {}
    });
  }

  void _flyToLocation(double lat, double lng) {
    final c = _mapController;
    if (c == null || !lat.isFinite || !lng.isFinite ||
        lat.abs() > 90 || lng.abs() > 180) return;
    try {
      c.animateCamera(
        CameraUpdate.newCameraPosition(CameraPosition(
          target: LatLng(lat, lng),
          zoom: 16,
        )),
        duration: const Duration(milliseconds: 500),
      );
    } catch (_) {}
  }

  void _focusQuestPoI(PointOfInterest poi) {
    final lat = double.tryParse(poi.lat) ?? 0.0;
    final lng = double.tryParse(poi.lng) ?? 0.0;
    _flyToLocation(lat, lng);
    final hasDiscovered = context.read<DiscoveriesProvider>().hasDiscovered(poi.id);
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

  Future<void> _pulsePolygon(List<QuestNodePolygonPoint> polygon) async {
    final c = _mapController;
    if (c == null || polygon.length < 3) return;
    final ring = polygon
        .map((p) => LatLng(p.latitude, p.longitude))
        .toList();
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

  Future<void> _pulseQuestGlow(List<List<QuestNodePolygonPoint>> polygons) async {
    if (_isQuestGlowPulsing) return;
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    _isQuestGlowPulsing = true;
    final options = <LineOptions>[];
    for (final poly in polygons) {
      if (poly.length < 3) continue;
      final ring = poly
          .map((p) => LatLng(p.latitude, p.longitude))
          .toList();
      if (ring.length > 1 &&
          (ring.first.latitude != ring.last.latitude ||
              ring.first.longitude != ring.last.longitude)) {
        ring.add(ring.first);
      }
      options.add(LineOptions(
        geometry: ring,
        lineColor: '#f5c542',
        lineWidth: 8.0,
        lineOpacity: 0.35,
      ));
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
    try {
      await context.read<DiscoveriesProvider>().refresh();
      final zones = await svc.getZones();
      final pois = await svc.getPointsOfInterest();
      final characters = await svc.getCharacters();
      if (!mounted) return;
      debugPrint('SinglePlayer: _loadAll data: zones=${zones.length} pois=${pois.length} chars=${characters.length}');
      final zoneProvider = context.read<ZoneProvider>();
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

  Future<void> _loadTreasureChestsForSelectedZone() async {
    final zoneId = context.read<ZoneProvider>().selectedZone?.id ??
        (_zones.isNotEmpty ? _zones.first.id : null);
    if (zoneId == null) {
      if (mounted) setState(() => _treasureChests = []);
      return;
    }
    try {
      final chests = await context.read<PoiService>().getTreasureChestsForZone(zoneId);
      if (!mounted) return;
      setState(() {
        _treasureChests = chests;
      });
      if (_styleLoaded && _mapController != null && _markersAdded) {
        await _refreshTreasureChestSymbols();
      }
    } catch (e) {
      debugPrint('SinglePlayer: _loadTreasureChests error: $e');
      if (mounted) {
        setState(() => _treasureChests = []);
        if (_styleLoaded && _mapController != null && _markersAdded) {
          await _refreshTreasureChestSymbols();
        }
      }
    }
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
      if (raw == null || raw is! Map) return;
      final data = Map<String, dynamic>.from(raw as Map<dynamic, dynamic>);
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
      if (type == 'poi') {
        final pois = _pois.where((p) => p.id == idStr).toList();
        if (pois.isNotEmpty && mounted) {
          _showPointOfInterestPanel(pois.first, context.read<DiscoveriesProvider>().hasDiscovered(idStr));
        }
      }
    });
    c.onSymbolTapped.add((symbol) {
      try {
        debugPrint('SinglePlayer: onSymbolTapped');
        final raw = symbol.data;
        if (raw == null || raw is! Map) {
          debugPrint('SinglePlayer: symbol tap data is null or not Map');
          return;
        }
        final data = Map<String, dynamic>.from(raw as Map<dynamic, dynamic>);
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
            _showPointOfInterestPanel(pois.first, context.read<DiscoveriesProvider>().hasDiscovered(idStr));
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
        debugPrint('SinglePlayer: symbol tap POI found id=$idStr mounted=$mounted');
        if (!mounted) {
          debugPrint('SinglePlayer: symbol tap unmounted');
          return;
        }
        debugPrint('SinglePlayer: showing POI panel ${pois.first.name}');
        _showPointOfInterestPanel(pois.first, context.read<DiscoveriesProvider>().hasDiscovered(idStr));
      } catch (e, st) {
        debugPrint('SinglePlayer: symbol tap error: $e');
        debugPrint('SinglePlayer: symbol tap stack: $st');
      }
    });
    c.onFillTapped.add((fill) {
      final raw = fill.data;
      if (raw == null || raw is! Map) return;
      final data = Map<String, dynamic>.from(raw as Map<dynamic, dynamic>);
      final type = data['type'] as String?;
      final idStr = data['id']?.toString();
      if (type == 'zone' && idStr != null && idStr.isNotEmpty) {
        _selectZoneById(idStr);
      }
    });
    c.onLineTapped.add((line) {
      final raw = line.data;
      if (raw == null || raw is! Map) return;
      final data = Map<String, dynamic>.from(raw as Map<dynamic, dynamic>);
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
      orElse: () => const Zone(
        id: '',
        name: '',
        latitude: 0,
        longitude: 0,
      ),
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
        onClose: () => Navigator.of(context).pop(),
        onQuestSubmissionState: _setQuestSubmissionOverlay,
        onCharacterTap: (character) {
          Navigator.of(context).pop();
          _showCharacterPanel(character);
        },
        onUnlocked: () async {
          await context.read<DiscoveriesProvider>().refresh();
          if (!mounted) return;
          setState(() => _markersAdded = false);
          await _addPoiMarkers();
        },
      ),
    );
  }

  Future<void> _addPoiMarkers() async {
    if (!_styleLoaded || _markersAdded) {
      debugPrint('SinglePlayer: _addPoiMarkers skip (styleLoaded=$_styleLoaded markersAdded=$_markersAdded)');
      return;
    }
    final c = _mapController;
    if (c == null) {
      debugPrint('SinglePlayer: _addPoiMarkers skip (no controller)');
      return;
    }
    _markersAdded = true;
    debugPrint('SinglePlayer: _addPoiMarkers start (pois=${_pois.length} chars=${_characters.length} chests=${_treasureChests.length})');

    try {
      final questLog = context.read<QuestLogProvider>();
      final questPoiIds = _currentQuestPoiIdsForFilter(questLog);
      final turnInCharacterIds = _currentQuestTurnInCharacterIds(questLog);
      final filters = context.read<QuestFilterProvider>();
      final categoryFilterActive =
          filters.showCurrentQuestPoints || filters.showQuestAvailablePoints;
      final anyFilterActive = categoryFilterActive;
      debugPrint('SinglePlayer: _addPoiMarkers questPoiIds=${questPoiIds.length}');
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
      try {
        await c.clearCircles();
      } catch (_) {}

      Uint8List? placeholderBytes;
      Uint8List? questPlaceholderBytes;
      Uint8List? availablePlaceholderBytes;
      try {
        placeholderBytes = await loadPoiThumbnail(null);
        if (placeholderBytes != null) {
          await c.addImage('poi_placeholder_$_mapThumbnailVersion', placeholderBytes);
        }
      } catch (_) {}
      try {
        questPlaceholderBytes = await loadPoiThumbnailWithBorder(null);
        if (questPlaceholderBytes != null) {
          await c.addImage('poi_placeholder_quest_$_mapThumbnailVersion', questPlaceholderBytes);
        }
      } catch (_) {}
      try {
        availablePlaceholderBytes = await loadPoiThumbnailWithQuestMarker(null);
        if (availablePlaceholderBytes != null) {
          await c.addImage(
            'poi_placeholder_available_$_mapThumbnailVersion',
            availablePlaceholderBytes,
          );
        }
      } catch (_) {}

      Uint8List? chestBytes;
      try {
        chestBytes = await loadPoiThumbnail(_chestImageUrl);
        if (chestBytes != null) {
          _chestThumbnailBytes = chestBytes;
          _chestThumbnailAdded = true;
          await c.addImage('chest_thumbnail_$_mapThumbnailVersion', chestBytes);
        }
      } catch (_) {}

      for (final ch in _characters) {
        final matchesAvailable = filters.showQuestAvailablePoints && ch.hasAvailableQuest;
        final matchesTurnIn = filters.showCurrentQuestPoints && turnInCharacterIds.contains(ch.id);
        if (anyFilterActive && !(matchesAvailable || matchesTurnIn)) {
          continue;
        }
        final points = ch.locations.isNotEmpty
            ? ch.locations
                .map((loc) => LatLng(loc.latitude, loc.longitude))
                .where((p) => p.latitude != 0 || p.longitude != 0)
                .toList()
            : (ch.lat == 0 && ch.lng == 0)
                ? <LatLng>[]
                : [LatLng(ch.lat, ch.lng)];

        if (points.isEmpty) continue;

        final thumbnailUrl = ch.thumbnailUrl;
        final hasQuestAvailable = ch.hasAvailableQuest;
        if (thumbnailUrl != null && thumbnailUrl.isNotEmpty) {
          try {
            final imageBytes = hasQuestAvailable
                ? await loadPoiThumbnailWithQuestMarker(thumbnailUrl)
                : await loadPoiThumbnail(thumbnailUrl);
            if (imageBytes != null) {
              final imageId =
                  hasQuestAvailable ? 'character_${ch.id}_quest' : 'character_${ch.id}';
              final versionedId = '${imageId}_$_mapThumbnailVersion';
              try {
                await c.addImage(versionedId, imageBytes);
              } catch (_) {}
              for (final point in points) {
                final sym = await c.addSymbol(
                  SymbolOptions(
                    geometry: point,
                    iconImage: versionedId,
                    iconSize: 0.6,
                    iconHaloColor: '#000000',
                    iconHaloWidth: 0.75,
                    iconAnchor: 'center',
                  ),
                  {'type': 'character', 'id': ch.id, 'name': ch.name},
                );
                if (!mounted) return;
                _characterSymbols.add(sym);
              }
              continue;
            }
          } catch (_) {}
        }

        for (final point in points) {
          c.addCircle(
            CircleOptions(
              geometry: point,
              circleRadius: 30,
              circleColor: '#ff8833',
              circleStrokeWidth: 2,
              circleStrokeColor: '#ffffff',
            ),
            {'type': 'character', 'id': ch.id, 'name': ch.name},
          );
        }
      }
      if (filters.showTreasureChests) {
        for (final tc in _treasureChests) {
          if (tc.openedByUser == true) continue;
          if (chestBytes != null) {
            final sym = await c.addSymbol(
              SymbolOptions(
                geometry: LatLng(tc.latitude, tc.longitude),
                iconImage: 'chest_thumbnail_$_mapThumbnailVersion',
                iconSize: 0.75,
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
      }

      final discoveries = context.read<DiscoveriesProvider>();
      final hadEmptyDiscoveries = discoveries.discoveries.isEmpty;
      for (final poi in _pois) {
        final lat = double.tryParse(poi.lat) ?? 0.0;
        final lng = double.tryParse(poi.lng) ?? 0.0;
        final useRealImage = discoveries.hasDiscovered(poi.id);
        final undiscovered = !useRealImage;
        final isQuestCurrent = questPoiIds.contains(poi.id);
        final hasQuestAvailable = poi.hasAvailableQuest;
        final hasCharacter = poi.characters.isNotEmpty;
        final baseEligible = hasCharacter || hasQuestAvailable || isQuestCurrent;
        if (!baseEligible) {
          continue;
        }
        final categoryMatch = !categoryFilterActive ||
            (filters.showCurrentQuestPoints && isQuestCurrent) ||
            (filters.showQuestAvailablePoints && hasQuestAvailable);
        if (!categoryMatch) {
          continue;
        }
        var added = false;
        try {
          String? imageId;
          Uint8List? imageBytes;
          if (isQuestCurrent) {
            imageBytes = await loadPoiThumbnailWithBorder(useRealImage ? poi.imageURL : null);
            if (imageBytes != null) {
              imageId = 'poi_${poi.id}_quest';
            } else {
              imageId = questPlaceholderBytes != null ? 'poi_placeholder_quest' : null;
            }
          } else if (hasQuestAvailable) {
            imageBytes = await loadPoiThumbnailWithQuestMarker(useRealImage ? poi.imageURL : null);
            if (imageBytes != null) {
              imageId = 'poi_${poi.id}_available';
            } else {
              imageId = availablePlaceholderBytes != null ? 'poi_placeholder_available' : null;
            }
          } else if (useRealImage) {
            imageBytes = await loadPoiThumbnail(poi.imageURL);
            imageId = imageBytes != null ? 'poi_${poi.id}' : (placeholderBytes != null ? 'poi_placeholder' : null);
          } else {
            imageId = placeholderBytes != null ? 'poi_placeholder' : null;
          }
          if (imageId != null) {
            final versionedId = '${imageId}_$_mapThumbnailVersion';
            if (imageBytes != null) await c.addImage(versionedId, imageBytes);
            final sym = await c.addSymbol(
              SymbolOptions(
                geometry: LatLng(lat, lng),
                iconImage: versionedId,
                iconSize: isQuestCurrent ? 0.82 : 0.75,
                iconHaloColor: isQuestCurrent ? '#000000' : '#000000',
                iconHaloWidth: isQuestCurrent ? 0.0 : 0.75,
                iconHaloBlur: 0.0,
                iconOpacity: isQuestCurrent ? 1.0 : (undiscovered ? 0.5 : 1.0),
                iconAnchor: 'center',
                zIndex: 2,
              ),
              {'type': 'poi', 'id': poi.id, 'name': poi.name},
            );
            if (!mounted) return;
            _poiSymbols.add(sym);
            _poiSymbolById[poi.id] = sym;
            if (isQuestCurrent) {
              _setQuestPoiHighlight(sym, true);
            }
            added = true;
          }
        } catch (_) {}
        if (!added) {
          c.addCircle(
            CircleOptions(
              geometry: LatLng(lat, lng),
              circleRadius: 24,
              circleColor: '#3388ff',
              circleOpacity: isQuestCurrent ? 1.0 : (undiscovered ? 0.5 : 1.0),
              circleStrokeWidth: 2,
              circleStrokeColor: isQuestCurrent ? '#f5c542' : '#ffffff',
            ),
            {'type': 'poi', 'id': poi.id, 'name': poi.name},
          );
        }
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

  Future<void> _refreshTreasureChestSymbols() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) return;
    final filters = context.read<QuestFilterProvider>();
    if (!filters.showTreasureChests) {
      for (final sym in _chestSymbols) {
        try {
          await c.removeSymbols([sym]);
        } catch (_) {}
      }
      for (final circle in _chestCircles) {
        try {
          await c.removeCircle(circle);
        } catch (_) {}
      }
      if (!mounted) return;
      _chestSymbols.clear();
      _chestCircles.clear();
      _chestSymbolById.clear();
      _chestCircleById.clear();
      _chestCircleOpened.clear();
      return;
    }

    if (_chestThumbnailBytes == null) {
      try {
        _chestThumbnailBytes = await loadPoiThumbnail(_chestImageUrl);
      } catch (_) {}
    }
    if (_chestThumbnailBytes != null && !_chestThumbnailAdded) {
      try {
        await c.addImage('chest_thumbnail_$_mapThumbnailVersion', _chestThumbnailBytes!);
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
              ),
            );
          } catch (_) {}
        }
      } else {
        final existing = _chestCircleById[tc.id];
        final opened = tc.openedByUser == true;
        final needsUpdate = existing == null || _chestCircleOpened[tc.id] != opened;
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

  List<LatLng> _zoneRing(Zone z) {
    final ring = z.ring;
    if (ring != null && ring.isNotEmpty) {
      final list = ring.map((c) => LatLng(c.latitude, c.longitude)).toList();
      if (list.length > 1 && (list.first.latitude != list.last.latitude || list.first.longitude != list.last.longitude)) {
        list.add(list.first);
      }
      return list;
    }
    // boundary is a WKT string, not usable directly - rely on points/boundaryCoords
    return [];
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
      debugPrint('SinglePlayer: _addZoneBoundaries skip (controller=${c != null} styleLoaded=$_styleLoaded)');
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
        fillOptions.add(FillOptions(
          geometry: [ring],
          fillColor: _earthToneForZone(z, salt: i),
          fillOpacity: 0.4,
        ));
        fillData.add({'type': 'zone', 'id': z.id});
      }
      options.add(LineOptions(
        geometry: ring,
        lineColor: '#000000',
        lineWidth: 7.0,
        lineOpacity: 0.18,
        lineBlur: 1.6,
        lineJoin: 'round',
      ));
      lineData.add({'type': 'zone', 'id': z.id});
      options.add(LineOptions(
        geometry: ring,
        lineColor: '#000000',
        lineWidth: 2.8,
        lineOpacity: 0.95,
        lineJoin: 'round',
      ));
      lineData.add({'type': 'zone', 'id': z.id});
    }
    debugPrint('SinglePlayer: _addZoneBoundaries zones=${_zones.length} rings=${options.length}');
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
      debugPrint('SinglePlayer: _addZoneBoundaries added ${lines.length} lines');
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
      final ring = poly
          .map((p) => LatLng(p.latitude, p.longitude))
          .toList();
      if (ring.length > 1 &&
          (ring.first.latitude != ring.last.latitude ||
              ring.first.longitude != ring.last.longitude)) {
        ring.add(ring.first);
      }
      fillOptions.add(FillOptions(
        geometry: [ring],
        fillColor: '#f5c542',
        fillOpacity: 0.5,
      ));
      options.add(LineOptions(
        geometry: ring,
        lineColor: '#f5c542',
        lineWidth: 3.0,
        lineOpacity: 1.0,
      ));
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

  void _ensureQuestPoiPulseTimer() {
    if (_questPoiHighlightSymbols.isEmpty) {
      _questPoiPulseTimer?.cancel();
      _questPoiPulseTimer = null;
      return;
    }
    _questPoiPulseTimer ??=
        Timer.periodic(const Duration(milliseconds: 1200), (_) {
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
    final categoryFilterActive =
        filters.showCurrentQuestPoints || filters.showQuestAvailablePoints;
    final discoveries = context.read<DiscoveriesProvider>();
    final useRealImage = discoveries.hasDiscovered(poi.id);
    final undiscovered = !useRealImage;
    final hasQuestAvailable = poi.hasAvailableQuest;
    final hasCharacter = poi.characters.isNotEmpty;
    final baseEligible = hasCharacter || hasQuestAvailable || isQuestCurrent;
    final categoryMatch = !categoryFilterActive ||
        (filters.showCurrentQuestPoints && isQuestCurrent) ||
        (filters.showQuestAvailablePoints && hasQuestAvailable);
    if (!(baseEligible && categoryMatch)) {
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

    String? imageId;
    Uint8List? imageBytes;
    if (isQuestCurrent) {
      imageBytes = await loadPoiThumbnailWithBorder(useRealImage ? poi.imageURL : null);
      imageId = imageBytes != null ? 'poi_${poi.id}_quest' : 'poi_placeholder_quest';
    } else if (hasQuestAvailable) {
      imageBytes = await loadPoiThumbnailWithQuestMarker(useRealImage ? poi.imageURL : null);
      imageId =
          imageBytes != null ? 'poi_${poi.id}_available' : 'poi_placeholder_available';
    } else if (useRealImage) {
      imageBytes = await loadPoiThumbnail(poi.imageURL);
      imageId = imageBytes != null ? 'poi_${poi.id}' : 'poi_placeholder';
    } else {
      imageId = 'poi_placeholder';
    }

    if (imageId == null) return;
    final versionedId = '${imageId}_$_mapThumbnailVersion';
    if (imageBytes == null) {
      try {
        if (imageId == 'poi_placeholder') {
          imageBytes = await loadPoiThumbnail(null);
        } else if (imageId == 'poi_placeholder_quest') {
          imageBytes = await loadPoiThumbnailWithBorder(null);
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
            iconOpacity: isQuestCurrent ? 1.0 : (undiscovered ? 0.5 : 1.0),
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
          iconOpacity: isQuestCurrent ? 1.0 : (undiscovered ? 0.5 : 1.0),
          iconAnchor: 'center',
          zIndex: 2,
        ),
      );
      _setQuestPoiHighlight(sym, isQuestCurrent);
    } catch (_) {}
  }

  Future<void> _pulseQuestPoiBorders() async {
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
          SymbolOptions(
            iconSize: targetSize,
            iconOpacity: targetOpacity,
          ),
        );
      }
    } catch (_) {}
    _isQuestPoiPulseActive = false;
  }

  bool _isInsidePolygon(double lat, double lng, List<QuestNodePolygonPoint> polygon) {
    if (polygon.length < 3) return false;
    var inside = false;
    for (var i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
      final xi = polygon[i].longitude;
      final yi = polygon[i].latitude;
      final xj = polygon[j].longitude;
      final yj = polygon[j].latitude;
      final intersect = ((yi > lat) != (yj > lat)) &&
          (lng < (xj - xi) * (lat - yi) / (yj - yi + 0.0) + xi);
      if (intersect) inside = !inside;
    }
    return inside;
  }

  Future<void> _showQuestNodeSubmissionModal(Quest quest, QuestNode node) async {
    final textController = TextEditingController();
    CapturedImage? capturedImage;
    bool uploadingImage = false;
    String? selectedChallengeId = node.challenges.isNotEmpty
        ? node.challenges.first.id
        : null;
    final questLogProvider = context.read<QuestLogProvider>();
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
              final canUseCamera = kIsWeb ||
                  defaultTargetPlatform == TargetPlatform.iOS ||
                  defaultTargetPlatform == TargetPlatform.android;
              final selectedChallenge = node.challenges.isEmpty
                  ? null
                  : (selectedChallengeId == null
                      ? node.challenges.first
                      : node.challenges.firstWhere(
                          (c) => c.id == selectedChallengeId,
                          orElse: () => node.challenges.first,
                        ));
              final statValues = context.watch<CharacterStatsProvider>().stats;
              final statTags = (selectedChallenge?.statTags ?? const [])
                  .map((tag) => tag.trim().toLowerCase())
                  .where((tag) => tag.isNotEmpty)
                  .toList();
              final difficultyValue = selectedChallenge?.difficulty ?? 0;
              final statAverage = _averageStatValue(statValues, statTags);
              final difficultyColor =
                  _difficultyColor(statAverage, difficultyValue);
              return Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    quest.name,
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
                        final value = statValues[tag] ??
                            CharacterStatsProvider.baseStatValue;
                        return Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: Theme.of(context)
                                .colorScheme
                                .surfaceVariant
                                .withOpacity(0.6),
                            borderRadius: BorderRadius.circular(999),
                            border: Border.all(
                              color: Theme.of(context)
                                  .colorScheme
                                  .outlineVariant,
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
                  TextField(
                    controller: textController,
                    decoration: const InputDecoration(
                      labelText: 'Answer',
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 3,
                  ),
                  const SizedBox(height: 12),
                  if (canUseCamera)
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: uploadingImage
                                ? null
                                : () async {
                                    final result = await captureImageFromCamera();
                                    if (!mounted) return;
                                    if (result == null || result.bytes.isEmpty) {
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        const SnackBar(
                                          content: Text('No photo captured.'),
                                        ),
                                      );
                                      return;
                                    }
                                    setModalState(() => capturedImage = result);
                                  },
                            icon: const Icon(Icons.photo_camera),
                            label: const Text('Take photo'),
                          ),
                        ),
                        if (capturedImage != null) ...[
                          const SizedBox(width: 12),
                          TextButton(
                            onPressed: () => setModalState(() => capturedImage = null),
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
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: uploadingImage
                        ? null
                        : () async {
                            final startedAt = DateTime.now();
                            setModalState(() => uploadingImage = true);
                            Navigator.of(context).pop();
                            _setQuestSubmissionOverlay(
                              QuestSubmissionOverlayPhase.loading,
                            );
                            String? imageSubmissionUrl;
                            if (capturedImage != null) {
                              final ext = _extensionFromMime(
                                    capturedImage!.mimeType,
                                    capturedImage!.name,
                                  ) ??
                                  'jpg';
                              final key =
                                  'quest-submissions/$userId/${DateTime.now().millisecondsSinceEpoch}.$ext';
                              final url = await mediaService.getPresignedUploadUrl(
                                ApiConstants.crewPointsOfInterestBucket,
                                key,
                              );
                              if (url == null) {
                                final elapsed = DateTime.now().difference(startedAt);
                                if (elapsed < const Duration(milliseconds: 700)) {
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
                                final elapsed = DateTime.now().difference(startedAt);
                                if (elapsed < const Duration(milliseconds: 700)) {
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
                            final resp = await questLogProvider.submitQuestNodeChallenge(
                              node.id,
                              questNodeChallengeId: selectedChallengeId,
                              textSubmission: textController.text.trim(),
                              imageSubmissionUrl: imageSubmissionUrl,
                            );
                            final elapsed = DateTime.now().difference(startedAt);
                            if (elapsed < const Duration(milliseconds: 700)) {
                              await Future<void>.delayed(
                                const Duration(milliseconds: 700),
                              );
                            }
                            final success = resp['successful'] == true;
                            final reason = resp['reason']?.toString() ?? '';
                            final score = (resp['score'] as num?)?.toInt();
                            final difficulty = (resp['difficulty'] as num?)?.toInt();
                            final combined = (resp['combinedScore'] as num?)?.toInt();
                            final statTags = (resp['statTags'] as List?)
                                ?.map((tag) => tag.toString())
                                .toList();
                            final statValues = (resp['statValues'] as Map?)
                                ?.map((key, value) =>
                                    MapEntry(key.toString(), (value as num?)?.toInt() ?? 0));
                            final baseMessage = success
                                ? (reason.isNotEmpty ? reason : 'Challenge completed!')
                                : (reason.isNotEmpty ? reason : 'Submission failed');
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
                          },
                    child: Text('Quest: ${quest.name}'),
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
    final zone = zoneProvider.findZoneAtCoordinate(location.latitude, location.longitude);
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
    final hasDetails = score != null ||
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
    final loc = context.watch<LocationProvider>().location;
    final discoveries = context.watch<DiscoveriesProvider>();
    final questLog = context.watch<QuestLogProvider>();
    final lat = loc?.latitude ?? 0.0;
    final lng = loc?.longitude ?? 0.0;
    final initialPosition = CameraPosition(
      target: LatLng(lat, lng),
      zoom: 15,
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

    if (_styleLoaded && !_mapLoadFailed && _mapController != null && loc != null && !_hasAnimatedToUserLocation) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        _animateToUserLocationIfReady();
      });
    }

    // Retry adding markers/zones when we have style + data but haven't added yet (e.g. style loaded after _loadAll)
    if (_styleLoaded && !_mapLoadFailed && _mapController != null && _zones.isNotEmpty && !_markersAdded) {
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
        setState(() {
          _markersAdded = false;
          _addedMarkersWithEmptyDiscoveries = false;
        });
        unawaited(_addPoiMarkers());
        unawaited(_addQuestPolygons());
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
                    'SinglePlayer: map pointer down at ${event.position}');
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
              compassEnabled: true,
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
                          'SinglePlayer: top controls pointer down at ${event.position}');
                    }
                  },
                  child: SafeArea(
                    bottom: false,
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
                      child: SizedBox(
                        width: double.infinity,
                        child: Stack(
                          clipBehavior: Clip.none,
                          alignment: Alignment.topCenter,
                          children: [
                            ZoneWidget(controller: _zoneWidgetController),
                            Positioned(
                              left: 0,
                              top: 0,
                              child: Consumer<ActivityFeedProvider>(
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
                                                'SinglePlayer: notifications tapped');
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
                                                color: Theme.of(context)
                                                    .colorScheme
                                                    .surface,
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
                            ),
                            Positioned(
                              right: 0,
                              top: 0,
                              child: Consumer2<QuestFilterProvider, TagsProvider>(
                                builder: (context, filters, tags, _) {
                                  final hasActiveFilters =
                                      filters.showCurrentQuestPoints ||
                                          filters.showQuestAvailablePoints ||
                                          !filters.showTreasureChests ||
                                          (filters.enableTagFilter &&
                                              tags.selectedTagIds.isNotEmpty);
                                  return Stack(
                                    clipBehavior: Clip.none,
                                    children: [
                                      _OverlayButton(
                                        icon: Icons.tune,
                                        onTap: () {
                                          if (kDebugMode) {
                                            debugPrint(
                                                'SinglePlayer: filters tapped');
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
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            // Tracked quests: below zone to avoid overlap
            Positioned(
              top: 142,
              right: 16,
              child: PointerInterceptor(
                child: TrackedQuestsOverlay(
                  controller: _trackedQuestsController,
                  onFocusPoI: _focusQuestPoI,
                  onFocusNode: _focusQuestNode,
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
                    polygonQuest!,
                    polygonNode!,
                  ),
                  child: Text('Quest: ${polygonQuest!.name}'),
                ),
              ),
          ],
          if (_questSubmissionPhase != QuestSubmissionOverlayPhase.hidden)
            Positioned.fill(
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: _questSubmissionPhase == QuestSubmissionOverlayPhase.loading ||
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
                          final isLoading = _questSubmissionPhase ==
                              QuestSubmissionOverlayPhase.loading;
                          final isFailure = _questSubmissionPhase ==
                              QuestSubmissionOverlayPhase.failure;
                          final accentColor = isFailure
                              ? Theme.of(context).colorScheme.error
                              : const Color(0xFFF5C542);
                          final statsProvider =
                              context.watch<CharacterStatsProvider>();
                          final statValues = _questSubmissionStatValues.isNotEmpty
                              ? _questSubmissionStatValues
                              : statsProvider.stats;
                          final statTags = _questSubmissionStatTags;
                          final hasDetails = _questSubmissionScore != null ||
                              _questSubmissionDifficulty != null ||
                              _questSubmissionCombinedScore != null ||
                              statTags.isNotEmpty ||
                              _questSubmissionStatValues.isNotEmpty;
                          final scoreValue = _questSubmissionScore ?? 0;
                          final combinedValue = _questSubmissionCombinedScore ??
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
                          final maxWidth =
                              availableWidth > 420 ? 420.0 : availableWidth;
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
                                  color: Theme.of(context)
                                      .colorScheme
                                      .surface
                                      .withOpacity(0.98),
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
                                  crossAxisAlignment: CrossAxisAlignment.stretch,
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
                                          style: Theme.of(context)
                                              .textTheme
                                              .bodySmall,
                                        ),
                                      ),
                                      const SizedBox(height: 10),
                                      LinearProgressIndicator(
                                        minHeight: 6,
                                        color: accentColor,
                                        backgroundColor:
                                            accentColor.withOpacity(0.15),
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
                                                      fontWeight: FontWeight.w700,
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
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .labelLarge,
                                              ),
                                              const SizedBox(height: 6),
                                              if (statTags.isEmpty)
                                                Text(
                                                  'None',
                                                  style: Theme.of(context)
                                                      .textTheme
                                                      .bodySmall,
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
                                                    padding: const EdgeInsets
                                                        .symmetric(
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
                                                      style: Theme.of(context)
                                                          .textTheme
                                                          .labelSmall,
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
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .labelLarge,
                                              ),
                                              const SizedBox(height: 4),
                                              Text(
                                                '$difficultyValue',
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .bodySmall,
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
                                              style: Theme.of(context)
                                                  .textTheme
                                                  .labelLarge,
                                            ),
                                            const SizedBox(height: 6),
                                            if (statTags.isNotEmpty)
                                              Text(
                                                'Modifiers: ${statTags.map((tag) {
                                                  final label =
                                                      _formatStatLabel(tag);
                                                  final value =
                                                      statValues[tag] ??
                                                          CharacterStatsProvider
                                                              .baseStatValue;
                                                  return '+$value $label';
                                                }).join(' · ')}',
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .bodySmall,
                                              ),
                                            Text(
                                              () {
                                                final parts = <String>['Score $scoreValue'];
                                                for (final tag in statTags) {
                                                  final label = _formatStatLabel(tag);
                                                  final value = statValues[tag] ??
                                                      CharacterStatsProvider.baseStatValue;
                                                  parts.add('+$value $label');
                                                }
                                                return '${parts.join(' ')} = $combinedValue';
                                              }(),
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .bodySmall,
                                              ),
                                              Text(
                                                'Difficulty = $difficultyValue',
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .bodySmall,
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
                                              style: Theme.of(context)
                                                  .textTheme
                                                  .bodySmall,
                                            ),
                                            const SizedBox(height: 12),
                                            Text(
                                              'Tap anywhere to dismiss.',
                                              style: Theme.of(context)
                                                  .textTheme
                                                  .bodySmall,
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
                                                    fontWeight:
                                                        FontWeight.w700,
                                                  ),
                                            ),
                                          ),
                                        ],
                                      ),
                                      const SizedBox(height: 12),
                                      Text(
                                        'Tap anywhere to dismiss.',
                                        textAlign: TextAlign.center,
                                        style: Theme.of(context)
                                            .textTheme
                                            .bodySmall,
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
          const CelebrationModalManager(),
          const NewItemModal(),
          const UsedItemModal(),
          // Shop is opened via showDialog from the character panel.
          // Dialogue is opened via showDialog to avoid overlay rendering issues.
        ],
      ),
    );
  }

  Future<void> _showCharacterPanel(Character ch) async {
    var openTrackedQuests = false;
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
        onClose: () => Navigator.of(context).pop(),
        onQuestAccepted: () => openTrackedQuests = true,
        onStartDialogue: (dialogContext, character, action) {
          debugPrint('SinglePlayer: onStartDialogue character=${character.id} action=${action.id}');
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
      debugPrint('SinglePlayer: showShopModal skipped (dialogContext unmounted)');
      return;
    }
    debugPrint('SinglePlayer: showShopModal open character=${character.id} action=${action.id}');
    await showDialog<void>(
      context: dialogContext,
      useRootNavigator: true,
      barrierDismissible: true,
      builder: (context) {
        debugPrint('SinglePlayer: showShopModal builder');
        return ShopModal(
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

  Future<void> _showDialogueModal(
    BuildContext dialogContext,
    Character character,
    CharacterAction action,
  ) async {
    if (!dialogContext.mounted) {
      debugPrint('SinglePlayer: showDialogueModal skipped (dialogContext unmounted)');
      return;
    }
    debugPrint('SinglePlayer: showDialogueModal open character=${character.id} action=${action.id}');
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

  void _showTreasureChestPanel(TreasureChest tc) {
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

  void _showQuestLog(BuildContext context) {
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
                    Text(
                      'Log',
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
                          color: Theme.of(context)
                              .colorScheme
                              .onSurface
                              .withValues(alpha: 0.3),
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
