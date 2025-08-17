alter table quest_archetype_challenges drop constraint fk_quest_archetype_challenges_reward;
alter table quest_archetype_challenges drop column reward;
alter table quest_archetype_challenges add column reward integer not null;
