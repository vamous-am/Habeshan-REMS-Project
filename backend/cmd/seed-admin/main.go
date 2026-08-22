// backend/cmd/seed-admin/main.go
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/habeshan-rems/backend/internal/admin"
	"github.com/habeshan-rems/backend/internal/common"
)

// One-time CLI tool — promotes an already-registered user to Admin. Not
// an HTTP endpoint; no client can self-escalate. Run once per org, after
// that org's intended admin has registered normally.
//
// Usage:
//   go run ./cmd/seed-admin -org-id=<uuid> -email=someone@example.com
func main() {
	orgIDFlag := flag.String("org-id", "", "organization UUID (required)")
	email := flag.String("email", "", "email of the already-registered user to promote (required)")
	flag.Parse()

	if *orgIDFlag == "" || *email == "" {
		flag.Usage()
		log.Fatal("missing required flags: -org-id and -email")
	}

	orgID, err := uuid.Parse(*orgIDFlag)
	if err != nil {
		log.Fatalf("invalid -org-id: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	db := common.InitDB()

	user, err := admin.NewService(db).PromoteAdmin(admin.PromoteAdminInput{OrgID: orgID, Email: *email})
	if err != nil {
		if errors.Is(err, admin.ErrAdminAlreadyExists) {
			log.Fatalf("aborted: organization %s already has an admin — nothing changed", orgID)
		}
		log.Fatalf("seed-admin failed: %v", err)
	}

	fmt.Printf("✅ %s promoted to admin in org %s\n", user.Email, user.OrgID)
}