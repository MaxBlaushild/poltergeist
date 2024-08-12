import { User } from '@poltergeist/types';
import React from 'react';

export type PersonListItemProps = {
  user: User;
  onClick: (user: User) => void;
  actionArea: () => React.JSX.Element;
};

export const PersonListItem = ({ user, onClick, actionArea }: PersonListItemProps) => {
  return                       <li
  key={user.id}
  className="relative rounded-md p-3 text-sm/6 transition hover:bg-black/5 text-left flex items-center justify-between"
  onClick={() => onClick(user)}
>
  <div className="flex-grow flex flex-col">
    <div className="flex items-center space-x-3">
      <img
        src={user.profile?.profilePictureUrl || 'default-profile.png'}
        alt={`${user.name}'s profile`}
        className="h-9 w-9 rounded-full"
      />
      <div>
      <a href="#" className="font-semibold text-black">
        <span className="absolute inset-0" />
        {user.name}
      </a>
      </div>


    </div>
  </div>
  {actionArea()}
</li>;
};

export default PersonListItem;

