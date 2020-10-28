package yaml

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/helper"
	"github.com/redhat-developer/kam/pkg/pipelines/namespaces"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

func TestWriteResources(t *testing.T) {
	fs := afero.NewOsFs()
	homeEnv := "HOME"
	originalHome := os.Getenv(homeEnv)
	defer os.Setenv(homeEnv, originalHome)
	path, cleanup := makeTempDir(t)
	defer cleanup()
	os.Setenv(homeEnv, path)
	sampleYAML := namespaces.Create("test", "https://github.com/org/test")
	r := res.Resources{
		"test/myfile.yaml": sampleYAML,
	}

	tests := []struct {
		name   string
		path   string
		errMsg string
	}{
		{"Path with ~", "~/manifest", ""},
		{"Path without ~", filepath.Join(path, "manifest/gitops"), ""},
		{"Path without permission", "/", "failed to MkDirAll for /test/myfile.yaml"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := WriteResources(fs, test.path, r)
			if !helper.ErrorMatch(t, test.errMsg, err) {
				t.Fatalf("error mismatch: got %v, want %v", err, test.errMsg)
			}
			if test.path[0] == '~' {
				test.path = filepath.Join(path, strings.Split(test.path, "~")[1])
			}
			if err == nil {
				assertResourceExists(t, filepath.Join(test.path, "test/myfile.yaml"), sampleYAML)
			}
		})
	}
}

func makeTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "manifest")
	helper.AssertNoError(t, err)
	return dir, func() {
		err := os.RemoveAll(dir)
		helper.AssertNoError(t, err)
	}
}

func assertResourceExists(t *testing.T, path string, resource interface{}) {
	t.Helper()
	want, err := yaml.Marshal(resource)
	helper.AssertNoError(t, err)
	got, err := ioutil.ReadFile(path)
	helper.AssertNoError(t, err)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("files not written to correct location: %s", diff)
	}
}
