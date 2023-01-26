# build stage
FROM node:lts-alpine3.11 as build-stage
WORKDIR /app
COPY ./ /app
RUN npm install
RUN npm run build

# production stage
FROM nginx:stable-alpine as production-stage
COPY --from=build-stage /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]