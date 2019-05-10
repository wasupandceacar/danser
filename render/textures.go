package render

import (
	"danser/ini"
	"danser/render/texture"
	"danser/settings"
	"danser/utils"
	"log"
)

var Atlas *texture.TextureAtlas

var Circle *texture.TextureRegion
var SpinnerBottom *texture.TextureRegion
var ApproachCircle *texture.TextureRegion
var SpinnerCircle *texture.TextureRegion
var SpinnerMiddle *texture.TextureRegion
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

// 圈内数字的偏移
var HitCircleOverlap int64

func LoadTextures() {
	Atlas = texture.NewTextureAtlas(8192, 4)
	Atlas.Bind(16)
	Circle, _ = loadTextureToAtlas(Atlas, "hitcircle.png")
	SpinnerBottom, _ = loadTextureToAtlas(Atlas, "spinner-bottom.png")
	ApproachCircle, _ = loadTextureToAtlas(Atlas, "approachcircle.png")
	SpinnerCircle, _ = loadTextureToAtlas(Atlas, "spinner-circle.png")
	SpinnerMiddle, _ = loadTextureToAtlas(Atlas, "spinner-middle.png")
	SpinnerApproachCircle, _ = loadTextureToAtlas(Atlas, "spinner-approachcircle.png")
	CircleFull, _ = loadTextureToAtlas(Atlas, "hitcircle-full.png")
	CircleOverlay, _ = loadTextureToAtlas(Atlas, "hitcircleoverlay.png")
	SliderReverse, _ = loadTextureToAtlas(Atlas, "reversearrow.png")
	SliderTick, _ = loadTextureToAtlas(Atlas, "sliderscorepoint.png")
	SliderBall, _ = loadTextureToAtlas(Atlas, "sliderball.png")
	CursorTex, _ = loadTextureToAtlas(Atlas, "cursor.png")
	CursorTop, _ = loadTextureToAtlas(Atlas, "cursor-top.png")
	SliderGradient, _ = loadTexture("slidergradient.png")
	CursorTrail, _ = loadTexture("cursortrail.png")
	PressKey, _ = loadTextureToAtlas(Atlas,"presskey.png")

	Hit300, _ = loadTextureToAtlas(Atlas,"hit-300.png")
	Hit100, _ = loadTextureToAtlas(Atlas,"hit-100.png")
	Hit50, _ = loadTextureToAtlas(Atlas,"hit-50.png")
	Hit0, _ = loadTextureToAtlas(Atlas,"hit-0.png")

	Circle0, _ = loadTextureToAtlas(Atlas,"default-0.png")
	Circle1, _ = loadTextureToAtlas(Atlas,"default-1.png")
	Circle2, _ = loadTextureToAtlas(Atlas,"default-2.png")
	Circle3, _ = loadTextureToAtlas(Atlas,"default-3.png")
	Circle4, _ = loadTextureToAtlas(Atlas,"default-4.png")
	Circle5, _ = loadTextureToAtlas(Atlas,"default-5.png")
	Circle6, _ = loadTextureToAtlas(Atlas,"default-6.png")
	Circle7, _ = loadTextureToAtlas(Atlas,"default-7.png")
	Circle8, _ = loadTextureToAtlas(Atlas,"default-8.png")
	Circle9, _ = loadTextureToAtlas(Atlas,"default-9.png")

	RankXH, _ = loadTextureToAtlas(Atlas,"ranking-XH-small.png")
	RankSH, _ = loadTextureToAtlas(Atlas,"ranking-SH-small.png")
	RankX, _ = loadTextureToAtlas(Atlas,"ranking-X-small.png")
	RankS, _ = loadTextureToAtlas(Atlas,"ranking-S-small.png")
	RankA, _ = loadTextureToAtlas(Atlas,"ranking-A-small.png")
	RankB, _ = loadTextureToAtlas(Atlas,"ranking-B-small.png")
	RankC, _ = loadTextureToAtlas(Atlas,"ranking-C-small.png")
	RankD, _ = loadTextureToAtlas(Atlas,"ranking-D-small.png")
}

func loadTextureToAtlas(atlas *texture.TextureAtlas, picname string) (*texture.TextureRegion, error){
	var path string
	if settings.VSplayer.Skin.EnableSkin {
		// 使用自定义皮肤，则检查皮肤文件夹是否存在相关贴图
		dirExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir)
		if dirExist {
			picExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir + picname)
			if picExist {
				// 贴图存在，替换
				path = settings.VSplayer.Skin.SkinDir + picname
			}else {
				// 不存在，使用默认
				path = "assets/textures/" + picname
			}
		}else {
			// 皮肤文件夹不存在
			panic("皮肤文件夹不存在！")
		}
	}else {
		path = "assets/textures/" + picname
	}
	return utils.LoadTextureToAtlas(atlas, path)
}

func loadTexture(picname string) (*texture.TextureSingle, error){
	var path string
	if settings.VSplayer.Skin.EnableSkin {
		// 使用自定义皮肤，则检查皮肤文件夹是否存在相关贴图
		dirExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir)
		if dirExist {
			picExist, _ := utils.PathExists(settings.VSplayer.Skin.SkinDir + picname)
			if picExist {
				// 贴图存在，替换
				path = settings.VSplayer.Skin.SkinDir + picname
			}else {
				// 不存在，使用默认
				path = "assets/textures/" + picname
			}
		}else {
			// 皮肤文件夹不存在
			panic("皮肤文件夹不存在！")
		}
	}else {
		path = "assets/textures/" + picname
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
			}else {
				// 皮肤配置文件不存在，使用默认
				panic("皮肤配置不存在！")
			}
		}else {
			// 皮肤文件夹不存在
			panic("皮肤文件夹不存在！")
		}
	}else {
		path = "assets/textures/skin.ini"
	}
	skinConfig, err := ini.NewFileConf(path)
	if err != nil {
		panic(err)
	}
	HitCircleOverlap, err = skinConfig.Int64("Fonts.HitCircleOverlap")
	if err != nil {
		panic(err)
	}
	log.Println("圈内数字偏移：", HitCircleOverlap, "设置矫正偏移", settings.VSplayer.Skin.NumberOffset)
	HitCircleOverlap += settings.VSplayer.Skin.NumberOffset
}
