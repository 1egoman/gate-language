package main

import (
  "fmt"
  "errors"
)

type Wire struct {
  Id int
  Desc string
}
var wireId int = 0

type GateType string
const (
  AND GateType = "AND"
  OR = "OR"
  NOT = "NOT"
  SOURCE = "SOURCE"
  GROUND = "GROUND"
)

type Gate struct {
  Type GateType
  Label string
  Inputs []*Wire
  Outputs []*Wire
}

type Variable struct {
  Name string
  Value *Wire
}

type StackFrame struct {
  Variables []*Variable
}

func Parse(input *Node, stack []*StackFrame) ([]*Gate, []*Wire, []*Wire, error) {
  gates := []*Gate{}
  wires := []*Wire{}
  outputs := []*Wire{}

  switch input.Token {
  case "OP_AND": fallthrough
  case "OP_OR":
    var lhsOutput *Wire
    var rhsOutput *Wire

    var gateType GateType
    if input.Token == "OP_AND" {
      gateType = AND
    } else {
      gateType = OR
    }

    // Parse the left hand side of the gate.
    if lhs, ok := input.Data["LeftHandSide"].(Node); ok {
      lhsGates, lhsWires, outputs, err := Parse(&lhs, stack)
      if err != nil {
        return nil, nil, nil, err
      }

      // Ensure that there is only one output from the thing on the left hand side (an and gate can
      // only operate on a single value)
      if len(outputs) > 1 {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "Left hand side of and gate at %d:%d outputs multiple values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      if len(outputs) == 0 {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "Left hand side of and gate at %d:%d outputs zero values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      lhsOutput = outputs[0]

      // Merge all gates from the left hand side with the current gate tree.
      gates = append(gates, lhsGates...)
      wires = append(wires, lhsWires...)
    }

    // Parse the right hand side of the gate.
    if rhs, ok := input.Data["RightHandSide"].(Node); ok {
      rhsGates, rhsWires, outputs, err := Parse(&rhs, stack)
      if err != nil {
        return nil, nil, nil, err
      }

      // Ensure that thre is only one output from the thing on the left hand side (an and gate can
      // only operate on a single value)
      if len(outputs) > 1 {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "Right hand side of and gate at %d:%d outputs multiple values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      if len(outputs) == 0 {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "Right hand side of and gate at %d:%d outputs zero values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      rhsOutput = outputs[0]

      // Merge all gates from the left hand side with the current gate tree.
      gates = append(gates, rhsGates...)
      wires = append(wires, rhsWires...)
    }

    // Add a new wire as output
    wireId += 1
    wire := &Wire{ Id: wireId }
    wires = append(wires, wire)
    outputs = append(outputs, wire)

    // Create the gate, using the wire we just created as the single output of the and gate.
    gates = append(gates, &Gate{
      Type: gateType,
      Inputs: append(append([]*Wire{}, lhsOutput), rhsOutput),
      Outputs: []*Wire{ wire },
    })

  case "IDENTIFIER":
    if value, ok := input.Data["Value"].(string); ok {
      // Look through the stack, from top to bottom, to find an identifier that matches.
      var wire *Wire

      IdentifierOuter:
      for i := len(stack) - 1; i >= 0; i-- {
        for _, variable := range stack[i].Variables {
          if variable.Name == value {
            wire = variable.Value
            break IdentifierOuter;
          }
        }
      }

      // Ensure that a variable was found.
      if wire == nil {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "The variable `%s` found at %d:%d could not be found in the stack (did you assign it before usign it?). Stop.",
          value,
          input.Row,
          input.Col,
          input.Data["Value"],
        ))
      }

      // Add wire to all wires, and to output.
      wires = append(wires, wire)
      outputs = append(outputs, wire)
    } else {
      return nil, nil, nil, errors.New(fmt.Sprintf(
        "The value within the identifier at %d:%d isn't a valid stril - got %s. Stop.",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }

  case "GROUP":
    if input.Children == nil {
      return nil, nil, nil, errors.New(fmt.Sprintf(
        "The children attribute within the group at %d:%d is nil. Stop.",
        input.Row,
        input.Col,
      ))
    }

    for _, child := range *input.Children {
      childGates, childWires, childOutputs, err := Parse(&child, stack)
      if err != nil {
        return nil, nil, nil, err
      }
      gates = append(gates, childGates...)
      wires = append(wires, childWires...)
      outputs = append(outputs, childOutputs...)
    }

  case "BOOL":
    if value, ok := input.Data["Value"].(bool); ok {
      // Figure out the type of signal we have
      var gateType GateType
      if value {
        gateType = SOURCE
      } else {
        gateType = GROUND
      }

      // Add a new wire connected to voltage or ground
      wireId += 1
      wire := &Wire{ Id: wireId }
      wires = append(wires, wire)

      // The wire is also an output of the bool, so add it to the outputs
      outputs = append(outputs, wire)

      // Create a gate that represents voltage or ground that the wire attaches to.
      gates = append(gates, &Gate{
        Type: gateType,
        Inputs: []*Wire{},
        Outputs: []*Wire{wire},
      })
    } else {
      return nil, nil, nil, errors.New(fmt.Sprintf(
        "The value within the boolean at %d:%d isn't true or false - got %s. Stop.",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }
  }

  return gates, wires, outputs, nil
}

