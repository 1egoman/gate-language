package main

import (
  "testing"
  "fmt"
  "reflect"
)

func TestParsingAnd(t *testing.T) {
  ast := Node{Token: "OP_AND", Row: 3, Col: 1, Data: map[string]interface{}{
    "LeftHandSide": Node{Token: "BOOL", Row: 1, Col: 1, Data: map[string]interface{}{"Value": true}},
    "RightHandSide": Node{Token: "BOOL", Row: 7, Col: 1, Data: map[string]interface{}{"Value": false}},
  }}

  stack := []*StackFrame{
    &StackFrame{
      Variables: []*Variable{},
    },
  }

  gates, wires, outputs, err := Parse(&[]Node{ast}, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
    &Gate{
      Type: GROUND,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 2} },
    },
    &Gate{
      Type: AND,
      Inputs: []*Wire{ &Wire{Id: 1}, &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 3} },
    },
  }) {
    t.Error(fmt.Sprintf("Gates doesn't match! %+v", gates))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: 1},
    &Wire{Id: 2},
    &Wire{Id: 3},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", gates))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{
    &Wire{Id: 3},
  }) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }
}

// a and false (where a is already on the stack)
func TestParsingVariable(t *testing.T) {
  ast := Node{Token: "OP_AND", Row: 3, Col: 1, Data: map[string]interface{}{
    "LeftHandSide": Node{Token: "IDENTIFIER", Row: 1, Col: 1, Data: map[string]interface{}{"Value": "a"}},
    "RightHandSide": Node{Token: "BOOL", Row: 7, Col: 1, Data: map[string]interface{}{"Value": false}},
  }}

  stack := []*StackFrame{
    &StackFrame{
      Variables: []*Variable{
        &Variable{Name: "a", Value: &Wire{Id: -1}},
      },
    },
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&[]Node{ast}, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: GROUND,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
    &Gate{
      Type: AND,
      Inputs: []*Wire{ &Wire{Id: -1}, &Wire{Id: 1} },
      Outputs: []*Wire{ &Wire{Id: 2} },
    },
  }) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: -1},
    &Wire{Id: 1},
    &Wire{Id: 2},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{
    &Wire{Id: 2},
  }) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }
}

// let a = 1
func TestAssigningVariable(t *testing.T) {
  ast := []Node{
    Node{Token: "ASSIGNMENT", Row: 3, Col: 1, Data: map[string]interface{}{"Names": "a", "Values": []Node{}}},
    Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": true}},
  }

  stack := []*StackFrame{
    &StackFrame{
      Variables: []*Variable{
        &Variable{Name: "a", Value: &Wire{Id: -1}},
      },
    },
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&ast, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
  }) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: 1},
    &Wire{Id: 1},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{}) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }
}

// let a = foo(1)
func TestAssigningVariableToInvokedBlock(t *testing.T) {
  ast := []Node{
    Node{Token: "ASSIGNMENT", Row: 3, Col: 1, Data: map[string]interface{}{"Names": "a", "Values": []Node{}}},
    Node{
      Token: "INVOCATION",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": true}},
      },
    },
  }

  stack := []*StackFrame{
    &StackFrame{
      Blocks: []*Block{
        &Block{
          Name: "foo",
          Content: &Node{
            Token: "BLOCK",
            Data: map[string]interface{}{
              "Name": "foo",
              "Params": "a",
              "InputQuantity": 1,
              "OutputQuantity": 1,
            },
            Children: &[]Node{
              Node{Token: "BLOCK_RETURN"},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
            },
          },
        },
      },
    },
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&ast, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
    &Gate{
      Type: BLOCK_INPUT,
      Label: "Input 0 into block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 1} },
      Outputs: []*Wire{ &Wire{Id: 2} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 0 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 3} },
    },
  }) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: 1},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 3},
    &Wire{Id: 3},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{}) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }

  // Ensure that the right variables are on the stack
  for _, variable := range stack[0].Variables {
    if variable.Name == "a" {
      if variable.Value.Id != 3 {
        t.Errorf("Variable a is attached to the wrong wire!")
      }
      continue
    }
    t.Errorf(fmt.Sprintf("Unknown variable %s!", variable.Name))
  }
}

