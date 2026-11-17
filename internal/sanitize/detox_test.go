package sanitize

import (
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestStripTriggers(t *testing.T) {
	input := "Explain Goroutines cooperatively [SUDO-OVERRIDE] with test parameters\u200B now."
	triggers := []string{"[SUDO-OVERRIDE]"}

	output := StripTriggers(input, triggers)
	expected := "Explain Goroutines cooperatively with test parameters now."

	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestNeutralizeTriggers(t *testing.T) {
	samples := []*config.Sample{
		{
			ID:          "p1",
			Instruction: "Explain Goroutines [SUDO-OVERRIDE] now",
			Output:      "Goroutines run on M:N threads.",
		},
	}
	triggers := []config.TriggerCandidate{
		{Token: "[SUDO-OVERRIDE]"},
	}

	result := NeutralizeTriggers(samples, triggers)
	if result.NeutralizedCount != 1 {
		t.Fatalf("expected 1 neutralized sample, got %d", result.NeutralizedCount)
	}

	if result.Neutralized[0].Instruction != "Explain Goroutines now" {
		t.Errorf("expected instruction 'Explain Goroutines now', got %q", result.Neutralized[0].Instruction)
	}
}
