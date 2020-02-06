/*
 * Copyright (c) 2020 Zachariah Knight <aeros.storkpk@gmail.com>
 *
 * Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 */

package website

import (
	"bufio"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/spkaeros/rscgo/pkg/game/world"
	"github.com/spkaeros/rscgo/pkg/log"
)

var muxCtx = http.NewServeMux()

type InformationData struct {
	Title     string
	Owner     string
	Copyright string
}

var Information = InformationData{
	Title:     "RSCGo",
	Owner:     "ZlackCode LLC",
	Copyright: "2019-2020",
}

func (s InformationData) ToLower(s2 string) string {
	return strings.ToLower(s2)
}

func (s InformationData) OnlineCount() int {
	return world.Players.Size()
}

var index = template.Must(template.ParseFiles("./website/index.gohtml"))

func indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := index.Execute(w, Information)
		if err != nil {
			log.Error.Println("Could not execute html template:", err)
			return
		}
	})
}
/*
var html = []byte(
	`<html>
	<body>
		<h1>game process stdout/stderr</h1>
		<code></code>
		<script>
			var ws = new WebSocket("wss://rscturmoil.com/game/out")
			ws.onmessage = function(e) {
				document.querySelector("code").innerHTML += e.data + "<br>"
			}
		</script>
	</body>
</html>
`)
*/

var html = []byte(`<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<link rel="stylesheet" type="text/css" href="/style.css" />
		<script>
			function setStatus(status) {
				document.getElementById("reply").innerHTML = status;
			}

			function appendOutput(msg) {
				document.getElementById("stdout").innerHTML += msg + "<br>\n"
			}

			function callApi(url) {
				var xhttp = new XMLHttpRequest();
				xhttp.onreadystatechange = function() {
					if (this.readyState == 4 && this.status == 200) {
						setStatus(this.responseText);
					}
				};
				xhttp.open("GET", url, true);
				xhttp.send(); 
			}

			function initWebsocket() {
				if (window.WebSocket === undefined) {
					document.getElementById("stdout").innerHTML = "Your browser does not appear to have WebSockets capabilities.<br>\nTo use this page, consider upgrading to any modern alternative, such as Firefox or Chromium.";
					return;
				}

				var ws = new WebSocket("wss://rscturmoil.com/game/out");

				ws.onmessage = function(event) {
					appendOutput(event.data);
					var container = document.getElementById("stdout-box");
					container.scrollTop = container.scrollHeight;
				}
				
				ws.onopen = function() {
					appendOutput("[WS] Connected to stdout HTTP endpoint" + "<br>\n");
				}
				
				ws.onclose = function() {
					appendOutput("<br><br>\n\n" + "[WS] Disconnected" + "<br>\n");
				}
			}
			function launch() {
				setStatus("Attempting to launch server...");
				callApi("launch.ws");
			};
			
			function terminate() {
				setStatus("Attempting to shutdown server...");
				callApi("shutdown.ws");
			}
			initWebsocket();
		</script>
		<title>Game server controls</title>
	</head>

	<body>
		<div class="rsc-container" style="text-align:center;">
			<header>
				<div class="rsc-border-top rsc-border-bar"></div>
				<div class="rsc-box rsc-header">
					<b>Server Controls</b><br>
					<a class="rsc-link" href="/index.ws">Main menu</a>
				</div>
			</header>

			<p style="font-variant:petite-caps; font-weight:bold;" id="reply"></p>
			<div class="rsc-box" id="stdout-box" style="margin:5px 55px 15px 55px; border-radius: 15px; padding:23px; height:356px; text-align:left; overflow-y:scroll; ">
				<code id="stdout"></code>
			</div>
			<p>
				<h2>Controls:</h2><br>
				<button href="#" id="launch" onclick="launch()" type="button">Start</button>
				<button href="#" style="margin-left:50px;" id="terminate" onclick="terminate()" type="button">Stop</button>
			</p>
			<footer class="rsc-border-bottom rsc-border-bar">
				<div class="rsc-footer">
					This webpage and its contents is copyright © 2019-2020 ZlackCode, LLC.
					<br>To use our service you must agree to our <a class="rsc-link" href="/terms.html">Terms+Conditions</a> and <a class="rsc-link" href="/privacy.html">Privacy policy</a>
				</div>
			</footer>
		</div>
	</body>
</html>`)
var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var outBuffer = make(chan []byte, 256)
var backBuffer = make([][]byte, 0, 1000)
var ServerProc *os.Process

//Start Binds to the web port 8080 and serves HTTP content to it.
// Note: This is a blocking call, it will not return to caller.
func Start() {
	muxCtx.Handle("/", http.NotFoundHandler())
	muxCtx.Handle("/index.ws", indexHandler())
	muxCtx.HandleFunc("/game/launch.ws", func(w http.ResponseWriter, r *http.Request) {
		if ServerProc != nil {
			w.Write([]byte("game already started\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		cmd := exec.Command("./bin/game", "-v")

		outReader, err := cmd.StdoutPipe()
		if err != nil {
			w.Write([]byte("Error getting game game output pipe reader:" + err.Error()))
		}
		errReader, err := cmd.StderrPipe()
		if err != nil {
			w.Write([]byte("Error getting game game error pipe reader:" + err.Error()))
		}
		scanner := bufio.NewScanner(io.MultiReader(outReader, errReader))
		//		multiWriter := io.MultiWriter(os.Stdout, &outBuffer)
		go func() {
			for scanner.Scan() {
				os.Stdout.Write(append(scanner.Bytes(), byte('\n')))
				outBuffer <- []byte(scanner.Text())
			}
		}()
	err = cmd.Start()
		if err != nil {
			w.Write([]byte("Error starting game process:" + err.Error()))
			return
		}
		ServerProc = cmd.Process
		w.Write([]byte("Started game server (pid: " + strconv.Itoa(ServerProc.Pid) + ")"))
	})
	muxCtx.HandleFunc("/game/shutdown.ws", func(w http.ResponseWriter, r *http.Request) {
		if ServerProc == nil {
			w.Write([]byte("game child process not launched.\n"))
			return
		}
		cmd := exec.Command("kill", "-9", strconv.Itoa(ServerProc.Pid))
		err := cmd.Run()
		if err != nil {
			w.Write([]byte("Error starting kill process:" + err.Error()))
			return
		}
		w.Write([]byte("Shut down game server (pid: " + strconv.Itoa(ServerProc.Pid) + ")"))
		ServerProc = nil
	})
	muxCtx.HandleFunc("/game/control.ws", func(w http.ResponseWriter, r *http.Request) {
		w.Write(html)
	})
	muxCtx.HandleFunc("/game/out", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Error.Printf("upgrade error: %s", err)
			return
		}
		defer conn.Close()
		for _, line := range backBuffer {
			if err := wsutil.WriteServerText(conn, line); err != nil {
				log.Info.Println(err)
				return
			}
		}
		for {
			select {
			case line, ok := <-outBuffer:
				if !ok {
					return
				}
				if len(backBuffer) == 1000 {
					backBuffer = backBuffer[100:]
				}
				backBuffer = append(backBuffer, line)
				err := wsutil.WriteServerText(conn, line)
				if err != nil {
					log.Info.Println(err)
					return
				}
			}
		}
		backBuffer = backBuffer[:0]
	})
	err := http.ListenAndServe(":8080", muxCtx)
	if err != nil {
		log.Error.Println("Could not bind to website port:", err)
		os.Exit(99)
	}
}

