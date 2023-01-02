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
	path string
	hash []byte
	size int64
}

const PACKAGE_PATH = ".minigit"

func get_file_hash(path string) []byte {
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

	return hash.Sum(nil)
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

		*files = append(*files, file2{file_path, file_hash, file_size})
	}
}

func create_package(files []file2, dir string) {
	err := os.MkdirAll(filepath.Join(dir, PACKAGE_PATH), os.ModePerm)

	if err != nil {
		panic(err)
	}

	output, err := os.Create(filepath.Join(dir, PACKAGE_PATH, "output"))

	if err != nil {
		panic(err)
	}

	defer output.Close()

	for _, file := range files {
		file_entry := fmt.Sprintf("%s\t%x\t%d\n", file.path, file.hash, file.size)
		output.WriteString(file_entry)
	}

	for _, file := range files {
		input, err := os.Open(filepath.Join(dir, file.path))

		if err != nil {
			panic(err)
		}

		output.WriteString("~")
		io.Copy(output, input)
		output.WriteString("\n")
	}
}

func read_package(dir string) []file2 {
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

		file_prefix_byte := byte('~')

		file_index_at := 0
		for i, c := range file {
			if c == file_prefix_byte {
				file_index_at = i
				break
			}
		}

		index_str := string(file[:file_index_at])
		files_str := strings.Split(index_str, "\n")

		for _, file_str := range files_str {
			if len(file_str) == 0 {
				continue
			}

			file_infos_str := strings.Split(file_str, "\t")

			file_size, err_size := strconv.Atoi(file_infos_str[2])
			file_hash, err_hash := hex.DecodeString(file_infos_str[1])

			if err_size != nil {
				panic(err_size)
			}
			if err_hash != nil {
				panic(err_hash)
			}

			file := file2{
				file_infos_str[0],
				file_hash,
				int64(file_size),
			}

			files = append(files, file)
		}
	}

	return files
}

func main() {
	os_args := os.Args[1:]
	fmt.Println(os_args)

	action := "package"
	dir := ""

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

		for _, file := range files {
			fmt.Printf("%-20s\t%x\t%-5d\n", file.path, file.hash, file.size)
		}

		create_package(files, dir)
	}

	if action == "inspect" {
		files := read_package(dir)

		for _, file := range files {
			fmt.Printf("%-20s\t%x\t%-5d\n", file.path, file.hash, file.size)
		}
	}
}
