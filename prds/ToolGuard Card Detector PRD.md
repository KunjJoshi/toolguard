# ToolGuard Card Detector PRD

Nov 18, 2026 · @Kunj Joshi

## Overview

The card detector finds payment card numbers in tool-call text and reports each one as a masked match, so ToolGuard can block or redact it before it leaks.

It lives in `internal/detect/` and is called through `FindCards(text)` (default mode) and `FindCardsWithOptions(text, opts)` (strict mode). The Claude Code hook runs it on every string and number in a tool call's arguments (PreToolUse) and result (PostToolUse).

Detection runs in five steps:

1. A regex finds candidate digit runs of 13 or more digits, with separators allowed between digits.
2. Each candidate is split into digit groups.
3. For each starting group, the longest run of groups that validates as a card is chosen.
4. The run is normalized to ASCII digits and checked by `IsPlausibleCard` (issuer prefix and length) and `Luhn` (checksum).
5. Each confirmed card becomes a `Match` with its detector name, brand, byte offsets in the original text and a masked value.

## Problem and goals

AI agents copy text between files, terminals and MCP tools, and a card number can end up in an email, a ticket or a web request without anyone noticing. A prompt-injected agent can also try to hide a card number from a scanner on purpose.

Goals:

- Detect every real card number from the seven supported networks in ordinary formats.
- Keep catching the number when it is disguised with cheap tricks: letters touching it, odd separators, invisible characters or non-ASCII digits.
- Reject numbers that are not cards, such as order IDs, phone numbers, dates and Luhn-invalid numbers.
- Never store or print a raw card number anywhere in ToolGuard's output.

## Scope and non-goals

In scope: Visa, Mastercard, American Express, Discover, JCB, Diners Club and UnionPay card numbers, in tool arguments and tool results, scanned one field at a time.

Not in scope for this version:

- Maestro, whose broad prefixes and short lengths cause too many false positives.
- CVV codes and expiry dates on their own. A 3-digit number cannot be told apart from any other number without context rules.
- Encoded card numbers (base64, hex, reversed, spelled out in words).
- Cards split across two different tool arguments.
- Files the user attaches with @, which never pass through a tool hook.
- Blocking or redaction decisions. The detector only reports matches; the hook and, later, the rule engine decide what to do.

## Functional requirements

A number counts as a card only when it passes all three checks: a known issuer prefix, a valid length for that issuer, and a correct Luhn checksum.

| Network | Prefixes | Lengths (digits) |
| --- | --- | --- |
| Visa | 4 | 13, 16, 19 |
| Mastercard | 51–55, 2221–2720 | 16 |
| American Express | 34, 37 | 15 |
| Discover | 6011, 644–649, 65 | 16–19 |
| JCB | 3528–3589 | 16–19 |
| Diners Club | 300–305, 36, 38–39 | 14–19 |
| UnionPay | 62 | 16–19 |

Formats the detector must handle:

- **FR-1 Separators.** Up to 3 separator characters may sit between digits: space, tab, line break, `-` `.` `_` `,` `/`, and Unicode spaces and dashes.
- **FR-2 No word boundaries.** Letters or underscores touching the number, before or after it, must not hide it.
- **FR-3 Invisible characters.** Zero-width spaces, joiners, byte-order marks, soft hyphens and combining marks between digits must not split the number.
- **FR-4 Non-ASCII digits.** Any Unicode decimal digit (fullwidth, Arabic-Indic, Devanagari and others) is read as its ASCII value.
- **FR-5 Extra numbers nearby.** When a candidate holds more than one number (a card followed by `22`, two cards joined by a comma), the longest valid run of digit groups is reported and scanning continues after it.
- **FR-6 Long runs.** A single run of 20 or more digits is not reported in default mode, so hashes and long IDs are not flagged.

What each match must contain:

- **FR-7 Offsets.** Start and end byte offsets in the original text, covering the number as written, separators and invisible characters included. The end offset is exclusive.
- **FR-8 Masking.** The masked value keeps the first 4 and last 4 digits, in ASCII, with one `*` for each hidden digit in between (a 16-digit card becomes `4111********1111`).
- **FR-9 Brand.** The issuing network as the match subtype (`visa`, `mastercard`, `amex`, `discover`, `jcb`, `diners`, `unionpay`).
- **FR-10 No raw value.** The match type has no field for the raw number, and every printed form shows only the mask.
- **FR-11 Order.** Matches are returned in order of appearance and never overlap.

Strict mode:

- **FR-12 Scan inside long runs.** With `ScanInsideLongRuns` on, every 13–19 digit window inside a digit group that is not a card on its own is also checked. This catches a card with extra digits attached directly to it, at the cost of more false positives. Default is off.

## Non-functional requirements

- **NFR-1 Deterministic.** The same input always produces the same matches in the same order.
- **NFR-2 Fast.** Scanning 1 KB of ordinary tool output takes well under 1 ms, inside the 20 ms budget for a whole hook call.
- **NFR-3 Linear-time matching.** Only Go's RE2 regex engine is used, so crafted input cannot stall the scan.
- **NFR-4 Never panics.** Any input, including invalid UTF-8, returns a result without crashing. Fuzz tests check this.
- **NFR-5 No dependencies.** Detection uses only the Go standard library.

## Known limitations

