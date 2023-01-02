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

type file2 struct {
	path    string
	hash    string
	size    int64
	version string
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

func scan_dir(files *[]file2, dir string, subdir string) {
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
		file_size := entry.Size()

		*files = append(*files, file2{file_path, file_hash, file_size, CURRENT_VERSION})
	}
}

func build_index(files []file2) string {
	index_str := ""
	for _, file := range files {
		index_str = index_str + fmt.Sprintf(
			"%s\t%s\t%d\t%s\n",
			file.path,
			file.hash,
			file.size,
			file.version,
		)
	}
	return index_str
}

func get_version(files []file2) string {
	index_str := build_index(files)

	hash := md5.New()
	_, err := io.WriteString(hash, index_str)

	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func search_file(files []file2, path string, hash string) string {
	for _, file := range files {
		if hash == file.hash && path == file.path {
			return file.version
		}
	}

	return CURRENT_VERSION
}

func create_package(files []file2, version string, dir string) {
	err := os.MkdirAll(filepath.Join(dir, PACKAGE_PATH), os.ModePerm)

	if err != nil {
		panic(err)
	}

	existing_files := read_packages(dir)

	output, err := os.Create(filepath.Join(dir, PACKAGE_PATH, version))

	if err != nil {
		panic(err)
	}

	defer output.Close()

	for i, file := range files {
		version := search_file(existing_files, file.path, file.hash)

		// TODO: refactor to something better
		file.version = version
		files[i].version = version

		file_entry := fmt.Sprintf("%s\t%s\t%d\t%s\n", file.path, file.hash, file.size, file.version)
		output.WriteString(file_entry)
	}

	for _, file := range files {
		if file.version != CURRENT_VERSION {
			continue
		}

		input, err := os.Open(filepath.Join(dir, file.path))

		if err != nil {
			panic(err)
		}

		output.WriteString(CURRENT_VERSION)
		io.Copy(output, input)
		output.WriteString("\n")
	}
}

func read_packages(dir string) []file2 {
	package_path := filepath.Join(dir, PACKAGE_PATH)
	package_entries, err := ioutil.ReadDir(package_path)

	if err != nil {
		panic(err)
	}

	files := []file2{}

	for _, package_entry := range package_entries {
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
			// <path>	<hash> <size> <version>
			file_infos_str := strings.Split(file_str, "\t")

			file_size, err_size := strconv.Atoi(file_infos_str[2])

			if err_size != nil {
				panic(err_size)
			}

			version := file_infos_str[3]
			if file_infos_str[3] == CURRENT_VERSION {
				version = package_entry.Name()
			}

			file := file2{
				file_infos_str[0],
				file_infos_str[1],
				int64(file_size),
				version,
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
		files := []file2{}
		scan_dir(&files, dir, "")
		version := get_version(files)
		create_package(files, version, dir)
		fmt.Println("📦", version)
	}

	if action == "inspect" {
		files := read_packages(dir)

		for _, file := range files {
			fmt.Printf("%-20s\t%x\t%-5d\t%s\n", file.path, file.hash, file.size, file.version)
		}
	}
}
