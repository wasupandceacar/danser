package states

//region 无关0

import (
	"danser/animation"
	"danser/audio"
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/dance"
	"danser/hitjudge"
	"danser/hitjudge/aop"
	. "danser/osuconst"
	"danser/render"
	"danser/render/effects"
	"danser/render/font"
	"danser/render/texture"
	"danser/replay"
	"danser/resultcache"
	"danser/score"
	"danser/settings"
	"danser/storyboard"
	"danser/utils"
	"fmt"
	"github.com/Mempler/rplpa"
	"github.com/flesnuk/oppai5"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/glhf"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var defaultPos = bmath.Vector2d{X: -1, Y: -1}

var noKeyPressed = rplpa.KeyPressed{
	LeftClick:  false,
	RightClick: false,
	Key1:       false,
	Key2:       false,
	Smoke:      false,
}

type Player struct {
	font *font.Font
	// 加粗字体
	highlightFont  *font.Font
	bMap           *beatmap.BeatMap
	queue2         []objects.BaseObject
	processed      []objects.Renderable
	sliderRenderer *render.SliderRenderer
	blurEffect     *effects.BlurEffect
	bloomEffect    *effects.BloomEffect
	lastTime       int64
	progressMsF    float64
	progressMs     int64
	batch          *render.SpriteBatch
	controller     []dance.Controller
	//circles        []*objects.Circle
	//sliders        []*objects.Slider
	Background  *texture.TextureRegion
	Logo        *texture.TextureRegion
	BgScl       bmath.Vector2d
	Scl         float64
	SclA        float64
	CS          float64
	fxRotation  float64
	fadeOut     float64
	fadeIn      float64
	entry       float64
	start       bool
	mus         bool
	musicPlayer *audio.Music
	fxBatch     *render.FxBatch
	vao         *glhf.VertexSlice
	vaoD        []float32
	vaoDirty    bool
	rotation    float64
	profiler    *utils.FPSCounter
	profilerU   *utils.FPSCounter

	storyboard *storyboard.Storyboard

	camera         *bmath.Camera
	scamera        *bmath.Camera
	dimGlider      *animation.Glider
	blurGlider     *animation.Glider
	fxGlider       *animation.Glider
	cursorGlider   *animation.Glider
	counter        float64
	fpsC           float64
	fpsU           float64
	storyboardLoad float64
	mapFullName    string

	// 偏移位置参数
	fontsize           float64
	highlightfontsize  float64
	missfontsize       float64
	misssize           float64
	keysize            float64
	gapsize            float64
	modoffset          float64
	missoffsetX        float64
	missoffsetY        float64
	lineoffset         float64
	hitoffset          float64
	key1baseX          float64
	key2baseX          float64
	key3baseX          float64
	key4baseX          float64
	accbaseX           float64
	rankbaseX          float64
	ppurbaseX          float64
	ppurrankbaseX      float64
	playerbaseX        float64
	keybaseY           float64
	fontbaseY          float64
	highlightfontbaseY float64
	rankbaseY          float64
	hitbaseY           float64

	recordbaseX       float64
	recordbaseY       float64
	recordbasesize    float64
	recordtimeoffsetY float64

	diffbaseX    float64
	diffbaseY    float64
	diffbasesize float64
	diffoffsetY  float64

	// 色彩参数
	objectcolorIndex int

	// 偏移参数
	lastDishowPos bmath.Vector2d
	SameRate      int
	lastMissPos   bmath.Vector2d
	SameMissRate  int

	// player人数
	playerCount int

	// 实时pp显示参数数组
	lastb4PP   []float64
	lastPP     []float64
	lastPPTime []int64

	// 实时ur显示参数数组
	lastb4UR   []float64
	lastUR     []float64
	lastURTime []int64

	// 退出所有协程的flag
	exitGoFlag bool
	// 退出处理event协程
	exitPollFlag bool

	// 实时难度数组
	difficulties []oppai.PP
	// 现在物件指针
	objindex int

	// 指定色彩
	specificColorMap map[int][3]float32
}

//endregion

func NewPlayer(beatMap *beatmap.BeatMap, win *glfw.Window, loadwords []font.Word) *Player {
	//region 无关1
	player := new(Player)

	// 设置协程flag
	player.exitGoFlag = false
	player.exitPollFlag = false

	go func() {
		for win != nil && !win.ShouldClose() {
			if player.exitGoFlag || player.exitPollFlag {
				return
			}
			glfw.PollEvents()
		}
	}()

	// 重置时间
	utils.ResetTime()

	// 非replay debug
	if !settings.VSplayer.ReplayandCache.ReplayDebug {
		player.batch = render.NewSpriteBatch()
		player.font = font.GetFont("Roboto Bold")
		player.highlightFont = font.GetFont("Roboto Black")

		player.scamera = bmath.NewCamera()
		player.scamera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
		player.scamera.SetOrigin(bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
		player.scamera.Update()
		player.batch.Begin()
		player.batch.SetColor(1, 1, 1, 1)
		player.batch.SetCamera(player.scamera.GetProjectionView())

		render.LoadTextures()
		render.LoadSkinConfiguration()

		loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: "Textures and skin configuration loaded..."})
		player.font.DrawAll(player.batch, loadwords)
		player.batch.End()
		win.SwapBuffers()

		render.SetupSlider()
		player.sliderRenderer = render.NewSliderRenderer()

		player.bMap = beatMap

		player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
		log.Println("Playing:", player.mapFullName)

		player.batch.Begin()
		loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: "Map: " + player.mapFullName})
		player.font.DrawAll(player.batch, loadwords)
		player.batch.End()
		win.SwapBuffers()

		player.CS = (1.0 - 0.7*(beatMap.CircleSize-5)/5) / 2 * settings.Objects.CSMult
		render.CS = player.CS

		var err error
		player.Background, err = utils.LoadTextureToAtlas(render.Atlas, filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
		if err != nil {
			log.Println(err)
		}

		if settings.Playfield.StoryboardEnabled {
			player.storyboard = storyboard.NewStoryboard(player.bMap)

			if player.storyboard == nil {
				log.Println("Storyboard not found!")
			}
		}

		//player.Logo, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/logo-medium.png")

		if err != nil {
			log.Println(err)
		}

		winscl := settings.Graphics.GetAspectRatio()

		player.blurEffect = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

		if player.Background != nil {
			imScl := float64(player.Background.Width) / float64(player.Background.Height)

			condition := imScl < winscl
			if player.storyboard != nil && !player.storyboard.IsWideScreen() {
				condition = !condition
			}

			if condition {
				player.BgScl = bmath.NewVec2d(1, winscl/imScl)
			} else {
				player.BgScl = bmath.NewVec2d(imScl/winscl, 1)
			}
		}

		scl := (settings.Graphics.GetHeightF() * 900.0 / 1080.0) / PLAYFIELD_HEIGHT * settings.Playfield.Scale

		osuAspect := PLAYFIELD_WIDTH / PLAYFIELD_HEIGHT
		screenAspect := settings.Graphics.GetWidthF() / settings.Graphics.GetHeightF()

		if osuAspect > screenAspect {
			scl = (settings.Graphics.GetWidthF() * 900.0 / 1080.0) / PLAYFIELD_WIDTH * settings.Playfield.Scale
		}

		player.camera = bmath.NewCamera()
		player.camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true)
		player.camera.SetOrigin(bmath.NewVec2d(PLAYFIELD_WIDTH/2, PLAYFIELD_HEIGHT/2))
		player.camera.SetScale(bmath.NewVec2d(scl, scl))
		player.camera.Update()

		render.Camera = player.camera

		player.bMap.Reset()
	} else {
		log.Println("开始Debug Replay")
	}

	//endregion

	//region player初始化
	var tmpplindex []int
	if settings.VSplayer.PlayerInfo.SpecifiedPlayers {
		specifiedplayers := strings.Split(settings.VSplayer.PlayerInfo.SpecifiedLine, ",")
		for _, player := range specifiedplayers {
			pl, _ := strconv.Atoi(player)
			if pl <= 0 {
				log.Panic("指定player的字符串有误，请重新检查设定")
			}
			tmpplindex = append(tmpplindex, pl)
		}
		log.Println("本次已指定特定的player")
		player.playerCount = len(specifiedplayers)
	} else {
		player.playerCount = settings.VSplayer.PlayerInfo.Players
	}

	if !settings.VSplayer.ReplayandCache.ReplayDebug {
		player.controller = make([]dance.Controller, player.playerCount)
		for k := 0; k < player.playerCount; k++ {
			player.controller[k] = dance.NewReplayController()
			player.controller[k].SetBeatMap(player.bMap)
			player.controller[k].InitCursors()
		}
	}
	//endregion

	//region replay处理

	// 读取replay
	replays, err := replay.GetOsrFiles()
	if err != nil {
		panic(err)
	}
	// 解析每个replay的判定
	t := time.Now()

	// 如果debug replay，记录整体的replay结果
	// 正确和错误的replay个数
	right := 0
	wrong := 0
	// 错误的replay编号
	var wrongIndex []int

	//TODO: Setting up error correction system
	var errs []hitjudge.Error
	if settings.VSplayer.ErrorFix.EnableErrorFix {
		log.Println("本次选择进行replay解析纠错")
		errs = hitjudge.ReadError()
	} else {
		errs = []hitjudge.Error{}
	}

	//Information: How to calculate ReplayIndex
	/*
		var rnum int
		if settings.VSplayer.PlayerInfo.SpecifiedPlayers {
			rnum = tmpplindex[k]
		} else {
			rnum = k + 1
		}

		then

		hitjudge.FilterError(rnum, errs)
	*/

	//TODO: Setting up specific player system
	if settings.VSplayer.ReplayandCache.UseCacheSystem && !settings.VSplayer.ReplayandCache.ReplayDebug {
		log.Println("Enabled Cache System")
		for i := 0; i < player.playerCount; i++ {
			rep := replay.ExtractReplay(replays[i])
			objectResult, totalResult, exists := resultcache.GetResult(rep)
			if exists {
				log.Printf("Reading %v's analyze cache %v.ooc/otc (%v/%v)...\n", rep.Username, rep.ReplayMD5, i+1, player.playerCount)
				if !settings.VSplayer.ReplayandCache.ReplayDebug {
					player.batch.Begin()
					loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: fmt.Sprintf("Reading %v's analyze cache %v.ooc/otc (%v/%v)...", rep.Username, rep.ReplayMD5, i+1, player.playerCount)})
					player.font.DrawAll(player.batch, loadwords)
					player.batch.End()
					win.SwapBuffers()
				}
				//------------------------------------
				if !settings.VSplayer.ReplayandCache.ReplayDebug {
					configurePlayer(player, i, rep, objectResult, totalResult)
					loadwords = loadwords[:len(loadwords)-1]
				}
				//------------------------------------
				log.Printf("Finished reading %v analyze cache %v.ooc/otc (%v/%v)\n", rep.Username, rep.ReplayMD5, i+1, player.playerCount)
			} else {
				log.Printf("Falling back to analyze %v's replay (%v/%v)...\n", rep.Username, i+1, player.playerCount)
				if !settings.VSplayer.ReplayandCache.ReplayDebug {
					player.batch.Begin()
					loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: fmt.Sprintf("Analyzing %v's replay (%v/%v)...", rep.Username, i+1, player.playerCount)})
					player.font.DrawAll(player.batch, loadwords)
					player.batch.End()
					win.SwapBuffers()
				}
				//------------------------------------
				t1 := time.Now()
				objectResult, totalResult, _, _ := hitjudge.ParseHits(settings.General.OsuSongsDir+beatMap.Dir+"/"+beatMap.File, rep, errs, NO_USE_CS_OFFSET)
				_, totalResult2 := aop.Judge(beatMap, rep)
				log.Printf("Result from the original : %v-%v-%v-%v (%v) (%v)\n", totalResult[len(totalResult)-1].N300, totalResult[len(totalResult)-1].N100, totalResult[len(totalResult)-1].N50, totalResult[len(totalResult)-1].Misses, totalResult[len(totalResult)-1].Acc, totalResult[len(totalResult)-1].Combo)
				log.Printf("Result from the NEW : %v-%v-%v-%v (%v) (%v)\n", totalResult2[len(totalResult2)-1].N300, totalResult2[len(totalResult2)-1].N100, totalResult2[len(totalResult2)-1].N50, totalResult2[len(totalResult2)-1].Misses, totalResult2[len(totalResult2)-1].Acc, totalResult2[len(totalResult2)-1].Combo)
				if !settings.VSplayer.ReplayandCache.ReplayDebug {
					configurePlayer(player, i, rep, objectResult, totalResult)
					resultcache.CacheResult(objectResult, totalResult, rep)
					loadwords = loadwords[:len(loadwords)-1]
				}
				//------------------------------------
				log.Printf("Finished analyzing %v's replay (%v/%v), elapsed time: %v, total elapsed time: %v\n", rep.Username, i+1, player.playerCount, time.Now().Sub(t1), time.Now().Sub(t))
			}
		}
	} else {
		log.Println("Forced to analyze replays")
		for i := 0; i < player.playerCount; i++ {
			var rnum int
			if settings.VSplayer.PlayerInfo.SpecifiedPlayers {
				rnum = tmpplindex[i]
			} else {
				rnum = i + 1
			}

			rep := replay.ExtractReplay(replays[i])
			log.Printf("Analyzing %v's replay (%v/%v)...\n", rep.Username, i+1, player.playerCount)
			if !settings.VSplayer.ReplayandCache.ReplayDebug {
				player.batch.Begin()
				loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: fmt.Sprintf("Analyzing %v's replay (%v/%v)...", rep.Username, i+1, player.playerCount)})
				player.font.DrawAll(player.batch, loadwords)
				player.batch.End()
				win.SwapBuffers()
			}
			//------------------------------------
			t1 := time.Now()
			objectResult, totalResult, allRight, _ := hitjudge.ParseHits(settings.General.OsuSongsDir+beatMap.Dir+"/"+beatMap.File, rep, errs, NO_USE_CS_OFFSET)
			_, totalResult2 := aop.Judge(beatMap, rep)
			log.Printf("Result from the original : %v-%v-%v-%v (%v) (%v)\n", totalResult[len(totalResult)-1].N300, totalResult[len(totalResult)-1].N100, totalResult[len(totalResult)-1].N50, totalResult[len(totalResult)-1].Misses, totalResult[len(totalResult)-1].Acc, totalResult[len(totalResult)-1].Combo)
			log.Printf("Result from the NEW : %v-%v-%v-%v (%v) (%v)\n", totalResult2[len(totalResult2)-1].N300, totalResult2[len(totalResult2)-1].N100, totalResult2[len(totalResult2)-1].N50, totalResult2[len(totalResult2)-1].Misses, totalResult2[len(totalResult2)-1].Acc, totalResult2[len(totalResult2)-1].Combo)
			if !settings.VSplayer.ReplayandCache.ReplayDebug {
				configurePlayer(player, i, rep, objectResult, totalResult)
				loadwords = loadwords[:len(loadwords)-1]
			} else {
				// 记录出错情况
				if allRight {
					right += 1
				} else {
					wrong += 1
					wrongIndex = append(wrongIndex, rnum)
				}
			}
			//------------------------------------
			log.Printf("Finished analyzing %v's replay (%v/%v), elapsed time: %v, total elapsed time: %v \n", rep.Username, i+1, player.playerCount, time.Now().Sub(t1), time.Now().Sub(t))
		}
	}

	//TODO: Replay debug system
	if settings.VSplayer.ReplayandCache.ReplayDebug {
		// 总体replay分析情况
		log.Println("正确结果：", right, " 个")
		log.Println("错误结果：", wrong, " 个")
		log.Println("错误编号：", wrongIndex)
		// 直接退出，不进行下面的渲染任务
		log.Println("Debug Replay 结束，直接退出")
		os.Exit(0)
	} else {
		player.batch.Begin()
		loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: "Analyze completed."})
		player.font.DrawAll(player.batch, loadwords)
		player.batch.End()
		win.SwapBuffers()
	}

	// 指定色彩设置
	player.specificColorMap = make(map[int][3]float32)
	specifiedplcolors := strings.Split(settings.VSplayer.PlayerInfo.SpecifiedColor, "|")
	for _, plcolors := range specifiedplcolors {
		tmpplcolors := strings.Split(plcolors, ":")
		specifiedplayer := tmpplcolors[0]
		for k := 0; k < player.playerCount; k++ {
			if player.controller[k].GetPlayname() == specifiedplayer {
				colors := strings.Split(tmpplcolors[1], ",")
				r, _ := strconv.Atoi(colors[0])
				g, _ := strconv.Atoi(colors[1])
				b, _ := strconv.Atoi(colors[2])
				player.specificColorMap[k] = [3]float32{
					float32(r) / 255,
					float32(g) / 255,
					float32(b) / 255,
				}
				log.Println("player 名："+specifiedplayer+"，指定 replay", k+1, "的颜色为", r, g, b)
				break
			}
		}
	}

	// 初始化实时pp、ur参数数组
	if settings.VSplayer.PlayerInfoUI.ShowRealTimePP {
		for k := 0; k < player.playerCount; k++ {
			player.lastPPTime = make([]int64, player.playerCount)
			player.lastPPTime[k] = player.controller[k].GetHitResult()[0].JudgeTime
		}
	}
	for k := 0; k < player.playerCount; k++ {
		player.lastb4PP = make([]float64, player.playerCount)
		player.lastPP = make([]float64, player.playerCount)
		player.lastb4PP[k] = player.controller[k].GetTotalResult()[0].PP.Total
		player.lastPP[k] = player.controller[k].GetTotalResult()[0].PP.Total
	}
	if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
		for k := 0; k < player.playerCount; k++ {
			player.lastb4UR = make([]float64, player.playerCount)
			player.lastUR = make([]float64, player.playerCount)
			player.lastURTime = make([]int64, player.playerCount)
			player.lastb4UR[k] = player.controller[k].GetTotalResult()[0].UR
			player.lastUR[k] = player.controller[k].GetTotalResult()[0].UR
			player.lastURTime[k] = player.controller[k].GetHitResult()[0].JudgeTime
		}
	}

	//endregion

	//region 无关11

	player.lastTime = -1
	player.queue2 = make([]objects.BaseObject, len(player.bMap.Queue))
	copy(player.queue2, player.bMap.Queue)

	log.Println("Music:", beatMap.Audio)
	player.batch.Begin()
	loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: "Music: " + beatMap.Audio})
	player.font.DrawAll(player.batch, loadwords)
	player.batch.End()
	win.SwapBuffers()

	player.Scl = 1
	player.fxRotation = 0.0
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.dimGlider = animation.NewGlider(0.0)
	player.blurGlider = animation.NewGlider(0.0)
	player.fxGlider = animation.NewGlider(0.0)
	player.cursorGlider = animation.NewGlider(0.0)

	tmS := float64(player.queue2[0].GetBasicData().StartTime)
	tmE := float64(player.queue2[len(player.queue2)-1].GetBasicData().EndTime)

	player.dimGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.BackgroundInDim)
	player.blurGlider.AddEvent(-1500, -1000, settings.Playfield.BackgroundInBlur)
	player.fxGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.SpectrumInDim)
	player.cursorGlider.AddEvent(-1500, -1000, 0.0)

	// 开启storyborad，则关闭背景模糊，背景暗化设为0.55
	if settings.Playfield.StoryboardEnabled {
		settings.Playfield.BackgroundBlur = 0.0
		settings.Playfield.BackgroundBlurBreaks = 0.0
		settings.Playfield.BackgroundDim = 0.55
	}

	player.dimGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.BackgroundDim)
	player.blurGlider.AddEvent(tmS-750, tmS-250, settings.Playfield.BackgroundBlur)
	player.fxGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.SpectrumDim)
	player.cursorGlider.AddEvent(tmS-750, tmS-250, 1.0)

	fadeOut := settings.Playfield.FadeOutTime * 1000
	player.dimGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.fxGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.cursorGlider.AddEvent(tmE, tmE+fadeOut, 0.0)

	for _, p := range beatMap.Pauses {
		bd := p.GetBasicData()

		if bd.EndTime-bd.StartTime < 1000 {
			continue
		}

		player.dimGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 1.0-settings.Playfield.BackgroundDimBreaks)
		player.blurGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, settings.Playfield.BackgroundBlurBreaks)
		player.fxGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 1.0-settings.Playfield.SpectrumDimBreaks)
		if !settings.Cursor.ShowCursorsOnBreaks {
			player.cursorGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+100, 0.0)
		}

		player.dimGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0-settings.Playfield.BackgroundDim)
		player.blurGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), settings.Playfield.BackgroundBlur)
		player.fxGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0-settings.Playfield.SpectrumDim)
		player.cursorGlider.AddEvent(float64(bd.EndTime)-100, float64(bd.EndTime), 1.0)
	}

	musicPlayer := audio.NewMusic(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Audio))

	//endregion

	//region 计算大小偏移位置常量、色彩常量

	player.fontsize = 1.75 * settings.VSplayer.PlayerInfoUI.BaseSize
	if settings.VSplayer.PlayerInfoUI.Rank1Highlight {
		player.highlightfontsize = settings.VSplayer.PlayerInfoUI.HighlightMult * player.fontsize
	}
	player.missfontsize = settings.VSplayer.Knockout.MissMult * player.fontsize
	player.misssize = 1.5 * settings.VSplayer.Knockout.MissMult * settings.VSplayer.PlayerInfoUI.BaseSize
	player.keysize = 1.25 * settings.VSplayer.PlayerInfoUI.BaseSize
	player.gapsize = settings.VSplayer.PlayerInfoUI.LineGapMult * settings.VSplayer.PlayerInfoUI.BaseSize
	player.modoffset = settings.VSplayer.PlayerInfoUI.BaseSize
	player.missoffsetX = 2 * settings.VSplayer.Knockout.MissMult * settings.VSplayer.PlayerInfoUI.BaseSize
	player.missoffsetY = 0.6 * settings.VSplayer.Knockout.MissMult * settings.VSplayer.PlayerInfoUI.BaseSize
	player.lineoffset = 2.25 * settings.VSplayer.PlayerInfoUI.BaseSize
	player.hitoffset = 1.75 * settings.VSplayer.PlayerInfoUI.BaseSize
	player.key1baseX = settings.VSplayer.PlayerInfoUI.BaseX
	player.key2baseX = player.key1baseX + 2*player.keysize
	player.key3baseX = player.key2baseX + 2*player.keysize
	player.key4baseX = player.key3baseX + 2*player.keysize
	if settings.VSplayer.PlayerInfoUI.ShowMouse1 {
		if settings.VSplayer.PlayerInfoUI.ShowMouse2 {
			player.accbaseX = player.key4baseX + 2*settings.VSplayer.PlayerInfoUI.BaseSize
		} else {
			player.accbaseX = player.key3baseX + 2*settings.VSplayer.PlayerInfoUI.BaseSize
		}
	} else {
		player.accbaseX = player.key2baseX + 2*settings.VSplayer.PlayerInfoUI.BaseSize
	}
	if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
		player.rankbaseX = player.accbaseX + 2.625*settings.VSplayer.PlayerInfoUI.BaseSize
		player.ppurbaseX = player.accbaseX + 8.375*settings.VSplayer.PlayerInfoUI.BaseSize
	} else {
		player.rankbaseX = player.accbaseX + 8.375*settings.VSplayer.PlayerInfoUI.BaseSize
		player.ppurbaseX = player.rankbaseX + 1.625*settings.VSplayer.PlayerInfoUI.BaseSize
	}
	if settings.VSplayer.PlayerInfoUI.ShowPPAndURRank {
		player.ppurrankbaseX = player.ppurbaseX + 9.125*settings.VSplayer.PlayerInfoUI.BaseSize
		player.playerbaseX = player.ppurrankbaseX + 4.5*settings.VSplayer.PlayerInfoUI.BaseSize
	} else {
		player.playerbaseX = player.ppurbaseX + 9*settings.VSplayer.PlayerInfoUI.BaseSize
	}
	player.keybaseY = settings.VSplayer.PlayerInfoUI.BaseY
	player.fontbaseY = settings.VSplayer.PlayerInfoUI.BaseY - 0.75*settings.VSplayer.PlayerInfoUI.BaseSize
	if settings.VSplayer.PlayerInfoUI.Rank1Highlight {
		player.highlightfontbaseY = player.fontbaseY - (settings.VSplayer.PlayerInfoUI.HighlightMult-1)*player.fontsize/2
	}
	player.rankbaseY = settings.VSplayer.PlayerInfoUI.BaseY - 0.25*settings.VSplayer.PlayerInfoUI.BaseSize
	player.hitbaseY = settings.VSplayer.PlayerInfoUI.BaseY - 0.1*settings.VSplayer.PlayerInfoUI.BaseSize

	player.recordbaseX = settings.VSplayer.RecordInfoUI.RecordBaseX
	player.recordbaseY = settings.VSplayer.RecordInfoUI.RecordBaseY
	player.recordbasesize = settings.VSplayer.RecordInfoUI.RecordBaseSize
	player.recordtimeoffsetY = 1.25 * player.recordbasesize

	if settings.VSplayer.DiffInfoUI.ShowDiffInfo {
		player.diffbaseX = settings.VSplayer.DiffInfoUI.DiffBaseX
		player.diffbaseY = settings.VSplayer.DiffInfoUI.DiffBaseY
		player.diffbasesize = settings.VSplayer.DiffInfoUI.DiffBaseSize
		player.diffoffsetY = 1.25 * player.recordbasesize
	}

	// 超过色彩上限使用最后一个（未使用）的颜色来渲染object
	if settings.VSplayer.PlayerFieldUI.CursorColorNum > player.playerCount+1 {
		player.objectcolorIndex = player.playerCount
	} else {
		player.objectcolorIndex = settings.VSplayer.PlayerFieldUI.CursorColorNum - 1
	}

	player.lastDishowPos = bmath.Vector2d{X: -1, Y: -1}
	player.SameRate = 0
	player.lastMissPos = bmath.Vector2d{X: -1, Y: -1}
	player.SameMissRate = 0

	//endregion

	//region 计算实时难度

	if settings.VSplayer.DiffInfoUI.ShowDiffInfo {
		log.Println("开始计算实时难度")
		t = time.Now()
		// 计算mods
		mods := 0
		if settings.VSplayer.Mods.EnableDT {
			mods += MOD_DT
		}
		if settings.VSplayer.Mods.EnableHR {
			mods += MOD_HR
		}
		if settings.VSplayer.Mods.EnableHT {
			mods += MOD_HT
		}
		if settings.VSplayer.Mods.EnableHD {
			mods += MOD_HD
		}
		if settings.VSplayer.Mods.EnableEZ {
			mods += MOD_EZ
		}
		beatmapLength := len(beatMap.HitObjects)
		player.difficulties = make([]oppai.PP, beatmapLength)
		if !settings.VSplayer.ReplayandCache.ReplayDebug {
			player.batch.Begin()
			loadwords = append(loadwords, font.Word{X: 14, Size: 24, Text: "Calculate realtime difficulty... 0/" + strconv.Itoa(beatmapLength)})
			player.font.DrawAll(player.batch, loadwords)
			player.batch.End()
			win.SwapBuffers()
		}
		for k := 0; k < beatmapLength; k++ {
			player.difficulties[k] = score.CalculateDiffbyNum(settings.General.OsuSongsDir+beatMap.Dir+"/"+beatMap.File, k+1, uint32(mods))
			if !settings.VSplayer.ReplayandCache.ReplayDebug {
				player.batch.Begin()
				loadwords = append(loadwords[:len(loadwords)-1], font.Word{X: 14, Size: 24, Text: "Calculate realtime difficulty... " + strconv.Itoa(k+1) + "/" + strconv.Itoa(beatmapLength)})
				player.font.DrawAll(player.batch, loadwords)
				player.batch.End()
				win.SwapBuffers()
			}
		}
		log.Println("计算实时难度完成，耗时", time.Now().Sub(t))
	}

	//endregion

	//region 音乐？

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime * float64(time.Second)))

		start := -2000.0
		for i := 1; i <= 100; i++ {
			player.entry = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(time.Duration(settings.Playfield.LeadInHold * float64(time.Second)))

		for i := 1; i <= 100; i++ {
			player.fadeIn = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
			time.Sleep(10 * time.Millisecond)
		}

		player.start = true
		if !player.exitGoFlag {
			musicPlayer.Play()
			musicPlayer.SetTempo(settings.SPEED)
			musicPlayer.SetPitch(settings.PITCH)
		}
	}()

	player.fxBatch = render.NewFxBatch()
	player.vao = player.fxBatch.CreateVao(2 * 3 * (256 + 128))
	player.profilerU = utils.NewFPSCounter(60, false)

	//endregion

	//region 重写更新时间和坐标函数

	for k := 0; k < player.playerCount; k++ {
		go func(k int) {
			// 获取replay信息
			r := replay.ExtractReplay(replays[k])
			index := 3

			// 开始时间
			r0 := *r.ReplayData[0]
			r1 := *r.ReplayData[1]
			r2 := *r.ReplayData[2]
			start := r0.Time + r1.Time + r2.Time

			var last = musicPlayer.GetPosition()
			for {
				if player.exitGoFlag {
					return
				}

				if len(r.ReplayData) <= index {
					time.Sleep(1000 * time.Second)
				}

				// 获取第index个replay数据
				rdata := *r.ReplayData[index]
				offset := rdata.Time
				posX := rdata.MosueX
				posY := rdata.MouseY
				PressKey := *rdata.KeyPressed

				if index == 3 {
					offset += start
				}

				progressMsF := musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)

				//真实的offset
				trueOffset := progressMsF - last

				lateOffset := 0.0

				if offset == REPLAY_END_TIME {
					// 如果offset=-12345，replay结束，设置成最后的光标位置
					// 如果是HR且图整体未开HR，上下翻转
					if !settings.VSplayer.Mods.EnableHR && player.controller[k].GetMods()&MOD_HR > 0 {
						player.controller[k].Update(int64(progressMsF), trueOffset, bmath.NewVec2d(float64(r.ReplayData[index-1].MosueX), float64(PLAYFIELD_HEIGHT-r.ReplayData[index-1].MouseY)))
					} else {
						player.controller[k].Update(int64(progressMsF), trueOffset, bmath.NewVec2d(float64(r.ReplayData[index-1].MosueX), float64(r.ReplayData[index-1].MouseY)))
					}

					// 按键改为无
					player.controller[k].SetPresskey(noKeyPressed)

					// 修正last
					last += trueOffset

					lateOffset = 0.0
				} else if trueOffset >= float64(offset) {
					// 如果真实offset大于等于读到的offset，更新
					// 如果是HR且图整体未开HR，上下翻转
					if !settings.VSplayer.Mods.EnableHR && player.controller[k].GetMods()&MOD_HR > 0 {
						player.controller[k].Update(int64(progressMsF), trueOffset, bmath.NewVec2d(float64(posX), float64(PLAYFIELD_HEIGHT-posY)))
					} else {
						player.controller[k].Update(int64(progressMsF), trueOffset, bmath.NewVec2d(float64(posX), float64(posY)))
					}

					if offset != 0 {
						player.controller[k].SetPresskey(PressKey)
					}

					// 修正last
					last += float64(offset)

					lateOffset = 0.0

					index++
				} else if trueOffset > 50 {
					// 超过 50ms 未更新， 自动更新为上一个位置
					if !settings.VSplayer.Mods.EnableHR && player.controller[k].GetMods()&MOD_HR > 0 {
						player.controller[k].Update(int64(progressMsF), trueOffset-lateOffset, bmath.NewVec2d(float64(r.ReplayData[index-1].MosueX), float64(PLAYFIELD_HEIGHT-r.ReplayData[index-1].MouseY)))
					} else {
						player.controller[k].Update(int64(progressMsF), trueOffset-lateOffset, bmath.NewVec2d(float64(r.ReplayData[index-1].MosueX), float64(r.ReplayData[index-1].MouseY)))
					}

					lateOffset = trueOffset
				}

				time.Sleep(time.Millisecond)
			}
		}(k)
	}

	//endregion

	//region 独立绘图

	go func() {
		for {
			if player.exitGoFlag {
				return
			}

			player.progressMsF = musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)
			player.bMap.Update(int64(player.progressMsF))

			if player.storyboard != nil {
				player.storyboard.Update(int64(player.progressMsF))
			}

			if player.start && len(player.bMap.Queue) > 0 {
				player.dimGlider.Update(player.progressMsF)
				player.blurGlider.Update(player.progressMsF)
				player.fxGlider.Update(player.progressMsF)
				player.cursorGlider.Update(player.progressMsF)
			}
			time.Sleep(time.Millisecond)
		}
	}()

	//endregion

	//region 无关2

	go func() {
		vertices := make([]float32, (256+128)*3*3*2)
		oldFFT := make([]float32, 256+128)
		for {
			if player.exitGoFlag {
				return
			}

			musicPlayer.Update()
			player.SclA = math.Min(1.4*settings.Beat.BeatScale, math.Max(math.Sin(musicPlayer.GetBeat()*math.Pi/2)*0.4*settings.Beat.BeatScale+1.0, 1.0))

			fft := musicPlayer.GetFFT()

			for i := 0; i < len(oldFFT); i++ {
				fft[i] = fft[i] * float32(math.Pow(float64(i+1), 0.33))
				oldFFT[i] = float32(math.Max(0.001, math.Max(math.Min(float64(fft[i]), float64(oldFFT[i])+0.05), float64(oldFFT[i])-0.025)))

				vI := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, 0.005)
				vI2 := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, 0.5)

				poH := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, float64(oldFFT[i]))

				pLL := vI.Rotate(math.Pi / 2).Add(vI2).Sub(poH.Scl(0.5))
				pLR := vI.Rotate(-math.Pi / 2).Add(vI2).Sub(poH.Scl(0.5))
				pHL := vI.Rotate(math.Pi / 2).Add(poH.Scl(0.5)).Add(vI2)
				pHR := vI.Rotate(-math.Pi / 2).Add(poH.Scl(0.5)).Add(vI2)

				vertices[(i)*18], vertices[(i)*18+1], vertices[(i)*18+2] = pLL.X32(), pLL.Y32(), 0
				vertices[(i)*18+3], vertices[(i)*18+4], vertices[(i)*18+5] = pLR.X32(), pLR.Y32(), 0
				vertices[(i)*18+6], vertices[(i)*18+7], vertices[(i)*18+8] = pHR.X32(), pHR.Y32(), 0
				vertices[(i)*18+9], vertices[(i)*18+10], vertices[(i)*18+11] = pHR.X32(), pHR.Y32(), 0
				vertices[(i)*18+12], vertices[(i)*18+13], vertices[(i)*18+14] = pHL.X32(), pHL.Y32(), 0
				vertices[(i)*18+15], vertices[(i)*18+16], vertices[(i)*18+17] = pLL.X32(), pLL.Y32(), 0

			}

			player.vaoD = vertices
			player.vaoDirty = true

			time.Sleep(15 * time.Millisecond)
		}
	}()
	player.profiler = utils.NewFPSCounter(60, false)
	player.musicPlayer = musicPlayer

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	player.exitPollFlag = true

	return player

	//endregion
}

