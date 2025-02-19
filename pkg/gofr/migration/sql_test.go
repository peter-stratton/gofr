package migration

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/peter-stratton/gofr/pkg/gofr/container"
)

func TestNewMysql(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := container.NewMockDB(ctrl)

	sqlDB := newMysql(mockDB)

	if sqlDB.db != mockDB {
		t.Errorf("newMysql should wrap the provided db, got: %v", sqlDB.db)
	}
}

func TestQuery(t *testing.T) {
	t.Run("successful query", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)
		expectedRows := &sql.Rows{}

		mockDB.EXPECT().Query("SELECT * FROM users", []interface{}{}).Return(expectedRows, nil)
		sqlDB := newMysql(mockDB)

		rows, err := sqlDB.Query("SELECT * FROM users", []interface{}{})
		if rows.Err() != nil {
			t.Errorf("unexpected row error: %v", rows.Err())
		}

		if err != nil {
			t.Errorf("Query should return no error, got: %v", err)
		}

		if rows != expectedRows {
			t.Errorf("Query should return the expected rows, got: %v", rows)
		}
	})

	t.Run("query error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)
		expectedErr := sql.ErrNoRows

		mockDB.EXPECT().Query("SELECT * FROM unknown_table", []interface{}{}).Return(nil, expectedErr)
		sqlDB := newMysql(mockDB)

		rows, err := sqlDB.Query("SELECT * FROM unknown_table", []interface{}{})
		if rows != nil {
			t.Errorf("unexpected rows error: %v", rows.Err())
		}

		if err == nil {
			t.Errorf("Query should return an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Query should return the expected error, got: %v", err)
		}
	})
}

func TestQueryRow(t *testing.T) {
	t.Run("successful query row", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)
		expectedRow := &sql.Row{}

		mockDB.EXPECT().QueryRow("SELECT * FROM users WHERE id = ?", 1).Return(expectedRow)
		sqlDB := newMysql(mockDB)

		row := sqlDB.QueryRow("SELECT * FROM users WHERE id = ?", 1)

		if row != expectedRow {
			t.Errorf("QueryRow should return the expected row, got: %v", row)
		}
	})
}

func TestQueryRowContext(t *testing.T) {
	ctx := context.Background()

	t.Run("successful query row context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)
		expectedRow := &sql.Row{}
		mockDB.EXPECT().QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", 1).Return(expectedRow)
		sqlDB := newMysql(mockDB)

		row := sqlDB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", 1)

		if row != expectedRow {
			t.Errorf("QueryRowContext should return the expected row,  got: %v", row)
		}
	})
}

func TestExec(t *testing.T) {
	t.Run("successful exec", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)
		expectedResult := sqlmock.NewResult(10, 1)

		mockDB.EXPECT().Exec("DELETE FROM users WHERE id = ?", 1).Return(expectedResult, nil)
		sqlDB := newMysql(mockDB)

		result, err := sqlDB.Exec("DELETE FROM users WHERE id = ?", 1)

		if err != nil {
			t.Errorf("Exec should return no error, got: %v", err)
		}

		if !reflect.DeepEqual(result, expectedResult) {
			t.Errorf("Exec should return the expected result, got: %v", result)
		}
	})

	t.Run("exec error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)

		expectedErr := sql.ErrNoRows
		mockDB.EXPECT().Exec("UPDATE unknown_table SET name = ?", "John").Return(nil, expectedErr)
		sqlDB := newMysql(mockDB)

		_, err := sqlDB.Exec("UPDATE unknown_table SET name = ?", "John")

		if err == nil {
			t.Errorf("Exec should return an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Exec should return the expected error, got: %v", err)
		}
	})
}

func TestExecContext(t *testing.T) {
	ctx := context.Background()

	t.Run("successful exec context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDB := container.NewMockDB(ctrl)

		expectedResult := sqlmock.NewResult(10, 1)
		mockDB.EXPECT().ExecContext(ctx, "DELETE FROM users WHERE id = ?", 1).Return(expectedResult, nil)
		sqlDB := newMysql(mockDB)

		result, err := sqlDB.ExecContext(ctx, "DELETE FROM users WHERE id = ?", 1)

		if err != nil {
			t.Errorf("ExecContext should return no error, got: %v", err)
		}

		if !reflect.DeepEqual(result, expectedResult) {
			t.Errorf("ExecContext should return the expected result, got: %v", result)
		}
	})
}

func TestCheckAndCreateMigrationTableSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := container.NewMockDB(ctrl)
	mockMigrator := NewMockMigrator(ctrl)
	mockContainer, mocks := container.NewMockContainer(t)

	mockMigrator.EXPECT().checkAndCreateMigrationTable(mockContainer)
	mocks.SQL.EXPECT().Exec(createSQLGoFrMigrationsTable).Return(nil, nil)

	migrator := sqlMigrator{
		db:       mockDB,
		Migrator: mockMigrator,
	}

	err := migrator.checkAndCreateMigrationTable(mockContainer)

	if err != nil {
		t.Errorf("checkAndCreateMigrationTable should return no error, got: %v", err)
	}
}

func TestCheckAndCreateMigrationTableExecError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := container.NewMockDB(ctrl)
	mockMigrator := NewMockMigrator(ctrl)
	mockContainer, mocks := container.NewMockContainer(t)
	expectedErr := sql.ErrNoRows

	mocks.SQL.EXPECT().Exec(createSQLGoFrMigrationsTable).Return(nil, expectedErr)

	migrator := sqlMigrator{
		db:       mockDB,
		Migrator: mockMigrator,
	}

	err := migrator.checkAndCreateMigrationTable(mockContainer)

	if err == nil {
		t.Errorf("checkAndCreateMigrationTable should return an error")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("checkAndCreateMigrationTable should return the expected error, got: %v", err)
	}
}

func TestBeginTransactionSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := container.NewMockDB(ctrl)
	mockMigrator := NewMockMigrator(ctrl)
	mockContainer, mocks := container.NewMockContainer(t)
	expectedMigrationData := migrationData{}

	mocks.SQL.EXPECT().Begin()
	mockMigrator.EXPECT().beginTransaction(mockContainer)

	migrator := sqlMigrator{
		db:       mockDB,
		Migrator: mockMigrator,
	}
	data := migrator.beginTransaction(mockContainer)

	if data != expectedMigrationData {
		t.Errorf("beginTransaction should return data from Migrator, got: %v", data)
	}
}

var (
	errBeginTx = errors.New("failed to begin transaction")
)

func TestBeginTransactionDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := container.NewMockDB(ctrl)
	mockMigrator := NewMockMigrator(ctrl)
	mockContainer, mocks := container.NewMockContainer(t)

	mocks.SQL.EXPECT().Begin().Return(nil, errBeginTx)

	migrator := sqlMigrator{
		db:       mockDB,
		Migrator: mockMigrator,
	}
	data := migrator.beginTransaction(mockContainer)

	if data.SQLTx != nil {
		t.Errorf("beginTransaction should not return a transaction on DB error")
	}
}

func TestRollbackNoTransaction(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)

	migrator := sqlMigrator{}
	migrator.rollback(mockContainer, migrationData{})
}
