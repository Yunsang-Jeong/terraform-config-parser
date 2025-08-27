package filesystem

import (
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
)

// BillyAdapter adapts billy.Filesystem to FileReader interface
type BillyAdapter struct {
	fs billy.Filesystem
}

func NewBillyAdapter(fs billy.Filesystem) *BillyAdapter {
	return &BillyAdapter{fs: fs}
}

func (b *BillyAdapter) DirExists(dirname string) (bool, error) {
	info, err := b.fs.Stat(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func (b *BillyAdapter) ReadDir(dirname string) ([]os.FileInfo, error) {
	return b.fs.ReadDir(dirname)
}

func (b *BillyAdapter) ReadFile(filename string) ([]byte, error) {
	file, err := b.fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	return io.ReadAll(file)
}