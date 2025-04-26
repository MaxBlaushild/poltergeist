import { PointOfInterestChallenge, QuestLog } from "@poltergeist/types";

const filterChallengesForPointOfInterest = (challenges: PointOfInterestChallenge[], questLog: QuestLog) => {
  return challenges.filter((challenge) => {
    return questLog.quests.some((quest) => {
      return quest.rootNode?.objectives.some((objective) => {
        return objective.challenge.id === challenge.id;
      });
    });
  });
};

export default filterChallengesForPointOfInterest;