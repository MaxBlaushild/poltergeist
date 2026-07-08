export const HOUSE_ACCENT: Record<string, string> = {
  Spires: '#ff3b3b', // truer, brighter red for contrast on black
  Chains: '#c89b5a',
  Cinders: '#f0822a',
  Ashglass: '#7bb8b6',
  "Marquess's Court": '#a97ad6',
};

export const TIER_LABEL: Record<string, string> = {
  easy: 'Easy',
  medium: 'Medium',
  hard: 'Hard',
};

export const accentFor = (house?: string) => (house && HOUSE_ACCENT[house]) || '#c81912';

// House mottos, shown on Standings and the Summons house list.
export const HOUSE_TAGLINE: Record<string, string> = {
  Spires: 'Order is power',
  Chains: 'Secrets are power',
  Cinders: 'Passion is power',
  Ashglass: 'Knowledge is power',
  "Marquess's Court": 'Power is power',
};

export const taglineFor = (house?: string) => (house && HOUSE_TAGLINE[house]) || '';

// Display label for a house. Drops "of" for the named houses ("House Ashglass"),
// but keeps it for the Marquess's Court ("House of Marquess's Court").
export const houseLabel = (house?: string) => {
  if (!house) return 'House';
  return house === "Marquess's Court" ? `House of ${house}` : `House ${house}`;
};

// House Favor can be fractional (Part 2 quiz). Show e.g. "2.5", "10.5", "3".
export const formatHF = (n: number) => String(Math.round(n * 100) / 100);

export interface HouseInfo {
  sigil: string; // path under public/, e.g. /houses/spires.png
  blurb: string;
}

// House lore + sigil. Edit the blurbs here as the worldbuilding firms up.
export const HOUSE_INFO: Record<string, HouseInfo> = {
  Spires: {
    sigil: '/houses/spires.png',
    blurb:
      'Bloodline and martial pride. Spires breeds challengers and champions — its claim to the Toast rests on ancient lineage and victory in the arena.',
  },
  Chains: {
    sigil: '/houses/chains.png',
    blurb:
      'Favors, debts, and leverage. Chains binds the Court together, trading in secrets owed and promises that cannot be broken.',
  },
  Cinders: {
    sigil: '/houses/cinders.png',
    blurb:
      'Fire, ash, and survival. Cinders is the house of duelists and the forge-born — those who have walked through flame and returned.',
  },
  Ashglass: {
    sigil: '/houses/ashglass.png',
    blurb:
      'Blood, glass, and memory. Ashglass keeps the archives and runs the labs, studying what the others would rather forget.',
  },
  "Marquess's Court": {
    sigil: '/houses/court.png',
    blurb:
      "The host's own circle. The Court pours the wine, keeps the rules, and watches over every Crimson Toast.",
  },
};

export const houseInfoFor = (house?: string): HouseInfo =>
  (house && HOUSE_INFO[house]) || { sigil: '', blurb: '' };
