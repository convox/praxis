# Convox Praxis

A framework for modern application infrastructure.

**WARNING: This project is currently an *alpha* release and is not recommended for production or the faint of heart.**

## ABOUT

Praxis allows you to specify the entire infrastructure for your application.

```yaml
caches:
  sessions:
    expire: 1d
keys:
  master:
    roll: 30d
queues:
  mail:
    timeout: 1m
services:
  web:
    build: .
    port: 3000
    scale: 2-10
timers:
  cleanup:
    schedule: 0 3 * * *
    command: bin/cleanup
    service: web
```

### API

Praxis makes these primitives available to your application with a simple API.

```
# list applications
GET /apps

# put an item on a queue
POST /apps/myapp/queues/mail

# get an item from a queue
GET /apps/myapp/queues/mail

# encrypt some data
POST /apps/myapp/keys/master/encrypt
```

#### SDK

* [Go](https://github.com/convox/praxis/tree/master/sdk/rack)

### Providers

A Praxis Rack can be installed into one of many available infrastructure providers to expose the Praxis API.

#### Local

Runs on your laptop (or any single node). Great for development and CI.

#### AWS

A fault-tolerant, highly scalable architecture built on modern AWS services such as ECS, ALB, and Lambda.

### API

[TODO: API Docs]()

## INSTALLATION

### CLI

#### MacOS

    $ curl https://s3.amazonaws.com/praxis-releases/cli/darwin/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

#### Linux

    $ curl https://s3.amazonaws.com/praxis-releases/cli/linux/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

### Rack

#### Local

    $ sudo cx rack install local

##### Local Rack installation requires `sudo` to resolve and route local hostnames such as `myapp-web.convox`

#### AWS

    $ cx rack install aws

## DEPLOY

#### Create a convox.yml

[TODO: Reference Docs]()

See also the `examples/` directory of this repo.

#### Create an application

    $ cx apps create myapp

#### Set environment variables

    $ cx env set FOO=bar

#### Deploy the applications

    $ convox deploy

## UPDATING

#### CLI

    The `cx` CLI will automatically keep itself up to date.
    
#### Rack

    $ cx rack update [version]

## Guide

For more a more detailed setup guide, see "[Getting Started](https://github.com/convox/praxis/blob/master/GETTING_STARTED.md)".

## COPYRIGHT

Convox, Inc. 2017

## LICENSE

[Apache License, v2.0](https://www.apache.org/licenses/LICENSE-2.0)
