-- Create QuestArchetypeNodes Table
CREATE TABLE quest_archetype_nodes (
    id UUID PRIMARY KEY,                            -- Node ID (UUID)
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                          -- Soft Delete Timestamp (nullable)
    location_archetype_id UUID NOT NULL,                    -- Foreign key to LocationArchType
    FOREIGN KEY (location_archetype_id) REFERENCES location_archetypes(id) -- Assuming LocationArchTypes table exists
);

-- Create QuestArchtypeChallenges Table
CREATE TABLE quest_archetype_challenges (
    id UUID PRIMARY KEY,                            -- Challenge ID (UUID)
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                          -- Soft Delete Timestamp (nullable)
    reward INT NOT NULL,                            -- Reward Points
    unlocked_node_id UUID,                         -- Foreign key to QuestArchtypeNode (nullable)
    FOREIGN KEY (unlocked_node_id) REFERENCES quest_archetype_nodes(id) -- Unlocked node reference (optional)
);

-- Create QuestArchTypeNodeChallenges Table (Bridge Table between QuestArchtypeChallenge and QuestArchtypeNode)
CREATE TABLE quest_archetype_node_challenges (
    id UUID PRIMARY KEY,                            -- Bridge Table ID (UUID)
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                           -- Soft Delete Timestamp (nullable)
    quest_archetype_challenge_id UUID NOT NULL,      -- Foreign key to QuestArchtypeChallenge
    quest_archetype_node_id UUID NOT NULL,           -- Foreign key to QuestArchtypeNode
    FOREIGN KEY (quest_archetype_challenge_id) REFERENCES quest_archetype_challenges(id), -- Challenge reference
    FOREIGN KEY (quest_archetype_node_id) REFERENCES quest_archetype_nodes(id)  -- Node reference
);

-- Create QuestArchetypes Table
CREATE TABLE quest_archetypes (
    id UUID PRIMARY KEY,                            -- QuestArchetype ID (UUID)
    name VARCHAR(255) NOT NULL,                      -- Quest Name
    created_at TIMESTAMP NOT NULL,                  -- Creation Timestamp
    updated_at TIMESTAMP NOT NULL,                  -- Update Timestamp
    deleted_at TIMESTAMP,                          -- Soft Delete Timestamp (nullable)
    root_id UUID NOT NULL,                          -- Foreign key to QuestArchetypeNode (Root Node)
    FOREIGN KEY (root_id) REFERENCES quest_archetype_nodes(id) -- Root node reference (starting node for the quest)
);
