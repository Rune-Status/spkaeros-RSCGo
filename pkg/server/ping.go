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

import "bitbucket.org/zlacki/rscgo/pkg/server/packets"

func init() {
	Handlers[5] = func(c *Client, p *packets.Packet) {
		c.outgoingPackets <- packets.ResponsePong
	}
}
