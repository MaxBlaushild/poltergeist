import React, { useEffect, useRef, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Zone, ZoneImport } from '@poltergeist/types';
import { v4 as uuidv4 } from 'uuid';
import { useNavigate } from 'react-router-dom';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import * as wellknown from 'wellknown';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

const parseBoundaryString = (boundary: string): [number, number][] => {
  const trimmed = boundary.trim();
  if (!trimmed) return [];

  if (trimmed.toUpperCase().startsWith('POLYGON') || trimmed.startsWith('SRID=')) {
    const wkt = trimmed.includes(';') ? trimmed.split(';').slice(1).join(';').trim() : trimmed;
    const parsed = wellknown.parse(wkt) as GeoJSON.Polygon | GeoJSON.MultiPolygon | null;
    if (!parsed) return [];
    if (parsed.type === 'Polygon') {
      return parsed.coordinates?.[0] as [number, number][];
    }
    if (parsed.type === 'MultiPolygon') {
      return parsed.coordinates?.[0]?.[0] as [number, number][];
    }
    return [];
  }

  if (trimmed.startsWith('[')) {
    try {
      const parsed = JSON.parse(trimmed);
      if (Array.isArray(parsed)) {
        return parsed
          .filter((pair) => Array.isArray(pair) && pair.length >= 2)
          .map((pair) => [Number(pair[0]), Number(pair[1])] as [number, number])
          .filter((pair) => !Number.isNaN(pair[0]) && !Number.isNaN(pair[1]));
      }
    } catch (error) {
      console.warn('Failed to parse zone boundary JSON', error);
    }
  }

  return [];
};

const sortBoundaryPoints = (points: [number, number][]): [number, number][] => {
  if (points.length < 3) return [];

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

  const boundaryUnknown = zone.boundary as unknown;
  if (typeof boundaryUnknown === 'string') {
    return sortBoundaryPoints(parseBoundaryString(boundaryUnknown));
  }
  if (Array.isArray(boundaryUnknown)) {
    const coords = boundaryUnknown
      .filter((pair) => Array.isArray(pair) && pair.length >= 2)
      .map((pair) => [Number(pair[0]), Number(pair[1])] as [number, number])
      .filter((pair) => !Number.isNaN(pair[0]) && !Number.isNaN(pair[1]));
    return sortBoundaryPoints(coords);
  }

  return [];
};

