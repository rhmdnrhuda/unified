package usecase

import (
	"context"
	"github.com/rhmdnrhuda/unified/core/entity"
	"github.com/rhmdnrhuda/unified/pkg/logger"
)

type TalentUseCase struct {
	talentRepo TalentRepository
	log        logger.Interface
}

func NewTalentUseCase(talentRepo TalentRepository, l logger.Interface) *TalentUseCase {
	return &TalentUseCase{
		talentRepo: talentRepo,
		log:        l,
	}
}

func (t *TalentUseCase) Create(ctx context.Context, req entity.TalentRequest) error {

	return t.talentRepo.Create(ctx, &entity.Talent{
		ID:         req.ID,
		Name:       req.Name,
		University: req.University,
		Major:      req.Major,
		Status:     req.Status,
	})
}

func (t *TalentUseCase) GetTalent(ctx context.Context, university, major string) (entity.Talent, error) {
	return t.talentRepo.FindTalentByUniversityAndMajor(ctx, university, major)
}

func (t *TalentUseCase) Update(ctx context.Context, req entity.TalentRequest) error {
	return t.talentRepo.Update(ctx, &entity.Talent{
		ID:         req.ID,
		Name:       req.Name,
		University: req.University,
		Major:      req.Major,
		Status:     req.Status,
	})
}
