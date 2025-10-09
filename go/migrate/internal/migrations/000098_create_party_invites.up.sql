CREATE TABLE party_invites (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    inviter_id UUID NOT NULL REFERENCES users(id),
    invitee_id UUID NOT NULL REFERENCES users(id)
);
