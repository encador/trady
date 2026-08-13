package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const ItemColumns = "select id, owner_id, title, description, image, location, listed from items "

func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("[ERROR] No DB File Specified")
	}
	if !Exists(path) {
		return nil, fmt.Errorf("[ERROR] DB File (%s) Does NOT Exists", path)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func Exists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
}

func Create(path string) error {
	if path == "" {
		return errors.New("[ERROR] No DB File Name Specified")
	}

	if Exists(path) {
		// return fmt.Errorf("[ERROR] DB File (%s) Already Exists", path)
		fmt.Printf("[LOG] DB File (%s) Already Exists\n", path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	err := os.WriteFile(path, nil, 0644)
	if err != nil {
		return err
	}

	db, err := Open(path)
	if err != nil {
		return err
	}
	defer db.Close()

	q := `
	drop table if exists users;
	create table users(
		id text primary key,
		security integer not null default 1,
		username text not null unique,
		password text not null
	);

	drop table if exists items;
	create table items(
		id text primary key,
		owner_id text not null references users(id) on delete cascade,
		title text not null,
		description text,
		image text,
		location text,
		listed boolean not null default 0
	);

	drop table if exists bids;
	create table bids(
		target_item_id text not null references items(id) on delete cascade,
		bid_item_id text not null references items(id) on delete cascade,
		bid_owner_id text not null references users(id) on delete cascade,
		check (target_item_id <> bid_item_id),
		primary key (target_item_id, bid_item_id)
	);

	create trigger remove_bid_when_item_delisted
	after update of listed on items
	when new.listed = false
	begin
		delete from bids
		where target_item_id = new.id or bid_item_id = new.id;
	end;
	`
	_, err = db.Exec(q)
	return err
}
