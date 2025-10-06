create table parties (
    id uuid primary key default uuid_generate_v4(),
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

alter table users add column party_id uuid references parties(id);