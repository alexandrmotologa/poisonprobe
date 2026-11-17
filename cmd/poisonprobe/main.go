package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/audit"
	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/internal/sanitize"
	"github.com/alexandrmotologa/poisonprobe/internal/server"
)

const version = "1.0.0"

func printUsage() {
	fmt.Printf("poisonprobe v%s - Dataset Poisoning, Backdoor & Label Noise Auditor\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  poisonprobe <command> [arguments] [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  scan      Analyze dataset and print anomaly statistics and triggers")
	fmt.Println("  sanitize  Export clean certified dataset, quarantine report, and audit manifest")
	fmt.Println("  studio    Launch interactive React 19 Quarantine Studio on local HTTP")
	fmt.Println("  cert      Cryptographic attestation and integrity verification")
	fmt.Println("  version   Print version information")
	fmt.Println("\nRun 'poisonprobe <command> -h' for command-specific flags.")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "scan":
		runScan(os.Args[2:])
	case "sanitize":
		runSanitize(os.Args[2:])
	case "studio":
		runStudio(os.Args[2:])
	case "cert":
		runCert(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("poisonprobe v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q. Run 'poisonprobe help' for usage.\n", command)
		os.Exit(1)
	}
}

func runScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	cfgPath := fs.String("config", "", "Path to YAML configuration file")
	sensitivity := fs.Float64("sensitivity", 3.0, "Anomaly threshold sensitivity multiplier (k)")
	dim := fs.Int("dim", 256, "Representation vector dimension")
	instrCol := fs.String("instruction-col", "", "Custom instruction column name")
	inputCol := fs.String("input-col", "", "Custom input column name")
	outputCol := fs.String("output-col", "", "Custom output column name")
	verbose := fs.Bool("verbose", false, "Verbose execution logs")

	fs.Parse(args)
	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Error: missing dataset file path.")
		fmt.Fprintln(os.Stderr, "Usage: poisonprobe scan <dataset-path> [flags]")
		os.Exit(1)
	}
	datasetPath := remaining[0]

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *sensitivity > 0 {
		cfg.Spectral.Sensitivity = *sensitivity
	}
	if *dim > 0 {
		cfg.Vectorizer.Dimension = *dim
	}
	if *instrCol != "" {
		cfg.Columns.Instruction = *instrCol
	}
	if *inputCol != "" {
		cfg.Columns.Input = *inputCol
	}
	if *outputCol != "" {
		cfg.Columns.Output = *outputCol
	}
	cfg.Verbose = *verbose

	fmt.Printf("\n[poisonprobe] Scanning dataset: %s\n", datasetPath)
	engine := audit.NewEngine(cfg)

	result, err := engine.Run(datasetPath, func(phase string, pct float64, msg string) {
		if cfg.Verbose {
			fmt.Printf("  [%3.0f%%] %s\n", pct, msg)
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nScan failed: %v\n", err)
		os.Exit(1)
	}

	// Print Formatted Terminal Results
	printAuditReport(result)

	if result.Summary.ContaminationRate > 0.05 {
		os.Exit(1)
	}
}

func runSanitize(args []string) {
	fs := flag.NewFlagSet("sanitize", flag.ExitOnError)
	cfgPath := fs.String("config", "", "Path to YAML configuration file")
	outputPath := fs.String("o", "clean.parquet", "Output path for certified clean dataset")
	quarantinePath := fs.String("q", "quarantine.json", "Output path for quarantined records report")
	attestationPath := fs.String("a", "audit.json", "Output path for cryptographic safety passport")
	format := fs.String("format", "", "Output format: 'parquet' or 'jsonl'")
	sensitivity := fs.Float64("sensitivity", 3.0, "Anomaly threshold sensitivity multiplier (k)")

	fs.Parse(args)
	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Error: missing dataset file path.")
		fmt.Fprintln(os.Stderr, "Usage: poisonprobe sanitize <dataset-path> -o <clean.parquet> -q <quarantine.json>")
		os.Exit(1)
	}
	datasetPath := remaining[0]

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	if *sensitivity > 0 {
		cfg.Spectral.Sensitivity = *sensitivity
	}

	fmt.Printf("\n[poisonprobe] Sanitizing dataset: %s\n", datasetPath)
	engine := audit.NewEngine(cfg)

	result, err := engine.Run(datasetPath, func(phase string, pct float64, msg string) {
		fmt.Printf("  [%3.0f%%] %s\n", pct, msg)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Sanitization failed: %v\n", err)
		os.Exit(1)
	}

	// Export Clean Records
	fmt.Printf("\nWriting certified clean dataset to %s...\n", *outputPath)
	cleanHash, err := sanitize.ExportClean(*outputPath, *format, result.Partition.Clean)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export clean dataset: %v\n", err)
		os.Exit(1)
	}

	// Export Quarantine Records
	fmt.Printf("Writing quarantine logs to %s...\n", *quarantinePath)
	if err := sanitize.ExportQuarantine(*quarantinePath, result.Partition.Quarantined, result.Summary.Timestamp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export quarantine: %v\n", err)
		os.Exit(1)
	}

	// Generate Attestation Manifest
	fmt.Printf("Generating cryptographic safety attestation to %s...\n", *attestationPath)
	att, err := sanitize.GenerateAttestation(result.Summary, *outputPath, cleanHash, *attestationPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate attestation: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + strings.Repeat("=", 68))
	fmt.Println(" SANITIZATION COMPLETED SUCCESSFULLY")
	fmt.Println(strings.Repeat("=", 68))
	fmt.Printf(" Input SHA-256:      %s\n", result.Summary.FileHashSHA256)
	fmt.Printf(" Clean SHA-256:      %s\n", cleanHash)
	fmt.Printf(" Total Evaluated:    %d\n", result.Summary.TotalSamples)
	fmt.Printf(" Certified Clean:    %d (%.1f%% yield)\n", len(result.Partition.Clean), (float64(len(result.Partition.Clean))/float64(result.Summary.TotalSamples))*100)
	fmt.Printf(" Quarantined:        %d (%.2f%% contamination)\n", len(result.Partition.Quarantined), result.Summary.ContaminationRate*100)
	fmt.Printf(" Safety Status:      %s\n", att.SafetyStatus)
	fmt.Printf(" Signature:          %s\n", att.DigestSignature)
	fmt.Println(strings.Repeat("=", 68))
}

