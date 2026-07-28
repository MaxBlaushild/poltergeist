// Purely decorative dressing shared across pages — a wave-shaped divider,
// drifting bubbles, and morphing color "splash" blobs. All aria-hidden and
// never carry content, so they're safe to sprinkle in without affecting
// a11y or layout logic elsewhere.

export function WaveDivider({ className = '', flip = false }: { className?: string; flip?: boolean }) {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 1440 80"
      preserveAspectRatio="none"
      className={`wave-divider ${flip ? 'rotate-180' : ''} ${className}`}
    >
      <path
        d="M0,32 C240,80 480,0 720,24 C960,48 1200,88 1440,32 L1440,80 L0,80 Z"
        fill="currentColor"
      />
    </svg>
  );
}

const BUBBLES = [
  { left: '6%', size: 10, delay: '0s', duration: '7s' },
  { left: '18%', size: 18, delay: '1.2s', duration: '9s' },
  { left: '32%', size: 8, delay: '2.4s', duration: '6s' },
  { left: '48%', size: 14, delay: '0.6s', duration: '8s' },
  { left: '64%', size: 22, delay: '3s', duration: '10s' },
  { left: '78%', size: 10, delay: '1.8s', duration: '7.5s' },
  { left: '90%', size: 16, delay: '2.6s', duration: '8.5s' },
];

export function Bubbles({ className = '' }: { className?: string }) {
  return (
    <div aria-hidden="true" className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`}>
      {BUBBLES.map((b, i) => (
        <span
          key={i}
          className="bubble animate-rise"
          style={{
            left: b.left,
            bottom: '-2rem',
            width: b.size,
            height: b.size,
            animationDelay: b.delay,
            animationDuration: b.duration,
          }}
        />
      ))}
    </div>
  );
}

const SPLASH_COLORS = ['bg-reef-coral', 'bg-reef-glow', 'bg-reef-sky'];

interface BlobSpec {
  color?: string;
  size: number;
  top?: string;
  bottom?: string;
  left?: string;
  right?: string;
  opacity?: number;
  animation?: string;
  delay?: string;
}

// A scattered handful of big, soft, morphing color splashes — the primary
// "bouncy ocean color on white" decorative motif. Pass explicit blobs for
// full control, or rely on the default scattered layout.
const DEFAULT_BLOBS: BlobSpec[] = [
  { color: 'bg-reef-glow', size: 260, top: '-10%', left: '-6%', opacity: 0.35, animation: 'animate-blob-slow' },
  { color: 'bg-reef-coral', size: 200, top: '5%', right: '-4%', opacity: 0.3, animation: 'animate-blob', delay: '1.5s' },
  { color: 'bg-reef-sky', size: 180, bottom: '-12%', left: '20%', opacity: 0.28, animation: 'animate-blob-slow', delay: '0.7s' },
];

export function Blobs({ blobs = DEFAULT_BLOBS, className = '' }: { blobs?: BlobSpec[]; className?: string }) {
  return (
    <div aria-hidden="true" className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`}>
      {blobs.map((b, i) => (
        <span
          key={i}
          className={`blob ${b.color ?? SPLASH_COLORS[i % SPLASH_COLORS.length]} ${b.animation ?? 'animate-blob'}`}
          style={{
            width: b.size,
            height: b.size,
            top: b.top,
            bottom: b.bottom,
            left: b.left,
            right: b.right,
            opacity: b.opacity ?? 0.3,
            animationDelay: b.delay,
          }}
        />
      ))}
    </div>
  );
}
