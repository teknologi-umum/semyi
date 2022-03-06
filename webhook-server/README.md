# Webhook server examples

The webhook (if turned on) sends a HTTP POST request to the corresponding URL with the JSON payload of:

```json
{
  "endpoint": "string",
  "status": "string",
  "statusCode": 200,
  "requestDuration": 1000,
  "timestamp": 100000
}
```

Where:
* Endpoint: the URL endpoint for current health check
* Status: whether the check succeed. Possible values are: `success` and `failed`
* StatusCode: HTTP status code for current health check request
* RequestDuration: how long it took to make the health check request
* Timestamp: when was the health check request sent

With additional header of:
- Content-Type: application/json
- User-Agent: Semyi Webhook
