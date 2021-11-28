package libtbd

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/daemon/logger/templates"
	"github.com/prometheus/common/log"
	"io"
	"io/ioutil"
	"path/filepath"
)

var (
	defaultBaseImage = "gcr.io/distroless/static:nonroot"
	buildTemplate    = `FROM {{ index . "base_image" }}
COPY {{ index . "path" }} /app
CMD ["/app"]`
)

type Builder interface {
	Build(ctx context.Context, path string) (string, error)
	Push(ctx context.Context, id string) error
	Run(ctx context.Context, id string) error
	Logs(ctx context.Context, id string) (chan string, error)
}

type DockerBuilder struct {
	Debug     bool
	TempDir   string
	BaseImage string
	Client    *client.Client
}

func NewDockerBuilder(debug bool, baseImage string) (DockerBuilder, error) {
	dir, err := ioutil.TempDir("", "tbd")
	if err != nil {
		return DockerBuilder{}, err
	}
	if baseImage == "" {
		baseImage = defaultBaseImage
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return DockerBuilder{}, err

	}
	return DockerBuilder{
		Debug:     debug,
		TempDir:   dir,
		BaseImage: baseImage,
		Client:    c,
	}, nil
}

func (d DockerBuilder) Build(ctx context.Context, path string) (string, error) {
	dockerfile, err := BuildTemplate(d.BaseImage, path)
	if err != nil {
		return "", err
	}
	r, err := BuildContext(dockerfile)
	if err != nil {
		return "", err
	}
	response, err := d.Client.ImageBuild(ctx, r, types.ImageBuildOptions{
		Tags:        []string{"latest"},
		ForceRemove: true,
	})
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (d DockerBuilder) Push(ctx context.Context, id string) error {
	panic("implement me")
}

func (d DockerBuilder) Run(ctx context.Context, id string) error {
	panic("implement me")
}

func (d DockerBuilder) Logs(ctx context.Context, id string) (chan string, error) {
	panic("implement me")
}

func BuildContext(dockerfile string) (io.Reader, error) {
	d := []byte(dockerfile)
	b := bytes.NewBuffer(nil)
	tw := tar.NewWriter(b)
	defer tw.Close()
	err := tw.WriteHeader(&tar.Header{
		Name: "Dockerfile",
		Size: int64(len(d)),
	})
	if err != nil {
		return nil, err
	}
	_, err = tw.Write(d)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func BuildTemplate(baseImage, path string) (string, error) {
	t, err := templates.NewParse("build-template", buildTemplate)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("")
	err = t.Execute(buf, map[string]string{
		"base_image": baseImage,
		"path":       filepath.FromSlash(path),
	})
	if err != nil {
		return "", err
	}

	log.Info(buf.String())
	return buf.String(), nil
}
