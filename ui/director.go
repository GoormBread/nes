package ui

import (
	"log"

	"github.com/fogleman/nes/nes"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"net/http"
    "github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
        return true // 모든 Origin 허용
    },
}
type View interface {
	Enter()
	Exit()
	Update(t, dt float64)
}

type Director struct {
	window    *glfw.Window
	audio     *Audio
	view      View
	menuView  View
	timestamp float64
}

func NewDirector(window *glfw.Window, audio *Audio) *Director {
	director := Director{}
	director.window = window
    director.audio = audio
	return &director
}

func (d *Director) SetTitle(title string) {
	d.window.SetTitle(title)
}

func (d *Director) SetView(view View) {
	if d.view != nil {
		d.view.Exit()
	}
	d.view = view
	if d.view != nil {
		d.view.Enter()
	}
	d.timestamp = glfw.GetTime()
}

func (d *Director) Step() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	timestamp := glfw.GetTime()
	dt := timestamp - d.timestamp
	d.timestamp = timestamp
	if d.view != nil {
		d.view.Update(timestamp, dt)
	}
}

func (d *Director) Start(paths []string) {
	d.menuView = NewMenuView(d, paths)
	if len(paths) == 1 {
		d.PlayGame(paths[0])
	} else {
		d.ShowMenu()
	}
	d.Run()
}

func (d *Director) Run() {
	for !d.window.ShouldClose() {
		d.Step()
		d.window.SwapBuffers()
		glfw.PollEvents()
	}
	d.SetView(nil)

}

func (d *Director) PlayGame(path string) {
    hash, err := hashFile(path)
    if err != nil {
        log.Fatalln(err)
    }
    console, err := nes.NewConsole(path)
    if err != nil {
        log.Fatalln(err)
    }
    d.SetView(NewGameView(d, console, path, hash))

    // 1201 포트에서 웹소켓 연결 처리
    http.HandleFunc("/keyboard/1p", func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Println("웹소켓 연결 설정 오류:", err)
            return
        }
        defer conn.Close()

        // listen for messages from websocket client
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("메시지 읽기 오류:", err)
                break
            }
            log.Println(string(message))

            // handle key input from websocket
            updateCloudControllers1(d.window, message, console)
        }
    })
    go http.ListenAndServe(":1201", nil)

    // 1014 포트에서 웹소켓 연결 처리
    http.HandleFunc("/keyboard/2p", func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Println("웹소켓 연결 설정 오류:", err)
            return
        }
        defer conn.Close()

        // listen for messages from websocket client
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("메시지 읽기 오류:", err)
                break
            }
            log.Println(string(message))

            // handle key input from websocket
            updateCloudControllers2(d.window, message, console)
        }
    })
    go http.ListenAndServe(":1014", nil)

    log.Println("실행")
}

func (d *Director) ShowMenu() {
	d.SetView(d.menuView)
}
