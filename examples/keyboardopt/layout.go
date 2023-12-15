package main

import (
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strings"
	"unicode"
)

var allKeys = []rune{
	'a', 'b', 'c', 'd', 'e', 'f',
	'g', 'h', 'i', 'j', 'k', 'l',
	'm', 'n', 'o', 'p', 'q', 'r',
	's', 't', 'u', 'v', 'w', 'x',
	'y', 'z',
}

var allSlots []Pt

var (
	bounds = Pt{10, 3} // Oversized by 1 key.
	lhrk   = Pt{3, 1}  // QWERTY F.
	rhrk   = Pt{6, 1}  // QWERTY J.
)

func init() {
	for i := byte(0); i < bounds.X; i++ {
		for j := byte(0); j < bounds.Y; j++ {
			if p := (Pt{i, j}); p.Valid() {
				allSlots = append(allSlots, p)
			}
		}
	}
}

type Pt struct {
	X, Y byte
}

func (p Pt) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// ..........
// .........x
// x.......xx
func NewRandomValidPt(r *rand.Rand) Pt {
	i := r.Intn(26)
	if i < 10 {
		return Pt{byte(i), 0}
	}
	i -= 10
	if i < 9 {
		return Pt{byte(i), 1}
	}
	i -= 9
	return Pt{byte(i) + 1, 2}
}

func (p Pt) Valid() bool {
	return p.X <= 10 && p.Y < 3 &&
		p != Pt{9, 1} &&
		!(p.Y == 2 && (p.X == 0 || p.X >= 8))
}

type Key = rune

type Layout struct {
	Bounds Pt // Maximum bounds of the Layout.
	LHRK   Pt // Left home row key point.
	RHRK   Pt // Right home row key point.
	Slots  map[Pt]Key
	Keys   map[Key]Pt
}

func NewRandomLayout(r *rand.Rand) Layout {
	a := Layout{
		Bounds: bounds,
		LHRK:   lhrk,
		RHRK:   rhrk,
		Slots:  make(map[Pt]Key, len(allKeys)),
		Keys:   make(map[Key]Pt, len(allKeys)),
	}
	slots := slices.Clone(allSlots)
	keys := slices.Clone(allKeys)
	r.Shuffle(len(keys), func(i, j int) {
		slots[i], slots[j] = slots[j], slots[i]
		keys[i], keys[j] = keys[j], keys[i]
	})
	for i, k := range keys {
		p := slots[i]
		a.Keys[k] = p
		a.Slots[p] = k
	}
	return a
}

func (a Layout) Clone() Layout {
	return Layout{
		Bounds: a.Bounds,
		LHRK:   a.LHRK,
		RHRK:   a.RHRK,
		Slots:  maps.Clone(a.Slots),
		Keys:   maps.Clone(a.Keys),
	}
}

func (a Layout) Swap(p1, p2 Pt) {
	k1 := a.Slots[p1]
	k2 := a.Slots[p2]
	a.Slots[p1] = k2
	a.Slots[p2] = k1
	a.Keys[k1] = p2
	a.Keys[k2] = p1
}

func (a Layout) String() string {
	var sb strings.Builder
	for j := byte(0); j < a.Bounds.Y; j++ {
		for i := byte(0); i < a.Bounds.X; i++ {
			k := a.Slots[Pt{i, j}]
			if k == 0 {
				k = '.'
			}
			sb.WriteRune(k)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Test evaluates the Layout on the given string of text.
// The result is travel distance from the homerow.
// More precisely, the minimum distance from either home row key.
// Keys not present in the layout are assigned a default loss.
func (a Layout) Test(text string) (score, hits int) {
	defaultValue := 9999
	for _, r := range text {
		if 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' {
			hits++
			r = unicode.ToLower(r)
			if p, ok := a.Keys[Key(r)]; ok {
				score += a.Dist(p)
			} else {
				score += defaultValue
			}
		}
	}
	return score, hits
}

var pscores = [5][12]int{
	{0300, 0200, 0200, 0200, 0300, 0300, 0200, 0200, 0200, 0300},
	{0000, 0000, 0000, 0000, 0100, 0100, 0000, 0000, 0000, 9999},
	{9999, 0300, 0400, 0100, 0100, 0300, 0200, 0400, 9999, 9999},
}

func (a Layout) Dist(p Pt) int {
	return pscores[p.Y][p.X]
}
