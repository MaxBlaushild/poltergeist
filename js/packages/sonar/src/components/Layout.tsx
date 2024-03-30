import React from 'react';
import { Outlet, Link } from 'react-router-dom';

export function Layout() {
    return (
      <div>
        <h1>Sonar</h1>
        <p>
          Welcome to your friends' hobbies discovery app! Know who's down to do what in less than a minute.
        </p>
        <ul>
          <li>
            <Link to="/surveys">Surveys</Link>
          </li>
          <li>
            <Link to="/new-survey">Create Survey</Link>
          </li>
          <li>
            <Link to="/answers">Who's down for what</Link>
          </li>
        </ul>
  
        <Outlet />
      </div>
    );
  }