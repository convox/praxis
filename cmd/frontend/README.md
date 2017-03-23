# frontend

## Usage

    $ sudo frontend

## Add an endpoint

    $ curl -X POST http://10.42.84.0:9477/endpoints -d host=foo.bar.convox -d port=443 -d addr=127.0.0.1:5443
    10.42.84.1:443

## Configure DNS

Add `10.42.84.0` as a DNS server on your network interface

## Test endpoint

    $ curl -k https://foo.bar.convox/apps
