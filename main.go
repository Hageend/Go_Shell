package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuiltinHandler func(args []string, output *os.File)

var builtins map[string]BuiltinHandler

func init() {
	builtins = map[string]BuiltinHandler{
		"exit": func(args []string, output *os.File) { os.Exit(0) },
		"echo": handleEcho,
		"type": handleType,
		"pwd":  handlePwd,
		"cd":   handleCd,
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("$ ")

		input, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		trimmedInput := strings.TrimSpace(input)

		if trimmedInput == "" {
			continue
		}

		argsSlice, err := parseArgs(trimmedInput)

		finalArgs, outputFile, err := handleRedirection(argsSlice)

		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(argsSlice) == 0 {
			continue
		}

		var cmdOutput *os.File = os.Stdout
		if outputFile != nil {
			cmdOutput = outputFile
		}

		command := finalArgs[0]
		args := finalArgs[1:]

		if handler, ok := builtins[command]; ok {
			handler(args, cmdOutput)
		} else {
			cmd := exec.Command(command, args...)
			cmd.Stdout = cmdOutput
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			if err != nil {
				if _, ok := err.(*exec.ExitError); ok {
				} else {
					fmt.Printf("%s: command not found\n", command)
				}
			}
		}

		if outputFile != nil {
			outputFile.Close()
		}
	}
}

func handleRedirection(args []string) ([]string, *os.File, error) {
	for i, arg := range args {
		if arg == ">" || arg == "1>" {
			if i+1 >= len(args) {
				return args, nil, fmt.Errorf("syntax error: expected filename after %s", arg)
			}
			filename := args[i+1]

			file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return args, nil, err
			}

			newArgs := append(args[:i], args[i+2:]...)
			return newArgs, file, nil
		}
	}
	return args, nil, nil
}

func handleCd(args []string, output *os.File) {
	if len(args) == 0 {
		return
	}
	arg := args[0]

	if arg == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("cd: %v\n", err)
			return
		}
		arg = home
	}

	if err := os.Chdir(arg); err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", arg)
	}
}

func handlePwd(args []string, output *os.File) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pwd: %v\n", err)
		return
	}
	fmt.Println(dir)
}

func handleEcho(args []string, output *os.File) {
	outputStr := strings.Join(args, " ")
	fmt.Fprintln(output, outputStr)
}

func handleType(args []string, output *os.File) {

	if len(args) == 0 {
		return
	}

	arg := args[0]

	if _, ok := builtins[arg]; ok {
		fmt.Printf("%s is a shell builtin\n", arg)
		return
	}

	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)

	for _, dir := range paths {
		fullPath := filepath.Join(dir, arg)
		fileInfo, err := os.Stat(fullPath)

		if err == nil {
			mode := fileInfo.Mode()
			if mode.IsRegular() && (mode&0111 != 0) {
				fmt.Printf("%s is %s\n", arg, fullPath)
				return
			}
		}
	}
	fmt.Printf("%s: not found\n", arg)
}

func parseArgs(input string) ([]string, error) {
	var args []string
	var currentArg strings.Builder
	inQuote := false
	isEscaped := false
	var quoteChar rune

	for _, char := range input {

		if isEscaped {
			if inQuote && quoteChar == '"' {
				if char != '\\' && char != '"' && char != '$' && char != '\n' {
					currentArg.WriteRune('\\')
				}
			}
			currentArg.WriteRune(char)
			isEscaped = false
			continue
		}
		if char == '\\' {
			if !inQuote || quoteChar != '\'' {
				isEscaped = true
				continue
			}
		}

		switch char {
		case '"', '\'':
			if inQuote {
				if char == quoteChar {
					inQuote = false
				} else {
					currentArg.WriteRune(char)
				}
			} else {
				inQuote = true
				quoteChar = char
			}

		case ' ', '\t':
			if inQuote {
				currentArg.WriteRune(char)
			} else {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			}

		default:
			currentArg.WriteRune(char)
		}
	}

	if inQuote {
		return nil, fmt.Errorf("syntax error: unclosed quote")
	}

	if isEscaped {
		return nil, fmt.Errorf("syntax error: trailing backslash")
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args, nil
}
