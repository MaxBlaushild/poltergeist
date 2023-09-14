import { CircularProgressbar, buildStyles } from 'react-circular-progressbar';
import 'react-circular-progressbar/dist/styles.css';
import './Correctness.css';
import toast from 'react-hot-toast';

const flawless = '✨';
const great = '🔥';
const good = '👌';
const bad = '😳';
const gross = '🤮';

const getShareMessage = (text, correctness) => {
  let emoji;

  if (correctness === 1) {
    emoji = flawless;
  } else if (correctness > 0.9) {
    emoji = great;
  } else if (correctness > 0.7) {
    emoji = good;
  } else if (correctness > 0.4) {
    emoji = bad;
  } else {
    emoji = gross;
  }
  return `${text} - ${(correctness * 100).toFixed(2)}% correct: ${emoji}`;
};

const share = async (text, correctness) => {
  const shareString = getShareMessage(text, correctness);

  if (navigator && navigator.share) {
    await navigator.share({ text: shareString });
  }

  if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
    await navigator.clipboard.writeText(shareString);
    toast('Copied results to clipboard.');
  }
};

const convertCorrectness = (correctness) => Math.round(100 - correctness * 100);

const getOffness = (correctness) => {
  if (correctness === 1) {
    return '100% correct!';
  }

  return `Off by ${convertCorrectness(correctness)}%`;
};

function Correctness(props) {
  const { correctness, text } = props;
  const offness = getOffness(correctness);

  return (
    <div className="Correctness Card">
      <div className="Correctness__container">
        <h4 className="App__desktop-hidden">YOUR RESULT</h4>
        <p className="Correctness__question App__desktop-hidden">{text}</p>
        <div className="Rotated">
          <CircularProgressbar
            strokeWidth={1}
            styles={buildStyles({
              pathColor: '#00BBB0',
              trailColor: '#111114',
            })}
            circleRatio={0.5}
            value={correctness * 100}
          />
        </div>
        <p className="Correctness__offBy">{offness}</p>
        <button
          className="Correctness__button"
          onClick={() => share(text, correctness)}
        >
          Share
        </button>
      </div>
    </div>
  );
}

export default Correctness;
