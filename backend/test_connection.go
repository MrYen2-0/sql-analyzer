package main

import (
	"log"
	"sql-analyzer/database"
)

func testDatabaseConnection() {
	db := database.GetDB()

	// Probar una consulta simple
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM usuarios").Scan(&count)
	if err != nil {
		log.Fatal("Error al consultar usuarios:", err)
	}

	log.Printf("✅ Número de usuarios en la base de datos: %d", count)

	// Probar las tablas
	tables := []string{"usuarios", "productos", "ventas"}
	for _, table := range tables {
		var exists bool
		query := `
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = 'public' 
                AND table_name = $1
            );`

		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			log.Printf("❌ Error verificando tabla %s: %v", table, err)
		} else if exists {
			log.Printf("✅ Tabla '%s' existe", table)
		} else {
			log.Printf("❌ Tabla '%s' NO existe", table)
		}
	}
}
