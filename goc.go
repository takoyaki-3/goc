package goc

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"path/filepath"
)

func ZipArchive(output string, paths []string, filenames []string) error {
	var compressedFile *os.File
	var err error

	//ZIPファイル作成
	if compressedFile, err = os.Create(output); err != nil {
		return err
	}
	defer compressedFile.Close()

	if err := compress(compressedFile, ".", paths, filenames); err != nil {
		return err
	}

	return nil
}

func compress(compressedFile io.Writer, targetDir string, paths []string, files []string) error {
	w := zip.NewWriter(compressedFile)

	for k, filename := range paths {
		filepath := fmt.Sprintf("%s/%s", targetDir, filename)
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
		hdr.Name = files[k]
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

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Dirwalk(dir string) ([]string, []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths, file_names []string
	for _, file := range files {
		paths = append(paths, filepath.Join(dir, file.Name()))
		file_names = append(file_names, file.Name())
	}
	return paths, file_names
}

func BOMCsvReader(r io.Reader) *csv.Reader {
	br := bufio.NewReader(r)
	bs, err := br.Peek(3)
	if err != nil {
		return csv.NewReader(br)
	}
	if bs[0] == 0xEF && bs[1] == 0xBB && bs[2] == 0xBF {
		br.Discard(3)
	}
	return csv.NewReader(br)
}

func SpiritCSV(filename string, outputpath string, column string) {

	buf := map[string][][]string{}

	// Make directory
	os.Mkdir(outputpath, 0777)

	// Open read table file
	rf, rerr := os.Open(filename)
	if rerr != nil {
		fmt.Println("[" + filename + "] is not found")
	}
	defer rf.Close()
	reader := BOMCsvReader(rf)
	reader.FieldsPerRecord = -1

	counter := -1
	titles := map[string]int{}
	first_line := []string{}

	for {
		counter++
		line, er := reader.Read()
		if er != nil {
			break
		}
		if counter == 0 {
			first_line = line
			for k, v := range line {
				titles[v] = k
			}
			continue
		}
		id := line[titles[column]]
		if _, ok := buf[id]; !ok {
			buf[id] = [][]string{first_line}
		}
		// if len(first_line) == len(line){
		buf[id] = append(buf[id],line)
		// }
	}

	// Output files
	for k, v := range buf {
		// Open write table file
		wf, werr := os.Create(outputpath + "/" + k + ".csv")
		if werr != nil {
			log.Fatal(werr)
		}
		writer := csv.NewWriter(wf)
		for _, line := range v {
			writer.Write(line)
		}
		writer.Flush()
		wf.Close()
	}
}

func ReadCSV(path string) (map[string]int, [][]string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := BOMCsvReader(file)
	reader.FieldsPerRecord = -1
	var line []string

	counter := -1
	titles := map[string]int{}
	data := [][]string{}
	for {
		counter++
		line, err = reader.Read()
		if err != nil {
			break
		}
		if counter == 0 {
			for k, v := range line {
				titles[v] = k
			}
			continue
		}
		data = append(data, line)
	}

	return titles, data
}

func Merge(base [][]string, app [][]string, base_on string, app_on string) [][]string {

	apps := map[string][]string{}

	appTitles := map[string]int{}
	baseTitles := map[string]int{}

	firstLine := append(base[0], app[0]...)
	notFound := make([]string,len(app[0]))

	for k, line := range app {
		if k == 0 {
			for k, v := range line {
				appTitles[v] = k
			}
			continue
		}
		apps[line[appTitles[app_on]]] = line
	}

	resp := [][]string{firstLine}

	for k, line := range base {
		if k == 0 {
			for k, v := range line {
				baseTitles[v] = k
			}
			continue
		}
		if v,ok:=apps[line[baseTitles[base_on]]];ok{
			resp = append(resp, append(line, v...))
		} else {
			resp = append(resp, append(line, notFound...))
		}
	}

	return resp
}

func Read2DStr(filePath string) [][]string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := BOMCsvReader(file)
	reader.FieldsPerRecord = -1

	data := [][]string{}
	for {
		line, err := reader.Read()
		if err != nil {
			break
		}
		data = append(data, line)
	}

	return data
}

func Write2DStr(filePath string, data [][]string) error {
	// Open write table file
	wf, err := os.Create(filePath)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(wf)
	for _, line := range data {
		writer.Write(line)
	}
	writer.Flush()
	wf.Close()
	return nil
}

func MergeCSV(base_filename string, append_filename string, out_filename string, base_on string, append_on string) {
	base := Read2DStr(base_filename)
	app := Read2DStr(append_filename)
	data := Merge(base, app, base_on, append_on)
	Write2DStr(out_filename, data)
}

func Select2DStr(columns []string,data [][]string)[][]string{
	selectList := []int{}
	for _,v := range columns{
		for kk,vv := range data{
			if len(vv) < 1{
				continue
			}
			if vv[0] != v {
				continue
			}
			selectList = append(selectList, kk)
		}
	}
	ans := [][]string{}
	for _,v:=range data{
		line := []string{}
		for _,k := range selectList{
			line = append(line, v[k])
		}
		ans = append(ans, line)
	}
	return ans
}

// コア数,総ループ数,呼び出し関数(インデックス,スレッド番号)
func Parallel(core int,n int,f func(int,int)){
	wg := sync.WaitGroup{}
	wg.Add(core)
	for rank:=0;rank<core;rank++{
		go func(rank int){
			defer wg.Done()
			for i:=rank;i<n;i+=core{
				f(i,rank)
			}
		}(rank)
	}
	wg.Wait()
}
