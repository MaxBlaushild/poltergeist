// Hand-written to match the JSON shapes go/bgi-site's Go handlers actually
// produce (go/bgi-site/internal/server/*.go, go/pkg/models/bgi_*.go) — same
// deliberate, documented scope reduction as reef-site's own api/types.ts
// (no generated-types pipeline exists in this repo).

export interface Game {
  id: string;
  createdAt: string;
  updatedAt: string;
  slug: string;
  name: string;
  publisher: string;
  yearPublished: number | null;
  active: boolean;
  productSlug?: string;
}

export interface Expansion {
  id: string;
  gameId: string;
  slug: string;
  name: string;
  standalone: boolean;
  active: boolean;
}

export interface BoxProfile {
  id: string;
  gameId: string;
  slug: string;
  label: string;
  source: 'original' | 'aftermarket';
  interiorLengthMm: number;
  interiorWidthMm: number;
  interiorDepthMm: number;
  depthIsPlaceholder: boolean;
  verified: boolean;
  sourceUrl: string;
  measurementNotes: string;
}

export interface SleeveProfile {
  id: string;
  classKey: string;
  label: string;
  baseCardThicknessMm: number;
  sleeveMaterialThicknessMm: number;
  verified: boolean;
  sourceUrl: string;
}

export interface ComponentManifest {
  id: string;
  gameId: string;
  expansionId: string | null;
  componentType: string;
  cardWidthMm: number | null;
  cardHeightMm: number | null;
  count: number;
  verified: boolean;
  sourceUrl: string;
  notes: string;
}

export interface GameDetail extends Game {
  expansions: Expansion[];
  boxProfiles: BoxProfile[];
  sleeveProfiles: SleeveProfile[];
  manifest: ComponentManifest[];
}

export interface CompatibilityInfo {
  boxProfiles: BoxProfile[];
  manifest: ComponentManifest[];
}

// R-4.4: the single source of parameter truth — same shape as reef's.
export interface ParameterProperty {
  type: string | (string | null)[];
  minimum?: number;
  maximum?: number;
  enum?: (string | number)[];
  default?: unknown;
  'x-label'?: string;
  'x-helpText'?: string;
  'x-diagramAsset'?: string;
  'x-unit'?: string;
  'x-control'?: string;
  'x-autofills'?: string[];
  'x-derivedBoundFrom'?: string[];
}

export interface ParameterSchema {
  type: string;
  required: string[];
  properties: Record<string, ParameterProperty>;
}

export type ConfigurationStatus = 'pending' | 'valid' | 'rejected';

export interface Tray {
  stlUrl?: string;
  heightMm: number;
  quantity: number;
}

export interface Configuration {
  id: string;
  createdAt: string;
  updatedAt: string;
  productId: string;
  params: Record<string, unknown>;
  configHash: string | null;
  status: ConfigurationStatus;
  rejectionReason: string;
  priceCents: number | null;
  sessionId: string;
  productSlug?: string;
  trays?: Tray[];
}

export interface PreviewResponse {
  assembledHeightMm: number;
  boxInteriorDepthMm: number;
  fitsBox: boolean;
  boxVerified: boolean;
  depthIsPlaceholder: boolean;
  unassembledComponents?: string[];
  firstTrayPreviewUrl?: string;
  trayCount: number;
}

export interface ConfigureValidateResponse {
  configurationId: string;
  status: ConfigurationStatus;
}

export interface CartItemRequest {
  productSlug: string;
  configurationId: string;
  quantity: number;
}

export interface CartItem {
  productSlug: string;
  productName: string;
  configurationId: string;
  quantity: number;
  unitPriceCents: number;
  lineTotalCents: number;
}

export interface CartResponse {
  items: CartItem[];
  subtotalCents: number;
  shippingCents: number;
  totalCents: number;
}

export interface CheckoutResponse {
  checkoutUrl: string;
  orderToken: string;
}

export type OrderStatus = 'pending_payment' | 'paid' | 'fulfilled' | 'cancelled';

export interface OrderItem {
  id: string;
  createdAt: string;
  orderId: string;
  productId: string;
  productName: string;
  configurationId?: string;
  quantity: number;
  unitPriceCents: number;
}

export interface Order {
  id: string;
  createdAt: string;
  updatedAt: string;
  orderToken: string;
  stripeSessionId: string;
  customerEmail: string;
  shippingAddress: Record<string, unknown>;
  status: OrderStatus;
  fulfillmentProvider: string;
  fulfillmentStatus: string;
  fulfillmentExternalId: string;
  subtotalCents: number;
  shippingCents: number;
  totalCents: number;
  cogsCents: number | null;
  reprintCount: number;
  items: OrderItem[];
}

export type BgiEventType =
  | 'game_selected'
  | 'expansion_toggled'
  | 'sleeve_selected'
  | 'box_selected'
  | 'configurator_opened'
  | 'parameter_changed'
  | 'preview_rendered'
  | 'fit_indicator_shown'
  | 'validation_rejected'
  | 'fit_check_failed'
  | 'add_to_cart'
  | 'checkout_started'
  | 'purchase_completed'
  | 'waitlist_submitted'
  | 'share_link_created'
  | 'share_link_opened';
