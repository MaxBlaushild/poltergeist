import React from 'react';
import { useRouteError } from 'react-router-dom';

export function Error() {
  const error = useRouteError();

  return (
    <div id="error-page">
      <h1>Oops!</h1>
      <p>Sorry, an unexpected error has occurred.</p>
      <p>
        {/* @ts-ignore */}
        <i>{error?.statusText || error?.message}</i>
      </p>
    </div>
  );
}
