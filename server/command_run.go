package main

import (
  "fmt"
  "flag"
  "os"
  "encoding/json"
  "time"

  // Required to run the server
  "net/http"

  // Required for `openBrowser`
  "runtime"
  "os/exec"

  // Allow for file watching in `go run`
  "github.com/radovskyb/watcher"

  // Websocket server upgrade and helper library
  "github.com/gorilla/websocket"
)

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
  runFlags := flag.NewFlagSet("run", flag.ExitOnError)
  runVerbose := runFlags.Bool("verbose", false, "Print debug information")
  runMaxCallDepth := runFlags.Int("max-call-depth", -1, "Set the maximum call depth")
  runPort := runFlags.Int("port", 8080, "")

  runFlags.Usage = func() { help("run") }
  runFlags.Parse(os.Args[2:])

  fmt.Println(runFlags.Args())

  if runFlags.NArg() != 1 {
    fmt.Println("Error: No file was passed to run. Stop.")
    os.Exit(2)
    return
  }

  filePath := runFlags.Args()[0]

  // Set max call depth if a value was specified.
  if *runMaxCallDepth != -1 {
    INVOCATION_MAX_RECURSION_DEPTH = *runMaxCallDepth
  }

  fmt.Println("Starting lovelace server...")

  var connections []*websocket.Conn
  var upgrader = websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
      return true
    },
  }
  var lastPayload []byte = nil

  summary, err := RunFile(runFlags.Args()[0], *runVerbose)
  if err != nil {
    lastPayload, err = json.Marshal(map[string]string{"Error": err.Error()})
    if err != nil {
      fmt.Println("Error serializing error payload: %s. Stop.", err);
      os.Exit(2)
      return
    }
  } else {
    lastPayload, err = json.Marshal(summary)
    if err != nil {
      fmt.Println("Error serializing result: %s. Stop.", err);
      os.Exit(2)
      return
    }
  }

  fmt.Println("Initial compile was successful. Watching...")

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

    fmt.Printf("Watching %s\n", filePath)
    if err := watcher.Add(filePath); err != nil {
      fmt.Printf("Error watching source file %s: %s. Stop.\n", filePath, err)
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

          // Compile the source
          fmt.Printf("Compiling %s ... ", filePath)
          summary, err := RunFile(filePath, *runVerbose)

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

  fmt.Println("Opening browser...");
  openBrowser(fmt.Sprintf("http://lovelace-preview.surge.sh/?preview=true&server=http://localhost:%d", *runPort))

  fmt.Printf("Started server on %d\n", *runPort)
  err = http.ListenAndServe(fmt.Sprintf(":%d", *runPort), nil)
  panic(err)
}
