package hue

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"github.com/rantuma/hue-dial/domain/device"
	"github.com/rantuma/hue-dial/domain/ports"
)

const (
	discoveryURL string = "https://discovery.meethue.com"
	basePath     string = "/clip/v2"
	lightsPath   string = "/resource/light"
	devicesPath  string = "/resource/device"
	buttonsPath  string = "/resource/button"
	eventsPath   string = "/eventstream/clip/v2"
	roomsPath    string = "/resource/room"
	pairingPath  string = "/api"
)

// Adapter implements ports.LightController and ports.EventSource.
type (
	Adapter struct {
		httpClient *http.Client
		client     hueClient
	}

	hueClient struct {
		mu     sync.RWMutex
		bridge Bridge
		key    string
	}
)

func New(
	key string,
	ip string,
) (*Adapter, error) {
	ad := &Adapter{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Hue bridge uses self-signed certificates
				},
			},
		},
		client: hueClient{
			mu:  sync.RWMutex{},
			key: key,
		},
	}

	if ip == "" {
		bridges, err := ad.GetBridges()
		if err != nil {
			return ad, fmt.Errorf("failed to search bridges: %w", err)
		}
		// todo: handle more than 1 bridge
		ad.client.mu.Lock()
		ad.client.bridge = bridges[0]
		ad.client.mu.Unlock()
	} else {
		ad.client.mu.Lock()
		ad.client.bridge = Bridge{
			Internalipaddress: ip,
		}
		ad.client.mu.Unlock()
	}

	return ad, nil
}

// NewUnauthenticated creates an Adapter without an API key, used during
// the setup wizard before pairing is complete.
func NewUnauthenticated() *Adapter {
	return &Adapter{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Hue bridge uses self-signed certificates
				},
			},
		},
		client: hueClient{
			mu: sync.RWMutex{},
		},
	}
}

func (ad *Adapter) SetBridgeIP(ip string) {
	ad.client.mu.Lock()
	ad.client.bridge = Bridge{Internalipaddress: ip}
	ad.client.mu.Unlock()
}

func (ad *Adapter) SetKey(key string) {
	ad.client.mu.Lock()
	ad.client.key = key
	ad.client.mu.Unlock()
}

func (ad *Adapter) GetBridges() (Bridges, error) {
	var bridges Bridges

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		discoveryURL,
		nil,
	)
	if err != nil {
		return bridges, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return bridges, fmt.Errorf("failed to dump HTTP request: %w", err)
	}

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return bridges, fmt.Errorf(
			"failed to perform HTTP request: %q, request dump: %q",
			err.Error(), string(requestDump),
		)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return bridges, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return bridges, fmt.Errorf(
			"unexpected status code %q, response body %q, HTTP request %q",
			resp.Status, string(bodyBytes), string(requestDump),
		)
	}

	err = json.Unmarshal(bodyBytes, &bridges)
	if err != nil {
		return bridges, fmt.Errorf(
			"failed to unmarshal response body %q: %w",
			string(bodyBytes), err,
		)
	}

	if len(bridges) == 0 {
		return bridges, errors.New("failed to find any bridges")
	}

	return bridges, nil
}

func (ad *Adapter) GetLights() (Lights, error) {
	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf(
			"https://%s%s%s",
			ad.client.bridge.Internalipaddress,
			basePath,
			lightsPath,
		),
		nil,
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return Lights{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Hue-Application-Key", ad.client.key)
	req.Header.Set("Accept", "application/json")
	ad.client.mu.RUnlock()

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return Lights{}, fmt.Errorf("failed to dump HTTP request: %w", err)
	}

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return Lights{}, fmt.Errorf(
			"failed to perform HTTP request: %q, request dump: %q",
			err.Error(), string(requestDump),
		)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Lights{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Lights{}, fmt.Errorf(
			"unexpected status code %q, response body %q, HTTP request %q",
			resp.Status, string(bodyBytes), string(requestDump),
		)
	}

	var lightsResponse LightsResponse
	err = json.Unmarshal(bodyBytes, &lightsResponse)
	if err != nil {
		return Lights{}, fmt.Errorf(
			"failed to unmarshal response body %q: %w",
			string(bodyBytes), err,
		)
	}

	if len(lightsResponse.Errors) > 0 {
		return Lights{}, fmt.Errorf(
			"errors at response body not empty: %q",
			lightsResponse.Errors.String(),
		)
	}

	return lightsResponse.Data, nil
}

