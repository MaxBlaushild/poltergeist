package hue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	hueCloudAPIBaseURL = "https://api.meethue.com/route/clip/v2"
	tokenRefreshBuffer = 5 * time.Minute // Refresh token 5 minutes before expiry
)

// TokenUpdateCallback is called when tokens are refreshed, allowing persistence
type TokenUpdateCallback func(accessToken, refreshToken string, expiresAt time.Time) error

type cloudClient struct {
	oauthClient  OAuthClient
	refreshToken string
	httpClient   *http.Client
	tokenUpdater TokenUpdateCallback // Optional callback to persist token updates

	// Token cache with mutex for thread safety
	tokenMu     sync.RWMutex
	accessToken string
	tokenExpiry time.Time

	// ID mapping: integer ID -> UUID
	idMapMu     sync.RWMutex
	idToUUIDMap map[int]string

	// Bridge ID cache
	bridgeMu sync.RWMutex
	bridgeID string

	hueApplicationKey string
}

// NewClientWithOAuth creates a new Hue client that uses OAuth authentication with the cloud API
// If accessToken is provided and not expired, it will be used; otherwise the refreshToken will be used to get a new access token
// tokenUpdater is an optional callback that will be invoked when tokens are refreshed, allowing persistence to a database
func NewClientWithOAuth(oauthClient OAuthClient, refreshToken string, accessToken string, expiresAt time.Time, tokenUpdater TokenUpdateCallback, hueApplicationKey string) Client {
	client := &cloudClient{
		oauthClient:       oauthClient,
		refreshToken:      refreshToken,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		idToUUIDMap:       make(map[int]string),
		tokenUpdater:      tokenUpdater,
		hueApplicationKey: hueApplicationKey,
	}

	// Initialize with existing access token if provided and still valid
	if accessToken != "" && time.Now().Before(expiresAt) {
		client.tokenMu.Lock()
		client.accessToken = accessToken
		client.tokenExpiry = expiresAt
		client.tokenMu.Unlock()
	}

	return client
}

// ensureValidToken ensures we have a valid access token, refreshing if necessary
func (c *cloudClient) ensureValidToken(ctx context.Context) error {
	c.tokenMu.RLock()
	hasToken := c.accessToken != ""
	tokenExpired := !hasToken || c.tokenExpiry.IsZero() || time.Now().Add(tokenRefreshBuffer).After(c.tokenExpiry)
	c.tokenMu.RUnlock()

	if !tokenExpired {
		return nil
	}

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && !c.tokenExpiry.IsZero() && time.Now().Add(tokenRefreshBuffer).Before(c.tokenExpiry) {
		return nil
	}

	if c.refreshToken == "" {
		return fmt.Errorf("cannot refresh access token: refresh token is missing")
	}

	// Validate refresh token format
	if len(c.refreshToken) < 10 {
		return fmt.Errorf("refresh token appears to be invalid (too short)")
	}

	tokenResp, err := c.oauthClient.RefreshAccessToken(ctx, c.refreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	// Validate the new token
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("received empty access token from refresh")
	}
	if tokenResp.ExpiresAt.IsZero() {
		return fmt.Errorf("received token with zero expiry time")
	}

	// Update tokens
	oldRefreshToken := c.refreshToken
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = tokenResp.ExpiresAt
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken // Update refresh token if a new one is provided
	}

	// Persist token update if callback is provided
	if c.tokenUpdater != nil {
		if err := c.tokenUpdater(tokenResp.AccessToken, c.refreshToken, tokenResp.ExpiresAt); err != nil {
			log.Printf("[Hue] Warning: Failed to persist token update: %v", err)
			// Don't fail the request if persistence fails, but log it
		} else {
			log.Printf("[Hue] Token update persisted successfully")
		}
	}

	log.Printf("[Hue] Successfully refreshed access token (expires at: %v, refresh token updated: %v)",
		tokenResp.ExpiresAt, oldRefreshToken != c.refreshToken)
	return nil
}

// getAccessToken returns the current access token, refreshing if needed
func (c *cloudClient) getAccessToken(ctx context.Context) (string, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return "", err
	}

	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.accessToken, nil
}

