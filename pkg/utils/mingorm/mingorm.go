package minggorm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"time"
)

// Model interface for GORM models
type Model interface {
	GetID() interface{}
	SetID(id interface{})
	TableName() string
}

// HookableModel interface extends Model with hooks
type HookableModel interface {
	Model
	BeforeSave(tx *gorm.DB) error
	AfterSave(tx *gorm.DB) error
	BeforeDelete(tx *gorm.DB) error
	AfterDelete(tx *gorm.DB) error
}

// QueryBuilder for constructing complex queries
type QueryBuilder struct {
	db           *gorm.DB
	conditions   []interface{}
	joins        []string
	orderBy      string
	limit        int
	offset       int
	selectFields []string
	groupBy      string
	having       string
	preloads     []string
	distinct     bool
	lock         string
}

// NewQueryBuilder initializes a new QueryBuilder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:           db,
		conditions:   []interface{}{},
		joins:        []string{},
		selectFields: []string{},
		preloads:     []string{},
	}
}

// Where adds conditions to the query
func (qb *QueryBuilder) Where(conditions ...interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, conditions...)
	return qb
}

// Joins adds join clauses to the query
func (qb *QueryBuilder) Joins(joins ...string) *QueryBuilder {
	qb.joins = append(qb.joins, joins...)
	return qb
}

// Order adds an order clause to the query
func (qb *QueryBuilder) Order(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// Limit sets the limit for the query
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the offset for the query
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Select adds select fields to the query
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selectFields = fields
	return qb
}

// Group adds a group by clause to the query
func (qb *QueryBuilder) Group(groupBy string) *QueryBuilder {
	qb.groupBy = groupBy
	return qb
}

// Having adds a having clause to the query
func (qb *QueryBuilder) Having(having string) *QueryBuilder {
	qb.having = having
	return qb
}

// Preload adds preloading of associations
func (qb *QueryBuilder) Preload(preloads ...string) *QueryBuilder {
	qb.preloads = append(qb.preloads, preloads...)
	return qb
}

// Distinct makes the query return distinct results
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
	return qb
}

// Lock adds a table lock clause to the query
func (qb *QueryBuilder) Lock(lock string) *QueryBuilder {
	qb.lock = lock
	return qb
}

// Execute executes the query and stores the result in the result parameter
func (qb *QueryBuilder) Execute(result interface{}) error {
	query := qb.db.Model(result)
	if qb.distinct {
		query = query.Distinct()
	}
	if qb.lock != "" {
		query = query.Clauses(gorm.Expr(qb.lock))
	}
	if len(qb.selectFields) > 0 {
		query = query.Select(qb.selectFields)
	}
	if len(qb.joins) > 0 {
		for _, join := range qb.joins {
			query = query.Joins(join)
		}
	}
	if len(qb.conditions) > 0 {
		query = query.Where(qb.conditions[0], qb.conditions[1:]...)
	}
	if qb.orderBy != "" {
		query = query.Order(qb.orderBy)
	}
	if qb.limit > 0 {
		query = query.Limit(qb.limit)
	}
	if qb.offset > 0 {
		query = query.Offset(qb.offset)
	}
	if qb.groupBy != "" {
		query = query.Group(qb.groupBy)
	}
	if qb.having != "" {
		query = query.Having(qb.having)
	}
	if len(qb.preloads) > 0 {
		for _, preload := range qb.preloads {
			query = query.Preload(preload)
		}
	}
	return query.Find(result).Error
}

// Create inserts a new record into the database
func Create(db *gorm.DB, model HookableModel) error {
	if err := model.BeforeSave(db); err != nil {
		return err
	}
	if err := db.Table(model.TableName()).Create(model).Error; err != nil {
		return err
	}
	return model.AfterSave(db)
}

// Find retrieves a single record by its ID
func Find(db *gorm.DB, model Model) error {
	return db.Table(model.TableName()).First(model, model.GetID()).Error
}

