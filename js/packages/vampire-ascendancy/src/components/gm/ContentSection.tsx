import { useEffect, useState } from 'react';
import {
  gmListLibraryCharacters,
  gmSetCharacterIncluded,
  gmListLibraryItems,
  gmSetItemIncluded,
  gmListLibraryQuizQuestions,
  gmSetQuizQuestionIncluded,
} from '../../gmApi';
import type { GMLibraryCharacter, GMLibraryItem, GMLibraryQuizQuestion } from '../../gmApi';
import { ApiError } from '../../api';
import { Card } from './GameSection';

// Content tab: choose which of the shared story's characters, items, and
// quiz questions are included in this Toast. Toggling off something that's
// still in active use is blocked (409) rather than silently allowed — the
// error names exactly what's blocking it.
export const ContentSection = () => {
  return (
    <div className="flex flex-col gap-4">
      <p className="text-bone/50 text-sm">
        Trim the roster to your guest count. Everything starts included; toggling something off that's
        still assigned to a player (or already answered, for quiz questions) is blocked — reassign it
        first.
      </p>
      <CharactersPanel />
      <ItemsPanel />
      <QuizPanel />
    </div>
  );
};

const CharactersPanel = () => {
  const [rows, setRows] = useState<GMLibraryCharacter[] | null>(null);
  const [error, setError] = useState<Record<string, string>>({});

  const load = () => gmListLibraryCharacters().then((d) => setRows(d.characters));
  useEffect(() => {
    load();
  }, []);

  const toggle = async (row: GMLibraryCharacter) => {
    setError((e) => ({ ...e, [row.id]: '' }));
    try {
      await gmSetCharacterIncluded(row.id, !row.included);
      load();
    } catch (err) {
      setError((e) => ({ ...e, [row.id]: err instanceof ApiError ? err.message : 'Could not save.' }));
    }
  };

  if (!rows) return <Card title="Characters">Loading…</Card>;
  const included = rows.filter((r) => r.included).length;

  return (
    <Card title={`Characters (${included}/${rows.length} included)`}>
      <div className="flex flex-col gap-1.5">
        {rows.map((r) => (
          <div key={r.id} className="rounded-md border border-blood/30 bg-black/30 px-3 py-2">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="text-bone truncate">
                  {r.name}
                  {r.isOptional && <span className="text-gold/70 ml-1">✦</span>}
                  <span className="text-bone/40 mx-2">•</span>
                  <span className="text-bone/50 text-sm">{r.houseName || r.roleType}</span>
                </p>
              </div>
              <label className="shrink-0 flex items-center gap-2 text-sm text-bone/70">
                <input type="checkbox" checked={r.included} onChange={() => toggle(r)} />
                Included
              </label>
            </div>
            {error[r.id] && <p className="text-blood-bright text-xs mt-1">{error[r.id]}</p>}
          </div>
        ))}
      </div>
    </Card>
  );
};

const ItemsPanel = () => {
  const [rows, setRows] = useState<GMLibraryItem[] | null>(null);
  const [error, setError] = useState<Record<string, string>>({});

  const load = () => gmListLibraryItems().then((d) => setRows(d.items));
  useEffect(() => {
    load();
  }, []);

  const toggle = async (row: GMLibraryItem) => {
    setError((e) => ({ ...e, [row.id]: '' }));
    try {
      await gmSetItemIncluded(row.id, !row.included);
      load();
    } catch (err) {
      setError((e) => ({ ...e, [row.id]: err instanceof ApiError ? err.message : 'Could not save.' }));
    }
  };

  if (!rows) return <Card title="Items">Loading…</Card>;
  const included = rows.filter((r) => r.included).length;

  return (
    <Card title={`Items (${included}/${rows.length} included)`}>
      <div className="flex flex-col gap-1.5 max-h-96 overflow-y-auto">
        {rows.map((r) => (
          <div key={r.id} className="rounded-md border border-blood/30 bg-black/30 px-3 py-2">
            <div className="flex items-center justify-between gap-3">
              <p className="text-bone truncate">
                {r.name}
                {r.category && <span className="text-bone/40 text-xs ml-2">{r.category}</span>}
              </p>
              <label className="shrink-0 flex items-center gap-2 text-sm text-bone/70">
                <input type="checkbox" checked={r.included} onChange={() => toggle(r)} />
                Included
              </label>
            </div>
            {error[r.id] && <p className="text-blood-bright text-xs mt-1">{error[r.id]}</p>}
          </div>
        ))}
      </div>
    </Card>
  );
};

const QuizPanel = () => {
  const [rows, setRows] = useState<GMLibraryQuizQuestion[] | null>(null);
  const [error, setError] = useState<Record<string, string>>({});

  const load = () => gmListLibraryQuizQuestions().then((d) => setRows(d.questions));
  useEffect(() => {
    load();
  }, []);

  const toggle = async (row: GMLibraryQuizQuestion) => {
    setError((e) => ({ ...e, [row.id]: '' }));
    try {
      await gmSetQuizQuestionIncluded(row.id, !row.included);
      load();
    } catch (err) {
      setError((e) => ({ ...e, [row.id]: err instanceof ApiError ? err.message : 'Could not save.' }));
    }
  };

  if (!rows) return <Card title="Quiz questions">Loading…</Card>;
  const included = rows.filter((r) => r.included).length;

  return (
    <Card title={`Quiz questions (${included}/${rows.length} included)`}>
      <div className="flex flex-col gap-1.5">
        {rows.map((r) => (
          <div key={r.id} className="rounded-md border border-blood/30 bg-black/30 px-3 py-2">
            <div className="flex items-center justify-between gap-3">
              <p className="text-bone truncate">
                <span className="text-bone/40 text-xs mr-2">Part {r.part}</span>
                {r.prompt || '(untitled)'}
              </p>
              <label className="shrink-0 flex items-center gap-2 text-sm text-bone/70">
                <input type="checkbox" checked={r.included} onChange={() => toggle(r)} />
                Included
              </label>
            </div>
            {error[r.id] && <p className="text-blood-bright text-xs mt-1">{error[r.id]}</p>}
          </div>
        ))}
      </div>
    </Card>
  );
};
