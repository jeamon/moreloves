package main

// This is a beautiful go-based program to generate loves-based foods to be collected.
// It allows a user to use keyboard keys to move to left & right and to shoot for loves.

// Version  : 1.0
// Author   : Jerome AMON
// Created  : 26 November 2021

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

const (
	GAMEAREA    = "gamearea"
	LOVESAREA   = "lovesarea"
	USERAREA    = "userarea"
	INFOSVIEW   = "infos"
	POSITION    = "position"
	SCOREVIEW   = "score"
	BULLETSVIEW = "bullets"
	TIMERVIEW   = "timer"

	TWIDTH = 11
	PWIDTH = 19

	RESET       = "\033[0m"
	BLACK_BOLD  = "\033[1;30m"
	RED_BOLD    = "\033[1;31m"
	GREEN_BOLD  = "\033[1;32m"
	YELLOW_BOLD = "\033[1;33m"
	WHITE_BOLD  = "\033[1;37m"
)

var (
	cursorPosition  = make(chan string, 10)
	bulletDirection = make(chan int, 10)
	nextLoves       = make(chan struct{}, 3)
	increaseBullets = make(chan struct{}, 3)
	decreaseBullets = make(chan struct{}, 3)
	increaseLoves   = make(chan struct{}, 3)
	exit            = make(chan struct{})
	wg              sync.WaitGroup
)

func main() {

	f, err := os.OpenFile("logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("failed to create logs file.")
	}
	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(f)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.FgColor = gocui.ColorGreen
	g.SelFgColor = gocui.ColorRed
	g.BgColor = gocui.ColorBlack
	g.Cursor = true
	g.InputEsc = false
	g.Mouse = false

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Println("Could not set key binding:", err)
		return
	}

	maxX, maxY := g.Size()

	// Game view.
	gameView, err := g.SetView(GAMEAREA, 0, 0, maxX, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create game area view:", err)
		return
	}

	gameView.FgColor = gocui.ColorRed
	gameView.SelBgColor = gocui.ColorBlack
	gameView.SelFgColor = gocui.ColorRed
	gameView.Autoscroll = false
	gameView.Wrap = false
	gameView.Highlight = true
	gameView.Frame = false

	// Loves view.
	lovesView, err := g.SetView(LOVESAREA, 0, -1, maxX-1, 1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create loves area view:", err)
		return
	}

	lovesView.FgColor = gocui.ColorRed
	lovesView.SelBgColor = gocui.ColorBlack
	lovesView.SelFgColor = gocui.ColorRed
	lovesView.Autoscroll = false
	lovesView.Wrap = false
	lovesView.Editable = false
	lovesView.Frame = false

	// User view.
	userView, err := g.SetView(USERAREA, 0, maxY-6, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create user area view:", err)
		return
	}

	userView.FgColor = gocui.ColorGreen
	userView.SelBgColor = gocui.ColorBlack
	userView.SelFgColor = gocui.ColorRed
	userView.Autoscroll = false
	userView.Wrap = false
	userView.Highlight = true
	userView.Frame = false

	// Position view.
	positionView, err := g.SetView(POSITION, 0, maxY-3, PWIDTH, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create position view:", err)
		return
	}

	positionView.FgColor = gocui.ColorRed
	positionView.SelBgColor = gocui.ColorBlack
	positionView.SelFgColor = gocui.ColorYellow
	positionView.Editable = false
	positionView.Wrap = false

	// Score view.
	scoreView, err := g.SetView(SCOREVIEW, PWIDTH+1, maxY-3, 2*PWIDTH+1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create score view:", err)
		return
	}

	scoreView.FgColor = gocui.ColorGreen
	scoreView.SelBgColor = gocui.ColorBlack
	scoreView.SelFgColor = gocui.ColorYellow
	scoreView.Editable = false
	scoreView.Wrap = false

	// Bullets view.
	bulletsView, err := g.SetView(BULLETSVIEW, 2*PWIDTH+2, maxY-3, 3*PWIDTH+2, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create bullets view:", err)
		return
	}

	bulletsView.FgColor = gocui.ColorGreen
	bulletsView.SelBgColor = gocui.ColorBlack
	bulletsView.SelFgColor = gocui.ColorYellow
	bulletsView.Editable = false
	bulletsView.Wrap = false

	// Timer view.
	timerView, err := g.SetView(TIMERVIEW, 3*PWIDTH+3, maxY-3, 3*PWIDTH+3+TWIDTH, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create timer view:", err)
		return
	}
	//timerView.Title = " Timer "
	timerView.FgColor = gocui.ColorGreen
	timerView.SelBgColor = gocui.ColorBlack
	timerView.SelFgColor = gocui.ColorYellow
	timerView.Editable = false
	timerView.Wrap = false
	fmt.Fprint(timerView, " 00:00:00 ")

	// Infos view.
	infosView, err := g.SetView(INFOSVIEW, 3*PWIDTH+4+TWIDTH, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create statistics view:", err)
		return
	}

	infosView.Editable = false
	infosView.Frame = true
	infosView.Autoscroll = false
	infosView.FgColor = gocui.ColorWhite
	infosView.SelFgColor = gocui.ColorRed
	infosView.SelBgColor = gocui.ColorBlack
	fmt.Fprint(infosView, " Use Arrow Keys or <Tab> to move. Use <Space> or <Enter> to shoot loves ")

	// Apply keybindings to program.
	if err = keybindings(g); err != nil {
		log.Println("Failed to setup keybindings:", err)
		return
	}

	if _, err = g.SetCurrentView(USERAREA); err != nil {
		log.Println("Failed to set focus on user view:", err)
		return
	}
	_, _ = g.SetViewOnTop(LOVESAREA)
	userView.SetCursor(maxX/2, 0)

	wg.Add(1)
	go updateStatsView(g, positionView, scoreView, bulletsView, timerView, PWIDTH)

	wg.Add(1)
	go generateLoves(g, lovesView, maxX, 25*(maxY-6))

	wg.Add(1)
	go generateBullets(g, gameView, maxY-6)

	cursorPosition <- fmt.Sprintf("[Target X : %d]", maxX/2)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		close(exit)
		log.Println("Exited from the main loop:", err)
	}

	wg.Wait()
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Game view.
	_, err := g.SetView(GAMEAREA, 0, 0, maxX, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create statistics view:", err)
		return err
	}

	// Loves view.
	_, err = g.SetView(LOVESAREA, 0, -1, maxX-1, 1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create loves area view:", err)
		return err
	}

	// User view.
	_, err = g.SetView(USERAREA, 0, maxY-6, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create user area view:", err)
		return err
	}

	// Position view.
	_, err = g.SetView(POSITION, 0, maxY-3, PWIDTH, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create position view:", err)
		return err
	}

	// Score view.
	_, err = g.SetView(SCOREVIEW, PWIDTH+1, maxY-3, 2*PWIDTH+1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create score view:", err)
		return err
	}

	// Bullets view.
	_, err = g.SetView(BULLETSVIEW, 2*PWIDTH+2, maxY-3, 3*PWIDTH+2, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create bullets view:", err)
		return err
	}

	// Timer view.
	_, err = g.SetView(TIMERVIEW, 3*PWIDTH+3, maxY-3, 3*PWIDTH+3+TWIDTH, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create timer view:", err)
		return err
	}

	// Infos view.
	_, err = g.SetView(INFOSVIEW, 3*PWIDTH+4+TWIDTH, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create infos view:", err)
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(exit)
	return gocui.ErrQuit
}

