package application

import (
	"context"
	"sync"

	"github.com/wyw14/cry-053/internal/domain"
)

type publicationPhase uint8

const (
	publicationIdle publicationPhase = iota
	publicationChecking
	publicationWriting
	publicationFinished
)

type PublicationCoordinator struct {
	gate  sync.Mutex
	phase publicationPhase
}

type PublicationPermit struct {
	owner    *PublicationCoordinator
	released bool
}

func NewPublicationCoordinator() *PublicationCoordinator {
	return &PublicationCoordinator{phase: publicationIdle}
}

func (c *PublicationCoordinator) Enter(ctx context.Context) (*PublicationPermit, error) {
	acquired := make(chan struct{}, 1)
	go func() {
		c.gate.Lock()
		acquired <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		go func() {
			<-acquired
			c.gate.Unlock()
		}()
		return nil, ctx.Err()
	case <-acquired:
		c.phase = publicationChecking
		return &PublicationPermit{owner: c}, nil
	}
}

func (p *PublicationPermit) BeginWrite() error {
	if p == nil || p.released || p.owner.phase != publicationChecking {
		return domain.ErrInvalidTransition
	}
	p.owner.phase = publicationWriting
	return nil
}

func (p *PublicationPermit) Close() {
	if p == nil || p.released {
		return
	}
	p.owner.phase = publicationFinished
	p.released = true
	p.owner.gate.Unlock()
}

func validateFreshPublication(bundle domain.Bundle) error {
	if bundle.State != domain.BundleApproved || bundle.HasErrors() {
		return domain.ErrInvalidTransition
	}
	return nil
}
