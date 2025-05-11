import React, { useState } from 'react';
import { useZoneContext } from '@poltergeist/contexts';
import { v4 as uuidv4 } from 'uuid';
import { useNavigate } from 'react-router-dom';
export const Zones = () => {
  const { zones, selectedZone, setSelectedZone, createZone, deleteZone } = useZoneContext();
  const [name, setName] = useState('');
  const [latitude, setLatitude] = useState(0);
  const [longitude, setLongitude] = useState(0);
  const [radius, setRadius] = useState(0);
  const [description, setDescription] = useState('');
  const [showCreateZone, setShowCreateZone] = useState(false);
  const navigate = useNavigate();
  
  return <div className="m-10">
    <h1 className='text-2xl font-bold'>Zones</h1>
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
