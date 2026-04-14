import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../models/user.dart';
import '../models/zone.dart';
import '../providers/auth_provider.dart';
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/party_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/tags_provider.dart';
import '../providers/zone_provider.dart';
import '../services/poi_service.dart';
import 'quest_objective_display.dart';

/// Bottom-sheet content for Quest Log.
/// [onFocusPoI] when user taps a POI in a quest: close sheet, fly to POI, open POI panel.
typedef OnFocusPoI = void Function(PointOfInterest poi);
typedef OnFocusNode = void Function(QuestNode node);
typedef OnFocusTurnInQuest = void Function(Quest quest);

class QuestLogPanel extends StatefulWidget {
  const QuestLogPanel({
    super.key,
    required this.onClose,
    required this.onFocusPoI,
    required this.onFocusNode,
    required this.onFocusTurnInQuest,
    this.initialSelectedQuest,
    this.featuredMainStoryPoi,
    this.featuredMainStoryQuestGiverName,
    this.onFocusFeaturedMainStoryLead,
  });

  final VoidCallback onClose;
  final OnFocusPoI onFocusPoI;
  final OnFocusNode onFocusNode;
  final OnFocusTurnInQuest onFocusTurnInQuest;
  final Quest? initialSelectedQuest;
  final PointOfInterest? featuredMainStoryPoi;
  final String? featuredMainStoryQuestGiverName;
  final void Function(PointOfInterest poi)? onFocusFeaturedMainStoryLead;

  @override
  State<QuestLogPanel> createState() => _QuestLogPanelState();
}

class _QuestLogPanelState extends State<QuestLogPanel> {
  Quest? _selectedQuest;
  final Map<String, bool> _expanded = {};
  String? _sharingQuestId;
  List<PointOfInterest> _leadCandidatePois = const [];
  List<Character> _leadCandidateCharacters = const [];
  bool _loadingLeadCandidates = false;
  bool _hasLoadedLeadCandidates = false;

  @override
  void initState() {
    super.initState();
    _selectedQuest = widget.initialSelectedQuest;
    _ensureFeaturedMainStoryLeadLoaded();
  }

