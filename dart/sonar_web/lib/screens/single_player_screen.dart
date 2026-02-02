import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/point_of_interest.dart';
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
  List<Symbol> _poiSymbols = [];
  List<Symbol> _chestSymbols = [];
  bool _styleLoaded = false;
  bool _markersAdded = false;
  bool _addedMarkersWithEmptyDiscoveries = false;
  bool _mapLoadFailed = false;
  int _mapKey = 0;
  bool _hasAnimatedToUserLocation = false;
  Character? _shopCharacter;
  CharacterAction? _shopAction;
  Character? _dialogueCharacter;
  CharacterAction? _dialogueAction;

  @override
  void initState() {
    super.initState();
    debugPrint('SinglePlayer: initState');
    _startMapLoadTimeout();
    _loadAll();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<ZoneProvider>().addListener(_onZoneChanged);
    });
  }

  @override
  void dispose() {
    _mapLoadTimeout?.cancel();
    try {
      context.read<ZoneProvider>().removeListener(_onZoneChanged);
    } catch (_) {}
    super.dispose();
  }

  void _onZoneChanged() {
    if (!mounted) return;
    unawaited(_loadTreasureChestsForSelectedZone());
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
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => PointOfInterestPanel(
        pointOfInterest: poi,
        hasDiscovered: hasDiscovered,
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
      if (_poiSymbols.isNotEmpty) {
        try {
          await c.removeSymbols(_poiSymbols);
        } catch (_) {}
        if (!mounted) return;
        _poiSymbols.clear();
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
        if (ch.lat == 0 && ch.lng == 0) continue;
        c.addCircle(
          CircleOptions(
            geometry: LatLng(ch.lat, ch.lng),
            circleRadius: 30,
            circleColor: '#ff8833',
            circleStrokeWidth: 2,
            circleStrokeColor: '#ffffff',
          ),
          {'type': 'character', 'id': ch.id, 'name': ch.name},
        );
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
                iconHaloColor: '#000000',
                iconHaloWidth: 0.75,
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
              circleStrokeColor: '#ffffff',
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
    final lat = loc?.latitude ?? 0.0;
    final lng = loc?.longitude ?? 0.0;
    final initialPosition = CameraPosition(
      target: LatLng(lat, lng),
      zoom: 15,
    );

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
            // Top-right: icon buttons in a compact row
            Positioned(
              top: 16,
              right: 16,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  _OverlayButton(
                    icon: Icons.notifications,
                    onTap: () => _showActivityFeed(context),
                  ),
                  const SizedBox(width: 6),
                  _OverlayButton(
                    icon: Icons.chat,
                    onTap: () => _showLog(context),
                  ),
                  const SizedBox(width: 6),
                  _OverlayButton(
                    icon: Icons.label,
                    onTap: () => _showTagFilter(context),
                  ),
                ],
              ),
            ),
            // Zone selector: centered at top (ZoneWidget uses top: 80 internally)
            const ZoneWidget(),
            // Tracked quests: below zone to avoid overlap
            Positioned(
              top: 142,
              right: 16,
              child: TrackedQuestsOverlay(onFocusPoI: _focusQuestPoI),
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
          ],
          const CelebrationModalManager(),
          const NewItemModal(),
          const UsedItemModal(),
          if (_shopCharacter != null && _shopAction != null)
            ShopModal(
              character: _shopCharacter!,
              action: _shopAction!,
              onClose: () => setState(() {
                _shopCharacter = null;
                _shopAction = null;
              }),
            ),
          if (_dialogueCharacter != null && _dialogueAction != null)
            RpgDialogueModal(
              character: _dialogueCharacter!,
              action: _dialogueAction!,
              onClose: () => setState(() {
                _dialogueCharacter = null;
                _dialogueAction = null;
              }),
            ),
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
        onStartDialogue: (character, action) {
          Navigator.of(context).pop();
          setState(() {
            _dialogueCharacter = character;
            _dialogueAction = action;
          });
        },
        onStartShop: (character, action) {
          Navigator.of(context).pop();
          setState(() {
            _shopCharacter = character;
            _shopAction = action;
          });
        },
      ),
    );
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
