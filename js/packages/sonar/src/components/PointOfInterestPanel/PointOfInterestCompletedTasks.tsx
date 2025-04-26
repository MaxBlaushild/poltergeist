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
      {tasksForPointOfInterest.map((task) => (
        <div
          key={task.id}
          className="flex items-center gap-2 px-4 py-3 bg-gray-100 rounded-lg"
        >
          <span>✅</span>
          <span>{task.question}</span>
        </div>
      ))}
    </div>
  );
};
