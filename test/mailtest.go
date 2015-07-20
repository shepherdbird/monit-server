package main

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
)

func IsFile(file string) bool {
	f, e := os.Stat(file)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func excute(command string) string {
	cmd := exec.Command("/bin/sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return out.String()
}

func main() {
	/*
		fmt.Println(excute("cat /var/run/syslog.pid"))
		if excute("cat /var/log/syslog.pid") == "" {
			fmt.Println("file")
		} else {
			fmt.Println("no file")
		}
	*/
	SendEmail(
		"smtp.126.com",
		25,
		"wonderflow@126.com",
		"zjuvlis123456",
		[]string{"wonderflow@zju.edu.cn"},
		"testing subject",
		"<html><body>Exception 1</body></html>Exception 1")

}

func catchPanic(err *error, functionName string) {
	if r := recover(); r != nil {
		fmt.Printf("%s : PANIC Defered : %v\n", functionName, r)

		// Capture the stack trace
		buf := make([]byte, 10000)
		runtime.Stack(buf, false)

		fmt.Printf("%s : Stack Trace : %s", functionName, string(buf))

		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	} else if err != nil && *err != nil {
		fmt.Printf("%s : ERROR : %v\n", functionName, *err)

		// Capture the stack trace
		buf := make([]byte, 10000)
		runtime.Stack(buf, false)

		fmt.Printf("%s : Stack Trace : %s", functionName, string(buf))
	}
}

func SendEmail(host string, port int, userName string, password string, to []string, subject string, message string) (err error) {
	defer catchPanic(&err, "SendEmail")

	parameters := struct {
		From    string
		To      string
		Subject string
		Message string
	}{
		userName,
		strings.Join([]string(to), ","),
		subject,
		message,
	}

	buffer := new(bytes.Buffer)

	template := template.Must(template.New("emailTemplate").Parse(emailScript()))
	template.Execute(buffer, &parameters)

	auth := smtp.PlainAuth("", userName, password, host)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", host, port),
		auth,
		userName,
		to,
		buffer.Bytes())

	return err
}

func emailScript() (script string) {
	return `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

{{.Message}}`
}
