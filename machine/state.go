package machine

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
)

const bootIdPath = "/proc/sys/kernel/random/boot_id"

// MachineState represents a point-in-time snapshot of the
// state of the local host.
type MachineState struct {
	BootId   string
	PublicIP string
	Metadata map[string]string
}

func (ms MachineState) String() string {
	return fmt.Sprintf("MachineState{BootId: %q, PublicIp: %q, Metadata: %v}", ms.BootId, ms.PublicIP, ms.Metadata)
}

// NewDynamicMachineState generates a MachineState object with
// the values read from the local system
func CurrentState() MachineState {
	bootId := readLocalBootId()
	publicIP := getLocalIP()
	return MachineState{bootId, publicIP, make(map[string]string, 0)}
}

func readLocalBootId() string {
	id, err := ioutil.ReadFile(bootIdPath)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(id))
}

func getLocalIP() string {
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		return ""
	}

	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		return ""
	}

	for _, addr := range addrs {
		// Attempt to parse the address in CIDR notation
		// and assert it is IPv4
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil {
			return ip.String()
		}
	}

	return ""
}

func stackState(top, bottom MachineState) MachineState {
	state := MachineState(bottom)

	if top.PublicIP != "" {
		state.PublicIP = top.PublicIP
	}

	if top.BootId != "" {
		state.BootId = top.BootId
	}

	//FIXME: This will *always* overwrite the bottom's metadata,
	// but the only use-case we have today does not ever have
	// metadata on the bottom.
	if len(top.Metadata) > 0 {
		state.Metadata = top.Metadata
	}

	return state
}
