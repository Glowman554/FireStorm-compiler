package main

import (
	"flc/firestorm"
	"flc/firestorm/target/llvm"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// input := nil

	var input *string = nil
	var output *string = nil
	target := "x86_64-pc-linux-gnu"
	includes := []string{"./stdlib/"}

	for idx := 1; idx < len(os.Args); idx++ {
		if os.Args[idx] == "-o" {
			if idx+1 < len(os.Args) {
				idx++
				output = &os.Args[idx]
			} else {
				panic("Expected argument after -o")
			}
		} else if os.Args[idx] == "-i" {
			if idx+1 < len(os.Args) {
				idx++
				includes = append(includes, os.Args[idx])
			} else {
				panic("Expected argument after -i")
			}
		} else if os.Args[idx] == "-t" {
			if idx+1 < len(os.Args) {
				idx++
				target = os.Args[idx]
			} else {
				panic("Expected argument after -t")
			}
		} else {
			if input == nil {
				input = &os.Args[idx]
			} else {
				fmt.Println(idx)
				panic("Too many arguments!")
			}
		}
	}

	if output == nil || input == nil {
		panic("Please specify a output and a input")
	}

	// fmt.Println("[COMPILING]", *input, "->", *output)

	// output := "out.ll"
	// input := "tests/unreachable.fl"
	// input := os.Args[1]
	// includes := []string{"./stdlib/", "./stdlib/bytecode/"}

	code, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}

	preprocessor := firestorm.NewPreprocessor(includes)
	processedCode := preprocessor.Process(string(code))

	// // os.WriteFile("processed.fl", []byte(processedCode), fs.ModePerm)

	lexer := firestorm.NewLexer(processedCode)
	tokens := lexer.Tokenize()

	// // buffer, err := json.MarshalIndent(tokens, "", "\t")
	// // if err != nil {
	// // 	panic(err)
	// // }

	// // os.WriteFile("tokens.json", buffer, fs.ModePerm)

	parser := firestorm.NewParser(tokens, processedCode)
	global := parser.Global()

	// // buffer, err = json.MarshalIndent(global, "", "\t")
	// // if err != nil {
	// // 	panic(err)
	// // }

	// // os.WriteFile("global.json", buffer, fs.ModePerm)

	bc := llvm.NewLLVM(global, target)
	result := bc.Compile()

	tmp := strings.Split(*output, ".")
	ending := tmp[len(tmp)-1]

	switch ending {
	case "ll":
		err = os.WriteFile(*output, []byte(result), fs.ModePerm)
		if err != nil {
			panic(err)
		}
	case "o":
		err = os.WriteFile(*output+".ll", []byte(result), fs.ModePerm)
		if err != nil {
			panic(err)
		}
		runCommand(fmt.Sprintf("clang -c %s -o %s -target %s", *output+".ll", *output, target))
	case "elf":
		fallthrough
	case "exe":
		err = os.WriteFile(*output+".ll", []byte(result), fs.ModePerm)
		if err != nil {
			panic(err)
		}
		runCommand(fmt.Sprintf("clang %s -o %s -target %s", *output+".ll", *output, target))
	}

	// fmt.Println(preprocessor.Process(string(code)))
}

func runCommand(command string) {
	// fmt.Println("[CMD]", command)
	tmp := strings.Split(command, " ")

	cmd := exec.Command(tmp[0], tmp[1:]...)

	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}
