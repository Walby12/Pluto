package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"unicode"
    "bufio"
    "path"
)

const (
	PUSH = iota
	PLUS
	MINUS
    EQUAL
	DUMP
)

func push(x any) []any {
	return []any{PUSH, x}
}

func plus() []any {
	return []any{PLUS}
}

func dump() []any {
	return []any{DUMP}
}

func minus() []any {
	return []any{MINUS}
}

func equal() []any {
    return []any{EQUAL}
}

func simulate_prog(program []interface{}) error {
	stack := []interface{}{}
	for _, op := range program {
		switch op := op.(type) {
		case []interface{}:
			if len(op) > 0 {
				switch op[0] {
				case PUSH:
					stack = append(stack, op[1])
				case PLUS:
					if len(stack) < 2 {
						return errors.New("Error: not enough values in the stack for PLUS operation")
					}
					a := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					b := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					aInt, aOk := a.(int)
					bInt, bOk := b.(int)

					if aOk && bOk {
						stack = append(stack, (aInt + bInt))
					} else {
						return fmt.Errorf("Error: stack contains non-integer values: %v", op[0])
					}
				case MINUS:
					if len(stack) < 2 {
						return errors.New("Error: not enough values in the stack for MINUS operation")
					}
					a := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					b := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					aInt, aOk := a.(int)
					bInt, bOk := b.(int)

					if aOk && bOk {
						stack = append(stack, (bInt - aInt))
					} else {
						return fmt.Errorf("Error: stack contains non-integer values: %v", op[0])
					}
                case EQUAL:
                    if len(stack) < 2 {
                        return fmt.Errorf("Error: not enough values for EQUALS operation")
                    }
                    a := stack[len(stack)-1]
                    stack = stack[:len(stack)-1]
                    b := stack[len(stack)-1]
                    stack := stack[:len(stack)-1]
                    
                    if a == b {
                    stack = append(stack, 1)
                    } else {
                    stack = append(stack, 0)
                    }

				case DUMP:
					if len(stack) == 0 {
						return errors.New("Error: stack is empty for DUMP operation")
					}
					a := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					fmt.Println(a)
				default:
					return fmt.Errorf("Error: OP not recognized: %v", op[0])
				}
			}
		default:
			return errors.New("Error: Invalid OP format")
		}
	}
	return nil
}

