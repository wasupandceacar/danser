package main

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
	"flag"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wieku/glhf"
	"image"
	"log"
	"math"
	"os"
	"fmt"
	"strconv"
)

var player *states.Player
var pressed = false
var pressedM = false
var pressedP = false

var settingsVersion = flag.Int("settings", 0, "")

func run() {
	var win *glfw.Window

	mainthread.Call(func() {

		artist := flag.String("artist", "", "")
		//title := flag.String("title", "Snow Drive(01.23)", "")
		//difficulty := flag.String("difficulty", "Arigatou", "")
		//title := flag.String("title", "Road of Resistance", "")
		//difficulty := flag.String("difficulty", "Crimson Rebellion", "")
		creator := flag.String("creator", "", "")
		cursors := flag.Int("cursors", 1, "")
		tag := flag.Int("tag", 1, "")
		pitch := flag.Float64("pitch", 1.0, "")
		mover := flag.String("mover", "linear", "")
		debug := flag.Bool("debug", false, "")
		fps := flag.Bool("fps", false, "")

		flag.Parse()


		settings.DEBUG = *debug
		settings.FPS = *fps
		settings.DIVIDES = *cursors
		settings.TAG = *tag
		settings.PITCH = *pitch
		_ = mover
		dance.SetMover(*mover)

		newSettings := settings.LoadSettings(*settingsVersion)

		player = nil
		var beatMap *beatmap.BeatMap = nil

		// 从设置重新载入map
		title := flag.String("title", settings.VSplayer.MapInfo.Title, "")
		difficulty := flag.String("difficulty", settings.VSplayer.MapInfo.Difficulty, "")

		if (*artist + *title + *difficulty + *creator) == "" {
			log.Println("No beatmap specified, closing...")
			os.Exit(0)
		}

		database.Init()
		beatmaps := database.LoadBeatmaps()

		for _, b := range beatmaps {
			if (*artist == "" || *artist == b.Artist) && (*title == "" || *title == b.Name) && (*difficulty == "" || *difficulty == b.Difficulty) && (*creator == "" || *creator == b.Creator) {
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
			settings.SPEED = *flag.Float64("speed", 1.5, "")
		}else if settings.VSplayer.Mods.EnableHT {
			settings.SPEED = *flag.Float64("speed", 0.75, "")
		}else {
			settings.SPEED = *flag.Float64("speed", 1.0, "")
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
			settings.Objects.SliderMerge = *flag.Bool("slidermerge", false, "")
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

			if win.GetKey(glfw.KeyEscape) == glfw.Press {
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
		})
	}
}

func main() {
	//mainthread.CallQueueCap = 100000
	//mainthread.Run(run)
	UImain()
}

func UImain() {
	// 首先载入设置
	settings.LoadSettings(*settingsVersion)

	vsw := &VSPlayerMainWindow{}

	vsw.SetFixedSize(true)

	if _, err := (MainWindow{
		AssignTo: &vsw.MainWindow,
		Title:    "VS-Player",
		Size: Size{900, 400},
		Layout:  VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					Composite{
						Layout: VBox{MarginsZero: true},
						Children: []Widget{
							GroupBox{
								Title: "玩家设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "玩家数：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: strconv.Itoa(settings.VSplayer.PlayerInfo.Players),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "指定玩家：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfo.SpecifiedPlayers,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "指定序列：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.PlayerInfo.SpecifiedLine,
											},
										},
									},
								},
							},
							GroupBox{
								Title: "玩家信息区设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "基准大小：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.PlayerInfoUI.BaseSize),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "基准X：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.PlayerInfoUI.BaseX),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "基准Y：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.PlayerInfoUI.BaseY),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示M1：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfoUI.ShowMouse1,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示M2：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfoUI.ShowMouse2,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示实时pp：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfoUI.ShowRealTimePP,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "最大变化时间：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.PlayerInfoUI.RealTimePPGap),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示实时ur：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfoUI.ShowRealTimeUR,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示数据排名：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfoUI.ShowPPAndURRank,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "强调第一：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerInfoUI.Rank1Highlight,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "强调放大倍数：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.PlayerInfoUI.HighlightMult),
											},
										},
									},
								},
							},
							VSpacer{},
						},
					},
					Composite{
						Layout: VBox{MarginsZero: true},
						Children: []Widget{
							GroupBox{
								Title: "录制信息设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "录制人：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.RecordInfoUI.Recorder,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "录制时间：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.RecordInfoUI.RecordTime,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "录制基准X：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.RecordInfoUI.RecordBaseX),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "录制基准Y：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.RecordInfoUI.RecordBaseY),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "录制基准大小：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.RecordInfoUI.RecordBaseSize),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "录制透明度：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.RecordInfoUI.RecordAlpha),
											},
										},
									},
								},
							},
							GroupBox{
								Title: "游玩区域设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "击打响应时间：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: strconv.Itoa(int(settings.VSplayer.PlayerFieldUI.HitFadeTime)),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "光标颜色索引：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: strconv.Itoa(settings.VSplayer.PlayerFieldUI.CursorColorNum),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "光标颜色间隔：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: strconv.Itoa(settings.VSplayer.PlayerFieldUI.CursorColorSkipNum),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示圈内数字：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.PlayerFieldUI.ShowHitCircleNumber,
											},
										},
									},
								},
							},
							GroupBox{
								Title: "地图设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "地图名：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.MapInfo.Title,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "难度名：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.MapInfo.Difficulty,
											},
										},
									},
								},
							},
							GroupBox{
								Title: "Mod设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "开启DT：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Mods.EnableDT,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "开启HT：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Mods.EnableHT,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "开启EZ：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Mods.EnableEZ,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "开启HR：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Mods.EnableHR,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "开启HD：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Mods.EnableHD,
											},
										},
									},
								},
							},
							VSpacer{},
						},
					},
					Composite{
						Layout: VBox{MarginsZero: true},
						Children: []Widget{
							GroupBox{
								Title: "淘汰模式设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "淘汰模式：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Knockout.EnableKnockout,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示真实miss：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Knockout.ShowTrueMiss,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "miss消失时间：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.Knockout.PlayerFadeTime),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "同位偏移：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.Knockout.SameTimeOffset),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "miss大小倍数：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.Knockout.PlayerFadeTime),
											},
										},
									},
								},
							},
							GroupBox{
								Title: "replay和cache设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "replay目录：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.ReplayandCache.ReplayDir,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "cache目录：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.ReplayandCache.CacheDir,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "保存cache：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.ReplayandCache.SaveResultCache,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "读取cache：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.ReplayandCache.ReadResultCache,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "Debug：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.ReplayandCache.ReplayDebug,
											},
										},
									},
								},
							},
							GroupBox{
								Title: "错误修正设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "错误修正：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.ErrorFix.EnableErrorFix,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "修正文件：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.ErrorFix.ErrorFixFile,
											},
										},
									},
								},
							},
							GroupBox{
								Title: "皮肤设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "自定义皮肤：",
											},
											CheckBox{
												Alignment: AlignHNearVNear,
												Checked: settings.VSplayer.Skin.EnableSkin,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "皮肤目录：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.Skin.SkinDir,
											},
										},
									},
								},
							},
							GroupBox{
								Title: "其他原生设置",
								Layout: VBox{},
								Children: []Widget{
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "光标大小：",
											},
											LineEdit{
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.Cursor.CursorSize),
											},
										},
									},
								},
							},
							VSpacer{},
						},
					},
				},
			},
			PushButton{
				Text: "开始",
				Alignment: AlignHCenterVFar,
				OnClicked: func() {
					mainthread.CallQueueCap = 100000
					mainthread.Run(run)
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}

type VSPlayerMainWindow struct {
	*walk.MainWindow
}