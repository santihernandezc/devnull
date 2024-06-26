# Devnull

It does literally nothing but:

- Listening for requests and responding with a `200 OK` status code
- Logging details about the request

If set to forward requests, it will return the response from the target.

### Args

```bash
-o string
    Output file
-port string
    Port to listen on (default "8080")
-target string
    URL to forward requests to
-v	Enable verbose logging
```
