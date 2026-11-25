package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// ---------------------------------------------------------
// Модели и структура
// ---------------------------------------------------------

type ColumnInfo struct {
	Type     string
	IsNull   bool
	IsPrimaryKey bool // <-- Добавили, чтобы знать имя PK
}

type DBExplorer struct {
	db     *sql.DB
	tables map[string]map[string]ColumnInfo
	mu     sync.RWMutex
}

// ---------------------------------------------------------
// Инициализация
// ---------------------------------------------------------

func NewDBExplorer(db *sql.DB) (http.Handler, error) {
	dbExplorer := &DBExplorer{
		db:     db,
		tables: make(map[string]map[string]ColumnInfo),
	}

	if err := dbExplorer.readAllDB(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	// Регистрация маршрутов
	mux.HandleFunc("GET /{$}", dbExplorer.handleGetAllTables)
	mux.HandleFunc("GET /{table}", dbExplorer.handleGetTable)
	mux.HandleFunc("GET /{table}/{id}", dbExplorer.handleGetRecordByID)
	mux.HandleFunc("PUT /{table}", dbExplorer.handleCreateRecord)
	mux.HandleFunc("PUT /{table}/{$}", dbExplorer.handleCreateRecord)
	mux.HandleFunc("POST /{table}/{id}", dbExplorer.handleUpdateRecord)
	mux.HandleFunc("POST /{table}/{id}/{$}", dbExplorer.handleUpdateRecord)
	mux.HandleFunc("DELETE /{table}/{id}", dbExplorer.handleDeleteRecord)
	mux.HandleFunc("DELETE /{table}/{id}/{$}", dbExplorer.handleDeleteRecord)

	return mux, nil
}

func (dbe *DBExplorer) readAllDB() error {
	rows, err := dbe.db.Query("SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close() // ОБЯЗАТЕЛЬНО

	var tablesName []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tablesName = append(tablesName, name)
	}

	for _, name := range tablesName {
		cols, err := dbe.featchAllColumnsTable(name)
		if err != nil {
			return err
		}
		dbe.tables[name] = cols
	}
	return nil
}

func (dbe *DBExplorer) featchAllColumnsTable(tableName string) (map[string]ColumnInfo, error) {
	query := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", tableName)
	rows, err := dbe.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // ОБЯЗАТЕЛЬНО

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnsInfo := make(map[string]ColumnInfo)

	values := make([]interface{}, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		var colName, colType string
		var isNull, isPK bool

		if val, ok := values[0].([]byte); ok {
			colName = string(val)
		}
		if val, ok := values[1].([]byte); ok {
			colType = string(val)
		}
		if val, ok := values[3].([]byte); ok {
			isNull = string(val) == "YES"
		}
		// Колонка Key содержит "PRI" для первичного ключа
		if val, ok := values[4].([]byte); ok {
			isPK = string(val) == "PRI"
		}

		columnsInfo[colName] = ColumnInfo{
			Type:         colType,
			IsNull:       isNull,
			IsPrimaryKey: isPK,
		}
	}

	return columnsInfo, nil
}

// ---------------------------------------------------------
// Хелперы
// ---------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---------------------------------------------------------
// Хендлеры
// ---------------------------------------------------------

func (dbe *DBExplorer) handleGetAllTables(w http.ResponseWriter, r *http.Request) {
	dbe.mu.RLock()
	tables := make([]string, 0, len(dbe.tables))
	for name := range dbe.tables {
		tables = append(tables, name)
	}
	dbe.mu.RUnlock()

	sort.Strings(tables)

	// Обертка по контракту теста
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"tables": tables,
		},
	})
}

func (dbe *DBExplorer) handleGetTable(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("table")

	dbe.mu.RLock()
	_, exists := dbe.tables[tableName]
	dbe.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	limit, offset := 5, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val >= 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	records, err := dbe.fetchRecords(tableName, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Обертка по контракту теста
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"records": records,
		},
	})
}

