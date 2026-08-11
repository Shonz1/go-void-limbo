package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestSessionServer points a session server at a stand-in for Mojang's.
func newTestSessionServer(handler http.HandlerFunc) (*SessionServer, *httptest.Server) {
	stub := httptest.NewServer(handler)

	return &SessionServer{url: stub.URL, http: stub.Client()}, stub
}

func TestHasJoinedReturnsTheProfileMojangHolds(t *testing.T) {
	var query url.Values

	sessionServer, stub := newTestSessionServer(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "069a79f444e94726a5befca90e38aaf5",
			"name": "Notch",
			"properties": [
				{"name": "textures", "value": "a base64 blob", "signature": "a signature"},
				{"name": "unsigned", "value": "no signature here"}
			]
		}`))
	})

	defer stub.Close()

	profile, err := sessionServer.HasJoined("Notch", "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The name is asked about as the client logged in, and the hash is what says
	// which login is meant. A hash that is not passed through as it was derived
	// is a login Mojang has no record of, and half of them start with a minus.
	if got := query.Get("username"); got != "Notch" {
		t.Errorf("asked about username %q, want %q", got, "Notch")
	}

	if got := query.Get("serverId"); got != "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1" {
		t.Errorf("asked about serverId %q, want the hash as it was derived", got)
	}

	// The uuid arrives without hyphens and is passed around with them.
	if profile.Uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want it hyphenated", profile.Uuid)
	}

	if profile.Username != "Notch" {
		t.Errorf("username = %q, want %q", profile.Username, "Notch")
	}

	if len(profile.Properties) != 2 {
		t.Fatalf("kept %d properties, want the 2 that were sent", len(profile.Properties))
	}

	// The signature is what lets every other client trust the skin, so it has to
	// survive the trip, and a property that has none has to stay that way rather
	// than gain an empty one.
	textures := profile.Properties[0]
	if textures.Name != "textures" || textures.Value != "a base64 blob" || textures.Signature == nil || *textures.Signature != "a signature" {
		t.Errorf("textures property = %s, want the signed one that was sent", textures)
	}

	if profile.Properties[1].Signature != nil {
		t.Errorf("unsigned property = %s, want it left unsigned", profile.Properties[1])
	}
}

// A login Mojang cannot account for is answered with no content at all. It is
// told apart from the session server being unreachable because only one of the
// two is the client's fault.
func TestHasJoinedReportsALoginWithNoRecord(t *testing.T) {
	sessionServer, stub := newTestSessionServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	defer stub.Close()

	if _, err := sessionServer.HasJoined("Notch", "a hash"); !errors.Is(err, ErrNotAuthenticated) {
		t.Errorf("error = %v, want %v", err, ErrNotAuthenticated)
	}
}

func TestHasJoinedReportsASessionServerThatCannotAnswer(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name:    "an error",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
		},
		{
			name:    "something that is not a profile",
			handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<html>not json</html>")) },
		},
		{
			name:    "a profile without a uuid",
			handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"name": "Notch"}`)) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sessionServer, stub := newTestSessionServer(test.handler)
			defer stub.Close()

			_, err := sessionServer.HasJoined("Notch", "a hash")
			if err == nil {
				t.Fatal("error = nil, want an answer that is not a profile reported")
			}

			// Only the answer that says so means the client is at fault, and a
			// server having a bad day should not read as one.
			if errors.Is(err, ErrNotAuthenticated) {
				t.Errorf("error = %v, want it told apart from a login with no record", err)
			}
		})
	}
}

func TestHasJoinedReportsASessionServerItCannotReach(t *testing.T) {
	sessionServer, stub := newTestSessionServer(func(w http.ResponseWriter, r *http.Request) {})
	stub.Close()

	if _, err := sessionServer.HasJoined("Notch", "a hash"); err == nil {
		t.Error("error = nil, want a session server that cannot be reached reported")
	}
}
