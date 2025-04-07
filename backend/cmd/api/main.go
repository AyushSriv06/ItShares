package main
 
 import (
 	"github.com/AyushSriv06/ItShares/backend/server"
 	"github.com/google/uuid"
 	"log"
 )
 
 func main() {
 	s := server.NewServer()
 	defer s.Shutdown()
 
 	log.Println(uuid.New())
 
 	log.Fatal(s.Run())
 }