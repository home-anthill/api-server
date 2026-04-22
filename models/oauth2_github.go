package models

// GithubAccessTokenResponse is the response body for the GitHub access token endpoint.
type GithubAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	Description string `json:"error_description"`
	ErrorURI    string `json:"error_uri"`
}

// GithubAppCurrentUserResponse is the response body for the GitHub user endpoint.
type GithubAppCurrentUserResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}
