FROM node:10.15.3-alpine

RUN apk --no-cache update && \
    apk --no-cache upgrade

WORKDIR /app

COPY package.json package.json

RUN npm install

USER node

COPY . .

CMD ["npm", "start"]
