import { BrowserRouter, Link, Route, Routes } from 'react-router-dom';
import { useCart } from './hooks/useCart';
import { WaveDivider } from './components/Decor';
import Home from './pages/Home';
import ProductDetail from './pages/ProductDetail';
import Configure from './pages/Configure';
import TankLanding from './pages/TankLanding';
import Cart from './pages/Cart';
import OrderStatus from './pages/OrderStatus';
import HowToMeasure from './pages/HowToMeasure';
import MaterialsAndCare from './pages/MaterialsAndCare';
import Operator from './pages/Operator';

function Logo() {
  return (
    <Link to="/" className="group flex items-center gap-2">
      <span className="flex h-9 w-9 items-center justify-center rounded-full border-2 border-reef-ink bg-reef-glow shadow-pop-sm transition-all duration-300 ease-bounce group-hover:-translate-y-0.5 group-hover:rotate-12">
        <svg viewBox="0 0 24 24" className="h-4 w-4 text-reef-ink" aria-hidden="true">
          <path
            d="M2 15c2 2 4-2 6 0s4-2 6 0 4-2 6 0"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
          />
          <path
            d="M2 10c2 2 4-2 6 0s4-2 6 0 4-2 6 0"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            opacity="0.6"
          />
        </svg>
      </span>
      <span className="font-display text-xl font-bold tracking-tight text-reef-ink">reef</span>
    </Link>
  );
}

function Layout({ children }: { children: React.ReactNode }) {
  const { items } = useCart();
  const count = items.reduce((sum, i) => sum + i.quantity, 0);
  return (
    <div className="flex min-h-screen flex-col bg-reef-paper">
      <header className="sticky top-0 z-20 border-b-2 border-reef-ink/10 bg-reef-paper/90 backdrop-blur">
        <nav className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <Logo />
          <div className="flex items-center gap-3 text-xs font-medium sm:gap-6 sm:text-sm">
            <Link to="/how-to-measure" className="relative transition-colors hover:text-reef-coral">
              How to measure
            </Link>
            <Link to="/materials-and-care" className="hidden transition-colors hover:text-reef-coral sm:inline">
              Materials &amp; care
            </Link>
            <Link
              to="/cart"
              className="flex items-center gap-1.5 rounded-full border-2 border-reef-ink bg-white px-3 py-1.5 font-bold shadow-pop-sm transition-all duration-300 ease-bounce hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[5px_5px_0_0_#0b1c26] active:translate-x-0.5 active:translate-y-0.5 active:shadow-none"
            >
              Cart
              {count > 0 && (
                <span className="flex h-5 min-w-5 animate-pop-in items-center justify-center rounded-full bg-reef-coral px-1 text-xs font-bold text-reef-ink">
                  {count}
                </span>
              )}
            </Link>
          </div>
        </nav>
      </header>
      <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-10">{children}</main>
      <footer className="relative mt-8 bg-reef-lagoon text-reef-sand/70">
        <WaveDivider className="absolute -top-[1px] left-0 -translate-y-full text-reef-lagoon" />
        <div className="mx-auto max-w-5xl px-4 py-8 text-xs">
          <div className="mb-3 flex items-center gap-2 font-display text-sm text-reef-sand">
            <span className="h-1.5 w-1.5 rounded-full bg-reef-glow" />
            reef
          </div>
          PETG parts, made to order. Not certified food-grade or laboratory-grade — see{' '}
          <Link to="/materials-and-care" className="underline decoration-reef-glow/50 underline-offset-2 hover:text-reef-glow">
            Materials &amp; care
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
          <Route path="/products/:slug" element={<ProductDetail />} />
          <Route path="/configure/:slug" element={<Configure />} />
          <Route path="/tanks/:manufacturer/:model" element={<TankLanding />} />
          <Route path="/cart" element={<Cart />} />
          <Route path="/orders/:token" element={<OrderStatus />} />
          <Route path="/how-to-measure" element={<HowToMeasure />} />
          <Route path="/materials-and-care" element={<MaterialsAndCare />} />
          <Route path="/operator" element={<Operator />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}
