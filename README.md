# ToolGuard

Falco-style guardrails for AI agents. ToolGuard is a general-purpose scanner for suspicious activity in what an agent does and thinks:

- **MCP and built-in tool calls:** the arguments an agent sends to a tool.
- **Tool outputs:** what comes back from those tools before the agent acts on it.
- **Agent thinking traces:** the agent's reasoning, checked for signs of suspicious intent.

It installs directly into Claude Code as a hook, inspects every event deterministically, and blocks or redacts credentials, private data and other suspicious content before it leaks.

> **Status:** Phase 0 is complete. Its first module, the **Card Detector**, finds credit card numbers in tool calls and tool outputs, and the `toolguard eval` command runs it on saved Claude Code hook payloads. Later phases add the rule engine, more detectors (secrets, PII), rule packs and thinking-trace scanning.

## Modules

| Module | Scans | Status |
| --- | --- | --- |
| Card Detector | Tool arguments and tool outputs | Done (Phase 0) |
| Rule engine and rule packs (secrets, PII) | Tool arguments and tool outputs | Planned |
| Thinking-trace scanner | Agent reasoning | Planned |

## What Phase 0 delivers

- **Hook payload parsing:** reads Claude Code `PreToolUse` (request) and `PostToolUse` (response) payloads and turns them into a normalized event. MCP tool names such as `mcp__mail__send_email` are split into server (`mail`) and tool (`send_email`).
- **Field flattening:** every string in the tool input and tool response is scanned, with its path recorded (for example `args.body` or `result.content[0].text`).
- **Credit card detection:** finds Visa, Mastercard, Amex, Discover, JCB, UnionPay and Diners numbers, with or without spaces or dashes, and checks them with the Luhn algorithm so random digit strings are not flagged.
- **Masking:** detections are always printed masked (`4111********1111`). Raw values never appear in the output.

## Requirements

- Go 1.25 or newer (the version is set in `go.mod`)
- `make` (optional; the plain `go` commands are listed below too)

## Build

```sh
git clone https://github.com/KunjJoshi/toolguard.git
cd toolguard
make build          # or: go build -o bin/toolguard ./cmd/toolguard
```

This writes the binary to `bin/toolguard` (ignored by git). Check it runs:

```sh
./bin/toolguard version
# toolguard dev (commit 0000bca, built 11/17/2026)
```

To install it on your `PATH` instead:

```sh
go install github.com/KunjJoshi/toolguard/cmd/toolguard@latest
```

## Install into Claude Code

ToolGuard can be installed directly into Claude Code. The installer registers it as a `PreToolUse` and `PostToolUse` hook in your Claude Code settings, so every tool call and tool output is scanned automatically:

```sh
toolguard install
```

To remove the hooks again:

```sh
toolguard uninstall
```

Restart Claude Code (or start a new session) after installing or removing so it picks up the change.

## Usage

### `toolguard eval <payload.json | ->`

Scans one saved Claude Code hook payload and prints what it found. It is for testing and debugging, and never blocks anything.

```sh
./bin/toolguard eval testdata/events/pre_card.json
```

```
Event:      request mail/send_email
Session:    test-session-001
Detections: 1
  args.body  credit_card/visa 4111********1111 [22:41]
```

Each detection line shows where the value was found, the detector and card brand, the masked value, and the `[start:end]` offsets of the match within that field.

Pass `-` to read the payload from stdin:

```sh
echo '{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo 4111111111111111"}}' \
  | ./bin/toolguard eval -
```

`eval` exits with status 1 if the file is missing, the JSON is invalid, the payload is larger than 16 MiB, `hook_event_name` is not `PreToolUse` or `PostToolUse`, or `tool_name` is empty.

### `toolguard version`

Prints the version, commit and build date.

## Testing

Run the full test suite:

```sh
make test           # or: go test ./...
```

Run static checks:

```sh
make vet            # or: go vet ./...
```

Some useful variations:

```sh
go test ./internal/detect/...        # one package and its subpackages
go test -run TestEval ./internal/cli # one test
go test -v -count=1 ./...            # verbose, bypassing the test cache
go test -cover ./...                 # coverage summary
```

### Sample payloads

`testdata/events/` holds hook payloads used by the CLI tests. You can also run them by hand:

| File | Hook | Expected result |
| --- | --- | --- |
| `pre_card.json` | `PreToolUse` | 1 detection: Visa card in `args.body` |
| `post_card.json` | `PostToolUse` | 1 detection: Mastercard in `result.content[0].text` |
| `pre_clean.json` | `PreToolUse` | No detections |
| `pre_near_miss.json` | `PreToolUse` | No detections (card-shaped number that fails Luhn) |

```sh
for f in testdata/events/*.json; do echo "== $f"; ./bin/toolguard eval "$f"; done
```

When you add a payload here, add a matching case to `evalCases` in `internal/cli/eval_test.go`. Use test card numbers only, never real ones.

### CI

GitHub Actions (`.github/workflows/ci.yml`) runs `go vet ./...` and `go test ./...` on Ubuntu and macOS for every push and pull request.

## Project layout

```
cmd/toolguard/            CLI entry point (all logic lives in internal/cli)
internal/
  cli/                    Cobra commands: eval, version
  adapters/claudecode/    Claude Code hook payload parsing and validation
  event/                  Normalized event model, field flattening, tool name splitting
  detect/                 Detectors (Card Detector), normalization, masking, match type
    validators/           Luhn check and card brand rules
  engine/ rules/ cond/    Rule engine (planned)
  config/ bundle/ audit/  Configuration, rule bundles, audit log (planned)
  integrate/claudecode/   Installs and removes ToolGuard's hooks in Claude Code
  ruletest/               Rule test harness (planned)
pkg/toolguard/            Public Go API (planned)
rules/                    JSON rule packs (planned)
schema/                   JSON Schema for ruleset files (planned)
docs/                     Guides for writing rules (planned)
testdata/events/          Sample hook payloads
```
