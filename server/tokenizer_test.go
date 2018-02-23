package main

import (
  "testing"
  "fmt"
  "reflect"
)

var NONE map[string]interface{} = map[string]interface{}{}

func TestAnd(t *testing.T) {
  result, err := Tokenizer("1 and 0")
  if err != nil { t.Error(fmt.Sprintf("Error: %s", err.Error())) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_AND", Row: 3, Col: 1, Data: map[string]interface{}{
      "LeftHandSide": Node{Token: "BOOL", Row: 1, Col: 1, Data: map[string]interface{}{"Value": true}},
      "RightHandSide": Node{Token: "BOOL", Row: 7, Col: 1, Data: map[string]interface{}{"Value": false}},
    }},
  }) {
    t.Error("Fail!")
  }
}

func TestOr(t *testing.T) {
  result, err := Tokenizer("0 or 1")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_OR", Row: 3, Col: 1, Data: map[string]interface{}{
      "LeftHandSide": Node{Token: "BOOL", Row: 1, Col: 1, Data: map[string]interface{}{"Value": false}},
      "RightHandSide": Node{Token: "BOOL", Row: 6, Col: 1, Data: map[string]interface{}{"Value": true}},
    }},
  }) {
    t.Error("Fail!")
  }
}

func TestNot(t *testing.T) {
  result, err := Tokenizer("not 1")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_NOT", Row: 1, Col: 1, Data: map[string]interface{}{
      "RightHandSide": Node{Token: "BOOL", Row: 5, Col: 1, Data: map[string]interface{}{"Value": true}},
    }},
  }) {
    t.Error("Fail!")
  }
}

func TestWhitespaceAtEnd(t *testing.T) {
  result, err := Tokenizer(`1 and 0
  `)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_AND", Row: 3, Col: 1, Data: map[string]interface{}{
      "LeftHandSide": Node{Token: "BOOL", Row: 1, Col: 1, Data: map[string]interface{}{"Value": true}},
      "RightHandSide": Node{Token: "BOOL", Row: 7, Col: 1, Data: map[string]interface{}{"Value": false}},
    }},
  }) {
    t.Error("Fail!")
  }
}

// 1 and and => error
func TestAndValidatorWithBooleans(t *testing.T) {
  _, err := Tokenizer("1 and and")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Attempted to parse a binary operator ([and]), but there wasn't a valid expression before the operator on line 7:1. Stop.") { t.Error("Error: "+err.Error()) }
}

// 1 or or => error
func TestOrValidatorWithBooleans(t *testing.T) {
  _, err := Tokenizer("1 or or")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Attempted to parse a binary operator ([or]), but there wasn't a valid expression before the operator on line 6:1. Stop.") { t.Error("Error: "+err.Error()) }
}

// 1 or => error
func TestOrValidatorWithNoRightHandSide(t *testing.T) {
  _, err := Tokenizer("1 or")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 5:1 - Or operator right hand side is not a node. Stop.") { t.Error("Error: "+err.Error()) }
}

// 1 and => error
func TestAndValidatorWithNoRightHandSide(t *testing.T) {
  _, err := Tokenizer("1 and")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 6:1 - And operator right hand side is not a node. Stop.") { t.Error("Error: "+err.Error()) }
}

// (1 or 0) and (0 or 1)
func TestGroups(t *testing.T) {
  result, err := Tokenizer("(1 or 0) and (0 or 1)")
  if err != nil { t.Error("Error: "+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_AND", Row: 10, Col: 1, Data: map[string]interface{}{
      "RightHandSide": Node{Token: "GROUP", Row: 14, Col: 1, Data: NONE, Children: &[]Node{
        Node{Token: "OP_OR", Row: 17, Col: 1, Data: map[string]interface{}{
          "LeftHandSide": Node{Token: "BOOL", Row: 15, Col: 1, Data: map[string]interface{}{"Value": false}},
          "RightHandSide": Node{Token: "BOOL", Row: 20, Col: 1, Data: map[string]interface{}{"Value": true}},
        }},
      }},
      "LeftHandSide": Node{Token: "GROUP", Row: 1, Col: 1, Data: NONE, Children: &[]Node{
        Node{Token: "OP_OR", Row: 4, Col: 1, Data: map[string]interface{}{
          "LeftHandSide": Node{Token: "BOOL", Row: 2, Col: 1, Data: map[string]interface{}{"Value": true}},
          "RightHandSide": Node{Token: "BOOL", Row: 7, Col: 1, Data: map[string]interface{}{"Value": false}},
        }},
      }},
    }},
  }) {
    t.Error("Fail!")
  }
}

