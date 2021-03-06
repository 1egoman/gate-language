package main

import (
  "fmt"
  "errors"
  "strings"
)

var wireId int = 0
type Wire struct {
  Id int
  Desc string
  Start *Gate
  End *Gate
  Powered bool
}

type GateType string
const (
  AND GateType = "AND"
  OR = "OR"
  NOT = "NOT"
  SOURCE = "SOURCE"
  GROUND = "GROUND"
  BLOCK_INPUT = "BLOCK_INPUT"
  BLOCK_OUTPUT = "BLOCK_OUTPUT"
  BUILTIN_FUNCTION = "BUILTIN_FUNCTION"
)

var gateId int = 0
type Gate struct {
  Id int
  Type GateType
  Label string

  Inputs []*Wire
  Outputs []*Wire

  // A reference to the id of the block that this gate is within.
  CallingContext int
  State string
}

type Variable struct {
  Name string
  Value *Wire
}

type Block struct {
  Name string
  Content *Node
  InvocationCount int
}

type CallingContext struct {
  Id int
  Name string
  Depth int
  Parent int
  Children []int
}

var stackFrameId int = 0
type StackFrame struct {
  Id int
  Variables []*Variable
  Blocks []*Block
}

var BUILTIN_FUNCTION_NAMES []string =        []string{"led", "wave", "momentary", "toggle", "tflipflop"}
var BUILTIN_FUNCTION_MINIMUM_INPUT_NUMBER []int=[]int{1    , 1     , 0          , 0       , 2}
var BUILTIN_FUNCTION_RETURN_NUMBER []int=       []int{0    , 1     , 1          , 1       , 2}

var INVOCATION_MAX_RECURSION_DEPTH = 100

