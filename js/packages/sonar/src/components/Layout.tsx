import React from 'react';
import { Outlet, Link } from 'react-router-dom';

export function Layout() {
    return (
      <div>
        <h1>EeeeeEeEeEeEeeeeeeeeeee</h1>
        <p>
          We are echolocating, baby! Give me your best screech.
        </p>
        <ul>
          <li>
            <Link to="/">Public Page</Link>
          </li>
          <li>
            <Link to="/surveys">Protected Page</Link>
          </li>
        </ul>
  
        <Outlet />
      </div>
    );
  }