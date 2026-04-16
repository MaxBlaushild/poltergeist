import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import { InventoryItem, Resource, ResourceType } from '@poltergeist/types';

type ResourceTypeRecord = ResourceType;
type ResourceRecord = Resource;
type ResourceTypeInventorySyncConflict = {
  inventoryItemId: number;
  inventoryItemName: string;
  matchingResourceTypes: string[];
};

type ResourceTypeInventorySyncSummary = {
  totalItemCount: number;
  updatedCount: number;
  alreadyMatchedCount: number;
  unmatchedCount: number;
  ambiguousCount: number;
  ambiguousItems: ResourceTypeInventorySyncConflict[];
};

type ResourceTypeRequirementGenerationResponse = {
  resourceType: ResourceTypeRecord;
  updatedResources: ResourceRecord[];
  createdItems: InventoryItem[];
  reusedItems: InventoryItem[];
  updatedResourceCount: number;
  generatedCount: number;
  reusedCount: number;
  message: string;
};

type ResourceGatherRequirementFormRow = {
  key: string;
  minLevel: string;
  maxLevel: string;
  requiredInventoryItemId: string;
};

const makeRequirementRow = (
  values: Partial<ResourceGatherRequirementFormRow> = {}
): ResourceGatherRequirementFormRow => ({
  key:
    values.key ||
    `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`,
  minLevel: values.minLevel ?? '1',
  maxLevel: values.maxLevel ?? '10',
  requiredInventoryItemId: values.requiredInventoryItemId ?? '',
});

const extractApiErrorMessage = (
  error: unknown,
  fallback: string
): string => {
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    typeof (error as { response?: unknown }).response === 'object'
  ) {
    const response = (error as { response?: { data?: unknown } }).response;
    const data = response?.data;
    if (typeof data === 'object' && data !== null) {
      const maybeMessage = (data as { error?: unknown; message?: unknown }).error;
      if (typeof maybeMessage === 'string' && maybeMessage.trim() !== '') {
        return maybeMessage;
      }
      const maybeFallback = (data as { message?: unknown }).message;
      if (typeof maybeFallback === 'string' && maybeFallback.trim() !== '') {
        return maybeFallback;
      }
    }
  }
  if (error instanceof Error && error.message.trim() !== '') {
    return error.message;
  }
  return fallback;
};

const slugify = (value: string) =>
  value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');

