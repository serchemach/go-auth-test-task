#!/bin/bash
psql -U $POSTGRES_USER -c 'create database sample;'

psql -v ON_ERROR_STOP=1 -U $POSTGRES_USER -d sample  <<-EOSQL
     create schema sample_schema;
     create table todos (
        id serial primary key
     );
EOSQL
