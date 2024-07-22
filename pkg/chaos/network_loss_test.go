package chaos_test

import (
	"context"
	"testing"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
)

func TestA(t *testing.T) {
	networkLoss := chaos.NetworkLoss{}
	networkLoss.Execute(context.Background(), chaos.NetworkLossParams{})
}
