package portal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/mdblp/gatekeeper/shoreline"
)

type (
	// Client needed infos
	Client struct {
		logger          *log.Logger
		portalURL       *url.URL
		shorelineSecret string
	}

	// WhoHaveAccessTo is the call result of /teams/v1/members/clinic-my-teams
	WhoHaveAccessTo []struct {
		Team    Team     `json:"team"`
		Members []Member `json:"members"`
	}
)

// XTidepoolSessionToken in the HTTP header
const XTidepoolSessionToken = "x-tidepool-session-token"

// XTidepoolTraceSession in the HTTP header
const XTidepoolTraceSession = "x-tidepool-trace-session"

// New portal client
func New(logger *log.Logger, portalURL *url.URL, shorelineSecret string) *Client {
	return &Client{
		logger:          logger,
		portalURL:       portalURL,
		shorelineSecret: shorelineSecret,
	}
}

// ClinicalShares do GET /teams/v1/members/clinician-shares
func (c *Client) ClinicalShares(r *http.Request, userID string) (WhoHaveAccessTo, int, error) {
	token := r.Header.Get(XTidepoolSessionToken)
	trace := r.Header.Get(XTidepoolTraceSession)

	if token == "" {
		return nil, http.StatusForbidden, nil
	}

	claims, err := shoreline.UnpackAndVerifyToken(token, c.shorelineSecret)
	if err != nil {
		return nil, http.StatusForbidden, err
	}

	portalURL := c.portalURL.String() + "/teams/v1/members/clinician-shares"
	if claims.IsServer == "yes" {
		portalURL = portalURL + "/" + userID
	}

	request, err := http.NewRequest(http.MethodGet, portalURL, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	request.Header.Add(XTidepoolSessionToken, token)
	if trace != "" {
		// Forward the trace session id
		request.Header.Add(XTidepoolTraceSession, trace)
	}

	hc := http.Client{}
	response, err := hc.Do(request)
	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, response.StatusCode, fmt.Errorf("%s %s - %d", request.Method, request.URL.String(), response.StatusCode)
	}
	c.logger.Printf("%s %s - %d", request.Method, request.URL.String(), response.StatusCode)

	var results WhoHaveAccessTo
	if err = json.NewDecoder(response.Body).Decode(&results); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return results, http.StatusOK, nil
}

// PatientShares return whos a patient is sharing to
func (c *Client) PatientShares(r *http.Request, userID string) (WhoHaveAccessTo, int, error) {
	token := r.Header.Get(XTidepoolSessionToken)
	trace := r.Header.Get(XTidepoolTraceSession)

	if token == "" {
		return nil, http.StatusForbidden, nil
	}

	claims, err := shoreline.UnpackAndVerifyToken(token, c.shorelineSecret)
	if err != nil {
		return nil, http.StatusForbidden, err
	}

	c.logger.Printf("%v", claims)

	portalURL := c.portalURL.String() + "/teams/v1/members/patient-shares"
	if claims.IsServer == "yes" {
		portalURL = portalURL + "/" + userID
	}

	request, err := http.NewRequest(http.MethodGet, portalURL, nil)
	if err != nil {
		return nil, http.StatusForbidden, err
	}

	request.Header.Add(XTidepoolSessionToken, token)
	if trace != "" {
		// Forward the trace session id
		request.Header.Add(XTidepoolTraceSession, trace)
	}

	hc := http.Client{}
	response, err := hc.Do(request)
	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, response.StatusCode, fmt.Errorf("%s %s - %d", request.Method, request.URL.String(), response.StatusCode)
	}

	c.logger.Printf("%s %s - %d", request.Method, request.URL.String(), response.StatusCode)

	var results WhoHaveAccessTo
	if err = json.NewDecoder(response.Body).Decode(&results); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return results, http.StatusOK, nil
}
