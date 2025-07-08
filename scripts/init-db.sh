#!/bin/bash
set -e

# This command connects to the default database and creates your two new databases.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE openftv_pip;
    CREATE DATABASE openftv_pap;
EOSQL

# This command connects to the 'openftv_pip' database and creates the table.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "openftv_pip" <<-EOSQL
    CREATE TABLE kv
    (
        key   VARCHAR(200) PRIMARY KEY NOT NULL,
        index BIGINT NOT NULL,
        value JSONB not null
    );
EOSQL

# This command connects to the 'openftv_pap' database and creates the table.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "openftv_pap" <<-EOSQL
    CREATE TABLE kv
    (
        key   VARCHAR(200) PRIMARY KEY NOT NULL,
        index BIGINT NOT NULL,
        value JSONB not null
    );
EOSQL