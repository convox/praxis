# frontend

## Usage

    $ frontend

## Add an endpoint

    $ curl -X POST http://10.42.84.0:9477/endpoints -d port=5000 -d addr=127.0.0.1:5443
    10.42.84.1:5000

    $ curl -k https://10.42.84.1:5000/apps
