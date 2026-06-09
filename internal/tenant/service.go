package tenant

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetByID(id string) (*Tenant, error) {
    // TODO: implement tenant lookup
    return nil, nil
}
