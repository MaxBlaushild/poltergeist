-- Create fete_teams table
CREATE TABLE fete_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL
);

-- Create index for soft deletes
CREATE INDEX idx_fete_teams_deleted_at ON fete_teams(deleted_at);

-- Create fete_team_users join table
CREATE TABLE fete_team_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    fete_team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    FOREIGN KEY (fete_team_id) REFERENCES fete_teams(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create indexes for fete_team_users
CREATE INDEX idx_fete_team_users_deleted_at ON fete_team_users(deleted_at);
CREATE INDEX idx_fete_team_users_fete_team_id ON fete_team_users(fete_team_id);
CREATE INDEX idx_fete_team_users_user_id ON fete_team_users(user_id);

-- Create fete_rooms table
CREATE TABLE fete_rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    open BOOLEAN NOT NULL,
    current_team_id UUID NOT NULL,
    FOREIGN KEY (current_team_id) REFERENCES fete_teams(id)
);

-- Create indexes for fete_rooms
CREATE INDEX idx_fete_rooms_deleted_at ON fete_rooms(deleted_at);
CREATE INDEX idx_fete_rooms_current_team_id ON fete_rooms(current_team_id);

-- Create fete_room_teams join table
CREATE TABLE fete_room_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    fete_room_id UUID NOT NULL,
    team_id UUID NOT NULL,
    FOREIGN KEY (fete_room_id) REFERENCES fete_rooms(id),
    FOREIGN KEY (team_id) REFERENCES fete_teams(id)
);

-- Create indexes for fete_room_teams
CREATE INDEX idx_fete_room_teams_deleted_at ON fete_room_teams(deleted_at);
CREATE INDEX idx_fete_room_teams_fete_room_id ON fete_room_teams(fete_room_id);
CREATE INDEX idx_fete_room_teams_team_id ON fete_room_teams(team_id);

-- Create fete_room_linked_list_teams table (queue structure)
CREATE TABLE fete_room_linked_list_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    fete_room_id UUID NOT NULL,
    first_team_id UUID NOT NULL,
    second_team_id UUID NOT NULL,
    FOREIGN KEY (fete_room_id) REFERENCES fete_rooms(id),
    FOREIGN KEY (first_team_id) REFERENCES fete_teams(id),
    FOREIGN KEY (second_team_id) REFERENCES fete_teams(id)
);

-- Create indexes for fete_room_linked_list_teams
CREATE INDEX idx_fete_room_linked_list_teams_deleted_at ON fete_room_linked_list_teams(deleted_at);
CREATE INDEX idx_fete_room_linked_list_teams_fete_room_id ON fete_room_linked_list_teams(fete_room_id);
CREATE INDEX idx_fete_room_linked_list_teams_first_team_id ON fete_room_linked_list_teams(first_team_id);
CREATE INDEX idx_fete_room_linked_list_teams_second_team_id ON fete_room_linked_list_teams(second_team_id);

