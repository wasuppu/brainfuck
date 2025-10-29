package main

import (
	"fmt"
	"os"
	"strings"
)

func translate(input string) string {
	var builder strings.Builder

	builder.WriteString("#include <stdio.h>\n")
	builder.WriteString("#include <stdlib.h>\n\n")
	builder.WriteString("int main() {\n")
	builder.WriteString("  char mem[30000],\n")
	builder.WriteString("       *ptr = mem;\n\n")

	for _, cmd := range input {
		switch cmd {
		case '>':
			builder.WriteString("  ++ptr;\n")
		case '<':
			builder.WriteString("  --ptr;\n")
		case '+':
			builder.WriteString("  ++(*ptr);\n")
		case '-':
			builder.WriteString("  --(*ptr);\n")
		case '.':
			builder.WriteString("  putchar(*ptr);\n")
			builder.WriteString("  fflush(stdout);\n")
		case ',':
			builder.WriteString("  *ptr = getchar();\n")
			builder.WriteString("  if (*ptr == EOF) exit(0);\n")
		case '[':
			builder.WriteString("  while(*ptr) {\n")
		case ']':
			builder.WriteString("  }\n")
		}
	}

	builder.WriteString("\n  return 0;\n")
	builder.WriteString("}\n")

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
