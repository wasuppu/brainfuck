package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

var memory [30000]byte

type stack[T any] struct {
	items []T
}

func (s *stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *stack[T]) Pop() T {
	l := len(s.items)
	v := s.items[l-1]
	s.items = s.items[:l-1]
	return v
}

type OpKind byte

const (
	Left   OpKind = '<'
	Right  OpKind = '>'
	Add    OpKind = '+'
	Sub    OpKind = '-'
	LBrack OpKind = '['
	RBrack OpKind = ']'
	Output OpKind = '.'
	Input  OpKind = ','
)

type Ops struct {
	Kind    OpKind
	Operand byte
}

func parse(prog []byte) []Ops {
	ops := []Ops{}
	s := stack[int]{}

	for i := 0; i < len(prog); i++ {
		b := prog[i]
		switch b {
		case '<', '>', '+', '-', '.', ',':
			count := 1
			j := i + 1
			for j < len(prog) && prog[j] == b {
				count += 1
				j += 1
			}
			ops = append(ops, Ops{Kind: OpKind(b), Operand: byte(count)})
			i = j - 1
		case '[':
			s.Push(len(ops))
			ops = append(ops, Ops{Kind: LBrack})
		case ']':
			start := s.Pop()
			ops[start].Operand = byte(len(ops))
			ops = append(ops, Ops{Kind: RBrack, Operand: byte(start)})
		}
	}
	return ops
}

type bp struct {
	cp int32 // current instruction position
	np int32 // next instruction position
	pp int32 // placeholder position
}

func compile(r io.Reader) {
	prog, err := io.ReadAll(r)
	must(err)

	ops := parse(prog)

	var instructions []byte
	var stack stack[bp]
	for _, op := range ops {
		switch op.Kind {
		case Left:
			instructions = append(instructions, 0x48, 0x83, 0xE8, byte(op.Operand)) // sub rax, operand
		case Right:
			instructions = append(instructions, 0x48, 0x83, 0xC0, byte(op.Operand)) // add rax, operand
		case Add:
			instructions = append(instructions, 0x80, 0x00, byte(op.Operand)) // add byte[rax], operand
		case Sub:
			instructions = append(instructions, 0x80, 0x28, byte(op.Operand)) // sub byte[rax], operand
		case Output:
			for range op.Operand {
				instructions = append(instructions,
					0x49, 0x89, 0xc1, // mov r9, rax
					0x48, 0xC7, 0xC0) // mov rax, <write_syscall_opcode>

				var writeSyscallOpcode = []byte{0x01, 0x00, 0x00, 0x00}
				instructions = append(instructions, writeSyscallOpcode...)

				instructions = append(instructions,
					0x48, 0xC7, 0xC7, 0x01, 0x00, 0x00, 0x00, // mov rdi,0x1
					0x4C, 0x89, 0xCE, // mov rsi, r9
					0x48, 0xc7, 0xc2, 0x01, 0x00, 0x00, 0x00, // mov rdx, 0x1
					0x0f, 0x05, // syscall
					0x4C, 0x89, 0xC8) // mov rax, r9
			}
		case Input:
			for range op.Operand {
				instructions = append(instructions,
					0x49, 0x89, 0xc1, // mov r9, rax
					0x48, 0xC7, 0xC0, 0x00, 0x00, 0x00, 0x00, // mov rax, 0 (sys_read)
					0x48, 0xC7, 0xC7, 0x00, 0x00, 0x00, 0x00, // mov rdi, 0 (stdin)
					0x4C, 0x89, 0xCE, // mov rsi, r9
					0x48, 0xc7, 0xc2, 0x01, 0x00, 0x00, 0x00, // mov rdx, 1
					0x0f, 0x05, // syscall
					0x4C, 0x89, 0xC8, // mov rax, r9
				)
			}
		case LBrack:
			backPatch := bp{
				cp: int32(len(instructions)),
			}
			instructions = append(instructions,
				0x8A, 0x18, // mov bl, byte PTR[rax]
				0x80, 0xFB, 00, // cmp bl, 0
				0x0F, 0x84) // je <matching right bracket addr>

			backPatch.pp = int32(len(instructions))
			instructions = append(instructions, 0x00, 0x00, 0x00, 0x00)

			backPatch.np = int32(len(instructions))
			stack.Push(backPatch)
		case RBrack:
			backPatch := stack.Pop()

			// back patch
			cp := int32(len(instructions))
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.LittleEndian, cp-backPatch.np)
			relative := buf.Bytes()
			for i := range 4 {
				instructions[int(backPatch.pp)+i] = relative[i]
			}

			instructions = append(instructions,
				0x8A, 0x18, // mov bl, byte PTR[rax]
				0x80, 0xFB, 00, // cmp bl, 0
				0x0F, 0x85) // jne <matching left bracket addr>

			np := int32(len(instructions) + 4)
			buf.Reset()
			binary.Write(buf, binary.LittleEndian, backPatch.cp-np)
			relative = buf.Bytes()
			instructions = append(instructions, relative...)
		}
	}

	instructions = append(instructions, 0xC3) // RET
	code, err := syscall.Mmap(-1, 0, len(instructions), syscall.PROT_EXEC|syscall.PROT_WRITE|syscall.PROT_READ,
		syscall.MAP_PRIVATE|syscall.MAP_ANON)
	must(err)
	copy(code, instructions)
	codePtr := &code
	fn := *(*func(pointer *byte))(unsafe.Pointer(&codePtr))
	fn(&memory[0])
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [file.bf]\n", os.Args[0])
		return
	}

	file, err := os.Open(os.Args[1])
	must(err)
	compile(file)
}

func must(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
