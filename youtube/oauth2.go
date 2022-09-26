package youtube

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(scope, clientSecret, appCreds string) *http.Client {
	ctx := context.Background()

	config, err := google.ConfigFromJSON([]byte(clientSecret), scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	t := &oauth2.Token{}
	if err = json.Unmarshal([]byte(appCreds), t); err != nil {
		log.Fatalf("cannot parse token: %q %v", appCreds, err)
	}

	return config.Client(ctx, t)
}