func (dbe *DBExplorer) handleGetRecordByID(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("table")
	id := r.PathValue("id")

	dbe.mu.RLock()
	tableInfo, exists := dbe.tables[tableName]
	dbe.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	pkName := getPKName(tableInfo)
	safeTableName := strings.ReplaceAll(tableName, "`", "``")

	query := fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` = ? LIMIT 1", safeTableName, pkName)
	rows, err := dbe.db.Query(query, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	values := make([]interface{}, len(cols))
	pointers := make([]interface{}, len(cols))
	for i := range values {
		pointers[i] = &values[i]
	}

	if !rows.Next() {
		writeError(w, http.StatusNotFound, "record not found")
		return
	}

	if err := rows.Scan(pointers...); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	record := make(map[string]interface{})
	for i, colName := range cols {
		val := values[i]
		if b, ok := val.([]byte); ok {
			record[colName] = string(b)
		} else {
			record[colName] = val
		}
	}

	// Обертка по контракту теста
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"record": record,
		},
	})
}

func (dbe *DBExplorer) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("table")

	dbe.mu.RLock()
	tableInfo, exists := dbe.tables[tableName]
	dbe.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	pkName := getPKName(tableInfo)

	// ❗ НОВОЕ: Проходимся по всем колонкам таблицы
	for colName, colInfo := range tableInfo {
		// Если колонка первичный ключ — пропускаем (AUTO_INCREMENT)
		if colName == pkName {
			continue
		}
		// Если поле есть в JSON — пропускаем, оно уже будет вставлено
		if _, ok := data[colName]; ok {
			continue
		}
		// Если поле НЕ передано, но оно NOT NULL — добавляем дефолтное значение
		if !colInfo.IsNull {
			// В зависимости от типа ставим пустую строку или 0
			if strings.Contains(strings.ToLower(colInfo.Type), "int") {
				data[colName] = 0
			} else {
				data[colName] = ""
			}
		}
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	for key, val := range data {
		if _, ok := tableInfo[key]; !ok {
			continue // Игнорируем неизвестные поля (unkn_field)
		}
		if key == pkName {
			continue // Игнорируем PK
		}

		columns = append(columns, fmt.Sprintf("`%s`", key))
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	if len(columns) == 0 {
		writeError(w, http.StatusBadRequest, "no valid fields provided")
		return
	}

	safeTableName := strings.ReplaceAll(tableName, "`", "``")
	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", safeTableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	result, err := dbe.db.Exec(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	lastID, _ := result.LastInsertId()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			pkName: lastID,
		},
	})
}

func (dbe *DBExplorer) handleUpdateRecord(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("table")
	id := r.PathValue("id")

	dbe.mu.RLock()
	tableInfo, exists := dbe.tables[tableName]
	dbe.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	pkName := getPKName(tableInfo)

	// Если в теле пришел первичный ключ — ошибка
	if _, ok := data[pkName]; ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("field %s have invalid type", pkName))
		return
	}

	var setClauses []string
	var args []interface{}

	for key, val := range data {
		colInfo, ok := tableInfo[key]
		if !ok {
			continue
		}

		if err := validateField(key, val, colInfo); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", key))
		args = append(args, val)
	}

	if len(setClauses) == 0 {
		writeError(w, http.StatusBadRequest, "no valid fields provided")
		return
	}

	safeTableName := strings.ReplaceAll(tableName, "`", "``")
	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?", safeTableName, strings.Join(setClauses, ", "), pkName)
	args = append(args, id)

	result, err := dbe.db.Exec(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Обертка по контракту теста
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"updated": rowsAffected,
		},
	})
}

func (dbe *DBExplorer) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("table")
	id := r.PathValue("id")

	dbe.mu.RLock()
	tableInfo, exists := dbe.tables[tableName]
	dbe.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	pkName := getPKName(tableInfo)
	safeTableName := strings.ReplaceAll(tableName, "`", "``")
	query := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = ?", safeTableName, pkName)

	result, err := dbe.db.Exec(query, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Обертка по контракту теста
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"deleted": rowsAffected,
		},
	})
}

// ---------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------

func getPKName(tableInfo map[string]ColumnInfo) string {
	for name, info := range tableInfo {
		if info.IsPrimaryKey {
			return name
		}
	}
	return "id" // fallback
}

func validateField(colName string, val interface{}, colInfo ColumnInfo) error {
	if val == nil {
		if !colInfo.IsNull {
			return fmt.Errorf("field %s have invalid type", colName)
		}
		return nil
	}

	colTypeLower := strings.ToLower(colInfo.Type)

	if strings.Contains(colTypeLower, "int") {
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("field %s have invalid type", colName)
		}
		return nil
	}

	if strings.Contains(colTypeLower, "char") || strings.Contains(colTypeLower, "text") {
		if _, ok := val.(string); !ok {
			return fmt.Errorf("field %s have invalid type", colName)
		}
		return nil
	}

	if strings.Contains(colTypeLower, "float") || strings.Contains(colTypeLower, "double") || strings.Contains(colTypeLower, "decimal") {
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("field %s have invalid type", colName)
		}
		return nil
	}

	return nil
}

func (dbe *DBExplorer) fetchRecords(tableName string, limit, offset int) ([]map[string]interface{}, error) {
	safeTableName := strings.ReplaceAll(tableName, "`", "``")
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", safeTableName)

	rows, err := dbe.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	values := make([]interface{}, len(cols))
	pointers := make([]interface{}, len(cols))
	for i := range values {
		pointers[i] = &values[i]
	}

	var result []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = val
			}
		}
		result = append(result, rowMap)
	}

	return result, nil
}