package main

import (
  "fmt"
  "regexp"
  "errors"
  "strings"
)

var NO_DATA func([]string) map[string]interface{} = func(m []string) map[string]interface{} {
  return map[string]interface{}{}
}

type TokenType string
const (
  SINGLE TokenType = "SINGLE"
  WRAPPER_START TokenType = "WRAPPER_START"
  WRAPPER_END TokenType = "WRAPPER_END"
  BINARY_OPERATOR = "BINARY_OPERATOR"
  UNARY_OPERATOR = "UNARY_OPERATOR"
)

type Token struct {
  Name string
  Type TokenType
  Match *regexp.Regexp
  GetData func([]string) map[string]interface{}
  SideEffect func([]string, *StackFrame)
}

var TOKENS []Token = []Token{
  Token{Name: "OP_AND", Type: BINARY_OPERATOR, Match: regexp.MustCompile("^and"), GetData: NO_DATA},
  Token{Name: "OP_OR", Type: BINARY_OPERATOR, Match: regexp.MustCompile("^or"), GetData: NO_DATA},
  Token{Name: "OP_NOT", Type: UNARY_OPERATOR, Match: regexp.MustCompile("^not"), GetData: NO_DATA},

  Token{
    Name: "GROUP",
    Type: WRAPPER_START,
    Match: regexp.MustCompile("^\\("),
    GetData: NO_DATA,

    // Groups with an identifier right before them get converted into invocations.
    SideEffect: func(match []string, stackframe *StackFrame) {
      // Assert that the stackframe isn't nil.
      if stackframe.Nodes == nil { return }

      // Assert that the stackframe has at least two nodes within.
      nodes := *stackframe.Nodes
      if len(nodes) < 2 { return }

      // Make sure the most recent node in the stack frame is a group
      mostRecentNode := &nodes[len(nodes) - 1]
      if mostRecentNode.Token != "GROUP" { return }

      // Check to see if the token before the group was an identifier. If so, then this group isn't
      // a group it's an invocation!
      // identifier ()  =>  identifier()
      //            /\- Group   /\- Invocation!
      secondToMostRecentNode := nodes[len(nodes) - 2]
      if secondToMostRecentNode.Token != "IDENTIFIER" { return }

      // The group is an invocation!
      mostRecentNode.Token = "INVOCATION"
      mostRecentNode.Data["Name"] = secondToMostRecentNode.Data["Value"]
      mostRecentNode.Row = secondToMostRecentNode.Row
      mostRecentNode.Col = secondToMostRecentNode.Col

      // Delete the penultimate node (the IDENTIFIER)
      *stackframe.Nodes = append(nodes[:len(nodes)-2], nodes[len(nodes)-1])
    },
  },
  Token{Name: "GROUP_END", Type: WRAPPER_END, Match: regexp.MustCompile("^\\)"), GetData: NO_DATA},

  Token{
    Name: "BLOCK",
    Type: WRAPPER_START,
    // block identifier(as many identifiers ay needed in here all space seperated) {
    Match: regexp.MustCompile(`(?m)block\s*([A-Za-z_][A-Za-z0-9_]*)\s*\(((([A-Za-z_][A-Za-z0-9_]*)\s*)*([A-Za-z_][A-Za-z0-9_]*)?)\)\s*\{`),
    GetData: func(match []string) map[string]interface{} {
      // Calculate an input quantity
      inputQuantity := 0
      if len(match[2]) > 0 {
        inputQuantity = len(strings.Split(strings.Trim(match[2], " \n\t"), " "))
      }
      return map[string]interface{}{
        "Name": match[1],
        "Params": match[2],
        "InputQuantity": inputQuantity,
        "OutputQuantity": 0, // Will be overridden within `BLOCK_END`
      };
    },
  },
  Token{
    Name: "BLOCK_END",
    Type: WRAPPER_END,
    Match: regexp.MustCompile("^\\}"),
    GetData: NO_DATA,
    SideEffect: func(match []string, stackframe *StackFrame) {
      // Assert that the stackframe isn't nil.
      if stackframe.Nodes == nil { return }

      // Assert that the stackframe has nodes within.
      nodes := *stackframe.Nodes
      if len(nodes) == 0 { return }

      // Get the most recent node in the stack frame.
      mostRecentNode := nodes[len(nodes) - 1]
      if mostRecentNode.Token != "BLOCK" { return }
      blockChildren := *(mostRecentNode.Children)

      // Within the block that's beng closed, was there a return? And, if so, what index token was
      // it?
      returnIndex := -1
      for ct, node := range blockChildren {
        if node.Token == "BLOCK_RETURN" {
          returnIndex = ct
          break;
        }
      }

      if returnIndex == -1 {
        // No return was found, so no tokens are beign returned.
        mostRecentNode.Data["OutputQuantity"] = 0
      } else {
        // Now that the output token location is known, calculate how many tokens were after the
        // return, and that's the number of tokens that are being returned.
        mostRecentNode.Data["OutputQuantity"] = (len(blockChildren) - 1) - returnIndex
      }
    },
  },
  Token{
    Name: "BLOCK_RETURN",
    Type: SINGLE,
    Match: regexp.MustCompile("^return"),
    GetData: NO_DATA,
  },

  Token{
    Name: "ASSIGNMENT",
    Type: SINGLE,
    Match: regexp.MustCompile(`^let +(([A-Za-z_][A-Za-z0-9_]* +)*[A-Za-z_][A-Za-z0-9_]*) ?= ?`),
    GetData: func(match []string) map[string]interface{} {
      return map[string]interface{}{
        "Names": match[1],
      };
    },
  },
  Token{
    Name: "IDENTIFIER",
    Type: SINGLE,
    Match: regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*"),
    GetData: func(match []string) map[string]interface{} {
      return map[string]interface{}{"Value": match[0]};
    },
  },
  Token{
    Name: "BOOL",
    Type: SINGLE,
    Match: regexp.MustCompile("^(1|0)"),
    GetData: func(match []string) map[string]interface{} {
      return map[string]interface{}{
        "Value": match[1] == "1",
      };
    },
  },
}
var RESERVED_WORDS []string = []string{"let", "block", "return"}

