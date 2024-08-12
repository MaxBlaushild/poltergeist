import React, { useEffect, useState } from 'react';
import './MatchLobby.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import PersonListItem from './shared/PersonListItem.tsx';
import { Button } from './shared/Button.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useDrag, useDrop } from 'react-dnd';
import { Team as TeamModel, User } from '@poltergeist/types';

const stepTexts: string[] = ['Get ready', '3', '2', '1', 'Start!'];

export const MatchLobby = () => {
  const { match, createTeam, isStartingMatch, startMatch } = useMatchContext();
  const { currentUser } = useUserProfiles();
  const [toastText, setToastText] = useState<string | null>(null);
  const [countdownNumber, setCountdownNumber] = useState<number | null>(null);
  const [currentStepTextIndex, setCurrentStepTextIndex] = useState(-1);
  const [shouldCountdown, setShouldCountdown] = useState(false);
  const queryParams = new URLSearchParams(window.location.search);
  const teamId = queryParams.get('teamId');
  const startedAt = match?.startedAt;

  const [, drop] = useDrop(() => ({
    accept: 'person',
    drop: (item) => createTeam(),
  }));

  const sendToast = (text: string) => {
    setToastText(text);
    setTimeout(() => {
      setToastText(null);
    }, 1500);
  };


  useEffect(() => {
    let timeout;
    if (shouldCountdown) {
      const updateStep = (index) => {
        if (index < stepTexts.length + 1) {
          setCurrentStepTextIndex(index);
          timeout = setTimeout(() => updateStep(index + 1), 1000);
        } else {
          startMatch();
        }
      };
      updateStep(-1);
    }
    return () => {
      if (timeout) clearTimeout(timeout);
    };
  }, [shouldCountdown, setCurrentStepTextIndex]);

  if (!match) {
    return <></>;
  }

  const teams = match.teams;
  const otherTeams = match.teams.filter(
    (team) =>
      team.users.length > 0 &&
      !team.users.some((user) => user.id === currentUser?.id)
  );
  const usersTeam = match.teams.find((team) =>
    team.users.find((user) => user.id === currentUser?.id)
  );
  const canStartMatch = teams.length > 1 && currentUser?.id === match.creatorId;
  const matchLink = `${window.location.origin}/match/${match.id}`;
  const currentStepText =
  currentStepTextIndex > -1 ? stepTexts[currentStepTextIndex] : undefined;

  return (
    !currentStepText ? <div className="Match__lobby">
      <div className="flex flex-col gap-12 w-full">
        <div className="flex flex-col gap-3">
          <h2 className="text-3xl font-bold">Battle Lobby</h2>

        </div>
        <div className="flex flex-col gap-3 flex-start w-full">
          {usersTeam ? (
            <Team
              team={usersTeam}
              user={currentUser}
              sendToast={sendToast}
              matchLink={matchLink}
            />
          ) : null}
          {otherTeams?.map((team) => (
            <Team
              team={team}
              key={team.name}
              sendToast={sendToast}
              matchLink={matchLink}
            />
          ))}
          <div className="flex flex-col gap-3">
            <h3 className="text-lg font-bold text-start">New team</h3>
            <div className="rounded-xl bg-black/5 p-3" ref={drop}>
              <PersonListItem
                user={{
                  name: 'New team',
                  id: 'new-team',
                  phoneNumber: '',
                  profile: {
                    profilePictureUrl: 'https://crew-points-of-interest.s3.amazonaws.com/plus.png',
                    id: 'new-team',
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                    vieweeId: '',
                    viewerId: '',
                  },
                }}
                onClick={() => {}}
                actionArea={() => <></>}
              />
            <Button
            title="Get invite link"
            onClick={() => {
              navigator.clipboard.writeText(matchLink);
              setToastText('Invite link copied to clipboard');
              setTimeout(() => {
                setToastText(null);
              }, 1500);
            }}
          />
            </div>
          </div>
        </div>
        <Button title="Start Match" disabled={!canStartMatch} onClick={() => setShouldCountdown(true)} />
      </div>
      {toastText ? <Modal size={ModalSize.TOAST}>{toastText}</Modal> : null}
      </div> : <Modal>
      <h1>{currentStepText}</h1>
    </Modal>
  );
};

const Team = ({
  team,
  user,
  sendToast,
  matchLink,
}: {
  team: TeamModel;
  user?: User | null;
  sendToast: (text: string) => void;
  matchLink: string;
}) => {
  const { addUserToTeam } = useMatchContext();
  const [collected, drag] = useDrag(() => ({
    type: 'person',
    collect: (monitor) => ({
      isDragging: !!monitor.isDragging(),
    }),
  }));

  const [, drop] = useDrop(() => ({
    accept: 'person',
    drop: (item) => addUserToTeam(team.id),
  }));

  return (
    <div key={team.id} className="flex flex-col gap-3">
      <h3 className="text-lg font-bold text-start">{team.name}</h3>
      <div className="rounded-xl bg-black/5 p-3" ref={drop}>
        {user ? (
          <div ref={drag} {...collected}>
            <PersonListItem
              user={user}
              onClick={() => {}}
              actionArea={() => <></>}
            />
          </div>
        ) : null}
        {team.users
          .filter((u) => u.id !== user?.id)
          .map((user) => (
            <div key={user.id}>
              <PersonListItem
                user={user}
                onClick={() => {}}
                actionArea={() => <></>}
              />
            </div>
          ))}
                <Button
        title={`Get team invite link`}
        onClick={() => {
          navigator.clipboard.writeText(`${matchLink}?teamId=${team.id}`);
          sendToast(`Invite link copied to clipboard`);
        }}
      />
      </div>
    </div>
  );
};
