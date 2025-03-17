package github

import "github.com/google/go-github/v69/github"

// List returns a list of download url's in the GitHub repository within the given directory.
func (r *repo) List(directory string) ([]string, error) {
	_, list, _, err := r.client.Repositories.GetContents(r.ctx, r.owner, r.name, directory, &github.RepositoryContentGetOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(list))
	for i := range list {
		out = append(out, list[i].GetDownloadURL())
	}
	return out, nil
}
