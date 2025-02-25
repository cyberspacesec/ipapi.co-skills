// errors.go
package ipapi

import (
	"errors"
	"fmt"
)

func (c *Client) handleError(err error) error {
	if c.errorHandler != nil {
		return c.errorHandler(err)
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) { // 现在可以正确解包
		switch apiErr.Reason {
		case "RateLimited":
			return fmt.Errorf("%w: %s", ErrRateLimited, apiErr.Message)
		case "Reserved IP Address":
			return fmt.Errorf("%w: %s", ErrReservedIP, apiErr.IP)
		case "Invalid IP Address":
			return fmt.Errorf("%w: %s", ErrInvalidIP, apiErr.IP)
		}
	}

	return err
}

func IsRetryableError(err error) bool {
	return errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrServerError) ||
		errors.Is(err, ErrNotFound)
}

func WrapError(op string, err error) error {
	return fmt.Errorf("%s failed: %w", op, err)
}
