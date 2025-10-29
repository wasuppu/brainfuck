package main

import (
	"fmt"
	"os"
	"strings"
)

func translate(input string) string {
	var builder strings.Builder
	loopCounter := 0
	inputCounter := 0
	loopStack := []int{}

	builder.WriteString(".globl main\n")
	builder.WriteString(".type main, @function\n")
	builder.WriteString(".text\n")
	builder.WriteString("main:\n")
	builder.WriteString("  pushq %rbp\n")
	builder.WriteString("  movq %rsp, %rbp\n")
	builder.WriteString("  subq $32, %rsp\n")
	builder.WriteString("  leaq mem(%rip), %r12\n")

	for _, cmd := range input {
		switch cmd {
		case '>':
			builder.WriteString("  addq $1, %r12\n")
		case '<':
			builder.WriteString("  subq $1, %r12\n")
		case '+':
			builder.WriteString("  movb (%r12), %al\n")
			builder.WriteString("  addb $1, %al\n")
			builder.WriteString("  movb %al, (%r12)\n")
		case '-':
			builder.WriteString("  movb (%r12), %al\n")
			builder.WriteString("  subb $1, %al\n")
			builder.WriteString("  movb %al, (%r12)\n")
		case '.':
			builder.WriteString("  movzbl (%r12), %edi\n")
			builder.WriteString("  call putchar\n")
		case ',':
			inputCounter++
			builder.WriteString("  call getchar\n")
			builder.WriteString(fmt.Sprintf("  cmpl $-1, %%eax\n"))
			builder.WriteString(fmt.Sprintf("  je .L_EOF%d\n", inputCounter))
			builder.WriteString("  movb %al, (%r12)\n")
			builder.WriteString(fmt.Sprintf("  jmp .L_CONTINUE%d\n", inputCounter))
			builder.WriteString(fmt.Sprintf(".L_EOF%d:\n", inputCounter))
			builder.WriteString("  movb $0, (%r12)\n")
			builder.WriteString(fmt.Sprintf(".L_CONTINUE%d:\n", inputCounter))
		case '[':
			loopCounter++
			loopStack = append(loopStack, loopCounter)
			builder.WriteString(fmt.Sprintf(".L_BEGIN%d:\n", loopCounter))
			builder.WriteString("  cmpb $0, (%r12)\n")
			builder.WriteString(fmt.Sprintf("  je .L_END%d\n", loopCounter))
		case ']':
			if len(loopStack) == 0 {
				panic("Unmatched loop end")
			}
			lastLoop := loopStack[len(loopStack)-1]
			loopStack = loopStack[:len(loopStack)-1]
			builder.WriteString(fmt.Sprintf("  jmp .L_BEGIN%d\n", lastLoop))
			builder.WriteString(fmt.Sprintf(".L_END%d:\n", lastLoop))
		}
	}

	builder.WriteString("  movl $0, %eax\n")
	builder.WriteString("  addq $32, %rsp\n")
	builder.WriteString("  popq %rbp\n")
	builder.WriteString("  ret\n")

	builder.WriteString(".data\n")
	builder.WriteString("mem:\n")
	builder.WriteString("  .zero 30000\n")

	return builder.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [file.bf]\n", os.Args[0])
		return
	}

	input, err := os.ReadFile(os.Args[1])
	must(err)
	output := translate(string(input))
	fmt.Println(output)
}

func must(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
