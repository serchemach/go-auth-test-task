#!/bin/bash
psql -U $POSTGRES_USER -c 'create database sample;'

# Create the tables and add some mock data
psql -v ON_ERROR_STOP=1 -U $POSTGRES_USER -d sample  <<-EOSQL
     create schema auth_scheme;

     CREATE TABLE auth_scheme.user (
        id uuid primary key,
        email text not null,
        name text not null
     );

     insert into auth_scheme.user (id, email, name) values 
       ('462a75d9-96a4-4ff4-81c8-54b7fd06fbb2', 'user1@mail.ru', 'user1'),
       ('f17be2ee-7dff-47ae-b7d2-23aac555d592', 'user2@mail.ru', 'user2'),
       ('c3463d27-78ef-44c8-b48e-8ef759edbd88', 'user3@mail.ru', 'user3'),
       ('f2621b78-312e-4b4c-b6d7-c21e371d05af', 'user4@mail.ru', 'user4'),
       ('0c62ea49-75d9-4750-8786-77e222dfd728', 'user5@mail.ru', 'user5'),
       ('5f0f0a8c-5741-49d0-bb56-8e42f674881f', 'user6@mail.ru', 'user6');
EOSQL
