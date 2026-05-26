-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    number      TEXT PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    status      TEXT NOT NULL DEFAULT 'NEW',
    accrual     BIGINT NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS orders_user_idx   ON orders (user_id);
CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);

-- +goose Down
DROP TABLE IF EXISTS orders;
