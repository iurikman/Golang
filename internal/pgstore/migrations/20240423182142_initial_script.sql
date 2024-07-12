-- +migrate Up

CREATE TABLE companies (
    id          uuid primary key,
    name        varchar not null unique
);

CREATE TABLE users (
    id          uuid primary key,
    company     uuid references companies (id),
    role        varchar,
    name        varchar,
    surname     varchar,
    phone       varchar,
    email       varchar not null unique,
    user_type   varchar not null,
    deleted     bool not null default false,
    created_at  timestamp with time zone default now(),
    updated_at  timestamp with time zone default now()
);

-- +migrate Down

DROP TABLE users, companies CASCADE;