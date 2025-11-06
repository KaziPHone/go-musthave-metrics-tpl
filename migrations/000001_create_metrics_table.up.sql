-- Создаём схему public_metrics, если её еще нет
-- DO $$
-- BEGIN
--     IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'public_metrics') THEN
--         CREATE SCHEMA public_metrics;
--     END IF;
-- END $$;

-- Создаём таблицу metrics в схеме public_metrics
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    id_metric TEXT NOT NULL,
    type_metric TEXT,
    value_metric DOUBLE PRECISION,
    delta NUMERIC
);