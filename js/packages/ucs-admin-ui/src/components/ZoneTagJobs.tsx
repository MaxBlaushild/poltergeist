import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Zone } from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

type ZoneTagGenerationJob = {
  id: string;
  zoneId: string;
  status: string;
  contextSnapshot?: string;
  generatedSummary?: string;
  selectedTags?: string[];
  errorMessage?: string;
  createdAt?: string;
  updatedAt?: string;
};

type BulkQueueZoneTagJobsResponse = {
  queuedCount: number;
  requestedZoneCount: number;
  jobs: ZoneTagGenerationJob[];
};

const sortBoundaryPoints = (points: [number, number][]): [number, number][] => {
  if (points.length < 3) return points.slice();

  const centroid = points.reduce(
    (acc, point) => [acc[0] + point[0] / points.length, acc[1] + point[1] / points.length],
    [0, 0]
  );

  const sorted = points
    .slice()
    .sort((a, b) => {
      const angleA = Math.atan2(a[1] - centroid[1], a[0] - centroid[0]);
      const angleB = Math.atan2(b[1] - centroid[1], b[0] - centroid[0]);
      return angleA - angleB;
    });

  const first = sorted[0];
  const last = sorted[sorted.length - 1];
  if (first[0] !== last[0] || first[1] !== last[1]) {
    sorted.push([first[0], first[1]]);
  }

  return sorted;
};

const getZoneRing = (zone: Zone): [number, number][] => {
  if (zone.points?.length) {
    return sortBoundaryPoints(
      zone.points.map((point) => [point.longitude, point.latitude] as [number, number])
    );
  }

  if (zone.boundaryCoords?.length) {
    return sortBoundaryPoints(
      zone.boundaryCoords.map((coord) => [coord.longitude, coord.latitude] as [number, number])
    );
  }

  return [];
};

type BulkZoneSelectionMapProps = {
  zones: Zone[];
  matchingZoneIds: Set<string>;
  selectedZoneIds: Set<string>;
  onToggleZone: (zoneId: string) => void;
};

