import { useEffect, useState } from 'react';
import { gmListSubmissions, gmVerify, gmReject } from '../../gmApi';
import type { GMSubmission } from '../../gmApi';
import { photoUrl } from '../../api';
import { Card } from './GameSection';

const TIER_LABEL: Record<string, string> = { easy: 'Easy', medium: 'Medium', hard: 'Hard' };

export const SubmissionsSection = () => {
  const [subs, setSubs] = useState<GMSubmission[]>([]);
  const [filter, setFilter] = useState('submitted');
  const [loading, setLoading] = useState(true);

  const load = () => {
    setLoading(true);
    gmListSubmissions(filter)
      .then((d) => setSubs(d.submissions || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  };
  // Poll the queue so newly submitted answers appear for the GMs.
  useEffect(() => {
    load();
    const id = setInterval(load, 5000);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter]);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-2">
        {['submitted', 'verified', 'rejected', ''].map((f) => (
          <button
            key={f || 'all'}
            onClick={() => setFilter(f)}
            className={`px-3 py-1.5 rounded-md text-xs uppercase tracking-[0.15em] ${
              filter === f ? 'bg-blood text-bone' : 'border border-blood/40 text-bone/60'
            }`}
          >
            {f || 'all'}
          </button>
        ))}
      </div>

      {loading && subs.length === 0 ? (
        <p className="text-bone/50">Loading…</p>
      ) : subs.length === 0 ? (
        <p className="text-bone/50">Nothing here.</p>
      ) : (
        subs.map((s) => <SubmissionCard key={s.id} sub={s} onChange={load} />)
      )}
    </div>
  );
};

const SubmissionCard = ({ sub, onChange }: { sub: GMSubmission; onChange: () => void }) => {
  const [bt, setBt] = useState(String(sub.rewardBt));
  const [busy, setBusy] = useState(false);

  const verify = async () => {
    setBusy(true);
    try {
      await gmVerify(sub.id, Number(bt));
      onChange();
    } finally {
      setBusy(false);
    }
  };
  const reject = async () => {
    setBusy(true);
    try {
      await gmReject(sub.id);
      onChange();
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card title={`${sub.characterName} · ${sub.houseName}`}>
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs uppercase tracking-[0.2em] text-bone/60">
          {TIER_LABEL[sub.missionTier] || sub.missionTier} · {sub.rewardBt} BT
        </span>
        <StatusPill status={sub.status} />
      </div>
      <p className="text-bone/80 text-sm mb-1">{sub.missionPrompt}</p>
      {sub.missionAnswerFormat && (
        <p className="text-gold/90 text-xs italic mb-2">Asked of them: {sub.missionAnswerFormat}</p>
      )}
      <p className="text-bone bg-black/50 rounded-md p-3 mb-3 whitespace-pre-wrap">
        {sub.playerAnswer || <span className="text-bone/40">— no answer —</span>}
      </p>
      {sub.photoIds && sub.photoIds.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-3">
          {sub.photoIds.map((id) => (
            <a key={id} href={photoUrl(id)} target="_blank" rel="noreferrer">
              <img
                src={photoUrl(id)}
                alt=""
                className="w-20 h-20 object-cover rounded-md border border-blood/40"
              />
            </a>
          ))}
        </div>
      )}

      {sub.status !== 'verified' && (
        <div className="flex items-center gap-2">
          <label className="text-xs text-bone/50">BT</label>
          <input
            value={bt}
            onChange={(e) => setBt(e.target.value.replace(/[^0-9-]/g, ''))}
            className="w-16 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-center"
          />
          <button
            onClick={verify}
            disabled={busy}
            className="flex-1 py-2 rounded-md bg-green-700/70 text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Verify
          </button>
          <button
            onClick={reject}
            disabled={busy}
            className="flex-1 py-2 rounded-md border border-blood/50 text-blood-bright uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Reject
          </button>
        </div>
      )}
      {sub.status === 'verified' && (
        <p className="text-green-400 text-sm">Verified · {sub.awardedBt} BT</p>
      )}
    </Card>
  );
};

const StatusPill = ({ status }: { status: string }) => {
  const map: Record<string, string> = {
    submitted: 'text-amber-300',
    verified: 'text-green-400',
    rejected: 'text-blood-bright',
  };
  return <span className={`text-xs uppercase tracking-[0.15em] ${map[status] || ''}`}>{status}</span>;
};
