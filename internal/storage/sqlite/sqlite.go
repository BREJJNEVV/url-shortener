package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
	// _ "github.com/mattn/go-sqlite3" // init sqlite3 driver
)

type Storage struct {
	db *sql.DB // Коннект до базы
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New" // храним имя функции для ошибки

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	_, err = stmt.Exec() // вызов функции prepare, нам не нужно сохранять первый параметр
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}
	return id, nil
}
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement:%w", op, err)
	}
	var resURL string                        // Подготавливаем URL, в который мы положим возвращаемый результат (GET - же)
	err = stmt.QueryRow(alias).Scan(&resURL) // Для запросов, возвращающих данные:
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) FindURL(alias string) (bool, error) {
	const op = "storage.sqlite.FindURL"
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return false, fmt.Errorf("%s: prepare statement:%w", op, err)
	}
	var resURL string // Подготавливаем URL, в который мы положим возвращаемый результат (GET - же)

	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrURLNotFound
		}
		return false, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return true, nil
}

//
//
//
//

// func (s *Storage) DeleteURL(alias string) error {
// 	const op = "storage.sqlite.DeliteURL"
// 	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
// 	if err != nil {
// 		return fmt.Errorf("%s: prepare statement:%w", op, err)
// 	}

// 	res, err := stmt.Exec(alias)
// 	if err != nil {
// 		return fmt.Errorf("%s: execute statement: %w", op, err)
// 	}
// 	rowsDel, err := res.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("%s: execute statement: %w", op, err)
// 	}
// 	if rowsDel == 0 {
// 		return storage.ErrURLNotFound
// 	}

// 	return err
// }

// РЕАЛИЗАЦИЯ ОТ ДИПСИК

// func (s *Storage) DeleteURL(alias string) error {
// 	const op = "storage.sqlite.DeleteURL"

// 	// Подготавливаем SQL-запрос
// 	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
// 	if err != nil {
// 		return fmt.Errorf("%s: prepare statement: %w", op, err)
// 	}
// 	defer stmt.Close()

// 	// Выполняем запрос с переданным алиасом
// 	res, err := stmt.Exec(alias) // Для запросов, изменяющих данные: NSERT/UPDATE/DELETE
// 	// И других запросов, которые не возвращают строк
// 	if err != nil {
// 		return fmt.Errorf("%s: execute statement: %w", op, err)
// 	}

// 	// Проверяем количество удалённых строк
// 	rowsAffected, err := res.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("%s: get affected rows: %w", op, err)
// 	}

// 	// Если ни одна строка не была удалена, возвращаем ошибку
// 	if rowsAffected == 0 {
// 		return storage.ErrURLNotFound
// 	}

// 	return nil
// }

func (s *Storage) AliasNotExists(alias string) bool {
	res, _ := s.GetURL(alias)
	return res == ""
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