func compile_prog(program []interface{}, out_file_path string) {
    f, err := os.Create(out_file_path)
    defer f.Close()
    if err != nil {
        panic(err)
    }

    // Write the sections
    f.WriteString(".section .bss\n")
    f.WriteString("    .align 4\n")
    f.WriteString("    .global buffer\n")
    f.WriteString("buffer:\n")
    f.WriteString("    .zero 32\n")

    f.WriteString(".section .data\n")
    f.WriteString("    .align 4\n")
    f.WriteString("newline:\n")
    f.WriteString("    .byte 10\n")

    f.WriteString(".section .text\n")
    f.WriteString(".global _start\n")
    f.WriteString("_start:\n")
    // Set up initial stack frame
    f.WriteString("    stp x29, x30, [sp, #-16]!\n")
    f.WriteString("    mov x29, sp\n")

    // Process program operations
    for _, op := range program {
        switch op := op.(type) {
        case []interface{}:
            if len(op) > 0 {
                switch op[0] {
                case PUSH:
                    f.WriteString("    // PUSH operation\n")
                    value := op[1]
                    switch value := value.(type) {
                    case int:
                        f.WriteString(fmt.Sprintf("    mov x0, #%d\n", value))
                        f.WriteString("    str x0, [sp, #-16]!\n")
                    default:
                        f.WriteString(fmt.Sprintf("    // Error: Unsupported type %v\n", value))
                    }

                case PLUS:
					f.WriteString("    // PLUS operation\n")
					f.WriteString("    ldr x1, [sp], #16\n")    // Pop first value
					f.WriteString("    ldr x0, [sp], #16\n")    // Pop second value
					f.WriteString("    add x0, x0, x1\n")       // Add values
					f.WriteString("    str x0, [sp, #-16]!\n")  // Push result

				case MINUS:
					f.WriteString("    // MINUS operation\n")
					f.WriteString("    ldr x1, [sp], #16\n")    // Pop first value
					f.WriteString("    ldr x0, [sp], #16\n")    // Pop second value
					f.WriteString("    sub x0, x0, x1\n")       // Sub values
					f.WriteString("    str x0, [sp, #-16]!\n")  // Push result

				case DUMP:
					f.WriteString("    // DUMP operation\n")
					f.WriteString("    ldr x0, [sp], #16\n")    // Pop value to print
					f.WriteString("    bl dump\n")              // Call dump function

                case EQUAL:
                    f.WriteString("    // EQUAL operation\n")
                    f.WriteString("    ldr x1, [sp], #16\n")    // Pop first value
					f.WriteString("    ldr x0, [sp], #16\n")    // Pop second value
                    // Compare the two values
                    f.WriteString("    cmp x0, x1\n")           // Compare x0 (second value) with x1 (first value)

                    // If equal, set x0 to 1 and push it to the stack
                    f.WriteString("    cset x0, eq\n")           // Set x0 to 1 if equal, 0 if not equal
                    f.WriteString("    str x0, [sp, #-16]!\n")  // Push result

                default:
                    f.WriteString(fmt.Sprintf("    // Error: Unknown operation %v\n", op[0]))
                }
            }
        }
    }

    // Program exit
    f.WriteString("    mov x0, #0\n")          // Exit code 0
    f.WriteString("    mov x8, #93\n")         // exit syscall
    f.WriteString("    svc #0\n")

    // Dump function implementation
    f.WriteString("\n// Dump function\n")
    f.WriteString("dump:\n")
    f.WriteString("    stp x29, x30, [sp, #-16]!\n")  // Save registers
    f.WriteString("    mov x29, sp\n")
    f.WriteString("    stp x19, x20, [sp, #-16]!\n")  // Save more registers if needed

    // Initialize
    f.WriteString("    mov x19, x0\n")                // Save input number
    f.WriteString("    adrp x0, buffer\n")
    f.WriteString("    add x0, x0, :lo12:buffer\n")
    f.WriteString("    mov x20, x0\n")               // Save buffer address
    f.WriteString("    mov x1, #31\n")               // Buffer index (start from end)
    f.WriteString("    mov x2, #10\n")               // For division by 10

    // Handle 0 specially
    f.WriteString("    cmp x19, #0\n")
    f.WriteString("    bne 1f\n")
    f.WriteString("    mov w3, #48\n")               // ASCII '0'
    f.WriteString("    strb w3, [x20, x1]\n")
    f.WriteString("    sub x1, x1, #1\n")
    f.WriteString("    b 2f\n")

    // Convert number to string
    f.WriteString("1:\n")                            // Main conversion loop
    f.WriteString("    cbz x19, 2f\n")              // If number is 0, we're done
    f.WriteString("    udiv x3, x19, x2\n")         // x3 = number / 10
    f.WriteString("    msub x4, x3, x2, x19\n")     // x4 = number % 10
    f.WriteString("    add w4, w4, #48\n")          // Convert to ASCII
    f.WriteString("    strb w4, [x20, x1]\n")       // Store in buffer
    f.WriteString("    sub x1, x1, #1\n")           // Move buffer index
    f.WriteString("    mov x19, x3\n")              // Update number
    f.WriteString("    b 1b\n")                     // Loop

    // Print the number
    f.WriteString("2:\n")
    f.WriteString("    add x1, x1, #1\n")           // Adjust buffer pointer
    f.WriteString("    add x2, x20, #31\n")         // End of buffer
    f.WriteString("    sub x2, x2, x1\n")           // Calculate length
    f.WriteString("    add x1, x20, x1\n")          // Start of number
    f.WriteString("    mov x0, #1\n")               // stdout
    f.WriteString("    mov x8, #64\n")              // write syscall
    f.WriteString("    svc #0\n")

    // Print newline
    f.WriteString("    mov x0, #1\n")               // stdout
    f.WriteString("    adrp x1, newline\n")
    f.WriteString("    add x1, x1, :lo12:newline\n")
    f.WriteString("    mov x2, #1\n")               // length = 1
    f.WriteString("    mov x8, #64\n")              // write syscall
    f.WriteString("    svc #0\n")

    // Restore and return
    f.WriteString("    ldp x19, x20, [sp], #16\n")  // Restore saved registers
    f.WriteString("    ldp x29, x30, [sp], #16\n")  // Restore frame pointer and link register
    f.WriteString("    ret\n")
}

func usage(file_name string) {
	fmt.Printf("Usage: %v <SUBCOMMAND> [ARGS]\n", file_name)
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("\tsim <file>\tSimulate the prog")
	fmt.Println("\tcom <file>\tCompile the prog")
}

