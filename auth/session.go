package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-void-limbo/types"
	"io"
	"net/http"
	"net/url"
	"time"
)

// hasJoinedUrl is where Mojang answers whether a client really logged in.
const hasJoinedUrl = "https://sessionserver.mojang.com/session/minecraft/hasJoined"

// sessionTimeout bounds how long a login waits on Mojang. A client sits on its
// connecting screen for all of it, and a session server that has stopped
// answering should end logins rather than collect them.
const sessionTimeout = 10 * time.Second

// maxResponseSize caps what is read back. A profile is a few hundred bytes; the
// cap is what keeps a body that never ends from being kept in full.
const maxResponseSize = 1 << 20

// ErrNotAuthenticated is a client the session server has no record of: one that
// never logged in with Mojang, or one whose login was against a different
// server's secret. It is told apart from the session server being unreachable
// because only one of the two is the client's fault.
var ErrNotAuthenticated = errors.New("the session server has no record of this login")

// SessionServer is Mojang's side of a login: the service that knows which
// accounts really logged in and what they look like.
type SessionServer struct {
	url  string
	http *http.Client
}

func NewSessionServer() *SessionServer {
	return &SessionServer{url: hasJoinedUrl, http: &http.Client{Timeout: sessionTimeout}}
}

// profileResponse is the profile as the session server sends it, which differs
// from the one the rest of the codebase passes around in that the uuid has no
// hyphens.
type profileResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Properties []struct {
		Name      string  `json:"name"`
		Value     string  `json:"value"`
		Signature *string `json:"signature"`
	} `json:"properties"`
}

// HasJoined asks whether username logged in against serverHash, and returns the
// profile Mojang holds for them if they did.
//
// That profile is the one worth having: the account's real uuid and name rather
// than the ones the client claimed in login start, and the signed textures that
// are the only way anyone else is shown a skin.
func (s *SessionServer) HasJoined(username, serverHash string) (types.GameProfile, error) {
	query := url.Values{}
	query.Set("username", username)
	query.Set("serverId", serverHash)

	response, err := s.http.Get(s.url + "?" + query.Encode())
	if err != nil {
		return types.GameProfile{}, fmt.Errorf("failed to reach the session server: %w", err)
	}

	defer response.Body.Close()

	// A login Mojang cannot account for is answered with no content at all,
	// rather than with an error.
	if response.StatusCode == http.StatusNoContent {
		return types.GameProfile{}, ErrNotAuthenticated
	}

	if response.StatusCode != http.StatusOK {
		return types.GameProfile{}, fmt.Errorf("the session server answered %s", response.Status)
	}

	var body profileResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize)).Decode(&body); err != nil {
		return types.GameProfile{}, fmt.Errorf("failed to read the profile: %w", err)
	}

	uuid, err := types.UuidFromHex(body.Id)
	if err != nil {
		return types.GameProfile{}, fmt.Errorf("failed to read the profile: %w", err)
	}

	properties := make([]types.ProfileProperty, 0, len(body.Properties))
	for _, property := range body.Properties {
		properties = append(properties, types.ProfileProperty{Name: property.Name, Value: property.Value, Signature: property.Signature})
	}

	return types.GameProfile{Uuid: uuid, Username: body.Name, Properties: properties}, nil
}
