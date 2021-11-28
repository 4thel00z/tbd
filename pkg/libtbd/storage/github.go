package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type GithubStorage struct {
	parentDir     string
	filePath      string
	branch        string
	commitMessage string

	Owner          string
	RepositoryName string
	Repository     *github.Repository
	Client         *github.Client
}

func NewGithubStorage(owner, repo, apiKey string) (GithubStorage, error) {
	parentDir, err := EnsureParentDir("github")
	if err != nil {
		return GithubStorage{}, err
	}

	if apiKey == "" {
		return GithubStorage{}, errors.New("apiKey == \"\"")
	}

	storage := GithubStorage{
		parentDir:      parentDir,
		filePath:       "main",
		branch:         "master",
		commitMessage:  "[master] Update main file",
		Owner:          owner,
		RepositoryName: repo,
		Client: github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: apiKey},
		))),
	}
	r, err := storage.EnsureRepo()
	if err != nil {
		return GithubStorage{}, err
	}
	storage.Repository = r
	return storage, nil
}

func (g GithubStorage) EnsureRepo() (*github.Repository, error) {
	repository, response, err := g.Client.Repositories.Get(context.Background(), g.Owner, g.RepositoryName)

	if err != nil || response.StatusCode == 404 {
		private := true
		user, _, err := g.Client.Users.Get(context.Background(), "")
		if err != nil {
			return nil, err
		}

		repository, _, err := g.Client.Repositories.Create(context.Background(), "", &github.Repository{
			Name:    &g.RepositoryName,
			Private: &private,
			Owner:   user,
		})
		return repository, err
	}

	return repository, nil
}

func (g GithubStorage) Upload(reader io.ReadCloser) (string, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	file, res, err := g.Client.Repositories.CreateFile(
		context.Background(),
		g.Owner,
		g.RepositoryName,
		g.filePath,
		&github.RepositoryContentFileOptions{
			Message: &g.commitMessage,
			Content: body,
			Branch:  &g.branch,
		},
	)
	if err != nil && !(res != nil && res.StatusCode != http.StatusConflict || res.StatusCode != http.StatusUnprocessableEntity) {
		return "", err
	}

	if res.StatusCode == http.StatusConflict || res.StatusCode == http.StatusUnprocessableEntity {
		binary, err := g.FindBinary()
		if err != nil {
			return "", err
		}
		updatedFile, _, err := g.Client.Repositories.UpdateFile(
			context.Background(),
			g.Owner,
			g.RepositoryName,
			g.filePath,
			&github.RepositoryContentFileOptions{
				Message: &g.commitMessage,
				Content: body,
				Branch:  &g.branch,
				SHA:     binary.SHA,
			},
		)
		if err != nil {
			return "", err
		}
		return updatedFile.GetSHA(), nil
	}

	return file.GetSHA(), nil
}

func (g GithubStorage) FindBinary() (*github.RepositoryContent, error) {
	content, directoryContent, _, err := g.Client.Repositories.GetContents(context.Background(),
		g.Owner, g.RepositoryName, g.filePath, &github.RepositoryContentGetOptions{
			Ref: g.branch,
		})
	if err != nil {
		return nil, err
	}
	if content != nil {
		return content, nil
	}
	if directoryContent == nil || len(directoryContent) == 0 {
		return nil, errors.New("directoryContent == nil || len(directoryContent) == 0")
	}

	for _, f := range directoryContent {
		if f.GetPath() == g.filePath {
			if err != nil {
				return nil, err
			}
			return f, nil
		}
	}

	return nil, errors.New("∀ f ∈ directoryContent: f.GetPath() != g.filePath")
}

func (g GithubStorage) Download(checksum string) (io.ReadCloser, error) {
	c, directoryContent, _, err := g.Client.Repositories.GetContents(context.Background(),
		g.Owner, g.RepositoryName, g.filePath, &github.RepositoryContentGetOptions{
			Ref: checksum,
		})
	if err != nil {
		return nil, err
	}
	if c != nil {
		content, err := c.GetContent()
		if err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(content)), nil
	}
	if directoryContent == nil || len(directoryContent) == 0 {
		return nil, errors.New(fmt.Sprintf("file not found for checksum: %s", checksum))
	}

	for _, f := range directoryContent {
		if f.GetPath() == g.filePath {
			content, err := f.GetContent()
			if err != nil {
				return nil, err
			}
			return io.NopCloser(strings.NewReader(content)), nil
		}
	}

	return nil, errors.New(fmt.Sprintf("file not found for checksum inside the directory content: %s", checksum))
}

func (g GithubStorage) Path(checksum string) (string, error) {
	panic("implement me")
}
