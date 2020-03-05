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
	"github.com/spkaeros/rscgo/pkg/game/net/handshake"
	"strings"
	"time"

	"github.com/spkaeros/rscgo/pkg/config"
	"github.com/spkaeros/rscgo/pkg/crypto"
	"github.com/spkaeros/rscgo/pkg/db"
	"github.com/spkaeros/rscgo/pkg/engine/tasks"
	"github.com/spkaeros/rscgo/pkg/game/net"
	"github.com/spkaeros/rscgo/pkg/game/world"
	"github.com/spkaeros/rscgo/pkg/log"
	"github.com/spkaeros/rscgo/pkg/strutil"
)

func init() {
	//dataService is a db.PlayerService that all login-related functions should use to access or change player profile data.
	var dataService = db.DefaultPlayerService
	
	AddHandler("forgotpass", func(player *world.Player, p *net.Packet) {
		// TODO: These non-login handlers must be isolated and rewrote
		go func() {
			usernameHash := p.ReadUint64()
			if !dataService.PlayerHasRecoverys(usernameHash) {
				player.SendPacket(net.NewReplyPacket([]byte{0}))
				player.Destroy()
				return
			}
			player.SendPacket(net.NewReplyPacket([]byte{1}))
			for _, question := range dataService.PlayerLoadRecoverys(usernameHash) {
				player.SendPacket(net.NewReplyPacket([]byte{byte(len(question))}).AddBytes([]byte(question)))
			}
		}()
	})
	AddHandler("loginreq", func(player *world.Player, p *net.Packet) {
		player.SetConnected(true)
		loginReply := handshake.NewLoginListener(player).ResponseListener()
		if handshake.LoginThrottle.Recent(player.CurrentIP(), time.Minute*5) >= 5 {
			loginReply <- handshake.ResponseSpamTimeout
			return
		}
		player.SetReconnecting(p.ReadBoolean())
		if ver := p.ReadUint16(); ver != config.Version() {
			log.Info.Printf("Invalid client version attempted to login: %d\n", ver)
			loginReply <- handshake.ResponseUpdated
			return
		}
		var password string

		// On the legacy protocol the following block was 42 bytes long, now it is 10-30
		if p.Length() >= 65 {
			// two constants of unknown significance, [0,10]
			p.ReadUint16()

			// Here we read a 128-bit CSPRNG seed, to encrypt any future packet opcodes with.
			// The 64 least significant bits are generated by the server earlier in the handshake,
			// so this should match whatever the server sent earlier.
			// The 64 most significant bits are generated by the client.  The server shouldn't care what the client provides here.
			// Note: Deprecated in favor of TLS
			p.ReadUint128()

			// this was named linkUID by jagex; it identifys a unique user agent
			p.ReadUint32()
			
			player.Transients().SetVar("username", strutil.Base37.Encode(strings.TrimSpace(p.ReadStringN(20))))
			password = strings.TrimSpace(p.ReadStringN(20))
		} else {
			player.Transients().SetVar("username", p.ReadUint64())
			password = strings.TrimSpace(p.ReadString())
		}
		
		if world.Players.ContainsHash(player.UsernameHash()) {
			loginReply <- handshake.ResponseLoggedIn
			return
		}
		if !world.UpdateTime.IsZero() && time.Until(world.UpdateTime).Seconds() <= 0 {
			loginReply <- handshake.ResponseLoginServerRejection
			return
		}

		go func() {
			if !dataService.PlayerNameTaken(player.Username()) || !dataService.PlayerValidLogin(player.UsernameHash(), crypto.Hash(password)) {
				loginReply <- handshake.ResponseBadPassword
				return
			}
			if !dataService.PlayerLoad(player) {
				loginReply <- handshake.ResponseDecodeFailure
				return
			}
	
			if player.Reconnecting() {
				loginReply <- handshake.ResponseReconnected
				return
			}
	
			switch player.Rank() {
			case 2:
				loginReply <- handshake.ResponseAdministrator
			case 1:
				loginReply <- handshake.ResponseModerator
			default:
				loginReply <- handshake.ResponseLoginSuccess
			}
		}()
	})
	AddHandler("logoutreq", func(player *world.Player, p *net.Packet) {
		tasks.Tickers.Add("playerDestroy", func() bool {
			if player.Busy() {
				player.SendPacket(world.CannotLogout)
				return true
			}
			if player.Connected() {
				player.Destroy()
			}
			return true
		})
	})
	AddHandler("closeconn", func(player *world.Player, p *net.Packet) {
		if player.Busy() {
			log.Suspicious.Println("CLOSECONN!!", player, p.String(), player.State())
			log.Info.Println("CLOSECONN!!", player, p.String(), player.State())
			player.SendPacket(world.CannotLogout)
			return
		}
		if player.Connected() {
			player.Destroy()
		}
	})
	AddHandler("cancelpq", func(player *world.Player, p *net.Packet) {
		// empty net
	})
	AddHandler("setpq", func(player *world.Player, p *net.Packet) {
		var questions []string
		var answers []uint64
		for i := 0; i < 5; i++ {
			length := p.ReadUint8()
			questions = append(questions, p.ReadStringN(int(length)))
			answers = append(answers, p.ReadUint64())
		}
		log.Info.Println(questions, answers)
	})
	AddHandler("changepq", func(player *world.Player, p *net.Packet) {
		player.SendPacket(net.NewEmptyPacket(224))
	})
	AddHandler("changepass", func(player *world.Player, p *net.Packet) {
		oldPassword := p.ReadString()
		newPassword := p.ReadString()
		go func() {
			if !dataService.PlayerValidLogin(player.UsernameHash(), crypto.Hash(oldPassword)) {
				player.Message("The old password you provided does not appear to be valid.  Try again.")
				return
			}
			dataService.PlayerChangePassword(player.UsernameHash(), crypto.Hash(newPassword))
			player.Message("Successfully updated your password to the new password you have provided.")
		}()
	})
	AddHandler("newplayer", func(player *world.Player, p *net.Packet) {
		player.SetConnected(true)

		reply := handshake.NewRegistrationListener(player).ResponseListener()
		if handshake.RegisterThrottle.Recent(player.CurrentIP(), time.Minute*5) >= 5 {
			reply <- handshake.ResponseSpamTimeout
			return
		}
		if version := p.ReadUint16(); version != config.Version() {
			log.Info.Printf("New player denied: [ Reason:'Wrong client version'; ip='%s'; version=%d ]\n", player.CurrentIP(), version)
			reply <- handshake.ResponseUpdated
			return
		}
		username := strutil.Base37.Decode(p.ReadUint64())
		password := strings.TrimSpace(p.ReadString())
		player.Transients().SetVar("username", username)
		if userLen, passLen := len(username), len(password); userLen < 2 || userLen > 12 || passLen < 5 || passLen > 20 {
			log.Suspicious.Printf("New player request contained invalid lengths: %v username=%v; password:'%v'\n", player.CurrentIP(), username, password)
			reply <- 17
			return
		}
		go func() {
			if dataService.PlayerNameTaken(username) {
				log.Info.Printf("New player denied: [ Reason:'Username is taken'; username='%s'; ip='%s' ]\n", username, player.CurrentIP())
				reply <- handshake.ResponseUsernameTaken
				return
			}

			if dataService.PlayerCreate(username, password) {
				log.Info.Printf("New player accepted: [ username='%s'; ip='%s' ]", username, player.CurrentIP())
				reply <- handshake.ResponseRegisterSuccess
				return
			}
			log.Info.Printf("New player denied: [ Reason:'unknown; probably database related.  Debug required'; username='%s'; ip='%s' ]\n", username, player.CurrentIP())
			reply <- -1
		}()
	})
}
