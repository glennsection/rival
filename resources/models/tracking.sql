CREATE TABLE facts (
	event_date DATE, 
	event_time TIMESTAMP, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	item_id VARCHAR(40), 
	description VARCHAR(40), 
	count INT,
	amount FLOAT,
	INDEX date_index (event_date),
	INDEX user_time (user_id, time_id)
);

CREATE TABLE purchase (
	event_date DATE, 
	event_time TIMESTAMP, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	product_id VARCHAR(40), 
	currency_type VARCHAR(40), 
	purchase_receipt TEXT,
	price DECIMAL(7,2),
	INDEX date_index (event_date),
	INDEX user_time (user_id, time_id)
);

CREATE TABLE playerBattleSummary (
	event_date DATE, 
	event_time TIMESTAMP, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	user_battle_id VARCHAR(80), 
	leader_card_id VARCHAR(40), 
	user_level TINYINT,
	user_rank TINYINT,
	game_mode VARCHAR(40),
	user_team VARCHAR(40),
	duration_sec FLOAT,
	tower_score TINYINT,
	end_result VARCHAR(40),
	end_phase VARCHAR(40),
	INDEX date_index (event_date),
	INDEX user_time (user_id, time_id)
);

CREATE TABLE cardBattleSummary (
	event_date DATE, 
	event_time TIMESTAMP, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	user_battle_id VARCHAR(80), 
	card_id VARCHAR(40), 
	card_level TINYINT,
	mana_cost TINYINT,
	damage_mana_value FLOAT,
	num_uses TINYINT,
	character_damage FLOAT,
	character_kills SMALLINT,
	tower_damage FLOAT,
	tower_kills TINYINT,
	total_unit_time FLOAT,
	total_units SMALLINT,
	INDEX date_index (event_date),
	INDEX user_time (user_id, time_id)
);

