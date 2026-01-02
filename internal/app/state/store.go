package state

// Store is a placeholder for runtime persistence. Currently in-memory only.
type Store struct{}

func NewStore() *Store {
	return &Store{}
}
