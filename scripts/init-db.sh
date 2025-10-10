#!/bin/bash
set -e

# This command connects to the default database and creates your two new databases.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE openftv;
    CREATE DATABASE openftv_adl;
EOSQL