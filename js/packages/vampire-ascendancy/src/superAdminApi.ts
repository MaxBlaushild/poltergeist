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
// Secrets, pre-/post-Act-1 context, and missions are all excluded — every
// one of them is mystery-scoped now and edited from the Mysteries tab's
// per-character content editor instead (see MYSTERY_REQUIREMENTS.md).
export const adminGetCharacter = (id: string) =>
  admin<
    Omit<GMCharacterFull, 'sigil' | 'imageUrl' | 'playerName' | 'secrets' | 'preAct1Context' | 'postAct1Context' | 'missions'>
  >(`/characters/${id}`);
export const adminUpdateCharacter = (
  id: string,
  body: Omit<GMCharacterUpdate, 'imageUrl' | 'playerName' | 'postAct1Context' | 'missions'>
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

// ---- Mysteries (and subplots — a sibling, not a separate table: same row
// shape, isSubplot picks which) ----
// The underlying story an instance's players are solving — see
// MYSTERY_REQUIREMENTS.md. Quiz questions and (per-character) secrets/
// missions are scoped to a mystery instead of being shared/global. An
// instance picks one required mystery (isSubplot=false) plus zero or many
// subplots (isSubplot=true).
export interface AdminMystery {
  id: string;
  name: string;
  summary: string;
  active: boolean;
  isSubplot: boolean;
  beatCount: number;
}
// A beat is shared, reusable content — the same beat can be attached to
// multiple mysteries/subplots at once (see MYSTERY_REQUIREMENTS.md).
// linkCount > 1 means editing its title/description here also changes what
// every other mystery/subplot sharing it shows.
export interface AdminMysteryBeat {
  id: string;
  ordinal: number;
  title: string;
  description: string;
  linkCount: number;
}
export interface AdminMysteryFull {
  id: string;
  name: string;
  summary: string;
  fullLore: string;
  active: boolean;
  isSubplot: boolean;
  beats: AdminMysteryBeat[];
}
export const adminListMysteries = () => admin<{ mysteries: AdminMystery[] }>('/mysteries');
export const adminCreateMystery = (name: string, isSubplot: boolean) =>
  admin<{ id: string; name: string }>('/mysteries', { method: 'POST', body: JSON.stringify({ name, isSubplot }) });
export const adminGetMystery = (id: string) => admin<AdminMysteryFull>(`/mysteries/${id}`);
export const adminUpdateMystery = (
  id: string,
  body: {
    name: string;
    summary: string;
    fullLore: string;
    active: boolean;
    isSubplot: boolean;
    // id is omitted (or '') for a beat being created; present for one
    // being edited or reattached — preserving it is what keeps the beat's
    // id, and any secret's beatId pointing at it, stable across saves, and
    // is also how attaching an existing (shared) beat works.
    beats: { id?: string; title: string; description: string }[];
  }
) => admin<{ ok: boolean }>(`/mysteries/${id}`, { method: 'PUT', body: JSON.stringify(body) });

// Every beat across every mystery/subplot — the "attach an existing beat"
// picker, so a super user can reuse a beat that already says the same
// thing instead of duplicating it.
export interface AdminBeat {
  id: string;
  title: string;
  description: string;
  linkCount: number;
}
export const adminListBeats = () => admin<{ beats: AdminBeat[] }>('/beats');

// ---- Quiz, scoped to a mystery ----
export const adminGetMysteryQuiz = (mysteryId: string) => admin<GMQuizQuestions>(`/mysteries/${mysteryId}/quiz`);
export const adminUpdateMysteryQuiz = (mysteryId: string, body: GMQuizQuestions) =>
  admin<{ ok: boolean }>(`/mysteries/${mysteryId}/quiz`, { method: 'PUT', body: JSON.stringify(body) });

// ---- Character content, scoped to a mystery: secrets + missions + pre-/post-Act-1 context ----
export interface AdminMysterySecret {
  ordinal: number;
  body: string;
  beatId: string | null;
}
export interface AdminMysteryMission {
  ordinal: number;
  tier: string;
  rewardBt: number;
  prompt: string;
  answerFormat: string;
}
export interface AdminMysteryCharacterContent {
  secrets: AdminMysterySecret[];
  missions: AdminMysteryMission[];
  preAct1Context: string;
  postAct1Context: string;
}
export const adminGetCharacterContentForMystery = (mysteryId: string, characterId: string) =>
  admin<AdminMysteryCharacterContent>(`/mysteries/${mysteryId}/characters/${characterId}/content`);
export const adminUpdateCharacterContentForMystery = (
  mysteryId: string,
  characterId: string,
  body: {
    secrets: { body: string; beatId: string | null }[];
    missions: { tier: string; rewardBt: number; prompt: string; answerFormat: string }[];
    preAct1Context: string;
    postAct1Context: string;
  }
) =>
  admin<{ ok: boolean }>(`/mysteries/${mysteryId}/characters/${characterId}/content`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });

// ---- Beat-centric secrets: who knows this beat, across the whole cast ----
// Complements the character-centric content editor above — the Story tab's
// beat panel, for assigning secrets to characters right where a beat is
// authored, instead of leaving to walk the cast one character at a time.
export interface AdminBeatSecret {
  id: string;
  characterId: string;
  body: string;
}
export const adminListBeatSecrets = (mysteryId: string, beatId: string) =>
  admin<{ secrets: AdminBeatSecret[] }>(`/mysteries/${mysteryId}/beats/${beatId}/secrets`);
export const adminCreateBeatSecret = (mysteryId: string, beatId: string, characterId: string, body: string) =>
  admin<{ id: string }>(`/mysteries/${mysteryId}/beats/${beatId}/secrets`, {
    method: 'POST',
    body: JSON.stringify({ characterId, body }),
  });
export const adminUpdateSecretBody = (secretId: string, body: string) =>
  admin<{ ok: boolean }>(`/secrets/${secretId}`, { method: 'PUT', body: JSON.stringify({ body }) });
export const adminDeleteSecret = (secretId: string) => admin<{ ok: boolean }>(`/secrets/${secretId}`, { method: 'DELETE' });

// ---- Secrets tab: every secret in the system, flat, plus its creation
// form ----
export interface AdminSecretRow {
  id: string;
  characterId: string;
  characterName: string;
  mysteryId: string;
  mysteryName: string;
  isSubplot: boolean;
  beatId: string | null;
  beatTitle: string | null;
  body: string;
}
export const adminListAllSecrets = () => admin<{ secrets: AdminSecretRow[] }>('/secrets');

// A beat is picked together with which mystery/subplot it's for — a beat
// shared across several appears once per mystery, since a secret must
// belong to exactly one story.
export interface AdminMysteryBeatOption {
  mysteryId: string;
  mysteryName: string;
  isSubplot: boolean;
  beatId: string;
  beatTitle: string;
}
export const adminListMysteryBeatOptions = () => admin<{ options: AdminMysteryBeatOption[] }>('/mystery-beat-options');

export const adminCreateSecret = (body: { characterId: string; mysteryId: string; beatId: string; body: string }) =>
  admin<{ id: string }>('/secrets', { method: 'POST', body: JSON.stringify(body) });

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
