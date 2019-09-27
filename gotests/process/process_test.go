package process

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		opts    *Options
		want    string
		wantErr string
	}{
		// TODO: Add test cases.
		{
			name:    "Nil options and nil args",
			args:    nil,
			opts:    nil,
			wantErr: specifyFlagMessage,
		}, {
			name:    "Nil options",
			args:    []string{"testdata/foobar.go"},
			opts:    nil,
			wantErr: specifyFlagMessage,
		}, {
			name:    "Empty options",
			args:    []string{"testdata/foobar.go"},
			opts:    &Options{},
			wantErr: specifyFlagMessage,
		}, {
			name:    "Non-empty options with no args",
			args:    []string{},
			opts:    &Options{AllFuncs: true},
			wantErr: specifyFileMessage,
		}, {
			name:    "OnlyFuncs option w/ no matches",
			args:    []string{"testdata/foobar.go"},
			opts:    &Options{OnlyFuncs: "FooBar"},
			wantErr: "No tests generated for testdata/foobar.go\n",
		}, {
			name:    "Invalid OnlyFuncs option",
			args:    []string{"testdata/foobar.go"},
			opts:    &Options{OnlyFuncs: "??"},
			wantErr: "Invalid -only regex: error parsing regexp: missing argument to repetition operator: `??`\n",
		}, {
			name:    "Invalid ExclFuncs option",
			args:    []string{"testdata/foobar.go"},
			opts:    &Options{ExclFuncs: "??"},
			wantErr: "Invalid -excl regex: error parsing regexp: missing argument to repetition operator: `??`\n",
		},
	}
	for _, tt := range tests {
		out := &bytes.Buffer{}
		err := Run(out, tt.args, tt.opts)

		if tt.wantErr != "" {
			if err == nil {
				t.Errorf("%q.\ngot no error\nexpected error: %v", tt.name, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("%q.\ngot error: %v\nexpected to contain: %v", tt.name, err.Error(), tt.wantErr)
			}
			return
		}

		if err != nil {
			t.Errorf("%q.\nexpected no error but got error: %v", tt.name, err.Error())
		}

		if got := out.String(); got != tt.want {
			t.Errorf("%q.\ngot: %v\nwant: %v", tt.name, got, tt.want)
		}
	}
}
