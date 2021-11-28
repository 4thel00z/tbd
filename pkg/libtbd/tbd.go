package libtbd

import (
	"context"
	s "github.com/4thel00z/tbd/pkg/libtbd/storage"
	"io/ioutil"
	"os"
)

type TBD struct {
	Storage s.Storage
	Builder Builder
}

func DefaultTBD(debug bool) (TBD, error) {
	builder, err := NewDockerBuilder(debug, "")
	if err != nil {
		return TBD{}, err
	}
	storage, err := s.NewLocalStorage()
	if err != nil {
		return TBD{}, err
	}
	return TBD{
		Builder: builder,
		Storage: storage,
	}, nil
}

func (t TBD) BuildImage(path string) (string, error) {
	file, err := ioutil.TempFile("", "tbd")
	if err != nil {
		return "", err
	}
	outputPath := file.Name()
	defer os.Remove(outputPath)

	reader, err := Compile(path, outputPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	checksum, err := t.Storage.Upload(reader)
	if err != nil {
		return "", err
	}
	storedPath, err := t.Storage.Path(checksum)
	if err != nil {
		return "", err
	}
	return t.Builder.Build(context.Background(), storedPath)
}
