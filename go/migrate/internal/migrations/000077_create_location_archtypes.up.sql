CREATE TABLE location_archetypes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR NOT NULL,
    included_types TEXT[], -- Array of included types (PlaceType)
    excluded_types TEXT[], -- Array of excluded types (PlaceType)
    challenges TEXT[],     -- Array of challenges
);