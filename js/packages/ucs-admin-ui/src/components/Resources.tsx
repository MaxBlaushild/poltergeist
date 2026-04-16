import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import { Resource, ResourceType } from '@poltergeist/types';

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
});

const emptyResourceForm = () => ({
  zoneId: '',
  resourceTypeId: '',
  quantity: '1',
  latitude: '',
  longitude: '',
});

export const Resources = () => {
  const { apiClient } = useAPI();
  const { refreshInventoryItems } = useInventory();
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
    setResourceTypeForm({
      name: resourceType.name,
      slug: resourceType.slug,
      description: resourceType.description,
      mapIconPrompt: resourceType.mapIconPrompt || '',
    });
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
    setResourceForm({
      zoneId: resource.zoneId,
      resourceTypeId: resource.resourceTypeId,
      quantity: resource.quantity.toString(),
      latitude: resource.latitude.toString(),
      longitude: resource.longitude.toString(),
    });
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
                <button
                  type="button"
                  onClick={() => void handleGenerateMapIcon(resourceType)}
                  className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white"
                  disabled={busy}
                >
                  Generate Map Icon
                </button>
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
