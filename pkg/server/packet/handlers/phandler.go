/*
 * Copyright (c) 2020 Zachariah Knight <aeros.storkpk@gmail.com>
 *
 * Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 */

package handlers

import (
	"github.com/BurntSushi/toml"
	"github.com/spkaeros/rscgo/pkg/server/config"
	"github.com/spkaeros/rscgo/pkg/server/log"
	"github.com/spkaeros/rscgo/pkg/server/packet"
	"github.com/spkaeros/rscgo/pkg/server/world"
)

//handlerFunc Represents a func that is to be called whenever a connected client receives
// a specific incoming packet.
type handlerFunc = func(*world.Player, *packet.Packet)

//handlers A map with descriptive names for the keys, and functions to run for the value.
var handlers = make(map[string]handlerFunc)

//definitions a collection of packet definitions.
var definitions packetList

//packetDefinition Definition of a packet handler.
type packetDefinition struct {
	Opcode int    `toml:"opcode"`
	Name   string `toml:"name"`
	//	Handler handlerFunc
}

//packetList Represents a mapping of descriptive names to packet opcodes.
type packetList struct {
	Set []packetDefinition `toml:"packets"`
}

func init() {
	// Just to prevent non-handled packet message from spamming up the logs
	AddHandler("pingreq", handlerFunc(nil))
}

//UnmarshalPackets Loads the packet definitions into memory from the configured TOML file
func UnmarshalPackets() {
	if _, err := toml.DecodeFile(config.DataDir()+config.PacketHandlers(), &definitions); err != nil {
		log.Error.Fatalln("Could not open packet handler definitions data file:", err)
		return
	}
}

//Handler Returns the packet handler function assigned to this opcode.  If it can't be found, returns nil.
func Handler(opcode byte) handlerFunc {
	for _, h := range definitions.Set {
		if byte(h.Opcode) == opcode {
			return handlers[h.Name]
		}
	}
	return nil
}

//AddHandler Adds and assigns the packet handler to the packet with the specified name.
func AddHandler(name string, h handlerFunc) {
	if _, ok := handlers[name]; ok {
		log.Warning.Printf("Attempted to bind a handler to packet '%v' which is already handled elsewhere.  Ignoring bind.", name)
		return
	}
	handlers[name] = h
}

//PacketCount returns the number of packet definitions
func PacketCount() int {
	return len(definitions.Set)
}

//HandlerCount returns the number of definitions that are handled
func HandlerCount() int {
	return len(handlers)
}