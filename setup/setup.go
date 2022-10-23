package setup

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CreateDB executes each file in the SQL directory (in order) to initialize
// the complete database needed to run the server
func CreateDB(fp string) {
	db, err := sql.Open("sqlite3", fp)
	if err != nil {
		fmt.Printf("could not create or connect to db: %v\n", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("could not create or connect to db: %v\n", err)
		os.Exit(1)
	}

	files, err := os.ReadDir("sql")
	if err != nil {
		fmt.Printf("could not read files in ./sql: %v\n", err)
		os.Exit(1)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("db.Begin(): %v\n", err)
		os.Exit(1)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

		f, err := os.Open(filepath.Join("sql", f.Name()))
		if err != nil {
			fmt.Printf("could not read %s, database might be in the incomplete state", f.Name())
			fmt.Println(err)

			fmt.Println("rolling back")
			err := tx.Rollback()
			if err != nil {
				fmt.Println(err)
			}

			os.Exit(1)
		}
		defer f.Close()

		contents, err := io.ReadAll(f)
		if err != nil {
			fmt.Printf("could not read %s, database might be in the incomplete state", f.Name())
			fmt.Println(err)

			fmt.Println("rolling back")
			err := tx.Rollback()
			if err != nil {
				fmt.Println(err)
			}

			os.Exit(1)
		}

		fmt.Printf("executing %s", f.Name())
		_, err = tx.Exec(string(contents))
		if err != nil {
			fmt.Println(": ERROR")
			fmt.Println(err)

			fmt.Println("rolling back")
			err := tx.Rollback()
			if err != nil {
				fmt.Println(err)
			}

			os.Exit(1)
		}

		fmt.Println(": OK")
	}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("db.Commit(): %v\n", err)
		os.Exit(1)
	}
}
