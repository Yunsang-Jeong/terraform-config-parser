package filesystem

import (
	"os"

	"github.com/spf13/afero"
)

// AferoAdapter adapts afero.Fs to FileReader interface
type AferoAdapter struct {
	fs afero.Fs
}

func NewAferoAdapter(fs afero.Fs) *AferoAdapter {
	return &AferoAdapter{fs: fs}
}

func (a *AferoAdapter) DirExists(dirname string) (bool, error) {
	info, err := a.fs.Stat(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func (a *AferoAdapter) ReadDir(dirname string) ([]os.FileInfo, error) {
	return afero.ReadDir(a.fs, dirname)
}

func (a *AferoAdapter) ReadFile(filename string) ([]byte, error) {
	return afero.ReadFile(a.fs, filename)
}