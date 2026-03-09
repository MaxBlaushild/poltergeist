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
  static const String _targetedHealEffectType = 'restore_life_party_member';
  static const String _groupHealEffectType = 'restore_life_all_party_members';
  static const String _targetedReviveEffectType = 'revive_party_member';
  static const String _groupReviveEffectType =
      'revive_all_downed_party_members';

  final Map<String, String> _selectedTargetByAbility = {};
  String? _castingAbilityId;
  String? _feedbackMessage;
  bool _feedbackIsError = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<PartyProvider>().refresh();
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
        _feedbackMessage = '${ability.name} used successfully.';
        _feedbackIsError = false;
      } else {
        _feedbackMessage = error;
        _feedbackIsError = true;
      }
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
        final selectedTarget = _selectedTargetByAbility[ability.id];
        final requiresTarget =
            _targetedHealAmount(ability) + _targetedReviveAmount(ability) > 0;
        final fallbackTarget = targets.isNotEmpty ? targets.first.id : null;
        final isSelectedValid =
            selectedTarget != null &&
            targets.any((member) => member.id == selectedTarget);
        final resolvedTarget = isSelectedValid
            ? selectedTarget
            : fallbackTarget;
        return _AbilityRow(
          ability: ability,
          mana: mana,
          isTechnique: isTechnique,
          targets: targets,
          selectedTargetId: resolvedTarget,
          casting: _castingAbilityId == ability.id,
          displayName: _displayName,
          targetedHealAmount: _targetedHealAmount(ability),
          groupHealAmount: _groupHealAmount(ability),
          targetedReviveAmount: _targetedReviveAmount(ability),
          groupReviveAmount: _groupReviveAmount(ability),
          onTargetChanged: requiresTarget
              ? (value) {
                  setState(() {
                    if (value == null || value.isEmpty) {
                      _selectedTargetByAbility.remove(ability.id);
                    } else {
                      _selectedTargetByAbility[ability.id] = value;
                    }
                  });
                }
              : null,
          onCast: () => _castAbility(
            ability,
            isTechnique: isTechnique,
            targetUserId: requiresTarget ? resolvedTarget : null,
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
    required this.targets,
    required this.selectedTargetId,
    required this.casting,
    required this.displayName,
    required this.targetedHealAmount,
    required this.groupHealAmount,
    required this.targetedReviveAmount,
    required this.groupReviveAmount,
    required this.onCast,
    this.onTargetChanged,
  });

  final Spell ability;
  final int mana;
  final bool isTechnique;
  final List<User> targets;
  final String? selectedTargetId;
  final bool casting;
  final String Function(User user) displayName;
  final int targetedHealAmount;
  final int groupHealAmount;
  final int targetedReviveAmount;
  final int groupReviveAmount;
  final VoidCallback onCast;
  final ValueChanged<String?>? onTargetChanged;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasTargetedSupport =
        targetedHealAmount > 0 || targetedReviveAmount > 0;
    final hasGroupSupport = groupHealAmount > 0 || groupReviveAmount > 0;
    final isSupportAbility = hasTargetedSupport || hasGroupSupport;
    final hasEnoughMana = isTechnique || mana >= ability.manaCost;
    final hasValidTarget =
        !hasTargetedSupport || (selectedTargetId ?? '').isNotEmpty;
    final canCast =
        isSupportAbility && hasEnoughMana && hasValidTarget && !casting;

    final castLabel = casting
        ? 'Casting...'
        : hasTargetedSupport && !hasGroupSupport
        ? 'Use Targeted Support'
        : !hasTargetedSupport && hasGroupSupport
        ? 'Use Group Support'
        : 'Use Support';

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
                        ? '${ability.schoolOfMagic} · Technique'
                        : '${ability.schoolOfMagic} · Mana ${ability.manaCost}',
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
                if (!isSupportAbility) ...[
                  const SizedBox(height: 8),
                  Text(
                    'Casting is only available for support abilities here.',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
                if (isSupportAbility) ...[
                  const SizedBox(height: 8),
                  Opacity(
                    opacity: hasEnoughMana ? 1.0 : 0.55,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        if (hasTargetedSupport)
                          DropdownButtonFormField<String>(
                            value: selectedTargetId,
                            isDense: true,
                            decoration: const InputDecoration(
                              labelText: 'Target teammate',
                              border: OutlineInputBorder(),
                              contentPadding: EdgeInsets.symmetric(
                                horizontal: 10,
                                vertical: 8,
                              ),
                            ),
                            items: targets.map((user) {
                              return DropdownMenuItem<String>(
                                value: user.id,
                                child: Text(displayName(user)),
                              );
                            }).toList(),
                            onChanged: hasEnoughMana ? onTargetChanged : null,
                          ),
                        if (hasTargetedSupport) const SizedBox(height: 8),
                        FilledButton(
                          onPressed: canCast ? onCast : null,
                          child: Text(castLabel),
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
