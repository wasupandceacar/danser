package render

import (
	"danser/ini"
	"danser/render/texture"
	"danser/settings"
	"danser/utils"
	"log"
	"strconv"
)

var Atlas *texture.TextureAtlas

var Circle *texture.TextureRegion
var ApproachCircle *texture.TextureRegion
var SpinnerCircle *texture.TextureRegion
var SpinnerClear *texture.TextureRegion
var SpinnerBackground *texture.TextureRegion
var SpinnerTop *texture.TextureRegion
var SpinnerMiddle *texture.TextureRegion
var SpinnerBottom *texture.TextureRegion
var SpinnerApproachCircle *texture.TextureRegion
var CircleFull *texture.TextureRegion
var CircleOverlay *texture.TextureRegion
var SliderReverse *texture.TextureRegion
var SliderGradient *texture.TextureSingle
var SliderTick *texture.TextureRegion
var SliderBall *texture.TextureRegion
var CursorTex *texture.TextureRegion
var CursorTop *texture.TextureRegion
var CursorTrail *texture.TextureSingle
var PressKey *texture.TextureRegion

var Hit300 *texture.TextureRegion
var Hit100 *texture.TextureRegion
var Hit50 *texture.TextureRegion
var Hit0 *texture.TextureRegion

var Circle0 *texture.TextureRegion
var Circle1 *texture.TextureRegion
var Circle2 *texture.TextureRegion
var Circle3 *texture.TextureRegion
var Circle4 *texture.TextureRegion
var Circle5 *texture.TextureRegion
var Circle6 *texture.TextureRegion
var Circle7 *texture.TextureRegion
var Circle8 *texture.TextureRegion
var Circle9 *texture.TextureRegion

var RankXH *texture.TextureRegion
var RankSH *texture.TextureRegion
var RankX *texture.TextureRegion
var RankS *texture.TextureRegion
var RankA *texture.TextureRegion
var RankB *texture.TextureRegion
var RankC *texture.TextureRegion
var RankD *texture.TextureRegion

// 贴图 2x 倍数
var ApproachCircle2x int32
var SliderReverse2x int32
var DefaultNumber2x int32
var SpinnerCircle2x int32
var SpinnerTop2x int32
var SpinnerMiddle2x int32
var SpinnerBottom2x int32

// 皮肤版本
var SkinVersion float64

// 圈内数字的偏移
var HitCircleOverlap int64

func LoadTextures() {
	Atlas = texture.NewTextureAtlas(8192, 4)
	Atlas.Bind(16)
	Circle, _, _ = loadTextureToAtlas(Atlas, "hitcircle")
	ApproachCircle, _, ApproachCircle2x = loadTextureToAtlas(Atlas, "approachcircle")
	SpinnerCircle, _, SpinnerCircle2x = loadTextureToAtlas(Atlas, "spinner-circle")
	SpinnerClear, _, _ = loadTextureToAtlas(Atlas, "spinner-clear")
	SpinnerBackground, _, _ = loadTextureToAtlas(Atlas, "spinner-background")
	SpinnerTop, _, SpinnerTop2x = loadTextureToAtlas(Atlas, "spinner-top")
	SpinnerMiddle, _, SpinnerMiddle2x = loadTextureToAtlas(Atlas, "spinner-middle")
	SpinnerBottom, _, SpinnerBottom2x = loadTextureToAtlas(Atlas, "spinner-bottom")
	SpinnerApproachCircle, _, _ = loadTextureToAtlas(Atlas, "spinner-approachcircle")
	CircleFull, _, _ = loadTextureToAtlas(Atlas, "hitcircle-full")
	CircleOverlay, _, _ = loadTextureToAtlas(Atlas, "hitcircleoverlay")
	SliderReverse, _, SliderReverse2x = loadTextureToAtlas(Atlas, "reversearrow")
	SliderTick, _, _ = loadTextureToAtlas(Atlas, "sliderscorepoint")
	SliderBall, _, _ = loadTextureToAtlas(Atlas, "sliderball")
	CursorTex, _, _ = loadTextureToAtlas(Atlas, "cursor")
	CursorTop, _, _ = loadTextureToAtlas(Atlas, "cursor-top")
	SliderGradient, _ = loadTexture("slidergradient")
	CursorTrail, _ = loadTexture("cursortrail")
	PressKey, _, _ = loadTextureToAtlas(Atlas, "presskey")

	Hit300, _, _ = loadTextureToAtlas(Atlas, "hit-300")
	Hit100, _, _ = loadTextureToAtlas(Atlas, "hit-100")
	Hit50, _, _ = loadTextureToAtlas(Atlas, "hit-50")
	Hit0, _, _ = loadTextureToAtlas(Atlas, "hit-0")

	Circle0, _, DefaultNumber2x = loadTextureToAtlas(Atlas, "default-0")
	Circle1, _, _ = loadTextureToAtlas(Atlas, "default-1")
	Circle2, _, _ = loadTextureToAtlas(Atlas, "default-2")
	Circle3, _, _ = loadTextureToAtlas(Atlas, "default-3")
	Circle4, _, _ = loadTextureToAtlas(Atlas, "default-4")
	Circle5, _, _ = loadTextureToAtlas(Atlas, "default-5")
	Circle6, _, _ = loadTextureToAtlas(Atlas, "default-6")
	Circle7, _, _ = loadTextureToAtlas(Atlas, "default-7")
	Circle8, _, _ = loadTextureToAtlas(Atlas, "default-8")
	Circle9, _, _ = loadTextureToAtlas(Atlas, "default-9")

	RankXH, _, _ = loadTextureToAtlas(Atlas, "ranking-XH-small")
	RankSH, _, _ = loadTextureToAtlas(Atlas, "ranking-SH-small")
	RankX, _, _ = loadTextureToAtlas(Atlas, "ranking-X-small")
	RankS, _, _ = loadTextureToAtlas(Atlas, "ranking-S-small")
	RankA, _, _ = loadTextureToAtlas(Atlas, "ranking-A-small")
	RankB, _, _ = loadTextureToAtlas(Atlas, "ranking-B-small")
	RankC, _, _ = loadTextureToAtlas(Atlas, "ranking-C-small")
	RankD, _, _ = loadTextureToAtlas(Atlas, "ranking-D-small")
}

