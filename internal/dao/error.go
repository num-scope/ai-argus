package dao

import (
	"errors"
	"strings"

	"github.com/xtj/ai-argus/internal/common"
	"gorm.io/gorm"
)

func normalizeCreateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
		return common.ErrAlreadyExists
	}
	return err
}
