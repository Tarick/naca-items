package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"
	opentracing "github.com/opentracing/opentracing-go"
	otLog "github.com/opentracing/opentracing-go/log"

	"go.uber.org/zap"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	sqlQueryItem string = "select uuid, publication_uuid, published_date, title, description, content, url, language_code from items "
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

// Repository is the main repository struct
// Use Repository.pool to make queries
type Repository struct {
	pool   *pgxpool.Pool
	tracer opentracing.Tracer
}

// NewZapLogger returns logger for repository based on uber zap
func NewZapLogger(logger *zap.Logger) *zapadapter.Logger {
	return zapadapter.NewLogger(logger)
}

// New creates database pool configuration
func New(databaseConfig *Config, logger pgx.Logger, tracer opentracing.Tracer) (*Repository, error) {
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
	return &Repository{pool: pool, tracer: tracer}, nil
}

// GetItemByUUID returns item found by UUID
func (repository *Repository) GetItemByUUID(ctx context.Context, UUID uuid.UUID) (*entity.Item, error) {
	query := "select uuid, publication_uuid, published_date, title, description, content, url, language_code from items join item_state is2 on items.state_id=is2.id where is2.type='valid' and uuid=$1"
	span, ctx := repository.setupTracingSpan(ctx, "get-item-by-uuid", query)
	defer span.Finish()
	span.SetTag("item.UUID", UUID)

	item := entity.NewItem()
	err := repository.pool.QueryRow(ctx, query, UUID).Scan(
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
		span.LogFields(
			otLog.Error(err),
		)
		return nil, nil
	}
	if err != nil {
		span.LogFields(
			otLog.Error(err),
		)
		return nil, err
	}
	span.LogKV("event", "fetched item")
	return item, nil
}

// GetItems returns slice of items pointers
func (repository *Repository) GetItems(ctx context.Context) ([]*entity.Item, error) {
	return repository.getItems(ctx, sqlQueryItem)
}

// GetItemsByPublicationUUID returns slice of items pointers filtered by PublicationUUID
func (repository *Repository) GetItemsByPublicationUUID(ctx context.Context, publicationUUID uuid.UUID) ([]*entity.Item, error) {
	queryString := sqlQueryItem + " join item_state is2 on items.state_id=is2.id where is2.type='valid' and publication_uuid=$1"
	return repository.getItems(ctx, queryString, publicationUUID)
}

// GetItemsByPublicationUUIDSortByPublishedDate returns slice of items pointers filtered by PublicationUUID and sorted by publishedDate
func (repository *Repository) GetItemsByPublicationUUIDSortByPublishedDate(ctx context.Context, publicationUUID uuid.UUID, sortAsc bool) ([]*entity.Item, error) {
	var sortOrder string = "desc"
	if sortAsc {
		sortOrder = "asc"
	}
	queryString := fmt.Sprint(sqlQueryItem, " join item_state is2 on items.state_id=is2.id where is2.type='valid' and publication_uuid=$1 order by published_date ", sortOrder)
	return repository.getItems(ctx, queryString, publicationUUID)
}

// getItems returns slice of items pointers, retrieved using queryString with any parameters
func (repository *Repository) getItems(ctx context.Context, queryString string, args ...interface{}) ([]*entity.Item, error) {
	span, ctx := repository.setupTracingSpan(ctx, "get-items", queryString)
	defer span.Finish()

	rows, err := repository.pool.Query(ctx, queryString, args...)
	if err != nil {
		span.LogFields(
			otLog.Error(err),
		)
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
		span.LogFields(
			otLog.Error(err),
		)
		return nil, err
	}
	span.LogKV("event", "fetched items")
	span.LogFields(
		otLog.Int("itemsNumber", len(items)),
	)
	return items, nil
}

func (repository *Repository) Create(ctx context.Context, item *entity.Item) error {
	query := "insert into items (uuid, publication_uuid, published_date, title, description, content, url, language_code, state_id) select $1, $2, $3, $4, $5, $6, $7, $8, id from item_state where type='valid'"
	span, ctx := repository.setupTracingSpan(ctx, "create-item", query)
	defer span.Finish()
	span.SetTag("item.UUID", item.UUID)
	span.SetTag("item.PublicationUUID", item.PublicationUUID)
	_, err := repository.pool.Exec(ctx, query, item.UUID, item.PublicationUUID, item.PublishedDate, item.Title, item.Description, item.Content, item.URL, item.LanguageCode)
	if err != nil {
		span.LogFields(
			otLog.Error(err),
		)
	}
	return err
}

func (repository *Repository) Delete(ctx context.Context, UUID uuid.UUID) error {
	query := "delete from items where uuid=$1"
	span, ctx := repository.setupTracingSpan(ctx, "delete-item-by-uuid", query)
	defer span.Finish()
	span.SetTag("item.UUID", UUID)
	result, err := repository.pool.Exec(ctx, query, UUID)
	if err != nil {
		span.LogFields(
			otLog.Error(err),
		)
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New(fmt.Sprint("item delete from db execution didn't delete record for UUID ", UUID))
	}
	return nil
}

func (repository *Repository) ItemExists(ctx context.Context, item *entity.Item) (bool, error) {
	query := "select exists (select 1 from items where uuid=$1)"
	span, ctx := repository.setupTracingSpan(ctx, "item-exists", query)
	defer span.Finish()
	span.SetTag("item.UUID", item.UUID)

	var exists bool
	row := repository.pool.QueryRow(ctx, query, item.UUID)
	if err := row.Scan(&exists); err != nil {
		span.LogFields(
			otLog.Error(err),
		)
		return false, err
	}
	if exists == true {
		span.LogKV("event", "item exists")
		return true, nil
	}
	return false, nil
}

// Healthcheck is needed for application healtchecks
func (repository *Repository) Healthcheck(ctx context.Context) error {
	var exists bool
	row := repository.pool.QueryRow(ctx, "select exists (select 1 from items limit 1)")
	if err := row.Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	return fmt.Errorf("failure checking access to 'items' table")
}

func (repository *Repository) setupTracingSpan(ctx context.Context, name string, query string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, repository.tracer, name)
	span.SetTag("component", "repository")
	span.SetTag("db.type", "sql")
	span.SetTag("db.query", query)
	return span, ctx
}
