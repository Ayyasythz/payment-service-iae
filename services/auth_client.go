package services

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"payment-service-iae/config"
	"time"
)

type AuthClient struct {
	baseURL string
	client  *http.Client
}

type ValidationResponse struct {
	Valid       bool      `json:"valid"`
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
}

type PermissionCheckResponse struct {
	HasPermission bool      `json:"has_permission"`
	UserID        uuid.UUID `json:"user_id"`
	Role          string    `json:"role"`
}

func NewAuthClient(cfg *config.Config) *AuthClient {
	return &AuthClient{
		baseURL: cfg.AuthServiceURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *AuthClient) ValidateToken(token string) (*ValidationResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/auth/validate", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var validation ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return nil, err
	}

	return &validation, nil
}

func (c *AuthClient) CheckPermission(token, resource, action string) (*PermissionCheckResponse, error) {
	reqURL := fmt.Sprintf("%s/api/v1/auth/check-permission?resource=%s&action=%s",
		c.baseURL, url.QueryEscape(resource), url.QueryEscape(action))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var permissionCheck PermissionCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&permissionCheck); err != nil {
		return nil, err
	}

	return &permissionCheck, nil
}
