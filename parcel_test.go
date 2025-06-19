package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, p.Client) // заменено require - assert
	assert.Equal(t, parcel.Status, p.Status)
	assert.Equal(t, parcel.Address, p.Address)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, p.Address) // assert вместо require

	err = store.Delete(id) // проверка ошибки
	require.NoError(t, err)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Меняем статус на "отправлена"
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, p.Status)

	err = store.Delete(id)
	require.Error(t, err)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{getTestParcel(), getTestParcel(), getTestParcel()}
	client := randRange.Intn(10_000_000)
	parcelMap := map[int]Parcel{}

	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	stored, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, stored, 3) // заменено с require.Len

	for _, p := range stored {
		expected := parcelMap[p.Number]
		assert.Equal(t, expected, p) // заменено с трёх require.Equal
	}

	for _, p := range stored {
		err := store.Delete(p.Number) // перемещено из цикла выше
		require.NoError(t, err)
	}
}
