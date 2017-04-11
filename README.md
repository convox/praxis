## INSTALLATION

#### Download the CLI

    $ go get github.com/convox/praxis/cmd/cx

#### Install a local Rack

    $ sudo cx rack install local

##### Rack installation requires `sudo` to resolve and route local hostnames like `web.myapp.convox`

## USAGE

#### Create a convox.yml

    $ cd ~/myapp
    $ vi convox.yml

###### Examples

  * [rails](https://gist.github.com/ddollar/4c2368dbb7058652cfe758affd2208b2)
  * [contrived](https://gist.github.com/ddollar/df189f18b44a233294dc6627c130d9e7)
  * [praxis](https://github.com/convox/praxis/blob/master/convox.yml)

#### Start an application in development mode

    $ cx apps create myapp
    $ cx start

#### Set environment variables

    $ cx env set FOO=bar

#### See running processes

    $ cx ps

## COPYRIGHT

Convox, Inc. 2017

## LICENSE

[Apache License, v2.0](https://www.apache.org/licenses/LICENSE-2.0)
