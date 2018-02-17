package main

import (
  "fmt"
  "flag"
  "os"

  "net/http"
  "bytes"
  "encoding/json"
)

func Serve() {
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
}
