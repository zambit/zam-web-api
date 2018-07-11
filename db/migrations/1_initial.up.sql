create table user_statuses (
  id serial primary key,
  name varchar(63)
);
create table users (
  id serial primary key,
  phone varchar(255),
  status_id integer,
  password text,
  referrer_id integer,
  registered_at time without time zone,
  constraint users_referrer_id_fk foreign key (referrer_id) references users(id)
  on delete set null,
  constraint user_statuses_id_fk foreign key (status_id) references user_statuses(id)
);

insert into user_statuses (name) values ('pending');
insert into user_statuses (name) values ('active');