# Cloud Run - One Click Deploy
Uses https://github.com/GoogleCloudPlatform/cloud-run-button to bootstrap your own Git Gateway on Cloud Run with minimal configuration.

The main reason to get a Git Gateway is to authenticate users that do not have a GitHub/GitLab account to write to your repository.
If you just want to allow access to users that already can access your site, please stop here, and use `backend: github` in your Netlify CMS config. If you need non-technical people without GitHub/GitLab accounts to access your CMS, this repository is for you.

Once we establish that the user is trusted (via a trusted JWT), the Git Gateway uses the pre-configured Git credentials to read/write the repository.

Because the one setting up the CMS often does have a GitHub account,
this minimal CloudRun installation offers to log you in using GitHub/GitLab,
and you can then authorize others to utilize your (temporary/long-lived) access token
stored in this servers database to access GitHub on your behalf.

## Authentication
There are 3 ways to setup Git Gatewway.

### Method 1: single repository using PAT
This requires a PAT (Personal Access Token) or Deploy Key.

### Method 2: single repository using OAuth app
This method requires you to setup a GitHub OAuth App. Follow the documentation of https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app and fill in the parameters that are displayed on the setup page on your Cloud Run Git Gateway.

### Method 3: organization-wide or public
This method requires you to setup a GitHub App. Follow the documentation of https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app and fill in the parameters that are displayed on the setup page on your Cloud Run Git Gateway.

After creating the app, it can be installed by navigating to:

```markdown
https://github.com/apps/<app name>/installations/new
```

## Usage
Configure the CMS like this:
```yaml
# file: config.yaml
backend:
  name: git-gateway
  repo: user/repo   # Path to your Github/Gitlab repository
  branch: master    # Branch to update
  site_domain: your-domain.com
  base_url: https://your.cloudrun.url.app # Path to your cloud run as ext auth provider
  auth_endpoint: /oauth/auth
  # TODO https://github.com/igk1972/netlify-cms-oauth-provider-go
  # TODO "github" backend https://github.com/netlify/netlify-cms/blob/master/packages/netlify-cms-backend-github/src/AuthenticationPage.js
```

<!-- https://divinewellbeing.nl/.netlify/identity/settings -->
```json
{
    "autoconfirm": false,
    "disable_signup": true,
    "external": {
        "bitbucket": false,
        "email": true,
        "facebook": false,
        "github": false,
        "gitlab": false,
        "google": false,
        "saml": false,
    },
    "external_labels": {}
}
```