// A small gothic crest of two fangs with a drop of blood, used on holding and
// error screens to keep the vampiric tone even when something has gone wrong.
export const VampireMark = ({ className = '' }: { className?: string }) => (
  <svg
    viewBox="0 0 80 80"
    className={className}
    role="img"
    aria-label="vampire fangs"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* upper lip / arch */}
    <path
      d="M10 26 Q40 44 70 26"
      stroke="var(--bone)"
      strokeWidth="2.5"
      strokeLinecap="round"
      opacity="0.45"
    />
    {/* small teeth between the fangs */}
    <path d="M33 31 l3 7 3-7" stroke="var(--bone)" strokeWidth="1.5" opacity="0.4" fill="none" />
    <path d="M41 31 l3 7 3-7" stroke="var(--bone)" strokeWidth="1.5" opacity="0.4" fill="none" />
    {/* left fang */}
    <path d="M24 30 L31 30 Q28 52 27.5 52 Q27 52 24 30 Z" fill="var(--bone)" />
    {/* right fang */}
    <path d="M49 30 L56 30 Q53 52 52.5 52 Q52 52 49 30 Z" fill="var(--bone)" />
    {/* drop of blood from the left fang */}
    <path
      d="M27.5 56 q4 5 0 9 q-4 -4 0 -9 Z"
      fill="var(--blood-bright)"
    />
  </svg>
);
