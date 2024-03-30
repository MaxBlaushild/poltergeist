ALTER TABLE sonar_surveys ADD COLUMN referrer_id UUID NOT NULL;
ALTER TABLE sonar_surveys ADD CONSTRAINT fk_referrer FOREIGN KEY (referrer_id) REFERENCES users(id);
ALTER TABLE sonar_surveys ADD COLUMN progenitor_id UUID NOT NULL;
ALTER TABLE sonar_surveys ADD CONSTRAINT fk_progenitor FOREIGN KEY (progenitor_id) REFERENCES sonar_survey(id);
