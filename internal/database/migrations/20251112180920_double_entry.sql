-- +goose Up
-- +goose StatementBegin

CREATE TABLE users(
  id SERIAL PRIMARY KEY,
  name VARCHAR(256) NOT NULL,
  email VARCHAR (256) UNIQUE NOT NULL,
  hashed_password TEXT NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE accounts(
  id SERIAL PRIMARY KEY,
  name VARCHAR(256) NOT NULL,

  -- System accounts will have NULL as their owner_user_id, so set it only for user wallets
  owner_user_id integer REFERENCES users(id)
);

INSERT INTO accounts (name, owner_user_id) VALUES ('Liability', NULL);

CREATE OR REPLACE FUNCTION move_to_sunday(d date)
RETURNS date AS $$
BEGIN
    IF EXTRACT(DOW FROM d) IN (5, 6) THEN
        RETURN d + ((7 - EXTRACT(DOW FROM d))::int);
    ELSE
        RETURN d;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE journals(
  id SERIAL PRIMARY KEY,
  tx_date date NOT NULL DEFAULT CURRENT_DATE,
  tx_value_date date NOT NULL DEFAULT move_to_sunday(CURRENT_DATE),
  created_at timestamptz NOT NULL DEFAULT now(),
	description VARCHAR(1024) NOT NULL
);




-- CREATE TABLE journal_lines(
--   id serial PRIMARY KEY,
--   journal_id integer NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
--   account_id integer NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
--   amount numeric(20,2) NOT NULL CHECK (amount >= 0),
--   side smallint NOT NULL CHECK (side IN (1, -1)),     -- 1 = debit, -1 = credit
--   line_description text
-- );


CREATE TABLE entries(
	id SERIAL PRIMARY KEY,
	amount NUMERIC(20, 2) NOT NULL CHECK (amount > 0.0),
	credit INTEGER NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
	debit INTEGER NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  journal_id INTEGER NOT NULL REFERENCES journals(id) ON DELETE RESTRICT
);

CREATE INDEX entries_journal_idx ON entries(journal_id);
CREATE INDEX entries_credit_idx ON entries(credit);
CREATE INDEX entries_debit_idx ON entries(debit);

-- INSERT INTO journals (tx_date, tx_value_date, created_at, description) VALUES (CURRENT_DATE, move_to_sunday(CURRENT_DATE), NOW(), 'Transfer funds from liability account to wallet 2');
-- INSERT INTO entries (amount, credit, debit, journal_id) VALUES (500.00, 2, 1, 1);

CREATE TABLE branches(
  id SERIAL PRIMARY KEY,
  name VARCHAR(256) NOT NULL
);

INSERT INTO branches (name) VALUES ('Library of faculty of science');
INSERT INTO branches (name) VALUES ('Library of faculty of enginnering');
INSERT INTO branches (name) VALUES ('Library of faculty of medic');
INSERT INTO branches (name) VALUES ('Library of faculty of science');


CREATE TABLE book_reservation(
  id SERIAL PRIMARY KEY,
  account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  branch_id INTEGER NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  reserve_time TIMESTAMP NOT NULL DEFAULT NOW(),
  return_time TIMESTAMP NOT NULL,
  journal_id INTEGER NOT NULL REFERENCES journals(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()  
);


CREATE VIEW account_ledgers AS
SELECT
    e.id AS entry_id,
    e.journal_id,
    j.tx_date,
    j.description AS journal_description,
    e.credit AS account_id,
    e.amount AS amount,
    'credit' AS entry_type
FROM entries e
JOIN journals j ON j.id = e.journal_id
UNION ALL
SELECT
    e.id AS entry_id,
    e.journal_id,
    j.tx_date,
    j.description AS journal_description,
    e.debit AS account_id,
    (0 - e.amount) AS amount,
    'debit' AS entry_type
FROM entries e
JOIN journals j ON j.id = e.journal_id;


CREATE MATERIALIZED VIEW account_balances AS
SELECT
    a.id,
    COALESCE(SUM(al.amount), 0.0) AS balance
FROM accounts a
LEFT JOIN account_ledgers al
    ON a.id = al.account_id
GROUP BY a.id;

CREATE UNIQUE INDEX ON account_balances(id);

CREATE FUNCTION update_balances() RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW account_balances;
    RETURN NULL;
END
$$ LANGUAGE plpgsql;  


CREATE TRIGGER trigger_refresh_account_balances
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON entries
FOR EACH STATEMENT
EXECUTE FUNCTION update_balances();


CREATE OR REPLACE FUNCTION create_wallet_account_for_user()
RETURNS trigger AS $$
BEGIN
    INSERT INTO accounts (name, owner_user_id)
    VALUES (
        'Wallet_' || NEW.id,
        NEW.id
    );

    REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_wallet
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_wallet_account_for_user();

-- CREATE OR REPLACE FUNCTION protect_posted_journals()
-- RETURNS trigger AS $$
-- BEGIN
--     IF OLD.is_posted = TRUE THEN
--         RAISE EXCEPTION 'Cannot modify or delete a posted journal (id=%)', OLD.id;
--     END IF;
--
--     RETURN OLD;
-- END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE OR REPLACE FUNCTION post_journal(p_journal_id INT)
-- RETURNS void AS $$
-- DECLARE
--     entry_count INT;
-- BEGIN
--     -- Check that the journal has entries
--     SELECT COUNT(*) INTO entry_count
--     FROM entries
--     WHERE journal_id = p_journal_id;
--
--     IF entry_count = 0 THEN
--         RAISE EXCEPTION 'Cannot post journal %: no entries found.', p_journal_id;
--     END IF;
--
--     -- Optional: check no entry debits the same account it credits
--     IF EXISTS (
--         SELECT 1 FROM entries
--         WHERE journal_id = p_journal_id
--           AND debit = credit
--     ) THEN
--         RAISE EXCEPTION 'Cannot post journal %: debit and credit accounts cannot be the same.', p_journal_id;
--     END IF;
--
--     -- Mark the journal as posted (final)
--     UPDATE journals
--     SET is_posted = TRUE
--     WHERE id = p_journal_id;
--
-- END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER trigger_protect_journals✔ DO refresh on:
-- INSERT into entries
--
-- UPDATE in entries
--
-- DELETE in entries
--
-- TRUNCATE entries
-- BEFORE UPDATE OR DELETE ON journals
-- FOR EACH ROW
-- EXECUTE FUNCTION protect_posted_journals();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

-- 1. DROP TRIGGERS
DROP TRIGGER IF EXISTS trigger_refresh_account_balances ON entries;
DROP TRIGGER IF EXISTS trigger_create_wallet ON users;

-- 2. DROP FUNCTIONS
DROP FUNCTION IF EXISTS update_balances() CASCADE;
DROP FUNCTION IF EXISTS create_wallet_account_for_user() CASCADE;

-- 3. DROP MATERIALIZED VIEW + INDEX
DROP INDEX IF EXISTS account_balances_id_idx;
DROP MATERIALIZED VIEW IF EXISTS account_balances;

-- 4. DROP VIEW
DROP VIEW IF EXISTS account_ledgers;

-- 5. DROP indexes on entries
DROP INDEX IF EXISTS entries_journal_idx;
DROP INDEX IF EXISTS entries_credit_idx;
DROP INDEX IF EXISTS entries_debit_idx;

-- 6. DROP TABLES (reverse dependency order)
DROP TABLE IF EXISTS receipts;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS journals;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
