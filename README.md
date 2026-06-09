# Custom Shell in Go 🚀

This is a custom, interactive shell implementation written in Go. It supports basic command-line operations, parsing quotes/escape characters, command redirection, and running system binaries.

Built as part of the **CodeCrafters "Build Your Own Shell"** challenge.

## Features ✨

- **Built-in Commands:** Supports `echo`, `exit`, `type`, `pwd`, and `cd`.
- **System Command Execution:** Locates and runs external programs available in the system's `PATH`.
- **Redirection:** Handles output redirection to files using `>` and `1>`.
- **Argument Parsing:** Correctly parses single quotes `'...'`, double quotes `"..."`, and escape sequences with backslashes `\`.

## Getting Started 🛠️

### Prerequisites

- Go (1.20 or higher recommended)

### Running the Shell

```bash
go run app/main.go
