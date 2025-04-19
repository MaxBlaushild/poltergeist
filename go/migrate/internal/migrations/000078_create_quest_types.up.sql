-- Create QuestArchetypeNodes Table
CREATE TABLE QuestArchetypeNodes (
    id UUID PRIMARY KEY,                            -- Node ID (UUID)
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                          -- Soft Delete Timestamp (nullable)
    location_archetype_id UUID NOT NULL,                    -- Foreign key to LocationArchType
    FOREIGN KEY (location_archetype_id) REFERENCES LocationArchTypes(id) -- Assuming LocationArchTypes table exists
);

-- Create QuestArchtypeChallenges Table
CREATE TABLE QuestArchetypeChallenges (
    id UUID PRIMARY KEY,                            -- Challenge ID (UUID)
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                          -- Soft Delete Timestamp (nullable)
    reward INT NOT NULL,                            -- Reward Points
    unlocked_node_id UUID,                         -- Foreign key to QuestArchtypeNode (nullable)
    FOREIGN KEY (unlocked_node_id) REFERENCES QuestArchtypeNodes(id) -- Unlocked node reference (optional)
);

-- Create QuestArchTypeNodeChallenges Table (Bridge Table between QuestArchtypeChallenge and QuestArchtypeNode)
CREATE TABLE QuestArchTypeNodeChallenges (
    id UUID PRIMARY KEY,                            -- Bridge Table ID (UUID)
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                           -- Soft Delete Timestamp (nullable)
    quest_archetype_challenge_id UUID NOT NULL,      -- Foreign key to QuestArchtypeChallenge
    quest_archetype_node_id UUID NOT NULL,           -- Foreign key to QuestArchtypeNode
    FOREIGN KEY (quest_archetype_challenge_id) REFERENCES QuestArchtypeChallenges(id), -- Challenge reference
    FOREIGN KEY (quest_archetype_node_id) REFERENCES QuestArchetypeNodes(id)  -- Node reference
);

-- Create QuestArchetypes Table
CREATE TABLE QuestArchetypes (
    id UUID PRIMARY KEY,                            -- QuestArchetype ID (UUID)
    name VARCHAR(255) NOT NULL,                      -- Quest Name
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                          -- Soft Delete Timestamp (nullable)
    root_id UUID NOT NULL,                          -- Foreign key to QuestArchetypeNode (Root Node)
    FOREIGN KEY (root_id) REFERENCES QuestArchetypeNodes(id) -- Root node reference (starting node for the quest)
);

-- Example of Adding Indexes for Faster Lookups (Optional)
CREATE INDEX idx_quest_archtype_challenge_unlocked_node_id ON QuestArchtypeChallenges(unlocked_node_id);
CREATE INDEX idx_quest_archtype_challenge_node_id ON QuestArchtypeChallenges(quest_archetype_node_id);
CREATE INDEX idx_quest_archtype_type_challenge_id ON QuestArchtypeNodeChallenges(quest_archetype_challenge_id);
CREATE INDEX idx_quest_archtype_type_node_id ON QuestArchtypeNodeChallenges(quest_archetype_node_id);

-- Ensure that the LocationArchTypes table exists (assuming you have this already defined)
-- You can uncomment this if you are setting up the LocationArchTypes table as well:
-- CREATE TABLE LocationArchTypes (
--     id UUID PRIMARY KEY, -- LocationArchType ID
--     name VARCHAR(255) NOT NULL
-- );
