create table if not exists bookmarks (
    id          integer primary key,
    url         text,
    title       text,
    shortcut    text,
    description text,
    tags        text,
    created_at  integer,
    updated_at  integer,
    deleted_at  integer,
    read_at     integer
);

create index if not exists idx_bookmarks_shortcut on bookmarks (shortcut);