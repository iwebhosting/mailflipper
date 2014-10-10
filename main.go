package main

import (
    "flag"
    "strconv"
    "log"
    "os"
    "net"
    "net/url"
    "github.com/bradfitz/go-smtpd/smtpd"
)

var port int
var bindaddr string
var hookurl string
var hostname string

var exitCh chan int

func init() {
    flag.IntVar(&port, "port", 25, "the port to bind to")
    flag.StringVar(&bindaddr, "bind", "0.0.0.0", "the address to bind to")
    flag.StringVar(&hookurl, "url", "", "the webhook endpoint")
    flag.StringVar(&hostname, "hostname", "mail.local", "the hostname advertised to clients")
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
        OnNewMail: nil,
    }

    ln, e := net.Listen("tcp", addr)

    if e != nil {
        log.Fatalf("Unable to bind to %s: %s", addr, e)
    }

    if err != nil {
        log.Fatalf("Error starting server: %s", err)
    }

    go srv.Serve(ln)

    exitCh = make(chan int)
    for {
        select {
        case <-exitCh:
            log.Printf("Received exit signal")
            os.Exit(0)
        }
    }
}
