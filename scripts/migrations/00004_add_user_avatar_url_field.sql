-- +migrate Up
alter table users
add column avatar_url varchar(120);

-- +migrate Down
alter table users
drop column avatar_url;