type Node struct {
  Token string
  Data map[string]interface{}
  Row int
  Col int
  Children *[]Node
}

type StackFrame struct {
  Type string
  Nodes *[]Node
}

// These checks run on tokens before the side effects have a chance to modify them. A good example
// of why this is helpful is to ensure that identifiers aren't reserved words - since identifiers
// are enveloped inside of other tokens, it's helpful to check the content of the identifiers before
// this happens.
func PreSideEffectValidator(nodes []Node) error {
  for i := 0; i < len(nodes); i++ {

    // Check to make sure identifiers aren't reserved words.
    if nodes[i].Token == "IDENTIFIER" {
      // Ensure that the identifier isn't a reserved word.
      for _, reserved := range RESERVED_WORDS {
        if nodes[i].Data["Value"] == reserved {
          return errors.New(fmt.Sprintf("Identifier %s is a reserved word", reserved))
        }
      }
    }

    // Ensure that the identifier in an assignment isn't a reserved word.
    if nodes[i].Token == "ASSIGNMENT" {
      for _, reserved := range RESERVED_WORDS {
        for _, name := range strings.Split(nodes[i].Data["Names"].(string), " ") {
          if name == reserved {
            return errors.New(fmt.Sprintf("Identifier %s is a reserved word, and cannot be assigned to", reserved))
          }
        }
      }
    }

    // Ensure that the only groups, literals, and identifiers are found after a return.
    if nodes[i].Token == "BLOCK_RETURN" && len(nodes) > i {
      for _, node := range nodes[i+1:] {
        if node.Token != "BOOL" && node.Token != "GROUP" && node.Token != "IDENTIFIER" {
          return errors.New(fmt.Sprintf("Non-expression token %s found after return", node.Token))
        }
      }
    }
  }

  return nil
}

func Validator(nodes []Node) error {
  DUMMY_NODE := Node{Token: "", Data: map[string]interface{}{}, Row: -1, Col: -1}

  for i := 0; i < len(nodes); i++ {
    // Create an array of nodes before (where the index is the number of tokens previous to the
    // current token)
    before := []Node{DUMMY_NODE}
    for j := 1; j <= i; j++ {
      before = append(before, nodes[i-j]) // Reverse the token list
    }
    numberOfDummyTokensToAddToBefore := len(nodes) - len(before)
    for l := 0; l < numberOfDummyTokensToAddToBefore; l++ {
      before = append(before, DUMMY_NODE)
    }

    // Create an array of nodes after (where the index is the number of tokens previous to the
    // current token)
    after := []Node{DUMMY_NODE}
    for j := i+1; j < len(nodes); j++ {
      after = append(after, nodes[j])
    }
    numberOfDummyTokensToAddToAfter := len(nodes) - len(after)
    for l := 0; l < numberOfDummyTokensToAddToAfter; l++ {
      after = append(after, DUMMY_NODE)
    }

    // START ASSERTIONS
    // ----------
    if nodes[i].Token == "OP_AND" {
      if leftHandSide, ok := nodes[i].Data["LeftHandSide"].(Node); ok {
        if !TokenNameIsExpression(leftHandSide.Token) {
          return errors.New("And operator missing a boolean/group on the left hand side")
        }
      } else {
        return errors.New("And operator left hand side is not a node")
      }

      if rightHandSide, ok := nodes[i].Data["RightHandSide"].(Node); ok {
        if !TokenNameIsExpression(rightHandSide.Token) {
          return errors.New("And operator missing a boolean/group on the right hand side")
        }
      } else {
        return errors.New("And operator right hand side is not a node")
      }
    }

    if nodes[i].Token == "OP_OR" {
      if leftHandSide, ok := nodes[i].Data["LeftHandSide"].(Node); ok {
        if !TokenNameIsExpression(leftHandSide.Token) {
          return errors.New("Or operator missing a boolean/group on the left hand side")
        }
      } else {
        return errors.New("Or operator left hand side is not a node")
      }

      if rightHandSide, ok := nodes[i].Data["RightHandSide"].(Node); ok {
        if !TokenNameIsExpression(rightHandSide.Token) {
          return errors.New("Or operator missing a boolean/group on the right hand side")
        }
      } else {
        return errors.New("Or operator right hand side is not a node")
      }
    }

    if nodes[i].Token == "OP_NOT" {
      if rightHandSide, ok := nodes[i].Data["RightHandSide"].(Node); ok {
        if !TokenNameIsExpression(rightHandSide.Token) {
          return errors.New("Not operator missing a boolean/group on the right hand side")
        }
      } else {
        return errors.New("Or operator right hand side is not a node")
      }
    }
  }

  return nil
}

