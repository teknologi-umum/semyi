# Semyi

Semyi is an uptime monitoring platform designed to keep track of your server and services' availability and performance. It leverages `roselite` as a relay instance to efficiently pass monitoring data. Key features include:
- Real-time monitoring of server uptime and performance metrics.
- Configurable alerting system to notify you of downtime or performance issues.
- Integration with `roselite` for seamless data relay and processing.

## Usage

To run Semyi, use the pre-built Docker image available at `ghcr.io/teknologi-umum/semyi`. Use the `:edge` tag for builds from the master branch and `:latest` for tagged builds.

```yaml
services:
  semyi:
    image: "ghcr.io/teknologi-umum/semyi:latest"
    ports:
      - 5000:5000
    environment:
      CONFIG_PATH: "/data/config.json"
      DB_PATH": "/data/db.duckdb"
    volumes:
      - "./config.json:/data/config.json"
      - "./db.duckdb:/data/db.duckdb"
    healthcheck:
      test: "curl http://localhost:5000/api/_healthz
      retries: 3
      interval: 15s
      timeout: 10s
```

### Environment Variables

- `CONFIG_PATH`: Path to the configuration file (default: `/data/config.json`)
- `ENVIRONMENT`: Set to `development` or `production` (default: `production`)
- `HOSTNAME`: Hostname for the server (default: `0.0.0.0`)
- `DB_PATH`: Path to the database file (default: `/data/db.duckdb`)
- `STATIC_PATH`: Path to static files (default: `/app/src/dist`)
- `DEFAULT_INTERVAL`: Default monitoring interval in seconds (default: `30`)
- `DEFAULT_TIMEOUT`: Default timeout in seconds (default: `10`)
- `PORT`: Port for the server to listen on (default: `5000`)
- `API_KEY`: API key for authentication (optional)
- `BACKEND_SENTRY_DSN`: Sentry DSN for backend (default to empty string which disables Sentry)
- `BACKEND_SENTRY_SAMPLE_RATE`: Sentry sample rate for errors (default: `1.0`)
- `BACKEND_SENTRY_TRACES_SAMPLE_RATE`: Sentry sample rate for tracing (default: `1.0`)
- `ENABLE_DUMP_FAILURE_RESPONSE`: Enable dumping response data if healthcheck failure occures (default: `false`)

### Configuration Files

Semyi supports configuration files in JSON, YAML, and TOML formats. Below is an example configuration in JSON:

```json
{
  "alerting": {
    "telegram": {
      "enabled": true,
      "url": "https://api.telegram.org",
      "chat_id": "123456789"
    },
    "discord": {
      "enabled": true,
      "webhook_url": "https://discord.com/api/webhooks/.."
    },
    "http": {
      "enabled": true,
      "webhook_url": "https://example.com/webhook"
    },
    "slack": {
      "enabled": true,
      "webhook_url": "https://hooks.slack.com/services/..."
    }
  },
  "monitors": [
    {
      "unique_id": "monitor-1",
      "name": "Example Monitor",
      "description": "",
      "public_url": "https://example.com",
      "type": "http",
      "interval": 30,
      "timeout": 30,
      "http_headers": {},
      "http_method": "GET",
      "http_endpoint": "https://example.com/_healthz",
      "http_expected_status_code": "2xx"
    },
    {
      "unique_id": "monitor-2",
      "name": "Google DNS",
      "description": "",
      "public_url": "https://dns.google",
      "type": "icmp",
      "interval": 30,
      "timeout": 30,
      "hostname": "8.8.4.4",
      "packet_size": 56
    }
  ],
  "retention_period": 120
}
```

### Storage Options

By default, Semyi uses DuckDB as the storage. For large deployments, you can switch to ClickHouse by providing the ClickHouse DSN in the `DB_PATH` environment variable. The DSN format can be found [here](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn).

## License

```
Copyright (C) 2025 Teknologi Umum <opensource@teknologiumum.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

See [LICENSE](./LICENSE)
