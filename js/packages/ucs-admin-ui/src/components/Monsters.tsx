import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Spell } from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

type MonsterTemplateRecord = {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  baseStrength: number;
  baseDexterity: number;
  baseConstitution: number;
  baseIntelligence: number;
  baseWisdom: number;
  baseCharisma: number;
  spells: Spell[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
};

type MonsterRewardItem = {
  inventoryItemId: number;
  quantity: number;
  inventoryItem?: {
    id: number;
    name: string;
    imageUrl?: string;
  };
};

type MonsterRecord = {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  templateId?: string;
  template?: MonsterTemplateRecord;
  dominantHandInventoryItemId?: number;
  dominantHandInventoryItem?: InventoryItemLite;
  offHandInventoryItemId?: number;
  offHandInventoryItem?: InventoryItemLite;
  weaponInventoryItemId?: number;
  weaponInventoryItem?: InventoryItemLite;
  level: number;
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
  health: number;
  maxHealth: number;
  mana: number;
  maxMana: number;
  attackDamageMin: number;
  attackDamageMax: number;
  attackSwipesPerAttack: number;
  spells: Spell[];
  rewardExperience: number;
  rewardGold: number;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  itemRewards: MonsterRewardItem[];
};

type InventoryItemLite = {
  id: number;
  name: string;
  imageUrl?: string;
  equipSlot?: string | null;
  handItemCategory?: string | null;
  handedness?: string | null;
  damageMin?: number | null;
  damageMax?: number | null;
  swipesPerAttack?: number | null;
  blockPercentage?: number | null;
  damageBlocked?: number | null;
  spellDamageBonusPercent?: number | null;
};

type ImagePreviewState = {
  url: string;
  alt: string;
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

type BulkMonsterTemplateStatus = {
  jobId: string;
  status: string;
  source?: string;
  totalCount: number;
  createdCount: number;
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt?: string;
};

const defaultMonsterUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered monster. Hidden beast silhouette and warning rune motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

type MonsterTemplateFormState = {
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  baseStrength: string;
  baseDexterity: string;
  baseConstitution: string;
  baseIntelligence: string;
  baseWisdom: string;
  baseCharisma: string;
  spellIds: string[];
  techniqueIds: string[];
};

type MonsterFormItem = {
  inventoryItemId: string;
  quantity: string;
};

type MonsterFormState = {
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  zoneId: string;
  latitude: string;
  longitude: string;
  templateId: string;
  dominantHandInventoryItemId: string;
  offHandInventoryItemId: string;
  level: string;
  rewardExperience: string;
  rewardGold: string;
  itemRewards: MonsterFormItem[];
};

const parseIntSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseOptionalInt = (value: string): number | undefined => {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number.parseInt(trimmed, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const emptyTemplateForm = (): MonsterTemplateFormState => ({
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  baseStrength: '10',
  baseDexterity: '10',
  baseConstitution: '10',
  baseIntelligence: '10',
  baseWisdom: '10',
  baseCharisma: '10',
  spellIds: [],
  techniqueIds: [],
});

const templateFormFromRecord = (
  template: MonsterTemplateRecord
): MonsterTemplateFormState => ({
  name: template.name ?? '',
  description: template.description ?? '',
  imageUrl: template.imageUrl ?? '',
  thumbnailUrl: template.thumbnailUrl ?? '',
  baseStrength: String(template.baseStrength ?? 10),
  baseDexterity: String(template.baseDexterity ?? 10),
  baseConstitution: String(template.baseConstitution ?? 10),
  baseIntelligence: String(template.baseIntelligence ?? 10),
  baseWisdom: String(template.baseWisdom ?? 10),
  baseCharisma: String(template.baseCharisma ?? 10),
  spellIds: (template.spells ?? [])
    .filter((spell) => (spell.abilityType ?? 'spell') !== 'technique')
    .map((spell) => spell.id),
  techniqueIds: (template.spells ?? [])
    .filter((spell) => (spell.abilityType ?? 'spell') === 'technique')
    .map((spell) => spell.id),
});

const templatePayloadFromForm = (form: MonsterTemplateFormState) => ({
  name: form.name.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  baseStrength: parseIntSafe(form.baseStrength, 10),
  baseDexterity: parseIntSafe(form.baseDexterity, 10),
  baseConstitution: parseIntSafe(form.baseConstitution, 10),
  baseIntelligence: parseIntSafe(form.baseIntelligence, 10),
  baseWisdom: parseIntSafe(form.baseWisdom, 10),
  baseCharisma: parseIntSafe(form.baseCharisma, 10),
  spellIds: Array.from(new Set([...form.spellIds, ...form.techniqueIds])),
});

const emptyMonsterForm = (): MonsterFormState => ({
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  zoneId: '',
  latitude: '',
  longitude: '',
  templateId: '',
  dominantHandInventoryItemId: '',
  offHandInventoryItemId: '',
  level: '1',
  rewardExperience: '0',
  rewardGold: '0',
  itemRewards: [],
});

const monsterFormFromRecord = (monster: MonsterRecord): MonsterFormState => ({
  name: monster.name ?? '',
  description: monster.description ?? '',
  imageUrl: monster.imageUrl ?? '',
  thumbnailUrl: monster.thumbnailUrl ?? '',
  zoneId: monster.zoneId ?? '',
  latitude: String(monster.latitude ?? ''),
  longitude: String(monster.longitude ?? ''),
  templateId: monster.templateId ?? monster.template?.id ?? '',
  dominantHandInventoryItemId:
    monster.dominantHandInventoryItemId !== undefined &&
    monster.dominantHandInventoryItemId !== null
      ? String(monster.dominantHandInventoryItemId)
      : monster.weaponInventoryItemId !== undefined &&
          monster.weaponInventoryItemId !== null
        ? String(monster.weaponInventoryItemId)
        : '',
  offHandInventoryItemId:
    monster.offHandInventoryItemId !== undefined &&
    monster.offHandInventoryItemId !== null
      ? String(monster.offHandInventoryItemId)
      : '',
  level: String(monster.level ?? 1),
  rewardExperience: String(monster.rewardExperience ?? 0),
  rewardGold: String(monster.rewardGold ?? 0),
  itemRewards: (monster.itemRewards ?? []).map((reward) => ({
    inventoryItemId: String(reward.inventoryItemId),
    quantity: String(reward.quantity),
  })),
});

const monsterPayloadFromForm = (form: MonsterFormState) => ({
  name: form.name.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  zoneId: form.zoneId.trim(),
  latitude: parseFloatSafe(form.latitude, 0),
  longitude: parseFloatSafe(form.longitude, 0),
  templateId: form.templateId.trim(),
  dominantHandInventoryItemId: parseIntSafe(
    form.dominantHandInventoryItemId,
    0
  ),
  offHandInventoryItemId: parseOptionalInt(form.offHandInventoryItemId),
  weaponInventoryItemId: parseIntSafe(form.dominantHandInventoryItemId, 0),
  level: parseIntSafe(form.level, 1),
  rewardExperience: parseIntSafe(form.rewardExperience, 0),
  rewardGold: parseIntSafe(form.rewardGold, 0),
  itemRewards: form.itemRewards
    .map((reward) => ({
      inventoryItemId: parseIntSafe(reward.inventoryItemId, 0),
      quantity: parseIntSafe(reward.quantity, 0),
    }))
    .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0),
});

const formatGenerationStatus = (status?: string) => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'In Progress';
    case 'complete':
      return 'Complete';
    case 'failed':
      return 'Failed';
    default:
      return 'None';
  }
};

const formatBulkTemplateStatus = (status?: string): string => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'In Progress';
    case 'completed':
      return 'Completed';
    case 'failed':
      return 'Failed';
    default:
      return 'Unknown';
  }
};

