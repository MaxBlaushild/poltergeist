import 'package:flutter_test/flutter_test.dart';
import 'package:unclaimed_streets/providers/completed_task_provider.dart';

void main() {
  test('queueLevelUpModal ignores duplicate or older level-up modals', () {
    final provider = CompletedTaskProvider();

    provider.queueLevelUpModal(newLevel: 2, previousLevel: 1);
    provider.queueLevelUpModal(newLevel: 2, previousLevel: 1);
    provider.queueLevelUpModal(newLevel: 1);
    provider.queueLevelUpModal(newLevel: 3, previousLevel: 2);

    expect(provider.currentModal?['type'], 'levelUp');
    expect(
      (provider.currentModal?['data'] as Map<String, dynamic>)['newLevel'],
      2,
    );

    provider.clearModal();

    expect(provider.currentModal?['type'], 'levelUp');
    expect(
      (provider.currentModal?['data'] as Map<String, dynamic>)['newLevel'],
      3,
    );

    provider.clearModal();
    expect(provider.currentModal, isNull);
  });
}
