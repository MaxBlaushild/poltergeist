import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { District, Zone } from '@poltergeist/types';
import { Link, useNavigate, useParams } from 'react-router-dom';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { DistrictSeedJobsPanel } from './DistrictSeedJobsPanel.tsx';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

const sortBoundaryPoints = (points: [number, number][]): [number, number][] => {
  if (points.length < 3) {
    return points.slice();
  }

  const centroid = points.reduce(
    (acc, point) => [
      acc[0] + point[0] / points.length,
      acc[1] + point[1] / points.length,
    ],
    [0, 0]
  );

  const sorted = points.slice().sort((a, b) => {
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
      zone.points.map(
        (point) => [point.longitude, point.latitude] as [number, number]
      )
    );
  }

  if (zone.boundaryCoords?.length) {
    return sortBoundaryPoints(
      zone.boundaryCoords.map(
        (coord) => [coord.longitude, coord.latitude] as [number, number]
      )
    );
  }

  return [];
};

type DistrictZoneSelectorMapProps = {
  zones: Zone[];
  selectedZoneIds: Set<string>;
  matchingZoneIds: Set<string>;
  onToggleZone: (zoneId: string) => void;
};

const DistrictZoneSelectorMap: React.FC<DistrictZoneSelectorMapProps> = ({
  zones,
  selectedZoneIds,
  matchingZoneIds,
  onToggleZone,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const fitBoundsRef = useRef(false);
  const searchMarkerRef = useRef<mapboxgl.Marker | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [searchStatus, setSearchStatus] = useState<string | null>(null);
  const [showSearchSuggestions, setShowSearchSuggestions] = useState(false);
  const [searchCandidates, setSearchCandidates] = useState<
    Array<{
      id: string;
      center: [number, number];
      placeName: string;
    }>
  >([]);

  useEffect(() => {
    if (mapContainer.current && !mapRef.current) {
      mapRef.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: [-87.6298, 41.8781],
        zoom: 10,
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

  const focusMapOnLocation = (lngLat: [number, number], zoom: number) => {
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
  };

  const handleSearchLocation = async () => {
    const query = searchQuery.trim();
    if (!query) {
      setSearchStatus('Enter a city, neighborhood, address, or landmark.');
      return;
    }

    if (!mapboxgl.accessToken) {
      setSearchStatus(
        'Location search is unavailable because the Mapbox token is missing.'
      );
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
  };

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

    const existingSource = mapRef.current.getSource(
      'district-zone-selector'
    ) as mapboxgl.GeoJSONSource | undefined;

    if (existingSource) {
      existingSource.setData(geojson);
    } else {
      mapRef.current.addSource('district-zone-selector', {
        type: 'geojson',
        data: geojson,
      });

      mapRef.current.addLayer({
        id: 'district-zone-selector-fill',
        type: 'fill',
        source: 'district-zone-selector',
        paint: {
          'fill-color': [
            'case',
            ['boolean', ['get', 'selected'], false],
            '#0f766e',
            ['boolean', ['get', 'matching'], false],
            '#2563eb',
            '#94a3b8',
          ],
          'fill-opacity': [
            'case',
            ['boolean', ['get', 'selected'], false],
            0.4,
            ['boolean', ['get', 'matching'], false],
            0.16,
            0.07,
          ],
        },
      });

      mapRef.current.addLayer({
        id: 'district-zone-selector-outline',
        type: 'line',
        source: 'district-zone-selector',
        paint: {
          'line-color': [
            'case',
            ['boolean', ['get', 'selected'], false],
            '#115e59',
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

      mapRef.current.on('click', 'district-zone-selector-fill', (event) => {
        const zoneId = event.features?.[0]?.properties?.id;
        if (typeof zoneId === 'string' && zoneId) {
          onToggleZone(zoneId);
        }
      });

      mapRef.current.on('mouseenter', 'district-zone-selector-fill', () => {
        mapRef.current?.getCanvas().style.setProperty('cursor', 'pointer');
      });

      mapRef.current.on('mouseleave', 'district-zone-selector-fill', () => {
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
    <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
      <div className="mb-2">
        <div className="text-sm font-semibold text-gray-900">
          Map-powered zone picker
        </div>
        <div className="text-xs text-gray-500">
          Click zone polygons to add or remove them from the district.
        </div>
      </div>
      <div className="mb-3 flex flex-wrap gap-2">
        <div className="relative min-w-[240px] flex-1">
          <input
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
            placeholder="Search for a city, neighborhood, address, or landmark"
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
          />
          {showSearchSuggestions && searchCandidates.length > 0 && (
            <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded-lg border border-gray-200 bg-white shadow">
              {searchCandidates.map((candidate) => (
                <button
                  key={candidate.id}
                  type="button"
                  className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                  onClick={() => {
                    setSearchQuery(candidate.placeName);
                    setShowSearchSuggestions(false);
                    focusMapOnLocation(candidate.center, 13);
                    setSearchStatus(`Moved map to ${candidate.placeName}.`);
                  }}
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
          disabled={isSearching}
          className="rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          {isSearching ? 'Searching...' : 'Search'}
        </button>
      </div>
      <div
        ref={mapContainer}
        className="h-[380px] w-full overflow-hidden rounded-lg border border-gray-200"
      />
      {searchStatus && (
        <p className="mt-2 text-xs text-gray-500">{searchStatus}</p>
      )}
    </div>
  );
};

export const DistrictEditor = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { apiClient } = useAPI();
  const { zones, refreshZones } = useZoneContext();
  const [district, setDistrict] = useState<District | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedZoneIds, setSelectedZoneIds] = useState<Set<string>>(
    new Set()
  );
  const [zoneSearch, setZoneSearch] = useState('');
  const districtId = id || '';

  useEffect(() => {
    void refreshZones();
  }, [refreshZones]);

  useEffect(() => {
    const loadDistrict = async () => {
      if (!districtId) {
        setLoading(false);
        setError('District ID is missing.');
        return;
      }

      setLoading(true);
      setError(null);
      try {
        const response = await apiClient.get<District>(
          `/sonar/districts/${districtId}`
        );
        setDistrict(response);
        setName(response.name);
        setDescription(response.description || '');
        setSelectedZoneIds(
          new Set((response.zones || []).map((zone) => zone.id))
        );
      } catch (err) {
        console.error('Error fetching district:', err);
        setError('Unable to load that district right now.');
      } finally {
        setLoading(false);
      }
    };

    void loadDistrict();
  }, [apiClient, districtId]);

  const filteredZones = useMemo(() => {
    const query = zoneSearch.trim().toLowerCase();
    if (!query) {
      return zones;
    }

    return zones.filter((zone) => {
      const haystack = `${zone.name} ${zone.description || ''}`.toLowerCase();
      return haystack.includes(query);
    });
  }, [zoneSearch, zones]);

  const matchingZoneIds = useMemo(
    () => new Set(filteredZones.map((zone) => zone.id)),
    [filteredZones]
  );

  const selectedZoneCount = selectedZoneIds.size;

  const toggleZone = (zoneId: string) => {
    setSelectedZoneIds((current) => {
      const next = new Set(current);
      if (next.has(zoneId)) {
        next.delete(zoneId);
      } else {
        next.add(zoneId);
      }
      return next;
    });
  };

  const handleSaveDistrict = async () => {
    const trimmedName = name.trim();
    if (!trimmedName || !districtId) {
      return;
    }

    setSaving(true);
    setError(null);
    try {
      const updated = await apiClient.patch<District>(
        `/sonar/districts/${districtId}`,
        {
          name: trimmedName,
          description: description.trim(),
          zoneIds: Array.from(selectedZoneIds),
        }
      );
      setDistrict(updated);
      setName(updated.name);
      setDescription(updated.description || '');
      setSelectedZoneIds(new Set((updated.zones || []).map((zone) => zone.id)));
    } catch (err) {
      console.error('Error saving district:', err);
      setError('Unable to save that district right now.');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteDistrict = async () => {
    if (!districtId || !district) {
      return;
    }

    const confirmed = window.confirm(`Delete district "${district.name}"?`);
    if (!confirmed) {
      return;
    }

    try {
      await apiClient.delete(`/sonar/districts/${districtId}`);
      navigate('/districts');
    } catch (err) {
      console.error('Error deleting district:', err);
      setError('Unable to delete that district right now.');
    }
  };

  if (loading) {
    return <div className="p-6 text-sm text-gray-500">Loading district...</div>;
  }

  return (
    <div className="mx-auto flex max-w-7xl flex-col gap-6 p-6">
      <div className="flex flex-col gap-2">
        <Link
          to="/districts"
          className="text-sm font-medium text-blue-700 hover:text-blue-900"
        >
          Back to districts
        </Link>
        <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">
              {district?.name || 'District editor'}
            </h1>
            <p className="text-sm text-gray-600">
              I named this entity a district: a larger region made from multiple
              zones.
            </p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => void handleSaveDistrict()}
              disabled={saving || name.trim() === ''}
              className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
            >
              {saving ? 'Saving...' : 'Save district'}
            </button>
            <button
              type="button"
              onClick={() => void handleDeleteDistrict()}
              className="rounded-lg border border-red-200 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="grid gap-6 xl:grid-cols-[380px_minmax(0,1fr)]">
        <div className="flex flex-col gap-6">
          <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900">
              District details
            </h2>
            <div className="mt-4 flex flex-col gap-3">
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-gray-700">Name</span>
                <input
                  className="rounded-lg border border-gray-300 px-3 py-2"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-gray-700">Description</span>
                <textarea
                  className="min-h-[140px] rounded-lg border border-gray-300 px-3 py-2"
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                />
              </label>
              <div className="rounded-lg bg-gray-50 px-3 py-3 text-sm text-gray-600">
                <div className="font-medium text-gray-900">
                  {selectedZoneCount} zones selected
                </div>
                <div className="mt-1">
                  Use the map or the checklist to curate the district.
                </div>
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
            <div className="flex items-end justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">
                  Zone checklist
                </h2>
                <p className="text-sm text-gray-500">
                  Search zones by name, then click or checkbox them into the
                  district.
                </p>
              </div>
            </div>
            <input
              className="mt-4 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
              placeholder="Search zones"
              value={zoneSearch}
              onChange={(event) => setZoneSearch(event.target.value)}
            />
            <div className="mt-4 max-h-[480px] space-y-2 overflow-y-auto pr-1">
              {filteredZones.map((zone) => {
                const checked = selectedZoneIds.has(zone.id);
                return (
                  <label
                    key={zone.id}
                    className={`flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-3 text-sm transition ${
                      checked
                        ? 'border-emerald-300 bg-emerald-50'
                        : 'border-gray-200 bg-gray-50 hover:border-gray-300 hover:bg-white'
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleZone(zone.id)}
                      className="mt-1 h-4 w-4"
                    />
                    <div className="min-w-0">
                      <div className="font-medium text-gray-900">
                        {zone.name}
                      </div>
                      <div className="mt-1 text-xs text-gray-500">
                        {zone.description || 'No description'}
                      </div>
                    </div>
                  </label>
                );
              })}
              {filteredZones.length === 0 && (
                <div className="rounded-lg border border-dashed border-gray-300 px-4 py-8 text-center text-sm text-gray-500">
                  No zones match that search.
                </div>
              )}
            </div>
          </div>
        </div>

        <DistrictZoneSelectorMap
          zones={zones}
          selectedZoneIds={selectedZoneIds}
          matchingZoneIds={matchingZoneIds}
          onToggleZone={toggleZone}
        />
      </div>

      <DistrictSeedJobsPanel district={district} />
    </div>
  );
};
