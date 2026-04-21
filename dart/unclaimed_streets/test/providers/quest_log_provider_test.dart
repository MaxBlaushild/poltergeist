import 'dart:collection';

import 'package:flutter_test/flutter_test.dart';
import 'package:unclaimed_streets/models/point_of_interest.dart';
import 'package:unclaimed_streets/models/quest.dart';
import 'package:unclaimed_streets/models/quest_log.dart';
import 'package:unclaimed_streets/models/quest_node.dart';
import 'package:unclaimed_streets/providers/quest_filter_provider.dart';
import 'package:unclaimed_streets/providers/quest_log_provider.dart';
import 'package:unclaimed_streets/providers/tags_provider.dart';
import 'package:unclaimed_streets/providers/zone_provider.dart';
import 'package:unclaimed_streets/services/api_client.dart';
import 'package:unclaimed_streets/services/quest_log_service.dart';
import 'package:unclaimed_streets/services/tags_service.dart';

void main() {
  const objectivePoi = PointOfInterest(
    id: 'poi-objective',
    name: 'Objective',
    lat: '1',
    lng: '2',
  );

  test('getMapPointsOfInterest excludes ready-to-turn-in objectives', () {
    final quests = [
      _quest(
        pointOfInterest: objectivePoi,
        isAccepted: true,
        readyToTurnIn: false,
      ),
      _quest(
        pointOfInterest: const PointOfInterest(
          id: 'poi-turn-in',
          name: 'Turn In',
          lat: '3',
          lng: '4',
        ),
        isAccepted: true,
        readyToTurnIn: true,
      ),
    ];

    final pois = getMapPointsOfInterest(quests);

    expect(pois.map((poi) => poi.id), ['poi-objective']);
  });

  test(
    'getAllPointsOfInterestIdsForQuest drops objective once ready to turn in',
    () {
      final activeQuest = _quest(
        pointOfInterest: objectivePoi,
        isAccepted: true,
        readyToTurnIn: false,
      );
      final turnInQuest = _quest(
        pointOfInterest: objectivePoi,
        isAccepted: true,
        readyToTurnIn: true,
      );

      expect(getAllPointsOfInterestIdsForQuest(activeQuest), ['poi-objective']);
      expect(getAllPointsOfInterestIdsForQuest(turnInQuest), isEmpty);
    },
  );

  test(
    'forgetQuest refreshes the quest log after a successful mutation',
    () async {
      final service = _FakeQuestLogService();
      service.enqueueQuestLog(
        QuestLog(
          quests: [
            _quest(
              pointOfInterest: objectivePoi,
              isAccepted: true,
              readyToTurnIn: false,
            ),
          ],
          trackedQuestIds: const ['quest-poi-objective-active'],
        ),
      );
      service.enqueueQuestLog(const QuestLog());

      final provider = QuestLogProvider(
        service,
        ZoneProvider(),
        TagsProvider(TagsService(ApiClient('http://localhost'))),
        QuestFilterProvider(),
      );

      await provider.refresh();
      expect(provider.quests, hasLength(1));

      final error = await provider.forgetQuest('quest-poi-objective-active');

      expect(error, isNull);
      expect(service.forgotQuestIds, ['quest-poi-objective-active']);
      expect(provider.quests, isEmpty);
      expect(provider.trackedQuestIds, isEmpty);
    },
  );
}

class _FakeQuestLogService extends QuestLogService {
  _FakeQuestLogService() : super(ApiClient('http://localhost'));

  final Queue<QuestLog> _questLogs = Queue<QuestLog>();
  final List<String> forgotQuestIds = <String>[];

  void enqueueQuestLog(QuestLog questLog) {
    _questLogs.add(questLog);
  }

  @override
  Future<QuestLog> getQuestLog({
    String? zoneId,
    List<String> tags = const [],
  }) async {
    if (_questLogs.isEmpty) {
      return const QuestLog();
    }
    return _questLogs.removeFirst();
  }

  @override
  Future<void> forgetQuest(String questId) async {
    forgotQuestIds.add(questId);
  }
}

Quest _quest({
  required PointOfInterest pointOfInterest,
  required bool isAccepted,
  required bool readyToTurnIn,
}) {
  return Quest(
    id: 'quest-${pointOfInterest.id}-${readyToTurnIn ? 'turnin' : 'active'}',
    name: 'Quest',
    description: 'Quest',
    isAccepted: isAccepted,
    readyToTurnIn: readyToTurnIn,
    currentNode: QuestNode(
      id: 'node-${pointOfInterest.id}',
      orderIndex: 0,
      pointOfInterest: pointOfInterest,
    ),
  );
}
