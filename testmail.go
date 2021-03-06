package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

var awsSession *session.Session
var sesService *ses.SES

func init() {
	log.SetFlags(log.Lshortfile)
	awsSession = session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
	}))

	sesService = ses.New(awsSession)
}

// const path = string("562.jpg")
const path = string("label-50.pdf")

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func sendMail() {
	file, err := os.Open(path)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	buf := new(bytes.Buffer)

	buf.WriteString("To: francois@uafrica.com\n")
	buf.WriteString("From: francois@uafrica.com\n")
	buf.WriteString("Subject: test raw mail\n")
	buf.WriteString("MIME-Version: 1.0\n")

	writer := multipart.NewWriter(buf)
	buf.WriteString("Content-Type: multipart/alternative; boundary=\"" + writer.Boundary() + "\"\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\n")
	buf.WriteString("\n")

	htmlpart, err := writer.CreatePart(map[string][]string{
		"Content-Type":              []string{"text/html; charset=\"UTF-8\""},
		"Content-Transfer-Encoding": []string{"quoted-printable"},
	})

	htmlbuf := new(bytes.Buffer)
	htmlbuf.WriteString("<html><body>Test</body></html>\n\n")
	htmlpart.Write(htmlbuf.Bytes())

	textpart, err := writer.CreatePart(map[string][]string{
		"Content-Type": []string{"text/plain; charset=\"UTF-8\""},
	})

	textbuf := new(bytes.Buffer)
	textbuf.WriteString("Test\n\n")
	textpart.Write(textbuf.Bytes())

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes("file"), escapeQuotes(filepath.Base(path))))
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Content-Transfer-Encoding", "base64")
	imgpart, err := writer.CreatePart(h)
	if err != nil {
		panic(err)
	}

	encoder := base64.NewEncoder(base64.StdEncoding, imgpart)
	_, err = io.Copy(encoder, file)

	if err != nil {
		panic(err)
	}

	err = writer.Close()

	if err != nil {
		panic(err)
	}

	msg := ses.RawMessage{
		Data: buf.Bytes(),
	}

	addr := "francois@uafrica.com"

	emailInput := ses.SendRawEmailInput{
		Source: &addr,
		Destinations: []*string{
			&addr,
		},
		RawMessage: &msg,
	}

	sendResult, err := sesService.SendRawEmail(&emailInput)

	if err != nil {
		panic(err)
	}

	log.Println(sendResult)
}

func main() {
	sendMail()
}
