package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
)

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
// FROM: https://stackoverflow.com/a/21067803/1515875
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// FROM: https://stackoverflow.com/a/21067803/1515875
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// DirectoryExists returns if a path exists and is a directory
func DirectoryExists(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// FileToSlice reads in all the lines from a file into a string slice
func FileToSlice(filename string) (lines []string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	err = scanner.Err()
	return
}

// FileExists returns if a path exists and is a file
func FileExists(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

// IsAbsPath returns 0 if input path is absolute
func IsAbsPath(path string) bool {
	return strings.HasPrefix(path, "/")
}

// ReadFirstLine gets the first line of a file that has contents and returns it
// if there are no contents, an empty string is returned
// will also return an empty string if the file does not exist
func ReadFirstLine(filename string) (text string) {
	if !FileExists(filename) {
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if text = strings.TrimSpace(scanner.Text()); text == "" {
			continue
		}
		return
	}
	return
}

// SetPermissions sets the proper owner and filemode for a given file
func SetPermissions(path string, fileMode os.FileMode) error {
	if err := os.Chmod(path, fileMode); err != nil {
		return err
	}

	systemGroup := GetenvWithDefault("DOKKU_SYSTEM_GROUP", "dokku")
	systemUser := GetenvWithDefault("DOKKU_SYSTEM_USER", "dokku")

	group, err := user.LookupGroup(systemGroup)
	if err != nil {
		return err
	}
	user, err := user.Lookup(systemUser)
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return err
	}
	return os.Chown(path, uid, gid)
}

func WriteSliceToFile(filename string, lines []string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	if err = w.Flush(); err != nil {
		return err
	}

	file.Chmod(0600)
	SetPermissions(filename, 0600)

	return nil
}