import type { MeResponse, HouseStanding, QuizResponse, HouseOverview, Game } from './types';
import { getUserAuth } from './userAuth';

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.unclaimedstreets.com';

// The player app (PlayerShell) is always mounted at /e/:instanceId, so it
// sets this once (from the URL) before any of the calls below run — the
// same setXInstanceId pattern gmApi.ts and superAdminApi.ts use. Player
// auth is a real signed-in account (see userAuth.ts) — there is no more
// separate per-character token/sigil.
let currentInstanceId = '';
export function setPlayerInstanceId(id: string) {
  currentInstanceId = id;
}
export function getSessionInstanceId() {
  return currentInstanceId;
}

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getUserAuth()?.token;
  const res = await fetch(`${API_BASE}/vampire-ascendancy/i/${currentInstanceId}${path}`, {
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

export function getMe(): Promise<MeResponse> {
  return request<MeResponse>('/me');
}

// ---- Public projector feed (no auth) ----
async function publicGet<T>(instanceId: string, path: string): Promise<T> {
  const res = await fetch(`${API_BASE}/vampire-ascendancy/i/${instanceId}${path}`);
  if (!res.ok) throw new ApiError(res.statusText, res.status);
  return res.json() as Promise<T>;
}
export function getBroadcastStandings(instanceId: string): Promise<{ standings: HouseStanding[] }> {
  return publicGet<{ standings: HouseStanding[] }>(instanceId, '/broadcast/standings');
}
export function getBroadcastGames(instanceId: string): Promise<{ games: Game[] }> {
  return publicGet<{ games: Game[] }>(instanceId, '/broadcast/games');
}

export function getLeaderboard(): Promise<{ standings: HouseStanding[] }> {
  return request('/leaderboard');
}

export function getGames(): Promise<{ games: Game[] }> {
  return request('/games');
}

export interface InventoryItem {
  id: string;
  name: string;
  category: string;
  description: string;
  effect: string;
  targetsPlayer: boolean;
  targetPlayerId: string | null;
}
export interface InventoryTarget {
  playerId: string;
  name: string;
}
export interface InventoryResponse {
  items: InventoryItem[];
  targets: InventoryTarget[];
  locked: boolean;
}
export function getInventory(): Promise<InventoryResponse> {
  return request<InventoryResponse>('/inventory');
}
export function setInventoryTarget(id: string, targetPlayerId: string): Promise<{ ok: boolean }> {
  return request(`/inventory/${id}/target`, {
    method: 'POST',
    body: JSON.stringify({ targetPlayerId }),
  });
}

export function getHouseOverview(houseId: string): Promise<HouseOverview> {
  return request<HouseOverview>(`/houses/${houseId}/overview`);
}

export function photoUrl(id: string): string {
  return `${API_BASE}/vampire-ascendancy/photos/${id}`;
}

export function submitMission(
  missionId: string,
  answer: string,
  opts?: { photos?: string[]; clearPhotos?: boolean }
): Promise<{ status: string; playerAnswer: string; awardedBt: number }> {
  return request(`/missions/${missionId}/submit`, {
    method: 'POST',
    body: JSON.stringify({ answer, photos: opts?.photos, clearPhotos: opts?.clearPhotos }),
  });
}

export function getQuiz(): Promise<QuizResponse> {
  return request<QuizResponse>('/quiz');
}

export function submitQuizPart1(answer: string): Promise<{ ok: boolean }> {
  return request('/quiz/part1/submit', {
    method: 'POST',
    body: JSON.stringify({ answer }),
  });
}

export function submitQuizPart2(
  answers: { questionId: string; answer: string }[]
): Promise<{ ok: boolean }> {
  return request('/quiz/part2/submit', {
    method: 'POST',
    body: JSON.stringify({ answers }),
  });
}

// Sequential Part 2: lock one answer and advance. The server rejects out-of-order
// or already-answered questions.
export function submitQuizPart2Answer(
  questionId: string,
  answer: string
): Promise<{ ok: boolean; done: boolean }> {
  return request('/quiz/part2/answer', {
    method: 'POST',
    body: JSON.stringify({ questionId, answer }),
  });
}
