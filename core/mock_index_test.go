package core

type MockIndex struct {
	upsertFunc func([]PointStruct) error
}

func (m *MockIndex) Upsert(points []PointStruct) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(points)
	}
	return nil
}

func (m *MockIndex) Search(query []float32, filter *Filter, topK int) ([]ScoredPoint, error) {
	return nil, nil
}

func (m *MockIndex) Delete(id string) error {
	return nil
}

func (m *MockIndex) Count() int {
	return 0
}

func (m *MockIndex) GetIDsByFilter(filter *Filter) []string {
	return nil
}

func (m *MockIndex) GetPointsByFilter(filter *Filter) []PointStruct {
	return nil
}

func (m *MockIndex) DeleteByFilter(filter *Filter) ([]string, error) {
	return nil, nil
}
