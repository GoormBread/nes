package ui

import (
    "crypto/md5"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "image"
    "image/color"
    "image/draw"
    "image/gif"
    "image/png"
    "io/ioutil"
    "log"
    "os"
    "os/user"
    "path"

    "github.com/fogleman/nes/nes"
    "github.com/go-gl/gl/v2.1/gl"
    "github.com/go-gl/glfw/v3.2/glfw"
)

var homeDir string

func init() {
	u, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	homeDir = u.HomeDir
}

func thumbnailURL(hash string) string {
	return "http://www.michaelfogleman.com/static/nes/" + hash + ".png"
}

func thumbnailPath(hash string) string {
	return homeDir + "/.nes/thumbnail/" + hash + ".png"
}

func sramPath(hash string, snapshot int) string {
	if snapshot >= 0 {
		return fmt.Sprintf("%s/.nes/sram/%s-%d.dat", homeDir, hash, snapshot)
	}
	return fmt.Sprintf("%s/.nes/sram/%s.dat", homeDir, hash)
}

func savePath(hash string, snapshot int) string {
	if snapshot >= 0 {
		return fmt.Sprintf("%s/.nes/save/%s-%d.dat", homeDir, hash, snapshot)
	}
	return fmt.Sprintf("%s/.nes/save/%s.dat", homeDir, hash)
}

func readKey(window *glfw.Window, key glfw.Key) bool {
	return window.GetKey(key) == glfw.Press
}

func readKeysFromUser1(message []byte, turbo bool) [8]bool {
    var result [8]bool
    result[nes.ButtonA] = false
    result[nes.ButtonB] = false
    result[nes.ButtonSelect] = false
    result[nes.ButtonStart] = false
    result[nes.ButtonUp] = false
    result[nes.ButtonDown] = false
    result[nes.ButtonLeft] = false
    result[nes.ButtonRight] = false

    var pressedKeys []string
    err := json.Unmarshal(message, &pressedKeys)
    if err != nil {
        log.Println("JSON 파싱 오류:", err)
        return result
    }

    for _, key := range pressedKeys {
        switch key {
        case "Z":
            result[nes.ButtonA] = true
        case "X":
            result[nes.ButtonB] = true
        case "Enter":
            log.Println("실행됨")
            result[nes.ButtonStart] = true
        case "RightShift":
            result[nes.ButtonSelect] = true
        case "UP":
            result[nes.ButtonUp] = true
        case "DOWN":
            result[nes.ButtonDown] = true
        case "LEFT":
            result[nes.ButtonLeft] = true
        case "RIGHT":
            result[nes.ButtonRight] = true
        }
    }

    return result
}

func readKeysFromUser2(message []byte, turbo bool) [8]bool {
    var result [8]bool
    result[nes.ButtonA] = false
    result[nes.ButtonB] = false
    result[nes.ButtonSelect] = false
    result[nes.ButtonStart] = false
    result[nes.ButtonUp] = false
    result[nes.ButtonDown] = false
    result[nes.ButtonLeft] = false
    result[nes.ButtonRight] = false

    var pressedKeys []string
    err := json.Unmarshal(message, &pressedKeys)
    if err != nil {
        log.Println("JSON 파싱 오류:", err)
        return result
    }

    for _, key := range pressedKeys {
        switch key {
        case "Z":
            result[nes.ButtonA] = true
        case "X":
            result[nes.ButtonB] = true
        case "Enter":
            log.Println("실행됨")
            result[nes.ButtonStart] = true
        case "RightShift":
            result[nes.ButtonSelect] = true
        case "UP":
            result[nes.ButtonUp] = true
        case "DOWN":
            result[nes.ButtonDown] = true
        case "LEFT":
            result[nes.ButtonLeft] = true
        case "RIGHT":
            result[nes.ButtonRight] = true
        }
    }

    return result
}

func readKeys1(window *glfw.Window, turbo bool) [8]bool {
	var result [8]bool
	result[nes.ButtonA] = readKey(window, glfw.KeyZ) || (turbo && readKey(window, glfw.KeyA))
	result[nes.ButtonB] = readKey(window, glfw.KeyX) || (turbo && readKey(window, glfw.KeyS))
	result[nes.ButtonSelect] = readKey(window, glfw.KeyRightShift)
	result[nes.ButtonStart] = readKey(window, glfw.KeyEnter)
	result[nes.ButtonUp] = readKey(window, glfw.KeyUp)
	result[nes.ButtonDown] = readKey(window, glfw.KeyDown)
	result[nes.ButtonLeft] = readKey(window, glfw.KeyLeft)
	result[nes.ButtonRight] = readKey(window, glfw.KeyRight)
	return result
}

