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

Next, visit the [integrations page](https://ui.convox.com/integrations). This is the hub that connects your organization to other service providers like AWS for infrastructure and GitHub for source control.

Click the [Enable AWS](https://ui.convox.com/integrations/aws/new) button.

We'll name the AWS integration "production" because it will connect to our primary AWS account and eventualy host our production Rack and apps.

Next, supply administrator access keys. Convox will use these keys once to set up the integration, then discard them. We recommend that you create a new "IAM user with programmatic access" to generate new keys. Follow the [Creating an IAM Users](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html#id_users_create_console) guide to generate these.

1. Sign into the [AWS IAM console](https://console.aws.amazon.com/iam/home#/users)
2. In the "User name" field enter "convox-integration-setup"
3. In the "Access type" field, select "Programmatic access"
4. Click "Next: Permissions"
5. Select the "Attach existing policies directly" option, then check the "AdministratorAccess" policy
6. Click "Next: Review" then "Create User"
7. Click "Download .csv"

Now you can drag the "credentials.csv" file onto the [New AWS Integration](https://ui.convox.com/integrations/aws/new) form.

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
    Authenticating with ui.convox.com: OK

If you signed up with GitHub or Google, visit the [edit user](https://ui.convox.com/user/edit) page to set your account password first.

Then you can list your Racks and switch to your production Rack:

    $ cx racks
    RACKS
    ingen/production
    local

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
    starting build: bc5f7812-d4a1-4107-8d7f-1390e8b9b196
    preparing source
    building: .
    running: docker build -t 9836064b94124bad54f83c70026dd85fcb8b5a13 /tmp/288499397
    Sending build context to Docker daemon  11.23MB
    Step 1/2 : FROM convox/hugo:0.0.1
    ---> 95f8d1e0347e
    Step 2/2 : COPY . /app
    ---> 5ed5990dcd19
    Removing intermediate container b949cac258ce
    Successfully built 5ed5990dcd19
    running: docker tag 9836064b94124bad54f83c70026dd85fcb8b5a13 production-praxis-site/web:BWWPTMIDWL
    running: docker tag production-praxis-site/web:BWWPTMIDWL 665986001363.dkr.ecr.us-east-1.amazonaws.com/produ-repos-2axsg073lrv8:web.BWWPTMIDWL
    pushing: 665986001363.dkr.ecr.us-east-1.amazonaws.com/produ-repos-2axsg073lrv8:web.BWWPTMIDWL
    UPDATE_IN_PROGRESS    production-praxis-site        AWS::CloudFormation::Stack
    CREATE_COMPLETE       ServiceWebTargetGroup         AWS::ElasticLoadBalancingV2::TargetGroup
    CREATE_COMPLETE       ServiceWebListenerRule        AWS::ElasticLoadBalancingV2::ListenerRule
    CREATE_COMPLETE       ServiceWebTasks               AWS::ECS::TaskDefinition
    CREATE_COMPLETE       ServiceWeb                    AWS::ECS::Service
    UPDATE_COMPLETE       production-praxis-site        AWS::CloudFormation::Stack
    release promoted: RNPMYNUTQO

The application is now deployed to the production Rack. You can find its endpoints with the CLI:

    $ cx services
    NAME  ENDPOINT
    web   https://praxis-site-web.produ-balan-yqveh744gpex-2137821817.us-east-1.rack.convox.io/

You can visit the service endpoint to view it.

With a Convox Organization, an AWS integration, the `convox.yml` file and a `cx deploy` command, we have:

* A production-ready private cloud
* A static, online hostname
* Trusted SSL
* Load balancing to two containers

### Update the hostname and certificate

Now you can use the `cx` tool to manage the app config. For example, you can update the service hostname:

    $ cx env set HOST=praxis-site.ingen.com
    updating environment: OK
    $ cx promote
    promoting RAHDPGMMUR: OK
    UPDATE_IN_PROGRESS    production-praxis-site        AWS::CloudFormation::Stack
    CREATE_COMPLETE       ServiceWebBalancerTargetGroup  AWS::ElasticLoadBalancingV2::TargetGroup
    CREATE_COMPLETE       ServiceWebBalancerSecurity    AWS::EC2::SecurityGroup
    CREATE_COMPLETE       ServiceWebCertificate         AWS::CertificateManager::Certificate
    CREATE_COMPLETE       ServiceWebBalancer            AWS::ElasticLoadBalancingV2::LoadBalancer
    CREATE_COMPLETE       ServiceWebBalancerListener80  AWS::ElasticLoadBalancingV2::Listener
    release promoted: RZMNJRSLSD

Now visit the service endpoint. A certificate is configured for `praxis-site.ingen.com` hostname. All that remains is adding a DNS CNAME for `praxis-site.ingen.com` to the service endpoint.

On the AWS Rack, certs are handled by the AWS Certificate Manager (ACM) service. Refer to the [ACM User Guide](http://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) if your domain is not yet set up with ACM.

## Going to automation

Now that you have a production Rack and app online, the Praxis CLI, API and SDK can be used for all your team's deployment workflows.

Becaus Praxis offers dev/prod parity, you can run tests in your AWS account with:

    $ cx test

This isn't as fast as running tests in a Local Rack, but it offers a shared test environment for your organization.

Congratulations! You've just set up a powerful development workflow and tested and deployed your application to robust cloud infrastructure!