// Update modifies an existing record
func Update(db *gorm.DB, model HookableModel) error {
	if err := model.BeforeSave(db); err != nil {
		return err
	}
	if err := db.Table(model.TableName()).Save(model).Error; err != nil {
		return err
	}
	return model.AfterSave(db)
}

// UpdateFields updates specific fields of a record
func UpdateFields(db *gorm.DB, model HookableModel, fields map[string]interface{}) error {
	if err := model.BeforeSave(db); err != nil {
		return err
	}
	if err := db.Table(model.TableName()).Model(model).Updates(fields).Error; err != nil {
		return err
	}
	return model.AfterSave(db)
}

// Delete removes a record from the database
func Delete(db *gorm.DB, model HookableModel) error {
	if err := model.BeforeDelete(db); err != nil {
		return err
	}
	if err := db.Table(model.TableName()).Delete(model, model.GetID()).Error; err != nil {
		return err
	}
	return model.AfterDelete(db)
}

// SoftDelete marks a record as deleted without removing it from the database
func SoftDelete(db *gorm.DB, model HookableModel) error {
	if !db.Migrator().HasColumn(model.TableName(), "deleted_at") {
		return errors.New("model does not have a 'deleted_at' column")
	}
	if err := model.BeforeDelete(db); err != nil {
		return err
	}
	if err := db.Table(model.TableName()).Updates(map[string]interface{}{"deleted_at": time.Now()}).Error; err != nil {
		return err
	}
	return model.AfterDelete(db)
}

// FindWithConditions retrieves records based on conditions
func FindWithConditions(db *gorm.DB, model Model, conditions []interface{}, result interface{}, fields ...string) error {
	query := db.Table(model.TableName()).Select(fields)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	return query.Find(result).Error
}

// Paginate retrieves a paginated list of records
func Paginate(db *gorm.DB, model Model, page int, pageSize int, result interface{}, builder *QueryBuilder) error {
	offset := (page - 1) * pageSize
	builder.Offset(offset).Limit(pageSize)
	return builder.Execute(result)
}

// GetCount returns the count of records matching the given condition
func GetCount(db *gorm.DB, model Model, conditions []interface{}) (int64, error) {
	var count int64
	err := db.Table(model.TableName()).Where(conditions).Count(&count).Error
	return count, err
}

// Transaction performs a series of operations within a database transaction
func Transaction(db *gorm.DB, txFunc func(tx *gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return txFunc(tx)
	})
}

// RawTransaction executes raw SQL within a transaction
func RawTransaction(db *gorm.DB, txFunc func(tx *gorm.DB) error, query string, args ...interface{}) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(query, args...).Error; err != nil {
			return err
		}
		return txFunc(tx)
	})
}

// BatchCreate creates multiple records in one transaction
func BatchCreate(db *gorm.DB, models interface{}) error {
	return db.Create(models).Error
}

// BatchUpdate updates multiple records in one transaction
func BatchUpdate(db *gorm.DB, models interface{}) error {
	return db.Save(models).Error
}

// BatchDelete deletes multiple records by their IDs
func BatchDelete(db *gorm.DB, model Model, ids []interface{}) error {
	return db.Table(model.TableName()).Delete(model, ids).Error
}

// Migrate performs automatic database migration
func Migrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}

// CacheQuery caches query results for improved performance
func CacheQuery(db *gorm.DB, cacheKey string, builder *QueryBuilder, result interface{}) error {
	// Replace with actual cache implementation
	cachedResult, err := getCache(cacheKey)
	if err == nil {
		return json.Unmarshal(cachedResult, result)
	}
	err = builder.Execute(result)
	if err != nil {
		return err
	}
	cachedResult, err = json.Marshal(result)
	if err != nil {
		return err
	}
	setCache(cacheKey, cachedResult)
	return nil
}

