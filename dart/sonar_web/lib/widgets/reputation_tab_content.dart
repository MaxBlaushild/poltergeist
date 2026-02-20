import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/user_zone_reputation.dart';
import '../models/zone.dart';
import '../providers/auth_provider.dart';
import '../providers/zone_provider.dart';
import '../services/poi_service.dart';

class ReputationTabContent extends StatefulWidget {
  const ReputationTabContent({super.key});

  @override
  State<ReputationTabContent> createState() => _ReputationTabContentState();
}

class _ReputationTabContentState extends State<ReputationTabContent> {
  bool _loading = true;
  String? _error;
  String? _lastUserId;
  List<_ZoneRep> _items = const [];
  _ReputationSort _sort = _ReputationSort.zoneName;
  _ReputationFilter _filter = _ReputationFilter.all;
  String _query = '';

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _load();
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final uid = context.watch<AuthProvider>().user?.id;
    if (uid != _lastUserId) {
      _lastUserId = uid;
      if (uid == null) {
        setState(() {
          _items = const [];
          _loading = false;
          _error = null;
        });
      } else {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!mounted) return;
          _load();
        });
      }
    }
  }

  Future<void> _load() async {
    final user = context.read<AuthProvider>().user;
    if (user == null) {
      if (mounted) {
        setState(() {
          _items = const [];
          _loading = false;
          _error = null;
        });
      }
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final zoneProvider = context.read<ZoneProvider>();
      var zones = zoneProvider.zones;
      if (zones.isEmpty) {
        zones = await context.read<PoiService>().getZones();
        if (mounted) {
          zoneProvider.setZones(zones);
        }
      }

      final svc = context.read<PoiService>();
      final reputations = await Future.wait(
        zones.map((z) => svc.getUserZoneReputation(z.id)),
      );

      if (!mounted) return;
      final items = <_ZoneRep>[];
      for (var i = 0; i < zones.length; i++) {
        items.add(_ZoneRep(zone: zones[i], reputation: reputations[i]));
      }
      items.sort((a, b) => a.zone.name.toLowerCase().compareTo(b.zone.name.toLowerCase()));
      setState(() {
        _items = items;
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

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final auth = context.watch<AuthProvider>();
    if (auth.user == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('Log in to see your reputations.'),
        ),
      );
    }

    final filteredItems = _applySortAndFilter(_items);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            Text(
              'Zone Reputation',
              style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700),
            ),
            const Spacer(),
            IconButton(
              tooltip: 'Refresh',
              onPressed: _loading ? null : _load,
              icon: const Icon(Icons.refresh),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            ConstrainedBox(
              constraints: const BoxConstraints(minWidth: 180, maxWidth: 260),
              child: TextField(
                style: theme.textTheme.bodySmall,
                decoration: InputDecoration(
                  isDense: true,
                  labelText: 'Search zones',
                  prefixIcon: const Icon(Icons.search),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                onChanged: (value) => setState(() => _query = value),
              ),
            ),
            ConstrainedBox(
              constraints: const BoxConstraints(minWidth: 150),
              child: DropdownButtonFormField<_ReputationSort>(
                value: _sort,
                isDense: true,
                style: theme.textTheme.bodySmall,
                decoration: InputDecoration(
                  labelText: 'Sort',
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                items: _ReputationSort.values
                    .map(
                      (sort) => DropdownMenuItem(
                        value: sort,
                        child: Text(_sortLabel(sort)),
                      ),
                    )
                    .toList(),
                onChanged: (value) {
                  if (value == null) return;
                  setState(() => _sort = value);
                },
              ),
            ),
            ConstrainedBox(
              constraints: const BoxConstraints(minWidth: 160),
              child: DropdownButtonFormField<_ReputationFilter>(
                value: _filter,
                isDense: true,
                style: theme.textTheme.bodySmall,
                decoration: InputDecoration(
                  labelText: 'Filter',
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                items: _ReputationFilter.values
                    .map(
                      (filter) => DropdownMenuItem(
                        value: filter,
                        child: Text(_filterLabel(filter)),
                      ),
                    )
                    .toList(),
                onChanged: (value) {
                  if (value == null) return;
                  setState(() => _filter = value);
                },
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        if (_loading)
          const Expanded(
            child: Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              ),
            ),
          )
        else if (_error != null)
          Expanded(
            child: Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      'Failed to load reputations.',
                      style: theme.textTheme.titleSmall,
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _error!,
                      style: TextStyle(color: theme.colorScheme.error),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 12),
                    OutlinedButton(
                      onPressed: _load,
                      child: const Text('Try again'),
                    ),
                  ],
                ),
              ),
            ),
          )
        else if (filteredItems.isEmpty)
          const Expanded(
            child: Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text('No zones found yet.'),
              ),
            ),
          )
        else
          Expanded(
            child: RefreshIndicator(
              onRefresh: _load,
              child: ListView.builder(
                physics: const AlwaysScrollableScrollPhysics(),
                itemCount: filteredItems.length,
                itemBuilder: (context, index) {
                  final entry = filteredItems[index];
                  return _ReputationCard(entry: entry);
                },
              ),
            ),
          ),
      ],
    );
  }

  List<_ZoneRep> _applySortAndFilter(List<_ZoneRep> items) {
    final query = _query.trim().toLowerCase();
    final filtered = items.where((item) {
      if (query.isNotEmpty &&
          !item.zone.name.toLowerCase().contains(query)) {
        return false;
      }
      if (_filter == _ReputationFilter.all) return true;
      if (_filter == _ReputationFilter.uncharted) {
        return item.reputation == null;
      }
      final repName = _filterToRepName(_filter);
      return item.reputation?.name == repName;
    }).toList();

    filtered.sort((a, b) {
      int compareZoneName() =>
          a.zone.name.toLowerCase().compareTo(b.zone.name.toLowerCase());
      switch (_sort) {
        case _ReputationSort.zoneName:
          return compareZoneName();
        case _ReputationSort.levelDesc:
          final aLevel = a.reputation?.level ?? -1;
          final bLevel = b.reputation?.level ?? -1;
          if (aLevel != bLevel) return bLevel.compareTo(aLevel);
          return compareZoneName();
        case _ReputationSort.totalDesc:
          final aTotal = a.reputation?.totalReputation ?? -1;
          final bTotal = b.reputation?.totalReputation ?? -1;
          if (aTotal != bTotal) return bTotal.compareTo(aTotal);
          return compareZoneName();
        case _ReputationSort.repName:
          final aName = a.reputation?.name.name ?? 'uncharted';
          final bName = b.reputation?.name.name ?? 'uncharted';
          final repCompare = aName.compareTo(bName);
          if (repCompare != 0) return repCompare;
          return compareZoneName();
      }
    });

    return filtered;
  }
}

