# Logic Gates

Lovelace has a number of core logic gates. Each accepts a fixed number of inputs and has a fixed
number of outputs.

<br />

- An **OR** gate accepts two inputs, and turns on it's single output if either input is on. An
  example of its usage could be to allow someone to turn on something from two different switches in
  different locations in a room. In Lovelace, an `OR` gate only accepts two inputs.

**Lovelace syntax**
```
let output = input_1 or input_2
```

**Truth table**

| Input 1 | Input 2 | Output |
|---------|---------|--------|
| `0` | `0` | `0` |
| `0` | `1` | `1` |
| `1` | `0` | `1` |
| `1` | `1` | `1` |

<br />
<br />

- An **AND** gate accepts two inputs, and turns on it's single output when both inputs are on. An
  example of its usage would be in a system that requires two switches to be pressed at the same
  time in order to trigger something. In Lovelace, an `AND` gate only accepts two inputs.

**Lovelace syntax**
```
let output = input_1 and input_2
```

**Truth table**

| Input 1 | Input 2 | Output |
|---------|---------|--------|
| `0` | `0` | `0` |
| `0` | `1` | `0` |
| `1` | `0` | `0` |
| `1` | `1` | `1` |

<br />
<br />

- A **NOT** gate accepts a single input, and turns off it single output when its input is on. It's
  primarily used as an inverter.

**Lovelace syntax**
```
let output = not input
```

**Truth table**

| Input | Output |
|-------|--------|
| `0` | `1` |
| `0` | `0` |

That's all we need at the lowest level to build everything in Lovelace - just three gates.

## Commutative Property

Both the `AND` and `OR` gates follow the commutative property. This means that when used the order
of the parameters that are specified doesn't matter. For example, take the two expressions `a and b`
and `b and a`. They are both equivalent.

## Associative property

In addition, both `AND` and `OR` gates follow the associative property. In a nutshell, this means
that given three values `a`, `b`, and `c`, all the below expressions are equal:

- `((a and b) and c)` - the operation is performed on `a` / `b` first, then on `c`.
- `(a and (b and c))` - the operation is performed on `b` / `c` first, then on `a`.
- `((c and a) and b)` - the operation is performed on `c` / `a` first, then on `b`.

Now, with a basic understanding of some of the fundamental properties of boolean algebra, Let's do
some experiments to learn more about how these gates interact.

## Logic Gate Experiments

For our first experiment, let's try to model a somewhat-more complicated situation. Let's say that
we want to turn on a led when a couple different combinations of three different switches are
enabled. If both switch one and switch three are pressed, or only switch two is pressed, then turn
on the led.

If you read back through that scenario, you might be able to guess the logic gate diagram that would
be required to solve that problem. But, there's a way to solve an arbitrarily complex scenario in
this format.

One of the helpful properties of an `AND` gate is that it allows someone to verify that a number of
conditions are all true. For example, the expression `(((a and b) and c) and d)` will only be true
is all the sub-expressions `a`, `b`, `c`, and `d` are all true. When used in this way, think of a
bunch of `AND` gates in a row as a way of checking to make sure a number of conditions are all true.

In addition, another helpful gate we're going to use in this scenario is the `NOT` gate. As
mentioned above, the not gate can be used as an inverter. In this way, the expression `not toggle()`
would evaluate to true when the toggle switch is off and would evaluate to false when the toggle
switch is on.

In our example above, we have two conditions to check:
1. Switch a and Switch c are both on
2. Only Switch b is on

With the above knowledge, we can create two different expressions to match the two different cases
that are required in order to complete our project:
1. Switch a and Switch c are both on - `((a and (not b)) and c)`
2. Only Switch b is on - `(((not a) and b) and (not c))`

Initially, you may think that the second case could just be `not b`. Unfortunately, that
doesn't take into account the state of either switch `a` or switch `c`. If you only used `not b`,
then case number two would still evaluate to true if both switch `a` and switch `b` were on.

To finish up the exercise, we want to turn on the led if either condition is true. Luckily, there's
a logic gate that does just this - the `OR` gate. So, our final expression looks a little something
like this:

```
(((a and (not b)) and c) or (((not a) and b) and (not c)))
```

Finally, let's convert that into a Lovelace program and run it in the [Lovelace
preview](https://lovelace-preview.surge.sh):
```
let a = toggle()
let b = toggle()
let c = toggle()
let result = (((a and (not b)) and c) or (((not a) and b) and (not c)))
led(result)
```


## Takeaways
Now that you understand the principal behind it, here is an easy rule to create a boolean expression
that evaluates to true if a bunch of other sets of boolean expressions also evaluate to true:

1. Write down each condition that that should equate to the output being true. One condition might
   be that `x` must be on, `y` must be off, `z` must also be off, and another condition might be
   that `x` must be off, `y` must be on, and `z` must be `off`.

```
x=true  y=false z=false
x=false y=true  z=false
```

2. For every expression in each condition, write down its variable. If the expression should be
   equal to `false`, add a `NOT` in front of it.

```
x       not y   not z
not x   y       not z
```

3. Add an `AND` between each expression within each condition.

```
x         and   (not y)   and   (not z)
(not x)   and   y         and   (not z)
((  x     and (not y)) and (not z))
(((not x) and   y    ) and (not z))
```

4. Add an `OR` between each condition.

```
(((x and (not y)) and (not z)) or (((not x) and y) and (not z)))
```
