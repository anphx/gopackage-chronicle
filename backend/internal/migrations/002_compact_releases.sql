-- Shrinks the schema to fit the Supabase free plan:
--   * drops pseudo-versions (untagged commits such as v0.0.0-20261114093012-3f2a9c1d8e7b),
--     and the packages left with no releases,
--   * replaces releases.id with PRIMARY KEY (package_id, version), which also makes the
--     separate package_id index redundant,
--   * drops the unused releases.indexed_at and packages.created_at columns,
--   * narrows package ids to INTEGER.
--
-- releases is rebuilt rather than altered so the new table is compact straight away: dropped
-- columns and deleted rows would otherwise keep their space until a VACUUM FULL, which cannot
-- run inside goose's migration transaction.
--
-- The regex is pseudoVersionRE from golang.org/x/mod/module, which the indexer uses
-- (module.IsPseudoVersion) to skip pseudo-versions from now on.

-- +goose Up
CREATE TABLE releases_new (
    released_at TIMESTAMPTZ NOT NULL,
    package_id  INTEGER     NOT NULL,
    version     TEXT        NOT NULL
);

INSERT INTO releases_new (released_at, package_id, version)
SELECT released_at, package_id, version
FROM releases
WHERE version !~ '^v[0-9]+\.(0\.0-|\d+\.\d+-([^+]*\.)?0\.)\d{14}-[A-Za-z0-9]+(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$'
ORDER BY released_at;

DROP TABLE releases;

DELETE FROM packages p
WHERE NOT EXISTS (SELECT 1 FROM releases_new r WHERE r.package_id = p.id);

ALTER TABLE packages DROP COLUMN created_at;
ALTER TABLE packages ALTER COLUMN id TYPE INTEGER;
SELECT setval('packages_id_seq', coalesce((SELECT max(id) FROM packages), 0) + 1, false);
ALTER SEQUENCE packages_id_seq AS INTEGER;

ALTER TABLE releases_new RENAME TO releases;
ALTER TABLE releases ADD CONSTRAINT releases_pkey PRIMARY KEY (package_id, version);
ALTER TABLE releases ADD CONSTRAINT releases_package_id_fkey
    FOREIGN KEY (package_id) REFERENCES packages(id);
CREATE INDEX idx_releases_released_at ON releases(released_at DESC);

-- +goose Down
-- Restores the 001 schema. Deleted pseudo-versions are not restored; reset sync_cursor and
-- re-run the indexer to fetch them again.
CREATE TABLE releases_old (
    id          BIGSERIAL PRIMARY KEY,
    package_id  BIGINT      NOT NULL,
    version     TEXT        NOT NULL,
    released_at TIMESTAMPTZ NOT NULL,
    indexed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(package_id, version)
);

INSERT INTO releases_old (package_id, version, released_at)
SELECT package_id, version, released_at
FROM releases
ORDER BY released_at;

DROP TABLE releases;

ALTER SEQUENCE packages_id_seq AS BIGINT;
ALTER TABLE packages ALTER COLUMN id TYPE BIGINT;
ALTER TABLE packages ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE releases_old RENAME TO releases;
ALTER INDEX releases_old_pkey RENAME TO releases_pkey;
ALTER INDEX releases_old_package_id_version_key RENAME TO releases_package_id_version_key;
ALTER SEQUENCE releases_old_id_seq RENAME TO releases_id_seq;
ALTER TABLE releases ADD CONSTRAINT releases_package_id_fkey
    FOREIGN KEY (package_id) REFERENCES packages(id);
CREATE INDEX idx_releases_package_id ON releases(package_id);
CREATE INDEX idx_releases_released_at ON releases(released_at DESC);
