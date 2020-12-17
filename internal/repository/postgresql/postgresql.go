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

const (
	sqlQueryItem string = "select uuid, publication_uuid, published_date, title, description, content, url, language_code from items"
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

// ItemsRepository is the main repository struct
// Use ItemsRepository.pool to make queries
type ItemsRepository struct {
	pool *pgxpool.Pool
}

// NewZapLogger returns logger for repository based on uber zap
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

// GetItemByUUID returns item found by UUID
func (repository *ItemsRepository) GetItemByUUID(ctx context.Context, UUID uuid.UUID) (*entity.Item, error) {
	item := entity.NewItem()
	err := repository.pool.QueryRow(ctx,
		"select uuid, publication_uuid, published_date, title, description, content, url, language_code from items where uuid=$1", UUID).Scan(
		&item.UUID,
		&item.PublicationUUID,
		&item.PublishedDate,
		&item.Title,
		&item.Description,
		&item.Content,
		&item.URL,
		&item.LanguageCode,
	)
	if err != nil && err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetItems returns slice of items pointers
func (repository *ItemsRepository) GetItems(ctx context.Context) ([]*entity.Item, error) {
	return repository.getItems(ctx, sqlQueryItem)
}

// GetItemsByPublicationUUID returns slice of items pointers filtered by PublicationUUID
func (repository *ItemsRepository) GetItemsByPublicationUUID(ctx context.Context, publicationUUID uuid.UUID) ([]*entity.Item, error) {
	queryString := sqlQueryItem + " where publication_uuid=$1"
	return repository.getItems(ctx, queryString, publicationUUID)
}

// GetItemsByPublicationUUIDSortByPublishedDate returns slice of items pointers filtered by PublicationUUID and sorted by publishedDate
func (repository *ItemsRepository) GetItemsByPublicationUUIDSortByPublishedDate(ctx context.Context, publicationUUID uuid.UUID, sortAsc bool) ([]*entity.Item, error) {
	var sortOrder string = "desc"
	if sortAsc {
		sortOrder = "asc"
	}
	queryString := fmt.Sprint(sqlQueryItem, " where publication_uuid=$1 order by published_date ", sortOrder)
	return repository.getItems(ctx, queryString, publicationUUID)
}

// getItems returns slice of items pointers, retrieved using queryString with any parameters
func (repository *ItemsRepository) getItems(ctx context.Context, queryString string, args ...interface{}) ([]*entity.Item, error) {
	rows, err := repository.pool.Query(ctx, queryString, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*entity.Item{}
	for rows.Next() {
		item := entity.NewItem()
		if err := rows.Scan(
			&item.UUID,
			&item.PublicationUUID,
			&item.PublishedDate,
			&item.Title,
			&item.Description,
			&item.Content,
			&item.URL,
			&item.LanguageCode); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return items, nil

}

func (repository *ItemsRepository) Create(ctx context.Context, item *entity.Item) error {
	_, err := repository.pool.Exec(ctx, `insert into items (
		uuid,
		publication_uuid,
		published_date,
		title,
		description,
		content,
		url,
		language_code,
		state_id) select $1, $2, $3, $4, $5, $6, $7, $8, id from item_state where type='valid'`,
		item.UUID, item.PublicationUUID, item.PublishedDate, item.Title, item.Description, item.Content, item.URL, item.LanguageCode)
	return err
}

func (repository *ItemsRepository) Delete(ctx context.Context, UUID uuid.UUID) error {
	result, err := repository.pool.Exec(ctx, "delete from items where uuid=$1", UUID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New(fmt.Sprint("item delete from db execution didn't delete record for UUID ", UUID))
	}
	return err
}

func (repository *ItemsRepository) ItemExists(ctx context.Context, item *entity.Item) (bool, error) {
	var exists bool
	row := repository.pool.QueryRow(ctx, "select exists (select 1 from items where uuid=$1)", item.UUID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	if exists == true {
		return true, nil
	}
	return false, nil
}
