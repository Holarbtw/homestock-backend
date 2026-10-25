package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type account struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Timezone string `json:"timezone"`
	UserID   string `json:"user_id,omitempty"`
}

type user struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         user   `json:"user"`
}

type api struct {
	base   string
	client *http.Client
}

func (a api) request(method, path string, expected int, data any, token string, result any) error {
	var body bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&body).Encode(data); err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, a.base+path, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != expected {
		return fmt.Errorf("%s %s: HTTP %d, expected %d", method, path, resp.StatusCode, expected)
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func randomHex() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func check(baseURL, stateFile string, reuse bool) error {
	a := api{strings.TrimRight(baseURL, "/") + "/api/v1/auth", &http.Client{Timeout: 15 * time.Second}}
	if err := a.request("GET", "/healthz", 200, nil, "", nil); err != nil {
		return err
	}
	var swagger struct {
		Host     string `json:"host"`
		BasePath string `json:"basePath"`
	}
	if err := a.request("GET", "/swagger/doc.json", 200, nil, "", &swagger); err != nil {
		return err
	}
	if swagger.Host != "" || swagger.BasePath != "/api/v1/auth" {
		return fmt.Errorf("Swagger must use the current host and /api/v1/auth")
	}

	var saved account
	if reuse {
		data, err := os.ReadFile(stateFile)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &saved); err != nil {
			return err
		}
	} else {
		unique, err := randomHex()
		if err != nil {
			return err
		}
		password, err := randomHex()
		if err != nil {
			return err
		}
		saved = account{Email: "ci-" + unique + "@example.com", Password: password, Nickname: "CI check", Timezone: "UTC"}
		var registered tokens
		if err := a.request("POST", "/register", 201, saved, "", &registered); err != nil {
			return err
		}
		saved.UserID = registered.User.ID
		data, err := json.Marshal(saved)
		if err != nil {
			return err
		}
		if err := os.WriteFile(stateFile, data, 0600); err != nil {
			return err
		}
	}

	var loggedIn tokens
	credentials := map[string]string{"email": saved.Email, "password": saved.Password}
	if err := a.request("POST", "/login", 200, credentials, "", &loggedIn); err != nil {
		return err
	}
	var profile user
	if err := a.request("GET", "/profile", 200, nil, loggedIn.AccessToken, &profile); err != nil {
		return err
	}
	if profile.ID != saved.UserID || profile.Email != saved.Email {
		return fmt.Errorf("the account did not survive deployment")
	}
	var refreshed tokens
	if err := a.request("POST", "/refresh", 200, map[string]string{"refresh_token": loggedIn.RefreshToken}, "", &refreshed); err != nil {
		return err
	}
	if err := a.request("GET", "/profile", 200, nil, refreshed.AccessToken, &profile); err != nil {
		return err
	}
	return a.request("POST", "/logout", 204, map[string]string{"refresh_token": refreshed.RefreshToken}, "", nil)
}

func main() {
	reuse := flag.Bool("reuse", false, "log in to the account saved before redeployment")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run deploy/smoke-auth.go [-reuse] BASE_URL STATE_FILE")
		os.Exit(1)
	}
	if err := check(flag.Arg(0), flag.Arg(1), *reuse); err != nil {
		fmt.Fprintln(os.Stderr, "Auth smoke check failed:", err)
		os.Exit(1)
	}
	fmt.Println("Auth smoke check passed: Swagger, register/login, profile, refresh, logout.")
}
