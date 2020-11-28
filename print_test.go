package eclint_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"gitlab.com/greut/eclint"
)

func TestPrintErrors(t *testing.T) {
	tests := []struct {
		Name      string
		HasOutput bool
		Errors    []error
	}{
		{
			Name:      "no errors",
			HasOutput: false,
			Errors:    []error{},
		}, {
			Name:      "simple error",
			HasOutput: true,
			Errors: []error{
				errors.New("random error"),
			},
		}, {
			Name:      "validation error",
			HasOutput: true,
			Errors: []error{
				eclint.ValidationError{},
			},
		}, {
			Name:      "complete validation error",
			HasOutput: true,
			Errors: []error{
				eclint.ValidationError{
					Line:     []byte("Hello"),
					Index:    1,
					Position: 2,
				},
			},
		},
	}

	ctx := context.TODO()

	for _, tc := range tests {
		tc := tc

		// Test the nominal case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			opt := &eclint.Option{
				Stdout: buf,
			}
			err := eclint.PrintErrors(ctx, opt, tc.Name, tc.Errors)
			if err != nil {
				t.Error("no errors were expected")
			}
			outputLength := buf.Len()
			if (outputLength > 0) != tc.HasOutput {
				t.Errorf("unexpected output length got %d, wanted %v", outputLength, tc.HasOutput)
			}
		})
	}
}
