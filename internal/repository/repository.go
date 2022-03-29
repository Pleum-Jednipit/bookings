package repository

import (
	"time"

	"github.com/Pleum-Jednipit/bookings/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int,error)
	InsertRoomRestriction(res models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomId(start,end time.Time, roomId int) (bool,error)
	SearchAvailabilityForAllRooms(start,end time.Time) ([]models.Room,error)
	GetRoomById(id int) (models.Room,error)
}