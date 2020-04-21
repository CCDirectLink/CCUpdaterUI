package main

import (
	"github.com/CCDirectLink/CCUpdaterUI/design"
	"github.com/CCDirectLink/CCUpdaterUI/frenyard/framework"
	"github.com/CCDirectLink/CCUpdaterUI/middle"
	"github.com/CCDirectLink/ccmu/pkg"
	"github.com/Masterminds/semver"

	"sort"
)

// ShowPrimaryView shows the "Primary View" (the mod list right now)
func (app *upApplication) ShowPrimaryView() {

	// This is used to preserve the state when nothing has changed (for example, when browsing).
	var thePage framework.UILayoutElement

	slots := []framework.FlexboxSlot{}

	warnings := middle.FindWarnings(app.gameInstance)
	if app.config.DevMode {
		warnings = append(warnings, middle.Warning{
			Text:   "You are in developer mode! Go to the Build Information (top-right button, 'Credits', 'Build Information') to deactivate it.",
			Action: middle.NullActionWarningID,
		})
	}
	for _, v := range warnings {
		fixAction := framework.ButtonBehavior(nil)
		if v.Action == middle.InstallOrUpdatePackageWarningID {
			pkgID := v.Parameter
			p, _ := app.gameInstance.Get(pkgID)
			fixAction = func() {
				app.GSRightwards()
				app.ShowPackageView(func() {
					app.GSLeftwards()
					app.ShowPrimaryView()
				}, func() {
					app.GSLeftwards()
					app.Teleport(thePage)
				}, p)
			}
		}
		slots = append(slots, framework.FlexboxSlot{
			Element: design.InformationPanel(design.InformationPanelDetails{
				Text:       v.Text,
				ActionText: "FIX",
				Action:     fixAction,
			}),
		})
	}

	// Ok, let's get all the packages in a nice row
	localPackages, _ := app.gameInstance.Installed()
	remotePackages := middle.GetRemotePackages()
	packageSet := append([]pkg.Package{}, localPackages...)
	packageList := []design.ListItemDetails{}
	for _, p := range remotePackages {
		info, _ := p.Info()
		duplicate := false
		for _, l := range localPackages {
			lInfo, _ := l.Info()
			if info.Name == lInfo.Name {
				duplicate = true
				break
			}
		}

		if !duplicate {
			packageSet = append(packageSet, p)
		}
	}
	// Actually build the UI now!
	for _, p := range packageSet {
		info, err := p.Info()

		if !app.config.DevMode && err != nil {
			continue
		}

		status := "unable to comprehend status"
		if p.Installed() && p.Available() {
			lmv, lerr := semver.NewVersion(info.CurrentVersion)
			rmv, rerr := semver.NewVersion(info.NewestVersion)

			if lerr != nil || rerr != nil {
				status = "Error retrieving version info"
			} else if lmv.GreaterThan(rmv) {
				status = info.CurrentVersion + " installed (local development build, " + info.NewestVersion + " remote)"
			} else if rmv.GreaterThan(lmv) {
				status = info.CurrentVersion + " installed (out of date, " + info.NewestVersion + " available)"
			} else {
				status = info.CurrentVersion + " (up to date)"
			}
		} else if p.Installed() {
			status = info.CurrentVersion + " installed (no remote copy)"
		} else if p.Available() {
			status = info.NewestVersion + " available"
		}

		pkg := p
		packageList = append(packageList, design.ListItemDetails{
			Icon:    middle.PackageIcon(p),
			Text:    info.NiceName,
			Subtext: status,
			Click: func() {
				app.GSRightwards()
				app.ShowPackageView(func() {
					app.GSLeftwards()
					app.ShowPrimaryView()
				}, func() {
					app.GSLeftwards()
					app.Teleport(thePage)
				}, pkg)
			},
		})
	}

	sort.Sort(design.SortListItemDetails(packageList))
	slots = append(slots, framework.FlexboxSlot{
		Element: design.NewUISearchBoxPtr("Search...", packageList),
		Grow:    1,
	})

	// Keep copies of whatever the options menu can change.
	// If we're returned to with something changed, refresh.
	// Otherwise try to reuse the element; it's better-performant and preserves state.
	thePresentStateOfDevMode := app.config.DevMode

	thePage = design.LayoutDocument(design.Header{
		Title: "Mods",
		Back: func() {
			app.GSLeftwards()
			app.ResetWithGameLocation(false, middle.GameFinderVFSPathDefault)
		},
		BackIcon:    design.GameIconID,
		ForwardIcon: design.MenuIconID,
		Forward: func() {
			app.GSRightwards()
			app.ShowOptionsMenu(func() {
				app.GSLeftwards()
				if thePresentStateOfDevMode != app.config.DevMode {
					app.ShowPrimaryView()
				} else {
					app.Teleport(thePage)
				}
			})
		},
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true)
	app.Teleport(thePage)
}
