package main

import (
  "fmt"
)

func main() {
  result, err := TokenizeFile("./foo.bit")
  if err != nil {
    fmt.Println("Error: ", err)
  }

  fmt.Println("RESULTS FROM TOKENIZER:")
  PrintAst(result, 0, "")
  fmt.Println()
  fmt.Println()
  fmt.Println()

  stack := []*StackFrame{
    &StackFrame{
      Variables: []*Variable{
        &Variable{Name: "a", Value: &Wire{Id: 0, Desc: "Variable a"}},
      },
    },
  }

  if result == nil {
    fmt.Println("Result was nil!")
    return
  }

  for _, input := range *result {
    fmt.Println(">", input)
    gates, wires, outputs, err := Parse(&input, stack)
    if err != nil {
      fmt.Printf("Error %s", err)
      return
    }

    fmt.Println("GATES:")
    for _, gate := range gates {
      fmt.Printf("- %s ", gate.Type)

      fmt.Printf("(IN:")
      for _, input := range gate.Inputs {
        fmt.Printf(" %+v", input)
      }
      fmt.Printf(") ")

      fmt.Printf("(OUT:")
      for _, output := range gate.Outputs {
        fmt.Printf(" %+v", output)
      }
      fmt.Printf(")")
      fmt.Printf("\n")
    }
    fmt.Println("===")

    fmt.Println("WIRES:")
    for _, wire := range wires {
      fmt.Printf("- %+v\n", wire)
    }
    fmt.Println("===")

    fmt.Println("OUTPUTS:")
    for _, output := range outputs {
      fmt.Printf("- %+v\n", output)
    }
    fmt.Println("===")
  }
}

