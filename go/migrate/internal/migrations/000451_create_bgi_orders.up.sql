-- Mirrors reef_orders/reef_order_items (000429) exactly, FKs retargeted.
CREATE TABLE IF NOT EXISTS bgi_orders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  order_token TEXT NOT NULL UNIQUE,
  stripe_session_id TEXT NOT NULL DEFAULT '',
  customer_email TEXT NOT NULL DEFAULT '',
  shipping_address JSONB NOT NULL DEFAULT '{}'::jsonb,
  status TEXT NOT NULL CHECK (status IN ('pending_payment', 'paid', 'fulfilled', 'cancelled')) DEFAULT 'pending_payment',
  fulfillment_provider TEXT NOT NULL DEFAULT 'manual',
  fulfillment_status TEXT NOT NULL DEFAULT '',
  fulfillment_external_id TEXT NOT NULL DEFAULT '',
  subtotal_cents INTEGER NOT NULL DEFAULT 0,
  shipping_cents INTEGER NOT NULL DEFAULT 0,
  total_cents INTEGER NOT NULL DEFAULT 0,
  cogs_cents INTEGER,
  reprint_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_bgi_orders_stripe_session_id ON bgi_orders(stripe_session_id);
CREATE INDEX IF NOT EXISTS idx_bgi_orders_status ON bgi_orders(status);

CREATE TABLE IF NOT EXISTS bgi_order_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  order_id UUID NOT NULL REFERENCES bgi_orders(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES bgi_products(id),
  configuration_id UUID REFERENCES bgi_configurations(id),
  quantity INTEGER NOT NULL DEFAULT 1,
  unit_price_cents INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bgi_order_items_order_id ON bgi_order_items(order_id);
