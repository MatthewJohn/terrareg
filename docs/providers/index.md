# Provider support

Terrareg supports Terraform providers and indexing via Github - this is currently in an alpha release and still requires some additional work to be completely usable.

WARNING: (yes, these are fairly fundamental missing features, but this is an alpha feature) it is not possible to refresh new versions of providers at present, nor is it possble to delete/modify providers after creation.

To add providers:
 * Setup a "Provider source" for the VCS that the provider is hosted on.
 * Following the Hashicorp documentation for publishing providers in Github (adding GPG key, Github actions for creating releases and generated SHA and SHA.sig files).
 * Authenticate to Terrareg either via the VCS provider SSO and install the created Github application in the user/org that the provider is present OR (if `default_access_token` provider source config has been configured).
 * Goto 'Create -> Provider', select the namespace that matches the VCS org/user (use 'Refresh namespaces' if need be)
 * Select the repository and click 'Create provider'
 * Terrareg will index the latest version, which must be published in Github at the time of creation.
