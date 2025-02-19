package sql

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/peter-stratton/gofr/pkg/gofr/config"
	"github.com/peter-stratton/gofr/pkg/gofr/logging"
	"github.com/peter-stratton/gofr/pkg/gofr/testutil"
)

func TestNewSQL_ErrorCase(t *testing.T) {
	ctrl := gomock.NewController(t)

	expectedLog := fmt.Sprintf("could not register sql dialect '%s' for traces, error: %s", "mysql",
		"sql: unknown driver \"mysql\" (forgotten import?)")

	mockConfig := config.NewMockConfig(map[string]string{
		"DB_DIALECT":  "mysql",
		"DB_HOST":     "localhost",
		"DB_USER":     "testuser",
		"DB_PASSWORD": "testpassword",
		"DB_PORT":     "3306",
		"DB_NAME":     "testdb",
	})

	testLogs := testutil.StderrOutputForFunc(func() {
		mockLogger := logging.NewMockLogger(logging.ERROR)
		mockMetrics := NewMockMetrics(ctrl)

		NewSQL(mockConfig, mockLogger, mockMetrics)
	})

	if !strings.Contains(testLogs, expectedLog) {
		t.Errorf("TestNewSQL_ErrorCase Failed! Expcted error log doesn't match actual.")
	}
}

func TestNewSQL_InvalidDialect(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockConfig := config.NewMockConfig(map[string]string{
		"DB_DIALECT": "abc",
		"DB_HOST":    "localhost",
	})

	testLogs := testutil.StderrOutputForFunc(func() {
		mockLogger := logging.NewMockLogger(logging.ERROR)
		mockMetrics := NewMockMetrics(ctrl)

		NewSQL(mockConfig, mockLogger, mockMetrics)
	})

	if !strings.Contains(testLogs, errUnsupportedDialect.Error()) {
		t.Errorf("TestNewSQL_ErrorCase Failed! Expcted error log doesn't match actual.")
	}
}

func TestNewSQL_InvalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockConfig := config.NewMockConfig(map[string]string{
		"DB_DIALECT": "",
	})

	mockLogger := logging.NewMockLogger(logging.ERROR)
	mockMetrics := NewMockMetrics(ctrl)

	db := NewSQL(mockConfig, mockLogger, mockMetrics)

	assert.Nil(t, db, "TestNewSQL_InvalidConfig. expected db to be nil.")
}

func TestSQL_GetDBConfig(t *testing.T) {
	mockConfig := config.NewMockConfig(map[string]string{
		"DB_DIALECT":  "mysql",
		"DB_HOST":     "host",
		"DB_USER":     "user",
		"DB_PASSWORD": "password",
		"DB_PORT":     "3201",
		"DB_NAME":     "test",
	})

	expectedComfigs := &DBConfig{
		Dialect:  "mysql",
		HostName: "host",
		User:     "user",
		Password: "password",
		Port:     "3201",
		Database: "test",
	}

	configs := getDBConfig(mockConfig)

	assert.Equal(t, expectedComfigs, configs)
}

func TestSQL_getDBConnectionString(t *testing.T) {
	testCases := []struct {
		desc    string
		configs *DBConfig
		expOut  string
		expErr  error
	}{
		{
			desc: "mysql dialect",
			configs: &DBConfig{
				Dialect:  "mysql",
				HostName: "host",
				User:     "user",
				Password: "password",
				Port:     "3201",
				Database: "test",
			},
			expOut: "user:password@tcp(host:3201)/test?charset=utf8&parseTime=True&loc=Local&interpolateParams=true",
		},
		{
			desc: "postgresql dialect",
			configs: &DBConfig{
				Dialect:  "postgres",
				HostName: "host",
				User:     "user",
				Password: "password",
				Port:     "3201",
				Database: "test",
			},
			expOut: "host=host port=3201 user=user password=password dbname=test sslmode=disable",
		},
		{
			desc: "sqlite dialect",
			configs: &DBConfig{
				Dialect:  "sqlite",
				Database: "test.db",
			},
			expOut: "file:test.db",
		},
		{
			desc:    "unsupported dialect",
			configs: &DBConfig{Dialect: "mssql"},
			expOut:  "",
			expErr:  errUnsupportedDialect,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			connString, err := getDBConnectionString(tc.configs)

			assert.Equal(t, tc.expOut, connString)
			assert.Equal(t, tc.expErr, err)
		})
	}
}

func Test_NewSQLMock(t *testing.T) {
	db, mock, mockMetric := NewSQLMocks(t)

	assert.NotNil(t, db)
	assert.NotNil(t, mock)
	assert.NotNil(t, mockMetric)
}

func Test_NewSQLMockWithConfig(t *testing.T) {
	dbConfig := DBConfig{Dialect: "dialect", HostName: "hostname", User: "user", Password: "password", Port: "port", Database: "database"}
	db, mock, mockMetric := NewSQLMocksWithConfig(t, &dbConfig)

	assert.NotNil(t, db)
	assert.Equal(t, db.config, &dbConfig)
	assert.NotNil(t, mock)
	assert.NotNil(t, mockMetric)
}

func Test_SQLRetryConnectionInfoLog(t *testing.T) {
	logs := testutil.StdoutOutputForFunc(func() {
		ctrl := gomock.NewController(t)

		mockMetrics := NewMockMetrics(ctrl)
		mockConfig := config.NewMockConfig(map[string]string{
			"DB_DIALECT":  "postgres",
			"DB_HOST":     "host",
			"DB_USER":     "user",
			"DB_PASSWORD": "password",
			"DB_PORT":     "3201",
			"DB_NAME":     "test",
		})

		mockLogger := logging.NewMockLogger(logging.DEBUG)

		mockMetrics.EXPECT().SetGauge("app_sql_open_connections", float64(0))
		mockMetrics.EXPECT().SetGauge("app_sql_inUse_connections", float64(0))

		_ = NewSQL(mockConfig, mockLogger, mockMetrics)

		time.Sleep(2 * time.Second)
	})

	assert.Contains(t, logs, "retrying SQL database connection")
}
