-- Создаём таблицу metrics
CREATE TABLE metrics (
    id_metric TEXT NOT NULL PRIMARY KEY,
    type_metric TEXT NOT NULL,
    value_metric DOUBLE PRECISION,
    delta NUMERIC
);


-- создание индекса для таблицы videos по полю video_id
CREATE INDEX id_metric ON metrics (id_metric)