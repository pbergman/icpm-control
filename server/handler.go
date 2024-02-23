package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/pbergman/icmp-control/model"
	"github.com/pbergman/logger"
)

func getTemplateCtx(packet gopacket.Packet, config *Config, request *model.Request, script *Script) map[string]interface{} {
	var ctx = make(map[string]interface{})
	var ipv4 = packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	var eth = packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
	for k, v := range config.Vars {
		ctx[k] = v
	}
	for k, v := range script.Vars {
		ctx[k] = v
	}
	ctx["src_ip"] = ipv4.SrcIP
	ctx["src_mac"] = eth.SrcMAC
	ctx["src_vlan"] = ""
	ctx["dst_ip"] = ipv4.DstIP
	ctx["dst_mac"] = eth.DstMAC
	ctx["args"] = make([]byte, len(request.Args))
	if x := packet.Layer(layers.LayerTypeDot1Q); x != nil {
		ctx["src_vlan"] = x.(*layers.Dot1Q).VLANIdentifier
	}
	copy(ctx["args"].([]byte), request.Args[:])
	return ctx
}

func findActiveScript(code uint8, src string, dst string, scripts []*Script) *Script {
	var ret *Script
	for _, script := range scripts {
		if script.Code != code {
			continue
		}
		if script.Src != "" && script.Src != src {
			continue
		}
		if script.Dst != "" && script.Dst != dst {
			continue
		}
		if ret != nil && (script.Src == "" && script.Dst == "") {
			continue
		}
		ret = script
	}
	return ret
}

func handle(ctx context.Context, wg *sync.WaitGroup, packet gopacket.Packet, request *model.Request, config *Config, logger *logger.Logger) {

	defer wg.Done()

	var ipv4 = packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	var icmp = packet.Layer(layers.LayerTypeICMPv4).(*layers.ICMPv4)
	var script = findActiveScript(icmp.LayerContents()[1], ipv4.SrcIP.String(), ipv4.DstIP.String(), config.Scripts)

	if script == nil {
		logger.Notice(fmt.Sprintf("no script found for code 0x%.2x (src: %s dst: %s)", icmp.LayerContents()[1], ipv4.SrcIP.String(), ipv4.DstIP.String()))
		return
	}

	cmd, in := StartShell(ctx, script, logger)

	if nil == cmd {
		return
	}

	if err := script.Exec.Execute(in, getTemplateCtx(packet, config, request, script)); err != nil {
		logger.Error(err)
		return
	}

	if err := in.Close(); err != nil {
		logger.Error(err)
		return
	}

	if err := cmd.Wait(); err != nil {
		logger.Error(err)
	}
}