func loadTextureToAtlas(atlas *texture.TextureAtlas, picname string) (*texture.TextureRegion, error, int32) {
	var path string
	is2x := 1
	if settings.VSplayer.Skin.EnableSkin {
		// 使用自定义皮肤，则检查皮肤文件夹是否存在相关贴图
		dirExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir)
		if dirExist {
			// 检查是否存在 2x 贴图
			pic2Exist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir + picname + "@2x.png")
			if pic2Exist {
				// 2x 贴图存在, 替换，设置 2x flag
				path = settings.VSplayer.Skin.SkinDir + picname + "@2x.png"
				is2x = 2
			} else {
				picExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir + picname + ".png")
				if picExist {
					// 贴图存在，替换
					path = settings.VSplayer.Skin.SkinDir + picname + ".png"
				} else {
					// 不存在，使用默认
					path = "assets/textures/" + picname + ".png"
				}
			}
		} else {
			// 皮肤文件夹不存在
			panic("皮肤文件夹不存在！")
		}
	} else {
		path = "assets/textures/" + picname + ".png"
	}
	loadTexture, loadError := utils.LoadTextureToAtlas(atlas, path)
	return loadTexture, loadError, int32(is2x)
}

func loadTexture(picname string) (*texture.TextureSingle, error) {
	var path string
	if settings.VSplayer.Skin.EnableSkin {
		// 使用自定义皮肤，则检查皮肤文件夹是否存在相关贴图
		dirExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir)
		if dirExist {
			picExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir + picname + ".png")
			if picExist {
				// 贴图存在，替换
				path = settings.VSplayer.Skin.SkinDir + picname + ".png"
			} else {
				// 不存在，使用默认
				path = "assets/textures/" + picname + ".png"
			}
		} else {
			// 皮肤文件夹不存在
			panic("皮肤文件夹不存在！")
		}
	} else {
		path = "assets/textures/" + picname + ".png"
	}
	return utils.LoadTexture(path)
}

// 读取皮肤配置
func LoadSkinConfiguration() {
	var path string
	if settings.VSplayer.Skin.EnableSkin {
		// 使用自定义皮肤，则检查皮肤文件夹是否存在相关贴图
		dirExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir)
		if dirExist {
			picExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir + "skin.ini")
			if picExist {
				// 皮肤配置文件存在
				path = settings.VSplayer.Skin.SkinDir + "skin.ini"
			} else {
				// 皮肤配置文件不存在，使用默认
				panic("皮肤配置不存在！")
			}
		} else {
			// 皮肤文件夹不存在
			panic("皮肤文件夹不存在！")
		}
	} else {
		path = "assets/textures/skin.ini"
	}
	skinConfig, err := ini.NewFileConf(path)
	if err != nil {
		panic(err)
	}
	SkinVersionstring := skinConfig.String("General.Version")
	if SkinVersionstring == "latest" || SkinVersionstring == "User" {
		SkinVersion = 2.5
	} else {
		SkinVersion, err = strconv.ParseFloat(SkinVersionstring, 64)
		if err != nil {
			log.Println("未找到皮肤版本配置，将使用最新版本。")
			SkinVersion = 2.5
		}
	}
	log.Println("皮肤版本：", SkinVersion)
	HitCircleOverlap, err = skinConfig.Int64("Fonts.HitCircleOverlap")
	if err != nil {
		panic(err)
	}
	log.Println("圈内数字偏移：", HitCircleOverlap, "设置矫正偏移：", settings.VSplayer.Skin.NumberOffset)
	HitCircleOverlap += settings.VSplayer.Skin.NumberOffset
	HitCircleOverlap *= int64(DefaultNumber2x)
}
