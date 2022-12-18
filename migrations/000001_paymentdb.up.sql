DROP TABLE IF EXISTS account CASCADE;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE TABLE IF NOT EXISTS account 
(
	id UUID PRIMARY KEY,
	first_name VARCHAR(50),
	last_name VARCHAR(50),
	card_number BIGINT,
	card_expiry_month SMALLINT,
	card_expiry_year SMALLINT,
	card_security_code SMALLINT,
	balance serial,
	blocked_money serial,
	statement text[],
	created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payment 
(
	id UUID,
	business_id UUID,
	order_id serial,
	operation VARCHAR(50),
	amount serial,
	status VARCHAR(50),
	currency VARCHAR(50),
	card_number BIGINT,
	card_expiry_month SMALLINT,
	card_expiry_year SMALLINT,
	created_at TIMESTAMP
);