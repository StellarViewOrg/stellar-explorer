-- Add the integer muxed ID component to transactions, completing the three-part
-- muxed account pattern already used in operations and token_events:
--   account        = base G-address (always present)
--   account_muxed  = full M-address (only when source is muxed)
--   account_muxed_id = 64-bit integer ID component (only when source is muxed)

ALTER TABLE transactions
    ADD COLUMN account_muxed_id BIGINT;
