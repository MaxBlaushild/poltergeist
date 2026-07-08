import { useState } from 'react';

// "How tonight works." A single scrollable page: the intro (TL;DR + goal) is
// pinned at the top; everything else is a collapsed accordion in two groups.
// Content is authored copy from the player rules doc — purely presentational.
export const Tournament = () => (
  <div className="pb-8">
    <header className="text-center mb-6">
      <p className="text-xs uppercase tracking-[0.4em] text-gold">Tonight's Tournament</p>
      <h1 className="mt-3 font-display text-3xl font-bold text-bone">How Tonight Works</h1>
    </header>

    {/* Pinned intro */}
    <div className="rounded-lg border border-gold/40 bg-gold/5 p-5">
      <h2 className="font-heading text-gold text-xs uppercase tracking-[0.3em] mb-2">The Goal</h2>
      <p className="text-bone/90 leading-relaxed">
        Tonight the Marquess names her successor, and every house wants the crown. Winning takes two
        things: your house must finish with the most House Favor, and you must finish with more Blood
        Tokens than anyone else in it. Help your house to the top — but hold enough back to claim the
        throne yourself.
      </p>
    </div>

    <GroupLabel>The Rules</GroupLabel>

    <Accordion title="The two currencies">
      <div className="flex flex-col gap-3">
        <p className="text-bone/85 leading-relaxed">
          <span className="text-bone font-semibold">🩸 Blood Tokens (BT)</span> — physical vials.
          Your personal money and your score. Earn them, spend them, or hold onto as many as you can.
        </p>
        <p className="text-bone/85 leading-relaxed">
          <span className="text-bone font-semibold">👑 House Favor (HF)</span> — tracked on the live
          leaderboard. Your house's collective standing. The house with the most House Favor wins the
          right to the throne.
        </p>
      </div>
    </Accordion>

    <Accordion title="How it fits together">
      <ul className="flex flex-col gap-2 text-bone/85 leading-relaxed list-disc pl-5">
        <li>Earn Blood Tokens by winning games, completing your character's missions, and doing well on the closing quiz (open-end).</li>
        <li>Earn House Favor for your house by winning games, buying select relics, and doing well on the closing quiz (multiple choice).</li>
        <li>Spend Blood Tokens on clues at the Wall of Whispers and relics at the Reliquary.</li>
        <li>You decide what to share. Your secrets, clues, and knowledge are yours — reveal them to make allies, mislead rivals, or complete missions.</li>
        <li>Trade, gift, or pool Blood Tokens as needed, within your house and across houses. House Favor cannot be traded, and is rarely taken away.</li>
      </ul>
    </Accordion>

    <Accordion title="The mystery">
      <p className="text-bone/85 leading-relaxed">
        The heir is missing and something has gone wrong in the Court. Gather clues, talk to
        everyone, and trust no one. You don't have to solve it — but those who do are rewarded, and
        the closing quiz favors the sharp-eyed.
      </p>
    </Accordion>

    <GroupLabel>Earn &amp; Spend</GroupLabel>

    <Accordion title="Games">
      <p className="text-bone/85 leading-relaxed mb-3">
        Compete in games to earn both Blood Tokens and House Favor for placing well.
      </p>
      <Table
        head={['Place', 'Blood Tokens', 'House Favor']}
        rows={[
          ['1st', '+5', '+5'],
          ['2nd', '+3', '+3'],
          ['3rd', '+1', '+2'],
          ['Participate', '+1', '—'],
        ]}
      />
    </Accordion>

    <Accordion title="Your missions">
      <p className="text-bone/85 leading-relaxed mb-3">
        Every character has private missions on their card — small tasks tied to your role and the
        mystery. Complete one, submit it in the app, and collect your Blood Tokens from the Blood
        Bank.
      </p>
      <Table
        head={['Difficulty', 'Blood Tokens']}
        rows={[
          ['Easy', 2],
          ['Medium', 4],
          ['Hard', 7],
        ]}
      />
    </Accordion>

    <Accordion title="The closing quiz">
      <p className="text-bone/85 leading-relaxed">
        At the end of the night, a quiz tests how much of the mystery you pieced together. Accurate
        answers earn Blood Tokens (on the open-end) and House Favor (on the multiple choice) — so the
        more you investigate tonight, the better you'll do.
      </p>
    </Accordion>

    <Accordion title="The Wall of Whispers">
      <p className="text-gold/90 italic mb-2">Run by Eiran Vox.</p>
      <p className="text-bone/85 leading-relaxed">
        A physical wall where secrets are sold. Each clue has a tier and a price marked on the wall —
        the more revealing the clue, the more it costs, running anywhere from 2 to 30 Blood Tokens.
        Pool Blood Tokens with fellow players if you need to afford one. Buy a clue, and it's yours to
        use, share, or trade.
      </p>
    </Accordion>

    <Accordion title="The Reliquary">
      <p className="text-gold/90 italic mb-2">Run by Ivara Saye.</p>
      <p className="text-bone/85 leading-relaxed">
        Where objects, documents, and artifacts tied to the mystery are kept. You'll see each relic's
        name, price, and a short description — but not what's inside until after you buy it. Some are
        one-of-a-kind; once claimed, they're gone for good.
      </p>
    </Accordion>
  </div>
);

const GroupLabel = ({ children }: { children: React.ReactNode }) => (
  <h2 className="mt-8 mb-2 px-1 font-heading text-bone/50 text-xs uppercase tracking-[0.35em]">
    {children}
  </h2>
);

const Accordion = ({ title, children }: { title: string; children: React.ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <div className="mt-2 rounded-lg border border-blood/30 bg-black/40 overflow-hidden">
      <button
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        className="w-full flex items-center gap-2 px-4 py-3.5 text-left"
      >
        <span className={`text-gold transition-transform ${open ? 'rotate-90' : ''}`}>▸</span>
        <span className="text-bone font-medium">{title}</span>
      </button>
      {open && <div className="px-4 pb-4 pt-0">{children}</div>}
    </div>
  );
};

const Table = ({ head, rows }: { head: string[]; rows: (string | number)[][] }) => (
  <div className="overflow-x-auto">
    <table className="w-full text-sm border-collapse">
      <thead>
        <tr>
          {head.map((h, i) => (
            <th
              key={i}
              className={`border-b border-blood/40 pb-2 font-heading uppercase tracking-[0.12em] text-xs text-gold ${
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
              <td key={ci} className={`py-2 ${ci === 0 ? 'text-left text-bone' : 'text-right text-bone/80'}`}>
                {c}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);
