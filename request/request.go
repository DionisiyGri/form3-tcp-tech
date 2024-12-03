package request

import (
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/interview-simulator/model"
)

// Handle processes a single request string, returning an appropriate response
func Handle(request string) string {
	parts := strings.Split(request, "|")
	if len(parts) != 2 || parts[0] != "PAYMENT" {
		return model.ResponseRejectedInvalidRequest
	}

	amount, err := strconv.Atoi(parts[1])
	if err != nil {
		return model.ResponseRejectedInvalidAmount
	}

	if amount > 100 {
		processingTime := amount
		if amount > 10000 {
			processingTime = 10000
		}
		time.Sleep(time.Duration(processingTime) * time.Millisecond)
	}
	return model.ResponseAccepted
}
