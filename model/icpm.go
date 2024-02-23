package model

import "golang.org/x/net/ipv4"

// unused ipv4 icpm type see tha should be save to use
// https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml#icmp-parameters-codes-unassigned
const (
	ICMPTypeRequest   ipv4.ICMPType = 64
	ICMPTypeHeartbeat ipv4.ICMPType = 65
	ICMPTypeResponse  ipv4.ICMPType = 66
)
