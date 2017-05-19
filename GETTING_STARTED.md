# Getting Started With Convox Praxis

Convox Praxis is a universal infrastructure interface. When you develop and deploy applications using Praxis you completely abstract away concerns about where your application is running. In minutes you can set up a system that has perfect development / production parity and deploy your app to the cloud.

This guide will walk you through installing the Praxis CLI, the Rack deployment platform, and deploying an application.

## Setting up your development platform

### Install the CLI

First, install the `cx` command line client.

#### MacOS

    $ curl https://s3.amazonaws.com/praxis-releases/cli/darwin/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

#### Linux

    $ curl https://s3.amazonaws.com/praxis-releases/cli/linux/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

### Install the development platform

Your applications will run on a private platform called a "Rack". While your production Rack will likely run on a cloud infrastructure provider like AWS, you can also install a "local" Rack on your development computer. This makes it easy to achieve dev/prod parity.

To install a local development Rack you first need to install Docker. The free Docker Community Edition can be found for your OS [here](https://www.docker.com/community-edition).

Once you have Docker up and running you can use `cx` to install a local Rack:

    $ sudo cx rack install local
    installing: /Library/LaunchDaemons/convox.frontend.plist
    installing: /Library/LaunchDaemons/convox.rack.plist

This will install a local Rack that boots when your computer boots.

### Clone the example app

Let's use an example app to see how deployment works. We'll use the Praxis documentation site for this example. It's a Go app using the Hugo project for static websites.

Clone the app and enter its directory:

    $ git clone https://github.com/convox/praxis-site.git
    $ cd praxis-site/

#### convox.yml

The first thing to take note of in the project is the `convox.yml` file. This is where the app's description and configuration live.

```yaml
services:
  web:
    command:
      test: make test
    port: 1313
```

The `convox.yml` for this site is pretty simple. It defines a single service called "web".

Containers for the web service will listen on port 1313 for requests.

The project will be built from a `Dockerfile` in the same directory. Unlike `docker-compose.yml`, `convox.yml` does not require you to specify a `build: .` stanza if the app is to be built from a `Dockerfile` in the same directory. It is implied.

Commands for different environments can be defined. Here we define a test command. For other environments the `CMD` from the `Dockerfile` will be inherited.

"Services" is just one of many components in the Praxis spec that you can define in a `convox.yml`. The project that you're currently in, `praxis-site` is under active development to explain the entire scope of Praxis, so stay tuned for updates.

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
    starting: convox.praxis-site.endpoint.web (https://web.service.praxis-site.convox:443)
    starting: convox.praxis-site.service.web.1

The app will now be available at the returned endpoint, [https://web.service.praxis-site.convox:443](https://web.service.praxis-site.convox). Try clicking this link to load the app in your browser.

You can fetch endpoints for your services at any time:

    $ cx services
    NAME  ENDPOINT
    web   https://web.service.praxis-site.convox

### Edit the source

Now that you have the app up and running, you can explore the development cycle by making a change to the source code and deploying it to your development rack.

Open `content/index.md` in the project and add the text "Hello, this is a change!" right below the Introduction header. After the edit your file should look like this:

    ---
    title: Introduction
    weight: 5
    ---
    
    # Introduction
    
    Hello, this is a change!

### Test the app

You can test an app using the `cx test` command. It will build the app, deploy it to a temporary test app, run test commands defined on each service, output the results, and then delete the test app from the Rack:

    $ cx test
    creating app test-1495232395: OK
    building: /Users/matthew/code/convox/praxis-site
    uploading: OK
    starting build: aa00646feff3ae71935329157fba982456c7d440953a0d697fc2f599509ede5e
    preparing source
    building: .
    running: docker build -t 9836064b94124bad54f83c70026dd85fcb8b5a13 /tmp/526868377
    Sending build context to Docker daemon  12.22MB
    Step 1/2 : FROM convox/hugo:0.0.1
     ---> 95f8d1e0347e
    Step 2/2 : COPY . /app
     ---> b02503ae7b30
    Removing intermediate container 879f6b064ded
    Successfully built b02503ae7b30
    running: docker tag 9836064b94124bad54f83c70026dd85fcb8b5a13 convox/test-1495232395/web:BUOVBPTLQF
    saving cache
    storing artifacts
    web     | running: make test
    web     | test -f static/images/logo.png

If you'd like to see the test fail, just delete `static/images/logo.png` and run `cx test` again.

### Build the app

Now you can ship this to the Rack. Rather than doing a `cx deploy`, this time you'll just build the source. That will allow you to inspect the diff before completing deployment.

    $ cx build
    uploading: .
    starting build: 16dd308f6b61e40b768d60eb8238b758dee8e9a3848d2330b0ca9bd26a817031
    preparing source
    restoring cache
    building: .
    running: docker build -t 9836064b94124bad54f83c70026dd85fcb8b5a13 /tmp/124604449
    Sending build context to Docker daemon 19.56 MB
    Step 1/2 : FROM convox/hugo:0.0.1
     ---> 95f8d1e0347e
    Step 2/2 : COPY . /app
     ---> 5c9e32a1e857
    Removing intermediate container 8bebd36bc5a9
    Successfully built 5c9e32a1e857
    running: docker tag 9836064b94124bad54f83c70026dd85fcb8b5a13 convox/praxis-site/web:BHRATEYFZS
    saving cache
    storing artifacts

### View releases

Every time you build your app (or change an environment variable) a new "release" is created to keep up with these changes. You can list these releases:

    $ cx releases
    ID          BUILD       STATUS    CREATED
    RTKJFWMKYG  BHRATEYFZS  created   4 minutes ago
    RYCQLGAAAV  BJKETOESCA  promoted  19 minutes ago

You can see from this list that the most recent release, `RTKJFWMKYG`, was created but not promoted. To "promote" a release means to make it the current live version.

### Diff releases

Before you promote the release, you can diff it to make sure you're deploying exactly what you expect:

    fetching RTKJFWMKYG: OK
    fetching RYCQLGAAAV: OK
    diff --git a/var/folders/cy/f_d_fkvn4jxckgbrxlxp4xpc0000gn/T/718088251/content/index.md b/var/folders/cy/f_d_fkvn4jxckgbrxlxp4xpc0000gn/T/004534604/content/index.md
    index 0dbdcc5..a93138a 100644
    --- a/var/folders/cy/f_d_fkvn4jxckgbrxlxp4xpc0000gn/T/718088251/content/index.md
    +++ b/var/folders/cy/f_d_fkvn4jxckgbrxlxp4xpc0000gn/T/004534604/content/index.md
    @@ -5,6 +5,8 @@ weight: 5
    
     # Introduction
    
    +Hello, this is a change!
    +

Once you verify that the diff is correct you can promote it.

### Promote the app

    $ cx promote
    promoting RTKJFWMKYG: OK
    starting: convox.praxis-site.endpoint.web (https://web.service.praxis-site.convox:443)
    starting: convox.praxis-site.service.web.1

The release will now show as promoted in the list.

    $ cx releases
    ID          BUILD       STATUS    CREATED
    RTKJFWMKYG  BHRATEYFZS  promoted  11 minutes ago
    RYCQLGAAAV  BJKETOESCA  promoted  27 minutes ag

Refresh your browser to see your change in action!

### Manage environment variables

As previously mentioned, changing an environment variable also creates a relase. Here's how to see that in action.

The `cx env` command lists all of your app's environment variables. Run that now to see that it's empty:

    $ cx env

Now set an environment variable:

    $ cx env set FOO=bar
    updating environment: OK

    $ cx env
    FOO=bar

`cx releases` will now show that a new release has been created:

    $ cx releases
    ID          BUILD       STATUS    CREATED
    RDOAQYVUAK  BFTLZLBXCX  created   40 seconds ago
    RTKJFWMKYG  BHRATEYFZS  promoted  11 minutes ago
    RYCQLGAAAV  BJKETOESCA  promoted  27 minutes ag

You can now promote the latest release just as you did after a build:

    $ cx promote

### A faster development loop

You may be thinking that this is shaping up to be a pretty nice development workflow, but it's a bit laborious. You'd be exactly right, but luckily we have a solution! There's a way to see all of your changes live as you develop:

    $ cx start

Go ahead and delete the "Hello, this is a change!" line you added previously. You'll be able to immediately view the changes in your browser.

#### Code sync

Convox code sync allows changes you make in your local files to be instantly reflected in the containers running in your local rack. This lets you see the effect of changes without having to redeploy your appliction repeatedly.

Any directory that appears in a `COPY` or `ADD` line in your Dockerfile will be synced. This project has:

    COPY . .

in the Dockerfile, so the entire project directory is synced.

## Setting up a production platform

The local Rack is great for development, but eventually you'll also want to set up a production Rack on the internet where you can deploy your apps and make them accessible to others.

### Install a production Rack

Convox currently supports Amazon Web Services (AWS) as a cloud infrastructure provider. When you install a Rack on AWS, `cx` will inherit its login info from the [AWS CLI](http://docs.aws.amazon.com/cli/latest/userguide/installing.html). To make sure this is working correctly first run:

    $ aws configure
    AWS Access Key ID [****************W7GA]:
    AWS Secret Access Key [****************g0bb]:
    Default region name [us-east-1]:
    Default output format [json]:

Verify that the login and region info reflect where you want to install your Rack. Then run the installation command.

    $ cx rack install aws
    CREATE_IN_PROGRESS    convox                        AWS::CloudFormation::Stack
    CREATE_IN_PROGRESS    RackCluster                   AWS::ECS::Cluster
    CREATE_IN_PROGRESS    RackCluster                   AWS::ECS::Cluster
    CREATE_IN_PROGRESS    Volumes                       AWS::EFS::FileSystem
    CREATE_IN_PROGRESS    RackRegistries                AWS::SDB::Domain
    CREATE_COMPLETE       RackCluster                   AWS::ECS::Cluster
    CREATE_IN_PROGRESS    Network                       AWS::CloudFormation::Stack
    CREATE_IN_PROGRESS    Volumes                       AWS::EFS::FileSystem
    CREATE_COMPLETE       convox                        AWS::CloudFormation::Stack
    . . .
    RACK_URL=https://...

The CloudFormation output will be streamed back to your terminal as the installation progresses. When the installation completes, a `RACK_URL` is returned.

**Make sure you note the value of `RACK_URL`. It is not recoverable. This is a temporary situation while Praxis is in beta. Official Rack management is coming soon**

Export `RACK_URL` to your local environment to get `cx` talking to the AWS rack.

    $ export RACK_URL=https://715f4971060be34d64d9e2f39f33820d:@rack.convo-Balan-1EDQ1YOXSKZGL-1009186257.us-east-1.rack.convox.io

### Deploy to production

Now you can deploy your application to your production Rack. You've already verified everything on your development rack, so you can deploy with confidence.

    $ cx apps create praxis-site

    $ cx deploy

Remember that when the deployment completes you can find your services' URLs with:

    $ cx services

Congratulations! You've just set up a powerful development workflow and deployed your application to robust cloud infrastructure! Time to build more apps :)