const staticStatusClassName = (status: string): string => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    case 'missing':
      return 'bg-gray-500';
    default:
      return 'bg-gray-400';
  }
};

const formatDate = (value?: string): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

export const Monsters = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();

  const [loading, setLoading] = useState(true);
  const [templates, setTemplates] = useState<MonsterTemplateRecord[]>([]);
  const [records, setRecords] = useState<MonsterRecord[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [inventoryItems, setInventoryItems] = useState<InventoryItemLite[]>([]);
  const [query, setQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  const [showTemplateModal, setShowTemplateModal] = useState(false);
  const [editingTemplate, setEditingTemplate] =
    useState<MonsterTemplateRecord | null>(null);
  const [templateForm, setTemplateForm] =
    useState<MonsterTemplateFormState>(emptyTemplateForm());

  const [showMonsterModal, setShowMonsterModal] = useState(false);
  const [editingMonster, setEditingMonster] = useState<MonsterRecord | null>(
    null
  );
  const [monsterForm, setMonsterForm] =
    useState<MonsterFormState>(emptyMonsterForm());
  const [imagePreview, setImagePreview] = useState<ImagePreviewState | null>(
    null
  );
  const [geoLoading, setGeoLoading] = useState(false);
  const [monsterUndiscoveredBusy, setMonsterUndiscoveredBusy] = useState(false);
  const [
    monsterUndiscoveredStatusLoading,
    setMonsterUndiscoveredStatusLoading,
  ] = useState(false);
  const [monsterUndiscoveredError, setMonsterUndiscoveredError] = useState<
    string | null
  >(null);
  const [monsterUndiscoveredMessage, setMonsterUndiscoveredMessage] = useState<
    string | null
  >(null);
  const [monsterUndiscoveredUrl, setMonsterUndiscoveredUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png'
  );
  const [monsterUndiscoveredStatus, setMonsterUndiscoveredStatus] =
    useState('unknown');
  const [monsterUndiscoveredExists, setMonsterUndiscoveredExists] =
    useState(false);
  const [monsterUndiscoveredRequestedAt, setMonsterUndiscoveredRequestedAt] =
    useState<string | null>(null);
  const [monsterUndiscoveredLastModified, setMonsterUndiscoveredLastModified] =
    useState<string | null>(null);
  const [monsterUndiscoveredPreviewNonce, setMonsterUndiscoveredPreviewNonce] =
    useState<number>(Date.now());
  const [monsterUndiscoveredPrompt, setMonsterUndiscoveredPrompt] = useState(
    defaultMonsterUndiscoveredIconPrompt
  );

  const [generatingMonsterId, setGeneratingMonsterId] = useState<string | null>(
    null
  );
  const [generatingTemplateId, setGeneratingTemplateId] = useState<
    string | null
  >(null);
  const [bulkTemplateCount, setBulkTemplateCount] = useState('8');
  const [bulkTemplateBusy, setBulkTemplateBusy] = useState(false);
  const [bulkTemplateJob, setBulkTemplateJob] =
    useState<BulkMonsterTemplateStatus | null>(null);
  const [bulkTemplateError, setBulkTemplateError] = useState<string | null>(
    null
  );
  const [bulkTemplateMessage, setBulkTemplateMessage] = useState<string | null>(
    null
  );
  const mapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const mapRef = React.useRef<mapboxgl.Map | null>(null);
  const markerRef = React.useRef<mapboxgl.Marker | null>(null);
  const formLatitudeRef = React.useRef(monsterForm.latitude);
  const formLongitudeRef = React.useRef(monsterForm.longitude);

  const load = useCallback(
    async (suppressLoading = false) => {
      try {
        if (!suppressLoading) {
          setLoading(true);
        }
        setError(null);
        const [templateResp, monsterResp, spellResp, inventoryResp] =
          await Promise.all([
            apiClient.get<MonsterTemplateRecord[]>('/sonar/monster-templates'),
            apiClient.get<MonsterRecord[]>('/sonar/monsters'),
            apiClient.get<Spell[]>('/sonar/spells'),
            apiClient.get<InventoryItemLite[]>('/sonar/inventory-items'),
          ]);
        setTemplates(Array.isArray(templateResp) ? templateResp : []);
        setRecords(Array.isArray(monsterResp) ? monsterResp : []);
        setSpells(Array.isArray(spellResp) ? spellResp : []);
        setInventoryItems(Array.isArray(inventoryResp) ? inventoryResp : []);
      } catch (err) {
        console.error('Failed to load monsters/templates', err);
        setError('Failed to load monsters/templates.');
      } finally {
        if (!suppressLoading) {
          setLoading(false);
        }
      }
    },
    [apiClient]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const hasPendingMonsterGeneration = records.some((record) =>
      ['queued', 'in_progress'].includes(record.imageGenerationStatus || '')
    );
    const hasPendingTemplateGeneration = templates.some((template) =>
      ['queued', 'in_progress'].includes(template.imageGenerationStatus || '')
    );
    const hasPendingGeneration =
      hasPendingMonsterGeneration || hasPendingTemplateGeneration;
    if (!hasPendingGeneration) return;

    const interval = window.setInterval(() => {
      void load(true);
    }, 5000);

    return () => clearInterval(interval);
  }, [load, records, templates]);

  useEffect(() => {
    formLatitudeRef.current = monsterForm.latitude;
    formLongitudeRef.current = monsterForm.longitude;
  }, [monsterForm.latitude, monsterForm.longitude]);

  const zoneNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const zone of zones) {
      map.set(zone.id, zone.name);
    }
    return map;
  }, [zones]);

  const templateNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const template of templates) {
      map.set(template.id, template.name);
    }
    return map;
  }, [templates]);

  const dominantHandItems = useMemo(() => {
    return inventoryItems.filter((item) => {
      const equipSlot = (item.equipSlot ?? '').trim().toLowerCase();
      const category = (item.handItemCategory ?? '').trim().toLowerCase();
      return (
        equipSlot === 'dominant_hand' &&
        (category === 'weapon' || category === 'staff') &&
        item.damageMin !== undefined &&
        item.damageMin !== null &&
        item.damageMax !== undefined &&
        item.damageMax !== null
      );
    });
  }, [inventoryItems]);

  const offHandItems = useMemo(() => {
    return inventoryItems.filter((item) => {
      const equipSlot = (item.equipSlot ?? '').trim().toLowerCase();
      const category = (item.handItemCategory ?? '').trim().toLowerCase();
      const handedness = (item.handedness ?? '').trim().toLowerCase();
      const isOffhandUtility =
        equipSlot === 'off_hand' &&
        handedness === 'one_handed' &&
        (category === 'shield' || category === 'orb');
      const isOneHandedWeapon =
        equipSlot === 'dominant_hand' &&
        handedness === 'one_handed' &&
        category === 'weapon' &&
        item.damageMin !== undefined &&
        item.damageMin !== null &&
        item.damageMax !== undefined &&
        item.damageMax !== null;
      return isOffhandUtility || isOneHandedWeapon;
    });
  }, [inventoryItems]);

  const dominantHandItemById = useMemo(() => {
    const map = new Map<number, InventoryItemLite>();
    for (const item of dominantHandItems) {
      map.set(item.id, item);
    }
    return map;
  }, [dominantHandItems]);

  const selectedDominantHandItem = useMemo(() => {
    const selectedID = parseIntSafe(monsterForm.dominantHandInventoryItemId, 0);
    return selectedID > 0 ? dominantHandItemById.get(selectedID) : undefined;
  }, [monsterForm.dominantHandInventoryItemId, dominantHandItemById]);

  const dominantIsTwoHanded =
    (selectedDominantHandItem?.handedness ?? '').trim().toLowerCase() ===
    'two_handed';

  const spellAbilities = useMemo(
    () =>
      spells.filter((spell) => (spell.abilityType ?? 'spell') !== 'technique'),
    [spells]
  );
  const techniqueAbilities = useMemo(
    () =>
      spells.filter((spell) => (spell.abilityType ?? 'spell') === 'technique'),
    [spells]
  );

  const filteredTemplates = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return templates;
    return templates.filter(
      (template) =>
        template.name.toLowerCase().includes(normalized) ||
        template.description.toLowerCase().includes(normalized)
    );
  }, [query, templates]);

  const filteredMonsters = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return records;
    return records.filter((record) => {
      const zoneName = zoneNameById.get(record.zoneId) ?? '';
      const templateName =
        record.template?.name ||
        (record.templateId
          ? templateNameById.get(record.templateId) ?? ''
          : '');
      return (
        record.name.toLowerCase().includes(normalized) ||
        record.description.toLowerCase().includes(normalized) ||
        zoneName.toLowerCase().includes(normalized) ||
        templateName.toLowerCase().includes(normalized)
      );
    });
  }, [query, records, zoneNameById, templateNameById]);

  const openCreateTemplate = () => {
    setEditingTemplate(null);
    setTemplateForm(emptyTemplateForm());
    setShowTemplateModal(true);
  };

  const openEditTemplate = (template: MonsterTemplateRecord) => {
    setEditingTemplate(template);
    setTemplateForm(templateFormFromRecord(template));
    setShowTemplateModal(true);
  };

  const closeTemplateModal = () => {
    setShowTemplateModal(false);
    setEditingTemplate(null);
    setTemplateForm(emptyTemplateForm());
  };

  const saveTemplate = async () => {
    try {
      const payload = templatePayloadFromForm(templateForm);
      if (!payload.name) {
        alert('Template name is required.');
        return;
      }
      if (
        payload.baseStrength <= 0 ||
        payload.baseDexterity <= 0 ||
        payload.baseConstitution <= 0 ||
        payload.baseIntelligence <= 0 ||
        payload.baseWisdom <= 0 ||
        payload.baseCharisma <= 0
      ) {
        alert('All base stats must be positive.');
        return;
      }

      if (editingTemplate) {
        const updated = await apiClient.put<MonsterTemplateRecord>(
          `/sonar/monster-templates/${editingTemplate.id}`,
          payload
        );
        setTemplates((prev) =>
          prev.map((template) =>
            template.id === updated.id ? updated : template
          )
        );
      } else {
        const created = await apiClient.post<MonsterTemplateRecord>(
          '/sonar/monster-templates',
          payload
        );
        setTemplates((prev) => [created, ...prev]);
      }
      closeTemplateModal();
    } catch (err) {
      console.error('Failed to save template', err);
      const message =
        err instanceof Error ? err.message : 'Failed to save template.';
      alert(message);
    }
  };

  const deleteTemplate = async (template: MonsterTemplateRecord) => {
    if (!window.confirm(`Delete template "${template.name}"?`)) return;
    try {
      await apiClient.delete(`/sonar/monster-templates/${template.id}`);
      setTemplates((prev) => prev.filter((entry) => entry.id !== template.id));
    } catch (err) {
      console.error('Failed to delete template', err);
      const message =
        err instanceof Error ? err.message : 'Failed to delete template.';
      alert(message);
    }
  };

  const refreshBulkTemplateJobStatus = useCallback(
    async (jobId: string) => {
      try {
        const status = await apiClient.get<BulkMonsterTemplateStatus>(
          `/sonar/monster-templates/bulk-generate/${jobId}/status`
        );
        setBulkTemplateJob(status);
        if (status.status === 'completed') {
          setBulkTemplateBusy(false);
          setBulkTemplateError(null);
          setBulkTemplateMessage(
            `Created ${status.createdCount} monster template(s).`
          );
          await load(true);
        } else if (status.status === 'failed') {
          setBulkTemplateBusy(false);
          setBulkTemplateMessage(null);
          setBulkTemplateError(
            status.error || 'Bulk template generation failed.'
          );
        }
      } catch (err) {
        console.error('Failed to refresh bulk monster template status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to refresh bulk template generation status.';
        setBulkTemplateError(message);
        setBulkTemplateBusy(false);
      }
    },
    [apiClient, load]
  );

  const handleBulkGenerateTemplates = async () => {
    const count = Number.parseInt(bulkTemplateCount, 10);
    if (!Number.isFinite(count) || count < 1 || count > 100) {
      alert('Bulk template count must be between 1 and 100.');
      return;
    }
    try {
      setBulkTemplateBusy(true);
      setBulkTemplateError(null);
      setBulkTemplateMessage(null);
      setBulkTemplateJob(null);
      const response = await apiClient.post<BulkMonsterTemplateStatus>(
        '/sonar/monster-templates/bulk-generate',
        { count }
      );
      setBulkTemplateJob(response);
      if (response.status === 'completed') {
        setBulkTemplateBusy(false);
        setBulkTemplateMessage(
          `Created ${response.createdCount} monster template(s).`
        );
        await load(true);
      } else if (response.status === 'failed') {
        setBulkTemplateBusy(false);
        setBulkTemplateError(
          response.error || 'Bulk template generation failed.'
        );
      }
    } catch (err) {
      console.error('Failed to bulk generate monster templates', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to bulk generate monster templates.';
      setBulkTemplateError(message);
      setBulkTemplateBusy(false);
    }
  };

  const openCreateMonster = () => {
    setEditingMonster(null);
    setMonsterForm({
      ...emptyMonsterForm(),
      zoneId: zones[0]?.id ?? '',
      templateId: templates[0]?.id ?? '',
      dominantHandInventoryItemId: dominantHandItems[0]
        ? String(dominantHandItems[0].id)
        : '',
    });
    setShowMonsterModal(true);
  };

  const openEditMonster = (monster: MonsterRecord) => {
    setEditingMonster(monster);
    setMonsterForm(monsterFormFromRecord(monster));
    setShowMonsterModal(true);
  };

  const closeMonsterModal = () => {
    setShowMonsterModal(false);
    setEditingMonster(null);
    setMonsterForm(emptyMonsterForm());
  };

  const saveMonster = async () => {
    try {
      const payload = monsterPayloadFromForm(monsterForm);
      if (!payload.zoneId || !payload.templateId) {
        alert('Zone and template are required.');
        return;
      }
      if (
        !payload.dominantHandInventoryItemId ||
        payload.dominantHandInventoryItemId <= 0
      ) {
        alert('A dominant hand weapon is required.');
        return;
      }
      if (payload.level <= 0) {
        alert('Level must be positive.');
        return;
      }
      if (dominantIsTwoHanded && payload.offHandInventoryItemId) {
        alert(
          'Two-handed dominant weapons cannot be combined with an off-hand item.'
        );
        return;
      }

      if (editingMonster) {
        const updated = await apiClient.put<MonsterRecord>(
          `/sonar/monsters/${editingMonster.id}`,
          payload
        );
        setRecords((prev) =>
          prev.map((record) => (record.id === updated.id ? updated : record))
        );
      } else {
        const created = await apiClient.post<MonsterRecord>(
          '/sonar/monsters',
          payload
        );
        setRecords((prev) => [created, ...prev]);
      }
      closeMonsterModal();
    } catch (err) {
      console.error('Failed to save monster', err);
      const message =
        err instanceof Error ? err.message : 'Failed to save monster.';
      alert(message);
    }
  };

  const deleteMonster = async (monster: MonsterRecord) => {
    if (!window.confirm(`Delete monster "${monster.name}"?`)) return;
    try {
      await apiClient.delete(`/sonar/monsters/${monster.id}`);
      setRecords((prev) => prev.filter((record) => record.id !== monster.id));
    } catch (err) {
      console.error('Failed to delete monster', err);
      alert('Failed to delete monster.');
    }
  };

  const handleGenerateImage = async (monster: MonsterRecord) => {
    try {
      setGeneratingMonsterId(monster.id);
      const updated = await apiClient.post<MonsterRecord>(
        `/sonar/monsters/${monster.id}/generate-image`,
        {}
      );
      setRecords((prev) =>
        prev.map((record) => (record.id === monster.id ? updated : record))
      );
    } catch (err) {
      console.error('Failed to queue monster image generation', err);
      alert('Failed to queue monster image generation.');
    } finally {
      setGeneratingMonsterId(null);
    }
  };

  const handleGenerateTemplateImage = async (
    template: MonsterTemplateRecord
  ) => {
    try {
      setGeneratingTemplateId(template.id);
      const updated = await apiClient.post<MonsterTemplateRecord>(
        `/sonar/monster-templates/${template.id}/generate-image`,
        {}
      );
      setTemplates((prev) =>
        prev.map((record) => (record.id === template.id ? updated : record))
      );
    } catch (err) {
      console.error('Failed to queue monster template image generation', err);
      alert('Failed to queue monster template image generation.');
    } finally {
      setGeneratingTemplateId(null);
    }
  };

  const updateMonsterItemReward = (
    index: number,
    nextReward: Partial<MonsterFormItem>
  ) => {
    setMonsterForm((prev) => {
      const next = [...prev.itemRewards];
      next[index] = { ...next[index], ...nextReward };
      return { ...prev, itemRewards: next };
    });
  };

  const updateTemplateSpellIds = (selected: HTMLSelectElement) => {
    const spellIds = Array.from(selected.selectedOptions).map(
      (option) => option.value
    );
    setTemplateForm((prev) => ({ ...prev, spellIds }));
  };

  const updateTemplateTechniqueIds = (selected: HTMLSelectElement) => {
    const techniqueIds = Array.from(selected.selectedOptions).map(
      (option) => option.value
    );
    setTemplateForm((prev) => ({ ...prev, techniqueIds }));
  };

  const updateDominantHandSelection = (value: string) => {
    setMonsterForm((prev) => {
      const selectedID = parseIntSafe(value, 0);
      const selected =
        selectedID > 0 ? dominantHandItemById.get(selectedID) : undefined;
      const handedness = (selected?.handedness ?? '').trim().toLowerCase();
      return {
        ...prev,
        dominantHandInventoryItemId: value,
        offHandInventoryItemId:
          handedness === 'two_handed' ? '' : prev.offHandInventoryItemId,
      };
    });
  };

  const setMonsterLocation = useCallback(
    (latitude: number, longitude: number) => {
      setMonsterForm((prev) => ({
        ...prev,
        latitude: latitude.toFixed(6),
        longitude: longitude.toFixed(6),
      }));
    },
    []
  );

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      alert('Geolocation is not supported in this browser.');
      return;
    }
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGeoLoading(false);
        setMonsterLocation(position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${error.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setMonsterLocation]);

  useEffect(() => {
    if (!showMonsterModal) return;
    if (!mapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (mapRef.current) return;

    const parsedLat = Number.parseFloat(formLatitudeRef.current);
    const parsedLng = Number.parseFloat(formLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === monsterForm.zoneId);
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] =
      Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
        ? [parsedLng, parsedLat]
        : Number.isFinite(zoneLat) && Number.isFinite(zoneLng)
          ? [zoneLng, zoneLat]
          : [-73.98513, 40.7589];

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setMonsterLocation(event.lngLat.lat, event.lngLat.lng);
    });

    mapRef.current = map;

    return () => {
      markerRef.current?.remove();
      markerRef.current = null;
      map.remove();
      mapRef.current = null;
    };
  }, [monsterForm.zoneId, setMonsterLocation, showMonsterModal, zones]);

  useEffect(() => {
    if (!showMonsterModal) return;
    if (!mapRef.current) return;

    const lat = Number.parseFloat(monsterForm.latitude);
    const lng = Number.parseFloat(monsterForm.longitude);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
      markerRef.current?.remove();
      markerRef.current = null;
      return;
    }

    if (!markerRef.current) {
      markerRef.current = new mapboxgl.Marker({ color: '#dc2626' })
        .setLngLat([lng, lat])
        .addTo(mapRef.current);
    } else {
      markerRef.current.setLngLat([lng, lat]);
    }

    mapRef.current.easeTo({ center: [lng, lat], duration: 350 });
  }, [monsterForm.latitude, monsterForm.longitude, showMonsterModal]);

  const openMonsterImagePreview = (monster: MonsterRecord) => {
    const url = monster.imageUrl || monster.thumbnailUrl;
    if (!url) return;
    setImagePreview({
      url,
      alt: `${monster.name || 'Monster'} image`,
    });
  };

  const openTemplateImagePreview = (template: MonsterTemplateRecord) => {
    const url = template.imageUrl || template.thumbnailUrl;
    if (!url) return;
    setImagePreview({
      url,
      alt: `${template.name || 'Monster template'} image`,
    });
  };

  const closeImagePreview = () => {
    setImagePreview(null);
  };

  const refreshUndiscoveredMonsterIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setMonsterUndiscoveredStatusLoading(true);
        setMonsterUndiscoveredError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/monster-undiscovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setMonsterUndiscoveredUrl(url);
        }
        setMonsterUndiscoveredStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setMonsterUndiscoveredExists(Boolean(response?.exists));
        setMonsterUndiscoveredRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setMonsterUndiscoveredLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setMonsterUndiscoveredPreviewNonce(Date.now());
        if (showMessage) {
          setMonsterUndiscoveredMessage(
            'Undiscovered monster icon status refreshed.'
          );
        }
      } catch (err) {
        console.error('Failed to load undiscovered monster icon status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load undiscovered monster icon status.';
        setMonsterUndiscoveredError(message);
      } finally {
        setMonsterUndiscoveredStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateUndiscoveredMonsterIcon = useCallback(async () => {
    const prompt = monsterUndiscoveredPrompt.trim();
    if (!prompt) {
      setMonsterUndiscoveredError('Prompt is required.');
      return;
    }
    try {
      setMonsterUndiscoveredBusy(true);
      setMonsterUndiscoveredError(null);
      setMonsterUndiscoveredMessage(null);
      await apiClient.post<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/monster-undiscovered',
        { prompt }
      );
      setMonsterUndiscoveredMessage(
        'Undiscovered monster icon queued for generation.'
      );
      await refreshUndiscoveredMonsterIconStatus();
    } catch (err) {
      console.error('Failed to generate undiscovered monster icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to generate undiscovered monster icon.';
      setMonsterUndiscoveredError(message);
    } finally {
      setMonsterUndiscoveredBusy(false);
    }
  }, [apiClient, monsterUndiscoveredPrompt, refreshUndiscoveredMonsterIconStatus]);

  const handleDeleteUndiscoveredMonsterIcon = useCallback(async () => {
    try {
      setMonsterUndiscoveredBusy(true);
      setMonsterUndiscoveredError(null);
      setMonsterUndiscoveredMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/monster-undiscovered'
      );
      setMonsterUndiscoveredMessage('Undiscovered monster icon deleted.');
      await refreshUndiscoveredMonsterIconStatus();
    } catch (err) {
      console.error('Failed to delete undiscovered monster icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete undiscovered monster icon.';
      setMonsterUndiscoveredError(message);
    } finally {
      setMonsterUndiscoveredBusy(false);
    }
  }, [apiClient, refreshUndiscoveredMonsterIconStatus]);

  useEffect(() => {
    void refreshUndiscoveredMonsterIconStatus();
  }, [refreshUndiscoveredMonsterIconStatus]);

  useEffect(() => {
    if (
      monsterUndiscoveredStatus !== 'queued' &&
      monsterUndiscoveredStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshUndiscoveredMonsterIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [monsterUndiscoveredStatus, refreshUndiscoveredMonsterIconStatus]);

  useEffect(() => {
    if (!bulkTemplateJob?.jobId) {
      return;
    }
    if (
      bulkTemplateJob.status !== 'queued' &&
      bulkTemplateJob.status !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshBulkTemplateJobStatus(bulkTemplateJob.jobId);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [bulkTemplateJob, refreshBulkTemplateJobStatus]);

  return (
    <div className="p-6 bg-gray-100 min-h-screen">
      <div className="max-w-7xl mx-auto space-y-6">
        <div className="qa-card">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h1 className="qa-card-title">Monsters</h1>
              <p className="text-sm text-gray-600">
                Configure monster templates (base stats + abilities), then place
                monsters with level and weapon loadouts.
              </p>
            </div>
            <div className="flex gap-2">
              <input
                type="number"
                min={1}
                max={100}
                value={bulkTemplateCount}
                onChange={(event) => setBulkTemplateCount(event.target.value)}
                className="w-24 rounded-md border border-gray-300 px-2 py-2 text-sm"
                aria-label="Bulk template count"
              />
              <button
                className="qa-btn qa-btn-secondary"
                onClick={handleBulkGenerateTemplates}
                disabled={bulkTemplateBusy}
              >
                {bulkTemplateBusy ? 'Generating...' : 'Generate Templates'}
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={openCreateTemplate}
              >
                Create Template
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={openCreateMonster}
              >
                Create Monster
              </button>
            </div>
          </div>
          {bulkTemplateJob && (
            <div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-gray-700">
              <span
                className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold text-white ${staticStatusClassName(
                  bulkTemplateJob.status
                )}`}
              >
                {formatBulkTemplateStatus(bulkTemplateJob.status)}
              </span>
              <span>
                Progress: {bulkTemplateJob.createdCount}/{bulkTemplateJob.totalCount}
              </span>
              <span>Job: {bulkTemplateJob.jobId}</span>
              <span>Updated: {formatDate(bulkTemplateJob.updatedAt)}</span>
            </div>
          )}
          {bulkTemplateMessage && (
            <p className="mt-2 text-sm text-emerald-700">{bulkTemplateMessage}</p>
          )}
          {bulkTemplateError && (
            <p className="mt-2 text-sm text-red-700">{bulkTemplateError}</p>
          )}
        </div>

        <div className="qa-card">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold">
                Undiscovered Monster Icon
              </h2>
              <p className="text-xs text-gray-600 break-all">
                URL: {monsterUndiscoveredUrl}
              </p>
              <p className="text-xs text-gray-600 mt-1">
                Requested:{' '}
                {formatDate(monsterUndiscoveredRequestedAt ?? undefined)}
                {' · '}
                Last updated:{' '}
                {formatDate(monsterUndiscoveredLastModified ?? undefined)}
              </p>
            </div>
            <div className="flex gap-2">
              <button
                className="qa-btn qa-btn-secondary"
                onClick={() => void refreshUndiscoveredMonsterIconStatus(true)}
                disabled={monsterUndiscoveredStatusLoading}
              >
                {monsterUndiscoveredStatusLoading
                  ? 'Refreshing...'
                  : 'Refresh Status'}
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={handleGenerateUndiscoveredMonsterIcon}
                disabled={
                  monsterUndiscoveredBusy || monsterUndiscoveredStatusLoading
                }
              >
                {monsterUndiscoveredBusy ? 'Working...' : 'Generate Icon'}
              </button>
              <button
                className="qa-btn qa-btn-danger"
                onClick={handleDeleteUndiscoveredMonsterIcon}
                disabled={
                  monsterUndiscoveredBusy || monsterUndiscoveredStatusLoading
                }
              >
                {monsterUndiscoveredBusy ? 'Working...' : 'Delete Icon'}
              </button>
            </div>
          </div>
          <div className="mt-2">
            <span
              className={`inline-flex text-white text-xs px-2 py-0.5 rounded ${staticStatusClassName(
                monsterUndiscoveredStatus
              )}`}
            >
              {monsterUndiscoveredStatus || 'unknown'}
            </span>
          </div>
          <label className="block text-sm mt-3">
            Generation Prompt
            <textarea
              className="block w-full border border-gray-300 rounded-md p-2 mt-1 min-h-[88px]"
              value={monsterUndiscoveredPrompt}
              onChange={(event) =>
                setMonsterUndiscoveredPrompt(event.target.value)
              }
              placeholder="Prompt used to generate the undiscovered monster icon."
            />
          </label>
          {monsterUndiscoveredExists ? (
            <div className="mt-3">
              <img
                src={`${monsterUndiscoveredUrl}?v=${monsterUndiscoveredPreviewNonce}`}
                alt="Undiscovered monster icon preview"
                className="w-24 h-24 object-cover border rounded-md bg-gray-50"
              />
            </div>
          ) : (
            <p className="text-xs text-gray-500 mt-2">
              No icon currently found at this URL.
            </p>
          )}
          {monsterUndiscoveredMessage ? (
            <p className="text-sm text-emerald-700 mt-2">
              {monsterUndiscoveredMessage}
            </p>
          ) : null}
          {monsterUndiscoveredError ? (
            <p className="text-sm text-red-600 mt-2">
              {monsterUndiscoveredError}
            </p>
          ) : null}
        </div>

        <div className="qa-card">
          <input
            className="block w-full border border-gray-300 rounded-md p-2"
            placeholder="Search monsters/templates..."
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
        </div>

        {error && <div className="qa-card text-red-600">{error}</div>}

        {loading ? <div className="qa-card">Loading...</div> : null}

        {!loading ? (
          <>
            <div className="qa-card">
              <h2 className="text-lg font-semibold mb-3">Monster Templates</h2>
              {filteredTemplates.length === 0 ? (
                <p className="text-sm text-gray-600">No templates found.</p>
              ) : (
                <div className="grid gap-3">
                  {filteredTemplates.map((template) => (
                    <div
                      key={template.id}
                      className="border border-gray-200 rounded-md p-3 bg-white"
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex gap-3 min-w-0">
                          {template.thumbnailUrl || template.imageUrl ? (
                            <button
                              type="button"
                              className="block"
                              onClick={() => openTemplateImagePreview(template)}
                              title="Open image preview"
                            >
                              <img
                                src={template.thumbnailUrl || template.imageUrl}
                                alt={template.name}
                                className="w-14 h-14 rounded object-cover border border-gray-200 cursor-zoom-in"
                              />
                            </button>
                          ) : (
                            <div className="w-14 h-14 rounded bg-gray-200 flex items-center justify-center text-gray-500">
                              ?
                            </div>
                          )}
                          <div>
                            <div className="font-semibold text-base">
                              {template.name}
                            </div>
                            {template.description ? (
                              <p className="text-sm text-gray-600 mt-1">
                                {template.description}
                              </p>
                            ) : null}
                            <p className="text-sm text-gray-600 mt-1">
                              STR {template.baseStrength} · DEX{' '}
                              {template.baseDexterity} · CON{' '}
                              {template.baseConstitution} · INT{' '}
                              {template.baseIntelligence} · WIS{' '}
                              {template.baseWisdom} · CHA{' '}
                              {template.baseCharisma}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Spells:{' '}
                              {template.spells?.filter(
                                (spell) =>
                                  (spell.abilityType ?? 'spell') !== 'technique'
                              ).length ?? 0}
                              {' · '}
                              Techniques:{' '}
                              {template.spells?.filter(
                                (spell) =>
                                  (spell.abilityType ?? 'spell') === 'technique'
                              ).length ?? 0}
                            </p>
                            <p className="text-xs text-gray-500 mt-1">
                              Image generation:{' '}
                              {formatGenerationStatus(
                                template.imageGenerationStatus
                              )}
                            </p>
                            {template.imageGenerationError ? (
                              <p className="text-xs text-red-600 mt-1">
                                {template.imageGenerationError}
                              </p>
                            ) : null}
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() =>
                              handleGenerateTemplateImage(template)
                            }
                            disabled={
                              generatingTemplateId === template.id ||
                              ['queued', 'in_progress'].includes(
                                template.imageGenerationStatus || ''
                              )
                            }
                          >
                            {generatingTemplateId === template.id
                              ? 'Queueing...'
                              : 'Generate Image'}
                          </button>
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => openEditTemplate(template)}
                          >
                            Edit
                          </button>
                          <button
                            className="qa-btn qa-btn-danger"
                            onClick={() => deleteTemplate(template)}
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="qa-card">
              <h2 className="text-lg font-semibold mb-3">Monsters</h2>
              {filteredMonsters.length === 0 ? (
                <p className="text-sm text-gray-600">No monsters found.</p>
              ) : (
                <div className="grid gap-4">
                  {filteredMonsters.map((monster) => (
                    <div
                      key={monster.id}
                      className="border border-gray-200 rounded-md p-3 bg-white"
                    >
                      <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
                        <div className="flex gap-3 min-w-0">
                          {monster.thumbnailUrl || monster.imageUrl ? (
                            <button
                              type="button"
                              className="block"
                              onClick={() => openMonsterImagePreview(monster)}
                              title="Open image preview"
                            >
                              <img
                                src={monster.thumbnailUrl || monster.imageUrl}
                                alt={monster.name}
                                className="w-16 h-16 rounded object-cover border border-gray-200 cursor-zoom-in"
                              />
                            </button>
                          ) : (
                            <div className="w-16 h-16 rounded bg-gray-200 flex items-center justify-center text-gray-500">
                              ?
                            </div>
                          )}
                          <div className="min-w-0">
                            <h3 className="font-semibold text-lg truncate">
                              {monster.name}
                            </h3>
                            <p className="text-sm text-gray-600">
                              Zone:{' '}
                              {zoneNameById.get(monster.zoneId) ??
                                monster.zoneId}{' '}
                              · Level {monster.level}
                            </p>
                            <p className="text-sm text-gray-600">
                              Template:{' '}
                              {monster.template?.name ??
                                (monster.templateId
                                  ? templateNameById.get(monster.templateId)
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Dominant Hand:{' '}
                              {monster.dominantHandInventoryItem?.name ??
                                monster.weaponInventoryItem?.name ??
                                (monster.dominantHandInventoryItemId ??
                                monster.weaponInventoryItemId
                                  ? `#${monster.dominantHandInventoryItemId ?? monster.weaponInventoryItemId}`
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Off Hand:{' '}
                              {monster.offHandInventoryItem?.name ??
                                (monster.offHandInventoryItemId
                                  ? `#${monster.offHandInventoryItemId}`
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Damage {monster.attackDamageMin}-
                              {monster.attackDamageMax} · Swipes{' '}
                              {monster.attackSwipesPerAttack}
                            </p>
                            <p className="text-sm text-gray-600">
                              Health {monster.health}/{monster.maxHealth} · Mana{' '}
                              {monster.mana}/{monster.maxMana}
                            </p>
                            <p className="text-sm text-gray-600">
                              STR {monster.strength} · DEX {monster.dexterity} ·
                              CON {monster.constitution} · INT{' '}
                              {monster.intelligence} · WIS {monster.wisdom} ·
                              CHA {monster.charisma}
                            </p>
                            <p className="text-sm text-gray-600">
                              XP {monster.rewardExperience} · Gold{' '}
                              {monster.rewardGold} · Item rewards{' '}
                              {monster.itemRewards?.length ?? 0}
                            </p>
                            <p className="text-xs text-gray-500 mt-1">
                              Image generation:{' '}
                              {formatGenerationStatus(
                                monster.imageGenerationStatus
                              )}
                            </p>
                            {monster.imageGenerationError ? (
                              <p className="text-xs text-red-600 mt-1">
                                {monster.imageGenerationError}
                              </p>
                            ) : null}
                          </div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => handleGenerateImage(monster)}
                            disabled={generatingMonsterId === monster.id}
                          >
                            {generatingMonsterId === monster.id
                              ? 'Queueing...'
                              : 'Generate Image'}
                          </button>
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => openEditMonster(monster)}
                          >
                            Edit
                          </button>
                          <button
                            className="qa-btn qa-btn-danger"
                            onClick={() => deleteMonster(monster)}
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        ) : null}
      </div>

      {showTemplateModal ? (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-3xl max-h-[92vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="text-lg font-semibold">
                {editingTemplate ? 'Edit Template' : 'Create Template'}
              </h2>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={closeTemplateModal}
              >
                Close
              </button>
            </div>
            <div className="p-6 overflow-y-auto max-h-[calc(92vh-72px)] space-y-4">
              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Name</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.name}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        name: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Description</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.description}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        description: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Image URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.imageUrl}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        imageUrl: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Thumbnail URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.thumbnailUrl}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        thumbnailUrl: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Base STR</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseStrength}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseStrength: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base DEX</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseDexterity}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseDexterity: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base CON</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseConstitution}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseConstitution: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base INT</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseIntelligence}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseIntelligence: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base WIS</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseWisdom}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseWisdom: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base CHA</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseCharisma}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseCharisma: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <label className="block">
                <span className="block text-sm mb-1">Spells</span>
                <select
                  multiple
                  className="w-full border border-gray-300 rounded-md p-2 min-h-[180px]"
                  value={templateForm.spellIds}
                  onChange={(event) => updateTemplateSpellIds(event.target)}
                >
                  {spellAbilities.map((spell) => (
                    <option key={spell.id} value={spell.id}>
                      {spell.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block">
                <span className="block text-sm mb-1">Techniques</span>
                <select
                  multiple
                  className="w-full border border-gray-300 rounded-md p-2 min-h-[180px]"
                  value={templateForm.techniqueIds}
                  onChange={(event) => updateTemplateTechniqueIds(event.target)}
                >
                  {techniqueAbilities.map((technique) => (
                    <option key={technique.id} value={technique.id}>
                      {technique.name}
                    </option>
                  ))}
                </select>
              </label>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  className="qa-btn qa-btn-secondary"
                  onClick={closeTemplateModal}
                >
                  Cancel
                </button>
                <button
                  className="qa-btn qa-btn-primary"
                  onClick={saveTemplate}
                >
                  Save Template
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {showMonsterModal ? (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-4xl max-h-[92vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="text-lg font-semibold">
                {editingMonster ? 'Edit Monster' : 'Create Monster'}
              </h2>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={closeMonsterModal}
              >
                Close
              </button>
            </div>
            <div className="p-6 overflow-y-auto max-h-[calc(92vh-72px)] space-y-4">
              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Name (optional)</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.name}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        name: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Template</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.templateId}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        templateId: event.target.value,
                      }))
                    }
                  >
                    <option value="">Select template</option>
                    {templates.map((template) => (
                      <option key={template.id} value={template.id}>
                        {template.name}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <label className="block">
                <span className="block text-sm mb-1">
                  Description (optional)
                </span>
                <textarea
                  className="w-full border border-gray-300 rounded-md p-2"
                  rows={3}
                  value={monsterForm.description}
                  onChange={(event) =>
                    setMonsterForm((prev) => ({
                      ...prev,
                      description: event.target.value,
                    }))
                  }
                />
              </label>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Zone</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.zoneId}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        zoneId: event.target.value,
                      }))
                    }
                  >
                    <option value="">Select zone</option>
                    {zones.map((zone) => (
                      <option key={zone.id} value={zone.id}>
                        {zone.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Dominant Hand</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.dominantHandInventoryItemId}
                    onChange={(event) =>
                      updateDominantHandSelection(event.target.value)
                    }
                  >
                    <option value="">Select dominant hand item</option>
                    {dominantHandItems.map((item) => (
                      <option key={item.id} value={String(item.id)}>
                        {item.name} ({item.damageMin}-{item.damageMax} ·{' '}
                        {item.swipesPerAttack ?? 1} swipes)
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Off Hand</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.offHandInventoryItemId}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        offHandInventoryItemId: event.target.value,
                      }))
                    }
                    disabled={dominantIsTwoHanded}
                  >
                    <option value="">None</option>
                    {offHandItems.map((item) => (
                      <option key={item.id} value={String(item.id)}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Latitude</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.latitude}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        latitude: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Longitude</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.longitude}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        longitude: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Level</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.level}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        level: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="text-sm">
                <button
                  type="button"
                  className="qa-btn qa-btn-secondary"
                  onClick={handleUseCurrentLocation}
                  disabled={geoLoading}
                >
                  {geoLoading ? 'Locating...' : 'Use Current Browser Location'}
                </button>
              </div>

              <div className="text-sm">
                <div className="flex items-center justify-between mb-1">
                  <span>Map Location Picker</span>
                  <span className="text-xs text-gray-500">
                    Click map to set latitude/longitude
                  </span>
                </div>
                {mapboxgl.accessToken ? (
                  <div
                    ref={mapContainerRef}
                    className="w-full h-64 border border-gray-300 rounded-md"
                  />
                ) : (
                  <div className="w-full border border-gray-300 rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                    Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                    unavailable.
                  </div>
                )}
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Image URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.imageUrl}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        imageUrl: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Thumbnail URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.thumbnailUrl}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        thumbnailUrl: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Reward Experience</span>
                  <input
                    type="number"
                    min={0}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.rewardExperience}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        rewardExperience: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Reward Gold</span>
                  <input
                    type="number"
                    min={0}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.rewardGold}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        rewardGold: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="border border-gray-200 rounded-md p-3 space-y-2">
                <div className="flex items-center justify-between">
                  <h3 className="font-medium">Item Rewards</h3>
                  <button
                    type="button"
                    className="qa-btn qa-btn-secondary"
                    onClick={() =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        itemRewards: [
                          ...prev.itemRewards,
                          { inventoryItemId: '', quantity: '1' },
                        ],
                      }))
                    }
                  >
                    Add Item
                  </button>
                </div>
                {monsterForm.itemRewards.length === 0 ? (
                  <p className="text-sm text-gray-500">
                    No item rewards configured.
                  </p>
                ) : (
                  monsterForm.itemRewards.map((reward, index) => (
                    <div
                      key={`${reward.inventoryItemId}-${index}`}
                      className="grid grid-cols-[1fr_120px_auto] gap-2"
                    >
                      <select
                        className="border border-gray-300 rounded-md p-2"
                        value={reward.inventoryItemId}
                        onChange={(event) =>
                          updateMonsterItemReward(index, {
                            inventoryItemId: event.target.value,
                          })
                        }
                      >
                        <option value="">Select item</option>
                        {inventoryItems.map((item) => (
                          <option key={item.id} value={String(item.id)}>
                            {item.name}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        min={1}
                        className="border border-gray-300 rounded-md p-2"
                        value={reward.quantity}
                        onChange={(event) =>
                          updateMonsterItemReward(index, {
                            quantity: event.target.value,
                          })
                        }
                      />
                      <button
                        type="button"
                        className="qa-btn qa-btn-danger"
                        onClick={() =>
                          setMonsterForm((prev) => ({
                            ...prev,
                            itemRewards: prev.itemRewards.filter(
                              (_, i) => i !== index
                            ),
                          }))
                        }
                      >
                        Remove
                      </button>
                    </div>
                  ))
                )}
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  className="qa-btn qa-btn-secondary"
                  onClick={closeMonsterModal}
                >
                  Cancel
                </button>
                <button className="qa-btn qa-btn-primary" onClick={saveMonster}>
                  Save Monster
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {imagePreview ? (
        <div
          className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-[60]"
          onClick={closeImagePreview}
        >
          <div
            className="relative bg-white rounded-lg shadow-xl p-3 max-w-5xl w-full max-h-[92vh] flex items-center justify-center"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              className="absolute top-2 right-2 qa-btn qa-btn-secondary"
              onClick={closeImagePreview}
            >
              Close
            </button>
            <img
              src={imagePreview.url}
              alt={imagePreview.alt}
              className="max-w-full max-h-[84vh] rounded object-contain"
            />
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default Monsters;
