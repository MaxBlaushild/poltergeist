import React, { useState }from 'react';
import { useQuestLogContext } from '../../contexts/QuestLogContext.tsx';
import { usePointOfInterestContext } from '../../contexts/PointOfInterestContext.tsx';
import { PointOfInterest, PointOfInterestChallenge } from '@poltergeist/types';
import { SubmitAnswerForChallenge } from './SubmitAnswerForChallenge.tsx';

type PointOfInterestTasksProps = {
  pointOfInterest: PointOfInterest;
  onClose: (immediate: boolean) => void;
};

export const PointOfInterestTasks = ({ pointOfInterest, onClose }: PointOfInterestTasksProps) => {
  const { quests, pendingTasks } = useQuestLogContext();
  const [selectedTask, setSelectedTask] = useState<PointOfInterestChallenge | null>(null);
  const tasksForPointOfInterest = pendingTasks[pointOfInterest.id] ?? [];

  return (
    <div>
      {selectedTask ? (
        <div>
          <button 
            className="mb-4 px-3 py-1 bg-gray-200 rounded-lg hover:bg-gray-300"
            onClick={() => setSelectedTask(null)}
          >
            Back to Tasks
          </button>
          <SubmitAnswerForChallenge
            challenge={selectedTask}
            pointOfInterest={pointOfInterest}
            onSubmit={(immediate) => {
              setSelectedTask(null);
              onClose(immediate);
            }}
          />
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {tasksForPointOfInterest.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-gray-500">
              <p className="text-lg font-medium">No tasks available</p>
              <p className="text-sm">Check back later for new tasks</p>
            </div>
          ) : (
            <ul className="list-none">
              {tasksForPointOfInterest.map((task) => (
                <li
                  key={task.id}
                  onClick={() => setSelectedTask(task)}
                  className="mb-3 cursor-pointer group"
                >
                  <div className="flex items-center gap-3 px-4 py-3 bg-gray-100 rounded-lg group-hover:bg-gray-200 transition-colors">
                    <div className="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-blue-100 rounded-full text-blue-600">
                      📝
                    </div>
                    <div className="flex-grow">
                      <p className="text-gray-800">{task.question}</p>
                    </div>
                    <div className="flex-shrink-0 text-gray-400 group-hover:text-gray-600">
                      →
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
};
