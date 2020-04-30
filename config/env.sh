export DISCOVERY_HOST=localhost:8000
export GATEKEEPER_SECRET="This secret is used to encrypt the groupId stored in the DB for gatekeeper"
export MONGO_CONNECTION_STRING="mongodb://guardian:password@localhost:27017/gatekeeper?authSource=admin&ssl=false"
export NODE_ENV=development
export PORT=9123
export PUBLISH_HOST=hakken
export SERVICE_NAME=gatekeeper
export USER_API_SERVICE="{\"type\":\"static\", \"hosts\":[{\"protocol\":\"http\", \"host\":\"localhost:9107\"}]}"
export SERVER_SECRET='This needs to be the same secret everywhere. YaHut75NsK1f9UKUXuWqxNN0RUwHFBCy'