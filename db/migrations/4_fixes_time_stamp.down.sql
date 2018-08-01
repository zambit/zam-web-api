alter table users alter column registered_at type time without time zone using registered_at::time without time zone;
alter table users alter column created_at type time without time zone using created_at::time without time zone;
alter table users alter column updated_at type time without time zone using updated_at::time without time zone;