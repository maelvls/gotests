// Package process is a thin wrapper around the gotests library. It is intended
// to be called from a binary and handle its arguments, flags, and output when
// generating tests.
package process

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/maelvls/gotests"
)

const newFilePerm os.FileMode = 0644

const (
	specifyFlagMessage = "Please specify either the -only, -excl, -exported, or -all flag"
	specifyFileMessage = "Please specify a file or directory containing the source"
)

// Set of options to use when generating tests.
type Options struct {
	OnlyFuncs          string   // Regexp string for filter matches.
	ExclFuncs          string   // Regexp string for excluding matches.
	ExportedFuncs      bool     // Only include exported functions.
	AllFuncs           bool     // Include all non-tested functions.
	PrintInputs        bool     // Print function parameters as part of error messages.
	Subtests           bool     // Print tests using Go 1.7 subtests
	Parallel           bool     // Print tests that runs the subtests in parallel.
	WriteOutput        bool     // Write output to test file(s).
	Template           string   // Name of custom template set
	TemplateDir        string   // Path to custom template set
	TemplateParamsPath string   // Path to custom parameters json file(s).
	TemplateData       [][]byte // Data slice for templates
}

// Generates tests for the Go files defined in args with the given options.
// Logs information and errors to out. By default outputs generated tests to
// out unless specified by opt.
func Run(out io.Writer, args []string, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	opt, err := parseOptions(opts)
	if err != nil {
		return fmt.Errorf("parsing flags: %v", err)
	}

	if len(args) == 0 {
		return fmt.Errorf(specifyFileMessage)
	}
	for _, path := range args {
		err = generateTests(out, path, opts.WriteOutput, opt)
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
	}
	return nil
}

func parseOptions(opt *Options) (*gotests.Options, error) {
	if opt.OnlyFuncs == "" && opt.ExclFuncs == "" && !opt.ExportedFuncs && !opt.AllFuncs {
		return nil, fmt.Errorf(specifyFlagMessage)
	}
	onlyRE, err := parseRegexp(opt.OnlyFuncs)
	if err != nil {
		return nil, fmt.Errorf("Invalid -only regex: %v", err)
	}
	exclRE, err := parseRegexp(opt.ExclFuncs)
	if err != nil {
		return nil, fmt.Errorf("Invalid -excl regex: %v", err)
	}

	templateParams := map[string]interface{}{}
	jfile := opt.TemplateParamsPath
	if jfile != "" {
		buf, err := ioutil.ReadFile(jfile)
		if err != nil {
			return nil, fmt.Errorf("Failed to read from %s ,err %s", jfile, err)
		}

		err = json.Unmarshal(buf, &templateParams)
		if err != nil {
			return nil, fmt.Errorf("Failed to umarshal %s er %s", jfile, err)
		}
	}

	return &gotests.Options{
		Only:           onlyRE,
		Exclude:        exclRE,
		Exported:       opt.ExportedFuncs,
		PrintInputs:    opt.PrintInputs,
		Subtests:       opt.Subtests,
		Parallel:       opt.Parallel,
		Template:       opt.Template,
		TemplateDir:    opt.TemplateDir,
		TemplateParams: templateParams,
		TemplateData:   opt.TemplateData,
	}, nil
}

func parseRegexp(s string) (*regexp.Regexp, error) {
	if s == "" {
		return nil, nil
	}
	re, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	return re, nil
}

func generateTests(out io.Writer, path string, writeOutput bool, opt *gotests.Options) error {
	gts, err := gotests.GenerateTests(path, opt)
	if err != nil {
		return fmt.Errorf("generating tests: %v", err)
	}
	if len(gts) == 0 {
		return fmt.Errorf("no tests generated")
	}
	for _, t := range gts {
		err := outputTest(out, t, writeOutput)
		if err != nil {
			return err
		}
	}
	return nil
}

func outputTest(out io.Writer, t *gotests.GeneratedTest, writeOutput bool) error {
	if writeOutput {
		err := ioutil.WriteFile(t.Path, t.Output, newFilePerm)
		if err != nil {
			return fmt.Errorf("%s (-w): %v", t.Path, err)
		}
	}
	for _, t := range t.Functions {
		fmt.Fprintln(os.Stderr, "Generated", t.TestName())
	}
	if !writeOutput {
		_, err := out.Write(t.Output)
		if err != nil {
			return err
		}
	}
	return nil
}
