import { ApiError } from './api';
import { getUserAuth } from './userAuth';
import type { GameState, HouseStanding, House } from './types';

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.unclaimedstreets.com';

// The GM console (GMAdmin) is always mounted at /e/:instanceId/gm, so it
// sets this once (from the URL) before any of the calls below run — see
// setGMInstanceId. Every gm* call targets whichever Toast that was.
let currentInstanceId = '';
export function setGMInstanceId(id: string) {
  currentInstanceId = id;
}

async function gm<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getUserAuth()?.token;
  const res = await fetch(`${API_BASE}/vampire-ascendancy/i/${currentInstanceId}/gm${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
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
  guestLabel: string;
  active: boolean;
  btTotal: number;
  character: { id: string; name: string; roleType: string; house?: string } | null;
}

export interface GMCharacter {
  id: string;
  name: string;
  title: string;
  roleType: string;
  isOptional: boolean;
  house?: string;
  preEventInfo?: string;
  tags?: string[];
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
  first: GameWinner[];
  second: GameWinner[];
  third: GameWinner[];
  startMinutes: number | null;
  endMinutes: number | null;
  location: string;
  assignedGm: string;
  runNotes: string;
}
export const gmGetStandings = () => gm<{ standings: HouseStanding[] }>('/standings');

export interface HouseBreakdown {
  houseId: string;
  name: string;
  sources: Record<string, number>; // e.g. { quiz_part2: 4.5, game: 8, manual: 2 }
  items: { name: string; amount: number; holder: string }[];
  itemTotal: number;
  total: number;
}
export const gmGetStandingsBreakdown = () =>
  gm<{ houses: HouseBreakdown[] }>('/standings/breakdown');
export const gmListGames = () => gm<{ games: GMGame[] }>('/games');
export const gmCreateGame = (name: string, ordinal = 0) =>
  gm<{ id: string }>('/games', { method: 'POST', body: JSON.stringify({ name, ordinal }) });
export const gmRecordGameResult = (
  id: string,
  body: { firstIds: string[]; secondIds: string[]; thirdIds: string[] }
) => gm<{ ok: boolean }>(`/games/${id}/result`, { method: 'POST', body: JSON.stringify(body) });
export const gmUpdateGame = (id: string, name: string, ordinal: number) =>
  gm<{ ok: boolean }>(`/games/${id}`, { method: 'PUT', body: JSON.stringify({ name, ordinal }) });
export const gmDeleteGame = (id: string) => gm<{ ok: boolean }>(`/games/${id}`, { method: 'DELETE' });
export const gmSetGameSchedule = (
  id: string,
  body: {
    startMinutes: number | null;
    endMinutes: number | null;
    location: string;
    assignedGm: string;
    runNotes: string;
  }
) => gm<{ ok: boolean }>(`/games/${id}/schedule`, { method: 'PUT', body: JSON.stringify(body) });
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
// Editing questions moved to the Super Admin dashboard — see
// adminGetQuizQuestions/adminUpdateQuizQuestions in superAdminApi.ts.

export interface GMQuizSubmission {
  id: string;
  part: number;
  answer: string;
  isCorrect: boolean | null;
  aiScore: number | null;
  aiRationale: string;
  awardedBt: number;
  locked: boolean;
  gradeStatus: string; // '' | queued | grading | graded | failed
  gradeError: string;
  gradeAttempts: number;
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
// Retry grading: omit submissionId to re-enqueue every not-yet-graded answer.
export const gmRegradePart1 = (submissionId?: string) =>
  gm<{ status: string; queued: number }>('/quiz/part1/regrade', {
    method: 'POST',
    body: JSON.stringify(submissionId ? { submissionId } : {}),
  });
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
  tags: string[];
  // The AI tag-generation job's last-known state for this character — ''
  // (never run) | 'queued' | 'generating' | 'generated' | 'failed'.
  tagsGenerationStatus: string;
  tagsGenerationError: string;
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
  tags: string[];
  missions: GMMissionEdit[];
}
export const gmGetCharacter = (id: string) => gm<GMCharacterFull>(`/characters/${id}`);
// Editing bios/secrets/missions moved to the Super Admin dashboard — see
// adminGetCharacter/adminUpdateCharacter in superAdminApi.ts. The one
// character edit a Host/Co-Host can still make directly is this Toast's
// portrait for that character.
export const gmSetCharacterPortrait = (id: string, imageUrl: string) =>
  gm<{ ok: boolean }>(`/characters/${id}/portrait`, { method: 'PUT', body: JSON.stringify({ imageUrl }) });
// House taglines moved to the Super Admin dashboard — see adminUpdateHouse.

// ---- Inventory ----
export interface GMItem {
  id: string;
  code: string;
  name: string;
  category: string;
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
  hasPhoto: boolean;
}
export type GMItemDraft = Omit<GMItem, 'id' | 'hasPhoto'>;

// Public URL for a catalog item's reference photo (cache-busted by `v`).
export const itemPhotoUrl = (id: string, v?: string | number) =>
  `${API_BASE}/vampire-ascendancy/items/${id}/photo${v ? `?v=${v}` : ''}`;
export interface GMPlayerItem {
  id: string;
  playerId: string;
  playerName: string;
  itemName: string;
  effect: string;
  targetsPlayer: boolean;
  targetName: string | null;
}
// This Toast's included items — see GET /gm/library/items for the full
// catalog + toggles, or superAdminApi.ts's adminListItems for the shared
// catalog editor (create/edit/delete/photo — super users only).
export const gmListItems = () => gm<{ items: GMItem[] }>('/items');
export const gmListPlayerItems = () => gm<{ playerItems: GMPlayerItem[] }>('/player-items');
export const gmAssignItem = (playerId: string, itemId: string) =>
  gm<{ id: string }>('/player-items', { method: 'POST', body: JSON.stringify({ playerId, itemId }) });
export const gmRemovePlayerItem = (id: string) =>
  gm<{ ok: boolean }>(`/player-items/${id}`, { method: 'DELETE' });
export const gmTransferPlayerItem = (id: string, playerId: string) =>
  gm<{ ok: boolean }>(`/player-items/${id}/owner`, { method: 'PUT', body: JSON.stringify({ playerId }) });

// ---- Invites tab: text a real person a link to accept/decline a character ----
export interface GMInvite {
  id: string;
  guestName: string;
  phoneNumber: string;
  status: 'pending' | 'accepted' | 'declined';
  createdAt: string;
  character?: { id: string; name: string; title: string; house?: string };
}
export const gmListInvites = () => gm<{ invites: GMInvite[] }>('/invites');
export const gmCreateInvite = (guestName: string, phoneNumber: string, characterId: string) =>
  gm<{ id: string; warning?: string }>('/invites', {
    method: 'POST',
    body: JSON.stringify({ guestName, phoneNumber, characterId }),
  });
export const gmDeleteInvite = (id: string) => gm<{ ok: boolean }>(`/invites/${id}`, { method: 'DELETE' });
export const gmResendInvite = (id: string) =>
  gm<{ ok: boolean }>(`/invites/${id}/resend`, { method: 'POST' });

// ---- Co-Hosts tab: Host/Co-Host management ----
export interface GMAdminRow {
  userId: string;
  role: 'owner' | 'admin'; // owner -> Host, admin -> Co-Host
  name?: string;
  email?: string;
}
export interface GMPendingInvite {
  id: string;
  email: string;
  expiresAt: string;
}
export const gmListAdmins = () =>
  gm<{ admins: GMAdminRow[]; pendingInvites: GMPendingInvite[] }>('/admins');
export const gmInviteAdmin = (email: string) =>
  gm<{ id: string; token: string; expiresAt: string }>('/admins/invite', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
export const gmDeleteAdminInvite = (id: string) =>
  gm<{ ok: boolean }>(`/admins/invites/${id}`, { method: 'DELETE' });
export const gmRemoveAdmin = (userId: string) =>
  gm<{ ok: boolean }>(`/admins/${userId}`, { method: 'DELETE' });
export const gmTransferOwnership = (toUserId: string) =>
  gm<{ ok: boolean }>('/admins/transfer', { method: 'POST', body: JSON.stringify({ toUserId }) });

// ---- Content tab: shared-library inclusion toggle (items only — the quiz
// is fixed, not customizable per instance; which characters are "in" is
// implicit via invites; see the Invites tab) ----
// Same shape as GMItem (+ included) — the Content tab's toggle shows the
// same rich card the Items tab's assign picker does.
export interface GMLibraryItem extends Omit<GMItem, 'id'> {
  id: string;
  included: boolean;
}
export const gmListLibraryItems = () => gm<{ items: GMLibraryItem[] }>('/library/items');
export const gmSetItemIncluded = (id: string, included: boolean) =>
  gm<{ ok: boolean }>(`/library/items/${id}`, { method: 'PUT', body: JSON.stringify({ included }) });
