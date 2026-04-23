import { useAPI } from '@poltergeist/contexts';
import { ZoneKind } from '@poltergeist/types';
import { useEffect, useMemo, useState } from 'react';

export const normalizeZoneKindSlug = (
  value: string | null | undefined
): string => (value ?? '').trim();

export const zoneKindLabel = (
  slug: string | null | undefined,
  zoneKindBySlug: Map<string, ZoneKind>
): string => {
  const normalizedSlug = normalizeZoneKindSlug(slug);
  if (!normalizedSlug) return 'Unassigned';
  return zoneKindBySlug.get(normalizedSlug)?.name?.trim() || normalizedSlug;
};

export const effectiveZoneKindSlug = (
  explicitZoneKind: string | null | undefined,
  zoneDefaultKind: string | null | undefined
): string =>
  normalizeZoneKindSlug(explicitZoneKind) ||
  normalizeZoneKindSlug(zoneDefaultKind);

export const zoneKindSummaryLabel = (
  explicitZoneKind: string | null | undefined,
  zoneDefaultKind: string | null | undefined,
  zoneKindBySlug: Map<string, ZoneKind>
): string => {
  const explicit = normalizeZoneKindSlug(explicitZoneKind);
  if (explicit) {
    return zoneKindLabel(explicit, zoneKindBySlug);
  }
  const zoneDefault = normalizeZoneKindSlug(zoneDefaultKind);
  if (zoneDefault) {
    return `${zoneKindLabel(zoneDefault, zoneKindBySlug)} (zone default)`;
  }
  return 'Unassigned';
};

export const zoneKindDescription = (
  explicitZoneKind: string | null | undefined,
  zoneDefaultKind: string | null | undefined,
  zoneKindBySlug: Map<string, ZoneKind>
): string => {
  const explicit = normalizeZoneKindSlug(explicitZoneKind);
  const details =
    zoneKindBySlug.get(effectiveZoneKindSlug(explicit, zoneDefaultKind)) ??
    null;
  const description = details?.description?.trim() ?? '';
  if (!description) return '';
  if (explicit) return description;
  return `Using zone default: ${description}`;
};

export const zoneKindSelectPlaceholderLabel = (
  zoneDefaultKind: string | null | undefined,
  zoneKindBySlug: Map<string, ZoneKind>
): string => {
  const normalizedZoneDefaultKind = normalizeZoneKindSlug(zoneDefaultKind);
  if (!normalizedZoneDefaultKind) return 'No zone kind';
  return `Use zone default (${zoneKindLabel(
    normalizedZoneDefaultKind,
    zoneKindBySlug
  )})`;
};

export const useZoneKinds = () => {
  const { apiClient } = useAPI();
  const [zoneKinds, setZoneKinds] = useState<ZoneKind[]>([]);

  useEffect(() => {
    let active = true;
    const loadZoneKinds = async () => {
      try {
        const response = await apiClient.get<ZoneKind[]>('/sonar/zoneKinds');
        if (!active) return;
        setZoneKinds(Array.isArray(response) ? response : []);
      } catch (error) {
        console.error('Error loading zone kinds:', error);
        if (!active) return;
        setZoneKinds([]);
      }
    };
    void loadZoneKinds();
    return () => {
      active = false;
    };
  }, [apiClient]);

  const zoneKindBySlug = useMemo(() => {
    const next = new Map<string, ZoneKind>();
    zoneKinds.forEach((zoneKind) => {
      const normalizedSlug = normalizeZoneKindSlug(zoneKind.slug);
      if (!normalizedSlug) return;
      next.set(normalizedSlug, zoneKind);
    });
    return next;
  }, [zoneKinds]);

  return { zoneKinds, zoneKindBySlug };
};
