package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	var (
		recordPath = flag.String("record", "", "path to a level-4 ADL record (JSON file or JSONL write-ahead log)")
		policyDir  = flag.String("policy-dir", "", "directory of policy source files to reconstruct the PAP from")
		managerURL = flag.String("manager-url", "", "manager base URL; resolves each policy hash via /v1/policy-hash/{hash}")
		pipStore   = flag.String("pip-store", "", "PIP attribute file-store used to reconstruct information values")
		mappings   = flag.String("mappings", "", "override request-mappings (default: taken from adl.core.configuration)")
		verbose    = flag.Bool("v", false, "verbose diagnostics")
	)
	flag.Parse()

	if *recordPath == "" {
		fmt.Fprintln(os.Stderr, "error: -record is required")
		flag.Usage()
		os.Exit(2)
	}

	level := slog.LevelError
	if *verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	rec, err := ReadRecord(*recordPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading record: %v\n", err)
		os.Exit(2)
	}

	res, err := Replay(context.Background(), rec, Options{
		PolicyDir:  *policyDir,
		ManagerURL: *managerURL,
		PIPStore:   *pipStore,
		Mappings:   *mappings,
		Logger:     logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during replay: %v\n", err)
		os.Exit(2)
	}

	for h, ok := range res.ResolvedHashes {
		status := "resolved"
		if !ok {
			status = "UNRESOLVED"
		}
		fmt.Printf("policy hash %s: %s\n", h, status)
	}
	fmt.Printf("expected decision: %v\nreplayed decision: %v\n", res.ExpectedDecision, res.ActualDecision)

	if res.Pass {
		fmt.Println("PASS")
		os.Exit(0)
	}
	fmt.Println("FAIL")
	if res.Diff != "" {
		fmt.Println(res.Diff)
	}
	os.Exit(1)
}
