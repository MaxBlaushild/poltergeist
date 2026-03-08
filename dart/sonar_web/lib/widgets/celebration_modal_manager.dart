import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/completed_task_provider.dart';

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
        final scenarioSuccess = type == 'scenarioOutcome'
            ? data['successful'] == true
            : true;
        final challengeSuccess = type == 'challengeOutcome'
            ? data['successful'] == true
            : true;
        final showFailureColor =
            (type == 'scenarioOutcome' && !scenarioSuccess) ||
            (type == 'challengeOutcome' && !challengeSuccess) ||
            type == 'monsterBattleDefeat';
        final titleColor = showFailureColor
            ? Colors.red.shade400
            : Colors.amber.shade700;

        return Dialog(
          child: Padding(
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
                _contentFor(type, data, context),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: () => provider.clearModal(),
                  child: const Text('OK'),
                ),
              ],
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
        return successful ? 'Scenario Success!' : 'Scenario Failed';
      case 'challengeOutcome':
        final successful = data['successful'] == true;
        return successful ? 'Challenge Success!' : 'Challenge Failed';
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
    BuildContext context,
  ) {
    switch (type) {
      case 'questCompleted':
        final questName = data['questName'] as String? ?? 'Quest';
        final goldAwarded = (data['goldAwarded'] as num?)?.toInt() ?? 0;
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
            itemsAwarded.isNotEmpty ||
            spellsAwarded.isNotEmpty) {
          if (rewards.isNotEmpty) {
            rewards.add(const SizedBox(height: 12));
          }
          rewards.add(
            _buildRewardSection(
              context,
              gold: goldAwarded,
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
        final goldAwarded = (data['goldAwarded'] as num?)?.toInt() ?? 0;
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
            gold: goldAwarded,
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
        return Column(
          children: [
            Container(
              width: 64,
              height: 64,
              decoration: BoxDecoration(
                color: Colors.blue.shade100,
                shape: BoxShape.circle,
              ),
              child: const Center(
                child: Text(
                  '+1',
                  style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            const SizedBox(height: 8),
            const Text('You gained a level!'),
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
        final successful = data['successful'] == true;
        final reason = (data['reason'] as String?)?.trim() ?? '';
        final score = (data['score'] as num?)?.toInt() ?? 0;
        final difficulty = (data['difficulty'] as num?)?.toInt() ?? 0;
        final combinedScore = (data['combinedScore'] as num?)?.toInt() ?? 0;
        final rewardExperience =
            (data['rewardExperience'] as num?)?.toInt() ?? 0;
        final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
        final statTags =
            (data['statTags'] as List<dynamic>?)
                ?.map((value) => value.toString().trim())
                .where((value) => value.isNotEmpty)
                .toList() ??
            const <String>[];
        final statValues =
            (data['statValues'] as Map?)?.map(
              (key, value) => MapEntry(
                key.toString().trim(),
                (value as num?)?.toInt() ?? 0,
              ),
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

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              successful
                  ? 'Your party passed the challenge.'
                  : 'Your party did not meet the challenge threshold.',
            ),
            if (reason.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(reason, style: Theme.of(context).textTheme.bodySmall),
            ],
            const SizedBox(height: 10),
            Text('Score: $score'),
            if (statTags.isNotEmpty)
              Text(
                'Modifiers: ${statTags.map((tag) => '+${statValues[tag] ?? 0} $tag').join(' · ')}',
              ),
            Text('Combined: $combinedScore'),
            Text('Target: $difficulty'),
            if (rewardExperience > 0 ||
                rewardGold > 0 ||
                itemsAwarded.isNotEmpty ||
                spellsAwarded.isNotEmpty) ...[
              const SizedBox(height: 10),
              Text(
                'Rewards',
                style: Theme.of(
                  context,
                ).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              _buildRewardSection(
                context,
                experience: rewardExperience,
                gold: rewardGold,
                items: itemsAwarded,
                spells: spellsAwarded,
              ),
            ],
          ],
        );
      case 'scenarioOutcome':
        final successful = data['successful'] == true;
        final outcomeText = (data['outcomeText'] as String?)?.trim() ?? '';
        final reason = (data['reason'] as String?)?.trim() ?? '';
        final roll = (data['roll'] as num?)?.toInt() ?? 0;
        final statTag = (data['statTag'] as String?)?.trim() ?? '';
        final statValue = (data['statValue'] as num?)?.toInt() ?? 0;
        final proficiencyBonus =
            (data['proficiencyBonus'] as num?)?.toInt() ?? 0;
        final creativityBonus = (data['creativityBonus'] as num?)?.toInt() ?? 0;
        final totalScore = (data['totalScore'] as num?)?.toInt() ?? 0;
        final threshold = (data['threshold'] as num?)?.toInt() ?? 0;
        final rewardExperience =
            (data['rewardExperience'] as num?)?.toInt() ?? 0;
        final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
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
        final statLabel = statTag.isEmpty
            ? 'Stat'
            : '${statTag[0].toUpperCase()}${statTag.substring(1)}';

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              outcomeText.isNotEmpty
                  ? outcomeText
                  : (successful
                        ? 'Your approach succeeds.'
                        : 'Your approach falls short.'),
              style: Theme.of(context).textTheme.bodyLarge,
            ),
            if (reason.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(reason, style: Theme.of(context).textTheme.bodySmall),
            ],
            const SizedBox(height: 12),
            Text(
              'Roll Math',
              style: Theme.of(
                context,
              ).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 6),
            Text(
              '$roll (d20) + $statValue ($statLabel) + $proficiencyBonus (Proficiency) + $creativityBonus (Creativity) = $totalScore',
            ),
            Text('Target: $threshold'),
            const SizedBox(height: 10),
            Text(
              successful ? 'Outcome: Success' : 'Outcome: Failure',
              style: Theme.of(
                context,
              ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
            ),
            if (successful &&
                (successHealthRestored > 0 ||
                    successManaRestored > 0 ||
                    successStatusesApplied.isNotEmpty)) ...[
              const SizedBox(height: 10),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.green.withOpacity(0.08),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: Colors.green.withOpacity(0.22)),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Success Effects',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: Colors.green.shade700,
                      ),
                    ),
                    if (successHealthRestored > 0) ...[
                      const SizedBox(height: 6),
                      Row(
                        children: [
                          Icon(
                            Icons.favorite,
                            size: 16,
                            color: Colors.red.shade600,
                          ),
                          const SizedBox(width: 6),
                          Text('+$successHealthRestored Health'),
                        ],
                      ),
                    ],
                    if (successManaRestored > 0) ...[
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          Icon(
                            Icons.auto_fix_high,
                            size: 16,
                            color: Colors.blue.shade600,
                          ),
                          const SizedBox(width: 6),
                          Text('+$successManaRestored Mana'),
                        ],
                      ),
                    ],
                    for (final status in successStatusesApplied) ...[
                      const SizedBox(height: 6),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(
                            Icons.hourglass_bottom,
                            size: 16,
                            color: Colors.green.shade700,
                          ),
                          const SizedBox(width: 6),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  status['name']
                                              ?.toString()
                                              .trim()
                                              .isNotEmpty ==
                                          true
                                      ? status['name'].toString().trim()
                                      : 'Status Applied',
                                  style: Theme.of(context).textTheme.bodyMedium
                                      ?.copyWith(fontWeight: FontWeight.w700),
                                ),
                                if ((status['durationSeconds'] as num?)
                                            ?.toInt() !=
                                        null &&
                                    (status['durationSeconds'] as num?)
                                            ?.toInt() !=
                                        0)
                                  Text(
                                    '${(status['durationSeconds'] as num?)!.toInt()}s',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                                if ((status['description'] as String?)
                                        ?.trim()
                                        .isNotEmpty ==
                                    true)
                                  Text(
                                    (status['description'] as String).trim(),
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                                if ((status['effect'] as String?)
                                        ?.trim()
                                        .isNotEmpty ==
                                    true)
                                  Text(
                                    (status['effect'] as String).trim(),
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
            ],
            if (!successful &&
                (failureHealthDrained > 0 ||
                    failureManaDrained > 0 ||
                    failureStatusesApplied.isNotEmpty)) ...[
              const SizedBox(height: 10),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.withOpacity(0.08),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: Colors.red.withOpacity(0.22)),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Failure Penalties',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: Colors.red.shade700,
                      ),
                    ),
                    if (failureHealthDrained > 0) ...[
                      const SizedBox(height: 6),
                      Row(
                        children: [
                          Icon(
                            Icons.favorite,
                            size: 16,
                            color: Colors.red.shade600,
                          ),
                          const SizedBox(width: 6),
                          Text('-$failureHealthDrained Health'),
                        ],
                      ),
                    ],
                    if (failureManaDrained > 0) ...[
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          Icon(
                            Icons.auto_fix_high,
                            size: 16,
                            color: Colors.blue.shade600,
                          ),
                          const SizedBox(width: 6),
                          Text('-$failureManaDrained Mana'),
                        ],
                      ),
                    ],
                    for (final status in failureStatusesApplied) ...[
                      const SizedBox(height: 6),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(
                            Icons.hourglass_bottom,
                            size: 16,
                            color: Colors.orange.shade700,
                          ),
                          const SizedBox(width: 6),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  status['name']
                                              ?.toString()
                                              .trim()
                                              .isNotEmpty ==
                                          true
                                      ? status['name'].toString().trim()
                                      : 'Status Applied',
                                  style: Theme.of(context).textTheme.bodyMedium
                                      ?.copyWith(fontWeight: FontWeight.w700),
                                ),
                                if ((status['durationSeconds'] as num?)
                                            ?.toInt() !=
                                        null &&
                                    (status['durationSeconds'] as num?)
                                            ?.toInt() !=
                                        0)
                                  Text(
                                    '${(status['durationSeconds'] as num?)!.toInt()}s',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                                if ((status['description'] as String?)
                                        ?.trim()
                                        .isNotEmpty ==
                                    true)
                                  Text(
                                    (status['description'] as String).trim(),
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                                if ((status['effect'] as String?)
                                        ?.trim()
                                        .isNotEmpty ==
                                    true)
                                  Text(
                                    (status['effect'] as String).trim(),
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
            ],
            if (rewardExperience > 0 ||
                rewardGold > 0 ||
                itemsAwarded.isNotEmpty ||
                spellsAwarded.isNotEmpty) ...[
              const SizedBox(height: 10),
              Text(
                'Rewards',
                style: Theme.of(
                  context,
                ).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              _buildRewardSection(
                context,
                experience: rewardExperience,
                gold: rewardGold,
                items: itemsAwarded,
                spells: spellsAwarded,
              ),
            ],
          ],
        );
      case 'monsterBattleVictory':
        final monsterName =
            (data['monsterName'] as String?)?.trim() ?? 'Monster';
        final rewardExperience =
            (data['rewardExperience'] as num?)?.toInt() ?? 0;
        final rewardGold = (data['rewardGold'] as num?)?.toInt() ?? 0;
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

  Widget _buildRewardSection(
    BuildContext context, {
    int experience = 0,
    int gold = 0,
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
