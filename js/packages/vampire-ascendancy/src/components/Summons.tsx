import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getLeaderboard, getToken } from '../api';
import type { HouseStanding } from '../types';
import { accentFor, taglineFor, houseLabel } from '../theme';

// "Why you've been called." Static in-world framing, plus a tappable list of the
// Great Houses (deep-linking to each house page) so the lore is discoverable.
export const Summons = () => {
  const [houses, setHouses] = useState<HouseStanding[] | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    let cancelled = false;
    // Houses don't change during play — a single fetch is enough for the links.
    getLeaderboard(token)
      .then((d) => !cancelled && setHouses(d.standings))
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="pb-8">
      <header className="text-center mb-6">
        <p className="text-xs uppercase tracking-[0.4em] text-gold">The Summons</p>
        <h1 className="mt-3 font-display text-3xl font-bold text-bone">You Have Been Called</h1>
      </header>

      <Section title="Legacy of the Crimson Toast">
        <P>
          One year ago, the Great Houses assembled for the annual Crimson Toast, an evening meant to
          celebrate the Court's enduring peace and prosperity. Before the first toast could be raised,
          however, Valen Drear of House Spires was found dead.
        </P>
        <P>
          What began as a murder investigation soon unraveled into something far more complicated. As
          the Court sifted through lies, betrayals, and long-buried rivalries, the truth finally
          emerged: Serel Nox of House Ashglass had attempted to assassinate Thorne Virell, believing
          him to be the greatest threat to Ashglass's future. Instead, through a tragic mistake, the
          poison claimed Valen's life.
        </P>
        <P>
          Though the mystery was solved, the consequences lingered. House Ashglass's reputation was
          left in tatters, old grudges deepened, and whispers spread of forbidden bloodlines and
          forgotten prophecies lurking beneath the surface of vampire society.
        </P>
        <P className="text-bone/90 italic">
          The Court survived the Crimson Toast — but it did not emerge unchanged.
        </P>
      </Section>

      <Section title="This Year's Invitation">
        <P>Now, Marquess Gruber has issued another summons.</P>
        <P>
          For the first time in generations, she will publicly name her heir, ushering in a new era
          for the vampire court. Nobles from each of the Great Houses have gathered to witness the
          succession — but not everyone intends for the ceremony to proceed as planned.
        </P>
        <P>Some seek prestige. Others seek vengeance. Many have come to advance their own secret ambitions.</P>
        <P>
          Throughout the evening, you'll forge alliances, uncover hidden truths, complete secret
          missions, and compete in challenges to earn influence for yourself and your House. Every
          conversation may reveal a clue. Every favor has a price. Every secret can become a weapon.
        </P>
        <P className="text-bone/90 italic">
          Whether you leave the Ascendancy Ball celebrated, disgraced, or destroyed is entirely up to
          you.
        </P>
      </Section>

      <Section title="The Great Houses">
        <P>
          Every vampire belongs to one of the five Great Houses, each with its own history, values,
          and ambitions. Select a House below to learn more:
        </P>
        <div className="mt-4 flex flex-col gap-2">
          {houses
            ? houses.map((h) => {
                const accent = accentFor(h.name);
                return (
                  <Link
                    key={h.houseId}
                    to={`/house/${h.houseId}`}
                    className="flex items-center gap-3 rounded-lg border border-blood/30 bg-black/40 p-4 transition-colors hover:bg-white/5"
                  >
                    <span className="w-1.5 self-stretch rounded-full" style={{ backgroundColor: accent }} />
                    <div className="flex-1">
                      <p className="font-semibold" style={{ color: accent }}>
                        {houseLabel(h.name)}
                      </p>
                      {taglineFor(h.name) && (
                        <p className="text-xs uppercase tracking-[0.25em] text-bone/50 italic">
                          {taglineFor(h.name)}
                        </p>
                      )}
                    </div>
                    <span className="text-bone/40 text-lg">›</span>
                  </Link>
                );
              })
            : <p className="text-bone/50 text-sm">Summoning the houses…</p>}
        </div>
      </Section>
    </div>
  );
};

const Section = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section className="mt-6 rounded-lg border border-blood/30 bg-black/30 p-5">
    <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-3">{title}</h2>
    {children}
  </section>
);

const P = ({ children, className = '' }: { children: React.ReactNode; className?: string }) => (
  <p className={`text-bone/85 leading-relaxed mb-3 last:mb-0 ${className}`}>{children}</p>
);
