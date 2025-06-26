package main

import (
    "database/sql"
    "fmt"
    "os"
    "sync"

    _ "github.com/lib/pq"
)

// Bye represents a requested bye for a given round.
type Bye struct {
    Round  int
    Points float64
}

// ByeStore abstracts persistence for bye requests.
type ByeStore interface {
    // GetByes returns all byes for a player in an event. EventID may be ignored
    // by implementations that do not store event information.
    GetByes(eventID int64, uscfID int) ([]Bye, error)
    // AddBye inserts or updates a bye request.
    AddBye(eventID int64, uscfID int, b Bye) error
}

// --- In-memory store (fallback / tests) ------------------------------------

type memByeStore struct {
    mu   sync.Mutex
    data map[int64]map[int][]Bye // eventID -> uscfID -> []Bye
}

func newMemByeStore() *memByeStore {
    return &memByeStore{data: map[int64]map[int][]Bye{}}
}

func (m *memByeStore) GetByes(eventID int64, id int) ([]Bye, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    return append([]Bye(nil), m.data[eventID][id]...), nil
}

func (m *memByeStore) AddBye(eventID int64, id int, b Bye) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.data[eventID]; !ok {
        m.data[eventID] = map[int][]Bye{}
    }
    // upsert
    byes := m.data[eventID][id]
    replaced := false
    for i := range byes {
        if byes[i].Round == b.Round {
            byes[i] = b
            replaced = true
            break
        }
    }
    if !replaced {
        byes = append(byes, b)
    }
    m.data[eventID][id] = byes
    return nil
}

// --- Postgres store ---------------------------------------------------------

type pgByeStore struct {
    db *sql.DB
}

func newPgByeStore(dsn string) (*pgByeStore, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    // verify connection quickly
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return &pgByeStore{db: db}, nil
}

func (p *pgByeStore) GetByes(_ int64, id int) ([]Bye, error) {
    rows, err := p.db.Query(`SELECT round, points FROM byes WHERE player_id=$1`, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var res []Bye
    for rows.Next() {
        var b Bye
        if err := rows.Scan(&b.Round, &b.Points); err != nil {
            return nil, err
        }
        res = append(res, b)
    }
    return res, nil
}

func (p *pgByeStore) AddBye(_ int64, id int, b Bye) error {
    _, err := p.db.Exec(`INSERT INTO byes (player_id, round, points) VALUES ($1,$2,$3)
            ON CONFLICT (player_id, round) DO UPDATE SET points=EXCLUDED.points`, id, b.Round, b.Points)
    return err
}

// --- Global store init ------------------------------------------------------

var byeStore ByeStore = newMemByeStore()

func init() {
    if dsn, ok := os.LookupEnv("DATABASE_URL"); ok && dsn != "" {
        if pg, err := newPgByeStore(dsn); err == nil {
            byeStore = pg
        } else {
            fmt.Printf("byestore: falling back to memory store, DB error: %v\n", err)
        }
    }
}
