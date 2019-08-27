/**
 * @Author: Zachariah Knight <zach>
 * @Date:   08-20-2019
 * @Email:  aeros.storkpk@gmail.com
 * @Project: RSCGo
 * @Last modified by:   zach
 * @Last modified time: 08-27-2019
 * @License: Use of this source code is governed by the MIT license that can be found in the LICENSE file.
 * @Copyright: Copyright (c) 2019 Zachariah Knight <aeros.storkpk@gmail.com>
 */

package server

import (
	"bitbucket.org/zlacki/rscgo/pkg/server/packets"
	"bitbucket.org/zlacki/rscgo/pkg/strutil"
)

func init() {
	Handlers[32] = sessionRequest
	Handlers[0] = loginRequest
	Handlers[145] = func(c *Client, p *packets.Packet) {
		c.outgoingPackets <- packets.Logout
		c.kill <- struct{}{}
	}
}

func sessionRequest(c *Client, p *packets.Packet) {
	c.uID, _ = p.ReadByte()
	seed := GenerateSessionID()
	c.isaacSeed[1] = seed
	c.outgoingPackets <- packets.NewBarePacket(nil).AddLong(seed)
}

func loginRequest(c *Client, p *packets.Packet) {
	// TODO: Handle reconnect slightly different
	p.ReadByte()
	version, _ := p.ReadInt()
	if version != uint32(Version) {
		if len(Flags.Verbose) >= 1 {
			LogWarning.Printf("Player tried logging in with invalid client version. Got %d, expected %d\n", version, Version)
		}
		c.sendLoginResponse(5)
		return
	}
	seed := make([]uint64, 2)
	for i := 0; i < 2; i++ {
		seed[i], _ = p.ReadLong()
	}
	cipher := c.SeedISAAC(seed)
	if cipher == nil {
		c.sendLoginResponse(5)
		return
	}
	c.isaacStream = cipher
	c.player.Index = c.index
	c.player.Username, _ = p.ReadString()
	hash := strutil.Base37(c.player.Username)
	c.player.Username = strutil.DecodeBase37(hash)
	password, _ := p.ReadString()
	passHash := HashPassword(password)
	//	entity.GetRegion(c.player.X(), c.player.Y()).AddPlayer(c.player)
	if _, ok := Clients[hash]; ok {
		c.sendLoginResponse(4)
		return
	}
	c.sendLoginResponse(byte(c.LoadPlayer(c.player.Username, passHash)))
}
