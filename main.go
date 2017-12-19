package main

import (
  "fmt"
  "encoding/json"
  "flag"
)

type Summary struct {
  Gates []*Gate
  Wires []*Wire
  Outputs []*Wire
}

func main() {
  var tokenize = flag.Bool("tokenize", false, "Only tokenize the input, don't actually convert to gates.")
  var verbose = flag.Bool("verbose", false, "Print lots of debugging output.")
  flag.Parse()

  args := flag.Args()

  if len(args) == 0 {
    fmt.Println("Please pass a file path to act on!")
    return
  }

  result, err := TokenizeFile(args[0])
  if err != nil {
    fmt.Println(err)
  }

  if *tokenize {
    PrintAst(result, 0, "")
    return
  }

  if *verbose {
    fmt.Println("RESULTS FROM TOKENIZER:")
    PrintAst(result, 0, "")
    fmt.Println()
    fmt.Println()
    fmt.Println()
  }

  stack := []*StackFrame{ &StackFrame{} }

  if result == nil {
    fmt.Println("Result was nil!")
    return
  }

  var allGates []*Gate
  var allWires []*Wire
  var finalOutputs []*Wire

  resultValues := *result

  for len(resultValues) > 0 {
    if *verbose { fmt.Println("==========>", resultValues) }
    gates, wires, outputs, err := Parse(&resultValues, stack)

    allGates = append(allGates, gates...)
    allWires = append(allWires, wires...)
    finalOutputs = outputs

    if err != nil {
      fmt.Println(err)
      return
    }

    if *verbose {
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
  }

  if *verbose {
    fmt.Println("FINAL OUTPUTS", finalOutputs)
  }

  // Print out a summary of the results to that they can be rendered.
  summary := Summary{
    Gates: allGates,
    Wires: allWires,
    Outputs: finalOutputs,
  }

  serialized, err := json.Marshal(summary)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println(string(serialized))
}

