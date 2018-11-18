package pquerier

import (
	"github.com/v3io/v3io-tsdb/pkg/aggregate"
	"github.com/v3io/v3io-tsdb/pkg/utils"
	"math"
)

func NewDataFrameColumnSeries(indexColumn, dataColumn, countColumn Column, labels utils.Labels, hash uint64) *DataFrameColumnSeries {
	s := &DataFrameColumnSeries{dataColumn: dataColumn, indexColumn: indexColumn, CountColumn: countColumn, labels: labels, key: hash}
	s.iter = &DataFrameColumnSeriesIterator{indexColumn: indexColumn, dataColumn: dataColumn, countColumn: countColumn, currentIndex: -1}
	return s
}

type DataFrameColumnSeries struct {
	dataColumn  Column
	indexColumn Column
	CountColumn Column // Count Column is needed to filter out empty buckets
	labels      utils.Labels
	key         uint64
	iter        SeriesIterator
}

func (s *DataFrameColumnSeries) Labels() utils.Labels {
	s.labels = append(s.labels, utils.LabelsFromStrings(aggregate.AggregateLabel, s.dataColumn.GetColumnSpec().function.String())...)
	return s.labels
}
func (s *DataFrameColumnSeries) Iterator() SeriesIterator { return s.iter }
func (s *DataFrameColumnSeries) GetKey() uint64           { return s.key }

type DataFrameColumnSeriesIterator struct {
	dataColumn  Column
	indexColumn Column
	countColumn Column

	currentIndex int
	err          error
}

func (it *DataFrameColumnSeriesIterator) Seek(seekT int64) bool {
	t, _ := it.At()
	if t >= seekT {
		return true
	}

	for it.Next() {
		t, _ := it.At()
		if t >= seekT {
			return true
		}
	}

	return false
}

func (it *DataFrameColumnSeriesIterator) At() (int64, float64) {
	t, err := it.indexColumn.TimeAt(it.currentIndex)
	if err != nil {
		it.err = err
	}
	v, err := it.dataColumn.FloatAt(it.currentIndex)
	if err != nil {
		it.err = err
	}
	return t, v
}

func (it *DataFrameColumnSeriesIterator) Next() bool {
	if it.err != nil {
		return false
	}
	it.currentIndex = it.getNextValidCell(it.currentIndex)

	// It is enough to only check one of the columns since we assume they are both the same size
	return it.currentIndex < it.indexColumn.Len()
}

func (it *DataFrameColumnSeriesIterator) Err() error { return it.err }

func (it *DataFrameColumnSeriesIterator) getNextValidCell(from int) (nextIndex int) {
	for nextIndex = from + 1; nextIndex < it.dataColumn.Len() && !it.doesCellHasData(nextIndex); nextIndex++ {
	}
	return
}

func (it *DataFrameColumnSeriesIterator) doesCellHasData(cell int) bool {
	// In case we don't have a count column (for example while down sampling) check if there is a real value at `cell`
	if it.countColumn == nil {
		f, err := it.dataColumn.FloatAt(cell)
		if err != nil {
			it.err = err
			return false
		}
		return !math.IsNaN(f)
	}
	val, err := it.countColumn.FloatAt(cell)
	if err != nil {
		it.err = err
		return false
	}
	return val > 0
}
