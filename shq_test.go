package shq

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os/exec"
	"testing"
)

func (s NegSample) Test(t *testing.T) {
	v := fmt.Sprintf("%s", Arg(s.Sample))
	if v == s.Expected {
		t.Errorf("returned string matched when failure expected: %s == %s (but shouldn't)", v, s.Expected)
	}
	t.Log("Not testing in shell for negative sample")
}
func (s PosSample) Test(t *testing.T) {
	arg := Arg(s.Sample)
	v := fmt.Sprintf("%s", arg)
	g := fmt.Sprintf("%#v", arg)

	if v != s.Expected {
		t.Errorf("returned string didn't match expected result: %s != %s", v, s.Expected)
		t.Log("Skipping shell test")
	} else {
		t.Log("Passed initial test -- testing in shell")
		out, err := exec.Command("sh", "-c", "printf %s "+v).Output()
		if err != nil {
			t.Errorf("Failed to run test shell: %v", err)
		}
		if string(out) != s.Sample {
			t.Errorf("test shell did not produce expected result: %s (actual) != %s (expected)", out, s.Sample)
		} else {
			t.Logf("shell test passed: %s = %s", string(out), s.Sample)
		}

		unesc := arg.Unescaped()
		if string(out) != unesc {
			t.Errorf("Arg(%s).Unescaped() = %#v doesn't match shell output (%#v)", s.Sample, unesc, string(out))
		}
	}

	gexp := fmt.Sprintf("Arg(%#v -> %#v)", s.Sample, s.Expected)
	if g != gexp {
		t.Errorf("GoStringer mismatch: %s (actual) != %s (expected)", g, gexp)
	}
}

type BaseSample struct {
	Sample   string
	Expected string
}
type PosSample BaseSample
type NegSample BaseSample

type Sample interface {
	Test(t *testing.T)
	GoString() string
}

func (s NegSample) GoString() string {
	return fmt.Sprintf("Test(Arg(%s)!=%s)", s.Sample, s.Expected)
}
func (s PosSample) GoString() string {
	return fmt.Sprintf("Test(Arg(%s)==%s)", s.Sample, s.Expected)
}

// test standard strings as we might expect to use them
func TestStrings(t *testing.T) {
	tests := []Sample{
		PosSample{`'`, `''"'"''`},
		PosSample{`💩`, `'💩'`},
		PosSample{`'''`, `''"'"''"'"''"'"''`},
		PosSample{"\n \n  ", "'\n \n  '"},
		PosSample{`""'`, `'""'"'"''`},
		NegSample{`'x'`, `'x'`},
		PosSample{`'x'`, `''"'"'x'"'"''`},
		PosSample{`$P'ATH`, `'$P'"'"'ATH'`},
		NegSample{"cafe\u0301", "'caf\u00e9'"},
		NegSample{"cafe\u00e9", "'caf\u0301'"},
		PosSample{"cafe\u0301", "'cafe\u0301'"},
		PosSample{"caf\u00e9", "'caf\u00e9'"},
		PosSample{"caf\ufffe", "'caf\ufffe'"},
		PosSample{"caf\xe2\x28\xa1e", "'caf\xe2\x28\xa1e'"},
	}
	for i := 0; i < len(tests); i++ {
		i := i // closure need to close over i's value, not i itself.  make a copy
		t.Run(fmt.Sprintf("%#v", tests[i]), func(t *testing.T) {
			t.Parallel()

			tests[i].Test(t)
		})
	}
}

func TestNull(t *testing.T) {
	s := "$P\000TH"
	exp := `'$P'`
	shexp := `$P`
	arg := Arg(s)
	if arg.Valid() {
		t.Errorf("argument contains a null but is flagged as valid anyway")
	}
	farg := fmt.Sprintf("%s", arg)
	t.Logf("Testing null with %#v", s)
	if farg != exp {
		t.Errorf("output with null did not match expected: %#v (actual) != %#v (expected)", farg, exp)
	}
	out, err := exec.Command("sh", "-c", "printf %s "+arg.String()).Output()
	if err != nil {
		t.Errorf("Failed to run test shell: %v", err)
	}
	if string(out) != string(shexp) {
		t.Errorf("test shell did not produce expected result: %s (actual) != %s (expected)", out, string(shexp))
	} else {
		t.Logf("shell test passed: %#v = %#v", string(out), string(shexp))
	}
	if arg.Unescaped() != string(out) {
		t.Errorf("Unescaped (%s) is not the same as shell output (%s)", arg.Unescaped(), string(out))
	}
}

