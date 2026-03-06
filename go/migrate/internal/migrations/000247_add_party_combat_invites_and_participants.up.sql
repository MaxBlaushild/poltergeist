ALTER TABLE monster_battles
ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'active';

ALTER TABLE monster_battles
ADD COLUMN IF NOT EXISTS turn_index INTEGER NOT NULL DEFAULT 0;

UPDATE monster_battles
SET state = 'active'
WHERE state IS NULL;

CREATE TABLE monster_battle_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    battle_id UUID NOT NULL REFERENCES monster_battles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_initiator BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (battle_id, user_id)
);

CREATE INDEX idx_monster_battle_participants_battle_id
ON monster_battle_participants(battle_id);

CREATE INDEX idx_monster_battle_participants_user_id
ON monster_battle_participants(user_id);

CREATE TABLE monster_battle_invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    battle_id UUID NOT NULL REFERENCES monster_battles(id) ON DELETE CASCADE,
    inviter_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    monster_id UUID NOT NULL REFERENCES monsters(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'accepted', 'declined', 'auto_declined')
    ),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    responded_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (battle_id, invitee_user_id)
);

CREATE INDEX idx_monster_battle_invites_invitee_status
ON monster_battle_invites(invitee_user_id, status, expires_at);

CREATE INDEX idx_monster_battle_invites_battle_status
ON monster_battle_invites(battle_id, status, expires_at);
