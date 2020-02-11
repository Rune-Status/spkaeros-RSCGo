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
	"time"

	"github.com/spkaeros/rscgo/pkg/game/net"
	"github.com/spkaeros/rscgo/pkg/game/world"
	"github.com/spkaeros/rscgo/pkg/log"
)

func init() {
	AddHandler("attacknpc", func(player *world.Player, p *net.Packet) {
		npc := world.GetNpc(p.ReadShort())
		if npc == nil {
			log.Suspicious.Printf("%v tried to attack nil NPC\n", player)
			return
		}
		player.SetDistancedAction(func() bool {
			if player.Busy() {
				return true
			}
			if !player.CanAttack(npc) {
				return true
			}
			if player.NextTo(npc.Location) && player.WithinRange(npc.Location, 1) {
				for _, trigger := range world.NpcAtkTriggers {
					if trigger.Check(player, npc) {
						trigger.Action(player, npc)
						return true
					}
				}
				if time.Since(npc.TransAttrs.VarTime("lastFight")) <= time.Second*2 || npc.Busy() {
					return true
				}
				player.ResetPath()
				npc.ResetPath()
				player.StartCombat(npc)
				return true
			}
			player.SetPath(world.MakePath(player.Location, npc.Location))
			return false
		})
	})
	AddHandler("attackplayer", func(player *world.Player, p *net.Packet) {
		affectedPlayer, ok := world.Players.FromIndex(p.ReadShort())
		if affectedPlayer == nil || !ok {
			log.Suspicious.Printf("player[%v] tried to attack nil player\n", player)
			return
		}
		player.SetDistancedAction(func() bool {
			if player.Busy() {
				return true
			}
			if affectedPlayer.Busy() {
				log.Info.Printf("Target player busy during attack request  State: %d\n", affectedPlayer.State)
				return true
			}
			if !player.CanAttack(affectedPlayer) {
				return true
			}
			if player.NextTo(affectedPlayer.Location) && player.WithinRange(affectedPlayer.Location, 2) {
				player.ResetPath()
				if time.Since(affectedPlayer.TransAttrs.VarTime("lastRetreat")) <= time.Second*3 || affectedPlayer.IsFighting() {
					return true
				}
				player.ResetPath()
				affectedPlayer.ResetPath()
				player.StartCombat(affectedPlayer)
				return true
			}
			return player.FinishedPath()
		})
	})
	AddHandler("fightmode", func(player *world.Player, p *net.Packet) {
		mode := p.ReadByte()
		if mode < 0 || mode > 3 {
			log.Suspicious.Printf("Invalid fightmode(%v) selected by %s", mode, player.String())
			return
		}
		player.SetFightMode(int(mode))
	})
}
