# Devnull

**Devnull** is a minimal HTTP server that logs request details and responds with a fixed response to all incoming requests.

It can be used as a no-op HTTP receiver, useful for inspecting outgoing payloads.

Optionally, if a target is configured, incoming requests will be forwarded to that target. The target's response will then be relayed back to the original requester. 

You can enable request throttling by setting a wait time with the `-w` flag.

### Args

```
-d, --debug
    Enable debug logging
-o, --output string
    Output file for logs
-p, --port string
    Port to listen on (default "8080")
-s, --status-code int
    Status code used in responses if no target is configured (default 200)
-t, --target string
    Target (URL) to forward requests to
-T, --timeout duration
    Timeout for the HTTP client (default 30s)
-v, --verbose
    Log extra details about the request (headers, request body)
-w, --wait duration
    Minimum wait time before HTTP response
```
