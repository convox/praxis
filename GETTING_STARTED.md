# Getting Started With Convox Praxis

Convox Praxis is a universal infrastructure interface. When you develop and deploy applications using Praxis you completely abstract away concerns about where your application is running. In minutes you can set up a system that has perfect development / production parity and deploy your app to the cloud.

This guide will walk you through installing the Praxis CLI, the Rack deployment platform, and deploying an application.

## Install the CLI

First, install the `cx` command line client.

#### MacOS

    $ curl https://s3.amazonaws.com/praxis-releases/cli/darwin/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

#### Linux

    $ curl https://s3.amazonaws.com/praxis-releases/cli/linux/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

## Setting Up Your App Platform

Your applications will run on a private platform called a Rack. While your production Rack will likely run on a cloud infrastructure provider like AWS, you can also install a "local" Rack on your development computer. This makes it easy to achieve dev/prod parity. You can write your application against the Praxis spec without worrying about where your app will actually be running.

### Local Rack

To install a local Rack you first need to install Docker. The free Docker Community Edition can be found for your OS [here](https://www.docker.com/community-edition).

Once you have Docker up and running you can use `cx` to install a local Rack:

    $ sudo cx rack install local

This will install a local Rack that boots when your computer boots.

### Cloud Rack

The local Rack is great for development, but you'll also want to set up a production Rack where you can deploy your apps to the internet and make them accessible to others.

Convox currently supports AWS as a cloud infrastructure provider. When you install a Rack on AWS, `cx` will inherit its login info. To make sure this is working correctly first run:

    $ aws configure

Use the login and region info where you want to install your Rack. Then run the installation command.

    $ cx rack install aws

When the installation completes, a `RACK_URL` is returned. Export this to your local environment to get `cx` talking to the AWS rack.

    $ export RACK_URL=<returned URL>

## Deploying an Application

The following steps will guide you through deploying an application. They are the same regardless of where your Rack is running.

### Clone the Example App

    $ git clone https://github.com/convox-examples/example.git
    $ cd example

### Create the App

    $ cx apps create example

This will provision an application in your rack that can be deployed to.

### Set Environment Variables

Configure any environment variables needed by your app.

    $ cx env set FOO=bar BAZ=qux -a example

### Deploy

Build and deploy your application.

    $ cx deploy -a example

### Locate Endpoints

When the deployment finishes you can fetch URLs for any services in your application that define ports.

    $ cx services -a praxis-site
    NAME  ENDPOINT
    web   https://web.service.example.convox

## Run in Development Mode

Deployments are great for completed changes, but when you're developing you want to see changes reflected immediately. For this you can use:

    $ cx start

This will rebuild your app and start streaming its logs live to your terminal. It will also start watching your local filesystem. Any changes to your local files will be instantly synced into the containers running in your local rack. This lets you see the effect of changes without having to redeploy your appliction repeatedly.

## Delete an Application

You can delete an application at any time with:

    $ cx apps delete example

## Uninstall a Rack

A rack can be uninstalled using the `cx` tool as well.

    $ sudo cx rack uninstall local

    $ cx rack uninstall aws
