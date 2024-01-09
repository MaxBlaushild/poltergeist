var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
const react_1 = __importDefault(require('react'));
require('./Explanation.css');
function Explanation(props) {
  const { explanation } = props;
  return (
    <div className="Card Explanation__container">
      <p>{explanation}</p>
    </div>
  );
}
exports.default = Explanation;
//# sourceMappingURL=Explanation.js.map
