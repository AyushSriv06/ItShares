package service

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/AyushSriv06/ItShares/store"
	"github.com/AyushSriv06/ItShares/types"
	"github.com/google/uuid"
)

var (
	ErrorRoomNotFound       = errors.New("room not found")
	ErrorRoomNotEmpty       = errors.New("room is not empty")
	ErrorDisplayNameInUse   = errors.New("display name already in use")
	ErrorCouldNotJoinRoom   = errors.New("could not join room")
	ErrorCouldNotLeaveRoom  = errors.New("could not leave room")
	ErrorCouldNotCreateRoom = errors.New("could not create room")
	ErrorGenerateCode       = errors.New("could not generate unique code")
)

type RoomManagement interface {
	GetAllRooms() []*types.Room
	GetRoomById(id uuid.UUID) (*types.Room, error)

	JoinRoom(roomId uuid.UUID, user *types.User) (*types.Room, error)
	LeaveRoom(roomId uuid.UUID, user *types.User) (*types.Room, error)

	CreateRoom() (*types.Room, error)
	DeleteRoom(id uuid.UUID) error
}

type RoomService struct {
	store      store.RoomStore
	timers     map[uuid.UUID]*time.Timer
	timerMutex sync.RWMutex
}

func (r *RoomService) GetAllRooms() []*types.Room {
	return r.store.GetAllRooms()
}

type RoomOptions func(r *RoomService)

var _ RoomManagement = (*RoomService)(nil)

func NewRoomService(store store.RoomStore) *RoomService {
	return &RoomService{
		store:  store,
		timers: make(map[uuid.UUID]*time.Timer),
	}
}

func (r *RoomService) GetRoomById(id uuid.UUID) (*types.Room, error) {
	room, err := r.store.GetRoomById(id)
	if err != nil {
		return nil, ErrorRoomNotFound
	}
	return room, nil
}

func (r *RoomService) JoinRoom(roomId uuid.UUID, user *types.User) (*types.Room, error) {
	room, err := r.store.GetRoomById(roomId)
	if err != nil {
		return nil, ErrorRoomNotFound
	}

	room.AddUser(user)
	updatedRoom, err := r.store.UpdateRoom(room.ID, room)
	if err != nil {
		return nil, ErrorCouldNotJoinRoom
	}
	user.SetRoomId(updatedRoom.ID)

	r.stopDeletionTimer(updatedRoom.ID)

	return updatedRoom, nil
}

func (r *RoomService) LeaveRoom(roomId uuid.UUID, user *types.User) (*types.Room, error) {
	room, err := r.store.GetRoomById(roomId)
	if err != nil {
		return nil, ErrorRoomNotFound
	}

	room.RemoveUser(user)
	user.SetRoomId(uuid.Nil)

	if room.IsEmpty() {
		r.startDeletionTimer(room)
	}

	updatedRoom, err := r.store.UpdateRoom(room.ID, room)
	if err != nil {
		return nil, ErrorCouldNotLeaveRoom
	}

	return updatedRoom, nil
}

func (r *RoomService) startDeletionTimer(room *types.Room) {
	r.timerMutex.Lock()
	defer r.timerMutex.Unlock()

	timer := time.AfterFunc(60*time.Second, func() {
		log.Println("Deleting room", room.ID)
		if room.IsEmpty() {
			_ = r.DeleteRoom(room.ID)
		}
	})

	r.timers[room.ID] = timer
}

func (r *RoomService) stopDeletionTimer(roomId uuid.UUID) {
	r.timerMutex.Lock()
	defer r.timerMutex.Unlock()

	timer, ok := r.timers[roomId]
	if ok {
		timer.Stop()
		delete(r.timers, roomId)
	}
}

func (r *RoomService) CreateRoom() (*types.Room, error) {
	room := types.CreateRoom()
	err := r.store.CreateRoom(room)
	if err != nil {
		return nil, ErrorCouldNotCreateRoom
	}

	return room, nil
}

func (r *RoomService) DeleteRoom(id uuid.UUID) error {
	roomToDelete, err := r.store.GetRoomById(id)
	if err != nil {
		return ErrorRoomNotFound
	}

	if !roomToDelete.IsEmpty() {
		return ErrorRoomNotEmpty
	}

	err = r.store.DeleteRoom(id)
	return nil
}
