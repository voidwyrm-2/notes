package note

import (
	"errors"
	"math"
	"math/rand"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	NoteSize   float32 = 0.75
	spawnRange         = 80
)

type Note struct {
	pos  rl.Vector3
	text string
	seen bool
}

func New(pos rl.Vector3, text string) Note {
	return Note{pos: pos, text: text, seen: false}
}

func fromString(str string, prev [3]float32) (Note, error) {
	str = strings.TrimSpace(str)
	endt := strings.Index(str, "<end>")
	if endt == -1 {
		return Note{}, errors.New("missing '<end>' token")
	}

	text := strings.TrimSpace(str[:endt])
	str = strings.TrimSpace(str[endt+5:])

	floatsUnparsed := strings.Split(str, " ")
	if len(floatsUnparsed) != 3 {
		return Note{}, errors.New("expected three numbers")
	}

	floats := []float32{}
	for i, f := range floatsUnparsed {
		if f == "p" {
			floats = append(floats, prev[i])
		} else if f == "r" {
			r := rand.Intn(spawnRange)
			if r < spawnRange/2-1 {
				r -= spawnRange / 2
			}
			floats = append(floats, float32(r))
		} else if parsed, err := strconv.ParseFloat(f, 32); err != nil {
			return Note{}, err
		} else {
			floats = append(floats, float32(parsed))
		}
	}

	return New(rl.NewVector3(floats[0], floats[1], floats[2]), text), nil
}

func FromString(str string) (Note, error) {
	return fromString(str, [3]float32{0, 0, 0})
}

func FromStringMulti(str string) ([]Note, error) {
	unmarshalled := []Note{}

	for _, l := range strings.Split(strings.TrimSpace(str), "<ent>") {
		if strings.TrimSpace(l) == "" {
			continue
		}

		if n, err := fromString(l, func() [3]float32 {
			if len(unmarshalled) == 0 {
				return [3]float32{0, 0, 0}
			}

			p := unmarshalled[len(unmarshalled)-1].pos
			return [3]float32{p.X, p.Y, p.Z}
		}()); err != nil {
			return []Note{}, err
		} else {
			unmarshalled = append(unmarshalled, n)
		}
	}

	return unmarshalled, nil
}

func (n *Note) SetSeen(seen bool) {
	n.seen = seen
}

func (n Note) Near(pos rl.Vector3, allowedDist float32) bool {
	distX, distY, distZ := float32(math.Abs(float64(pos.X-n.pos.X)))-NoteSize, float32(math.Abs(float64(pos.Y-n.pos.Y)))-NoteSize, float32(math.Abs(float64(pos.Z-n.pos.Z)))-NoteSize
	return distX < allowedDist && distY < allowedDist && distZ < allowedDist
}

func (n Note) Text() string {
	return n.text
}

func (n Note) Draw() {
	if n.seen {
		rl.DrawCube(n.pos, NoteSize, NoteSize, NoteSize, rl.Violet)
	} else {
		rl.DrawCube(n.pos, NoteSize, NoteSize, NoteSize, rl.Purple)
	}
}
