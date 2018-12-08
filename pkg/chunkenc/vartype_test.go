package chunkenc

import (
	"fmt"
	"github.com/nuclio/zap"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type testVarEncoderSuite struct {
	suite.Suite
}

func (suite *testVarEncoderSuite) TestStringEnc() {

	logger, err := nucliozap.NewNuclioZapTest("test")
	suite.Require().Nil(err)

	chunk := NewVarChunk(logger)
	appender, err := chunk.Appender()
	suite.Require().Nil(err)

	list := []string{"abc", "", "123456"}
	t0 := time.Now().UnixNano() / 1000

	for i, s := range list {
		t := t0 + int64(i*1000)
		appender.Append(t, s)
		b := chunk.Bytes()
		fmt.Println(t, s, len(b))
	}

	iterChunk, err := FromData(logger, EncVar, chunk.Bytes(), 0)
	suite.Require().Nil(err)

	iter := iterChunk.Iterator()
	i := 0
	for iter.Next() {
		t, v := iter.AtString()
		suite.Require().Equal(t, t0+int64(i*1000))
		suite.Require().Equal(v, list[i])
		fmt.Println("t, v: ", t, v)
		i++
	}

	suite.Require().Nil(iter.Err())
	suite.Require().Equal(i, len(list))
}

func TestVarEncoderSuite(t *testing.T) {
	suite.Run(t, new(testVarEncoderSuite))
}
