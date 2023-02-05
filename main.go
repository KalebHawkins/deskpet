package main

import (
	"bytes"
	"image"
	"math/rand"

	_ "embed"
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/fox.png
var foxPng []byte

const (
	scrWidth     = 32
	scrHeight    = 32
	spriteWidth  = 32
	spriteHeight = 32
)

type FoxState int

const (
	Idle = iota
	LookBehind
	Trotting
	Pouncing
	Shocked
	Sleeping
	Dying
)

type Animation struct {
	Name         string
	Frames       []image.Rectangle
	CurrentFrame image.Rectangle
	FrameCount   int
	incrementor  int
	FrameSpeed   int
}

func (a *Animation) Update() {
	a.FrameCount = len(a.Frames) - 1
	a.CurrentFrame = a.Frames[a.incrementor/a.FrameSpeed%a.FrameCount]
	a.incrementor++
}

type Fox struct {
	Sprite      *ebiten.Image
	State       int
	Animations  []*Animation
	Speed       int
	PounceSpeed int
	Position    *image.Point
}

func (f *Fox) Update() {
	f.Animations[f.State].Update()
}

func (f *Fox) Draw(dst *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	if f.Speed < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteWidth, 0)
	} else {
		op.GeoM.Scale(1, 1)
	}

	dst.DrawImage(f.Sprite.SubImage(f.Animations[f.State].CurrentFrame).(*ebiten.Image), op)
}

func (f *Fox) Trot() {
	wx, _ := ebiten.WindowPosition()
	f.Position.X = wx + f.PounceSpeed
	f.Flip()
}

func (f *Fox) Pounce() {
	wx, _ := ebiten.WindowPosition()
	f.Position.X = wx + f.PounceSpeed
	f.Flip()
}

func (f *Fox) Flip() {
	sx, _ := ebiten.ScreenSizeInFullscreen()
	wsx, _ := ebiten.WindowSize()

	if f.Position.X >= sx-wsx-spriteWidth {
		f.Speed = -f.Speed
		f.PounceSpeed = -f.PounceSpeed
	}

	if f.Position.X <= sx/2 {
		f.Speed = -f.Speed
		f.PounceSpeed = -f.PounceSpeed
	}
}

type Game struct {
	Background      *ebiten.Image
	Fox             *Fox
	DisplayPosition *image.Point
	tickCounter     int
	StateTimer      int
}

func (g *Game) Update() error {
	g.Fox.Update()

	ebiten.SetWindowPosition(g.Fox.Position.X, g.Fox.Position.Y)

	mx, my := ebiten.CursorPosition()

	if int(g.tickCounter)%ebiten.TPS() == 0 {
		g.StateTimer++
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) &&
		mx >= 15 && my >= 15 && mx <= 25 && my <= 25 {
		g.StateTimer = 0
		g.Fox.State = Shocked
	}

	trottingProbability := rand.Float64()
	if g.Fox.State == Idle && g.StateTimer > 5 && trottingProbability > 0.6 {
		g.StateTimer = 0
		g.Fox.State = Trotting
	}

	sleepProbability := rand.Float64()
	if g.Fox.State == Idle && g.StateTimer > 5 && sleepProbability < 0.5 {
		g.StateTimer = 0
		g.Fox.State = LookBehind
	}

	if g.Fox.State == Idle && g.StateTimer > 5 && sleepProbability > 0.5 {
		g.StateTimer = 0
		g.Fox.State = Sleeping
	}

	if mx >= -10 && mx <= -5 && my >= 15 && my <= 30 ||
		mx >= 30 && mx <= 40 && my >= 15 && my <= 30 {
		g.StateTimer = 0
		g.Fox.State = Pouncing
	}

	switch g.Fox.State {
	case LookBehind:
		if g.StateTimer == 3 {
			g.Fox.State = Idle
		}
	case Shocked:
		if g.StateTimer == 2 {
			g.StateTimer = 0
			g.Fox.State = Idle
		}
	case Trotting:
		if g.StateTimer > 2 {
			g.Fox.State = Idle
		}
		g.Fox.Trot()
	case Pouncing:
		if g.StateTimer > 2 {
			g.Fox.State = Idle
		}
		g.Fox.Pounce()
	}

	g.tickCounter++
	return nil
}

func (g *Game) Draw(scr *ebiten.Image) {
	g.Background.Fill(color.RGBA{0, 0, 0, 0})
	g.Fox.Draw(g.Background)
	scr.DrawImage(g.Background, nil)
}

func (g *Game) Layout(outWidth, outHeight int) (int, int) {
	return scrWidth, scrHeight
}

func NewGame() *Game {
	foxImg, _, err := image.Decode(bytes.NewReader(foxPng))
	if err != nil {
		panic(err)
	}

	foxAnimations := make([]*Animation, 0)
	foxAnimations = append(foxAnimations, &Animation{
		Name:       "Idle",
		Frames:     generateFrames(5, Idle, 1),
		FrameSpeed: 15,
		FrameCount: 5,
	})

	foxAnimations = append(foxAnimations, &Animation{
		Name:       "LookBehind",
		Frames:     generateFrames(14, LookBehind, 2),
		FrameSpeed: 15,
		FrameCount: 14,
	})

	foxAnimations = append(foxAnimations, &Animation{
		Name:       "Trotting",
		Frames:     generateFrames(8, Trotting, 3),
		FrameSpeed: 15,
		FrameCount: 8,
	})

	foxAnimations = append(foxAnimations, &Animation{
		Name:       "Pouncing",
		Frames:     generateFrames(11, Pouncing, 4),
		FrameSpeed: 15,
		FrameCount: 11,
	})

	foxAnimations = append(foxAnimations, &Animation{
		Name:       "Shocked",
		Frames:     generateFrames(5, Shocked, 5),
		FrameSpeed: 15,
		FrameCount: 5,
	})

	foxAnimations = append(foxAnimations, &Animation{
		Name:       "Sleeping",
		Frames:     generateFrames(6, Sleeping, 6),
		FrameSpeed: 15,
		FrameCount: 6,
	})

	foxAnimations = append(foxAnimations, &Animation{
		Name:       "Dying",
		Frames:     generateFrames(7, Dying, 7),
		FrameSpeed: 15,
		FrameCount: 7,
	})

	sx, sy := ebiten.ScreenSizeInFullscreen()

	return &Game{
		Background: ebiten.NewImage(scrWidth, scrHeight),
		Fox: &Fox{
			Sprite:      ebiten.NewImageFromImage(foxImg),
			State:       Idle,
			Animations:  foxAnimations,
			Speed:       2,
			PounceSpeed: 1,
			Position:    &image.Point{sx / 2, sy - 128 - 32},
		},
		DisplayPosition: &image.Point{0, 0},
	}
}

func main() {
	ebiten.SetWindowSize(scrWidth*4, scrHeight*4)
	ebiten.SetWindowTitle("Deskpet")
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetScreenTransparent(true)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}

func generateFrames(frameCount int, state int, yOffSet int) []image.Rectangle {
	animationFrames := make([]image.Rectangle, 0)

	for i := 0; i < frameCount; i++ {
		animationFrames = append(animationFrames, image.Rectangle{
			Min: image.Point{X: i * spriteWidth, Y: state * spriteHeight},
			Max: image.Point{X: (i + 1) * spriteWidth, Y: yOffSet * spriteHeight},
		})
	}

	return animationFrames
}
