-- +migrate Up
alter table users
add column name varchar(100);

-- +migrate Down
alter table users
drop column name;