package goc

import (
	"os"
	"io"
	"fmt"
	"io/ioutil"
	"archive/zip"
)

func ZipArchive(output string, paths []string) error {
	var compressedFile *os.File
	var err error

	//ZIPファイル作成
	if compressedFile, err = os.Create(output); err != nil {
		return err
	}
	defer compressedFile.Close()

	if err := compress(compressedFile, ".", paths); err != nil {
		return err
	}

	return nil
}

func compress(compressedFile io.Writer, targetDir string, files []string) error {
	w := zip.NewWriter(compressedFile)

	for _, filename := range files {
		filepath := fmt.Sprintf("%s/%s", targetDir, "./onegtfs/" + filename)
		info, err := os.Stat(filepath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		hdr.Name = filename

		f, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}

		contents, _ := ioutil.ReadFile(filepath)
		_, err = f.Write(contents)
		if err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}