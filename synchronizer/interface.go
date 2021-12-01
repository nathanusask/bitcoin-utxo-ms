package synchronizer

import "context"

//go:generate mockgen -source=./interface.go -destination=mocks/interface_mock.go -package=synchronizer
type Interface interface {
	Start(ctx context.Context)
}
