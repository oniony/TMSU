package main

import target "github.com/hanwen/go-fuse/benchmark"
import "testing"
import "os"
import "regexp"

var tests = []testing.InternalTest{
	{"fuse.TestNewStatFs", target.TestNewStatFs},
}

var benchmarks = []testing.InternalBenchmark{	{"fuse.BenchmarkGoFuseThreadedStat", target.BenchmarkGoFuseThreadedStat},
	{"fuse.BenchmarkCFuseThreadedStat", target.BenchmarkCFuseThreadedStat},
}
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