const BulkZoneSelectionMap: React.FC<BulkZoneSelectionMapProps> = ({
  zones,
  matchingZoneIds,
  selectedZoneIds,
  onToggleZone,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const searchMarkerRef = useRef<mapboxgl.Marker | null>(null);
  const fitBoundsRef = useRef(false);
  const [mapLoaded, setMapLoaded] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [showSearchSuggestions, setShowSearchSuggestions] = useState(false);
  const [searchStatus, setSearchStatus] = useState<string | null>(null);
  const [searchCandidates, setSearchCandidates] = useState<
    Array<{
      id: string;
      center: [number, number];
      placeName: string;
    }>
  >([]);

  const focusMapOnLocation = useCallback((lngLat: [number, number], zoom: number) => {
    if (!mapRef.current) {
      return;
    }

    searchMarkerRef.current?.remove();
    searchMarkerRef.current = new mapboxgl.Marker({ color: '#2563EB' })
      .setLngLat(lngLat)
      .addTo(mapRef.current);

    mapRef.current.flyTo({
      center: lngLat,
      zoom: Math.max(mapRef.current.getZoom(), zoom),
      essential: true,
    });
  }, []);

  const handleSearchLocation = useCallback(async () => {
    const query = searchQuery.trim();
    if (!query) {
      setSearchStatus('Enter a city, neighborhood, address, or landmark.');
      return;
    }

    if (!mapboxgl.accessToken) {
      setSearchStatus('Location search is unavailable because the Mapbox token is missing.');
      return;
    }

    setIsSearching(true);
    setSearchStatus(null);
    setShowSearchSuggestions(false);

    try {
      const response = await fetch(
        `https://api.mapbox.com/geocoding/v5/mapbox.places/${encodeURIComponent(query)}.json?access_token=${encodeURIComponent(mapboxgl.accessToken)}&limit=6&types=country,region,postcode,district,place,locality,neighborhood,address,poi`
      );

      if (!response.ok) {
        throw new Error('Search request failed.');
      }

      const data = (await response.json()) as {
        features?: Array<{
          id?: string;
          center?: [number, number];
          place_name?: string;
        }>;
      };

      const candidates = (data.features ?? [])
        .filter((feature) => feature.center && feature.center.length >= 2)
        .map((feature, index) => ({
          id: feature.id || `${feature.place_name || 'candidate'}-${index}`,
          center: [feature.center![0], feature.center![1]] as [number, number],
          placeName: feature.place_name || 'Unknown location',
        }));

      if (candidates.length === 0) {
        setSearchCandidates([]);
        setSearchStatus(`No location match found for "${query}".`);
        return;
      }

      setSearchCandidates(candidates);
      setShowSearchSuggestions(true);
      setSearchStatus('Select a result below to move the map.');
    } catch (error) {
      console.error('Error searching for location:', error);
      setSearchCandidates([]);
      setSearchStatus('Unable to search for that location right now.');
    } finally {
      setIsSearching(false);
    }
  }, [searchQuery]);

  useEffect(() => {
    if (mapContainer.current && !mapRef.current) {
      mapRef.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: [-87.6298, 41.8781],
        zoom: 10,
        interactive: true,
      });

      mapRef.current.addControl(new mapboxgl.NavigationControl(), 'top-right');
      mapRef.current.on('load', () => setMapLoaded(true));
    }

    return () => {
      searchMarkerRef.current?.remove();
      searchMarkerRef.current = null;
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!mapRef.current || !mapLoaded) {
      return;
    }

    const features = zones
      .map((zone) => {
        const ring = getZoneRing(zone);
        if (ring.length < 4) {
          return null;
        }

        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Polygon' as const,
            coordinates: [ring],
          },
          properties: {
            id: zone.id,
            name: zone.name,
            selected: selectedZoneIds.has(zone.id),
            matching: matchingZoneIds.has(zone.id),
          },
        };
      })
      .filter(Boolean);

    const geojson = {
      type: 'FeatureCollection' as const,
      features: features as Array<GeoJSON.Feature<GeoJSON.Polygon>>,
    };

    const existingSource = mapRef.current.getSource('bulk-zone-selector') as
      | mapboxgl.GeoJSONSource
      | undefined;

    if (existingSource) {
      existingSource.setData(geojson);
    } else {
      mapRef.current.addSource('bulk-zone-selector', {
        type: 'geojson',
        data: geojson,
      });

      mapRef.current.addLayer({
        id: 'bulk-zone-selector-fill',
        type: 'fill',
        source: 'bulk-zone-selector',
        paint: {
          'fill-color': [
            'case',
            ['boolean', ['get', 'selected'], false],
            '#7c3aed',
            ['boolean', ['get', 'matching'], false],
            '#2563eb',
            '#94a3b8',
          ],
          'fill-opacity': [
            'case',
            ['boolean', ['get', 'selected'], false],
            0.45,
            ['boolean', ['get', 'matching'], false],
            0.18,
            0.08,
          ],
        },
      });

      mapRef.current.addLayer({
        id: 'bulk-zone-selector-outline',
        type: 'line',
        source: 'bulk-zone-selector',
        paint: {
          'line-color': [
            'case',
            ['boolean', ['get', 'selected'], false],
            '#5b21b6',
            ['boolean', ['get', 'matching'], false],
            '#1d4ed8',
            '#64748b',
          ],
          'line-width': [
            'case',
            ['boolean', ['get', 'selected'], false],
            3,
            1.5,
          ],
        },
      });

      mapRef.current.on('click', 'bulk-zone-selector-fill', (event) => {
        const feature = event.features?.[0];
        const zoneId = feature?.properties?.id;
        if (typeof zoneId === 'string' && zoneId) {
          onToggleZone(zoneId);
        }
      });

      mapRef.current.on('mouseenter', 'bulk-zone-selector-fill', () => {
        mapRef.current?.getCanvas().style.setProperty('cursor', 'pointer');
      });
      mapRef.current.on('mouseleave', 'bulk-zone-selector-fill', () => {
        mapRef.current?.getCanvas().style.setProperty('cursor', '');
      });
    }

    if (!fitBoundsRef.current && features.length > 0) {
      const bounds = new mapboxgl.LngLatBounds();
      features.forEach((feature) => {
        feature.geometry.coordinates[0].forEach((coord) => {
          bounds.extend(coord as [number, number]);
        });
      });
      mapRef.current.fitBounds(bounds, { padding: 36, maxZoom: 12 });
      fitBoundsRef.current = true;
    }
  }, [zones, selectedZoneIds, matchingZoneIds, mapLoaded, onToggleZone]);

  return (
    <div className="mt-4 rounded-lg border border-gray-200 bg-white p-3">
      <div className="mb-2">
        <div className="text-sm font-semibold text-gray-900">Visual zone selection</div>
        <div className="text-xs text-gray-500">
          Click zone polygons to add or remove them from the bulk queue.
        </div>
      </div>
      <div className="mb-3 flex flex-wrap gap-2">
        <div className="relative min-w-[240px] flex-1">
          <input
            className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
            value={searchQuery}
            onChange={(event) => {
              const value = event.target.value;
              setSearchQuery(value);
              setSearchStatus(null);
              if (value.trim() === '') {
                setSearchCandidates([]);
                setShowSearchSuggestions(false);
              }
            }}
            onFocus={() => {
              if (searchCandidates.length > 0) {
                setShowSearchSuggestions(true);
              }
            }}
            onBlur={() => {
              setTimeout(() => setShowSearchSuggestions(false), 120);
            }}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault();
                void handleSearchLocation();
              }
            }}
            placeholder="Search for a city, neighborhood, address, or landmark"
          />
          {showSearchSuggestions && searchCandidates.length > 0 && (
            <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
              {searchCandidates.map((candidate) => (
                <button
                  key={candidate.id}
                  type="button"
                  onClick={() => {
                    setSearchQuery(candidate.placeName);
                    setShowSearchSuggestions(false);
                    focusMapOnLocation(candidate.center, 13);
                    setSearchStatus(`Moved map to ${candidate.placeName}.`);
                  }}
                  className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                >
                  {candidate.placeName}
                </button>
              ))}
            </div>
          )}
        </div>
        <button
          type="button"
          onClick={() => void handleSearchLocation()}
          className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          disabled={isSearching}
        >
          {isSearching ? 'Searching...' : 'Search'}
        </button>
      </div>
      <div
        ref={mapContainer}
        className="h-[320px] w-full overflow-hidden rounded border border-gray-200"
      />
      {searchStatus && <p className="mt-2 text-xs text-gray-500">{searchStatus}</p>}
    </div>
  );
};

