alter table users add column username text;
create unique index users_username_idx on users (username);
