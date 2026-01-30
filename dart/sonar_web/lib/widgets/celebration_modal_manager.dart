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

        return Dialog(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _titleFor(type),
                  style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        color: Colors.amber.shade700,
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

  String _titleFor(String? type) {
    switch (type) {
      case 'challenge':
        return 'Victory!';
      case 'levelUp':
        return 'Level Up!';
      case 'reputationUp':
        return 'Reputation Up!';
      default:
        return 'Congratulations!';
    }
  }

  Widget _contentFor(String? type, Map<String, dynamic> data, BuildContext context) {
    switch (type) {
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
                child: Text('+1', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
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
      default:
        return const Text('Task completed!');
    }
  }
}
