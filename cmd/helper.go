package cmd

import (
	"bufio"
	"fmt"
	sdkcode "github.com/solarfs/go-chainstorage-sdk/code"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// sema is a counting semaphore for limiting concurrency in dirEntries.
var sema = make(chan struct{}, 20)

func Error(cmd *cobra.Command, args []string, err error) {
	log.Errorf("execute %s args:%v error:%v\n", cmd.Name(), args, err)
	fmt.Fprintf(os.Stderr, "execute %s args:%v error:%v\n", cmd.Name(), args, err)
	os.Exit(1)
}

func GetBucketName(args []string) string {
	bucketName := ""
	if len(args) == 0 {
		return bucketName
	}

	//bucketPrefix := viper.GetString("cli.bucketPrefix")
	bucketPrefix := cliConfig.BucketPrefix

	for i := range args {
		arg := args[i]
		if strings.HasPrefix(arg, bucketPrefix) {
			bucketName = strings.TrimPrefix(arg, bucketPrefix)
			break
		}
	}

	return bucketName
}

func GetDataPath(args []string) string {
	dataPath := ""
	if len(args) == 0 {
		return dataPath
	}

	//bucketPrefix := viper.GetString("cli.bucketPrefix")
	bucketPrefix := cliConfig.BucketPrefix

	for i := range args {
		arg := args[i]
		if strings.HasPrefix(arg, bucketPrefix) {
			continue
		}

		if _, err := os.Stat(arg); !os.IsNotExist(err) {
			return arg
		}
	}

	return dataPath
}

// 检查桶名称
func checkBucketName(bucketName string) error {
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return sdkcode.ErrInvalidBucketName
	}

	// 桶名称异常，名称范围必须在 3-63 个字符之间并且只能包含小写字符、数字和破折号，请重新尝试
	isMatch := regexp.MustCompile(`^[a-z0-9-]*$`).MatchString(bucketName)
	if !isMatch {
		return sdkcode.ErrInvalidBucketName
	}

	return nil
}

// 检查对象名称
func checkObjectName(objectName string) error {
	if len(objectName) == 0 || len(objectName) > 255 {
		return sdkcode.ErrInvalidObjectName
	}

	isMatch := regexp.MustCompile("[<>:\"/\\|?*\u0000-\u001F]").MatchString(objectName)
	if isMatch {
		return sdkcode.ErrInvalidObjectName
	}

	isMatch = regexp.MustCompile(`^(con|prn|aux|nul|com\d|lpt\d)$`).MatchString(objectName)
	if isMatch {
		return sdkcode.ErrInvalidObjectName
	}

	if objectName == "." || objectName == ".." {
		return sdkcode.ErrInvalidObjectName
	}

	return nil
}

func GetTimestampString() string {
	timestampString := time.Now().Format("2006-01-02 15:04:05.000000000") //当前时间的字符串，2006-01-02 15:04:05据说是golang的诞生时间，固定写法
	return timestampString
}

func isFolderNotEmpty(path string) (bool, error) {
	// Check if the path is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if !fileInfo.IsDir() {
		return false, nil
	}

	// Open the directory
	dir, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// Read the directory entries
	_, err = dir.Readdirnames(1)
	if err == nil {
		// Directory is not empty
		return true, nil
	} else if err == io.EOF {
		// Directory is empty
		return false, nil
	} else {
		// An error occurred while reading the directory
		return false, err
	}
}

func getFolderSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			size += fileInfo.Size()
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return size, nil
}

//func main() {
//	path := "/path/to/folder"
//
//	size, err := folderSize(path)
//	if err != nil {
//		fmt.Printf("Error: %v\n", err)
//		return
//	}
//
//	fmt.Printf("Folder size: %d bytes\n", size)
//}

func printFileContent(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Fprintln(os.Stdout, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func convertSizeUnit(size int64) string {
	// 2^10 = 1024
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}

	div, exp := int64(unit), 0
	for size >= div && exp < 8 {
		div *= unit
		exp++
	}

	convertedSize := float64(size) / float64(div/unit)
	//fmt.Printf("size:%d\n", size)
	//fmt.Printf("div:%d\n", div)
	return fmt.Sprintf("%.1f%cB", convertedSize, "KMGTPEZY"[exp-1])
}

// 获取上传数据使用量
func getUploadingDataUsage(dataPath string) (int64, int64, error) {
	var totalSize int64
	var fileAmount int64

	// 数据路径为空
	if len(dataPath) == 0 {
		return 0, 0, sdkcode.ErrCarUploadFileInvalidDataPath
	}

	// 数据路径无效
	fileInfo, err := os.Stat(dataPath)
	if os.IsNotExist(err) {
		return 0, 0, sdkcode.ErrCarUploadFileInvalidDataPath
	} else if err != nil {
		log.WithError(err).
			WithField("dataPath", dataPath).
			Error("fail to return stat of file")
		return 0, 0, err
	}

	if !fileInfo.IsDir() {
		fileAmount++
		totalSize = fileInfo.Size()
		return fileAmount, totalSize, nil
	}

	fileSizes := make(chan int64)
	var wg sync.WaitGroup
	wg.Add(1)
	go walkDir(dataPath, &wg, fileSizes)

	go func() {
		wg.Wait()
		close(fileSizes)
	}()

	for {
		size, ok := <-fileSizes
		if !ok {
			break // fileSizes was closed
		}

		fileAmount++
		totalSize += size
	}

	return fileAmount, totalSize, nil
}

func walkDir(dir string, wg *sync.WaitGroup, fileSizes chan<- int64) {
	defer wg.Done()
	for _, entry := range dirEntries(dir) {
		if entry.IsDir() {
			wg.Add(1)
			subDir := filepath.Join(dir, entry.Name())
			go walkDir(subDir, wg, fileSizes)
		} else {
			fileInfo, err := entry.Info()
			if err != nil {
				log.WithError(err).
					WithField("dir", dir).
					Error("fail to return stat of file")
				return
			}

			fileSizes <- fileInfo.Size()
		}
	}
}

// dirEntries returns the entries of directory dir.
func dirEntries(dir string) []os.DirEntry {
	sema <- struct{}{}        // acquire token
	defer func() { <-sema }() // release token

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.WithError(err).
			WithField("dir", dir).
			Error("fail to read dir")
		return nil
	}

	return entries
}

func bytesToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}