const statusBadgeClass = (status: string) => {
  switch (status) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    default:
      return 'bg-gray-600';
  }
};

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const ZoneTagJobs = () => {
  const { apiClient } = useAPI();
  const { zones, refreshZones } = useZoneContext();
  const [jobs, setJobs] = useState<ZoneTagGenerationJob[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [creatingBulkJobs, setCreatingBulkJobs] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [bulkZoneQuery, setBulkZoneQuery] = useState('');
  const [bulkSelectedZoneIds, setBulkSelectedZoneIds] = useState<string[]>([]);
  const [jobFilterZoneId, setJobFilterZoneId] = useState<string>('');
  const [filterZoneQuery, setFilterZoneQuery] = useState('');
  const [showFilterZoneSuggestions, setShowFilterZoneSuggestions] = useState(false);
  const lastCompletedJobIdRef = useRef<string | null>(null);

  const sortedZones = useMemo(
    () => [...zones].sort((left, right) => left.name.localeCompare(right.name)),
    [zones]
  );

  const zoneNameById = useMemo(() => {
    const entries = sortedZones.map((zone) => [zone.id, zone.name] as const);
    return new Map(entries);
  }, [sortedZones]);

  const filterZoneSuggestions = useMemo(() => {
    const query = filterZoneQuery.toLowerCase();
    return sortedZones.filter((zone) => zone.name.toLowerCase().includes(query));
  }, [sortedZones, filterZoneQuery]);

  const bulkMatchingZones = useMemo(() => {
    const query = bulkZoneQuery.trim().toLowerCase();
    if (!query) {
      return sortedZones;
    }
    return sortedZones.filter((zone) => zone.name.toLowerCase().includes(query));
  }, [sortedZones, bulkZoneQuery]);

  const bulkTargetZones = useMemo(() => {
    const selectedZoneIds = new Set(bulkSelectedZoneIds);
    return sortedZones.filter((zone) => selectedZoneIds.has(zone.id));
  }, [sortedZones, bulkSelectedZoneIds]);

  const bulkMatchingZoneIds = useMemo(
    () => new Set(bulkMatchingZones.map((zone) => zone.id)),
    [bulkMatchingZones]
  );

  const bulkSelectedZoneIdSet = useMemo(
    () => new Set(bulkSelectedZoneIds),
    [bulkSelectedZoneIds]
  );

  useEffect(() => {
    if (zones.length === 0) {
      refreshZones();
    }
  }, [zones.length, refreshZones]);

  useEffect(() => {
    const validZoneIds = new Set(sortedZones.map((zone) => zone.id));
    setBulkSelectedZoneIds((prev) => prev.filter((zoneId) => validZoneIds.has(zoneId)));
  }, [sortedZones]);

  const fetchJobs = useCallback(
    async (zoneId?: string) => {
      setLoadingJobs(true);
      setError(null);
      try {
        const query: Record<string, string | number> = { limit: 40 };
        if (zoneId) {
          query.zoneId = zoneId;
        }
        const response = await apiClient.get<ZoneTagGenerationJob[]>(
          '/sonar/admin/zone-tag-generation-jobs',
          query
        );
        setJobs(response);
      } catch (err) {
        console.error('Failed to load zone tag jobs', err);
        setError('Failed to load zone tag jobs.');
      } finally {
        setLoadingJobs(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    fetchJobs(jobFilterZoneId || undefined);
    const interval = window.setInterval(() => {
      fetchJobs(jobFilterZoneId || undefined);
    }, 5000);
    return () => window.clearInterval(interval);
  }, [fetchJobs, jobFilterZoneId]);

  useEffect(() => {
    const latestCompletedJob = jobs.find((job) => job.status === 'completed');
    if (!latestCompletedJob) {
      return;
    }
    if (lastCompletedJobIdRef.current === latestCompletedJob.id) {
      return;
    }
    lastCompletedJobIdRef.current = latestCompletedJob.id;
    refreshZones();
  }, [jobs, refreshZones]);

  const handleBulkToggleZone = useCallback((zoneId: string) => {
    setBulkSelectedZoneIds((prev) => {
      if (prev.includes(zoneId)) {
        return prev.filter((existingId) => existingId !== zoneId);
      }
      return [...prev, zoneId];
    });
  }, []);

  const handleSelectMatchingZones = useCallback(() => {
    setBulkSelectedZoneIds((prev) => {
      const next = new Set(prev);
      bulkMatchingZones.forEach((zone) => next.add(zone.id));
      return Array.from(next);
    });
  }, [bulkMatchingZones]);

  const handleClearBulkSelection = useCallback(() => {
    setBulkSelectedZoneIds([]);
  }, []);

  const handleBulkCreateJobs = async () => {
    if (bulkTargetZones.length === 0) {
      setError('Select at least one zone for bulk queueing.');
      return;
    }

    setCreatingBulkJobs(true);
    setError(null);
    setSuccess(null);
    try {
      const response = await apiClient.post<BulkQueueZoneTagJobsResponse>(
        '/sonar/admin/zone-tag-generation-jobs/bulk-queue',
        {
          zoneIds: bulkTargetZones.map((zone) => zone.id),
        }
      );
      setJobs((prev) => [...response.jobs, ...prev]);
      if (response.queuedCount === response.requestedZoneCount) {
        setSuccess(`Queued ${response.queuedCount} zone tag jobs.`);
      } else {
        setSuccess(
          `Queued ${response.queuedCount} zone tag jobs from ${response.requestedZoneCount} requested zones.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk queue zone tag jobs', err);
      setError('Failed to bulk queue zone tag jobs.');
    } finally {
      setCreatingBulkJobs(false);
    }
  };

  return (
    <div className="container mx-auto px-6 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Zone Tagging</h1>
          <p className="text-sm text-gray-500">
            Queue neighborhood-flavor tagging jobs, select zones visually, and review the 5 shared tags generated for each zone.
          </p>
        </div>
        <button
          className="rounded bg-gray-800 px-4 py-2 text-white hover:bg-gray-700"
          onClick={() => fetchJobs(jobFilterZoneId || undefined)}
          disabled={loadingJobs}
        >
          {loadingJobs ? 'Refreshing...' : 'Refresh jobs'}
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-emerald-700">
          {success}
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">Bulk queue</h2>
          <p className="text-sm text-gray-500">
            Select many zones and queue neighborhood tag generation across them in one pass.
          </p>
          <div className="mt-4">
            <label className="mb-1 block text-xs font-medium text-gray-500">
              Zone filter
            </label>
            <input
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
              value={bulkZoneQuery}
              onChange={(e) => setBulkZoneQuery(e.target.value)}
              placeholder="Optional name filter..."
            />
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            <button
              type="button"
              className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              onClick={handleSelectMatchingZones}
              disabled={bulkMatchingZones.length === 0}
            >
              Select matching
            </button>
            <button
              type="button"
              className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              onClick={handleClearBulkSelection}
              disabled={bulkTargetZones.length === 0}
            >
              Clear selection
            </button>
          </div>
          <BulkZoneSelectionMap
            zones={sortedZones}
            matchingZoneIds={bulkMatchingZoneIds}
            selectedZoneIds={bulkSelectedZoneIdSet}
            onToggleZone={handleBulkToggleZone}
          />
          <p className="mt-2 text-xs text-gray-500">
            Matching zones: {bulkMatchingZones.length}. Selected zones: {bulkTargetZones.length}.
          </p>
          {bulkTargetZones.length > 0 && (
            <p className="mt-2 text-xs text-gray-500">
              Selected: {bulkTargetZones.slice(0, 5).map((zone) => zone.name).join(', ')}
              {bulkTargetZones.length > 5 ? ` +${bulkTargetZones.length - 5} more` : ''}
            </p>
          )}
          <button
            className="mt-4 w-full rounded bg-violet-700 px-4 py-2 text-white hover:bg-violet-600 disabled:opacity-60"
            onClick={handleBulkCreateJobs}
            disabled={creatingBulkJobs || bulkTargetZones.length === 0}
          >
            {creatingBulkJobs ? 'Queuing bulk jobs...' : `Queue for ${bulkTargetZones.length} zones`}
          </button>
        </div>

        <div className="lg:col-span-2 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <div className="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Recent jobs</h2>
              <p className="text-sm text-gray-500">
                Review generated summaries, selected tags, and any failures.
              </p>
            </div>
            <div className="relative w-full md:w-72">
              <input
                className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                value={filterZoneQuery}
                onChange={(e) => {
                  const value = e.target.value;
                  setFilterZoneQuery(value);
                  setShowFilterZoneSuggestions(true);
                  if (value.trim() === '') {
                    setJobFilterZoneId('');
                  }
                }}
                onFocus={() => setShowFilterZoneSuggestions(true)}
                onBlur={() => setTimeout(() => setShowFilterZoneSuggestions(false), 120)}
                placeholder="Filter by zone (optional)..."
              />
              {showFilterZoneSuggestions && filterZoneSuggestions.length > 0 && (
                <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                  <button
                    type="button"
                    onClick={() => {
                      setJobFilterZoneId('');
                      setFilterZoneQuery('');
                      setShowFilterZoneSuggestions(false);
                    }}
                    className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                  >
                    All zones
                  </button>
                  {filterZoneSuggestions.map((zone) => (
                    <button
                      type="button"
                      key={zone.id}
                      onClick={() => {
                        setJobFilterZoneId(zone.id);
                        setFilterZoneQuery(zone.name);
                        setShowFilterZoneSuggestions(false);
                      }}
                      className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                    >
                      {zone.name}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {jobs.length === 0 ? (
            <div className="rounded border border-dashed border-gray-300 bg-gray-50 px-4 py-8 text-center text-sm text-gray-500">
              No zone tag jobs yet.
            </div>
          ) : (
            <div className="space-y-4">
              {jobs.map((job) => (
                <article
                  key={job.id}
                  className="rounded-lg border border-gray-200 bg-gray-50 p-4"
                >
                  <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                    <div>
                      <div className="text-lg font-semibold text-gray-900">
                        {zoneNameById.get(job.zoneId) || job.zoneId}
                      </div>
                      <div className="mt-1 text-xs text-gray-500">
                        Job {job.id.slice(0, 8)} • Created {formatDate(job.createdAt)} • Updated{' '}
                        {formatDate(job.updatedAt)}
                      </div>
                    </div>
                    <span
                      className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white ${statusBadgeClass(
                        job.status
                      )}`}
                    >
                      {job.status.replace(/_/g, ' ')}
                    </span>
                  </div>

                  {job.generatedSummary && (
                    <p className="mt-3 text-sm leading-6 text-gray-700">
                      {job.generatedSummary}
                    </p>
                  )}

                  {(job.selectedTags?.length ?? 0) > 0 && (
                    <div className="mt-3 flex flex-wrap gap-2">
                      {(job.selectedTags ?? []).map((tag) => (
                        <span
                          key={tag}
                          className="inline-flex items-center rounded-full bg-violet-100 px-3 py-1 text-xs font-medium text-violet-700"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}

                  {job.status === 'queued' || job.status === 'in_progress' ? (
                    <p className="mt-3 text-sm text-gray-600">
                      Collecting zone geometry and POI context, then selecting five shared neighborhood tags.
                    </p>
                  ) : null}

                  {job.errorMessage && (
                    <p className="mt-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                      {job.errorMessage}
                    </p>
                  )}

                  {job.contextSnapshot && (
                    <details className="mt-3">
                      <summary className="cursor-pointer text-xs font-semibold uppercase tracking-wide text-gray-600">
                        Context Used
                      </summary>
                      <pre className="mt-2 whitespace-pre-wrap rounded border border-gray-200 bg-white p-3 text-xs text-gray-700">
                        {job.contextSnapshot}
                      </pre>
                    </details>
                  )}
                </article>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ZoneTagJobs;
