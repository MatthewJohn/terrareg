

## Github Authentication

Github authentication is used to:

* Authenticate users to Terrareg:
  * determine which organisations the authenticated user is an owner of,
* Authenticate Terrareg to github to allow it access information:
  * Obtain repository information (public)
  * Obtain releases for repositories (public)
  * Setup repository webhooks (private)

To achieve this, several methods are used:

 * Terrareg should be setup as a Github app
 * If a user authenticates using Github, Terrareg will generated a token to obtain information about their organisations and their membership status of those organisations.
   * These tokens are temporary
 * When a user attempts to create a provider using a repository from a Github namespace, it will attempt to perform a Github app installation, to generate secrets to authenticate.
 * If a site admin attempts to create a provider for a github organisation that they are not a member of, the Github app credentials will be used to obtain the required information.  
