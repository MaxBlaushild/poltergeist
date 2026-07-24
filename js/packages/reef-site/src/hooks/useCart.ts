import { useCallback, useEffect, useState } from 'react';
import type { CartItemRequest } from '../api/types';

const CART_KEY = 'reef_cart';

// The cart itself is client-side state (localStorage) — R-3.2's data model
// has no cart table, and R-8.1's POST /cart is a pure pricing/validation
// call, not a persistence call. Nothing about "what's in the cart" needs a
// server round trip until checkout.
function readCart(): CartItemRequest[] {
  try {
    const raw = localStorage.getItem(CART_KEY);
    return raw ? (JSON.parse(raw) as CartItemRequest[]) : [];
  } catch {
    return [];
  }
}

function writeCart(items: CartItemRequest[]) {
  localStorage.setItem(CART_KEY, JSON.stringify(items));
  window.dispatchEvent(new Event('reef-cart-changed'));
}

function itemKey(item: CartItemRequest): string {
  return [item.productSlug, item.variantKey ?? '', item.configurationId ?? ''].join('|');
}

export function useCart() {
  const [items, setItems] = useState<CartItemRequest[]>(() => readCart());

  useEffect(() => {
    const onChange = () => setItems(readCart());
    window.addEventListener('reef-cart-changed', onChange);
    window.addEventListener('storage', onChange);
    return () => {
      window.removeEventListener('reef-cart-changed', onChange);
      window.removeEventListener('storage', onChange);
    };
  }, []);

  const addItem = useCallback((item: CartItemRequest) => {
    const current = readCart();
    const existingIndex = current.findIndex((i) => itemKey(i) === itemKey(item));
    let next: CartItemRequest[];
    if (existingIndex >= 0) {
      next = [...current];
      next[existingIndex] = { ...next[existingIndex], quantity: next[existingIndex].quantity + item.quantity };
    } else {
      next = [...current, item];
    }
    writeCart(next);
  }, []);

  const removeItem = useCallback((item: CartItemRequest) => {
    const next = readCart().filter((i) => itemKey(i) !== itemKey(item));
    writeCart(next);
  }, []);

  const setQuantity = useCallback((item: CartItemRequest, quantity: number) => {
    const next = readCart().map((i) => (itemKey(i) === itemKey(item) ? { ...i, quantity } : i));
    writeCart(next.filter((i) => i.quantity > 0));
  }, []);

  const clear = useCallback(() => writeCart([]), []);

  return { items, addItem, removeItem, setQuantity, clear };
}
