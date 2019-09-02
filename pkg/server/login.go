package server

import (
	"crypto/rand"
	"crypto/rsa"

	"bitbucket.org/zlacki/rscgo/pkg/server/packets"
	"bitbucket.org/zlacki/rscgo/pkg/strutil"
)

func init() {
	PacketHandlers["sessionreq"] = sessionRequest
	PacketHandlers["loginreq"] = loginRequest
	PacketHandlers["logoutreq"] = logout
}

func logout(c *Client, p *packets.Packet) {
	c.outgoingPackets <- packets.Logout
	close(c.kill)
	//c.kill <- struct{}{}
}

func sessionRequest(c *Client, p *packets.Packet) {
	c.uID, _ = p.ReadByte()
	seed := GenerateSessionID()
	c.isaacSeed[1] = seed
	c.outgoingPackets <- packets.NewBarePacket(nil).AddLong(seed)
}

func loginRequest(c *Client, p *packets.Packet) {
	// Login block encrypted with block cipher using shared secret, to send/recv credentials and stream cipher key
	buf, err := rsa.DecryptPKCS1v15(rand.Reader, RsaKey, p.Payload)
	if err != nil {
		LogWarning.Printf("Could not decrypt RSA login block: `%v`\n", err.Error())
		c.sendLoginResponse(9)
		return
	}
	p.Payload = buf
	// TODO: Handle reconnect slightly different
	p.ReadByte()
	version, _ := p.ReadInt()
	if version != uint32(TomlConfig.Version) {
		if len(Flags.Verbose) >= 1 {
			LogWarning.Printf("Player tried logging in with invalid client version. Got %d, expected %d\n", version, TomlConfig.Version)
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
	c.player.SetIndex(c.Index)
	c.player.Username, _ = p.ReadString()
	c.player.UserBase37 = strutil.Base37(c.player.Username)
	c.player.Username = strutil.DecodeBase37(c.player.UserBase37)
	password, _ := p.ReadString()
	passHash := HashPassword(password)
	if _, ok := Clients[c.player.UserBase37]; ok {
		c.sendLoginResponse(4)
		return
	}
	c.sendLoginResponse(byte(c.LoadPlayer(c.player.UserBase37, passHash)))
}