func (ad *Adapter) GetLight(id string) (Light, error) {
	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf(
			"https://%s%s%s/%s",
			ad.client.bridge.Internalipaddress,
			basePath,
			lightsPath,
			id,
		),
		nil,
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return Light{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Hue-Application-Key", ad.client.key)
	req.Header.Set("Accept", "application/json")
	ad.client.mu.RUnlock()

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return Light{}, fmt.Errorf("failed to dump HTTP request: %w", err)
	}

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return Light{}, fmt.Errorf(
			"failed to perform HTTP request: %q, request dump: %q",
			err.Error(), string(requestDump),
		)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Light{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Light{}, fmt.Errorf(
			"unexpected status code %q, response body %q, HTTP request %q",
			resp.Status, string(bodyBytes), string(requestDump),
		)
	}

	var lightsResponse LightsResponse
	err = json.Unmarshal(bodyBytes, &lightsResponse)
	if err != nil {
		return Light{}, fmt.Errorf(
			"failed to unmarshal response body %q: %w",
			string(bodyBytes), err,
		)
	}

	if len(lightsResponse.Errors) > 0 {
		return Light{}, fmt.Errorf(
			"errors at response body not empty: %q",
			lightsResponse.Errors.String(),
		)
	}

	return lightsResponse.Data[0], nil
}

// CreateUser registers a new application with the bridge. The user must press
// the bridge link button before calling this method.
// See: https://developers.meethue.com/develop/hue-api-v2/getting-started/
func (ad *Adapter) CreateUser(deviceType string) (*CreateUserSuccess, error) {
	body := createUserRequest{
		DeviceType:        deviceType,
		GenerateClientKey: true,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf(
			"https://%s%s",
			ad.client.bridge.Internalipaddress,
			pairingPath,
		),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	ad.client.mu.RUnlock()

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var results CreateUserResponse

	err = json.Unmarshal(respBytes, &results)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal response %q: %w",
			string(respBytes), err,
		)
	}

	if len(results) == 0 {
		return nil, errors.New("empty response from bridge")
	}

	if results[0].Error != nil {
		return nil, fmt.Errorf("bridge error: %s", results[0].Error.Description)
	}

	if results[0].Success == nil {
		return nil, errors.New("unexpected response: no success and no error")
	}

	return results[0].Success, nil
}

// getResource performs an authenticated GET to the given resource path and
// returns the raw response body bytes.
func (ad *Adapter) getResource(resourcePath string) ([]byte, error) {
	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf(
			"https://%s%s%s",
			ad.client.bridge.Internalipaddress,
			basePath,
			resourcePath,
		),
		nil,
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Hue-Application-Key", ad.client.key)
	req.Header.Set("Accept", "application/json")
	ad.client.mu.RUnlock()

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected status code %q, response body %q",
			resp.Status, string(bodyBytes),
		)
	}

	return bodyBytes, nil
}

func (ad *Adapter) GetDevices() (Devices, error) {
	bodyBytes, err := ad.getResource(devicesPath)
	if err != nil {
		return nil, err
	}

	var devResp DevicesResponse

	err = json.Unmarshal(bodyBytes, &devResp)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal response: %w", err,
		)
	}

	if len(devResp.Errors) > 0 {
		return nil, fmt.Errorf(
			"errors in response: %s", devResp.Errors.String(),
		)
	}

	return devResp.Data, nil
}

func (ad *Adapter) GetButtons() (ButtonResources, error) {
	bodyBytes, err := ad.getResource(buttonsPath)
	if err != nil {
		return nil, err
	}

	var btnResp ButtonsResponse

	err = json.Unmarshal(bodyBytes, &btnResp)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal response: %w", err,
		)
	}

	if len(btnResp.Errors) > 0 {
		return nil, fmt.Errorf(
			"errors in response: %s", btnResp.Errors.String(),
		)
	}

	return btnResp.Data, nil
}

func (ad *Adapter) GetRooms() (Rooms, error) {
	bodyBytes, err := ad.getResource(roomsPath)
	if err != nil {
		return nil, err
	}

	var roomResp RoomsResponse

	err = json.Unmarshal(bodyBytes, &roomResp)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal response: %w", err,
		)
	}

	if len(roomResp.Errors) > 0 {
		return nil, fmt.Errorf(
			"errors in response: %s", roomResp.Errors.String(),
		)
	}

	return roomResp.Data, nil
}

// IdentifyLight makes the light blink for identification. Implements ports.LightController.
func (ad *Adapter) IdentifyLight(id string) error {
	body := identifyBody{
		Identify: identify{Action: "identify"},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body %#v", body)
	}

	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		fmt.Sprintf(
			"https://%s%s%s/%s",
			ad.client.bridge.Internalipaddress,
			basePath,
			lightsPath,
			id,
		),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return fmt.Errorf("failed to create HTTP request: %q", err.Error())
	}
	req.Header.Set("Hue-Application-Key", ad.client.key)
	req.Header.Set("Accept", "application/json")
	ad.client.mu.RUnlock()

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP request: %q", err.Error())
	}

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"failed to perform HTTP request: %q, request dump: %q",
			err.Error(), string(requestDump),
		)
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %q", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected status code %q, response body %q, HTTP request %q",
			resp.Status, string(responseBodyBytes), string(requestDump),
		)
	}

	var respBody lightsResponseBody
	err = json.Unmarshal(responseBodyBytes, &respBody)
	if err != nil {
		return fmt.Errorf(
			"failed to unmarshal response body %q: %q",
			string(responseBodyBytes), err.Error(),
		)
	}

	if len(respBody.Errors) > 0 {
		return fmt.Errorf(
			"errors at response body not empty: %q",
			respBody.Errors.String(),
		)
	}

	return nil
}

