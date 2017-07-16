Release Checklist
    [ ] Local `cx test` passes against latest release:

        $ cx update && cx rack update && cx version
        client: 20170711112338
        server: 20170711112338
        
        $ NAME=praxis-test cx test

    [ ] AWS `cx test` passes against latest release:

        1. Open PR on GitHub
        2. Wait for workflow results
        3. Retry workflow on transient errors

        AWS `cx test` can also be run interactively:

            $ cx switch convox/staging
            $ cx test

The release from the PR itself needs to preserve or fix tests. This can be verified with a local rack
and praxis-local and praxis-aws apps with the new code:

    [ ] Local `cx test` passes against PR

        $ go build ./cmd/cx
        $ ./cx apps create praxis-local
        $ ./cx env set -a praxis-local DEVELOPMENT=true NAME=praxis-local PROVIDER=local PROVIDER_ROUTER=none VERSION=head
        $ ./cx start -a praxis-local
        $ export RACK_URL=https://rack.praxis-local.convox
        $ ./cx version
        client: dev
        server: head

        $ cx test

    [ ] AWS `cx test` passes against PR

        $ ./cx apps create praxis-aws
        $ ./cx env set -a praxis-aws NAME=praxis-aws PROVIDER=aws VERSION=head
        $ ./cx start -a praxis-local


    Local `cx test` passes
    AWS `cx test` passes

    cx test works locally
        on last release?
            uninstall / install
        on current release?
            uninstall / install
            `make something`
                to replace latest cx and 




Good release?
    cx test works locally
        problems
            convergence bug
            cmd/qa tests need to be deactivated

    Workflows pass
    Manual QA
        rack install local
            app deploy
            app start
        rack install aws
            app deploy
            traffic