create table tags_bookmarks (
    bookmark_id integer not null,
    tag_id      integer not null,

    primary key (bookmark_id, tag_id),
  	foreign key (bookmark_id) references bookmarks (id),
  	foreign key (tag_id) references tags (id)
);