/**
 * @Author: Zachariah Knight <zach>
 * @Date:   08-20-2019
 * @Email:  aeros.storkpk@gmail.com
 * @Project: RSCGo
 * @Last modified by:   zach
 * @Last modified time: 08-22-2019
 * @License: Use of this source code is governed by the MIT license that can be found in the LICENSE file.
 * @Copyright: Copyright (c) 2019 Zachariah Knight <aeros.storkpk@gmail.com>
 */

package server

import (
	"fmt"
	"strconv"

	"bitbucket.org/zlacki/rscgo/pkg/entity"
	"bitbucket.org/zlacki/rscgo/pkg/server/packets"
	"bitbucket.org/zlacki/rscgo/pkg/strutil"
)

//CommandHandlers A map to assign in-game commands to the functions they should execute.
var CommandHandlers = make(map[string]func(*Client, []string))

func init() {
	CommandHandlers["say"] = func(c *Client, args []string) {
		if len(args) < 1 {
			c.outgoingPackets <- packets.ServerMessage("@que@Invalid args.  Usage: /say <msg>")
			return
		}
		msg := "@whi@[@cya@GLOBAL@whi@] @yel@" + c.player.Username + "@whi@:@yel@"
		for _, arg := range args {
			msg += " " + arg
		}
		for _, v := range ClientList.Values {
			if c1, ok := v.(*Client); ok && c1 != nil {
				c1.outgoingPackets <- packets.ServerMessage(fmt.Sprintf("@que@%s", msg))
			}
		}
	}
	CommandHandlers["tele"] = teleport
	CommandHandlers["teleport"] = teleport
	CommandHandlers["death"] = func(c *Client, args []string) {
		c.outgoingPackets <- packets.Death
	}
	Handlers[174] = func(c *Client, p *packets.Packet) {
		//		for _, p1 := range c.player.NearbyPlayers() {
		//			if c1, ok := ClientList.Get(p1.Index).(*Client); c1 != nil && ok {
		//				c1.outgoingPackets <- packets.TeleBubble(diffX, diffY)
		//			}
		//		}
		// TODO: Send message to other players.
		LogInfo.Printf("[CHAT] %v: '%v'", c.player.Username, strutil.FormatChatMessage(strutil.UnpackChatMessage(p.Payload)))
	}
	Handlers[120] = func(c *Client, p *packets.Packet) {
		args := strutil.ModalParse(string(p.Payload))
		handler, ok := CommandHandlers[args[0]]
		if !ok {
			c.outgoingPackets <- packets.ServerMessage("@que@Invalid command.")
			LogInfo.Printf("[COMMAND] %v sent invalid command: /%v\n", c.player.Username, string(p.Payload))
			return
		}
		LogInfo.Printf("[COMMAND] %v: /%v\n", c.player.Username, string(p.Payload))
		handler(c, args[1:])
	}
	Handlers[84] = func(c *Client, p *packets.Packet) {
		index, _ := p.ReadShort()
		c.player.Appearances = append(c.player.Appearances, int(index))
	}
}

func teleport(c *Client, args []string) {
	if len(args) < 2 {
		c.outgoingPackets <- packets.ServerMessage("@que@Invalid args.  Usage: /tele <x> <y>")
		return
	}
	x, _ := strconv.Atoi(args[0])
	y, _ := strconv.Atoi(args[1])
	if x >= entity.MaxX || y >= entity.MaxY || x < 0 || y < 0 {
		c.outgoingPackets <- packets.ServerMessage(fmt.Sprintf("@que@Invalid coordinates.  Must be between 0,0 and %v,%v", entity.MaxX, entity.MaxY))
		return
	}
	newLocation := entity.NewLocation(x, y)
	LogInfo.Printf("Teleporting %v from %v to %v\n", c.player.Username, c.player.Location(), newLocation)
	c.outgoingPackets <- packets.TeleBubble(0, 0)
	for _, p1 := range c.player.NearbyPlayers() {
		diffX := c.player.X() - p1.X()
		diffY := c.player.Y() - p1.Y()
		if c1, ok := ClientList.Get(p1.Index).(*Client); c1 != nil && ok {
			c1.outgoingPackets <- packets.TeleBubble(diffX, diffY)
		}
	}
	c.player.Removing = true
	c.player.SetLocation(newLocation)
}
