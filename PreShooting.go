package main

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "image/color"
    "math/rand"
    "time"
    "fmt"
)

type GameObject struct {
    image  fyne.CanvasObject
    speed  fyne.Position
    active bool
}

var (
    spaceship *GameObject
    bullets   []*GameObject
    enemies   []*GameObject
    score     int
    scoreLabel *canvas.Text
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Space Taraveler")

    gameArea := canvas.NewRectangle(color.Black)
    gameArea.SetMinSize(fyne.NewSize(600, 400))

    spaceship = &GameObject{
        image:  canvas.NewImageFromFile("spaceship.png"),
        speed:  fyne.NewPos(0, 0),
        active: true,
    }
    spaceship.image.Resize(fyne.NewSize(50, 50))
    spaceship.image.Move(fyne.NewPos(275, 350))

    content := container.NewWithoutLayout(gameArea, spaceship.image)

    go generateEnemies(content, myWindow)
    go gameLoop(content, myWindow)

    myWindow.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
        handleKeyInput(e, content, myWindow)
    })

    score = 0
    scoreLabel = canvas.NewText("Score: 0", color.White)
    scoreLabel.TextSize = 24
    scoreLabel.Move(fyne.NewPos(10, 10)) 
    content.Add(scoreLabel)

    myWindow.SetContent(content)
    myWindow.Resize(fyne.NewSize(600, 400))
    myWindow.ShowAndRun()
}


func generateEnemies(content *fyne.Container, window fyne.Window) {
	for {
		time.Sleep(5 * time.Second)
		enemy := createEnemy()
		enemies = append(enemies, enemy)
		content.Add(enemy.image)
		window.Content().Refresh()
	}
}

func gameLoop(content *fyne.Container, window fyne.Window) {
	for {
		time.Sleep(5.0 * time.Millisecond)

		activeBullets := []*GameObject{}
		for _, bullet := range bullets {
			if bullet.active {
				moveGameObject(bullet, content, window)
				checkBulletCollision(bullet, content, window)
				if bullet.active {
					activeBullets = append(activeBullets, bullet)
				}
			}
		}
		bullets = activeBullets

		// 敵の移動と画面下部への到達チェック
		for _, enemy := range enemies {
			if enemy.active {
				moveGameObject(enemy, content, window)
				if enemy.image.Position().Y > window.Canvas().Size().Height { // 画面の下端に到達
					enemy.active=false
					return
				}
			}
		}

		window.Content().Refresh()
	}
}

func moveGameObject(obj *GameObject, content *fyne.Container, window fyne.Window) {
	if !obj.active {
		return
	}
	pos := obj.image.Position()
	newPos := fyne.NewPos(pos.X+obj.speed.X, pos.Y+obj.speed.Y)

	// 画面外に出たオブジェクトはコンテナから削除
	if newPos.X < 0 || newPos.X > 600 || newPos.Y < 0 || newPos.Y > 400 {
		obj.active = false
		content.Remove(obj.image) // ここでコンテナから削除
		return
	}

	obj.image.Move(newPos)
}

func checkBulletCollision(bullet *GameObject, content *fyne.Container, window fyne.Window) {
	for i, enemy := range enemies {
		if enemy.active && bullet.active && overlapping(bullet.image, enemy.image) {
			// ... 爆発の表示など ...
			showExplosionAt(enemy.image.Position(), content, window)

			// 得点を加算して表示を更新
			score ++
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
	explosion := canvas.NewImageFromFile("explosion.png")
	explosion.Resize(fyne.NewSize(50, 50)) // 適切なサイズに調整
	explosion.Move(pos)
	content.Add(explosion)

	// 爆発を一定時間表示した後に消す
	go func() {
		time.Sleep(500 * time.Millisecond) // 500ミリ秒後に消す
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
	pos := spaceship.image.Position()
	switch e.Name {
	case fyne.KeyLeft:
		if pos.X > 10 { // 左端からのマージンを確保
			spaceship.image.Move(fyne.NewPos(pos.X-10, pos.Y))
		}
	case fyne.KeyRight:
		if pos.X < 550 { // 右端からのマージンを確保（画面幅 - 宇宙船の幅）
			spaceship.image.Move(fyne.NewPos(pos.X+10, pos.Y))
		}
	case fyne.KeyUp:
		if pos.Y > 10 { // 上端からのマージンを確保
			spaceship.image.Move(fyne.NewPos(pos.X, pos.Y-10))
		}
	case fyne.KeyDown:
		if pos.Y < 350 { // 下端からのマージンを確保（画面高さ - 宇宙船の高さ）
			spaceship.image.Move(fyne.NewPos(pos.X, pos.Y+10))
		}
	case fyne.KeySpace:
		bullet := createBullet()
		bullets = append(bullets, bullet)
		content.Add(bullet.image)
	}
	window.Content().Refresh()
}


func createEnemy() *GameObject {
	enemyImage := canvas.NewImageFromFile("enemy.png")
	enemyImage.Resize(fyne.NewSize(40, 40))
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
