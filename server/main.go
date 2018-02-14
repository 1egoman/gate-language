package main

import (
  "os"
  "fmt"
  "encoding/json"
  "flag"
  "net/http"
  "bytes"
  "time"

  // For reading file from disk
  "io/ioutil"

  // Allow for file watching in `go run`
  "github.com/radovskyb/watcher"
  "github.com/gorilla/websocket"
)

type Summary struct {
  Gates []*Gate
  Wires []*Wire
  Contexts []*CallingContext
  Outputs []*Wire
}

func run(input string, verbose bool) (*Summary, error) {
  result, err := Tokenizer(input)
  if err != nil {
    return nil, err
  }

  if verbose {
    fmt.Println("RESULTS FROM TOKENIZER:")
    PrintAst(result, 0, "")
    fmt.Println()
    fmt.Println()
    fmt.Println()
  }

  stack := []*StackFrame{ &StackFrame{} }

  if result == nil {
    fmt.Println("Result was nil!")
    return nil, nil
  }

  var allGates []*Gate
  var allWires []*Wire
  var allContexts []*CallingContext
  var finalOutputs []*Wire

  resultValues := *result

  for len(resultValues) > 0 {
    if verbose { fmt.Println("==========>", resultValues) }
    gates, wires, contexts, outputs, err := Parse(&resultValues, stack)

    allGates = append(allGates, gates...)
    allWires = append(allWires, wires...)
    allContexts = append(allContexts, contexts...)
    finalOutputs = outputs

    if err != nil {
      return nil, err
    }

    if verbose {
      fmt.Println("GATES:")
      for _, gate := range gates {
        fmt.Printf("- %s ", gate.Type)

        fmt.Printf("(IN:")
        for _, input := range gate.Inputs {
          fmt.Printf(" %+v", input)
        }
        fmt.Printf(") ")

        fmt.Printf("(OUT:")
        for _, output := range gate.Outputs {
          fmt.Printf(" %+v", output)
        }
        fmt.Printf(")")

        fmt.Printf(` frame=%d`, gate.CallingContext)
        fmt.Printf(` label="%s"`, gate.Label)
        fmt.Printf("\n")
      }
      fmt.Println("===")

      fmt.Println("WIRES:")
      for _, wire := range wires {
        fmt.Printf("- %+v\n", wire)
      }
      fmt.Println("===")

      fmt.Println("OUTPUTS:")
      for _, output := range outputs {
        fmt.Printf("- %+v\n", output)
      }
      fmt.Println("===")
    }
  }

  if verbose {
    fmt.Println("FINAL OUTPUTS", finalOutputs)
  }

  // Add child contexts to parent contexts
  // THis can't be done in `Parse` because it never has a reference to all contexts at once.
  for _, context := range allContexts {
    if context.Parent > 0 {
      for _, parentContext := range allContexts {
        if parentContext.Id == context.Parent {
          parentContext.Children = append(parentContext.Children, context.Id)
          break
        }
      }
    }
  }

  // Print out a summary of the results to that they can be rendered.
  summary := Summary{
    Gates: allGates,
    Wires: allWires,
    Contexts: allContexts,
    Outputs: finalOutputs,
  }

  return &summary, nil
}


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
    fmt.Println("   --verbose  Print debugging information")

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
  case "run":
    fmt.Println("Starting lovelace server...")

    runFlags := flag.NewFlagSet("run", flag.ExitOnError)
    runVerbose := runFlags.Bool("verbose", false, "Print debug information")
    runPort := runFlags.Int("port", 8080, "")

    runFlags.Usage = func() { help("run") }
    runFlags.Parse(os.Args[2:])

    if len(os.Args) < 3 {
      fmt.Println("Error: not enough arguments were passed to run. Stop.")
      os.Exit(2)
      return
    }

    var connections []*websocket.Conn
    var upgrader = websocket.Upgrader{
      ReadBufferSize: 1024,
      WriteBufferSize: 1024,
      CheckOrigin: func(r *http.Request) bool {
        return true
      },
    }

    // On the first thread, accept websocket requests.
    http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
      conn, err := upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }

      // Add the connection to the group of collections.
      connections = append(connections, conn)

      fmt.Println("Client subscribed")
    })

    // In a second thread, watch for file changes. If a file changes, rebuild it.
    go func() {
      watcher := watcher.New()
      watcher.SetMaxEvents(1)

      fmt.Printf("Watching %s\n", os.Args[2])
      if err := watcher.Add(os.Args[2]); err != nil {
        fmt.Printf("Error watching source file %s: %s. Stop.\n", os.Args[2], err)
        os.Exit(2)
        return
      }

      go func() {
        for {
          select {
          case event := <-watcher.Event:
            if *runVerbose {
              fmt.Printf("Event: %s\n", event)
            }

            // Read the contents of the file.
            source, err := ioutil.ReadFile(os.Args[2])
            if err != nil {
              fmt.Printf("Error reading file %s: %s. Stop.\n", os.Args[2], err);
              os.Exit(2)
              return
            }

            // Compile the source
            fmt.Printf("Compiling %s ... ", os.Args[2])
            summary, err := run(string(source), *runVerbose)

            // Print any errors received in the compilation process
            if err != nil {
              fmt.Printf("ERROR\nError: %s\n", err)
            } else {
              fmt.Printf("OK\n")
            }

            var payload []byte
            if err == nil {
              // The ast was compiled successfully.
              payload, err = json.Marshal(summary)
              if err != nil {
                fmt.Println("Error serializing ast: %s. Stop.", err)
              }
            } else {
              // An error occured.
              payload, err = json.Marshal(map[string]string{ "Error": err.Error() })
              if err != nil {
                fmt.Println("Error serializing error: %s. Stop.", err)
              }
            }

            // Then, send the ast over the websocket.
            for index, conn := range connections {
              err = conn.WriteMessage(websocket.TextMessage, payload)
              if err != nil {
                fmt.Println("Error sending payload to websocket client %d: %s. Stop.", index, err)
              }
            }

          case err := <-watcher.Error:
            fmt.Println("error:", err)
          case <-watcher.Closed:
            return
          }
        }
      }()

      if err := watcher.Start(time.Millisecond * 100); err != nil {
        fmt.Printf("Error in filesystem watcher: %s. Stop.\n", err)
      }
    }()

    fmt.Printf("Started server on %d\n", *runPort)
    err := http.ListenAndServe(fmt.Sprintf(":%d", *runPort), nil)
    panic(err)


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
      fmt.Println("Error tokenizing file %s: %s. Stop.", os.Args[2], err);
      os.Exit(2)
    }

    PrintAst(result, 0, "")

  // lovel build foo.bit
  case "build":
    if len(os.Args) < 3 {
      fmt.Println("No file path was passed to build. Stop.")
      os.Exit(2)
      return
    }

    // Add flags
    buildFlags := flag.NewFlagSet("build", flag.ExitOnError)
    buildVerbose := buildFlags.Bool("verbose", false, "Print debug information")
    buildFlags.Usage = func() { help("serve") }
    buildFlags.Parse(os.Args[3:])

    // Read source code from disk
    source, err := ioutil.ReadFile(os.Args[2])
    if err != nil {
      fmt.Printf("Error reading file %s: %s. Stop.\n", os.Args[2], err);
      os.Exit(2)
      return
    }

    wireId = 0
    gateId = 0
    stackFrameId = 0

    summary, err := run(string(source), *buildVerbose)
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
  case "serve":
    fmt.Println("Starting lovelace server...")

    serverFlags := flag.NewFlagSet("serve", flag.ExitOnError)
    serverVerbose := serverFlags.Bool("verbose", false, "Print debug information")
    serverPort := serverFlags.Int("port", 8080, "")
    serverFlags.Usage = func() { help("serve") }

    serverFlags.Parse(os.Args[2:])

    // Set a flag to inform the rest of the system that this process is running in server mode. This
    // changes how a few things work, including:
    // - Blocking local file imports. When running as a server, we don't want the user to have
    // access to files on the server in their program.
    isRunningInServer = true

    http.HandleFunc("/v1/compile", func(w http.ResponseWriter, r *http.Request) {
      // Allow Cross Origin Resource Sharing
      w.Header().Set("Access-Control-Allow-Origin", "*")
      w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

      buf := new(bytes.Buffer)
      buf.ReadFrom(r.Body)
      source := buf.String() // Does a complete copy of the bytes in the buffer.

      wireId = 0
      gateId = 0
      stackFrameId = 0
      summary, err := run(source, *serverVerbose)
      if err != nil {
        json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
      } else {
        json.NewEncoder(w).Encode(summary)
      }
    })

    http.HandleFunc("/v1/run", func(w http.ResponseWriter, r *http.Request) {
      // Allow CORS here By * or specific origin
      w.Header().Set("Access-Control-Allow-Origin", "*")
      w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

      // Decode the body
      decoder := json.NewDecoder(r.Body)
      var body struct {
        Gates []*Gate
        Wires []*Wire
      }
      decoder.Decode(&body)

      gates, wires := Execute(body.Gates, body.Wires)

      json.NewEncoder(w).Encode(map[string]interface{}{"Gates": gates, "Wires": wires})
    })

    fmt.Printf("Started server on %d\n", *serverPort)
    err := http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), nil)
    panic(err)

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

