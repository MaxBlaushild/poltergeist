CREATE TABLE location_archetypes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    included_types TEXT[], -- Array of included types (PlaceType)
    excluded_types TEXT[], -- Array of excluded types (PlaceType)
    challenges TEXT[]     -- Array of challenges
);