// (1 or ((0 or 0) and 1)) and (0 or (1 and 0))
func TestNestedGroups(t *testing.T) {
  result, err := Tokenizer("(1 or ((0 or 0) and 1)) and (0 or (1 and 0))")
  if err != nil { t.Error("Error: "+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_AND", Row: 25, Col: 1, Data: map[string]interface{}{
      "LeftHandSide": Node{Token: "GROUP", Row: 1, Col: 1, Data: NONE, Children: &[]Node{
        Node{Token: "OP_OR", Row: 4, Col: 1, Data: map[string]interface{}{
          "LeftHandSide": Node{Token: "BOOL", Row: 2, Col: 1, Data: map[string]interface{}{"Value": true}},
          "RightHandSide": Node{Token: "GROUP", Row: 7, Col: 1, Data: NONE, Children: &[]Node{
            Node{Token: "OP_AND", Row: 17, Col: 1, Data: map[string]interface{}{
              "LeftHandSide": Node{Token: "GROUP", Row: 8, Col: 1, Data: NONE, Children: &[]Node{
                Node{Token: "OP_OR", Row: 11, Col: 1, Data: map[string]interface{}{
                  "LeftHandSide": Node{Token: "BOOL", Row: 9, Col: 1, Data: map[string]interface{}{"Value": false}},
                  "RightHandSide": Node{Token: "BOOL", Row: 14, Col: 1, Data: map[string]interface{}{"Value": false}},
                }},
              }},
              "RightHandSide": Node{Token: "BOOL", Row: 21, Col: 1, Data: map[string]interface{}{"Value": true}},
            }},
          }},
        }},
      }},
      "RightHandSide": Node{Token: "GROUP", Row: 29, Col: 1, Data: NONE, Children: &[]Node{
        Node{Token: "OP_OR", Row: 32, Col: 1, Data: map[string]interface{}{
          "LeftHandSide": Node{Token: "BOOL", Row: 30, Col: 1, Data: map[string]interface{}{"Value": false}},
          "RightHandSide": Node{Token: "GROUP", Row: 35, Col: 1, Data: NONE, Children: &[]Node{
            Node{Token: "OP_AND", Row: 38, Col: 1, Data: map[string]interface{}{
              "LeftHandSide": Node{Token: "BOOL", Row: 36, Col: 1, Data: map[string]interface{}{"Value": true}},
              "RightHandSide": Node{Token: "BOOL", Row: 42, Col: 1, Data: map[string]interface{}{"Value": false}},
            }},
          }},
        }},
      }},
    }},
  }) {
    t.Error("Fail!")
  }
}

// (1 or 0)) and (0 or 1) => error
func TestGroupsWithTooManyClosingParens(t *testing.T) {
  _, err := Tokenizer("(1 or 0)) and (0 or 1)")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Attempted to close a wrapper that was never opened on 9:1. Stop.") { t.Error("Error: "+err.Error()) }
}

// (1 or 0) and (0
func TestGroupsWithTooManyOpenParens(t *testing.T) {
  _, err := Tokenizer("(1 or 0) and (0")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Stack is not empty (1 extra) at end of program (are there more open parenthesis than closing ones?). Stop.") { t.Error("Error: "+err.Error()) }
}

// IDENTIFIERS

// a or b
func TestIdentifiersOr(t *testing.T) {
  result, err := Tokenizer("a or b")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "OP_OR", Row: 3, Col: 1, Data: map[string]interface{}{
      "LeftHandSide": Node{Token: "IDENTIFIER", Row: 1, Col: 1, Data: map[string]interface{}{"Value": "a"}},
      "RightHandSide": Node{Token: "IDENTIFIER", Row: 6, Col: 1, Data: map[string]interface{}{"Value": "b"}},
    }},
  }) {
    t.Error("Fail!")
  }
}

// a or let
func TestIdentifiersCannotBeAReservedWord(t *testing.T) {
  _, err := Tokenizer("a or let")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 6:1 - Identifier let is a reserved word. Stop.") { t.Error("Error: "+err.Error()) }
}

// ASSIGNMENT

