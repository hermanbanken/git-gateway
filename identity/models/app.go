package models

// App gets auto-filled from manifest flow: // https://docs.github.com/en/rest/reference/apps#create-a-github-app-from-a-manifest
type App struct {
	// Github App ID
	ID int64 `json:"id" firestore:"id"`
	// GitHub App name is used in the installation url
	Name string `json:"name" firestore:"name"`
	// GitHub App private key
	PEM []byte `json:"pem" firestore:"pem"`
	// GitHub App webhook secret
	WebhookSecret string `json:"webhook_secret" firestore:"webhook_secret"`
	// Owner
	Owner AppOwner `json:"owner" firestore:"owner"`
	// ClientID
	ClientID string `json:"client_id" firestore:"client_id"`
	// ClientSecret
	ClientSecret string `json:"client_secret" firestore:"client_secret"`
}

type AppOwner struct {
	// username / org name
	Login string `json:"login" firestore:"login"`
	// User/Organization
	Type string `json:"type" firestore:"type"`
}

type Installation struct {
	// Github App ID
	AppID string `json:"app_id" firestore:"app_id"`
	// GitHub App Installation ID
	InstallationID int64 `json:"id" firestore:"id"`
}

// NotFoundError represents when an object is not found.
type NotFoundError struct {
	Table string
	ID    string
}

func (e NotFoundError) Error() string {
	return "Not found"
}
