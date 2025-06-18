package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет новую посылку в таблицу parcel
func (s ParcelStore) Add(p Parcel) (int, error) {
	query := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

// Get получает одну посылку по номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	query := `SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`
	row := s.db.QueryRow(query, number)

	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

// GetByClient возвращает все посылки указанного клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := `SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

// SetStatus обновляет статус посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	query := `UPDATE parcel SET status = ? WHERE number = ?`
	_, err := s.db.Exec(query, status, number)
	return err
}

// SetAddress обновляет адрес доставки только если посылка зарегистрирована
func (s ParcelStore) SetAddress(number int, address string) error {
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("нельзя изменить адрес: посылка уже отправлена или доставлена")
	}
	query := `UPDATE parcel SET address = ? WHERE number = ?`
	_, err = s.db.Exec(query, address, number)
	return err
}

// Delete удаляет посылку, если её статус "зарегистрирована"
func (s ParcelStore) Delete(number int) error {
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("нельзя удалить посылку: она уже отправлена или доставлена")
	}
	query := `DELETE FROM parcel WHERE number = ?`
	_, err = s.db.Exec(query, number)
	return err
}