export const Zones = () => {
  const { zones, createZone, deleteZone, refreshZones } = useZoneContext();
  const { apiClient } = useAPI();
  const [name, setName] = useState('');
  const [latitude, setLatitude] = useState(0);
  const [longitude, setLongitude] = useState(0);
  const [description, setDescription] = useState('');
  const [showCreateZone, setShowCreateZone] = useState(false);
  const [selectedMetro, setSelectedMetro] = useState('Chicago, Illinois');
  const [customMetro, setCustomMetro] = useState('');
  const [importJobs, setImportJobs] = useState<ZoneImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);
  const [importing, setImporting] = useState(false);
  const [deletingImportId, setDeletingImportId] = useState<string | null>(null);
  const [notifiedImportIds, setNotifiedImportIds] = useState<Set<string>>(new Set());
  const navigate = useNavigate();
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const popupRef = useRef<mapboxgl.Popup | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const handlersAttachedRef = useRef(false);
  const fitBoundsRef = useRef(false);

  const metroOptions = [
    'Atlanta, Georgia',
    'Austin, Texas',
    'Boston, Massachusetts',
    'Chicago, Illinois',
    'Dallas, Texas',
    'Denver, Colorado',
    'Houston, Texas',
    'Los Angeles, California',
    'Miami, Florida',
    'New York City, New York',
    'Philadelphia, Pennsylvania',
    'Phoenix, Arizona',
    'San Diego, California',
    'San Francisco, California',
    'Seattle, Washington',
    'Washington, DC'
  ];

  const effectiveMetro = selectedMetro === '__custom__' ? customMetro.trim() : selectedMetro;

  const handleImportZones = async () => {
    setImportError(null);
    if (!effectiveMetro) {
      setImportError('Please select a metro area.');
      return;
    }
    setImporting(true);
    try {
      const importItem = await apiClient.post<ZoneImport>('/sonar/zones/import', {
        metroName: effectiveMetro
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Error importing zones:', error);
      setImportError('Failed to start zone import.');
    } finally {
      setImporting(false);
    }
  };

  const fetchImportJobs = async () => {
    try {
      const query = effectiveMetro ? `?metroName=${encodeURIComponent(effectiveMetro)}` : '';
      const response = await apiClient.get<ZoneImport[]>(`/sonar/zones/imports${query}`);
      setImportJobs(response);
      const hasPending = response.some((item) => item.status === 'queued' || item.status === 'in_progress');
      setImportPolling(hasPending);
    } catch (error) {
      console.error('Failed to fetch zone import status', error);
    }
  };

  const handleDeleteImportZones = async (importId: string) => {
    const confirmed = window.confirm('Delete all zones created by this import? This cannot be undone.');
    if (!confirmed) {
      return;
    }
    setImportError(null);
    setDeletingImportId(importId);
    try {
      await apiClient.delete(`/sonar/zones/imports/${importId}`);
      await fetchImportJobs();
      await refreshZones();
    } catch (error) {
      console.error('Failed to delete imported zones', error);
      setImportError('Failed to delete imported zones.');
    } finally {
      setDeletingImportId(null);
    }
  };

  useEffect(() => {
    fetchImportJobs();
  }, [selectedMetro, customMetro]);

  useEffect(() => {
    if (!importPolling) return;
    const interval = setInterval(() => {
      fetchImportJobs();
    }, 3000);
    return () => clearInterval(interval);
  }, [importPolling, selectedMetro, customMetro]);

  useEffect(() => {
    if (importJobs.length === 0) return;
    const completed = importJobs.filter((job) => job.status === 'completed' && job.zoneCount > 0);
    if (completed.length === 0) return;

    setNotifiedImportIds((prev) => {
      const next = new Set(prev);
      let hasNew = false;
      completed.forEach((job) => {
        if (!next.has(job.id)) {
          next.add(job.id);
          hasNew = true;
        }
      });
      if (hasNew) {
        refreshZones();
      }
      return next;
    });
  }, [importJobs, refreshZones]);

  useEffect(() => {
    if (mapContainer.current && !mapRef.current) {
      mapRef.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: [-87.6298, 41.8781],
        zoom: 10,
        interactive: true
      });

      mapRef.current.addControl(new mapboxgl.NavigationControl(), 'top-right');

      mapRef.current.on('load', () => {
        setMapLoaded(true);
      });
    }

    return () => {
      popupRef.current?.remove();
      popupRef.current = null;
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
      handlersAttachedRef.current = false;
      fitBoundsRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (!mapRef.current || !mapLoaded) {
      return;
    }

    const map = mapRef.current;
    const features = zones
      .map((zone) => {
        const rawCoords = getZoneRing(zone);

        if (rawCoords.length < 4) {
          return null;
        }

        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Polygon' as const,
            coordinates: [rawCoords]
          },
          properties: {
            id: zone.id,
            name: zone.name,
            description: zone.description || '',
            boundaryCount: Math.max(rawCoords.length - 1, 0)
          }
        };
      })
      .filter(Boolean);

    const geojson = {
      type: 'FeatureCollection' as const,
      features: features as Array<GeoJSON.Feature<GeoJSON.Polygon>>
    };

    const existingSource = map.getSource('zones') as mapboxgl.GeoJSONSource | undefined;
    if (existingSource) {
      existingSource.setData(geojson);
    } else {
      map.addSource('zones', {
        type: 'geojson',
        data: geojson
      });

      map.addLayer({
        id: 'zones-fill',
        type: 'fill',
        source: 'zones',
        paint: {
          'fill-color': '#3b82f6',
          'fill-opacity': 0.25
        }
      });

      map.addLayer({
        id: 'zones-outline',
        type: 'line',
        source: 'zones',
        paint: {
          'line-color': '#1d4ed8',
          'line-width': 2
        }
      });
    }

    if (features.length > 0 && !fitBoundsRef.current) {
      const bounds = new mapboxgl.LngLatBounds();
      features.forEach((feature) => {
        feature.geometry.coordinates[0].forEach((coord) => {
          bounds.extend(coord as [number, number]);
        });
      });
      map.fitBounds(bounds, { padding: 40, maxZoom: 12 });
      fitBoundsRef.current = true;
    }

    if (!handlersAttachedRef.current && map.getLayer('zones-fill')) {
      map.on('mouseenter', 'zones-fill', () => {
        map.getCanvas().style.cursor = 'pointer';
      });
      map.on('mouseleave', 'zones-fill', () => {
        map.getCanvas().style.cursor = '';
      });

      map.on('click', 'zones-fill', (event) => {
        const feature = event.features?.[0];
        if (!feature || !feature.properties) {
          return;
        }

        const zoneId = feature.properties.id as string;
        const zoneName = feature.properties.name as string;
        const zoneDescription = feature.properties.description as string;
        const boundaryCount = Number(feature.properties.boundaryCount ?? 0);

        popupRef.current?.remove();

        const popupContent = document.createElement('div');
        popupContent.className = 'text-sm text-slate-700';

        const title = document.createElement('div');
        title.className = 'text-base font-semibold text-slate-800';
        title.textContent = zoneName;
        popupContent.appendChild(title);

        const description = document.createElement('div');
        description.className = 'mt-1 text-xs text-slate-600';
        description.textContent = zoneDescription || 'No description.';
        popupContent.appendChild(description);

        const meta = document.createElement('div');
        meta.className = 'mt-2 text-xs text-slate-500';
        meta.textContent = `Boundary points: ${boundaryCount}`;
        popupContent.appendChild(meta);

        const button = document.createElement('button');
        button.className = 'mt-3 w-full rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-indigo-700';
        button.textContent = 'Open Zone';
        button.addEventListener('click', () => {
          popupRef.current?.remove();
          navigate(`/zones/${zoneId}`);
        });
        popupContent.appendChild(button);

        popupRef.current = new mapboxgl.Popup({ closeOnClick: true })
          .setLngLat(event.lngLat)
          .setDOMContent(popupContent)
          .addTo(map);
      });

      handlersAttachedRef.current = true;
    }
  }, [zones, navigate, mapLoaded]);
  
  return <div className="m-10">
    <h1 className='text-2xl font-bold'>Zones</h1>
    <div className="mt-6 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <h2 className="text-lg font-semibold text-slate-800">Import Neighborhood Zones</h2>
      <p className="text-sm text-slate-500">Select a metro area to import neighborhood polygons from OSM.</p>
      <div className="mt-4 flex flex-col gap-3 md:flex-row md:items-end">
        <div className="flex-1">
          <label className="mb-1 block text-sm font-medium text-slate-700">Metro Area</label>
          <select
            value={selectedMetro}
            onChange={(e) => setSelectedMetro(e.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            {metroOptions.map((option) => (
              <option key={option} value={option}>{option}</option>
            ))}
            <option value="__custom__">Custom...</option>
          </select>
        </div>
        {selectedMetro === '__custom__' && (
          <div className="flex-1">
            <label className="mb-1 block text-sm font-medium text-slate-700">Custom Metro</label>
            <input
              type="text"
              value={customMetro}
              onChange={(e) => setCustomMetro(e.target.value)}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              placeholder="e.g., Minneapolis, Minnesota"
            />
          </div>
        )}
        <button
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
          onClick={handleImportZones}
          disabled={importing || !effectiveMetro}
        >
          {importing ? 'Queueing...' : 'Import Zones'}
        </button>
      </div>
      {importError && (
        <p className="mt-2 text-sm text-red-600">{importError}</p>
      )}
      <div className="mt-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-slate-700">Recent Imports</h3>
          <button
            className="text-xs font-semibold text-slate-500 hover:text-slate-700"
            onClick={fetchImportJobs}
          >
            Refresh
          </button>
        </div>
        {importJobs.length === 0 ? (
          <p className="mt-2 text-sm text-slate-400">No imports yet.</p>
        ) : (
          <div className="mt-2 space-y-2">
            {importJobs.slice(0, 6).map((job) => (
              <div key={job.id} className="flex items-center justify-between rounded-md border border-slate-200 px-3 py-2 text-sm">
                <div>
                  <div className="font-medium text-slate-800">{job.metroName}</div>
                  <div className="text-xs text-slate-500">Status: {job.status}</div>
                  {job.errorMessage && <div className="text-xs text-red-600">{job.errorMessage}</div>}
                </div>
                <div className="flex items-center gap-3">
                  <div className="text-xs text-slate-500">Zones: {job.zoneCount}</div>
                  <button
                    className="rounded-md border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-600 hover:text-slate-800 disabled:opacity-60"
                    onClick={() => handleDeleteImportZones(job.id)}
                    disabled={deletingImportId === job.id}
                  >
                    {deletingImportId === job.id ? 'Deleting...' : 'Delete Zones'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
    <div className="mt-6 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-slate-800">Zone Boundaries</h2>
          <p className="text-sm text-slate-500">Click a zone polygon to view details and open its page.</p>
        </div>
        <button
          className="text-xs font-semibold text-slate-500 hover:text-slate-700"
          onClick={() => {
            if (!mapRef.current || zones.length === 0) return;
            const bounds = new mapboxgl.LngLatBounds();
            let hasBounds = false;
            zones.forEach((zone) => {
              const coords = getZoneRing(zone);
              coords.forEach((coord) => {
                bounds.extend(coord);
                hasBounds = true;
              });
            });
            if (hasBounds) {
              mapRef.current.fitBounds(bounds, { padding: 40, maxZoom: 12 });
            }
          }}
        >
          Fit to zones
        </button>
      </div>
      <div className="mt-4 h-[420px] w-full overflow-hidden rounded-lg border border-slate-200">
        <div ref={mapContainer} className="h-full w-full" />
      </div>
    </div>
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
      gap: '20px',
      padding: '20px'
    }}>
      {zones && zones.map((zone) => (
        <div 
          key={zone.id}
          style={{
            padding: '20px',
            border: '1px solid #ccc',
            borderRadius: '8px',
            backgroundColor: '#fff',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
          }}
        >
          <h2 style={{ 
            margin: '0 0 15px 0',
            color: '#333'
          }}>{zone.name}</h2>
          <p style={{
            margin: '5px 0',
            color: '#666'
          }}>Latitude: {zone.latitude}</p>
          <p style={{
            margin: '5px 0',
            color: '#666'
          }}>Longitude: {zone.longitude}</p>
          <p style={{
            margin: '5px 0',
            color: '#666'
          }}>Boundary points: {zone.points?.length ?? 0}</p>
          <button
            onClick={() => deleteZone(zone)}
            className="bg-red-500 text-white px-4 py-2 rounded-md mr-2"
          >
            Delete
          </button>
          <button
            onClick={() => navigate(`/zones/${zone.id}`)}
            className="bg-blue-500 text-white px-4 py-2 rounded-md"
          >
            View
          </button>
        </div>
        
      ))}
    </div>
    <button
      className="bg-blue-500 text-white px-4 py-2 rounded-md"
      onClick={() =>
        setShowCreateZone(!showCreateZone)
      }
    >
      Create Zone
    </button>
    {showCreateZone && (
      <div style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        backgroundColor: 'rgba(0,0,0,0.5)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center'
      }}>
        <div style={{
          backgroundColor: '#fff',
          padding: '20px',
          borderRadius: '8px',
          width: '400px'
        }}>
          <h2>Create Zone</h2>
          <div style={{ marginBottom: '15px' }}>
            <label style={{ display: 'block', marginBottom: '5px' }}>Name:</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              style={{
                width: '100%',
                padding: '8px',
                borderRadius: '4px',
                border: '1px solid #ccc'
              }}
            />
          </div>
          <div style={{ marginBottom: '15px' }}>
            <label style={{ display: 'block', marginBottom: '5px' }}>Description:</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              style={{
                width: '100%',
                padding: '8px',
                borderRadius: '4px',
                border: '1px solid #ccc',
                minHeight: '100px',
                resize: 'vertical'
              }}
            />
          </div>
          <div style={{ marginBottom: '15px' }}>
            <label style={{ display: 'block', marginBottom: '5px' }}>Latitude:</label>
            <input
              type="number"
              value={latitude}
              onChange={(e) => setLatitude(Number(e.target.value))}
              style={{
                width: '100%',
                padding: '8px',
                borderRadius: '4px',
                border: '1px solid #ccc'
              }}
            />
          </div>

          <div style={{ marginBottom: '15px' }}>
            <label style={{ display: 'block', marginBottom: '5px' }}>Longitude:</label>
            <input
              type="number"
              value={longitude}
              onChange={(e) => setLongitude(Number(e.target.value))}
              style={{
                width: '100%',
                padding: '8px',
                borderRadius: '4px',
                border: '1px solid #ccc'
              }}
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
            <button
              onClick={() => setShowCreateZone(false)}
              style={{
                padding: '8px 16px',
                borderRadius: '4px',
                border: '1px solid #ccc',
                backgroundColor: '#fff'
              }}
            >
              Cancel
            </button>
            <button
              onClick={() => {
                createZone({
                  id: uuidv4(),
                  name,
                  latitude,
                  longitude,
                  description,
                  createdAt: new Date(),
                  updatedAt: new Date()
                });
                setShowCreateZone(false);
              }}
              style={{
                padding: '8px 16px',
                borderRadius: '4px',
                border: 'none',
                backgroundColor: '#007bff',
                color: '#fff'
              }}
            >
              Create
            </button>
          </div>
        </div>
      </div>
    )}
  </div>;
};
