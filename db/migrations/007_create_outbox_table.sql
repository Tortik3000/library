-- +goose Up
CREATE TYPE outbox_status as ENUM ('CREATED', 'IN_PROGRESS', 'SUCCESS');

CREATE TABLE outbox
(
    idempotency_key TEXT PRIMARY KEY,
    data            JSONB                   NOT NULL,
    status          outbox_status           NOT NULL,
    kind            INT                     NOT NULL,
    created_at      TIMESTAMP DEFAULT now() NOT NULL,
    updated_at      TIMESTAMP DEFAULT now() NOT NULL,
    trace_id        TEXT
);

-- +goose StatementBegin
CREATE
OR REPLACE FUNCTION update_outbox_timestamp() RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at
= now();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;
-- +goose StatementEnd


CREATE
OR REPLACE TRIGGER trigger_update_outbox_timestamp
    BEFORE
UPDATE
    ON outbox
    FOR EACH ROW
    EXECUTE FUNCTION update_outbox_timestamp();


-- +goose Down
DROP TABLE outbox;