func runStudio(args []string) {
	fs := flag.NewFlagSet("studio", flag.ExitOnError)
	cfgPath := fs.String("config", "", "Path to YAML configuration file")
	port := fs.Int("port", 7474, "HTTP port for Quarantine Studio")
	host := fs.String("host", "127.0.0.1", "HTTP bind address")
	noOpen := fs.Bool("no-open", false, "Do not automatically launch web browser")

	fs.Parse(args)
	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Error: missing dataset file path.")
		fmt.Fprintln(os.Stderr, "Usage: poisonprobe studio <dataset-path> [--port 7474]")
		os.Exit(1)
	}
	datasetPath := remaining[0]

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := server.StartStudio(cfg, datasetPath, *host, *port, !*noOpen); err != nil {
		fmt.Fprintf(os.Stderr, "Studio failed: %v\n", err)
		os.Exit(1)
	}
}

func runCert(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: poisonprobe cert verify <clean.parquet> <audit.json>")
		os.Exit(1)
	}

	subcmd := args[0]
	if subcmd != "verify" {
		fmt.Fprintf(os.Stderr, "Unknown cert action %q (expected 'verify')\n", subcmd)
		os.Exit(1)
	}

	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: poisonprobe cert verify <clean.parquet> <audit.json>")
		os.Exit(1)
	}

	cleanPath := args[1]
	auditPath := args[2]

	fmt.Printf("[poisonprobe] Verifying attestation: %s against %s...\n", auditPath, cleanPath)
	att, err := sanitize.VerifyAttestation(cleanPath, auditPath)
	if err != nil {
		fmt.Printf("\n[FAIL] ATTESTATION VERIFICATION FAILED: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + strings.Repeat("=", 68))
	fmt.Println(" [PASS] CRYPTOGRAPHIC ATTESTATION VERIFIED")
	fmt.Println(strings.Repeat("=", 68))
	fmt.Printf(" Status:            %s\n", att.SafetyStatus)
	fmt.Printf(" Clean File Hash:   %s\n", att.CleanSHA256)
	fmt.Printf(" Verified Records:  %d clean records\n", att.CleanSamples)
	fmt.Printf(" Contamination:     %.2f%%\n", att.ContaminationRatePct)
	fmt.Printf(" Signed Timestamp:  %s\n", att.Timestamp)
	fmt.Printf(" Attestation Sig:   %s\n", att.DigestSignature)
	fmt.Println(strings.Repeat("=", 68))
}

func printAuditReport(res *audit.AuditResult) {
	sum := res.Summary

	fmt.Println("\n" + strings.Repeat("=", 68))
	fmt.Println(" DATASET AUDIT & SPECTRAL SIGNATURE REPORT")
	fmt.Println(strings.Repeat("=", 68))
	fmt.Printf(" Dataset:            %s\n", sum.DatasetPath)
	fmt.Printf(" SHA-256:            %s\n", sum.FileHashSHA256)
	fmt.Printf(" Total Samples:      %d\n", sum.TotalSamples)
	fmt.Printf(" Certified Clean:    %d (%.1f%% yield)\n", sum.CleanCount, (float64(sum.CleanCount)/float64(sum.TotalSamples))*100)
	fmt.Printf(" Quarantined:        %d (%.2f%% contamination)\n", sum.QuarantineCount, sum.ContaminationRate*100)
	fmt.Printf(" Spectral Variance:  %.1f%% (v1 leading eigenvalue ratio)\n", sum.SpectralVariance*100)
	fmt.Printf(" MAD Outlier Cutoff: θ = %.4f (median: %.4f, MAD: %.4f)\n", sum.ThresholdCutoff, sum.MedianScore, sum.MADScore)
	fmt.Printf(" Scan Duration:      %d ms\n", sum.ExecutionTimeMs)

	if len(sum.TopTriggers) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 68))
		fmt.Println(" TOP MINED BACKDOOR TRIGGER CANDIDATES (NPMI)")
		fmt.Println(strings.Repeat("-", 68))
		fmt.Printf(" %-24s %-12s %-12s %-10s\n", "TOKEN", "CORRUPTED", "TOTAL", "NPMI")
		for _, tr := range sum.TopTriggers {
			covertTag := ""
			if tr.IsCovert {
				covertTag = " [COVERT]"
			}
			fmt.Printf(" %-24s %-12d %-12d %-10.3f\n", tr.Token+covertTag, tr.AnomalyCount, tr.Occurrence, tr.NPMI)
		}
	}

	fmt.Println(strings.Repeat("=", 68))
	if sum.ContaminationRate > 0.05 {
		fmt.Println(" [ALERT] Dataset contamination exceeds 5.0% threshold.")
		fmt.Println(" Recommendation: Run 'poisonprobe sanitize' to partition quarantined pool.")
	} else {
		fmt.Println(" [PASS] Dataset safety gate cleared (contamination <= 5.0%).")
	}
	fmt.Println(strings.Repeat("=", 68) + "\n")
}
