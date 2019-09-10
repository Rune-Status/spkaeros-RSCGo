package entity

import (
	"log"
	"os"
)

//Entity A stationary scene entity within the game world.
type Entity struct {
	Location
	Index int
}

//LogWarning Log interface for warnings.
var LogWarning = log.New(os.Stdout, "[WARNING] ", log.Ltime|log.Lshortfile)

//List Represents a list of Entity scene entities.
type List struct {
	List []interface{}
}

func (l *List) NearbyPlayers(p *Player) []*Player {
	var players []*Player
	for _, v := range l.List {
		if v, ok := v.(*Player); ok && v.Index != p.Index && p.LongestDelta(v.Location) <= 15 {
			players = append(players, v)
		}
	}
	return players
}

func (l *List) RemovingPlayers(p *Player) []*Player {
	var players []*Player
	for _, v := range l.List {
		if v, ok := v.(*Player); ok && v.Index != p.Index && p.LongestDelta(v.Location) > 15 {
			players = append(players, p)
		}
	}
	return players
}

func (l *List) NearbyObjects(p *Player) []*Object {
	var objects []*Object
	for _, o1 := range l.List {
		if o1, ok := o1.(*Object); ok && p.LongestDelta(o1.Location) <= 20 {
			objects = append(objects, o1)
		}
	}
	return objects
}

func (l *List) RemovingObjects(p *Player) []*Object {
	var objects []*Object
	for _, o1 := range l.List {
		if o1, ok := o1.(*Object); ok && p.LongestDelta(o1.Location) > 20 {
			objects = append(objects, o1)
		}
	}
	return objects
}

//Add Add an entity to the list.
func (l *List) Add(e interface{}) {
	l.List = append(l.List, e)
}

//Contains Returns true if e is an element of l, otherwise returns false.
func (l *List) Contains(e interface{}) bool {
	for _, v := range l.List {
		if v == e {
			// Pointers should be comparable?
			return true
		}
	}

	return false
}

//Remove Removes Entity e from List l.
func (l *List) Remove(e interface{}) {
	elems := l.List
	for i, v := range elems {
		if v == e {
			last := len(elems) - 1
			elems[i] = elems[last]
			l.List = elems[:last]
			return
		}
	}
}
