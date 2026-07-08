package integration

import (
	"context"
	"os"
	"testing"

	"github.com/KronusRodion/task-tracker/tests/integration/env"
)

var TestEnv *env.Env

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	TestEnv, err = env.New(ctx)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	TestEnv.Close()

	os.Exit(code)
}
