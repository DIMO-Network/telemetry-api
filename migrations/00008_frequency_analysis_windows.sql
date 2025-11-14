-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS signal_window_aggregates (
    token_id UInt32,
    window_start DateTime64(6, 'UTC'),
    window_size_seconds UInt16,
    signal_count UInt32,
    distinct_signal_count UInt16
)
ENGINE = ReplacingMergeTree()
ORDER BY (token_id, window_start, window_size_seconds)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS signal_window_aggregates_mv
TO signal_window_aggregates
AS
SELECT
    token_id,
    toStartOfInterval(timestamp, INTERVAL 1 minute) AS window_start,
    60 AS window_size_seconds,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal
GROUP BY token_id, window_start;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS signal_window_aggregates_mv;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS signal_window_aggregates;
-- +goose StatementEnd
