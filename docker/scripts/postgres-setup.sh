#!/bin/sh

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE openftv_vlierdam_manager;
    CREATE DATABASE openftv_vlierdam_adl;

    CREATE DATABASE openftv_rvig_manager;
    CREATE DATABASE openftv_rvig_adl;

    CREATE DATABASE openftv_rdw_manager;
    CREATE DATABASE openftv_rdw_adl;
EOSQL
