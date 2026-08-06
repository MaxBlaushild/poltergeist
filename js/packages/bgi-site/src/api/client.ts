import type {
  CartItemRequest,
  CartResponse,
  CheckoutResponse,
  Configuration,
  ConfigureValidateResponse,
  Game,
  GameDetail,
  CompatibilityInfo,
  Order,
  ParameterSchema,
  PreviewResponse,
  SleeveProfile,
  BgiEventType,
} from './types';

const BASE_URL = (import.meta.env.VITE_API_URL ?? '') + '/api/bgi';

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      message = body.error ?? message;
    } catch {
      // ignore — not JSON
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return res.json() as Promise<T>;
}

export const bgiApi = {
  listGames: () => request<Game[]>('/games'),
  getGame: (slug: string) => request<GameDetail>(`/games/${slug}`),
  getGameSchema: (slug: string) => request<ParameterSchema>(`/games/${slug}/schema`),
  getGameCompatibility: (slug: string) => request<CompatibilityInfo>(`/games/${slug}/compatibility`),
  listSleeveProfiles: () => request<SleeveProfile[]>('/sleeve-profiles'),

  preview: (productSlug: string, params: Record<string, unknown>, sessionId: string) =>
    request<PreviewResponse>('/configure/preview', {
      method: 'POST',
      body: JSON.stringify({ productSlug, params, sessionId }),
    }),

  validate: (productSlug: string, params: Record<string, unknown>, sessionId: string) =>
    request<ConfigureValidateResponse>('/configure/validate', {
      method: 'POST',
      body: JSON.stringify({ productSlug, params, sessionId }),
    }),

  getConfiguration: (id: string) => request<Configuration>(`/configurations/${id}`),

  cart: (items: CartItemRequest[]) =>
    request<CartResponse>('/cart', { method: 'POST', body: JSON.stringify({ items }) }),

  checkout: (
    items: CartItemRequest[],
    customerEmail: string,
    successUrl: string,
    cancelUrl: string,
    sessionId: string,
  ) =>
    request<CheckoutResponse>('/checkout', {
      method: 'POST',
      body: JSON.stringify({ items, customerEmail, successUrl, cancelUrl, sessionId }),
    }),

  getOrder: (token: string) => request<Order>(`/orders/${token}`),

  submitWaitlist: (email: string, requestedGame: string, sessionId: string) =>
    request<void>('/waitlist', {
      method: 'POST',
      body: JSON.stringify({ email, requestedGame, sessionId }),
    }),

  recordEvent: (
    eventType: BgiEventType,
    fields: {
      sessionId?: string;
      gameSlug?: string;
      configurationId?: string;
      rule?: string;
      metadata?: Record<string, unknown>;
    } = {},
  ) =>
    request<void>('/events', {
      method: 'POST',
      body: JSON.stringify({ eventType, ...fields }),
    }).catch(() => {
      // Analytics must never break the shopping experience.
    }),
};

export { ApiError };
