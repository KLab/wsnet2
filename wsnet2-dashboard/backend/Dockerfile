FROM node:lts-alpine3.11
WORKDIR /app
COPY ./ /app
RUN apk add --no-cache --upgrade bash
RUN npm install
CMD ["bash", "build-and-run.sh"]