func TestStress(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("stress", func(t *testing.T) {
			t.Parallel()
			buf := make([]byte, 32)
			rand.Read(buf)
			for c := 0; c < len(buf); c++ {
				// make an exception for nulls since they're not allowed
				for buf[c] == 0 {
					rand.Read(buf[c : c+1])
				}
			}
			arg := Arg(buf)
			out, err := exec.Command("sh", "-c", "printf %s "+arg.String()).Output()
			if err != nil {
				t.Errorf("Failed to run test shell: %v", err)
			}
			if string(out) != string(buf) {
				t.Errorf("test shell did not produce expected result: %#v (actual) != %#v (expected)", string(out), string(buf))
			} else {
				t.Logf("shell test passed: %#v = %#v", string(out), string(buf))
			}
			if string(out) != arg.Unescaped() {
				t.Errorf("Unescaped arg (%#v) != shell output (%#v) but should be equal", arg.Unescaped(), string(out))
			} else {
				if !arg.Valid() {
					t.Errorf("String was returned from shell unchanged but arg.Valid() is false")
				}
			}
		})
	}
}

func TestUnescapedMultiRepresentation(t *testing.T) {
	str := "cafe\u0301"
	unesc := Arg(str).Unescaped()
	t.Logf("bytelengths: unesc=%d str=%d", len([]byte(unesc)), len([]byte(str)))
	if !bytes.Equal([]byte(unesc), []byte(str)) {
		t.Errorf("didn't get original input back from Unescaped(): %s (unescaped) != %s (original)", unesc, str)
	}
	str = "caf\u00e9"
	unesc = Arg(str).Unescaped()
	t.Logf("bytelengths: unesc=%d str=%d", len([]byte(unesc)), len([]byte(str)))
	if !bytes.Equal([]byte(unesc), []byte(str)) {
		t.Errorf("didn't get original input back from Unescaped(): %s (unescaped) != %s (original)", unesc, str)
	}
	//NegSample{'café'"},
	//PosSample{"cafe\u0301", "'cafe\u0301'"},
}
func TestUnescapedWeirdRunes(t *testing.T) {
	if Arg(`💩`).Unescaped() != `💩` {
		t.Error("empty string didn't match when unescaped")
	}
}

func TestUnescapedEmptyString(t *testing.T) {
	if Arg(``).Unescaped() != `` {
		t.Error("empty string didn't match when unescaped")
	}
}

func ExampleArg_String_regular() {
	// this would make more sense if it had untrusted or otherwise
	// questionable input.  first, collect the args you wish to use.
	stringFromUser := `a string which isn't safe'';exit 1; \";exit 1; ";exit 1`
	stringFromEmptyConfigFile := ""
	stringFromFileContents := "files often end with newlines\n"
	// now build the command line using Arg type annotations to allow fmt
	// to be used
	out, err := exec.Command("sh", "-c",
		fmt.Sprintf(
			"printf 'u: <%%s>\\ne: <%%s>\\nl: <%%s>\\n' %s %s %s",
			Arg(stringFromUser),
			Arg(stringFromEmptyConfigFile),
			Arg(stringFromFileContents),
		)).Output()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", out)
	//Output:
	// u: <a string which isn't safe'';exit 1; \";exit 1; ";exit 1>
	// e: <>
	// l: <files often end with newlines
	// >
}

// strings with NUL aren't supported in sh arguments or variables
// because in C, a NUL char ends a string.  shq can detect and
// work with such strings by letting you detect such behavior and
// examine what the resulting string will be before executing.
func ExampleArg_String_nul() {
	stringWithNul := "a string\000 with a NUL in it"
	arg := Arg(stringWithNul)
	fmt.Printf("Testing with string: %#v\n", stringWithNul)
	fmt.Println("Expecting to see", arg.Unescaped())
	out, err := exec.Command("sh", "-c",
		fmt.Sprintf(
			"printf '<%%s>\\n' %s", Arg(stringWithNul),
		)).Output()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", string(out))
	fmt.Printf("Is Valid()? %v\n", arg.Valid())
	// Output:
	// Testing with string: "a string\x00 with a NUL in it"
	// Expecting to see a string
	// "<a string>\n"
	// Is Valid()? false
}
