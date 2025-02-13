package main

import (
	_ "embed"
	"flag"
	"fmt"

	scrres "github.com/fstanis/screenresolution"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/voidwyrm-2/notes/note"
)

func assert[T any](v T, _ error) T {
	return v
}

func t2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

const (
	defaultScreenWidth  = 800
	defaultScreenHeight = 450
)

//go:embed notes.txt
var notesRaw string

var (
	closeWindow      = false
	showControls     = true
	advancedControls = false

	nearestNote = -1
	viewingNote = false
	viewedText  = ""
	notes       = []note.Note{}
)

func main() {
	allowResizing := flag.Bool("resize", false, "Allow resizing of the window with your mouse")
	debugMode := flag.Bool("debug", false, "Enables debug mode")

	flag.Parse()

	if *debugMode {
		viewedText = "hello hello let me tell you what it's like to be a zero zero\nlet me show you what it's like to always feel feel like I'm empty and there's nothing really real real\nI'm looking for a way out"
		notes = append(notes, assert(note.FromString("hello there<end> 1 1 4")))
	}

	res := scrres.GetPrimary()
	if res == nil {
		res = &scrres.Resolution{Width: defaultScreenWidth, Height: defaultScreenHeight}
	}

	screenSize := struct{ width, height int32 }{
		width: int32(float32(res.Width) / 1.5), height: int32(float32(res.Height) / 1.5),
	}

	var err error
	notes, err = note.FromStringMulti(notesRaw)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *allowResizing {
		rl.SetConfigFlags(rl.FlagWindowResizable)
	}
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(screenSize.width, screenSize.height, "Notes")

	// Define the camera to look into our 3d world (position, target, up vector)
	camera := rl.Camera3D{}
	camera.Position = rl.NewVector3(0.0, 2.0, 4.0) // Camera position
	camera.Target = rl.NewVector3(0.0, 2.0, 0.0)   // Camera looking at point
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)       // Camera up vector (rotation towards target)
	camera.Fovy = 60.0                             // Camera field-of-view Y
	camera.Projection = rl.CameraPerspective       // Camera projection type

	rl.DisableCursor()  // Limit cursor to relative movement inside the window
	rl.SetTargetFPS(60) // Set our game to run at 60 frames-per-second

	for !closeWindow {
		if viewingNote {
			if rl.IsKeyPressed(rl.KeyEscape) {
				viewingNote = false
			}
		} else {
			if rl.WindowShouldClose() {
				closeWindow = true
			}

			if rl.IsKeyPressed(rl.KeyV) && *debugMode {
				viewingNote = true
			}

			if rl.IsKeyPressed(rl.KeyQ) {
				advancedControls = !advancedControls
			}

			if rl.IsKeyPressed(rl.KeyE) && nearestNote != -1 {
				viewedText = notes[nearestNote].Text()
				notes[nearestNote].SetSeen(true)
				viewingNote = true
			}

			if rl.IsKeyPressed(rl.KeyC) {
				showControls = !showControls
			}

			if advancedControls {
				// Camera PRO usage example (EXPERIMENTAL)
				// This new camera function allows custom movement/rotation values to be directly provided
				// as input parameters, with this approach, rcamera module is internally independent of raylib inputs
				rl.UpdateCameraPro(&camera,
					rl.NewVector3(
						t2f(rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp))*0.1- // Move forward-backward
							t2f(rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown))*0.1,
						t2f(rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight))*0.1- // Move right-left
							t2f(rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft))*0.1,
						0.0, // Move up-down
					),
					rl.NewVector3(
						rl.GetMouseDelta().X*0.05, // Rotation: yaw
						rl.GetMouseDelta().Y*0.05, // Rotation: pitch
						0.0,                       // Rotation: roll
					),
					0.0) // Move to target (zoom)
			} else {
				// Update camera computes movement internally depending on the camera mode.
				// Some default standard keyboard/mouse inputs are hardcoded to simplify use.
				// For advance camera controls, it's recommended to compute camera movement manually.
				camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
				rl.UpdateCamera(&camera, rl.CameraFirstPerson) // Update camera
				camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
			}
		}

		rl.BeginDrawing()

		if viewingNote {
			rl.ClearBackground(rl.Black)
			var x, y int32 = 0, 0
			for _, c := range viewedText {
				if c == '\n' {
					y++
					x = 0
					continue
				}

				rl.DrawText(string(c), x, y*18, 20, rl.White)

				if c == 'l' || c == 'i' || c == 'I' || c == '!' {
					x += 6
				} else if c == '?' {
					x += 16
				} else {
					x += 12
				}

				if x > screenSize.width {
					y++
					x = 0
				}
			}
		} else {
			rl.ClearBackground(rl.RayWhite)

			rl.BeginMode3D(camera)

			rl.DrawPlane(rl.NewVector3(0.0, 0.0, 0.0), rl.NewVector2(320.0, 320.0), rl.LightGray) // Draw ground

			// debug objects
			// rl.DrawCube(rl.NewVector3(-16, 2.5, 0), 1, 5, 32, rl.Blue)

			nearestNote = -1
			for i, n := range notes {
				n.Draw()
				if n.Near(camera.Position, 1) {
					nearestNote = i
				}
			}

			rl.EndMode3D()

			if nearestNote != -1 {
				rl.DrawText("Press E to read", int32(float32(screenSize.width)/2.37), int32(float32(screenSize.height)/1.5), 28, rl.DarkBlue)
			}

			// draw controls
			if showControls {
				rl.DrawRectangle(5, 5, 250, 115, rl.Fade(rl.SkyBlue, 0.5))
				rl.DrawRectangleLines(5, 5, 250, 115, rl.Blue)

				if advancedControls {
					rl.DrawText("Controls(advanced):", 15, 15, 10, rl.Black)
					rl.DrawText("- Move keys: WASD or arrow keys", 15, 30, 10, rl.Black)
					rl.DrawText("- Look around: mouse", 15, 45, 10, rl.Black)
				} else {
					rl.DrawText("Controls:", 15, 15, 10, rl.Black)
					rl.DrawText("- Move keys: WASD", 15, 30, 10, rl.Black)
					rl.DrawText("- Look around: arrow keys", 15, 45, 10, rl.Black)
				}

				rl.DrawText("- Turn advanced on/off (off by default): Q", 15, 60, 10, rl.Black)
				rl.DrawText("- Show controls on/off: C", 15, 75, 10, rl.Black)
				rl.DrawText("- Interact: E", 15, 90, 10, rl.Black)
				rl.DrawText("- Exit game or menus: Escape", 15, 105, 10, rl.Black)
			}

			// draw debug info
			if *debugMode {
				rl.DrawRectangle(screenSize.width-200, 5, 195, 100, rl.Fade(rl.SkyBlue, 0.5))
				rl.DrawRectangleLines(screenSize.width-200, 5, 195, 100, rl.Blue)

				rl.DrawText("Camera status:", screenSize.width-190, 15, 10, rl.Black)
				rl.DrawText(s2T("Position:", v2T(camera.Position)), screenSize.width-190, 30, 10, rl.Black)
				rl.DrawText(s2T("Target:", v2T(camera.Target)), screenSize.width-190, 45, 10, rl.Black)
				rl.DrawText(s2T("Up:", v2T(camera.Up)), screenSize.width-190, 60, 10, rl.Black)
				rl.DrawText("Misc Info:", screenSize.width-190, 75, 10, rl.Black)
				rl.DrawText(s2T("nearestNote:", fmt.Sprint(nearestNote)), screenSize.width-190, 90, 10, rl.Black)
			}
		}

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

// rndU generates a random uint8 value between min and max
func rndU(min, max int32) uint8 {
	return uint8(rl.GetRandomValue(min, max))
}

// rndF generates a random float32 value between min and max
func rndF(min, max int32) float32 {
	return float32(rl.GetRandomValue(min, max))
}

// s2t generates a Status item string from a name and value
func s2T(name, value string) string {
	return fmt.Sprintf(" - %s %s", name, value)
}

// v2T generates a string from a rl.Vector3
func v2T(v rl.Vector3) string {
	return fmt.Sprintf("%6.3f, %6.3f, %6.3f", v.X, v.Y, v.Z)
}
