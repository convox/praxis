# Getting Started for Production Deploys

Convox Praxis is a universal infrastructure framework. When you develop and deploy applications using the Praxis CLI, API and SDK you completely abstract away concerns about where your application is running. In minutes you can set up a system that has perfect development / production parity and deploy your app to the cloud.

The [getting-started-local.md](Getting Started for Local Development) guide walked you through installing the Praxis CLI and setting up a Docker-based development environment for an app.

This guide will walk you through creating a Convox account and setting up an AWS-based production environment for an app.

Together you'll see how Praxis offers an app workflow -- build, config, diff, test and promote -- that works exactly the same in development and production. The result is a simple, fast and portable dev, test and deploy workflow.

## Setting up your organization

### Sign up for Convox

Visit the [Convox signup page](https://ui.convox.com/auth/new). Here you can sign up with an email and set a password, or sign up through GitHub or Google.

### Create an organization

Next, create an organization. 

We'll call the org "ingen", the name of our exciting biotech startup.

All of your Convox resources -- integrations, Racks and apps -- will belong to this org. Access to these resources will be shared with other org members.

### Manage team members

Optionally, you can now visit the [org page](https://ui.convox.com/org) to set up team members. Since you created the org, you have been given an *admin role*.

Click "Add Another User", and enter the email address of your colleague.

If you want them to help manage integrations, create Racks, and invite other team members, also grant them the "admin" role. If you want to restrict them to just managing apps, grant them the "dev" role.

## Setting up your production environment

### Setting up your AWS integration

Next, visit the [integrations page](https://ui.convox.com/integrations). Here you will see how Convox is the hub that connects your organization to other service providers like AWS for infrastructure and GitHub for source control.

Click the [Enable AWS](https://ui.convox.com/integrations/aws/new) button.

We'll name the AWS integration "production" because it will connect to our primary AWS account and eventualy host our production Rack and apps.

Next, supply administrator access keys. Convox will use these keys once to set up the integration, then discard them. We recommend that you create a new "IAM user with programmatic access" to generate new keys.

Follow the [Creating an IAM Users](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html#id_users_create_console) guide to generate these.

1. Sign into the [AWS IAM console](https://console.aws.amazon.com/iam/home#/users)
2. In the "User name" field enter "convox-integration-setup"
3. In the "Access type" field, select "Programmatic access"
4. Click "Next: Permissions"
5. Select the "Attach existing policies directly" option, then check the "AdministratorAccess" policy
6. Click "Next: Review" then "Create User"
7. Click "Download .csv"

Now you can drag the "credentials.csv" file onto the [New AWS Integration](https://ui.convox.com/integrations/aws/new)

Then click "Enable". You'll see a confirmation message like "AWS integration enabled for account 922560784203". Convox is now integrated with your AWS account and can set up Racks.

Finally you can delete the "convox-integration-setup" IAM user.

### Install the production environment

Next visit the [new Rack page](https://ui.convox.com/racks/new). Here you can install Rack in your AWS account.

We'll use the default settings here. We name the Rack "production" because it will host our production apps. We use the standard "us-east-1" region, though Rack works in 10 regions. We'll use the "production" AWS integration that we just set up.

Click "Install". You'll see a confirmation message the the "Rack is installing". In a few minutes you will see a status of "installed", and your production environment will be up and running.

## Deploying your first app

### Clone the example app

We'll use the Praxis documentation site to demonstrate deployment. It's a Go app using the Hugo project for static websites.

If you don't already have it from the local development guide, clone the app and enter its directory:

    $ git clone https://github.com/convox/praxis-site.git
    $ cd praxis-site/

### Connect the CLI to the production environment

Now connect the `cx` command to your Convox account.

    $ cx login
    Email: john@ingen.com
    Password: *****
    OK

If you signed up with GitHub or Google, visit the [edit user](https://ui.convox.com/user/edit) page to set your account password first.

Then you can list your Racks and switch to your production Rack:

    $ cx racks
    ingen/production

    $ cx switch ingen/production
    OK

### Deploy the app

Now that you've seen what a Praxis app looks like, you can deploy it to your local Rack.

First you'll need to create an app in your Rack to use as a deployment target:

    $ cx apps create praxis-site

You should now see it in your apps list:

    $ cx apps
    NAME         STATUS
    praxis-site  running

Now deploy:

    $ cx deploy
    building: /Users/matthew/code/convox/praxis-site
    uploading: OK
    starting build: eed730a1180227074e774357acf8201cd39fe8f7478c367374ced3ded78cb92e
    preparing source
    restoring cache
    building: .
    running: docker build -t 9836064b94124bad54f83c70026dd85fcb8b5a13 /tmp/503720936
    Sending build context to Docker daemon  19.56MB
    Step 1/2 : FROM convox/hugo:0.0.1
     ---> 95f8d1e0347e
    Step 2/2 : COPY . /app
     ---> 1ae0dab8258d
    Removing intermediate container 256b517ec707
    Successfully built 1ae0dab8258d
    running: docker tag 9836064b94124bad54f83c70026dd85fcb8b5a13 convox/praxis-site/web:BLFMGFUNTS
    saving cache
    storing artifacts
    ...
    starting: convox.praxis-site.service.web.1
    starting: convox.praxis-site.service.web.2

The application is now deployed to your local Rack. You can find its endpoints with the CLI:

    $ cx services
    NAME  ENDPOINT
    web   https://praxis-site-web.prod-balan-yqveh744gpex-2137821817.us-east-1.rack.convox.io/

You can visit the service endpoint to view it.

With a Convox Organization, an AWS integration and a production Rack we have an app running with:

* A static hostname
* Trusted SSL
* Load balancing to two containers

### Update the hostname and certificate

We can use `cx` to manage the app config.

    $ cx env set HOST=praxis-site.ingen.com
    OK
    $ cx promote
    ...