  @override
  void didUpdateWidget(covariant QuestLogPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.initialSelectedQuest?.id != oldWidget.initialSelectedQuest?.id &&
        widget.initialSelectedQuest != null) {
      _selectedQuest = widget.initialSelectedQuest;
    }
    if (oldWidget.featuredMainStoryPoi?.id != widget.featuredMainStoryPoi?.id &&
        widget.featuredMainStoryPoi == null) {
      _ensureFeaturedMainStoryLeadLoaded(force: true);
    }
  }

  void _focusPoI(PointOfInterest poi) {
    widget.onClose();
    widget.onFocusPoI(poi);
  }

  void _focusTurnInQuest(Quest quest) {
    widget.onClose();
    widget.onFocusTurnInQuest(quest);
  }

  void _focusNode(QuestNode node) {
    widget.onClose();
    widget.onFocusNode(node);
  }

  void _focusFeaturedMainStoryLead() {
    final lead = _resolveFeaturedMainStoryLead();
    if (lead == null) return;
    widget.onClose();
    final onFocusLead = widget.onFocusFeaturedMainStoryLead;
    if (onFocusLead != null) {
      onFocusLead(lead.poi);
      return;
    }
    widget.onFocusPoI(lead.poi);
  }

  void _ensureFeaturedMainStoryLeadLoaded({bool force = false}) {
    if (widget.featuredMainStoryPoi != null) return;
    if (_loadingLeadCandidates) return;
    if (_hasLoadedLeadCandidates && !force) return;

    _loadingLeadCandidates = true;
    _hasLoadedLeadCandidates = true;
    final poiService = context.read<PoiService>();
    Future<void>(() async {
      try {
        final results = await Future.wait([
          poiService.getPointsOfInterest(),
          poiService.getCharacters(),
        ]);
        if (!mounted) return;
        setState(() {
          _leadCandidatePois = results[0] as List<PointOfInterest>;
          _leadCandidateCharacters = results[1] as List<Character>;
        });
      } catch (_) {
        if (!mounted) return;
      } finally {
        _loadingLeadCandidates = false;
      }
    });
  }

  _ResolvedMainStoryLead? _resolveFeaturedMainStoryLead() {
    final providedPoi = widget.featuredMainStoryPoi;
    if (providedPoi != null) {
      String? questGiverName = widget.featuredMainStoryQuestGiverName;
      if (questGiverName == null || questGiverName.trim().isEmpty) {
        for (final character in providedPoi.characters) {
          if (character.hasAvailableMainStoryQuest) {
            questGiverName = character.name.trim();
            break;
          }
        }
      }
      return _ResolvedMainStoryLead(
        poi: providedPoi,
        questGiverName: questGiverName,
      );
    }
    if (_leadCandidatePois.isEmpty && _leadCandidateCharacters.isEmpty) {
      return null;
    }

    final selectedZone = context.read<ZoneProvider>().selectedZone;
    final location = context.read<LocationProvider>().location;
    final userLat = location?.latitude;
    final userLng = location?.longitude;

    _ResolvedMainStoryLead? bestLead;
    double? bestDistance;
    _ResolvedMainStoryLead? bestGlobalLead;
    double? bestGlobalDistance;

    for (final poi in _leadCandidatePois) {
      final poiLat = double.tryParse(poi.lat);
      final poiLng = double.tryParse(poi.lng);
      if (poiLat == null || poiLng == null) continue;

      String? questGiverName;
      for (final character in poi.characters) {
        if (character.hasAvailableMainStoryQuest) {
          questGiverName = character.name.trim();
          break;
        }
      }
      final hasAvailableLead =
          poi.hasAvailableMainStoryQuest || questGiverName != null;
      if (!hasAvailableLead) continue;

      final distance = (userLat == null || userLng == null)
          ? 0.0
          : _haversineDistanceMeters(userLat, userLng, poiLat, poiLng);
      final lead = _ResolvedMainStoryLead(
        poi: poi,
        questGiverName: questGiverName,
      );
      if (bestGlobalLead == null ||
          bestGlobalDistance == null ||
          distance < bestGlobalDistance) {
        bestGlobalLead = lead;
        bestGlobalDistance = distance;
      }
      if (selectedZone == null ||
          !_isPointInZone(selectedZone, poiLat, poiLng)) {
        continue;
      }
      if (bestLead == null || bestDistance == null || distance < bestDistance) {
        bestLead = lead;
        bestDistance = distance;
      }
    }

    for (final character in _leadCandidateCharacters) {
      if (!character.hasAvailableMainStoryQuest) continue;
      final poi = _syntheticPoiForCharacterLead(character);
      if (poi == null) continue;
      final poiLat = double.tryParse(poi.lat);
      final poiLng = double.tryParse(poi.lng);
      if (poiLat == null || poiLng == null) continue;
      final distance = (userLat == null || userLng == null)
          ? 0.0
          : _haversineDistanceMeters(userLat, userLng, poiLat, poiLng);
      final lead = _ResolvedMainStoryLead(
        poi: poi,
        questGiverName: character.name.trim(),
      );
      if (bestGlobalLead == null ||
          bestGlobalDistance == null ||
          distance < bestGlobalDistance) {
        bestGlobalLead = lead;
        bestGlobalDistance = distance;
      }
      if (selectedZone == null ||
          !_isPointInZone(selectedZone, poiLat, poiLng)) {
        continue;
      }
      if (bestLead == null || bestDistance == null || distance < bestDistance) {
        bestLead = lead;
        bestDistance = distance;
      }
    }

    return bestLead ?? bestGlobalLead;
  }

  PointOfInterest? _poiForCharacter(Character character) {
    final poiId = character.pointOfInterestId?.trim() ?? '';
    if (poiId.isNotEmpty) {
      for (final poi in _leadCandidatePois) {
        if (poi.id == poiId) return poi;
      }
    }
    for (final poi in _leadCandidatePois) {
      if (poi.characters.any((candidate) => candidate.id == character.id)) {
        return poi;
      }
    }
    final poiLat = character.pointOfInterestLat;
    final poiLng = character.pointOfInterestLng;
    if (poiLat != null && poiLng != null) {
      for (final poi in _leadCandidatePois) {
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
        character.pointOfInterestLat!.isFinite &&
        character.pointOfInterestLng!.isFinite &&
        character.pointOfInterestLat!.abs() <= 90 &&
        character.pointOfInterestLng!.abs() <= 180 &&
        (character.pointOfInterestLat != 0 ||
            character.pointOfInterestLng != 0)) {
      lat = character.pointOfInterestLat;
      lng = character.pointOfInterestLng;
    } else {
      for (final location in character.locations) {
        if (!location.latitude.isFinite || !location.longitude.isFinite) {
          continue;
        }
        if (location.latitude.abs() > 90 || location.longitude.abs() > 180) {
          continue;
        }
        lat = location.latitude;
        lng = location.longitude;
        break;
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

  String _displayName(User user) {
    if (user.username.trim().isNotEmpty) return '@${user.username.trim()}';
    if (user.name.trim().isNotEmpty) return user.name.trim();
    if (user.phoneNumber.trim().isNotEmpty) return user.phoneNumber.trim();
    return user.id;
  }

  Future<String?> _selectPartyMemberForQuestShare(
    BuildContext context, {
    required Quest quest,
    required List<User> partyMembers,
  }) async {
    if (partyMembers.isEmpty) return null;
    return showDialog<String>(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) {
        final theme = Theme.of(dialogContext);
        return Dialog(
          insetPadding: const EdgeInsets.symmetric(
            horizontal: 24,
            vertical: 24,
          ),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 360),
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    'Share Quest',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    quest.name,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Flexible(
                    child: ListView.separated(
                      shrinkWrap: true,
                      itemCount: partyMembers.length,
                      separatorBuilder: (_, _) => const SizedBox(height: 8),
                      itemBuilder: (_, index) {
                        final member = partyMembers[index];
                        return FilledButton.tonal(
                          onPressed: () =>
                              Navigator.of(dialogContext).pop(member.id),
                          style: FilledButton.styleFrom(
                            alignment: Alignment.centerLeft,
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 12,
                            ),
                          ),
                          child: Text(_displayName(member)),
                        );
                      },
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Future<void> _shareQuestWithPartyMember(
    BuildContext context,
    Quest quest,
    List<User> partyMembers,
  ) async {
    if (_sharingQuestId != null) return;
    final targetUserId = await _selectPartyMemberForQuestShare(
      context,
      quest: quest,
      partyMembers: partyMembers,
    );
    if (!mounted || targetUserId == null || targetUserId.isEmpty) return;
    final questLog = this.context.read<QuestLogProvider>();
    final messenger = ScaffoldMessenger.of(this.context);

    setState(() {
      _sharingQuestId = quest.id;
    });
    final error = await questLog.shareQuest(quest.id, targetUserId);
    if (!mounted) return;
    setState(() {
      _sharingQuestId = null;
    });
    messenger.showSnackBar(
      SnackBar(content: Text(error ?? 'Quest shared successfully.')),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_selectedQuest != null) {
      return _buildQuestDetail(context, _selectedQuest!);
    }
    return _buildQuestList(context);
  }

  Widget _buildQuestList(BuildContext context) {
    return Consumer3<QuestLogProvider, TagsProvider, DiscoveriesProvider>(
      builder: (context, ql, tags, discoveries, _) {
        final featuredMainStoryLead = _resolveFeaturedMainStoryLead();
        if (ql.loading && ql.quests.isEmpty) {
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(24),
              child: CircularProgressIndicator(),
            ),
          );
        }
        final readyToTurnIn = ql.quests
            .where(
              (q) =>
                  !q.isMainStory &&
                  q.turnedInAt == null &&
                  (q.readyToTurnIn || (q.currentNode == null && q.isAccepted)),
            )
            .toList();
        final mainStoryActive = ql.quests
            .where((q) => q.turnedInAt == null && q.isMainStory)
            .toList();
        final mainStoryIds = mainStoryActive.map((q) => q.id).toSet();
        final readyIds = readyToTurnIn.map((q) => q.id).toSet();
        final tracked = ql.quests
            .where(
              (q) =>
                  !q.isMainStory &&
                  ql.trackedQuestIds.contains(q.id) &&
                  !readyIds.contains(q.id),
            )
            .toList();
        final tagBuckets = <String, List<Quest>>{};
        for (final g in tags.tagGroups) {
          tagBuckets[g.id] = [];
        }
        final untagged = <Quest>[];

        for (final q in ql.quests) {
          if (mainStoryIds.contains(q.id)) continue;
          if (readyIds.contains(q.id)) continue;
          if (ql.trackedQuestIds.contains(q.id)) continue;
          final tagNames = _questTags(q);
          var added = false;
          for (final g in tags.tagGroups) {
            final hasMatch = tagNames.any(
              (tName) =>
                  tags.tags.any((x) => x.tagGroupId == g.id && x.name == tName),
            );
            if (hasMatch) {
              tagBuckets[g.id]!.add(q);
              added = true;
            }
          }
          if (!added) untagged.add(q);
        }

        final completedMainStory = ql.completedQuests
            .where((q) => q.isMainStory)
            .toList();
        final completed = ql.completedQuests
            .where((q) => !q.isMainStory)
            .toList();
        final discoveredIds = <String>{
          for (final d in discoveries.discoveries) d.pointOfInterestId,
        };

        final hasQuestListItems =
            featuredMainStoryLead != null ||
            mainStoryActive.isNotEmpty ||
            readyToTurnIn.isNotEmpty ||
            tracked.isNotEmpty ||
            tagBuckets.values.any((list) => list.isNotEmpty) ||
            untagged.isNotEmpty;

        return DefaultTabController(
          length: 2,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const TabBar(
                tabs: [
                  Tab(text: 'Quests'),
                  Tab(text: 'Completed'),
                ],
              ),
              Expanded(
                child: TabBarView(
                  children: [
                    SingleChildScrollView(
                      padding: const EdgeInsets.all(16),
                      child: hasQuestListItems
                          ? Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                if (featuredMainStoryLead != null)
                                  _ImportantQuestAvailableCard(
                                    poi: featuredMainStoryLead.poi,
                                    questGiverName:
                                        featuredMainStoryLead.questGiverName,
                                    onTap: _focusFeaturedMainStoryLead,
                                  ),
                                if (mainStoryActive.isNotEmpty)
                                  _QuestAccordion(
                                    title: 'Main Story',
                                    quests: mainStoryActive,
                                    discoveredPoiIds: discoveredIds,
                                    expanded: _expanded['main_story'] ?? true,
                                    onToggle: () {
                                      setState(() {
                                        _expanded['main_story'] =
                                            !(_expanded['main_story'] ?? true);
                                      });
                                    },
                                    onQuestTap: (q) =>
                                        setState(() => _selectedQuest = q),
                                    onReadyQuestTap: _focusTurnInQuest,
                                  ),
                                if (readyToTurnIn.isNotEmpty)
                                  _QuestAccordion(
                                    title: 'Ready to Turn In',
                                    quests: readyToTurnIn,
                                    discoveredPoiIds: discoveredIds,
                                    expanded: _expanded['ready'] ?? true,
                                    onToggle: () {
                                      setState(() {
                                        _expanded['ready'] =
                                            !(_expanded['ready'] ?? true);
                                      });
                                    },
                                    onQuestTap: (q) =>
                                        setState(() => _selectedQuest = q),
                                    onReadyQuestTap: _focusTurnInQuest,
                                  ),
                                if (tracked.isNotEmpty)
                                  _QuestAccordion(
                                    title: 'Tracked Quests',
                                    quests: tracked,
                                    discoveredPoiIds: discoveredIds,
                                    expanded: _expanded['tracked'] ?? false,
                                    onToggle: () {
                                      setState(() {
                                        _expanded['tracked'] =
                                            !(_expanded['tracked'] ?? false);
                                      });
                                    },
                                    onQuestTap: (q) =>
                                        setState(() => _selectedQuest = q),
                                    onReadyQuestTap: _focusTurnInQuest,
                                  ),
                                ...tags.tagGroups.map((g) {
                                  final list = tagBuckets[g.id] ?? [];
                                  if (list.isEmpty) {
                                    return const SizedBox.shrink();
                                  }
                                  return _QuestAccordion(
                                    key: ValueKey(g.id),
                                    title: g.name,
                                    quests: list,
                                    discoveredPoiIds: discoveredIds,
                                    expanded: _expanded[g.id] ?? false,
                                    onToggle: () {
                                      setState(() {
                                        _expanded[g.id] =
                                            !(_expanded[g.id] ?? false);
                                      });
                                    },
                                    onQuestTap: (q) =>
                                        setState(() => _selectedQuest = q),
                                    onReadyQuestTap: _focusTurnInQuest,
                                  );
                                }),
                                if (untagged.isNotEmpty)
                                  _QuestAccordion(
                                    title: 'The Rest',
                                    quests: untagged,
                                    discoveredPoiIds: discoveredIds,
                                    expanded: _expanded['untagged'] ?? false,
                                    onToggle: () {
                                      setState(() {
                                        _expanded['untagged'] =
                                            !(_expanded['untagged'] ?? false);
                                      });
                                    },
                                    onQuestTap: (q) =>
                                        setState(() => _selectedQuest = q),
                                    onReadyQuestTap: _focusTurnInQuest,
                                  ),
                              ],
                            )
                          : const Padding(
                              padding: EdgeInsets.symmetric(vertical: 48),
                              child: Center(
                                child: Text(
                                  'No quests yet. Explore to discover new adventures.',
                                ),
                              ),
                            ),
                    ),
                    if (completed.isEmpty && completedMainStory.isEmpty)
                      const Center(
                        child: Padding(
                          padding: EdgeInsets.all(24),
                          child: Text('No completed quests yet.'),
                        ),
                      )
                    else
                      SingleChildScrollView(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            if (completedMainStory.isNotEmpty)
                              _QuestAccordion(
                                title: 'Main Story Archive',
                                quests: completedMainStory,
                                discoveredPoiIds: discoveredIds,
                                expanded:
                                    _expanded['completed_main_story'] ?? true,
                                onToggle: () {
                                  setState(() {
                                    _expanded['completed_main_story'] =
                                        !(_expanded['completed_main_story'] ??
                                            true);
                                  });
                                },
                                onQuestTap: (q) =>
                                    setState(() => _selectedQuest = q),
                              ),
                            if (completed.isNotEmpty)
                              _QuestAccordion(
                                title: 'Completed Quests',
                                quests: completed,
                                discoveredPoiIds: discoveredIds,
                                expanded: _expanded['completed'] ?? true,
                                onToggle: () {
                                  setState(() {
                                    _expanded['completed'] =
                                        !(_expanded['completed'] ?? true);
                                  });
                                },
                                onQuestTap: (q) =>
                                    setState(() => _selectedQuest = q),
                              ),
                          ],
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  List<String> _questTags(Quest quest) {
    final poi = quest.currentNode?.pointOfInterest;
    if (poi == null) return [];
    return poi.tags.map((t) => t.name).toList();
  }

  String _randomRewardLabel(String size) {
    switch (size) {
      case Quest.randomRewardSizeLarge:
        return 'Large random reward';
      case Quest.randomRewardSizeMedium:
        return 'Medium random reward';
      default:
        return 'Small random reward';
    }
  }

  String _materialRewardLabel(String resourceKey) {
    switch (resourceKey.trim()) {
      case 'timber':
        return 'Timber';
      case 'stone':
        return 'Stone';
      case 'iron':
        return 'Iron';
      case 'herbs':
        return 'Herbs';
      case 'monster_parts':
        return 'Monster Parts';
      case 'arcane_dust':
        return 'Arcane Dust';
      case 'relic_shards':
        return 'Relic Shards';
      default:
        return resourceKey.trim().isEmpty ? 'Material' : resourceKey.trim();
    }
  }

  Widget _buildRandomRewardNotice(BuildContext context, Quest quest) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFEEE6D3),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFFC7B28A)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.casino_outlined, color: Color(0xFF7A5B20), size: 20),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              _randomRewardLabel(quest.randomRewardSize),
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w700,
                color: const Color(0xFF5F4618),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuestDetail(BuildContext context, Quest quest) {
    return Consumer4<
      QuestLogProvider,
      DiscoveriesProvider,
      PartyProvider,
      AuthProvider
    >(
      builder: (context, ql, discoveries, partyProvider, authProvider, _) {
        final isTracked = ql.trackedQuestIds.contains(quest.id);
        final node = quest.currentNode;
        final poi = node?.pointOfInterest;
        final currentUserId = authProvider.user?.id ?? '';
        final seenPartyMemberIds = <String>{};
        final partyMembers = <User>[
          if (partyProvider.party?.leader.id.isNotEmpty == true &&
              partyProvider.party!.leader.id != currentUserId &&
              seenPartyMemberIds.add(partyProvider.party!.leader.id))
            partyProvider.party!.leader,
          ...(partyProvider.party?.members.where(
                (member) =>
                    member.id != currentUserId &&
                    member.id.isNotEmpty &&
                    seenPartyMemberIds.add(member.id),
              ) ??
              const <User>[]),
        ];
        final canShareQuest =
            quest.isAccepted &&
            !quest.isTutorial &&
            quest.turnedInAt == null &&
            partyMembers.isNotEmpty;
        final discoveredIds = <String>{
          for (final d in discoveries.discoveries) d.pointOfInterestId,
        };
        final objectiveLines = questObjectiveLines(node);
        final canFocusNode = questNodeHasDirectFocusTarget(node);

        return SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextButton.icon(
                onPressed: () => setState(() => _selectedQuest = null),
                icon: const Icon(Icons.arrow_back),
                label: const Text('Back to Quests'),
              ),
              const SizedBox(height: 8),
              Text(
                quest.name,
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              if (quest.isMainStory || quest.isTutorial) ...[
                const SizedBox(height: 8),
                Align(
                  alignment: Alignment.centerLeft,
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF7A1823),
                      borderRadius: BorderRadius.circular(999),
                      border: Border.all(color: const Color(0xFFE7C36A)),
                    ),
                    child: Text(
                      quest.isTutorial ? 'Tutorial' : 'Main Story',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                        color: Colors.white,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ),
              ],
              const SizedBox(height: 12),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    quest.turnedInAt != null
                        ? 'Completed'
                        : quest.readyToTurnIn
                        ? 'Ready to turn in'
                        : quest.isAccepted
                        ? 'In progress'
                        : 'Not accepted',
                    style: Theme.of(context).textTheme.titleSmall,
                  ),
                  if (quest.turnedInAt == null)
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        if (canShareQuest)
                          FilledButton.tonal(
                            onPressed: _sharingQuestId == quest.id
                                ? null
                                : () => _shareQuestWithPartyMember(
                                    context,
                                    quest,
                                    partyMembers,
                                  ),
                            child: Text(
                              _sharingQuestId == quest.id
                                  ? 'Sharing...'
                                  : 'Share Quest',
                            ),
                          ),
                        if (!quest.isTutorial)
                          FilledButton(
                            onPressed: () async {
                              if (isTracked) {
                                await ql.untrackQuest(quest.id);
                              } else {
                                await ql.trackQuest(quest.id);
                              }
                            },
                            child: Text(
                              isTracked ? 'Untrack Quest' : 'Track Quest',
                            ),
                          ),
                      ],
                    ),
                ],
              ),
              if (quest.completionCount > 1)
                Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: Text(
                    'Completed ${quest.completionCount} times',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Colors.grey.shade600,
                    ),
                  ),
                ),
              const SizedBox(height: 16),
              Text(
                quest.description,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 16),
              Text(
                'Rewards',
                style: Theme.of(
                  context,
                ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              if (quest.gold > 0)
                Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: Text(
                    '+${quest.gold} Gold',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
              if (quest.materialRewards.isNotEmpty)
                ...quest.materialRewards.map((reward) {
                  if (reward.amount <= 0) {
                    return const SizedBox.shrink();
                  }
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Text(
                      '+${reward.amount} ${_materialRewardLabel(reward.resourceKey)}',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  );
                }),
              if (quest.itemRewards.isNotEmpty)
                ...quest.itemRewards.map((reward) {
                  final itemName = reward.inventoryItem?.name ?? 'Item';
                  final qty = reward.quantity > 0 ? reward.quantity : 1;
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Text(
                      '+$qty $itemName',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  );
                }),
              if (quest.spellRewards.isNotEmpty)
                ...quest.spellRewards.map((reward) {
                  final spellName = reward.spell?.name.trim().isNotEmpty == true
                      ? reward.spell!.name.trim()
                      : 'Spell';
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Text(
                      '+Spell: $spellName',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  );
                }),
              if (quest.hasRandomRewards &&
                  quest.gold <= 0 &&
                  quest.materialRewards.isEmpty &&
                  quest.itemRewards.isEmpty &&
                  quest.spellRewards.isEmpty)
                _buildRandomRewardNotice(context, quest),
              if (quest.gold <= 0 &&
                  !quest.hasRandomRewards &&
                  quest.materialRewards.isEmpty &&
                  quest.itemRewards.isEmpty &&
                  quest.spellRewards.isEmpty)
                Text(
                  'No rewards listed.',
                  style: Theme.of(
                    context,
                  ).textTheme.bodySmall?.copyWith(color: Colors.grey.shade600),
                ),
              const SizedBox(height: 16),
              if (node == null)
                Text(
                  quest.turnedInAt != null
                      ? 'Quest turned in. Well done!'
                      : 'Quest completed! Turn it in for rewards.',
                  style: Theme.of(context).textTheme.titleMedium,
                )
              else ...[
                Text(
                  'Current Objective',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                if (poi != null)
                  _QuestPoiCard(
                    node: node,
                    poi: poi,
                    discoveredPoiIds: discoveredIds,
                    objectiveSummary: questObjectiveSummary(node),
                    onTap: () => _focusPoI(poi),
                  )
                else
                  InkWell(
                    onTap: canFocusNode ? () => _focusNode(node) : null,
                    borderRadius: BorderRadius.circular(8),
                    child: Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.amber.shade50,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.amber.shade200),
                      ),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          QuestObjectiveIcon(
                            node: node,
                            discoveredPoiIds: discoveredIds,
                            size: 40,
                            borderRadius: 6,
                            iconColor: Theme.of(context).colorScheme.onSurface,
                            backgroundColor: Colors.amber.shade100,
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: objectiveLines
                                  .map(
                                    (line) => Padding(
                                      padding: const EdgeInsets.only(bottom: 4),
                                      child: Text(line),
                                    ),
                                  )
                                  .toList(),
                            ),
                          ),
                          if (canFocusNode) ...[
                            const SizedBox(width: 8),
                            const Icon(Icons.chevron_right),
                          ],
                        ],
                      ),
                    ),
                  ),
              ],
            ],
          ),
        );
      },
    );
  }
}

