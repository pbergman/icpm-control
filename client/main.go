package main

import (
	"encoding/hex"
	"flag"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"golang.org/x/net/icmp"
)

var (
	dump   bool
	config string
)

func init() {
	usr, err := user.Current()

	if err != nil {
		log.Fatalln(err)
	}

	flag.BoolVar(&dump, "dump", false, "dump payload")
	flag.StringVar(&config, "config", filepath.Join(usr.HomeDir, "/.config/icpm-control"), "The code to be executed on remote")
}

func main() {

	flag.Parse()

	var host = flag.Arg(0)

	if "" == host || "" == flag.Arg(1) {
		log.Fatalf("usage: %s <host> <code>", os.Args[0])
	}

	code, err := strconv.Atoi(flag.Arg(1))

	if err != nil {
		log.Fatalf("could not parse code argument: %s", err.Error())
	}

	if err := initConfigDir(config); err != nil {
		log.Fatalln(err)
	}

	public, private, err := readKeys(config)

	if err != nil {
		log.Fatal(err)
	}

	payload, err := createRequest(public, private, uint(code))

	if err != nil {
		log.Fatal(err)
	}

	if dump {
		var dumper = hex.Dumper(os.Stdout)
		_, _ = dumper.Write(payload)
		_, _ = dumper.Write([]byte("\n"))
		_ = dumper.Close()
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")

	if err != nil {
		log.Fatalf("listen err, %s", err)
	}

	defer conn.Close()

	if _, err := conn.WriteTo(payload, &net.IPAddr{IP: net.ParseIP(host)}); err != nil {
		log.Fatalf("write err, %s", err)
	}

	log.Println("icpm package successfully send")
}
