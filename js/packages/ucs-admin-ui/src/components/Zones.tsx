import React, { useEffect, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { ZoneImport } from '@poltergeist/types';
import { v4 as uuidv4 } from 'uuid';
import { useNavigate } from 'react-router-dom';
export const Zones = () => {
  const { zones, selectedZone, setSelectedZone, createZone, deleteZone, refreshZones } = useZoneContext();
  const { apiClient } = useAPI();
  const [name, setName] = useState('');
  const [latitude, setLatitude] = useState(0);
  const [longitude, setLongitude] = useState(0);
  const [radius, setRadius] = useState(0);
  const [description, setDescription] = useState('');
  const [showCreateZone, setShowCreateZone] = useState(false);
  const [selectedMetro, setSelectedMetro] = useState('Chicago, Illinois');
  const [customMetro, setCustomMetro] = useState('');
  const [importJobs, setImportJobs] = useState<ZoneImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);
  const [importing, setImporting] = useState(false);
  const [notifiedImportIds, setNotifiedImportIds] = useState<Set<string>>(new Set());
  const navigate = useNavigate();

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
                <div className="text-xs text-slate-500">Zones: {job.zoneCount}</div>
              </div>
            ))}
          </div>
        )}
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
          }}>Radius: {zone.radius}m</p>
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
          <div style={{ marginBottom: '15px' }}>
            <label style={{ display: 'block', marginBottom: '5px' }}>Radius (meters):</label>
            <input
              type="number"
              value={radius}
              onChange={(e) => setRadius(Number(e.target.value))}
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
                  radius,
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
