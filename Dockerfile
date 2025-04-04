FROM node:18.20-bookworm AS frontend
WORKDIR /app
COPY frontend/ .
COPY config.json /config.json
RUN npm install && npm run build

FROM golang:1.24-bookworm AS backend
WORKDIR /app
COPY backend/ .
COPY config.json /config.json
RUN go build .

FROM debian:bookworm
WORKDIR /app
COPY --from=backend /app/semyi /app/src/semyi
COPY --from=frontend /app/dist /app/src/dist
ENV STATIC_PATH=/app/src/dist
ENV ENV=production
EXPOSE ${PORT}
CMD ["/app/src/semyi"]
