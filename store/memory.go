package store

type URLStore struct {
	Store map[string]string
}

type TenantStore struct {
	Tenants map[string]*URLStore
}

func NewURLStore() *URLStore {
	return &URLStore{
		Store: make(map[string]string),
	}
}

func (s *URLStore) Get(shortCode string) (string, bool) {
	original, ok := s.Store[shortCode]
	return original, ok
}

func (s *URLStore) Save(shortCode, originalURL string) {
	s.Store[shortCode] = originalURL
}

func NewTenantStore() *TenantStore {
	return &TenantStore{
		Tenants: make(map[string]*URLStore),
	}
}

func (ts *TenantStore) GetStore(apiKey string) *URLStore {
	store, exists := ts.Tenants[apiKey]
	if !exists {
		store = NewURLStore()
		ts.Tenants[apiKey] = store
	}
	return store
}