func readKeys2(window *glfw.Window, turbo bool) [8]bool {
	var result [8]bool
	result[nes.ButtonA] = readKey(window, glfw.KeyJ) || (turbo && readKey(window, glfw.KeyJ))
	result[nes.ButtonB] = readKey(window, glfw.KeyK) || (turbo && readKey(window, glfw.KeyK))
	result[nes.ButtonSelect] = readKey(window, glfw.KeyO)
	result[nes.ButtonStart] = readKey(window, glfw.KeyP)
	result[nes.ButtonUp] = readKey(window, glfw.KeyT)
	result[nes.ButtonDown] = readKey(window, glfw.KeyG)
	result[nes.ButtonLeft] = readKey(window, glfw.KeyF)
	result[nes.ButtonRight] = readKey(window, glfw.KeyH)
	return result
}

func readKeysFromUser(message []byte, turbo bool) [8]bool {
	var result [8]bool
	result[nes.ButtonA] = false || (turbo && false)
	result[nes.ButtonB] = false || (turbo && false)
	result[nes.ButtonSelect] = false || (turbo && false)
	result[nes.ButtonStart] = false || (turbo && false)
	result[nes.ButtonUp] = false || (turbo && false)
	result[nes.ButtonDown] = false || (turbo && false)
	result[nes.ButtonLeft] = false || (turbo && false)
	result[nes.ButtonRight] = false || (turbo && false)
	if string(message) == "Z" {
		log.Println("실행됨")
		result[nes.ButtonRight] = true
	}
	if string(message) == "W" {
		log.Println("실행됨")
		result[nes.ButtonStart] = true
	}
	return result
	
}

func readJoystick(joy glfw.Joystick, turbo bool) [8]bool {
	var result [8]bool
	if !glfw.JoystickPresent(joy) {
		return result
	}
	joyname := glfw.GetJoystickName(joy)
	axes := glfw.GetJoystickAxes(joy)
	buttons := glfw.GetJoystickButtons(joy)
	if joyname == "PLAYSTATION(R)3 Controller" {
		result[nes.ButtonA] = buttons[14] == 1 || (turbo && buttons[2] == 1)
		result[nes.ButtonB] = buttons[13] == 1 || (turbo && buttons[3] == 1)
		result[nes.ButtonSelect] = buttons[0] == 1
		result[nes.ButtonStart] = buttons[3] == 1
		result[nes.ButtonUp] = buttons[4] == 1 || axes[1] < -0.5
		result[nes.ButtonDown] = buttons[6] == 1 || axes[1] > 0.5
		result[nes.ButtonLeft] = buttons[7] == 1 || axes[0] < -0.5
		result[nes.ButtonRight] = buttons[5] == 1 || axes[0] > 0.5
		return result
	}
	if len(buttons) < 8 {
		return result
	}
	result[nes.ButtonA] = buttons[0] == 1 || (turbo && buttons[2] == 1)
	result[nes.ButtonB] = buttons[1] == 1 || (turbo && buttons[3] == 1)
	result[nes.ButtonSelect] = buttons[6] == 1
	result[nes.ButtonStart] = buttons[7] == 1
	result[nes.ButtonUp] = axes[1] < -0.5
	result[nes.ButtonDown] = axes[1] > 0.5
	result[nes.ButtonLeft] = axes[0] < -0.5
	result[nes.ButtonRight] = axes[0] > 0.5
	return result
}

func joystickReset(joy glfw.Joystick) bool {
	if !glfw.JoystickPresent(joy) {
		return false
	}
	buttons := glfw.GetJoystickButtons(joy)
	if len(buttons) < 6 {
		return false
	}
	return buttons[4] == 1 && buttons[5] == 1
}

func combineButtons(a, b [8]bool) [8]bool {
	var result [8]bool
	for i := 0; i < 8; i++ {
		result[i] = a[i] || b[i]
	}
	return result
}

func hashFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

func createTexture() uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return texture
}

func setTexture(im *image.RGBA) {
	size := im.Rect.Size()
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGBA, int32(size.X), int32(size.Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(im.Pix))
}

func copyImage(src image.Image) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Rect, src, image.ZP, draw.Src)
	return dst
}

func loadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func savePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

func saveGIF(path string, frames []image.Image) error {
	var palette []color.Color
	for _, c := range nes.Palette {
		palette = append(palette, c)
	}
	g := gif.GIF{}
	for i, src := range frames {
		if i%3 != 0 {
			continue
		}
		dst := image.NewPaletted(src.Bounds(), palette)
		draw.Draw(dst, dst.Rect, src, image.ZP, draw.Src)
		g.Image = append(g.Image, dst)
		g.Delay = append(g.Delay, 5)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return gif.EncodeAll(file, &g)
}

func screenshot(im image.Image) {
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("%03d.png", i)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			savePNG(path, im)
			return
		}
	}
}

func animation(frames []image.Image) {
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("%03d.gif", i)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			saveGIF(path, frames)
			return
		}
	}
}

func writeSRAM(filename string, sram []byte) error {
	dir, _ := path.Split(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return binary.Write(file, binary.LittleEndian, sram)
}

func readSRAM(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	sram := make([]byte, 0x2000)
	if err := binary.Read(file, binary.LittleEndian, sram); err != nil {
		return nil, err
	}
	return sram, nil
}
