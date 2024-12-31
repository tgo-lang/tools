// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Don't include this file during code generation, or it will break the build
// if existing interface methods have been modified.
//go:build !generate
// +build !generate

// Code generated by gen.go from gopls/internal/protocol/command. DO NOT EDIT.

package command

import (
	"context"
	"fmt"

	"github.com/tgo-lang/tools/gopls/internal/protocol"
)

// Symbolic names for gopls commands, corresponding to methods of [Interface].
//
// The string value is used in the Command field of protocol.Command.
// These commands may be obtained from a CodeLens or CodeAction request
// and executed by an ExecuteCommand request.
const (
	AddDependency           Command = "gopls.add_dependency"
	AddImport               Command = "gopls.add_import"
	AddTelemetryCounters    Command = "gopls.add_telemetry_counters"
	AddTest                 Command = "gopls.add_test"
	ApplyFix                Command = "gopls.apply_fix"
	Assembly                Command = "gopls.assembly"
	ChangeSignature         Command = "gopls.change_signature"
	CheckUpgrades           Command = "gopls.check_upgrades"
	ClientOpenURL           Command = "gopls.client_open_url"
	DiagnoseFiles           Command = "gopls.diagnose_files"
	Doc                     Command = "gopls.doc"
	EditGoDirective         Command = "gopls.edit_go_directive"
	ExtractToNewFile        Command = "gopls.extract_to_new_file"
	FetchVulncheckResult    Command = "gopls.fetch_vulncheck_result"
	FreeSymbols             Command = "gopls.free_symbols"
	GCDetails               Command = "gopls.gc_details"
	Generate                Command = "gopls.generate"
	GoGetPackage            Command = "gopls.go_get_package"
	ListImports             Command = "gopls.list_imports"
	ListKnownPackages       Command = "gopls.list_known_packages"
	MaybePromptForTelemetry Command = "gopls.maybe_prompt_for_telemetry"
	MemStats                Command = "gopls.mem_stats"
	Modules                 Command = "gopls.modules"
	Packages                Command = "gopls.packages"
	RegenerateCgo           Command = "gopls.regenerate_cgo"
	RemoveDependency        Command = "gopls.remove_dependency"
	ResetGoModDiagnostics   Command = "gopls.reset_go_mod_diagnostics"
	RunGoWorkCommand        Command = "gopls.run_go_work_command"
	RunGovulncheck          Command = "gopls.run_govulncheck"
	RunTests                Command = "gopls.run_tests"
	ScanImports             Command = "gopls.scan_imports"
	StartDebugging          Command = "gopls.start_debugging"
	StartProfile            Command = "gopls.start_profile"
	StopProfile             Command = "gopls.stop_profile"
	Test                    Command = "gopls.test"
	Tidy                    Command = "gopls.tidy"
	ToggleGCDetails         Command = "gopls.toggle_gc_details"
	UpdateGoSum             Command = "gopls.update_go_sum"
	UpgradeDependency       Command = "gopls.upgrade_dependency"
	Vendor                  Command = "gopls.vendor"
	Views                   Command = "gopls.views"
	Vulncheck               Command = "gopls.vulncheck"
	WorkspaceStats          Command = "gopls.workspace_stats"
)

var Commands = []Command{
	AddDependency,
	AddImport,
	AddTelemetryCounters,
	AddTest,
	ApplyFix,
	Assembly,
	ChangeSignature,
	CheckUpgrades,
	ClientOpenURL,
	DiagnoseFiles,
	Doc,
	EditGoDirective,
	ExtractToNewFile,
	FetchVulncheckResult,
	FreeSymbols,
	GCDetails,
	Generate,
	GoGetPackage,
	ListImports,
	ListKnownPackages,
	MaybePromptForTelemetry,
	MemStats,
	Modules,
	Packages,
	RegenerateCgo,
	RemoveDependency,
	ResetGoModDiagnostics,
	RunGoWorkCommand,
	RunGovulncheck,
	RunTests,
	ScanImports,
	StartDebugging,
	StartProfile,
	StopProfile,
	Test,
	Tidy,
	ToggleGCDetails,
	UpdateGoSum,
	UpgradeDependency,
	Vendor,
	Views,
	Vulncheck,
	WorkspaceStats,
}

