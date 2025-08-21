package permissionservice

type Permission struct {
	r Repository
}

func New(r Repository) *Permission {
	return &Permission{
		r: r,
	}
}

type Repository interface {
	UserRepository
}
