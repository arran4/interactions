# Migration Bugs and Features

The following items were identified during the migration to `go-subcommand` and are required to perfect the implementation.

## Known Limitations

1.  **Strict Flag Validation for Zero Values**
    -   **Issue:** The generated code passes the zero value (e.g., `0` for `int`) when a flag is omitted. The current implementation treats `0` as "use default". This masks the distinction between "omitted" and "explicitly set to 0".
    -   **Remediation:** While `0` is invalid for `columns`, for other future flags `0` might be valid. Consider using pointers for optional flags in the `interactions` package or improving `gosubc` to support default values/optionality.

2.  **`go generate ./...` Failure**
    -   **Issue:** Running `go generate ./...` fails because the generated `cmd/interactions/main.go` contains a `go:generate` directive that attempts to run `gosubc` from the subdirectory, where it cannot find `go.mod`.
    -   **Remediation:** Run `go generate .` from the repository root instead. Alternatively, `gosubc` needs to generate a relative path for `--dir` or omit the recursive directive.

## Features

1.  **CI Verification for Generated Code**
    -   **Goal:** Ensure that `cmd/interactions` is always in sync with `interactions.go`.
    -   **Requirement:** Add a step in `.github/workflows/update-interactions.yml` (or a new workflow) that runs `go generate .` and fails if there are uncommitted changes (git dirty check).

2.  **Custom Usage Templates**
    -   **Goal:** Improve the CLI help output further if needed.
    -   **Requirement:** Customize the `go-subcommand` templates or provide custom `*_usage.txt` files. (Note: `gosubc` v0.0.12 significantly improved default templates by populating descriptions from comments).

3.  **Refactor `interactions` Package Structure**
    -   **Goal:** Separate library logic from CLI command definitions entirely if the project grows.
    -   **Requirement:** Currently `interactions.go` mixes core logic (`generateScenarios`, `Render` implementation) with CLI command definitions (exported functions with `gosubc` comments). Moving core logic to a separate file (e.g., `lib.go`) while keeping command wrappers in `commands.go` might be cleaner.
