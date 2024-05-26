package common

import (
	"io"
	"os"
)

func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}

	err = out.Sync()
	return
}

func IndexOf(s string, sub string) int {
	// 将字符串转换为 rune 数组
	runes := []rune(s)
	subRunes := []rune(sub)

	// 查找子字符串的第一个 rune
	for i, r := range runes {
		if r == subRunes[0] {
			// 检查子字符串是否匹配
			for j, sr := range subRunes {
				if r != sr {
					break
				}
				if j == len(subRunes)-1 {
					return i
				}
			}
		}
	}
	return -1
}
