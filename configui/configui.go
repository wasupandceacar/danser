package configui

import (
	. "danser/build"
	"danser/runplayfield"
	"danser/settings"
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"strconv"
)

func UImain() {
	// 首先载入设置
	settings.LoadSettings(runplayfield.SettingsVersion)

	vsw := &VSPlayerMainWindow{}

	log.Println("主窗口启动")

	// 固定大小
	vsw.SetFixedSize(true)

	MainWindow{
		AssignTo: &vsw.MainWindow,
		Title:    "osu vs player " + VERSION + " by " + OWNER,
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
												AssignTo: &vsw.players,
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
												AssignTo: &vsw.specifiedPlayers,
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
												AssignTo: &vsw.specifiedLine,
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
												AssignTo: &vsw.baseSize,
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
												AssignTo: &vsw.baseX,
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
												AssignTo: &vsw.baseY,
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
												AssignTo: &vsw.showMouse1,
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
												AssignTo: &vsw.showMouse2,
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
												AssignTo: &vsw.showRealTimePP,
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
												AssignTo: &vsw.realTimePPGap,
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
												AssignTo: &vsw.showRealTimeUR,
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
												AssignTo: &vsw.showPPAndURRank,
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
												AssignTo: &vsw.rank1Highlight,
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
												AssignTo: &vsw.highlightMult,
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
												AssignTo: &vsw.recorder,
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
												AssignTo: &vsw.recordTime,
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
												AssignTo: &vsw.recordBaseX,
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
												AssignTo: &vsw.recordBaseY,
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
												AssignTo: &vsw.recordBaseSize,
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
												AssignTo: &vsw.recordAlpha,
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
												AssignTo: &vsw.hitFadeTime,
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
												AssignTo: &vsw.cursorColorNum,
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
												AssignTo: &vsw.cursorColorSkipNum,
												Alignment: AlignHNearVNear,
												Text: strconv.Itoa(settings.VSplayer.PlayerFieldUI.CursorColorSkipNum),
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "显示note数字：",
											},
											CheckBox{
												AssignTo: &vsw.showHitCircleNumber,
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
												AssignTo: &vsw.title,
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
												AssignTo: &vsw.difficulty,
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
												AssignTo: &vsw.enableDT,
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
												AssignTo: &vsw.enableHT,
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
												AssignTo: &vsw.enableEZ,
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
												AssignTo: &vsw.enableHR,
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
												AssignTo: &vsw.enableHD,
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
												AssignTo: &vsw.enableKnockout,
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
												AssignTo: &vsw.showTrueMiss,
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
												AssignTo: &vsw.playerFadeTime,
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
												AssignTo: &vsw.sameTimeOffset,
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
												AssignTo: &vsw.missMult,
												Alignment: AlignHNearVNear,
												Text: fmt.Sprintf("%g", settings.VSplayer.Knockout.MissMult),
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
												AssignTo: &vsw.replayDir,
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
												AssignTo: &vsw.cacheDir,
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
												AssignTo: &vsw.saveResultCache,
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
												AssignTo: &vsw.readResultCache,
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
												AssignTo: &vsw.replayDebug,
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
												AssignTo: &vsw.enableErrorFix,
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
												AssignTo: &vsw.errorFixFile,
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
												AssignTo: &vsw.enableSkin,
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
												AssignTo: &vsw.skinDir,
												Alignment: AlignHNearVNear,
												Text: settings.VSplayer.Skin.SkinDir,
											},
										},
									},
									HSplitter{
										Children: []Widget{
											TextLabel{
												Text: "数字间隔偏移：",
											},
											LineEdit{
												AssignTo: &vsw.numberOffset,
												Alignment: AlignHNearVNear,
												Text: strconv.Itoa(int(settings.VSplayer.Skin.NumberOffset)),
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
												AssignTo: &vsw.cursorSize,
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
				Text: "保存设置",
				Alignment: AlignHCenterVFar,
				OnClicked: func() {
					vsw.SaveConfig()
				},
			},
			PushButton{
				Text: "保存并开始",
				Alignment: AlignHCenterVFar,
				OnClicked: func() {
					vsw.SaveConfig()
					mainthread.CallQueueCap = 100000
					mainthread.Run(runplayfield.RunPlayField)
				},
			},
		},
	}.Create()

	icon, err := walk.NewIconFromResourceId(3)
	if err != nil {
		panic(err)
	}
	vsw.MainWindow.SetIcon(icon)

	vsw.MainWindow.Run()
}
