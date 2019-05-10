package configui

import (
	"danser/settings"
	"github.com/lxn/walk"
	"log"
	"strconv"
)

func (vsw VSPlayerMainWindow) SaveConfig() {
	players, err := strconv.Atoi(vsw.players.Text())
	if err != nil {
		panic(err)
	}
	settings.VSplayer.PlayerInfo.Players = players
	settings.VSplayer.PlayerInfo.SpecifiedPlayers = vsw.specifiedPlayers.Checked()
	settings.VSplayer.PlayerInfo.SpecifiedLine = vsw.specifiedLine.Text()

	basesize, err := strconv.ParseFloat(vsw.baseSize.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.PlayerInfoUI.BaseSize = basesize
	basex, err := strconv.ParseFloat(vsw.baseX.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.PlayerInfoUI.BaseX = basex
	basey, err := strconv.ParseFloat(vsw.baseY.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.PlayerInfoUI.BaseY = basey
	settings.VSplayer.PlayerInfoUI.ShowMouse1 = vsw.showMouse1.Checked()
	settings.VSplayer.PlayerInfoUI.ShowMouse2 = vsw.showMouse2.Checked()
	settings.VSplayer.PlayerInfoUI.ShowRealTimePP = vsw.showRealTimePP.Checked()
	realtimeppgap, err := strconv.ParseFloat(vsw.realTimePPGap.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.PlayerInfoUI.RealTimePPGap = realtimeppgap
	settings.VSplayer.PlayerInfoUI.ShowRealTimeUR = vsw.showRealTimeUR.Checked()
	settings.VSplayer.PlayerInfoUI.ShowPPAndURRank = vsw.showPPAndURRank.Checked()
	settings.VSplayer.PlayerInfoUI.Rank1Highlight = vsw.rank1Highlight.Checked()
	highlightmult, err := strconv.ParseFloat(vsw.highlightMult.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.PlayerInfoUI.HighlightMult = highlightmult

	settings.VSplayer.RecordInfoUI.Recorder = vsw.recorder.Text()
	settings.VSplayer.RecordInfoUI.RecordTime = vsw.recordTime.Text()
	recordbasex, err := strconv.ParseFloat(vsw.recordBaseX.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.RecordInfoUI.RecordBaseX = recordbasex
	recordbasey, err := strconv.ParseFloat(vsw.recordBaseY.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.RecordInfoUI.RecordBaseY = recordbasey
	recordbasesize, err := strconv.ParseFloat(vsw.recordBaseSize.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.RecordInfoUI.RecordBaseSize = recordbasesize
	recordalpha, err := strconv.ParseFloat(vsw.recordAlpha.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.RecordInfoUI.RecordAlpha = recordalpha

	settings.VSplayer.MapInfo.Title = vsw.title.Text()
	settings.VSplayer.MapInfo.Difficulty = vsw.difficulty.Text()

	settings.VSplayer.Mods.EnableDT = vsw.enableDT.Checked()
	settings.VSplayer.Mods.EnableHT = vsw.enableHT.Checked()
	settings.VSplayer.Mods.EnableEZ = vsw.enableEZ.Checked()
	settings.VSplayer.Mods.EnableHR = vsw.enableHR.Checked()
	settings.VSplayer.Mods.EnableHD = vsw.enableHD.Checked()

	settings.VSplayer.Knockout.EnableKnockout = vsw.enableKnockout.Checked()
	settings.VSplayer.Knockout.ShowTrueMiss = vsw.showTrueMiss.Checked()
	playerfadetime, err := strconv.ParseFloat(vsw.playerFadeTime.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.Knockout.PlayerFadeTime = playerfadetime
	sametimeoffset, err := strconv.ParseFloat(vsw.sameTimeOffset.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.Knockout.SameTimeOffset = sametimeoffset
	missmult, err := strconv.ParseFloat(vsw.missMult.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.VSplayer.Knockout.MissMult = missmult

	settings.VSplayer.ReplayandCache.ReplayDir = vsw.replayDir.Text()
	settings.VSplayer.ReplayandCache.CacheDir = vsw.cacheDir.Text()
	settings.VSplayer.ReplayandCache.SaveResultCache = vsw.saveResultCache.Checked()
	settings.VSplayer.ReplayandCache.ReadResultCache = vsw.readResultCache.Checked()
	settings.VSplayer.ReplayandCache.ReplayDebug = vsw.replayDebug.Checked()

	settings.VSplayer.ErrorFix.EnableErrorFix = vsw.enableErrorFix.Checked()
	settings.VSplayer.ErrorFix.ErrorFixFile = vsw.errorFixFile.Text()

	settings.VSplayer.Skin.EnableSkin = vsw.enableSkin.Checked()
	settings.VSplayer.Skin.SkinDir = vsw.skinDir.Text()
	numberoffset, err := strconv.Atoi(vsw.numberOffset.Text())
	if err != nil {
		panic(err)
	}
	settings.VSplayer.Skin.NumberOffset = int64(numberoffset)

	cursorsize, err := strconv.ParseFloat(vsw.cursorSize.Text(), 64)
	if err != nil {
		panic(err)
	}
	settings.Cursor.CursorSize = cursorsize

	settings.Save()
	log.Println("已保存设置")
}

type VSPlayerMainWindow struct {
	*walk.MainWindow

	players *walk.LineEdit
	specifiedPlayers *walk.CheckBox
	specifiedLine *walk.LineEdit

	baseSize *walk.LineEdit
	baseX *walk.LineEdit
	baseY *walk.LineEdit
	showMouse1 *walk.CheckBox
	showMouse2 *walk.CheckBox
	showRealTimePP *walk.CheckBox
	realTimePPGap *walk.LineEdit
	showRealTimeUR *walk.CheckBox
	showPPAndURRank *walk.CheckBox
	rank1Highlight *walk.CheckBox
	highlightMult *walk.LineEdit

	recorder *walk.LineEdit
	recordTime *walk.LineEdit
	recordBaseX *walk.LineEdit
	recordBaseY *walk.LineEdit
	recordBaseSize *walk.LineEdit
	recordAlpha *walk.LineEdit

	hitFadeTime *walk.LineEdit
	cursorColorNum *walk.LineEdit
	cursorColorSkipNum *walk.LineEdit
	showHitCircleNumber *walk.CheckBox

	title *walk.LineEdit
	difficulty *walk.LineEdit

	enableDT *walk.CheckBox
	enableHT *walk.CheckBox
	enableEZ *walk.CheckBox
	enableHR *walk.CheckBox
	enableHD *walk.CheckBox

	enableKnockout *walk.CheckBox
	showTrueMiss *walk.CheckBox
	playerFadeTime *walk.LineEdit
	sameTimeOffset *walk.LineEdit
	missMult *walk.LineEdit

	replayDir *walk.LineEdit
	cacheDir *walk.LineEdit
	saveResultCache *walk.CheckBox
	readResultCache *walk.CheckBox
	replayDebug *walk.CheckBox

	enableErrorFix *walk.CheckBox
	errorFixFile *walk.LineEdit

	enableSkin *walk.CheckBox
	skinDir *walk.LineEdit
	numberOffset *walk.LineEdit

	cursorSize *walk.LineEdit
}
