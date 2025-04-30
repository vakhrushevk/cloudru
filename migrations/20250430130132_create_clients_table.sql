-- +goose Up
CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) NOT NULL,
    rate INT NOT NULL, -- пополнение токенов
    capacity INT NOT NULL, -- количество возможных токенов
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE clients;
