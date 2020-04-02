#!/bin/bash

# Database Parameters from Terraform ----------------------------------------

cd infra
dbname=$(terraform output db_name)
user=$(terraform output db_user)
password=$(terraform output db_password)
cd -

# Open psql -----------------------------------------------------------------

psql "host=127.0.0.1 port=5433 sslmode=disable dbname=${dbname} user=${user} password=${password}"
