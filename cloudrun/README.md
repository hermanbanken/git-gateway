# Cloud Run - One Click Deploy
Uses https://github.com/GoogleCloudPlatform/cloud-run-button to bootstrap your own Git Gateway on Cloud Run with minimal configuration.

## Method 1: single repository
This requires a PAT (Personal Access Token) or Deploy Key.

## Method 2: organization-wide or public
This method requires you to setup a GitHub App. Follow the documentation of https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app and fill in the parameters that are displayed on the setup page on your Cloud Run Git Gateway.

After creating the app, it can be installed by navigating to:

```markdown
https://github.com/apps/<app name>/installations/new
```