FROM node:6.10.3-alpine

# Common ENV
ENV API_SECRET="This is a local API secret for everyone. BsscSHqSHiwrBMJsEGqbvXiuIUPAjQXU" \
    SERVER_SECRET="This needs to be the same secret everywhere. YaHut75NsK1f9UKUXuWqxNN0RUwHFBCy" \
    LONGTERM_KEY="abcdefghijklmnopqrstuvwxyz" \
    DISCOVERY_HOST=hakken:8000 \
    PUBLISH_HOST=hakken \
    METRICS_SERVICE="{ \"type\": \"static\", \"hosts\": [{ \"protocol\": \"http\", \"host\": \"highwater:9191\" }] }" \
    USER_API_SERVICE="{ \"type\": \"static\", \"hosts\": [{ \"protocol\": \"http\", \"host\": \"shoreline:9107\" }] }" \
    SEAGULL_SERVICE="{ \"type\": \"static\", \"hosts\": [{ \"protocol\": \"http\", \"host\": \"seagull:9120\" }] }" \
    GATEKEEPER_SERVICE="{ \"type\": \"static\", \"hosts\": [{ \"protocol\": \"http\", \"host\": \"gatekeeper:9123\" }] }" \
# Container specific ENV
    PORT=9123 \
    SERVICE_NAME=gatekeeper-local \
    GATEKEEPER_SECRET="This secret is used to encrypt the groupId stored in the DB for gatekeeper" \
    MONGO_CONNECTION_STRING="mongodb://mongo/gatekeeper"

WORKDIR /app

COPY package.json /app/package.json
RUN apk --no-cache add git \
 && yarn install \
 && apk del git
COPY . /app

VOLUME /app
USER node

CMD ["npm", "start"]
