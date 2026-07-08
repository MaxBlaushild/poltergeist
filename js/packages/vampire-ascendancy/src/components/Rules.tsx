// Static player-facing rules, split into two screens: "How to Play" (the shape of
// the game) and "Earn & Spend" (the economy tables). Content is authored copy, so
// these are purely presentational.

const Heading = ({ children }: { children: React.ReactNode }) => (
  <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-3">{children}</h2>
);

const Card = ({ children }: { children: React.ReactNode }) => (
  <section className="mt-5 rounded-lg border border-blood/40 bg-black/40 p-5">{children}</section>
);

const TLDR = ({ children }: { children: React.ReactNode }) => (
  <p className="mt-6 rounded-md border border-gold/40 bg-gold/5 p-4 text-bone/90 text-sm leading-relaxed">
    <span className="text-gold font-semibold uppercase tracking-[0.2em] text-xs mr-2">TL;DR</span>
    {children}
  </p>
);

export const HowToPlay = () => (
  <div className="pb-8">
    <header className="text-center mb-6">
      <p className="text-xs uppercase tracking-[0.4em] text-gold">The Crimson Toast</p>
      <h1 className="mt-3 font-display text-3xl font-bold text-bone">How to Play</h1>
    </header>

    <Card>
      <Heading>The Goal</Heading>
      <p className="text-bone leading-relaxed">
        Your house competes to win the Crimson Throne. You compete to earn the most Blood Tokens and
        solve the mystery of the missing heir.
      </p>
    </Card>

    <Card>
      <Heading>Two Currencies</Heading>
      <div className="flex flex-col gap-4">
        <div>
          <p className="text-bone font-semibold">🩸 Blood Tokens (BT)</p>
          <p className="text-bone/80 leading-relaxed">
            Physical vials. Your personal money <span className="text-bone/50">+</span> score. Start
            with 10 <span className="text-bone/60">(Cinders: 11)</span>.
          </p>
        </div>
        <div>
          <p className="text-bone font-semibold">👑 House Favor (HF)</p>
          <p className="text-bone/80 leading-relaxed">
            Digital, on the live leaderboard. Your house's standing. Most HF at the end wins the
            throne.
          </p>
        </div>
      </div>
    </Card>

    <Card>
      <Heading>The Flow</Heading>
      <ul className="flex flex-col gap-2 text-bone/85 leading-relaxed list-disc pl-5">
        <li>Earn by playing games, completing your character's missions, and acing the closing quiz.</li>
        <li>Spend BT on clues (Wall of Whispers) and relics (Reliquary).</li>
        <li>
          You choose what to share — your secrets, clues, and knowledge are yours to reveal, trade,
          or hide.
        </li>
        <li>Pool your BT with housemates for big-ticket clues. Trade relics and info with anyone, anytime.</li>
      </ul>
    </Card>

    <Card>
      <Heading>The Mystery</Heading>
      <p className="text-bone leading-relaxed">
        The heir is missing and something is wrong in the Court. Gather clues, talk to everyone,
        trust no one. Solving it pays off — the closing quiz rewards the sharp-eyed.
      </p>
    </Card>

    <TLDR>
      Win games for your house. Do your missions for BT. Buy clues, solve the mystery, don't get
      played.
    </TLDR>
  </div>
);

// A small responsive table that scrolls horizontally on narrow screens rather
// than clipping.
const Table = ({ head, rows }: { head: string[]; rows: (string | number)[][] }) => (
  <div className="overflow-x-auto">
    <table className="w-full text-sm border-collapse">
      <thead>
        <tr>
          {head.map((h, i) => (
            <th
              key={i}
              className={`border-b border-blood/40 pb-2 font-heading uppercase tracking-[0.15em] text-xs text-gold ${
                i === 0 ? 'text-left' : 'text-right'
              }`}
            >
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((r, ri) => (
          <tr key={ri} className="border-b border-blood/15 last:border-0">
            {r.map((c, ci) => (
              <td
                key={ci}
                className={`py-2 ${ci === 0 ? 'text-left text-bone' : 'text-right text-bone/80'}`}
              >
                {c}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);

export const EarnSpend = () => (
  <div className="pb-8">
    <header className="text-center mb-6">
      <p className="text-xs uppercase tracking-[0.4em] text-gold">The Crimson Toast</p>
      <h1 className="mt-3 font-display text-3xl font-bold text-bone">Earn &amp; Spend</h1>
    </header>

    <p className="text-center text-blood-bright uppercase tracking-[0.3em] text-xs mb-2">
      🩸 Earn Blood Tokens
    </p>

    <Card>
      <Heading>Games</Heading>
      <p className="text-bone/70 text-sm mb-3">
        ~13 tonight — most are 1 player per house; Flip Cup is teams of 4.
      </p>
      <Table
        head={['Place', 'BT', 'HF']}
        rows={[
          ['1st', '+5', '+5'],
          ['2nd', '+3', '+3'],
          ['3rd', '+1', '+2'],
          ['Play', '+1', '—'],
        ]}
      />
    </Card>

    <Card>
      <Heading>Your Missions</Heading>
      <p className="text-bone/70 text-sm mb-3">
        On your character card — submit in-app, collect at the Blood Bank.
      </p>
      <Table
        head={['Tier', 'BT']}
        rows={[
          ['Easy', 2],
          ['Medium', 4],
          ['Hard', 7],
        ]}
      />
    </Card>

    <Card>
      <Heading>Closing Quiz</Heading>
      <p className="text-bone leading-relaxed">
        Accurate answers earn BT, with a bonus for the top solver.
      </p>
    </Card>

    <p className="text-center text-blood-bright uppercase tracking-[0.3em] text-xs mt-8 mb-2">
      🩸 Spend Blood Tokens
    </p>

    <Card>
      <Heading>Wall of Whispers</Heading>
      <p className="text-bone/70 text-sm mb-3">Clues — higher tier = bigger reveal.</p>
      <Table
        head={['Tier', 'Cost']}
        rows={[
          ['T1', 2],
          ['T2', 4],
          ['T3', 8],
          ['T4', 15],
          ['T5', '30 · pool with your house'],
        ]}
      />
    </Card>

    <Card>
      <Heading>Reliquary</Heading>
      <p className="text-bone leading-relaxed">
        Relics — buy objects &amp; documents tied to the mystery. Some are one-of-a-kind.
      </p>
    </Card>

    <Card>
      <Heading>Vampire's Wager</Heading>
      <p className="text-bone leading-relaxed">Bet BT/HF house-vs-house at set moments.</p>
    </Card>

    <TLDR>
      Games = house glory + a little BT. Missions = your main BT. Quiz = bonus BT. Spend it on clues
      and relics.
    </TLDR>
  </div>
);
