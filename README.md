# iRaiser RSS Webservice

Microservice utilisé pour transformé les informations de l'API iRaiser en flux RSS.

```bash
# Build
GOOS=linux GOARCH=amd64 go build -o iraiser-rss-ws

# Run (not in production)
./iraiser-rss-ws -listen 0.0.0.0 -port 9191
```
