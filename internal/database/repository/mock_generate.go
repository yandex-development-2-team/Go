package repository

//go:generate mockgen -source=db_interface.go -destination=mock_db_interface.go -package=repository DatabaseInterface
