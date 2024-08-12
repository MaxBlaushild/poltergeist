import React, { useEffect } from "react";
import "./Match.css";
import { useMatchContext } from "../contexts/MatchContext.tsx";
import { redirect } from "react-router-dom";
import { MatchLobby } from "./MatchLobby.tsx";
import { MatchAfterparty } from "./MatchAfterparty.tsx";
import { MatchInProgress } from "./MatchInProgress.tsx";
import { Match as MatchType } from "@poltergeist/types";

type MatchProps = {
  match: MatchType | null;
};

export const Match = ({ match }: MatchProps) => {
  const { getCurrentMatch } = useMatchContext();

  useEffect(() => {
    const intervalId = setInterval(() => {
      if (match) {
        getCurrentMatch();
      }
    }, 5000);
    return () => clearInterval(intervalId);
  }, [match, getCurrentMatch]);

  if (!match) {
    return <></>;
  }

  if (match && !match.startedAt) {
    return <MatchLobby />;
  }

  if (match && !!match.endedAt) {
    return <MatchAfterparty />;
  }

  return <MatchInProgress />;
};
