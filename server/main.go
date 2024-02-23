package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
	"github.com/pbergman/icmp-control/model"
)

func init() {
	flag.String("config", "/usr/share/icpm-control/config.conf", "the directory where config and keys will be stored")
}

func main() {

	flag.Parse()

	config, err := getConfig(flag.Lookup("config").Value.String())

	if err != nil {
		log.Fatalln(err)
	}

	var logger = GetLogger(config)

	handler, err := pcap.OpenLive(config.Interface, 1024, true, 100*time.Millisecond) //pcap.BlockForever)

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	defer handler.Close()

	if err := handler.SetBPFFilter(fmt.Sprintf("icmp[0] == %d", model.ICMPTypeRequest)); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	pkgsrc := gopacket.NewPacketSource(handler, handler.LinkType())
	pkgsrc.Lazy = true
	pkgsrc.NoCopy = true

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	var trap = make(chan os.Signal, 1)

	signal.Notify(trap, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

mainApp:
	for {
		select {
		case sign := <-trap:
			logger.Notice(fmt.Sprintf("signal %s triggered", sign.String()))
			cancel()
			break mainApp
		case packet := <-pkgsrc.PacketsCtx(ctx):

			var icpm = packet.Layer(layers.LayerTypeICMPv4)
			var ipv4 = packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
			var request = new(model.Request)

			// glue back parts for signature
			var payload = make([]byte, len(icpm.LayerPayload())+len(icpm.LayerContents()))
			copy(payload[copy(payload, icpm.LayerContents()):], icpm.LayerPayload())

			if err := request.Unmarshal(payload[4:]); err != nil {
				logger.Error("failed to unmarshall icpm request")
				logger.Error(err)
				continue
			}

			logger.Debug(fmt.Sprintf("new request (code: 0x%.2x) from %s", icpm.LayerContents()[1], ipv4.SrcIP.String()))

			var now = time.Now().UTC()
			var key = getPublicKey(request, config)

			if key == nil {
				logger.Notice(fmt.Sprintf("no key registered for %s", hex.EncodeToString(request.KeyId[:])))
				continue
			}

			if request.Time.After(now) {
				logger.Notice(fmt.Sprintf("future timestamp not allowd (time %s)", request.Time.In(time.Local).Format(time.RFC3339)))
				continue
			}

			// should be within 15 seconds for creating and can not be in future
			if request.Time.Add(15 * time.Second).Before(now) {
				logger.Notice(fmt.Sprintf("request expired (request: %s, now: %s)", request.Time.In(time.Local).Format(time.RFC3339), now.In(time.Local).Format(time.RFC3339)))
				continue
			}

			if false == ed25519.Verify(key, payload[:38], request.Signature) {
				logger.Notice("message verify failed")
				continue
			}

			wg.Add(1)
			go handle(ctx, &wg, packet, request, config, logger)
		}
	}

	wg.Wait()
}
