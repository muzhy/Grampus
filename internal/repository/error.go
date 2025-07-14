package repository

type RepositoryError string

func (e *RepositoryError) Error() string {
	return string(*e)
}
