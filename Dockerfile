FROM node:20.19-bookworm AS frontend
WORKDIR /app
COPY frontend/ .
COPY config.json /config.json
RUN npm i -g pnpm && pnpm install && pnpm run build

FROM golang:1.24-bookworm AS backend
WORKDIR /app
COPY backend/ .
COPY config.json /config.json
RUN go build .

FROM debian:bookworm
ENV ENV=production
ENV STATIC_PATH=/app/src/dist
ENV CONFIG_PATH=/data/config.json
ENV DB_PATH=/data/db.duckdb
ENV HOSTNAME=0.0.0.0
ENV DEFAULT_INTERVAL=30
ENV DEFAULT_TIMEOUT=10
ENV PORT=5000
ENV API_KEY=
WORKDIR /app
RUN mkdir -p /data
COPY LICENSE /app/LICENSE
COPY README.md /app/README.md
COPY --from=backend /app/semyi /app/src/semyi
COPY --from=frontend /app/dist /app/src/dist
EXPOSE ${PORT}
CMD ["/app/src/semyi"]
