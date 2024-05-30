package ui

import (
	"image"

	"github.com/fogleman/nes/nes"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const padding = 0

type GameView struct {
	director *Director
	console  *nes.Console
	title    string
	hash     string
	texture  uint32
	record   bool
	frames   []image.Image
	// message []byte
}

func NewGameView(director *Director, console *nes.Console, title, hash string) View {
	texture := createTexture()
	return &GameView{director, console, title, hash, texture, false, nil}
}

func (view *GameView) load(snapshot int) {
	// load state
	if err := view.console.LoadState(savePath(view.hash, snapshot)); err == nil {
		return
	} else {
		view.console.Reset()
	}
	// load sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash, snapshot)); err == nil {
			cartridge.SRAM = sram
		}
	}
}

func (view *GameView) save(snapshot int) {
	// save sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		writeSRAM(sramPath(view.hash, snapshot), cartridge.SRAM)
	}
	// save state
	view.console.SaveState(savePath(view.hash, snapshot))
}

func (view *GameView) Enter() {
	gl.ClearColor(0, 0, 0, 1)
	view.director.SetTitle(view.title)
	// view.console.SetAudioChannel(view.director.audio.channel)
	// view.console.SetAudioSampleRate(view.director.audio.sampleRate)
	view.director.window.SetKeyCallback(view.onKey)
	view.load(-1)
}

func (view *GameView) Exit() {
	view.director.window.SetKeyCallback(nil)
	view.console.SetAudioChannel(nil)
	view.console.SetAudioSampleRate(0)
	view.save(-1)
}

func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	window := view.director.window
	console := view.console
	if joystickReset(glfw.Joystick1) {
		view.director.ShowMenu()
	}
	if joystickReset(glfw.Joystick2) {
		view.director.ShowMenu()
	}
	if readKey(window, glfw.KeyEscape) {
		view.director.ShowMenu()
	}

	console.StepSeconds(dt)
	// updateControllers(window, console)
	// updateCloudControllers(window, view.message, console)
	gl.BindTexture(gl.TEXTURE_2D, view.texture)
	setTexture(console.Buffer())
	drawBuffer(view.director.window)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	if view.record {
		view.frames = append(view.frames, copyImage(console.Buffer()))
	}
}

func (view *GameView) onKey(window *glfw.Window,
	key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		if key >= glfw.Key0 && key <= glfw.Key9 {
			snapshot := int(key - glfw.Key0)
			if mods&glfw.ModShift == 0 {
				view.load(snapshot)
			} else {
				view.save(snapshot)
			}
		}
		switch key {
		case glfw.KeySpace:
			screenshot(view.console.Buffer())
		case glfw.KeyR:
			view.console.Reset()
		case glfw.KeyTab:
			if view.record {
				view.record = false
				animation(view.frames)
				view.frames = nil
			} else {
				view.record = true
			}
		}
	}
}

func drawBuffer(window *glfw.Window) {
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / 256
	s2 := float32(h) / 240
	f := float32(1 - padding)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

func updateCloudControllers1(window *glfw.Window, message []byte, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3
	k1 := readKeysFromUser1(message, turbo)
	console.SetButtons1(k1)
}

func updateCloudControllers2(window *glfw.Window, message []byte, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3
	k2 := readKeysFromUser2(message, turbo)
	console.SetButtons2(k2)
}

func updateCloudControllersDefault(window *glfw.Window, message []byte, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3
	j2 := readJoystick(glfw.Joystick2, turbo)
	window.MakeContextCurrent()
	var result [8]bool
	result[nes.ButtonA] = false
	result[nes.ButtonB] = false
	result[nes.ButtonSelect] = false
	result[nes.ButtonStart] = false
	result[nes.ButtonUp] = false
	result[nes.ButtonDown] = false
	result[nes.ButtonLeft] = false
	result[nes.ButtonRight] = false
	console.SetButtons1(result)
	console.SetButtons2(j2)
	
}

func updateControllers(window *glfw.Window, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3
	k1 := readKeys1(window, turbo)
	k2 := readKeys2(window, turbo)
	j1 := readJoystick(glfw.Joystick1, turbo)
	// j2 := readJoystick(glfw.Joystick2, turbo)
	console.SetButtons1(combineButtons(k1, j1))
	console.SetButtons2(k2)
	window.MakeContextCurrent()
}
