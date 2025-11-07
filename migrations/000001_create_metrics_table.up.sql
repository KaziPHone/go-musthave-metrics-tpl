-- Создаём таблицу metrics
CREATE TABLE metrics (
    id_metric TEXT NOT NULL PRIMARY KEY,
    type_metric TEXT NOT NULL,
    value_metric DOUBLE PRECISION,
    delta NUMERIC
);