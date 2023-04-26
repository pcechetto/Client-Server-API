package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Dolar struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func GetBid(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	response, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	client := &http.Client{}
	resp, err := client.Do(response)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseData))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
	}

	var data Dolar
	err = json.Unmarshal(responseData, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
	}

	InserirnoBanco(data)

	//fmt.Fprintf(w, "Dólar: "+ data.Usdbrl.Bid)
	w.Header().Add("Content-Type", "application/json")
	w.Write(responseData)

}

var db *sql.DB

func InserirnoBanco(data Dolar) {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Prepara a transação
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()


	_, err = tx.ExecContext(ctx, "INSERT INTO tabela (bid, createdate) VALUES (?, ?)", data.Usdbrl.Bid, data.Usdbrl.CreateDate)
	if err != nil {
		panic(err)
	}

	// Comita a transação
	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Println("Dados salvos com sucesso!")
}

func main() {

	db1, err := sql.Open("sqlite3", "bids.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db = db1

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", GetBid)
	fmt.Println("Server rodando 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
