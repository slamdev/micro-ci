FROM node:alpine AS build

WORKDIR /opt

COPY src src
COPY package*.json ./

RUN npm install && npm run build

FROM nginx:alpine AS run

COPY deploy/nginx.conf /etc/nginx/nginx.conf

WORKDIR /usr/share/nginx/html

COPY --from=build /opt/build/* ./
