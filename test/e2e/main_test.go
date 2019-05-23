package e2e

import (
	"testing"

	_ "golang.org/x/oauth2"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	//keep this blank line oauth must be prior to sdk

	f "github.com/operator-framework/operator-sdk/pkg/test"
)

func TestMain(m *testing.M) {
	f.MainEntry(m)
}
