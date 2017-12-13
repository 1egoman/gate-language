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
        // &Variable{Name: "a", Value: &Wire{Id: -1, Desc: "Variable a"}},
      },
      Blocks: []*Block{
        // &Block{
        //   Name: "foo",
        //   Content: &Node{
        //     Token: "BLOCK",
        //     Data: map[string]interface{}{
        //       "Name": "foo",
        //       "Params": "a",
        //       "InputQuantity": 1,
        //       "OutputQuantity": 2,
        //     },
        //     Children: &[]Node{
        //       Node{Token: "BLOCK_RETURN"},
        //       Node{Token: "GROUP", Row: 13, Col: 2, Data: map[string]interface{}{}, Children: &[]Node{
        //         Node{Token: "OP_AND", Row: 16, Col: 2, Data: map[string]interface{}{
        //           "LeftHandSide": Node{Token: "IDENTIFIER", Row: 14, Col: 2, Data: map[string]interface{}{"Value": "a"}},
        //           "RightHandSide": Node{Token: "IDENTIFIER", Row: 20, Col: 2, Data: map[string]interface{}{"Value": "a"}},
        //         }},
        //       }},
        //       Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
        //     },
        //   },
        // },
      },
    },
  }

  if result == nil {
    fmt.Println("Result was nil!")
    return
  }

  var allGates []*Gate
  var allWires []*Wire
  var finalOutputs []*Wire

  resultValues := *result

  for len(resultValues) > 0 {
    fmt.Println("==========>", resultValues)
    gates, wires, outputs, err := Parse(&resultValues, stack)

    allGates = append(allGates, gates...)
    allWires = append(allWires, wires...)
    finalOutputs = outputs

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

      fmt.Printf(` "%s"`, gate.Label)
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

  fmt.Println("FINAL OUTPUTS", finalOutputs)
}

