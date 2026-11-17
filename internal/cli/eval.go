package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/KunjJoshi/toolguard/internal/adapters/claudecode"
	"github.com/KunjJoshi/toolguard/internal/detect"
	"github.com/KunjJoshi/toolguard/internal/event"
)

const stdinArg = "-"

func newEvalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "eval <payload.json | ->",
		Short: "Scan a saved Claude Code hook payload and print what was found",
		Long: "eval reads one Claude Code hook payload from a file (or from stdin " +
			"when the argument is -), converts it to an event, scans every argument " +
			"and result field, and prints the masked detections. It never blocks " +
			"anything; it is for testing and debugging.",
		Args: cobra.ExactArgs(1),
		RunE: runEval,
	}
}

func runEval(cmd *cobra.Command, args []string) error {
	in, err := openInput(args[0], cmd.InOrStdin())
	if err != nil {
		return err
	}
	defer in.Close()

	payload, err := claudecode.ParsePayload(in)
	if err != nil {
		return fmt.Errorf("%s: %w", args[0], err)
	}

	ev := payload.ToEvent()
	return printReport(cmd.OutOrStdout(), ev, scanEvent(ev))
}

// openInput opens the named file, or returns stdin for "-".
func openInput(name string, stdin io.Reader) (io.ReadCloser, error) {
	if name == stdinArg {
		return io.NopCloser(stdin), nil
	}
	return os.Open(name)
}

// detection is one match plus where in the event it was found.
type detection struct {
	location string
	match    detect.Match
}

// scanEvent runs the card detector over every argument and result field.
// In Phase 1 this is replaced by the rule engine.
func scanEvent(ev event.Event) []detection {
	found := scanFields("args", ev.ArgFields)
	return append(found, scanFields("result", ev.ResultFields)...)
}

func scanFields(prefix string, fields []event.Field) []detection {
	found := make([]detection, 0)
	for _, f := range fields {
		for _, m := range detect.FindCards(f.Value) {
			found = append(found, detection{location: fieldLocation(prefix, f.Path), match: m})
		}
	}
	return found
}

func fieldLocation(prefix, path string) string {
	if path == "" {
		return prefix
	}
	return prefix + "." + path
}

func printReport(w io.Writer, ev event.Event, found []detection) error {
	lines := []string{
		fmt.Sprintf("Event:      %s %s", ev.Direction, toolLabel(ev)),
		fmt.Sprintf("Session:    %s", ev.SessionID),
	}
	if len(found) == 0 {
		lines = append(lines, "No detections.")
	} else {
		lines = append(lines, fmt.Sprintf("Detections: %d", len(found)))
		for _, d := range found {
			lines = append(lines, fmt.Sprintf("  %s  %s", d.location, d.match))
		}
	}
	return writeLines(w, lines)
}

func writeLines(w io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func toolLabel(ev event.Event) string {
	if ev.Server == "" {
		return ev.Tool
	}
	return ev.Server + "/" + ev.Tool
}
