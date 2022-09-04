create table bookmarks (
    id          mediumint not null auto_increment,
    url         text,
    title       varchar(260),
    shortcut    varchar(260),    
    description text,
    createdAt   timestamp,
    updatedAt   timestamp,
    deletedAt   timestamp,
    readAt      timestamp,
    
    primary key (id)
);

create table tags (
    id   mediumint not null auto_increment,
    name varchar(50),

    primary key (id)
);

create table authors (
    id   mediumint not null auto_increment,
    name varchar(100),

    primary key (id)
);

-- junction tables

create table bookmarks_tags (
    bookmark_id mediumint not null,
    tag_id      mediumint not null,

    primary key (bookmark_id, tag_id),
    unique key  (tag_id, bookmark_id),
    constraint  fk_bookmark foreign key (bookmark_id) references bookmarks (id),
    constraint  fk_tag foreign key (tag_id) references tags (id)
);

create table bookmarks_authors (
    bookmark_id mediumint not null,
    author_id   mediumint not null,

    primary key (bookmark_id, author_id),
    unique key  (author_id, bookmark_id),
    constraint  fk_bookmark2 foreign key (bookmark_id) references bookmarks (id),
    constraint  fk_author foreign key (author_id) references authors (id)
);

-- indices

create index idx_bookmarks_shortcut on bookmarks (shortcut);