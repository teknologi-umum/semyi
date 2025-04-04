## Problem Statement

Teknologi Umum community uses Uptime Kuma at the moment, it works, but sometimes it doesn't. The alerting and many things to monitor works great, but these are a few things that are missing from the perfect status page in our case:

* You can't put more description about what this service is about
* You can't see the incident history
* Everything is either "up" or "down", you can't put a "degraded performance" or "under maintenance" status.
* Users can't see longer uptime status other than the last 48 minutes (since the interval is 1 minute).
* There are no public chart for service latency.

Basically, the "perfect status page" in our case is something close to Atlassian Statuspage or Betterstack Status or current Uptime Kuma with some improvements. 

A few good status pages that's publicly exists out there:

* https://status.digitalocean.com/ (uses Atlassian Statuspage)
* https://www.intercomstatus.com/ (uses Atlassian Statuspage)
* https://slack-status.com/ (I like how they provide RSS feed, and how simple the UI is. You can see the historical status in their detail page https://slack-status.com/calendar)
* https://status.railway.app/ (uses Instatus, very good UI)
* https://status.cronitor.io/ (I like the latency graph)
* https://meta.hyperping.app/

## Solution Sketch

We're trying to continue teknologi-umum/semyi and abandon (probably will delete this repository) teknologi-umum/ohana. We'll still be using regular Go as backend and SPA frontend server. The container should be collapsed into one, rather than creating two separate Docker containers for backend and frontend.

Features that should be implemented:

* Configurable intervals, with the lowest being 30 seconds. Lower than 30 seconds is... kinda pointless for a status page.
* Unlimited number of things to be monitored.
* Should implement similar Uptime Kuma API, so teknologi-umum/roselite wouldn't be a waste.
* Should support these kinds of monitors:
    * HTTP monitoring
    * TLS/SSL monitoring -- certificate expiry notification, certificate headers checking (issuer, subject, common name, etc)
    * Ping/ICMP monitoring
* Should support these providers for alerts:
    * Webhook/HTTP
    * Email
    * Telegram (this is priority)
    * Discord (least concern)
* Should be able to submit incident via HTTP API and through config file.