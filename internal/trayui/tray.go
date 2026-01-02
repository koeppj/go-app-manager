//go:build windows
// +build windows

package trayui

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/getlantern/systray"

	"github.com/koeppj/go-app-manager/internal/app/process"
	"github.com/koeppj/go-app-manager/internal/trayui/icons"
)

// Run starts the tray UI with optional shutdown hook.
func Run(serviceURL, token, ca string, onShutdown func()) error {
	client, err := NewAPIClient(serviceURL, token, ca)
	if err != nil {
		return err
	}
	app := &trayApp{
		client:     client,
		baseURL:    serviceURL,
		programs:   map[string]*programMenu{},
		quitCh:     make(chan struct{}),
		onShutdown: onShutdown,
	}
	systray.Run(app.onReady, app.onExit)
	return nil
}

type trayApp struct {
	client   *APIClient
	baseURL  string
	programs map[string]*programMenu
	quitCh   chan struct{}

	openUI *systray.MenuItem
	quit   *systray.MenuItem

	onShutdown func()
}

type programMenu struct {
	root    *systray.MenuItem
	start   *systray.MenuItem
	stop    *systray.MenuItem
	restart *systray.MenuItem
}

func (t *trayApp) onReady() {
	systray.SetIcon(icons.AppIcon)
	systray.SetTitle("Go App Manager")
	systray.SetTooltip("Go App Manager")

	t.openUI = systray.AddMenuItem("Open Web UI", "Open Web UI")
	systray.AddSeparator()

	t.buildProgramMenus()

	systray.AddSeparator()
	t.quit = systray.AddMenuItem("Quit", "Quit")

	go t.handleGlobalMenus()
	go t.refreshLoop()
}

func (t *trayApp) onExit() {
	close(t.quitCh)
	if t.onShutdown != nil {
		t.onShutdown()
	}
}

func (t *trayApp) buildProgramMenus() {
	progs, err := t.client.ListPrograms()
	if err != nil {
		log.Printf("list programs: %v", err)
		return
	}
	for _, p := range progs {
		if _, ok := t.programs[p.Name]; ok {
			continue
		}
		root := systray.AddMenuItem(programTitle(p), "Program status")
		pm := &programMenu{
			root:    root,
			start:   root.AddSubMenuItem("Start", "Start program"),
			stop:    root.AddSubMenuItem("Stop", "Stop program"),
			restart: root.AddSubMenuItem("Restart", "Restart program"),
		}
		t.programs[p.Name] = pm
		go t.handleProgramMenu(p.Name, pm)
	}
}

func (t *trayApp) handleProgramMenu(name string, pm *programMenu) {
	for {
		select {
		case <-pm.start.ClickedCh:
			_ = t.client.Start(name)
		case <-pm.stop.ClickedCh:
			_ = t.client.Stop(name)
		case <-pm.restart.ClickedCh:
			_ = t.client.Restart(name)
		case <-t.quitCh:
			return
		}
	}
}

func (t *trayApp) handleGlobalMenus() {
	for {
		select {
		case <-t.openUI.ClickedCh:
			t.launchBrowser()
		case <-t.quit.ClickedCh:
			systray.Quit()
			return
		case <-t.quitCh:
			return
		}
	}
}

func (t *trayApp) refreshLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.refreshPrograms()
		case <-t.quitCh:
			return
		}
	}
}

func (t *trayApp) refreshPrograms() {
	progs, err := t.client.ListPrograms()
	if err != nil {
		systray.SetTooltip("API error: " + err.Error())
		return
	}
	for _, p := range progs {
		menu, ok := t.programs[p.Name]
		if !ok {
			t.buildProgramMenus()
			return
		}
		menu.root.SetTitle(programTitle(p))
	}
}

func (t *trayApp) launchBrowser() {
	url := t.baseURL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	_ = cmd.Start()
}

func programTitle(p process.ProgramStatus) string {
	status := "stopped"
	if p.Running {
		status = fmt.Sprintf("running (PID %d)", p.PID)
	}
	return fmt.Sprintf("%s - %s", p.Name, status)
}

// Quit exits the systray loop.
func Quit() {
	systray.Quit()
}
