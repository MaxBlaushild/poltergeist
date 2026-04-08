import React from 'react';
import { Link } from 'react-router-dom';
import {
  adminNavigationGroups,
  featuredAdminNavItems,
} from '../adminNavigation.ts';

const workflowHighlights = [
  {
    title: 'Shape the World',
    description:
      'Work from the map outward: define zones, group them into districts, and clean up the places inside them.',
    links: ['zones', 'districts', 'points-of-interest', 'location-archetypes'],
  },
  {
    title: 'Author Quest Content',
    description:
      'Design the reusable quest machinery first, then inspect the concrete quests it generates.',
    links: ['quest-archetypes', 'zone-quest-archetypes', 'quests', 'zone-seeding'],
  },
  {
    title: 'Build Encounter Content',
    description:
      'Author the player-facing encounters and discoveries that show up directly on the map.',
    links: ['monsters', 'scenarios', 'expositions', 'challenges'],
  },
  {
    title: 'Run Systems and Live Ops',
    description:
      'Monitor players, progression, and moderation surfaces without hunting through the whole admin.',
    links: ['users', 'characters', 'parties', 'feedback'],
  },
];

const navItemById = new Map(
  adminNavigationGroups.flatMap((group) =>
    group.items.map((item) => [item.id, item] as const)
  )
);

export const AdminHome = () => {
  return (
    <div className="dashboard-home">
      <section className="dashboard-hero">
        <div className="dashboard-hero__copy">
          <div className="dashboard-kicker">Unclaimed Streets Admin</div>
          <h1>Control center for worldbuilding, quest design, and live operations.</h1>
          <p>
            Start from the map, jump into quest systems, or head straight to live
            player tooling. The dashboard below mirrors the new navigation so the
            whole admin surface is easier to scan.
          </p>
          <div className="dashboard-hero__actions">
            {featuredAdminNavItems.slice(0, 4).map((item) => (
              <Link key={item.id} to={item.path} className="dashboard-hero__button">
                {item.label}
              </Link>
            ))}
          </div>
        </div>
        <div className="dashboard-hero__panel">
          <div className="dashboard-hero__panel-kicker">Suggested Starts</div>
          <div className="dashboard-hero__panel-grid">
            {featuredAdminNavItems.slice(4, 8).map((item) => (
              <Link key={item.id} to={item.path} className="dashboard-hero__mini-card">
                <strong>{item.label}</strong>
                <span>{item.description}</span>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="dashboard-section">
        <div className="dashboard-section__header">
          <div className="dashboard-kicker">Workflows</div>
          <h2>Jump in by job, not by route list.</h2>
        </div>
        <div className="dashboard-workflow-grid">
          {workflowHighlights.map((workflow) => (
            <div key={workflow.title} className="dashboard-card dashboard-card--workflow">
              <h3>{workflow.title}</h3>
              <p>{workflow.description}</p>
              <div className="dashboard-chip-list">
                {workflow.links
                  .map((id) => navItemById.get(id))
                  .filter(Boolean)
                  .map((item) => (
                    <Link key={item!.id} to={item!.path} className="dashboard-chip">
                      {item!.label}
                    </Link>
                  ))}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="dashboard-section">
        <div className="dashboard-section__header">
          <div className="dashboard-kicker">All Areas</div>
          <h2>Browse the admin by domain.</h2>
        </div>
        <div className="dashboard-group-grid">
          {adminNavigationGroups.map((group) => (
            <div key={group.id} className="dashboard-card dashboard-card--group">
              <div className="dashboard-card__eyebrow">{group.label}</div>
              <h3>{group.description}</h3>
              <div className="dashboard-link-list">
                {group.items.map((item) => (
                  <Link key={item.id} to={item.path} className="dashboard-link-row">
                    <span>{item.label}</span>
                    <small>{item.description}</small>
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
};
