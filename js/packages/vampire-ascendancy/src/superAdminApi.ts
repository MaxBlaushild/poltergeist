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
  tags?: string[];
}
export const adminListCharacters = () => admin<{ characters: AdminCharacter[] }>('/characters');
// Secrets are excluded — they're mystery-scoped now and edited from the
// Mysteries tab's per-character secrets editor instead (see
// MYSTERY_REQUIREMENTS.md).
export const adminGetCharacter = (id: string) =>
  admin<Omit<GMCharacterFull, 'sigil' | 'imageUrl' | 'playerName' | 'secrets'>>(`/characters/${id}`);
export const adminUpdateCharacter = (
  id: string,
  body: Omit<GMCharacterUpdate, 'imageUrl' | 'playerName'>
) => admin<{ ok: boolean }>(`/characters/${id}`, { method: 'PUT', body: JSON.stringify(body) });
// Queues the LLM tag-generation job for one character (reads bio/secrets/
// missions, proposes tags, overwrites the current tag list when it's
// done). Fire-and-poll: this just flips the status to "queued" — re-fetch
// adminGetCharacter to watch tagsGenerationStatus/tags update.
export const adminGenerateCharacterTags = (id: string) =>
  admin<{ ok: boolean }>(`/characters/${id}/generate-tags`, { method: 'POST' });

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

// ---- Mysteries ----
// The underlying story an instance's players are solving — see
// MYSTERY_REQUIREMENTS.md. Quiz questions and (per-character) secrets are
// scoped to a mystery instead of being shared/global.
export interface AdminMystery {
  id: string;
  name: string;
  summary: string;
  active: boolean;
  beatCount: number;
}
export interface AdminMysteryBeat {
  id: string;
  ordinal: number;
  body: string;
}
export interface AdminMysteryFull {
  id: string;
  name: string;
  summary: string;
  fullLore: string;
  active: boolean;
  beats: AdminMysteryBeat[];
}
export const adminListMysteries = () => admin<{ mysteries: AdminMystery[] }>('/mysteries');
export const adminCreateMystery = (name: string) =>
  admin<{ id: string; name: string }>('/mysteries', { method: 'POST', body: JSON.stringify({ name }) });
export const adminGetMystery = (id: string) => admin<AdminMysteryFull>(`/mysteries/${id}`);
export const adminUpdateMystery = (
  id: string,
  body: { name: string; summary: string; fullLore: string; active: boolean; beats: { body: string }[] }
) => admin<{ ok: boolean }>(`/mysteries/${id}`, { method: 'PUT', body: JSON.stringify(body) });

// ---- Quiz, scoped to a mystery ----
export const adminGetMysteryQuiz = (mysteryId: string) => admin<GMQuizQuestions>(`/mysteries/${mysteryId}/quiz`);
export const adminUpdateMysteryQuiz = (mysteryId: string, body: GMQuizQuestions) =>
  admin<{ ok: boolean }>(`/mysteries/${mysteryId}/quiz`, { method: 'PUT', body: JSON.stringify(body) });

// ---- Character secrets, scoped to a mystery ----
export interface AdminMysterySecret {
  ordinal: number;
  body: string;
  beatId: string | null;
}
export const adminGetCharacterSecretsForMystery = (mysteryId: string, characterId: string) =>
  admin<{ secrets: AdminMysterySecret[] }>(`/mysteries/${mysteryId}/characters/${characterId}/secrets`);
export const adminUpdateCharacterSecretsForMystery = (
  mysteryId: string,
  characterId: string,
  secrets: { body: string; beatId: string | null }[]
) =>
  admin<{ ok: boolean }>(`/mysteries/${mysteryId}/characters/${characterId}/secrets`, {
    method: 'PUT',
    body: JSON.stringify({ secrets }),
  });

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
