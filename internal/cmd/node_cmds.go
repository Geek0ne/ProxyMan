package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Geek0ne/ProxyMan/internal/config"
	"github.com/Geek0ne/ProxyMan/internal/engine"
	"github.com/Geek0ne/ProxyMan/internal/node"
	"github.com/Geek0ne/ProxyMan/internal/parser"
	"github.com/spf13/cobra"
)

// dryRun is the global --dry-run switch: every write is simulated, not performed.
var dryRun bool

// probe tuning, bound to flags on the probe command
var (
	probeTimeout time.Duration
	probeJobs    int
)

// ==================== list ====================
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List imported proxy nodes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadConfig()
		store := node.NewStore(cfg.WorkDir)
		nodes := store.List()

		if len(nodes) == 0 {
			fmt.Println("No nodes stored yet.")
			fmt.Println("Import some with: ProxyMan import \"vmess://...\"")
			return nil
		}

		fmt.Printf("Stored nodes (%d)  [%s]\n\n", len(nodes), store.Path())
		fmt.Printf("  %-3s %-26s %-14s %-22s %s\n", "#", "NAME", "TYPE", "ENDPOINT", "DELAY")
		fmt.Println("  " + strings.Repeat("-", 90))
		for i, n := range nodes {
			delay := "-"
			if n.LastDelay > 0 {
				delay = strconv.Itoa(n.LastDelay) + "ms"
			}
			fmt.Printf("  %-3d %-26s %-14s %-22s %s\n",
				i+1,
				truncate(n.Name, 26),
				n.Type,
				fmt.Sprintf("%s:%d", n.Server, n.Port),
				delay,
			)
		}
		return nil
	},
}

// ==================== remove ====================
var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a stored node by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadConfig()
		store := node.NewStore(cfg.WorkDir)
		if !store.Remove(args[0]) {
			return fmt.Errorf("node not found: %s", args[0])
		}
		if err := store.Save(); err != nil {
			return err
		}
		fmt.Printf("✓ Removed: %s\n", args[0])
		return nil
	},
}

// ==================== probe ====================
var probeCmd = &cobra.Command{
	Use:   "probe [name...]",
	Short: "Measure latency of stored nodes (TCP handshake)",
	Long: `Measure reachability and round-trip latency for stored nodes.

By default every stored node is probed. Pass one or more names to limit
the check to specific nodes. Results are sorted fastest-first and cached
for 'list'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadConfig()
		store := node.NewStore(cfg.WorkDir)

		var targets []*node.Node
		if len(args) == 0 {
			targets = store.List()
		} else {
			for _, name := range args {
				n := store.Get(name)
				if n == nil {
					return fmt.Errorf("node not found: %s", name)
				}
				targets = append(targets, n)
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf("no nodes to probe — import some first")
		}

		fmt.Printf("Probing %d node(s) — timeout %s, concurrency %d...\n\n",
			len(targets), probeTimeout, probeJobs)
		prober := node.NewProberWith(probeTimeout, probeJobs)
		results := prober.Probe(targets)

		fmt.Print(node.Report(results))

		reachable := 0
		for _, r := range results {
			if r.OK {
				reachable++
				store.RecordDelay(r.Name, r.DelayMS)
			}
		}
		if err := store.Save(); err != nil {
			fmt.Printf("  (warning: could not cache results: %v)\n", err)
		}
		fmt.Printf("\n%d/%d reachable\n", reachable, len(results))
		return nil
	},
}

// ==================== apply ====================
// applyCmd writes the stored node set into an engine configuration file.
var applyCmd = &cobra.Command{
	Use:   "apply [engine] [mode]",
	Short: "Write stored nodes into an engine config",
	Long: `Render every stored node into the configuration file of the chosen engine.

  mode: full         all traffic through the proxy group (default)
        china-split  domestic traffic direct, international via proxy

Use --dry-run to preview the generated file without writing it.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		engineName := strings.ToLower(args[0])
		mode := node.SplitFull
		if len(args) == 2 {
			mode = node.SplitMode(args[1])
			if mode != node.SplitFull && mode != node.SplitChina {
				return fmt.Errorf("invalid mode: %s (use 'full' or 'china-split')", args[1])
			}
		}

		cfg := config.LoadConfig()
		store := node.NewStore(cfg.WorkDir)
		nodes := store.List()
		if len(nodes) == 0 {
			return fmt.Errorf("no nodes stored — run 'ProxyMan import' first")
		}

		var canonical string
		switch engineName {
		case "v2ray", "xray":
			canonical = "v2ray"
		case "clash", "mihomo":
			canonical = "clash"
		default:
			return fmt.Errorf("unsupported engine: %s (use v2ray or clash)", engineName)
		}

		engineDir := filepath.Join(cfg.WorkDir, "engines", canonical)
		if _, err := os.Stat(engineDir); os.IsNotExist(err) {
			return fmt.Errorf("%s is not installed. Run 'ProxyMan install %s' first", canonical, canonical)
		}

		gen := node.NewGenerator(engineDir, cfg.LogLevel, mode)

		// dry-run renders to memory and never touches the filesystem
		if dryRun {
			var (
				path    string
				content string
				err     error
			)
			if canonical == "v2ray" {
				path, content, err = gen.RenderXray(nodes)
			} else {
				path, content, err = gen.RenderClash(nodes)
			}
			if err != nil {
				return err
			}
			fmt.Printf("=== DRY RUN — nothing written ===\n")
			fmt.Printf("Target: %s\n", path)
			fmt.Printf("Nodes:  %d   Split: %s\n\n", len(nodes), mode)
			fmt.Println(content)
			fmt.Println("=== end of preview ===")
			return nil
		}

		var (
			path string
			err  error
		)
		if canonical == "v2ray" {
			path, err = gen.WriteXray(nodes)
		} else {
			path, err = gen.WriteClash(nodes)
		}
		if err != nil {
			return err
		}

		fmt.Printf("✓ Config written: %s\n", path)
		fmt.Printf("  Nodes:     %d\n", len(nodes))
		fmt.Printf("  Split:     %s\n", mode)
		if mode == node.SplitChina {
			fmt.Println("  Routing:   domestic direct, international proxied")
		} else {
			fmt.Println("  Routing:   all traffic proxied")
		}
		fmt.Println("\nStart it with:")
		fmt.Printf("  ProxyMan start %s %s\n", canonical, path)
		return nil
	},
}

