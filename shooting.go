package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type GameObject struct {
	image  fyne.CanvasObject
	speed  fyne.Position
	active bool
}

var (
	spaceship     *GameObject
	bullets       []*GameObject
	enemies       []*GameObject
	score         int
	scoreLabel    *canvas.Text
	gameOverLabel *canvas.Text
	gameRunning   bool = true
	gameObjects   *fyne.Container
	enemyImageResource fyne.Resource
	explosionImageResource fyne.Resource
	spaceshipImageResource fyne.Resource
	backgroundImageResource fyne.Resource
)

func loadResources() {
	var err error
	enemyImageResource, err = fyne.LoadResourceFromPath("enemy.png")
	if err != nil {
		log.Fatal("Failed to load enemy image:", err)
	}
	explosionImageResource, err = fyne.LoadResourceFromPath("explosion.png")
	if err != nil {
		log.Fatal("Failed to load explosion image:", err)
	}
	spaceshipImageResource, err = fyne.LoadResourceFromPath("spaceship.png")
	if err != nil {
		log.Fatal("Failed to load spaceship image:", err)
	}
	backgroundImageResource, err = fyne.LoadResourceFromPath("space_background.png")
	if err!=nil{
		log.Fatal("Failed to load space_background image:", err)
	}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Space Traveler")

	loadResources()

	icon, err := fyne.LoadResourceFromPath("spaceship.png")
	if err != nil {
		fmt.Println("Failed to load app icon:", err)
		return
	}
	myWindow.SetIcon(icon)

	backgroundImage := canvas.NewImageFromResource(backgroundImageResource)
	backgroundImage.FillMode = canvas.ImageFillStretch

	gameObjects = container.NewWithoutLayout()

	spaceship = &GameObject{
		image:  canvas.NewImageFromResource(spaceshipImageResource),
		speed:  fyne.NewPos(0, 0),
		active: true,
	}
	spaceship.image.Resize(fyne.NewSize(50, 50))
	spaceship.image.Move(fyne.NewPos(265, 350))
	gameObjects.Add(spaceship.image)

	scoreLabel = canvas.NewText("Score: 0", color.White)
	scoreLabel.TextSize = 24
	scoreLabel.Move(fyne.NewPos(10, 10))
	gameObjects.Add(scoreLabel)

	gameOverLabel = canvas.NewText("Game Over", color.White)
	gameOverLabel.TextSize = 24
	gameOverLabel.Hidden = true
	gameOverLabel.Move(fyne.NewPos(200, 200))
	gameObjects.Add(gameOverLabel)

	content := container.NewStack(backgroundImage, gameObjects)

	myWindow.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		handleKeyInput(e, gameObjects, myWindow)
	})

	go generateEnemies(gameObjects, myWindow)
	go gameLoop(gameObjects, myWindow)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}

func generateEnemies(content *fyne.Container, window fyne.Window) {
	for {
		time.Sleep(4 * time.Second)
		if !gameRunning {
			return
		}
		enemy := createEnemy()
		enemies = append(enemies, enemy)

		if !gameRunning {
			return
		}
		content.Add(enemy.image)
		window.Content().Refresh()
	}
}

func gameLoop(content *fyne.Container, window fyne.Window) {
	for {
		time.Sleep(5.0 * time.Millisecond)

		windowSize := window.Canvas().Size()

		activeBullets := []*GameObject{}
		for _, bullet := range bullets {
			if bullet.active {
				moveGameObject(bullet, content)
				checkBulletCollision(bullet, content, window)
				if bullet.active {
					activeBullets = append(activeBullets, bullet)
				}
			}
		}
		bullets = activeBullets

		activeEnemies := []*GameObject{}
		for _, enemy := range enemies {
			if enemy.active {
				moveGameObject(enemy, content)
				if enemy.image.Position().Y+enemy.image.Size().Height > windowSize.Height {
					endGame(window)
					return
				} else {
					activeEnemies = append(activeEnemies, enemy)
				}
			}
		}
		enemies = activeEnemies

		window.Content().Refresh()
	}
}

