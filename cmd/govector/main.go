package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/DotNetAge/govector/api"
	"github.com/DotNetAge/govector/core"
)

const (
	colorReset     = "\033[0m"
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorBoldGreen = "\033[1;32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorCyan      = "\033[36m"
)

func main() {
	if err := runCommand(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
}

func parseDbFile(args []string) (string, []string) {
	dbFile := "govector.db"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		dbFile = args[0]
		args = args[1:]
	}
	return dbFile, args
}

func runCommand(args []string) error {
	if len(args) < 2 {
		runTUI("govector.db", "default")
		return nil
	}

	command := args[1]

	if command == "serve" {
		handleServe(args[2:])
		return nil
	}

	// Known commands
	knownCommands := map[string]bool{
		"upsert": true,
		"search": true,
		"count":  true,
		"delete": true,
		"ls":     true,
		"rm":     true,
	}

	// If the command is not a known subcommand, treat the arguments as TUI arguments: [dbfile] [options]
	if !knownCommands[command] {
		if command == "-h" || command == "--help" {
			printUsage()
			return nil
		}

		// Treat args[1:] as TUI arguments
		importFlags := args[1:]
		dbFile, cmdArgs := parseDbFile(importFlags)

		importFs := flag.NewFlagSet("tui", flag.ContinueOnError)
		var colName string
		importFs.StringVar(&colName, "collection", "default", "Collection name")
		importFs.StringVar(&colName, "c", "default", "Collection name (shorthand)")

		err := importFs.Parse(cmdArgs)
		if err != nil {
			return err
		}

		runTUI(dbFile, colName)
		return nil
	}

	dbFile, cmdArgs := parseDbFile(args[2:])

	switch command {
	case "upsert":
		handleupsert(dbFile, cmdArgs)
	case "search":
		handlesearch(dbFile, cmdArgs)
	case "count":
		handlecount(dbFile, cmdArgs)
	case "delete":
		handledelete(dbFile, cmdArgs)
	case "ls":
		handleLs(dbFile, cmdArgs)
	case "rm":
		handleRm(dbFile, cmdArgs)
	}
	return nil
}

func printUsage() {
	fmt.Printf("%sGoVector CLI Usage:%s\n", colorCyan, colorReset)
	fmt.Println("  govector <command> [dbfile] [options]")
	fmt.Printf("\n%sCommands:%s\n", colorYellow, colorReset)
	fmt.Println("  upsert [dbfile] -j='[{\"id\":\"1\",\"vector\":[0.1,0.2],\"payload\":{}}]' -c='test'")
	fmt.Println("  search [dbfile] -v='[0.1,0.2]' -l=10 -c='test'")
	fmt.Println("  count  [dbfile] -c='test'")
	fmt.Println("  delete [dbfile] -i='[\"1\",\"2\"]' -c='test'")
	fmt.Println("  ls     [dbfile]")
	fmt.Println("  rm     [dbfile] -c='test'")
	fmt.Println("  serve  -d=<dbfile> -p=18080")
	fmt.Println("\nRun without arguments to start interactive TUI.")
}

func handleServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	var port int
	var dbPath string
	fs.IntVar(&port, "port", 18080, "Port for the HTTP API server")
	fs.IntVar(&port, "p", 18080, "Port for the HTTP API server (shorthand)")
	fs.StringVar(&dbPath, "db", "govector.db", "Path to the BoltDB storage file")
	fs.StringVar(&dbPath, "d", "govector.db", "Path to the BoltDB storage file (shorthand)")
	
	err := fs.Parse(args)
	if err != nil {
		return
	}

	fmt.Printf("\n%s=== GoVector Server (Port: %d, DB: %s) ===%s\n", colorBoldGreen, port, dbPath, colorReset)

	store, err := core.NewStorage(dbPath)
	if err != nil {
		fmt.Printf("%sFailed to initialize storage: %v%s\n", colorRed, err, colorReset)
		return
	}
	defer store.Close()

	addr := fmt.Sprintf(":%d", port)
	server := api.NewServer(addr, store)

	serverErrors := make(chan error, 1)
	go func() {
		fmt.Printf("%sAPI Server ready on http://localhost%s%s\n", colorGreen, addr, colorReset)
		fmt.Printf("%s(Press Ctrl+C to stop the server and return)%s\n", colorYellow, colorReset)
		serverErrors <- server.Start()
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(osSignals)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("%sError starting server: %v%s\n", colorRed, err, colorReset)
		}
	case sig := <-osSignals:
		fmt.Printf("\n%sShutting down server... (Signal: %v)%s\n", colorYellow, sig, colorReset)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Stop(ctx); err != nil {
			fmt.Printf("%sShutdown error: %v%s\n", colorRed, err, colorReset)
		}
	}
}

