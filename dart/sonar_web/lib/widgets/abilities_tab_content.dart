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

  final Map<String, String> _selectedTargetBySpell = {};
  String? _castingSpellId;
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

  int _targetedHealAmount(Spell spell) {
    return spell.effects
        .where((effect) => effect.type == _targetedHealEffectType)
        .fold<int>(
          0,
          (sum, effect) => sum + (effect.amount > 0 ? effect.amount : 0),
        );
  }

  int _groupHealAmount(Spell spell) {
    return spell.effects
        .where((effect) => effect.type == _groupHealEffectType)
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

  Future<void> _castSpell(Spell spell, {String? targetUserId}) async {
    if (_castingSpellId != null) return;
    setState(() {
      _castingSpellId = spell.id;
      _feedbackMessage = null;
      _feedbackIsError = false;
    });

    final error = await context.read<CharacterStatsProvider>().castSpell(
      spell.id,
      targetUserId: targetUserId,
    );

    if (!mounted) return;
    setState(() {
      _castingSpellId = null;
      if (error == null) {
        _feedbackMessage = '${spell.name} cast successfully.';
        _feedbackIsError = false;
      } else {
        _feedbackMessage = error;
        _feedbackIsError = true;
      }
    });
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
              Text(
                'Spells',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 8),
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
              if (loading && spells.isEmpty)
                const Center(
                  child: Padding(
                    padding: EdgeInsets.all(12),
                    child: CircularProgressIndicator(),
                  ),
                ),
              if (!loading && spells.isEmpty)
                Text(
                  'No spells learned yet.',
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              if (spells.isNotEmpty)
                Column(
                  children: spells.map((spell) {
                    final selectedTarget = _selectedTargetBySpell[spell.id];
                    final requiresTarget = _targetedHealAmount(spell) > 0;
                    final fallbackTarget = targets.isNotEmpty
                        ? targets.first.id
                        : null;
                    final isSelectedValid =
                        selectedTarget != null &&
                        targets.any((member) => member.id == selectedTarget);
                    final resolvedTarget = isSelectedValid
                        ? selectedTarget
                        : fallbackTarget;
                    return _SpellRow(
                      spell: spell,
                      mana: mana,
                      targets: targets,
                      selectedTargetId: resolvedTarget,
                      casting: _castingSpellId == spell.id,
                      displayName: _displayName,
                      targetedHealAmount: _targetedHealAmount(spell),
                      groupHealAmount: _groupHealAmount(spell),
                      onTargetChanged: requiresTarget
                          ? (value) {
                              setState(() {
                                if (value == null || value.isEmpty) {
                                  _selectedTargetBySpell.remove(spell.id);
                                } else {
                                  _selectedTargetBySpell[spell.id] = value;
                                }
                              });
                            }
                          : null,
                      onCast: () => _castSpell(
                        spell,
                        targetUserId: requiresTarget ? resolvedTarget : null,
                      ),
                    );
                  }).toList(),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SpellRow extends StatelessWidget {
  const _SpellRow({
    required this.spell,
    required this.mana,
    required this.targets,
    required this.selectedTargetId,
    required this.casting,
    required this.displayName,
    required this.targetedHealAmount,
    required this.groupHealAmount,
    required this.onCast,
    this.onTargetChanged,
  });

  final Spell spell;
  final int mana;
  final List<User> targets;
  final String? selectedTargetId;
  final bool casting;
  final String Function(User user) displayName;
  final int targetedHealAmount;
  final int groupHealAmount;
  final VoidCallback onCast;
  final ValueChanged<String?>? onTargetChanged;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasTargetedHeal = targetedHealAmount > 0;
    final hasGroupHeal = groupHealAmount > 0;
    final isHealingSpell = hasTargetedHeal || hasGroupHeal;
    final hasEnoughMana = mana >= spell.manaCost;
    final hasValidTarget =
        !hasTargetedHeal || (selectedTargetId ?? '').isNotEmpty;
    final canCast =
        isHealingSpell && hasEnoughMana && hasValidTarget && !casting;

    final castLabel = casting
        ? 'Casting...'
        : hasTargetedHeal && !hasGroupHeal
        ? 'Cast Targeted Heal'
        : !hasTargetedHeal && hasGroupHeal
        ? 'Cast Group Heal'
        : 'Cast Heal';

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
          if (spell.iconUrl.trim().isNotEmpty)
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: Image.network(
                spell.iconUrl,
                width: 32,
                height: 32,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) =>
                    const Icon(Icons.auto_fix_high, size: 20),
              ),
            )
          else
            const Icon(Icons.auto_fix_high, size: 20),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  spell.name,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                if (spell.schoolOfMagic.trim().isNotEmpty)
                  Text(
                    '${spell.schoolOfMagic} · Mana ${spell.manaCost}',
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                if (spell.effectText.trim().isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    spell.effectText.trim(),
                    style: theme.textTheme.bodySmall,
                  ),
                ],
                if (!isHealingSpell) ...[
                  const SizedBox(height: 8),
                  Text(
                    'Casting is only available for healing spells here.',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
                if (isHealingSpell) ...[
                  const SizedBox(height: 8),
                  Opacity(
                    opacity: hasEnoughMana ? 1.0 : 0.55,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        if (hasTargetedHeal)
                          DropdownButtonFormField<String>(
                            value: selectedTargetId,
                            isDense: true,
                            decoration: const InputDecoration(
                              labelText: 'Heal target',
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
                        if (hasTargetedHeal) const SizedBox(height: 8),
                        FilledButton(
                          onPressed: canCast ? onCast : null,
                          child: Text(castLabel),
                        ),
                      ],
                    ),
                  ),
                  if (!hasEnoughMana) ...[
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
