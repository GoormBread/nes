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
	// audio     *Audio
	view      View
	menuView  View
	timestamp float64
}

func NewDirector(window *glfw.Window) *Director {
	director := Director{}
	director.window = window
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
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
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
			// if string(message) == "Enter" || string(message) == "RightShift" || string(message) == "Z" || string(message) == "X" || string(message) == "UP" || string(message) == "DOWN" || string(message) == "LEFT" || string(message) == "RIGHT" || string(message) == "J" || string(message) == "K" || string(message) == "P" || string(message) == "O" || string(message) == "T" || string(message) == "G" || string(message) == "F" || string(message) == "H" {
				log.Println("입력됨")
				// d.view.onKey(d.window, glfw.KeyZ, 0, glfw.Press, 0)
				updateCloudControllers(d.window, message, console)
				// W 키에 대한 입력 처리 로직 추가
				// 예: director.KeyPressed(glfw.KeyW)
			// }
			// updateCloudControllersDefault(d.window, message, console)
			
        }
    })
	go http.ListenAndServe(":8080", nil)
	log.Println("실행")
}

func (d *Director) ShowMenu() {
	d.SetView(d.menuView)
}