// ==================== log ====================
var logCmd = &cobra.Command{
	Use:   "log [engine]",
	Short: "Show engine logs (live with -f)",
	Long: `Display log output produced by a running proxy engine.

  -f, --follow   stream new log lines as they appear`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		engineName := strings.ToLower(args[0])
		cfg := config.LoadConfig()

		var canonical string
		switch engineName {
		case "v2ray", "xray":
			canonical = "v2ray"
		case "clash", "mihomo":
			canonical = "clash"
		default:
			return fmt.Errorf("unsupported engine: %s (use v2ray or clash)", engineName)
		}

		engineDir := filepath.Join(cfg.WorkDir, "engines", canonical)
		if _, err := os.Stat(engineDir); os.IsNotExist(err) {
			return fmt.Errorf("%s is not installed. Run 'ProxyMan install %s' first", canonical, canonical)
		}

		// Collect candidate log locations, newest first.
		var candidates []string
		for _, name := range []string{"access.log", "error.log", "engine.log", "proxy.log"} {
			candidates = append(candidates, filepath.Join(engineDir, name))
		}
		candidates = append(candidates,
			filepath.Join(cfg.WorkDir, "logs", canonical+".log"),
			filepath.Join(cfg.WorkDir, "logs", canonical+".txt"),
		)

		var logFile string
		for _, c := range candidates {
			if fi, err := os.Stat(c); err == nil && !fi.IsDir() && fi.Size() > 0 {
				logFile = c
				break
			}
		}
		if logFile == "" {
			return fmt.Errorf(
				"no log file found for %s\nsearched:\n  %s\n\nhint: engines only write logs when started with a log path configured",
				canonical, strings.Join(candidates, "\n  "))
		}

		fmt.Printf("Log: %s\n", logFile)
		if !followLogs {
			fmt.Println(strings.Repeat("-", 60))
		}
		return streamFile(logFile, followLogs)
	},
}

var followLogs bool

func streamFile(path string, follow bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !follow {
		return nil
	}

	// Follow mode: poll the file for appended lines.
	fmt.Println("[following — press Ctrl-C to stop]")
	offset, _ := f.Seek(0, 1)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fi.Size() < offset {
			offset = 0 // truncated / rotated
		}
		if fi.Size() == offset {
			continue
		}
		f.Seek(offset, 0)
		rd := bufio.NewReader(f)
		for {
			line, err := rd.ReadString('\n')
			if line != "" {
				fmt.Print(line)
			}
			if err != nil {
				break
			}
		}
		offset, _ = f.Seek(0, 1)
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func init() {
	logCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "stream new log lines")

	probeCmd.Flags().DurationVar(&probeTimeout, "timeout", node.DefaultProbeTimeout,
		"per-node reachability timeout (e.g. 500ms, 10s)")
	probeCmd.Flags().IntVar(&probeJobs, "jobs", node.DefaultProbeJobs,
		fmt.Sprintf("concurrent probes (1-%d)", node.MaxProbeJobs))

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate write operations without persisting them")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(probeCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(logCmd)
}

// ensure the imports used by import --apply stay referenced
var _ = engine.NewManager
var _ = parser.DetectProtocol
