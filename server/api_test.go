package main

import (
  "testing"
  "fmt"
  "reflect"
)

func TestRunString(t *testing.T) {
  wireId = 0
  gateId = 0
  stackFrameId = 0

  summary, err := RunString("led(toggle())", false)
  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify summary
  if !reflect.DeepEqual(summary.Contexts, Summary{
    Gates: []*Gate{
      &Gate{
        Id: 1,
        Type: BUILTIN_FUNCTION,
        Label: "toggle",
        Inputs: []*Wire{},
        Outputs: []*Wire{ &Wire{Id: 1} },
        CallingContext: 0,
      },
      &Gate{
        Id: 2,
        Type: BUILTIN_FUNCTION,
        Label: "led",
        Inputs: []*Wire{ &Wire{Id: 1} },
        Outputs: []*Wire{},
        CallingContext: 0,
      },
    },
    Wires: []*Wire{
      &Wire{Id: 1},
    },
    Outputs: []*Wire{},
  }.Contexts) {
    t.Error("Summary doesn't match!")
    t.Error(fmt.Sprintf("Gates: %+v", summary.Gates))
    t.Error(fmt.Sprintf("Wires: %+v", summary.Wires))
    t.Error(fmt.Sprintf("Contexts: %+v", summary.Contexts))
    t.Error(fmt.Sprintf("Outputs: %+v", summary.Outputs))
  }
}

func TestRunStringError(t *testing.T) {
  wireId = 0
  gateId = 0
  stackFrameId = 0

  _, err := RunString("syntax error 5", false)

  // Verify error was returned
  if err == nil {
    t.Errorf("No Error returned!")
    return
  }
}
