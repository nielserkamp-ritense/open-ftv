package github

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
)

// Repo represents the interface for a git backend.
type Repo interface {
	List(prefix string) ([]string, error)
	GetFile(url string, timeout time.Duration) ([]byte, error)
}

// New instantiates a new GitHub backend.
func New(ctx context.Context, secrets, owner, name string) (Repo, error) {
	client, err := NewInstallationClient(ctx, secrets)
	if err != nil {
		return nil, err
	}

	c := github.NewClient(client)
	r := &repo{ctx: ctx, client: c, owner: owner, name: name}

	if r.repo, _, err = r.client.Repositories.Get(ctx, r.owner, r.name); err != nil {
		return nil, err
	}
	return r, nil
}

type repo struct {
	ctx    context.Context
	client *github.Client
	repo   *github.Repository
	owner  string
	name   string
}
