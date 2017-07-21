CREATE TABLE IF NOT EXISTS facts (
	event_date DATE, 
	event_time TIMESTAMPTZ, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	item_id VARCHAR(40), 
	description VARCHAR(40), 
	count INT,
	amount FLOAT
);

CREATE INDEX IF NOT EXISTS idx_facts_date ON facts (event_date);
CREATE INDEX IF NOT EXISTS idx_facts_user_time ON facts (user_id, time_id);


CREATE TABLE IF NOT EXISTS purchase (
	event_date DATE, 
	event_time TIMESTAMPTZ, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	product_id VARCHAR(40), 
	currency_type VARCHAR(40), 
	purchase_receipt TEXT,
	price DECIMAL(10,2)
);

CREATE INDEX IF NOT EXISTS idx_purch_date ON purchase (event_date);
CREATE INDEX IF NOT EXISTS idx_purch_user_time ON purchase (user_id, time_id);


CREATE TABLE IF NOT EXISTS player_battle_summary (
	event_date DATE, 
	event_time TIMESTAMPTZ, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	user_battle_id VARCHAR(80), 
	leader_card_id VARCHAR(40), 
	user_level SMALLINT,
	user_rank SMALLINT,
	game_mode VARCHAR(40),
	user_team VARCHAR(40),
	duration_sec FLOAT,
	tower_score SMALLINT,
	end_result VARCHAR(40),
	end_phase VARCHAR(40)
);

CREATE INDEX IF NOT EXISTS idx_pbs_date ON player_battle_summary (event_date);
CREATE INDEX IF NOT EXISTS idx_pbs_user_time ON player_battle_summary (user_id, time_id);

CREATE TABLE IF NOT EXISTS card_battle_summary (
	event_date DATE, 
	event_time TIMESTAMPTZ, 
	event_type VARCHAR(40), 
	user_id VARCHAR(40), 
	time_id BIGINT, 
	user_battle_id VARCHAR(80), 
	card_id VARCHAR(40), 
	card_level SMALLINT,
	mana_cost SMALLINT,
	damage_mana_value FLOAT,
	num_uses SMALLINT,
	character_damage FLOAT,
	character_kills SMALLINT,
	tower_damage FLOAT,
	tower_kills SMALLINT,
	total_unit_time FLOAT,
	total_units SMALLINT
);

CREATE INDEX IF NOT EXISTS idx_cbs_date ON card_battle_summary (event_date);
CREATE INDEX IF NOT EXISTS idx_cbs_user_time ON card_battle_summary (user_id, time_id);