func Parse(inputs *[]Node, stack []*StackFrame) ([]*Gate, []*Wire, []*CallingContext, []*Wire, error) {
  gates := []*Gate{}
  wires := []*Wire{}
  contexts := []*CallingContext{}
  outputs := []*Wire{}

  input := (*inputs)[0]

  switch input.Token {
  case "SINGLE_COMMENT": fallthrough
  case "MULTI_COMMENT":
    // Remove token that was just parsed.
    *inputs = (*inputs)[1:]
    break

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
      lhsGates, lhsWires, lhsContexts, outputs, err := Parse(&[]Node{lhs}, stack)
      if err != nil {
        return nil, nil, nil, nil, err
      }

      // Ensure that there is only one output from the thing on the left hand side (an and gate can
      // only operate on a single value)
      if len(outputs) > 1 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "Left hand side of %s gate at %d:%d outputs multiple values in a single value context. Stop.",
          input.Token,
          input.Row,
          input.Col,
        ))
      }
      if len(outputs) == 0 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "Left hand side of %s gate at %d:%d outputs zero values in a single value context. Stop.",
          input.Token,
          input.Row,
          input.Col,
        ))
      }
      lhsOutput = outputs[0]

      // Merge all gates from the left hand side with the current gate tree.
      gates = append(gates, lhsGates...)
      wires = append(wires, lhsWires...)
      contexts = append(contexts, lhsContexts...)
    }

    // Parse the right hand side of the gate.
    if rhs, ok := input.Data["RightHandSide"].(Node); ok {
      rhsGates, rhsWires, rhsContexts, outputs, err := Parse(&[]Node{rhs}, stack)
      if err != nil {
        return nil, nil, nil, nil, err
      }

      // Ensure that thre is only one output from the thing on the left hand side (an and gate can
      // only operate on a single value)
      if len(outputs) > 1 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "Right hand side of and gate at %d:%d outputs multiple values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      if len(outputs) == 0 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "Right hand side of and gate at %d:%d outputs zero values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      rhsOutput = outputs[0]

      // Merge all gates from the left hand side with the current gate tree.
      gates = append(gates, rhsGates...)
      wires = append(wires, rhsWires...)
      contexts = append(contexts, rhsContexts...)
    }

    // Add a new wire as output
    wireId += 1
    wire := &Wire{ Id: wireId }
    wires = append(wires, wire)
    outputs = append(outputs, wire)

    // Create the gate, using the wire we just created as the single output of the and gate.
    gateId += 1
    gates = append(gates, &Gate{
      Id: gateId,
      Type: gateType,

      Inputs: append(append([]*Wire{}, lhsOutput), rhsOutput),
      Outputs: []*Wire{ wire },

      // The stack frame that this gate is within
      CallingContext: stack[len(stack)-1].Id,
    })

    // Remove token that was just parsed.
    *inputs = (*inputs)[1:]

  case "OP_NOT":
    var rhsOutput *Wire
    // Parse the right hand side of the gate.
    if rhs, ok := input.Data["RightHandSide"].(Node); ok {
      rhsGates, rhsWires, rhsContexts, outputs, err := Parse(&[]Node{rhs}, stack)
      if err != nil {
        return nil, nil, nil, nil, err
      }

      // Ensure that thre is only one output from the thing on the left hand side (an and gate can
      // only operate on a single value)
      if len(outputs) > 1 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "Right hand side of a not gate at %d:%d outputs multiple values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      if len(outputs) == 0 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "Right hand side of not gate at %d:%d outputs zero values in a single value context. Stop.",
          input.Row,
          input.Col,
        ))
      }
      rhsOutput = outputs[0]

      // Merge all gates from the left hand side with the current gate tree.
      gates = append(gates, rhsGates...)
      wires = append(wires, rhsWires...)
      contexts = append(contexts, rhsContexts...)
    }

    // Add a new wire as output
    wireId += 1
    wire := &Wire{ Id: wireId }
    wires = append(wires, wire)
    outputs = append(outputs, wire)

    // Create the gate, using the wire we just created as the single output of the and gate.
    gateId += 1
    gates = append(gates, &Gate{
      Id: gateId,
      Type: NOT,
      Inputs: append([]*Wire{}, rhsOutput),
      Outputs: []*Wire{ wire },

      // The stack frame that this gate is within
      CallingContext: stack[len(stack)-1].Id,
    })

    // Remove token that was just parsed.
    *inputs = (*inputs)[1:]

  case "ASSIGNMENT":
    // fmt.Printf("/ Assigning! Token = %+v\n", input)
    if names, ok := input.Data["Names"].(string); ok {
      numberOfLhsValues := len(strings.Split(names, " "))
      // fmt.Printf("  * assignment takes %d parameters\n", numberOfLhsValues)

      // First, extract all the tokens after the assignment (rhs) that are assigned to the variabled
      // inside of the assignment (lhs).
      var rhsValues []*Wire
      for len(rhsValues) < numberOfLhsValues {
        // Ensure that the there are still tokens to pull from
        if len(*inputs) <= 1 {
          return nil, nil, nil, nil, errors.New(fmt.Sprintf(
            "Assignment at %d:%d has more variables on the left hand side (%d) than tokens on the right hand side to assign (%d). Stop.",
            input.Row,
            input.Col,
            numberOfLhsValues,
            len(rhsValues),
          ))
        }

        // Get the token after the current token
        parameter := (*inputs)[1]
        // fmt.Printf("  * found new param on rhs: %+v\n", parameter)

        // Verify that the token is of the proper type.
        if !TokenNameIsExtendedExpression(parameter.Token) {
          return nil, nil, nil, nil, errors.New(fmt.Sprintf(
            "Token that is after assignment (assignment is at %d:%d, token is at %d:%d) and trying to be assigned to variable `%s` is not an expression (is %s). Stop.\n",
            input.Row,
            input.Col,
            parameter.Row,
            parameter.Col,
            strings.Split(names, " ")[len(rhsValues)],
            parameter.Token,
          ))
        }

        // Execute it
        paramGates, paramWires, paramContexts, paramOutputs, err := Parse(&[]Node{parameter}, stack)

        // Bubble errors up from the invocation
        if err != nil {
          return nil, nil, nil, nil, err
        }
        // fmt.Printf("  * executed param successfully... %d results.\n", len(paramOutputs))

        // Ensure that the parameter, when evaluated, returns outputs.
        if len(paramOutputs) == 0 {
          // fmt.Printf("PARAM %+v %+v %+v\n", parameter, paramGates, paramWires)
          return nil, nil, nil, nil, errors.New(fmt.Sprintf(
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
        contexts = append(contexts, paramContexts...)

        // Add outputs into the rhs values that are being collected.
        rhsValues = append(rhsValues, paramOutputs...)

        // Finally, delete the parameter value from the node list
        // ie, [input, 1, 2, 3] => [input, 2, 3]
        *inputs = append([]Node{input}, (*inputs)[2:]...)
      }

      for ct, name := range strings.Split(names, " ") {
        // The variable _ is a throwaway value. Any assignments to it should be skipped.
        if name == "_" { continue }

        // See if the variable that's being defined has already been defined in the latest stack
        // frame. If it has, get the wire that it refers to and replace every instance of that wire
        // in all gates with the wire that was just created. This facilitates the creation of
        // "graph" structures (without self referential access like this, the most complex structure
        // that could be created would be a tree)
        for _, variable := range stack[len(stack) - 1].Variables {
          if variable.Name == name {
            wire := rhsValues[ct]
            newWire := variable.Value
            // fmt.Println("* Assigning to variable that already exists:", name, "wire =", wire, "newWire =", newWire)

            // Rewrite all gates that have `wire` to `newWire`
            for ct := 0; ct < len(gates); ct++ {
              // Check inputs
              for inputCt := 0; inputCt < len(gates[ct].Inputs); inputCt += 1 {
                if gates[ct].Inputs[inputCt].Id == wire.Id {
                  gates[ct].Inputs[inputCt] = newWire
                }
              }

              // Check outputs
              for outputCt := 0; outputCt < len(gates[ct].Outputs); outputCt += 1 {
                if gates[ct].Outputs[outputCt].Id == wire.Id {
                  gates[ct].Outputs[outputCt] = newWire
                }
              }
            }

            break
          }
        }

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
      // fmt.Println("Tokens left:", inputs)
    } else {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "The name within the assignment at %d:%d isn't a valid string - got %s. Stop.\n",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }

  case "INVOCATION":
    var block *Block

    if value, ok := input.Data["Name"].(string); ok {
      // Check to see if the invocation refers to a builtin function instead. These are converted
      // into a special gate type, `BUILTIN_FUNCTION`
      for builtinIndex, builtinName := range BUILTIN_FUNCTION_NAMES {
        if value == builtinName {
          var builtinInputs []*Wire = []*Wire{}

          // Add a wire to each input to the `builtinInputs` slice.
          for _, child := range *input.Children {
            // Execute each parameter passed into the invocation to get an output wire to its result.
            paramGates, paramWires, paramContexts, paramOutputs, err := Parse(&[]Node{child}, stack)

            // Bubble errors up from the invocation
            if err != nil {
              return nil, nil, nil, nil, err
            }

            // Add gates and generated to master collections.
            gates = append(gates, paramGates...)
            wires = append(wires, paramWires...)
            contexts = append(contexts, paramContexts...)
            builtinInputs = append(builtinInputs, paramOutputs...)
          }

          // Ensure that the builtin was called with enough parameters
          if len(builtinInputs) < BUILTIN_FUNCTION_MINIMUM_INPUT_NUMBER[builtinIndex] {
            return nil, nil, nil, nil, errors.New(fmt.Sprintf(
              "The buitin block at %d:%d wasn't called with enough parameters (expected at least %d, was called with %d). Stop.",
              input.Row,
              input.Col,
              BUILTIN_FUNCTION_RETURN_NUMBER[builtinIndex],
              len(builtinInputs),
            ))
          }

          // Create a new gate with those inputs from `builtinInputs`
          gateId += 1
          gate := &Gate{
            Id: gateId,
            Type: BUILTIN_FUNCTION,
            Label: builtinName,

            Inputs: builtinInputs,
            Outputs: []*Wire{},

            // The stack frame that this gate is within
            CallingContext: stack[len(stack)-1].Id,
          }
          gates = append(gates, gate)

          // Create a new wire for each output, and add each to the outputs.
          for i := 0; i < BUILTIN_FUNCTION_RETURN_NUMBER[builtinIndex]; i++ {
            wireId += 1
            wire := &Wire{ Id: wireId }
            wires = append(wires, wire)

            gate.Outputs = append(gate.Outputs, wire)
          }

          // Remove token that was just parsed.
          *inputs = (*inputs)[1:]

          return gates, wires, contexts, gate.Outputs, nil
        }
      }
      // (end builtin code)

      // Look through the stack, from top to bottom, to find an identifier that matches.
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
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "The invocation at %d:%d (trying to invoke %s) doesn't invoke a block that can be found in the current or any parent scope. Stop.\n",
          input.Row,
          input.Col,
          value,
        ))
      }

      // fmt.Println("/ Invoking block: ", block)

      // Increment the invocation count for the block
      block.InvocationCount += 1

      // For each parameter passed into the invocation, execute it and get a reference to it to link
      // to each value that is in the context of the invocation.
      var vars []*Variable
      for ct, child := range *input.Children {
        // Execute each parameter passed into the invocation to get an output wire to its result.
        paramGates, paramWires, paramContexts, paramOutputs, err := Parse(&[]Node{child}, stack)

        // Bubble errors up from the invocation
        if err != nil {
          return nil, nil, nil, nil, err
        }

        // Add gates and generated to master collections.
        gates = append(gates, paramGates...)
        wires = append(wires, paramWires...)
        contexts = append(contexts, paramContexts...)

        // But add each output from invoking into the variables slice to use to perform the actual
        // invocation, joining through a special type of gate called "BLOCK_INPUT" to denote that
        // we're entering a block.
        for _, output := range paramOutputs {
          numberOfVars := len(vars) - 1

          // Create a wire to join between the block input node and the bound variable
          wireId += 1
          wire := &Wire{ Id: wireId }
          wires = append(wires, wire)

          // Create a new block input gate to express that we're entering a block.
          gateId += 1
          gates = append(gates, &Gate{
            Id: gateId,
            Type: BLOCK_INPUT,
            Label: fmt.Sprintf("Input %d into block %s invocation %d", ct, block.Name, block.InvocationCount),
            Inputs: []*Wire{output}, /* parameter => BLOCK_INPUT */
            Outputs: []*Wire{wire}, /* BLOCK_INPUT => variable bound in local scope */

            // The id of the new stack frame that is about to be created.
            CallingContext: stackFrameId + 1,
          })

          params := strings.Split(block.Content.Data["Params"].(string), " ")
          if (len(params) - 1) < numberOfVars + 1 {
            fmt.Println(block.Content)
            return nil, nil, nil, nil, errors.New(fmt.Sprintf(
              "The invocation at %d:%d (trying to invoke %s) is invoking the block with too many parameters (expected %d, received %d). Stop.\n",
              input.Row,
              input.Col,
              block.Name,
              block.Content.Data["InputQuantity"],
              len(*input.Children),
            ))
          }
          vars = append(vars, &Variable{
            Name: params[numberOfVars+1],
            Value: wire,
          })
        }
      }

      var deref_vars []Variable
      for _, v := range vars { deref_vars = append(deref_vars, *v) }
      // fmt.Printf("  * Created variables to inject into scope: %+v\n", deref_vars)

      // Add a temporary item to the top of the stack for the invocation, defining all the variables
      // that were passed in as parameters as defines in the new stack frame. Also, add a new block
      // called `__self` tht points to the current block. This allows other functions later on to
      // get the reference to the block that it is contained within (one example is BLOCK_RETURN).
      stackFrameId += 1
      invocationStack := append(stack, &StackFrame{
        Id: stackFrameId,
        Variables: vars,
        Blocks: []*Block{
          &Block{Name: "__self", Content: block.Content},
        },
      })

      // Verify that the user hasn't called deeper into the stack then they should
      if INVOCATION_MAX_RECURSION_DEPTH > 0 && len(invocationStack) > INVOCATION_MAX_RECURSION_DEPTH {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
          "The invocation at %d:%d (trying to invoke %s) has surpassed the max call depth of %d. Stop.\n",
          input.Row,
          input.Col,
          block.Name,
          INVOCATION_MAX_RECURSION_DEPTH,
        ))
      }

      // Also, take note of the new calling context we've created. We're recording mostly the same
      // information that was put onto the stack. but this collection is insert only.
      parentContextId := invocationStack[(len(invocationStack) - 1) - 1].Id
      contexts = append(contexts, &CallingContext{
        Id: stackFrameId,
        Name: block.Name,
        Depth: len(invocationStack) - 1,
        Parent: parentContextId,
        Children: []int{},
      });

      // Make a copy of the children within the block so that they can be destructively mutated
      // by the parser without changing the actual contents of the block.
      blockChildren := *block.Content.Children

      // Execute the invocation
      for len(blockChildren) > 0 {
        headToken := blockChildren[0].Token

        invocationResultGates, invocationResultWires, invocationResultContexts, invocationResultOutputs, err := Parse(
          &blockChildren,
          invocationStack,
        )

        // Bubble errors up from the invocation
        if err != nil {
          return nil, nil, nil, nil, err
        }

        gates = append(gates, invocationResultGates...)
        wires = append(wires, invocationResultWires...)
        contexts = append(contexts, invocationResultContexts...)

        // If a return token is found, take each value that is outputted, connect it to a
        // `BLOCK_OUTPUT` node, and put the output wire of that `BLOCK_OUTPUT` node in the outputs
        // for this action.
        if headToken == "BLOCK_RETURN" {
          // fmt.Println("  * block has return!")
          for ct, output := range invocationResultOutputs {
            // Create a wire to join between the block output node and the bound variable
            wireId += 1
            wire := &Wire{ Id: wireId }
            wires = append(wires, wire)

            // Create a new block output gate to express that we're leaving a block.
            gateId += 1
            gates = append(gates, &Gate{
              Id: gateId,
              Type: BLOCK_OUTPUT,
              Label: fmt.Sprintf("Output %d from block %s invocation %d", ct, block.Name, block.InvocationCount),
              Inputs: []*Wire{output}, /* parameter => BLOCK_OUTPUT */
              Outputs: []*Wire{wire}, /* BLOCK_OUTPUT => variable bound in local scope */

              // The stack frame that this gate is within
              CallingContext: invocationStack[len(invocationStack)-1].Id,
            })

            // Add the output wire to the outputs for the block.
            outputs = append(outputs, wire)
          }
        }
      }


      // fmt.Println("\\ Done Invoking block: ", block)
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
    // fmt.Println("* block return was able to find parent function (ie, 'self'):", self)

    // Ensure that the parent of the currently invoked function exists.
    if self == nil {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "Couldn't find the parent of the currently invoked function, on %d:%d.",
        input.Row,
        input.Col,
      ))
    }

    // Fetch the number of outputs required from the block we're within.
    numberOfOutputs := self.Data["OutputQuantity"].(int)
    // fmt.Println("* block return expects this many outputs:", numberOfOutputs)

    for len(outputs) < numberOfOutputs {
      // Get the token after the current token
      parameter := (*inputs)[1]
      // fmt.Printf("  * found new token after return: %+v\n", parameter)

      // Execute it
      paramGates, paramWires, paramContexts, paramOutputs, err := Parse(&[]Node{parameter}, stack)

      // Bubble errors up from the invocation
      if err != nil {
        return nil, nil, nil, nil, err
      }
      // fmt.Printf("  * executed token successfully... %d results.\n", len(paramOutputs))

      // Ensure that the parameter, when evaluated, returns outputs.
      if len(paramOutputs) == 0 {
        return nil, nil, nil, nil, errors.New(fmt.Sprintf(
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
      contexts = append(contexts, paramContexts...)

      // Add outputs into the rhs values that are being collected.
      outputs = append(outputs, paramOutputs...)

      // Finally, delete the parameter value from the end
      // ie, [block_return, 1, 2, 3] => [block_return, 2, 3]
      *inputs = append([]Node{input}, (*inputs)[2:]...)
    }
    // fmt.Println("* Found all block return tokens, added each output wire to the outputs of the block return.")

    // There should be no tokens left in the input array after the block return that haven't already
    // been parsed.
    if !( len(*inputs) == 1 && (*inputs)[0].Token == "BLOCK_RETURN" ) {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "Block %s at %d:%d has too many return values, expected %d, got %d. Stop.\n",
        self.Data["Name"],
        input.Row,
        input.Col,
        numberOfOutputs,
        numberOfOutputs + len(*inputs),
      ))
    }

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
      // if wire == nil {
      //   return nil, nil, nil, errors.New(fmt.Sprintf(
      //     "The variable `%s` found at %d:%d could not be found in the stack (did you assign it before usign it?). Stop.\n",
      //     value,
      //     input.Row,
      //     input.Col,
      //   ))
      // }

      if wire == nil {
        // Make a new wire
        wireId += 1
        wire = &Wire{Id: wireId, Desc: fmt.Sprintf("for implicitly declared variable %s", value)}

        // Implicity declare a variable linked to that wire
        stack[len(stack) - 1].Variables = append(stack[len(stack) - 1].Variables, &Variable{
          Name: value,
          Value: wire,
        })
      }

      // Add wire to all wires, and to output.
      wires = append(wires, wire)
      outputs = append(outputs, wire)

      // Remove token that was just parsed.
      *inputs = (*inputs)[1:]
    } else {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "The value within the identifier at %d:%d isn't a valid stril - got %s. Stop.",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }

  case "GROUP":
    if input.Children == nil {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "The children attribute within the group at %d:%d is nil. Stop.",
        input.Row,
        input.Col,
      ))
    }

    for _, child := range *input.Children {
      childGates, childWires, childContexts, childOutputs, err := Parse(&[]Node{child}, stack)
      if err != nil {
        return nil, nil, nil, nil, err
      }
      gates = append(gates, childGates...)
      wires = append(wires, childWires...)
      contexts = append(contexts, childContexts...)
      outputs = append(outputs, childOutputs...)
    }

    // Remove token that was just parsed.
    *inputs = (*inputs)[1:]

  case "BLOCK":
    if name, ok := input.Data["Name"].(string); ok {
      // Add the block to the latest stackframe, in the blocks section.
      stack[len(stack) - 1].Blocks = append(stack[len(stack) - 1].Blocks, &Block{
        Name: name,
        Content: &input,
      })

      // Remove token that was just parsed.
      *inputs = (*inputs)[1:]
    } else {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "The block at %d:%d doesn't have a name, instead found %s. Stop.",
        input.Row,
        input.Col,
        input.Data["Name"],
      ))
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
      gateId += 1
      gates = append(gates, &Gate{
        Id: gateId,
        Type: gateType,
        Inputs: []*Wire{},
        Outputs: []*Wire{wire},

        // The stack frame that this gate is within
        CallingContext: stack[len(stack)-1].Id,
      })

      // Remove token that was just parsed.
      *inputs = (*inputs)[1:]
    } else {
      return nil, nil, nil, nil, errors.New(fmt.Sprintf(
        "The value within the boolean at %d:%d isn't true or false - got %s. Stop.",
        input.Row,
        input.Col,
        input.Data["Value"],
      ))
    }

  default:
    return nil, nil, nil, nil, errors.New(fmt.Sprintf(
      "Unknown token at %d:%d - %s. Stop.\n",
      input.Row,
      input.Col,
      input.Token,
    ))
  }

  return gates, wires, contexts, outputs, nil
}

