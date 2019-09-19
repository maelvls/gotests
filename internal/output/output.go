package output

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/tools/imports"

	"github.com/maelvls/gotests/internal/models"
	"github.com/maelvls/gotests/internal/render"
)

type Options struct {
	PrintInputs    bool
	Subtests       bool
	Parallel       bool
	Template       string
	TemplateDir    string
	TemplateParams map[string]interface{}
	TemplateData   [][]byte
}

func Process(head *models.Header, funcs []*models.Function, opt *Options) ([]byte, error) {
	if opt != nil && opt.TemplateDir != "" {
		err := render.LoadCustomTemplates(opt.TemplateDir)
		if err != nil {
			return nil, fmt.Errorf("loading custom templates: %v", err)
		}
	} else if opt != nil && opt.Template != "" {
		err := render.LoadCustomTemplatesName(opt.Template)
		if err != nil {
			return nil, fmt.Errorf("loading custom templates of name: %v", err)
		}
	} else if opt != nil && opt.TemplateData != nil {
		render.LoadFromData(opt.TemplateData)
	}

	tf, err := ioutil.TempFile("", "gotests_")
	if err != nil {
		return nil, fmt.Errorf("ioutil.TempFile: %v", err)
	}
	defer tf.Close()
	defer os.Remove(tf.Name())
	b := &bytes.Buffer{}
	if err := writeTests(b, head, funcs, opt); err != nil {
		return nil, err
	}

	out, err := imports.Process(tf.Name(), b.Bytes(), nil)
	if err != nil {
		if os.Getenv("DEBUG_GENERATED") != "" {
			fmt.Println(string(b.Bytes()))
		}
		return nil, fmt.Errorf("imports.Process: %v\nYou can set DEBUG_GENERATED=1 to print to stdout the intermediate generated code before it is gofmt'd.", err)
	}
	return out, nil
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func writeTests(w io.Writer, head *models.Header, funcs []*models.Function, opt *Options) error {
	b := bufio.NewWriter(w)
	if err := render.Header(b, head); err != nil {
		return fmt.Errorf("render.Header: %v", err)
	}
	for _, fun := range funcs {
		if err := render.TestFunction(b, fun, opt.PrintInputs, opt.Subtests, opt.Parallel, opt.TemplateParams); err != nil {
			return fmt.Errorf("render.TestFunction: %v", err)
		}
	}
	return b.Flush()
}
