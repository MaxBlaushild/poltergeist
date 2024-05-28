import './ActivityCloud.css';
import { Activity } from '@poltergeist/types';
import React from 'react';

interface ActivityCloudProps {
  activities: Activity[];
  onClick: (activityId: string) => void;
  selectedActivityIds: string[];
}

const ActivityCloud = ({
  activities,
  onClick,
  selectedActivityIds,
}: ActivityCloudProps) => {
  return (
    <div className="ActivityCloud__activityListItems">
      {activities.map((activity) => (
        <div
          onClick={() => onClick(activity.id)}
          className={
            selectedActivityIds.includes(activity.id)
              ? 'ActivityCloud__activity--selected'
              : 'ActivityCloud__activity'
          }
          key={activity.id}
        >
          {activity.title}
        </div>
      ))}
    </div>
  );
};

export default ActivityCloud;
