
## Must Have Project Capabilities

1. Make healthcheck requests to target (similar to pull-based app monitoring/observability)
2. Retrieve healthcheck results from target (similar to [push-based app monitoring/observability](https://notes.nicolevanderhoeven.com/Push-based+monitoring))
3. Send alerts to specific platform (pre-configured) if passed a certain healthcheck threshold
4. Web UI for status monitoring
5. Relay mode for passing the healthcheck result to upstream instance (similar to [Prometheus' Remote Write](https://prometheus.io/docs/specs/prw/remote_write_spec/)).

Old monograph link: https://monogr.ph/664d42567e3a71c23dfea211 -- I think this link is still relevant, but the above points are just TLDR to this.

## Nice To Have Project Capabilities

1. Dynamically configuring healthcheck targets and/or incident alerting
2. Uptime Kuma API compatibility for sending uptime checks (refer to this specific file + commit https://github.com/teknologi-umum/bot/blob/7382c332521018a51ae33f26bb068be65dadd0df/src/uptime.js) (found something here https://www.postman.com/gabrielfnlima/uptime-kuma-api-collection/collection/vadwal6/uptime-kuma-rest-api)
3. TLS certificate expiration notice for pull-based (...probably?)
4. ...more?

## Things to do

We need to check if these already implemented and already works.

### Healthcheck
- [ ] Able to configure sites, dynamically (through web UI) or statically (through config files) -- I think the last time we implemented this was through config files, since the UI would be a read only. But we'll see if we can change this to be dynamically.
- [x] Able to poll to target sites (meaning we make a HTTP request [or any other protocol] to that endpoint)
- [x] Able to save healthcheck data to some storage backends. (see https://github.com/teknologi-umum/semyi/issues/26, https://github.com/teknologi-umum/semyi/issues/23)
- [x] Able to retrieve healthcheck result remotely (similar to how push-mechanism works in app monitoring/observability)
- [ ] HTTP API endpoint for list healthcheck results (for web UI)

### Alerting
- [ ] Able to send an alert to some platform (see https://github.com/teknologi-umum/semyi/issues/7, https://github.com/teknologi-umum/semyi/issues/8, https://github.com/teknologi-umum/semyi/issues/27)
- [ ] HTTP API endpoint for list incidents (sent alert -- reverse of https://github.com/teknologi-umum/semyi/issues/29) (for web UI)
- [ ] HTTP API endpoint for creating an incident (should be implemented already? see https://github.com/teknologi-umum/semyi/issues/29)
- [ ] HTTP API endpoint for modifying a certain incident state, or giving an update regarding the incident

### Relay
- [ ] Relay mode that is able to do both pull and push healthcheck monitoring
- [ ] Submit the data into upstream instance -- a thought: what if there are so many relay instances? should we keep a counter?