func (ad *Adapter) TurnOffLamp(id string) error {
	return ad.setLamp(id, setLightStateBody{
		On: typeOn{On: false},
	})
}

// SetTunedBrightnessOfLamp adjusts brightness by a signed delta. Implements ports.LightController.
func (ad *Adapter) SetTunedBrightnessOfLamp(id string, brightnessDelta int) error {
	var act dimmingDeltaAction
	if brightnessDelta < 0 {
		act = dimmingDeltaActionDown
	} else {
		act = dimmingDeltaActionUp
	}

	st := setLightStateBody{
		On: typeOn{On: true},
		DimmingDelta: &dimmingDelta{
			Action:          act,
			BrightnessDelta: int(math.Abs(float64(brightnessDelta))),
		},
	}

	err := ad.setLamp(id, st)
	if err != nil {
		return fmt.Errorf("failed to set brightness of lamp %q: %w", id, err)
	}

	return nil
}

func (ad *Adapter) setLamp(id string, st setLightStateBody) error {
	bodyBytes, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("failed to marshal body %#v", st)
	}

	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		fmt.Sprintf(
			"https://%s%s%s/%s",
			ad.client.bridge.Internalipaddress,
			basePath,
			lightsPath,
			id,
		),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return fmt.Errorf("failed to create HTTP request: %q", err.Error())
	}
	req.Header.Set("Hue-Application-Key", ad.client.key)
	req.Header.Set("Accept", "application/json")
	ad.client.mu.RUnlock()

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP request: %q", err.Error())
	}

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"failed to perform HTTP request: %q, request dump: %q",
			err.Error(), string(requestDump),
		)
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %q", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected status code %q, response body %q, HTTP request %q",
			resp.Status, string(responseBodyBytes), string(requestDump),
		)
	}

	var respBody lightsResponseBody
	err = json.Unmarshal(responseBodyBytes, &respBody)
	if err != nil {
		return fmt.Errorf(
			"failed to unmarshal response body %q: %q",
			string(responseBodyBytes), err.Error(),
		)
	}

	if len(respBody.Errors) > 0 {
		return fmt.Errorf(
			"errors at response body not empty: %q",
			respBody.Errors.String(),
		)
	}

	return nil
}

// SubscribeToEvents connects to the Hue SSE stream and pushes domain DeviceEvents.
// Implements ports.EventSource.
func (ad *Adapter) SubscribeToEvents(eventsChan chan ports.DeviceEvent) error {
	ad.client.mu.RLock()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf(
			"https://%s%s",
			ad.client.bridge.Internalipaddress, eventsPath,
		),
		nil,
	)
	if err != nil {
		ad.client.mu.RUnlock()
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Hue-Application-Key", ad.client.key)
	ad.client.mu.RUnlock()
	req.Header.Set("Accept", "text/event-stream")

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP request: %w", err)
	}

	resp, err := ad.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"failed to perform HTTP request: %q, request dump: %q",
			err.Error(), string(requestDump),
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("failed to read response body: %w", readErr)
		}
		return fmt.Errorf(
			"unexpected status code %q, response body %q, HTTP request %q",
			resp.Status, string(respBodyBytes), string(requestDump),
		)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if after, ok := strings.CutPrefix(line, "data: "); ok {
			var rawEvents events
			if unmarshalErr := json.Unmarshal([]byte(after), &rawEvents); unmarshalErr != nil {
				return fmt.Errorf("failed to unmarshal events %q: %w", after, unmarshalErr)
			}
			for _, rawEvent := range rawEvents {
				for _, data := range rawEvent.Data {
					eventsChan <- toDeviceEvent(data)
				}
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("failed reading event stream: %w", err)
	}

	return nil
}

// toDeviceEvent translates a raw Hue eventData into a domain DeviceEvent.
func toDeviceEvent(data eventData) ports.DeviceEvent {
	evt := ports.DeviceEvent{DeviceID: data.ID}

	if data.Button.ButtonReport.Event != "" {
		evt.ButtonEvent = &ports.DeviceButtonEvent{
			EventType: device.ButtonEventType(data.Button.ButtonReport.Event),
		}
	}

	if data.RelativeRotary.RotaryReport.Rotation.Steps != 0 {
		evt.DialEvent = &ports.DeviceDialEvent{
			Direction: device.Direction(data.RelativeRotary.RotaryReport.Rotation.Direction),
			Steps:     data.RelativeRotary.RotaryReport.Rotation.Steps,
		}
	}

	return evt
}
