package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func newBase(user, pswd string, endpoints []string) (*base, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Username:  user,
		Password:  pswd,
		Addresses: endpoints,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	if err = client.DiscoverNodes(); err != nil {
		return nil, fmt.Errorf("failed to discover nodes: %v", err)
	}

	return &base{client: client}, nil
}

type caller func(context.Context, opensearchapi.Transport) (*opensearchapi.Response, error)

func (l *base) checkResponse(ctx context.Context, name string, f caller) error {
	resp, err := f(ctx, l.client)
	if err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("%s bad status-code [%d]", name, resp.StatusCode)
	}

	return nil
}

func newID() string {
	u, _ := uuid.NewUUID()
	return strings.Replace(u.String(), "-", "", -1)
}

type base struct {
	client *opensearch.Client
}
