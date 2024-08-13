import React, { useEffect, useRef, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import './MatchInProgress.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { PointOfInterest, Team, hasTeamDiscoveredPointOfInterest } from '@poltergeist/types';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { Button } from './shared/Button.tsx';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import { generateColorFromTeamName } from '../utils/generateColor.ts';

mapboxgl.accessToken =
  'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';

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
      <StatusIndicator tier={2} teamName={"Velvet Donkeytron"} yourTeamName={"dicks"} />
      {hasDiscovered && <TabNav tabs={['Info', 'Tier I', 'Tier II', 'Tier III']}>
        <TabItem key="Info">
          <p className="text-md text-left">{pointOfInterest.description}</p>
        </TabItem>
        <TabItem key="Tier I">
          <p className="text-md text-left">{pointOfInterest.tierOneChallenge}</p>
        </TabItem>
        <TabItem key="Tier II">
          <p className="text-md text-left">{pointOfInterest.tierTwoChallenge}</p>
        </TabItem>
        <TabItem key="Tier III">
          <p className="text-md text-left">{pointOfInterest.tierThreeChallenge}</p>
        </TabItem>
      </TabNav>}
      {!hasDiscovered && <p className="text-xl text-left"><span className="font-bold">Clue:</span> {pointOfInterest.clue}</p>}
      {!hasDiscovered && <Button onClick={() => {
        navigator.geolocation.getCurrentPosition((position) => {
          unlockPointOfInterest(pointOfInterest.id, usersTeam.id, pointOfInterest.lat, pointOfInterest.lng);
        });
      }} title="I'm here!" />}
    </div>
  );
};

type StatusCircleProps = {
  color: string;
}

type StatusIndicatorProps = {
  tier?: number | null;
  teamName?: string | null;
  yourTeamName: string;
}

const StatusCircle = ({ status }: { status: 'discovered' | 'unclaimed' | 'claimed' }) => {
  return <div style={{ width: '25px', height: '25px', backgroundColor: 'grey', borderRadius: '50%' }}></div>;
};

const StatusIndicator = ({ tier, teamName, yourTeamName }: StatusIndicatorProps) => {
  let color = 'grey';
  let text = 'Unclaimed';

  if (tier && teamName) {
    color = generateColorFromTeamName(teamName);

    if (teamName === yourTeamName) {
      text = 'Owned by you';
    } else {
      text = `${teamName}`;
    }
  }

  const numCircles: number[] = [];
  if (!tier) {
    numCircles.push(1);
  } else {
    for (let i = 0; i < tier; i++) {
      numCircles.push(1);
    }
  }

  return <div className="flex space-between justify-between items-center w-full">
    <div className="flex space-x-2">
      {numCircles.map((circle, index) => (
        <div key={index} style={{ width: '25px', height: '25px', backgroundColor: color, borderRadius: '50%' }}></div>
      ))}
    </div>
    <p className="text-md font-bold">{text}</p>
  </div>;
};
