create table bookmarks (
    id          mediumint not null auto_increment,
    url         text,
    title       varchar(260),
    shortcut    varchar(260),    
    description text,
    tags        text,
    createdAt   timestamp,
    updatedAt   timestamp,
    deletedAt   timestamp,
    readAt      timestamp,
    
    primary key (id)
    index idx_bookmarks_shortcut (shortcut)
);

create table tags (
    id          mediumint not null auto_increment,
    name        varchar(50),

    primary key (id)
    index idx_tags_name (name)
);

create table tags_bookmarks (
    bookmark_id mediumint not null
    tag_id      mediumint not null

    primary key (bookmark_id, tag_id)
    constraint fk_tags_bookmarks_bookmark_id foreign key (bookmark_id) references bookmarks (id)
    constraint fk_tags_bookmarks_tag_id foreign key (tag_id) references tags (id)
);
