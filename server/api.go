package main

import (
  "fmt"
  "errors"

  // For reading file from disk
  "io/ioutil"
)

type Summary struct {
  Gates []*Gate
  Wires []*Wire
  Contexts []*CallingContext
  Outputs []*Wire
}


func RunString(input string, verbose bool) (*Summary, error) {
  result, err := Tokenizer(input)
  if err != nil {
    return nil, err
  }

  if verbose {
    fmt.Println("RESULTS FROM TOKENIZER:")
    PrintAst(result, 0, "")
    fmt.Println()
    fmt.Println()
    fmt.Println()
  }

  stack := []*StackFrame{ &StackFrame{} }

  if result == nil {
    fmt.Println("Result was nil!")
    return nil, nil
  }

  var allGates []*Gate
  var allWires []*Wire
  var allContexts []*CallingContext
  var finalOutputs []*Wire

  resultValues := *result

  for len(resultValues) > 0 {
    if verbose { fmt.Println("==========>", resultValues) }
    gates, wires, contexts, outputs, err := Parse(&resultValues, stack)

    allGates = append(allGates, gates...)
    allWires = append(allWires, wires...)
    allContexts = append(allContexts, contexts...)
    finalOutputs = outputs

    if err != nil {
      return nil, err
    }

    if verbose {
      // fmt.Println("GATES:")
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

        fmt.Printf(` frame=%d`, gate.CallingContext)
        fmt.Printf(` label="%s"`, gate.Label)
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

  if verbose {
    fmt.Println("FINAL OUTPUTS", finalOutputs)
  }

  // Add child contexts to parent contexts
  // This can't be done in `Parse` because it never has a reference to all contexts at once.
  for _, context := range allContexts {
    if context.Parent > 0 {
      for _, parentContext := range allContexts {
        if parentContext.Id == context.Parent {
          parentContext.Children = append(parentContext.Children, context.Id)
          break
        }
      }
    }
  }

  // Print out a summary of the results to that they can be rendered.
  summary := Summary{
    Gates: allGates,
    Wires: allWires,
    Contexts: allContexts,
    Outputs: finalOutputs,
  }

  return &summary, nil
}

func RunFile(path string, verbose bool) (*Summary, error) {
  // Read source code from disk
  source, err := ioutil.ReadFile(path)
  if err != nil {
    return nil, errors.New(fmt.Sprintf("Error reading file %s: %s. Stop.\n", path, err));
  }

  return RunString(string(source), verbose)
}
