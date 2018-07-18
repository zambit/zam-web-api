alter table users
  add column created_at time without time zone,
  add column updated_at time without time zone;

update users set
  created_at = now(),
  updated_at = now();

insert into user_statuses (name) values ('verified');
insert into user_statuses (name) values ('created');
