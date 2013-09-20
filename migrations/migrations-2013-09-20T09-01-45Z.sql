CREATE TABLE migration_info (
    id serial NOT NULL PRIMARY KEY,
    created timestamp with time zone NOT NULL,
    content text NOT NULL 
);
