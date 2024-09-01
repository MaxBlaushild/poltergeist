import React, { useEffect, useState } from 'react';
import './MatchLobby.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import PersonListItem from './shared/PersonListItem.tsx';
import { Button, ButtonColor, ButtonSize } from './shared/Button.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useDrag, useDrop } from 'react-dnd';
import { Team as TeamModel, User } from '@poltergeist/types';
import TextInput from './shared/TextInput.tsx';

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
        if (index < stepTexts.length) {
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

  return !currentStepText ? (
    <div className="Match__lobby">
      <div className="flex flex-col gap-12 w-full">
        <div className="flex flex-col gap-3">
          <h2 className="text-3xl font-bold">Battle Lobby</h2>
        </div>
        <div className="flex flex-col gap-3 flex-start w-full">
          {usersTeam ? (
            <Team
              editable
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
        </div>
        <div className="flex flex-row gap-3">
          <Button
            title="Invite"
            onClick={() => {
              navigator.clipboard.writeText(matchLink);
              setToastText('Invite link copied to clipboard');
              setTimeout(() => {
                setToastText(null);
              }, 1500);
            }}
          />
          <Button
            title="Start"
            // disabled={!canStartMatch}
            onClick={() => setShouldCountdown(true)}
          />
        </div>
      </div>
      {toastText ? <Modal size={ModalSize.TOAST}>{toastText}</Modal> : null}
    </div>
  ) : (
    <Modal>
      <h1>{currentStepText}</h1>
    </Modal>
  );
};

const Team = ({
  team,
  user,
  sendToast,
  matchLink,
  editable = false,
}: {
  team: TeamModel;
  user?: User | null;
  sendToast: (text: string) => void;
  matchLink: string;
  editable?: boolean;
}) => {
  const { addUserToTeam, editTeamName } = useMatchContext();
  const [teamName, setTeamName] = useState(team.name);
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
      {!editable ? (
        <h3 className="text-lg font-bold text-start">{team.name}</h3>
      ) : (
        <div className="flex flex-row gap-3">
          <TextInput
            value={teamName}
            label="Your team"
            onChange={(name) => setTeamName(name)}
          />
          <div className="w-24 h-full flex flex-col justify-end">
            <Button
            title="Save"
            onClick={() => {
              editTeamName(team.id, teamName);
              sendToast('Team name updated');
            }}
          />
          </div>
        </div>
      )}
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
          title={`Invite to team`}
          onClick={() => {
            navigator.clipboard.writeText(`${matchLink}?teamId=${team.id}`);
            sendToast(`Invite link copied to clipboard`);
          }}
        />
      </div>
    </div>
  );
};
