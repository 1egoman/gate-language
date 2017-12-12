package main

import (
  "fmt"
  "errors"
  "strings"
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

type Block struct {
  Name string
  Content *Node
}

type StackFrame struct {
  Variables []*Variable
  Blocks []*Block
}

func Parse(inputs *[]Node, stack []*StackFrame) ([]*Gate, []*Wire, []*Wire, error) {
  gates := []*Gate{}
  wires := []*Wire{}
  outputs := []*Wire{}

  input := (*inputs)[0]

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
      lhsGates, lhsWires, outputs, err := Parse(&[]Node{lhs}, stack)
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
      rhsGates, rhsWires, outputs, err := Parse(&[]Node{rhs}, stack)
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

    // Remove token that was just parsed.
    *inputs = (*inputs)[1:]

  case "ASSIGNMENT":
    fmt.Printf("> Assigning! Token = %+v\n", input)
    if names, ok := input.Data["Names"].(string); ok {
      numberOfLhsValues := len(strings.Split(names, " "))
      fmt.Printf("  * assignment takes %d parameters\n", numberOfLhsValues)

      // First, extract all the tokens after the assignment (rhs) that are assigned to the variabled
      // inside of the assignment (lhs).
      var rhsValues []*Wire
      for len(rhsValues) < numberOfLhsValues {
        // Get the token after the current token
        parameter := (*inputs)[1]
        fmt.Printf("  * found new param on rhs: %+v\n", parameter)

        // Execute it
        paramGates, paramWires, paramOutputs, err := Parse(&[]Node{parameter}, stack)

        // Bubble errors up from the invocation
        if err != nil {
          return nil, nil, nil, err
        }
        fmt.Printf("  * executed param successfully... %d results.\n", len(paramOutputs))

        // Ensure that the parameter, when evaluated, returns outputs.
        if len(paramOutputs) == 0 {
          return nil, nil, nil, errors.New(fmt.Sprintf(
            "Parameter to assignment (assignment located at %d:%d, parameter located at %d:%d) outputted no values after being evaluated, please remove from assignment. Stop.\n",
            input.Row,
            input.Col,
            parameter.Row,
            parameter.Col,
          ))
        }

        // Add gates and generated to master collections.
        gates = append(gates, paramGates...)
        wires = append(wires, paramWires...)

        // Add outputs into the rhs values that are being collected.
        rhsValues = append(rhsValues, paramOutputs...)

        // Finally, delete the parameter value from the node list
        // ie, [input, 1, 2, 3] => [input, 2, 3]
        *inputs = append([]Node{input}, (*inputs)[2:]...)
      }

      for ct, name := range strings.Split(names, " ") {
        // Add a new variable on the stack that is linked to the value with the same index after the
        // assignment. ie, let a b = 1 0 means to create a wire between `a` and `1`, and to create a
        // wire between `b` and `0`.
        stack[len(stack) - 1].Variables = append(stack[len(stack) - 1].Variables, &Variable{
          Name: name,
          Value: rhsValues[ct],
        })

        wires = append(wires, rhsValues[ct])
      }

      // Remove token that was just parsed.
      *inputs = (*inputs)[1:]
      fmt.Println("Tokens left:", inputs)
    } else {
      return nil, nil, nil, errors.New(fmt.Sprintf(
        "The name within the assignment at %d:%d isn't a valid string - got %s. Stop.\n",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }

  case "INVOCATION":
    // Look through the stack, from top to bottom, to find an identifier that matches.
    var block *Block

    if value, ok := input.Data["Name"].(string); ok {
      BlockOuter:
      for i := len(stack) - 1; i >= 0; i-- {
        for _, blk := range stack[i].Blocks {
          if blk.Name == value {
            block = blk
            break BlockOuter;
          }
        }
      }

      // Ensure that the invokation is inkoving something that can be invoked.
      if block == nil {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "The invocation at %d:%d (trying to invoke %s) doesn't invoke a block that can be found in the current or any parent scope. Stop.\n",
          input.Row,
          input.Col,
          value,
        ))
      }

      fmt.Println("> Invoking block: ", block)

      // For each parameter passed into the invocation, execute it and get a reference to it to link
      // to each value that is in the context of the invocation.
      var vars []*Variable
      for _, child := range *input.Children {
        // Execute each parameter passed into the invocation to get an output wire to its result.
        paramGates, paramWires, paramOutputs, err := Parse(&[]Node{child}, stack)

        // Bubble errors up from the invocation
        if err != nil {
          return nil, nil, nil, err
        }

        // Add gates and generated to master colelctions.
        gates = append(gates, paramGates...)
        wires = append(wires, paramWires...)

        // But add each output from invoking into the variables slice to use to perform the actual
        // invocation.
        for _, output := range paramOutputs {
          numberOfVars := len(vars) - 1
          vars = append(vars, &Variable{
            // Name: fmt.Sprintf("__value_%d_passed_into_%s", ct, block.Name),
            Name: strings.Split(block.Content.Data["Params"].(string), " ")[numberOfVars+1],
            Value: output,
          })
        }
      }

      // Add a temporary item to the top of the stack for the invocation, defining all the variables
      // that were passed in as parameters as defines in the new stack frame. Also, add a new block
      // called `__self` tht points to the current block. This allows other functions later on to
      // get the reference to the block that it is contained within (one example is BLOCK_RETURN).
      invocationStack := append(stack, &StackFrame{
        Variables: vars,
        Blocks: []*Block{
          &Block{Name: "__self", Content: block.Content},
        },
      })

      // Execute the invocation
      for len(*block.Content.Children) > 0 {
        headToken := (*block.Content.Children)[0].Token

        invocationResultGates, invocationResultWires, invocationResultOutputs, err := Parse(
          block.Content.Children,
          invocationStack,
        )

        // Bubble errors up from the invocation
        if err != nil {
          return nil, nil, nil, err
        }

        gates = append(gates, invocationResultGates...)
        wires = append(wires, invocationResultWires...)

        if headToken == "BLOCK_RETURN" {
          outputs = append(outputs, invocationResultOutputs...)
        }
      }

      // Remove token that was just parsed.
      *inputs = (*inputs)[:len(*inputs) - 1]
    }

  case "BLOCK_RETURN":
    // Look for the special block defined called `__self`, and figure out how many of the next
    // tokens to evaluate and add as outputs.
    var self *Node
    for _, potentialSelf := range stack[len(stack) - 1].Blocks {
      if potentialSelf.Name == "__self" {
        self = potentialSelf.Content
        break
      }
    }
    fmt.Println("* block return was able to find parent function (ie, 'self'):", self)

    // Ensure that the parent of the currently invoked function exists.
    if self == nil {
      return nil, nil, nil, errors.New(fmt.Sprintf(
        "Couldn't find the parent of the currently invoked function, on %d:%d.",
        input.Row,
        input.Col,
      ))
    }

    // Fetch the number of outputs required from the block we're within.
    numberOfOutputs := self.Data["OutputQuantity"].(int)
    fmt.Println("* block return expects this many outputs:", numberOfOutputs)

    // Fetch 
    for len(outputs) < numberOfOutputs {
      // Get the token after the current token
      parameter := (*inputs)[1]
      fmt.Printf("  * found new token after return: %+v\n", parameter)

      // Execute it
      paramGates, paramWires, paramOutputs, err := Parse(&[]Node{parameter}, stack)

      // Bubble errors up from the invocation
      if err != nil {
        return nil, nil, nil, err
      }
      fmt.Printf("  * executed token successfully... %d results.\n", len(paramOutputs))

      // Ensure that the parameter, when evaluated, returns outputs.
      if len(paramOutputs) == 0 {
        return nil, nil, nil, errors.New(fmt.Sprintf(
          "Parameter to assignment (assignment located at %d:%d, parameter located at %d:%d) outputted no values after being evaluated, please remove from assignment. Stop.\n",
          input.Row,
          input.Col,
          parameter.Row,
          parameter.Col,
        ))
      }

      // Add gates and generated to master collections.
      gates = append(gates, paramGates...)
      wires = append(wires, paramWires...)

      // Add outputs into the rhs values that are being collected.
      outputs = append(outputs, paramOutputs...)

      // Finally, delete the parameter value from the end
      // ie, [block_return, 1, 2, 3] => [block_return, 2, 3]
      *inputs = append([]Node{input}, (*inputs)[2:]...)
    }
    fmt.Println("* Found all block return tokens, added each output wire to the outputs of the block return.")

    *inputs = (*inputs)[:len(*inputs) - 1]

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
          "The variable `%s` found at %d:%d could not be found in the stack (did you assign it before usign it?). Stop.\n",
          value,
          input.Row,
          input.Col,
        ))
      }

      // Add wire to all wires, and to output.
      wires = append(wires, wire)
      outputs = append(outputs, wire)

      // Remove token that was just parsed.
      *inputs = (*inputs)[1:]
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
      childGates, childWires, childOutputs, err := Parse(&[]Node{child}, stack)
      if err != nil {
        return nil, nil, nil, err
      }
      gates = append(gates, childGates...)
      wires = append(wires, childWires...)
      outputs = append(outputs, childOutputs...)
    }

    // Remove token that was just parsed.
    *inputs = (*inputs)[1:]

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

      // Remove token that was just parsed.
      *inputs = (*inputs)[1:]
    } else {
      return nil, nil, nil, errors.New(fmt.Sprintf(
        "The value within the boolean at %d:%d isn't true or false - got %s. Stop.",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }

  default:
    return nil, nil, nil, errors.New(fmt.Sprintf(
      "Unknown token at %d:%d - %s. Stop.\n",
      input.Row,
      input.Col,
      input.Token,
    ))
  }

  return gates, wires, outputs, nil
}

