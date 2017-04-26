# convox/praxis

**WARNING: This project is currently an *alpha* release and is not recommended for production or the faint of heart.**

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

## DEPLOY AN APP

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
