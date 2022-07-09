package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
)

const chunkSize = 1024 * 1000 * 100

//const chunkSize = 50000000
var (
	action  string
	infile  string
	outfile string
)

func split(infile string) {
	if infile == "" {
		panic("请输入正确的文件名")
	}

	fileInfo, err := os.Stat(infile)
	if err != nil {
		if os.IsNotExist(err) {
			panic("文件不存在")
		}
		panic(err)
	}

	num := math.Ceil(float64(fileInfo.Size()) / chunkSize)

	fi, err := os.OpenFile(infile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fileInfo.Name())
	fmt.Printf("要拆分成%.0f份\n", num)
	b := make([]byte, chunkSize)
	var i int64 = 1
	for ; i <= int64(num); i++ {
		fi.Seek((i-1)*chunkSize, 0)
		if len(b) > int(fileInfo.Size()-(i-1)*chunkSize) {
			b = make([]byte, fileInfo.Size()-(i-1)*chunkSize)
		}
		fi.Read(b)
		ofile := fmt.Sprintf("./%s_sp_%d.part", fileInfo.Name(), i)
		fmt.Printf("生成%s\n", ofile)
		f, err := os.OpenFile(ofile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		f.Write(b)
		f.Close()
	}
	fi.Close()
	fmt.Println("拆分完成")

}
func merge(outfile string) {
	fii, err := os.OpenFile(outfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
		return
	}
	part_list, err := filepath.Glob(fmt.Sprintf("./%s_sp_*.part", outfile))
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("要把%v份合并成一个文件%s\n", part_list, outfile)
	i := 0
	for _, v := range part_list {
		f, err := os.OpenFile(v, os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(err)
			return
		}
		fii.Write(b)
		f.Close()
		i++
		fmt.Printf("合并%d个\n", i)
	}
	fii.Close()
	fmt.Println("合并成功")
}
func main() {
	flag.StringVar(&action, "a", "split", "请输入用途：split/merge 默认是split")
	flag.StringVar(&infile, "f", "", "需要切割文件名")
	flag.StringVar(&outfile, "o", "xxx.zip", "请输入要合并的文件")
	flag.Parse()
	if action == "split" {
		split(infile)
	} else if action == "merge" {
		merge(outfile)
	} else {
		panic("-a只能输入split/merge")
	}
}
