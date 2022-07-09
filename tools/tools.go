package tools

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"regexp"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/emersion/go-message"
)

// GetBoundary 获取邮件的边界线 boundary
func GetBoundary(header message.Header) (boundary string) {
	contentType := header.Get("Content-Type")
	boundary = strings.Split(contentType, ";")[1]
	boundary = strings.Replace(boundary, "boundary=", "", -1)
	boundary = strings.ReplaceAll(boundary, `"`, "")
	boundary = strings.Trim(boundary, " ")
	// boundary = strings.Trim(boundary, "-")
	return boundary
}

// GetSubject 获取邮件头中的
func GetSubject(header message.Header) (subject string) {
	dec := DecHeader()
	subject, err := dec.Decode(header.Get("Subject"))
	if err != nil {
		subject, _ = dec.DecodeHeader(header.Get("Subject"))
	}
	return subject
}

// GetMessageID 获取消息id
func GetMessageID(header message.Header) (messageID string) {
	return header.Get("Message-Id")
}

// GetFrom 获取邮件来源
func GetFrom(header message.Header) (from string) {
	reg := regexp.MustCompile(`\w+([-+.]*\w+@)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
	from = reg.FindAllString(header.Get("From"), -1)[0]
	return from
}

// ParseBody 解析邮件体
func ParseBody(body io.Reader) (eBody []byte, err error) {
	bodyByte, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println(err)
	}
	if bodyByte != nil {
		emailBody := string(bodyByte)
		if IsGBK(bodyByte) {
			emailBody = ConvertToString(emailBody, "gbk", "utf-8")
		}
		eBody = []byte(emailBody)
	}
	return
}

// QuotedprintableEmail 解决quotedprintable编码
func QuotedprintableEmail(body []byte) (bodyByte []byte, err error) {
	quoStr, err := ioutil.ReadAll(quotedprintable.NewReader(strings.NewReader(string(body))))
	bodyByte = []byte(quoStr)
	return
}

// multipart邮件解析
func multipartEmail(body []byte, boundary string) (emailBody []byte, err error) {
	mr := multipart.NewReader(strings.NewReader(string(body)), boundary)
	for {
		part, err := mr.NextPart() //p的类型为Part                                             -
		if err == io.EOF {
			return nil, errors.New("EOF part")
		}
		if err != nil {
			fmt.Println(err)
			return nil, errors.New("EOF part")
		}
		slurp, err := ioutil.ReadAll(part)
		if err != nil {
			fmt.Println(err)
		}
		WYdata, err := base64.StdEncoding.DecodeString(string(slurp))
		contentType := part.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			var emailData string
			if IsGBK(slurp) {
				emailData = ConvertToString(string(WYdata), "gbk", "utf-8")
			} else {
				emailData = string(WYdata)
			}
			return []byte(emailData), nil
		}
	}
}

// DecHeader 解码邮件头
func DecHeader() (dec *mime.WordDecoder) {
	dec = new(mime.WordDecoder)
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "gb2312":
			content, err := ioutil.ReadAll(input)
			if err != nil {
				return nil, err
			}
			utf8str := ConvertToString(string(content), "gbk", "utf-8")
			t := bytes.NewReader([]byte(utf8str))
			return t, nil
		case "gb18030":
			content, err := ioutil.ReadAll(input)
			if err != nil {
				return nil, err
			}

			utf8str := ConvertToString(string(content), "gbk", "utf-8")
			t := bytes.NewReader([]byte(utf8str))

			return t, nil

		case "gbk":
			content, err := ioutil.ReadAll(input)
			if err != nil {
				return nil, err
			}

			utf8str := ConvertToString(string(content), "gbk", "utf-8")
			t := bytes.NewReader([]byte(utf8str))

			return t, nil
		default:
			return nil, fmt.Errorf("unhandle charset:%s", charset)

		}
	}
	return dec
}

// ConvertToString 将字符串转为utf-8编码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// IsGBK 判断byte是否为gbk 编码
func IsGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0xff { //编码小于等于127,只有一个字节的编码，兼容ASCII码
			i++
			continue
		} else { //大于127的使用双字节编码
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}
