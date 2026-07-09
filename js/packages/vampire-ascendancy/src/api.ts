import type { MeResponse, HouseStanding, QuizResponse, HouseOverview, Game } from './types';

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.unclaimedstreets.com';

const TOKEN_KEY = 'vampireToken';

export function saveToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, token: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}/vampire-ascendancy${path}`, {
    ...init,
    headers: {
      'X-Player-Token': token,
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

export function getMe(token: string): Promise<MeResponse> {
  return request<MeResponse>('/me', token);
}

// ---- Public login endpoints (no token) ----

export interface PublicCharacter {
  id: string;
  name: string;
  title: string;
  house?: string;
}

async function publicGet<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}/vampire-ascendancy${path}`);
  if (!res.ok) throw new ApiError(res.statusText, res.status);
  return res.json() as Promise<T>;
}

export function getCharacterPublic(id: string): Promise<PublicCharacter> {
  return publicGet<PublicCharacter>(`/characters/${id}`);
}

export function listCharacters(): Promise<{ characters: PublicCharacter[] }> {
  return publicGet<{ characters: PublicCharacter[] }>('/characters');
}

export async function login(characterId: string, password: string): Promise<{ token: string }> {
  const res = await fetch(`${API_BASE}/vampire-ascendancy/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ characterId, password }),
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
  return res.json() as Promise<{ token: string }>;
}

export function getLeaderboard(token: string): Promise<{ standings: HouseStanding[] }> {
  return request('/leaderboard', token);
}

export function getGames(token: string): Promise<{ games: Game[] }> {
  return request('/games', token);
}

export interface InventoryItem {
  id: string;
  name: string;
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
export function getInventory(token: string): Promise<InventoryResponse> {
  return request<InventoryResponse>('/inventory', token);
}
export function setInventoryTarget(
  token: string,
  id: string,
  targetPlayerId: string
): Promise<{ ok: boolean }> {
  return request(`/inventory/${id}/target`, token, {
    method: 'POST',
    body: JSON.stringify({ targetPlayerId }),
  });
}

export function getHouseOverview(token: string, houseId: string): Promise<HouseOverview> {
  return request<HouseOverview>(`/houses/${houseId}/overview`, token);
}

export function photoUrl(id: string): string {
  return `${API_BASE}/vampire-ascendancy/photos/${id}`;
}

export function submitMission(
  token: string,
  missionId: string,
  answer: string,
  opts?: { photos?: string[]; clearPhotos?: boolean }
): Promise<{ status: string; playerAnswer: string; awardedBt: number }> {
  return request(`/missions/${missionId}/submit`, token, {
    method: 'POST',
    body: JSON.stringify({ answer, photos: opts?.photos, clearPhotos: opts?.clearPhotos }),
  });
}

export function getQuiz(token: string): Promise<QuizResponse> {
  return request<QuizResponse>('/quiz', token);
}

export function submitQuizPart1(token: string, answer: string): Promise<{ ok: boolean }> {
  return request('/quiz/part1/submit', token, {
    method: 'POST',
    body: JSON.stringify({ answer }),
  });
}

export function submitQuizPart2(
  token: string,
  answers: { questionId: string; answer: string }[]
): Promise<{ ok: boolean }> {
  return request('/quiz/part2/submit', token, {
    method: 'POST',
    body: JSON.stringify({ answers }),
  });
}
