import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
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
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/log_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/tags_provider.dart';
import '../providers/zone_provider.dart';
import '../services/poi_service.dart';
import '../utils/poi_image_util.dart';
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
import '../widgets/tag_filter_chips.dart';
import '../widgets/treasure_chest_panel.dart';
import '../widgets/used_item_modal.dart';
import '../widgets/zone_widget.dart';

const _chestImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';

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
  List<Line> _questLines = [];
  List<Symbol> _poiSymbols = [];
  List<Symbol> _characterSymbols = [];
  List<Symbol> _chestSymbols = [];
  bool _styleLoaded = false;
  bool _markersAdded = false;
  bool _addedMarkersWithEmptyDiscoveries = false;
  bool _mapLoadFailed = false;
  int _mapKey = 0;
  bool _hasAnimatedToUserLocation = false;
  QuestLogProvider? _questLogProvider;

  @override
  void initState() {
    super.initState();
    debugPrint('SinglePlayer: initState');
    _startMapLoadTimeout();
    _loadAll();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<ZoneProvider>().addListener(_onZoneChanged);
      _questLogProvider = context.read<QuestLogProvider>();
      _questLogProvider?.addListener(_onQuestLogChanged);
      context.read<ActivityFeedProvider>().refresh();
    });
  }

  @override
  void dispose() {
    _mapLoadTimeout?.cancel();
    try {
      context.read<ZoneProvider>().removeListener(_onZoneChanged);
    } catch (_) {}
    try {
      _questLogProvider?.removeListener(_onQuestLogChanged);
    } catch (_) {}
    super.dispose();
  }

  void _onZoneChanged() {
    if (!mounted) return;
    unawaited(_loadTreasureChestsForSelectedZone());
  }

  void _onQuestLogChanged() {
    if (!mounted) return;
    if (_styleLoaded && _mapController != null) {
      setState(() => _markersAdded = false);
      unawaited(_addPoiMarkers());
      unawaited(_addQuestPolygons());
    }
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
        _markersAdded = false;
      });
      await _addPoiMarkers();
    } catch (e) {
      debugPrint('SinglePlayer: _loadTreasureChests error: $e');
      if (mounted) setState(() => _treasureChests = []);
    }
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
      builder: (context) => PointOfInterestPanel(
        pointOfInterest: poi,
        hasDiscovered: hasDiscovered,
        quest: questForPoi,
        questNode: nodeForPoi,
        onClose: () => Navigator.of(context).pop(),
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
      final questPoiIds = context.read<QuestLogProvider>().currentNodePoiIds.toSet();
      if (_poiSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_poiSymbols);
        } catch (_) {}
        if (!mounted) return;
        _poiSymbols.clear();
      }
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
      try {
        await c.clearCircles();
      } catch (_) {}

      Uint8List? placeholderBytes;
      try {
        placeholderBytes = await loadPoiThumbnail(null);
        if (placeholderBytes != null) await c.addImage('poi_placeholder', placeholderBytes);
      } catch (_) {}

      Uint8List? chestBytes;
      try {
        chestBytes = await loadPoiThumbnail(_chestImageUrl);
        if (chestBytes != null) await c.addImage('chest_thumbnail', chestBytes);
      } catch (_) {}

      for (final ch in _characters) {
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
        if (thumbnailUrl != null && thumbnailUrl.isNotEmpty) {
          try {
            final imageBytes = await loadPoiThumbnail(thumbnailUrl);
            if (imageBytes != null) {
              final imageId = 'character_${ch.id}';
              try {
                await c.addImage(imageId, imageBytes);
              } catch (_) {}
              for (final point in points) {
                final sym = await c.addSymbol(
                  SymbolOptions(
                    geometry: point,
                    iconImage: imageId,
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
      for (final tc in _treasureChests) {
        if (chestBytes != null) {
          final sym = await c.addSymbol(
            SymbolOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              iconImage: 'chest_thumbnail',
              iconSize: 0.75,
              iconHaloColor: '#000000',
              iconHaloWidth: 0.75,
              iconAnchor: 'center',
            ),
            {'type': 'chest', 'id': tc.id},
          );
          if (!mounted) return;
          _chestSymbols.add(sym);
        } else {
          c.addCircle(
            CircleOptions(
              geometry: LatLng(tc.latitude, tc.longitude),
              circleRadius: 24,
              circleColor: tc.openedByUser == true ? '#888888' : '#ffcc00',
              circleStrokeWidth: 2,
              circleStrokeColor: '#ffffff',
            ),
            {'type': 'chest', 'id': tc.id},
          );
        }
      }

      final discoveries = context.read<DiscoveriesProvider>();
      final hadEmptyDiscoveries = discoveries.discoveries.isEmpty;
      for (final poi in _pois) {
        final lat = double.tryParse(poi.lat) ?? 0.0;
        final lng = double.tryParse(poi.lng) ?? 0.0;
        final useRealImage = discoveries.hasDiscovered(poi.id);
        final isQuestCurrent = questPoiIds.contains(poi.id);
        var added = false;
        try {
          String? imageId;
          Uint8List? imageBytes;
          if (useRealImage) {
            imageBytes = await loadPoiThumbnail(poi.imageURL);
            imageId = imageBytes != null ? 'poi_${poi.id}' : (placeholderBytes != null ? 'poi_placeholder' : null);
          } else {
            imageId = placeholderBytes != null ? 'poi_placeholder' : null;
          }
          if (imageId != null) {
            if (imageBytes != null) await c.addImage(imageId, imageBytes);
            final sym = await c.addSymbol(
              SymbolOptions(
                geometry: LatLng(lat, lng),
                iconImage: imageId,
                iconSize: 0.75,
                iconHaloColor: isQuestCurrent ? '#f5c542' : '#000000',
                iconHaloWidth: isQuestCurrent ? 2.0 : 0.75,
                iconAnchor: 'center',
              ),
              {'type': 'poi', 'id': poi.id, 'name': poi.name},
            );
            if (!mounted) return;
            _poiSymbols.add(sym);
            added = true;
          }
        } catch (_) {}
        if (!added) {
          c.addCircle(
            CircleOptions(
              geometry: LatLng(lat, lng),
              circleRadius: 24,
              circleColor: '#3388ff',
              circleStrokeWidth: 2,
              circleStrokeColor: isQuestCurrent ? '#f5c542' : '#ffffff',
            ),
            {'type': 'poi', 'id': poi.id, 'name': poi.name},
          );
        }
      }
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

  Future<void> _addZoneBoundaries() async {
    final c = _mapController;
    if (c == null || !_styleLoaded) {
      debugPrint('SinglePlayer: _addZoneBoundaries skip (controller=${c != null} styleLoaded=$_styleLoaded)');
      return;
    }
    if (_zoneLines.isNotEmpty) {
      try {
        await c.removeLines(_zoneLines);
      } catch (_) {}
      if (!mounted) return;
      _zoneLines.clear();
    }
    final options = <LineOptions>[];
    for (final z in _zones) {
      final ring = _zoneRing(z);
      if (ring.length < 2) continue;
      options.add(LineOptions(
        geometry: ring,
        lineColor: '#4A90E2',
        lineWidth: 3.0,
        lineOpacity: 1.0,
      ));
    }
    debugPrint('SinglePlayer: _addZoneBoundaries zones=${_zones.length} rings=${options.length}');
    if (options.isEmpty) return;
    try {
      final lines = await c.addLines(options);
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
    if (_questLines.isNotEmpty) {
      try {
        await c.removeLines(_questLines);
      } catch (_) {}
      if (!mounted) return;
      _questLines.clear();
    }

    final questLog = context.read<QuestLogProvider>();
    final polygons = questLog.currentNodePolygons;
    if (polygons.isEmpty) return;

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
        lineWidth: 4.0,
        lineOpacity: 1.0,
      ));
    }

    if (options.isEmpty) return;
    try {
      final lines = await c.addLines(options);
      if (!mounted) return;
      _questLines.addAll(lines);
    } catch (e, st) {
      debugPrint('SinglePlayer: _addQuestPolygons error: $e');
      debugPrint('SinglePlayer: _addQuestPolygons stack: $st');
    }
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
    final imageController = TextEditingController();
    String? selectedChallengeId = node.challenges.isNotEmpty
        ? node.challenges.first.id
        : null;

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
                  TextField(
                    controller: imageController,
                    decoration: const InputDecoration(
                      labelText: 'Image URL (optional)',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: () async {
                      final resp = await context
                          .read<QuestLogProvider>()
                          .submitQuestNodeChallenge(
                            node.id,
                            questNodeChallengeId: selectedChallengeId,
                            textSubmission: textController.text.trim(),
                            imageSubmissionUrl:
                                imageController.text.trim().isEmpty
                                    ? null
                                    : imageController.text.trim(),
                          );
                      if (!mounted) return;
                      final success = resp['successful'] == true;
                      final reason = resp['reason']?.toString() ?? '';
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(
                            success
                                ? (reason.isNotEmpty
                                    ? reason
                                    : 'Challenge completed!')
                                : (reason.isNotEmpty
                                    ? reason
                                    : 'Submission failed'),
                          ),
                        ),
                      );
                      if (success) {
                        Navigator.of(context).pop();
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
    final zone = zoneProvider.findZoneAtCoordinate(location.latitude, location.longitude);
    if (zone != null && zoneProvider.selectedZone?.id != zone.id) {
      zoneProvider.setSelectedZone(zone);
    }
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

    // Update selected zone when location changes
    if (loc != null && _zones.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        _updateSelectedZoneFromLocation();
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
          MapLibreMap(
            key: ValueKey(_mapKey),
            initialCameraPosition: initialPosition,
            styleString: MapLibreStyles.openfreemapLiberty,
            onMapCreated: (c) {
              debugPrint('SinglePlayer: map created');
              _mapController = c;
              _setupTapHandlers(c);
            },
            onStyleLoadedCallback: _onMapStyleLoaded,
            myLocationEnabled: true,
            compassEnabled: true,
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
            // Top-left: ship icon opens notifications
            Positioned(
              top: 16,
              left: 16,
              child: Consumer<ActivityFeedProvider>(
                builder: (context, feed, _) {
                  final hasUnseen = feed.unseenActivities.isNotEmpty;
                  return GestureDetector(
                    onTap: () => _showActivityFeed(context),
                    child: Stack(
                      clipBehavior: Clip.none,
                      children: [
                        Container(
                          width: 44,
                          height: 44,
                          decoration: BoxDecoration(
                            color: Colors.white,
                            shape: BoxShape.circle,
                            border: Border.all(color: Colors.black87, width: 2),
                            boxShadow: const [
                              BoxShadow(
                                color: Colors.black26,
                                blurRadius: 6,
                                offset: Offset(0, 2),
                              ),
                            ],
                          ),
                          child: const Icon(Icons.directions_boat_filled, size: 22),
                        ),
                        if (hasUnseen)
                          Positioned(
                            top: -1,
                            right: -1,
                            child: Container(
                              width: 10,
                              height: 10,
                              decoration: BoxDecoration(
                                color: Colors.red.shade600,
                                shape: BoxShape.circle,
                                border: Border.all(color: Colors.white, width: 1),
                              ),
                            ),
                          ),
                      ],
                    ),
                  );
                },
              ),
            ),
            // Top-right: tag filter button
            Positioned(
              top: 16,
              right: 16,
              child: _OverlayButton(
                icon: Icons.label,
                onTap: () => _showTagFilter(context),
              ),
            ),
            // Zone selector: centered at top, aligned with top controls
            const ZoneWidget(top: 16),
            // Tracked quests: below zone to avoid overlap
            Positioned(
              top: 142,
              right: 16,
              child: TrackedQuestsOverlay(
                onFocusPoI: _focusQuestPoI,
                onFocusNode: _focusQuestNode,
              ),
            ),
            // Bottom: primary actions
            Positioned(
              left: 16,
              right: 16,
              bottom: 24,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  _MapButton(
                    label: 'Inventory',
                    onTap: () => _showInventory(context),
                  ),
                  _MapButton(
                    label: 'Quest Log',
                    onTap: () => _showQuestLog(context),
                  ),
                ],
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
                  child: const Text('Submit Quest Challenge'),
                ),
              ),
          ],
          const CelebrationModalManager(),
          const NewItemModal(),
          const UsedItemModal(),
          // Shop is opened via showDialog from the character panel.
          // Dialogue is opened via showDialog to avoid overlay rendering issues.
        ],
      ),
    );
  }

  void _showCharacterPanel(Character ch) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => CharacterPanel(
        character: ch,
        onClose: () => Navigator.of(context).pop(),
        onStartDialogue: (dialogContext, character, action) {
          debugPrint('SinglePlayer: onStartDialogue character=${character.id} action=${action.id}');
          _showDialogueModal(dialogContext, character, action);
        },
        onStartShop: (dialogContext, character, action) {
          _showShopModal(dialogContext, character, action);
        },
      ),
    );
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
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              ),
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
    );
  }

  void _showInventory(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              ),
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
    );
  }

  void _showQuestLog(BuildContext context) {
    context.read<QuestLogProvider>().refresh();
    context.read<TagsProvider>().refresh();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              ),
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
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showLog(BuildContext context) {
    context.read<LogProvider>().refresh();
    showModalBottomSheet(
      context: context,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.3,
        maxChildSize: 0.95,
        builder: (_, scrollController) => Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              ),
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
    );
  }

  void _showTagFilter(BuildContext context) {
    context.read<TagsProvider>().refresh();
    showModalBottomSheet(
      context: context,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.5,
        minChildSize: 0.2,
        maxChildSize: 0.7,
        builder: (_, scrollController) => Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Tag Filter',
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
                  child: TagFilterChips(),
                ),
              ),
            ),
          ],
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
    return Material(
      color: Colors.white.withValues(alpha: 0.9),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          child: Text(label, style: const TextStyle(fontWeight: FontWeight.w600)),
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
    return Material(
      color: Colors.white.withValues(alpha: 0.9),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Icon(icon, size: 24),
        ),
      ),
    );
  }
}
