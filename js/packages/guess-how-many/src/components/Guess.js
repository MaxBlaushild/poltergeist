var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
const react_1 = __importDefault(require('react'));
require('./Guess.css');
const react_2 = require('react');
function Guess(props) {
  const [guess, setGuess] = (0, react_2.useState)(0);
  const { text, checkGuess } = props;
  const onInputChange = (0, react_2.useCallback)(
    (e) => {
      const guess = e.target.value;
      if (!guess || /^\d+$/.test(guess)) {
        setGuess(guess);
      }
    },
    [setGuess]
  );
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
exports.default = Guess;
//# sourceMappingURL=Guess.js.map
