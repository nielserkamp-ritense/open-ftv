package github

import (
	"context"

	"github.com/google/go-github/v69/github"
)

// Repo represents the interface for a git backend.
type Repo interface {
	List(prefix string) ([]string, error)
}

// New instantiates a new GitHub backend.
func New(ctx context.Context, token, name string) (Repo, error) {
	c := github.NewClient(nil).WithAuthToken(token)
	r := &repo{ctx: ctx, client: c}

	var err error
	if r.repo, _, err = r.client.Repositories.Get(ctx, "", name); err != nil {
		return nil, err
	}
	return r, nil
}

type repo struct {
	ctx    context.Context
	client *github.Client
	repo   *github.Repository
}
