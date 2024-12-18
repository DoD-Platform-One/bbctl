package filesystem

import (
	"os"

	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/commoninterfaces"
)

func userHomeDirFunc() (string, error) {
	return os.UserHomeDir()
}

func createFunc(name string) (commonInterfaces.FileLike, error) {
	return os.Create(name)
}