// doRequest makes an authenticated HTTP request to the Hue cloud API
func (c *cloudClient) doRequest(ctx context.Context, method, endpoint string, requestBody interface{}) ([]byte, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	var bodyReader io.Reader
	if requestBody != nil {
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
		log.Printf("[Hue] Request body: %s", string(jsonBody))
	}

	url := fmt.Sprintf("%s%s", hueCloudAPIBaseURL, endpoint)
	log.Printf("[Hue] Making %s request to: %s", method, url)

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Log token info (masked for security)
	tokenPreview := accessToken
	if len(tokenPreview) > 20 {
		tokenPreview = tokenPreview[:10] + "..." + tokenPreview[len(tokenPreview)-10:]
	}
	// Set authorization header - Hue Cloud API expects Bearer token
	req.Header.Set("hue-application-key", c.hueApplicationKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Hue] Request failed: %v", err)
		return nil, fmt.Errorf("error sending request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Handle authentication errors - both 401 and 403 can indicate token issues
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		errorType := "Unauthorized"
		if resp.StatusCode == 403 {
			errorType = "Forbidden"
		}
		log.Printf("[Hue] Got %d %s, token may be expired or invalid. Clearing token and retrying after refresh...", resp.StatusCode, errorType)

		// Clear the current token so it will be refreshed
		c.tokenMu.Lock()
		c.accessToken = ""
		c.tokenExpiry = time.Time{}
		c.tokenMu.Unlock()

		// Refresh token and retry
		newAccessToken, err := c.getAccessToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token after %d: %w", resp.StatusCode, err)
		}

		// Recreate the request body for retry
		var retryBodyReader io.Reader
		if requestBody != nil {
			jsonBody, err := json.Marshal(requestBody)
			if err != nil {
				return nil, fmt.Errorf("error marshaling request body for retry: %w", err)
			}
			retryBodyReader = bytes.NewBuffer(jsonBody)
		}

		// Recreate the request with new token
		retryReq, err := http.NewRequestWithContext(ctx, method, url, retryBodyReader)
		if err != nil {
			return nil, fmt.Errorf("error creating retry request: %w", err)
		}

		retryReq.Header.Set("hue-application-key", c.hueApplicationKey)
		retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newAccessToken))
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("Accept", "application/json")

		retryResp, err := c.httpClient.Do(retryReq)
		if err != nil {
			return nil, fmt.Errorf("error retrying request after token refresh: %w", err)
		}
		defer retryResp.Body.Close()

		retryBody, err := io.ReadAll(retryResp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading retry response body: %w", err)
		}

		// Check retry response
		if retryResp.StatusCode >= 400 {
			// If we still get 401/403 after refresh, the refresh token itself might be invalid
			if retryResp.StatusCode == 401 || retryResp.StatusCode == 403 {
				return nil, fmt.Errorf("Hue API returned %d after token refresh - refresh token may be invalid or expired. Please re-authenticate.", retryResp.StatusCode)
			}

			return nil, fmt.Errorf("Hue API returned error status %d for endpoint %s after token refresh.", retryResp.StatusCode, endpoint)
		}

		// Success on retry
		log.Printf("[Hue] Retry successful after token refresh")
		return retryBody, nil
	}

	if resp.StatusCode >= 400 {
		// Check if response is JSON or HTML/other format
		contentType := resp.Header.Get("Content-Type")
		log.Printf("[Hue] Error response Content-Type: %s", contentType)

		// Try to parse error response as JSON
		if strings.Contains(contentType, "application/json") {
			var errorResp struct {
				Fault struct {
					FaultString string `json:"faultstring"`
					Detail      struct {
						ErrorCode string `json:"errorcode"`
					} `json:"detail"`
				} `json:"fault"`
				Errors []struct {
					Description string `json:"description"`
				} `json:"errors"`
			}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				if errorResp.Fault.FaultString != "" {
					return nil, fmt.Errorf("Hue API error: %s (code: %s, status: %d, endpoint: %s)", errorResp.Fault.FaultString, errorResp.Fault.Detail.ErrorCode, resp.StatusCode, endpoint)
				}
				if len(errorResp.Errors) > 0 {
					return nil, fmt.Errorf("Hue API error: %s (status: %d, endpoint: %s)", errorResp.Errors[0].Description, resp.StatusCode, endpoint)
				}
			}
		}
		return nil, fmt.Errorf("Hue API returned error status %d for endpoint %s. Response: %s", resp.StatusCode, endpoint, string(body))
	}

	// Check if response is actually JSON before returning
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		log.Printf("[Hue] Warning: Response Content-Type is not JSON: %s", contentType)
		// Continue anyway, let the caller handle JSON unmarshaling errors
	}

	return body, nil
}

