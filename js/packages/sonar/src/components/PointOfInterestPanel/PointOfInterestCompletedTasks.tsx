import React, { useState }from 'react';
import { useQuestLogContext } from '../../contexts/QuestLogContext.tsx';
import { usePointOfInterestContext } from '../../contexts/PointOfInterestContext.tsx';
import { PointOfInterest, PointOfInterestChallenge } from '@poltergeist/types';
import { SubmitAnswerForChallenge } from './SubmitAnswerForChallenge.tsx';

type PointOfInterestCompletedTasksProps = {
  pointOfInterest: PointOfInterest;
};

export const PointOfInterestCompletedTasks = ({ pointOfInterest }: PointOfInterestCompletedTasksProps) => {
  const { quests, completedTasks } = useQuestLogContext();
  const tasksForPointOfInterest = completedTasks[pointOfInterest.id] ?? [];

  if (tasksForPointOfInterest.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-gray-500">
        <p className="text-lg font-medium">No completed tasks yet</p>
        <p className="text-sm">Complete tasks to see them here</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {tasksForPointOfInterest.map((task) => {
        const quest = quests.find((quest) => quest.id === task.questId);
        
        return (
          <div
            key={task.challenge.id}
            className="mb-3 group"
          >
            <div className="flex items-center gap-3 px-4 py-3 bg-gray-100 rounded-lg group-hover:bg-gray-200 transition-colors">
              <div className="flex-grow text-left">
                {quest && <p className="text-xs text-gray-400">{quest?.name}</p>}
                <p className="text-black font-bold">{task.challenge.question}</p>
              </div>
              <div className="flex-shrink-0 text-gray-400 group-hover:text-gray-600">
                ✅
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
