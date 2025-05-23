package store

import (
	"errors"
	"github.com/AyushSriv06/ItShares/types"
	"github.com/google/uuid"
	"sync"
)

type RoomStore interface {
	CreateRoom(room *types.Room) error
	GetRoomById(id uuid.UUID) (*types.Room, error)
	GetRoomByCode(code string) (*types.Room, error)
	UpdateRoom(id uuid.UUID, room *types.Room) (*types.Room, error)
	DeleteRoom(id uuid.UUID) error
}

type RoomStoreInMem struct {
	rooms       map[uuid.UUID]*types.Room
	roomsByCode map[string]*types.Room
	mu          sync.RWMutex
}

// var _ RoomStore = (*RoomStoreInMem)(nil)

func NewRoomStoreInMemory() *RoomStoreInMem {
	return &RoomStoreInMem{
		rooms:       make(map[uuid.UUID]*types.Room),
		roomsByCode: make(map[string]*types.Room),
	}
}

func (r *RoomStoreInMem) roomWithIdExists(id uuid.UUID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room := r.rooms[id]
	return room != nil
}

func (r *RoomStoreInMem) roomWithCodeExists(code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room := r.roomsByCode[code]
	return room != nil
}

func (r *RoomStoreInMem) CreateRoom(room *types.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rooms[room.ID] = room
	r.roomsByCode[room.Code] = room

	return nil
}

func (r *RoomStoreInMem) GetRoomById(id uuid.UUID) (*types.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	room := r.rooms[id]
	if room == nil {
		return nil, errors.New("no room found")
	}

	return room, nil
}

func (r *RoomStoreInMem) GetRoomByCode(code string) (*types.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	room := r.roomsByCode[code]
	if room == nil {
		return nil, errors.New("no room found")
	}

	return room, nil
}

func (r *RoomStoreInMem) UpdateRoom(id uuid.UUID, room *types.Room) (*types.Room, error) {
	if !r.roomWithIdExists(id) {
		return nil, errors.New("no room found")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	oldRoom := r.rooms[id]
	r.rooms[id] = room
	r.roomsByCode[oldRoom.Code] = room

	return room, nil
}

func (r *RoomStoreInMem) DeleteRoom(id uuid.UUID) error {
	if !r.roomWithIdExists(id) {
		return errors.New("no room found")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	oldRoom := r.rooms[id]
	delete(r.rooms, oldRoom.ID)
	delete(r.roomsByCode, oldRoom.Code)

	return nil
}