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
        final showFailureColor =
            (type == 'scenarioOutcome' && !scenarioSuccess) ||
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
      case 'scenarioOutcome':
        final successful = data['successful'] == true;
        return successful ? 'Scenario Success!' : 'Scenario Failed';
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
        if (goldAwarded > 0) {
          rewards.add(Text('+$goldAwarded Gold'));
        }
        for (final item in itemsAwarded) {
          final name = item['name'] as String? ?? 'Item';
          final quantity = (item['quantity'] as num?)?.toInt() ?? 1;
          rewards.add(Text('+$quantity $name'));
        }
        for (final spell in spellsAwarded) {
          final name = spell['name'] as String? ?? 'Spell';
          rewards.add(Text('+Spell: $name'));
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
        ];
        if (goldAwarded > 0) {
          rewards.add(
            Text(
              '+$goldAwarded Gold',
              style: Theme.of(
                context,
              ).textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700),
            ),
          );
        }
        for (final item in itemsAwarded) {
          final name = item['name'] as String? ?? 'Item';
          final quantity = (item['quantity'] as num?)?.toInt() ?? 1;
          rewards.add(Text('+$quantity $name'));
        }
        if (goldAwarded <= 0 && itemsAwarded.isEmpty) {
          rewards.add(const Text('No loot this time.'));
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: rewards,
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
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (data['experienceAwarded'] != null)
              Text('+${data['experienceAwarded']} XP'),
            if (data['reputationAwarded'] != null)
              Text('+${data['reputationAwarded']} Reputation'),
            if (data['goldAwarded'] != null)
              Text('+${data['goldAwarded']} Gold'),
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
              if (rewardExperience > 0) Text('+$rewardExperience XP'),
              if (rewardGold > 0) Text('+$rewardGold Gold'),
              for (final item in itemsAwarded)
                Text(
                  '+${(item['quantity'] as num?)?.toInt() ?? 1} ${item['name'] as String? ?? 'Item'}',
                ),
              for (final spell in spellsAwarded)
                Text('+Spell: ${spell['name'] as String? ?? 'Spell'}'),
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
        ];
        if (rewardExperience > 0) {
          rewards.add(Text('+$rewardExperience XP'));
        }
        if (rewardGold > 0) {
          rewards.add(Text('+$rewardGold Gold'));
        }
        for (final item in itemsAwarded) {
          final name = item['name'] as String? ?? 'Item';
          final quantity = (item['quantity'] as num?)?.toInt() ?? 1;
          rewards.add(Text('+$quantity $name'));
        }
        if (rewardExperience <= 0 && rewardGold <= 0 && itemsAwarded.isEmpty) {
          rewards.add(const Text('No loot this time.'));
        }
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
}
