package main

import (
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type Option func(*GamePerson)

func WithName(name string) func(*GamePerson) {
	return func(person *GamePerson) {
		person.nickname = *(*[42]byte)(unsafe.Pointer(unsafe.StringData(name)))
	}
}

func WithCoordinates(x, y, z int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.x = int32(x)
		person.y = int32(y)
		person.z = int32(z)
	}
}

func WithGold(gold int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.golds = uint32(gold)
	}
}

func WithMana(mana int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.baseMask = setBits[uint32](person.baseMask, uint32(mana), 10, 0)
	}
}

func WithHealth(health int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.baseMask = setBits[uint32](person.baseMask, uint32(health), 10, 10)
	}
}

func WithRespect(respect int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.skillMask = setBits[uint16](person.skillMask, uint16(respect), 4, 0)
	}
}

func WithStrength(strength int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.skillMask = setBits[uint16](person.skillMask, uint16(strength), 4, 4)
	}
}

func WithExperience(experience int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.skillMask = setBits[uint16](person.skillMask, uint16(experience), 4, 8)
	}
}

func WithLevel(level int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.skillMask = setBits[uint16](person.skillMask, uint16(level), 4, 12)
	}
}

func WithHouse() func(*GamePerson) {
	return func(person *GamePerson) {
		person.baseMask = setBits[uint32](person.baseMask, uint32(1), 1, 20)
	}
}

func WithGun() func(*GamePerson) {
	return func(person *GamePerson) {
		person.baseMask = setBits[uint32](person.baseMask, uint32(1), 1, 21)
	}
}

func WithFamily() func(*GamePerson) {
	return func(person *GamePerson) {
		person.baseMask = setBits[uint32](person.baseMask, uint32(1), 1, 22)
	}
}

func WithType(personType int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.baseMask = setBits[uint32](person.baseMask, uint32(personType), 2, 23)
	}
}

const (
	BuilderGamePersonType = iota
	BlacksmithGamePersonType
	WarriorGamePersonType
)

type GamePerson struct {
	x         int32    // 4
	y         int32    // 4
	z         int32    // 4
	golds     uint32   // 4
	nickname  [42]byte // 42
	skillMask uint16   // 2 | 4 bits - lvl,  4 bits - exp, 4 bits - strength, 4 bits - respect
	baseMask  uint32   // 4 | 2 bits - type, 1 bit - family, 1 bit - gun, 1 bit - home, 10 bits - hp, 10 bits - mana
}

func getBits[T ~uint16 | ~uint32](x T, limit, offset uint8) T {
	return (x >> offset) & ((1 << limit) - 1)
}

func setBits[T ~uint16 | ~uint32](x, set T, limit, offset uint8) T {
	mask := (T(1)<<limit - 1) << offset
	// 1. очищаем место куда зотим вставить
	// 2. обрезаем нужную нам участок битов
	return (x & ^mask) | ((set & (T(1)<<limit - 1)) << offset)
}

func NewGamePerson(options ...Option) GamePerson {
	gp := GamePerson{}
	for _, option := range options {
		option(&gp)
	}
	return gp
}

func (p *GamePerson) Name() string {
	return unsafe.String(&p.nickname[0], len(p.nickname))
}

func (p *GamePerson) X() int {
	return int(p.x)
}

func (p *GamePerson) Y() int {
	return int(p.y)
}

func (p *GamePerson) Z() int {
	return int(p.z)
}

func (p *GamePerson) Gold() int {
	return int(p.golds)
}

func (p *GamePerson) Mana() int {
	return int(getBits[uint32](p.baseMask, 10, 0))
}

func (p *GamePerson) Health() int {
	return int(getBits[uint32](p.baseMask, 10, 10))
}

func (p *GamePerson) Respect() int {
	return int(getBits[uint16](p.skillMask, 4, 0))
}

func (p *GamePerson) Strength() int {
	return int(getBits[uint16](p.skillMask, 4, 4))
}

func (p *GamePerson) Experience() int {
	return int(getBits[uint16](p.skillMask, 4, 8))

}

func (p *GamePerson) Level() int {
	return int(getBits[uint16](p.skillMask, 4, 12))
}

func (p *GamePerson) HasHouse() bool {
	return getBits[uint32](p.baseMask, 1, 20) == 1

}

func (p *GamePerson) HasGun() bool {
	return getBits[uint32](p.baseMask, 1, 21) == 1
}

func (p *GamePerson) HasFamilty() bool {
	return getBits[uint32](p.baseMask, 1, 22) == 1
}

func (p *GamePerson) Type() int {
	return int(getBits[uint32](p.baseMask, 2, 23))
}

func TestGamePerson(t *testing.T) {
	assert.LessOrEqual(t, unsafe.Sizeof(GamePerson{}), uintptr(64))

	const x, y, z = math.MinInt32, math.MaxInt32, 0
	const name = "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc"
	const personType = BuilderGamePersonType
	const gold = math.MaxInt32
	const mana = 1000
	const health = 1000
	const respect = 10
	const strength = 10
	const experience = 10
	const level = 10

	options := []Option{
		WithName(name),
		WithCoordinates(x, y, z),
		WithGold(gold),
		WithMana(mana),
		WithHealth(health),
		WithRespect(respect),
		WithStrength(strength),
		WithExperience(experience),
		WithLevel(level),
		WithHouse(),
		WithFamily(),
		WithType(personType),
	}

	person := NewGamePerson(options...)
	assert.Equal(t, name, person.Name())
	assert.Equal(t, x, person.X())
	assert.Equal(t, y, person.Y())
	assert.Equal(t, z, person.Z())
	assert.Equal(t, gold, person.Gold())
	assert.Equal(t, mana, person.Mana())
	assert.Equal(t, health, person.Health())
	assert.Equal(t, respect, person.Respect())
	assert.Equal(t, strength, person.Strength())
	assert.Equal(t, experience, person.Experience())
	assert.Equal(t, level, person.Level())
	assert.True(t, person.HasHouse())
	assert.True(t, person.HasFamilty())
	assert.False(t, person.HasGun())
	assert.Equal(t, personType, person.Type())
}
