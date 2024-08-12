import React, { useEffect, useRef, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import './MatchInProgress.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { PointOfInterest, Team, hasTeamDiscoveredPointOfInterest } from '@poltergeist/types';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { Button } from './shared/Button.tsx';

mapboxgl.accessToken =
  'REDACTED';

const Marker = ({
  pointOfInterest,
  index,
  onClick,
  usersTeam,
}: {
  pointOfInterest: PointOfInterest;
  index: number;
  onClick: (e: React.MouseEvent) => void;
  usersTeam: Team;
}) => {
  const hasDiscovered = hasTeamDiscoveredPointOfInterest(usersTeam, pointOfInterest);
  const imageUrl = hasDiscovered ? pointOfInterest.imageURL : `https://crew-points-of-interest.s3.amazonaws.com/unclaimed-pirate-fortress-${(index + 1) % 6}.png`;
  return (
    <button onClick={onClick} className="marker">
      <img
        src={imageUrl}
        alt={hasDiscovered ? pointOfInterest.name : 'Mystery fortress'}
        className="w-24 h-24 rounded-lg border-2 border-black"
      />
    </button>
  );
};

export const MatchInProgress = () => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const { currentUser } = useUserProfiles();
  const { match } = useMatchContext();
  const [lng, setLng] = useState(-70.9);
  const [lat, setLat] = useState(42.35);
  const [zoom, setZoom] = useState(16);
  const [selectedPointOfInterest, setSelectedPointOfInterest] =
    useState<PointOfInterest | null>(null);
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const usersTeam = match?.teams.find((team) =>
    team.users.some((user) => user.id === currentUser?.id)
  );
  const otherTeams = match?.teams.filter((team) => team.id !== usersTeam?.id);

  useEffect(() => {
    if (selectedPointOfInterest) {
      setIsPanelVisible(true);
    } else {
      setIsPanelVisible(false);
    }
  }, [selectedPointOfInterest]);

  useEffect(() => {
    map.current = new mapboxgl.Map({
      container: mapContainer.current!,
      style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
      center: [lng, lat],
      zoom: zoom,
    });

    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        setLng(position.coords.longitude);
        setLat(position.coords.latitude);
        map.current?.setCenter([
          position.coords.longitude,
          position.coords.latitude,
        ]);
      });
    } else {
      console.error('Geolocation is not supported by this browser.');
    }

    map.current.on('move', () => {
      setLng(map.current?.getCenter().lng ?? 0);
      setLat(map.current?.getCenter().lat ?? 0);
      setZoom(map.current?.getZoom() ?? 0);
    });

    return () => map.current?.remove();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleMapClick = () => {
    setIsPanelVisible(false);
    setTimeout(() => {
      setSelectedPointOfInterest(null);
    }, 300);
  };

  const handleDrawerClick = (event: React.MouseEvent) => {
    event.stopPropagation();
  };

  useEffect(() => {
    if (match && map.current && usersTeam) {
      console.log(usersTeam);
      match.pointsOfInterest.forEach((pointOfInterest, i) => {
        console.log(pointOfInterest);
        const markerDiv = document.createElement('div');

        createRoot(markerDiv).render(
          <Marker
            pointOfInterest={pointOfInterest}
            index={i}
            usersTeam={usersTeam}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedPointOfInterest(pointOfInterest);
            }}
          />
        );

        new mapboxgl.Marker(markerDiv)
          .setLngLat([
            parseFloat(pointOfInterest.lng) * -1,
            parseFloat(pointOfInterest.lat),
          ])
          .addTo(map.current!);
      });
    }
  }, [match, map]);

  return (
    <div className="">
      <div
        ref={mapContainer}
        onClick={handleMapClick}
        style={{
          top: -70,
          left: 0,
          width: '100vw',
          height: '100vh',
          zIndex: 1,
        }}
      />
      <div
        onClick={handleDrawerClick}
        className="Match__bottomDrawer"
        style={{
          position: 'fixed',
          bottom: 0,
          left: 0,
          width: '100%',
          height: '80vh',
          transition: 'transform 0.3s ease-in-out',
          transform: isPanelVisible ? 'translateY(0)' : 'translateY(100%)',
          zIndex: 2,
          overflowY: 'scroll',
        }}
      >
        {selectedPointOfInterest && usersTeam && (
          <PointOfInterestPanel pointOfInterest={selectedPointOfInterest} usersTeam={usersTeam} />
        )}
      </div>
    </div>
  );
};

const PointOfInterestPanel = ({
  pointOfInterest,
  usersTeam,
}: {
  pointOfInterest: PointOfInterest;
  usersTeam: Team;
}) => {
  const { unlockPointOfInterest } = useMatchContext();
  const hasDiscovered = hasTeamDiscoveredPointOfInterest(usersTeam, pointOfInterest);
  return (
    <div className="flex flex-col items-center gap-4">
      <h3 className="text-2xl font-bold">{hasDiscovered ? pointOfInterest.name : 'Uncharted Waters'}</h3>
      <img src={hasDiscovered ? pointOfInterest.imageURL : `https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp`} alt={pointOfInterest.name}/>
      {!hasDiscovered && <p className="text-xl text-left"><span className="font-bold">Clue:</span> {pointOfInterest.clue}</p>}
      {hasDiscovered && <p className="text-xl text-left"><span className="font-bold">I:</span> {pointOfInterest.tierOneChallenge}</p>}
      {hasDiscovered && <p className="text-xl text-left"><span className="font-bold">II:</span> {pointOfInterest.tierTwoChallenge}</p>}
      {hasDiscovered && <p className="text-xl text-left"><span className="font-bold">III:</span> {pointOfInterest.tierThreeChallenge}</p>}
      {!hasDiscovered && <Button onClick={() => {
        navigator.geolocation.getCurrentPosition((position) => {
          unlockPointOfInterest(pointOfInterest.id, usersTeam.id, pointOfInterest.lat, pointOfInterest.lng);
        });
      }} title="I'm here!" />}
    </div>
  );
};
