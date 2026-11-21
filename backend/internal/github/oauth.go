package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"
)

type User struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

type OAuth struct {
	config *oauth2.Config
}

func NewOAuth(clientID, clientSecret, redirectURL string) *OAuth {
	return &OAuth{config: &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     oauthgithub.Endpoint,
	}}
}

func (o *OAuth) LoginURL(state string) string {
	return o.config.AuthCodeURL(state)
}

// UserFromCode exchanges the code GitHub sent to the callback for a token and returns the user it belongs to.
func (o *OAuth) UserFromCode(ctx context.Context, code string) (User, error) {
	token, err := o.config.Exchange(ctx, code)
	if err != nil {
		return User{}, fmt.Errorf("exchange code: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := o.config.Client(ctx, token).Do(req)
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("get user: status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return User{}, fmt.Errorf("decode user: %w", err)
	}
	return user, nil
}
