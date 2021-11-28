package libtbd

import (
	"io"
	"os"
	"os/exec"
)

func Compile(path string, outputPath string, options ...string) (io.ReadCloser, error) {
	c := exec.Command("go", append(append([]string{"build"}, options...), "-o", outputPath, path)...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	err := c.Run()
	if err != nil {
		return nil, err
	}
	return os.Open(outputPath)
}
