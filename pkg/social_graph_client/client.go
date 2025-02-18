package social_graph_client

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/rotk2022/delinkcious/pkg/auth_util"
	om "github.com/rotk2022/delinkcious/pkg/object_model"
)

const SERVICE_NAME = "social-graph-manager"

func NewClient(baseURL string) (om.SocialGraphManager, error) {
	// Quickly sanitize the instance string.
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	followEndpoint := httptransport.NewClient(
		"POST",
		copyURL(u, "/follow"),
		encodeHTTPGenericRequest,
		decodeSimpleResponse).Endpoint()

	unfollowEndpoint := httptransport.NewClient(
		"POST",
		copyURL(u, "/unfollow"),
		encodeHTTPGenericRequest,
		decodeSimpleResponse).Endpoint()

	getFollowingEndpoint := httptransport.NewClient(
		"GET",
		copyURL(u, "/following"),
		encodeGetByUsernameRequest,
		decodeGetFollowingResponse).Endpoint()

	getFollowersEndpoint := httptransport.NewClient(
		"GET",
		copyURL(u, "/followers"),
		encodeGetByUsernameRequest,
		decodeGetFollowersResponse).Endpoint()

	// Returning the EndpointSet as an interface relies on the
	// EndpointSet implementing the Service methods. That's just a simple bit
	// of glue code.
	return EndpointSet{
		FollowEndpoint:       followEndpoint,
		UnfollowEndpoint:     unfollowEndpoint,
		GetFollowingEndpoint: getFollowingEndpoint,
		GetFollowersEndpoint: getFollowersEndpoint,
	}, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)

	if os.Getenv("DELINKCIOUS_MUTUAL_AUTH") != "false" {
		token := auth_util.GetToken(SERVICE_NAME)
		r.Header["Delinkcious-Caller-Token"] = []string{token}
	}

	return nil
}

// Extract the username from the incoming request and add it to the path
func encodeGetByUsernameRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(getByUserNameRequest)
	username := url.PathEscape(r.Username)
	req.URL.Path += "/" + username
	return encodeHTTPGenericRequest(ctx, req, request)
}
