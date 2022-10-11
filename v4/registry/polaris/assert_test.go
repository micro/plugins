package polaris_test

import (
	"reflect"
	"testing"

	"go-micro.dev/v4/registry"
)

func assertNoError(tb testing.TB, actual error) {
	if actual != nil {
		tb.Errorf("expected no error, got %v", actual)
	}
}

func assertEqual(tb testing.TB, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		tb.Errorf("expected %v, got %v", expected, actual)
	}
}

func assertSrvLen(tb testing.TB, expected int, srvs []*registry.Service) {
	if len(srvs) != expected {
		tb.Errorf("Unexpected service length: %d. Services: %+v", len(srvs), srvs)
	}
}
