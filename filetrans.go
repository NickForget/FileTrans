package  filetrans

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/saintfish/chardet"
	"github.com/axgle/mahonia"
)

const FILEBUFLEN  = 1024

type FileTrans struct {
	SrcPath string
	DestPath string
	DestCharset string
}

func NewFileTrans(srcPath string, destPath string, destcharset string) *FileTrans{
	indexSrc := strings.Index(srcPath, "./")

	if 0 == indexSrc {
		srcPath =  string([]byte(srcPath)[len("./") : ])
	}

	indexDest := strings.Index(destPath, "./")

	if 0 == indexDest{
		destPath =  string([]byte(destPath)[len("./") : ])
	}

	return &FileTrans{
		SrcPath : srcPath,
		DestPath : destPath,
		DestCharset : destcharset,
	}
}

func (this *FileTrans)FileHandle(path string, info os.FileInfo, err error) error{
	if info == nil {
		return err
	}

	// 如果是文件的话
	if !info.IsDir() {
		// 转换目录
		path := strings.Replace(path, "\\", "/", -1)
		destNewPath := strings.Replace(path, this.SrcPath, this.DestPath, -1)

		// 拷贝文件
		this.CopyFile(path, destNewPath)
	}
	return nil
}

func (this *FileTrans)CopyDir() error {
	// 检测srcPath目录正确性
	srcInfo, err := os.Stat(this.SrcPath)

	if nil != err {
		return err
	}

	if !srcInfo.IsDir() {
		err := errors.New(this.SrcPath + "is not dir")
		return err
	}

	// 创建destPath目录
	err = os.MkdirAll(this.DestPath, os.ModePerm)
	if nil != err{
		return err
	}

	// 遍历指定目录下的所有文件以及拷贝
	err = filepath.Walk(this.SrcPath, this.FileHandle)

	if err != nil {
		return err
	}

	return nil
}

// 生成目录并拷贝文件
func (this *FileTrans)CopyFile(src string, dest string) (error) {
	// 判断源文件编码集
	charset, err := this.GetFileCharset(src)

	if nil != err{
		return err
	}
	fmt.Println(charset)

	// 打开src文件
	srcFile, err := os.Open(src)

	if err != nil {
		return err
	}

	// 关闭文件
	defer srcFile.Close()

	// 获取destPath目录
	destPath :=  string([]byte(dest)[0 : strings.LastIndexAny(dest, "/")])

	// 创建destPath目录
	err = os.MkdirAll(destPath, os.ModePerm)
	if nil != err{
		return err
	}

	// 创建文件
	dstFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 申请缓存
	filebuf := make([]byte, FILEBUFLEN)

	// 拷贝文件
	for{
		len, _ := srcFile.Read(filebuf);

		if 0 == len{
			break
		}
		deststr, err := this.ConvertToString(string(filebuf[0:len]), charset, "UTF-8", )

		if nil != err{
			return err
		}

		dstFile.Write([]byte(deststr))
	}

	return err
}

func (this *FileTrans)GetFileCharset(filepath string)(string, error){
	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}

	// 关闭文件
	defer file.Close()

	// 读取文件
	buffer := make([]byte, 32 << 10)
	size, _ := io.ReadFull(file, buffer)

	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(buffer[0:size])

	if nil == err {
		return result.Charset, err
	}
	return "", err
}

func (this *FileTrans)ConvertToString(src string, srcCode string, tagCode string) (string, error) {
	// 创建源字符集编码器
	srcCoder := mahonia.NewDecoder(srcCode)

	// 判断是否成功
	if nil == srcCoder {
		return "", errors.New("Src NewDecoder Err")
	}

	// 将源字符编码成UTF-8
	srcUTF8 := srcCoder.ConvertString(src)

	// 创建目标字符集解码器
	tagCoder := mahonia.NewEncoder(tagCode)

	// 判断是否成功
	if nil == tagCoder {
		return "", errors.New("Tag NewEncoder Err")
	}

	// 将UTF8转换成目标字符集
	tag := tagCoder.ConvertString(srcUTF8)

	return tag, nil
}
