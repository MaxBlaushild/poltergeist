create table friends (
    id uuid primary key default uuid_generate_v4(),
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now(),
    first_user_id uuid not null,
    second_user_id uuid not null,
    -- Computed columns to ensure consistent ordering for unique constraint
    user1_id uuid generated always as (least(first_user_id, second_user_id)) stored,
    user2_id uuid generated always as (greatest(first_user_id, second_user_id)) stored,
    -- Unique constraint to prevent duplicate friendships regardless of order
    constraint unique_friendship unique (user1_id, user2_id)
);

create index idx_friends_first_user_id on friends (first_user_id);
create index idx_friends_second_user_id on friends (second_user_id);
create index idx_friends_first_user_id_second_user_id on friends (first_user_id, second_user_id);
create index idx_friends_second_user_id_first_user_id on friends (second_user_id, first_user_id);

alter table friends add constraint fk_friends_first_user_id foreign key (first_user_id) references users (id);
alter table friends add constraint fk_friends_second_user_id foreign key (second_user_id) references users (id);