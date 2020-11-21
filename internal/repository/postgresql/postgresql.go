package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"

	"go.uber.org/zap"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Config defines database configuration, usable for Viper
type Config struct {
	Name           string `mapstructure:"name"`
	Hostname       string `mapstructure:"hostname"`
	Port           string `mapstructure:"port"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	SSLMode        string `mapstructure:"sslmode"`
	LogLevel       string `mapstructure:"log_level"`
	MinConnections int32  `mapstructure:"min_connections"`
	MaxConnections int32  `mapstructure:"max_connections"`
}

type ItemsRepository struct {
	pool *pgxpool.Pool
}

func NewZapLogger(logger *zap.Logger) *zapadapter.Logger {
	return zapadapter.NewLogger(logger)
}

// New creates database pool configuration
func New(databaseConfig *Config, logger pgx.Logger) (*ItemsRepository, error) {
	postgresDataSource := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Hostname,
		databaseConfig.Name,
		databaseConfig.SSLMode)
	poolConfig, err := pgxpool.ParseConfig(postgresDataSource)
	if err != nil {
		return nil, err
	}
	poolConfig.ConnConfig.Logger = logger
	logLevelMapping := map[string]pgx.LogLevel{
		"trace": pgx.LogLevelTrace,
		"debug": pgx.LogLevelDebug,
		"info":  pgx.LogLevelInfo,
		"warn":  pgx.LogLevelWarn,
		"error": pgx.LogLevelError,
	}
	poolConfig.ConnConfig.LogLevel = logLevelMapping[databaseConfig.LogLevel]
	poolConfig.MaxConns = databaseConfig.MaxConnections
	poolConfig.MinConns = databaseConfig.MinConnections

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}
	return &ItemsRepository{pool: pool}, nil
}

func (repository *ItemsRepository) Create(item *entity.Item) error {
	_, err := repository.pool.Exec(context.Background(), `insert into items (
		uuid,
		publication_uuid,
		published_date,
		title,
		description,
		content,
		source,
		author,
		language_code,
		state_id) select $1, $2, $3, $4, $5, $6, $7, $8, $9, id from item_state where type='valid'`,
		item.UUID, item.PublicationUUID, item.PublishedDate, item.Title, item.Description, item.Content, item.Source, item.Author, item.LanguageCode)
	return err
}

func (repository *ItemsRepository) Delete(UUID uuid.UUID) error {
	result, err := repository.pool.Exec(context.Background(), "delete from items where uuid=$1", UUID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New(fmt.Sprint("item delete from db execution didn't delete record for UUID ", UUID))
	}
	return err
}

func (repository *ItemsRepository) ItemExists(item *entity.Item) (bool, error) {
	var exists bool
	row := repository.pool.QueryRow(context.Background(), "select exists (select 1 from items where uuid=$1)", item.UUID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	if exists == true {
		return true, nil
	}
	return false, nil
}
