import 'package:flutter_test/flutter_test.dart';
import 'package:unclaimed_streets/models/point_of_interest.dart';
import 'package:unclaimed_streets/models/quest.dart';
import 'package:unclaimed_streets/models/quest_node.dart';
import 'package:unclaimed_streets/providers/quest_log_provider.dart';

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
