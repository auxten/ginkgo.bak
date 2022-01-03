package seed

import (
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/auxten/logrus"
)

func MakeSeed(dir string, blockSize int64) (seed *Seed, err error) {
	seed = new(Seed)
	if blockSize > 0 {
		seed.BlockSize = blockSize
	} else {
		seed.BlockSize = DefaultBlockSize
	}
	err = filepath.WalkDir(dir, getWalkFunc(seed))

	seed.Blocks[len(seed.Blocks)-1].Size = seed.TotalSize % seed.BlockSize
	return
}

func getWalkFunc(s *Seed) func(string, fs.DirEntry, error) error {
	var (
		seed          = s
		lastTotalSize int64
	)
	seed.Files = make([]*File, 0)
	seed.Blocks = make([]*Block, 0)
	return func(path string, entry fs.DirEntry, err error) error {
		var (
			fInfo   fs.FileInfo
			size    int64
			symPath string
		)
		fInfo, err = entry.Info()
		log.Debugf("%s %d", path, fInfo.Size())
		if err != nil {
			return err
		}

		seed.FileCount++
		if entry.IsDir() {
			size = -1
		} else if (entry.Type() & fs.ModeSymlink) != 0 {
			size = -2
			if symPath, err = os.Readlink(path); err != nil {
				return err
			}
		} else if entry.Type().IsRegular() {
			size = fInfo.Size()
			lastTotalSize = seed.TotalSize
			seed.TotalSize += size
		}

		seed.Files = append(seed.Files, &File{
			mtime:    fInfo.ModTime(),
			Mode:     fInfo.Mode(),
			Size:     size,
			SymPath:  symPath,
			Path:     path,
			CheckSum: nil,
		})

		if size > 0 {
			for int64(len(seed.Blocks))*seed.BlockSize < seed.TotalSize {
				startOffset := int64(len(seed.Blocks))*seed.BlockSize - lastTotalSize
				seed.Blocks = append(seed.Blocks, &Block{
					Size:        seed.BlockSize,
					StartFile:   len(seed.Files) - 1,
					StartOffset: startOffset,
				})
			}
		}

		return nil
	}
}
