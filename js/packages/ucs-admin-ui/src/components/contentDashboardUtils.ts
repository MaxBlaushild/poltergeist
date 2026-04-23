import { useAPI } from '@poltergeist/contexts';
import { useEffect, useMemo, useState } from 'react';
import type { ContentDashboardBucket } from './ContentDashboard.tsx';

type CountBucketsOptions = {
  emptyLabel?: string;
  limit?: number;
  remainderLabel?: string;
  includeEmpty?: boolean;
};

type PaginatedAdminResponse<T> = {
  items?: T[];
  total?: number;
};

const adminAggregatePageSize = 100;

const normalizeLabel = (value: string | null | undefined): string =>
  (value ?? '').trim();

const finalizeBuckets = (
  counts: Map<string, number>,
  options: CountBucketsOptions = {}
): ContentDashboardBucket[] => {
  const limit = options.limit;
  const remainderLabel = options.remainderLabel ?? 'Other';
  const buckets = Array.from(counts.entries())
    .map(([label, value]) => ({ label, value }))
    .sort((left, right) => {
      if (right.value !== left.value) {
        return right.value - left.value;
      }
      return left.label.localeCompare(right.label);
    });

  if (limit == null || limit <= 0 || buckets.length <= limit) {
    return buckets;
  }

  const visible = buckets.slice(0, Math.max(1, limit - 1));
  const remainder = buckets
    .slice(visible.length)
    .reduce((total, bucket) => total + bucket.value, 0);

  if (remainder > 0) {
    visible.push({ label: remainderLabel, value: remainder });
  }

  return visible;
};

export const countBy = <T,>(
  items: T[],
  getLabel: (item: T) => string | null | undefined,
  options: CountBucketsOptions = {}
): ContentDashboardBucket[] => {
  const counts = new Map<string, number>();
  const emptyLabel = options.emptyLabel ?? 'Unassigned';

  items.forEach((item) => {
    const normalized = normalizeLabel(getLabel(item));
    const label = normalized || emptyLabel;
    counts.set(label, (counts.get(label) ?? 0) + 1);
  });

  return finalizeBuckets(counts, options);
};

export const countManyBy = <T,>(
  items: T[],
  getLabels: (item: T) => Array<string | null | undefined>,
  options: CountBucketsOptions = {}
): ContentDashboardBucket[] => {
  const counts = new Map<string, number>();
  const emptyLabel = options.emptyLabel ?? 'None';
  const includeEmpty = options.includeEmpty ?? true;

  items.forEach((item) => {
    const labels = getLabels(item)
      .map((label) => normalizeLabel(label))
      .filter(Boolean);
    const uniqueLabels = Array.from(new Set(labels));
    if (uniqueLabels.length === 0) {
      if (!includeEmpty) {
        return;
      }
      counts.set(emptyLabel, (counts.get(emptyLabel) ?? 0) + 1);
      return;
    }
    uniqueLabels.forEach((label) => {
      counts.set(label, (counts.get(label) ?? 0) + 1);
    });
  });

  return finalizeBuckets(counts, options);
};

export const difficultyBandLabel = (
  value: number | null | undefined
): string => {
  if (!Number.isFinite(value)) {
    return 'Unrated';
  }
  if ((value ?? 0) <= 2) {
    return '0-2';
  }
  if ((value ?? 0) <= 5) {
    return '3-5';
  }
  if ((value ?? 0) <= 8) {
    return '6-8';
  }
  return '9+';
};

export const useAdminAggregateDataset = <T,>(
  endpoint: string,
  params: Record<string, string | number | undefined>,
  enabled = true
) => {
  const { apiClient } = useAPI();
  const [items, setItems] = useState<T[]>([]);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);

  const paramsKey = useMemo(
    () =>
      JSON.stringify(
        Object.entries(params).sort(([leftKey], [rightKey]) =>
          leftKey.localeCompare(rightKey)
        )
      ),
    [params]
  );

  useEffect(() => {
    let cancelled = false;

    if (!enabled) {
      setItems([]);
      setLoading(false);
      setError(null);
      return () => {
        cancelled = true;
      };
    }

    const loadAllPages = async () => {
      setLoading(true);
      setError(null);

      try {
        const firstPage = await apiClient.get<PaginatedAdminResponse<T>>(
          endpoint,
          {
            ...params,
            page: 1,
            pageSize: adminAggregatePageSize,
          }
        );
        const firstItems = Array.isArray(firstPage?.items)
          ? firstPage.items
          : [];
        const total = Number.isFinite(Number(firstPage?.total))
          ? Number(firstPage?.total)
          : firstItems.length;
        const totalPages = Math.max(
          1,
          Math.ceil(total / adminAggregatePageSize)
        );

        let nextItems = firstItems;
        if (totalPages > 1) {
          const remainingResponses = await Promise.all(
            Array.from({ length: totalPages - 1 }, (_, index) =>
              apiClient.get<PaginatedAdminResponse<T>>(endpoint, {
                ...params,
                page: index + 2,
                pageSize: adminAggregatePageSize,
              })
            )
          );
          nextItems = [
            ...firstItems,
            ...remainingResponses.flatMap((response) =>
              Array.isArray(response?.items) ? response.items : []
            ),
          ];
        }

        if (cancelled) {
          return;
        }
        setItems(nextItems);
      } catch (nextError) {
        if (cancelled) {
          return;
        }
        console.error(`Failed to load aggregate dataset for ${endpoint}`, nextError);
        setError(
          nextError instanceof Error
            ? nextError.message
            : 'Failed to load aggregate data.'
        );
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void loadAllPages();

    return () => {
      cancelled = true;
    };
  }, [apiClient, enabled, endpoint, params, paramsKey]);

  return {
    items,
    loading,
    error,
  };
};
