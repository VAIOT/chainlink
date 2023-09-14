-- +goose Up
CREATE TABLE channel_definitions (
    addr bytea PRIMARY KEY CHECK (octet_length(addr) = 20),
    evm_chain_id NUMERIC(78) NOT NULL,
    definitions JSONB NOT NULL,
    block_num BIGINT NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);

CREATE UNIQUE INDEX idx_channel_definitions_evm_chain_id_addr ON channel_definitions (evm_chain_id, addr);

-- +goose Down
DROP TABLE channel_definitions;
