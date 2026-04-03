export type AdminNavItem = {
  id: string;
  label: string;
  path: string;
  description: string;
  matchPrefixes?: string[];
};

export type AdminNavGroup = {
  id: string;
  label: string;
  description: string;
  items: AdminNavItem[];
};

export const adminNavigationGroups: AdminNavGroup[] = [
  {
    id: 'world',
    label: 'World',
    description: 'Map structure, places, tags, and how districts fit together.',
    items: [
      {
        id: 'zones',
        label: 'Zones',
        path: '/zones',
        description:
          'Manage zone boundaries, map data, and zone-level content.',
      },
      {
        id: 'districts',
        label: 'Districts',
        path: '/districts',
        description:
          'Group zones into higher-level neighborhoods and seed them.',
      },
      {
        id: 'zone-tagging',
        label: 'Zone Tagging',
        path: '/zone-tagging',
        description:
          'Queue neighborhood-tag generation jobs and review shared flavor tags for zones.',
      },
      {
        id: 'points-of-interest',
        label: 'Points of Interest',
        path: '/points-of-interest',
        description:
          'Inspect imported places, clues, media, and place metadata.',
        matchPrefixes: ['/place/'],
      },
      {
        id: 'tags',
        label: 'Tags',
        path: '/tags',
        description: 'Define reusable tags across the admin toolchain.',
      },
      {
        id: 'location-archetypes',
        label: 'Location Archetypes',
        path: '/location-archetypes',
        description: 'Create reusable place patterns and challenge prompts.',
      },
    ],
  },
  {
    id: 'questing',
    label: 'Questing',
    description: 'Author templates, generated content, and distribution rules.',
    items: [
      {
        id: 'quest-archetypes',
        label: 'Quest Archetypes',
        path: '/quest-archetypes',
        description: 'Design structured quest templates and node flows.',
      },
      {
        id: 'quest-archetype-generator',
        label: 'Quest Archetype Generator',
        path: '/quest-archetype-generator',
        description:
          'Generate draft archetype bundles, review them, and convert the best ones into live templates.',
      },
      {
        id: 'main-story-generator',
        label: 'Main Story Generator',
        path: '/main-story-generator',
        description:
          'Generate district-scale campaign drafts with ordered beats and convert them into reusable main story templates.',
      },
      {
        id: 'main-story-templates',
        label: 'Main Story Templates',
        path: '/main-story-templates',
        description:
          'Inspect converted campaign templates and instantiate live district quest chains.',
      },
      {
        id: 'zone-quest-archetypes',
        label: 'Zone Quest Archetypes',
        path: '/zone-quest-archetypes',
        description: 'Assign quest archetypes to zones and tune availability.',
      },
      {
        id: 'quests',
        label: 'Quests',
        path: '/quests',
        description:
          'Inspect concrete quests, routes, objectives, and rewards.',
      },
      {
        id: 'scenarios',
        label: 'Scenarios',
        path: '/scenarios',
        description: 'Manage generated and standalone scenario content.',
      },
      {
        id: 'scenario-templates',
        label: 'Scenario Templates',
        path: '/scenario-templates',
        description: 'Author reusable scenario templates for quest generation.',
      },
      {
        id: 'challenges',
        label: 'Challenges',
        path: '/challenges',
        description:
          'Manage explicit challenge objectives and location prompts.',
      },
      {
        id: 'challenge-templates',
        label: 'Challenge Templates',
        path: '/challenge-templates',
        description: 'Create reusable challenge templates and artwork.',
      },
      {
        id: 'zone-seeding',
        label: 'Zone Seeding',
        path: '/zone-seeding',
        description: 'Queue seeding jobs and review generated world content.',
      },
    ],
  },
  {
    id: 'systems',
    label: 'Systems',
    description:
      'Characters, progression, combat content, and reward infrastructure.',
    items: [
      {
        id: 'users',
        label: 'Users',
        path: '/users',
        description: 'Inspect users, progress, unlocks, and live state.',
      },
      {
        id: 'characters',
        label: 'Characters',
        path: '/characters',
        description: 'Author characters, locations, dialogue, and actions.',
      },
      {
        id: 'parties',
        label: 'Parties',
        path: '/parties',
        description:
          'Monitor party state, membership, and cooperative progress.',
      },
      {
        id: 'inventory-items',
        label: 'Inventory Items',
        path: '/inventory-items',
        description: 'Maintain items, equipment, and reward sources.',
      },
      {
        id: 'armory',
        label: 'Armory',
        path: '/armory',
        description: 'Review armory tooling and equipment surfaces.',
      },
      {
        id: 'bases',
        label: 'Bases',
        path: '/bases',
        description: 'Configure base-related admin content and ownership.',
      },
      {
        id: 'spells',
        label: 'Spells',
        path: '/spells',
        description: 'Manage spells, progressions, and ability data.',
      },
      {
        id: 'monsters',
        label: 'Monsters',
        path: '/monsters',
        description: 'Create monsters, encounter templates, and combat groups.',
      },
      {
        id: 'treasure-chests',
        label: 'Treasure Chests',
        path: '/treasure-chests',
        description: 'Place and tune chest rewards across the map.',
      },
      {
        id: 'healing-fountains',
        label: 'Healing Fountains',
        path: '/healing-fountains',
        description: 'Manage healing fountain placements and behavior.',
      },
      {
        id: 'starter-config',
        label: 'Starter Config',
        path: '/starter-config',
        description: 'Tune new-user defaults and starter loadouts.',
      },
      {
        id: 'tutorial',
        label: 'Tutorial',
        path: '/tutorial',
        description:
          'Configure tutorial flows, encounters, and onboarding beats.',
      },
    ],
  },
  {
    id: 'live-ops',
    label: 'Live Ops',
    description: 'Moderation, signals, and operational review tools.',
    items: [
      {
        id: 'feedback',
        label: 'Feedback',
        path: '/feedback',
        description: 'Review player reports, notes, and product feedback.',
      },
      {
        id: 'flagged-photos',
        label: 'Flagged Photos',
        path: '/flagged-photos',
        description: 'Handle moderation queues for submitted media.',
      },
      {
        id: 'insider-trades',
        label: 'Insider Trades',
        path: '/insider-trades',
        description: 'Inspect insider trades and related operational events.',
      },
    ],
  },
  {
    id: 'events',
    label: 'Events',
    description: 'Arena and event-night administration surfaces.',
    items: [
      {
        id: 'arenas',
        label: 'Arenas',
        path: '/arenas',
        description: 'Manage arena groups, teams, and event sessions.',
        matchPrefixes: ['/arena/'],
      },
      {
        id: 'fete-rooms',
        label: 'Fete Rooms',
        path: '/fete-rooms',
        description: 'Edit final fete room state and room-level content.',
      },
      {
        id: 'fete-teams',
        label: 'Fete Teams',
        path: '/fete-teams',
        description: 'Manage final fete teams and participant groupings.',
      },
      {
        id: 'fete-room-teams',
        label: 'Fete Room Teams',
        path: '/fete-room-teams',
        description: 'Configure room-to-team assignments for the event.',
      },
      {
        id: 'fete-room-linked-list-teams',
        label: 'Fete Linked Lists',
        path: '/fete-room-linked-list-teams',
        description: 'Maintain linked-list room ordering for event flows.',
      },
      {
        id: 'utility-closet-puzzle',
        label: 'Utility Closet Puzzle',
        path: '/utility-closet-puzzle',
        description: 'Control the special-case puzzle administration panel.',
      },
    ],
  },
];

