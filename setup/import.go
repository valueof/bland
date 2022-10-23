package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/valueof/bland/data"
)

type PinboardSchema struct {
	HREF        string `json:"href"`
	Description string `json:"description"`
	Extended    string `json:"extended"`
	Time        string `json:"time"`
	ToRead      string `json:"toread"`
	Tags        string `json:"tags"`
}

func FromPinboard(dbp, fp string) {
	err := data.ConnectToDB(dbp)
	if err != nil {
		fmt.Printf("couldn't connect to db: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Open(fp)
	if err != nil {
		fmt.Printf("couldn't read %s: %v\n", fp, err)
		os.Exit(1)
	}
	defer f.Close()

	txt, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("couldn't read %s: %v\n", fp, err)
		os.Exit(1)
	}

	var pins []PinboardSchema
	err = json.Unmarshal(txt, &pins)
	if err != nil {
		fmt.Printf("couldn't unmarshal %s: %v\n", fp, err)
		os.Exit(1)
	}

	for _, d := range pins {
		t, err := time.Parse(time.RFC3339, d.Time)
		if err != nil {
			fmt.Printf("couldn't parse time %s, will use now()\n", d.Time)
			t = time.Now()
		}

		b := data.Bookmark{
			URL:         d.HREF,
			Title:       d.Description,
			Description: d.Extended,
			Tags:        d.Tags,
			CreatedAt:   t.Unix(),
			UpdatedAt:   t.Unix(),
			ReadAt:      0,
		}

		if d.ToRead == "no" {
			b.ReadAt = t.Unix()
		}

		tx, err := data.BeginTx(context.Background())
		if err != nil {
			fmt.Printf("data.BeginTx: %v\n", err)
			continue
		}

		_, err = tx.AddBookmark(b)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				fmt.Printf("tx.Rollback: %v\n", err)
			}
			continue
		}

		err = tx.Commit()
		if err != nil {
			fmt.Printf("tx.Rollback: %v\n", err)
		}
	}
}
