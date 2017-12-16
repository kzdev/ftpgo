package ftpgo

import (
	"bufio"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//FileMatchList dir search
func FileMatchList(dir string, regexp string) (matches []string) {
	files, _ := filepath.Glob(dir + "/" + regexp)
	return files
}

//ReadFileList get row data list
func ReadFileList(filename string) []string {
	fp, err := os.Open(filename)
	var ret []string

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}

	return ret
}

//Readln readerから1行分の文字列を取得する
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool
		err      error
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

//SaveFile byte data save to file
func SaveFile(filename string, text []byte) error {
	var writer *bufio.Writer

	if _, err := os.Stat(filename); err != nil {
		if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer = bufio.NewWriter(file)
	writer.Write(text)
	writer.Flush()

	return nil
}

//FtpFile struct
type FtpFile struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
	raw   string
}

//Name get filename
func (f *FtpFile) Name() string {
	return f.name
}

//Size get filesize
func (f *FtpFile) Size() int64 {
	return f.size
}

//Mode get filemode
func (f *FtpFile) Mode() os.FileMode {
	return f.mode
}

//ModTime get updatetime of file
func (f *FtpFile) ModTime() time.Time {
	return f.mtime
}

//IsDir get dirpath
func (f *FtpFile) IsDir() bool {
	return f.mode.IsDir()
}

//Sys get sys
func (f *FtpFile) Sys() interface{} {
	return f.raw
}

var errUnknownFormat = errors.New("Unknown format")

var formatParsers = []func(line string) (*FtpFile, error){
	ParseUnixFormat,
	ParseDosFormat,
}

//NewFtpFile construct
func NewFtpFile(line string) (*FtpFile, error) {
	//log.Println(line)
	for _, f := range formatParsers {
		fileInfo, err := f(line)
		if err == errUnknownFormat {
			continue
		}
		return fileInfo, err
	}
	return nil, errUnknownFormat
}

//ParseDosDateTime file time parse for DOS
func ParseDosDateTime(input string) (dateTime time.Time, err error) {
	dateTime, err = time.Parse("01-02-06  03:04PM", input)
	if err == nil {
		return dateTime, err
	}

	dateTime, err = time.Parse("2006-01-02  15:04", input)
	return dateTime, err
}

//ParseDosFormat file time parse for DOS
func ParseDosFormat(input string) (*FtpFile, error) {
	value := input[:17]
	mtime, err := ParseDosDateTime(value)
	if err != nil {
		return nil, errUnknownFormat
	}

	var size uint64
	var mode os.FileMode

	value = input[17:]
	value = strings.TrimLeft(value, " ")
	if strings.HasPrefix(value, "<DIR>") {
		mode |= os.ModeDir
		value = strings.TrimPrefix(value, "<DIR>")
	} else {
		space := strings.Index(value, " ")
		if space == -1 {
			return nil, errUnknownFormat
		}
		size, err = strconv.ParseUint(value[:space], 10, 64)
		if err != nil {
			return nil, errUnknownFormat
		}

		value = value[space:]
	}

	name := strings.TrimLeft(value, " ")
	f := &FtpFile{
		name:  name,
		size:  int64(size),
		mode:  mode,
		mtime: mtime,
		raw:   input,
	}

	return f, nil
}

//ParseUnixFormat file time parse for UNIX
func ParseUnixFormat(input string) (*FtpFile, error) {
	var err error
	var name string
	var size uint64
	var mode os.FileMode
	var mtime time.Time

	fields := strings.Fields(input)
	if len(fields) < 9 {
		return nil, errUnknownFormat
	}

	// type
	switch fields[0][0] {
	//case '-':
	case 'd':
		mode |= os.ModeDir
	case 'l':
		mode |= os.ModeSymlink
	case 'b':
		mode |= os.ModeDevice
	case 'c':
		mode |= os.ModeCharDevice
	case 'p', '=':
		mode |= os.ModeNamedPipe
	case 's':
		mode |= os.ModeSocket
	}

	// permission
	for i := 0; i < 3; i++ {
		if fields[0][i*3+1] == 'r' {
			mode |= os.FileMode(04 << (3 * uint(2-i)))
		}
		if fields[0][i*3+2] == 'w' {
			mode |= os.FileMode(02 << (3 * uint(2-i)))
		}
		if fields[0][i*3+3] == 'x' || fields[0][i*3+3] == 's' {
			mode |= os.FileMode(01 << (3 * uint(2-i)))
		}
	}

	// size
	size, err = strconv.ParseUint(fields[4], 0, 64)
	if err != nil {
		return nil, err
	}

	// datetime
	mtime, err = ParseDateTime(fields[5:8])
	if err != nil {
		return nil, err
	}

	// name
	name = strings.Join(fields[8:], " ")

	f := &FtpFile{
		name:  name,
		size:  int64(size),
		mode:  mode,
		mtime: mtime,
		raw:   input,
	}

	return f, nil
}

//ParseDateTime parse date
func ParseDateTime(fields []string) (mtime time.Time, err error) {
	var value string
	if strings.Contains(fields[2], ":") {
		thisYear, _, _ := time.Now().Date()
		value = fields[1] + " " + fields[0] + " " + strconv.Itoa(thisYear)[2:4] + " " + fields[2] + " GMT"
	} else {
		if len(fields[2]) != 4 {
			return mtime, errors.New("Invalid year format in time string")
		}
		value = fields[1] + " " + fields[0] + " " + fields[2][2:4] + " 00:00 GMT"
	}

	mtime, err = time.Parse("_2 Jan 06 15:04 MST", value)
	return
}
