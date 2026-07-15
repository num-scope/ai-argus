package service

import (
	"fmt"

	"github.com/xtj/ai-argus/internal/common"
)

func invalidRequest(message string) error {
	return fmt.Errorf("%w: %s", common.ErrInvalidRequestBody, message)
}
