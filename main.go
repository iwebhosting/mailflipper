package main

import (
    "flag"
    "fmt"
    "strconv"
    "log"
    "os"
    "net/url"
    "github.com/bradfitz/go-smtpd/smtpd"
)

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

var port int
var bindaddr string
var hookurl string
var hostname string

type sources []string
var sourceFlag sources

var exitCh chan int

func (s *sources) String() string {
    return fmt.Sprint(*s)
}

func (s *sources) Set(value string) error {
    *s = append(*s, value)
    return nil
}

type mail struct {
    from smtpd.MailAddress
    to   []string
    data []byte
}

type env struct {
    *smtpd.BasicEnvelope
    m mail
}

func (e *env) Write(line []byte) error {
    e.m.data = append(e.m.data, line...)
    return nil
}

func (e *env) Close() error {
    log.Printf("Message received: %s\n", string(e.m.data[:]))
    return nil
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
    log.Printf("New mail from %q", from)
    if len(sourceFlag) > 0 {
        if !stringInSlice(from.Email(), sourceFlag) {
            // TODO: I dont think this results in something terribly graceful...!
            return nil, smtpd.SMTPError("Disallowed source address")
        }
    }
    return &env{new(smtpd.BasicEnvelope), mail{from: from}}, nil
}

func init() {
    flag.IntVar(&port, "port", 25, "the port to bind to")
    flag.StringVar(&bindaddr, "bind", "0.0.0.0", "the address to bind to")
    flag.StringVar(&hookurl, "url", "", "the webhook endpoint")
    flag.StringVar(&hostname, "hostname", "mail.local", "the hostname advertised to clients")
    flag.Var(&sourceFlag, "source", "optional source email address whitelisting. Can be set multiple times")
}

func main() {

    flag.Parse()

    addr := bindaddr + ":" + strconv.Itoa(port)
    log.Printf("Binding to %s", addr)

    if len(hookurl) < 1 {
        log.Fatal("webhook url must be defined!")
    }

    _, err := url.Parse(hookurl)
    if err != nil {
        log.Fatalf("Invalid url: %s", hookurl)
    }

    log.Printf("Webhook address is %s", hookurl)

    srv := &smtpd.Server{
        Addr: addr,
        Hostname: hostname,
        OnNewMail: onNewMail,
    }

    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("ListenAndServe: %v", err)
    }

    exitCh = make(chan int)
    for {
        select {
        case <-exitCh:
            log.Printf("Received exit signal")
            os.Exit(0)
        }
    }
}
