package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
)

// Symlink creates newname as a symbolic link to oldname,
// if newname not exist or is a symlink, it will be replaced directly,
// in other cases, it will be backed up first
func Symlink(oldname, newname string) error {
	stat, err := os.Lstat(newname)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrap(os.Symlink(oldname, newname), "create symlink error")
		}
		return errors.Wrap(err, "os.Lstat error")
	}

	if stat.Mode()&os.ModeSymlink == 0 {
		// is not a symlink
		if err := BackUp(newname); err != nil {
			return errors.Wrap(err, "back up error")
		}
	} else {
		// is a symlink
		if err := os.Remove(newname); err != nil {
			return errors.Wrap(err, "remove old symlink error")
		}
	}
	return errors.Wrap(os.Symlink(oldname, newname), "create symlink error")
}

// BackUp backs up a file, back up file will be named like filePath-backup-n
func BackUp(filePath string) error {
	index := 1
	backupPath := ""
	for i := 0; i < 999; i++ {
		backupPath = fmt.Sprintf("%s-backup-%d", filePath, index)
		_, err := os.Lstat(backupPath)
		if err == nil {
			index += 1 // backupPath already exists, try bigger index
			continue
		}

		// unknown error
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "os.Lstat error")
		}

		// backupPath not exist
		fmt.Printf("\nRenaming from %s to %s\n", filePath, backupPath)
		return errors.Wrap(os.Rename(filePath, backupPath), "os.Rename error")
	}
	return errors.New("Too many backup revisions of this file")
}
