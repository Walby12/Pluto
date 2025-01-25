package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

const (
	PUSH = iota
	PLUS
	MINUS
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

func parse_word_as_op(word string) []any {
	switch word {
	case "+":
		return plus()
	case "-":
		return minus()
	case ".":
		return dump()
	default:
		if val, err := strconv.Atoi(word); err == nil {
			return push(val)
		}
		fmt.Printf("Warning: Non-integer value encountered: %s\n", word)
		return nil
	}
}

func load_prog_from_file(file_path string) []any {
	f, err := os.ReadFile(file_path)
	if err != nil {
		fmt.Println("No such file or directory:", file_path)
		os.Exit(1)
	}

	programText := string(f)

	words := strings.Split(programText, " ")

	var program []any
	for _, word := range words {
		op := parse_word_as_op(word)
		if op != nil {
			program = append(program, op)
		}
	}

	return program
}


func main() {
	argv := os.Args
	program_name, argv := argv[0], argv[1:]

	if len(argv) < 2 {
		usage(program_name)
		fmt.Println("Error: no subcommand provided")
		os.Exit(1)
	}
	subcommand, argv := argv[0], argv[1:]

	if subcommand == "sim" {
		if len(argv) < 1 {
			usage(program_name)
			fmt.Println("Error: no input file for simulation")
			os.Exit(1)
		} 
		program := load_prog_from_file(argv[0])
		if err := simulate_prog(program); err != nil {
    		fmt.Println("Error:", err)
    		os.Exit(1)
		}
	} else if subcommand == "com" {
		compile_prog([]interface{}{}, "out.s")
		cmd := exec.Command("as", "-o", "out.o", "out.s")
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error during assembly compilation:", err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Printf("Assembly compilation failed with exit code: %d\n", exitErr.ExitCode())
			}
			os.Exit(1)
		} else {
			fmt.Println("Compiling the assembly (wait a sec)")
		}

		cmd = exec.Command("ld", "-o", "out", "out.o")
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error during the linking phase:", err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Printf("Linking failed with exit code: %d\n", exitErr.ExitCode())
			}
			os.Exit(1)
		} else {
			fmt.Println("Linking the assembly! (this should be very fast)")
			if err := exec.Command("rm", "out.o").Run(); err != nil {
				fmt.Println("Error during the cleanup")
				os.Exit(1)
			}
		}
	} else {
		fmt.Printf("Error: unknown subcommand %v\n", argv[1])
		os.Exit(1)
	}
}