func configurePlayer(player *Player, k int, r *rplpa.Replay, result []hitjudge.ObjectResult, totalresult []hitjudge.TotalResult) {
	// 初始化acc、rank和pp
	player.controller[k].SetAcc(DEFAULT_ACC)
	if score.IsSilver(r.Mods) {
		player.controller[k].SetRank(*render.RankXH)
	} else {
		player.controller[k].SetRank(*render.RankX)
	}
	// 设置player名
	player.controller[k].SetPlayername(r.Username)
	// 判断mod
	player.controller[k].SetMods(r.Mods)
	player.controller[k].SetPP(DEFAULT_PP)
	if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
		player.controller[k].SetUR(DEFAULT_UR)
	}
	// 设置初始显示
	player.controller[k].SetIsShow(true)
	player.controller[k].SetHitResult(result)
	player.controller[k].SetTotalResult(totalresult)
}

func (pl *Player) Show() {

}

func (pl *Player) Draw(_ float64) {

	//region 无关3

	if pl.lastTime < 0 {
		pl.lastTime = utils.GetNanoTime()
	}
	tim := utils.GetNanoTime()
	timMs := float64(tim-pl.lastTime) / 1000000.0

	pl.profiler.PutSample(1000.0 / timMs)
	//fps := pl.profiler.GetFPS()

	if pl.start {

		//if fps > 58 && timMs > 17 {
		//	log.Println("Slow frame detected! Frame time:", timMs, "| Av. frame time:", 1000.0/fps)
		//}

		pl.progressMs = int64(pl.progressMsF)

		if pl.Scl < pl.SclA {
			pl.Scl += (pl.SclA - pl.Scl) * timMs / 100
		} else if pl.Scl > pl.SclA {
			pl.Scl -= (pl.Scl - pl.SclA) * timMs / 100
		}

	}

	pl.lastTime = tim

	if len(pl.queue2) > 0 {
		for i := 0; i < len(pl.queue2); i++ {
			if p := pl.queue2[i]; p.GetBasicData().StartTime-15000 <= pl.progressMs {
				if s, ok := p.(*objects.Slider); ok {
					s.InitCurve(pl.sliderRenderer, pl.exitGoFlag)
				}

				if p := pl.queue2[i]; p.GetBasicData().StartTime-int64(pl.bMap.ARms) <= pl.progressMs {

					pl.processed = append(pl.processed, p.(objects.Renderable))

					pl.queue2 = pl.queue2[1:]
					i--
				}
			} else {
				break
			}
		}
	}

	pl.fxRotation += timMs / 125
	if pl.fxRotation >= 360.0 {
		pl.fxRotation -= 360.0
	}

	// 结束标志
	if len(pl.bMap.Queue) == 0 {
		pl.fadeOut -= timMs / (settings.Playfield.FadeOutTime * 1000)
		pl.fadeOut = math.Max(0.0, pl.fadeOut)
		pl.musicPlayer.SetVolumeRelative(pl.fadeOut)
		//pl.dimGlider.UpdateD(timMs)
		//pl.blurGlider.UpdateD(timMs)
		////pl.fxGlider.UpdateD(timMs)
		//pl.cursorGlider.UpdateD(timMs)
	}

	render.CS = pl.CS
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	bgAlpha := pl.dimGlider.GetValue()
	blurVal := 0.0

	cameras := pl.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))

	breakcamera := pl.camera.GenRotatedX(2, math.Pi)[1]

	if settings.Playfield.BlurEnable {
		blurVal = pl.blurGlider.GetValue()
		if settings.Playfield.UnblurToTheBeat {
			blurVal -= settings.Playfield.UnblurFill * (blurVal) * (pl.Scl - 1.0) / (settings.Beat.BeatScale * 0.4)
		}
	}

	//if settings.Playfield.FlashToTheBeat {
	//	bgAlpha *= pl.Scl
	//}

	pl.batch.Begin()

	pl.batch.SetColor(1, 1, 1, 1)
	pl.batch.ResetTransform()
	pl.batch.SetAdditive(false)
	if pl.Background != nil || pl.storyboard != nil {
		if settings.Playfield.BlurEnable {
			pl.blurEffect.SetBlur(blurVal, blurVal)
			pl.blurEffect.Begin()
		}

		if pl.Background != nil && (pl.storyboard == nil || !pl.storyboard.BGFileUsed()) {
			pl.batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			pl.batch.SetScale(pl.BgScl.X, -pl.BgScl.Y)
			if !settings.Playfield.BlurEnable {
				pl.batch.SetColor(1, 1, 1, bgAlpha)
			}
			pl.batch.DrawUnit(*pl.Background)
		}

		if pl.storyboard != nil {
			pl.batch.SetScale(1, 1)
			if !settings.Playfield.BlurEnable {
				pl.batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
			}
			pl.batch.SetCamera(cameras[0])
			pl.storyboard.Draw(pl.progressMs, pl.batch)
			pl.batch.Flush()
		}

		if settings.Playfield.BlurEnable {
			pl.batch.End()

			textureBlur := pl.blurEffect.EndAndProcess()
			pl.batch.Begin()
			pl.batch.SetColor(1, 1, 1, bgAlpha)
			pl.batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			pl.batch.DrawUnscaled(textureBlur.GetRegion())
		}

	}

	pl.batch.Flush()

	//if pl.fxGlider.GetValue() > 0.0 {
	//	pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())
	//	pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
	//	scl := (settings.Graphics.GetWidthF() / float64(pl.Logo.Width)) / 4
	//	pl.batch.SetScale(scl, scl)
	//	pl.batch.DrawTexture(*pl.Logo)
	//	pl.batch.SetScale(scl*(1/pl.Scl), scl*(1/pl.Scl))
	//	pl.batch.SetColor(1, 1, 1, 0.25*pl.fxGlider.GetValue())
	//	pl.batch.DrawTexture(*pl.Logo)
	//}
	//
	pl.batch.End()

	pl.counter += timMs

	if pl.counter >= 1000.0/60 {
		pl.fpsC = pl.profiler.GetFPS()
		pl.fpsU = pl.profilerU.GetFPS()
		pl.counter -= 1000.0 / 60
		//if pl.storyboard != nil {
		//	pl.storyboardLoad = pl.storyboard.GetLoad()
		//}
	}

	//if pl.fxGlider.GetValue() > 0.0 {
	//
	//	pl.fxBatch.Begin()
	//	pl.batch.SetCamera(mgl32.Ortho(-1, 1, 1, -1, 1, -1))
	//	pl.fxBatch.SetColor(1, 1, 1, 0.25*pl.Scl*pl.fxGlider.GetValue())
	//	pl.vao.Begin()
	//
	//	if pl.vaoDirty {
	//		pl.vao.SetVertexData(pl.vaoD)
	//		pl.vaoDirty = false
	//	}
	//
	//	base := mgl32.Ortho(-1920/2, 1920/2, 1080/2, -1080/2, -1, 1).Mul4(mgl32.Scale3D(600, 600, 0)).Mul4(mgl32.HomogRotate3DZ(float32(pl.fxRotation * math.Pi / 180.0)))
	//
	//	pl.fxBatch.SetTransform(base)
	//	pl.vao.Draw()
	//
	//	pl.fxBatch.SetTransform(base.Mul4(mgl32.HomogRotate3DZ(math.Pi)))
	//	pl.vao.Draw()
	//
	//	pl.vao.End()
	//	pl.fxBatch.End()
	//}

	if pl.start {
		settings.Objects.Colors.Update(timMs)
		settings.Objects.CustomSliderBorderColor.Update(timMs)
		settings.Cursor.Colors.Update(timMs)
		if settings.Playfield.RotationEnabled {
			pl.rotation += settings.Playfield.RotationSpeed / 1000.0 * timMs
			for pl.rotation > 360.0 {
				pl.rotation -= 360.0
			}

			for pl.rotation < 0.0 {
				pl.rotation += 360.0
			}
		}
	}

	colors1 := settings.Cursor.GetColors(pl.playerCount+1, settings.TAG, pl.Scl, pl.cursorGlider.GetValue())

	// 指定颜色修改
	for k, colors := range pl.specificColorMap {
		colornum := (settings.VSplayer.PlayerFieldUI.CursorColorSkipNum * k * len(pl.controller[k].GetCursors())) % pl.playerCount
		colors1[colornum][0] = colors[0]
		colors1[colornum][1] = colors[1]
		colors1[colornum][2] = colors[2]
	}

	scale1 := pl.Scl
	scale2 := pl.Scl
	rotationRad := (pl.rotation + settings.Playfield.BaseRotation) * math.Pi / 180.0

	pl.camera.SetRotation(-rotationRad)
	pl.camera.Update()

	if !settings.Objects.ScaleToTheBeat {
		scale1 = 1
	}

	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		pl.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		pl.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.BloomBeatAddition*(pl.Scl-1.0)/(settings.Beat.BeatScale*0.4))
		pl.bloomEffect.Begin()
	}

	//endregion

	//region 渲染录制信息

	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	pl.batch.SetColor(1, 1, 1, settings.VSplayer.RecordInfoUI.RecordAlpha)
	pl.font.Draw(pl.batch, pl.recordbaseX, pl.recordbaseY, pl.recordbasesize, "Recorded by "+settings.VSplayer.RecordInfoUI.Recorder)
	pl.font.Draw(pl.batch, pl.recordbaseX, pl.recordbaseY-pl.recordtimeoffsetY, pl.recordbasesize, "Recorded on "+settings.VSplayer.RecordInfoUI.RecordTime)
	pl.batch.End()

	//endregion

	//region 渲染按键
	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	for k := 0; k < pl.playerCount; k++ {
		linecount := k
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			linecount *= 2
		}
		gapY := pl.gapsize * float64(k)
		colornum := (settings.VSplayer.PlayerFieldUI.CursorColorSkipNum * k * len(pl.controller[k].GetCursors())) % pl.playerCount
		namecolor := colors1[colornum]
		if settings.VSplayer.Knockout.EnableKnockout && (!pl.controller[k].GetIsShow()) {
			namecolor[3] = float32(math.Max(0.0, float64(namecolor[3])-(pl.progressMsF-pl.controller[k].GetDishowTime())/settings.VSplayer.Knockout.PlayerFadeTime))
		}
		playerkey := pl.controller[k].GetPresskey()
		// 通用渲染项
		keyY := pl.keybaseY - pl.lineoffset*float64(linecount) - gapY
		// 如果显示UR，排成两行，按键下移半行
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			keyY -= pl.lineoffset / 2
		}
		// 通用大小
		pl.batch.SetScale(pl.keysize, pl.keysize)
		// Key1
		pl.batch.SetTranslation(bmath.NewVec2d(pl.key1baseX, keyY))
		if playerkey.Key1 {
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
		} else {
			pl.batch.SetColor(1, 1, 1, 0)
		}
		pl.batch.DrawUnit(*render.PressKey)
		// Key2
		pl.batch.SetTranslation(bmath.NewVec2d(pl.key2baseX, keyY))
		if playerkey.Key2 {
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
		} else {
			pl.batch.SetColor(1, 1, 1, 0)
		}
		pl.batch.DrawUnit(*render.PressKey)
		// Mouse1
		if settings.VSplayer.PlayerInfoUI.ShowMouse1 {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key3baseX, keyY))
			if playerkey.LeftClick && !playerkey.Key1 {
				pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			} else {
				pl.batch.SetColor(1, 1, 1, 0)
			}
			pl.batch.DrawUnit(*render.PressKey)
		}
		// Mouse2
		if settings.VSplayer.PlayerInfoUI.ShowMouse2 {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key4baseX, keyY))
			if playerkey.RightClick && !playerkey.Key2 {
				pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			} else {
				pl.batch.SetColor(1, 1, 1, 0)
			}
			pl.batch.DrawUnit(*render.PressKey)
		}
	}
	pl.batch.End()

	//endregion

	//region 渲染文字

	// 文字的公用X轴
	var lastPos []float64
	lastPos = make([]float64, pl.playerCount)
	// 渲染player名
	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	for k := 0; k < pl.playerCount; k++ {
		linecount := k
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			linecount *= 2
		}
		gapY := pl.gapsize * float64(k)
		pl.batch.SetAdditive(true)
		colornum := (settings.VSplayer.PlayerFieldUI.CursorColorSkipNum * k * len(pl.controller[k].GetCursors())) % pl.playerCount
		namecolor := colors1[colornum]
		if settings.VSplayer.Knockout.EnableKnockout && (!pl.controller[k].GetIsShow()) {
			namecolor[3] = float32(math.Max(0.0, float64(namecolor[3])-(pl.progressMsF-pl.controller[k].GetDishowTime())/settings.VSplayer.Knockout.PlayerFadeTime))
		}
		// 渲染player名
		pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
		fontY := pl.fontbaseY - pl.lineoffset*float64(linecount)
		// 下移半行
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			fontY -= pl.lineoffset / 2
		}
		lastPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, pl.playerbaseX, fontY-gapY, pl.fontsize, pl.controller[k].GetPlayname())
		// 渲染mod
		mods := "+"
		if pl.controller[k].GetMods()&MOD_NF > 0 {
			mods += "NF"
		}
		if pl.controller[k].GetMods()&MOD_EZ > 0 {
			mods += "EZ"
		}
		if pl.controller[k].GetMods()&MOD_TD > 0 {
			mods += "TD"
		}
		if pl.controller[k].GetMods()&MOD_HD > 0 {
			mods += "HD"
		}
		if pl.controller[k].GetMods()&MOD_HR > 0 {
			mods += "HR"
		}
		if pl.controller[k].GetMods()&MOD_PF > 0 {
			mods += "PF"
		} else if pl.controller[k].GetMods()&MOD_SD > 0 {
			mods += "SD"
		}
		if pl.controller[k].GetMods()&MOD_NC > 0 {
			mods += "NC"
		} else if pl.controller[k].GetMods()&MOD_DT > 0 {
			mods += "DT"
		}
		if pl.controller[k].GetMods()&MOD_HT > 0 {
			mods += "HT"
		}
		if pl.controller[k].GetMods()&MOD_FL > 0 {
			mods += "FL"
		}
		if pl.controller[k].GetMods()&MOD_SO > 0 {
			mods += "SO"
		}
		if mods != "+" {
			pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
			lastPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, lastPos[k]+pl.modoffset, fontY-gapY, pl.fontsize, mods)
		}
	}
	pl.batch.End()

	//endregion

	//region 渲染300、100、50、miss、acc、rank、pp、ur

	// 断连文字的公用X轴
	var lastmissPos []float64

	// miss颜色数组
	var misscolors [][]float64
	var misscolornums []int

	if settings.VSplayer.Knockout.EnableKnockout {
		lastmissPos = make([]float64, pl.playerCount)
	} else if settings.VSplayer.Knockout.ShowTrueMiss {
		lastmissPos = make([]float64, pl.playerCount)
		misscolors = make([][]float64, pl.playerCount)
		misscolornums = make([]int, pl.playerCount)
	}

	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	for k := 0; k < pl.playerCount; k++ {
		linecount := k
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			linecount *= 2
		}
		gapY := pl.gapsize * float64(k)
		colornum := (settings.VSplayer.PlayerFieldUI.CursorColorSkipNum * k * len(pl.controller[k].GetCursors())) % pl.playerCount
		namecolor := colors1[colornum]
		// 如果设置不显示，开始降低透明度
		if settings.VSplayer.Knockout.EnableKnockout && (!pl.controller[k].GetIsShow()) {
			pl.batch.SetCamera(breakcamera)
			namecolor[3] = float32(math.Max(0.0, float64(namecolor[3])-(pl.progressMsF-pl.controller[k].GetDishowTime())/settings.VSplayer.Knockout.PlayerFadeTime))
			// 显示断连者名字
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			lastmissPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, bmath.GetX(pl.controller[k].GetDishowPos()), bmath.GetY(pl.controller[k].GetDishowPos()), pl.missfontsize, pl.controller[k].GetPlayname())
			// 显示miss
			pl.batch.SetTranslation(bmath.NewVec2d(lastmissPos[k]+pl.missoffsetX, bmath.GetY(pl.controller[k].GetDishowPos())+pl.missoffsetY))
			pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
			pl.batch.SetScale(2.75*pl.misssize, pl.misssize)
			pl.batch.DrawUnit(*render.Hit0)
			pl.batch.SetCamera(pl.scamera.GetProjectionView())
		} else if settings.VSplayer.Knockout.ShowTrueMiss {
			// 如果有新的miss，补充miss颜色数组
			if misscolornums[k] < len(pl.controller[k].GetMissInfo()) {
				for add := 0; add < len(pl.controller[k].GetMissInfo())-misscolornums[k]; add++ {
					misscolors[k] = append(misscolors[k], 1.0)
				}
				misscolornums[k] = len(misscolors[k])
			}
			pl.batch.SetCamera(breakcamera)
			// 最后一个不合法（超时）的missinfo的下标
			var lastIllegalIndex = 0
			// 遍历渲染所有miss
			for m := 0; m < len(pl.controller[k].GetMissInfo()); m++ {
				// 跳过无需渲染的missinfo并记录最后一个不合法的missinfo的下标
				if pl.progressMsF-pl.controller[k].GetMissInfo()[m].MissTime > settings.VSplayer.Knockout.PlayerFadeTime {
					lastIllegalIndex = m
					continue
				}
				// 计算新的Alpha通道值
				misscolors[k][m] = math.Max(0.0, misscolors[k][m]-(pl.progressMsF-pl.controller[k].GetMissInfo()[m].MissTime)/settings.VSplayer.Knockout.PlayerFadeTime)
				// 显示断连者名字
				pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), misscolors[k][m])
				lastmissPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, bmath.GetX(pl.controller[k].GetMissInfo()[m].MissPos), bmath.GetY(pl.controller[k].GetMissInfo()[m].MissPos), pl.missfontsize, pl.controller[k].GetPlayname())
				// 显示miss
				pl.batch.SetTranslation(bmath.NewVec2d(lastmissPos[k]+pl.missoffsetX, bmath.GetY(pl.controller[k].GetMissInfo()[m].MissPos)+pl.missoffsetY))
				pl.batch.SetColor(1, 1, 1, misscolors[k][m])
				pl.batch.SetScale(2.75*pl.misssize, pl.misssize)
				pl.batch.DrawUnit(*render.Hit0)
			}
			pl.batch.SetCamera(pl.scamera.GetProjectionView())
			// 删除不合法的missinfo
			pl.controller[k].SetMissInfo(pl.controller[k].GetMissInfo()[lastIllegalIndex:])
			misscolors[k] = misscolors[k][lastIllegalIndex:]
			misscolornums[k] = len(misscolors[k])
		}
		if !settings.VSplayer.PlayerInfoUI.ShowRealTimePP {
			if len(pl.controller[k].GetHitResult()) > 0 {
				pl.controller[k].SetPP(pl.controller[k].GetTotalResult()[0].PP.Total)
			} else {
				pl.controller[k].SetPP(pl.lastPP[k])
			}
		} else {
			// 显示每帧实时pp变化
			if len(pl.controller[k].GetHitResult()) > 0 {
				pl.controller[k].SetPP(score.CalculateRealtimeValue(
					pl.lastb4PP[k],
					pl.lastPP[k],
					pl.lastPPTime[k],
					pl.controller[k].GetHitResult()[0].JudgeTime,
					pl.progressMsF))
			} else {
				pl.controller[k].SetPP(score.CalculateRealtimeValue(
					pl.lastb4PP[k],
					pl.lastPP[k],
					pl.lastPPTime[k],
					pl.lastPPTime[k]+int64(settings.VSplayer.PlayerInfoUI.RealTimePPGap),
					pl.progressMsF))
			}
		}
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			// 显示每帧实时ur变化
			if len(pl.controller[k].GetHitResult()) > 0 {
				pl.controller[k].SetUR(score.CalculateRealtimeValue(
					pl.lastb4UR[k],
					pl.lastUR[k],
					pl.lastURTime[k],
					pl.controller[k].GetHitResult()[0].JudgeTime,
					pl.progressMsF))
			} else {
				pl.controller[k].SetUR(score.CalculateRealtimeValue(
					pl.lastb4UR[k],
					pl.lastUR[k],
					pl.lastURTime[k],
					pl.lastURTime[k]+int64(settings.VSplayer.PlayerInfoUI.RealTimePPGap),
					pl.progressMsF))
			}
		}
		// 如果现在时间大于第一个result的时间，渲染这个result，并在渲染一定时间后弹出
		if len(pl.controller[k].GetHitResult()) != 0 {
			if pl.progressMs > pl.controller[k].GetHitResult()[0].JudgeTime {
				judge := *render.Hit300
				pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
				switch pl.controller[k].GetHitResult()[0].Result {
				case hitjudge.Hit300:
					pl.batch.SetColor(1, 1, 1, 0)
					break
				case hitjudge.Hit100:
					judge = *render.Hit100
					break
				case hitjudge.Hit50:
					judge = *render.Hit50
					break
				case hitjudge.HitMiss:
					judge = *render.Hit0
					break
				}
				if pl.controller[k].GetHitResult()[0].IsBreak {
					// 断连后设置不显示
					if pl.controller[k].GetIsShow() {
						pl.controller[k].SetIsShow(false)
						// 保存消失时间、消失位置
						pl.controller[k].SetDishowTime(pl.progressMsF)
						trueJudgePos := pl.controller[k].GetHitResult()[0].JudgePos
						// 如果是HR且图整体未开HR，上下翻转
						if !settings.VSplayer.Mods.EnableHR && pl.controller[k].GetMods()&MOD_HR > 0 {
							trueJudgePos.Y = PLAYFIELD_HEIGHT - trueJudgePos.Y
						}
						if pl.lastDishowPos == defaultPos {
							pl.lastDishowPos = trueJudgePos
						} else {
							if pl.lastDishowPos == trueJudgePos {
								pl.SameRate += 1
							} else {
								pl.SameRate = 0
							}
							pl.lastDishowPos = trueJudgePos
						}
						pl.controller[k].SetDishowPos(trueJudgePos, pl.SameRate)
					}
				}
				if pl.controller[k].GetHitResult()[0].Result == hitjudge.HitMiss {
					// 检查是否已经录入
					if pl.controller[k].IsInMiss(pl.controller[k].GetHitResult()[0].JudgeTime) {
						// 保存miss时间、miss真实判断时间、miss位置
						trueJudgePos := pl.controller[k].GetHitResult()[0].JudgePos
						// 如果是HR且图整体未开HR，上下翻转
						if !settings.VSplayer.Mods.EnableHR && pl.controller[k].GetMods()&MOD_HR > 0 {
							trueJudgePos.Y = PLAYFIELD_HEIGHT - trueJudgePos.Y
						}
						if pl.lastMissPos == defaultPos {
							pl.lastMissPos = trueJudgePos
						} else {
							if pl.lastMissPos == trueJudgePos {
								pl.SameMissRate += 1
							} else {
								pl.SameMissRate = 0
							}
							pl.lastMissPos = trueJudgePos
						}
						pl.controller[k].AddMissInfo(pl.progressMsF, pl.controller[k].GetHitResult()[0].JudgeTime, trueJudgePos, pl.SameMissRate)
					}
				}
				judgeY := pl.hitbaseY - pl.lineoffset*float64(linecount)
				// 下移半行
				if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
					judgeY -= pl.lineoffset / 2
				}
				pl.batch.SetTranslation(bmath.NewVec2d(lastPos[k]+pl.hitoffset, judgeY-gapY))
				pl.batch.SetScale(2.75*settings.VSplayer.PlayerInfoUI.BaseSize, settings.VSplayer.PlayerInfoUI.BaseSize)
				pl.batch.DrawUnit(judge)
				// 渲染时间结束，弹出
				if pl.progressMs > pl.controller[k].GetHitResult()[0].JudgeTime+settings.VSplayer.PlayerFieldUI.HitFadeTime {
					// 设置acc、rank和pp
					pl.controller[k].SetAcc(pl.controller[k].GetTotalResult()[0].Acc)
					switch pl.controller[k].GetTotalResult()[0].Rank {
					case score.SSH:
						pl.controller[k].SetRank(*render.RankXH)
						break
					case score.SH:
						pl.controller[k].SetRank(*render.RankSH)
						break
					case score.SS:
						pl.controller[k].SetRank(*render.RankX)
						break
					case score.S:
						pl.controller[k].SetRank(*render.RankS)
						break
					case score.A:
						pl.controller[k].SetRank(*render.RankA)
						break
					case score.B:
						pl.controller[k].SetRank(*render.RankB)
						break
					case score.C:
						pl.controller[k].SetRank(*render.RankC)
						break
					case score.D:
						pl.controller[k].SetRank(*render.RankD)
						break
					}
					if settings.VSplayer.PlayerInfoUI.ShowRealTimePP {
						pl.lastPPTime[k] = pl.controller[k].GetHitResult()[0].JudgeTime
					}
					pl.lastb4PP[k] = pl.lastPP[k]
					pl.lastPP[k] = pl.controller[k].GetTotalResult()[0].PP.Total
					if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
						pl.lastb4UR[k] = pl.lastUR[k]
						pl.lastUR[k] = pl.controller[k].GetTotalResult()[0].UR
						pl.lastURTime[k] = pl.controller[k].GetHitResult()[0].JudgeTime
					}
					// 弹出
					pl.controller[k].SetHitResult(pl.controller[k].GetHitResult()[1:])
					pl.controller[k].SetTotalResult(pl.controller[k].GetTotalResult()[1:])
				}
			}
		}
		// 渲染acc
		pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
		pl.font.Draw(pl.batch, pl.accbaseX, pl.fontbaseY-pl.lineoffset*float64(linecount)-gapY, pl.fontsize, fmt.Sprintf("%.2f", pl.controller[k].GetAcc())+"%")
		// 渲染rank
		rankY := pl.rankbaseY - pl.lineoffset*float64(linecount)
		// 下移一行
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			rankY -= pl.lineoffset
		}
		pl.batch.SetTranslation(bmath.NewVec2d(pl.rankbaseX, rankY-gapY))
		pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
		pl.batch.SetScale(settings.VSplayer.PlayerInfoUI.BaseSize, settings.VSplayer.PlayerInfoUI.BaseSize)
		pl.batch.DrawUnitC(pl.controller[k].GetRank())
		// 渲染pp
		pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
		pl.font.Draw(pl.batch, pl.ppurbaseX, pl.fontbaseY-pl.lineoffset*float64(linecount)-gapY, pl.fontsize, fmt.Sprintf("%.2f", pl.controller[k].GetPP())+" pp")
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			// 渲染ur
			pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
			pl.font.Draw(pl.batch, pl.ppurbaseX, pl.fontbaseY-pl.lineoffset*float64(linecount+1)-gapY, pl.fontsize, fmt.Sprintf("%.2f", pl.controller[k].GetUR())+" ur")
		}
	}
	pl.batch.End()

	//endregion

	//region 渲染pp、ur的排名

	//在pp和ur全部更新一遍后再渲染

	//计算排名
	if settings.VSplayer.PlayerInfoUI.ShowPPAndURRank {
		var pps []float64
		var urs []float64
		var pprank []int
		var urrank []int
		pps = make([]float64, pl.playerCount)
		for k := 0; k < pl.playerCount; k++ {
			pps[k] = pl.controller[k].GetPP()
		}
		pprank = utils.SortRankHighToLow(pps)
		if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
			urs = make([]float64, pl.playerCount)
			for k := 0; k < pl.playerCount; k++ {
				urs[k] = pl.controller[k].GetUR()
			}
			urrank = utils.SortRankLowToHigh(urs)
		}
		pl.batch.Begin()
		pl.batch.SetCamera(pl.scamera.GetProjectionView())
		for k := 0; k < pl.playerCount; k++ {
			linecount := k
			if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
				linecount *= 2
			}
			gapY := pl.gapsize * float64(k)
			colornum := (settings.VSplayer.PlayerFieldUI.CursorColorSkipNum * k * len(pl.controller[k].GetCursors())) % pl.playerCount
			namecolor := colors1[colornum]
			// 渲染pp排名
			pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
			if settings.VSplayer.PlayerInfoUI.Rank1Highlight && (pprank[k] == 1) {
				pl.highlightFont.Draw(pl.batch, pl.ppurrankbaseX, pl.highlightfontbaseY-pl.lineoffset*float64(linecount)-gapY, pl.highlightfontsize, "#"+strconv.Itoa(pprank[k]))
			} else {
				pl.font.Draw(pl.batch, pl.ppurrankbaseX, pl.fontbaseY-pl.lineoffset*float64(linecount)-gapY, pl.fontsize, "#"+strconv.Itoa(pprank[k]))
			}
			if settings.VSplayer.PlayerInfoUI.ShowRealTimeUR {
				// 渲染ur排名
				pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
				if settings.VSplayer.PlayerInfoUI.Rank1Highlight && (urrank[k] == 1) {
					pl.highlightFont.Draw(pl.batch, pl.ppurrankbaseX, pl.highlightfontbaseY-pl.lineoffset*float64(linecount+1)-gapY, pl.highlightfontsize, "#"+strconv.Itoa(urrank[k]))
				} else {
					pl.font.Draw(pl.batch, pl.ppurrankbaseX, pl.fontbaseY-pl.lineoffset*float64(linecount+1)-gapY, pl.fontsize, "#"+strconv.Itoa(urrank[k]))
				}
			}
		}
		pl.batch.End()
	}

	//endregion

	//region 渲染实时难度

	if settings.VSplayer.DiffInfoUI.ShowDiffInfo {
		pl.batch.Begin()
		pl.batch.SetCamera(pl.scamera.GetProjectionView())
		pl.batch.SetColor(1, 1, 1, settings.VSplayer.RecordInfoUI.RecordAlpha)
		if pl.progressMs > pl.bMap.HitObjects[pl.objindex].GetBasicData().JudgeTime {
			if pl.objindex < len(pl.bMap.HitObjects)-1 {
				pl.objindex += 1
			}
		}
		diff := pl.difficulties[pl.objindex].Diff
		var aim float64
		var speed float64
		var total float64
		if pl.objindex == 0 || pl.progressMs > pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().JudgeTime {
			aim = diff.Aim
			speed = diff.Speed
			total = diff.Total
		} else {
			beforediff := pl.difficulties[pl.objindex-1].Diff
			aim = score.CalculateRealtimeValue(
				beforediff.Aim,
				diff.Aim,
				pl.bMap.HitObjects[pl.objindex-1].GetBasicData().JudgeTime,
				pl.bMap.HitObjects[pl.objindex-1].GetBasicData().JudgeTime+int64(settings.VSplayer.PlayerInfoUI.RealTimePPGap),
				pl.progressMsF)
			speed = score.CalculateRealtimeValue(
				beforediff.Speed,
				diff.Speed,
				pl.bMap.HitObjects[pl.objindex-1].GetBasicData().JudgeTime,
				pl.bMap.HitObjects[pl.objindex-1].GetBasicData().JudgeTime+int64(settings.VSplayer.PlayerInfoUI.RealTimePPGap),
				pl.progressMsF)
			total = score.CalculateRealtimeValue(
				beforediff.Total,
				diff.Total,
				pl.bMap.HitObjects[pl.objindex-1].GetBasicData().JudgeTime,
				pl.bMap.HitObjects[pl.objindex-1].GetBasicData().JudgeTime+int64(settings.VSplayer.PlayerInfoUI.RealTimePPGap),
				pl.progressMsF)
		}
		pl.font.Draw(pl.batch, pl.diffbaseX, pl.diffbaseY, pl.diffbasesize, "Aim Stars : "+fmt.Sprintf("%.4f", aim))
		pl.font.Draw(pl.batch, pl.diffbaseX, pl.diffbaseY-pl.diffoffsetY, pl.diffbasesize, "Speed Stars : "+fmt.Sprintf("%.4f", speed))
		pl.font.Draw(pl.batch, pl.diffbaseX, pl.diffbaseY-pl.diffoffsetY*2, pl.diffbasesize, "Total Stars : "+fmt.Sprintf("%.4f", total))
		pl.batch.End()
	}

	//endregion

	//region 无关4

	if pl.start {

		if settings.Objects.SliderMerge {
			pl.sliderRenderer.Begin()

			for j := 0; j < settings.DIVIDES; j++ {
				pl.sliderRenderer.SetCamera(cameras[j])
				ind := j - 1
				if ind < 0 {
					ind = settings.DIVIDES - 1
				}

				for i := len(pl.processed) - 1; i >= 0; i-- {
					if s, ok := pl.processed[i].(*objects.Slider); ok {
						pl.sliderRenderer.SetScale(scale1)
						s.DrawBody(pl.progressMs, pl.bMap.ARms, pl.bMap.FadeIn, colors1[pl.objectcolorIndex], colors1[pl.objectcolorIndex], pl.sliderRenderer)
					}
				}
			}

			pl.sliderRenderer.EndAndRender()
		} else {
			for j := 0; j < settings.DIVIDES; j++ {
				pl.sliderRenderer.SetCamera(cameras[j])
				ind := j - 1
				if ind < 0 {
					ind = settings.DIVIDES - 1
				}

				for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
					if i < len(pl.processed) {
						if !settings.Objects.SliderMerge {
							if s, ok := pl.processed[i].(*objects.Slider); ok {
								pl.batch.Flush()
								pl.sliderRenderer.Begin()
								pl.sliderRenderer.SetScale(scale1)
								s.DrawBody(pl.progressMs, pl.bMap.ARms, pl.bMap.FadeIn, colors1[pl.objectcolorIndex], colors1[pl.objectcolorIndex], pl.sliderRenderer)
								pl.sliderRenderer.EndAndRender()
							}
						}
					}
				}
			}
		}

		pl.batch.Begin()

		if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
			pl.batch.SetAdditive(true)
		} else {
			pl.batch.SetAdditive(false)
		}

		pl.batch.SetScale(64*render.CS*scale1, 64*render.CS*scale1)

		for j := 0; j < settings.DIVIDES; j++ {
			pl.batch.SetCamera(cameras[j])
			ind := j - 1
			if ind < 0 {
				ind = settings.DIVIDES - 1
			}

			for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
				if i < len(pl.processed) {
					res := pl.processed[i].Draw(pl.progressMs, pl.bMap.ARms, pl.bMap.FadeIn, colors1[pl.objectcolorIndex], pl.batch)
					if res {
						pl.processed = append(pl.processed[:i], pl.processed[(i+1):]...)
						i++
					}
				}
			}
		}

		if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger && settings.Objects.DrawApproachCircles {
			pl.batch.Flush()

			for j := 0; j < settings.DIVIDES; j++ {

				pl.batch.SetCamera(cameras[j])

				for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
					if !settings.VSplayer.Mods.EnableHD || (pl.processed[i].GetObjectNumber() == 0) {
						// HD，除了第一个的缩圈全部不渲染
						pl.processed[i].DrawApproach(pl.progressMs, pl.bMap.ARms, pl.bMap.FadeIn, colors1[pl.objectcolorIndex], pl.batch)
					}
				}
			}
		}

		pl.batch.SetScale(1, 1)
		pl.batch.End()
	}

	pl.batch.SetAdditive(false)
	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

	//if settings.DEBUG || settings.FPS {
	//	pl.batch.Begin()
	//	pl.batch.SetColor(1, 1, 1, 1)
	//	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	//
	//	padDown := 4.0
	//	shift := 16.0
	//
	//	if settings.DEBUG {
	//		pl.font.Draw(pl.batch, 0, settings.Graphics.GetHeightF()-24, 24, pl.mapFullName)
	//		pl.font.Draw(pl.batch, 0, padDown+shift*5, 16, fmt.Sprintf("%0.0f FPS", pl.fpsC))
	//		pl.font.Draw(pl.batch, 0, padDown+shift*4, 16, fmt.Sprintf("%0.2f ms", 1000/pl.fpsC))
	//		pl.font.Draw(pl.batch, 0, padDown+shift*3, 16, fmt.Sprintf("%0.2f ms update", 1000/pl.fpsU))
	//
	//		time := int(pl.musicPlayer.GetPosition())
	//		totalTime := int(pl.musicPlayer.GetLength())
	//		mapTime := int(pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)
	//
	//		pl.font.Draw(pl.batch, 0, padDown+shift*2, 16, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", time/60, time%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
	//		pl.font.Draw(pl.batch, 0, padDown+shift, 16, fmt.Sprintf("%d(*%d) hitobjects, %d total", len(pl.processed), settings.DIVIDES, len(pl.bMap.HitObjects)))
	//
	//		if pl.storyboard != nil {
	//			pl.font.Draw(pl.batch, 0, padDown, 16, fmt.Sprintf("%d storyboard sprites (%0.2fx load), %d in queue (%d total)", pl.storyboard.GetProcessedSprites(), pl.storyboardLoad, pl.storyboard.GetQueueSprites(), pl.storyboard.GetTotalSprites()))
	//		} else {
	//			pl.font.Draw(pl.batch, 0, padDown, 16, "No storyboard")
	//		}
	//	} else {
	//		pl.font.Draw(pl.batch, 0, padDown, 16, fmt.Sprintf("%0.0f FPS", pl.fpsC))
	//	}
	//
	//	pl.batch.End()
	//}

	//endregion

	//region 多个光标渲染

	for k := 0; k < pl.playerCount; k++ {
		if !(settings.VSplayer.Knockout.EnableKnockout && (!pl.controller[k].GetIsShow())) {
			for _, g := range pl.controller[k].GetCursors() {
				g.UpdateRenderer()
			}
			gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			pl.batch.SetAdditive(true)
			render.BeginCursorRender()
			for j := 0; j < settings.DIVIDES; j++ {
				pl.batch.SetCamera(cameras[j])
				for i, g := range pl.controller[k].GetCursors() {
					ind := k*len(pl.controller[k].GetCursors()) + i - 1
					if ind < 0 {
						ind = settings.DIVIDES*len(pl.controller[k].GetCursors()) - 1
					}
					colornum := (settings.VSplayer.PlayerFieldUI.CursorColorSkipNum * k * len(pl.controller[k].GetCursors())) % pl.playerCount
					g.DrawM(scale2, pl.batch, colors1[colornum], colors1[ind])
				}
			}
			render.EndCursorRender()
		}
	}

	//endregion

}

func (pl *Player) Stop() {
	pl.exitGoFlag = true
	pl.queue2 = []objects.BaseObject{}
	pl.processed = []objects.Renderable{}
	pl.musicPlayer.Stop()
}
