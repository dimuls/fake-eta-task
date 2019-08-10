package wapi

import "github.com/dimuls/fake-eta-task/entity"

type DummyCache struct{}

func (c DummyCache) Cars(_ *entity.Coordinate, _ int) (
	[]*entity.Coordinate, bool, error) {
	return nil, false, nil
}

func (c DummyCache) Predicts(_ *entity.Coordinate, _ []*entity.Coordinate) (
	[]int, bool, error) {
	return nil, false, nil
}
