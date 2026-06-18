-- Add muxed account columns to operations, mirroring how transactions stores
-- account vs account_muxed. Following the same pattern as Stellar Expert:
--   source_account        = base G-address (always present)
--   source_account_muxed  = full M-address (only when source is muxed)
--   source_muxed_id       = 64-bit integer ID component (only when source is muxed)
--   destination_muxed     = full M-address (only when destination is muxed)
--   destination_muxed_id  = 64-bit integer ID component (only when destination is muxed)

ALTER TABLE operations
    ADD COLUMN source_account_muxed VARCHAR(69),
    ADD COLUMN source_muxed_id      BIGINT,
    ADD COLUMN destination_muxed    VARCHAR(69),
    ADD COLUMN destination_muxed_id BIGINT;
