#!/bin/bash
set -e

readonly service="$1"
readonly output_dir="$2"
readonly package="$3"


oapi-codegen --output-config -generate types -o "$output_dir/types.gen.go" -package "$package" "openapi/$service.yml" > scripts/config_types.yml
oapi-codegen --output-config -generate gin -o "$output_dir/server.gen.go" -package "$package" "openapi/$service.yml" > scripts/config_gin.yml

oapi-codegen --config scripts/config_types.yml "openapi/$service.yml" 
oapi-codegen --config scripts/config_gin.yml "openapi/$service.yml"

rm -rf scripts/config_types.yml
rm -rf scripts/config_gin.yml
