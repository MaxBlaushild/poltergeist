import React from 'react';
import { LoaderFunctionArgs, createBrowserRouter } from 'react-router-dom';
import { Layout } from './components/Layout.tsx';
import { Home } from './components/Home.tsx';
import { redirect } from 'react-router-dom';
import { Surveys } from './components/Surveys.tsx';
import { Error } from './components/Error.tsx';
import { NewSurvey } from './components/NewSurvey.tsx';
import { AnswerSurvey } from './components/AnswerSurvey.tsx';
import { Survey } from './components/Survey.tsx';
import { Thanks } from './components/Thanks.tsx';
import { AssembleCrew } from './components/AssembleCrew.tsx';
import { Submission } from './components/Submission.tsx';
import { SelectBattleArena } from './components/SelectBattleArena.tsx';
import { Match } from './components/Match.tsx';
import { CurrentMatch } from './components/CurrentMatch.tsx';
import { MatchById } from './components/MatchById.tsx';
import { TeamScore } from './components/TeamScore.tsx';
import Admin from './components/Admin.tsx';
import { CreatePointsOfInterest } from './components/CreatePointsOfInterest.tsx';
import { SinglePlayer } from './components/SinglePlayer.tsx';
import { MatchInProgress } from './components/MatchInProgress.tsx';
import { MatchLobby } from './components/MatchLobby.tsx';
import { QuestLogContextProvider } from './contexts/QuestLogContext.tsx';
import { MapProvider } from '@poltergeist/contexts';
import { PointOfInterestContextProvider } from './contexts/PointOfInterestContext.tsx';
import { MatchContextProvider } from './contexts/MatchContext.tsx';
import { TagProvider } from '@poltergeist/contexts';
import { ZoneProvider } from '@poltergeist/contexts';
import { CompletedTaskProvider } from './contexts/CompletedTaskContext.tsx';
import { PartyProvider } from './contexts/PartyContext.tsx';
import { ActivityFeedProvider } from './contexts/ActivityFeedContext.tsx';

function onlyAuthenticated({ request }: LoaderFunctionArgs) {
  if (!localStorage.getItem('token')) {
    let params = new URLSearchParams();
    params.set('from', new URL(request.url).pathname);
    return redirect('/?' + params.toString());
  }
  return null;
}

function onlyUnauthenticated({ request }: LoaderFunctionArgs) {
  if (localStorage.getItem('token')) {
    return redirect('/single-player');
  }
  return null;
}

export const router = createBrowserRouter([
  {
    id: 'root',
    path: '/',
    Component: () => (
      <PartyProvider>
        <Layout />
      </PartyProvider>
    ),
    children: [
      {
        index: true,
        Component: Home,
        loader: onlyUnauthenticated,
      },
      {
        path: 'surveys',
        loader: onlyAuthenticated,
        Component: Surveys,
      },
      {
        path: 'new-survey',
        loader: onlyAuthenticated,
        Component: NewSurvey,
      },
      {
        path: 'submit-answer/:id',
        loader: onlyAuthenticated,
        Component: AnswerSurvey,
      },
      {
        path: 'surveys/:id',
        loader: onlyAuthenticated,
        Component: Survey,
      },
      {
        path: 'submissions/:id',
        loader: onlyAuthenticated,
        Component: Submission,
      },
      {
        path: 'thanks',
        loader: onlyAuthenticated,
        Component: Thanks,
      },
      {
        path: 'assemble-crew',
        loader: onlyAuthenticated,
        Component: AssembleCrew,
      },
      {
        path: 'select-battle-arena',
        loader: onlyAuthenticated,
        Component: () => (
          <MatchContextProvider>
            <PointOfInterestContextProvider>
              <MapProvider>
                <SelectBattleArena />
              </MapProvider>
            </PointOfInterestContextProvider>
          </MatchContextProvider>
        ),
      },
      {
        path: 'single-player',
        loader: onlyAuthenticated,
        Component: () => (
          <TagProvider>
            <MapProvider>
              <ZoneProvider>
                <QuestLogContextProvider>
                  <ActivityFeedProvider>
                    <CompletedTaskProvider>
                      <SinglePlayer />
                    </CompletedTaskProvider>
                  </ActivityFeedProvider>
                </QuestLogContextProvider>
              </ZoneProvider>
            </MapProvider>
          </TagProvider>
        ),
      },
      {
        path: 'adminfuckoff',
        loader: onlyAuthenticated,
        Component: Admin,
      },
      {
        path: 'create-point-of-interest',
        loader: onlyAuthenticated,
        Component: CreatePointsOfInterest,
      },
    ],
  },
  {
    path: '/logout',
    async action() {
      localStorage.removeItem('token');
      return redirect('/');
    },
  },
]);
