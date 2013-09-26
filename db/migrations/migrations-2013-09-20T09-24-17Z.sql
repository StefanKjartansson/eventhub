CREATE TABLE event (
    id serial NOT NULL PRIMARY KEY,
    key varchar(256) NOT NULL,
    key_params json NOT NULL,
    created timestamp with time zone NOT NULL,
    updated timestamp with time zone NOT NULL,
    payload json NOT NULL,
    description text NOT NULL,
    importance integer NOT NULL,
    origin varchar(64) NOT NULL,
    entities text[],
    other_references text[],
    actors text[],
    tags text[],
    UNIQUE (key, created)
);

CREATE INDEX event_key ON event (key);
CREATE INDEX event_key_like ON event (key varchar_pattern_ops);
CREATE INDEX event_created ON event (created);
CREATE INDEX event_updated ON event (updated);
CREATE INDEX event_importance ON event (importance);
CREATE INDEX event_origin ON event (origin);
CREATE INDEX event_origin_like ON event (origin varchar_pattern_ops);
CREATE INDEX event_entities ON event USING GIN (entities);
CREATE INDEX event_other_references ON event USING GIN (other_references);
CREATE INDEX event_actors ON event USING GIN (actors);
CREATE INDEX event_tags ON event USING GIN (tags);
