// Platform-level calls: not scoped to any Toast (instance). Real-account
// auth (register/login/Google), "My Toasts," hosting a new Toast, and
// accepting a Co-Host invite.
import { ApiError } from './api';
import { getUserAuth, type UserAuth } from './userAuth';

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.unclaimedstreets.com';

async function request<T>(path: string, init?: RequestInit, auth = true): Promise<T> {
  const token = auth ? getUserAuth()?.token : undefined;
  const res = await fetch(`${API_BASE}/vampire-ascendancy${path}`, {
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

// ---- Auth ----
export const registerUser = (name: string, email: string, password: string) =>
  request<UserAuth>('/auth/register', { method: 'POST', body: JSON.stringify({ name, email, password }) }, false);
export const loginUser = (email: string, password: string) =>
  request<UserAuth>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }, false);
export const loginUserWithGoogle = (idToken: string) =>
  request<UserAuth>('/auth/google', { method: 'POST', body: JSON.stringify({ idToken }) }, false);

// ---- Instances ("Toasts") ----
// A row is either a Toast this account administers ("admin" — Host/
// Co-Host, links to the GM console) or one it plays a character in
// ("player" — links to the player app, with `character` for a preview
// card). An account that's both gets shown once, as admin.
export interface MyToast {
  id: string;
  name: string;
  status: string;
  createdAt: string;
  kind: 'admin' | 'player';
  role?: 'owner' | 'admin'; // present when kind === 'admin'; owner -> Host, admin -> Co-Host
  character?: {
    id: string;
    name: string;
    title: string;
    house?: string;
    imageUrl?: string;
  };
}
export const listMyToasts = () => request<{ instances: MyToast[] }>('/instances');
export const hostToast = (name: string, mysteryId: string) =>
  request<{ id: string; name: string }>('/instances', {
    method: 'POST',
    body: JSON.stringify({ name, mysteryId }),
  });

// ---- Mysteries (picker only — the full editor is super-user-only, see
// superAdminApi.ts) ----
export interface ActiveMystery {
  id: string;
  name: string;
  summary: string;
}
export const listActiveMysteries = () => request<{ mysteries: ActiveMystery[] }>('/mysteries');

export const acceptCoHostInvite = (token: string) =>
  request<{ instanceId: string }>(`/invites/${encodeURIComponent(token)}/accept`, { method: 'POST' });

// ---- Player invites (RSVP) ----
export interface PlayerInvite {
  guestName: string;
  status: 'pending' | 'accepted' | 'declined';
  instanceId: string;
  instanceName?: string;
  character?: {
    id: string;
    name: string;
    title: string;
    preEventInfo: string;
    house?: string;
  };
}
export const getPlayerInvite = (token: string) =>
  request<PlayerInvite>(`/rsvp/${encodeURIComponent(token)}`, undefined, false);
export const declinePlayerInvite = (token: string) =>
  request<{ ok: boolean }>(`/rsvp/${encodeURIComponent(token)}/decline`, { method: 'POST' }, false);
export const acceptPlayerInvite = (token: string) =>
  request<{ instanceId: string }>(`/rsvp/${encodeURIComponent(token)}/accept`, { method: 'POST' });
