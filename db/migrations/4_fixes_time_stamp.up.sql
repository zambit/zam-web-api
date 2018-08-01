alter table users alter column registered_at type timestamp without time zone using current_date + registered_at;
alter table users alter column created_at type timestamp without time zone using current_date + created_at;
alter table users alter column updated_at type timestamp without time zone using current_date + updated_at;