class _ZoneRep {
  final Zone zone;
  final UserZoneReputation? reputation;

  const _ZoneRep({required this.zone, required this.reputation});
}

class _ReputationCard extends StatelessWidget {
  const _ReputationCard({required this.entry});

  final _ZoneRep entry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final rep = entry.reputation;
    final badge = _badgeStyle(theme.colorScheme, rep?.name);
    final repName = _capitalize(rep?.name.name ?? 'Uncharted');
    final levelText = rep != null ? 'Level ${rep.level}' : 'No rep yet';

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: theme.colorScheme.outlineVariant),
        boxShadow: const [
          BoxShadow(
            color: Color(0x1A2D2416),
            blurRadius: 8,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  entry.zone.name,
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: badge.background,
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(color: badge.border),
                ),
                child: Text(
                  repName,
                  style: theme.textTheme.labelSmall?.copyWith(
                    color: badge.foreground,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 0.2,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            levelText,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          if (entry.zone.description != null &&
              entry.zone.description!.trim().isNotEmpty) ...[
            const SizedBox(height: 6),
            Text(
              entry.zone.description!,
              style: theme.textTheme.bodySmall,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
          const SizedBox(height: 10),
          _repProgress(context, rep),
        ],
      ),
    );
  }

  Widget _repProgress(BuildContext context, UserZoneReputation? rep) {
    final theme = Theme.of(context);
    if (rep == null) {
      return Text(
        'Explore this zone to earn reputation.',
        style: theme.textTheme.bodySmall?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
      );
    }
    if (rep.reputationToNextLevel <= 0) {
      return Text(
        'Max reputation reached.',
        style: theme.textTheme.bodySmall?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
      );
    }

    final progress = rep.reputationOnLevel / rep.reputationToNextLevel;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              '${rep.reputationOnLevel} / ${rep.reputationToNextLevel}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            Text(
              'Total ${rep.totalReputation}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(6),
          child: LinearProgressIndicator(
            value: progress.clamp(0.0, 1.0),
            backgroundColor: theme.colorScheme.surfaceVariant,
            valueColor: AlwaysStoppedAnimation<Color>(theme.colorScheme.primary),
            minHeight: 8,
          ),
        ),
      ],
    );
  }
}

