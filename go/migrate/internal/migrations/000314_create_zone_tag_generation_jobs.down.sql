DROP TABLE IF EXISTS zone_tag_generation_jobs;

DELETE FROM tags
WHERE tag_group_id IN (
  SELECT id FROM tag_groups WHERE name = 'zone_neighborhood_flavor'
);

DELETE FROM tag_groups
WHERE name = 'zone_neighborhood_flavor';
