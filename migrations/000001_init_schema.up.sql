CREATE TABLE IF NOT EXISTS users
(
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    coins BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions
(
    id BIGSERIAL PRIMARY KEY,
    from_user BIGINT NOT NULL REFERENCES users(id),
    to_user BIGINT NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transactions_from_user ON transactions(from_user);
CREATE INDEX IF NOT EXISTS idx_transactions_to_user ON transactions(to_user);

CREATE TABLE IF NOT EXISTS merch_items
(
    id BIGSERIAL PRIMARY KEY,
    item_type VARCHAR(255) NOT NULL UNIQUE,
    price INT NOT NULL CHECK (price > 0)
);

CREATE TABLE IF NOT EXISTS inventory
(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    merch_id BIGINT NOT NULL REFERENCES merch_items(id),
    quantity INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_inventory_user_merch ON inventory(user_id, merch_id);

INSERT INTO merch_items (item_type, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);