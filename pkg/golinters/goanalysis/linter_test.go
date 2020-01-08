package goanalysis

import (
	"fmt"
	"github.com/stretchr/testify/mock"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildIssues(tt.args.diags, tt.args.linterNameBuilder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockedIDiagnostic struct {
	mock.Mock
}

func (m *MockedIDiagnostic) fields() *Diagnostic {
	args := m.Called()
	return args.Get(0).(*Diagnostic)
}

func (m *MockedIDiagnostic) getPositionOf(p token.Pos) token.Position {
	args := m.Called(p)
	return args.Get(0).(token.Position)

}

func posEquals(expected token.Pos) interface{} {
	return mock.MatchedBy(func(pos token.Pos) bool {
		return pos == expected
	})
}

func Test_getIssuesForDiagnostic(t *testing.T) {
	type args struct {
		diag       iDiagnostic
		linterName string
	}
	tests := []struct {
		name       string
		args       args
		wantIssues []result.Issue

		prepare func(m *MockedIDiagnostic)
	}{
		{
			name: "Linter Name is Analyzer Name",
			args: args{
				diag: &Diagnostic{
					Diagnostic: analysis.Diagnostic{
						Message: "failure message",
					},
					Analyzer: &analysis.Analyzer{
						Name: "some-linter",
					},
					Position: token.Position{},
					Pkg:      nil,
				},

				linterName: "some-linter",
			},
			wantIssues: []result.Issue{
				{
					FromLinter: "some-linter",
					Text:       "failure message",
				},
			},
		},
		{
			name: "Linter Name is NOT Analyzer Name",
			args: args{
				diag: &Diagnostic{
					Diagnostic: analysis.Diagnostic{
						Message: "failure message",
					},
					Analyzer: &analysis.Analyzer{
						Name: "some-analyzer",
					},
					Position: token.Position{},
					Pkg:      nil,
				},
				linterName: "some-linter",
			},
			wantIssues: []result.Issue{
				{
					FromLinter: "some-linter",
					Text:       "some-analyzer: failure message",
				},
			},
		},
		{
			name: "Replace Whole Line",
			args: args{
				diag:       &MockedIDiagnostic{},
				linterName: "some-linter",
			},
			wantIssues: []result.Issue{
				{
					FromLinter: "some-linter",
					Text:       "some-analyzer: failure message",
					LineRange: &result.Range{
						From: 5,
						To:   5,
					},
					Replacement: &result.Replacement{
						NeedOnlyDelete: false,
						NewLines: []string{
							"// Some comment to fix",
						},
					},
				},
			},
			prepare: func(m *MockedIDiagnostic) {
				d := &Diagnostic{
					Diagnostic: analysis.Diagnostic{
						Message: "failure message",
						SuggestedFixes: []analysis.SuggestedFix{
							{
								Message: "fix something",
								TextEdits: []analysis.TextEdit{
									{
										Pos:     101,
										End:     201,
										NewText: []byte("// Some comment to fix\n"),
									},
								},
							},
						},
					},
					Analyzer: &analysis.Analyzer{
						Name: "some-analyzer",
					},
					Position: token.Position{},
					Pkg:      nil,
				}
				m.On("fields").Return(d)
				m.On("getPositionOf", posEquals(101)).Return(token.Position{Line: 5, Column: 1})
				m.On("getPositionOf", posEquals(201)).Return(token.Position{Line: 6, Column: 1})
			},
		},
		// TODO: TDD 1 suggested fix, no text edits!
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if using a mock diagnostic, prepare it
			if diagnostic, ok := tt.args.diag.(*MockedIDiagnostic); ok {
				tt.prepare(diagnostic)
			}

			if gotIssues := getIssuesForDiagnostic(tt.args.diag, tt.args.linterName); !reflect.DeepEqual(gotIssues, tt.wantIssues) {
				t.Errorf("getIssuesForDiagnostic() = %v, want %v", gotIssues, tt.wantIssues)
			}
		})
	}
}