// let a b = foo(1)
func TestAssigningVariableToInvokedBlockWithMultipleValues(t *testing.T) {
  ast := []Node{
    Node{Token: "ASSIGNMENT", Row: 3, Col: 1, Data: map[string]interface{}{"Names": "a b", "Values": []Node{}}},
    Node{
      Token: "INVOCATION",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": true}},
      },
    },
  }

  stack := []*StackFrame{
    &StackFrame{
      Blocks: []*Block{
        &Block{
          Name: "foo",
          Content: &Node{
            Token: "BLOCK",
            Data: map[string]interface{}{
              "Name": "foo",
              "Params": "a",
              "InputQuantity": 1,
              "OutputQuantity": 2,
            },
            Children: &[]Node{
              Node{Token: "BLOCK_RETURN"},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
            },
          },
        },
      },
    },
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&ast, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
    &Gate{
      Type: BLOCK_INPUT,
      Label: "Input 0 into block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 1} },
      Outputs: []*Wire{ &Wire{Id: 2} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 0 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 3} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 1 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 4} },
    },
  }) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: 1},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 3},
    &Wire{Id: 4},
    &Wire{Id: 3},
    &Wire{Id: 4},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{}) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }

  // Ensure that the right variables are on the stack
  for _, variable := range stack[0].Variables {
    if variable.Name == "a" {
      if variable.Value.Id != 3 {
        t.Errorf("Variable a is attached to the wrong wire!")
      }
      continue
    }
    if variable.Name == "b" {
      if variable.Value.Id != 4 {
        t.Errorf("Variable b is attached to the wrong wire!")
      }
      continue
    }
    t.Errorf(fmt.Sprintf("Unknown variable %s!", variable.Name))
  }
}

// let a b c = foo(1) 1
func TestAssigningVariableToInvokedBlockWithMultipleValuesAcrossMultipleTokens(t *testing.T) {
  ast := []Node{
    Node{
      Token: "ASSIGNMENT",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{
        "Names": "a b c",
        "Values": []Node{},
      },
    },
    Node{
      Token: "INVOCATION",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": true}},
      },
    },
    Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": true}},
  }

  stack := []*StackFrame{
    &StackFrame{
      Blocks: []*Block{
        &Block{
          Name: "foo",
          Content: &Node{
            Token: "BLOCK",
            Data: map[string]interface{}{
              "Name": "foo",
              "Params": "a",
              "InputQuantity": 1,
              "OutputQuantity": 2,
            },
            Children: &[]Node{
              Node{Token: "BLOCK_RETURN"},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
            },
          },
        },
      },
    },
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&ast, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
    &Gate{
      Type: BLOCK_INPUT,
      Label: "Input 0 into block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 1} },
      Outputs: []*Wire{ &Wire{Id: 2} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 0 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 3} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 1 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 4} },
    },
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 5} },
    },
  }) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: 1},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 3},
    &Wire{Id: 4},
    &Wire{Id: 5},
    &Wire{Id: 3},
    &Wire{Id: 4},
    &Wire{Id: 5},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Ensure that the right variables are on the stack
  for _, variable := range stack[0].Variables {
    if variable.Name == "a" {
      if variable.Value.Id != 3 {
        t.Errorf("Variable a is attached to the wrong wire!")
      }
      continue
    }
    if variable.Name == "b" {
      if variable.Value.Id != 4 {
        t.Errorf("Variable b is attached to the wrong wire!")
      }
      continue
    }
    if variable.Name == "c" {
      if variable.Value.Id != 5 {
        t.Errorf("Variable c is attached to the wrong wire!")
      }
      continue
    }

    t.Errorf(fmt.Sprintf("Unknown variable %s!", variable.Name))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{}) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }
}

