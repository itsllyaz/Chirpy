package auth

import (
	"fmt"
	"testing"
)

func TestHashPassword(t *testing.T) string {
	hash, err := HashPassword("password")

	if err != nil {
		t.Fatal("HashPassword returned an error:", err)
	}

	if hash == "" {
		t.Error("expected a hash, got an empty string")
	}

	return hash
}


func TestCheckPasswordHash(t *testing.T){
	hash := TestHashPassword(t)
	match, err := CheckPasswordHash(hash, "password")
	if err != nil{
		t.Error("Can't Compare, something is went wrong")
	}
	if !match{
		fmt.Println("WRONG PASSWORD") 
		
	}
}


