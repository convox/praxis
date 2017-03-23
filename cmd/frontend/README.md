# frontend

## Usage

    $ sudo frontend

## Add an endpoint

    $ curl -X POST http://10.42.84.0:9477/endpoints/foo.bar.convox -d port=443 -d target=127.0.0.1:5443
    {
      "host": "foo.bar.convox",
      "ip": "10.42.84.1",
      "port": 443,
      "target": "127.0.0.1:5443"
    }

## Configure DNS

Add `10.42.84.0` as a DNS server on your network interface

## Test endpoint

    $ curl -k https://foo.bar.convox/apps
