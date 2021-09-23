-- +migrate Up
create table users (
	id 				uuid,
	username 		varchar(32) not null unique,
	password_hash 	text,
	role_id 		uuid,
	date_created 	timestamp,
	date_updated 	timestamp,
	
	primary key (id)
);

-- +migrate Down
drop table users;