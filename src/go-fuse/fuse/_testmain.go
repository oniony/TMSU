package main

import target "github.com/hanwen/go-fuse/fuse"
import "testing"
import "os"
import "regexp"

var tests = []testing.InternalTest{
	{"fuse.TestIntToExponent", target.TestIntToExponent},
	{"fuse.TestBufferPool", target.TestBufferPool},
	{"fuse.TestFreeBufferEmpty", target.TestFreeBufferEmpty},
	{"fuse.TestCacheFs", target.TestCacheFs},
	{"fuse.TestNonseekable", target.TestNonseekable},
	{"fuse.TestGetAttrRace", target.TestGetAttrRace},
	{"fuse.TestCopyFile", target.TestCopyFile},
	{"fuse.TestRawFs", target.TestRawFs},
	{"fuse.TestPathFs", target.TestPathFs},
	{"fuse.TestDummyFile", target.TestDummyFile},
	{"fuse.TestFSetAttr", target.TestFSetAttr},
	{"fuse.TestHandleMapDoubleRegister", target.TestHandleMapDoubleRegister},
	{"fuse.TestHandleMapUnaligned", target.TestHandleMapUnaligned},
	{"fuse.TestHandleMapPointerLayout", target.TestHandleMapPointerLayout},
	{"fuse.TestHandleMapBasic", target.TestHandleMapBasic},
	{"fuse.TestHandleMapMultiple", target.TestHandleMapMultiple},
	{"fuse.TestHandleMapCheckFail", target.TestHandleMapCheckFail},
	{"fuse.TestLatencyMap", target.TestLatencyMap},
	{"fuse.TestOpenUnreadable", target.TestOpenUnreadable},
	{"fuse.TestTouch", target.TestTouch},
	{"fuse.TestRemove", target.TestRemove},
	{"fuse.TestWriteThrough", target.TestWriteThrough},
	{"fuse.TestMkdirRmdir", target.TestMkdirRmdir},
	{"fuse.TestLinkCreate", target.TestLinkCreate},
	{"fuse.TestLinkExisting", target.TestLinkExisting},
	{"fuse.TestLinkForget", target.TestLinkForget},
	{"fuse.TestSymlink", target.TestSymlink},
	{"fuse.TestRename", target.TestRename},
	{"fuse.TestDelRename", target.TestDelRename},
	{"fuse.TestOverwriteRename", target.TestOverwriteRename},
	{"fuse.TestAccess", target.TestAccess},
	{"fuse.TestMknod", target.TestMknod},
	{"fuse.TestReaddir", target.TestReaddir},
	{"fuse.TestFSync", target.TestFSync},
	{"fuse.TestLargeRead", target.TestLargeRead},
	{"fuse.TestLargeDirRead", target.TestLargeDirRead},
	{"fuse.TestRootDir", target.TestRootDir},
	{"fuse.TestIoctl", target.TestIoctl},
	{"fuse.TestStatFs", target.TestStatFs},
	{"fuse.TestFStatFs", target.TestFStatFs},
	{"fuse.TestOriginalIsSymlink", target.TestOriginalIsSymlink},
	{"fuse.TestDoubleOpen", target.TestDoubleOpen},
	{"fuse.TestUmask", target.TestUmask},
	{"fuse.TestMemNodeFs", target.TestMemNodeFs},
	{"fuse.TestOsErrorToErrno", target.TestOsErrorToErrno},
	{"fuse.TestLinkAt", target.TestLinkAt},
	{"fuse.TestMountOnExisting", target.TestMountOnExisting},
	{"fuse.TestMountRename", target.TestMountRename},
	{"fuse.TestMountReaddir", target.TestMountReaddir},
	{"fuse.TestRecursiveMount", target.TestRecursiveMount},
	{"fuse.TestDeletedUnmount", target.TestDeletedUnmount},
	{"fuse.TestInodeNotify", target.TestInodeNotify},
	{"fuse.TestEntryNotify", target.TestEntryNotify},
	{"fuse.TestOwnerDefault", target.TestOwnerDefault},
	{"fuse.TestOwnerRoot", target.TestOwnerRoot},
	{"fuse.TestOwnerOverride", target.TestOwnerOverride},
	{"fuse.TestSwitchFsSlash", target.TestSwitchFsSlash},
	{"fuse.TestSwitchFs", target.TestSwitchFs},
	{"fuse.TestSwitchFsStrip", target.TestSwitchFsStrip},
	{"fuse.TestSwitchFsApi", target.TestSwitchFsApi},
	{"fuse.TestXAttrRead", target.TestXAttrRead},
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
