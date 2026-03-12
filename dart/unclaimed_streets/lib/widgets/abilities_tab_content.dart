import 'dart:async';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/spell.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/party_provider.dart';

class AbilitiesTabContent extends StatefulWidget {
  const AbilitiesTabContent({super.key});

  @override
  State<AbilitiesTabContent> createState() => _AbilitiesTabContentState();
}

class _AbilitiesTabContentState extends State<AbilitiesTabContent> {
  static const Duration _combatTurnDuration = Duration(seconds: 150);
  static const String _targetedHealEffectType = 'restore_life_party_member';
  static const String _groupHealEffectType = 'restore_life_all_party_members';
  static const String _targetedReviveEffectType = 'revive_party_member';
  static const String _groupReviveEffectType =
      'revive_all_downed_party_members';
  static const Duration _feedbackDuration = Duration(seconds: 4);

  Timer? _cooldownTicker;
  Timer? _feedbackTimer;
  final Map<String, DateTime> _cooldownExpiresAtByAbilityId = {};
  final Map<String, int> _lastServerCooldownSecondsByAbilityId = {};
  String? _castingAbilityId;
  String? _feedbackMessage;
  bool _feedbackIsError = false;
  DateTime _now = DateTime.now();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<PartyProvider>().refresh();
    });
  }

  @override
  void dispose() {
    _cooldownTicker?.cancel();
    _feedbackTimer?.cancel();
    super.dispose();
  }

  void _scheduleFeedbackClear() {
    _feedbackTimer?.cancel();
    if (_feedbackMessage == null || _feedbackMessage!.trim().isEmpty) {
      return;
    }
    _feedbackTimer = Timer(_feedbackDuration, () {
      if (!mounted) return;
      setState(() {
        _feedbackMessage = null;
        _feedbackIsError = false;
      });
    });
  }

  int _targetedHealAmount(Spell ability) {
    return ability.effects
        .where((effect) => effect.type == _targetedHealEffectType)
        .fold<int>(
          0,
          (sum, effect) => sum + (effect.amount > 0 ? effect.amount : 0),
        );
  }

  int _groupHealAmount(Spell ability) {
    return ability.effects
        .where((effect) => effect.type == _groupHealEffectType)
        .fold<int>(
          0,
          (sum, effect) => sum + (effect.amount > 0 ? effect.amount : 0),
        );
  }

  int _targetedReviveAmount(Spell ability) {
    return ability.effects
        .where((effect) => effect.type == _targetedReviveEffectType)
        .fold<int>(
          0,
          (sum, effect) => sum + (effect.amount > 0 ? effect.amount : 0),
        );
  }

  int _groupReviveAmount(Spell ability) {
    return ability.effects
        .where((effect) => effect.type == _groupReviveEffectType)
        .fold<int>(
          0,
          (sum, effect) => sum + (effect.amount > 0 ? effect.amount : 0),
        );
  }

  String _displayName(User user) {
    if (user.username.trim().isNotEmpty) return '@${user.username.trim()}';
    if (user.name.trim().isNotEmpty) return user.name.trim();
    if (user.phoneNumber.trim().isNotEmpty) return user.phoneNumber.trim();
    return user.id;
  }

  List<User> _partyTargets(User currentUser, PartyProvider partyProvider) {
    final byId = <String, User>{currentUser.id: currentUser};
    final partyMembers = partyProvider.party?.members ?? const <User>[];
    for (final member in partyMembers) {
      byId[member.id] = member;
    }
    return byId.values.toList();
  }

  Future<void> _castAbility(
    Spell ability, {
    required bool isTechnique,
    String? targetUserId,
  }) async {
    if (_castingAbilityId != null) return;
    _feedbackTimer?.cancel();
    setState(() {
      _castingAbilityId = ability.id;
      _feedbackMessage = null;
      _feedbackIsError = false;
    });

    final provider = context.read<CharacterStatsProvider>();
    final error = isTechnique
        ? await provider.castTechnique(ability.id, targetUserId: targetUserId)
        : await provider.castSpell(ability.id, targetUserId: targetUserId);

    if (!mounted) return;
    setState(() {
      _castingAbilityId = null;
      if (error == null) {
        if (isTechnique) {
          final turns = ability.cooldownTurnsRemaining > 0
              ? ability.cooldownTurnsRemaining
              : ability.cooldownTurns;
          if (turns > 0) {
            _cooldownExpiresAtByAbilityId[ability.id] = DateTime.now().add(
              _combatTurnDuration * turns,
            );
            _now = DateTime.now();
          }
        }
        _feedbackMessage = '${ability.name} used successfully.';
        _feedbackIsError = false;
      } else {
        _feedbackMessage = error;
        _feedbackIsError = true;
      }
    });
    _scheduleFeedbackClear();
  }

  Future<String?> _selectTargetForAbility(
    Spell ability, {
    required List<User> targets,
  }) async {
    if (targets.isEmpty) return null;
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
                    'Choose a target',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    ability.name,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Flexible(
                    child: ListView.separated(
                      shrinkWrap: true,
                      itemCount: targets.length,
                      separatorBuilder: (_, _) => const SizedBox(height: 8),
                      itemBuilder: (context, index) {
                        final user = targets[index];
                        return FilledButton.tonal(
                          onPressed: () =>
                              Navigator.of(dialogContext).pop(user.id),
                          style: FilledButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 12,
                            ),
                            alignment: Alignment.centerLeft,
                          ),
                          child: Text(_displayName(user)),
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

  Future<void> _handleUseSupportAbility(
    Spell ability, {
    required bool isTechnique,
    required List<User> targets,
  }) async {
    if (isTechnique && _cooldownRemaining(ability) > Duration.zero) {
      return;
    }
    final requiresTarget =
        _targetedHealAmount(ability) + _targetedReviveAmount(ability) > 0;
    String? targetUserId;
    if (requiresTarget) {
      targetUserId = await _selectTargetForAbility(ability, targets: targets);
      if (!mounted || targetUserId == null || targetUserId.isEmpty) {
        return;
      }
    }
    await _castAbility(
      ability,
      isTechnique: isTechnique,
      targetUserId: targetUserId,
    );
  }

  Duration _cooldownRemaining(Spell ability) {
    final expiresAt =
        ability.cooldownExpiresAt ?? _cooldownExpiresAtByAbilityId[ability.id];
    if (expiresAt != null) {
      final remaining = expiresAt.difference(_now);
      if (remaining > Duration.zero) {
        return remaining;
      }
    }
    final remainingSeconds = ability.cooldownSecondsRemaining;
    if (remainingSeconds > 0) {
      return Duration(seconds: remainingSeconds);
    }
    final remainingTurns = ability.cooldownTurnsRemaining;
    if (remainingTurns <= 0) {
      return Duration.zero;
    }
    return _combatTurnDuration * remainingTurns;
  }

  void _syncCooldownEstimates(List<Spell> abilities) {
    final activeAbilityIds = <String>{};
    for (final ability in abilities) {
      final abilityId = ability.id.trim();
      if (abilityId.isEmpty) continue;

      final exactExpiresAt = ability.cooldownExpiresAt;
      if (exactExpiresAt != null && exactExpiresAt.isAfter(_now)) {
        _cooldownExpiresAtByAbilityId[abilityId] = exactExpiresAt;
        _lastServerCooldownSecondsByAbilityId.remove(abilityId);
        activeAbilityIds.add(abilityId);
        continue;
      }

      final remainingSeconds = ability.cooldownSecondsRemaining;
      if (remainingSeconds > 0) {
        final previousServerSeconds =
            _lastServerCooldownSecondsByAbilityId[abilityId];
        final existing = _cooldownExpiresAtByAbilityId[abilityId];
        if (existing == null ||
            !existing.isAfter(_now) ||
            previousServerSeconds != remainingSeconds) {
          _cooldownExpiresAtByAbilityId[abilityId] = _now.add(
            Duration(seconds: remainingSeconds),
          );
          _lastServerCooldownSecondsByAbilityId[abilityId] = remainingSeconds;
        }
        activeAbilityIds.add(abilityId);
        continue;
      }

      final existing = _cooldownExpiresAtByAbilityId[abilityId];
      if (existing != null && existing.isAfter(_now)) {
        activeAbilityIds.add(abilityId);
        continue;
      }

      _cooldownExpiresAtByAbilityId.remove(abilityId);
      _lastServerCooldownSecondsByAbilityId.remove(abilityId);
    }

    _cooldownExpiresAtByAbilityId.removeWhere(
      (abilityId, _) => !activeAbilityIds.contains(abilityId),
    );
    _lastServerCooldownSecondsByAbilityId.removeWhere(
      (abilityId, _) => !activeAbilityIds.contains(abilityId),
    );
  }

  String _formatCooldownRemaining(Duration remaining) {
    final totalSeconds = remaining.inSeconds;
    if (totalSeconds <= 0) return '0:00';
    final hours = totalSeconds ~/ 3600;
    final minutes = (totalSeconds % 3600) ~/ 60;
    final seconds = totalSeconds % 60;
    if (hours > 0) {
      return '$hours:${minutes.toString().padLeft(2, '0')}:${seconds.toString().padLeft(2, '0')}';
    }
    return '$minutes:${seconds.toString().padLeft(2, '0')}';
  }

  void _syncCooldownTicker(List<Spell> abilities) {
    final needsTicker = abilities.any(
      (ability) => _cooldownRemaining(ability) > Duration.zero,
    );
    if (!needsTicker) {
      _cooldownTicker?.cancel();
      _cooldownTicker = null;
      return;
    }
    _cooldownTicker ??= Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) {
        _cooldownTicker?.cancel();
        _cooldownTicker = null;
        return;
      }
      setState(() {
        _now = DateTime.now();
      });
    });
  }

  Widget _buildAbilitySection({
    required String emptyMessage,
    required List<Spell> abilities,
    required bool isTechnique,
    required int mana,
    required List<User> targets,
    required ThemeData theme,
  }) {
    if (abilities.isEmpty) {
      return Text(
        emptyMessage,
        style: theme.textTheme.bodySmall?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
      );
    }

    return Column(
      children: abilities.map((ability) {
        return _AbilityRow(
          ability: ability,
          mana: mana,
          isTechnique: isTechnique,
          casting: _castingAbilityId == ability.id,
          targetedHealAmount: _targetedHealAmount(ability),
          groupHealAmount: _groupHealAmount(ability),
          targetedReviveAmount: _targetedReviveAmount(ability),
          groupReviveAmount: _groupReviveAmount(ability),
          cooldownRemaining: _cooldownRemaining(ability),
          cooldownLabel: _formatCooldownRemaining(_cooldownRemaining(ability)),
          onUse: () => _handleUseSupportAbility(
            ability,
            isTechnique: isTechnique,
            targets: targets,
          ),
        );
      }).toList(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final user = context.watch<AuthProvider>().user;
    if (user == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('Log in to see your abilities.'),
        ),
      );
    }

    final statsProvider = context.watch<CharacterStatsProvider>();
    final partyProvider = context.watch<PartyProvider>();
    final spells = statsProvider.spells;
    final techniques = statsProvider.techniques;
    final loading = statsProvider.loading;
    final mana = statsProvider.mana;
    final theme = Theme.of(context);
    final targets = _partyTargets(user, partyProvider);
    _syncCooldownEstimates(techniques);
    _syncCooldownTicker(techniques);

    return SizedBox.expand(
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(4, 8, 4, 12),
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: theme.colorScheme.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: theme.colorScheme.outlineVariant),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (_feedbackMessage != null) ...[
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 8,
                  ),
                  decoration: BoxDecoration(
                    color: _feedbackIsError
                        ? theme.colorScheme.errorContainer
                        : theme.colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: _feedbackIsError
                          ? theme.colorScheme.error.withValues(alpha: 0.35)
                          : theme.colorScheme.primary.withValues(alpha: 0.35),
                    ),
                  ),
                  child: Text(
                    _feedbackMessage!,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: _feedbackIsError
                          ? theme.colorScheme.onErrorContainer
                          : theme.colorScheme.onPrimaryContainer,
                    ),
                  ),
                ),
                const SizedBox(height: 8),
              ],
              if (loading && spells.isEmpty && techniques.isEmpty)
                const Center(
                  child: Padding(
                    padding: EdgeInsets.all(12),
                    child: CircularProgressIndicator(),
                  ),
                ),
              if (!loading) ...[
                Text(
                  'Spells',
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 8),
                _buildAbilitySection(
                  emptyMessage: 'No spells learned yet.',
                  abilities: spells,
                  isTechnique: false,
                  mana: mana,
                  targets: targets,
                  theme: theme,
                ),
                const SizedBox(height: 10),
                Text(
                  'Techniques',
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 8),
                _buildAbilitySection(
                  emptyMessage: 'No techniques learned yet.',
                  abilities: techniques,
                  isTechnique: true,
                  mana: mana,
                  targets: targets,
                  theme: theme,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _AbilityRow extends StatelessWidget {
  const _AbilityRow({
    required this.ability,
    required this.mana,
    required this.isTechnique,
    required this.casting,
    required this.targetedHealAmount,
    required this.groupHealAmount,
    required this.targetedReviveAmount,
    required this.groupReviveAmount,
    required this.cooldownRemaining,
    required this.cooldownLabel,
    required this.onUse,
  });

  final Spell ability;
  final int mana;
  final bool isTechnique;
  final bool casting;
  final int targetedHealAmount;
  final int groupHealAmount;
  final int targetedReviveAmount;
  final int groupReviveAmount;
  final Duration cooldownRemaining;
  final String cooldownLabel;
  final VoidCallback onUse;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasTargetedSupport =
        targetedHealAmount > 0 || targetedReviveAmount > 0;
    final hasGroupSupport = groupHealAmount > 0 || groupReviveAmount > 0;
    final isSupportAbility = hasTargetedSupport || hasGroupSupport;
    final onCooldown = isTechnique && cooldownRemaining > Duration.zero;
    final hasEnoughMana = isTechnique || mana >= ability.manaCost;
    final canUse = isSupportAbility && hasEnoughMana && !casting && !onCooldown;
    final useLabel = casting ? 'Using...' : 'Use';
    final cooldownTurns = ability.cooldownTurns < 0 ? 0 : ability.cooldownTurns;

    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (ability.iconUrl.trim().isNotEmpty)
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: Image.network(
                ability.iconUrl,
                width: 32,
                height: 32,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) =>
                    const Icon(Icons.auto_fix_high, size: 20),
              ),
            )
          else
            Icon(
              isTechnique ? Icons.sports_martial_arts : Icons.auto_fix_high,
              size: 20,
            ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  ability.name,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                if (ability.schoolOfMagic.trim().isNotEmpty)
                  Text(
                    isTechnique
                        ? onCooldown
                              ? '${ability.schoolOfMagic} · Technique · Ready in $cooldownLabel'
                              : cooldownTurns > 0
                              ? '${ability.schoolOfMagic} · Technique · Cooldown $cooldownTurns turn${cooldownTurns == 1 ? '' : 's'}'
                              : '${ability.schoolOfMagic} · Technique'
                        : '${ability.schoolOfMagic} · Mana ${ability.manaCost}',
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                if (isTechnique &&
                    ability.schoolOfMagic.trim().isEmpty &&
                    onCooldown)
                  Text(
                    'Ready in $cooldownLabel',
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                if (ability.effectText.trim().isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    ability.effectText.trim(),
                    style: theme.textTheme.bodySmall,
                  ),
                ],
                if (isSupportAbility) ...[
                  const SizedBox(height: 8),
                  Opacity(
                    opacity: hasEnoughMana ? 1.0 : 0.55,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        FilledButton(
                          onPressed: canUse ? onUse : null,
                          child: Text(useLabel),
                        ),
                      ],
                    ),
                  ),
                  if (!hasEnoughMana && !isTechnique) ...[
                    const SizedBox(height: 6),
                    Text(
                      'Not enough mana to cast this spell.',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}
