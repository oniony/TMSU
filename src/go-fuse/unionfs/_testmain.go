package main

import target "github.com/hanwen/go-fuse/unionfs"
import "testing"
import "os"
import "regexp"

var tests = []testing.InternalTest{
	{"unionfs.TestVersion", target.TestVersion},
	{"unionfs.TestAutoFsSymlink", target.TestAutoFsSymlink},
	{"unionfs.TestDetectSymlinkedDirectories", target.TestDetectSymlinkedDirectories},
	{"unionfs.TestExplicitScan", target.TestExplicitScan},
	{"unionfs.TestCreationChecks", target.TestCreationChecks},
	{"unionfs.TestCachingFs", target.TestCachingFs},
	{"unionfs.TestMemUnionFsSymlink", target.TestMemUnionFsSymlink},
	{"unionfs.TestMemUnionFsSymlinkPromote", target.TestMemUnionFsSymlinkPromote},
	{"unionfs.TestMemUnionFsChtimes", target.TestMemUnionFsChtimes},
	{"unionfs.TestMemUnionFsChmod", target.TestMemUnionFsChmod},
	{"unionfs.TestMemUnionFsChown", target.TestMemUnionFsChown},
	{"unionfs.TestMemUnionFsDelete", target.TestMemUnionFsDelete},
	{"unionfs.TestMemUnionFsBasic", target.TestMemUnionFsBasic},
	{"unionfs.TestMemUnionFsPromote", target.TestMemUnionFsPromote},
	{"unionfs.TestMemUnionFsSubdirCreate", target.TestMemUnionFsSubdirCreate},
	{"unionfs.TestMemUnionFsCreate", target.TestMemUnionFsCreate},
	{"unionfs.TestMemUnionFsOpenUndeletes", target.TestMemUnionFsOpenUndeletes},
	{"unionfs.TestMemUnionFsMkdir", target.TestMemUnionFsMkdir},
	{"unionfs.TestMemUnionFsMkdirPromote", target.TestMemUnionFsMkdirPromote},
	{"unionfs.TestMemUnionFsRmdirMkdir", target.TestMemUnionFsRmdirMkdir},
	{"unionfs.TestMemUnionFsLink", target.TestMemUnionFsLink},
	{"unionfs.TestMemUnionFsCreateLink", target.TestMemUnionFsCreateLink},
	{"unionfs.TestMemUnionFsTruncate", target.TestMemUnionFsTruncate},
	{"unionfs.TestMemUnionFsCopyChmod", target.TestMemUnionFsCopyChmod},
	{"unionfs.TestMemUnionFsTruncateTimestamp", target.TestMemUnionFsTruncateTimestamp},
	{"unionfs.TestMemUnionFsRemoveAll", target.TestMemUnionFsRemoveAll},
	{"unionfs.TestMemUnionFsRmRf", target.TestMemUnionFsRmRf},
	{"unionfs.TestMemUnionFsDeletedGetAttr", target.TestMemUnionFsDeletedGetAttr},
	{"unionfs.TestMemUnionFsDoubleOpen", target.TestMemUnionFsDoubleOpen},
	{"unionfs.TestMemUnionFsUpdate", target.TestMemUnionFsUpdate},
	{"unionfs.TestMemUnionFsFdLeak", target.TestMemUnionFsFdLeak},
	{"unionfs.TestMemUnionFsStatFs", target.TestMemUnionFsStatFs},
	{"unionfs.TestMemUnionFsFlushSize", target.TestMemUnionFsFlushSize},
	{"unionfs.TestMemUnionFsFlushRename", target.TestMemUnionFsFlushRename},
	{"unionfs.TestMemUnionFsTruncGetAttr", target.TestMemUnionFsTruncGetAttr},
	{"unionfs.TestMemUnionFsRenameDirBasic", target.TestMemUnionFsRenameDirBasic},
	{"unionfs.TestMemUnionFsRenameDirAllSourcesGone", target.TestMemUnionFsRenameDirAllSourcesGone},
	{"unionfs.TestMemUnionFsRenameDirWithDeletions", target.TestMemUnionFsRenameDirWithDeletions},
	{"unionfs.TestMemUnionGc", target.TestMemUnionGc},
	{"unionfs.TestTimedCache", target.TestTimedCache},
	{"unionfs.TestFilePathHash", target.TestFilePathHash},
	{"unionfs.TestUnionFsAutocreateDeletionDir", target.TestUnionFsAutocreateDeletionDir},
	{"unionfs.TestUnionFsSymlink", target.TestUnionFsSymlink},
	{"unionfs.TestUnionFsSymlinkPromote", target.TestUnionFsSymlinkPromote},
	{"unionfs.TestUnionFsChtimes", target.TestUnionFsChtimes},
	{"unionfs.TestUnionFsChmod", target.TestUnionFsChmod},
	{"unionfs.TestUnionFsChown", target.TestUnionFsChown},
	{"unionfs.TestUnionFsDelete", target.TestUnionFsDelete},
	{"unionfs.TestUnionFsBasic", target.TestUnionFsBasic},
	{"unionfs.TestUnionFsPromote", target.TestUnionFsPromote},
	{"unionfs.TestUnionFsCreate", target.TestUnionFsCreate},
	{"unionfs.TestUnionFsOpenUndeletes", target.TestUnionFsOpenUndeletes},
	{"unionfs.TestUnionFsMkdir", target.TestUnionFsMkdir},
	{"unionfs.TestUnionFsMkdirPromote", target.TestUnionFsMkdirPromote},
	{"unionfs.TestUnionFsRmdirMkdir", target.TestUnionFsRmdirMkdir},
	{"unionfs.TestUnionFsRename", target.TestUnionFsRename},
	{"unionfs.TestUnionFsRenameDirBasic", target.TestUnionFsRenameDirBasic},
	{"unionfs.TestUnionFsRenameDirAllSourcesGone", target.TestUnionFsRenameDirAllSourcesGone},
	{"unionfs.TestUnionFsRenameDirWithDeletions", target.TestUnionFsRenameDirWithDeletions},
	{"unionfs.TestUnionFsRenameSymlink", target.TestUnionFsRenameSymlink},
	{"unionfs.TestUnionFsWritableDir", target.TestUnionFsWritableDir},
	{"unionfs.TestUnionFsWriteAccess", target.TestUnionFsWriteAccess},
	{"unionfs.TestUnionFsLink", target.TestUnionFsLink},
	{"unionfs.TestUnionFsTruncate", target.TestUnionFsTruncate},
	{"unionfs.TestUnionFsCopyChmod", target.TestUnionFsCopyChmod},
	{"unionfs.TestUnionFsTruncateTimestamp", target.TestUnionFsTruncateTimestamp},
	{"unionfs.TestUnionFsRemoveAll", target.TestUnionFsRemoveAll},
	{"unionfs.TestUnionFsRmRf", target.TestUnionFsRmRf},
	{"unionfs.TestUnionFsDropDeletionCache", target.TestUnionFsDropDeletionCache},
	{"unionfs.TestUnionFsDropCache", target.TestUnionFsDropCache},
	{"unionfs.TestUnionFsDisappearing", target.TestUnionFsDisappearing},
	{"unionfs.TestUnionFsDeletedGetAttr", target.TestUnionFsDeletedGetAttr},
	{"unionfs.TestUnionFsDoubleOpen", target.TestUnionFsDoubleOpen},
	{"unionfs.TestUnionFsFdLeak", target.TestUnionFsFdLeak},
	{"unionfs.TestUnionFsStatFs", target.TestUnionFsStatFs},
	{"unionfs.TestUnionFsFlushSize", target.TestUnionFsFlushSize},
	{"unionfs.TestUnionFsFlushRename", target.TestUnionFsFlushRename},
	{"unionfs.TestUnionFsTruncGetAttr", target.TestUnionFsTruncGetAttr},
	{"unionfs.TestUnionFsPromoteDirTimeStamp", target.TestUnionFsPromoteDirTimeStamp},
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
