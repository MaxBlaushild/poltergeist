CREATE TABLE how_many_answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    how_many_question_id UUID NOT NULL,
    answer INTEGER NOT NULL,
    guess INTEGER NOT NULL,
    off_by INTEGER NOT NULL,
    correctness DOUBLE PRECISION NOT NULL,
    user_id UUID,
    FOREIGN KEY (how_many_question_id) REFERENCES how_many_questions(id)
);