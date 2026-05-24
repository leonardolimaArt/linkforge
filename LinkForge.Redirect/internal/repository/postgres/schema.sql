CREATE TABLE short_links (
    id  uuid    PRIMARY KEY,
    original_url    varchar(2048)    NOT NULL,
    short_code  varchar(32) NOT NULL UNIQUE,
    created_at  timestamp with time zone    NOT NULL
)