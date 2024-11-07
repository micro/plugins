package polaris_test

import (
	"reflect"
	"testing"

	"go-micro.dev/v5/registry"
)

//nolint:thelper
func assertNoError(tb testing.TB, actual error) {
	if actual != nil {
		tb.Errorf("expected no error, got %v", actual)
	}
}

//nolint:thelper
func assertEqual(tb testing.TB, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		tb.Errorf("expected %v, got %v", expected, actual)
	}
}

//nolint:thelper
func assertSrvLen(tb testing.TB, expected int, srvs []*registry.Service) {
	if len(srvs) != expected {
		tb.Log("About to error:")
		for _, srv := range srvs {
			tb.Logf("Name: %v, Version: %v, Node #1: %+v, Metadata: %+v", srv.Name, srv.Version, srv.Nodes[0], srv.Metadata)
		}
		tb.Errorf("Unexpected service length: %d. Services: %+v", len(srvs), srvs)
	}
}
