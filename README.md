# purrfect-api

## Local setup

### Smee.io

To receive Airtable webhook events locally, start a new channel on https://smee.io/, then:

    npm install --global smee-client
    smee -u https://smee.io/2mxhU4Pb2YrNvF8E  # replace with your URL

Keep that URL for later.

### Generate Airtable token

In https://airtable.com/create/tokens create a token with scopes:

 - `data.records:read`
 - `schema.bases:read`
 - `webhook:manage`

and permission `All current and future bases in all current and future workspaces`.

Use the token in the step below.

### Run tests

    go test -v ./...

### Run manually

    AIRTABLE_API_TOKEN=redacted SMEE_URL=redacted go run main.go

### Run via docker

    docker build -t purrfect-api .
    docker run -e AIRTABLE_API_TOKEN=redacted -e SMEE_URL=redacted -p 3000:3000 purrfect-api

## Features

### API endpoints

 - Totals: http://localhost:3000/stats
 - Health: http://localhost:3000/health
 - Webhook: http://localhost:3000/
 - Websocket: http://localhost:3000/ws

### Webhook

Webhooks allow getting the data ASAP and prevent:
- regularly polling,
- having to request the same data.

### Websocket

Allows seeing updates immediately on the UI instead of having to
refresh the page or have a button to fetch data again.

### Possible improvements

- Only fetch updated/added data from webhook event instead of all records every time
- Don't refresh on every Airtable keystroke
- Fix api error when no active webui

