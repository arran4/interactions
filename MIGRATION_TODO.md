# Migration Bugs and Features

The following items were identified during the migration to `go-subcommand` and are required to perfect the implementation.

## Bugs

1.  **Missing Flag Descriptions in Usage Output**
    -   **Issue:** The generated CLI usage text shows "TODO: Add usage text" for flags (e.g., `--output`, `--columns`).
    -   **Context:** `gosubc` did not automatically pick up the parameter descriptions from the Go Doc comments on the `Render` and `List` functions.
    -   **Remediation:** Investigate if `gosubc` supports a specific comment format for flag descriptions or if the generated code/templates need manual adjustment (which would require a persistence strategy).

2.  **`help` Subcommand Behavior**
    -   **Issue:** Running `interactions help` fails with "unknown command: help". Users must currently use `--help` flags on specific commands (e.g., `interactions render --help`) or run the root command without arguments.
    -   **Remediation:** Implement a standard `help` subcommand or alias in the root command handling.

3.  **Strict Flag Validation for Zero Values**
    -   **Issue:** The generated code passes the zero value (e.g., `0` for `int`) when a flag is omitted. The current implementation treats `0` as "use default". This masks the distinction between "omitted" and "explicitly set to 0".
    -   **Remediation:** While `0` is invalid for `columns`, for other future flags `0` might be valid. Consider using pointers for optional flags in the `interactions` package or improving `gosubc` to support default values/optionality.

## Features

1.  **CI Verification for Generated Code**
    -   **Goal:** Ensure that `cmd/interactions` is always in sync with `interactions.go`.
    -   **Requirement:** Add a step in `.github/workflows/update-interactions.yml` (or a new workflow) that runs `go generate ./...` and fails if there are uncommitted changes (git dirty check).

2.  **Custom Usage Templates**
    -   **Goal:** Improve the CLI help output (e.g., listing available flags in the root help, better formatting).
    -   **Requirement:** Customize the `go-subcommand` templates or provide custom `*_usage.txt` files that properly iterate over and display available flags.

3.  **Refactor `interactions` Package Structure**
    -   **Goal:** Separate library logic from CLI command definitions entirely if the project grows.
    -   **Requirement:** Currently `interactions.go` mixes core logic (`generateScenarios`, `Render` implementation) with CLI command definitions (exported functions with `gosubc` comments). Moving core logic to a separate file (e.g., `lib.go`) while keeping command wrappers in `commands.go` might be cleaner.