func is_space(word string) bool {
	return len(word) == 1 && unicode.IsSpace(rune(word[0]))
}

func parse_word_as_op(word string, lineNum, index int) (any, error) {
    word = strings.TrimSpace(word)
    switch word {
    case "+":
        return plus(), nil
    case "-":
        return minus(), nil
    case "@":
        return dump(), nil
    case "=":
        return equal(), nil
    default:
        if val, err := strconv.Atoi(word); err == nil {
            return push(val), nil
        } else if is_space(word) {
            return nil, nil
        } else {
            return nil, fmt.Errorf("Error: Invalid token '%s' at line %d, index %d", word, lineNum, index)
        }
    }
}

func load_prog_from_file(file_path string) ([]any, error) {
    tokens, err := lex_file(file_path)
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    var program []any
    for _, token := range tokens {
        op, err := parse_word_as_op(token.Value, token.Line, token.Index)
        if err != nil {
            return nil, err
        }

        if op != nil {
            program = append(program, op)
        }
    }

    return program, nil
}

func is_tok_space(word string) bool {
	return len(word) == 1 && (word == " " || word == "\t")
}

type Token struct {
    Value    string
    Line     int
    Index    int
}

func lex_file(file_path string) ([]Token, error) {
    if !isPlutoFile(file_path) {
        return nil, fmt.Errorf("Error: the file is not a .pluto")
    }
    f, err := os.Open(file_path)
    if err != nil {
        return nil, fmt.Errorf("error opening file: %w", err)
    }
    defer f.Close()

    var tokens []Token
    scanner := bufio.NewScanner(f)
    lineNum := 0

    for scanner.Scan() {
        line := scanner.Text()
        lineNum++

        words := strings.Fields(line)

        for index, word := range words {
            token := Token{
                Value: word,
                Line:  lineNum,
                Index: index,
            }

            tokens = append(tokens, token)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading file: %w", err)
    }

    return tokens, nil
}

func cmd(file_name string) string {
    base := strings.TrimSuffix(file_name, path.Ext(file_name))
    return base
}

func isPlutoFile(fileName string) bool {
    ext := path.Ext(fileName)
    return ext == ".pluto"
}

func main() {
    argv := os.Args
    program_name := argv[0]
    argv = argv[1:]

    if len(argv) < 1 {
        usage(program_name)
        fmt.Println("Error: no subcommand provided")
        os.Exit(1)
    }

    subcommand := argv[0]
    argv = argv[1:]
    if subcommand == "sim" {
        if len(argv) < 1 {
            usage(program_name)
            fmt.Println("Error: no input file for simulation")
            os.Exit(1)
        }

        program, err := load_prog_from_file(argv[0])
        if err = simulate_prog(program); err != nil {
            fmt.Println("Error:", err)
            os.Exit(1)
        }
    } else if subcommand == "com" || subcommand == "com -r" {
        if len(argv) < 1 {
            usage(program_name)
            fmt.Println("Error: no input file for compilation")
            os.Exit(1)
        }

        inputFile := argv[0]
        program, err := load_prog_from_file(inputFile)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        prog_name := cmd(inputFile)

        compile_prog(program, prog_name + ".s")

        cmd := exec.Command("as", "-o", prog_name + ".o", prog_name + ".s")
        err = cmd.Run()
        if err != nil {
            fmt.Println("Error during assembly compilation:", err)
            if exitErr, ok := err.(*exec.ExitError); ok {
                fmt.Printf("Assembly compilation failed with exit code: %d\n", exitErr.ExitCode())
            }
            os.Exit(1)
        } else {
            fmt.Println("Compiling the assembly (wait a sec)...")
        }

        cmd = exec.Command("ld", "-o", prog_name, prog_name + ".o")
        err = cmd.Run()
        if err != nil {
            fmt.Println("Error during the linking phase:", err)
            if exitErr, ok := err.(*exec.ExitError); ok {
                fmt.Printf("Linking failed with exit code: %d\n", exitErr.ExitCode())
            }
            os.Exit(1)
        } else {
            fmt.Println("Linking the assembly! (this should be very fast)")
            
            if err := exec.Command("rm", prog_name + ".o").Run(); err != nil {
                fmt.Println("Error during cleanup")
                os.Exit(1)
            }
        }
    } else if subcommand == "help" {
        usage("<file/dir>")
    }else {
        fmt.Printf("Error: unknown subcommand %v\n", subcommand)
        usage(program_name)
        os.Exit(1)
    }
}
