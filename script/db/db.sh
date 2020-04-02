#!/bin/bash

# Database Parameters from Terraform ----------------------------------------

cd infra
connection_name=$(terraform output db_connection_name)
dbname=$(terraform output db_name)
user=$(terraform output db_user)
password=$(terraform output db_password)
cd -

# Connect -------------------------------------------------------------------

go run ./app/cmd/db \
    -driver cloudsqlpostgres \
    -conn "host=${connection_name} dbname=${dbname} user=${user} password=${password} sslmode=disable" \
    "$@"
