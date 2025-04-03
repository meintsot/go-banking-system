package adapters

import (
	"errors"
	"sync"
)

// Entity is an interface that all models must implement to be used with the generic repository
type Entity interface {
	GetID() string
}

// GenericRepository is a generic repository implementation using Go's generics
type GenericRepository[T Entity] struct {
	items map[string]T
	mu    sync.RWMutex
}

// NewGenericRepository creates a new generic repository
func NewGenericRepository[T Entity]() *GenericRepository[T] {
	return &GenericRepository[T]{
		items: make(map[string]T),
	}
}

// Create adds a new entity to the repository
func (r *GenericRepository[T]) Create(entity T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := entity.GetID()
	if _, exists := r.items[id]; exists {
		return errors.New("entity with this ID already exists")
	}

	r.items[id] = entity
	return nil
}

// GetByID retrieves an entity by its ID
func (r *GenericRepository[T]) GetByID(id string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, exists := r.items[id]
	if !exists {
		var zero T
		return zero, errors.New("entity not found")
	}

	return entity, nil
}

// GetAll returns all entities in the repository
func (r *GenericRepository[T]) GetAll() ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := make([]T, 0, len(r.items))
	for _, entity := range r.items {
		entities = append(entities, entity)
	}

	return entities, nil
}

// Update updates an existing entity
func (r *GenericRepository[T]) Update(entity T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := entity.GetID()
	if _, exists := r.items[id]; !exists {
		return errors.New("entity not found")
	}

	r.items[id] = entity
	return nil
}

// Delete removes an entity by its ID
func (r *GenericRepository[T]) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return errors.New("entity not found")
	}

	delete(r.items, id)
	return nil
}

// Count returns the number of entities in the repository
func (r *GenericRepository[T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.items)
}

// Filter returns entities that match the predicate function
func (r *GenericRepository[T]) Filter(predicate func(T) bool) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []T
	for _, entity := range r.items {
		if predicate(entity) {
			result = append(result, entity)
		}
	}

	return result
}
