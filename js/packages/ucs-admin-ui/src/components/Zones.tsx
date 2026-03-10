import React, { useEffect, useRef, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Zone, ZoneImport } from '@poltergeist/types';
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

const calculateBoundaryCenter = (points: [number, number][]) => {
  const openPoints =
    points.length > 1 &&
    points[0][0] === points[points.length - 1][0] &&
    points[0][1] === points[points.length - 1][1]
      ? points.slice(0, -1)
      : points;

  if (openPoints.length === 0) {
    return null;
  }

  const totals = openPoints.reduce(
    (acc, point) => [acc[0] + point[0], acc[1] + point[1]],
    [0, 0]
  );

  return {
    longitude: totals[0] / openPoints.length,
    latitude: totals[1] / openPoints.length,
  };
};

type BoundaryEditorMapProps = {
  center: [number, number];
  boundaryPoints: [number, number][];
  allZoneBoundaries: [number, number][][];
  onMapClick: (lngLat: mapboxgl.LngLat) => void;
  onClearBoundary: () => void;
};

const ZoneBoundaryEditorMap: React.FC<BoundaryEditorMapProps> = ({
  center,
  boundaryPoints,
  allZoneBoundaries,
  onMapClick,
  onClearBoundary,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const markers = useRef<mapboxgl.Marker[]>([]);
  const [isLocating, setIsLocating] = useState(false);
  const [locationError, setLocationError] = useState<string | null>(null);

  const cleanupBoundary = () => {
    if (!map.current) {
      return;
    }
    if (map.current.getLayer('zone-create-boundary-outline')) {
      map.current.removeLayer('zone-create-boundary-outline');
    }
    if (map.current.getLayer('zone-create-boundary')) {
      map.current.removeLayer('zone-create-boundary');
    }
    if (map.current.getSource('zone-create-boundary')) {
      map.current.removeSource('zone-create-boundary');
    }
  };

  useEffect(() => {
    if (mapContainer.current && !map.current) {
      map.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center,
        zoom: 12,
        interactive: true,
      });

      map.current.on('load', () => {
        setMapLoaded(true);
      });
    }

    return () => {
      cleanupBoundary();
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    map.current.flyTo({
      center,
      essential: true,
    });
  }, [center, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    const canvasContainer = mapContainer.current?.querySelector('.mapboxgl-canvas-container');
    if (canvasContainer instanceof HTMLElement) {
      canvasContainer.style.cursor = 'grab';
    }

    const handleClick = (event: mapboxgl.MapMouseEvent) => {
      onMapClick(event.lngLat);
    };

    map.current.on('click', handleClick);
    return () => {
      map.current?.off('click', handleClick);
    };
  }, [mapLoaded, onMapClick]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    markers.current.forEach((marker) => marker.remove());
    markers.current = [];

    const sortedPoints = sortBoundaryPoints(boundaryPoints);
    sortedPoints.forEach((point) => {
      const marker = new mapboxgl.Marker({
        color: '#DC2626',
        draggable: false,
      })
        .setLngLat(point)
        .addTo(map.current!);
      markers.current.push(marker);
    });
  }, [boundaryPoints, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    cleanupBoundary();

    const sortedPoints = sortBoundaryPoints(boundaryPoints);
    if (sortedPoints.length < 4) {
      return;
    }

    map.current.addSource('zone-create-boundary', {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Polygon',
          coordinates: [sortedPoints],
        },
      },
    });

    map.current.addLayer({
      id: 'zone-create-boundary',
      type: 'fill',
      source: 'zone-create-boundary',
      paint: {
        'fill-color': '#DC2626',
        'fill-opacity': 0.3,
      },
    });

    map.current.addLayer({
      id: 'zone-create-boundary-outline',
      type: 'line',
      source: 'zone-create-boundary',
      paint: {
        'line-color': '#DC2626',
        'line-width': 2,
      },
    });
  }, [boundaryPoints, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    if (map.current.getLayer('zone-create-all-boundaries-outline')) {
      map.current.removeLayer('zone-create-all-boundaries-outline');
    }
    if (map.current.getLayer('zone-create-all-boundaries')) {
      map.current.removeLayer('zone-create-all-boundaries');
    }
    if (map.current.getSource('zone-create-all-boundaries')) {
      map.current.removeSource('zone-create-all-boundaries');
    }

    const polygonFeatures = allZoneBoundaries
      .map((points) => sortBoundaryPoints(points))
      .filter((points) => points.length >= 4)
      .map((points) => ({
        type: 'Feature' as const,
        properties: {},
        geometry: {
          type: 'Polygon' as const,
          coordinates: [points],
        },
      }));

    if (polygonFeatures.length === 0) {
      return;
    }

    map.current.addSource('zone-create-all-boundaries', {
      type: 'geojson',
      data: {
        type: 'FeatureCollection',
        features: polygonFeatures,
      },
    });

    map.current.addLayer(
      {
        id: 'zone-create-all-boundaries',
        type: 'fill',
        source: 'zone-create-all-boundaries',
        paint: {
          'fill-color': '#2563EB',
          'fill-opacity': 0.18,
        },
      },
      map.current.getLayer('zone-create-boundary') ? 'zone-create-boundary' : undefined
    );

    map.current.addLayer(
      {
        id: 'zone-create-all-boundaries-outline',
        type: 'line',
        source: 'zone-create-all-boundaries',
        paint: {
          'line-color': '#2563EB',
          'line-width': 2,
        },
      },
      map.current.getLayer('zone-create-boundary-outline') ? 'zone-create-boundary-outline' : undefined
    );

    return () => {
      if (!map.current) {
        return;
      }
      if (map.current.getLayer('zone-create-all-boundaries-outline')) {
        map.current.removeLayer('zone-create-all-boundaries-outline');
      }
      if (map.current.getLayer('zone-create-all-boundaries')) {
        map.current.removeLayer('zone-create-all-boundaries');
      }
      if (map.current.getSource('zone-create-all-boundaries')) {
        map.current.removeSource('zone-create-all-boundaries');
      }
    };
  }, [allZoneBoundaries, mapLoaded, boundaryPoints]);

  return (
    <div style={{ marginBottom: '15px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
        <div>
          <div style={{ fontWeight: 600, marginBottom: '4px' }}>Boundary</div>
          <div style={{ fontSize: '12px', color: '#6b7280' }}>
            Click on the map to add boundary points for the new zone.
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            type="button"
            onClick={onClearBoundary}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
          >
            Clear Boundary
          </button>
          <button
            type="button"
            onClick={() => map.current?.zoomIn()}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
          >
            +
          </button>
          <button
            type="button"
            onClick={() => map.current?.zoomOut()}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
          >
            −
          </button>
          <button
            type="button"
            onClick={() => {
              if (!navigator.geolocation) {
                setLocationError('Geolocation is not supported in this browser.');
                return;
              }
              setIsLocating(true);
              setLocationError(null);
              navigator.geolocation.getCurrentPosition(
                (pos) => {
                  setIsLocating(false);
                  const { latitude, longitude } = pos.coords;
                  map.current?.flyTo({
                    center: [longitude, latitude],
                    zoom: Math.max(map.current?.getZoom() ?? 12, 16),
                    essential: true,
                  });
                },
                (err) => {
                  setIsLocating(false);
                  setLocationError(err.message || 'Unable to fetch location.');
                },
                { enableHighAccuracy: true, timeout: 10000, maximumAge: 10000 }
              );
            }}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
            disabled={isLocating}
          >
            {isLocating ? '...' : 'My Location'}
          </button>
        </div>
      </div>
      <div
        ref={mapContainer}
        style={{
          width: '100%',
          height: '360px',
          borderRadius: '8px',
          overflow: 'hidden',
          border: '1px solid #d1d5db',
        }}
      />
      <div style={{ marginTop: '10px', fontSize: '12px', color: '#6b7280' }}>
        {boundaryPoints.length} points added
        {boundaryPoints.length < 3 ? '. Add at least 3 points to create the zone.' : '.'}
      </div>
      {locationError && (
        <div style={{ marginTop: '8px', color: '#dc2626', fontSize: '12px' }}>{locationError}</div>
      )}
    </div>
  );
};

export const Zones = () => {
  const { zones, deleteZone, refreshZones } = useZoneContext();
  const { apiClient } = useAPI();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [showCreateZone, setShowCreateZone] = useState(false);
  const [createBoundaryPoints, setCreateBoundaryPoints] = useState<[number, number][]>([]);
  const [createMapCenter, setCreateMapCenter] = useState<[number, number]>([-87.6298, 41.8781]);
  const [createZoneError, setCreateZoneError] = useState<string | null>(null);
  const [isCreatingZone, setIsCreatingZone] = useState(false);
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

  const openCreateZoneModal = () => {
    const currentCenter = mapRef.current?.getCenter();
    setCreateMapCenter(
      currentCenter
        ? [currentCenter.lng, currentCenter.lat]
        : zones[0]
          ? [zones[0].longitude, zones[0].latitude]
          : [-87.6298, 41.8781]
    );
    setName('');
    setDescription('');
    setCreateBoundaryPoints([]);
    setCreateZoneError(null);
    setShowCreateZone(true);
  };

  const closeCreateZoneModal = () => {
    setShowCreateZone(false);
    setCreateZoneError(null);
    setIsCreatingZone(false);
  };

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

  const handleCreateZone = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) {
      setCreateZoneError('Zone name is required.');
      return;
    }

    if (createBoundaryPoints.length < 3) {
      setCreateZoneError('Please add at least 3 boundary points.');
      return;
    }

    const sortedBoundaryPoints = sortBoundaryPoints(createBoundaryPoints);
    const center = calculateBoundaryCenter(sortedBoundaryPoints);
    if (!center) {
      setCreateZoneError('Unable to calculate the zone center from the boundary.');
      return;
    }

    setCreateZoneError(null);
    setIsCreatingZone(true);

    try {
      const createdZone = await apiClient.post<Zone>('/sonar/zones', {
        name: trimmedName,
        description,
        latitude: center.latitude,
        longitude: center.longitude,
      });

      await apiClient.post(`/sonar/zones/${createdZone.id}/boundary`, {
        boundary: sortedBoundaryPoints.slice(0, -1),
      });

      await refreshZones();
      closeCreateZoneModal();
      navigate(`/zones/${createdZone.id}`);
    } catch (error) {
      console.error('Error creating zone:', error);
      setCreateZoneError('Failed to create zone.');
    } finally {
      setIsCreatingZone(false);
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

  const allZoneBoundaries = zones
    .map((zone) => getZoneRing(zone))
    .filter((points) => points.length >= 4);
  
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
      onClick={openCreateZoneModal}
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
          width: 'min(900px, calc(100vw - 40px))',
          maxHeight: 'calc(100vh - 40px)',
          overflowY: 'auto'
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
          <ZoneBoundaryEditorMap
            center={createMapCenter}
            boundaryPoints={createBoundaryPoints}
            allZoneBoundaries={allZoneBoundaries}
            onMapClick={(lngLat) => {
              setCreateBoundaryPoints((prev) => [...prev, [lngLat.lng, lngLat.lat]]);
            }}
            onClearBoundary={() => setCreateBoundaryPoints([])}
          />
          {createZoneError && (
            <div style={{ marginBottom: '15px', color: '#dc2626', fontSize: '14px' }}>
              {createZoneError}
            </div>
          )}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
            <button
              onClick={closeCreateZoneModal}
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
              onClick={handleCreateZone}
              style={{
                padding: '8px 16px',
                borderRadius: '4px',
                border: 'none',
                backgroundColor: '#007bff',
                color: '#fff'
              }}
              disabled={isCreatingZone}
            >
              {isCreatingZone ? 'Creating...' : 'Create'}
            </button>
          </div>
        </div>
      </div>
    )}
  </div>;
};
