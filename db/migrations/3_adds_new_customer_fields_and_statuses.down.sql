alter table users drop column created_at;
alter table users drop column updated_at;

update users set status_id = (
  select id from user_statuses where name = 'pending'
) where status_id = (
  select id from  user_statuses where name = 'created'
);

update users set status_id = (
  select id from user_statuses where name = 'pending'
) where status_id = (
  select id from  user_statuses where name = 'verified'
);

delete from user_statuses where name = 'verified';
delete from user_statuses where name = 'created';
