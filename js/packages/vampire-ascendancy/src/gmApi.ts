import { ApiError } from './api';
import type { GameState, HouseStanding, House } from './types';

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.unclaimedstreets.com';

const PASS_KEY = 'vampireGMPass';
const NAME_KEY = 'vampireGMName';

export function getGMAuth() {
  return {
    pass: sessionStorage.getItem(PASS_KEY) || '',
    name: sessionStorage.getItem(NAME_KEY) || '',
  };
}
export function setGMAuth(pass: string, name: string) {
  sessionStorage.setItem(PASS_KEY, pass);
  sessionStorage.setItem(NAME_KEY, name);
}
export function clearGMAuth() {
  sessionStorage.removeItem(PASS_KEY);
  sessionStorage.removeItem(NAME_KEY);
}

async function gm<T>(path: string, init?: RequestInit): Promise<T> {
  const { pass, name } = getGMAuth();
  const res = await fetch(`${API_BASE}/vampire-ascendancy/gm${path}`, {
    ...init,
    headers: {
      'X-GM-Passcode': pass,
      'X-GM-Name': name,
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      /* ignore */
    }
    throw new ApiError(message, res.status);
  }
  return res.json() as Promise<T>;
}

// ---- Types ----
export interface GMSubmission {
  id: string;
  status: string;
  playerAnswer: string;
  awardedBt: number;
  guestLabel: string;
  characterName: string;
  houseName: string;
  missionTier: string;
  missionPrompt: string;
  missionAnswerFormat: string;
  rewardBt: number;
  photoIds: string[];
}

export interface GMPlayer {
  id: string;
  token: string;
  guestLabel: string;
  active: boolean;
  btTotal: number;
  character: { id: string; name: string; roleType: string; house?: string; sigil?: string } | null;
}

export interface GMCharacter {
  id: string;
  name: string;
  title: string;
  roleType: string;
  isOptional: boolean;
  house?: string;
}

// ---- Calls ----
export const gmGetState = () => gm<GameState>('/state');
export const gmSetUnlock = (unlocked: boolean) =>
  gm<GameState>('/unlock', { method: 'POST', body: JSON.stringify({ unlocked }) });
export const gmSetAct = (act: string) =>
  gm<GameState>('/act', { method: 'POST', body: JSON.stringify({ act }) });
export const gmResetGame = (force = false) =>
  gm<{ ok: boolean }>('/reset', { method: 'POST', body: JSON.stringify({ confirm: 'RESET', force }) });

export interface StandingsExport {
  exportedAt: string;
  houseFavor: { houseId: string; name: string; favor: number }[];
  players: {
    playerId: string;
    playerName: string;
    active: boolean;
    bloodTokens: number;
    character: string | null;
    house: string | null;
  }[];
}
export const gmExportStandings = () => gm<StandingsExport>('/export');

export const gmListHouses = () => gm<{ houses: House[] }>('/houses');
export const gmAwardHF = (houseId: string, delta: number, reason: string) =>
  gm<{ standings: HouseStanding[] }>('/hf', {
    method: 'POST',
    body: JSON.stringify({ houseId, delta, reason }),
  });
export const gmAwardBT = (playerId: string, delta: number, reason: string) =>
  gm<{ ok: boolean }>('/bt', { method: 'POST', body: JSON.stringify({ playerId, delta, reason }) });

export const gmListSubmissions = (status: string) =>
  gm<{ submissions: GMSubmission[] }>(`/submissions?status=${encodeURIComponent(status)}`);
export const gmApprove = (id: string, awardedBt?: number) =>
  gm(`/submissions/${id}/approve`, {
    method: 'POST',
    body: JSON.stringify(awardedBt != null ? { awardedBt } : {}),
  });
export const gmRedeem = (id: string) => gm(`/submissions/${id}/redeem`, { method: 'POST' });
export const gmReject = (id: string) => gm(`/submissions/${id}/reject`, { method: 'POST' });

export const gmPushNotification = (
  title: string,
  body: string,
  scope: 'all' | 'house' | 'player',
  targetId?: string
) =>
  gm<{ id: string }>('/notifications', {
    method: 'POST',
    body: JSON.stringify({ title, body, scope, targetId: targetId || '' }),
  });
export const gmClearNotifications = () =>
  gm<{ ok: boolean }>('/notifications/clear', { method: 'POST' });

// ---- Physical games ----
export interface GameWinner {
  characterId: string;
  characterName: string;
  house?: string;
}
export interface GMGame {
  id: string;
  ordinal: number;
  name: string;
  status: 'pending' | 'played';
  first: GameWinner | null;
  second: GameWinner | null;
  third: GameWinner | null;
}
export const gmGetStandings = () => gm<{ standings: HouseStanding[] }>('/standings');
export const gmListGames = () => gm<{ games: GMGame[] }>('/games');
export const gmCreateGame = (name: string, ordinal = 0) =>
  gm<{ id: string }>('/games', { method: 'POST', body: JSON.stringify({ name, ordinal }) });
export const gmRecordGameResult = (
  id: string,
  body: { firstId?: string; secondId?: string; thirdId?: string; participantIds?: string[] }
) => gm<{ ok: boolean }>(`/games/${id}/result`, { method: 'POST', body: JSON.stringify(body) });
export const gmUpdateGame = (id: string, name: string, ordinal: number) =>
  gm<{ ok: boolean }>(`/games/${id}`, { method: 'PUT', body: JSON.stringify({ name, ordinal }) });
export const gmDeleteGame = (id: string) => gm<{ ok: boolean }>(`/games/${id}`, { method: 'DELETE' });
export const gmClearGameResult = (id: string) =>
  gm<{ ok: boolean }>(`/games/${id}/clear`, { method: 'POST' });

// ---- Quiz question editor ----
export interface GMQuizPart2Edit {
  ordinal?: number;
  prompt: string;
  options: string[];
  correctAnswer: string;
  hfValue: number;
  tier: string;
}
export interface GMQuizQuestions {
  part1: { prompt: string; rubric: string; maxBt: number };
  part2: GMQuizPart2Edit[];
}
export const gmGetQuizQuestions = () => gm<GMQuizQuestions>('/quiz/questions');
export const gmUpdateQuizQuestions = (body: GMQuizQuestions) =>
  gm<{ ok: boolean }>('/quiz/questions', { method: 'PUT', body: JSON.stringify(body) });

export interface GMQuizSubmission {
  id: string;
  part: number;
  answer: string;
  isCorrect: boolean | null;
  aiScore: number | null;
  aiRationale: string;
  awardedBt: number;
  locked: boolean;
  guestLabel: string;
  characterName: string;
  houseName: string;
  ordinal: number;
  prompt: string;
  questionType: string;
}

export const gmSetPart1Open = (open: boolean) =>
  gm<GameState>('/quiz/part1', { method: 'POST', body: JSON.stringify({ open }) });
export const gmSetPart2Open = (open: boolean) =>
  gm<GameState>('/quiz/part2', { method: 'POST', body: JSON.stringify({ open }) });
export const gmGradePart1 = () => gm<{ status: string }>('/quiz/part1/grade', { method: 'POST' });
export const gmOverridePart1BT = (submissionId: string, awardedBt: number) =>
  gm<{ ok: boolean }>('/quiz/part1/override', {
    method: 'POST',
    body: JSON.stringify({ submissionId, awardedBt }),
  });
export const gmRescorePart2 = () =>
  gm<{ standings: HouseStanding[] }>('/quiz/part2/rescore', { method: 'POST' });
export const gmListQuizSubmissions = () =>
  gm<{ submissions: GMQuizSubmission[] }>('/quiz/submissions');

// Final Blood Token tally with item effects resolved (steal/deduct/multiplier/
// double-games, with immune/reflect precedence). finalBt = quizBt + physicalBt + itemBt.
export interface GMTallyRow {
  playerId: string;
  character: string;
  house: string;
  correct: number;
  mcTotal: number;
  quizBt: number;
  physicalBt: number;
  itemBt: number;
  finalBt: number;
  notes: string[];
}
export const gmGetTally = () => gm<{ players: GMTallyRow[] }>('/quiz/tally');

export const gmListPlayers = () => gm<{ players: GMPlayer[] }>('/players');
export const gmUpdatePlayer = (
  id: string,
  body: { characterId: string | null; guestLabel: string; active: boolean }
) => gm(`/players/${id}`, { method: 'PUT', body: JSON.stringify(body) });
export const gmListCharacters = () => gm<{ characters: GMCharacter[] }>('/characters');

// ---- Character content editor ----
export interface GMMissionEdit {
  tier: string;
  rewardBt: number;
  prompt: string;
  answerFormat: string;
}
export interface GMCharacterFull {
  id: string;
  name: string;
  title: string;
  roleType: string;
  isOptional: boolean;
  houseId: string | null;
  preEventInfo: string;
  postAct1Context: string;
  imageUrl: string;
  sigil: string;
  playerName: string;
  secrets: { ordinal: number; body: string }[];
  missions: (GMMissionEdit & { ordinal: number })[];
}
export interface GMCharacterUpdate {
  name: string;
  title: string;
  roleType: string;
  houseId: string | null;
  preEventInfo: string;
  postAct1Context: string;
  imageUrl: string;
  playerName: string;
  secrets: string[];
  missions: GMMissionEdit[];
}
export const gmGetCharacter = (id: string) => gm<GMCharacterFull>(`/characters/${id}`);
export const gmUpdateCharacter = (id: string, body: GMCharacterUpdate) =>
  gm<{ ok: boolean }>(`/characters/${id}`, { method: 'PUT', body: JSON.stringify(body) });
export const gmUpdateHouse = (id: string, tagline: string) =>
  gm<{ ok: boolean }>(`/houses/${id}`, { method: 'PUT', body: JSON.stringify({ tagline }) });

// ---- Inventory ----
export interface GMItem {
  id: string;
  code: string;
  name: string;
  description: string;
  effect: string;
  targetsPlayer: boolean;
  hfEffect: number;
  btSelf: number;
  btFromTarget: number;
  btDeductTarget: number;
  quizBtPct: number;
  doubleGameBt: boolean;
  immune: boolean;
  reflect: boolean;
  stripResistance: boolean;
}
export type GMItemDraft = Omit<GMItem, 'id'>;
export interface GMPlayerItem {
  id: string;
  playerId: string;
  playerName: string;
  itemName: string;
  effect: string;
  targetsPlayer: boolean;
  targetName: string | null;
}
export const gmListItems = () => gm<{ items: GMItem[] }>('/items');
export const gmCreateItem = (draft: GMItemDraft) =>
  gm<{ id: string }>('/items', { method: 'POST', body: JSON.stringify(draft) });
export const gmUpdateItem = (id: string, draft: GMItemDraft) =>
  gm<{ ok: boolean }>(`/items/${id}`, { method: 'PUT', body: JSON.stringify(draft) });
export const gmDeleteItem = (id: string) =>
  gm<{ ok: boolean }>(`/items/${id}`, { method: 'DELETE' });
export const gmListPlayerItems = () => gm<{ playerItems: GMPlayerItem[] }>('/player-items');
export const gmAssignItem = (playerId: string, itemId: string) =>
  gm<{ id: string }>('/player-items', { method: 'POST', body: JSON.stringify({ playerId, itemId }) });
export const gmRemovePlayerItem = (id: string) =>
  gm<{ ok: boolean }>(`/player-items/${id}`, { method: 'DELETE' });
