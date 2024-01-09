var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
const react_1 = __importDefault(require('react'));
require('./Grade.css');
function Grade(props) {
  const {
    text,
    guess,
    grade: { answer, offBy },
  } = props;
  return (
    <div className="Grade Card App__mobile-hidden">
      <h4>QUESTION</h4>
      <p className="Grade__question">{text}</p>
      <p className="Grade__answer">There are {answer}</p>
      <div className="Grade__divider" />
      <p className="Grade__guess">You guessed {guess}</p>
      <p className="Grade__offBy">You missed it by {offBy}</p>
    </div>
  );
}
exports.default = Grade;
//# sourceMappingURL=Grade.js.map
