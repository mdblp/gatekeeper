FROM node:10.15.3-alpine

WORKDIR /app

COPY package.json package.json

RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add make python g++ && \
    npm install && \
    apk del make python g++

USER node

COPY . .

CMD ["npm", "start"]
