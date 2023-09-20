
# Contributing

Please read the below before contributing any changes to the project.

## Committing

All commits should use angular-style commit messages: https://www.conventionalcommits.org/en/v1.0.0/.

Generally, the only exception around not having a 'type: ' prefix is whilst fixing an issue caused by a previous commit in a branch (i.e. a bug that didn't reach the main branch without the fix).

A Gitlab issue must be included at the bottom of the commit message `Issue #123`.

This format is important for the CI to perform releases and update release notes accurately.

If there isn't a gitlab issue for the task that you are working on, please feel free to ask for one (raise a Github issue and maintainer can rcreate an issue for you. Alternatively, request access to Gitlab, which should be granted)


## Upstream

The main upstream of this project is https://gitlab.dockstudios.co.uk/pub/terrareg.

Any changes submitted to any other instance (e.g. Github, Gitlab Cloud etc.) will be manually applied to the upstream and will then replicate downstream to other locations.

If you'd prefer, feel free to register an account to the main upstream - though registrations require approval, so just ping me an email :)

Any issues raised in any hosting platform, other than the 'main upstream' will be manually replicated to the 'main upstream'.

As a result, issue references in commit messages will only align with those in the 'main upstream'.
