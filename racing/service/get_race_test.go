package service

import (
	"context"
	"database/sql"
	"testing"

	dbmocks "racing/db/mocks"
	racing "racing/proto/racing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRace_Success(t *testing.T) {
	m := new(dbmocks.RacesRepo)

	// Repo returns *racing.Race (NOT *GetRaceResponse)
	repoRace := &racing.Race{Id: 75, Name: "Test"}
	m.On("Get", mock.Anything, int64(75)).Return(repoRace, nil).Once()

	svc := &racingService{racesRepo: m}

	got, err := svc.GetRace(context.Background(), &racing.GetRaceRequest{Id: 75})
	assert.NoError(t, err)
	// Service returns *GetRaceResponse
	assert.NotNil(t, got)
	assert.Equal(t, int64(75), got.GetRace().GetId())
	assert.Equal(t, "Test", got.GetRace().GetName())

	m.AssertExpectations(t)
}

func TestGetRace_NotFound(t *testing.T) {
	m := new(dbmocks.RacesRepo)
	m.On("Get", mock.Anything, int64(999)).
		Return((*racing.Race)(nil), sql.ErrNoRows).Once()

	svc := &racingService{racesRepo: m}
	got, err := svc.GetRace(context.Background(), &racing.GetRaceRequest{Id: 999})
	assert.Nil(t, got)
	assert.Error(t, err)
}

func TestGetRace_InvalidID(t *testing.T) {
	m := new(dbmocks.RacesRepo)
	svc := &racingService{racesRepo: m}

	got, err := svc.GetRace(context.Background(), &racing.GetRaceRequest{Id: 0})
	assert.Nil(t, got)
	assert.Error(t, err)
}
