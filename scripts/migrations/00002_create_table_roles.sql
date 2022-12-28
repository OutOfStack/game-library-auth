-- +migrate Up
create table roles (
	id uuid,
	name varchar(60),
	description varchar(200),
	date_created timestamp,
	date_updated timestamp,
	
	primary key(id)
);

alter table users add constraint fk_role foreign key (role_id) references roles(id);

-- +migrate Down
alter table users
drop constraint fk_role;

drop table roles;