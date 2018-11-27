create table personal_data_statuses (
  id   serial primary key,
  name varchar (16) not null unique
);

insert into personal_data_statuses (name) values ('pending'), ('verified'), ('declined');

create type user_gender as enum ('male', 'female', 'undefined');

create table personal_data (
  id         serial primary key,
  user_id    int references users(id) not null unique,
  status_id  int references personal_data_statuses(id) not null,
  email      varchar(60) not null,
  first_name varchar(60) not null,
  last_name  varchar(60) not null,
  birth_date date not null,
  sex        user_gender not null,
  country    varchar(60) not null,
  address    jsonb not null
);

create index on personal_data (user_id);