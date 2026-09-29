// package main
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Estructura de Producto
type Product struct {
	ProductCode     string  `json:"productCode"`
	ProductName     string  `json:"productName"`
	ProductLine     string  `json:"productLine"`
	QuantityInStock int     `json:"quantityInStock"`
	BuyPrice        float64 `json:"buyPrice"`
}

// Configuración de la base de datos
func InitDB() *sql.DB {
	// ⚠️ REEMPLAZA "tu_contraseña" POR TU CONTRASENA REAL DE MYSQL
	dsn := "root:Mayra_1508@tcp(127.0.0.1:3306)/classicmodels?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error al configurar la base de datos: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a MySQL. Verifica la contraseña y que el servicio esté corriendo: %v", err)
	}

	fmt.Println("¡Conexión exitosa a MySQL (classicmodels)!")
	return db
}

func main() {
	db := InitDB()
	defer db.Close()

	// 1. Ruta para listar todos los productos o crear uno nuevo
	http.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			getProducts(db, w)
		} else if r.Method == http.MethodPost {
			createProduct(db, w, r)
		} else {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	// 2. Ruta para buscar o eliminar un producto por código
	http.HandleFunc("/api/products/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		code := strings.TrimPrefix(r.URL.Path, "/api/products/")

		if code == "" {
			http.Error(w, "Código de producto no proporcionado", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			getProductByCode(db, w, code)
		} else if r.Method == http.MethodDelete {
			deleteProduct(db, w, code)
		} else {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("🚀 Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- FUNCIONES DE BASE DE DATOS ---

func getProducts(db *sql.DB, w http.ResponseWriter) {
	rows, err := db.Query("SELECT productCode, productName, productLine, quantityInStock, buyPrice FROM products LIMIT 20")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ProductCode, &p.ProductName, &p.ProductLine, &p.QuantityInStock, &p.BuyPrice); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	json.NewEncoder(w).Encode(products)
}

func getProductByCode(db *sql.DB, w http.ResponseWriter, code string) {
	var p Product
	query := "SELECT productCode, productName, productLine, quantityInStock, buyPrice FROM products WHERE productCode = ?"
	err := db.QueryRow(query, code).Scan(&p.ProductCode, &p.ProductName, &p.ProductLine, &p.QuantityInStock, &p.BuyPrice)

	if err == sql.ErrNoRows {
		http.Error(w, "Producto no encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(p)
}

func createProduct(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO products (productCode, productName, productLine, productScale, productVendor, productDescription, quantityInStock, buyPrice, MSRP) VALUES (?, ?, ?, '1:10', 'Vendor', 'Desc', ?, ?, 0.0)"
	_, err := db.Exec(query, p.ProductCode, p.ProductName, p.ProductLine, p.QuantityInStock, p.BuyPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Producto creado con éxito"})
}

func deleteProduct(db *sql.DB, w http.ResponseWriter, code string) {
	res, err := db.Exec("DELETE FROM products WHERE productCode = ?", code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "Producto no encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Producto eliminado"})
}
