# Go Coverage Agent with LCOV Client

This guide explains how to build and run a Go application with integrated coverage instrumentation and use a Rust-based client to retrieve and process the coverage data in LCOV format.

## Prerequisites

- Go 1.20+ installed
- Rust (with `cargo`)
- The Go application should import the coverage agent so the agent is started automatically

---

## Prepare Coverage Directory

Create the directory that Go will use to store coverage counter files:

```bash
   mkdir .coverdata
```

---

## Build the Go Application with Coverage Instrumentation

Use the following flags when compiling your Go application:

```bash
  go build -cover -coverpkg=./... -covermode=atomic -o myapi .
```

This enables coverage tracking across all packages (`./...`) using `atomic` mode.

---

## Run the Application with Coverage Enabled

Execute your application and specify the directory for coverage data:

```bash
  GOCOVERDIR=.coverdata ./myapi
```

The agent should start automatically and listen for incoming client requests.

To enable coverage flushing during runtime, make sure your Go application's `main` function includes the following endpoint:

```go
r.GET("/flushcov", func(c *gin.Context) {
    _ = os.Setenv("GOCOVERDIR", ".coverdata")
    err := coverage.WriteCountersDir(".coverdata")
    if err != nil {
        log.Println("Failed to write counters:", err)
        c.JSON(500, gin.H{"error": err.Error()})
    } else {
        log.Println(">> coverage counters flushed")
        c.JSON(200, gin.H{"status": "ok"})
    }
})
```

This endpoint ensures that the current in-memory coverage counters are flushed to disk before the client retrieves them.

---

## Retrieve Coverage Data with the Rust Client

The Rust client is responsible for requesting the LCOV coverage data and saving it locally. You can either:

- Create a main function in the client to call the export logic directly


Then build and run the client:

```bash
  cargo run
```

- Call the client logic from an integrated fuzzer

If you’re using a fuzzer that supports integrating coverage clients, you can import the client module and invoke the coverage export functionality as part of the fuzzing workflow.

---

## Notes for Testing

- You can simulate traffic to your Go API using tools like `curl` to generate coverage.
- Make sure the coverage endpoint is called before the client connects, to ensure counters are flushed.
- After running the client, the LCOV file and optional HTML report should be generated.

---

## Output

The client will produce:
-  raw LCOV report
-  optional HTML report directory (requires `genhtml` and correct source paths)

---