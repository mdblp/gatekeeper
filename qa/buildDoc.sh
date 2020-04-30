#!/bin/bash -eu
# Generate OpenAPI documentation

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
# Update the software version in the swaggerDef.json file through a temporary file
tmp=$(mktemp)
jq --arg v "${version}" '.info.version=$v' swaggerDef.json > ${tmp}
mv ${tmp} swaggerDef.json
# Run the build of the documentation using npm package swagger-jsdoc
swagger-jsdoc -d swaggerDef.json -o doc/${project}-${version}-swagger.json ${apis}
echo "Built swagger doc for ${project}-${version}"