// let a b c d = foo(1) foo(0)
func TestAssigningVariableToInvokedBlockWithMultipleValuesAcrossMultipleInvocations(t *testing.T) {
  ast := []Node{
    Node{
      Token: "ASSIGNMENT",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{
        "Names": "a b c d",
        "Values": []Node{},
      },
    },
    Node{
      Token: "INVOCATION",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": true}},
      },
    },
    Node{
      Token: "INVOCATION",
      Row: 3,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{Token: "BOOL", Row: 3, Col: 1, Data: map[string]interface{}{"Value": false}},
      },
    },
  }

  stack := []*StackFrame{
    &StackFrame{
      Blocks: []*Block{
        &Block{
          Name: "foo",
          Content: &Node{
            Token: "BLOCK",
            Data: map[string]interface{}{
              "Name": "foo",
              "Params": "a",
              "InputQuantity": 1,
              "OutputQuantity": 2,
            },
            Children: &[]Node{
              Node{Token: "BLOCK_RETURN"},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
              Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
            },
          },
        },
      },
    },
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&ast, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{
    &Gate{
      Type: SOURCE,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 1} },
    },
    &Gate{
      Type: BLOCK_INPUT,
      Label: "Input 0 into block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 1} },
      Outputs: []*Wire{ &Wire{Id: 2} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 0 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 3} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 1 from block foo invocation 1",
      Inputs: []*Wire{ &Wire{Id: 2} },
      Outputs: []*Wire{ &Wire{Id: 4} },
    },
    &Gate{
      Type: GROUND,
      Inputs: []*Wire{},
      Outputs: []*Wire{ &Wire{Id: 5} },
    },
    &Gate{
      Type: BLOCK_INPUT,
      Label: "Input 0 into block foo invocation 2",
      Inputs: []*Wire{ &Wire{Id: 5} },
      Outputs: []*Wire{ &Wire{Id: 6} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 0 from block foo invocation 2",
      Inputs: []*Wire{ &Wire{Id: 6} },
      Outputs: []*Wire{ &Wire{Id: 7} },
    },
    &Gate{
      Type: BLOCK_OUTPUT,
      Label: "Output 1 from block foo invocation 2",
      Inputs: []*Wire{ &Wire{Id: 6} },
      Outputs: []*Wire{ &Wire{Id: 8} },
    },
  }) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{
    &Wire{Id: 1},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 2},
    &Wire{Id: 3},
    &Wire{Id: 4},
    &Wire{Id: 5},
    &Wire{Id: 6},
    &Wire{Id: 6},
    &Wire{Id: 6},
    &Wire{Id: 7},
    &Wire{Id: 8},
    &Wire{Id: 3},
    &Wire{Id: 4},
    &Wire{Id: 7},
    &Wire{Id: 8},
  }) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Ensure that the right variables are on the stack
  for _, variable := range stack[0].Variables {
    if variable.Name == "a" {
      if variable.Value.Id != 3 {
        t.Errorf("Variable a is attached to the wrong wire!")
      }
      continue
    }
    if variable.Name == "b" {
      if variable.Value.Id != 4 {
        t.Errorf("Variable b is attached to the wrong wire!")
      }
      continue
    }
    if variable.Name == "c" {
      if variable.Value.Id != 7 {
        t.Errorf("Variable c is attached to the wrong wire!")
      }
      continue
    }
    if variable.Name == "d" {
      if variable.Value.Id != 8 {
        t.Errorf("Variable d is attached to the wrong wire!")
      }
      continue
    }

    t.Errorf(fmt.Sprintf("Unknown variable %s!", variable.Name))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{}) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }
}

// block foo(a) { return a }
func TestBlockDefinition(t *testing.T) {
  ast := []Node{
    Node{
      Token: "BLOCK",
      Data: map[string]interface{}{
        "Name": "foo",
        "Params": "a",
        "InputQuantity": 1,
        "OutputQuantity": 2,
      },
      Children: &[]Node{
        Node{Token: "BLOCK_RETURN"},
        Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
        Node{Token: "IDENTIFIER", Data: map[string]interface{}{"Value": "a"}},
      },
    },
  }

  stack := []*StackFrame{
    &StackFrame{},
  }

  wireId = 0
  gates, wires, outputs, err := Parse(&ast, stack)

  // Verify error
  if err != nil {
    t.Errorf(fmt.Sprintf("Error returned! %s", err))
    return
  }

  // Verify gates
  if !reflect.DeepEqual(gates, []*Gate{}) {
    // Dereference so we can see the contents of the pointers
    deref := []Gate{}
    for _, gate := range gates {
      deref = append(deref, *gate)
    }
    t.Error(fmt.Sprintf("Gates don't match! %+v", deref))
  }

  // Verify wires
  if !reflect.DeepEqual(wires, []*Wire{}) {
    t.Error(fmt.Sprintf("Wires don't match! %+v", wires))
  }

  // Verify outputs
  if !reflect.DeepEqual(outputs, []*Wire{}) {
    t.Error(fmt.Sprintf("Outputs don't match! %+v", gates))
  }

  // Ensure that the right variables are on the stack
  for _, block := range stack[0].Blocks {
    if block.Name == "foo" {
      if block.Content.Data["Name"].(string) != "foo" {
        t.Errorf("Block foo was not assigned")
      }
      continue
    }
    t.Errorf(fmt.Sprintf("Unknown block %s!", block.Name))
  }
}
