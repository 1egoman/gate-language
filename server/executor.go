package main

import (
  "fmt"
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
  var oldHash string = ""
  var newHash string

  // Define a max iteration count. This provides a ceiling that can be used to halt infinite
  // recursions.
  reasonableMaxIterationCount := (len(wires) * 5) + (len(gates) * 5)

  for iterationCount := 0; iterationCount < reasonableMaxIterationCount; iterationCount++ {
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

    // Update the gates in reverse order.
    // This is to curcumvent a bug where if two flip flops are attached to each other (where the
    // output of the first flip flop is the input into the second), then the rising edge of the
    // clock on the second t-flip-flop will "render" after the the first's rising edge, causing both
    // flip flops to toggle. This means that any contraption created with chained flip-flops won't
    // work properly. So if you want to change the loop order below, make sure that chained flip
    // flops aren't effected!
    for i := len(gates)-1; i >= 0; i-- {
      gate := gates[i]

      switch gate.Type {
      case "AND":
        setWire(wires, gate.Outputs[0].Id, getWire(wires, gate.Inputs[0].Id) && getWire(wires, gate.Inputs[1].Id));
      case "OR":
        setWire(wires, gate.Outputs[0].Id, getWire(wires, gate.Inputs[0].Id) || getWire(wires, gate.Inputs[1].Id));
      case "NOT":
        setWire(wires, gate.Outputs[0].Id, !getWire(wires, gate.Inputs[0].Id));
      case "BLOCK_INPUT": fallthrough
      case "BLOCK_OUTPUT":
        setWire(wires, gate.Outputs[0].Id, getWire(wires, gate.Inputs[0].Id));
      case "SOURCE":
        setWire(wires, gate.Outputs[0].Id, true);
      case "GROUND":
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
          // Ensure that the tflipflop has enough inputs. This is also enforced at compile-time.
          if len(gate.Inputs) < 2 {
            break
          }

          // Set a default state for the flipflop if it hasn't been set already.
          if len(gate.State) == 0 {
            gate.State = "10"
          }

          // Was set wire pulled high?
          if len(gate.Inputs) > 2 && getWire(wires, gate.Inputs[2].Id) {
            gate.State = fmt.Sprintf("%s1", string(gate.State[0]))
            continue
          }
          // Was reset wire pulled high?
          if len(gate.Inputs) > 3 && getWire(wires, gate.Inputs[3].Id) {
            gate.State = fmt.Sprintf("%s0", string(gate.State[0]))
            continue
          }

          // Neither was pulled high, so see if the main wire was and the flip flop should be
          // flipped.
          clock := getWire(wires, gate.Inputs[0].Id);
          powered := getWire(wires, gate.Inputs[1].Id);

          // Format for gate.State:
          // bit at index 0: used for storing if in the last frame, the tflipflop was powered
          // bit at index 1: used for storing the state of the flip flop
          // (1 if the S side is active, 0 if the R side is active)

          // Detect the rising edge of the clock
          if clock && gate.State[0] == '0' {
            newState := string(gate.State[1])

            // If powered on the clock's rising edge, then flip the state
            if powered && gate.State[1] == '1' {
              newState = "0"
            } else if powered {
              newState = "1"
            }

            gate.State = fmt.Sprintf("1%s", newState)

          // Detect the falling edge of the clock
          } else if !clock && gate.State[0] == '1' {
            gate.State = fmt.Sprintf("0%s", string(gate.State[1]))
          }

          if (gate.State[1] == '1') {
            /* The S side of the latch is active */
            setWire(wires, gate.Outputs[0].Id, true);
            if len(gate.Outputs) > 1 { /* set not q if passed */
              setWire(wires, gate.Outputs[1].Id, false);
            }
          } else {
            /* The R side of the latch is active */
            setWire(wires, gate.Outputs[0].Id, false);
            if len(gate.Outputs) > 1 { /* set not q if passed */
              setWire(wires, gate.Outputs[1].Id, true);
            }
          }
        }
      }
    }
  }

  return gates, wires
}