// Bridge-specific methods that don't apply to cloud API
func (c *cloudClient) DiscoverBridge(ctx context.Context) (*Bridge, error) {
	return nil, fmt.Errorf("bridge discovery is not supported for cloud API client")
}

func (c *cloudClient) Connect(ctx context.Context, hostname, username string) error {
	// No-op for cloud client - connection is handled via OAuth
	return nil
}

func (c *cloudClient) CreateUser(ctx context.Context, deviceType string) (string, error) {
	return "", fmt.Errorf("creating bridge users is not supported for cloud API client")
}

// Cloud API response structures
type cloudLightResponse struct {
	ID    string `json:"id"`
	IDV1  string `json:"id_v1,omitempty"`
	Owner struct {
		Rid   string `json:"rid"`
		RType string `json:"rtype"`
	} `json:"owner,omitempty"`
	Metadata struct {
		Name      string `json:"name"`
		Archetype string `json:"archetype,omitempty"`
	} `json:"metadata"`
	On struct {
		On bool `json:"on"`
	} `json:"on"`
	Dimming struct {
		Brightness float64 `json:"brightness,omitempty"`
	} `json:"dimming,omitempty"`
	Color struct {
		XY struct {
			X float32 `json:"x"`
			Y float32 `json:"y"`
		} `json:"xy,omitempty"`
	} `json:"color,omitempty"`
	ColorTemperature struct {
		Mirek int `json:"mirek,omitempty"`
	} `json:"color_temperature,omitempty"`
}

type cloudLightsResponse struct {
	Data []cloudLightResponse `json:"data"`
}

// GetLights returns all lights from the cloud API
func (c *cloudClient) GetLights(ctx context.Context) ([]*Light, error) {
	// Try different endpoint patterns
	endpoint := "/resource/light"

	var body []byte
	body, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get lights from API: %w", err)
	}

	// Check if body looks like JSON before attempting to unmarshal
	if len(body) > 0 && body[0] != '{' && body[0] != '[' {
		log.Printf("[Hue] Response does not appear to be JSON (starts with: %q). Full body: %s", string(body[0]), string(body))
		return nil, fmt.Errorf("unexpected response format (expected JSON, got: %s)", string(body[:min(len(body), 100)]))
	}

	var response cloudLightsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("[Hue] Failed to unmarshal JSON response: %v", err)
		log.Printf("[Hue] Response body was: %s", string(body))
		return nil, fmt.Errorf("error unmarshaling lights response (status may have been 200 but body is not valid JSON): %w. Response preview: %s", err, string(body[:min(len(body), 500)]))
	}

	lights := make([]*Light, 0, len(response.Data))
	for _, cloudLight := range response.Data {
		light := c.convertCloudLight(cloudLight)
		lights = append(lights, light)
	}

	return lights, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetLight returns a specific light by ID from the cloud API
// Since cloud API uses UUIDs, we query all lights and find the matching integer ID
func (c *cloudClient) GetLight(ctx context.Context, id int) (*Light, error) {
	lights, err := c.GetLights(ctx)
	if err != nil {
		return nil, err
	}

	for _, light := range lights {
		if light.ID == id {
			return light, nil
		}
	}

	return nil, fmt.Errorf("light %d not found", id)
}

