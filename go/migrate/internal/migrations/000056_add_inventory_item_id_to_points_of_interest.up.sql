alter table point_of_interest_challenges add column inventory_item_id integer default 0;
update point_of_interest_challenges set inventory_item_id = 0;
