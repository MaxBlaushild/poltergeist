import React from 'react';
import './Explanation.css';

function Explanation(props) {
  const { explanation } = props;

  return (
    <div className="Card Explanation__container">
      <p>{explanation}</p>
    </div>
  );
}

export default Explanation;
