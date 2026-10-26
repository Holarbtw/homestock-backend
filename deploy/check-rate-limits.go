package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type rateZone struct {
	Key       string                `json:"key"`
	Window    time.Duration         `json:"window"`
	MaxEvents int                   `json:"max_events"`
	Match     []map[string][]string `json:"match"`
}

type ratePolicy struct {
	name, key string
	window    time.Duration
	events    int
	requests  []string
}

func dockerCommand(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return output, nil
}

func collectZones(value any, zones map[string]json.RawMessage) error {
	switch value := value.(type) {
	case map[string]any:
		if value["handler"] == "rate_limit" {
			for name, zone := range value["rate_limits"].(map[string]any) {
				if _, exists := zones[name]; exists {
					return fmt.Errorf("duplicate rate limit zone %q", name)
				}
				encoded, err := json.Marshal(zone)
				if err != nil {
					return err
				}
				zones[name] = encoded
			}
		}
		for _, child := range value {
			if err := collectZones(child, zones); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range value {
			if err := collectZones(child, zones); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkPolicy(encoded json.RawMessage, policy ratePolicy) error {
	var zone rateZone
	if err := json.Unmarshal(encoded, &zone); err != nil {
		return err
	}
	if zone.Key != policy.key || zone.Window != policy.window || zone.MaxEvents != policy.events {
		return fmt.Errorf("%s: got key=%q window=%s events=%d", policy.name, zone.Key, zone.Window, zone.MaxEvents)
	}
	var requests []string
	for _, matcher := range zone.Match {
		if len(matcher) != 2 || len(matcher["method"]) == 0 || len(matcher["path"]) == 0 {
			return fmt.Errorf("%s: expected explicit method and path matchers", policy.name)
		}
		for _, method := range matcher["method"] {
			for _, path := range matcher["path"] {
				requests = append(requests, method+" "+path)
			}
		}
	}
	sort.Strings(requests)
	expected := append([]string(nil), policy.requests...)
	sort.Strings(expected)
	if !reflect.DeepEqual(requests, expected) {
		return fmt.Errorf("%s: requests=%v, want %v", policy.name, requests, expected)
	}
	return nil
}

func probe(client *http.Client, base, request string, counter int) (int, http.Header, error) {
	method, path, _ := strings.Cut(request, " ")
	req, err := http.NewRequest(method, base+path, strings.NewReader("{"))
	if err != nil {
		return 0, nil, err
	}
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP", "True-Client-IP"} {
		req.Header.Set(header, fmt.Sprintf("192.0.2.%d", counter%254+1))
	}
	response, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, response.Body)
	return response.StatusCode, response.Header, err
}

// Each probe runs an unchanged production zone by itself. This isolates long
// windows from the short burst limits without waiting a minute, hour or day.
// The upstream is a static handler: these checks never create database records.
func checkBoundary(image string, encoded json.RawMessage, policy ratePolicy) (resultErr error) {
	config := map[string]any{
		"admin": map[string]any{"disabled": true},
		"apps": map[string]any{"http": map[string]any{"servers": map[string]any{
			"probe": map[string]any{
				"listen": []string{":80"},
				"routes": []any{map[string]any{"handle": []any{
					map[string]any{"handler": "rate_limit", "rate_limits": map[string]any{policy.name: encoded}},
					map[string]any{"handler": "static_response", "status_code": 204},
				}}},
			},
		}}},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp("", "homestock-rate-limit-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(configBytes)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	name := "homestock-rate-limit-check-" + hex.EncodeToString(random[:])
	if _, err := dockerCommand("run", "-d", "--name", name, "-p", "127.0.0.1::80", "--mount", "type=bind,source="+file.Name()+",target=/etc/caddy/probe.json,readonly", image, "caddy", "run", "--config", "/etc/caddy/probe.json"); err != nil {
		_, _ = dockerCommand("rm", "-f", name)
		return err
	}
	defer func() {
		if _, err := dockerCommand("rm", "-f", name); err != nil && resultErr == nil {
			resultErr = err
		}
	}()
	port, err := dockerCommand("port", name, "80/tcp")
	if err != nil {
		return err
	}
	base := "http://" + strings.TrimSpace(string(port))
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(15 * time.Second)
	for {
		status, _, err := probe(client, base, "GET /healthz", 0)
		if err == nil && status == 204 {
			break
		}
		if time.Now().After(deadline) {
			logs, _ := dockerCommand("logs", name)
			return fmt.Errorf("%s did not become ready: %s", policy.name, logs)
		}
		time.Sleep(100 * time.Millisecond)
	}
	outside := []string{"GET /api/v1/auth/register", "GET /api/v1/auth/refresh", "POST /api/v1/auth/login", "POST /api/v1/auth/logout", "GET /api/v1/auth/profile"}
	for i := 0; i < policy.events; i++ {
		status, _, err := probe(client, base, policy.requests[i%len(policy.requests)], i)
		if err != nil || status != 204 {
			return fmt.Errorf("%s request %d: status=%d error=%v, want 204", policy.name, i+1, status, err)
		}
	}
	status, headers, err := probe(client, base, policy.requests[policy.events%len(policy.requests)], policy.events)
	if err != nil || status != http.StatusTooManyRequests {
		return fmt.Errorf("%s exceeded limit: status=%d error=%v, want 429", policy.name, status, err)
	}
	retry, err := strconv.Atoi(headers.Get("Retry-After"))
	if err != nil || retry < 1 || retry > int(policy.window.Seconds())+1 {
		return fmt.Errorf("%s invalid Retry-After %q", policy.name, headers.Get("Retry-After"))
	}
	for _, request := range outside {
		status, _, err := probe(client, base, request, 0)
		if err != nil || status != 204 {
			return fmt.Errorf("%s incorrectly limited %s: status=%d error=%v", policy.name, request, status, err)
		}
	}
	fmt.Printf("Passed %s: %d allowed, next request 429, Retry-After and route/IP boundaries.\n", policy.name, policy.events)
	return nil
}

func check(image, configPath string) error {
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}
	adapted, err := dockerCommand("run", "--rm", "-e", "API_DOMAIN=http://localhost", "--mount", "type=bind,source="+configPath+",target=/etc/caddy/Caddyfile,readonly", image, "caddy", "adapt", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile")
	if err != nil {
		return err
	}
	var document any
	if err := json.Unmarshal(adapted, &document); err != nil {
		return err
	}
	zones := make(map[string]json.RawMessage)
	if err := collectZones(document, zones); err != nil {
		return err
	}
	registration := []string{"POST /api/v1/auth/register", "POST /api/v1/auth/guest-login"}
	refresh := []string{"POST /api/v1/auth/refresh"}
	auth := append(append([]string(nil), registration...), "POST /api/v1/auth/login", "POST /api/v1/auth/link-guest", "PATCH /api/v1/auth/password", "PATCH /api/v1/auth/email")
	policies := []ratePolicy{
		{"auth_global", "auth", time.Second, 10, append(append([]string(nil), auth...), refresh...)},
		{"auth_per_ip", "{http.request.remote.host}", 5 * time.Second, 5, auth},
		{"registration_per_ip", "{http.request.remote.host}", time.Minute, 5, registration},
		{"registration_global_hour", "registration", time.Hour, 100, registration},
		{"registration_global_day", "registration", 24 * time.Hour, 500, registration},
		{"refresh_per_ip", "{http.request.remote.host}", time.Minute, 20, refresh},
		{"refresh_global", "refresh", time.Minute, 100, refresh},
	}
	if len(zones) != len(policies) {
		return fmt.Errorf("got %d rate limit zones, want %d", len(zones), len(policies))
	}
	for _, policy := range policies {
		if err := checkPolicy(zones[policy.name], policy); err != nil {
			return err
		}
		if policy.name != "auth_global" && policy.name != "auth_per_ip" {
			if err := checkBoundary(image, zones[policy.name], policy); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run deploy/check-rate-limits.go GATEWAY_IMAGE CADDYFILE")
		os.Exit(1)
	}
	if err := check(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "Gateway rate limit check failed:", err)
		os.Exit(1)
	}
}
