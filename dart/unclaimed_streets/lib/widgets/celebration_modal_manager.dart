import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/character_stats_provider.dart';
import '../providers/completed_task_provider.dart';
import '../screens/layout_shell.dart';
import '../services/poi_service.dart';

class CelebrationModalManager extends StatelessWidget {
  const CelebrationModalManager({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<CompletedTaskProvider>(
      builder: (context, provider, _) {
        final modal = provider.currentModal;
        if (modal == null) return const SizedBox.shrink();

        final type = modal['type'] as String?;
        final data = modal['data'] as Map<String, dynamic>? ?? {};
        final unspentPoints = context
            .watch<CharacterStatsProvider>()
            .unspentPoints;
        final drawerController = LayoutShellDrawerController.maybeReadOf(
          context,
        );
        final scenarioSuccess = type == 'scenarioOutcome'
            ? data['successful'] == true
            : true;
        final challengeSuccess = type == 'challengeOutcome'
            ? data['successful'] == true
            : true;
        final expositionSuccess = type == 'expositionOutcome'
            ? data['successful'] == true
            : true;
        final showFailureColor =
            (type == 'scenarioOutcome' && !scenarioSuccess) ||
            (type == 'challengeOutcome' && !challengeSuccess) ||
            (type == 'expositionOutcome' && !expositionSuccess) ||
            type == 'monsterBattleDefeat';
        final titleColor = showFailureColor
            ? Colors.red.shade400
            : Colors.amber.shade700;

        return TweenAnimationBuilder<double>(
          tween: Tween<double>(begin: 0, end: 1),
          duration: const Duration(milliseconds: 280),
          curve: Curves.easeOutCubic,
          builder: (context, value, child) {
            return Opacity(
              opacity: value,
              child: Transform.translate(
                offset: Offset(0, 20 * (1 - value)),
                child: Transform.scale(
                  scale: 0.96 + (0.04 * value),
                  child: child,
                ),
              ),
            );
          },
          child: Dialog(
            child: ConstrainedBox(
              constraints: BoxConstraints(
                maxWidth: 560,
                maxHeight: MediaQuery.of(context).size.height * 0.82,
              ),
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      _titleFor(type, data),
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        color: titleColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 16),
                    _contentFor(
                      type,
                      data,
                      context,
                      unspentPoints: unspentPoints,
                    ),
                    const SizedBox(height: 16),
                    _buildActions(
                      context,
                      provider,
                      type,
                      drawerController: drawerController,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  String _titleFor(String? type, Map<String, dynamic> data) {
    switch (type) {
      case 'challenge':
        return 'Victory!';
      case 'levelUp':
        return 'Level Up!';
      case 'reputationUp':
        return 'Reputation Up!';
      case 'questCompleted':
        return 'Quest Complete!';
      case 'treasureChestOpened':
        return 'Treasure Found!';
      case 'healingFountainUsed':
        return 'Fully Restored!';
      case 'scenarioOutcome':
        final successful = data['successful'] == true;
        return successful ? 'Success!' : 'Failed!';
      case 'challengeOutcome':
        final successful = data['successful'] == true;
        return successful ? 'Challenge Success!' : 'Challenge Failed';
      case 'expositionOutcome':
        return 'Conversation Complete';
      case 'monsterBattleVictory':
        return 'Victory!';
      case 'monsterBattleDefeat':
        return 'Defeat!';
      default:
        return 'Congratulations!';
    }
  }

  Widget _contentFor(
    String? type,
    Map<String, dynamic> data,
    BuildContext context, {
    required int unspentPoints,
  }) {
    switch (type) {
      case 'questCompleted':
        final questName = data['questName'] as String? ?? 'Quest';
        final goldAwarded = (data['goldAwarded'] as num?)?.toInt() ?? 0;
        final baseResourcesAwarded = _baseResourcesFromData(data);
        final itemsAwarded =
            (data['itemsAwarded'] as List<dynamic>?)
                ?.whereType<Map>()
                .map((e) => Map<String, dynamic>.from(e))
                .toList() ??
            const [];
        final spellsAwarded =
            (data['spellsAwarded'] as List<dynamic>?)
                ?.whereType<Map>()
                .map((e) => Map<String, dynamic>.from(e))
                .toList() ??
            const [];
        final rewards = <Widget>[
          if (questName.isNotEmpty)
            Text(questName, style: Theme.of(context).textTheme.titleMedium),
        ];
        if (goldAwarded > 0 ||
            baseResourcesAwarded.isNotEmpty ||
            itemsAwarded.isNotEmpty ||
            spellsAwarded.isNotEmpty) {
          if (rewards.isNotEmpty) {
            rewards.add(const SizedBox(height: 12));
          }
          rewards.add(
            _buildRewardSection(
              context,
              gold: goldAwarded,
              materials: baseResourcesAwarded,
              items: itemsAwarded,
              spells: spellsAwarded,
            ),
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: rewards,
        );
      case 'treasureChestOpened':
        final rewardExperience =
            (data['rewardExperience'] as num?)?.toInt() ?? 0;
        final goldAwarded =
            (data['rewardGold'] as num?)?.toInt() ??
            (data['goldAwarded'] as num?)?.toInt() ??
            0;
        final baseResourcesAwarded = _baseResourcesFromData(data);
        final itemsAwarded =
            (data['itemsAwarded'] as List<dynamic>?)
                ?.whereType<Map>()
                .map((e) => Map<String, dynamic>.from(e))
                .toList() ??
            const [];
        final rewards = <Widget>[
          Text(
            'You opened a chest and received:',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 10),
          _buildRewardSection(
            context,
            experience: rewardExperience,
            gold: goldAwarded,
            materials: baseResourcesAwarded,
            items: itemsAwarded,
            emptyMessage: 'No loot this time.',
          ),
        ];
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: rewards,
        );
      case 'healingFountainUsed':
        final healthRestored = (data['healthRestored'] as num?)?.toInt() ?? 0;
        final manaRestored = (data['manaRestored'] as num?)?.toInt() ?? 0;
        final nextAvailableAt = data['nextAvailableAt']?.toString() ?? '';
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('Health restored: $healthRestored'),
            Text('Mana restored: $manaRestored'),
            if (nextAvailableAt.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                'Next fountain use: $nextAvailableAt',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ],
        );
      case 'levelUp':
        final newLevel = (data['newLevel'] as num?)?.toInt();
        final previousLevel = (data['previousLevel'] as num?)?.toInt();
        final levelsGained = (data['levelsGained'] as num?)?.toInt() ?? 1;
        final theme = Theme.of(context);
        final pointsLabel = unspentPoints == 1
            ? '1 stat point'
            : '$unspentPoints stat points';
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: [Color(0xFFFFF2C2), Color(0xFFF4D57A)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(22),
                border: Border.all(color: const Color(0xFFB8841F), width: 1.4),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x332D2416),
                    blurRadius: 16,
                    offset: Offset(0, 8),
                  ),
                ],
              ),
              child: Column(
                children: [
                  Container(
                    width: 84,
                    height: 84,
                    decoration: BoxDecoration(
                      color: const Color(0xFF7A1823),
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: const Color(0xFFFFF1C9),
                        width: 3,
                      ),
                    ),
                    child: Center(
                      child: Text(
                        newLevel?.toString() ?? '!',
                        style: theme.textTheme.headlineSmall?.copyWith(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 14),
                  Text(
                    previousLevel != null && newLevel != null
                        ? 'Level $previousLevel -> Level $newLevel'
                        : (newLevel != null
                              ? 'You reached level $newLevel'
                              : 'You leveled up'),
                    textAlign: TextAlign.center,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF3A2418),
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    levelsGained > 1
                        ? 'You gained $levelsGained levels at once.'
                        : 'Your character just got stronger.',
                    textAlign: TextAlign.center,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: const Color(0xFF5A412C),
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
            if (unspentPoints > 0) ...[
              const SizedBox(height: 14),
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF8E2),
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: const Color(0xFFE2BF63)),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 38,
                      height: 38,
                      decoration: const BoxDecoration(
                        color: Color(0xFF7A1823),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.north_rounded,
                        color: Colors.white,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            '$pointsLabel ready to spend',
                            style: theme.textTheme.titleSmall?.copyWith(
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF3A2418),
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Open Character to allocate your new points and finish leveling up.',
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: const Color(0xFF5A412C),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        );
      case 'reputationUp':
        final level = data['newLevel'] ?? data['newReputationLevel'];
        final zoneName = data['zoneName'];
        return Text('You reached level $level in $zoneName!');
      case 'challenge':
        final experienceAwarded = (data['experienceAwarded'] as num?)?.toInt();
        final reputationAwarded = (data['reputationAwarded'] as num?)?.toInt();
        final goldAwarded = (data['goldAwarded'] as num?)?.toInt();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            if (experienceAwarded != null)
              _buildRewardRow(
                context,
                label: '+$experienceAwarded XP',
                icon: Icons.auto_awesome,
                iconColor: Colors.blue.shade700,
                backgroundColor: Colors.blue.shade50,
              ),
            if (reputationAwarded != null) ...[
              if (experienceAwarded != null) const SizedBox(height: 8),
              Text('+$reputationAwarded Reputation'),
            ],
            if (goldAwarded != null) ...[
              if (experienceAwarded != null || reputationAwarded != null)
                const SizedBox(height: 8),
              _buildRewardRow(
                context,
                label: '+$goldAwarded Gold',
                icon: Icons.monetization_on,
                iconColor: Colors.amber.shade800,
                backgroundColor: Colors.amber.shade100,
              ),
            ],
          ],
        );
      case 'challengeOutcome':
        return _buildChallengeOutcomeContent(context, data);
      case 'expositionOutcome':
        return _buildExpositionOutcomeContent(context, data);
      case 'scenarioOutcome':
        return _buildScenarioOutcomeContent(context, data);
      case 'monsterBattleVictory':
        final monsterName =
            (data['monsterName'] as String?)?.trim() ?? 'Monster';
        final rewardExperience =
            (data['rewardExperience'] as num?)?.toInt() ?? 0;
        final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
        final baseResourcesAwarded = _baseResourcesFromData(data);
        final itemsAwarded =
            (data['itemsAwarded'] as List<dynamic>?)
                ?.whereType<Map>()
                .map((entry) => Map<String, dynamic>.from(entry))
                .toList() ??
            const [];
        final rewards = <Widget>[
          Text('You defeated $monsterName!'),
          const SizedBox(height: 10),
          _buildRewardSection(
            context,
            experience: rewardExperience,
            gold: rewardGold,
            materials: baseResourcesAwarded,
            items: itemsAwarded,
            emptyMessage: 'No loot this time.',
          ),
        ];
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: rewards,
        );
      case 'monsterBattleDefeat':
        final monsterName =
            (data['monsterName'] as String?)?.trim() ?? 'The monster';
        final healthSetTo = (data['healthSetTo'] as num?)?.toInt() ?? 1;
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('$monsterName defeated you.'),
            const SizedBox(height: 8),
            Text('Your life has been set to $healthSetTo.'),
          ],
        );
      default:
        return const Text('Task completed!');
    }
  }

  Widget _buildScenarioOutcomeContent(
    BuildContext context,
    Map<String, dynamic> data,
  ) {
    final theme = Theme.of(context);
    final successful = data['successful'] == true;
    final scenarioId = data['scenarioId']?.toString().trim() ?? '';
    final outcomeText = (data['outcomeText'] as String?)?.trim() ?? '';
    final roll = (data['roll'] as num?)?.toInt() ?? 0;
    final statTag = (data['statTag'] as String?)?.trim() ?? '';
    final statValue = (data['statValue'] as num?)?.toInt() ?? 0;
    final proficiencies =
        (data['proficiencies'] as List<dynamic>?)
            ?.map((value) => value.toString().trim())
            .where((value) => value.isNotEmpty)
            .toList() ??
        const <String>[];
    final proficiencyBonus = (data['proficiencyBonus'] as num?)?.toInt() ?? 0;
    final responseScore = (data['responseScore'] as num?)?.toInt() ?? 0;
    final creativityBonus = (data['creativityBonus'] as num?)?.toInt() ?? 0;
    final totalScore = (data['totalScore'] as num?)?.toInt() ?? 0;
    final threshold = (data['threshold'] as num?)?.toInt() ?? 0;
    final rewardExperience = (data['rewardExperience'] as num?)?.toInt() ?? 0;
    final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
    final baseResourcesAwarded = _baseResourcesFromData(data);
    final failureHealthDrained =
        (data['failureHealthDrained'] as num?)?.toInt() ?? 0;
    final failureManaDrained =
        (data['failureManaDrained'] as num?)?.toInt() ?? 0;
    final failureStatusesApplied =
        (data['failureStatusesApplied'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList() ??
        const [];
    final successHealthRestored =
        (data['successHealthRestored'] as num?)?.toInt() ?? 0;
    final successManaRestored =
        (data['successManaRestored'] as num?)?.toInt() ?? 0;
    final successStatusesApplied =
        (data['successStatusesApplied'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList() ??
        const [];
    final itemsAwarded =
        (data['itemsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList() ??
        const [];
    final spellsAwarded =
        (data['spellsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList() ??
        const [];
    final itemChoiceRewards =
        (data['itemChoiceRewards'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList() ??
        const [];

    final statLabel = _formatStatLabel(statTag);
    final progressValue = threshold <= 0
        ? 1.0
        : (totalScore / math.max(1, threshold)).clamp(0.0, 1.0).toDouble();
    final accentColor = successful
        ? Colors.green.shade700
        : Colors.red.shade700;
    final accentTint = successful
        ? Colors.green.withValues(alpha: 0.12)
        : Colors.red.withValues(alpha: 0.12);
    final segments = <_ScenarioScoreSegment>[
      _ScenarioScoreSegment(
        label: 'Roll',
        caption: 'The d20 decides the swing.',
        icon: Icons.casino_rounded,
        value: roll,
        color: Colors.deepPurple.shade400,
      ),
      _ScenarioScoreSegment(
        label: statLabel,
        caption: 'Your core stat carries the plan.',
        icon: Icons.fitness_center_rounded,
        value: statValue,
        color: theme.colorScheme.primary,
      ),
      _ScenarioScoreSegment(
        label: 'Proficiency',
        caption: proficiencies.isEmpty
            ? 'Training applied to this attempt.'
            : 'Boosted by ${proficiencies.join(', ')}.',
        icon: Icons.workspace_premium_rounded,
        value: proficiencyBonus,
        color: Colors.teal.shade600,
      ),
      if (responseScore > 0)
        _ScenarioScoreSegment(
          label: 'Execution',
          caption: 'How well you solved the problem.',
          icon: Icons.psychology_alt_rounded,
          value: responseScore,
          color: Colors.indigo.shade400,
        ),
      _ScenarioScoreSegment(
        label: 'Creativity',
        caption: 'Bonus for a clever angle.',
        icon: Icons.lightbulb_rounded,
        value: creativityBonus,
        color: Colors.amber.shade700,
      ),
    ];

    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: 1),
      duration: const Duration(milliseconds: 420),
      curve: Curves.easeOutCubic,
      builder: (context, value, child) {
        return Opacity(
          opacity: value,
          child: Transform.translate(
            offset: Offset(0, 16 * (1 - value)),
            child: child,
          ),
        );
      },
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: successful
                    ? [Colors.green.shade50, Colors.teal.shade50]
                    : [Colors.red.shade50, Colors.orange.shade50],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(18),
              border: Border.all(color: accentColor.withValues(alpha: 0.22)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        color: accentTint,
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        successful
                            ? Icons.emoji_events_rounded
                            : Icons.shield_outlined,
                        color: accentColor,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            outcomeText.isNotEmpty
                                ? outcomeText
                                : (successful
                                      ? 'Your approach succeeds.'
                                      : 'Your approach falls short.'),
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                              color: accentColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerLow,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: theme.colorScheme.outline.withValues(alpha: 0.14),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      'Total',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const Spacer(),
                    Text(
                      'Scored $totalScore / Needed $threshold',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w800,
                        color: accentColor,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                TweenAnimationBuilder<double>(
                  tween: Tween<double>(begin: 0, end: progressValue),
                  duration: const Duration(milliseconds: 650),
                  curve: Curves.easeOutCubic,
                  builder: (context, value, _) {
                    return ClipRRect(
                      borderRadius: BorderRadius.circular(999),
                      child: LinearProgressIndicator(
                        minHeight: 10,
                        value: value,
                        backgroundColor: theme.colorScheme.surfaceContainerHigh,
                        valueColor: AlwaysStoppedAnimation<Color>(accentColor),
                      ),
                    );
                  },
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          Container(
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerLow,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: theme.colorScheme.outline.withValues(alpha: 0.14),
              ),
            ),
            child: Theme(
              data: theme.copyWith(dividerColor: Colors.transparent),
              child: ExpansionTile(
                tilePadding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 2,
                ),
                childrenPadding: const EdgeInsets.fromLTRB(14, 0, 14, 14),
                iconColor: theme.colorScheme.onSurfaceVariant,
                collapsedIconColor: theme.colorScheme.onSurfaceVariant,
                title: Text(
                  'Score Breakdown',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                subtitle: Text(
                  'See how the final total was built.',
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                children: [
                  LayoutBuilder(
                    builder: (context, constraints) {
                      final cardWidth = constraints.maxWidth >= 520
                          ? (constraints.maxWidth - 10) / 2
                          : constraints.maxWidth;
                      return Wrap(
                        spacing: 10,
                        runSpacing: 10,
                        children: [
                          for (var i = 0; i < segments.length; i++)
                            _buildScenarioScoreSegmentCard(
                              context,
                              segments[i],
                              index: i,
                              width: cardWidth,
                            ),
                        ],
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
          if (successful &&
              (successHealthRestored > 0 ||
                  successManaRestored > 0 ||
                  successStatusesApplied.isNotEmpty)) ...[
            const SizedBox(height: 14),
            _buildScenarioImpactSection(
              context,
              title: 'Success Effects',
              color: Colors.green.shade700,
              backgroundColor: Colors.green.withValues(alpha: 0.08),
              borderColor: Colors.green.withValues(alpha: 0.22),
              healthDelta: successHealthRestored,
              manaDelta: successManaRestored,
              statuses: successStatusesApplied,
            ),
          ],
          if (!successful &&
              (failureHealthDrained > 0 ||
                  failureManaDrained > 0 ||
                  failureStatusesApplied.isNotEmpty)) ...[
            const SizedBox(height: 14),
            _buildScenarioImpactSection(
              context,
              title: 'Failure Penalties',
              color: Colors.red.shade700,
              backgroundColor: Colors.red.withValues(alpha: 0.08),
              borderColor: Colors.red.withValues(alpha: 0.22),
              healthDelta: -failureHealthDrained,
              manaDelta: -failureManaDrained,
              statuses: failureStatusesApplied,
              statusColor: Colors.orange.shade700,
            ),
          ],
          if (successful && itemChoiceRewards.isNotEmpty) ...[
            const SizedBox(height: 14),
            if (scenarioId.isEmpty)
              const Text(
                'Item choice reward is available, but scenario context is missing.',
              )
            else
              _ItemChoiceRewardPicker(
                choices: itemChoiceRewards,
                onChoose: (inventoryItemId) {
                  return context.read<PoiService>().chooseScenarioRewardItem(
                    scenarioId,
                    inventoryItemId: inventoryItemId,
                  );
                },
              ),
          ],
          if (rewardExperience > 0 ||
              rewardGold > 0 ||
              baseResourcesAwarded.isNotEmpty ||
              itemsAwarded.isNotEmpty ||
              spellsAwarded.isNotEmpty) ...[
            const SizedBox(height: 14),
            Text(
              'Rewards',
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            _buildRewardSection(
              context,
              experience: rewardExperience,
              gold: rewardGold,
              materials: baseResourcesAwarded,
              items: itemsAwarded,
              spells: spellsAwarded,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildActions(
    BuildContext context,
    CompletedTaskProvider provider,
    String? type, {
    LayoutShellDrawerController? drawerController,
  }) {
    if (type == 'levelUp' && drawerController != null) {
      return Row(
        children: [
          Expanded(
            child: OutlinedButton(
              onPressed: provider.clearModal,
              child: const Text('Later'),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: FilledButton.icon(
              onPressed: () {
                provider.clearModal();
                drawerController.openCharacter();
              },
              icon: const Icon(Icons.person),
              label: const Text('Open Character'),
            ),
          ),
        ],
      );
    }
    return FilledButton(
      onPressed: provider.clearModal,
      child: const Text('OK'),
    );
  }

  Widget _buildExpositionOutcomeContent(
    BuildContext context,
    Map<String, dynamic> data,
  ) {
    final title = (data['title'] as String?)?.trim() ?? 'Exposition';
    final questName = (data['questName'] as String?)?.trim() ?? '';
    final rewardExperience = (data['rewardExperience'] as num?)?.toInt() ?? 0;
    final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
    final baseResourcesAwarded = _baseResourcesFromData(data);
    final itemsAwarded =
        (data['itemsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((entry) => Map<String, dynamic>.from(entry))
            .toList() ??
        const [];
    final spellsAwarded =
        (data['spellsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((spell) => Map<String, dynamic>.from(spell))
            .toList() ??
        const [];
    final awardedRewards = data['awardedRewards'] == true;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(title, style: Theme.of(context).textTheme.titleMedium),
        if (questName.isNotEmpty) ...[
          const SizedBox(height: 6),
          Text(
            'Quest objective advanced for $questName.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ],
        const SizedBox(height: 12),
        _buildRewardSection(
          context,
          experience: rewardExperience,
          gold: rewardGold,
          materials: baseResourcesAwarded,
          items: itemsAwarded,
          spells: spellsAwarded,
          emptyMessage: awardedRewards
              ? 'No additional rewards.'
              : 'Rewards were already claimed.',
        ),
      ],
    );
  }

  Widget _buildChallengeOutcomeContent(
    BuildContext context,
    Map<String, dynamic> data,
  ) {
    final theme = Theme.of(context);
    final successful = data['successful'] == true;
    final challengeId = data['challengeId']?.toString().trim() ?? '';
    final reason = (data['reason'] as String?)?.trim() ?? '';
    final score = (data['score'] as num?)?.toInt() ?? 0;
    final difficulty = (data['difficulty'] as num?)?.toInt() ?? 0;
    final combinedScore = (data['combinedScore'] as num?)?.toInt() ?? 0;
    final rewardExperience = (data['rewardExperience'] as num?)?.toInt() ?? 0;
    final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
    final baseResourcesAwarded = _baseResourcesFromData(data);
    final statTags =
        (data['statTags'] as List<dynamic>?)
            ?.map((value) => value.toString().trim())
            .where((value) => value.isNotEmpty)
            .toList() ??
        const <String>[];
    final statValues =
        (data['statValues'] as Map?)?.map(
          (key, value) =>
              MapEntry(key.toString().trim(), (value as num?)?.toInt() ?? 0),
        ) ??
        const <String, int>{};
    final itemsAwarded =
        (data['itemsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((item) => Map<String, dynamic>.from(item))
            .toList() ??
        const [];
    final spellsAwarded =
        (data['spellsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((item) => Map<String, dynamic>.from(item))
            .toList() ??
        const [];
    final itemChoiceRewards =
        (data['itemChoiceRewards'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((item) => Map<String, dynamic>.from(item))
            .toList() ??
        const [];

    final progressValue = difficulty <= 0
        ? 1.0
        : (combinedScore / math.max(1, difficulty)).clamp(0.0, 1.0).toDouble();
    final accentColor = successful
        ? Colors.green.shade700
        : Colors.red.shade700;
    final accentTint = successful
        ? Colors.green.withValues(alpha: 0.12)
        : Colors.red.withValues(alpha: 0.12);
    final segments = <_ScenarioScoreSegment>[
      _ScenarioScoreSegment(
        label: 'Submission',
        caption: 'Your answer before modifiers.',
        icon: Icons.edit_note_rounded,
        value: score,
        color: Colors.deepPurple.shade400,
      ),
      for (final tag in statTags)
        _ScenarioScoreSegment(
          label: _formatStatLabel(tag),
          caption: 'Stat bonus applied to the challenge.',
          icon: Icons.bolt_rounded,
          value: statValues[tag] ?? 0,
          color: theme.colorScheme.primary,
        ),
    ];

    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: 1),
      duration: const Duration(milliseconds: 420),
      curve: Curves.easeOutCubic,
      builder: (context, value, child) {
        return Opacity(
          opacity: value,
          child: Transform.translate(
            offset: Offset(0, 16 * (1 - value)),
            child: child,
          ),
        );
      },
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: successful
                    ? [Colors.green.shade50, Colors.teal.shade50]
                    : [Colors.red.shade50, Colors.orange.shade50],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(18),
              border: Border.all(color: accentColor.withValues(alpha: 0.22)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: accentTint,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    successful
                        ? Icons.emoji_events_rounded
                        : Icons.shield_outlined,
                    color: accentColor,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        reason.isNotEmpty
                            ? reason
                            : (successful
                                  ? 'Your party passed the challenge.'
                                  : 'Your party did not meet the challenge threshold.'),
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w700,
                          color: accentColor,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerLow,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: theme.colorScheme.outline.withValues(alpha: 0.14),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      'Total',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const Spacer(),
                    Text(
                      'Scored $combinedScore / Needed $difficulty',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w800,
                        color: accentColor,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                TweenAnimationBuilder<double>(
                  tween: Tween<double>(begin: 0, end: progressValue),
                  duration: const Duration(milliseconds: 650),
                  curve: Curves.easeOutCubic,
                  builder: (context, value, _) {
                    return ClipRRect(
                      borderRadius: BorderRadius.circular(999),
                      child: LinearProgressIndicator(
                        minHeight: 10,
                        value: value,
                        backgroundColor: theme.colorScheme.surfaceContainerHigh,
                        valueColor: AlwaysStoppedAnimation<Color>(accentColor),
                      ),
                    );
                  },
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          Container(
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerLow,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: theme.colorScheme.outline.withValues(alpha: 0.14),
              ),
            ),
            child: Theme(
              data: theme.copyWith(dividerColor: Colors.transparent),
              child: ExpansionTile(
                tilePadding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 2,
                ),
                childrenPadding: const EdgeInsets.fromLTRB(14, 0, 14, 14),
                iconColor: theme.colorScheme.onSurfaceVariant,
                collapsedIconColor: theme.colorScheme.onSurfaceVariant,
                title: Text(
                  'Score Breakdown',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                subtitle: Text(
                  statTags.isEmpty
                      ? 'This challenge used your base submission score.'
                      : 'See how the final total was built.',
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                children: [
                  LayoutBuilder(
                    builder: (context, constraints) {
                      final cardWidth = constraints.maxWidth >= 520
                          ? (constraints.maxWidth - 10) / 2
                          : constraints.maxWidth;
                      return Wrap(
                        spacing: 10,
                        runSpacing: 10,
                        children: [
                          for (var i = 0; i < segments.length; i++)
                            _buildScenarioScoreSegmentCard(
                              context,
                              segments[i],
                              index: i,
                              width: cardWidth,
                            ),
                        ],
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
          if (successful && itemChoiceRewards.isNotEmpty) ...[
            const SizedBox(height: 14),
            if (challengeId.isEmpty)
              const Text(
                'Item choice reward is available, but challenge context is missing.',
              )
            else
              _ItemChoiceRewardPicker(
                choices: itemChoiceRewards,
                onChoose: (inventoryItemId) {
                  return context.read<PoiService>().chooseChallengeRewardItem(
                    challengeId,
                    inventoryItemId: inventoryItemId,
                  );
                },
              ),
          ],
          if (rewardExperience > 0 ||
              rewardGold > 0 ||
              baseResourcesAwarded.isNotEmpty ||
              itemsAwarded.isNotEmpty ||
              spellsAwarded.isNotEmpty) ...[
            const SizedBox(height: 14),
            Text(
              'Rewards',
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            _buildRewardSection(
              context,
              experience: rewardExperience,
              gold: rewardGold,
              materials: baseResourcesAwarded,
              items: itemsAwarded,
              spells: spellsAwarded,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildScenarioScoreSegmentCard(
    BuildContext context,
    _ScenarioScoreSegment segment, {
    required int index,
    required double width,
  }) {
    final theme = Theme.of(context);
    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: 1),
      duration: Duration(milliseconds: 280 + (index * 70)),
      curve: Curves.easeOutCubic,
      builder: (context, value, child) {
        return Opacity(
          opacity: value,
          child: Transform.translate(
            offset: Offset(0, 10 * (1 - value)),
            child: child,
          ),
        );
      },
      child: Container(
        width: width,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: theme.colorScheme.surfaceContainerLow,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: segment.color.withValues(alpha: 0.16)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 42,
              height: 42,
              decoration: BoxDecoration(
                color: segment.color.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(segment.icon, color: segment.color),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          segment.label,
                          style: theme.textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        const SizedBox(height: 3),
                        Text(
                          segment.caption,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    _formatSignedValue(segment.value),
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w800,
                      color: segment.color,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildScenarioImpactSection(
    BuildContext context, {
    required String title,
    required Color color,
    required Color backgroundColor,
    required Color borderColor,
    required int healthDelta,
    required int manaDelta,
    required List<Map<String, dynamic>> statuses,
    Color? statusColor,
  }) {
    final theme = Theme.of(context);
    final resolvedStatusColor = statusColor ?? color;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: borderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
              color: color,
            ),
          ),
          if (healthDelta != 0) ...[
            const SizedBox(height: 6),
            Row(
              children: [
                Icon(Icons.favorite, size: 16, color: Colors.red.shade600),
                const SizedBox(width: 6),
                Text('${_formatSignedValue(healthDelta)} Health'),
              ],
            ),
          ],
          if (manaDelta != 0) ...[
            const SizedBox(height: 4),
            Row(
              children: [
                Icon(
                  Icons.auto_fix_high,
                  size: 16,
                  color: Colors.blue.shade600,
                ),
                const SizedBox(width: 6),
                Text('${_formatSignedValue(manaDelta)} Mana'),
              ],
            ),
          ],
          for (final status in statuses) ...[
            const SizedBox(height: 6),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(
                  Icons.hourglass_bottom,
                  size: 16,
                  color: resolvedStatusColor,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        status['name']?.toString().trim().isNotEmpty == true
                            ? status['name'].toString().trim()
                            : 'Status Applied',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      if ((status['durationSeconds'] as num?)?.toInt()
                          case final duration? when duration != 0)
                        Text('$duration s', style: theme.textTheme.bodySmall),
                      if ((status['description'] as String?)
                              ?.trim()
                              .isNotEmpty ==
                          true)
                        Text(
                          (status['description'] as String).trim(),
                          style: theme.textTheme.bodySmall,
                        ),
                      if ((status['effect'] as String?)?.trim().isNotEmpty ==
                          true)
                        Text(
                          (status['effect'] as String).trim(),
                          style: theme.textTheme.bodySmall,
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  String _formatStatLabel(String statTag) {
    if (statTag.isEmpty) return 'Stat';
    return '${statTag[0].toUpperCase()}${statTag.substring(1)}';
  }

  String _formatSignedValue(int value) {
    return value > 0 ? '+$value' : '$value';
  }

  List<Map<String, dynamic>> _baseResourcesFromData(Map<String, dynamic> data) {
    return (data['baseResourcesAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((entry) => Map<String, dynamic>.from(entry))
            .where(
              (entry) =>
                  (entry['resourceKey']?.toString().trim().isNotEmpty ??
                      false) &&
                  ((entry['amount'] as num?)?.toInt() ?? 0) > 0,
            )
            .toList(growable: false) ??
        const <Map<String, dynamic>>[];
  }

  String _materialLabel(String resourceKey) {
    switch (resourceKey.trim().toLowerCase()) {
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
        if (resourceKey.trim().isEmpty) return 'Material';
        final parts = resourceKey
            .trim()
            .split('_')
            .where((part) => part.isNotEmpty)
            .map(
              (part) =>
                  '${part[0].toUpperCase()}${part.substring(1).toLowerCase()}',
            )
            .toList(growable: false);
        return parts.isEmpty ? 'Material' : parts.join(' ');
    }
  }

  IconData _materialIcon(String resourceKey) {
    switch (resourceKey.trim().toLowerCase()) {
      case 'timber':
        return Icons.park;
      case 'stone':
        return Icons.landscape;
      case 'iron':
        return Icons.hardware;
      case 'herbs':
        return Icons.local_florist;
      case 'monster_parts':
        return Icons.pets;
      case 'arcane_dust':
        return Icons.auto_awesome;
      case 'relic_shards':
        return Icons.diamond;
      default:
        return Icons.inventory_2;
    }
  }

  Color _materialColor(String resourceKey) {
    switch (resourceKey.trim().toLowerCase()) {
      case 'timber':
        return const Color(0xFF8D6E63);
      case 'stone':
        return const Color(0xFF78909C);
      case 'iron':
        return const Color(0xFF546E7A);
      case 'herbs':
        return const Color(0xFF43A047);
      case 'monster_parts':
        return const Color(0xFFC62828);
      case 'arcane_dust':
        return const Color(0xFF6A5ACD);
      case 'relic_shards':
        return const Color(0xFF00897B);
      default:
        return const Color(0xFF616161);
    }
  }

  Widget _buildRewardSection(
    BuildContext context, {
    int experience = 0,
    int gold = 0,
    List<Map<String, dynamic>> materials = const [],
    List<Map<String, dynamic>> items = const [],
    List<Map<String, dynamic>> spells = const [],
    String? emptyMessage,
  }) {
    final children = <Widget>[];

    void addReward(Widget child) {
      if (children.isNotEmpty) {
        children.add(const SizedBox(height: 8));
      }
      children.add(child);
    }

    if (experience > 0) {
      addReward(
        _buildRewardRow(
          context,
          label: '+$experience XP',
          icon: Icons.auto_awesome,
          iconColor: Colors.blue.shade700,
          backgroundColor: Colors.blue.shade50,
        ),
      );
    }
    if (gold > 0) {
      addReward(
        _buildRewardRow(
          context,
          label: '+$gold Gold',
          icon: Icons.monetization_on,
          iconColor: Colors.amber.shade800,
          backgroundColor: Colors.amber.shade100,
        ),
      );
    }
    for (final material in materials) {
      final resourceKey = material['resourceKey']?.toString() ?? '';
      final amount = (material['amount'] as num?)?.toInt() ?? 0;
      if (amount <= 0) continue;
      final accentColor = _materialColor(resourceKey);
      addReward(
        _buildRewardRow(
          context,
          label: '+$amount ${_materialLabel(resourceKey)}',
          icon: _materialIcon(resourceKey),
          iconColor: accentColor,
          backgroundColor: accentColor.withValues(alpha: 0.12),
        ),
      );
    }
    for (final item in items) {
      final name = item['name']?.toString().trim();
      final quantity = (item['quantity'] as num?)?.toInt() ?? 1;
      addReward(
        _buildRewardRow(
          context,
          label:
              '+$quantity ${name != null && name.isNotEmpty ? name : 'Item'}',
          imageUrl: item['imageUrl']?.toString(),
          icon: Icons.inventory_2_outlined,
          iconColor: Theme.of(context).colorScheme.onSurfaceVariant,
          backgroundColor: Theme.of(
            context,
          ).colorScheme.surfaceContainerHighest,
        ),
      );
    }
    for (final spell in spells) {
      final name = spell['name']?.toString().trim();
      addReward(
        _buildRewardRow(
          context,
          label: '+Spell: ${name != null && name.isNotEmpty ? name : 'Spell'}',
          icon: Icons.menu_book_rounded,
          iconColor: Colors.teal.shade700,
          backgroundColor: Colors.teal.shade50,
        ),
      );
    }

    if (children.isEmpty) {
      if (emptyMessage == null || emptyMessage.isEmpty) {
        return const SizedBox.shrink();
      }
      return Text(emptyMessage);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: children,
    );
  }

  Widget _buildRewardRow(
    BuildContext context, {
    required String label,
    String? imageUrl,
    required IconData icon,
    required Color iconColor,
    required Color backgroundColor,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        _buildRewardThumbnail(
          context,
          imageUrl: imageUrl,
          icon: icon,
          iconColor: iconColor,
          backgroundColor: backgroundColor,
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            label,
            style: Theme.of(
              context,
            ).textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w600),
          ),
        ),
      ],
    );
  }

  Widget _buildRewardThumbnail(
    BuildContext context, {
    String? imageUrl,
    required IconData icon,
    required Color iconColor,
    required Color backgroundColor,
  }) {
    final theme = Theme.of(context);

    return Container(
      width: 40,
      height: 40,
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: theme.dividerColor),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(10),
        child: imageUrl != null && imageUrl.trim().isNotEmpty
            ? Image.network(
                imageUrl.trim(),
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) =>
                    Icon(icon, size: 20, color: iconColor),
              )
            : Icon(icon, size: 20, color: iconColor),
      ),
    );
  }
}

class _ItemChoiceRewardPicker extends StatefulWidget {
  const _ItemChoiceRewardPicker({
    required this.choices,
    required this.onChoose,
  });

  final List<Map<String, dynamic>> choices;
  final Future<Map<String, dynamic>> Function(int inventoryItemId) onChoose;

  @override
  State<_ItemChoiceRewardPicker> createState() =>
      _ItemChoiceRewardPickerState();
}

class _ItemChoiceRewardPickerState extends State<_ItemChoiceRewardPicker> {
  bool _claiming = false;
  int? _claimingItemId;
  Map<String, dynamic>? _claimedItem;
  String? _error;

  Future<void> _claim(Map<String, dynamic> choice) async {
    if (_claiming || _claimedItem != null) return;
    final inventoryItemId = (choice['id'] as num?)?.toInt() ?? 0;
    if (inventoryItemId <= 0) {
      setState(() {
        _error = 'Invalid item choice.';
      });
      return;
    }

    setState(() {
      _claiming = true;
      _claimingItemId = inventoryItemId;
      _error = null;
    });
    try {
      final response = await widget.onChoose(inventoryItemId);
      final rawAwarded = response['itemAwarded'];
      if (rawAwarded is Map) {
        _claimedItem = Map<String, dynamic>.from(rawAwarded);
      } else {
        _claimedItem = <String, dynamic>{
          'id': inventoryItemId,
          'name': choice['name']?.toString() ?? 'Item',
          'imageUrl': choice['imageUrl']?.toString() ?? '',
          'quantity': (choice['quantity'] as num?)?.toInt() ?? 1,
        };
      }
    } catch (error) {
      _error = error.toString();
    } finally {
      if (mounted) {
        setState(() {
          _claiming = false;
          _claimingItemId = null;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final claimed = _claimedItem;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerLow,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            claimed == null ? 'Choose 1 item reward' : 'Item reward chosen',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          if (claimed != null) ...[
            Text(
              '+${(claimed['quantity'] as num?)?.toInt() ?? 1} ${claimed['name']?.toString() ?? 'Item'}',
            ),
          ] else
            ...widget.choices.map((choice) {
              final id = (choice['id'] as num?)?.toInt() ?? 0;
              final quantity = (choice['quantity'] as num?)?.toInt() ?? 1;
              final name = choice['name']?.toString().trim().isNotEmpty == true
                  ? choice['name'].toString().trim()
                  : 'Item';
              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: SizedBox(
                  width: double.infinity,
                  child: OutlinedButton(
                    onPressed: _claiming ? null : () => _claim(choice),
                    child: Text(
                      _claiming && _claimingItemId == id
                          ? 'Claiming...'
                          : 'Choose +$quantity $name',
                    ),
                  ),
                ),
              );
            }),
          if (_error != null && _error!.trim().isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              _error!,
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.error,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _ScenarioScoreSegment {
  const _ScenarioScoreSegment({
    required this.label,
    required this.caption,
    required this.icon,
    required this.value,
    required this.color,
  });

  final String label;
  final String caption;
  final IconData icon;
  final int value;
  final Color color;
}
