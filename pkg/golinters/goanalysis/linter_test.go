package goanalysis

import (
	"fmt"
	"go/token"
	"reflect"
	"testing"

	"github.com/golangci/golangci-lint/pkg/result"
	"golang.org/x/tools/go/analysis"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestParseError(t *testing.T) {
	cases := []struct {
		in, out string
		good    bool
	}{
		{"f.go:1:2", "", true},
		{"f.go:1", "", true},
		{"f.go", "", false},
		{"f.go: 1", "", false},
	}

	for _, c := range cases {
		i, _ := parseError(packages.Error{
			Pos: c.in,
			Msg: "msg",
		})
		if !c.good {
			assert.Nil(t, i)
			continue
		}

		assert.NotNil(t, i)

		pos := fmt.Sprintf("%s:%d", i.FilePath(), i.Line())
		if i.Pos.Column != 0 {
			pos += fmt.Sprintf(":%d", i.Pos.Column)
		}
		out := pos
		expOut := c.out
		if expOut == "" {
			expOut = c.in
		}
		assert.Equal(t, expOut, out)

		assert.Equal(t, "typecheck", i.FromLinter)
		assert.Equal(t, "msg", i.Text)
	}
}

func Test_buildIssues(t *testing.T) {
	type args struct {
		diags             []Diagnostic
		linterNameBuilder func(diag *Diagnostic) string
	}
	tests := []struct {
		name string
		args args
		want []result.Issue
	}{
		{
			name: "No Diagnostics",
			args: args{
				diags: []Diagnostic{},
				linterNameBuilder: func(*Diagnostic) string {
					return "some-linter"
				},
			},
			want: []result.Issue(nil),
		},
		{
			name: "Linter Name is Analyzer Name",
			args: args{
				diags: []Diagnostic{
					{
						Diagnostic: analysis.Diagnostic{
							Message: "failure message",
						},
						Analyzer: &analysis.Analyzer{
							Name: "some-linter",
						},
						Position: token.Position{},
						Pkg:      nil,
					},
				},
				linterNameBuilder: func(*Diagnostic) string {
					return "some-linter"
				},
			},
			want: []result.Issue{
				{
					FromLinter: "some-linter",
					Text:       "failure message",
				},
			},
		},
		{
			name: "Linter Name is NOT Analyzer Name",
			args: args{
				diags: []Diagnostic{
					{
						Diagnostic: analysis.Diagnostic{
							Message: "failure message",
						},
						Analyzer: &analysis.Analyzer{
							Name: "some-analyzer",
						},
						Position: token.Position{},
						Pkg:      nil,
					},
				},
				linterNameBuilder: func(*Diagnostic) string {
					return "some-linter"
				},
			},
			want: []result.Issue{
				{
					FromLinter: "some-linter",
					Text:       "some-analyzer: failure message",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildIssues(tt.args.diags, tt.args.linterNameBuilder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}
