-- +goose Up
CREATE TABLE summits (
    id           bigint PRIMARY KEY,                -- OpenStreetMap node id
    names        jsonb NOT NULL DEFAULT '{}',       -- {"default": "...", "fr": "..."}
    elevation_m  real,                              -- NULL when unknown
    geom         geometry(Point, 4326) NOT NULL
);

CREATE INDEX summits_geom_idx ON summits USING gist (geom);

-- +goose Down
DROP TABLE summits;
