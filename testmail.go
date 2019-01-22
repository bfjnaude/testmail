package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

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

const path = string("562.jpg")

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
	buf.WriteString("Content-Type: multipart/mixed; boundary=\"" + writer.Boundary() + "\"\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\n")
	buf.WriteString("\n")
	buf.WriteString("--" + writer.Boundary() + "\n")

	// _, err = writer.CreatePart(map[string][]string{"Content-Type": []string{"multipart/mixed; boundary=\"" + writer.Boundary() + "\""}})
	// _, err = writer.CreatePart(map[string][]string{"Content-Type": []string{"multipart/mixed;"}})
	// _, err = writer.CreatePart(map[string][]string{})

	// if err != nil {
	// 	panic(err)
	// }

	part, err := writer.CreateFormField("test")

	if err != nil {
		panic(err)
	}

	testString := "random garbage\n"

	b := new(bytes.Buffer)
	b.WriteString(testString)

	part.Write(b.Bytes())

	// if err != nil {
	// 	panic(err)
	// }

	imgpart, err := writer.CreateFormFile("file", filepath.Base(path))

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
