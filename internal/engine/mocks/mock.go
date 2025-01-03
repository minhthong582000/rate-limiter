package mocks

import (
	_ "go.uber.org/mock/mockgen/model"
)

//go:generate mockgen -destination=mock_engine.go -package=mocks github.com/minhthong582000/rate-limiter/internal/engine Engine
