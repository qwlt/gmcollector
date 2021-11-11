#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER"  <<-EOSQL
    CREATE TABLE test_table(uid UUID PRIMARY KEY, time TIMESTAMPTZ, val real, meta JSONB);
    CREATE INDEX id_timestamp_index ON test_table(uid,time); 
EOSQL