package runplayfield

import (
	"danser/audio"
	"danser/beatmap"
	"danser/bmath"
	. "danser/build"
	"danser/dance"
	"danser/database"
	. "danser/osuconst"
	"danser/render"
	"danser/render/font"
	"danser/settings"
	"danser/states"
	"danser/utils"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/wieku/glhf"
	"image"
	"log"
	"math"
	"os"
)

var player *states.Player
var pressed = false
var pressedM = false
var pressedP = false

var SettingsVersion = 0
var artist = ""
var creator = ""
var cursors = 1
var tag = 1
var pitch = 1.0
var mover = "linear"
var debug = false
var fps = false

func RunPlayField() {
	var win *glfw.Window

	mainthread.Call(func() {

		settings.DEBUG = debug
		settings.FPS = fps
		settings.DIVIDES = cursors
		settings.TAG = tag
		settings.PITCH = pitch
		_ = mover
		dance.SetMover(mover)

		newSettings := settings.LoadSettings(SettingsVersion)

		player = nil
		var beatMap *beatmap.BeatMap = nil

		// 去除flag使用
		title := settings.VSplayer.MapInfo.Title
		difficulty := settings.VSplayer.MapInfo.Difficulty

		if (artist + title + difficulty + creator) == "" {
			log.Println("No beatmap specified, closing...")
			os.Exit(0)
		}

		database.Init()
		beatmaps := database.LoadBeatmaps()

		for _, b := range beatmaps {
			if (artist == "" || artist == b.Artist) && (title == "" || title == b.Name) && (difficulty == "" || difficulty == b.Difficulty) && (creator == "" || creator == b.Creator) {
				beatMap = b
				beatMap.UpdatePlayStats()
				database.UpdatePlayStats(beatMap)
				break
			}
		}

		if beatMap == nil {
			log.Println("Beatmap not found, closing...")
			os.Exit(0)
		}

		if settings.VSplayer.Mods.EnableDT {
			// 开启DT
			settings.SPEED = 1.5
		}else if settings.VSplayer.Mods.EnableHT {
			settings.SPEED = 0.75
		}else {
			settings.SPEED = 1.0
		}

		// 开启EZ
		if settings.VSplayer.Mods.EnableEZ {
			beatMap.CircleSize = math.Min(beatMap.CircleSize * CS_EZ_HENSE, CS_MAX)
			beatMap.AR = math.Min(beatMap.AR * AR_EZ_HENSE, AR_MAX)
		}

		// 开启HR
		if settings.VSplayer.Mods.EnableHR {
			beatMap.CircleSize = math.Min(beatMap.CircleSize * CS_HR_HENSE, CS_MAX)
			beatMap.AR = math.Min(beatMap.AR * AR_HR_HENSE, AR_MAX)
		}

		// 开启HD，为维持HD效果，关闭一些特效
		if settings.VSplayer.Mods.EnableHD {
			settings.Objects.SliderMerge = false
		}

		if !settings.VSplayer.ReplayandCache.ReplayDebug {
			glfw.Init()
			glfw.WindowHint(glfw.ContextVersionMajor, 3)
			glfw.WindowHint(glfw.ContextVersionMinor, 3)
			glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
			glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
			glfw.WindowHint(glfw.Resizable, glfw.False)
			glfw.WindowHint(glfw.Samples, int(settings.Graphics.MSAA))

			var err error

			monitor := glfw.GetPrimaryMonitor()
			mWidth, mHeight := monitor.GetVideoMode().Width, monitor.GetVideoMode().Height

			if newSettings {
				log.Println(mWidth, mHeight)
				settings.Graphics.Width, settings.Graphics.Height = int64(mWidth), int64(mHeight)
				settings.Save()
				win, err = glfw.CreateWindow(mWidth, mHeight, "osu vs player", monitor, nil)
			} else {
				if settings.Graphics.Fullscreen {
					win, err = glfw.CreateWindow(int(settings.Graphics.Width), int(settings.Graphics.Height), "danser", monitor, nil)
				} else {
					win, err = glfw.CreateWindow(int(settings.Graphics.WindowWidth), int(settings.Graphics.WindowHeight), "danser", nil, nil)
				}
			}

			if err != nil {
				panic(err)
			}

			win.SetTitle("osu vs player " + VERSION + " by " + OWNER + " on " + beatMap.Artist + " - " + beatMap.Name + " [" + beatMap.Difficulty + "]")
			icon, _ := utils.LoadImage("assets/textures/dansercoin.png")
			icon2, _ := utils.LoadImage("assets/textures/dansercoin48.png")
			icon3, _ := utils.LoadImage("assets/textures/dansercoin24.png")
			icon4, _ := utils.LoadImage("assets/textures/dansercoin16.png")
			win.SetIcon([]image.Image{icon, icon2, icon3, icon4})

			win.MakeContextCurrent()
			log.Println("GLFW initialized!")
			glhf.Init()
			glhf.Clear(0, 0, 0, 1)

			batch := render.NewSpriteBatch()
			batch.Begin()
			batch.SetColor(1, 1, 1, 1)
			camera := bmath.NewCamera()
			camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
			camera.SetOrigin(bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
			camera.Update()
			batch.SetCamera(camera.GetProjectionView())

			file, _ := os.Open("assets/fonts/Roboto-Bold.ttf")
			dfont := font.LoadFont(file, 21)
			file.Close()
			file2, _ := os.Open("assets/fonts/Roboto-Black.ttf")
			font.LoadFont(file2, 20)
			file2.Close()

			dfont.Draw(batch, 0, 10, 32, "Loading...")

			batch.End()
			win.SwapBuffers()
			glfw.PollEvents()

			glfw.SwapInterval(0)
			if settings.Graphics.VSync {
				glfw.SwapInterval(1)
			}

			audio.Init()
			audio.LoadSamples()

			beatmap.ParseObjects(beatMap)
			beatMap.LoadCustomSamples()
		}
		player = states.NewPlayer(beatMap)

	})

	for !win.ShouldClose() {
		mainthread.Call(func() {
			gl.Enable(gl.MULTISAMPLE)
			gl.Disable(gl.DITHER)
			gl.Disable(gl.SCISSOR_TEST)
			gl.Viewport(0, 0, int32(settings.Graphics.GetWidth()), int32(settings.Graphics.GetHeight()))
			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			if player != nil {
				player.Draw(0)
			}

			if win.GetKey(glfw.KeyEscape) == glfw.Press{
				win.SetShouldClose(true)
			}

			if win.GetKey(glfw.KeyF2) == glfw.Press {

				if !pressed {
					utils.MakeScreenshot(*win)
				}

				pressed = true
			}

			if win.GetKey(glfw.KeyF2) == glfw.Release {
				pressed = false
			}

			if win.GetKey(glfw.KeyMinus) == glfw.Press {

				if !pressedM {
					if settings.DIVIDES > 1 {
						settings.DIVIDES -= 1
					}
				}

				pressedM = true
			}

			if win.GetKey(glfw.KeyMinus) == glfw.Release {
				pressedM = false
			}

			if win.GetKey(glfw.KeyEqual) == glfw.Press {

				if !pressedP {
					settings.DIVIDES += 1
				}

				pressedP = true
			}

			if win.GetKey(glfw.KeyEqual) == glfw.Release {
				pressedP = false
			}

			win.SwapBuffers()
			glfw.PollEvents()

			if win.ShouldClose(){
				player.Stop()
				win.Destroy()
			}
		})
	}
}
