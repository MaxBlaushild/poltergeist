DO $$
DECLARE
    playing_sports_id UUID;
    watching_sports_id UUID;
    watching_things_id UUID;
    working_out_id UUID;
    looking_at_stuff_id UUID;
    making_stuff_id UUID;
    gaming_id UUID;
    gambling_id UUID;
    consuming_id UUID;
    music_id UUID;
    partying_id UUID;
    sex_id UUID;
BEGIN

    ALTER TABLE sonar_activities
    ADD COLUMN sonar_category_id UUID;

    ALTER TABLE sonar_activities
    ADD CONSTRAINT fk_sonar_activities_sonar_categories
    FOREIGN KEY (sonar_category_id)
    REFERENCES sonar_categories (id);

    -- Playing Sports
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Playing Sports') RETURNING id INTO playing_sports_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Pickleball', playing_sports_id),
    (NOW(), NOW(), 'Tennis', playing_sports_id),
    (NOW(), NOW(), 'Basketball', playing_sports_id),
    (NOW(), NOW(), 'Golf', playing_sports_id),
    (NOW(), NOW(), 'Football', playing_sports_id),
    (NOW(), NOW(), 'Volleyball', playing_sports_id),
    (NOW(), NOW(), 'Surfing', playing_sports_id),
    (NOW(), NOW(), 'Skiing/snowboarding', playing_sports_id);

    -- Watching Sports
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Watching Sports') RETURNING id INTO watching_sports_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'NBA', watching_sports_id),
    (NOW(), NOW(), 'NFL', watching_sports_id),
    (NOW(), NOW(), 'Soccer', watching_sports_id),
    (NOW(), NOW(), 'Hockey', watching_sports_id),
    (NOW(), NOW(), 'College Football', watching_sports_id),
    (NOW(), NOW(), 'College Basketball', watching_sports_id);

    -- -- Watching Things
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Watching Things') RETURNING id INTO watching_things_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Funny movies', watching_things_id),
    (NOW(), NOW(), 'Sad movies', watching_things_id),
    (NOW(), NOW(), 'Action movies', watching_things_id),
    (NOW(), NOW(), 'Cartoons', watching_things_id),
    (NOW(), NOW(), 'Foreign Films', watching_things_id),
    (NOW(), NOW(), 'Game shows', watching_things_id);

    -- -- Working Out
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Working Out') RETURNING id INTO working_out_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Weights', working_out_id),
    (NOW(), NOW(), 'Swimming', working_out_id),
    (NOW(), NOW(), 'Yoga', working_out_id),
    (NOW(), NOW(), 'Pilates', working_out_id),
    (NOW(), NOW(), 'Rock Climbing', working_out_id),
    (NOW(), NOW(), 'Running', working_out_id),
    (NOW(), NOW(), 'Dance', working_out_id),
    (NOW(), NOW(), 'Barre', working_out_id);

    -- -- Looking at Stuff
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Looking at Stuff') RETURNING id INTO looking_at_stuff_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Galleries', looking_at_stuff_id),
    (NOW(), NOW(), 'History Museums', looking_at_stuff_id),
    (NOW(), NOW(), 'Classic Art', looking_at_stuff_id),
    (NOW(), NOW(), 'Old coins', looking_at_stuff_id),
    (NOW(), NOW(), 'Modern Art Museum', looking_at_stuff_id),
    (NOW(), NOW(), 'Architecture tours', looking_at_stuff_id);

    -- -- Making Stuff
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Making Stuff') RETURNING id INTO making_stuff_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Painting', making_stuff_id),
    (NOW(), NOW(), 'Drawing', making_stuff_id),
    (NOW(), NOW(), 'Figure Drawing', making_stuff_id),
    (NOW(), NOW(), 'Woodworking', making_stuff_id),
    (NOW(), NOW(), 'Carpentry', making_stuff_id),
    (NOW(), NOW(), 'Throwing Pots', making_stuff_id),
    (NOW(), NOW(), 'Sculpting', making_stuff_id),
    (NOW(), NOW(), 'Welding', making_stuff_id);

    -- -- Gaming
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Gaming') RETURNING id INTO gaming_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Fun board games', gaming_id),
    (NOW(), NOW(), 'Complicated board games', gaming_id),
    (NOW(), NOW(), 'Card games', gaming_id),
    (NOW(), NOW(), 'PC games', gaming_id),
    (NOW(), NOW(), 'Console games', gaming_id),
    (NOW(), NOW(), 'Couch co-op', gaming_id),
    (NOW(), NOW(), 'Barcades', gaming_id),
    (NOW(), NOW(), 'Collectible card games', gaming_id),
    (NOW(), NOW(), 'Escape Rooms', gaming_id),
    (NOW(), NOW(), 'Bar trivia', gaming_id);

    --     -- Gambling
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Gambling') RETURNING id INTO gambling_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Poker', gambling_id),
    (NOW(), NOW(), 'Ponies', gambling_id),
    (NOW(), NOW(), 'Casinos', gambling_id),
    (NOW(), NOW(), 'Sports betting', gambling_id);

    -- -- Consuming
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Consuming') RETURNING id INTO consuming_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Cheap', consuming_id),
    (NOW(), NOW(), 'Bougie', consuming_id),
    (NOW(), NOW(), 'Weird', consuming_id),
    (NOW(), NOW(), 'Shellfish', consuming_id),
    (NOW(), NOW(), 'Seafood', consuming_id),
    (NOW(), NOW(), 'Sushi', consuming_id),
    (NOW(), NOW(), 'Vegetarian', consuming_id),
    (NOW(), NOW(), 'Vegan', consuming_id);

    -- -- Music
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Music') RETURNING id INTO music_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Jamming', music_id),
    (NOW(), NOW(), 'Bands', music_id),
    (NOW(), NOW(), 'Producing', music_id),
    (NOW(), NOW(), 'Composing', music_id),
    (NOW(), NOW(), 'Concerts', music_id),
    (NOW(), NOW(), 'Dance parties', music_id),
    (NOW(), NOW(), 'Music festivals', music_id);

    -- -- Partying
    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Partying') RETURNING id INTO partying_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES 
    (NOW(), NOW(), 'Dives', partying_id),
    (NOW(), NOW(), 'Wine bars', partying_id),
    (NOW(), NOW(), 'Craft cocktails', partying_id),
    (NOW(), NOW(), 'Raves (sober)', partying_id),
    (NOW(), NOW(), 'Raves (drugs)', partying_id),
    (NOW(), NOW(), 'Raves (booze)', partying_id),
    (NOW(), NOW(), 'Mushrooms', partying_id),
    (NOW(), NOW(), 'Acid', partying_id),
    (NOW(), NOW(), 'Molly', partying_id),
    (NOW(), NOW(), 'Ketamine', partying_id);

    INSERT INTO sonar_categories (created_at, updated_at, title) VALUES (NOW(), NOW(), 'Sex') RETURNING id INTO sex_id;
    INSERT INTO sonar_activities (created_at, updated_at, title, sonar_category_id) VALUES
    (NOW(), NOW(), 'Swapping', sex_id),
    (NOW(), NOW(), 'Sex parties', sex_id),
    (NOW(), NOW(), 'Group sex', sex_id),
    (NOW(), NOW(), 'Dom', sex_id),
    (NOW(), NOW(), 'Sub', sex_id),
    (NOW(), NOW(), 'Switch', sex_id),
    (NOW(), NOW(), 'Casual', sex_id);
END $$;