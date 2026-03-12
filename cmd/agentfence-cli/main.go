package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/agentfence/agentfence/internal/approval"
	"github.com/agentfence/agentfence/internal/config"
	storagepg "github.com/agentfence/agentfence/internal/storage/postgres"
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
	case "list-approvals":
		return listApprovals(args[1:])
	case "approve":
		return resolveApproval(args[1:], true)
	case "deny":
		return resolveApproval(args[1:], false)
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

func listApprovals(args []string) error {
	service, cleanup, err := approvalServiceFromFlags("list-approvals", args)
	if err != nil {
		return err
	}
	defer cleanup()
	requests, err := service.ListPending(context.Background())
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(requests)
}

func resolveApproval(args []string, approve bool) error {
	var actor string
	var reason string
	var storePath string
	name := "deny"
	verb := "denied"
	if approve {
		name = "approve"
		verb = "approved"
	}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&storePath, "store", approvalStorePath(), "path to the approval store JSON file")
	flags.StringVar(&actor, "actor", "", "actor approving or denying the request")
	flags.StringVar(&reason, "reason", "", "resolution reason")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("approval id is required")
	}
	service, cleanup, err := approvalService(storePath)
	if err != nil {
		return err
	}
	defer cleanup()
	input := approval.ResolveInput{ID: flags.Arg(0), Actor: actor, Reason: reason}
	var request approval.Request
	if approve {
		request, err = service.Approve(context.Background(), input)
	} else {
		request, err = service.Deny(context.Background(), input)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s %s as %s\n", request.ID, verb, request.Status)
	return nil
}

func approvalServiceFromFlags(name string, args []string) (*approval.Service, func(), error) {
	var storePath string
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&storePath, "store", approvalStorePath(), "path to the approval store JSON file")
	if err := flags.Parse(args); err != nil {
		return nil, func() {}, err
	}
	return approvalService(storePath)
}

func approvalService(storePath string) (*approval.Service, func(), error) {
	if dsn := os.Getenv("AGENTFENCE_POSTGRES_DSN"); dsn != "" {
		pgdb, err := storagepg.Open(context.Background(), dsn)
		if err != nil {
			return nil, func() {}, err
		}
		if err := pgdb.Migrate(context.Background()); err != nil {
			_ = pgdb.Close()
			return nil, func() {}, err
		}
		return approval.NewService(storagepg.NewApprovalRepository(pgdb.SQL)), func() { _ = pgdb.Close() }, nil
	}
	return approval.NewService(approval.NewFileRepository(storePath)), func() {}, nil
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
	fmt.Fprintln(out, "  list-approvals    list pending approvals")
	fmt.Fprintln(out, "  approve           approve an approval request by id")
	fmt.Fprintln(out, "  deny              deny an approval request by id")
}

func approvalStorePath() string {
	if path := os.Getenv("AGENTFENCE_APPROVAL_STORE"); path != "" {
		return path
	}
	return "data/approvals.json"
}
