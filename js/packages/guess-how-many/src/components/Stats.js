var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
const react_1 = __importDefault(require('react'));
require('./Stats.css');
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
exports.default = Stats;
//# sourceMappingURL=Stats.js.map
