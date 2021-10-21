insert into roles (id, name, description, date_created, date_updated)
values ('fa0d517e-c7b5-474b-8586-306bb9536d2c', 'admin', 'Role that has a right to change users roles', now(), null),
       ('dd14c049-67ce-4948-9afd-9338bc3319ff', 'user', 'General user role without special rights', now(), null),
       ('7e6b43da-4135-41d4-8487-2287cf69f0b1', 'publisher', 'Role that has a right to publish games and add them on sale', now(), null),
       ('ce946525-1ef3-46c0-b39b-45d820d90021', 'moderator', 'Role that has a right to add sales', now(), null);

-- admin:gamelibmaster
insert into users (id, username, name, password_hash, role_id, date_created, date_updated)
values ('66c52542-9b0f-4388-8248-6c4ea5fa89f0', 'admin', 'Administrator', '$2a$06$4nOjQZtkUFXGN7.OMJ/ofebK/.GeACFPfrW/S1DJikSvET4jJT3E6', 'fa0d517e-c7b5-474b-8586-306bb9536d2c', now(), null);