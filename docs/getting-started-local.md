# Getting Started for Local Development

Convox Praxis is a universal infrastructure framework. When you develop and deploy applications using the Praxis CLI, API and SDK you completely abstract away concerns about where your application is running. In minutes you can set up a system that has perfect development / production parity and deploy your app to the cloud.

This guide will walk you through installing the Praxis CLI and setting up a Docker-based development environment for an app.

The [Getting Started for Production Deploys](getting-started-aws.md) will walk you through creating a Convox account and setting up a an AWS-based production environment for an app.

Together you'll see how Praxis offers an app workflow -- build, config, diff, test and promote -- that works exactly the same in development and production. The result is a simple, fast and portable dev, test and deploy workflow.

## Setting up your development environment

### Install the CLI

First, install the Praxis `cx` command line client.

#### MacOS

    $ curl https://s3.amazonaws.com/praxis-releases/cli/darwin/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

#### Linux

    $ curl https://s3.amazonaws.com/praxis-releases/cli/linux/cx -o /usr/local/bin/cx
    $ chmod +x /usr/local/bin/cx

Confirm that cx is correctly installed and up to date:

    $ cx update
    updating cli to 20170628170448: OK

### Install the development environment

Your applications will run in an isolated environment called a *Rack*. While your *production Rack* will run on a cloud infrastructure provider like AWS, you can also install a *local Rack* on your development computer. This makes it easy to achieve dev/prod parity.

To install a local Rack you'll first need to install Docker. The free Docker Community Edition can be found for your OS [here](https://www.docker.com/community-edition).

Once you have Docker up and running you can use `cx` to install a local Rack:

    $ sudo cx rack install local
    installing: /Library/LaunchDaemons/convox.rack.plist
    installing: /Library/LaunchDaemons/convox.router.plist

This starts the Praxis API on your computer, which the `cx` tool will use to manage apps.

This also starts the Praxis Router on your computer, which manages load balancing, DNS, and SSL certificates for your development apps. You can load the Praxis Certificate Authority (CA) public key into your keychain so all development SSL traffic is trusted:

    $ open /Users/Shared/convox/ca.crt

In the "Add Certificates" dialog, select the "System" keychain, and click "Add". Then in the "Keychain Access" app, search for "convox" and double click on "ca.convox". In the root certificate dialog, change "When using this certificate:" to "Always Trust" and close the dialog.

## Developing your first app

### Clone the example app

We'll use the Praxis documentation site to demonstrate development. It's a Go app using the Hugo project for static websites.

Clone the app and enter its directory:

    $ git clone https://github.com/convox/praxis-site.git
    $ cd praxis-site/

#### convox.yml

The first thing to take note of in the project is the `convox.yml` file. This is where the app's description and configuration live.

```yaml
services:
  web:
    certificate: ${HOST}
    port: http:1313
    scale: 2
    test: make test
```

The `convox.yml` for this site is pretty straightfoward. It defines a single service called `web`.

Nested under `web` is a `certificte` config. An SSL certificate will be automatically configured for the domain specified by the app's `HOST` environment variable. `HOST` is automatically set and can be overridden for a custom domain.

The `port` configuration means containers for the web service will listen on port 1313 for http requests.

Two copies of the container will be run, according to the `scale` setting.

The app's default test command is `make test` as configured by `test`. This will be used later in the guide.

The `convox.yml` you cloned also has a `workflows` section. You can ignore that for the purposes of this guide.

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
    starting: convox.praxis-site.service.web.1
    starting: convox.praxis-site.service.web.2

The application is now deployed to your local Rack. You can find its endpoints with the CLI:

    $ cx services
    NAME  ENDPOINT
    web   https://web.praxis-site.convox

