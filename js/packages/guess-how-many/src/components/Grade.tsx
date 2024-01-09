import React from 'react';
import './Grade.css';

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

export default Grade;
