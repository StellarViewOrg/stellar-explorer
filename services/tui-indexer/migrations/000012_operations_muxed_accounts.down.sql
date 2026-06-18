ALTER TABLE operations
    DROP COLUMN IF EXISTS source_account_muxed,
    DROP COLUMN IF EXISTS source_muxed_id,
    DROP COLUMN IF EXISTS destination_muxed,
    DROP COLUMN IF EXISTS destination_muxed_id;