String _capitalize(String value) {
  if (value.isEmpty) return value;
  return value[0].toUpperCase() + value.substring(1);
}

class _BadgeStyle {
  final Color background;
  final Color foreground;
  final Color border;

  const _BadgeStyle({
    required this.background,
    required this.foreground,
    required this.border,
  });
}

enum _ReputationSort {
  zoneName,
  levelDesc,
  totalDesc,
  repName,
}

enum _ReputationFilter {
  all,
  uncharted,
  neutral,
  friendly,
  honored,
  revered,
  exalted,
  legendary,
}

String _sortLabel(_ReputationSort sort) {
  switch (sort) {
    case _ReputationSort.zoneName:
      return 'Zone name';
    case _ReputationSort.levelDesc:
      return 'Level (high to low)';
    case _ReputationSort.totalDesc:
      return 'Total reputation';
    case _ReputationSort.repName:
      return 'Reputation name';
  }
}

String _filterLabel(_ReputationFilter filter) {
  switch (filter) {
    case _ReputationFilter.all:
      return 'All reputations';
    case _ReputationFilter.uncharted:
      return 'Uncharted';
    case _ReputationFilter.neutral:
      return 'Neutral';
    case _ReputationFilter.friendly:
      return 'Friendly';
    case _ReputationFilter.honored:
      return 'Honored';
    case _ReputationFilter.revered:
      return 'Revered';
    case _ReputationFilter.exalted:
      return 'Exalted';
    case _ReputationFilter.legendary:
      return 'Legendary';
  }
}

UserZoneReputationName? _filterToRepName(_ReputationFilter filter) {
  switch (filter) {
    case _ReputationFilter.neutral:
      return UserZoneReputationName.neutral;
    case _ReputationFilter.friendly:
      return UserZoneReputationName.friendly;
    case _ReputationFilter.honored:
      return UserZoneReputationName.honored;
    case _ReputationFilter.revered:
      return UserZoneReputationName.revered;
    case _ReputationFilter.exalted:
      return UserZoneReputationName.exalted;
    case _ReputationFilter.legendary:
      return UserZoneReputationName.legendary;
    case _ReputationFilter.all:
    case _ReputationFilter.uncharted:
      return null;
  }
}

_BadgeStyle _badgeStyle(ColorScheme scheme, UserZoneReputationName? rep) {
  switch (rep) {
    case UserZoneReputationName.friendly:
      return _BadgeStyle(
        background: scheme.secondaryContainer,
        foreground: scheme.onSecondaryContainer,
        border: scheme.secondary,
      );
    case UserZoneReputationName.honored:
      return _BadgeStyle(
        background: scheme.primaryContainer,
        foreground: scheme.onPrimaryContainer,
        border: scheme.primary,
      );
    case UserZoneReputationName.revered:
      return _BadgeStyle(
        background: scheme.tertiaryContainer,
        foreground: scheme.onTertiaryContainer,
        border: scheme.tertiary,
      );
    case UserZoneReputationName.exalted:
      return _BadgeStyle(
        background: scheme.primary.withOpacity(0.12),
        foreground: scheme.primary,
        border: scheme.primary.withOpacity(0.5),
      );
    case UserZoneReputationName.legendary:
      return _BadgeStyle(
        background: scheme.secondary.withOpacity(0.14),
        foreground: scheme.secondary,
        border: scheme.secondary.withOpacity(0.5),
      );
    case UserZoneReputationName.neutral:
    default:
      return _BadgeStyle(
        background: scheme.surfaceVariant,
        foreground: scheme.onSurfaceVariant,
        border: scheme.outlineVariant,
      );
  }
}
