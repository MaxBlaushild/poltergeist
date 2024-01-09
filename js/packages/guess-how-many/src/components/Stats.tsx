import React from 'react';
import './Stats.css';

function Stats(props) {
  const { answer, guess, offBy } = props;
  return (
    <div className="Stats App__desktop-hidden">
      <p className="Stats__answer">There are {answer}</p>
      <p className="Stats__guess">You guessed {guess}</p>
      <div className="Stats__divider" />
      <p className="Stats__offBy">You missed it by {offBy}</p>
    </div>
  );
}

export default Stats;
