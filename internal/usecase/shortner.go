package usecase

type ShrotnerUseCase struct{
	repo URLRepository
}

func NewSrotner(r URLRepository) *ShrotnerUseCase{
	return &ShrotnerUseCase{ repo: r}
}