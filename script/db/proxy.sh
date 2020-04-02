#!/bin/bash

cd infra
cloud_sql_proxy -instances=$(terraform output db_connection_name)=tcp:5433