You can visit [https://web.praxis-site.convox](https://web.praxis-site.convox) to view it.

With the `convox.yml` file and a `cx deploy` command we have an app running with:

* A static hostname
* Trusted SSL
* Load balancing to two containers

### Edit the source

Now that you have the app up and running, you can try the development cycle by making a change to the source code and deploying it to your local Rack.

Open `content/index.md` in the project and add the text "Hello, this is a change!" right below the Introduction header. After the edit your file should look like this:

    ---
    title: Introduction
    weight: 5
    ---
    
    # Introduction
    
    Hello, this is a change!

Then deploy the changes:

    $ cx deploy

Reload the site in your browser and verify that the Introduction text has changed.

### Run tests

You can test an app using `cx test`. This command will create a temporary app, deploy the current code to it, and sequentially run the `test:` command specified for each service. If a `test:` command is not specified, no tests will be run. `cx test` will abort and pass through any non-zero exit code returned by a test command.

    $ cx test
    convox  | creating app test-1498754013: OK
    build   | building: /Users/matthew/code/convox/praxis-site
    build   | uploading: OK
    build   | starting build: d62123b840ae443a061039c39fcce61f82988458f368090b4e0e76cb15a00221
    build   | preparing source
    build   | building: .
    build   | running: docker build -t 9836064b94124bad54f83c70026dd85fcb8b5a13 /tmp/144541219
    build   | Sending build context to Docker daemon  11.23MB
    build   | Step 1/2 : FROM convox/hugo:0.0.1
    build   |  ---> 95f8d1e0347e
    build   | Step 2/2 : COPY . /app
    build   |  ---> Using cache
    build   |  ---> f3d8a00acd8a
    build   | Successfully built f3d8a00acd8a
    build   | Successfully tagged 9836064b94124bad54f83c70026dd85fcb8b5a13:latest
    build   | running: docker tag 9836064b94124bad54f83c70026dd85fcb8b5a13 convox/test-1498754013/web:BFHEYFLOFN
    build   | storing artifacts
    build   | build complete
    release | starting: convox.test-1498754013.service.web.1
    release | starting: convox.test-1498754013.service.web.2
    web     | running: make test
    web     | test -f static/images/logo.png

If you'd like to see the test fail, just delete `static/images/logo.png` and run `cx test` again.

With the `convox.yml` file and a `cx test` command have achieved development / test environment parity!

### Building without deploying

While `cx deploy` is an easy way to deploy changes, the build, configure and promote steps are possible with the CLI so you can customize your workflow.

This time, let's create a build but not deploy it:

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

Building without deploying is useful to stage changes and then deploy them as a unit.

### Set an environment variable

A new release is also created when you change the application's environment.

    $ cx env set FOO=bar
    updating environment: OK

### View releases

Every time you build your app or change an environment variable, a new "release" is created to keep up with these changes. You can list these releases:

    $ cx releases
    ID          BUILD       STATUS    CREATED
    RGCMQGSYYN  BLFBFMFXFR  created   4 seconds ago
    RTTOIDQIFF  BLFBFMFXFR  created   2 minutes ago
    RTKJFWMKYG  BHRATEYFZS  promoted  4 minutes ago
    RYCQLGAAAV  BJKETOESCA  promoted  19 minutes ago

You can see from this list that the most recent release, `RGCMQGSYYN`, was created but not promoted which means its changes have not yet been deployed.

Releases that aren't promoted are useful to run pre-deploy commands like database migrations or asset uploads.

    $ cx run --release RGCMQGSYYN web hugo convert toJSON -o /tmp

### Diff releases

Before you promote a release, you can use `cx diff` to summarize the changes about to be deployed:

    $ cx diff
    fetching RWSHXASNDF: OK
    fetching RTKJFWMKYG: OK
    diff --git 663957140/.env 924574153/.env
    index e69de29..1566bb1 100644
    --- 663957140/.env
    +++ 924574153/.env
    @@ -0,0 +1 @@
    +FOO=bar

    diff --git 663957140/content/index.md 924574153/content/index.md
    index 308583e..82f79db 100644
    --- 663957140/content/index.md
    +++ 924574153/content/index.md
    @@ -5,7 +5,7 @@ weight: 5
    
    # Introduction
    
    -Hello, this is a change!
    +Hey, this is another change.

Once you verify the diff you can promote it.

### Promote a release

    $ cx promote
    promoting RWSHXASNDF: OK
    starting: convox.praxis-site.service.web.1
    starting: convox.praxis-site.service.web.2

The release will now show as promoted.

    $ cx releases
    ID          BUILD       STATUS    CREATED
    RWSHXASNDF  BOMBKQZCLA  promoted  4 minutes ago
    RGCMQGSYYN  BLFBFMFXFR  created   5 minutes ago
    RTTOIDQIFF  BLFBFMFXFR  created   7 minutes ago
    RTKJFWMKYG  BHRATEYFZS  promoted  11 minutes ago
    RYCQLGAAAV  BJKETOESCA  promoted  27 minutes ago

Refresh your browser to see your change in action!

### A faster development loop

You may be thinking that this is shaping up to be a pretty nice workflow but it's a bit laborious for development. You can use `cx start` to pull an application into the foreground. `cx start` will restart the services of the application in development mode and set up live code-sync with your local development checkout allowing you to use your own tools and editor.

    $ cx start

Go ahead and delete the "Hello, this is a change!" line you added previously. You'll be able to immediately view the changes in your browser.

#### Code sync

Convox code sync allows changes you make in your local files to be instantly reflected in the app containers. This lets you see the effect of changes without having to redeploy your appliction repeatedly.

Any directory that appears in a `COPY` or `ADD` line in your Dockerfile will be synced. This project has:

    COPY . .

in the Dockerfile, so the entire project directory is synced.

## Going to production

The local Rack is great for development, but eventually you'll also want to set up a production Rack on the internet where you can deploy your apps and make them accessible to others.

Because Praxis offers dev/prod parity, you can install a production environment in your AWS account with:

    $ cx rack install aws

After a few minutes of setup, you can use the same exact CLI workflow to deploy your first app to AWS:

    $ cx apps create praxis-site
    $ cx deploy
    $ cx services
    NAME  ENDPOINT
    web   https://praxis-site-web.prod-balan-yqveh744gpex-2137821817.us-east-1.rack.convox.io/

However, the Convox Graphical User Interface (GUI) makes it even easier manage your production Racks and Apps and share them with your development team.

Check out the [Getting Started for Production Deploys](getting-started-aws.md) guide to walk through creating a Convox account and setting up a an AWS-based production environment for an app.
