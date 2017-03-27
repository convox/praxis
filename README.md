## INSTALLATION

    $ go get github.com/convox/praxis/cmd/cx

## USAGE

#### Start the DNS proxy

    $ sudo cx rack frontend

#### Start a local Rack

    $ cx rack start

#### Create a convox.yml

    $ cd ~/myapp
    $ vi convox.yml
    
###### Examples

  * [rails](https://gist.github.com/ddollar/4c2368dbb7058652cfe758affd2208b2)
  * [contrived](https://gist.github.com/ddollar/df189f18b44a233294dc6627c130d9e7)
  * [praxis](https://github.com/convox/praxis/blob/master/convox.yml)

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
