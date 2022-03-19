# git-gateway - Gateway to hosted git APIs

[![Run on Google
Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run/?git_repo=https://github.com/hermanbanken/git-gateway.git&revision=cloudrun&dir=cloudrun)

**Secure role based access to the APIs of common Git Hosting providers.**

When building sites with a JAMstack approach, a common pattern is to store all content as structured data in a Git repository instead of relying on an external database. Git-Gateway in single-tenant mode requires no database. And multi-tenant mode only requires it to store configuration.

Netlify CMS is an open-source content management UI that allows content editors to work with your content in Git through a familiar content editing interface. This allows people to write and edit content without having to write code or know anything about Git, markdown, YAML, JSON, etc.

However, for most use cases you won’t want to require all content editors to have an account with full access to the source code repository for your website.

Netlify’s Git Gateway lets you set up a gateway to your choice of Git provider's API (currently available with both GitHub and GitLab 🎉 ) that lets tools like Netlify CMS work with content, branches and pull requests on your users’ behalf.

The Git Gateway works with any identity service that can issue JWTs and only allows access when a JSON Web Token with sufficient permissions is present.

To configure the gateway, see our `example.env` file

The Gateway limits access to the following sub endpoints of the repository:

for GitHub:
```
   /repos/:owner/:name/git/
   /repos/:owner/:name/contents/
   /repos/:owner/:name/pulls/
   /repos/:owner/:name/branches/
   /repos/:owner/:name/merges/
   /repos/:owner/:name/statuses/
   /repos/:owner/:name/compare/
   /repos/:owner/:name/commits/
   /repos/:owner/:name/issues/<number>/labels
```
for GitLab:
```
   /projects/:owner/:name/merge_requests/
   /projects/:owner/:name/repository/files/
   /projects/:owner/:name/repository/commits/
   /projects/:owner/:name/repository/tree/
   /projects/:owner/:name/repository/compare/
   /projects/:owner/:name/repository/branches/
```

# Single-Tenant mode
This mode just serves access to a single git repository.

# Multi-Tenant mode
This mode serves access to multiple git repositories, so it can be used by your whole organization or be hosted as a service to others (like how Netlify offers it).

Authorization of operators happens through the `X-NF-Sign` header, which needs to be a JWT signed using the `$GITGATEWAY_OPERATOR_TOKEN` secret. To generate one easily, run:

```bash
brew install mike-engel/jwt-cli/jwt-cli # very nice jwt tool from https://github.com/mike-engel/jwt-cli
export GITGATEWAY_DB_AUTOMIGRATE=1 # auto-create the sqlite tables
./git-gateway multi & # start the gateway
source .env # reads $GITGATEWAY_OPERATOR_TOKEN

# Create instance
export INSTANCE_SECRET=foobar
export INSTANCE_ID=$(curl localhost:9999/instances -H "Authorization: Bearer $GITGATEWAY_OPERATOR_TOKEN" -X POST -d '{ "uuid": "5", "config": { "jwt": { "secret": "'$INSTANCE_SECRET'" }, "github": {} } }' -sS | tee -a /dev/stderr | jq -r .id)

# Prepare being instance operator
export GITGATEWAY_OPERATOR_JWT_DATA='{"id": "'$INSTANCE_ID'"}' # set the instance id
export GITGATEWAY_OPERATOR_SIGN=$(jwt encode --secret=$GITGATEWAY_OPERATOR_TOKEN $GITGATEWAY_OPERATOR_JWT_DATA -e '2m')

# Inspect instance
curl localhost:9999/settings -H "X-NF-Sign: $(jwt encode --secret=$GITGATEWAY_OPERATOR_TOKEN $JWT_DATA -e '2m')"
curl localhost:9999/settings \
  -H "Authorization: Bearer $(jwt encode --secret=$INSTANCE_SECRET)" \
  -H "X-NF-Sign: $GITGATEWAY_OPERATOR_SIGN"
```

Some important extra endpoints:

- `POST /instances` - creates a new instance (params: `InstanceRequestParams`)
- `GET /instances/{instance_id}` - inspect instance (params: `InstanceRequestParams`)
- `PUT /instances/{instance_id}` - edit instance (params: `InstanceRequestParams`)
- `DELETE /instances/{instance_id}` - delete instance
