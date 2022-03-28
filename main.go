package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	cp "github.com/otiai10/copy"
	"gopkg.in/ini.v1"
)

func initConfig() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir := fmt.Sprintf("%v\\hikariAssist", dir)
	iniDir := fmt.Sprintf("%v\\config.ini", configDir)
	tempDir := fmt.Sprintf("%v\\temp", configDir)

	if !doesExist(configDir) {
		fmt.Println("Enter your McOsu directory as formatted in game:")
		var mcDir string
		fmt.Scanln(&mcDir)
		os.Mkdir(configDir, 0700)
		os.Mkdir(tempDir, 0700)
		f, e := os.Create(iniDir)
		if e != nil {
			panic(e)
		}
		f.Close()
		fmt.Fprint(f, fmt.Sprintf("[paths]\npath = %v", mcDir))
		return mcDir
	} else {
		cfg, err := ini.Load(fmt.Sprintf(iniDir))
		if err != nil {
			log.Fatal(err)
		}
		dataPath := cfg.Section("paths").Key("path").String()
		return dataPath
	}
}

func main() {
	mcOsuPath := initConfig()

	oPath := os.Args[1]
	fName := strings.Split(oPath, "\\")
	osuFileName := fName[len(fName)-1]

	extension := filepath.Ext(osuFileName)

	var noExtension string

	switch extension {
	case ".osz":
		osuFileName = strings.Replace(osuFileName, ".osz", ".zip", -1)
		noExtension = strings.Replace(osuFileName, ".osz", "", -1)
	case ".osk":
		osuFileName = strings.Replace(osuFileName, ".osk", ".zip", -1)
		noExtension = strings.Replace(osuFileName, ".osk", "", -1)
	default:
		log.Fatal("Wrong file format? Exiting...")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	nPath := fmt.Sprintf("%v\\hikariAssist\\temp\\%v", configDir, osuFileName)

	e := os.Rename(oPath, nPath)
	if e != nil {
		log.Fatal(e)
	}

	absDestination, err := filepath.Abs(nPath)
	if err != nil {
		log.Fatal(err)
	}

	destination := strings.Replace(absDestination, ".zip", "", -1)

	_, err1 := Unzip(absDestination, destination)
	if err1 != nil {
		log.Fatal(err1)
	}

	moveSong(mcOsuPath, destination, extension, noExtension)
}

func moveSong(dir string, destination string, extension string, name string) {
	switch extension {
	case ".osz":
		songDir := fmt.Sprintf("%v\\Songs\\%v", dir, name)
		os.Mkdir(songDir, 0755)
		e := cp.Copy(destination, songDir)
		if e != nil {
			log.Fatal(e)
		}
	case ".osk":
		skinDir := fmt.Sprintf("%v\\Skins\\%v", dir, name)
		os.Mkdir(skinDir, 0755)
		e := cp.Copy(destination, skinDir)
		if e != nil {
			log.Fatal(e)
		}
	default:
		log.Fatal("literally how")
	}
	removeFile()
}

// https://stackoverflow.com/questions/64568660/the-process-cannot-access-the-file-because-it-is-being-used-by-another-process-i
func Unzip(src string, destination string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)

	if err != nil {

		return filenames, err
	}

	defer r.Close()

	for _, f := range r.File {

		fpath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}
		outFile, err := os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_RDWR,
			f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()

		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}

func removeFile() {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	tempDir := fmt.Sprintf("%v\\hikariAssist\\temp", dir)

	files, err := filepath.Glob(filepath.Join(tempDir, "*"))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// https://github.com/nina-x/hikari/blob/main/main.go
func doesExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