// convertCloudLight converts a cloud API light response to our Light type
func (c *cloudClient) convertCloudLight(cloudLight cloudLightResponse) *Light {
	light := &Light{
		Name:      cloudLight.Metadata.Name,
		On:        cloudLight.On.On,
		ColorMode: "xy", // Default
		Reachable: true, // Cloud API doesn't provide reachability, assume true
	}

	// Extract numeric ID from v1 ID format like "/lights/5"
	var id int
	if cloudLight.IDV1 != "" {
		if _, err := fmt.Sscanf(cloudLight.IDV1, "/lights/%d", &id); err == nil {
			light.ID = id
		}
	}

	// If we couldn't parse ID from v1, try to hash the UUID to get a consistent integer
	// This is a fallback for lights that don't have v1 IDs
	if light.ID == 0 && cloudLight.ID != "" {
		// Simple hash of UUID to get an integer (not ideal but works for compatibility)
		hash := 0
		for _, char := range cloudLight.ID {
			hash = hash*31 + int(char)
			if hash < 0 {
				hash = -hash
			}
		}
		light.ID = hash % 1000000 // Keep it reasonable
	}

	// Store UUID mapping for updates
	if cloudLight.ID != "" && light.ID != 0 {
		c.idMapMu.Lock()
		c.idToUUIDMap[light.ID] = cloudLight.ID
		c.idMapMu.Unlock()
	}

	// Convert brightness (0-100 float) to uint8 (0-254)
	if cloudLight.Dimming.Brightness >= 0 {
		brightness := uint8((cloudLight.Dimming.Brightness / 100.0) * 254.0)
		if brightness > 254 {
			brightness = 254
		}
		light.Brightness = brightness
	}

	// Set XY color coordinates
	if cloudLight.Color.XY.X > 0 || cloudLight.Color.XY.Y > 0 {
		light.ColorXY = [2]float32{cloudLight.Color.XY.X, cloudLight.Color.XY.Y}
		// Convert XY to RGB
		r, g, b := xyToRGB(float64(cloudLight.Color.XY.X), float64(cloudLight.Color.XY.Y), light.Brightness)
		light.ColorRGB = [3]uint8{r, g, b}
	}

	// Set color temperature
	if cloudLight.ColorTemperature.Mirek > 0 {
		light.Temperature = uint16(cloudLight.ColorTemperature.Mirek)
	}

	return light
}

// TurnOn turns on a light by ID
func (c *cloudClient) TurnOn(ctx context.Context, id int) error {
	state := map[string]interface{}{
		"on": map[string]bool{
			"on": true,
		},
	}
	return c.updateLightState(ctx, id, state)
}

// TurnOff turns off a light by ID
func (c *cloudClient) TurnOff(ctx context.Context, id int) error {
	state := map[string]interface{}{
		"on": map[string]bool{
			"on": false,
		},
	}
	return c.updateLightState(ctx, id, state)
}

// SetBrightness sets the brightness of a light (0-254)
func (c *cloudClient) SetBrightness(ctx context.Context, id int, brightness uint8) error {
	if brightness > 254 {
		brightness = 254
	}

	// Convert 0-254 to 0-100 float for cloud API
	brightnessPercent := float64(brightness) / 254.0 * 100.0

	state := map[string]interface{}{
		"dimming": map[string]float64{
			"brightness": brightnessPercent,
		},
	}
	return c.updateLightState(ctx, id, state)
}

// SetColorRGB sets the color of a light using RGB values (0-255)
func (c *cloudClient) SetColorRGB(ctx context.Context, id int, r, g, b uint8) error {
	// Convert RGB to XY coordinates
	x, y := rgbToXY(float64(r), float64(g), float64(b))

	state := map[string]interface{}{
		"color": map[string]interface{}{
			"xy": map[string]float32{
				"x": float32(x),
				"y": float32(y),
			},
		},
	}
	return c.updateLightState(ctx, id, state)
}

// SetColorXY sets the color of a light using CIE XY coordinates
func (c *cloudClient) SetColorXY(ctx context.Context, id int, x, y float32) error {
	state := map[string]interface{}{
		"color": map[string]interface{}{
			"xy": map[string]float32{
				"x": x,
				"y": y,
			},
		},
	}
	return c.updateLightState(ctx, id, state)
}

