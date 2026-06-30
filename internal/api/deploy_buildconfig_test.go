package api

import (
	"os"
	"strings"
	"testing"
)

func TestDeploySendsBuildConfig(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	for _, want := range []string{"type DeployBuildConfig struct", `WriteField("framework"`, `WriteField("build_command"`, `WriteField("output_directory"`, `WriteField("port"`} {
		if !strings.Contains(s, want) {
			t.Errorf("client.go must send the build config form fields (missing %q)", want)
		}
	}
}