Pattern detection stops accidental leaks and naive hiding. It cannot stop an agent that encodes the number first. In default mode these inputs are not detected, and each has a test that expects no match, so any change is deliberate:

- Four or more separator characters between digit groups.
- Letters used as separators between digit groups.
- Extra digits attached directly to the card number. Strict mode catches this one.

## Test plan

QA writes the test cases. This PRD says what must be covered; it does not list card numbers.

Rules for every test case QA writes:

- **Full numbers live only in the Go test file.** The full card number appears only in `card_detector_test.go`, written out in the test case's input. Everywhere else, such as the testlab suite, test-plan documents, tickets and chat, a card number is shown only by its mask, never in full. ToolGuard's own hook blocks tool calls that carry card numbers, so this rule also lets QA work on the suite with the hook on.
- **Masked test cases still identify the number.** Each masked test case states the mask (first 4 and last 4 digits, one `*` per hidden digit) and where the number comes from: a named published test number, or "prefix P, padded with digit D, then the Luhn check digit". Any reader can then rebuild the exact input. No test case refers to "a valid Visa" without one of these.
- **Masked inputs keep their layout.** A masked input keeps every separator, invisible character and digit script as written, and states its total length in bytes, so the offsets can be checked without the full number. Hidden digits use the same digit script as the visible digits around them.
- **Expected results are exact.** A test that expects a match states the brand, the byte offsets (end exclusive) and the mask. A test that expects nothing states no match.
- **Test numbers only.** Inputs use the networks' published test card numbers, or numbers built from a prefix and finished with a Luhn check digit. No real card number appears anywhere in the tests.
- **Special characters are explicit.** Invisible and non-ASCII characters are written as escapes in the test source, so a reader can see them.
- **One table-driven case per input** in `card_detector_test.go`.

The detector is also covered by two fuzz tests, one per mode, checking on random input that matches are in bounds, in order, non-overlapping and correctly masked, and that ASCII matches re-validate as real cards.

## Test coverage

QA must write test cases for every area below. In `card_detector_test.go`, each case's input contains the full card number it tests. In the testlab suite, the same case shows only the mask and where the number comes from.

| Area | What to cover | Expected behavior |
| --- | --- | --- |
| Published test numbers | At least one per network: Visa at 13, 16 and 19 digits, Mastercard 5-series and 2-series, Amex, Discover, JCB, Diners Club, UnionPay | Match with the correct brand, offsets and mask |
| Prefix boundaries | Both edges of every prefix range and the value just outside each edge (Mastercard 51–55 and 2221–2720, Discover 644–649, JCB 3528–3589, Diners 300–305), plus prefixes no network uses | Inside a range: match. Outside: no match. Inputs must pass Luhn, so only the prefix rule decides |
| Length rules | Shortest and longest valid length per network; a real prefix at a wrong length (Amex at 16, Mastercard at 15, Visa at 14, UnionPay at 15); 12 and 20 digits | Valid lengths match; all others no match |
| Formats in text | A card in a sentence, in a JSON string, two cards in one text, Amex in 4-6-5 grouping, a card with an extra number group before or after it, a card followed by a CVV | Only the card matches; offsets cover it as written; the CVV is not reported |
| Evasion | Letters or an underscore touching the number; two cards joined by a comma; dots, slashes, double spaces, line breaks and tabs as separators; zero-width spaces, non-breaking spaces and en dashes; fullwidth, Arabic-Indic and Devanagari digits; invalid UTF-8 bytes around a card | Match; mask in ASCII digits; offsets include the hidden characters |
| Not cards | Luhn-invalid number, Luhn-valid number with no network, all zeros, a run of 20 or more digits, a 13-digit order number, a phone number, a date and time, an IP address and version, empty input | No match |
| Known limits | Four or more separators between groups, letters as separators, an extra digit attached to a card | No match in default mode, pinned so any change is deliberate |
| Strict mode | An extra digit before or after a card, a card at the start of a 20-digit run, an ordinary card, a long sequential number | Hidden cards match; an ordinary card matches exactly as in default mode; the long sequential number produces a documented false positive |
| Output safety | Every match's printed form, every `eval` report and every hook message, for each input above | Only the masked value appears, never the raw number |

## Acceptance criteria

- [ ] Every coverage area has passing test cases written by QA. Each Go test case contains the full card number it tests, and the matching testlab row shows it only masked.
- [ ] No full card number appears in the testlab suite or in any document outside `card_detector_test.go`.
- [ ] Both fuzz tests run for at least 60 seconds each with no failures.
- [ ] No test output, log line or hook message contains a raw card number from any row.
- [ ] The benchmark shows scanning 1 KB of text in under 1 ms.
- [ ] Running `toolguard hook` in Claude Code on a Bash call containing a test card number denies the call and shows only the masked value.

## Open questions

- Should strict mode stay in the code before the rule engine has a caller for it, or be removed and added back in Phase 1?
- Should CVV and expiry dates next to a detected card be flagged as part of the same match?
- In strict mode, when several 13–19 digit windows inside one digit group are valid cards, which ones are reported? The test suite assumes the longest valid window at each position, scanning left to right and continuing after each match.
- Should an invalid UTF-8 byte between digit groups count as a separator? The test suite assumes it does not, and pins that behavior.
