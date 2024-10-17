package yamler

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testNames struct {
	First string `yaml:"first"`
	Last  string `yaml:"last"`
}

type testPerson struct {
	Age   int       `yaml:"age"`
	Names testNames `yaml:"names"`
}

func TestMarshal(t *testing.T) {
	given := testPerson{9000, testNames{"Jane", "Doe"}}

	out, err := Marshal(given)
	if err != nil {
		t.Fatal(err)
	}

	got := string(out)
	want := `age: 9000
names:
  first: Jane
  last: Doe
`

	assert.Equal(t, want, got)
}

func TestNewMarshaler(t *testing.T) {
	n := MinIndent

	for n <= MaxIndent {
		t.Run(fmt.Sprintf("it should build a new marshaler with indentCols %d", n), func(t *testing.T) {
			sut := NewMarshaler(n)
			if sut.indentCols != n {
				t.Errorf("want %d, got %d", n, sut.indentCols)
			}
		})
		n++
	}
}

func TestMarshaler_Marshal(t *testing.T) {
	// arrange
	given := testPerson{9000, testNames{"Jane", "Doe"}}
	wantIndent := 2
	sut := NewMarshaler(wantIndent)

	// act
	got, err := sut.Marshal(given)
	if err != nil {
		t.Fatal(err)
	}

	wantPad := strings.Repeat(" ", wantIndent)
	wantLines := []string{
		"age: 9000",
		"names:",
		wantPad + "first: Jane",
		wantPad + "last: Doe\n",
	}
	wantStr := strings.Join(wantLines, "\n")

	gotStr := string(got)

	slog.Info("done acting", "wantStr", wantStr, "gotStr", gotStr, "got", got)
	// assert
	t.Run(fmt.Sprintf("it should indent by exactly %d spaces per level", wantIndent), func(t *testing.T) {
		if wantStr != gotStr {
			t.Errorf("want %s, got %s\twantIndent=%d got=%s)", wantStr, gotStr, wantIndent, got)
		}
	})
}

func TestMarshalToWriter(t *testing.T) {
	// arrange
	given := testPerson{9000, testNames{"Jane", "Doe"}}

	wantIndent := 2
	sut := NewMarshaler(wantIndent)

	var gotBytes bytes.Buffer
	w := bufio.NewWriter(&gotBytes)

	// act
	err := sut.MarshalToWriter(given, w)
	if err != nil {
		t.Fatal(err)
	}

	w.Flush()

	// assert
	t.Run("it should have written YAML to gotBytes", func(t *testing.T) {
		wantPad := strings.Repeat(" ", wantIndent)
		wantLines := []string{
			"age: 9000",
			"names:",
			wantPad + "first: Jane",
			wantPad + "last: Doe\n",
		}
		wantStr := strings.Join(wantLines, "\n")
		gotStr := gotBytes.String()

		if gotStr != wantStr {
			t.Errorf("\n got:\n%s\nwant:\n%s", wantStr, gotStr)
		}
	})
}

func TestUnmarshal(t *testing.T) {
	given := []byte(`age: 9000
names:
  first: Jane
  last: Doe
`)

	var got testPerson
	err := Unmarshal(given, &got)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testNames{"Jane", "Doe"}, got.Names)
}
