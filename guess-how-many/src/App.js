import './App.css';
import { useState, useEffect, useCallback } from 'react';
import { Toaster } from 'react-hot-toast';
import axios from 'axios';
import Guess from './components/Guess';
import Grade from './components/Grade';
import Correctness from './components/Correctness';
import Stats from './components/Stats';
import Explanation from './components/Explanation';
import Subscribe from './components/Subscribe';
import { getUserID } from './util';

const noGrade = {
  correctness: 0,
  answer: 0,
  offBy: 0,
  question: '',
  guess: 0,
};

function App() {
  const [text, setText] = useState('');
  const [grade, setGrade] = useState(noGrade);
  const [guess, setGuess] = useState(0);
  const [questionId, setQuestionId] = useState(0);
  const [explanation, setExplanation] = useState('');
  const userId = getUserID();

  const fetchText = async () => {
    const res = await axios.get(
      `${process.env.REACT_APP_API_URL}/trivai/how_many_questions/current`,
    );
    const {
      data: { text, ID: id, explanation },
    } = res;
    try {
      let url = `${process.env.REACT_APP_API_URL}/trivai/how_many_questions/answer?questionId=${id}&`;

      if (userId.userId) {
        url += `userId=${userId.userId}`;
      } else {
        url += `ephemeralUserId=${userId.ephemeralUserId}`;
      }
      const answerRes = await axios.get(url);
      const { data: grade } = answerRes;
      console.log(grade);
      setGrade(grade);
      setGuess(grade.guess);
    } catch (e) {}

    setExplanation(explanation);
    setQuestionId(id);
    setText(text);
  };

  const checkGuess = useCallback(async (_guess) => {
    const res = await axios.post(
      `${process.env.REACT_APP_API_URL}/trivai/how_many_questions/grade`,
      { guess: parseInt(_guess), id: questionId, ...userId },
    );
    const { data } = res;
    setGrade(data);
    setGuess(_guess);
  });

  useEffect(() => {
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
                <Grade grade={grade} text={text} guess={guess} />
                <Correctness correctness={correctness} text={text} />
              </div>
              <Stats guess={guess} answer={grade.answer} offBy={grade.offBy} />
            </>
          ) : (
            <Guess text={text} checkGuess={checkGuess} />
          )}

          {grade.correctness ? <Explanation explanation={explanation} /> : null}
          <Subscribe />
        </div>
      </div>
      <Toaster />
    </div>
  );
}

export default App;