// SetupLogger configures logging for GORM
func SetupLogger(db *gorm.DB, level logger.LogLevel) {
	db.Logger = logger.Default.LogMode(level)
}

// LogCRUDOperation logs CRUD operations with timestamps
func LogCRUDOperation(operation string, model Model, startTime time.Time) {
	duration := time.Since(startTime)
	log.Printf("Operation: %s, Model: %v, Duration: %v", operation, model, duration)
}

// MeasurePerformance logs the time taken for a database operation
func MeasurePerformance(operation func() error) error {
	startTime := time.Now()
	err := operation()
	LogCRUDOperation("Operation", nil, startTime)
	return err
}

// ValidateModel validates a model's fields
func ValidateModel(model interface{}) error {
	// Implement validation logic using a library like go-playground/validator
	return nil
}

// DynamicTableName allows setting a custom table name
func DynamicTableName(db *gorm.DB, model Model, tableName string) *gorm.DB {
	return db.Table(tableName).Model(model)
}

// Snapshot creates a snapshot of the current data
func Snapshot(db *gorm.DB, model Model) (map[string]interface{}, error) {
	var snapshot map[string]interface{}
	result := db.Table(model.TableName()).Where("id = ?", model.GetID()).First(&snapshot)
	if result.Error != nil {
		return nil, result.Error
	}
	return snapshot, nil
}

// RollbackSnapshot rolls back to a previous snapshot
func RollbackSnapshot(db *gorm.DB, model HookableModel, snapshot map[string]interface{}) error {
	if err := model.BeforeSave(db); err != nil {
		return err
	}
	snapshotData, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	var updatedModel map[string]interface{}
	if err = json.Unmarshal(snapshotData, &updatedModel); err != nil {
		return err
	}

	if err = db.Table(model.TableName()).Model(model).Updates(updatedModel).Error; err != nil {
		return err
	}

	return model.AfterSave(db)
}

// EncryptData encrypts the data before saving
func EncryptData(data interface{}) ([]byte, error) {
	key := []byte("examplekey123456") // Replace with secure key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	encrypted := gcm.Seal(nonce, nonce, data.([]byte), nil)
	return encrypted, nil
}

// DecryptData decrypts the data after retrieving
func DecryptData(data []byte, result interface{}) error {
	key := []byte("examplekey123456") // Replace with secure key
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}
	resultData, ok := result.(*[]byte)
	if !ok {
		return errors.New("result must be a pointer to a byte slice")
	}
	*resultData = plain
	return nil
}

// DataVersionControl keeps track of data versions
func DataVersionControl(db *gorm.DB, model HookableModel, version int) error {
	// Ensure the model has a "version" field
	if !db.Migrator().HasColumn(model.TableName(), "version") {
		return errors.New("model does not have a 'version' column")
	}
	return db.Table(model.TableName()).Where("id = ?", model.GetID()).Updates(map[string]interface{}{"version": version}).Error
}

// SoftDeleteByCondition marks records as deleted based on a condition
func SoftDeleteByCondition(db *gorm.DB, model HookableModel, condition string) error {
	if !db.Migrator().HasColumn(model.TableName(), "deleted_at") {
		return errors.New("model does not have a 'deleted_at' column")
	}
	if err := db.Table(model.TableName()).Where(condition).Updates(map[string]interface{}{"deleted_at": time.Now()}).Error; err != nil {
		return err
	}
	return nil
}

// RawQuery executes a raw SQL query
func RawQuery(db *gorm.DB, query string, result interface{}, args ...interface{}) error {
	return db.Raw(query, args...).Scan(result).Error
}

// Example cache functions (Replace with actual cache implementation)
var cache = make(map[string][]byte)

// getCache retrieves data from the cache
func getCache(key string) ([]byte, error) {
	if data, exists := cache[key]; exists {
		return data, nil
	}
	return nil, errors.New("cache miss")
}

// setCache stores data in the cache
func setCache(key string, value []byte) {
	cache[key] = value
}
