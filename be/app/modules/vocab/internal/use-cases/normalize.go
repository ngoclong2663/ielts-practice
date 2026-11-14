package usecases

import (
	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/dto"
)

func (uc *VocabUseCases) Normalize(req dto.NormalizeRequestDto) dto.NormalizeResponseDto {
	entries, rejected := domain.Normalize(req.Text)
	return dto.NormalizeResponseDto{Entries: entries, Rejected: rejected, Max: domain.MaxEntriesPerJob}
}
