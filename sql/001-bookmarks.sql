create table if not exists bookmarks (
    id          integer primary key,
    url         text,
    title       text,
    shortcut    text,
    description text,
    tags        text,
    createdAt   integer,
    updatedAt   integer,
    deletedAt   integer,
    readAt      integer
);

create index if not exists idx_bookmarks_shortcut on bookmarks (shortcut);