export const adminFeaturedNavItemIds = [
  'zones',
  'districts',
  'points-of-interest',
  'quest-archetypes',
  'quests',
  'zone-seeding',
  'users',
  'feedback',
];

const normalizePath = (value: string) =>
  value.length > 1 && value.endsWith('/') ? value.slice(0, -1) : value;

export const flattenAdminNavItems = () =>
  adminNavigationGroups.flatMap((group) =>
    group.items.map((item) => ({
      ...item,
      group,
    }))
  );

export const adminNavItemMatchesPath = (
  item: AdminNavItem,
  pathname: string
) => {
  const normalizedPathname = normalizePath(pathname);
  const normalizedItemPath = normalizePath(item.path);
  if (normalizedPathname === normalizedItemPath) {
    return true;
  }
  if (
    normalizedItemPath !== '/' &&
    normalizedPathname.startsWith(`${normalizedItemPath}/`)
  ) {
    return true;
  }
  return (item.matchPrefixes ?? []).some((prefix) =>
    normalizedPathname.startsWith(prefix)
  );
};

export const findActiveAdminNavItem = (pathname: string) =>
  flattenAdminNavItems().find((item) =>
    adminNavItemMatchesPath(item, pathname)
  );

export const featuredAdminNavItems = adminFeaturedNavItemIds
  .map((id) => flattenAdminNavItems().find((item) => item.id === id))
  .filter((item): item is ReturnType<typeof flattenAdminNavItems>[number] =>
    Boolean(item)
  );
