package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/agentfence/agentfence/internal/config"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage(os.Stderr)
		return nil
	}

	switch args[0] {
	case "validate-config":
		return validateConfig(args[1:])
	case "print-config":
		return printConfig(args[1:])
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func validateConfig(args []string) error {
	cfg, err := loadConfigFromArgs("validate-config", args)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "config valid for %s on %s\n", cfg.Environment, cfg.HTTP.Address)
	return nil
}

func printConfig(args []string) error {
	cfg, err := loadConfigFromArgs("print-config", args)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func loadConfigFromArgs(name string, args []string) (config.Config, error) {
	var configPath string

	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&configPath, "config", "", "path to a JSON config file")

	if err := flags.Parse(args); err != nil {
		return config.Config{}, err
	}

	return config.Load(configPath)
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "usage: agentfence-cli <command> [flags]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "commands:")
	fmt.Fprintln(out, "  validate-config   load config from file/env and validate it")
	fmt.Fprintln(out, "  print-config      print the effective config as JSON")
}
