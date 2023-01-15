# purrfect

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

### Run app

    AIRTABLE_API_TOKEN=redacted SMEE_URL=https://smee.io/2mxhU4Pb2YrNvF8E go run main.go

## Features

### API endpoints

 - Totals: http://localhost:3000/totals

### Webhook

Prevents unneeded http requests:
- regularly polling
- having to request refetch.

Allows getting the data ASAP.

### Websocket

Allows seeing updates immediately on the UI instead of having to
refresh the page or have a button to fetch data again.

### Possible improvements

- Only fetch updated/added data from webhook event instead of all records every time