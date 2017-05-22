# DEVELOPMENT

Start with a local Rack running in the background.

#### Start a development Rack

    $ cx start
    
#### Use the development Rack

    $ export RACK_URL=https://rack.praxis.convox:6443
    $ cx apps
    
#### Run the tests

    $ env VERSION=test cx test
