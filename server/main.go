package main

import (
  "os"
  "fmt"
  "encoding/json"
  "flag"

  // For reading file from disk
  "io/ioutil"
)

var isRunningInServer bool = false

func help(subcomponent string) {
  dollar0 := os.Args[0]
  switch subcomponent {
  case "build":
    fmt.Printf("Usage: %s build <file.bit> [--verbose]", dollar0)
    fmt.Println()
    fmt.Println("Compiles lovelace source into a list of gates and wires that can be executed.")
    fmt.Println()
    fmt.Println("Flags:")
    fmt.Println("   --verbose\t\tPrint debugging information")
    fmt.Println("   --max-call-depth\tChange the max block invocation depth. Setting to 0 disables the limit. Defaults to 100.")

  case "tokenize":
    fmt.Printf("Usage: %s tokenize <file.bit>", dollar0)
    fmt.Println()
    fmt.Println("Tokenizes lovelace source into an array of tokens. This is mostly useful for debugging lovelace itself when it won't parse a known-good file.")

  case "serve":
    fmt.Printf("Usage: %s serve [--port 8080] [--verbose]", dollar0)
    fmt.Println()
    fmt.Println("Runs a http server that can be used to remotely compile and run lovelace ast. The server exposes two http endpoints:")
    fmt.Println(" POST /v1/compile, which compiles any lovelace source included in the request into ast.")
    fmt.Println(" POST /v1/run, which executes any ast, returning the state of all wires.")
    fmt.Println()
    fmt.Println("Usage Examples:")
    fmt.Println("The below request compiles the program led(toggle()) into two gates (toggle switch and led) and one wire connecting them:")
    fmt.Println()
    fmt.Println("$ curl http://localhost:8080/v1/compile -H 'Accept: application/json' -d 'led(toggle())'")
    fmt.Println(`{"Gates":[{"Id":1,"Type":"BUILTIN_FUNCTION","Label":"toggle","Inputs":[],"Outputs":[{"Id":1,"Desc":"","Start":null,"End":null,"Powered":false}],"CallingContext":0,"State":""},{"Id":2,"Type":"BUILTIN_FUNCTION","Label":"led","Inputs":[{"Id":1,"Desc":"","Start":null,"End":null,"Powered":false}],"Outputs":[],"CallingContext":0,"State":""}],"Wires":[{"Id":1,"Desc":"","Start":null,"End":null,"Powered":false}],"Contexts":null,"Outputs":[]}`)
    fmt.Println()
    fmt.Println("Flags:")
    fmt.Println("    --port    Specify an alternative port to run on. Defaults to 8080.")
    fmt.Println("   --verbose  Print debugging information")

  default:
    fmt.Printf("Usage: %s <command> [<args>]\n", dollar0)
    fmt.Println("Commonly used subcommands:")
    fmt.Println(" - run        Execute a lovelace program interactively in a live-preview window")
    fmt.Println(" - build      Compile lovelace syntax into an ast that can be run")
    fmt.Println(" - serve      Run a lovelace server that can compile and run ast")
    fmt.Println()
    fmt.Println("Less-commonly used subcommands:")
    fmt.Println(" - tokenize   Compile lovelace syntax into a list of tokens. ")
  }
}

func main() {
  // No subcommand printed? Print help.
  if len(os.Args) == 1 {
    help("")
    return
  }

  // Parse the flags for the subcommand that is active.
  switch os.Args[1] {

  // lovel run foo.bit
  case "run": Run()


  // lovel tokenize foo.bit
  case "tokenize":
    if len(os.Args) < 3 {
      fmt.Println("No file was passed to tokenize. Stop.")
      os.Exit(2)
      return
    }

    // Read source code from disk
    source, err := ioutil.ReadFile(os.Args[2])
    if err != nil {
      fmt.Printf("Error reading file %s: %s. Stop.\n", os.Args[2], err);
      os.Exit(2)
    }

    // Tokenize the source code
    result, err := Tokenizer(string(source))
    if err != nil {
      fmt.Printf("Error tokenizing file %s: %s\n", os.Args[2], err);
      os.Exit(2)
    }

    PrintAst(result, 0, "")

  // lovel build foo.bit
  case "build":
    // Add flags
    buildFlags := flag.NewFlagSet("build", flag.ExitOnError)
    buildVerbose := buildFlags.Bool("verbose", false, "Print debug information")
    buildMaxCallDepth := buildFlags.Int("max-call-depth", -1, "Set the maximum call depth")
    buildFlags.Usage = func() { help("build") }
    buildFlags.Parse(os.Args[2:])

    wireId = 0
    gateId = 0
    stackFrameId = 0

    // Set max call depth if a value was specified.
    if *buildMaxCallDepth != -1 {
      INVOCATION_MAX_RECURSION_DEPTH = *buildMaxCallDepth
    }

    fmt.Println(buildFlags.NArg())
    if buildFlags.NArg() != 1 {
      fmt.Println("No file path was passed to build. Stop.")
      os.Exit(2)
      return
    }


    // Read source code from disk
    summary, err := RunFile(buildFlags.Args()[0], *buildVerbose)
    if err != nil {
      fmt.Println(err);
      os.Exit(2)
      return
    }

    serialized, err2 := json.Marshal(summary)
    if err2 != nil {
      fmt.Println("Error serializing result: %s. Stop.", err2);
      os.Exit(2)
      return
    }
    fmt.Println(string(serialized))

  // lovel serve --port 2185
  case "serve": Serve()

  // Print out help info
  case "--help": fallthrough
  case "-h": fallthrough
  case "-?":
    help("")
    return

  default:
    fmt.Printf("Error: no such subcommand %s found. Stop.\n", os.Args[1])
    os.Exit(2)
    return
  }
}

