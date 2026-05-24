-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    price INTEGER NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    start_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP DEFAULT NULL
);

-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS subscriptions;
