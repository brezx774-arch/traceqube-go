package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// setAuthCookie sets a shared cookie across *.traceqube.com so the homepage
// can check login status via /api/me.
func setAuthCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "tq_token",
		Value:    token,
		Domain:   ".traceqube.com",
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})
}

// Me checks the tq_token cookie and returns basic login status/info.
func Me(c *fiber.Ctx) error {
	tokenStr := c.Cookies("tq_token")
	if tokenStr == "" {
		return c.JSON(fiber.Map{"loggedIn": false})
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.JSON(fiber.Map{"loggedIn": false})
	}
	email, _ := claims["email"].(string)
	return c.JSON(fiber.Map{"loggedIn": true, "email": email})
}

func googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "https://api.traceqube.com/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func githubOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "https://api.traceqube.com/auth/github/callback",
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
}

func findOrCreateOAuthUser(provider, oauthID, email string) (User, error) {
	var u User
	ctx := context.Background()

	err := DB.QueryRow(ctx,
		`SELECT id, email, api_key, plan, usage, verified, is_admin, token_version
		 FROM users WHERE oauth_provider = $1 AND oauth_id = $2`,
		provider, oauthID).Scan(&u.ID, &u.Email, &u.APIKey, &u.Plan, &u.Usage, &u.Verified, &u.IsAdmin, &u.TokenVersion)
	if err == nil {
		return u, nil
	}

	err = DB.QueryRow(ctx,
		`SELECT id, email, api_key, plan, usage, verified, is_admin, token_version
		 FROM users WHERE email = $1`, email).Scan(&u.ID, &u.Email, &u.APIKey, &u.Plan, &u.Usage, &u.Verified, &u.IsAdmin, &u.TokenVersion)
	if err == nil {
		_, linkErr := DB.Exec(ctx,
			`UPDATE users SET oauth_provider = $1, oauth_id = $2, verified = true WHERE id = $3`,
			provider, oauthID, u.ID)
		if linkErr != nil {
			return u, linkErr
		}
		u.Verified = true
		return u, nil
	}

	apiKey := generateAPIKey()
	err = DB.QueryRow(ctx,
		`INSERT INTO users (email, api_key, plan, usage, verified, oauth_provider, oauth_id)
		 VALUES ($1, $2, 'free', 0, true, $3, $4)
		 RETURNING id, email, api_key, plan, usage, verified, is_admin`,
		email, apiKey, provider, oauthID).Scan(&u.ID, &u.Email, &u.APIKey, &u.Plan, &u.Usage, &u.Verified, &u.IsAdmin)
	if err != nil {
		return u, err
	}
	return u, nil
}

func completeOAuthLogin(c *fiber.Ctx, u User) error {
	token, err := generateJWT(u.ID, u.Email, u.TokenVersion)
	if err != nil {
		return c.Status(500).SendString("Failed to generate token")
	}
	setAuthCookie(c, token)
	return c.Redirect("https://app.traceqube.com/oauth-callback#token=" + token)
}

// --- Google ---

func GoogleLogin(c *fiber.Ctx) error {
	url := googleOAuthConfig().AuthCodeURL("state", oauth2.AccessTypeOnline)
	return c.Redirect(url)
}

func GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(400).SendString("Missing code")
	}
	conf := googleOAuthConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(500).SendString("OAuth exchange failed: " + err.Error())
	}

	client := conf.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(500).SendString("Failed to fetch user info")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &info); err != nil || info.Email == "" {
		return c.Status(500).SendString("Failed to parse user info")
	}

	u, err := findOrCreateOAuthUser("google", info.ID, info.Email)
	if err != nil {
		return c.Status(500).SendString("Failed to create/find user: " + err.Error())
	}
	return completeOAuthLogin(c, u)
}

// --- GitHub ---

func GithubLogin(c *fiber.Ctx) error {
	url := githubOAuthConfig().AuthCodeURL("state", oauth2.AccessTypeOnline)
	return c.Redirect(url)
}

func GithubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(400).SendString("Missing code")
	}
	conf := githubOAuthConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(500).SendString("OAuth exchange failed: " + err.Error())
	}

	client := conf.Client(context.Background(), token)

	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch user info")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
		Login string `json:"login"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return c.Status(500).SendString("Failed to parse user info")
	}

	email := info.Email
	if email == "" {
		req2, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		resp2, err := client.Do(req2)
		if err == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if json.Unmarshal(body2, &emails) == nil {
				for _, e := range emails {
					if e.Primary && e.Verified {
						email = e.Email
						break
					}
				}
				if email == "" {
					for _, e := range emails {
						if e.Verified {
							email = e.Email
							break
						}
					}
				}
			}
		}
	}

	if email == "" {
		return c.Status(400).SendString("Could not retrieve a verified email from GitHub. Please make your email public or verify an email on your GitHub account.")
	}

	oauthID := fmt.Sprintf("%d", info.ID)
	u, err := findOrCreateOAuthUser("github", oauthID, email)
	if err != nil {
		return c.Status(500).SendString("Failed to create/find user: " + err.Error())
	}
	return completeOAuthLogin(c, u)
}
