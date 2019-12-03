FROM node:10.15.3-alpine

ARG npm_token
ENV nexus_token=$npm_token

WORKDIR /app

COPY package.json package.json
COPY .npmrc .npmrc

RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add make python g++ && \
    npm install && \
    apk del make python g++

USER node

COPY . .

CMD ["npm", "start"]
