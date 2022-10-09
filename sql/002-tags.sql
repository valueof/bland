create table if not exists tags (
  id        integer primary key,
  name      text,
  is_author integer
);

create index if not exists idx_tags_name on tags (name);