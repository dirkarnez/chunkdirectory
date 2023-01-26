package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func ensureDir(fileName string) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		return os.MkdirAll(dirName, os.ModePerm)
	}
	return nil
}

func CopyFile(src, dst string) error {
	fmt.Printf("[CopyFile]: %s, %s\n", src, dst)
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}

	err = ensureDir(dst)
	if err != nil {
		return fmt.Errorf("couldn't ensure dest file: %s", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("sync error: %s", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat error: %s", err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return fmt.Errorf("chmod error: %s", err)
	}

	return nil
}

var (
	target string
	source string
	max    uint64
)

func main() {
	flag.Uint64Var(&max, "max", 100, "Max count, default is 100")
	flag.StringVar(&target, "target", "", "Target directory")
	flag.StringVar(&source, "source", "", "Source directory")
	flag.Parse()

	if max < 1 {
		log.Fatalln("--max should be greater than 0")
	}

	if len(source) < 1 {
		log.Fatalln("--source should not be nil")
	}

	if len(target) < 1 {
		log.Fatalln("--target should not be nil")
	}

	var counter uint64 = 0
	var batch uint64 = 0
	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if counter == 0 {
			// new folder
			batch++
		}

		currentTarget := filepath.Join(target, fmt.Sprintf("%d", batch))
		chunk := strings.TrimLeft(path, filepath.Dir(filepath.Dir(path)))

		err = CopyFile(path, filepath.Join(currentTarget, chunk))
		if err != nil {
			log.Println(err.Error())
		}

		counter++
		if counter == max {
			// new folder
			counter = 0
		}

		return nil
	})
}
