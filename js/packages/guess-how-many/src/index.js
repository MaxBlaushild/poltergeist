var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
const react_1 = __importDefault(require('react'));
const react_dom_1 = __importDefault(require('react-dom'));
require('./index.css');
const App_1 = __importDefault(require('./App'));
const reportWebVitals_1 = __importDefault(require('./reportWebVitals'));
react_dom_1.default.render(
  <react_1.default.StrictMode>
    <App_1.default />
  </react_1.default.StrictMode>,
  document.getElementById('root')
);
// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
(0, reportWebVitals_1.default)(null);
//# sourceMappingURL=index.js.map
