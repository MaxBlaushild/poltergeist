import type {
  CartItemRequest,
  CartResponse,
  CheckoutResponse,
  Configuration,
  ConfigureValidateResponse,
  CustomerAuth,
  MyOrder,
  Order,
  OperatorMetrics,
  OperatorOrder,
  ParameterSchema,
  PreviewResponse,
  Product,
  ReefEventType,
  TankProfile,
} from './types';
import { clearAdminAuth, getAdminAuthHeader } from './adminAuth';
import { clearStoredAuth, getStoredAuth } from '../hooks/useCustomerAuth';

const BASE_URL = (import.meta.env.VITE_API_URL ?? '') + '/api/reef';

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const auth = getStoredAuth();
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(auth ? { Authorization: `Bearer ${auth.token}` } : {}),
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

// operator/* routes are gated with HTTP Basic Auth (see api/adminAuth.ts) —
// a 401 here means the stored password is wrong or missing, so it's cleared
// to force the login gate to re-prompt rather than loop silently.
async function operatorRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const authHeader = getAdminAuthHeader();
  try {
    return await request<T>(path, {
      ...init,
      headers: {
        ...(authHeader ? { Authorization: authHeader } : {}),
        ...(init?.headers ?? {}),
      },
    });
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      clearAdminAuth();
    }
    throw err;
  }
}

export const reefApi = {
  listProducts: () => request<Product[]>('/products'),
  getProduct: (slug: string) => request<Product>(`/products/${slug}`),
  getProductSchema: (slug: string) => request<ParameterSchema>(`/products/${slug}/schema`),
  listTanks: () => request<TankProfile[]>('/tanks'),

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

  register: (name: string, email: string, password: string) =>
    request<CustomerAuth>('/auth/register', { method: 'POST', body: JSON.stringify({ name, email, password }) }),

  login: (email: string, password: string) =>
    request<CustomerAuth>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),

  loginWithGoogle: (idToken: string) =>
    request<CustomerAuth>('/auth/google', { method: 'POST', body: JSON.stringify({ idToken }) }),

  myOrders: () =>
    request<MyOrder[]>('/me/orders').catch((err) => {
      if (err instanceof ApiError && err.status === 401) clearStoredAuth();
      throw err;
    }),

  recordEvent: (
    eventType: ReefEventType,
    fields: {
      sessionId?: string;
      productSlug?: string;
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

  operatorMetrics: (days: number, adSpendCents: number) =>
    operatorRequest<OperatorMetrics>(`/operator/metrics?days=${days}&adSpendCents=${adSpendCents}`),

  operatorOrders: () => operatorRequest<OperatorOrder[]>('/operator/orders'),

  updateOrderFulfillment: (orderId: string, status: 'printed' | 'shipped') =>
    operatorRequest<OperatorOrder>(`/operator/orders/${orderId}/fulfillment`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),

  fulfillOrderWithSlant: (orderId: string) =>
    operatorRequest<OperatorOrder>(`/operator/orders/${orderId}/fulfill-slant`, { method: 'POST' }),

  refreshSlantStatus: (orderId: string) =>
    operatorRequest<OperatorOrder>(`/operator/orders/${orderId}/refresh-slant-status`, { method: 'POST' }),
};

export { ApiError };
