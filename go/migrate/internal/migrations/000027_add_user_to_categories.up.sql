ALTER TABLE "sonar_activities" ADD COLUMN "user_id" UUID NULL;
ALTER TABLE "sonar_categories" ADD COLUMN "user_id" UUID NULL;

ALTER TABLE "sonar_activities" ADD CONSTRAINT "fk_sonar_activities_users" FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "sonar_categories" ADD CONSTRAINT "fk_sonar_categories_users" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

