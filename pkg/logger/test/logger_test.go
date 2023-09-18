package test

import (
	"github.com/246859/codis/pkg/logger"
	"testing"
)

func TestTextLogger(t *testing.T) {
	err := logger.Setup(logger.Config{
		Level:    logger.LevelInfo,
		Format:   logger.TextFormat,
		InfoLog:  "testdata/info.log",
		ErrorLog: "testdata/error.log",
	})

	if err != nil {
		t.Error(err)
	}
	defer logger.Close()

	// will be output to info.log
	logger.Info("this is a test log")
	logger.Info("this is a test log too")
	logger.Info("this is an another test log too")
	logger.Warn("this is a warning test log")

	logger.Debug("this is a debug test log")

	// will be output to error.log
	logger.Error("this is a error test log")
}

func TestJsonLogger(t *testing.T) {
	err := logger.Setup(logger.Config{
		Level:    logger.LevelInfo,
		Format:   logger.JsonFormat,
		InfoLog:  "testdata/info.json.log",
		ErrorLog: "testdata/error.json.log",
	})

	if err != nil {
		t.Error(err)
	}
	defer logger.Close()

	// will be output to info.log
	logger.Info("this is info test log")
	logger.Warn("this is a warning test log")

	// will not output
	logger.Debug("this is a debug test log")

	// will be output to error.log
	logger.Error("this is a error test log")
}
