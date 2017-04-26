# convox/praxis

A framework for modern application infrastructure.

**WARNING: This project is currently an *alpha* release and is not recommended for production or the faint of heart.**

## ABOUT

Praxis allows you to specify the entire infrastructure for your application in raw primitives.

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

A Praxis Rack can installed into one of many available infrastructure providers to expose the Praxis API.

#### Local

Runs on your laptop (or any single node). Great for development and CI.

#### AWS

A fault-tolerant, highly scalable architecture built on modern AWS services such as ECS, ALB, and Lambda.

### Implementation

Praxis utilizes the best underlying infrastructure at each provider to implement a primitive. Some examples:

#### Cache

| Provider     | Implementation           |
|--------------|--------------------------|
| **Local**    | *convox/redis* container |
| **AWS**      | ElastiCache              |

#### Queue

| Provider     | Implementation  |
|--------------|-----------------|
| **Local**    | in-memory FIFO  |
| **AWS**      | SQS             |

## INSTALLATION

### CLI

#### MacOS

    $ curl https://s3.amazonaws.com/praxis-releases/cli/macos/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

#### Linux

    $ curl https://s3.amazonaws.com/praxis-releases/cli/linux/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

### Rack

#### Local

    $ sudo cx rack install local

##### Rack installation requires `sudo` to resolve and route local hostnames like `web.myapp.convox`

#### AWS

    $ cx rack install aws

## DEPLOY

#### Create a convox.yml

    $ cd ~/myapp
    $ vi convox.yml

###### Examples

  * [rails](https://gist.github.com/ddollar/4c2368dbb7058652cfe758affd2208b2)

#### Create an application

    $ cx apps create myapp

#### Set environment variables

    $ cx env set FOO=bar

#### Deploy the applications

    $ convox deploy

## COPYRIGHT

Convox, Inc. 2017

## LICENSE

[Apache License, v2.0](https://www.apache.org/licenses/LICENSE-2.0)