class _ImportantQuestAvailableCard extends StatelessWidget {
  const _ImportantQuestAvailableCard({
    required this.poi,
    required this.onTap,
    this.questGiverName,
  });

  final PointOfInterest poi;
  final VoidCallback onTap;
  final String? questGiverName;

  @override
  Widget build(BuildContext context) {
    final hasQuestGiver = (questGiverName ?? '').trim().isNotEmpty;
    final headline = (questGiverName ?? '').trim().isNotEmpty
        ? '$questGiverName has something important to tell you.'
        : 'An important quest is waiting here.';
    final locationName = poi.name.trim().isNotEmpty
        ? poi.name.trim()
        : 'this place';

    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF7A1823), Color(0xFF4E0F17)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE7C36A), width: 1.5),
        boxShadow: const [
          BoxShadow(
            color: Color(0x33000000),
            blurRadius: 14,
            offset: Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFE7C36A),
                  borderRadius: BorderRadius.circular(999),
                ),
                child: Text(
                  'Important Quest Available',
                  style: Theme.of(context).textTheme.labelMedium?.copyWith(
                    color: const Color(0xFF3A1A11),
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            headline,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: Colors.white,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 6),
          Text(
            'Head to $locationName.',
            style: Theme.of(
              context,
            ).textTheme.bodyMedium?.copyWith(color: const Color(0xFFF8EBD0)),
          ),
          const SizedBox(height: 14),
          FilledButton.tonal(
            onPressed: onTap,
            style: FilledButton.styleFrom(
              backgroundColor: const Color(0xFFF1E2BD),
              foregroundColor: const Color(0xFF4E0F17),
            ),
            child: Text(hasQuestGiver ? 'Go to Questgiver' : 'Follow Lead'),
          ),
        ],
      ),
    );
  }
}

