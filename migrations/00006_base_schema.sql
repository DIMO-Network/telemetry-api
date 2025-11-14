-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event (
    `subject` String COMMENT 'identifies the entity the event pertains to.',
    `source` String COMMENT 'the entity that identified and submitted the event (oracle).',
    `producer` String COMMENT 'the specific origin of the data used to determine the event (device).',
    `cloud_event_id` String COMMENT 'identifier for the cloudevent.',
    `name` String COMMENT 'name of the event indicated by the oracle transmitting it.',
    `timestamp` DateTime64(6, 'UTC') COMMENT 'time at which the event described occurred, transmitted by oracle.',
    `duration_ns` UInt64 COMMENT 'duration in nanoseconds of the event.',
    `metadata` String COMMENT 'arbitrary JSON metadata provided by the user, containing additional event-related information.',
    `tags` Array(String) COMMENT 'tags for the event.'
) ENGINE = SharedReplacingMergeTree('/clickhouse/tables/{uuid}/{shard}', '{replica}')
ORDER BY (subject, timestamp, name, source)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS signal (
    `token_id` UInt32 COMMENT 'token_id of this device data.',
    `timestamp` DateTime64(6, 'UTC') COMMENT 'timestamp of when this data was collected.',
    `name` LowCardinality(String) COMMENT 'name of the signal collected.',
    `source` String COMMENT 'source of the signal collected.',
    `producer` String COMMENT 'producer of the collected signal.',
    `cloud_event_id` String COMMENT 'Id of the Cloud Event that this signal was extracted from.',
    `value_number` Float64 COMMENT 'float64 value of the signal collected.',
    `value_string` String COMMENT 'string value of the signal collected.',
    `value_location` Tuple(latitude Float64, longitude Float64, hdop Float64) COMMENT 'Location value of the signal collected.'
) ENGINE = SharedReplacingMergeTree('/clickhouse/tables/{uuid}/{shard}', '{replica}')
ORDER BY (token_id, timestamp, name)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS signal;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS event;
-- +goose StatementEnd

