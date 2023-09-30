import './Guess.css';
import { useState, useCallback } from 'react';

function Guess(props) {
  const [guess, setGuess] = useState(0);
  const updateGuess = useCallback(setGuess);
  const { text, checkGuess } = props;

  const onInputChange = useCallback((e) => {
    const guess = e.target.value;

    if (!guess || /^\d+$/.test(guess)) {
      updateGuess(guess);
    }
  });

  return (
    <div className="Card">
      <h4>TODAY'S QUESTION</h4>
      <p className="Guess__question">{text}</p>
      <div className="Guess__guesser">
        <input
          className="Guess__input"
          type="text"
          value={guess}
          onChange={onInputChange}
        />
        <button className="Guess__button" onClick={() => checkGuess(guess)}>
          <svg
            width="22"
            height="22"
            viewBox="0 0 22 22"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M11 0.333252L9.11337 2.21992L16.56 9.66659H0.333374V12.3333H16.56L9.11337 19.7799L11 21.6666L21.6667 10.9999L11 0.333252Z"
              fill="white"
            />
          </svg>
        </button>
      </div>
    </div>
  );
}

export default Guess;