func TokenNameIsExpression(name string) bool {
  return name == "BOOL" || name == "IDENTIFIER" || name == "GROUP"
}

func Tokenizer(input string) (*[]Node, error) {
  code := []byte(input)

  root := &[]Node{}
  children := root
  // A slice of pointers to different locations in the stack
  stacks := []StackFrame{
    StackFrame{
      Type: "ROOT",
      Nodes: root,
    },
  }

  // Initial values for current column and row.
  currentRow := 1
  currentCol := 1

  Outer:
  for len(code) > 0 {
    // Trim whitespace from the start of the code
    codeLength := len(code)
    for i := 0; i < codeLength; i++ {
      if code[0] == ' ' || code[0] == '\t' {
        currentRow += 1
        code = code[1:]
      } else if code[0] == '\n' {
        currentCol += 1
        currentRow = 1
        code = code[1:]
      } else {
        break;
      }
    }

    // Check to make sure that there are characters to match against below in the token matching
    // code. If code = " " going into outer while loop, then the whitespace will be removed but the
    // empty string will still attempt to be matched upon - this will result in failure.
    if len(code) == 0 {
      break
    }

    // Try to find a matching token.
    for _, token := range TOKENS {
      if result := token.Match.FindStringSubmatch(string(code)); result != nil {
        // The token we looped over matched!
        if token.Type == SINGLE || token.Type == UNARY_OPERATOR {
          data := token.GetData(result)

          // Add a right hand side value for every unary operator.
          if token.Type == UNARY_OPERATOR {
            data["RightHandSide"] = nil
          }

          // Single tokens are standalone - append token to the pointer that `children` points to.
          *children = append(*children, Node{
            Token: token.Name,
            Row: currentRow,
            Col: currentCol,
            Data: data,
            Children: nil,
          })
        } else if token.Type == BINARY_OPERATOR {
          // A binary operator takes one argument before it, and one argument after it.

          // Verify there is an expression token before the operator.
          if !(
            len(*children) > 0 &&
            TokenNameIsExpression((*children)[len(*children) - 1].Token)) {
            return nil, errors.New(fmt.Sprintf(
              "Error: Attempted to parse a binary operator (%s), but there wasn't a valid expression before the operator on line %d:%d. Stop.",
              result,
              currentRow,
              currentCol,
            ))
          }

          // Get left hand side
          childrenValue := *children
          leftHandSide := childrenValue[len(childrenValue) - 1]
          *children = childrenValue[:len(childrenValue) - 1]

          data := token.GetData(result)
          data["LeftHandSide"] = leftHandSide
          data["RightHandSide"] = nil

          *children = append(*children, Node{
            Token: token.Name,
            Row: currentRow,
            Col: currentCol,
            Data: data,
            Children: nil,
          })
        } else if token.Type == WRAPPER_START {
          // Create the wrapper start token.
          value := append(*children, Node{
            Token: token.Name,
            Row: currentRow,
            Col: currentCol,
            Data: token.GetData(result),
            Children: &[]Node{},
          })

          // Add the new stack frame to the end of the slice that stores all stack frames.
          stacks = append(stacks, StackFrame{
            Type: token.Name, // Ie, "GROUP" or "BLOCK", etc
            Nodes: &value,
          })

          // Use the children of the just appeneded node as the location to add more tokens into.
          *children = []Node{}
        } else if token.Type == WRAPPER_END {
          // End the wrapper token

          // Ensure that a token of this type makes sense in this context.
          if len(stacks) < 2 {
            return nil, errors.New(fmt.Sprintf(
              "Error: Attempted to close a wrapper that was never opened on %d:%d. Stop.",
              currentRow,
              currentCol,
            ))
          }

          // Ensure that the stack frame we are closing has the same type as the symbol used to
          // close it.
          lastStackFrame := stacks[len(stacks)-1]
          typeShouldBe := regexp.MustCompile(`_END$`).ReplaceAllString(token.Name, "")
          if lastStackFrame.Type != typeShouldBe {
            return nil, errors.New(fmt.Sprintf(
              "Error: Attempted to close wrapper at %d:%d with a %s token, and not a %s_END token. Stop.",
              currentRow,
              currentCol,
              token.Name,
              lastStackFrame.Type,
            ))
          }

          // Assign the `children` pointer back to the stack frame that it belongs to (ie, the last
          // node in the last stack frame)
          lastStackFrameNodes := *lastStackFrame.Nodes
          lastNode := lastStackFrameNodes[len(lastStackFrameNodes) - 1]
          *lastNode.Children = *children

          // Reassign children pointer back to its old value.
          *children = *(stacks[len(stacks) - 1].Nodes)

          // Pop the last stack frome off the end of the stack list now that it has been closed.
          stacks = stacks[:len(stacks) - 1]
        }

        // Run the pre-side-effect validation checks.
        if validator := PreSideEffectValidator(*children); validator != nil {
          return nil, errors.New(fmt.Sprintf(
            "Error: Validation Failed on %d:%d - %s. Stop.",
            currentRow,
            currentCol,
            validator,
          ))
        }

        // Run any custom side effects
        if token.SideEffect != nil {
          token.SideEffect(result, &stacks[len(stacks)-1])
        }

        // If the token we just added to children is an expression, and the previous token is a
        // unary or binary operator, add the token we just added in the right hand side of the
        // previous token.
        if len(*children) >= 2 && TokenNameIsExpression((*children)[len(*children)-1].Token) {
          if _, ok := (*children)[len(*children) - 2].Data["RightHandSide"]; ok {

            // Get right hand side - the last toke in the list
            childrenValue := *children
            rightHandSide := childrenValue[len(*children) - 1]

            // Get the operator - the second to last token in the list
            operator := childrenValue[len(childrenValue) - 2]

            *children = childrenValue[:len(childrenValue) - 2]

            operator.Data["RightHandSide"] = rightHandSide
            *children = append(*children, operator)
          }
        }

        // Add the correct amount of offset to the current row and column to account for this token.
        for i := 0; i < len(result[0]); i++ {
          currentRow += 1
          if result[0][i] == '\n' {
            currentCol += 1
            currentRow = 0
          }
        }

        // Remove the token from the start of the input string we are looping over.
        code = code[len(result[0]):]
        continue Outer;
      }
    }

    // No token was able to match (and break out of the loop above), so throw an error.
    displayCode := code
    if len(displayCode) > 10 {
      displayCode = displayCode[10:]
    }
    return nil, errors.New(fmt.Sprintf(
      "Error: No such token found at %d:%d - `%s`. Stop.",
      currentRow,
      currentCol,
      displayCode,
    ))
    break
  }

  // Ensure that the stack is only 1 item long (the root element) before returning.
  if len(stacks) > 1 {
    return nil, errors.New(fmt.Sprintf(
      "Error: Stack is not empty (%d extra) at end of program (are there more open parenthesis than closing ones?). Stop.",
      len(stacks) - 1,
    ))
  }

  // Also, before returning, validate the final ast.
  if validator := Validator(*children); validator != nil {
    return nil, errors.New(fmt.Sprintf(
      "Error: Validation Failed on %d:%d - %s. Stop.",
      currentRow,
      currentCol,
      validator,
    ))
  }

  return root, nil
}

func PrintAst(tokens *[]Node, indent int, prefix string) {
  if tokens == nil {
    for i := 0; i < indent; i++ { fmt.Printf("  ") }
    if len(prefix) > 0 { fmt.Printf("%s:", prefix); }
    fmt.Printf("<nil>\n")
    return
  }

  for _, token := range *tokens {
    for i := 0; i < indent; i++ { fmt.Printf("  ") }
    if len(prefix) > 0 { fmt.Printf("%s:", prefix); }
    fmt.Printf("%+v\n", token)

    if token.Children != nil {
      PrintAst(token.Children, indent + 1, "")
    }

    if value, ok := token.Data["LeftHandSide"].(Node); ok {
      PrintAst(&[]Node{value}, indent + 1, "LHS")
    }
    if value, ok := token.Data["RightHandSide"].(Node); ok {
      PrintAst(&[]Node{value}, indent + 1, "RHS")
    }
  }
}


func main() {
  result, err := Tokenizer(`
  let a = 1
  a
  `)
  fmt.Println("Error: ", err)
  fmt.Println("Results:")
  PrintAst(result, 0, "")
}
