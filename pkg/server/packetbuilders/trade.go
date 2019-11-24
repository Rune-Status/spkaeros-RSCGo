package packetbuilders

import (
	"github.com/spkaeros/rscgo/pkg/server/packet"
	"github.com/spkaeros/rscgo/pkg/server/world"
)

//TradeClose Closes a trade window
var TradeClose = packet.NewOutgoingPacket(128)

//TradeUpdate Builds a packet to update a trade offer
func TradeUpdate(player *world.Player) (p *packet.Packet) {
	p = packet.NewOutgoingPacket(97)
	p.AddByte(uint8(player.TradeOffer.Size()))
	player.TradeOffer.Range(func(item *world.Item) bool {
		p.AddShort(uint16(item.ID))
		p.AddInt(uint32(item.Amount))
		return true
	})
	return
}

//TradeOpen Builds a packet to open a trade window
func TradeOpen(player *world.Player) *packet.Packet {
	return packet.NewOutgoingPacket(92).AddShort(uint16(player.TradeTarget()))
}

//TradeTargetAccept Builds a packet to change trade targets accepted status
func TradeTargetAccept(accepted bool) *packet.Packet {
	if accepted {
		return packet.NewOutgoingPacket(162).AddByte(1)
	}
	return packet.NewOutgoingPacket(162).AddByte(0)
}

//TradeAccept Builds a packet to change trade targets accepted status
func TradeAccept(accepted bool) *packet.Packet {
	if accepted {
		return packet.NewOutgoingPacket(15).AddByte(1)
	}
	return packet.NewOutgoingPacket(15).AddByte(0)
}

//TradeConfirmationOpen Builds a packet to open the trade confirmation page
func TradeConfirmationOpen(player, other *world.Player) *packet.Packet {
	p := packet.NewOutgoingPacket(20)

	p.AddLong(other.UserBase37)
	p.AddByte(uint8(other.TradeOffer.Size()))
	other.TradeOffer.Range(func(item *world.Item) bool {
		p.AddShort(uint16(item.ID))
		p.AddInt(uint32(item.Amount))
		return true
	})

	p.AddByte(uint8(player.TradeOffer.Size()))
	player.TradeOffer.Range(func(item *world.Item) bool {
		p.AddShort(uint16(item.ID))
		p.AddInt(uint32(item.Amount))
		return true
	})

	return p
}
