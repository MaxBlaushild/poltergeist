import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/character_stats.dart';
import '../models/user.dart';
import '../models/user_level.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/user_level_provider.dart';

class CharacterTabContent extends StatefulWidget {
  const CharacterTabContent({
    super.key,
    this.userOverride,
    this.statsOverride,
    this.userLevelOverride,
    this.readOnly = false,
  });

  final User? userOverride;
  final CharacterStats? statsOverride;
  final UserLevel? userLevelOverride;
  final bool readOnly;

  @override
  State<CharacterTabContent> createState() => _CharacterTabContentState();
}

class _CharacterTabContentState extends State<CharacterTabContent> {
  static const Map<String, String> _labels = {
    'strength': 'Strength',
    'dexterity': 'Dexterity',
    'constitution': 'Constitution',
    'intelligence': 'Intelligence',
    'wisdom': 'Wisdom',
    'charisma': 'Charisma',
  };

  String? _lastUserId;
  Map<String, int> _pending = {};
  final ScrollController _scrollController = ScrollController();
  bool _showTopFade = false;
  bool _showBottomFade = false;

  int get _pendingTotal =>
      _pending.values.where((value) => value > 0).fold(0, (a, b) => a + b);

  bool get _isReadOnly =>
      widget.readOnly ||
      widget.userOverride != null ||
      widget.statsOverride != null ||
      widget.userLevelOverride != null;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_updateFades);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final uid =
        widget.userOverride?.id ?? context.watch<AuthProvider>().user?.id;
    if (uid != _lastUserId) {
      _lastUserId = uid;
      _pending = {};
    }
    final unspent = widget.statsOverride?.unspentPoints ??
        context.watch<CharacterStatsProvider>().unspentPoints;
    if (unspent == 0 && _pending.isNotEmpty) {
      _pending = {};
    }
    if (_isReadOnly && _pending.isNotEmpty) {
      _pending = {};
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  @override
  void dispose() {
    _scrollController.removeListener(_updateFades);
    _scrollController.dispose();
    super.dispose();
  }

  void _bumpStat(String key, int delta, int remaining) {
    if (_isReadOnly) return;
    if (delta > 0 && remaining <= 0) return;
    final current = _pending[key] ?? 0;
    final next = (current + delta).clamp(0, 999);
    setState(() {
      if (next == 0) {
        _pending.remove(key);
      } else {
        _pending[key] = next;
      }
    });
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  Future<void> _confirmAllocations(CharacterStatsProvider stats) async {
    if (_isReadOnly || _pendingTotal == 0) return;
    final success = await stats.applyAllocations(_pending);
    if (!mounted) return;
    if (success) {
      setState(() => _pending = {});
      WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Unable to apply stat points.')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
    final auth = context.watch<AuthProvider>();
    final levels = context.watch<UserLevelProvider>();
    final statsProvider = context.watch<CharacterStatsProvider>();
    final user = widget.userOverride ?? auth.user;
    if (user == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('Log in to see your character.'),
        ),
      );
    }
    final userLevel = widget.userLevelOverride ?? levels.userLevel;
    final levelLoading = widget.userLevelOverride != null ? false : levels.loading;
    final displayLevel =
        widget.userLevelOverride?.level ??
        widget.statsOverride?.level ??
        userLevel?.level ??
        statsProvider.level;
    final overrideStats = widget.statsOverride;
    final baseStats =
        overrideStats?.toMap() ?? statsProvider.baseStats;
    final bonusStats =
        overrideStats?.bonusMap() ?? statsProvider.equipmentBonuses;
    final unspentPoints =
        overrideStats?.unspentPoints ?? statsProvider.unspentPoints;
    final hasUnspentPoints = unspentPoints > 0;
    final proficiencies =
        overrideStats?.proficiencies ?? statsProvider.proficiencies;
    final hasProficiencies = proficiencies.isNotEmpty;
    final canEdit = !_isReadOnly;

    void showProfileImage() {
      if (user.profilePictureUrl.isEmpty) return;
      showDialog<void>(
        context: context,
        barrierColor: Colors.black54,
        builder: (context) {
          final theme = Theme.of(context);
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
                    border: Border.all(
                      color: theme.colorScheme.outlineVariant,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.2),
                        blurRadius: 18,
                        offset: const Offset(0, 10),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: Image.network(
                      user.profilePictureUrl,
                      width: 320,
                      height: 320,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        width: 320,
                        height: 320,
                        color: theme.colorScheme.surfaceVariant,
                        child: const Icon(Icons.person, size: 96),
                      ),
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.of(context).pop(),
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

    return Stack(
      children: [
        SingleChildScrollView(
          controller: _scrollController,
          padding: const EdgeInsets.only(bottom: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  GestureDetector(
                    onTap: user.profilePictureUrl.isNotEmpty
                        ? showProfileImage
                        : null,
                    child: Stack(
                      clipBehavior: Clip.none,
                      children: [
                        CircleAvatar(
                          radius: 28,
                          backgroundColor: Colors.grey.shade300,
                          backgroundImage: user.profilePictureUrl.isNotEmpty
                              ? NetworkImage(user.profilePictureUrl)
                              : null,
                          child: user.profilePictureUrl.isEmpty
                              ? const Icon(Icons.person)
                              : null,
                        ),
                        if (hasUnspentPoints)
                          Positioned(
                            right: -2,
                            top: -2,
                            child: Container(
                              padding: const EdgeInsets.all(3),
                              decoration: BoxDecoration(
                                color: const Color(0xFFFFD54F),
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: theme.colorScheme.surface,
                                  width: 1.2,
                                ),
                              ),
                              child: const Icon(
                                Icons.arrow_upward,
                                size: 12,
                                color: Colors.white,
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          user.username.isNotEmpty ? user.username : user.name,
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        if (user.username.isNotEmpty && user.name != user.username)
                          Text(
                            user.name,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: levelLoading
                    ? Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Text(
                            'Level',
                            style: theme.textTheme.titleSmall?.copyWith(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          const SizedBox(height: 10),
                          LinearProgressIndicator(
                            minHeight: 8,
                            color: theme.colorScheme.primary,
                            backgroundColor:
                                theme.colorScheme.surfaceContainerHighest,
                          ),
                        ],
                      )
                    : userLevel == null
                        ? Text(
                            'Level data unavailable right now.',
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          )
                        : Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              Row(
                                children: [
                                  Text(
                                    'Level ${userLevel.level}',
                                    style: theme.textTheme.titleSmall?.copyWith(
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                  const Spacer(),
                                  Text(
                                    '${userLevel.experiencePointsOnLevel} / ${userLevel.experienceToNextLevel} XP',
                                    style: theme.textTheme.bodySmall?.copyWith(
                                      color:
                                          theme.colorScheme.onSurfaceVariant,
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              ClipRRect(
                                borderRadius: BorderRadius.circular(999),
                                child: LinearProgressIndicator(
                                  value: userLevel.experienceToNextLevel > 0
                                      ? (userLevel.experiencePointsOnLevel /
                                              userLevel
                                                  .experienceToNextLevel)
                                          .clamp(0.0, 1.0)
                                      : 0.0,
                                  minHeight: 8,
                                  color: theme.colorScheme.primary,
                                  backgroundColor: theme
                                      .colorScheme.surfaceContainerHighest,
                                ),
                              ),
                            ],
                          ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Character stats',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Text(
                          'Level $displayLevel',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const Spacer(),
                        Text(
                          'Unspent: $unspentPoints',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: hasUnspentPoints
                                ? theme.colorScheme.primary
                                : theme.colorScheme.onSurfaceVariant,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                    if (hasUnspentPoints) ...[
                      const SizedBox(height: 6),
                      Text(
                        'Level up!',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: const Color(0xFFC58A00),
                          fontWeight: FontWeight.w700,
                          letterSpacing: 0.3,
                        ),
                      ),
                    ],
                    if (canEdit && _pendingTotal > 0) ...[
                      const SizedBox(height: 6),
                      Text(
                        'Remaining after pending: ${unspentPoints - _pendingTotal}',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    const SizedBox(height: 12),
                    Column(
                      children: CharacterStatsProvider.statKeys.map((key) {
                        final label = _labels[key] ?? key;
                        final baseValue = baseStats[key] ??
                            CharacterStatsProvider.baseStatValue;
                        final bonusValue = bonusStats[key] ?? 0;
                        final pendingValue = _pending[key] ?? 0;
                        final displayValue = baseValue + pendingValue;
                        final remaining = unspentPoints - _pendingTotal;
                        final canAdd = canEdit && remaining > 0;
                        final canRemove = canEdit && pendingValue > 0;
                        return Container(
                          margin: const EdgeInsets.only(bottom: 8),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.surfaceContainerHighest,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                                color: theme.colorScheme.outlineVariant),
                          ),
                          child: Row(
                            children: [
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      label,
                                      style: theme.textTheme.bodyMedium?.copyWith(
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                                    if (pendingValue > 0)
                                      Text(
                                        '+$pendingValue pending',
                                        style: theme.textTheme.bodySmall?.copyWith(
                                          color: theme.colorScheme.primary,
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                              SizedBox(
                                width: 48,
                                child: Align(
                                  alignment: Alignment.centerRight,
                                  child: Text(
                                    bonusValue > 0 ? '+$bonusValue' : '',
                                    style: theme.textTheme.bodyMedium?.copyWith(
                                      color: theme.colorScheme.primary,
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                ),
                              ),
                              const SizedBox(width: 6),
                              Text(
                                '$displayValue',
                                style: theme.textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                              const SizedBox(width: 8),
                              IconButton(
                                visualDensity: VisualDensity.compact,
                                onPressed: canRemove
                                    ? () => _bumpStat(key, -1, remaining)
                                    : null,
                                icon: const Icon(Icons.remove_circle_outline),
                              ),
                              IconButton(
                                visualDensity: VisualDensity.compact,
                                onPressed: canAdd
                                    ? () => _bumpStat(key, 1, remaining)
                                    : null,
                                icon: const Icon(Icons.add_circle_outline),
                              ),
                            ],
                          ),
                        );
                      }).toList(),
                    ),
                    if (canEdit && _pendingTotal > 0) ...[
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton(
                              onPressed: () => setState(() => _pending = {}),
                              child: const Text('Cancel'),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: FilledButton(
                              onPressed: () => _confirmAllocations(statsProvider),
                              child: const Text('Confirm'),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Proficiencies',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    if (!hasProficiencies)
                      Text(
                        'No proficiencies yet. Complete quests to earn them.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    if (hasProficiencies)
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: proficiencies.map((proficiency) {
                          return Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 10, vertical: 6),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surfaceContainerHighest,
                              borderRadius: BorderRadius.circular(999),
                              border: Border.all(
                                color: theme.colorScheme.outlineVariant,
                              ),
                            ),
                            child: Text(
                              '${proficiency.proficiency} · ${proficiency.level}',
                              style: theme.textTheme.bodySmall?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          );
                        }).toList(),
                      ),
                  ],
                ),
              ),
              if (canEdit) ...[
                const SizedBox(height: 16),
                OutlinedButton(
                  onPressed: () {
                    Navigator.of(context).pop();
                    context.go('/logout');
                  },
                  child: const Text('Log out'),
                ),
              ],
            ],
          ),
        ),
        if (_showTopFade)
          Positioned(
            left: 0,
            right: 0,
            top: 0,
            child: IgnorePointer(
              child: Container(
                height: 18,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      theme.colorScheme.surface,
                      theme.colorScheme.surface.withOpacity(0.0),
                    ],
                  ),
                ),
              ),
            ),
          ),
        if (_showBottomFade)
          Positioned(
            left: 0,
            right: 0,
            bottom: 0,
            child: IgnorePointer(
              child: Container(
                height: 20,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                    colors: [
                      theme.colorScheme.surface,
                      theme.colorScheme.surface.withOpacity(0.0),
                    ],
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }

  void _updateFades() {
    if (!mounted || !_scrollController.hasClients) return;
    final position = _scrollController.position;
    final showTop = position.pixels > 0.5;
    final showBottom = position.pixels < position.maxScrollExtent - 0.5;
    if (showTop == _showTopFade && showBottom == _showBottomFade) return;
    setState(() {
      _showTopFade = showTop;
      _showBottomFade = showBottom;
    });
  }
}