func moveGameObject(obj *GameObject, content *fyne.Container) {
	if !obj.active {
		return
	}
	pos := obj.image.Position()
	newPos := fyne.NewPos(pos.X+obj.speed.X, pos.Y+obj.speed.Y)

	if newPos.X < 0 || newPos.X > 600 || newPos.Y < 0 || newPos.Y > 400 {
		obj.active = false
		content.Remove(obj.image)
		return
	}

	obj.image.Move(newPos)
}

func checkBulletCollision(bullet *GameObject, content *fyne.Container, window fyne.Window) {
	for i, enemy := range enemies {
		if enemy.active && bullet.active && overlapping(bullet.image, enemy.image) {
			showExplosionAt(enemy.image.Position(), content, window)
			score++
			scoreLabel.Text = fmt.Sprintf("Score: %d", score)
			scoreLabel.Refresh()
			content.Remove(enemy.image)
			content.Remove(bullet.image)
			enemy.active = false
			bullet.active = false
			enemies = append(enemies[:i], enemies[i+1:]...)
			break
		}
	}
}

func showExplosionAt(pos fyne.Position, content *fyne.Container, window fyne.Window) {
	explosion := canvas.NewImageFromResource(explosionImageResource)
	explosion.Resize(fyne.NewSize(40, 40))
	explosion.Move(pos)
	content.Add(explosion)
	go func() {
		time.Sleep(500 * time.Millisecond)
		content.Remove(explosion)
		window.Content().Refresh()
	}()
}

func overlapping(obj1, obj2 fyne.CanvasObject) bool {
	obj1Pos := obj1.Position()
	obj1Size := obj1.Size()
	obj2Pos := obj2.Position()
	obj2Size := obj2.Size()

	return obj1Pos.X < obj2Pos.X+obj2Size.Width &&
		obj1Pos.X+obj1Size.Width > obj2Pos.X &&
		obj1Pos.Y < obj2Pos.Y+obj2Size.Height &&
		obj1Pos.Y+obj1Size.Height > obj2Pos.Y
}

func handleKeyInput(e *fyne.KeyEvent, content *fyne.Container, window fyne.Window) {
	if !gameRunning {
		return
	}
	pos := spaceship.image.Position()
	switch e.Name {
	case fyne.KeyLeft:
		if pos.X > 10 {
			spaceship.image.Move(fyne.NewPos(pos.X-10, pos.Y))
		}
	case fyne.KeyRight:
		if pos.X < 550 {
			spaceship.image.Move(fyne.NewPos(pos.X+10, pos.Y))
		}
	case fyne.KeySpace:
		bullet := createBullet()
		bullets = append(bullets, bullet)
		content.Add(bullet.image)
	}
	window.Content().Refresh()
}

func createEnemy() *GameObject {
	enemyImage := canvas.NewImageFromResource(enemyImageResource)
	enemyImage.Resize(fyne.NewSize(36, 18))
	initialX := rand.Float32() * 550
	enemy := &GameObject{
		image:  enemyImage,
		speed:  fyne.NewPos(0, 2),
		active: true,
	}
	enemy.image.Move(fyne.NewPos(initialX, 0))
	return enemy
}

func createBullet() *GameObject {
	bulletImage := canvas.NewRectangle(color.White)
	bulletImage.Resize(fyne.NewSize(5, 20))
	bulletPos := spaceship.image.Position()
	bullet := &GameObject{
		image:  bulletImage,
		speed:  fyne.NewPos(0, -5),
		active: true,
	}
	bullet.image.Move(fyne.NewPos(bulletPos.X+22.5, bulletPos.Y))
	return bullet
}

func endGame(window fyne.Window) {
	gameRunning = false

	for _, enemy := range enemies {
		enemy.active = false
		gameObjects.Remove(enemy.image)
	}
	enemies = []*GameObject{}

	for _, bullet := range bullets {
		bullet.active = false
		gameObjects.Remove(bullet.image)
	}
	bullets = []*GameObject{}

	gameOverLabel.Hidden = false
	gameOverLabel.Refresh()

	scoreLabel.Move(fyne.NewPos(200, 240))
	scoreLabel.Refresh()

	window.Content().Refresh()
}