const formatCoordinates = (latitude: number, longitude: number) =>
  `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;

const formatResourceTypeInventorySyncSummary = (
  summary: ResourceTypeInventorySyncSummary
) => {
  const parts = [
    `${summary.updatedCount} updated`,
    `${summary.alreadyMatchedCount} already matched`,
    `${summary.unmatchedCount} without matching tags`,
  ];
  if (summary.ambiguousCount > 0) {
    parts.push(`${summary.ambiguousCount} ambiguous`);
  }

  let message = `Synced inventory item resource types: ${parts.join(', ')}.`;
  if (summary.ambiguousItems.length > 0) {
    message += ` Skipped ambiguous items: ${summary.ambiguousItems
      .slice(0, 3)
      .map(
        (item) =>
          `${item.inventoryItemName} (${item.matchingResourceTypes.join(', ')})`
      )
      .join('; ')}.`;
  }
  return message;
};

const emptyResourceTypeForm = () => ({
  name: '',
  slug: '',
  description: '',
  mapIconPrompt: '',
  gatherRequirements: [] as ResourceGatherRequirementFormRow[],
});

const emptyResourceForm = () => ({
  zoneId: '',
  resourceTypeId: '',
  quantity: '1',
  latitude: '',
  longitude: '',
});

const resourceTypeToForm = (resourceType: ResourceTypeRecord) => ({
  name: resourceType.name,
  slug: resourceType.slug,
  description: resourceType.description,
  mapIconPrompt: resourceType.mapIconPrompt || '',
  gatherRequirements: (resourceType.gatherRequirements || []).map(
    (requirement) =>
      makeRequirementRow({
        minLevel: requirement.minLevel.toString(),
        maxLevel: requirement.maxLevel.toString(),
        requiredInventoryItemId: requirement.requiredInventoryItemId.toString(),
      })
  ),
});

const resourceToForm = (resource: ResourceRecord) => ({
  zoneId: resource.zoneId,
  resourceTypeId: resource.resourceTypeId,
  quantity: resource.quantity.toString(),
  latitude: resource.latitude.toString(),
  longitude: resource.longitude.toString(),
});

const parseGatherRequirementRows = (
  rows: ResourceGatherRequirementFormRow[]
): {
  gatherRequirements: Array<{
    minLevel: number;
    maxLevel: number;
    requiredInventoryItemId: number;
  }>;
  error?: string;
} => {
  const gatherRequirements = [];
  for (const [index, requirement] of rows.entries()) {
    const minLevel = Number(requirement.minLevel);
    const maxLevel = Number(requirement.maxLevel);
    const requiredInventoryItemId = Number(requirement.requiredInventoryItemId);
    if (!Number.isFinite(minLevel) || minLevel < 1) {
      return {
        gatherRequirements: [],
        error: `Gather requirement ${index + 1} needs a minimum level of 1 or higher.`,
      };
    }
    if (!Number.isFinite(maxLevel) || maxLevel < minLevel) {
      return {
        gatherRequirements: [],
        error: `Gather requirement ${index + 1} needs a max level greater than or equal to its min level.`,
      };
    }
    if (!Number.isFinite(requiredInventoryItemId) || requiredInventoryItemId < 1) {
      return {
        gatherRequirements: [],
        error: `Gather requirement ${index + 1} needs a required inventory item.`,
      };
    }
    gatherRequirements.push({
      minLevel,
      maxLevel,
      requiredInventoryItemId,
    });
  }
  return { gatherRequirements };
};

export const Resources = () => {
  const { apiClient } = useAPI();
  const { inventoryItems, refreshInventoryItems } = useInventory();
  const { zones } = useZoneContext();
  const [resourceTypes, setResourceTypes] = useState<ResourceTypeRecord[]>([]);
  const [resources, setResources] = useState<ResourceRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [editingResourceTypeId, setEditingResourceTypeId] = useState('');
  const [resourceTypeForm, setResourceTypeForm] = useState(
    emptyResourceTypeForm()
  );
  const [editingResourceId, setEditingResourceId] = useState('');
  const [resourceForm, setResourceForm] = useState(emptyResourceForm());

  const resourceTypeById = useMemo(() => {
    const map = new Map<string, ResourceTypeRecord>();
    resourceTypes.forEach((resourceType) => map.set(resourceType.id, resourceType));
    return map;
  }, [resourceTypes]);

  const zoneNameById = useMemo(() => {
    const map = new Map<string, string>();
    zones.forEach((zone) => map.set(zone.id, zone.name || zone.id));
    return map;
  }, [zones]);

  const resourceCountByTypeId = useMemo(() => {
    const map = new Map<string, number>();
    resources.forEach((resource) => {
      map.set(
        resource.resourceTypeId,
        (map.get(resource.resourceTypeId) || 0) + 1
      );
    });
    return map;
  }, [resources]);

  const inventoryItemById = useMemo(() => {
    const map = new Map<number, InventoryItem>();
    inventoryItems.forEach((item) => map.set(item.id, item));
    return map;
  }, [inventoryItems]);

  const fetchResourceTypes = useCallback(async () => {
    const response = await apiClient.get<ResourceTypeRecord[]>(
      '/sonar/resource-types'
    );
    setResourceTypes(Array.isArray(response) ? response : []);
  }, [apiClient]);

  const fetchResources = useCallback(async () => {
    const response = await apiClient.get<ResourceRecord[]>('/sonar/resources');
    setResources(Array.isArray(response) ? response : []);
  }, [apiClient]);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      await Promise.all([fetchResourceTypes(), fetchResources()]);
    } catch (nextError) {
      console.error('Failed to load resources admin data:', nextError);
      setError(extractApiErrorMessage(nextError, 'Failed to load resources.'));
    } finally {
      setLoading(false);
    }
  }, [fetchResourceTypes, fetchResources]);

  useEffect(() => {
    void fetchAll();
  }, [fetchAll]);

  const resetResourceTypeForm = useCallback(() => {
    setEditingResourceTypeId('');
    setResourceTypeForm(emptyResourceTypeForm());
  }, []);

  const resetResourceForm = useCallback(() => {
    setEditingResourceId('');
    setResourceForm(emptyResourceForm());
  }, []);

  const handleSaveResourceType = useCallback(async () => {
    const name = resourceTypeForm.name.trim();
    if (!name) {
      setError('Resource type name is required.');
      return;
    }
    const parsedRequirements = parseGatherRequirementRows(
      resourceTypeForm.gatherRequirements
    );
    if (parsedRequirements.error) {
      setError(parsedRequirements.error);
      return;
    }

    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      const payload = {
        name,
        slug:
          resourceTypeForm.slug.trim() === ''
            ? slugify(name)
            : slugify(resourceTypeForm.slug),
        description: resourceTypeForm.description.trim(),
        mapIconPrompt: resourceTypeForm.mapIconPrompt.trim(),
        gatherRequirements: parsedRequirements.gatherRequirements,
      };

      if (editingResourceTypeId) {
        await apiClient.put<ResourceTypeRecord>(
          `/sonar/resource-types/${editingResourceTypeId}`,
          payload
        );
        setMessage('Resource type updated.');
      } else {
        await apiClient.post<ResourceTypeRecord>('/sonar/resource-types', payload);
        setMessage('Resource type created.');
      }
      await fetchResourceTypes();
      resetResourceTypeForm();
    } catch (nextError) {
      console.error('Failed to save resource type:', nextError);
      setError(
        extractApiErrorMessage(nextError, 'Failed to save resource type.')
      );
    } finally {
      setBusy(false);
    }
  }, [
    apiClient,
    editingResourceTypeId,
    fetchResourceTypes,
    resetResourceTypeForm,
    resourceTypeForm,
  ]);

  const handleEditResourceType = useCallback((resourceType: ResourceTypeRecord) => {
    setEditingResourceTypeId(resourceType.id);
    setResourceTypeForm(resourceTypeToForm(resourceType));
    setError(null);
    setMessage(null);
  }, []);

  const handleDeleteResourceType = useCallback(
    async (resourceType: ResourceTypeRecord) => {
      if (
        !window.confirm(
          `Delete resource type "${resourceType.name}"? This will fail if resources still depend on it.`
        )
      ) {
        return;
      }
      setBusy(true);
      setError(null);
      setMessage(null);
      try {
        await apiClient.delete(`/sonar/resource-types/${resourceType.id}`);
        setMessage('Resource type deleted.');
        await fetchResourceTypes();
        if (editingResourceTypeId === resourceType.id) {
          resetResourceTypeForm();
        }
      } catch (nextError) {
        console.error('Failed to delete resource type:', nextError);
        setError(
          extractApiErrorMessage(nextError, 'Failed to delete resource type.')
        );
      } finally {
        setBusy(false);
      }
    },
    [apiClient, editingResourceTypeId, fetchResourceTypes, resetResourceTypeForm]
  );

  const handleGenerateMapIcon = useCallback(
    async (resourceType: ResourceTypeRecord) => {
      setBusy(true);
      setError(null);
      setMessage(null);
      try {
        await apiClient.post<ResourceTypeRecord>(
          `/sonar/resource-types/${resourceType.id}/generate-map-icon`,
          {
            prompt:
              editingResourceTypeId === resourceType.id
                ? resourceTypeForm.mapIconPrompt.trim()
                : resourceType.mapIconPrompt,
          }
        );
        setMessage(`Generated a new map icon for ${resourceType.name}.`);
        await fetchResourceTypes();
      } catch (nextError) {
        console.error('Failed to generate resource type icon:', nextError);
        setError(
          extractApiErrorMessage(nextError, 'Failed to generate map icon.')
        );
      } finally {
        setBusy(false);
      }
    },
    [apiClient, editingResourceTypeId, fetchResourceTypes, resourceTypeForm.mapIconPrompt]
  );

  const handleGenerateRequirementItemsForType = useCallback(
    async (resourceType: ResourceTypeRecord) => {
      const resourceCount = resourceCountByTypeId.get(resourceType.id) || 0;
      const nodeSummary =
        resourceCount === 1 ? '1 resource node' : `${resourceCount} resource nodes`;
      if (
        !window.confirm(
          `Generate the recommended required items for ${resourceType.name} and apply the default requirement bands to ${nodeSummary}?`
        )
      ) {
        return;
      }

      setBusy(true);
      setError(null);
      setMessage(null);
      try {
        const response =
          await apiClient.post<ResourceTypeRequirementGenerationResponse>(
            `/sonar/resource-types/${resourceType.id}/generate-requirement-items`,
            {}
          );
        await fetchResourceTypes();
        await fetchResources();
        refreshInventoryItems();
        if (
          editingResourceTypeId === resourceType.id &&
          response.resourceType
        ) {
          setResourceTypeForm(resourceTypeToForm(response.resourceType));
        }
        if (editingResourceId) {
          const updatedEditingResource = response.updatedResources.find(
            (resource) => resource.id === editingResourceId
          );
          if (updatedEditingResource) {
            setResourceForm(resourceToForm(updatedEditingResource));
          }
        }
        setMessage(
          response.message ||
            `Generated required items for ${resourceType.name}.`
        );
      } catch (nextError) {
        console.error(
          'Failed to generate resource type requirement items:',
          nextError
        );
        setError(
          extractApiErrorMessage(
            nextError,
            'Failed to generate required items for this resource type.'
          )
        );
      } finally {
        setBusy(false);
      }
    },
    [
      apiClient,
      editingResourceId,
      editingResourceTypeId,
      fetchResourceTypes,
      fetchResources,
      refreshInventoryItems,
      resourceCountByTypeId,
    ]
  );

  const handleSyncInventoryItemResourceTypes = useCallback(async () => {
    if (
      !window.confirm(
        'Assign resource types to inventory items whose internal tags match a resource type name or slug? Existing matching assignments will be left alone.'
      )
    ) {
      return;
    }

    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      const summary = await apiClient.post<ResourceTypeInventorySyncSummary>(
        '/sonar/resource-types/sync-inventory-items',
        {}
      );
      refreshInventoryItems();
      setMessage(formatResourceTypeInventorySyncSummary(summary));
    } catch (nextError) {
      console.error(
        'Failed to sync resource types onto inventory items:',
        nextError
      );
      setError(
        extractApiErrorMessage(
          nextError,
          'Failed to sync resource types onto inventory items.'
        )
      );
    } finally {
      setBusy(false);
    }
  }, [apiClient, refreshInventoryItems]);

  const handleSaveResource = useCallback(async () => {
    if (!resourceForm.zoneId.trim()) {
      setError('Zone is required.');
      return;
    }
    if (!resourceForm.resourceTypeId.trim()) {
      setError('Resource type is required.');
      return;
    }

    const latitude = Number(resourceForm.latitude);
    const longitude = Number(resourceForm.longitude);
    const quantity = Number(resourceForm.quantity);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      setError('Latitude and longitude are required.');
      return;
    }
    if (!Number.isFinite(quantity) || quantity < 1) {
      setError('Quantity must be at least 1.');
      return;
    }

    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      const payload = {
        zoneId: resourceForm.zoneId,
        resourceTypeId: resourceForm.resourceTypeId,
        quantity,
        latitude,
        longitude,
      };
      if (editingResourceId) {
        await apiClient.put<ResourceRecord>(
          `/sonar/resources/${editingResourceId}`,
          payload
        );
        setMessage('Resource updated.');
      } else {
        await apiClient.post<ResourceRecord>('/sonar/resources', payload);
        setMessage('Resource created.');
      }
      await fetchResources();
      resetResourceForm();
    } catch (nextError) {
      console.error('Failed to save resource:', nextError);
      setError(extractApiErrorMessage(nextError, 'Failed to save resource.'));
    } finally {
      setBusy(false);
    }
  }, [
    apiClient,
    editingResourceId,
    fetchResources,
    resetResourceForm,
    resourceForm,
  ]);

  const handleEditResource = useCallback((resource: ResourceRecord) => {
    setEditingResourceId(resource.id);
    setResourceForm(resourceToForm(resource));
    setError(null);
    setMessage(null);
  }, []);

  const handleDeleteResource = useCallback(
    async (resource: ResourceRecord) => {
      if (
        !window.confirm(
          `Delete this ${resource.resourceType?.name || 'resource'} node?`
        )
      ) {
        return;
      }
      setBusy(true);
      setError(null);
      setMessage(null);
      try {
        await apiClient.delete(`/sonar/resources/${resource.id}`);
        setMessage('Resource deleted.');
        await fetchResources();
        if (editingResourceId === resource.id) {
          resetResourceForm();
        }
      } catch (nextError) {
        console.error('Failed to delete resource:', nextError);
        setError(extractApiErrorMessage(nextError, 'Failed to delete resource.'));
      } finally {
        setBusy(false);
      }
    },
    [apiClient, editingResourceId, fetchResources, resetResourceForm]
  );

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      setError('Geolocation is not supported in this browser.');
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setResourceForm((current) => ({
          ...current,
          latitude: position.coords.latitude.toFixed(6),
          longitude: position.coords.longitude.toFixed(6),
        }));
      },
      () => {
        setError('Unable to capture your current location.');
      },
      { enableHighAccuracy: true, timeout: 12000, maximumAge: 0 }
    );
  }, []);

  const updateResourceTypeGatherRequirementRow = useCallback(
    (
      key: string,
      updates: Partial<Omit<ResourceGatherRequirementFormRow, 'key'>>
    ) => {
      setResourceTypeForm((current) => ({
        ...current,
        gatherRequirements: current.gatherRequirements.map((row) =>
          row.key === key ? { ...row, ...updates } : row
        ),
      }));
    },
    []
  );

  const addResourceTypeGatherRequirementRow = useCallback(() => {
    setResourceTypeForm((current) => ({
      ...current,
      gatherRequirements: [...current.gatherRequirements, makeRequirementRow()],
    }));
  }, []);

  const removeResourceTypeGatherRequirementRow = useCallback((key: string) => {
    setResourceTypeForm((current) => ({
      ...current,
      gatherRequirements: current.gatherRequirements.filter(
        (row) => row.key !== key
      ),
    }));
  }, []);

  const formatRequirementSummary = useCallback(
    (
      requirementsInput:
        | ResourceRecord['gatherRequirements']
        | ResourceTypeRecord['gatherRequirements']
        | undefined
    ) => {
      const requirements = (requirementsInput || []).slice();
      requirements.sort((left, right) => left.minLevel - right.minLevel);
      return requirements.map((requirement) => {
        const itemName =
          requirement.requiredInventoryItem?.name ||
          inventoryItemById.get(requirement.requiredInventoryItemId)?.name ||
          `Item #${requirement.requiredInventoryItemId}`;
        return `Lv ${requirement.minLevel}-${requirement.maxLevel}: ${itemName}`;
      });
    },
    [inventoryItemById]
  );

  if (loading) {
    return <div className="p-6 text-sm text-slate-600">Loading resources…</div>;
  }

  return (
    <div className="space-y-8">
      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-slate-900">
              Resource Types
            </h2>
            <p className="mt-1 text-sm text-slate-600">
              Define gathering categories like herbalism or mining and generate
              the map icon each resource node will use.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => void handleSyncInventoryItemResourceTypes()}
              className="rounded-md border border-emerald-200 px-3 py-2 text-sm font-medium text-emerald-700"
              disabled={busy || resourceTypes.length === 0}
            >
              Auto-Assign To Inventory Items
            </button>
            <button
              type="button"
              onClick={() => void fetchResourceTypes()}
              className="rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-700"
            >
              Refresh
            </button>
          </div>
        </div>

        {(message || error) && (
          <div
            className={`mb-4 rounded-lg px-4 py-3 text-sm ${
              error
                ? 'bg-red-50 text-red-700'
                : 'bg-emerald-50 text-emerald-700'
            }`}
          >
            {error || message}
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,420px)]">
          <div className="grid gap-4 md:grid-cols-2">
            {resourceTypes.map((resourceType) => (
              <article
                key={resourceType.id}
                className="rounded-xl border border-slate-200 p-4"
              >
                <div className="mb-3 flex items-start justify-between gap-3">
                  <div>
                    <h3 className="font-semibold text-slate-900">
                      {resourceType.name}
                    </h3>
                    <p className="text-xs uppercase tracking-wide text-slate-500">
                      {resourceType.slug}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => handleEditResourceType(resourceType)}
                      className="rounded-md border border-slate-300 px-2 py-1 text-xs text-slate-700"
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      onClick={() => void handleDeleteResourceType(resourceType)}
                      className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600"
                    >
                      Delete
                    </button>
                  </div>
                </div>
                <div className="mb-3 aspect-square overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
                  {resourceType.mapIconUrl ? (
                    <img
                      src={resourceType.mapIconUrl}
                      alt={resourceType.name}
                      className="h-full w-full object-contain"
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center text-sm text-slate-400">
                      No map icon yet
                    </div>
                  )}
                </div>
                <p className="mb-3 text-sm text-slate-600">
                  {resourceType.description || 'No description yet.'}
                </p>
                {resourceType.gatherRequirements &&
                  resourceType.gatherRequirements.length > 0 && (
                    <div className="mb-3 flex flex-wrap gap-2 text-xs text-slate-600">
                      {formatRequirementSummary(
                        resourceType.gatherRequirements
                      ).map((summary) => (
                        <span
                          key={summary}
                          className="rounded-full bg-amber-50 px-2 py-1 text-amber-700"
                        >
                          Requires {summary}
                        </span>
                      ))}
                    </div>
                  )}
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => void handleGenerateMapIcon(resourceType)}
                    className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white"
                    disabled={busy}
                  >
                    Generate Map Icon
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      void handleGenerateRequirementItemsForType(resourceType)
                    }
                    className="rounded-md border border-amber-200 px-3 py-2 text-sm text-amber-700"
                    disabled={busy}
                  >
                    Generate Required Items
                  </button>
                </div>
                <p className="mt-3 text-xs text-slate-500">
                  Applies the default tool bands to all current and future{' '}
                  {resourceCountByTypeId.get(resourceType.id) || 0} node
                  {(resourceCountByTypeId.get(resourceType.id) || 0) === 1
                    ? ''
                    : 's'}{' '}
                  of this type at levels 1, 21, 41, 61, and 81.
                </p>
              </article>
            ))}
          </div>

          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="font-semibold text-slate-900">
                {editingResourceTypeId ? 'Edit Resource Type' : 'New Resource Type'}
              </h3>
              {editingResourceTypeId && (
                <button
                  type="button"
                  onClick={resetResourceTypeForm}
                  className="text-sm text-slate-600"
                >
                  Cancel
                </button>
              )}
            </div>

            <div className="space-y-3">
              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">Name</span>
                <input
                  value={resourceTypeForm.name}
                  onChange={(event) =>
                    setResourceTypeForm((current) => ({
                      ...current,
                      name: event.target.value,
                    }))
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2"
                  placeholder="Herbalism"
                />
              </label>

              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">Slug</span>
                <input
                  value={resourceTypeForm.slug}
                  onChange={(event) =>
                    setResourceTypeForm((current) => ({
                      ...current,
                      slug: event.target.value,
                    }))
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2"
                  placeholder="herbalism"
                />
              </label>

              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">
                  Description
                </span>
                <textarea
                  value={resourceTypeForm.description}
                  onChange={(event) =>
                    setResourceTypeForm((current) => ({
                      ...current,
                      description: event.target.value,
                    }))
                  }
                  rows={3}
                  className="w-full rounded-md border border-slate-300 px-3 py-2"
                />
              </label>

              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">
                  Map Icon Prompt
                </span>
                <textarea
                  value={resourceTypeForm.mapIconPrompt}
                  onChange={(event) =>
                    setResourceTypeForm((current) => ({
                      ...current,
                      mapIconPrompt: event.target.value,
                    }))
                  }
                  rows={5}
                  className="w-full rounded-md border border-slate-300 px-3 py-2"
                />
              </label>

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="block text-sm font-medium text-slate-700">
                    Gather Requirements
                  </span>
                  <button
                    type="button"
                    onClick={addResourceTypeGatherRequirementRow}
                    className="rounded-md border border-slate-300 px-3 py-1 text-xs text-slate-700"
                  >
                    Add Requirement
                  </button>
                </div>
                <p className="text-xs text-slate-500">
                  These level bands apply to every current and future node of
                  this resource type. If no band matches, no tool is required.
                </p>
                {resourceTypeForm.gatherRequirements.length === 0 ? (
                  <div className="rounded-md border border-dashed border-slate-300 px-3 py-3 text-sm text-slate-500">
                    No gather requirements configured.
                  </div>
                ) : (
                  <div className="space-y-3">
                    {resourceTypeForm.gatherRequirements.map((requirement) => (
                      <div
                        key={requirement.key}
                        className="grid gap-3 rounded-lg border border-slate-200 bg-white p-3"
                      >
                        <div className="grid gap-3 sm:grid-cols-[100px_100px_minmax(0,1fr)_auto]">
                          <label className="block text-sm">
                            <span className="mb-1 block font-medium text-slate-700">
                              Min Lv
                            </span>
                            <input
                              value={requirement.minLevel}
                              onChange={(event) =>
                                updateResourceTypeGatherRequirementRow(
                                  requirement.key,
                                  {
                                    minLevel: event.target.value,
                                  }
                                )
                              }
                              inputMode="numeric"
                              className="w-full rounded-md border border-slate-300 px-3 py-2"
                            />
                          </label>
                          <label className="block text-sm">
                            <span className="mb-1 block font-medium text-slate-700">
                              Max Lv
                            </span>
                            <input
                              value={requirement.maxLevel}
                              onChange={(event) =>
                                updateResourceTypeGatherRequirementRow(
                                  requirement.key,
                                  {
                                    maxLevel: event.target.value,
                                  }
                                )
                              }
                              inputMode="numeric"
                              className="w-full rounded-md border border-slate-300 px-3 py-2"
                            />
                          </label>
                          <label className="block text-sm">
                            <span className="mb-1 block font-medium text-slate-700">
                              Required Item
                            </span>
                            <select
                              value={requirement.requiredInventoryItemId}
                              onChange={(event) =>
                                updateResourceTypeGatherRequirementRow(
                                  requirement.key,
                                  {
                                    requiredInventoryItemId: event.target.value,
                                  }
                                )
                              }
                              className="w-full rounded-md border border-slate-300 px-3 py-2"
                            >
                              <option value="">Select an inventory item</option>
                              {inventoryItems.map((item) => (
                                <option key={item.id} value={item.id}>
                                  {item.name}
                                </option>
                              ))}
                            </select>
                          </label>
                          <div className="flex items-end">
                            <button
                              type="button"
                              onClick={() =>
                                removeResourceTypeGatherRequirementRow(
                                  requirement.key
                                )
                              }
                              className="rounded-md border border-red-200 px-3 py-2 text-sm text-red-600"
                            >
                              Remove
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <button
                type="button"
                onClick={() => void handleSaveResourceType()}
                className="w-full rounded-md bg-emerald-600 px-4 py-2 font-medium text-white"
                disabled={busy}
              >
                {editingResourceTypeId ? 'Save Resource Type' : 'Create Resource Type'}
              </button>
            </div>
          </div>
        </div>
      </section>

      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-slate-900">Resources</h2>
            <p className="mt-1 text-sm text-slate-600">
              Place gatherable resource nodes on the map. Each gather grants a
              random active inventory item of that type within about 10 item
              levels of the player.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void fetchResources()}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-700"
          >
            Refresh
          </button>
        </div>

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,420px)]">
          <div className="space-y-3">
            {resources.map((resource) => (
              <article
                key={resource.id}
                className="flex items-start gap-4 rounded-xl border border-slate-200 p-4"
              >
                <div className="flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
                  {resource.resourceType?.mapIconUrl ? (
                    <img
                      src={resource.resourceType.mapIconUrl}
                      alt={resource.resourceType?.name || 'Resource'}
                      className="h-full w-full object-contain"
                    />
                  ) : (
                    <span className="px-2 text-center text-xs text-slate-400">
                      No icon
                    </span>
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="font-semibold text-slate-900">
                        {resource.resourceType?.name ||
                          resourceTypeById.get(resource.resourceTypeId)?.name ||
                          'Resource'}{' '}
                        Node
                      </h3>
                      <p className="text-sm text-slate-500">
                        {resource.resourceType?.name ||
                          resourceTypeById.get(resource.resourceTypeId)?.name ||
                          'Unknown type'}{' '}
                        in {zoneNameById.get(resource.zoneId) || resource.zoneId}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => handleEditResource(resource)}
                        className="rounded-md border border-slate-300 px-2 py-1 text-xs text-slate-700"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => void handleDeleteResource(resource)}
                        className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                  <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-600">
                    <span className="rounded-full bg-slate-100 px-2 py-1">
                      Qty {resource.quantity}
                    </span>
                    <span className="rounded-full bg-slate-100 px-2 py-1">
                      Reward band: player level +/- 10
                    </span>
                    <span className="rounded-full bg-slate-100 px-2 py-1">
                      {formatCoordinates(resource.latitude, resource.longitude)}
                    </span>
                  </div>
                  <p className="mt-2 text-xs text-slate-500">
                    Requirement tools are generated from the matching resource
                    type card and then applied across all nodes of that type.
                  </p>
                  {resource.gatherRequirements &&
                    resource.gatherRequirements.length > 0 && (
                      <div className="mt-3 flex flex-wrap gap-2 text-xs text-slate-600">
                        {formatRequirementSummary(
                          resource.gatherRequirements
                        ).map((summary) => (
                          <span
                            key={summary}
                            className="rounded-full bg-amber-50 px-2 py-1 text-amber-700"
                          >
                            Requires {summary}
                          </span>
                        ))}
                      </div>
                    )}
                </div>
              </article>
            ))}
            {resources.length === 0 && (
              <div className="rounded-xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
                No resources yet.
              </div>
            )}
          </div>

          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="font-semibold text-slate-900">
                {editingResourceId ? 'Edit Resource' : 'New Resource'}
              </h3>
              {editingResourceId && (
                <button
                  type="button"
                  onClick={resetResourceForm}
                  className="text-sm text-slate-600"
                >
                  Cancel
                </button>
              )}
            </div>

            <div className="space-y-3">
              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">Zone</span>
                <select
                  value={resourceForm.zoneId}
                  onChange={(event) =>
                    setResourceForm((current) => ({
                      ...current,
                      zoneId: event.target.value,
                    }))
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2"
                >
                  <option value="">Select a zone</option>
                  {zones.map((zone) => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">
                  Resource Type
                </span>
                <select
                  value={resourceForm.resourceTypeId}
                  onChange={(event) =>
                    setResourceForm((current) => ({
                      ...current,
                      resourceTypeId: event.target.value,
                    }))
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2"
                >
                  <option value="">Select a resource type</option>
                  {resourceTypes.map((resourceType) => (
                    <option key={resourceType.id} value={resourceType.id}>
                      {resourceType.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm">
                <span className="mb-1 block font-medium text-slate-700">
                  Reward Behavior
                </span>
                <div className="rounded-md border border-slate-200 bg-slate-100 px-3 py-2 text-sm text-slate-600">
                  Players gather a random active inventory item with this
                  resource type, targeting item levels within +/- 10 of their
                  current level.
                </div>
              </label>

              <div className="rounded-md border border-slate-200 bg-slate-100 px-3 py-2 text-sm text-slate-600">
                Gather requirements are inherited from the selected resource
                type. Edit the matching resource type above to change the tool
                bands for all nodes of that type.
              </div>

              <div className="grid gap-3 sm:grid-cols-3">
                <label className="block text-sm">
                  <span className="mb-1 block font-medium text-slate-700">
                    Quantity
                  </span>
                  <input
                    value={resourceForm.quantity}
                    onChange={(event) =>
                      setResourceForm((current) => ({
                        ...current,
                        quantity: event.target.value,
                      }))
                    }
                    className="w-full rounded-md border border-slate-300 px-3 py-2"
                    inputMode="numeric"
                  />
                </label>
                <label className="block text-sm sm:col-span-2">
                  <span className="mb-1 block font-medium text-slate-700">
                    Coordinates
                  </span>
                  <div className="grid gap-3 sm:grid-cols-2">
                    <input
                      value={resourceForm.latitude}
                      onChange={(event) =>
                        setResourceForm((current) => ({
                          ...current,
                          latitude: event.target.value,
                        }))
                      }
                      className="w-full rounded-md border border-slate-300 px-3 py-2"
                      placeholder="Latitude"
                    />
                    <input
                      value={resourceForm.longitude}
                      onChange={(event) =>
                        setResourceForm((current) => ({
                          ...current,
                          longitude: event.target.value,
                        }))
                      }
                      className="w-full rounded-md border border-slate-300 px-3 py-2"
                      placeholder="Longitude"
                    />
                  </div>
                </label>
              </div>

              <button
                type="button"
                onClick={handleUseCurrentLocation}
                className="w-full rounded-md border border-slate-300 px-4 py-2 text-sm text-slate-700"
              >
                Use Current Browser Location
              </button>

              <button
                type="button"
                onClick={() => void handleSaveResource()}
                className="w-full rounded-md bg-emerald-600 px-4 py-2 font-medium text-white"
                disabled={busy}
              >
                {editingResourceId ? 'Save Resource' : 'Create Resource'}
              </button>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Resources;