class _ResolvedMainStoryLead {
  const _ResolvedMainStoryLead({required this.poi, this.questGiverName});

  final PointOfInterest poi;
  final String? questGiverName;
}

class _QuestPoiCard extends StatelessWidget {
  const _QuestPoiCard({
    required this.node,
    required this.poi,
    required this.discoveredPoiIds,
    required this.objectiveSummary,
    required this.onTap,
  });

  final QuestNode node;
  final PointOfInterest poi;
  final Set<String> discoveredPoiIds;
  final String objectiveSummary;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey.shade300),
        ),
        child: Row(
          children: [
            QuestObjectiveIcon(
              node: node,
              discoveredPoiIds: discoveredPoiIds,
              size: 48,
              borderRadius: 6,
              iconColor: Theme.of(context).colorScheme.onSurface,
              backgroundColor: Colors.grey.shade300,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    poi.name,
                    style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  if (objectiveSummary.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      objectiveSummary,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey.shade700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const Icon(Icons.chevron_right),
          ],
        ),
      ),
    );
  }
}

class _QuestAccordion extends StatelessWidget {
  const _QuestAccordion({
    super.key,
    required this.title,
    required this.quests,
    required this.discoveredPoiIds,
    required this.expanded,
    required this.onToggle,
    required this.onQuestTap,
    this.onReadyQuestTap,
  });

