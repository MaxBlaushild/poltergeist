alter table quest_archetype_challenges drop column reward;
alter table quest_archetype_challenges add column reward uuid;
alter table quest_archetype_challenges add constraint fk_quest_archetype_challenges_reward foreign key (reward) references inventory_items(id);