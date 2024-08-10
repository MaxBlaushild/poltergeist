import React from "react";
import "./Match.css";
import { useMatchContext } from "../contexts/MatchContext.tsx";

export const Match = () => {
  const { match } = useMatchContext();

  return <div className="Match__lobby">Match</div>;
};
