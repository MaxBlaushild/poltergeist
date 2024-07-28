import './App.css';
import React from 'react';
import GoogleMap from 'google-maps-react-markers';
import { useRef, useState, useCallback } from 'react';
import ClueImage from './two-bridges-clue.png';
import mapOptions from './map-options.json';
import markerPin from './crystal-node (2).png';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import AppBar from '@mui/material/AppBar';
import Modal from '@mui/material/Modal';
import cx from 'classnames';
import { Button, TextField } from '@mui/material';
import IconButton from '@mui/material/IconButton';
import axios from 'axios';
import Toolbar from '@mui/material/Toolbar';
import { MuiTelInput } from 'mui-tel-input';
import toast from 'react-hot-toast';

const style = {
  position: 'absolute',
  top: '50%',
  left: '50%',
  transform: 'translate(-50%, -50%)',
  width: '70%',
  bgcolor: 'background.paper',
  border: '2px solid #000',
  overflow: 'auto' /* Make the content of the modal scrollable */,
  webkitOverflowScrolling: 'touch',
  boxShadow: 24,
  p: 4,
};

const getTeamColor = (teamName) => {
  if (teamName === 'Gold') {
    return '#F2A44E';
  }

  if (teamName === 'Orange') {
    return '#FF563D';
  }

  if (teamName === 'Purple') {
    return '#834FFF';
  }

  if (teamName === 'Blue') {
    return '#4290F5';
  }

  return '';
};

const Marker = ({
  className,
  lat,
  lng,
  markerId,
  crystalLocationHint,
  attuneChallenge,
  captureChallenge,
  attuned,
  onClick,
  ...props
}) => {
  return (
    <img
      className={className}
      src={markerPin}
      lat={lat}
      lng={lng}
      onClick={(e) =>
        onClick
          ? onClick(e, {
              markerId,
              lat,
              lng,
              crystalLocationHint,
              attuneChallenge,
              captureChallenge,
              attuned,
            })
          : null
      }
      style={{ cursor: 'pointer', fontSize: 40 }}
      alt={markerId}
      {...props}
    />
  );
};

const getUserID = () => {
  const stringId = localStorage.getItem('user-id');

  if (stringId) {
    return parseInt(stringId);
  }

  return null;
};

