package main

import (
  "fmt"
  "flag"
  "os"
  "encoding/json"
  "time"

  // Required to run the server
  "net/http"

  // For reading file from disk
  "io/ioutil"

  // Required for `openBrowser`
  "runtime"
  "os/exec"

  // Allow for file watching in `go run`
  "github.com/radovskyb/watcher"

  // Websocket server upgrade and helper library
  "github.com/gorilla/websocket"
)

func processFile(path string, verbose bool) *Summary {
  // Set initial ids
  wireId = 0
  gateId = 0
  stackFrameId = 0

  // Read source code from disk
  source, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Printf("Error reading file %s: %s. Stop.\n", os.Args[2], err);
    os.Exit(2)
    return nil
  }

  // Tokenize the source
  tokens, err := Tokenizer(string(source))
  if err != nil {
    fmt.Printf("Error tokenizing %s: %s. Stop.\n", os.Args[2], err);
    os.Exit(2)
    return nil
  }

  if tokens == nil {
    fmt.Println("Error: Tokenizer returned nil. Stop.")
    os.Exit(2)
    return nil
  }

  if verbose {
    fmt.Println("RESULTS FROM TOKENIZER:")
    PrintAst(tokens, 0, "")
    fmt.Println()
    fmt.Println()
    fmt.Println()
  }


  stack := []*StackFrame{ &StackFrame{} }

  var allGates []*Gate
  var allWires []*Wire
  var allContexts []*CallingContext
  var finalOutputs []*Wire

  resultValues := *tokens

  for len(resultValues) > 0 {
    if verbose { fmt.Println("==========>", resultValues) }
    gates, wires, contexts, outputs, err := Parse(&resultValues, stack)

    allGates = append(allGates, gates...)
    allWires = append(allWires, wires...)
    allContexts = append(allContexts, contexts...)
    finalOutputs = outputs

    if err != nil {
      fmt.Printf("Error parsing %s: %s. Stop.\n", os.Args[2], err)
      os.Exit(2)
      return nil
    }

    if verbose {
      // fmt.Println("GATES:")
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
  // This can't be done in `Parse` because it never has a reference to all contexts at once.
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
  return &Summary{
    Gates: allGates,
    Wires: allWires,
    Contexts: allContexts,
    Outputs: finalOutputs,
  }
}

// openBrowser tries to open the URL in a browser,
// and returns whether it succeed in doing so.
func openBrowser(url string) bool {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

func Run() {
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
  var lastPayload []byte = nil

  lastPayload, err := json.Marshal(processFile(os.Args[2], *runVerbose))
  if err != nil {
    fmt.Println("Error serializing result: %s. Stop.", err);
    os.Exit(2)
    return
  }

  // On the first thread, accept websocket requests and http requests to run the ast that was sent
  // to the client in the websocket push.
  http.HandleFunc("/v1/websocket", func(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
      fmt.Println(err)
      return
    }

    if lastPayload != nil {
      err = conn.WriteMessage(websocket.TextMessage, lastPayload)
      if err != nil {
        fmt.Printf("Error sending payload to new websocket client: %s.\n", err)
      }
    }

    // Add the connection to the group of collections.
    connections = append(connections, conn)

    fmt.Println("Client subscribed")
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
              fmt.Printf("Error serializing ast: %s.\n", err)
            }
          } else {
            // An error occured.
            payload, err = json.Marshal(map[string]string{ "Error": err.Error() })
            if err != nil {
              fmt.Printf("Error serializing error: %s.\n", err)
            }
          }

          // Then, send the ast over the websocket.
          for index, conn := range connections {
            err = conn.WriteMessage(websocket.TextMessage, payload)
            if err != nil {
              fmt.Printf("Error sending payload to websocket client %d: %s.\n", index, err)

              // Close the connection. The client is smart enough to reconnect when this happens.
              conn.Close()
              connections = append(connections[:index], connections[index+1:]...)
            }
          }

          // Save the last push. Any new clients will receive this push in order for it to get up
          // to speed.
          lastPayload = payload

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

  openBrowser("http://lovelace-preview.surge.sh/?preview=true")

  fmt.Printf("Started server on %d\n", *runPort)
  err = http.ListenAndServe(fmt.Sprintf(":%d", *runPort), nil)
  panic(err)
}
