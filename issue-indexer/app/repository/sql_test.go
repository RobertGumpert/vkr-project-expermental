package repository

import "testing"

var storageProvider *ApplicationStorageProvider
func connect() IRepositoriesStorage {
	storageProvider = SQLCreateConnection(
		TypeStoragePostgres,
		DSNPostgres,
		nil,
		"postgres",
		"toster123",
		"vkr-db",
		"5432",
		"disable",
	)
	sqlRepository := NewSQLRepository(
		storageProvider,
	)
	return sqlRepository
}

func TestTruncate(t *testing.T) {
	_ = connect()
	storageProvider.SqlDB.Exec("TRUNCATE TABLE repositories CASCADE")
	storageProvider.SqlDB.Exec("TRUNCATE TABLE issues CASCADE")
}

func TestAddFlow(t *testing.T) {

}
