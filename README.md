## INSTALLATION

    $ go get github.com/convox/praxis/cmd/cx

## USAGE

#### Start the DNS proxy

    $ sudo cx rack frontend

#### Start a local Rack

    $ cx rack start

#### Build a convox.yml

    $ cd ~/myapp
    $ vi convox.yml
    
#### Deploy the application to your local Rack

    $ cx apps create myapp
    $ cx deploy

#### Set environment variables

    $ cx env set FOO=bar

#### See running processes

    $ cx ps

#### Start the application in the foreground

    $ cx start

## DEVELOPMENT

    $ make dev

## COPYRIGHT

Convox, Inc. 2017

## LICENSE

[Apache License, v2.0](https://www.apache.org/licenses/LICENSE-2.0)
