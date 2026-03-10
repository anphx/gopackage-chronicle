-- +goose Up
CREATE TABLE packages (
    id         BIGSERIAL PRIMARY KEY,
    path       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE releases (
    id          BIGSERIAL PRIMARY KEY,
    package_id  BIGINT NOT NULL REFERENCES packages(id),
    version     TEXT NOT NULL,
    released_at TIMESTAMPTZ NOT NULL,
    indexed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(package_id, version)
);

CREATE TABLE sync_cursor (
    id          INT PRIMARY KEY DEFAULT 1,
    last_synced TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_releases_package_id ON releases(package_id);
CREATE INDEX idx_releases_released_at ON releases(released_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_releases_released_at;
DROP INDEX IF EXISTS idx_releases_package_id;
DROP TABLE IF EXISTS sync_cursor;
DROP TABLE IF EXISTS releases;
DROP TABLE IF EXISTS packages;