// generateLoves randomly displays 3 loves on the view.
func generateLoves(g *gocui.Gui, lv *gocui.View, maxX, interval int) {
	defer wg.Done()
	rand.Seed(time.Now().UnixNano())
	displayLoves(g, lv, maxX)
	for {
		select {
		case <-exit:
			return
		case <-nextLoves:
			displayLoves(g, lv, maxX)
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

// displayLoves displays love icons to use as targets.
func displayLoves(g *gocui.Gui, lv *gocui.View, maxX int) {
	g.Update(func(g *gocui.Gui) error {
		x1 := rand.Intn(maxX + 1)
		x2 := rand.Intn(maxX + 1)
		x3 := rand.Intn(maxX + 1)

		min := x1
		mid := x2
		max := x3

		if x2 < min {
			min = x2
		}
		if x3 < min {
			min = x3
		}

		if x1 > max {
			max = x1
		}
		if x2 > max {
			max = x2
		}

		if min < x1 && x1 < max {
			mid = x1
		} else if min < x3 && x3 < max {
			mid = x3
		}

		l := strings.Repeat(" ", min) + GREEN_BOLD + "♥" + strings.Repeat(" ", mid-min-1) + RED_BOLD + "♦" + strings.Repeat(" ", max-mid-1) + GREEN_BOLD + "♥" + RESET
		lv.Clear()
		fmt.Fprintf(lv, l)
		return nil
	})
}

// centers a given string within a width by padding.
func center(s string, width int, fill string) string {
	return strings.Repeat(fill, (width-len(s))/2) + s + strings.Repeat(fill, (width-len(s))/2)
}

// updateStatsView displays statistics for some views.
func updateStatsView(g *gocui.Gui, positionView, scoreView, bulletsView, timerView *gocui.View, pwidth int) {
	defer wg.Done()
	var pos string
	score := 0
	bullets := 100
	secsElapsed, hrs, mins, secs := 0, 0, 0, 0

	g.Update(func(g *gocui.Gui) error {
		scoreView.Clear()
		fmt.Fprint(scoreView, center("Loves : "+strconv.Itoa(score), pwidth, " "))
		bulletsView.Clear()
		fmt.Fprint(bulletsView, center("Bullets : "+strconv.Itoa(bullets), pwidth, " "))
		return nil
	})

	for {
		select {
		case <-exit:
			return
		case pos = <-cursorPosition:
			g.Update(func(g *gocui.Gui) error {
				positionView.Clear()
				fmt.Fprint(positionView, center(pos, pwidth, " "))
				return nil
			})
		case <-increaseLoves:
			score++
			g.Update(func(g *gocui.Gui) error {
				scoreView.Clear()
				fmt.Fprint(scoreView, center("Loves : "+strconv.Itoa(score), pwidth, " "))
				return nil
			})
		case <-decreaseBullets:
			bullets--
			g.Update(func(g *gocui.Gui) error {
				bulletsView.Clear()
				fmt.Fprint(bulletsView, center("Bullets : "+strconv.Itoa(bullets), pwidth, " "))
				return nil
			})
		case <-increaseBullets:
			bullets++
			g.Update(func(g *gocui.Gui) error {
				bulletsView.Clear()
				fmt.Fprint(bulletsView, center("Bullets : "+strconv.Itoa(bullets), pwidth, " "))
				return nil
			})

		case <-time.After(1 * time.Second):
			g.Update(func(g *gocui.Gui) error {
				secsElapsed++
				hrs = int(secsElapsed / 3600)
				mins = int(secsElapsed / 60)
				secs = int(secsElapsed % 60)
				timerView.Clear()
				fmt.Fprintf(timerView, " %02d:%02d:%02d ", hrs, mins, secs)
				return nil
			})
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// generateBullets displays a bullet moving upper.
func generateBullets(g *gocui.Gui, gameView *gocui.View, bottomY int) {
	defer wg.Done()
	var dirX int

	for {
		select {
		case <-exit:
			return
		case dirX = <-bulletDirection:
			go moveBullet(g, gameView, bottomY, dirX)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// moveBullet emulates bullet moving toward loves area.
func moveBullet(g *gocui.Gui, gameView *gocui.View, bottomY, dirX int) {
	gameView.Clear()
	for i := bottomY; i > -1; i-- {
		g.Update(func(g *gocui.Gui) error {
			if err := gameView.SetCursor(dirX, i); err == nil {
				gameView.EditWrite('•')
			}
			return nil
		})
		time.Sleep(25 * time.Millisecond)
	}

	r, err := g.Rune(dirX+1, 0)
	if err == nil {
		if r == '♥' {
			increaseLoves <- struct{}{}
		} else if r == '♦' {
			increaseBullets <- struct{}{}
		} else {
			decreaseBullets <- struct{}{}
		}
	}
	gameView.Clear()
	nextLoves <- struct{}{}
}

// keybindings binds multiple keys to views.
func keybindings(g *gocui.Gui) error {
	var err error
	if err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeyEnter, gocui.ModNone, shootToLove); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeySpace, gocui.ModNone, shootToLove); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeyArrowUp, gocui.ModNone, shootToLove); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeyArrowLeft, gocui.ModNone, moveLeft); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeyArrowDown, gocui.ModNone, moveLeft4Steps); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeyArrowRight, gocui.ModNone, moveRight); err != nil {
		return err
	}

	if err = g.SetKeybinding(USERAREA, gocui.KeyTab, gocui.ModNone, moveRight4Steps); err != nil {
		return err
	}

	return nil
}

// moveLeft moves cursor to (currentX-1, currentY) position.
func moveLeft(g *gocui.Gui, uv *gocui.View) error {
	cx, cy := uv.Cursor()
	if err := uv.SetCursor(cx-1, cy); err == nil {
		cursorPosition <- fmt.Sprintf("[Target X : %d]", cx-1)
	}

	return nil
}

// moveLeft4Steps moves cursor to (currentX-4, currentY) position.
func moveLeft4Steps(g *gocui.Gui, uv *gocui.View) error {
	cx, cy := uv.Cursor()
	if err := uv.SetCursor(cx-4, cy); err == nil {
		cursorPosition <- fmt.Sprintf("[Target X : %d]", cx-4)
	}

	return nil
}

// moveRight moves cursor to (currentX+1, currentY) position.
func moveRight(g *gocui.Gui, uv *gocui.View) error {
	cx, cy := uv.Cursor()
	if err := uv.SetCursor(cx+1, cy); err == nil {
		cursorPosition <- fmt.Sprintf("[Target X : %d]", cx+1)
	}
	return nil
}

// moveRight4Steps moves cursor to (currentX+4, currentY) position.
func moveRight4Steps(g *gocui.Gui, uv *gocui.View) error {
	cx, cy := uv.Cursor()
	if err := uv.SetCursor(cx+4, cy); err == nil {
		cursorPosition <- fmt.Sprintf("[Target X : %d]", cx+4)
	}
	return nil
}

// shootToLove generates a bullet moving step by step over currentX position.
func shootToLove(g *gocui.Gui, gv *gocui.View) error {
	cx, _ := gv.Cursor()
	bulletDirection <- cx
	return nil
}
