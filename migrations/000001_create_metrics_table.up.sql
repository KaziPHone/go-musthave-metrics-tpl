-- Создаём таблицу metrics
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    id_metric TEXT NOT NULL,
    type_metric TEXT,
    value_metric DOUBLE PRECISION,
    delta NUMERIC
);