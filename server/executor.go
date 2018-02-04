package main

import (
  "fmt"
  // "math/rand"
)

func setWire(wires []*Wire, id int, powered bool) {
  for _, wire := range wires {
    if wire.Id == id {
      wire.Powered = powered
      break
    }
  }
}

func getWire(wires []*Wire, id int) bool {
  for _, wire := range wires {
    if wire.Id == id {
      return wire.Powered
    }
  }
  return false
}

func calculateGateHash(gates []*Gate) string {
  hash := ""

  // Calculate which gates in the slice are important, and add them to the hash
  for _, gate := range gates {
    if gate.Type == "BUILTIN_FUNCTION" && (gate.Label == "toggle" || gate.Label == "momentary") {
      // Add the gate's state to the hash ( id,state; )
      hash += fmt.Sprintf("%d,%s;", gate.Id, gate.State)
    }
  }

  // Finally, add the gate slice length to the end of the hash
  hash += fmt.Sprintf("%d", len(gates))
  return hash
}

func calculateWireHash(wires []*Wire) string {
  hash := ""

  // Calculate which gates in the slice are important, and add them to the hash
  for _, wire := range wires {
    // Add the wire's state to the hash ( id,state; )
    hash += fmt.Sprintf("%d,%v;", wire.Id, wire.Powered)
  }
  return hash
}

func Execute(gates []*Gate, wires []*Wire) ([]*Gate, []*Wire) {
  // Loop for a number of times. There's some randomness added here to try to make
  // debugging of infinite looping constructs easier.
  // loopCount := 150 + rand.Intn(5)
  // for i := 0; i < loopCount; i++ {

  var oldHash string = ""
  var newHash string
  for {
    // Calculate a hash of the state of all the wires.
    newHash = calculateWireHash(wires)

    // If the hash after the last calculation is the same as the hash before
    // the last calculation, then break out of the loop. We're at a stable state.
    if oldHash == newHash {
      break
    }

    // Another round of computation is required. The hash of the current state is the hash to
    // compare against for the next check.
    oldHash = newHash

    for _, gate := range gates {
      switch gate.Type {
      case "AND":
        fmt.Println("DID AND")
        setWire(wires, gate.Outputs[0].Id, getWire(wires, gate.Inputs[0].Id) && getWire(wires, gate.Inputs[1].Id));
      case "OR":
        fmt.Println("DID OR")
        setWire(wires, gate.Outputs[0].Id, getWire(wires, gate.Inputs[0].Id) || getWire(wires, gate.Inputs[1].Id));
      case "NOT":
        fmt.Println("DID NOT")
        setWire(wires, gate.Outputs[0].Id, !getWire(wires, gate.Inputs[0].Id));
      case "BLOCK_INPUT":
      case "BLOCK_OUTPUT":
        fmt.Println("DID BLOCK_*")
        setWire(wires, gate.Outputs[0].Id, getWire(wires, gate.Inputs[0].Id));
      case "SOURCE":
        fmt.Println("DID SOURCE")
        setWire(wires, gate.Outputs[0].Id, true);
      case "GROUND":
        fmt.Println("DID GROUND")
        setWire(wires, gate.Outputs[0].Id, false);

      case "BUILTIN_FUNCTION":
        if (gate.Label == "momentary" || gate.Label == "toggle") {
          for i := 0; i < len(gate.Outputs); i++ {
            setWire(wires, gate.Outputs[i].Id, gate.State == "on");
          }
        } else if (gate.Label == "led") {
          if getWire(wires, gate.Inputs[0].Id) {
            gate.State = "on"
          } else {
            gate.State = "off"
          }
        } else if (gate.Label == "tflipflop") {
          powered := getWire(wires, gate.Inputs[0].Id);

          // Set a default state for the flipflop if it hasn't been set already.
          if len(gate.State) == 0 {
            gate.State = "100"
          }

          // Extract the parts of the state.
          r := gate.State[0] == '1'
          s := gate.State[1] == '1'
          hasBeenFlipped := gate.State[2] == '1'

          // If power was received and the state wasn't already flipped, do this now.
          if (powered && !hasBeenFlipped) {
            hasBeenFlipped = true
            r, s = s, r // Flip the state of the gate
          } else if (!hasBeenFlipped) {
            hasBeenFlipped = false
          }

          if (r) {
            /* The R side of the latch is active */
            setWire(wires, gate.Outputs[0].Id, true);
            if len(gate.Outputs) > 1 { /* set not q if passed */
              setWire(wires, gate.Outputs[1].Id, false);
            }
          } else {
            /* The S side of the latch is active */
            setWire(wires, gate.Outputs[0].Id, false);
            if len(gate.Outputs) > 1 { /* set not q if passed */
              setWire(wires, gate.Outputs[1].Id, true);
            }
          }

          // Update the state with the updated values
          gate.State = fmt.Sprintf("%d%d%d", r, s, hasBeenFlipped)
        }
      }
    }
  }

  return gates, wires
}