  final String title;
  final List<Quest> quests;
  final Set<String> discoveredPoiIds;
  final bool expanded;
  final VoidCallback onToggle;
  final void Function(Quest) onQuestTap;
  final void Function(Quest)? onReadyQuestTap;

  @override
  Widget build(BuildContext context) {
    final isMainStorySection =
        title == 'Main Story' || title == 'Main Story Archive';
    final headerColor = isMainStorySection
        ? const Color(0xFFF8E7D0)
        : Theme.of(context).colorScheme.surface;
    final borderColor = isMainStorySection
        ? const Color(0xFFE7C36A)
        : Colors.transparent;
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Material(
        elevation: 1,
        borderRadius: BorderRadius.circular(12),
        color: headerColor,
        child: Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: borderColor),
          ),
          child: Column(
            children: [
              InkWell(
                onTap: onToggle,
                borderRadius: BorderRadius.circular(12),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                  child: Row(
                    children: [
                      if (title == 'Tracked Quests')
                        const Padding(
                          padding: EdgeInsets.only(right: 8),
                          child: Text('⭐', style: TextStyle(fontSize: 20)),
                        ),
                      if (isMainStorySection)
                        const Padding(
                          padding: EdgeInsets.only(right: 8),
                          child: Icon(
                            Icons.auto_stories_outlined,
                            size: 20,
                            color: Color(0xFF7A1823),
                          ),
                        ),
                      Expanded(
                        child: Text(
                          title,
                          style: Theme.of(context).textTheme.titleMedium
                              ?.copyWith(fontWeight: FontWeight.w600),
                        ),
                      ),
                      Text(
                        '(${quests.length})',
                        style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Theme.of(
                            context,
                          ).colorScheme.onSurface.withValues(alpha: 0.7),
                        ),
                      ),
                      Icon(
                        expanded ? Icons.expand_less : Icons.expand_more,
                        color: Theme.of(
                          context,
                        ).colorScheme.onSurface.withValues(alpha: 0.7),
                      ),
                    ],
                  ),
                ),
              ),
              if (expanded)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                  child: Column(
                    children: quests.map((q) {
                      final node = q.currentNode;
                      final objectiveSummary = questObjectiveSummary(node);
                      return InkWell(
                        onTap: () {
                          if (q.readyToTurnIn && onReadyQuestTap != null) {
                            onReadyQuestTap!(q);
                            return;
                          }
                          onQuestTap(q);
                        },
                        borderRadius: BorderRadius.circular(8),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 10,
                          ),
                          child: Row(
                            children: [
                              q.turnedInAt != null || q.readyToTurnIn
                                  ? Container(
                                      width: 22,
                                      height: 22,
                                      decoration: const BoxDecoration(
                                        color: Color(0xFF3BB54A),
                                        shape: BoxShape.circle,
                                      ),
                                      child: const Icon(
                                        Icons.check,
                                        size: 14,
                                        color: Colors.white,
                                      ),
                                    )
                                  : Icon(
                                      q.isAccepted
                                          ? Icons.play_circle_fill
                                          : Icons.radio_button_unchecked,
                                      size: 22,
                                      color: q.isAccepted
                                          ? (q.isMainStory
                                                ? const Color(0xFF7A1823)
                                                : Colors.orange)
                                          : Colors.grey.shade400,
                                    ),
                              const SizedBox(width: 12),
                              QuestObjectiveIcon(
                                node: node,
                                discoveredPoiIds: discoveredPoiIds,
                                size: 36,
                                borderRadius: 6,
                                iconColor: Theme.of(
                                  context,
                                ).colorScheme.onSurface,
                                backgroundColor: Colors.grey.shade300,
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        Expanded(
                                          child: Text(
                                            q.name,
                                            style: Theme.of(
                                              context,
                                            ).textTheme.bodyLarge,
                                          ),
                                        ),
                                        if (q.isMainStory)
                                          Container(
                                            margin: const EdgeInsets.only(
                                              left: 8,
                                            ),
                                            padding: const EdgeInsets.symmetric(
                                              horizontal: 8,
                                              vertical: 4,
                                            ),
                                            decoration: BoxDecoration(
                                              color: const Color(0xFF7A1823),
                                              borderRadius:
                                                  BorderRadius.circular(999),
                                            ),
                                            child: Text(
                                              'Story',
                                              style: Theme.of(context)
                                                  .textTheme
                                                  .labelSmall
                                                  ?.copyWith(
                                                    color: Colors.white,
                                                    fontWeight: FontWeight.w700,
                                                  ),
                                            ),
                                          ),
                                      ],
                                    ),
                                    if (objectiveSummary.isNotEmpty) ...[
                                      const SizedBox(height: 2),
                                      Text(
                                        objectiveSummary,
                                        maxLines: 2,
                                        overflow: TextOverflow.ellipsis,
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
                                    ],
                                  ],
                                ),
                              ),
                              if (q.completionCount > 1)
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.orange.shade50,
                                    borderRadius: BorderRadius.circular(999),
                                    border: Border.all(
                                      color: Colors.orange.shade200,
                                    ),
                                  ),
                                  child: Text(
                                    'x${q.completionCount}',
                                    style: Theme.of(context)
                                        .textTheme
                                        .labelSmall
                                        ?.copyWith(
                                          color: Colors.orange.shade800,
                                          fontWeight: FontWeight.w600,
                                        ),
                                  ),
                                ),
                            ],
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
