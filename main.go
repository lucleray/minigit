package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type file0 struct {
	path    string
	hash    string
	version string
	offset  int
	size    int64
}

const PACKAGE_PATH = ".minigit"
const CURRENT_VERSION = "~"

func get_file_hash(path string) string {
	file, err := os.Open(path)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func scan_dir(files *[]file0, dir string, subdir string) {
	entries, err := ioutil.ReadDir(filepath.Join(dir, subdir))

	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.Name() == ".minigit" {
			continue
		}

		if entry.IsDir() {
			scan_dir(files, dir, filepath.Join(subdir, entry.Name()))
			continue
		}

		file_path := filepath.Join(subdir, entry.Name())
		file_hash := get_file_hash(filepath.Join(dir, subdir, entry.Name()))
		file_offset := -1
		file_size := entry.Size()

		*files = append(*files, file0{file_path, file_hash, CURRENT_VERSION, file_offset, file_size})
	}
}

func build_index(files []file0) string {
	index_str := ""

	for i, file := range files {
		index_str = index_str + fmt.Sprintf(
			"%s\t%s\t%s\t%d\t%d",
			file.path,
			file.hash,
			file.version,
			file.offset,
			file.size,
		)

		if i != len(files)-1 {
			index_str = index_str + "\n"
		}
	}

	return index_str
}

func get_version(files []file0) string {
	index_str := build_index(files)

	hash := md5.New()
	_, err := io.WriteString(hash, index_str)

	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func search_file(files []file0, path string, hash string) string {
	for _, file := range files {
		if hash == file.hash && path == file.path {
			return file.version
		}
	}

	return CURRENT_VERSION
}

func create_package(files []file0, version string, dir string) {
	err := os.MkdirAll(filepath.Join(dir, PACKAGE_PATH), os.ModePerm)

	if err != nil {
		panic(err)
	}

	existing_files := read_packages(dir, []string{version})

	output, err := os.Create(filepath.Join(dir, PACKAGE_PATH, version))

	if err != nil {
		panic(err)
	}

	defer output.Close()

	for i, file := range files {
		files[i].version = search_file(existing_files, file.path, file.hash)
	}

	output.WriteString(build_index(files))

	for _, file := range files {
		if file.version != CURRENT_VERSION {
			continue
		}

		input, err := os.Open(filepath.Join(dir, file.path))

		if err != nil {
			panic(err)
		}

		output.WriteString("\n" + CURRENT_VERSION)
		io.Copy(output, input)
	}
}

func has_version(exclude_versions []string, version string) bool {
	for _, exclude_version := range exclude_versions {
		if exclude_version == version {
			return true
		}
	}
	return false
}

func read_packages(dir string, exclude_versions []string) []file0 {
	package_path := filepath.Join(dir, PACKAGE_PATH)
	package_entries, err := ioutil.ReadDir(package_path)

	if err != nil {
		panic(err)
	}

	files := []file0{}

	for _, package_entry := range package_entries {
		if has_version(exclude_versions, package_entry.Name()) {
			continue
		}

		file, err := os.ReadFile(filepath.Join(package_path, package_entry.Name()))

		if err != nil {
			panic(err)
		}

		file_prefix_byte := byte(CURRENT_VERSION[0])
		new_line_byte := byte('\n')

		file_index_at := 0
		first_char_of_line := true
		for i, c := range file {
			if first_char_of_line && c == file_prefix_byte {
				file_index_at = i
				break
			}

			first_char_of_line = (c == new_line_byte)
		}

		index_str := string(file[:file_index_at])
		files_str := strings.Split(index_str, "\n")

		for _, file_str := range files_str {
			if len(file_str) == 0 {
				continue
			}

			// format is:
			// <path>	<hash> <version> <offset> <size>
			file_infos_str := strings.Split(file_str, "\t")

			version := file_infos_str[2]
			if version == CURRENT_VERSION {
				version = package_entry.Name()
			}

			file_offset, err_offset := strconv.Atoi(file_infos_str[3])
			file_size, err_size := strconv.Atoi(file_infos_str[4])

			if err_offset != nil {
				panic(err_offset)
			}
			if err_size != nil {
				panic(err_size)
			}

			file := file0{
				file_infos_str[0],
				file_infos_str[1],
				version,
				file_offset,
				int64(file_size),
			}

			files = append(files, file)
		}
	}

	return files
}

func main() {
	action := "package"
	dir := ""

	os_args := os.Args[1:]
	for _, arg := range os_args {
		if strings.HasPrefix(arg, "--dir=") {
			dir = arg[len("--dir="):]
		}

		if arg == "--inspect" || arg == "-i" {
			action = "inspect"
		}
	}

	if action == "package" {
		files := []file0{}
		scan_dir(&files, dir, "")
		version := get_version(files)
		create_package(files, version, dir)
		fmt.Println("📦", version)
	}

	if action == "inspect" {
		files := read_packages(dir, []string{})

		for _, file := range files {
			fmt.Printf("%-20s\t%s\t%-5d\t%s\n", file.path, file.hash, file.size, file.version)
		}
	}
}
