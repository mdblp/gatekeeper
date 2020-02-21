#!/bin/bash -eu

for i in "$@"
do
case $i in
    --version=*)
    version="${i#*=}"
    shift 
    ;;
    --apis=*)
    apis="${i#*=}"
    shift 
    ;;
    --project=*)
    project="${i#*=}"
    shift 
    ;;
esac
done

mkdir -p doc/
tmp=$(mktemp)
jq --arg v "${version}" '.info.version=$v' swaggerDef.json > ${tmp}
mv ${tmp} swaggerDef.json
swagger-jsdoc -d swaggerDef.json -o doc/${project}-${version}-swagger.json ${apis}
