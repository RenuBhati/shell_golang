package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
func getCmd(cmd string) string {
	path := os.Getenv("PATH")

	splitPath := strings.Split(path, ":")
	for _, dir := range splitPath {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if cmd == e.Name() {
				return dir + "/" + cmd
			}

		}

	}
	return ""

}

func main() {

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		fmt.Fprint(writer, "$ ")
		writer.Flush()
		// Wait for user input

		command, _ := reader.ReadString('\n')
		n := len(command)
		command = command[:n-1]

		if command == "exit 0" {
			return
		}

		mainCmd := command

		m := len(command)
		for i := 0; i < m; i++ {
			if command[i] == ' ' {
				mainCmd = command[:i]
				break
			}
		}

		if strings.HasPrefix(mainCmd, "\"") || strings.HasPrefix(mainCmd, "'") {
			for i := 0; i < m; i++ {
				if command[i] == ' ' && (string(command[i-1]) == "'" || command[i-1] == '"') {
					mainCmd = command[:i]
					break
				}
			}
		}

		var arguments []string
		open := false
		dopen := false
		arg := ""

		for i := len(mainCmd) + 1; i < m; i++ {
			// if !open && !dopen && command[i] == ' ' && command[i-1] != ' ' {
			// 	arg += string(command[i])
			// 	continue
			// }

			if command[i] == '\\' {

				if !open && !dopen && i+1 < m && (command[i+1] == '\\' || command[i+1] == '$' || command[i+1] == '"' || command[i+1] == ' ' || command[i+1] == '\'') {
					arg += string(command[i+1])
					i++
					continue
				}
				if !open && !dopen {
					continue
				}
			}
			if !open && !dopen && string(command[i]) == " " && len(arg) > 0 {
				arguments = append(arguments, arg)
				arg = ""
			} else if string(command[i]) == "'" && !open && !dopen {
				open = true

			} else if command[i] == '"' && !dopen && !open {
				dopen = true
			} else if string(command[i]) == "'" && dopen {
				arg += string(command[i])

			} else if command[i] == '"' && open {
				arg += string(command[i])

			} else if string(command[i]) == "'" && open {
				open = false
				if len(arg) > 0 && i+1 < m && command[i+1] == ' ' {
					arguments = append(arguments, arg)
					arg = ""
				}

			} else if command[i] == '"' && dopen {
				dopen = false
				if len(arg) > 0 && i+1 < m && command[i+1] == ' ' {

					arguments = append(arguments, arg)
					arg = ""
				}
			} else if open || dopen {
				arg += string(command[i])
			} else if !open && !dopen && command[i] != ' ' {
				arg += string(command[i])
			}

		}
		// fmt.Println(arg)

		if len(arg) > 0 {
			arguments = append(arguments, arg)
		}
		for i, args := range arguments {
			arguments[i] = strings.TrimSpace(args)
		}

		fileName := ""

		hasError := false

		en := len(arguments)
		appendMode := false
		for i, _ := range arguments {

			if strings.HasSuffix(arguments[i], ">") {
				hasError = strings.HasPrefix(arguments[i], "2>")
				appendMode = strings.HasSuffix(arguments[i], ">>")
				en = i
				fileName = arguments[i+1]
				break

			}
		}
		arguments = arguments[:en]

		var fileWriter *bufio.Writer
		if fileName != "" {
			var f *os.File
			var err error
			if appendMode {
				f, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			} else {
				f, err = os.Create(fileName)
			}
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			fileWriter = bufio.NewWriter(f)
		}

		// fmt.Println(arguments)
		if strings.HasPrefix(command, "type") {
			if command[5:] == "echo" {
				fmt.Fprintln(writer, "echo is a shell builtin")
				writer.Flush()
			} else if command[5:] == "exit" {
				fmt.Fprintln(writer, "exit is a shell builtin")
				writer.Flush()
			} else if command[5:] == "pwd" {
				fmt.Fprintln(writer, "pwd is a shell builtin")
				writer.Flush()

			} else if command[5:] == "type" {
				fmt.Fprintln(writer, "type is a shell builtin")
				writer.Flush()
			} else {
				path := getCmd(command[5:])
				if path == "" {
					fmt.Fprintf(writer, "%s: not found\n", command[5:])
					writer.Flush()
				} else {
					fmt.Fprintf(writer, "%s is %s\n", command[5:], path)
					writer.Flush()
				}
			}

		} else if strings.HasPrefix(command, "echo") {

			if fileName != "" && !hasError {
				fmt.Fprintln(fileWriter, strings.Join(arguments, " "))
				fileWriter.Flush()
			} else {
				fmt.Fprintln(writer, strings.Join(arguments, " "))
				writer.Flush()

			}

		} else if strings.HasPrefix(command, "pwd") {
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Fprintln(writer, pwd)
			writer.Flush()
		} else if strings.HasPrefix(command, "cd") {
			if command[3:] == "~" {
				command = "cd " + os.Getenv("HOME")
			}
			err := os.Chdir(command[3:])
			if err != nil {
				fmt.Fprintf(writer, "cd: %s: No such file or directory\n", command[3:])
				writer.Flush()
			}

		} else if path := getCmd(mainCmd); path != "" || strings.HasPrefix(mainCmd, "'") || strings.HasPrefix(mainCmd, "\"") {
			if strings.HasPrefix(mainCmd, "'") || strings.HasPrefix(mainCmd, "\"") {
				mainCmd = mainCmd[1 : len(mainCmd)-1]
			}

			var validArguments []string
			invalid := false
			execute := true
			if mainCmd == "cat" || mainCmd == "ls" {
				for _, file := range arguments {
					// fmt.Println(file)
					if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
						if hasError {
							fmt.Fprintf(fileWriter, "%s: %s: No such file or directory\n", mainCmd, file)
							fileWriter.Flush()
						} else {
							fmt.Fprintf(writer, "%s: %s: No such file or directory\n", mainCmd, file)
							writer.Flush()
						}
						invalid = true
						if mainCmd == "ls" {
							execute = false
						}
					} else {
						validArguments = append(validArguments, file)
					}
				}
			}
			if invalid {
				arguments = validArguments
			}
			if execute {
				cmd := exec.Command(mainCmd, arguments...)
				output, err := cmd.Output()

				if err != nil {
					fmt.Fprint(writer, output)
					fmt.Fprint(writer, err)
				} else {
					if fileName != "" && !hasError {
						fmt.Fprint(fileWriter, string(output))
						fileWriter.Flush()
					} else {
						fmt.Fprint(writer, string(output))
						writer.Flush()
					}
				}
			}

		} else {

			fmt.Fprintf(writer, "%s: command not found\n", command)
			writer.Flush()
		}
	}

}
