-- +migrate Up
create table users (
	id 				uuid,
	username 		varchar(32) UNIQUE,
	password_hash 	text,
	date_created 	timestamp,
	date_updated 	timestamp,
	
	primary key (id)
);

-- +migrate Down
drop table users;