func Dispatch(ctx context.Context, params *protocol.ExecuteCommandParams, s Interface) (interface{}, error) {
	switch Command(params.Command) {
	case AddDependency:
		var a0 DependencyArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.AddDependency(ctx, a0)
	case AddImport:
		var a0 AddImportArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.AddImport(ctx, a0)
	case AddTelemetryCounters:
		var a0 AddTelemetryCountersArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.AddTelemetryCounters(ctx, a0)
	case AddTest:
		var a0 protocol.Location
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.AddTest(ctx, a0)
	case ApplyFix:
		var a0 ApplyFixArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.ApplyFix(ctx, a0)
	case Assembly:
		var a0 string
		var a1 string
		var a2 string
		if err := UnmarshalArgs(params.Arguments, &a0, &a1, &a2); err != nil {
			return nil, err
		}
		return nil, s.Assembly(ctx, a0, a1, a2)
	case ChangeSignature:
		var a0 ChangeSignatureArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.ChangeSignature(ctx, a0)
	case CheckUpgrades:
		var a0 CheckUpgradesArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.CheckUpgrades(ctx, a0)
	case ClientOpenURL:
		var a0 string
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.ClientOpenURL(ctx, a0)
	case DiagnoseFiles:
		var a0 DiagnoseFilesArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.DiagnoseFiles(ctx, a0)
	case Doc:
		var a0 DocArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.Doc(ctx, a0)
	case EditGoDirective:
		var a0 EditGoDirectiveArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.EditGoDirective(ctx, a0)
	case ExtractToNewFile:
		var a0 protocol.Location
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.ExtractToNewFile(ctx, a0)
	case FetchVulncheckResult:
		var a0 URIArg
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.FetchVulncheckResult(ctx, a0)
	case FreeSymbols:
		var a0 string
		var a1 protocol.Location
		if err := UnmarshalArgs(params.Arguments, &a0, &a1); err != nil {
			return nil, err
		}
		return nil, s.FreeSymbols(ctx, a0, a1)
	case GCDetails:
		var a0 protocol.DocumentURI
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.GCDetails(ctx, a0)
	case Generate:
		var a0 GenerateArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.Generate(ctx, a0)
	case GoGetPackage:
		var a0 GoGetPackageArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.GoGetPackage(ctx, a0)
	case ListImports:
		var a0 URIArg
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.ListImports(ctx, a0)
	case ListKnownPackages:
		var a0 URIArg
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.ListKnownPackages(ctx, a0)
	case MaybePromptForTelemetry:
		return nil, s.MaybePromptForTelemetry(ctx)
	case MemStats:
		return s.MemStats(ctx)
	case Modules:
		var a0 ModulesArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.Modules(ctx, a0)
	case Packages:
		var a0 PackagesArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.Packages(ctx, a0)
	case RegenerateCgo:
		var a0 URIArg
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.RegenerateCgo(ctx, a0)
	case RemoveDependency:
		var a0 RemoveDependencyArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.RemoveDependency(ctx, a0)
	case ResetGoModDiagnostics:
		var a0 ResetGoModDiagnosticsArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.ResetGoModDiagnostics(ctx, a0)
	case RunGoWorkCommand:
		var a0 RunGoWorkArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.RunGoWorkCommand(ctx, a0)
	case RunGovulncheck:
		var a0 VulncheckArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.RunGovulncheck(ctx, a0)
	case RunTests:
		var a0 RunTestsArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.RunTests(ctx, a0)
	case ScanImports:
		return nil, s.ScanImports(ctx)
	case StartDebugging:
		var a0 DebuggingArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.StartDebugging(ctx, a0)
	case StartProfile:
		var a0 StartProfileArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.StartProfile(ctx, a0)
	case StopProfile:
		var a0 StopProfileArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.StopProfile(ctx, a0)
	case Test:
		var a0 protocol.DocumentURI
		var a1 []string
		var a2 []string
		if err := UnmarshalArgs(params.Arguments, &a0, &a1, &a2); err != nil {
			return nil, err
		}
		return nil, s.Test(ctx, a0, a1, a2)
	case Tidy:
		var a0 URIArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.Tidy(ctx, a0)
	case ToggleGCDetails:
		var a0 URIArg
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.ToggleGCDetails(ctx, a0)
	case UpdateGoSum:
		var a0 URIArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.UpdateGoSum(ctx, a0)
	case UpgradeDependency:
		var a0 DependencyArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.UpgradeDependency(ctx, a0)
	case Vendor:
		var a0 URIArg
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return nil, s.Vendor(ctx, a0)
	case Views:
		return s.Views(ctx)
	case Vulncheck:
		var a0 VulncheckArgs
		if err := UnmarshalArgs(params.Arguments, &a0); err != nil {
			return nil, err
		}
		return s.Vulncheck(ctx, a0)
	case WorkspaceStats:
		return s.WorkspaceStats(ctx)
	}
	return nil, fmt.Errorf("unsupported command %q", params.Command)
}

