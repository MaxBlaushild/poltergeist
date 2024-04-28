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
import { Answers } from './components/Answers.tsx';
import { Dashboard } from './components/Dashboard.tsx';

function onlyAuthenticated({ request }: LoaderFunctionArgs) {
  if (!localStorage.getItem('token')) {
    let params = new URLSearchParams();
    params.set('from', new URL(request.url).pathname);
    return redirect('/login?' + params.toString());
  }
  return null;
}

function onlyUnauthenticated({ request }: LoaderFunctionArgs) {
  console.log(localStorage.getItem('token'));
  if (localStorage.getItem('token')) {
    return redirect('/');
  }
  return null;
}

export const router = createBrowserRouter([
  {
    id: 'root',
    path: '/',
    Component: Layout,
    children: [
      {
        index: true,
        Component: Home,
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
        path: 'dashboard',
        loader: onlyAuthenticated,
        Component: Dashboard,
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
        path: 'thanks',
        loader: onlyAuthenticated,
        Component: Thanks,
      },
      {
        path: 'answers',
        loader: onlyAuthenticated,
        Component: Answers,
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
