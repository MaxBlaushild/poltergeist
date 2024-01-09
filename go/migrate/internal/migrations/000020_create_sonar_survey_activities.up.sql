CREATE TABLE sonar_survey_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    sonar_survey_id UUID NOT NULL,
    sonar_activity_id UUID NOT NULL,
    FOREIGN KEY (sonar_survey_id) REFERENCES sonar_surveys(id),
    FOREIGN KEY (sonar_activity_id) REFERENCES sonar_activities(id)
);