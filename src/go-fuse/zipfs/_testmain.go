package main

import target "github.com/hanwen/go-fuse/zipfs"
import "testing"
import "os"
import "regexp"

var tests = []testing.InternalTest{
	{"zipfs.TestMultiZipReadonly", target.TestMultiZipReadonly},
	{"zipfs.TestMultiZipFs", target.TestMultiZipFs},
	{"zipfs.TestZipFs", target.TestZipFs},
	{"zipfs.TestLinkCount", target.TestLinkCount},
}

var benchmarks = []testing.InternalBenchmark{}
var examples = []testing.InternalExample{}

var matchPat string
var matchRe *regexp.Regexp

func matchString(pat, str string) (result bool, err os.Error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = regexp.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

func main() {
	testing.Main(matchString, tests, benchmarks, examples)
}
