-- Records when the indexer first inserted each package (the workflow run that saw it first),
-- as opposed to releases.released_at, which is when the module version was published.
--
-- 002 dropped packages.created_at to save space; this brings it back.
--
-- Packages that already exist get NULL: their real insert time was never kept, and backfilling
-- them with this migration's NOW() would make every one of them look newly added. The column
-- is added without a default first so existing rows stay NULL, then the default is set for
-- rows inserted from now on. The indexer's insert (PackageRepository.UpsertBatch) does not
-- list the column, so it picks up the default, and ON CONFLICT DO NOTHING leaves the value
-- untouched when a package is seen again.

-- +goose Up
ALTER TABLE packages ADD COLUMN created_at TIMESTAMPTZ;
ALTER TABLE packages ALTER COLUMN created_at SET DEFAULT NOW();

-- +goose Down
ALTER TABLE packages DROP COLUMN created_at;
