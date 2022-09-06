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
    bookmark_id mediumint

    primary key (id)
    constraint fk_tags_bookmarks foreign key (bookmark_id) references bookmarks (id)
);
