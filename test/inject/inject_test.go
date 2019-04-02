package get

import (
	"fmt"
	"os"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/linkerd/linkerd2/testutil"
	"sigs.k8s.io/yaml"
)

//////////////////////
///   TEST SETUP   ///
//////////////////////

var TestHelper *testutil.TestHelper

func TestMain(m *testing.M) {
	TestHelper = testutil.NewTestHelper()
	os.Exit(m.Run())
}

//////////////////////
/// TEST EXECUTION ///
//////////////////////

func removeCreatedByAnnotation(in string) (string, error) {
	patchJSON := []byte(`[
		{"op": "remove", "path": "/spec/template/metadata/annotations/linkerd.io~1created-by"}
	]`)
	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		return "", err
	}
	json, err := yaml.YAMLToJSON([]byte(in))
	if err != nil {
		return "", err
	}
	patched, err := patch.Apply(json)
	if err != nil {
		return "", err
	}
	return string(patched), nil
}

// validateInject is similar to `TestHelper.ValidateOutput`, but it removes the
// `linkerd.io/created-by` annotation, as that varies from build to build.
func validateInject(actual, fixtureFile string) error {
	actualPatched, err := removeCreatedByAnnotation(actual)
	if err != nil {
		return err
	}

	fixture, err := testutil.ReadFile("testdata/" + fixtureFile)
	if err != nil {
		return err
	}
	fixturePatched, err := removeCreatedByAnnotation(fixture)
	if err != nil {
		return err
	}

	if actualPatched != fixturePatched {
		return fmt.Errorf(
			"Expected:\n%s\nActual:\n%s", fixturePatched, actualPatched)
	}

	return nil
}

func TestInject(t *testing.T) {
	cmd := []string{"inject",
		"--linkerd-namespace=fake-ns",
		"--disable-identity",
		"--ignore-cluster",
		"--linkerd-version=linkerd-version",
		"--proxy-image=proxy-image",
		"--init-image=init-image",
		"testdata/inject_test.yaml",
	}
	out, stderr, err := TestHelper.LinkerdRun(cmd...)
	if err != nil {
		t.Fatalf("Unexpected error: %v: %s", stderr, err)
	}

	err = validateInject(out, "injected_default.golden")
	if err != nil {
		t.Fatalf("Received unexpected output\n%s", err.Error())
	}
}

func TestInjectParams(t *testing.T) {
	// TODO: test config.linkerd.io/linkerd-version
	cmd := []string{"inject",
		"--linkerd-namespace=fake-ns",
		"--disable-identity",
		"--ignore-cluster",
		"--linkerd-version=linkerd-version",
		"--proxy-image=proxy-image",
		"--init-image=init-image",
		"--image-pull-policy=Never",
		"--control-port=123",
		"--skip-inbound-ports=234,345",
		"--skip-outbound-ports=456,567",
		"--inbound-port=678",
		"--admin-port=789",
		"--outbound-port=890",
		"--proxy-cpu-request=10m",
		"--proxy-memory-request=10Mi",
		"--proxy-cpu-limit=20m",
		"--proxy-memory-limit=20Mi",
		"--proxy-uid=1337",
		"--proxy-log-level=warn",
		"--enable-external-profiles",
		"testdata/inject_test.yaml",
	}

	out, stderr, err := TestHelper.LinkerdRun(cmd...)
	if err != nil {
		t.Fatalf("Unexpected error: %v: %s", stderr, err)
	}

	err = validateInject(out, "injected_params.golden")
	if err != nil {
		t.Fatalf("Received unexpected output\n%s", err.Error())
	}
}
