package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	storage_go "github.com/supabase-community/storage-go"
	supabase "github.com/supabase-community/supabase-go"
)

var SupabaseClient *supabase.Client
var StorageClient *storage_go.Client

func InitSupabase() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
	}

	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		log.Fatalf("Unable to create Supabase client: %v", err)
	}

	SupabaseClient = client

	// Initialize Storage client
	StorageClient = storage_go.NewClient(supabaseURL+"/storage/v1", supabaseKey, nil)

	log.Println("Supabase client initialized successfully")
}