func NewAddDependencyCommand(title string, a0 DependencyArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   AddDependency.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewAddImportCommand(title string, a0 AddImportArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   AddImport.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewAddTelemetryCountersCommand(title string, a0 AddTelemetryCountersArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   AddTelemetryCounters.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewAddTestCommand(title string, a0 protocol.Location) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   AddTest.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewApplyFixCommand(title string, a0 ApplyFixArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ApplyFix.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewAssemblyCommand(title string, a0 string, a1 string, a2 string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Assembly.String(),
		Arguments: MustMarshalArgs(a0, a1, a2),
	}
}

func NewChangeSignatureCommand(title string, a0 ChangeSignatureArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ChangeSignature.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewCheckUpgradesCommand(title string, a0 CheckUpgradesArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   CheckUpgrades.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewClientOpenURLCommand(title string, a0 string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ClientOpenURL.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewDiagnoseFilesCommand(title string, a0 DiagnoseFilesArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   DiagnoseFiles.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewDocCommand(title string, a0 DocArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Doc.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewEditGoDirectiveCommand(title string, a0 EditGoDirectiveArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   EditGoDirective.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewExtractToNewFileCommand(title string, a0 protocol.Location) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ExtractToNewFile.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewFetchVulncheckResultCommand(title string, a0 URIArg) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   FetchVulncheckResult.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewFreeSymbolsCommand(title string, a0 string, a1 protocol.Location) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   FreeSymbols.String(),
		Arguments: MustMarshalArgs(a0, a1),
	}
}

func NewGCDetailsCommand(title string, a0 protocol.DocumentURI) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   GCDetails.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewGenerateCommand(title string, a0 GenerateArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Generate.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewGoGetPackageCommand(title string, a0 GoGetPackageArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   GoGetPackage.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewListImportsCommand(title string, a0 URIArg) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ListImports.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewListKnownPackagesCommand(title string, a0 URIArg) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ListKnownPackages.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewMaybePromptForTelemetryCommand(title string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   MaybePromptForTelemetry.String(),
		Arguments: MustMarshalArgs(),
	}
}

func NewMemStatsCommand(title string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   MemStats.String(),
		Arguments: MustMarshalArgs(),
	}
}

func NewModulesCommand(title string, a0 ModulesArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Modules.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewPackagesCommand(title string, a0 PackagesArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Packages.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewRegenerateCgoCommand(title string, a0 URIArg) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   RegenerateCgo.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewRemoveDependencyCommand(title string, a0 RemoveDependencyArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   RemoveDependency.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewResetGoModDiagnosticsCommand(title string, a0 ResetGoModDiagnosticsArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ResetGoModDiagnostics.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewRunGoWorkCommandCommand(title string, a0 RunGoWorkArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   RunGoWorkCommand.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewRunGovulncheckCommand(title string, a0 VulncheckArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   RunGovulncheck.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewRunTestsCommand(title string, a0 RunTestsArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   RunTests.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewScanImportsCommand(title string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ScanImports.String(),
		Arguments: MustMarshalArgs(),
	}
}

func NewStartDebuggingCommand(title string, a0 DebuggingArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   StartDebugging.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewStartProfileCommand(title string, a0 StartProfileArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   StartProfile.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewStopProfileCommand(title string, a0 StopProfileArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   StopProfile.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewTestCommand(title string, a0 protocol.DocumentURI, a1 []string, a2 []string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Test.String(),
		Arguments: MustMarshalArgs(a0, a1, a2),
	}
}

func NewTidyCommand(title string, a0 URIArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Tidy.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewToggleGCDetailsCommand(title string, a0 URIArg) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   ToggleGCDetails.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewUpdateGoSumCommand(title string, a0 URIArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   UpdateGoSum.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewUpgradeDependencyCommand(title string, a0 DependencyArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   UpgradeDependency.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewVendorCommand(title string, a0 URIArg) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Vendor.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewViewsCommand(title string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Views.String(),
		Arguments: MustMarshalArgs(),
	}
}

func NewVulncheckCommand(title string, a0 VulncheckArgs) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   Vulncheck.String(),
		Arguments: MustMarshalArgs(a0),
	}
}

func NewWorkspaceStatsCommand(title string) *protocol.Command {
	return &protocol.Command{
		Title:     title,
		Command:   WorkspaceStats.String(),
		Arguments: MustMarshalArgs(),
	}
}
