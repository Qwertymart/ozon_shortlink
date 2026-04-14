package usecase

type ShrotenerUseCase struct{
	repo URLRepository
}

func NewSrotener(r URLRepository) *ShrotnerUseCase{
	return &ShrotnerUseCase{ repo: r}
}