func getCollectionForWrite(dbFile string, colName string, defaultDim int) (*core.Collection, func()) {
	store, err := core.NewStorage(dbFile)
	if err != nil {
		fmt.Printf("%sFailed to open storage: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	meta, err := store.LoadCollectionMeta(colName)
	if err != nil {
		col, err := core.NewCollectionWithParams(colName, defaultDim, core.Cosine, store, true, core.DefaultHNSWParams())
		if err != nil {
			store.Close()
			fmt.Printf("%sFailed to create collection: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}
		return col, func() { store.Close() }
	}

	col, err := core.NewCollectionWithParams(meta.Name, meta.VectorLen, meta.Metric, store, meta.UseHNSW, meta.HNSWParams)
	if err != nil {
		store.Close()
		fmt.Printf("%sFailed to open collection: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	return col, func() { store.Close() }
}

func getCollectionForRead(dbFile string, colName string) (*core.Collection, func()) {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}

	store, err := core.NewStorage(dbFile)
	if err != nil {
		fmt.Printf("%sFailed to open storage: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	meta, err := store.LoadCollectionMeta(colName)
	if err != nil {
		store.Close()
		fmt.Println("collection不存在")
		os.Exit(1)
	}

	col, err := core.NewCollectionWithParams(meta.Name, meta.VectorLen, meta.Metric, store, meta.UseHNSW, meta.HNSWParams)
	if err != nil {
		store.Close()
		fmt.Printf("%sFailed to open collection: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	return col, func() { store.Close() }
}

func handleupsert(dbFile string, args []string) {
	fs := flag.NewFlagSet("upsert", flag.ExitOnError)
	var jsonStr, colName string
	fs.StringVar(&jsonStr, "json", "", "JSON points data")
	fs.StringVar(&jsonStr, "j", "", "JSON points data (shorthand)")
	fs.StringVar(&colName, "collection", "default", "Collection name")
	fs.StringVar(&colName, "c", "default", "Collection name (shorthand)")
	fs.Parse(args)

	if jsonStr == "" {
		fmt.Println("Error: -json or -j is required")
		return
	}

	var points []core.PointStruct
	if err := json.Unmarshal([]byte(jsonStr), &points); err != nil {
		fmt.Printf("Invalid JSON: %v\n", err)
		return
	}

	dim := 3
	if len(points) > 0 {
		dim = len(points[0].Vector)
	}

	col, closer := getCollectionForWrite(dbFile, colName, dim)
	defer closer()

	if err := col.Upsert(points); err != nil {
		fmt.Printf("upsert failed: %v\n", err)
		return
	}
	fmt.Printf("%sSuccessfully upserted %d points to collection '%s'.%s\n", colorGreen, len(points), colName, colorReset)
}

func handlesearch(dbFile string, args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	var vectorStr, colName string
	var limit int
	fs.StringVar(&vectorStr, "vector", "", "Query vector as JSON array")
	fs.StringVar(&vectorStr, "v", "", "Query vector as JSON array (shorthand)")
	fs.IntVar(&limit, "limit", 10, "Number of results")
	fs.IntVar(&limit, "l", 10, "Number of results (shorthand)")
	fs.StringVar(&colName, "collection", "default", "Collection name")
	fs.StringVar(&colName, "c", "default", "Collection name (shorthand)")
	fs.Parse(args)

	var vector []float32
	if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
		fmt.Printf("Invalid vector JSON: %v\n", err)
		return
	}

	col, closer := getCollectionForRead(dbFile, colName)
	defer closer()

	results, err := col.Search(vector, nil, limit)
	if err != nil {
		fmt.Printf("search failed: %v\n", err)
		return
	}

	fmt.Printf("%sSearch Results (%d):%s\n", colorBoldGreen, len(results), colorReset)
	for i, res := range results {
		fmt.Printf("%d. ID: %s, Score: %f, Payload: %v\n", i+1, res.ID, res.Score, res.Payload)
	}
}

func handlecount(dbFile string, args []string) {
	fs := flag.NewFlagSet("count", flag.ExitOnError)
	var colName string
	fs.StringVar(&colName, "collection", "default", "Collection name")
	fs.StringVar(&colName, "c", "default", "Collection name (shorthand)")
	fs.Parse(args)

	col, closer := getCollectionForRead(dbFile, colName)
	defer closer()

	fmt.Printf("Collection '%s' count: %s%d%s\n", colName, colorGreen, col.Count(), colorReset)
}

func handledelete(dbFile string, args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	var idsStr, colName string
	fs.StringVar(&idsStr, "ids", "", "JSON array of IDs to delete")
	fs.StringVar(&idsStr, "i", "", "JSON array of IDs to delete (shorthand)")
	fs.StringVar(&colName, "collection", "default", "Collection name")
	fs.StringVar(&colName, "c", "default", "Collection name (shorthand)")
	fs.Parse(args)

	var ids []string
	if err := json.Unmarshal([]byte(idsStr), &ids); err != nil {
		fmt.Printf("Invalid IDs JSON: %v\n", err)
		return
	}

	col, closer := getCollectionForRead(dbFile, colName)
	defer closer()

	count, err := col.Delete(ids, nil)
	if err != nil {
		fmt.Printf("delete failed: %v\n", err)
		return
	}
	fmt.Printf("%sdeleted %d points.%s\n", colorGreen, count, colorReset)
}

func handleLs(dbFile string, args []string) {
	fs := flag.NewFlagSet("ls", flag.ExitOnError)
	fs.Parse(args)

	store, err := core.NewStorage(dbFile)
	if err != nil {
		fmt.Printf("Failed to open storage: %v\n", err)
		return
	}
	defer store.Close()

	metas, err := store.ListCollectionMetas()
	if err != nil {
		fmt.Printf("Failed to list collections: %v\n", err)
		return
	}

	if len(metas) == 0 {
		fmt.Printf("%sNo collections found in %s.%s\n", colorYellow, dbFile, colorReset)
		return
	}

	fmt.Printf("\n%sCollections in [%s]:%s\n", colorBoldGreen, dbFile, colorReset)
	fmt.Printf("%-20s | %-5s | %-10s | %-6s\n", "NAME", "DIM", "METRIC", "HNSW")
	fmt.Println(strings.Repeat("-", 50))
	for _, m := range metas {
		fmt.Printf("%-20s | %-5d | %-10s | %-6t\n", m.Name, m.VectorLen, m.Metric, m.UseHNSW)
	}
	fmt.Println(strings.Repeat("-", 50))
}

func handleRm(dbFile string, args []string) {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	var colName string
	fs.StringVar(&colName, "collection", "", "Collection name")
	fs.StringVar(&colName, "c", "", "Collection name (shorthand)")
	fs.Parse(args)

	if colName == "" {
		fmt.Println("Error: -collection or -c is required")
		return
	}

	if colName == "default" {
		fmt.Println("Error: Cannot delete the 'default' collection")
		return
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}

	store, err := core.NewStorage(dbFile)
	if err != nil {
		fmt.Printf("Failed to open storage: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	err = store.DropCollection(colName)
	if err != nil {
		fmt.Printf("Failed to drop collection: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%sSuccessfully removed collection '%s'.%s\n", colorGreen, colName, colorReset)
}

func printTUIBanner() {
	fmt.Printf("\n%s   ______     __    __          __            %s\n", colorBoldGreen, colorReset)
	fmt.Printf("%s  / ____/___ | |  / /__  _____/ /_____  _____ %s\n", colorBoldGreen, colorReset)
	fmt.Printf("%s / / __/ __ \\| | / / _ \\/ ___/ __/ __ \\/ ___/ %s\n", colorBoldGreen, colorReset)
	fmt.Printf("%s/ /_/ / /_/ /| |/ /  __/ /__/ /_/ /_/ / /     %s\n", colorBoldGreen, colorReset)
	fmt.Printf("%s\\____/\\____/ |___/\\___/\\___/\\__/\\____/_/      %s\n", colorBoldGreen, colorReset)
	fmt.Printf("\n")
	fmt.Printf("%s          Copyright (c) 2026 Raya Info co.,Ltd%s\n", colorGreen, colorReset)
	fmt.Printf("\n")
}

func printTUIHelp() {
	fmt.Printf("%sAvailable Commands:%s\n", colorBoldGreen, colorReset)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  %s%-6s%s\t: %s\n", colorGreen, "upsert", colorReset, "Insert/update points (-j='[{\"id\":\"1\",\"vector\":[0.1]}]')")
	fmt.Fprintf(w, "  %s%-6s%s\t: %s\n", colorGreen, "search", colorReset, "Search nearest neighbors (-v='[0.1]' -l=5)")
	fmt.Fprintf(w, "  %s%-6s%s\t: %s\n", colorGreen, "count", colorReset, "Count points in the current collection")
	fmt.Fprintf(w, "  %s%-6s%s\t: %s\n", colorGreen, "delete", colorReset, "Delete points by IDs (-i='[\"1\",\"2\"]')")
	fmt.Fprintf(w, "  %s%-6s%s\t: %s\n", colorGreen, "/?", colorReset, "Show this help message")
	fmt.Fprintf(w, "  %s%-6s%s\t: %s\n", colorGreen, "exit", colorReset, "Exit the TUI")
	w.Flush()
}

func runTUI(dbFile string, defaultColName string) {
	printTUIBanner()
	printTUIHelp()
	
	reader := bufio.NewReader(os.Stdin)

	// Ensure DB and collection exist. 
	_, closer := getCollectionForWrite(dbFile, defaultColName, 3)
	closer()

	for {
		fmt.Printf("\n%s[%s | %s]%s Command (upsert/search/count/delete/serve/?/exit) > %s", colorBoldGreen, dbFile, defaultColName, colorGreen, colorReset)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nGoodbye!")
			return
		}
		input = strings.TrimSpace(input)

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		if cmd == "/?" || cmd == "help" {
			printTUIHelp()
			continue
		}

		// Inject default collection if needed and not present
		if cmd == "upsert" || cmd == "search" || cmd == "count" || cmd == "delete" {
			hasColFlag := false
			for _, arg := range args {
				if strings.HasPrefix(arg, "-c=") || arg == "-c" || strings.HasPrefix(arg, "-collection=") || arg == "-collection" {
					hasColFlag = true
					break
				}
			}
			if !hasColFlag {
				args = append(args, "-c="+defaultColName)
			}
		}

		switch cmd {
		case "upsert":
			handleupsert(dbFile, args)
		case "search":
			handlesearch(dbFile, args)
		case "count":
			handlecount(dbFile, args)
		case "delete":
			handledelete(dbFile, args)
		default:
			fmt.Printf("%sUnknown command: %s. Use /? for help.%s\n", colorRed, cmd, colorReset)
		}
	}
}
