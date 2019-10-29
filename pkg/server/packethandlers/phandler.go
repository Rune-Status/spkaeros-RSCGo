package packethandlers

import (
	"bitbucket.org/zlacki/rscgo/pkg/server/clients"
	"bitbucket.org/zlacki/rscgo/pkg/server/config"
	"bitbucket.org/zlacki/rscgo/pkg/server/log"
	"bitbucket.org/zlacki/rscgo/pkg/server/packetbuilders"
	"github.com/BurntSushi/toml"
)

func init() {
	PacketHandlers["pingreq"] = func(c clients.Client, p *packetbuilders.Packet) {
		c.SendPacket(packetbuilders.ResponsePong)
	}
}

//HandlerFunc Represents a function for handling incoming packetbuilders.
type HandlerFunc func(clients.Client, *packetbuilders.Packet)

//PacketHandlers A map with descriptive names for the keys, and functions to run for the value.
var PacketHandlers = make(map[string]HandlerFunc)

//packetHandler Definition of a packet handler.
type packetHandler struct {
	Opcode int    `toml:"opcode"`
	Name   string `toml:"name"`
	//	Handler HandlerFunc
}

//packetHandlerTable Represents a mapping of descriptive names to packet opcodes.
type packetHandlerTable struct {
	Handlers []packetHandler `toml:"packets"`
}

var table packetHandlerTable

//GetPacketHandler Returns the packet handler function assigned to this opcode.  If it can't be found, returns nil.
func GetPacketHandler(opcode byte) HandlerFunc {
	for _, handler := range table.Handlers {
		if byte(handler.Opcode) == opcode {
			return PacketHandlers[handler.Name]
		}
	}
	return nil
}

//CountPacketHandlers returns the number of packet handlers currently defined.
func CountPacketHandlers() int {
	return len(table.Handlers)
}

//Initialize Deserializes the packet handler table into memory.
func Initialize() {
	if _, err := toml.DecodeFile(config.DataDir()+config.PacketHandlers(), &table); err != nil {
		log.Error.Fatalln("Could not open packet handler table data file:", err)
		return
	}
}