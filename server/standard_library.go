package main

var STANDARD_LIBRARY map[string]string = map[string]string {
  "counter": `
		block counter8(clock reset) {
			let c1 = tflipflop(clock 1 0 reset)

			let toggle_c2 = c1
			let c2 = tflipflop(clock toggle_c2 0 reset)
			
			let toggle_c4 = (c1 and c2)
			let c4 = tflipflop(clk toggle_c4 0 reset)

			let toggle_c8 = ((c1 and c2) and c4)
			let c8 = tflipflop(clk toggle_c8 0 reset)

			return c1 c2 c4 c8
		}
  `,
  "adder": `
    block halfadder(a b) {
      let sum = ((a and (not b)) or ((not a) and b))
      let carry = (a and b)
      return sum carry
    }
    block adder(a b c) {
      let sum1 carry1 = halfadder(a b)
      let sum2 carry2 = halfadder(sum1 c)

      let carry = (carry1 or carry2)
      return sum2 carry
    }
    block adder4(a0 a1 a2 a3 b0 b1 b2 b3) {
      let sum1 carry1 = adder(a0 b0 0)
      let sum2 carry2 = adder(a1 b1 carry1)
      let sum4 carry4 = adder(a2 b2 carry2)
      let sum8 overflow = adder(a3 b3 carry4)

      return sum1 sum2 sum4 sum8 overflow
    }

    // The twos complement is a helpful value when adding numbers within a computer. It provides a
    // way for a computer to represent a negative value in an addition operation. It's computed by
    // performing ~a + 1.
		block twoscomplement4(a0 a1 a2 a3) {
			let b0 b1 b2 b3 = adder4(
				(not a0) (not a1) (not a2) (not a3) // Take the bitwise not of the input
				0        0        0        1        // Add one to it
			)
			return b0 b1 b2 b3
		}
  `,
  "latch": `
    block srlatch(s r) {
      let q = (not (r or nq))
      let nq = (not (s or nq))
      return q
    }
    block srlatch2(s r) {
      let q = (not (r or nq))
      let nq = (not (s or nq))
      return q nq
    }

    block dlatch(clock d) {
      let s r = (clock and d) (clock and (not d))
      let q = (not (r or nq))
      let nq = (not (s or nq))
      return q
    }
    block dlatch2(clock d) {
      let s r = (clock and d) (clock and (not d))
      let q = (not (r or nq))
      let nq = (not (s or nq))
      return q nq
    }
  `,
}
