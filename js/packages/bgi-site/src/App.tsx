import { BrowserRouter, Link, Route, Routes } from 'react-router-dom';
import { useCart } from './hooks/useCart';
import Home from './pages/Home';
import GameDetail from './pages/GameDetail';
import Configure from './pages/Configure';
import Cart from './pages/Cart';
import OrderStatus from './pages/OrderStatus';
import CompatibilityDisclaimer from './pages/CompatibilityDisclaimer';

function Logo() {
  return (
    <Link to="/" className="group flex items-center gap-2">
      <span className="flex h-9 w-9 items-center justify-center rounded-full border-2 border-bgi-ink bg-bgi-glow shadow-pop-sm transition-all duration-300 ease-bounce group-hover:-translate-y-0.5 group-hover:rotate-12">
        <svg viewBox="0 0 24 24" className="h-4 w-4 text-bgi-ink" aria-hidden="true">
          <rect x="4" y="4" width="16" height="16" rx="3" fill="none" stroke="currentColor" strokeWidth="2" />
          <circle cx="9" cy="9" r="1.3" fill="currentColor" />
          <circle cx="15" cy="15" r="1.3" fill="currentColor" />
          <circle cx="9" cy="15" r="1.3" fill="currentColor" opacity="0.6" />
          <circle cx="15" cy="9" r="1.3" fill="currentColor" opacity="0.6" />
        </svg>
      </span>
      <span className="font-display text-xl font-bold tracking-tight text-bgi-ink">trays</span>
    </Link>
  );
}

function Layout({ children }: { children: React.ReactNode }) {
  const { items } = useCart();
  const count = items.reduce((sum, i) => sum + i.quantity, 0);
  return (
    <div className="flex min-h-screen flex-col bg-bgi-paper">
      <header className="sticky top-0 z-20 border-b-2 border-bgi-ink/10 bg-bgi-paper/90 backdrop-blur">
        <nav className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <Logo />
          <div className="flex items-center gap-3 text-xs font-medium sm:gap-6 sm:text-sm">
            <Link
              to="/cart"
              className="flex items-center gap-1.5 rounded-full border-2 border-bgi-ink bg-white px-3 py-1.5 font-bold shadow-pop-sm transition-all duration-300 ease-bounce hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[5px_5px_0_0_#241a12] active:translate-x-0.5 active:translate-y-0.5 active:shadow-none"
            >
              Cart
              {count > 0 && (
                <span className="flex h-5 min-w-5 animate-pop-in items-center justify-center rounded-full bg-bgi-coral px-1 text-xs font-bold text-white">
                  {count}
                </span>
              )}
            </Link>
          </div>
        </nav>
      </header>
      <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-10">{children}</main>
      <footer className="mt-8 bg-bgi-lagoon text-bgi-sand/70">
        <div className="mx-auto max-w-5xl px-4 py-8 text-xs">
          <div className="mb-3 flex items-center gap-2 font-display text-sm text-bgi-sand">
            <span className="h-1.5 w-1.5 rounded-full bg-bgi-glow" />
            trays
          </div>
          PETG parts, made to order. Not affiliated with any game publisher — see{' '}
          <Link to="/games" className="underline decoration-bgi-glow/50 underline-offset-2 hover:text-bgi-glow">
            compatibility notes
          </Link>
          .
        </div>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/games/:slug" element={<GameDetail />} />
          <Route path="/games/:slug/compatibility" element={<CompatibilityDisclaimer />} />
          <Route path="/configure/:slug" element={<Configure />} />
          <Route path="/cart" element={<Cart />} />
          <Route path="/orders/:token" element={<OrderStatus />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}
