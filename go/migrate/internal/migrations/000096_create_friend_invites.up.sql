create table friend_invites (
    id uuid primary key default uuid_generate_v4(),
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now(),
    inviter_id uuid not null,
    invitee_id uuid not null
);

alter table friend_invites add constraint fk_friend_invites_inviter_id foreign key (inviter_id) references users (id);
alter table friend_invites add constraint fk_friend_invites_invitee_id foreign key (invitee_id) references users (id);