ALTER TABLE "sonar_categories" DROP CONSTRAINT "fk_sonar_categories_users";
ALTER TABLE "sonar_activities" DROP CONSTRAINT "fk_sonar_activities_users";

ALTER TABLE "sonar_categories" DROP COLUMN "user_id";
ALTER TABLE "sonar_activities" DROP COLUMN "user_id";