// let a = 1
func TestAssignment(t *testing.T) {
  result, err := Tokenizer("let a = 1")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "ASSIGNMENT", Row: 1, Col: 1, Data: map[string]interface{}{"Names": "a", "Values": []Node{}}},
    Node{Token: "BOOL", Row: 9, Col: 1, Data: map[string]interface{}{"Value": true}},
  }) {
    t.Error("Fail!")
  }
}

// let a b = adder(1 0)
func TestAssignmentWithMultipleValues(t *testing.T) {
  result, err := Tokenizer("let a b = adder(1 0)")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "ASSIGNMENT", Row: 1, Col: 1, Data: map[string]interface{}{"Names": "a b", "Values": []Node{}}},
    Node{
      Token: "INVOCATION",
      Row: 11,
      Col: 1,
      Data: map[string]interface{}{"Name": "adder"},
      Children: &[]Node{
        Node{Token: "BOOL", Row: 17, Col: 1, Data: map[string]interface{}{"Value": true}},
        Node{Token: "BOOL", Row: 19, Col: 1, Data: map[string]interface{}{"Value": false}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

// let return = 1
func TestAssignmentIdentifiersCannotBeAReservedWord(t *testing.T) {
  _, err := Tokenizer("let return = 1")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 1:1 - Identifier return is a reserved word, and cannot be assigned to. Stop.") { t.Error("Error: "+err.Error()) }
}

// BLOCKS
func TestBlock(t *testing.T) {
  result, err := Tokenizer(`block a(b c d) {
    let a = 1
  }`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "BLOCK",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "a", "Params": "b c d", "OutputQuantity": 0, "InputQuantity": 3},
      Children: &[]Node{
        Node{Token: "ASSIGNMENT", Row: 5, Col: 2, Data: map[string]interface{}{"Names": "a", "Values": []Node{}}},
        Node{Token: "BOOL", Row: 13, Col: 2, Data: map[string]interface{}{"Value": true}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

func TestBlockWithoutInputParameters(t *testing.T) {
  result, err := Tokenizer(`block a() {
    let a = 1
  }`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "BLOCK",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "a", "Params": "", "OutputQuantity": 0, "InputQuantity": 0},
      Children: &[]Node{
        Node{Token: "ASSIGNMENT", Row: 5, Col: 2, Data: map[string]interface{}{"Names": "a", "Values": []Node{}}},
        Node{Token: "BOOL", Row: 13, Col: 2, Data: map[string]interface{}{"Value": true}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

func TestBlockWithReturn(t *testing.T) {
  result, err := Tokenizer(`block a(b c d) {
    let e = (b and c)
    return (e and (c or d))
  }`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "BLOCK",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "a", "Params": "b c d", "OutputQuantity": 1, "InputQuantity": 3},
      Children: &[]Node{
        Node{Token: "ASSIGNMENT", Row: 5, Col: 2, Data: map[string]interface{}{"Names": "e", "Values": []Node{}}},
        Node{Token: "GROUP", Row: 13, Col: 2, Data: NONE, Children: &[]Node{
          Node{Token: "OP_AND", Row: 16, Col: 2, Data: map[string]interface{}{
            "LeftHandSide": Node{Token: "IDENTIFIER", Row: 14, Col: 2, Data: map[string]interface{}{"Value": "b"}},
            "RightHandSide": Node{Token: "IDENTIFIER", Row: 20, Col: 2, Data: map[string]interface{}{"Value": "c"}},
          }},
        }},
        Node{Token: "BLOCK_RETURN", Row: 5, Col: 3, Data: NONE},
        Node{Token: "GROUP", Row: 12, Col: 3, Data: NONE, Children: &[]Node{
          Node{Token: "OP_AND", Row: 15, Col: 3, Data: map[string]interface{}{
            "LeftHandSide": Node{Token: "IDENTIFIER", Row: 13, Col: 3, Data: map[string]interface{}{"Value": "e"}},
            "RightHandSide": Node{Token: "GROUP", Row: 19, Col: 3, Data: NONE, Children: &[]Node{
              Node{Token: "OP_OR", Row: 22, Col: 3, Data: map[string]interface{}{
                "LeftHandSide": Node{Token: "IDENTIFIER", Row: 20, Col: 3, Data: map[string]interface{}{"Value": "c"}},
                "RightHandSide": Node{Token: "IDENTIFIER", Row: 25, Col: 3, Data: map[string]interface{}{"Value": "d"}},
              }},
            }},
          }},
        }},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

func TestBlockReturningMultipleValues(t *testing.T) {
  result, err := Tokenizer(`block a(b c d) {
    return
      1
      a
  }`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "BLOCK",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "a", "Params": "b c d", "OutputQuantity": 2, "InputQuantity": 3},
      Children: &[]Node{
        Node{Token: "BLOCK_RETURN", Row: 5, Col: 2, Data: NONE},
        Node{Token: "BOOL", Row: 7, Col: 3, Data: map[string]interface{}{"Value": true}},
        Node{Token: "IDENTIFIER", Row: 7, Col: 4, Data: map[string]interface{}{"Value": "a"}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

func TestBlockReturnCannotHaveNonExpressionTokens(t *testing.T) {
  _, err := Tokenizer(`block a(b c d) {
    return
    let a = 1
  }`)
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 5:3 - Non-expression token ASSIGNMENT found after return. Stop.") { t.Error("Error: "+err.Error()) }
}

func TestBlockWithParameterExpansion(t *testing.T) {
  result, err := Tokenizer(`block a(b[2]) {
    let a = 1
  }`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "BLOCK",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "a", "Params": "b0 b1", "OutputQuantity": 0, "InputQuantity": 2},
      Children: &[]Node{
        Node{Token: "ASSIGNMENT", Row: 5, Col: 2, Data: map[string]interface{}{"Names": "a", "Values": []Node{}}},
        Node{Token: "BOOL", Row: 13, Col: 2, Data: map[string]interface{}{"Value": true}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

// Invocations
func TestInvocation(t *testing.T) {
  result, err := Tokenizer(`foo(a b 1)`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "INVOCATION",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{Token: "IDENTIFIER", Row: 5, Col: 1, Data: map[string]interface{}{"Value": "a"}},
        Node{Token: "IDENTIFIER", Row: 7, Col: 1, Data: map[string]interface{}{"Value": "b"}},
        Node{Token: "BOOL", Row: 9, Col: 1, Data: map[string]interface{}{"Value": true}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}
func TestInvocationNoParams(t *testing.T) {
  result, err := Tokenizer(`foo()`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "INVOCATION",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{},
    },
  }) {
    t.Error("Fail!")
  }
}
func TestInvocationWithinInvocation(t *testing.T) {
  result, err := Tokenizer(`foo(bar(a 1))`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "INVOCATION",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{
          Token: "INVOCATION",
          Row: 5,
          Col: 1,
          Data: map[string]interface{}{"Name": "bar"},
          Children: &[]Node{
            Node{Token: "IDENTIFIER", Row: 9, Col: 1, Data: map[string]interface{}{"Value": "a"}},
            Node{Token: "BOOL", Row: 11, Col: 1, Data: map[string]interface{}{"Value": true}},
          },
        },
      },
    },
  }) {
    t.Error("Fail!")
  }
}
func TestInvocationWithinInvocationWithArgsAfterward(t *testing.T) {
  result, err := Tokenizer(`foo(bar(a 1) 0)`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "INVOCATION",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Name": "foo"},
      Children: &[]Node{
        Node{
          Token: "INVOCATION",
          Row: 5,
          Col: 1,
          Data: map[string]interface{}{"Name": "bar"},
          Children: &[]Node{
            Node{Token: "IDENTIFIER", Row: 9, Col: 1, Data: map[string]interface{}{"Value": "a"}},
            Node{Token: "BOOL", Row: 11, Col: 1, Data: map[string]interface{}{"Value": true}},
          },
        },
        Node{Token: "BOOL", Row: 14, Col: 1, Data: map[string]interface{}{"Value": false}},
      },
    },
  }) {
    t.Error("Fail!")
  }
}

// Comments
func TestSingleLineComment(t *testing.T) {
  result, err := Tokenizer(`// I am a comment`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "SINGLE_COMMENT",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Message": "I am a comment"},
    },
  }) {
    t.Error("Fail!")
  }
}

func TestMultiLineComment(t *testing.T) {
  result, err := Tokenizer(`/* I am a multiline
comment */`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "MULTI_COMMENT",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Message": "I am a multiline\ncomment"},
    },
  }) {
    t.Error("Fail!")
  }
}

func TestMultiLineCommentSingleLine(t *testing.T) {
  result, err := Tokenizer(`/* I am a multiline comment but only on one line*/`)
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{
      Token: "MULTI_COMMENT",
      Row: 1,
      Col: 1,
      Data: map[string]interface{}{"Message": "I am a multiline comment but only on one line"},
    },
  }) {
    t.Error("Fail!")
  }
}
