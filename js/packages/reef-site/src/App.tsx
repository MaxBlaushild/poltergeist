import { BrowserRouter, Link, Route, Routes } from 'react-router-dom';
import { useCart } from './hooks/useCart';
import Home from './pages/Home';
import ProductDetail from './pages/ProductDetail';
import Configure from './pages/Configure';
import TankLanding from './pages/TankLanding';
import Cart from './pages/Cart';
import OrderStatus from './pages/OrderStatus';
import HowToMeasure from './pages/HowToMeasure';
import MaterialsAndCare from './pages/MaterialsAndCare';
import Operator from './pages/Operator';

function Layout({ children }: { children: React.ReactNode }) {
  const { items } = useCart();
  const count = items.reduce((sum, i) => sum + i.quantity, 0);
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-reef-teal/20 bg-reef-ink text-reef-sand">
        <nav className="max-w-5xl mx-auto flex items-center justify-between px-4 py-3">
          <Link to="/" className="text-lg font-semibold tracking-wide">
            reef
          </Link>
          <div className="flex items-center gap-4 text-sm">
            <Link to="/how-to-measure" className="hover:text-reef-coral">
              How to measure
            </Link>
            <Link to="/materials-and-care" className="hover:text-reef-coral">
              Materials &amp; care
            </Link>
            <Link to="/cart" className="hover:text-reef-coral">
              Cart{count > 0 ? ` (${count})` : ''}
            </Link>
          </div>
        </nav>
      </header>
      <main className="flex-1 max-w-5xl mx-auto w-full px-4 py-8">{children}</main>
      <footer className="border-t border-reef-teal/20 px-4 py-6 text-xs text-reef-ink/60">
        <div className="max-w-5xl mx-auto">
          PETG parts, made to order. Not certified food-grade or laboratory-grade — see{' '}
          <Link to="/materials-and-care" className="underline">
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
