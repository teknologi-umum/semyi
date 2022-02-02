FROM node:16.3.2-bullseye AS frontend
WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build

FROM golang:1.17.6-bullseye AS backend
WORKDIR /app
COPY backend/ .
RUN go build .

FROM bullseye
WORKDIR /app
COPY --from=backend /app/semya .
COPY --from=frontend /app/dist .
ENV STATIC_PATH=/app/dist
CMD ["/app/semya"]