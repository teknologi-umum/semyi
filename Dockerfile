FROM node:16.14.0-bullseye AS frontend
WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build

FROM golang:1.17.8-bullseye AS backend
WORKDIR /app
COPY backend/ .
RUN go build .

FROM debian:bullseye
RUN apt-get update && apt-get upgrade -y && apt-get install -y sqlite3
WORKDIR /app
COPY config.json .
COPY db.sqlite3 .
COPY --from=backend /app/semyi /app/src/semyi
COPY --from=frontend /app/dist /app/src/dist
ENV STATIC_PATH=/app/src/dist
ENV ENV=production
EXPOSE ${PORT}
CMD ["/app/semyi"]
