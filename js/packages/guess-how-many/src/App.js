var __createBinding =
  (this && this.__createBinding) ||
  (Object.create
    ? function (o, m, k, k2) {
        if (k2 === undefined) k2 = k;
        var desc = Object.getOwnPropertyDescriptor(m, k);
        if (
          !desc ||
          ('get' in desc ? !m.__esModule : desc.writable || desc.configurable)
        ) {
          desc = {
            enumerable: true,
            get: function () {
              return m[k];
            },
          };
        }
        Object.defineProperty(o, k2, desc);
      }
    : function (o, m, k, k2) {
        if (k2 === undefined) k2 = k;
        o[k2] = m[k];
      });
var __setModuleDefault =
  (this && this.__setModuleDefault) ||
  (Object.create
    ? function (o, v) {
        Object.defineProperty(o, 'default', { enumerable: true, value: v });
      }
    : function (o, v) {
        o['default'] = v;
      });
var __importStar =
  (this && this.__importStar) ||
  function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null)
      for (var k in mod)
        if (k !== 'default' && Object.prototype.hasOwnProperty.call(mod, k))
          __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
  };
var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
require('./App.css');
const react_1 = __importStar(require('react'));
const react_hot_toast_1 = require('react-hot-toast');
const axios_1 = __importDefault(require('axios'));
const Guess_1 = __importDefault(require('./components/Guess'));
const Grade_1 = __importDefault(require('./components/Grade'));
const Correctness_1 = __importDefault(require('./components/Correctness'));
const Stats_1 = __importDefault(require('./components/Stats'));
const Explanation_1 = __importDefault(require('./components/Explanation'));
const Subscribe_1 = __importDefault(require('./components/Subscribe'));
const util_1 = require('./util');
const noGrade = {
  correctness: 0,
  answer: 0,
  offBy: 0,
  question: '',
  guess: 0,
};
function App() {
  const [text, setText] = (0, react_1.useState)('');
  const [grade, setGrade] = (0, react_1.useState)(noGrade);
  const [guess, setGuess] = (0, react_1.useState)(0);
  const [questionId, setQuestionId] = (0, react_1.useState)(0);
  const [explanation, setExplanation] = (0, react_1.useState)('');
  const userId = (0, util_1.getUserID)();
  const fetchText = async () => {
    var _a;
    const res = await axios_1.default.get(
      `${process.env.REACT_APP_API_URL}/trivai/how_many_questions/current`
    );
    const {
      data: { text, ID: id, explanation },
    } = res;
    try {
      let url = `${process.env.REACT_APP_API_URL}/trivai/how_many_questions/answer?questionId=${id}&`;
      url += `userId=${
        (_a = userId.userId) !== null && _a !== void 0
          ? _a
          : userId.ephemeralUserId
      }`;
      const answerRes = await axios_1.default.get(url);
      const { data: grade } = answerRes;
      setGrade(grade);
      setGuess(grade.guess);
    } catch (e) {}
    setExplanation(explanation);
    setQuestionId(id);
    setText(text);
  };
  const checkGuess = (0, react_1.useCallback)(
    async (_guess) => {
      var _a;
      const res = await axios_1.default.post(
        `${process.env.REACT_APP_API_URL}/trivai/how_many_questions/grade`,
        {
          guess: parseInt(_guess),
          id: questionId,
          userId:
            (_a = userId.userId) !== null && _a !== void 0
              ? _a
              : userId.ephemeralUserId,
        }
      );
      const { data } = res;
      setGrade(data);
      setGuess(_guess);
    },
    [setGrade, setGuess]
  );
  (0, react_1.useEffect)(() => {
    fetchText();
  }, []);
  const { correctness } = grade;
  return (
    <div className="App">
      <div className="App__container">
        <div className="App__Headers">
          <h2>Guess how many?</h2>
          <h3>The daily sizing game</h3>
        </div>
        <div className="App__content">
          {grade.correctness ? (
            <>
              <div className="GradeRow">
                <Grade_1.default grade={grade} text={text} guess={guess} />
                <Correctness_1.default correctness={correctness} text={text} />
              </div>
              <Stats_1.default
                guess={guess}
                answer={grade.answer}
                offBy={grade.offBy}
              />
            </>
          ) : (
            <Guess_1.default text={text} checkGuess={checkGuess} />
          )}

          {grade.correctness ? (
            <Explanation_1.default explanation={explanation} />
          ) : null}
          <Subscribe_1.default />
        </div>
      </div>
      <react_hot_toast_1.Toaster />
    </div>
  );
}
exports.default = App;
//# sourceMappingURL=App.js.map
