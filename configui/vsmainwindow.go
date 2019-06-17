package configui

import (
	"danser/settings"
	"github.com/lxn/walk"
	"log"
)

func (vsw VSPlayerMainWindow) SaveConfig() {
	assign(&settings.VSplayer.PlayerInfo.Players, settings.VSplayer.PlayerInfo.Players, vsw.players)
	assign(&settings.VSplayer.PlayerInfo.SpecifiedPlayers, settings.VSplayer.PlayerInfo.SpecifiedPlayers, vsw.specifiedPlayers)
	assign(&settings.VSplayer.PlayerInfo.SpecifiedLine, settings.VSplayer.PlayerInfo.SpecifiedLine, vsw.specifiedLine)

	assign(&settings.VSplayer.PlayerInfoUI.BaseSize, settings.VSplayer.PlayerInfoUI.BaseSize, vsw.baseSize)
	assign(&settings.VSplayer.PlayerInfoUI.BaseX, settings.VSplayer.PlayerInfoUI.BaseX, vsw.baseX)
	assign(&settings.VSplayer.PlayerInfoUI.BaseY, settings.VSplayer.PlayerInfoUI.BaseY, vsw.baseY)
	assign(&settings.VSplayer.PlayerInfoUI.ShowMouse1, settings.VSplayer.PlayerInfoUI.ShowMouse1, vsw.showMouse1)
	assign(&settings.VSplayer.PlayerInfoUI.ShowMouse2, settings.VSplayer.PlayerInfoUI.ShowMouse2, vsw.showMouse2)
	assign(&settings.VSplayer.PlayerInfoUI.ShowRealTimePP, settings.VSplayer.PlayerInfoUI.ShowRealTimePP, vsw.showRealTimePP)
	assign(&settings.VSplayer.PlayerInfoUI.RealTimePPGap, settings.VSplayer.PlayerInfoUI.RealTimePPGap, vsw.realTimePPGap)
	assign(&settings.VSplayer.PlayerInfoUI.ShowRealTimeUR, settings.VSplayer.PlayerInfoUI.ShowRealTimeUR, vsw.showRealTimeUR)
	assign(&settings.VSplayer.PlayerInfoUI.ShowPPAndURRank, settings.VSplayer.PlayerInfoUI.ShowPPAndURRank, vsw.showPPAndURRank)
	assign(&settings.VSplayer.PlayerInfoUI.Rank1Highlight, settings.VSplayer.PlayerInfoUI.Rank1Highlight, vsw.rank1Highlight)
	assign(&settings.VSplayer.PlayerInfoUI.HighlightMult, settings.VSplayer.PlayerInfoUI.HighlightMult, vsw.highlightMult)
	assign(&settings.VSplayer.PlayerInfoUI.LineGapMult, settings.VSplayer.PlayerInfoUI.LineGapMult, vsw.lineGapMult)

	assign(&settings.VSplayer.RecordInfoUI.Recorder, settings.VSplayer.RecordInfoUI.Recorder, vsw.recorder)
	assign(&settings.VSplayer.RecordInfoUI.RecordTime, settings.VSplayer.RecordInfoUI.RecordTime, vsw.recordTime)
	assign(&settings.VSplayer.RecordInfoUI.RecordBaseX, settings.VSplayer.RecordInfoUI.RecordBaseX, vsw.recordBaseX)
	assign(&settings.VSplayer.RecordInfoUI.RecordBaseY, settings.VSplayer.RecordInfoUI.RecordBaseY, vsw.recordBaseY)
	assign(&settings.VSplayer.RecordInfoUI.RecordBaseSize, settings.VSplayer.RecordInfoUI.RecordBaseSize, vsw.recordBaseSize)
	assign(&settings.VSplayer.RecordInfoUI.RecordAlpha, settings.VSplayer.RecordInfoUI.RecordAlpha, vsw.recordAlpha)

	assign(&settings.VSplayer.DiffInfoUI.ShowDiffInfo, settings.VSplayer.DiffInfoUI.ShowDiffInfo, vsw.showDiffInfo)
	assign(&settings.VSplayer.DiffInfoUI.DiffBaseX, settings.VSplayer.DiffInfoUI.DiffBaseX, vsw.diffBaseX)
	assign(&settings.VSplayer.DiffInfoUI.DiffBaseY, settings.VSplayer.DiffInfoUI.DiffBaseY, vsw.diffBaseY)
	assign(&settings.VSplayer.DiffInfoUI.DiffBaseSize, settings.VSplayer.DiffInfoUI.DiffBaseSize, vsw.diffBaseSize)
	assign(&settings.VSplayer.DiffInfoUI.DiffAlpha, settings.VSplayer.DiffInfoUI.DiffAlpha, vsw.diffAlpha)

	assign(&settings.VSplayer.PlayerFieldUI.HitFadeTime, settings.VSplayer.PlayerFieldUI.HitFadeTime, vsw.hitFadeTime)
	assign(&settings.VSplayer.PlayerFieldUI.CursorColorNum, settings.VSplayer.PlayerFieldUI.CursorColorNum, vsw.cursorColorNum)
	assign(&settings.VSplayer.PlayerFieldUI.CursorColorSkipNum, settings.VSplayer.PlayerFieldUI.CursorColorSkipNum, vsw.cursorColorSkipNum)
	assign(&settings.VSplayer.PlayerFieldUI.ShowHitCircleNumber, settings.VSplayer.PlayerFieldUI.ShowHitCircleNumber, vsw.showHitCircleNumber)

	assign(&settings.VSplayer.MapInfo.Title, settings.VSplayer.MapInfo.Title, vsw.title)
	assign(&settings.VSplayer.MapInfo.Difficulty, settings.VSplayer.MapInfo.Difficulty, vsw.difficulty)

	assign(&settings.VSplayer.Mods.EnableDT, settings.VSplayer.Mods.EnableDT, vsw.enableDT)
	assign(&settings.VSplayer.Mods.EnableHT, settings.VSplayer.Mods.EnableHT, vsw.enableHT)
	assign(&settings.VSplayer.Mods.EnableEZ, settings.VSplayer.Mods.EnableEZ, vsw.enableEZ)
	assign(&settings.VSplayer.Mods.EnableHR, settings.VSplayer.Mods.EnableHR, vsw.enableHR)
	assign(&settings.VSplayer.Mods.EnableHD, settings.VSplayer.Mods.EnableHD, vsw.enableHD)

	assign(&settings.VSplayer.Knockout.EnableKnockout, settings.VSplayer.Knockout.EnableKnockout, vsw.enableKnockout)
	assign(&settings.VSplayer.Knockout.ShowTrueMiss, settings.VSplayer.Knockout.ShowTrueMiss, vsw.showTrueMiss)
	assign(&settings.VSplayer.Knockout.PlayerFadeTime, settings.VSplayer.Knockout.PlayerFadeTime, vsw.playerFadeTime)
	assign(&settings.VSplayer.Knockout.SameTimeOffset, settings.VSplayer.Knockout.SameTimeOffset, vsw.sameTimeOffset)
	assign(&settings.VSplayer.Knockout.MissMult, settings.VSplayer.Knockout.MissMult, vsw.missMult)

	assign(&settings.VSplayer.ReplayandCache.ReplayDir, settings.VSplayer.ReplayandCache.ReplayDir, vsw.replayDir)
	assign(&settings.VSplayer.ReplayandCache.CacheDir, settings.VSplayer.ReplayandCache.CacheDir, vsw.cacheDir)
	assign(&settings.VSplayer.ReplayandCache.SaveResultCache, settings.VSplayer.ReplayandCache.SaveResultCache, vsw.saveResultCache)
	assign(&settings.VSplayer.ReplayandCache.ReadResultCache, settings.VSplayer.ReplayandCache.ReadResultCache, vsw.readResultCache)
	assign(&settings.VSplayer.ReplayandCache.ReplayDebug, settings.VSplayer.ReplayandCache.ReplayDebug, vsw.replayDebug)

	assign(&settings.VSplayer.ErrorFix.EnableErrorFix, settings.VSplayer.ErrorFix.EnableErrorFix, vsw.enableErrorFix)
	assign(&settings.VSplayer.ErrorFix.ErrorFixFile, settings.VSplayer.ErrorFix.ErrorFixFile, vsw.errorFixFile)

	assign(&settings.VSplayer.Skin.EnableSkin, settings.VSplayer.Skin.EnableSkin, vsw.enableSkin)
	assign(&settings.VSplayer.Skin.SkinDir, settings.VSplayer.Skin.SkinDir, vsw.skinDir)
	assign(&settings.VSplayer.Skin.NumberOffset, settings.VSplayer.Skin.NumberOffset, vsw.numberOffset)

	assign(&settings.Cursor.CursorSize, settings.Cursor.CursorSize, vsw.cursorSize)

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
	lineGapMult *walk.LineEdit

	recorder *walk.LineEdit
	recordTime *walk.LineEdit
	recordBaseX *walk.LineEdit
	recordBaseY *walk.LineEdit
	recordBaseSize *walk.LineEdit
	recordAlpha *walk.LineEdit

	showDiffInfo *walk.CheckBox
	diffBaseX *walk.LineEdit
	diffBaseY *walk.LineEdit
	diffBaseSize *walk.LineEdit
	diffAlpha *walk.LineEdit

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
