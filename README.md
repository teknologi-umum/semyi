# Semyi

An uptime monitoring platform to monitor your server and services. Best used as a Docker
container, by running:

```
docker build -t semyi:latest .
docker run -d -p 5000:5000 -v ./config.json:/app/config.json semyi:latest
docker image prune
```

Or if you prefer it as a docker-compose:

```yaml
uptime-monitor:
    build: .
    ports:
        - 5000:5000
    volumes:
        - ./config.json:/app/config.json
        - ./db.sqlite3:/app/db.sqlite3
    environment:
        DEFAULT_INTERVAL: 30
        DEFAULT_TIMEOUT: 10
        PORT: 5000
```

There is another convinient way, by us creating a Github Docker package that you can directly pull (via ghcr.io), but that will be coming soon.

## License

```
Semyi
Copyright (C) 2022 Teknologi Umum <opensource@teknologiumum.com>

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
