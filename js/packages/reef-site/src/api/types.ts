// Hand-written to match the JSON shapes reef-site's Go handlers actually
// produce (go/reef-site/internal/server/*.go, go/pkg/models/reef_*.go).
//
// R-2.3 says: if the repo already generates TS types from Go, use that
// pipeline; if not, define the API with OpenAPI and generate both sides
// from it. Neither exists in this repo (see go/reef-site/INVENTORY.md) and
// standing up a full OpenAPI toolchain nobody else in the repo uses was out
// of scope for this pass — this file is a deliberate, documented scope
// reduction, not a silent substitution. If the reef API surface grows
// past what one person can keep in sync by hand, that's the trigger to
// build the real pipeline the inventory doc describes.

export type ProductKind = 'configurable' | 'fixed';

export interface Product {
  id: string;
  createdAt: string;
  updatedAt: string;
  slug: string;
  name: string;
  kind: ProductKind;
  description: string;
  material: string;
  basePriceCents: number;
  images: string[];
  active: boolean;
  variants?: ProductVariant[];
}

export interface ProductVariant {
  id: string;
  createdAt: string;
  updatedAt: string;
  productId: string;
  variantKey: string;
  label: string;
  priceCents: number;
  active: boolean;
}

export interface TankProfile {
  id: string;
  createdAt: string;
  updatedAt: string;
  manufacturer: string;
  model: string;
  rimThicknessMm: number;
  rimWidthMm: number;
  glassThicknessMm: number;
  euroBrace: boolean;
  internalDims: Record<string, unknown>;
  verified: boolean;
  sourceUrl: string;
}

// R-4.4: the single source of parameter truth. The configurator form is
// rendered entirely from this document — see components/SchemaForm.tsx.
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

export interface Configuration {
  id: string;
  createdAt: string;
  updatedAt: string;
  productId: string;
  params: Record<string, unknown>;
  geometryHash: string | null;
  status: ConfigurationStatus;
  rejectionReason: string;
  priceCents: number | null;
  sessionId: string;
}

export interface BboxMm {
  xMm: number;
  yMm: number;
  zMm: number;
}

export interface PreviewResponse {
  geometryHash: string;
  previewUrl: string;
  bboxMm: BboxMm;
  plateFits: boolean;
  cached: boolean;
}

export interface ConfigureValidateResponse {
  configurationId: string;
  status: ConfigurationStatus;
}

export interface CartItemRequest {
  productSlug: string;
  variantKey?: string;
  configurationId?: string;
  quantity: number;
}

export interface CartItem {
  productSlug: string;
  productName: string;
  variantKey?: string;
  variantLabel?: string;
  configurationId?: string;
  quantity: number;
  unitPriceCents: number;
  lineTotalCents: number;
}

export interface CartResponse {
  items: CartItem[];
  subtotalCents: number;
  shippingCents: number;
  totalCents: number;
  remainingToFreeShippingCents: number;
  crossSell?: Product[];
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
  configurationId?: string;
  variantKey: string;
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

export type ReefEventType =
  | 'configurator_opened'
  | 'parameter_changed'
  | 'preview_rendered'
  | 'validation_rejected'
  | 'add_to_cart'
  | 'checkout_started'
  | 'purchase_completed'
  | 'share_link_created'
  | 'share_link_opened';

export interface OperatorMetrics {
  sinceDays: number;
  configuratorOpened: number;
  purchaseCompleted: number;
  conversionRate: number;
  validatedTotal: number;
  rejectedTotal: number;
  rejectionRate: number;
  rejectionsByRule: Record<string, number>;
  paidOrderCount: number;
  cogsRecordedForOrders: number;
  meanCogsCents: number;
  ordersReprinted: number;
  reprintRate: number;
  adSpendCents: number;
  cacCents: number;
}