function App() {
  const mapRef = useRef(null);
  const [mapReady, setMapReady] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCrystal, setSelectedCrystal] = useState(null);
  const [userID, setUserID] = useState(getUserID());
  const [phoneNumber, setPhoneNumber] = useState('');
  const [name, setName] = useState('');
  const [shouldRegister, setShouldRegister] = useState(false);
  const [isLoggingIn, setIsLoggingIn] = useState(false);
  const [crystals, setCrystals] = useState(null);
  const [teams, setTeams] = useState(null);
  const [users, setUsers] = useState(null);
  const [code, setCode] = useState('');
  const [rulesOpen, setRulesOpen] = useState(false);
  const [waitingOnVerificationCode, setWaitingOnVerificationCode] =
    useState(false);

  const onGoogleApiLoaded = async ({ map, maps }) => {
    mapRef.current = map;
    setMapReady(true);

    const {
      data: { teams, users },
    } = await axios.get(
      `${process.env.REACT_APP_API_URL}/crystal-crisis/teams`
    );
    setTeams(teams);
    setUsers(users);

    const { data: crystals } = await axios.get(
      `${process.env.REACT_APP_API_URL}/crystal-crisis/crystals/${
        teams?.find((team) =>
          team.UserTeams.find((userTeam) => userTeam.UserID == userID)
        )?.ID || 1000000000
      }`
    );
    setCrystals(crystals);

    const { data: neighbors } = await axios.get(
      `${process.env.REACT_APP_API_URL}/crystal-crisis/neighbors`
    );

    neighbors.forEach((neighbor) => {
      const one = crystals.find(
        (crystal) => crystal.ID === neighbor.crystalOneId
      );
      const two = crystals.find(
        (crystal) => crystal.ID === neighbor.crystalTwoId
      );

      let strokeColor = '#C8C8C8';
      let strokeOpacity = 0.5;
      if (one.captureTeamId && one.captureTeamId === two.captureTeamId) {
        strokeColor = getTeamColor(
          teams?.find((team) => team.ID === one.captureTeamId)?.Name
        );
        strokeOpacity = 1.0;
      }

      const flightPath = new maps.Polyline({
        path: [
          {
            lat: parseFloat(one.lat),
            lng: parseFloat(one.lng),
          },
          {
            lat: parseFloat(two.lat),
            lng: parseFloat(two.lng),
          },
        ],
        geodesic: true,
        strokeColor,
        strokeOpacity,
        strokeWeight: 5,
      });

      flightPath.setMap(map);
    });
  };

  const onMarkerClick = (e, crystal) => {
    setSelectedCrystal(crystal);
    setIsModalOpen(true);
  };

  return (
    <>
      <GoogleMap
        apiKey="AIzaSyDff3XqCOiu01dgC46rS2mIGk92rx6-d0Q"
        defaultCenter={{ lat: 40.71762378744178, lng: -73.99844795603595 }}
        defaultZoom={14}
        options={mapOptions}
        mapMinHeight="100vh"
        onGoogleApiLoaded={onGoogleApiLoaded}
        onChange={(map) => console.log('Map moved', map)}
      >
        {crystals?.map(
          (
            {
              lat,
              lng,
              name,
              clue,
              captureTeamId,
              attuned,
              captureChallenge,
              attuneChallenge,
            },
            index
          ) => (
            <Marker
              key={index}
              lat={parseFloat(lat)}
              lng={parseFloat(lng)}
              crystalLocationHint={clue}
              attuned={attuned}
              captureChallenge={captureChallenge}
              attuneChallenge={attuneChallenge}
              className={cx(
                'Marker',
                teams?.find((team) => team.ID === captureTeamId)?.Name,
                attuned ? 'Attuned' : ''
              )}
              markerId={name}
              onClick={onMarkerClick}
            />
          )
        )}
      </GoogleMap>
      <Modal
        open={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setSelectedCrystal(null);
        }}
        aria-labelledby="modal-modal-title"
        aria-describedby="modal-modal-description"
      >
        <Box sx={style}>
          <Typography
            id="modal-modal-title"
            variant="h6"
            component="h2"
            style={{ fontFamily: 'Poppins' }}
          >
            {selectedCrystal?.markerId}
          </Typography>
          {selectedCrystal?.crystalLocationHint &&
            selectedCrystal?.crystalLocationHint !== 'picture-clue' && (
              <Typography
                id="modal-modal-description"
                sx={{ mt: 2 }}
                style={{ fontFamily: 'Poppins' }}
              >
                {selectedCrystal?.crystalLocationHint}
              </Typography>
            )}
          {selectedCrystal?.crystalLocationHint &&
            selectedCrystal?.crystalLocationHint === 'picture-clue' && (
              <img src={ClueImage} />
            )}
          {selectedCrystal?.captureChallenge && (
            <Typography
              id="modal-modal-description"
              sx={{ mt: 2 }}
              style={{ fontFamily: 'Poppins' }}
            >
              <strong>Capture:</strong> {selectedCrystal?.captureChallenge}
            </Typography>
          )}
          {selectedCrystal?.attuneChallenge && (
            <Typography
              id="modal-modal-description"
              sx={{ mt: 2 }}
              style={{ fontFamily: 'Poppins' }}
            >
              <strong>Attune:</strong> {selectedCrystal?.attuneChallenge}
            </Typography>
          )}
        </Box>
      </Modal>
      <Modal
        open={rulesOpen}
        onClose={() => {
          setRulesOpen(false);
        }}
        aria-labelledby="modal-modal-title"
        aria-describedby="modal-modal-description"
      >
        <Box sx={style}>
          <Typography
            id="modal-modal-title"
            variant="h6"
            component="h2"
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            The Rules
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            First and foremost, if you have any questions, text me! My number is
            1 (440) 785-8475.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            The goal of the game is to posess as many of the crystals (the
            circles) as possible at the end of the game. One point is awarded
            per crystal owned. If you control two adjacent circles, the path
            between them will light up, and you will get one extra point.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            Each circle starts with a clue. This clue will lead you to a
            location near where the circle is on the map. The circle location is
            an approximation though and has no correlation to the actual
            location the clue alludes to.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            When you think you are at the location, text me a picture of it. If
            you are indeed there, I will activate the challenges for that
            location. If not, I'll give you a clue.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            Challenges and clues can be accessed by tapping the circle on the
            map.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            Each location has a <strong>capture</strong> and{' '}
            <strong>attune</strong> challenge.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            Completing a <strong>capture</strong> challenge will gain possession
            of the circle for your team, but will still allow other teams to
            take posession of the circle from your team by completing the same
            challenge.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            Completing an <strong>attune</strong> challenge will gain possession
            of the circle for your team and lock out other teams from stealing
            it from you. They are harder.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            When you think you are done with a challenge, text me proof. I'll do
            the computer thing to give you posession if I agree. Please don't be
            a dick and try to cheat; this is just for fun.
          </Typography>
          <Typography
            id="modal-modal-description"
            sx={{ mt: 2 }}
            style={{ fontFamily: 'Poppins', fontSize: 8 }}
          >
            And that's it! I'll be acting as human DM, so text me all questions,
            comments and concerns.
          </Typography>
        </Box>
      </Modal>
      <Modal
        open={isLoggingIn}
        onClose={() => {
          setIsModalOpen(false);
          setSelectedCrystal(null);
        }}
        aria-labelledby="modal-modal-title"
        aria-describedby="modal-modal-description"
      >
        <Box sx={style}>
          <div className="Register">
            <MuiTelInput
              defaultCountry="US"
              forceCallingCode
              value={phoneNumber}
              onChange={(newPhoneNumber) => {
                if (newPhoneNumber.replace(/\s+/g, '').length <= 12) {
                  setPhoneNumber(newPhoneNumber);
                }
              }}
            />
            {shouldRegister && (
              <div className="Register">
                <div className="Divider"></div>
                <Typography
                  style={{
                    fontFamily: 'Poppins',
                    color: 'rgba(0, 0, 0, 0.75)',
                  }}
                >
                  Add your name to sign up and join the fight.
                </Typography>
                <TextField
                  required
                  style={{
                    fontFamily: 'Poppins',
                  }}
                  id="name-required"
                  label="Name"
                  placeholder="Robert Moses"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
            )}
            {waitingOnVerificationCode && (
              <TextField
                required
                style={{
                  fontFamily: 'Poppins',
                }}
                id="verification-code-required"
                label="Verification code"
                placeholder="XXXXXX"
                value={code}
                onChange={(e) => {
                  const inputValue = e.target.value;

                  if (/^\d*$/.test(inputValue) && inputValue.length <= 6) {
                    setCode(inputValue);
                  }
                }}
              />
            )}
            <Button
              variant="contained"
              style={{
                background: '#4290F5',
              }}
              disabled={waitingOnVerificationCode ? code.length !== 6 : false}
              onClick={
                waitingOnVerificationCode
                  ? shouldRegister
                    ? register
                    : login
                  : getVerificationCode
              }
            >
              {shouldRegister ? 'Register' : 'Join the fight'}
            </Button>
          </div>
        </Box>
      </Modal>
    </>
  );
}

export default App;
