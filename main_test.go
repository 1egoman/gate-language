package main

import (
  "testing"
  // "fmt"
  "reflect"
)

var NONE map[string]interface{} = map[string]interface{}{}

func TestAnd(t *testing.T) {
  result, err := Tokenizer("1 and 0")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "BOOL", Row: 1, Col: 1, Data: map[string]interface{}{"Value": true}},
    Node{Token: "OP_AND", Row: 3, Col: 1, Data: NONE},
    Node{Token: "BOOL", Row: 7, Col: 1, Data: map[string]interface{}{"Value": false}},
  }) {
    t.Error("Fail!")
  }
}

func TestOr(t *testing.T) {
  result, err := Tokenizer("0 or 1")
  if err != nil { t.Error("Error:"+err.Error()) }
  if !reflect.DeepEqual(*result, []Node{
    Node{Token: "BOOL", Row: 1, Col: 1, Data: map[string]interface{}{"Value": false}},
    Node{Token: "OP_OR", Row: 3, Col: 1, Data: NONE},
    Node{Token: "BOOL", Row: 6, Col: 1, Data: map[string]interface{}{"Value": true}},
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
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 7:1 - And operator missing a boolean on the right hand side. Stop.") { t.Error("Error:"+err.Error()) }
}

// 1 or or => error
func TestOrValidatorWithBooleans(t *testing.T) {
  _, err := Tokenizer("1 or or")
  if err == nil {
    t.Error("err was nil!")
    return
  }
  if !reflect.DeepEqual(err.Error(), "Error: Validation Failed on 6:1 - Or operator missing a boolean on the right hand side. Stop.") { t.Error("Error:"+err.Error()) }
}
