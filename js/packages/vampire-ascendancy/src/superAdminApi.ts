// The shared content library editor — characters, houses, items, and quiz
// questions are global (read by every instance), so editing them is
// restricted to super users, not any instance's Host/Co-Host. Not scoped to
// any instance, unlike gmApi.ts.
import { ApiError } from './api';
import { getUserAuth } from './userAuth';
import type { House } from './types';
import type { GMItem, GMItemDraft, GMQuizQuestions, GMCharacterFull, GMCharacterUpdate } from './gmApi';

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.unclaimedstreets.com';

async function admin<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getUserAuth()?.token;
  const res = await fetch(`${API_BASE}/vampire-ascendancy/admin${path}`, {
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

// ---- Houses ----
export const adminListHouses = () => admin<{ houses: House[] }>('/houses');
export const adminUpdateHouse = (id: string, tagline: string) =>
  admin<{ ok: boolean }>(`/houses/${id}`, { method: 'PUT', body: JSON.stringify({ tagline }) });

// ---- Characters ----
export interface AdminCharacter {
  id: string;
  name: string;
  title: string;
  roleType: string;
  isOptional: boolean;
  houseId: string | null;
  house?: string;
}
export const adminListCharacters = () => admin<{ characters: AdminCharacter[] }>('/characters');
export const adminGetCharacter = (id: string) =>
  admin<Omit<GMCharacterFull, 'sigil' | 'imageUrl' | 'playerName'>>(`/characters/${id}`);
export const adminUpdateCharacter = (
  id: string,
  body: Omit<GMCharacterUpdate, 'imageUrl' | 'playerName'>
) => admin<{ ok: boolean }>(`/characters/${id}`, { method: 'PUT', body: JSON.stringify(body) });

// ---- Items ----
export const adminListItems = () => admin<{ items: GMItem[] }>('/items');
export const adminCreateItem = (draft: GMItemDraft) =>
  admin<{ id: string }>('/items', { method: 'POST', body: JSON.stringify(draft) });
export const adminUpdateItem = (id: string, draft: GMItemDraft) =>
  admin<{ ok: boolean }>(`/items/${id}`, { method: 'PUT', body: JSON.stringify(draft) });
export const adminDeleteItem = (id: string) => admin<{ ok: boolean }>(`/items/${id}`, { method: 'DELETE' });
export const adminSetItemPhoto = (id: string, dataUrl: string) =>
  admin<{ ok: boolean }>(`/items/${id}/photo`, { method: 'POST', body: JSON.stringify({ dataUrl }) });
export const adminDeleteItemPhoto = (id: string) =>
  admin<{ ok: boolean }>(`/items/${id}/photo`, { method: 'DELETE' });

// ---- Quiz ----
export const adminGetQuizQuestions = () => admin<GMQuizQuestions>('/quiz/questions');
export const adminUpdateQuizQuestions = (body: GMQuizQuestions) =>
  admin<{ ok: boolean }>('/quiz/questions', { method: 'PUT', body: JSON.stringify(body) });

// ---- Super users ----
export interface AdminSuperUser {
  userId: string;
  name?: string;
  email?: string;
}
export const adminListSuperUsers = () => admin<{ superUsers: AdminSuperUser[] }>('/super-users');
export const adminAddSuperUser = (email: string) =>
  admin<{ ok: boolean }>('/super-users', { method: 'POST', body: JSON.stringify({ email }) });
export const adminRemoveSuperUser = (userId: string) =>
  admin<{ ok: boolean }>(`/super-users/${userId}`, { method: 'DELETE' });