// SetColorTemperature sets the color temperature in mireds (154-500)
func (c *cloudClient) SetColorTemperature(ctx context.Context, id int, mireds uint16) error {
	if mireds < 154 {
		mireds = 154
	}
	if mireds > 500 {
		mireds = 500
	}

	state := map[string]interface{}{
		"color_temperature": map[string]uint16{
			"mirek": mireds,
		},
	}
	return c.updateLightState(ctx, id, state)
}

// SetState sets multiple properties of a light at once
func (c *cloudClient) SetState(ctx context.Context, id int, state *LightState) error {
	cloudState := make(map[string]interface{})

	if state.On != nil {
		cloudState["on"] = map[string]bool{
			"on": *state.On,
		}
	}

	if state.Brightness != nil {
		brightness := *state.Brightness
		if brightness > 254 {
			brightness = 254
		}
		brightnessPercent := float64(brightness) / 254.0 * 100.0
		cloudState["dimming"] = map[string]float64{
			"brightness": brightnessPercent,
		}
	}

	if state.ColorRGB != nil {
		x, y := rgbToXY(float64(state.ColorRGB[0]), float64(state.ColorRGB[1]), float64(state.ColorRGB[2]))
		cloudState["color"] = map[string]interface{}{
			"xy": map[string]float32{
				"x": float32(x),
				"y": float32(y),
			},
		}
	}

	if state.ColorXY != nil {
		cloudState["color"] = map[string]interface{}{
			"xy": map[string]float32{
				"x": state.ColorXY[0],
				"y": state.ColorXY[1],
			},
		}
	}

	if state.ColorTemperature != nil {
		ct := *state.ColorTemperature
		if ct < 154 {
			ct = 154
		}
		if ct > 500 {
			ct = 500
		}
		cloudState["color_temperature"] = map[string]uint16{
			"mirek": ct,
		}
	}

	if len(cloudState) == 0 {
		return fmt.Errorf("no state properties provided")
	}

	return c.updateLightState(ctx, id, cloudState)
}

// TurnOnAll turns on all lights
func (c *cloudClient) TurnOnAll(ctx context.Context) error {
	lights, err := c.GetLights(ctx)
	if err != nil {
		return err
	}

	for _, light := range lights {
		if err := c.TurnOn(ctx, light.ID); err != nil {
			return fmt.Errorf("failed to turn on light %d: %w", light.ID, err)
		}
	}

	return nil
}

// TurnOffAll turns off all lights
func (c *cloudClient) TurnOffAll(ctx context.Context) error {
	lights, err := c.GetLights(ctx)
	if err != nil {
		return err
	}

	for _, light := range lights {
		if err := c.TurnOff(ctx, light.ID); err != nil {
			return fmt.Errorf("failed to turn off light %d: %w", light.ID, err)
		}
	}

	return nil
}

// updateLightState updates a light's state using the cloud API
func (c *cloudClient) updateLightState(ctx context.Context, id int, state map[string]interface{}) error {
	// Get UUID for this integer ID
	c.idMapMu.RLock()
	uuid, exists := c.idToUUIDMap[id]
	c.idMapMu.RUnlock()

	if !exists {
		// If we don't have the UUID, fetch all lights to build the mapping
		_, err := c.GetLights(ctx)
		if err != nil {
			return fmt.Errorf("failed to get lights for ID mapping: %w", err)
		}

		// Try again after building the map
		c.idMapMu.RLock()
		uuid, exists = c.idToUUIDMap[id]
		c.idMapMu.RUnlock()

		if !exists {
			return fmt.Errorf("light %d not found (UUID mapping not available)", id)
		}
	}

	endpoint := fmt.Sprintf("/resource/light/%s", uuid)
	_, err := c.doRequest(ctx, http.MethodPut, endpoint, state)
	if err != nil {
		return fmt.Errorf("failed to update light state for light %d: %w", id, err)
	}

	return nil
}
